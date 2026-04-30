package summarizer

func (ts *TextSummarizer) GenerateSummary() string {
	var summary string

	// Set the requested method
	ts.requestedMethod = ts.getMethodName(ts.summarizerType)

	// Decide which summarizer to use
	switch ts.summarizerType {
	case AbstractiveOpenAI:

	case AbstractiveHuggingFace:

	case Hybrid:

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
