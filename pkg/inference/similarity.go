package inference

import (
	"math"
	"strings"
	"unicode"
)

// SimilarityScorer computes TF-IDF based text similarity
type SimilarityScorer struct {
	docFreq   map[string]int
	docCount  int
	idfCache  map[string]float64
}

// NewSimilarityScorer creates a scorer from a corpus of documents
func NewSimilarityScorer(documents []string) *SimilarityScorer {
	s := &SimilarityScorer{
		docFreq:  make(map[string]int),
		docCount: len(documents),
		idfCache: make(map[string]float64),
	}

	for _, doc := range documents {
		seen := make(map[string]bool)
		for _, token := range tokenize(doc) {
			if !seen[token] {
				s.docFreq[token]++
				seen[token] = true
			}
		}
	}

	return s
}

// Score returns TF-IDF cosine similarity between query and document
func (s *SimilarityScorer) Score(query, document string) float64 {
	if query == "" || document == "" {
		return 0.0
	}

	qTokens := tokenize(query)
	dTokens := tokenize(document)

	if len(qTokens) == 0 || len(dTokens) == 0 {
		return 0.0
	}

	qTFIDF := s.tfidfVector(qTokens)
	dTFIDF := s.tfidfVector(dTokens)

	return cosineSimilarity(qTFIDF, dTFIDF)
}

func (s *SimilarityScorer) tfidfVector(tokens []string) map[string]float64 {
	tf := make(map[string]int)
	for _, t := range tokens {
		tf[t]++
	}

	vec := make(map[string]float64)
	for term, count := range tf {
		tfVal := float64(count) / float64(len(tokens))
		idf := s.idf(term)
		vec[term] = tfVal * idf
	}
	return vec
}

func (s *SimilarityScorer) idf(term string) float64 {
	if cached, ok := s.idfCache[term]; ok {
		return cached
	}

	df := s.docFreq[term]
	if df == 0 {
		df = 1
	}
	idf := math.Log(float64(s.docCount+1) / float64(df+1))
	s.idfCache[term] = idf
	return idf
}

func cosineSimilarity(a, b map[string]float64) float64 {
	var dotProduct, normA, normB float64

	for term, valA := range a {
		normA += valA * valA
		if valB, ok := b[term]; ok {
			dotProduct += valA * valB
		}
	}

	for _, valB := range b {
		normB += valB * valB
	}

	if normA == 0 || normB == 0 {
		return 0.0
	}

	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}

func tokenize(text string) []string {
	text = strings.ToLower(text)
	var tokens []string
	var current strings.Builder

	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(r)
		} else if current.Len() > 0 {
			token := current.String()
			if len(token) > 2 && !isStopWord(token) {
				tokens = append(tokens, token)
			}
			current.Reset()
		}
	}

	if current.Len() > 0 {
		token := current.String()
		if len(token) > 2 && !isStopWord(token) {
			tokens = append(tokens, token)
		}
	}

	return tokens
}

var stopWords = map[string]bool{
	"the": true, "and": true, "for": true, "are": true, "but": true,
	"not": true, "you": true, "all": true, "can": true, "had": true,
	"her": true, "was": true, "one": true, "our": true, "out": true,
	"has": true, "have": true, "been": true, "will": true, "more": true,
	"when": true, "who": true, "this": true, "that": true, "with": true,
	"from": true, "they": true, "what": true, "there": true, "which": true,
}

func isStopWord(word string) bool {
	return stopWords[word]
}

// ComputeSimilarity is a convenience function for one-off comparisons
func ComputeSimilarity(text1, text2 string) float64 {
	scorer := NewSimilarityScorer([]string{text1, text2})
	return scorer.Score(text1, text2)
}
