package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	// Load env variables, ignoring variables
	_ = godotenv.Load()

	// Define and parse command line flags

	apiMode := flag.Bool("api", false, "Run in api mode")
	port := flag.String("port", "8080", "Port to run the api server on")
	filePath := flag.String("file", "", "File path to summarize")

	// Summary Configuration
	summaryRate := flag.Float64("rate", 0.3, "Summary rate for extractive summarization")
	targetPercent := flag.Float64("percent", 30.0, "Target summary percentage for abstractive summarization")
	summarizerType := flag.String("type", "extractive", "Summarizer type (extractive, abstractive_openai, abstractive_huggingface, hybrid)")

	// OpenAI configuration
	openAIKey := flag.String("openai-key", os.Getenv("OPENAI_API_KEY"), "OpenAI API key")
	openAIModel := flag.String("openai-model", os.Getenv("OPEN_AI_MODEL"), "OpenAI model to use")
	openAIBaseURL := flag.String("openai-url", os.Getenv("OPEN_AI_URL"), "BAse URL for OpenAI API requests")

	// Hugging Face configuration
	huggingFaceKey := flag.String("hf-key", os.Getenv("HUGGING_FACE_KEY"), "Hugging face api key")
	huggingFaceModel := flag.String("hf-model", "", "Hugging face model to use")
	huggingFaceURL := flag.String("hf-url", "", "Custom URL for hugging face inference API")
	maxLength := flag.Int("max-length", 0, "Maximum length for hugging face summary (0 for auto)")
	minLength := flag.Int("min-length", 0, "Minimum length for hugging face summary (0 for auto)")

	flag.Parse()

	// Create a configuration for the app from parsed flags

	config := Config{
		FilePath:         *filePath,
		SummaryRate:      *summaryRate,
		TargetPercent:    *targetPercent,
		SummarizerType:   *summarizerType,
		OpenAIKey:        *openAIKey,
		OpenAIModel:      *openAIModel,
		OpenAIBaseURL:    *openAIBaseURL,
		HuggingFaceKey:   *huggingFaceKey,
		HuggingFaceModel: *huggingFaceModel,
		HuggingFaceURL:   *huggingFaceURL,
		MaxLength:        *maxLength,
		MinLength:        *minLength,
	}

	if *apiMode {
		fmt.Printf("Starting api server on port %s...\n", *port)
	} else {
		if config.FilePath == "" {
			fmt.Println("Error: Please provide a file path with -file flag")
			flag.Usage()
			os.Exit(1)
		}
	}

}
