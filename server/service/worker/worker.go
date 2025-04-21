package worker

import (
	structure "RAGScholar/service/structure"
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// PublishToQueue publishes JSON data to the specified RabbitMQ queue
func PublishToQueue(channel *amqp.Channel, queueName string, jsonData []byte) error {
	return channel.Publish(
		"",        // Exchange
		queueName, // Routing key (queue name)
		false,     // Mandatory
		false,     // Immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        jsonData,
		},
	)
}

// GetData fetches and processes arXiv data
func GetData(topics []string) ([]structure.SimplifiedEntry, []byte, error) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	topic := topics[rng.Intn(len(topics))]
	start := rng.Intn(200) + 1
	log.Printf("Selected Topic and Start: %s\n , %v", topic, start)

	queryURL := "http://export.arxiv.org/api/query?search_query=" + url.QueryEscape(topic) + "&max_results=5&start=" + url.QueryEscape(strconv.Itoa(start))

	res, err := http.Get(queryURL)
	if err != nil {
		log.Printf("HTTP Request Error: %v\n", err)
		return nil, nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Read Body Error: %v\n", err)
		return nil, nil, err
	}

	var feed structure.Feed
	if err := xml.Unmarshal(body, &feed); err != nil {
		log.Printf("XML Parsing Error: %v\n", err)
		return nil, nil, err
	}

	var simplifiedEntries []structure.SimplifiedEntry
	for _, entry := range feed.Entries {
		simplifiedEntry := structure.SimplifiedEntry{
			ID:         entry.ID,
			Updated:    entry.Updated,
			Published:  entry.Published,
			Title:      entry.Title,
			Summary:    entry.Summary,
			Authors:    entry.Authors,
			Comment:    entry.Comment,
			Links:      entry.Links,
			DOI:        entry.DOI,
			JournalRef: entry.JournalRef,
		}

		for _, cat := range entry.Categories {
			simplifiedEntry.Categories = append(simplifiedEntry.Categories, cat.Term)
		}
		simplifiedEntries = append(simplifiedEntries, simplifiedEntry)
	}

	jsonData, err := json.Marshal(simplifiedEntries)
	if err != nil {
		log.Printf("JSON Marshal Error: %v\n", err)
		return nil, nil, err
	}

	return simplifiedEntries, jsonData, nil
}
