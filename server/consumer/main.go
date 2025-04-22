package main

import (
	"RAGScholar/consumer/structure"
	"RAGScholar/consumer/worker"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/qdrant/go-client/qdrant"
	amqp "github.com/rabbitmq/amqp091-go"
	"golang.org/x/time/rate"
	"google.golang.org/api/option"
)

const (
	queueName      = "paper-fetcher"
	collectionName = "papers"
	vectorSize     = 768 // Gemini text embeddings are 768-dimensional vectors
)

func initQdrant() (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	ctx := context.Background()

	
	exists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection exists: %w", err)
	}

	if exists {
		log.Printf("Collection '%s' already exists", collectionName)
		return client, nil
	}

	log.Printf("Creating collection '%s' with vector size %d", collectionName, vectorSize)
	err = client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: collectionName,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     vectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	log.Printf("Collection '%s' created successfully", collectionName)
	return client, nil
}

func main() {
	// Initialize RabbitMQ
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672")
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %v", err)
	}
	defer channel.Close()

	// Initialize Qdrant
	qdrantClient, err := initQdrant()
	if err != nil {
		log.Fatalf("Failed to initialize Qdrant: %v", err)
	}
	defer qdrantClient.Close()

	// Initialize Gemini client
	ctx := context.Background()
	geminiAPIKey := "AIzaSyCwMtDVOhs6n9x2MdcwmRFiieoBGKIrGVU"
	if geminiAPIKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable not set")
	}

	geminiClient, err := genai.NewClient(ctx, option.WithAPIKey(geminiAPIKey))
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}
	defer geminiClient.Close()

	// Create an embedding model instance
	embeddingModel := geminiClient.EmbeddingModel("models/embedding-001")
	if embeddingModel == nil {
		log.Fatal("Failed to initialize Gemini embedding model")
	}

	limiter := rate.NewLimiter(rate.Every(time.Minute/60), 1)

	// Worker pool
	const numWorkers = 10 
	taskChan := make(chan []structure.SimplifiedEntry, 10)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerId)
			for entries := range taskChan {
				err := worker.StoreInQdrant(qdrantClient, entries, geminiClient, embeddingModel, limiter)
				if err != nil {
					log.Printf("Worker %d: Failed to store entries in Qdrant: %v", workerId, err)
				} else {
					log.Printf("Worker %d: Stored %d entries in Qdrant", workerId, len(entries))
				}
			}
			log.Printf("Worker %d stopped", workerId)
		}(i)
	}

	// Consume messages
	messages, err := channel.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to consume messages: %v", err)
	}

	log.Println("Connected to RabbitMQ, waiting for messages...")

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// Main loop
loop:
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				log.Println("Message channel closed")
				break loop
			}

			var entries []structure.SimplifiedEntry
			if err := json.Unmarshal(message.Body, &entries); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			log.Printf("Received message with %d entries", len(entries))

			taskChan <- entries

		case sig := <-sigchan:
			log.Printf("Received signal %v, shutting down...", sig)
			break loop
		}
	}

	close(taskChan)
	log.Println("Waiting for in-flight tasks to complete...")
	wg.Wait()
	log.Println("Shutdown complete")
}
