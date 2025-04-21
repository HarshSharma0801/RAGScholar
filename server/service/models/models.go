package models

type TextAnalysisRequest struct {
	SelectedText  string `json:"selectedText"` 
	PaperContext  string `json:"paperContext"`  
	SearchQuery   string `json:"searchQuery"`  
	CustomPrompt  string `json:"customPrompt"`  
}

type TextAnalysisResponse struct {
	RelatedPapers []PaperResult `json:"relatedPapers"` 
	Explanation   string        `json:"explanation"` 
}

type PaperResult struct {
	Paper interface{} `json:"paper"`
	Score float32     `json:"score"` 
}
