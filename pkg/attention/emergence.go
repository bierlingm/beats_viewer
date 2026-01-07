package attention

import (
	"sort"
	"strings"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// EmergenceConfig holds detection parameters
type EmergenceConfig struct {
	Window   time.Duration // Look at recent beats (default 30 days)
	MinBeats int           // Minimum beats to consider emergent (default 3)
}

// DefaultEmergenceConfig returns default emergence detection settings
func DefaultEmergenceConfig() EmergenceConfig {
	return EmergenceConfig{
		Window:   30 * 24 * time.Hour,
		MinBeats: 3,
	}
}

// EmergentPattern represents a potential new pattern forming
type EmergentPattern struct {
	BeatIDs     []string
	BeatCount   int
	FirstSeen   time.Time
	LastSeen    time.Time
	CommonTerms []string // Shared keywords/entities
	Signal      string   // Why this is flagged
}

// DetectEmergence finds beats not matching existing clusters
func DetectEmergence(beats []model.Beat, clusters []model.Cluster, config EmergenceConfig) []EmergentPattern {
	// Get unclustered beats in the window
	recentBeats := BeatsInWindow(beats, config.Window)
	unclustered := UnclusteredBeats(recentBeats, clusters)

	// Not enough unclustered beats
	if len(unclustered) < config.MinBeats {
		return nil
	}

	// Sort by time
	sort.Slice(unclustered, func(i, j int) bool {
		return unclustered[i].CreatedAt.Before(unclustered[j].CreatedAt)
	})

	// Extract beat IDs
	beatIDs := make([]string, len(unclustered))
	for i, b := range unclustered {
		beatIDs[i] = b.ID
	}

	// Find common terms
	commonTerms := FindCommonTerms(unclustered)

	// Build pattern
	pattern := EmergentPattern{
		BeatIDs:     beatIDs,
		BeatCount:   len(unclustered),
		FirstSeen:   unclustered[0].CreatedAt,
		LastSeen:    unclustered[len(unclustered)-1].CreatedAt,
		CommonTerms: commonTerms,
		Signal:      generateEmergenceSignal(len(unclustered), len(commonTerms)),
	}

	return []EmergentPattern{pattern}
}

// UnclusteredBeats returns beats not assigned to any cluster
func UnclusteredBeats(beats []model.Beat, clusters []model.Cluster) []model.Beat {
	// Build set of all clustered beat IDs
	clustered := make(map[string]bool)
	for _, c := range clusters {
		for _, bid := range c.BeatIDs {
			clustered[bid] = true
		}
	}

	var unclustered []model.Beat
	for _, b := range beats {
		if !clustered[b.ID] {
			unclustered = append(unclustered, b)
		}
	}
	return unclustered
}

// FindCommonTerms extracts shared terms from a set of beats
func FindCommonTerms(beats []model.Beat) []string {
	if len(beats) == 0 {
		return nil
	}

	// Count entity occurrences
	entityCounts := make(map[string]int)
	for _, b := range beats {
		for _, e := range b.Entities {
			entityCounts[strings.ToLower(e)]++
		}
	}

	// Find entities that appear in at least 2 beats
	threshold := 2
	if len(beats) > 5 {
		threshold = len(beats) / 3
	}

	var common []string
	for entity, count := range entityCounts {
		if count >= threshold {
			common = append(common, entity)
		}
	}

	// Sort by frequency descending
	sort.Slice(common, func(i, j int) bool {
		return entityCounts[common[i]] > entityCounts[common[j]]
	})

	// Limit to top 5
	if len(common) > 5 {
		common = common[:5]
	}

	return common
}

// generateEmergenceSignal creates a human-readable explanation
func generateEmergenceSignal(beatCount, commonTermCount int) string {
	if commonTermCount > 0 {
		return "unclustered beats with shared themes"
	}
	return "unclustered recent beats"
}

// HasEmergence returns true if emergence patterns were detected
func HasEmergence(patterns []EmergentPattern) bool {
	return len(patterns) > 0
}

// TotalEmergentBeats returns the total number of emergent beats
func TotalEmergentBeats(patterns []EmergentPattern) int {
	total := 0
	for _, p := range patterns {
		total += p.BeatCount
	}
	return total
}
