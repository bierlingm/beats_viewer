package ripeness

import (
	"math"
	"strings"
	"time"

	"beats_viewer/pkg/model"
)

const (
	AgeFactor        = 0.2
	RevisitFactor    = 0.25
	ConnectionFactor = 0.25
	ActionFactor     = 0.2
	CompletenessFactor = 0.1
)

// Calculate computes the ripeness score (0.0-1.0) for a beat
func Calculate(beat model.Beat, allBeats []model.Beat, viewStat model.ViewStat) float64 {
	age := calculateAgeFactor(beat.CreatedAt)
	revisit := calculateRevisitFactor(viewStat.ViewCount)
	connection := calculateConnectionFactor(beat, allBeats)
	action := DetectActionLanguage(beat.Content) * ActionFactor
	completeness := calculateCompletenessFactor(beat)

	score := age + revisit + connection + action + completeness

	if score > 1.0 {
		score = 1.0
	}
	return score
}

func calculateAgeFactor(createdAt time.Time) float64 {
	ageDays := time.Since(createdAt).Hours() / 24
	factor := math.Min(ageDays/30, 1.0) * AgeFactor
	return factor
}

func calculateRevisitFactor(viewCount int) float64 {
	factor := math.Min(float64(viewCount)/5, 1.0) * RevisitFactor
	return factor
}

func calculateConnectionFactor(beat model.Beat, allBeats []model.Beat) float64 {
	connections := len(beat.LinkedBeads)
	connections += countRelatedBeats(beat, allBeats)
	factor := math.Min(float64(connections)/3, 1.0) * ConnectionFactor
	return factor
}

func countRelatedBeats(beat model.Beat, allBeats []model.Beat) int {
	count := 0
	for _, entity := range beat.Entities {
		for _, other := range allBeats {
			if other.ID == beat.ID {
				continue
			}
			for _, otherEntity := range other.Entities {
				if strings.EqualFold(entity, otherEntity) {
					count++
					break
				}
			}
		}
	}
	if count > 5 {
		count = 5
	}
	return count
}

func calculateCompletenessFactor(beat model.Beat) float64 {
	factors := CompletenessFactors{
		HasEntities:    len(beat.Entities) > 0,
		HasGoodImpetus: beat.Impetus.Label != "" && strings.ToLower(beat.Impetus.Label) != "manual entry",
		HasReferences:  len(beat.References) > 0,
		HasLinkedBeads: len(beat.LinkedBeads) > 0,
		ContentLength:  len(beat.Content),
	}
	return CalculateCompleteness(factors) * CompletenessFactor
}

// CalculateAll computes ripeness scores for all beats
func CalculateAll(beats []model.Beat, viewStats map[string]model.ViewStat) map[string]float64 {
	result := make(map[string]float64)
	for _, beat := range beats {
		stat := viewStats[beat.ID]
		result[beat.ID] = Calculate(beat, beats, stat)
	}
	return result
}

// RipenessBreakdown provides detailed scoring factors for a beat
type RipenessBreakdown struct {
	Total       float64 `json:"total"`
	Age         float64 `json:"age"`
	Revisit     float64 `json:"revisit"`
	Connection  float64 `json:"connection"`
	Action      float64 `json:"action"`
	Completeness float64 `json:"completeness"`
}

// CalculateWithBreakdown returns both the score and its component factors
func CalculateWithBreakdown(beat model.Beat, allBeats []model.Beat, viewStat model.ViewStat) RipenessBreakdown {
	age := calculateAgeFactor(beat.CreatedAt)
	revisit := calculateRevisitFactor(viewStat.ViewCount)
	connection := calculateConnectionFactor(beat, allBeats)
	action := DetectActionLanguage(beat.Content) * ActionFactor
	completeness := calculateCompletenessFactor(beat)

	total := age + revisit + connection + action + completeness
	if total > 1.0 {
		total = 1.0
	}

	return RipenessBreakdown{
		Total:       total,
		Age:         age,
		Revisit:     revisit,
		Connection:  connection,
		Action:      action,
		Completeness: completeness,
	}
}
