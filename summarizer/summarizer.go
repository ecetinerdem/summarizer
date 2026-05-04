package summarizer

import (
	"fmt"
	"log"
)

func (ts *TextSummarizer) GenerateSummary() string {
	var summary string

	// Set the requested method
	ts.requestedMethod = ts.getMethodName(ts.summarizerType)

	// Decide which summarizer to use
	switch ts.summarizerType {
	case AbstractiveOpenAI:
		summary, err := ts.generateOpenAISummary()
		if err != nil {
			ts.fallbackReason = fmt.Sprintf("Error from openAI abstractive: %v", err)
			log.Printf("Falling back to extractive\n", ts.fallbackReason)
			summary = ts.generateExtractiveSummary()
			ts.actualMethod = "Extractive"
		} else {
			ts.actualMethod = "AbstractiveOpenAI"
		}

	case AbstractiveHuggingFace:

	case Hybrid:
		summary, err := ts.generateHybridSummary()
		if err != nil {
			ts.fallbackReason = fmt.Sprintf("Error from hybrid: %v", err)
			log.Printf("Falling back to extractive\n", ts.fallbackReason)
			summary = ts.generateExtractiveSummary()
			ts.actualMethod = "Extractive"
		} else {
			ts.actualMethod = "Hybrid"
		}

	default:
		// Default to extractive
		summary = ts.generateExtractiveSummary()
		ts.actualMethod = "Extractive"
	}

	ts.summary = summary
	return summary
}

func (ts *TextSummarizer) getMethodName(t Type) string {
	switch t {
	case Extractive:
		return "Extractive"
	case AbstractiveOpenAI:
		return "AbstractiveOpenAI"
	case AbstractiveHuggingFace:
		return "AbstractiveHuggingFace"
	case Hybrid:
		return "Hybrid"
	default:
		return "Unknown"
	}
}
