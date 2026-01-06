package ripeness

import (
	"strings"
)

// Action language patterns that indicate readiness for action
var actionPatterns = []string{
	"should", "need to", "must", "will", "plan to",
	"implement", "build", "create", "fix", "add",
	"todo", "action", "next step", "follow up",
	"want to", "going to", "have to", "try to",
}

// DetectActionLanguage returns a score 0.0-1.0 based on actionable phrasing
func DetectActionLanguage(content string) float64 {
	lower := strings.ToLower(content)
	matches := 0
	for _, pattern := range actionPatterns {
		if strings.Contains(lower, pattern) {
			matches++
		}
	}
	score := float64(matches) / 3.0
	if score > 1.0 {
		score = 1.0
	}
	return score
}

// Completeness factors for scoring
type CompletenessFactors struct {
	HasEntities    bool
	HasGoodImpetus bool
	HasReferences  bool
	HasLinkedBeads bool
	ContentLength  int
}

// CalculateCompleteness returns a score 0.0-1.0 for beat completeness
func CalculateCompleteness(f CompletenessFactors) float64 {
	var score float64

	if f.HasEntities {
		score += 0.2
	}
	if f.HasGoodImpetus {
		score += 0.3
	}
	if f.HasReferences {
		score += 0.2
	}
	if f.HasLinkedBeads {
		score += 0.2
	}
	if f.ContentLength > 100 {
		score += 0.1
	}

	if score > 1.0 {
		score = 1.0
	}
	return score
}
