package summarizer

import (
	"context"

	"github.com/sashabaranov/go-openai"
)

func (ts *TextSummarizer) generateHybridSummary() (string, error) {
	extractiveSummary := ts.generateExtractiveSummary()

	// Try open ai if configured
	if ts.openAIKey != "" {
		client, err := ts.getOpenAIClient()
		if err == nil {
			resp, err := client.CreateChatCompletion(
				context.Background(),
				openai.ChatCompletionRequest{
					Model: ts.openAIModel,
					Messages: []openai.ChatCompletionMessage{
						{
							Role: openai.ChatMessageRoleSystem,
							Content: "You are a sophisticated text refinement system. The extracted key sentences" +
								"Rewrite them into coherent, cluent summary. Connenct ideas smoothly," +
								"eliminate redundancy, and ensure the text flows naturally while preserving all key information",
						},
						{
							Role:    openai.ChatMessageRoleUser,
							Content: extractiveSummary,
						},
					},
					Temperature: 0.3,
				},
			)
			if err == nil && len(resp.Choices) > 0 {
				return resp.Choices[0].Message.Content, nil
			}
		}
	}

	// Fallback to hugging face configured
	if ts.huggingFaceConfig.APIKey != "" {
		summary, err := ts.generateHuggingFaceSummaryFromText(extractiveSummary)
		if err == nil {
			return summary, nil
		}
	}

	return extractiveSummary, nil
}

func (ts *TextSummarizer) generateHuggingFaceSummaryFromText(text string) (string, error) {
	originalText := ts.text

	ts.text = text

	summary, err := ts.generateHuggingFaceSummary()
	ts.text = originalText
	return summary, err
}
