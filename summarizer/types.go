package summarizer

import (
	"sync"

	"github.com/sashabaranov/go-openai"
	"github.com/tsawler/prose/v3"
)

type Type int

const (
	Extractive             Type = iota // Selects existing sentences from the original text
	AbstractiveOpenAI                  // Generates new text using OpenAI models
	AbstractiveHuggingFace             // Generates new text using Hugging Face models
	Hybrid                             // Use both extractive and abstractive
)

type SummaryResponse struct {
	OriginalText          string   `json:"original_text"`
	Summary               string   `json:"summary"`
	Keywords              []string `json:"keywords"`
	OriginalSentenceCount int      `json:"original_sentence_count"`
	SummarySentenceCount  int      `json:"summary_sentence_count"`
	OriginalWordCount     int      `json:"original_word_count"`
	SummaryWordCount      int      `json:"summary_word_count"`
	CompressionRatio      float64  `json:"compression_rate"`
	SummaryPercentage     float64  `json:"summary_percentage"`
	TargetPercentage      float64  `json:"target_percentage"`
	AbstractiveSummary    bool     `json:"abstractive_summary"`
	RequestedMethod       string   `json:"requested_method"`
	ActualMethod          string   `json:"actual_method"`
	FallBackReason        string   `json:"fallback_reason,omitempty"`
}

type Summarizer interface {
	GenerateSummary() string
	GetResponse() SummaryResponse
	ExtractKeyWords(count int) []string
}

type TextSummarizer struct {
	doc               *prose.Document
	text              string
	summaru           string
	keywords          []string
	summaryRate       float64
	targetPercentage  float64
	summarizerType    Type
	openAIKey         string
	openAIBaseURL     string
	openAIModel       string
	openAIClient      *openai.Client
	clientMutex       sync.Mutex
	huggingFaceConfig HuggingFaceConfig
	requestedMethod   string
	actualMethod      string
	fallbackReason    string
}

type Options func(*TextSummarizer)
