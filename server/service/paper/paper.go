package paper

import (
	"RAGScholar/service/structure"
	"context"
	"fmt"
	"time"

	"github.com/qdrant/go-client/qdrant"
	"golang.org/x/exp/rand"
)

func FetchRandomPapers(ctx context.Context, client *qdrant.Client, collectionName string, limit uint64) ([]structure.SimplifiedEntry, error) {

	points, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Query:          nil, // No query for arbitrary points
		Limit:          &limit,
		WithPayload:    &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch papers: %w", err)
	}

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
		papers = append(papers, paper)
	}

	// Shuffle papers for randomness
	seed := uint64(time.Now().UnixNano())
	rand.Seed(seed)
	rand.Shuffle(len(papers), func(i, j int) {
		papers[i], papers[j] = papers[j], papers[i]
	})

	// Limit to the specified number of papers
	if len(papers) > int(limit) {
		papers = papers[:limit]
	}

	return papers, nil
}

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

// Helper function to extract list of Links from payload
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

type Paper struct {
	ID              string             `json:"id"`
	Updated         string             `json:"updated"`
	Published       string             `json:"published"`
	Title           string             `json:"title"`
	Summary         string             `json:"summary"`
	Authors         []structure.Author `json:"authors"`
	Comment         string             `json:"comment"`
	Links           []structure.Link   `json:"links"`
	PrimaryCategory string             `json:"primaryCategory"`
	Categories      []string           `json:"categories"`
	DOI             string             `json:"doi"`
	JournalRef      string             `json:"journalRef"`
}

func FetchPaperByID(ctx context.Context, client *qdrant.Client, collectionName, paperID string) (*Paper, error) {
	points, err := client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: collectionName,
		Filter: &qdrant.Filter{
			Must: []*qdrant.Condition{
				{
					ConditionOneOf: &qdrant.Condition_Field{
						Field: &qdrant.FieldCondition{
							Key: "id",
							Match: &qdrant.Match{
								MatchValue: &qdrant.Match_Text{
									Text: paperID,
								},
							},
						},
					},
				},
			},
		},
		WithPayload: &qdrant.WithPayloadSelector{SelectorOptions: &qdrant.WithPayloadSelector_Enable{Enable: true}},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query Qdrant: %v", err)
	}

	if len(points) == 0 {
		return nil, fmt.Errorf("paper with ID %s not found", paperID)
	}

	payload := points[0].Payload
	paper := &Paper{
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

	return paper, nil
}

