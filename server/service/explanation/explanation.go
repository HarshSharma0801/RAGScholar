package explanation

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
)

// SystemPrompt is the default system prompt for the Gemini model
const SystemPrompt = `You are a helpful academic assistant. Your task is to explain the given text from a research paper.
Provide a clear, concise explanation that:
1. Summarizes the key points or concepts in the text
2. Explains any technical terms or jargon
3. Places the text in the broader context of the research field
4. Highlights the significance or implications of the content

Keep your explanation focused, accurate, and helpful for someone trying to understand this research.`

func ExplainText(ctx context.Context, client *genai.Client, selectedText string, paperContext string) (string, error) {
	model := client.GenerativeModel("gemini-1.5-pro")
	if model == nil {
		return "", fmt.Errorf("failed to initialize Gemini model")
	}

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(SystemPrompt),
		},
	}

	prompt := fmt.Sprintf("The following text is from a research paper titled '%s':\n\n%s\n\nPlease explain this text.",
		paperContext, selectedText)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating explanation: %v", err)
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received empty response from Gemini")
	}

	explanation, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Gemini")
	}

	return string(explanation), nil
}

func CustomExplainText(ctx context.Context, client *genai.Client, selectedText string, paperContext string, customPrompt string) (string, error) {
	model := client.GenerativeModel("gemini-1.5-flash")
	if model == nil {
		return "", fmt.Errorf("failed to initialize Gemini model")
	}

	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(customPrompt),
		},
	}

	prompt := fmt.Sprintf("The following text is from a research paper titled '%s':\n\n%s",
		paperContext, selectedText)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		log.Printf("Error generating explanation: %v", err)
		return "", err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("received empty response from Gemini")
	}

	explanation, ok := resp.Candidates[0].Content.Parts[0].(genai.Text)
	if !ok {
		return "", fmt.Errorf("unexpected response format from Gemini")
	}

	return string(explanation), nil
}
