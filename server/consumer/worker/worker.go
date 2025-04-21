package worker

import (
	"RAGScholar/consumer/structure"
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/google/uuid"
	"github.com/qdrant/go-client/qdrant"
	"golang.org/x/time/rate"
)

const (
	queueName      = "paper-fetcher"
	collectionName = "papers"
	vectorSize     = 768 // Gemini text embeddings are 768-dimensional vectors
)

func GenerateEmbedding(client *genai.Client, model *genai.EmbeddingModel, text string, limiter *rate.Limiter) ([]float32, error) {
	ctx := context.Background()

	if err := limiter.Wait(ctx); err != nil {
		return nil, err
	}

	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("cannot generate embedding for empty text")
	}

	resp, err := model.EmbedContent(ctx, genai.Text(text))
	if err != nil {
		return nil, err
	}

	if len(resp.Embedding.Values) == 0 {
		return nil, fmt.Errorf("received empty embedding from Gemini")
	}

	embedding := make([]float32, len(resp.Embedding.Values))
	for i, v := range resp.Embedding.Values {
		embedding[i] = float32(v)
	}

	log.Printf("Generated embedding with size: %d", len(embedding))

	return embedding, nil
}

func StoreInQdrant(qdrantClient *qdrant.Client, entries []structure.SimplifiedEntry, geminiClient *genai.Client, embeddingModel *genai.EmbeddingModel, limiter *rate.Limiter) error {
	ctx := context.Background()

	if len(entries) == 0 {
		log.Println("Warning: Received empty batch of entries to store")
		return nil
	}

	points := make([]*qdrant.PointStruct, 0, len(entries))

	for _, entry := range entries {
		if strings.TrimSpace(entry.Summary) == "" {
			log.Printf("Entry %s has empty summary, skipping", entry.ID)
			continue
		}

		vector, err := GenerateEmbedding(geminiClient, embeddingModel, entry.Summary, limiter)
		if err != nil {
			log.Printf("Failed to generate embedding for entry %s: %v", entry.ID, err)
			continue
		}

		if len(vector) != vectorSize {
			log.Printf("Warning: Generated vector size (%d) doesn't match expected size (%d)",
				len(vector), vectorSize)
		}

		payload := ConvertToQdrantPayload(entry)

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
		return nil
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

func ConvertToQdrantPayload(entry structure.SimplifiedEntry) map[string]*qdrant.Value {
	payload := make(map[string]*qdrant.Value)

	payload["id"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.ID}}
	payload["updated"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Updated}}
	payload["published"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Published}}
	payload["title"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Title}}
	payload["summary"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Summary}}
	payload["comment"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.Comment}}
	payload["primaryCategory"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.PrimaryCategory}}
	payload["doi"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.DOI}}
	payload["journalRef"] = &qdrant.Value{Kind: &qdrant.Value_StringValue{StringValue: entry.JournalRef}}

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
