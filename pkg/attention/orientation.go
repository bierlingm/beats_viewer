package attention

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// OrientationItem represents a topic in the orientation summary
type OrientationItem struct {
	ClusterID   string
	ClusterName string
	Weight      float64 // 0.0-1.0, normalized
	BeatCount   int
	Trend       string // ↑ ↓ →
	Signal      string // Why this is notable
}

// OrientationSummary provides a computed summary of attention direction
type OrientationSummary struct {
	ComputedAt    time.Time
	Window        time.Duration
	TopTopics     []OrientationItem // Weighted by recency + volume
	Growing       []OrientationItem // Positive drift
	Crystallizing []OrientationItem // High ripeness + recent activity
	Emerging      []OrientationItem // New clusters forming
}

// OrientationConfig holds computation parameters
type OrientationConfig struct {
	Window         time.Duration // Default 30 days
	TopN           int           // Number of top topics to return (default 5)
	RecencyWeight  float64       // Weight for recency vs volume (default 0.6)
	RipenessThresh float64       // Threshold for crystallizing (default 0.6)
}

// DefaultOrientationConfig returns default orientation settings
func DefaultOrientationConfig() OrientationConfig {
	return OrientationConfig{
		Window:         30 * 24 * time.Hour,
		TopN:           5,
		RecencyWeight:  0.6,
		RipenessThresh: 0.6,
	}
}

// ComputeOrientation generates an orientation summary
func ComputeOrientation(
	beats []model.Beat,
	clusters []model.Cluster,
	activations []Activation,
	drift *DriftReport,
	ripeness map[string]float64,
	config OrientationConfig,
) *OrientationSummary {
	now := time.Now()

	summary := &OrientationSummary{
		ComputedAt:    now,
		Window:        config.Window,
		TopTopics:     []OrientationItem{},
		Growing:       []OrientationItem{},
		Crystallizing: []OrientationItem{},
		Emerging:      []OrientationItem{},
	}

	// Build cluster scores for top topics
	clusterScores := computeClusterScores(beats, clusters, config)

	// Top topics by weighted score
	summary.TopTopics = topNByScore(clusterScores, clusters, config.TopN)

	// Growing topics from drift report
	if drift != nil {
		summary.Growing = driftToOrientation(drift.Rising, "rising attention")
	}

	// Crystallizing: high ripeness + recent activity (from activations)
	summary.Crystallizing = findCrystallizing(clusters, activations, ripeness, config.RipenessThresh)

	// Emerging: from drift emerged items
	if drift != nil {
		summary.Emerging = driftToOrientation(drift.Emerged, "new pattern")
	}

	return summary
}

// computeClusterScores calculates weighted scores for each cluster
func computeClusterScores(beats []model.Beat, clusters []model.Cluster, config OrientationConfig) map[string]float64 {
	now := time.Now()
	windowBeats := BeatsInWindow(beats, config.Window)

	// Count by cluster
	counts := CountByCluster(windowBeats, clusters)

	// Find max count for normalization
	maxCount := 0
	for _, c := range counts {
		if c > maxCount {
			maxCount = c
		}
	}

	// Calculate recency-weighted scores
	scores := make(map[string]float64)

	// Build beat ID to cluster mapping
	beatToCluster := make(map[string]string)
	for _, c := range clusters {
		for _, bid := range c.BeatIDs {
			beatToCluster[bid] = c.ID
		}
	}

	// Recency score per cluster (more recent = higher weight)
	recencyScores := make(map[string]float64)
	for _, b := range windowBeats {
		cid, ok := beatToCluster[b.ID]
		if !ok {
			continue
		}

		// Days ago (0 = today, higher = older)
		daysAgo := now.Sub(b.CreatedAt).Hours() / 24
		// Recency weight: 1.0 for today, decaying over window
		windowDays := config.Window.Hours() / 24
		recencyWeight := 1.0 - (daysAgo / windowDays)
		if recencyWeight < 0 {
			recencyWeight = 0
		}
		recencyScores[cid] += recencyWeight
	}

	// Find max recency for normalization
	maxRecency := 0.0
	for _, r := range recencyScores {
		if r > maxRecency {
			maxRecency = r
		}
	}

	// Combine volume and recency
	for cid, count := range counts {
		volumeScore := 0.0
		if maxCount > 0 {
			volumeScore = float64(count) / float64(maxCount)
		}

		recencyScore := 0.0
		if maxRecency > 0 {
			recencyScore = recencyScores[cid] / maxRecency
		}

		// Weighted combination
		scores[cid] = config.RecencyWeight*recencyScore + (1-config.RecencyWeight)*volumeScore
	}

	return scores
}

// topNByScore returns top N clusters by score
func topNByScore(scores map[string]float64, clusters []model.Cluster, n int) []OrientationItem {
	// Build name map
	names := make(map[string]string)
	beatCounts := make(map[string]int)
	for _, c := range clusters {
		names[c.ID] = c.Name
		beatCounts[c.ID] = len(c.BeatIDs)
	}

	// Convert to slice
	type scoredCluster struct {
		id    string
		score float64
	}

	var scored []scoredCluster
	for id, score := range scores {
		scored = append(scored, scoredCluster{id, score})
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Take top N
	if len(scored) > n {
		scored = scored[:n]
	}

	// Convert to OrientationItems
	var items []OrientationItem
	for _, sc := range scored {
		items = append(items, OrientationItem{
			ClusterID:   sc.id,
			ClusterName: names[sc.id],
			Weight:      sc.score,
			BeatCount:   beatCounts[sc.id],
			Trend:       "→",
			Signal:      "active topic",
		})
	}

	return items
}

// driftToOrientation converts drift items to orientation items
func driftToOrientation(items []DriftItem, signal string) []OrientationItem {
	var result []OrientationItem
	for _, d := range items {
		result = append(result, OrientationItem{
			ClusterID:   d.ClusterID,
			ClusterName: d.ClusterName,
			Weight:      normalizeChangePercent(d.ChangePercent),
			BeatCount:   d.CurrentCount,
			Trend:       d.Direction.Symbol(),
			Signal:      signal,
		})
	}
	return result
}

// findCrystallizing identifies topics with high ripeness and recent activity
func findCrystallizing(clusters []model.Cluster, activations []Activation, ripeness map[string]float64, threshold float64) []OrientationItem {
	// Build set of activating cluster IDs
	activating := make(map[string]bool)
	for _, a := range activations {
		activating[a.ClusterID] = true
	}

	var items []OrientationItem
	for _, c := range clusters {
		// Check if activating
		if !activating[c.ID] {
			continue
		}

		// Check if cluster has high ripeness
		if c.RipenessScore < threshold {
			continue
		}

		items = append(items, OrientationItem{
			ClusterID:   c.ID,
			ClusterName: c.Name,
			Weight:      c.RipenessScore,
			BeatCount:   len(c.BeatIDs),
			Trend:       "↑",
			Signal:      "high ripeness + active",
		})
	}

	// Sort by ripeness descending
	sort.Slice(items, func(i, j int) bool {
		return items[i].Weight > items[j].Weight
	})

	return items
}

// normalizeChangePercent converts change percent to 0-1 weight
func normalizeChangePercent(change float64) float64 {
	// Cap at 200% change for normalization
	if change > 200 {
		change = 200
	}
	if change < -200 {
		change = -200
	}
	// Convert to 0-1 scale (100% = 0.5 weight)
	return (change + 200) / 400
}

// IsEmpty returns true if the orientation summary has no content
func (o *OrientationSummary) IsEmpty() bool {
	return len(o.TopTopics) == 0 && len(o.Growing) == 0 &&
		len(o.Crystallizing) == 0 && len(o.Emerging) == 0
}
