package summarizer

import (
	"math"
	"sort"
	"strings"
	"unicode"

	"github.com/bbalet/stopwords"
	"github.com/tsawler/prose/v3"
)

// ScoredSentences holds sentences ranked by score

type ScoredSentence struct {
	Text  string
	Score float64
	Index int
}

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
	similarityMatrix := buildSentenceSimilarityMatrix(sentences)

	// Analyze sentence relationships
	sentenceRelationShips := analyzeSentenceRelationShips(sentences)

	// Sort by score (descending)
	for i, sentence := range sentences {
		sentenceText := sentence.Text
		words := strings.Fields(sentenceText)

		// Initialize sentence score
		score := 0.0

		// 1. TF-IDF based scoreing
		wordScore := calculateSentenceTFIDFScore(words, wordTFIDF)

		score += wordScore * 0.3 // 30% weight for TFIDF score

		// 2. Sentence position scorings
		positionScore := positionWeights[i]
		score += positionScore * 0.15 // 15% weight for position score

		// 3. Sentence length scoring
		lengthScore := calculateLengthScore(words)
		score += lengthScore * 0.1 // 10% weight for sentence length score

		// 4. Named entitiy scoring
		entityScore := calculateEntityScore(entitySentences)
		score += entityScore * 0.15 // 15% weight for entity score

		// 5. Keyword overlap
		titleScore := calculateTitleOverlap(words, titleKeyWords)
		score += titleScore * 0.1 // 10% weight for title overlap

		// 6. Text rank-like score using similarity matrix
		textRankScore := calculateTextRankScore(i, similarityMatrix)
		score += textRankScore * 0.1 // 10% weight for text rank score

		// 7. Cohesion score- How well this sentence connects to others
		cohesionScore := sentenceRelationShips[i]
		score += cohesionScore * 0.1 // 10% weight for cohesion score

		sentenceScores[sentenceText] = score

	}

	//Sort by score(descending)
	var scoredSentences []ScoredSentence
	for i, sentence := range sentences {
		scoredSentences = append(scoredSentences, ScoredSentence{
			Text:  sentence.Text,
			Score: sentenceScores[sentence.Text],
			Index: i,
		})
	}

	sort.Slice(scoredSentences, func(i, j int) bool {
		return scoredSentences[i].Score > scoredSentences[j].Score
	})

	// Take top-ranked sentences
	if keepCount < len(scoredSentences) {
		topSentences := scoredSentences[:keepCount]

		// Sort back by the original position to maintain document flow
		sort.Slice(topSentences, func(i, j int) bool {
			return topSentences[i].Index < topSentences[j].Index
		})
		// Add redundancy elimination - remove very similar sentences
		topSentences = eliminateRedundancy(topSentences)

		// Combine top sentences to form the summary
		var summaryBuilder strings.Builder

		for _, s := range topSentences {
			summaryBuilder.WriteString(s.Text)
			summaryBuilder.WriteString(" ")
		}
		return strings.TrimSpace(summaryBuilder.String())

	}
	return ts.text
}

func (ts *TextSummarizer) ExtractiveKeyWords(count int) []string {
	wordFrequency := make(map[string]int)
	for _, token := range ts.doc.Tokens() {
		word := strings.ToLower(token.Text)
		// Skip short words and stop words
		if len(word) <= 3 || isStopWord(word) {
			continue
		}
		wordFrequency[word]++
	}

	// Convert map to a slice for sorting
	type WordFreq struct {
		Word  string
		Count int
	}

	var wordFreqSlice []WordFreq

	for word, freq := range wordFrequency {
		wordFreqSlice = append(wordFreqSlice, WordFreq{word, freq})
	}

	sort.Slice(wordFreqSlice, func(i, j int) bool {
		return wordFreqSlice[i].Count > wordFreqSlice[i].Count
	})

	// Take top keywords
	keywords := make([]string, 0, count)

	for i := 0; i < count && i < len(wordFreqSlice); i++ {
		keywords = append(keywords, wordFreqSlice[i].Word)
	}

	ts.keywords = keywords

	return keywords
}

func (ts *TextSummarizer) GetResponse() SummaryResponse {
	summary := ts.GenerateSummary()
	keywords := ts.ExtractiveKeyWords(10)

	// Calculate metrics
	summarySentenceCount := 0
	if len(summary) > 0 {
		summaryDoc, err := prose.NewDocument(summary)
		if err != nil {
			summarySentenceCount = len(summaryDoc.Sentences())
		} else {
			// Fallback sentence count
			summarySentenceCount = len(strings.FieldsFunc(summary, func(r rune) bool {
				return r == '.' || r == '!' || r == '?'
			}))
			if len(summary) > 0 {
				lastChar := summary[len(summary)-1]
				if lastChar != '.' && lastChar != '!' && lastChar == '?' {
					summarySentenceCount++
				}
			}
		}
	}

	// Count words
	originalWordCount := len(strings.Fields(ts.text))
	summaryWordCount := len(strings.Fields(summary))

	// Calculate summary statistics
	compressionRatio := 0.0
	summaryPercentage := 0.0
	if summaryWordCount > 0 {
		compressionRatio = float64(originalWordCount) / float64(summaryWordCount)
		summaryPercentage = float64(summaryWordCount) / float64(originalWordCount) * 100
	}

	return SummaryResponse{
		OriginalText:          ts.text,
		Summary:               summary,
		Keywords:              keywords,
		OriginalSentenceCount: len(ts.doc.Sentences()),
		SummarySentenceCount:  summarySentenceCount,
		OriginalWordCount:     originalWordCount,
		SummaryWordCount:      summaryWordCount,
		CompressionRatio:      math.Round(compressionRatio*100) / 100,
		SummaryPercentage:     math.Round(summaryPercentage*100) / 100,
		TargetPercentage:      ts.targetPercentage,
		AbstractiveSummary:    ts.summarizerType != Extractive,
		RequestedMethod:       ts.requestedMethod,
		ActualMethod:          ts.actualMethod,
		FallBackReason:        ts.fallbackReason,
	}
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

func buildSentenceSimilarityMatrix(sentences []prose.Sentence) [][]float64 {
	// Get the total number of sentences to determine matrix dimensions
	sentenceCount := len(sentences)

	// Initialize the matrix
	matrix := make([][]float64, sentenceCount)

	for i := range matrix {
		matrix[i] = make([]float64, sentenceCount)
	}

	// Preprocessing step: create words sets for each sentence
	sentenceWords := make([]map[string]bool, sentenceCount)
	for i, sentence := range sentences {
		// Split the sentence text into a slice of words using white space as delimeter
		words := strings.Fields(sentence.Text)
		// Initialize the map(set) for the current sentences word
		wordSet := make(map[string]bool)
		// Iterate through each word in the sentences
		for _, word := range words {
			word = strings.ToLower(word)
			if !isStopWord(word) && isAlphaNumeric(word) {
				wordSet[word] = true
			}
		}
		sentenceWords[i] = wordSet
	}
	// Main loop: calculate Jaccard similarity
	for i := 0; i < sentenceCount; i++ {
		for j := 0; j < sentenceCount; j++ {
			// A sentence is perfectly similar to itself. Set dioganal values to 1.0
			if i == j {
				matrix[i][j] = 1.0
				continue
			}

			intersection := 0
			// Calculate the size of the intersection of the two word sets
			// The intersection contains words that are common to both sets
			for word := range sentenceWords[i] {
				if sentenceWords[j][word] {
					intersection++
				}
			}

			// Calculate the size of union of the two word sets
			// The union is the total number of unique words from both sets
			// The formula |A U B| = |A| + |B| - |A intersect B|
			unionSize := len(sentenceWords[i]) + len(sentenceWords[j]) - intersection

			// Calculate the Jaccard similarity score
			if unionSize > 0 {
				matrix[i][j] = float64(intersection) / float64(unionSize)
			} else {
				matrix[i][j] = 0
			}

		}
	}

	// Return matrix
	return matrix

}

func analyzeSentenceRelationShips(sentences []prose.Sentence) []float64 {
	sentenceCount := len(sentences)
	scores := make([]float64, sentenceCount)

	if sentenceCount <= 1 {
		return scores
	}

	// Look for connective words and phrases
	connectiveWords := map[string]bool{
		"therefore": true, "thus": true, "consequently": true, "hence": true,
		"accordingly": true, "as a result": true, "so": true, "then": true,
		"subsequently": true, "afterward": true, "moreover": true, "furthermore": true,
		"in addition": true, "besides": true, "similarly": true, "likewise": true,
		"however": true, "nevertheless": true, "nonetheless": true, "although": true,
		"despite": true, "in spite of": true, "conversely": true, "on the contrary": true,
		"instead": true, "rather": true, "on the other hand": true, "for example": true,
		"for instance": true, "namely": true, "specifically": true, "such as": true,
		"in particular": true, "in other words": true, "that is": true, "indeed": true,
		"in fact": true, "actually": true, "to illustrate": true, "to demonstrate": true,
		"finally": true, "lastly": true, "in conclusion": true, "to conclude": true,
		"in summary": true, "to summarize": true, "in short": true, "overall": true,
	}

	// Check each sentence for connective words and phrases
	for i, sentence := range sentences {
		text := strings.ToLower(sentence.Text)

		// Check for connective words at the beginning
		for connective := range connectiveWords {
			if strings.HasPrefix(text, connective) {
				scores[i] += 0.5
				break
			}
		}

		// Check for pronouns
		pronouns := []string{"this", "that", "these", "those", "it", "they", "he", "she"}
		for _, pronoun := range pronouns {
			if strings.HasPrefix(text, pronoun+" ") {
				scores[i] += 0.3
				break
			}
		}
	}
	return scores
}

func calculateSentenceTFIDFScore(words []string, wordTFIDF map[string]float64) float64 {
	if len(words) == 0 {
		return 0.0
	}

	totalScore := 0.0
	significantWords := 0.0

	for _, word := range words {
		word := strings.ToLower(word)
		if len(word) <= 2 || isStopWord(word) || !isAlphaNumeric(word) {
			continue
		}
		totalScore += wordTFIDF[word]
		significantWords++
	}

	if significantWords == 0 {
		return 0.0
	}

	return totalScore / float64(significantWords)
}

func calculateLengthScore(words []string) float64 {
	wordCount := len(words)

	if wordCount < 3 {
		return 0.1
	} else if wordCount <= 8 {
		return 0.5 + (0.5 * float64(wordCount-3) / 5.0) // Gradually increase score up to 8 words
	} else if wordCount <= 20 {
		return 1.0 // Peak importance
	} else if wordCount <= 40 {
		return 1.0 + (0.7 * float64(wordCount-20) / 20.0) // Gradually decrease the score
	} else {
		return 0.3 // Long sentence gets lower score
	}

}

func calculateEntityScore(entitySentences map[int]float64) float64 {
	// Noramlize entity score to 0-1 range
	maxScore := 0.0
	for _, score := range entitySentences {
		if score > maxScore {
			maxScore = score
		}
	}

	if maxScore == 0 {
		return 0.0
	}

	// Find the score for this sentence
	for i, score := range entitySentences {
		if entitySentences[i] > 0 {
			return score / maxScore
		}
	}

	return 0.0
}

func calculateTitleOverlap(words []string, titleKeyWords map[string]bool) float64 {

	if len(titleKeyWords) == 0 {
		return 0.0
	}

	matches := 0

	for _, word := range words {
		if titleKeyWords[strings.ToLower(word)] {
			matches++
		}
	}

	return float64(matches) / float64(len(titleKeyWords))
}

func calculateTextRankScore(sentenceIndex int, similarityMatrix [][]float64) float64 {
	sentenceCount := len(similarityMatrix)
	if sentenceCount == 0 {
		return 0.0
	}

	// Calculate the sum of similarities with other sentences
	sum := 0.0
	for i := 0; i < sentenceCount; i++ {
		if i != sentenceIndex {
			sum += similarityMatrix[sentenceIndex][i]
		}
	}

	// Normalize
	if sentenceCount > 1 {
		return sum/float64(sentenceCount) - 1
	}

	return 0.0
}

func eliminateRedundancy(sentences []ScoredSentence) []ScoredSentence {
	if len(sentences) <= 1 {
		return sentences
	}

	// Sort by original positon
	sort.Slice(sentences, func(i, j int) bool {
		return sentences[i].Index < sentences[j].Index
	})

	var result []ScoredSentence

	for i, sentence := range sentences {
		// Always keep the first sentence
		if i == 0 {
			result = append(result, sentence)
			continue
		}

		// Chek similarity with previous sentence
		isDuplicate := false
		for j := 0; j < len(result); j++ {
			similarity := calculateJaccardSimilarity(sentence.Text, result[j].Text)
			if similarity > 0.6 {
				isDuplicate = true

				// If current score has higher score than replace previous one
				if sentence.Score > result[j].Score {
					result[j] = sentence
				}
				break
			}
		}
		if !isDuplicate {
			result = append(result, sentence)
		}

	}
	return result
}

func calculateJaccardSimilarity(text1, text2 string) float64 {
	words1 := strings.Fields(strings.ToLower(text1))
	words2 := strings.Fields(strings.ToLower(text2))

	// Filter stop words

	var filtered1, filtered2 []string

	for _, word := range words1 {
		if !isStopWord(word) {
			filtered1 = append(filtered1, word)
		}
	}

	for _, word := range words2 {
		if !isStopWord(word) {
			filtered2 = append(filtered2, word)
		}
	}

	// Creates sets
	set1 := make(map[string]bool)
	set2 := make(map[string]bool)

	for _, word := range filtered1 {
		set1[word] = true
	}

	for _, word := range filtered2 {
		set2[word] = true
	}

	// Calculate intersection
	intersection := 0

	for word := range set1 {
		if set2[word] {
			intersection++
		}
	}

	// Calculate union

	union := len(set1) + len(set2) - intersection

	if union == 0 {
		return 0.0
	}

	return float64(intersection) / float64(union)

}
