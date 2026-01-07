package attention

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// DormancyConfig holds detection parameters
type DormancyConfig struct {
	InactivityThreshold time.Duration // Default 30 days
	MinRipeness         float64       // Default 0.6 (ripe or overripe)
}

// DefaultDormancyConfig returns default dormancy detection settings
func DefaultDormancyConfig() DormancyConfig {
	return DormancyConfig{
		InactivityThreshold: 30 * 24 * time.Hour,
		MinRipeness:         0.6,
	}
}

// DormantCluster represents a cluster that may need attention
type DormantCluster struct {
	ClusterID      string
	ClusterName    string
	RipenessScore  float64
	LastActivityAt time.Time
	InactiveDays   int
	RipeBeatCount  int      // Number of ripe beats
	RipeBeatIDs    []string // IDs of ripe beats
}

// DetectDormancy finds clusters with ripe beats but no recent activity
func DetectDormancy(beats []model.Beat, clusters []model.Cluster, ripeness map[string]float64, config DormancyConfig) []DormantCluster {
	var dormant []DormantCluster
	now := time.Now()
	inactivityCutoff := now.Add(-config.InactivityThreshold)

	for _, cluster := range clusters {
		// Get beats in this cluster
		clusterBeats := BeatsInCluster(beats, cluster)
		if len(clusterBeats) == 0 {
			continue
		}

		// Find last activity
		lastActivity := LastActivityInCluster(beats, cluster)

		// Skip if there's been recent activity
		if !lastActivity.Before(inactivityCutoff) {
			continue
		}

		// Find ripe beats in this cluster
		var ripeBeatIDs []string
		for _, b := range clusterBeats {
			if score, ok := ripeness[b.ID]; ok && score >= config.MinRipeness {
				ripeBeatIDs = append(ripeBeatIDs, b.ID)
			}
		}

		// Only flag if there are ripe beats
		if len(ripeBeatIDs) == 0 {
			continue
		}

		inactiveDays := int(now.Sub(lastActivity).Hours() / 24)

		dormant = append(dormant, DormantCluster{
			ClusterID:      cluster.ID,
			ClusterName:    cluster.Name,
			RipenessScore:  cluster.RipenessScore,
			LastActivityAt: lastActivity,
			InactiveDays:   inactiveDays,
			RipeBeatCount:  len(ripeBeatIDs),
			RipeBeatIDs:    ripeBeatIDs,
		})
	}

	// Sort by inactive days descending (most dormant first)
	sort.Slice(dormant, func(i, j int) bool {
		return dormant[i].InactiveDays > dormant[j].InactiveDays
	})

	return dormant
}

// LastActivityInCluster returns the most recent beat timestamp in a cluster
func LastActivityInCluster(beats []model.Beat, cluster model.Cluster) time.Time {
	beatSet := make(map[string]bool)
	for _, bid := range cluster.BeatIDs {
		beatSet[bid] = true
	}

	var latest time.Time
	for _, b := range beats {
		if beatSet[b.ID] && b.CreatedAt.After(latest) {
			latest = b.CreatedAt
		}
	}
	return latest
}

// IsDormant returns true if the cluster appears in the dormant list
func IsDormant(clusterID string, dormant []DormantCluster) bool {
	for _, d := range dormant {
		if d.ClusterID == clusterID {
			return true
		}
	}
	return false
}
