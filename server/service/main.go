package main

import (
	"RAGScholar/service/paper"
	"RAGScholar/service/structure"
	"RAGScholar/service/worker"
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/generative-ai-go/genai"
	"github.com/qdrant/go-client/qdrant"
	amqp "github.com/rabbitmq/amqp091-go"
	"google.golang.org/api/option"
)

const queueName = "paper-fetcher"

func main() {
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

	_, err = channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		log.Fatalf("Failed to declare queue: %v", err)
	}

	log.Print("Connected With Producer RabbitMQ!")

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

	log.Print("Connected With Gemini Client!")

	qDrantclient, err := qdrant.NewClient(&qdrant.Config{
		Host: "localhost",
		Port: 6334,
	})
	if err != nil {
		log.Fatalf("failed to create Qdrant client: %v", err)
	}
	defer qDrantclient.Close()

	log.Print("Connected With Qdrant Client!")

	router := gin.Default()

	router.GET("/data", func(ctx *gin.Context) {
		var wg sync.WaitGroup
		dataChan := make(chan struct {
			entries []structure.SimplifiedEntry
			data    []byte
			err     error
		}, 10)

		// Start 10 goroutines to fetch data
		for i := 0; i < 50; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				entries, data, err := worker.GetData(structure.Topics)
				dataChan <- struct {
					entries []structure.SimplifiedEntry
					data    []byte
					err     error
				}{entries, data, err}
				log.Printf("Got data from goroutine %v", i)
			}()
		}

		go func() {
			wg.Wait()
			close(dataChan)
		}()

		var allEntries []structure.SimplifiedEntry
		for result := range dataChan {
			if result.err != nil {
				log.Printf("Data Fetch Error: %v\n", result.err)
				return
			}

			err := worker.PublishToQueue(channel, queueName, result.data)
			if err != nil {
				log.Printf("RabbitMQ Publish Error: %v\n", err)
				return
			}

			log.Printf("Published %d entries to RabbitMQ queue: %s\n", len(result.entries), queueName)
			allEntries = append(allEntries, result.entries...)
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Published all messages to RabbitMQ", "total_entries": len(allEntries)})
	})

	router.GET("/check", func(ctx *gin.Context) {

		ctx.JSON(http.StatusOK, gin.H{"message": "Welcome to RAGScholar API!"})

	})

	router.GET("/", func(ctx *gin.Context) {
		collectionName := "papers"
		limit := uint64(10)

		papers, err := paper.FetchRandomPapers(context.Background(), qDrantclient, collectionName, limit)
		if err != nil {
			log.Printf("Failed to fetch random papers: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch papers"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"papers": papers})
	})
	log.Fatal(router.Run(":8040"))
}

