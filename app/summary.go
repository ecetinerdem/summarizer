package main

import (
	"fmt"
	"summarizer"
)

func summarizeText(config Config, printOutput bool) (AppResponse, error) {
	// Determine the text source - either from the (CLI mode) or directly provided (API mode)
	var text string
	if config.FilePath != "" {
		// We are in CLI mode
	} else {
		// In API mode
		text = config.Text
	}

	// What summarizer are we going to use
	var summarizerType summarizer.Type
	switch config.SummarizerType {
	case "abstractive_opemai":
		summarizerType = summarizer.AbstractiveOpenAI
	case "abstractive_huggingface":
		summarizerType = summarizer.AbstractiveHuggingFace
	case "hybrid":
		summarizerType = summarizer.Hybrid
	default:
		summarizerType = summarizer.Extractive
	}

	// Set default values if not provided
	if config.SummaryRate == 0 {
		config.SummaryRate = 0.3
	}

	if config.TargetPercent == 0 {
		config.TargetPercent = 30.0
	}

	// Create a slice to hold summarizer options
	var options []summarizer.Options

	// Add summarizer type options
	options = append(options, summarizer.WithSummarizerType(summarizerType))

	// Add OpenAI specific options if an api key provided

	// Add HuggingFace specific options if an api key provided

	// Create the summarizer with text, rate parameters and option
	ts, err := summarizer.NewTextSummarizer(text, config.SummaryRate, config.TargetPercent, options...)

	if err != nil {
		return AppResponse{}, fmt.Errorf("failed to initialize summarizer: %v", err)
	}

	// Generate the summary
	summaryResponse := ts.GetResponse()

	response := AppResponse{
		Summary:               summaryResponse.Summary,
		Keywords:              summaryResponse.Keywords,
		OriginalSentenceCount: summaryResponse.OriginalSentenceCount,
		OriginalWordCount:     summaryResponse.OriginalWordCount,
		SummaryWordCount:      summaryResponse.SummaryWordCount,
		SummarySentenceCount:  summaryResponse.SummarySentenceCount,
		CompressionRatio:      summaryResponse.CompressionRatio,
		SummaryPercentage:     summaryResponse.SummaryPercentage,
	}

	// Print results to console if in CLI mode

	return response, nil

}
