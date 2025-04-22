package main

import (
	"RAGScholar/service/explanation"
	"RAGScholar/service/models"
	"RAGScholar/service/paper"
	"RAGScholar/service/search"
	"RAGScholar/service/structure"
	"RAGScholar/service/worker"
	"context"
	"log"
	"net/http"
	"sync"

	"github.com/gin-contrib/cors"
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

	// Initialize embedding model
	embeddingModel := geminiClient.EmbeddingModel("models/embedding-001")
	if embeddingModel == nil {
		log.Fatal("Failed to initialize Gemini embedding model")
	}

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

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

	// New route for fetching a paper by ID
	router.GET("/paper/:id", func(ctx *gin.Context) {
		paperID := ctx.Param("id")
		collectionName := "papers"

		// Query Qdrant for the paper by ID
		paper, err := paper.FetchPaperByID(context.Background(), qDrantclient, collectionName, paperID)
		if err != nil {
			log.Printf("Failed to fetch paper with ID %s: %v", paperID, err)
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Paper not found"})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"paper": paper})
	})

	// New route for analyzing selected text and finding related papers
	router.POST("/analyze", func(ctx *gin.Context) {
		var request models.TextAnalysisRequest
		if err := ctx.ShouldBindJSON(&request); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate request
		if request.SelectedText == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Selected text is required"})
			return
		}

		// Use search query if provided, otherwise use selected text
		searchQuery := request.SearchQuery
		if searchQuery == "" {
			searchQuery = request.SelectedText
		}

		// Perform similarity search to find related papers
		collectionName := "papers"
		limit := uint64(5) // Get top 5 papers as requested

		relatedPapers, err := search.SimilaritySearch(context.Background(), qDrantclient, collectionName, searchQuery, limit)
		if err != nil {
			log.Printf("Failed to perform similarity search: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find related papers"})
			return
		}

		// Generate explanation for the selected text
		var textExplanation string
		if request.CustomPrompt != "" {
			// Use custom prompt if provided
			textExplanation, err = explanation.CustomExplainText(context.Background(), geminiClient, request.SelectedText, request.PaperContext, request.CustomPrompt)
		} else {
			// Use default prompt
			textExplanation, err = explanation.ExplainText(context.Background(), geminiClient, request.SelectedText, request.PaperContext)
		}

		if err != nil {
			log.Printf("Failed to generate explanation: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate explanation"})
			return
		}

		// Prepare response
		var paperResults []models.PaperResult
		for _, paper := range relatedPapers {
			paperResults = append(paperResults, models.PaperResult{
				Paper: paper,
				Score: paper.Score,
			})
		}

		response := models.TextAnalysisResponse{
			RelatedPapers: paperResults,
			Explanation:   textExplanation,
		}

		ctx.JSON(http.StatusOK, response)
	})

	log.Fatal(router.Run(":8040"))
}
