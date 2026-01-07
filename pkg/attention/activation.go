package attention

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// ActivationType represents the type of activation detected
type ActivationType int

const (
	ActivationBurst        ActivationType = iota // Sudden increase in existing cluster
	ActivationReactivation                        // Dormant cluster waking up
	ActivationEmergent                            // New pattern forming
)

func (t ActivationType) String() string {
	switch t {
	case ActivationBurst:
		return "burst"
	case ActivationReactivation:
		return "reactivation"
	case ActivationEmergent:
		return "emergent"
	default:
		return "unknown"
	}
}

// Activation represents a detected activation event
type Activation struct {
	ClusterID     string
	ClusterName   string
	Type          ActivationType
	BeatCount     int           // Beats in detection window
	Window        time.Duration // Detection window
	Beats         []string      // Beat IDs that triggered this
	PriorActivity int           // Beats in prior equivalent window
}

// ActivationConfig holds detection thresholds
type ActivationConfig struct {
	Window         time.Duration // Default 72h
	BurstThreshold int           // Default 3 beats in window
	DormancyPeriod time.Duration // Default 14 days
	RateMultiplier float64       // 2x prior window rate triggers burst
}

// DefaultActivationConfig returns default activation detection settings
func DefaultActivationConfig() ActivationConfig {
	return ActivationConfig{
		Window:         72 * time.Hour,
		BurstThreshold: 3,
		DormancyPeriod: 14 * 24 * time.Hour,
		RateMultiplier: 2.0,
	}
}

// DetectActivations finds all activation events
func DetectActivations(beats []model.Beat, clusters []model.Cluster, config ActivationConfig) []Activation {
	var activations []Activation

	// Detect bursts
	bursts := DetectBursts(beats, clusters, config)
	activations = append(activations, bursts...)

	// Detect reactivations
	reactivations := DetectReactivations(beats, clusters, config)
	activations = append(activations, reactivations...)

	// Sort by beat count descending (most active first)
	sort.Slice(activations, func(i, j int) bool {
		return activations[i].BeatCount > activations[j].BeatCount
	})

	return activations
}

// DetectBursts finds clusters with sudden activity
func DetectBursts(beats []model.Beat, clusters []model.Cluster, config ActivationConfig) []Activation {
	var activations []Activation
	now := time.Now()

	currentStart := now.Add(-config.Window)
	priorStart := currentStart.Add(-config.Window)

	for _, cluster := range clusters {
		clusterBeats := BeatsInCluster(beats, cluster)

		// Count beats in current window
		var currentBeats []model.Beat
		for _, b := range clusterBeats {
			if !b.CreatedAt.Before(currentStart) {
				currentBeats = append(currentBeats, b)
			}
		}

		// Count beats in prior window
		priorCount := 0
		for _, b := range clusterBeats {
			if !b.CreatedAt.Before(priorStart) && b.CreatedAt.Before(currentStart) {
				priorCount++
			}
		}

		currentCount := len(currentBeats)

		// Check burst conditions:
		// 1. At least BurstThreshold beats in current window
		// 2. OR current rate is RateMultiplier times prior rate
		isBurst := currentCount >= config.BurstThreshold ||
			(priorCount > 0 && float64(currentCount) >= float64(priorCount)*config.RateMultiplier)

		if isBurst && currentCount > 0 {
			beatIDs := make([]string, len(currentBeats))
			for i, b := range currentBeats {
				beatIDs[i] = b.ID
			}

			activations = append(activations, Activation{
				ClusterID:     cluster.ID,
				ClusterName:   cluster.Name,
				Type:          ActivationBurst,
				BeatCount:     currentCount,
				Window:        config.Window,
				Beats:         beatIDs,
				PriorActivity: priorCount,
			})
		}
	}

	return activations
}

// DetectReactivations finds dormant clusters with new activity
func DetectReactivations(beats []model.Beat, clusters []model.Cluster, config ActivationConfig) []Activation {
	var activations []Activation
	now := time.Now()

	windowStart := now.Add(-config.Window)
	dormancyCutoff := windowStart.Add(-config.DormancyPeriod)

	for _, cluster := range clusters {
		clusterBeats := BeatsInCluster(beats, cluster)
		if len(clusterBeats) == 0 {
			continue
		}

		// Sort by time descending
		sortedBeats := make([]model.Beat, len(clusterBeats))
		copy(sortedBeats, clusterBeats)
		sort.Slice(sortedBeats, func(i, j int) bool {
			return sortedBeats[i].CreatedAt.After(sortedBeats[j].CreatedAt)
		})

		// Find beats in current window
		var recentBeats []model.Beat
		for _, b := range sortedBeats {
			if !b.CreatedAt.Before(windowStart) {
				recentBeats = append(recentBeats, b)
			}
		}

		if len(recentBeats) == 0 {
			continue
		}

		// Check if cluster was dormant (no activity for DormancyPeriod before window)
		hadRecentPriorActivity := false
		for _, b := range sortedBeats {
			if b.CreatedAt.Before(windowStart) && !b.CreatedAt.Before(dormancyCutoff) {
				hadRecentPriorActivity = true
				break
			}
		}

		if !hadRecentPriorActivity {
			// Count prior activity (before dormancy period)
			priorCount := 0
			for _, b := range sortedBeats {
				if b.CreatedAt.Before(dormancyCutoff) {
					priorCount++
				}
			}

			if priorCount > 0 { // Only flag as reactivation if there was prior activity
				beatIDs := make([]string, len(recentBeats))
				for i, b := range recentBeats {
					beatIDs[i] = b.ID
				}

				activations = append(activations, Activation{
					ClusterID:     cluster.ID,
					ClusterName:   cluster.Name,
					Type:          ActivationReactivation,
					BeatCount:     len(recentBeats),
					Window:        config.Window,
					Beats:         beatIDs,
					PriorActivity: priorCount,
				})
			}
		}
	}

	return activations
}

// IsActivating returns true if any activation is detected for the cluster
func IsActivating(clusterID string, activations []Activation) bool {
	for _, a := range activations {
		if a.ClusterID == clusterID {
			return true
		}
	}
	return false
}
