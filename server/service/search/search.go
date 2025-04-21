package search

import (
	"RAGScholar/service/structure"
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/qdrant/go-client/qdrant"
	"golang.org/x/time/rate"
)

// GenerateEmbedding creates an embedding vector for the given text using Gemini API
func GenerateEmbedding(client *genai.Client, model *genai.EmbeddingModel, text string, limiter *rate.Limiter) ([]float32, error) {
	ctx := context.Background()

	if err := limiter.Wait(ctx); err != nil {
		return nil, err
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

// SimilaritySearch retrieves papers and ranks them based on relevance to the query text
func SimilaritySearch(ctx context.Context, client *qdrant.Client, collectionName string, 
	queryText string, limit uint64) ([]structure.SimplifiedEntry, error) {
	
	// First, fetch papers using the existing Query method
	points, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          nil, // No query filter to get a diverse set of papers
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch papers: %w", err)
	}

	// Convert points to papers
	var papers []structure.SimplifiedEntry
	for _, point := range points {
		payload := point.Payload
		paper := structure.SimplifiedEntry{
			ID:              getStringFromPayload(payload, "id"),
			Updated:         getStringFromPayload(payload, "updated"),
			Published:       getStringFromPayload(payload, "published"),
			Title:           getStringFromPayload(payload, "title"),
			Summary:         getStringFromPayload(payload, "summary"),
			Authors:         getAuthorsFromPayload(payload, "authors"),
			Comment:         getStringFromPayload(payload, "comment"),
			Links:           getLinksFromPayload(payload, "links"),
			PrimaryCategory: getStringFromPayload(payload, "primaryCategory"),
			Categories:      getStringListFromPayload(payload, "categories"),
			DOI:             getStringFromPayload(payload, "doi"),
			JournalRef:      getStringFromPayload(payload, "journalRef"),
		}
		
		// Calculate a simple relevance score based on text matching
		// This is a fallback since we're not using vector search
		relevanceScore := calculateRelevanceScore(paper, queryText)
		paper.Score = relevanceScore
		
		papers = append(papers, paper)
	}

	// Sort papers by relevance score (descending)
	sort.Slice(papers, func(i, j int) bool {
		return papers[i].Score > papers[j].Score
	})

	// Limit to the requested number of papers
	if len(papers) > int(limit) {
		papers = papers[:limit]
	}

	return papers, nil
}

// calculateRelevanceScore computes a simple text-based relevance score
func calculateRelevanceScore(paper structure.SimplifiedEntry, queryText string) float32 {
	queryLower := strings.ToLower(queryText)
	titleLower := strings.ToLower(paper.Title)
	summaryLower := strings.ToLower(paper.Summary)
	
	var score float32 = 0
	
	// Check title for query terms
	if strings.Contains(titleLower, queryLower) {
		score += 10.0 // High weight for title matches
	}
	
	// Check summary for query terms
	if strings.Contains(summaryLower, queryLower) {
		score += 5.0 // Medium weight for summary matches
	}
	
	// Check categories for relevance
	for _, category := range paper.Categories {
		if strings.Contains(strings.ToLower(category), queryLower) {
			score += 3.0 // Lower weight for category matches
			break
		}
	}
	
	return score
}

// Helper functions to extract data from Qdrant payload

func getStringFromPayload(payload map[string]*qdrant.Value, key string) string {
	if value, ok := payload[key]; ok && value.GetStringValue() != "" {
		return value.GetStringValue()
	}
	return ""
}

func getAuthorsFromPayload(payload map[string]*qdrant.Value, key string) []structure.Author {
	var authors []structure.Author
	if value, ok := payload[key]; ok && value.GetListValue() != nil {
		for _, v := range value.GetListValue().Values {
			if structVal := v.GetStructValue(); structVal != nil {
				if nameVal, exists := structVal.Fields["name"]; exists && nameVal.GetStringValue() != "" {
					authors = append(authors, structure.Author{Name: nameVal.GetStringValue()})
				}
			}
		}
	}
	return authors
}

func getLinksFromPayload(payload map[string]*qdrant.Value, key string) []structure.Link {
	var links []structure.Link
	if value, ok := payload[key]; ok && value.GetListValue() != nil {
		for _, v := range value.GetListValue().Values {
			if structVal := v.GetStructValue(); structVal != nil {
				link := structure.Link{
					Href: getStructString(structVal, "href"),
					Rel:  getStructString(structVal, "rel"),
					Type: getStructString(structVal, "type"),
				}
				links = append(links, link)
			}
		}
	}
	return links
}

func getStringListFromPayload(payload map[string]*qdrant.Value, key string) []string {
	var strings []string
	if value, ok := payload[key]; ok && value.GetListValue() != nil {
		for _, v := range value.GetListValue().Values {
			if str := v.GetStringValue(); str != "" {
				strings = append(strings, str)
			}
		}
	}
	return strings
}

func getStructString(structVal *qdrant.Struct, key string) string {
	if val, exists := structVal.Fields[key]; exists && val.GetStringValue() != "" {
		return val.GetStringValue()
	}
	return ""
}
