package main

import (
	"RAGScholar/service/structure"
	"RAGScholar/service/worker"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	amqp "github.com/rabbitmq/amqp091-go"
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

	router := gin.Default()

	router.GET("/", func(ctx *gin.Context) {
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

	log.Fatal(router.Run(":8040"))
}
