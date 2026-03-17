package crystallize

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// BeadSuggestion is the recommended bead to create from a beat
type BeadSuggestion struct {
	Title       string `json:"title"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Priority    int    `json:"priority"`
}

// CrystallizationSuggestion represents a beat ready for bead conversion
type CrystallizationSuggestion struct {
	BeatID         string         `json:"beat_id"`
	ContentPreview string         `json:"content_preview"`
	Ripeness       float64        `json:"ripeness"`
	SuggestedBead  BeadSuggestion `json:"suggested_bead"`
	RelatedBeats   []string       `json:"related_beats"`
	Confidence     float64        `json:"confidence"`
	Reason         string         `json:"reason"`
}

// SuggestOptions configures suggestion generation
type SuggestOptions struct {
	MinRipeness    float64
	MaxSuggestions int
	ProjectFilter  string
	IncludeRelated bool
}

// DefaultSuggestOptions returns sensible defaults
func DefaultSuggestOptions() SuggestOptions {
	return SuggestOptions{
		MinRipeness:    0.5,
		MaxSuggestions: 10,
		IncludeRelated: true,
	}
}

// GenerateSuggestions finds ripe beats and suggests bead conversions
func GenerateSuggestions(beats []model.EnrichedBeat, cache *model.Cache, opts SuggestOptions) []CrystallizationSuggestion {
	var suggestions []CrystallizationSuggestion

	// Filter to ripe beats
	var ripeBeats []model.EnrichedBeat
	for _, b := range beats {
		if b.RipenessScore >= opts.MinRipeness {
			// Skip beats already linked to beads
			if len(b.LinkedBeads) > 0 {
				continue
			}
			ripeBeats = append(ripeBeats, b)
		}
	}

	// Sort by ripeness (highest first)
	sort.Slice(ripeBeats, func(i, j int) bool {
		return ripeBeats[i].RipenessScore > ripeBeats[j].RipenessScore
	})

	// Generate suggestions for top N
	for i, beat := range ripeBeats {
		if i >= opts.MaxSuggestions {
			break
		}

		suggestion := CrystallizationSuggestion{
			BeatID:         beat.ID,
			ContentPreview: beat.ContentPreview(100),
			Ripeness:       beat.RipenessScore,
			SuggestedBead:  InferBeadFromBeat(beat),
			Confidence:     calculateConfidence(beat, cache),
			Reason:         generateReason(beat),
		}

		if opts.IncludeRelated {
			suggestion.RelatedBeats = findRelatedBeats(beat, beats, cache)
		}

		suggestions = append(suggestions, suggestion)
	}

	return suggestions
}

// InferBeadFromBeat generates a bead suggestion from beat content
func InferBeadFromBeat(beat model.EnrichedBeat) BeadSuggestion {
	title := extractActionableTitle(beat.Content)
	beadType := inferBeadType(beat)
	description := fmt.Sprintf("From beat %s: %s", beat.ID, beat.ContentPreview(200))

	return BeadSuggestion{
		Title:       title,
		Type:        beadType,
		Description: description,
		Priority:    inferPriority(beat.RipenessScore),
	}
}

var actionPatterns = []struct {
	pattern     *regexp.Regexp
	transformer func([]string) string
}{
	{
		regexp.MustCompile(`(?i)need to\s+(.+)`),
		func(m []string) string { return strings.Title(strings.TrimSpace(m[1])) },
	},
	{
		regexp.MustCompile(`(?i)should\s+(.+)`),
		func(m []string) string { return strings.Title(strings.TrimSpace(m[1])) },
	},
	{
		regexp.MustCompile(`(?i)want to\s+(.+)`),
		func(m []string) string { return strings.Title(strings.TrimSpace(m[1])) },
	},
	{
		regexp.MustCompile(`(?i)todo:?\s*(.+)`),
		func(m []string) string { return strings.Title(strings.TrimSpace(m[1])) },
	},
	{
		regexp.MustCompile(`(?i)^(implement|add|fix|create|build|design|refactor|update|remove)\s+(.+)`),
		func(m []string) string { return strings.Title(strings.TrimSpace(m[0])) },
	},
}

func extractActionableTitle(content string) string {
	// Clean content
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n", " ")

	for _, p := range actionPatterns {
		if matches := p.pattern.FindStringSubmatch(content); len(matches) > 0 {
			title := p.transformer(matches)
			// Truncate first sentence if too long
			if idx := strings.Index(title, "."); idx > 0 && idx < 60 {
				title = title[:idx]
			}
			if len(title) > 60 {
				return title[:60] + "..."
			}
			return title
		}
	}

	// Fallback: clean first sentence
	firstSentence := strings.Split(content, ".")[0]
	firstSentence = strings.TrimSpace(firstSentence)
	if len(firstSentence) > 60 {
		return firstSentence[:60] + "..."
	}
	if firstSentence == "" {
		return "Review beat content"
	}
	return firstSentence
}

func inferBeadType(beat model.EnrichedBeat) string {
	content := strings.ToLower(beat.Content)
	impetus := strings.ToLower(beat.Impetus.Label)

	switch {
	case strings.Contains(content, "bug") || strings.Contains(content, "fix") ||
		strings.Contains(content, "broken") || strings.Contains(content, "error"):
		return "bug"
	case strings.Contains(content, "feature") || strings.Contains(content, "add") ||
		strings.Contains(content, "implement") || strings.Contains(content, "new"):
		return "feature"
	case strings.Contains(impetus, "research") || strings.Contains(impetus, "discovery"):
		return "task"
	default:
		return "task"
	}
}

func inferPriority(ripeness float64) int {
	switch {
	case ripeness >= 0.8:
		return 1 // P1: Urgent
	case ripeness >= 0.6:
		return 2 // P2: High
	default:
		return 3 // P3: Medium
	}
}

func calculateConfidence(beat model.EnrichedBeat, cache *model.Cache) float64 {
	confidence := 0.5 // Base confidence

	// Higher ripeness = higher confidence
	confidence += beat.RipenessScore * 0.3

	// Has entities = better understanding
	if len(beat.ExtractedEntities) > 0 {
		confidence += 0.1
	}

	// In a cluster = related context available
	if beat.ClusterID != "" {
		confidence += 0.1
	}

	if confidence > 1.0 {
		confidence = 1.0
	}
	return confidence
}

func generateReason(beat model.EnrichedBeat) string {
	var reasons []string

	tier := model.RipenessTier(beat.RipenessScore)
	reasons = append(reasons, fmt.Sprintf("%s (%.2f)", tier, beat.RipenessScore))

	content := strings.ToLower(beat.Content)
	if strings.Contains(content, "need") || strings.Contains(content, "should") ||
		strings.Contains(content, "todo") || strings.Contains(content, "want to") {
		reasons = append(reasons, "actionable content")
	}

	if len(beat.ExtractedEntities) > 0 {
		reasons = append(reasons, fmt.Sprintf("%d entities", len(beat.ExtractedEntities)))
	}

	return strings.Join(reasons, ", ")
}

func findRelatedBeats(target model.EnrichedBeat, all []model.EnrichedBeat, cache *model.Cache) []string {
	var related []string
	seen := make(map[string]bool)

	// Same cluster
	if target.ClusterID != "" {
		for _, b := range all {
			if b.ID != target.ID && b.ClusterID == target.ClusterID {
				if !seen[b.ID] {
					related = append(related, b.ID)
					seen[b.ID] = true
				}
			}
		}
	}

	// Entity overlap
	targetEntities := make(map[string]bool)
	for _, e := range target.ExtractedEntities {
		targetEntities[e.Name] = true
	}

	for _, b := range all {
		if b.ID != target.ID && !seen[b.ID] {
			for _, e := range b.ExtractedEntities {
				if targetEntities[e.Name] {
					related = append(related, b.ID)
					seen[b.ID] = true
					break
				}
			}
		}
	}

	// Limit to 5 related beats
	if len(related) > 5 {
		related = related[:5]
	}

	return related
}

// SuggestionsResponse is the JSON output format
type SuggestionsResponse struct {
	Suggestions   []CrystallizationSuggestion `json:"suggestions"`
	TotalRipe     int                         `json:"total_ripe"`
	ProjectFilter *string                     `json:"project_filter"`
}
