package models

// TextAnalysisRequest represents the request for analyzing selected text from a paper
type TextAnalysisRequest struct {
	SelectedText  string `json:"selectedText"`  // The text selected by the user from the paper
	PaperContext  string `json:"paperContext"`  // Title or context of the paper the text is from
	SearchQuery   string `json:"searchQuery"`   // Optional query to use for similarity search instead of selected text
	CustomPrompt  string `json:"customPrompt"`  // Optional custom prompt for the explanation
}

// TextAnalysisResponse represents the response for the text analysis
type TextAnalysisResponse struct {
	RelatedPapers []PaperResult `json:"relatedPapers"` // Related papers based on similarity search
	Explanation   string        `json:"explanation"`   // Explanation of the selected text
}

// PaperResult represents a paper with its similarity score
type PaperResult struct {
	Paper interface{} `json:"paper"` // The paper data
	Score float32     `json:"score"` // Similarity score
}
