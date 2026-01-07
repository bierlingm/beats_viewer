package attention

import (
	"math"
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// DriftDirection represents how attention is shifting
type DriftDirection int

const (
	DriftRising   DriftDirection = iota // Increasing attention
	DriftStable                          // Stable attention
	DriftFading                          // Decreasing attention
	DriftEmerged                         // New (not in prior window)
	DriftVanished                        // Gone (was in prior, zero now)
)

func (d DriftDirection) String() string {
	switch d {
	case DriftRising:
		return "rising"
	case DriftStable:
		return "stable"
	case DriftFading:
		return "fading"
	case DriftEmerged:
		return "emerged"
	case DriftVanished:
		return "vanished"
	default:
		return "unknown"
	}
}

// Symbol returns the visual symbol for drift direction
func (d DriftDirection) Symbol() string {
	switch d {
	case DriftRising:
		return "↑"
	case DriftStable:
		return "→"
	case DriftFading:
		return "↓"
	case DriftEmerged:
		return "+"
	case DriftVanished:
		return "×"
	default:
		return "?"
	}
}

// DriftItem represents drift for a single cluster
type DriftItem struct {
	ClusterID     string
	ClusterName   string
	CurrentCount  int
	PriorCount    int
	ChangePercent float64
	Direction     DriftDirection
}

// DriftReport contains the full drift analysis
type DriftReport struct {
	Window      time.Duration // Current window (default 30 days)
	PriorWindow time.Duration // Comparison window
	ComputedAt  time.Time
	Rising      []DriftItem
	Stable      []DriftItem
	Fading      []DriftItem
	Emerged     []DriftItem
	Vanished    []DriftItem
}

// DriftConfig holds analysis parameters
type DriftConfig struct {
	Window          time.Duration // Default 30 days
	StableThreshold float64       // Default 0.2 (20% change = stable)
}

// DefaultDriftConfig returns default drift analysis settings
func DefaultDriftConfig() DriftConfig {
	return DriftConfig{
		Window:          30 * 24 * time.Hour,
		StableThreshold: 0.2,
	}
}

// ComputeDrift analyzes attention drift between time windows
func ComputeDrift(beats []model.Beat, clusters []model.Cluster, config DriftConfig) *DriftReport {
	now := time.Now()

	currentStart := now.Add(-config.Window)
	priorStart := currentStart.Add(-config.Window)

	// Get beats in each window
	currentBeats := BeatsInTimeRange(beats, currentStart, now)
	priorBeats := BeatsInTimeRange(beats, priorStart, currentStart)

	// Count by cluster for each window
	currentCounts := CountByCluster(currentBeats, clusters)
	priorCounts := CountByCluster(priorBeats, clusters)

	report := &DriftReport{
		Window:      config.Window,
		PriorWindow: config.Window,
		ComputedAt:  now,
		Rising:      []DriftItem{},
		Stable:      []DriftItem{},
		Fading:      []DriftItem{},
		Emerged:     []DriftItem{},
		Vanished:    []DriftItem{},
	}

	// Build cluster ID to name map
	clusterNames := make(map[string]string)
	for _, c := range clusters {
		clusterNames[c.ID] = c.Name
	}

	// Collect all cluster IDs from both windows
	allClusters := make(map[string]bool)
	for id := range currentCounts {
		allClusters[id] = true
	}
	for id := range priorCounts {
		allClusters[id] = true
	}

	// Analyze each cluster
	for clusterID := range allClusters {
		current := currentCounts[clusterID]
		prior := priorCounts[clusterID]

		direction := CategorizeDrift(current, prior, config.StableThreshold)
		changePercent := computeChangePercent(current, prior)

		item := DriftItem{
			ClusterID:     clusterID,
			ClusterName:   clusterNames[clusterID],
			CurrentCount:  current,
			PriorCount:    prior,
			ChangePercent: changePercent,
			Direction:     direction,
		}

		switch direction {
		case DriftRising:
			report.Rising = append(report.Rising, item)
		case DriftStable:
			report.Stable = append(report.Stable, item)
		case DriftFading:
			report.Fading = append(report.Fading, item)
		case DriftEmerged:
			report.Emerged = append(report.Emerged, item)
		case DriftVanished:
			report.Vanished = append(report.Vanished, item)
		}
	}

	// Sort each category by absolute change magnitude
	sortByChange := func(items []DriftItem) {
		sort.Slice(items, func(i, j int) bool {
			return math.Abs(items[i].ChangePercent) > math.Abs(items[j].ChangePercent)
		})
	}

	sortByChange(report.Rising)
	sortByChange(report.Fading)
	sortByChange(report.Stable)

	// Sort emerged/vanished by count
	sort.Slice(report.Emerged, func(i, j int) bool {
		return report.Emerged[i].CurrentCount > report.Emerged[j].CurrentCount
	})
	sort.Slice(report.Vanished, func(i, j int) bool {
		return report.Vanished[i].PriorCount > report.Vanished[j].PriorCount
	})

	return report
}

// CategorizeDrift determines the drift direction based on counts
func CategorizeDrift(current, prior int, threshold float64) DriftDirection {
	// Emerged: no prior activity, has current activity
	if prior == 0 && current > 0 {
		return DriftEmerged
	}

	// Vanished: had prior activity, no current activity
	if prior > 0 && current == 0 {
		return DriftVanished
	}

	// No activity in either window - treat as stable
	if prior == 0 && current == 0 {
		return DriftStable
	}

	// Calculate change percentage
	changePercent := computeChangePercent(current, prior)

	if changePercent > threshold*100 {
		return DriftRising
	}
	if changePercent < -threshold*100 {
		return DriftFading
	}

	return DriftStable
}

// computeChangePercent calculates the percentage change
func computeChangePercent(current, prior int) float64 {
	if prior == 0 {
		if current == 0 {
			return 0
		}
		return 100 // Treat as 100% increase for emerged
	}
	return (float64(current) - float64(prior)) / float64(prior) * 100
}

// TotalChanges returns the total number of non-stable clusters
func (r *DriftReport) TotalChanges() int {
	return len(r.Rising) + len(r.Fading) + len(r.Emerged) + len(r.Vanished)
}

// AllItems returns all drift items as a flat list
func (r *DriftReport) AllItems() []DriftItem {
	var items []DriftItem
	items = append(items, r.Rising...)
	items = append(items, r.Stable...)
	items = append(items, r.Fading...)
	items = append(items, r.Emerged...)
	items = append(items, r.Vanished...)
	return items
}
