package summarizer

import (
	"math"
	"strings"
	"unicode"

	"github.com/bbalet/stopwords"
	"github.com/tsawler/prose/v3"
)

func (ts *TextSummarizer) generateExtractiveSummary() string {
	sentences := ts.doc.Sentences()
	sentenceCount := len(sentences)

	if sentenceCount == 0 {
		return ""
	}

	// Determine the number of sentences to keep
	keepCount := int(float64(sentenceCount) * ts.summaryRate)

	if keepCount < 1 {
		keepCount = 1
	}

	if keepCount > sentenceCount {
		keepCount = sentenceCount
	}

	// Calculate the TF-IDF for all words in the document
	wordTFIDF := calculateTFIDF(ts.doc)

	// Score sentences using multiple features
	sentenceScores := make(map[string]float64)

	// Get named entities from the document for entity based scoring
	entities := extractNamedEntities(ts.doc)

	// Calculate sentence position weights (first and last paragraph typically more important)
	positionWeights := calculatePositionWeights(sentences, sentenceCount)

	// Create a map of sentences containing important entities
	entitySentences := mapEntitiesToSentences(sentences, entities)

	// Find title keywords if available
	titleKeyWords := extractTitleKeyWords(sentences)
	// Create a graph representation for Textrank like algorithm

	// Analyze sentence relationships

	// Sort by score (descending)

	// Take top-ranked sentences

}

func calculateTFIDF(doc *prose.Document) map[string]float64 {
	wordTF := make(map[string]int)
	wordIDF := make(map[string]float64)
	wordTFIDF := make(map[string]float64)
	totalWords := 0

	// Calculate term frequency
	for _, token := range doc.Tokens() {
		word := strings.ToLower(token.Text)
		if len(word) <= 2 || isStopWord(word) || isAlphaNumeric(word) {
			continue
		}

		wordTF[word]++
		totalWords++
	}

	// Split the document into pseudo-documents for idf calculation
	pseudoDocs := splitIntoPseudoDocuments(doc)
	docCount := len(pseudoDocs)

	// Calculate document frequency
	for word := range wordTF {
		docFreq := 0
		for _, pseudoDoc := range pseudoDocs {
			if strings.Contains(strings.ToLower(pseudoDoc), word) {
				docFreq++
			}
		}
		// Calculate IDF
		if docFreq > 0 {
			wordIDF[word] = math.Log(float64(docCount) / float64(docFreq))
		} else {
			wordIDF[word] = 0
		}

		// Calculate TF-IDF
		wordTFIDF[word] = float64(wordTF[word]) * wordIDF[word] / float64(totalWords)
	}

	return wordTFIDF

}

func isStopWord(word string) bool {
	cleaned := stopwords.CleanString(strings.TrimSpace(word), "en", false)

	return strings.TrimSpace(cleaned) == ""
}

func isAlphaNumeric(word string) bool {
	for _, r := range word {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

func splitIntoPseudoDocuments(doc *prose.Document) []string {
	sentences := doc.Sentences()

	// If document is short use sentences as pseudo docs

	if len(sentences) < 10 {
		docs := make([]string, len(sentences))
		for i, sent := range sentences {
			docs[i] = sent.Text
		}
		return docs
	}

	// For longer documents, create paragraph like chunks
	var docs []string
	var currentDoc strings.Builder

	sentencesPerDoc := 3

	for i, sent := range sentences {
		currentDoc.WriteString((sent.Text))
		currentDoc.WriteString(" ")
		if (i+1)%sentencesPerDoc == 0 || i == len(sentences)-1 {
			docs = append(docs, currentDoc.String())
			currentDoc.Reset()
		}
	}
	return docs
}

func extractNamedEntities(doc *prose.Document) map[string]float64 {
	entities := make(map[string]float64)

	// Use prose's NER functionality

	for _, ent := range doc.Entities() {
		// Weight entities by type
		weight := 1.0
		switch ent.Label {
		case "PERSON", "ORG", "GPE":
			weight = 1.2
		case "DATE", "TIME", "MONEY", "PERCENT":
			weight = 1.0
		default:
			weight = 0.8

		}

		entities[strings.ToLower(ent.Text)] = weight
	}

	return entities
}

func calculatePositionWeights(sentences []prose.Sentence, sentenceCount int) []float64 {
	weights := make([]float64, sentenceCount)

	// Position based weighting - ifrst and last paragraph tend to be more important

	for i := range sentences {
		// First few sentences get higher weight
		if i < sentenceCount/5 {
			weights[i] = 1.0 - (0.8 * float64(i) / float64(sentenceCount/5))
		} else if i >= sentenceCount*4/5 {
			weights[i] = 0.4 + (0.4 * float64(i-sentenceCount*4/5) / float64(sentenceCount/5))
		} else {
			weights[i] = 0.2
		}
	}

	return weights
}

func mapEntitiesToSentences(sentences []prose.Sentence, entities map[string]float64) map[int]float64 {
	sentenceEntityScores := make(map[int]float64)

	for i, sentence := range sentences {
		score := 0.0
		text := strings.ToLower(sentence.Text)

		for entity, weight := range entities {
			if strings.Contains(text, strings.ToLower(entity)) {
				score += weight
			}
		}
		sentenceEntityScores[i] = score
	}

	return sentenceEntityScores
}

func extractTitleKeyWords(sentences []prose.Sentence) map[string]bool {
	keyWords := make(map[string]bool)

	// No sentences, return empty keywords
	if len(sentences) == 0 {
		return keyWords
	}

	// Assume first sentence might be title
	potentialTitle := sentences[0].Text
	words := strings.Fields(potentialTitle)

	for _, word := range words {
		word = strings.ToLower(word)
		// Keep only significant word
		if len(word) > 3 || !isStopWord(word) || !isAlphaNumeric(word) {
			keyWords[word] = true
		}
	}
	return keyWords
}
