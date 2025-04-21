package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
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

// SimplifiedEntry matches the structure of the incoming JSON data
type SimplifiedEntry struct {
	ID              string   `json:"id"`
	Updated         string   `json:"updated"`
	Published       string   `json:"published"`
	Title           string   `json:"title"`
	Summary         string   `json:"summary"`
	Authors         []Author `json:"authors"`
	Comment         string   `json:"comment"`
	Links           []Link   `json:"links"`
	PrimaryCategory string   `json:"primaryCategory"`
	Categories      []string `json:"categories"`
	DOI             string   `json:"doi"`
	JournalRef      string   `json:"journalRef"`
}

type Author struct {
	Name string `json:"name"`
}

type Link struct {
	Href string `json:"href"`
	Rel  string `json:"rel"`
	Type string `json:"type"`
}

// initQdrant initializes the Qdrant client and creates the papers collection
func initQdrant() (*qdrant.Client, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	ctx := context.Background()

	// Check if collection exists
	exists, err := client.CollectionExists(ctx, collectionName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if collection exists: %w", err)
	}

	if exists {
		log.Printf("Collection '%s' already exists", collectionName)
		return client, nil
	}

	// Create collection if it doesn't exist
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

// generateEmbedding uses Google Gemini API to generate a vector for the given text
func generateEmbedding(client *genai.Client, model *genai.EmbeddingModel, text string, limiter *rate.Limiter) ([]float32, error) {
	ctx := context.Background()

	// Wait for rate limiter
	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	// Make sure the text isn't empty
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("cannot generate embedding for empty text")
	}

	// Generate embedding using Gemini
	resp, err := model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}

	// Validate response
	if len(resp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("received empty embedding from Gemini")
	}

	// Convert the embedding to float32
	embedding := make([]float32, len(resp.Embedding.Values))
	for i, v := range resp.Embedding.Values {
		embedding[i] = float32(v)
	}

	// Log the embedding size for debugging
	log.Printf("Generated embedding with size: %d", len(embedding))

	return embedding, nil
}

// storeInQdrant stores a batch of SimplifiedEntry data in Qdrant
func storeInQdrant(qdrantClient *qdrant.Client, entries []SimplifiedEntry, geminiClient *genai.Client, embeddingModel *genai.EmbeddingModel, limiter *rate.Limiter) error {
	ctx := context.Background()

	if len(entries) == 0 {
		log.Println("Warning: Received empty batch of entries to store")
		return nil // Nothing to store, no need to send request
	}

	points := make([]*qdrant.PointStruct, 0, len(entries)) // Using capacity but initializing empty

	for _, entry := range entries {
		// Skip entries with empty summaries
		if strings.TrimSpace(entry.Summary) == "" {
			log.Printf("Entry %s has empty summary, skipping", entry.ID)
			continue
		}

		// Generate embedding for summary
		vector, err := generateEmbedding(geminiClient, embeddingModel, entry.Summary, limiter)
		if err != nil {
			log.Printf("Failed to generate embedding for entry %s: %v", entry.ID, err)
			continue // Skip this entry but try others
		}

		if len(vector) != vectorSize {
			log.Printf("Warning: Generated vector size (%d) doesn't match expected size (%d)",
				len(vector), vectorSize)
		}

		// Convert entry to proper Qdrant payload
		payload := convertToQdrantPayload(entry)

		point := &qdrant.PointStruct{
			Id: &qdrant.PointId{
				PointIdOptions: &qdrant.PointId_Uuid{
					Uuid: uuid.New().String(),
				},
			},
			Vectors: &qdrant.Vectors{
				VectorsOptions: &qdrant.Vectors_Vector{
					Vector: &qdrant.Vector{
						Data: vector,
					},
				},
			},
			Payload: payload,
		}

		points = append(points, point)
	}

	if len(points) == 0 {
		log.Println("Warning: No valid points to store after processing")
		return nil // Don't send empty request
	}

	log.Printf("Sending upsert request with %d points", len(points))

	_, err := qdrantClient.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: collectionName,
		Points:         points,
	})

	if err != nil {
		log.Printf("Upsert error details: %v", err)
		return err
	}

	log.Printf("Successfully stored %d points in Qdrant", len(points))
	return nil
}

// convertToQdrantPayload converts a SimplifiedEntry to the expected Qdrant payload format
func convertToQdrantPayload(entry SimplifiedEntry) map[string]*qdrant.Value {
	payload := make(map[string]*qdrant.Value)

	// Add string fields
	payload["id"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.ID}}
	payload["updated"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Updated}}
	payload["published"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Published}}
	payload["title"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Title}}
	payload["summary"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Summary}}
	payload["comment"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Comment}}
	payload["primaryCategory"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.PrimaryCategory}}
	payload["doi"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.DOI}}
	payload["journalRef"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.JournalRef}}

	// Add authors array
	authorsArray := make([]*qdrant.Value, len(entry.Authors))
	for i, author := range entry.Authors {
		authorMap := map[string]*qdrant.Value{
			"name": {Kind: &qdrant.Value_StringValue{StringValue: author.Name}},
		}
		authorsArray[i] = &qdrant.Value{Kind: &qdrant.Value_StructValue{
			StructValue: &qdrant.Struct{
				Fields: authorMap,
			},
		}}
	}
	payload["authors"] = &qdrant.Value{Kind: &qdrant.Value_ListValue{
		ListValue: &qdrant.ListValue{
			Values: authorsArray,
		},
	}}

	// Add links array
	linksArray := make([]*qdrant.Value, len(entry.Links))
	for i, link := range entry.Links {
		linkMap := map[string]*qdrant.Value{
			"href": {Kind: &qdrant.Value_StringValue{StringValue: link.Href}},
			"rel":  {Kind: &qdrant.Value_StringValue{StringValue: link.Rel}},
			"type": {Kind: &qdrant.Value_StringValue{StringValue: link.Type}},
		}
		linksArray[i] = &qdrant.Value{Kind: &qdrant.Value_StructValue{
			StructValue: &qdrant.Struct{
				Fields: linkMap,
			},
		}}
	}
	payload["links"] = &qdrant.Value{Kind: &qdrant.Value_ListValue{
		ListValue: &qdrant.ListValue{
			Values: linksArray,
		},
	}}

	// Add categories array
	categoriesArray := make([]*qdrant.Value, len(entry.Categories))
	for i, category := range entry.Categories {
		categoriesArray[i] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: category}}
	}
	payload["categories"] = &qdrant.Value{Kind: &qdrant.Value_ListValue{
		ListValue: &qdrant.ListValue{
			Values: categoriesArray,
		},
	}}

	return payload
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

	// Rate limiter: 60 requests per minute (adjust based on your Gemini API limits)
	limiter := rate.NewLimiter(rate.Every(time.Minute/60), 1)

	// Worker pool
	const numWorkers = 4 // Adjust based on system resources
	taskChan := make(chan []SimplifiedEntry, 10)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerId int) {
			defer wg.Done()
			log.Printf("Worker %d started", workerId)
			for entries := range taskChan {
				err := storeInQdrant(qdrantClient, entries, geminiClient, embeddingModel, limiter)
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

			// Parse JSON message
			var entries []SimplifiedEntry
			if err := json.Unmarshal(message.Body, &entries); err != nil {
				log.Printf("Failed to parse message: %v", err)
				continue
			}

			log.Printf("Received message with %d entries", len(entries))

			// Send entries to worker pool
			taskChan <- entries

		case sig := <-sigchan:
			log.Printf("Received signal %v, shutting down...", sig)
			break loop
		}
	}

	// Close task channel and wait for workers to finish
	close(taskChan)
	log.Println("Waiting for in-flight tasks to complete...")
	wg.Wait()
	log.Println("Shutdown complete")
}
