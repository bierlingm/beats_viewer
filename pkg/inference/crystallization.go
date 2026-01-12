package inference

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// SignalType represents the type of inference signal
type SignalType int

const (
	SignalContentSimilarity SignalType = iota
	SignalTemporalProximity
	SignalEntityOverlap
	SignalSessionCorrelation
)

func (s SignalType) String() string {
	switch s {
	case SignalContentSimilarity:
		return "content_similarity"
	case SignalTemporalProximity:
		return "temporal_proximity"
	case SignalEntityOverlap:
		return "entity_overlap"
	case SignalSessionCorrelation:
		return "session_correlation"
	default:
		return "unknown"
	}
}

// InferenceSignal represents one component of inference
type InferenceSignal struct {
	Type   SignalType `json:"type"`
	Score  float64    `json:"score"`
	Detail string     `json:"detail"`
}

// CrystallizationResult represents inferred beat→bead connection
type CrystallizationResult struct {
	BeatIDs    []string          `json:"beat_ids"`
	BeadID     string            `json:"bead_id"`
	BeadTitle  string            `json:"bead_title"`
	Confidence float64           `json:"confidence"`
	Signals    []InferenceSignal `json:"signals"`
	InferredAt time.Time         `json:"inferred_at"`
}

// CrystallizationConfig holds inference settings
type CrystallizationConfig struct {
	ConfidenceThreshold float64
	TemporalWindow      time.Duration
	MinSimilarity       float64
	MaxBeatsPerBead     int
}

// DefaultCrystallizationConfig returns sensible defaults
func DefaultCrystallizationConfig() CrystallizationConfig {
	return CrystallizationConfig{
		ConfidenceThreshold: 0.6,
		TemporalWindow:      90 * 24 * time.Hour,
		MinSimilarity:       0.1,
		MaxBeatsPerBead:     10,
	}
}

// Signal weights
const (
	weightContentSimilarity  = 0.4
	weightTemporalProximity  = 0.3
	weightEntityOverlap      = 0.2
	weightSessionCorrelation = 0.1
)

// InferCrystallizations finds likely beat→bead connections
func InferCrystallizations(beats []model.Beat, beads []Bead, config CrystallizationConfig) []CrystallizationResult {
	if len(beats) == 0 || len(beads) == 0 {
		return nil
	}

	// Build corpus for similarity scoring
	var corpus []string
	for _, b := range beats {
		corpus = append(corpus, b.Content)
	}
	for _, b := range beads {
		corpus = append(corpus, b.Title+" "+b.Description)
	}
	scorer := NewSimilarityScorer(corpus)

	var results []CrystallizationResult

	for _, bead := range beads {
		beadText := bead.Title + " " + bead.Description
		var candidates []beatScore

		for _, beat := range beats {
			// Only consider beats captured before or around bead creation
			if beat.CreatedAt.After(bead.CreatedAt.Add(24 * time.Hour)) {
				continue
			}

			// Check temporal window
			if bead.CreatedAt.Sub(beat.CreatedAt) > config.TemporalWindow {
				continue
			}

			signals := computeSignals(beat, bead, beadText, scorer)
			confidence := computeConfidence(signals)

			if confidence >= config.MinSimilarity {
				candidates = append(candidates, beatScore{
					beat:       beat,
					signals:    signals,
					confidence: confidence,
				})
			}
		}

		// Sort by confidence descending
		sort.Slice(candidates, func(i, j int) bool {
			return candidates[i].confidence > candidates[j].confidence
		})

		// Take top candidates above threshold
		var beatIDs []string
		var topSignals []InferenceSignal
		avgConfidence := 0.0

		for i, c := range candidates {
			if i >= config.MaxBeatsPerBead {
				break
			}
			if c.confidence >= config.ConfidenceThreshold {
				beatIDs = append(beatIDs, c.beat.ID)
				avgConfidence += c.confidence
				if i == 0 {
					topSignals = c.signals
				}
			}
		}

		if len(beatIDs) > 0 {
			avgConfidence /= float64(len(beatIDs))
			results = append(results, CrystallizationResult{
				BeatIDs:    beatIDs,
				BeadID:     bead.ID,
				BeadTitle:  bead.Title,
				Confidence: avgConfidence,
				Signals:    topSignals,
				InferredAt: time.Now(),
			})
		}
	}

	// Sort by confidence descending
	sort.Slice(results, func(i, j int) bool {
		return results[i].Confidence > results[j].Confidence
	})

	return results
}

type beatScore struct {
	beat       model.Beat
	signals    []InferenceSignal
	confidence float64
}

func computeSignals(beat model.Beat, bead Bead, beadText string, scorer *SimilarityScorer) []InferenceSignal {
	var signals []InferenceSignal

	// Content similarity
	simScore := scorer.Score(beat.Content, beadText)
	signals = append(signals, InferenceSignal{
		Type:   SignalContentSimilarity,
		Score:  simScore,
		Detail: "",
	})

	// Temporal proximity
	tempScore := temporalProximityScore(beat.CreatedAt, bead.CreatedAt)
	signals = append(signals, InferenceSignal{
		Type:   SignalTemporalProximity,
		Score:  tempScore,
		Detail: "",
	})

	// Entity overlap (simple keyword matching)
	overlapScore := entityOverlapScore(beat, bead)
	signals = append(signals, InferenceSignal{
		Type:   SignalEntityOverlap,
		Score:  overlapScore,
		Detail: "",
	})

	// Session correlation (check impetus for session info)
	sessionScore := sessionCorrelationScore(beat, bead)
	signals = append(signals, InferenceSignal{
		Type:   SignalSessionCorrelation,
		Score:  sessionScore,
		Detail: "",
	})

	return signals
}

func computeConfidence(signals []InferenceSignal) float64 {
	var confidence float64
	for _, s := range signals {
		switch s.Type {
		case SignalContentSimilarity:
			confidence += s.Score * weightContentSimilarity
		case SignalTemporalProximity:
			confidence += s.Score * weightTemporalProximity
		case SignalEntityOverlap:
			confidence += s.Score * weightEntityOverlap
		case SignalSessionCorrelation:
			confidence += s.Score * weightSessionCorrelation
		}
	}
	return confidence
}

func temporalProximityScore(beatTime, beadTime time.Time) float64 {
	diff := beadTime.Sub(beatTime)
	if diff < 0 {
		return 0.0
	}

	days := diff.Hours() / 24
	switch {
	case days <= 1:
		return 1.0
	case days <= 3:
		return 0.8
	case days <= 7:
		return 0.6
	case days <= 30:
		return 0.3
	default:
		return 0.0
	}
}

func entityOverlapScore(beat model.Beat, bead Bead) float64 {
	beatTokens := tokenize(beat.Content)
	beadTokens := tokenize(bead.Title + " " + bead.Description)

	if len(beatTokens) == 0 || len(beadTokens) == 0 {
		return 0.0
	}

	beadSet := make(map[string]bool)
	for _, t := range beadTokens {
		beadSet[t] = true
	}

	overlap := 0
	for _, t := range beatTokens {
		if beadSet[t] {
			overlap++
		}
	}

	// Jaccard-like overlap score
	return float64(overlap) / float64(len(beatTokens)+len(beadTokens)-overlap)
}

func sessionCorrelationScore(beat model.Beat, bead Bead) float64 {
	// Check if beat impetus contains session identifiers
	impetusText := beat.Impetus.Label + " " + beat.Impetus.Raw
	if impetusText == " " {
		return 0.0
	}

	// Check for tag overlap between beat and bead
	if len(bead.Tags) == 0 {
		return 0.0
	}

	for _, tag := range bead.Tags {
		if contains(beat.Content, tag) || contains(impetusText, tag) {
			return 0.5
		}
	}

	return 0.0
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && 
		(s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 1; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
