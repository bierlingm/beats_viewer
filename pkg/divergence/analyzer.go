package divergence

import (
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// DivergenceItem represents a topic with human/agent counts
type DivergenceItem struct {
	Topic      string  `json:"topic"`
	HumanCount int     `json:"human_count"`
	AgentCount int     `json:"agent_count"`
	Ratio      float64 `json:"ratio"`
}

// DivergenceReport summarizes human vs agent attention
type DivergenceReport struct {
	Window     time.Duration     `json:"window"`
	HumanBeats int               `json:"human_beats"`
	AgentBeats int               `json:"agent_beats"`
	HumanOnly  []DivergenceItem  `json:"human_only"`
	AgentOnly  []DivergenceItem  `json:"agent_only"`
	Amplified  []DivergenceItem  `json:"amplified"`
	BlindSpots []string          `json:"blind_spots"`
}

// AnalyzerConfig holds analysis settings
type AnalyzerConfig struct {
	Window           time.Duration
	MinAgentForBlind int
}

// DefaultAnalyzerConfig returns defaults
func DefaultAnalyzerConfig() AnalyzerConfig {
	return AnalyzerConfig{
		Window:           30 * 24 * time.Hour,
		MinAgentForBlind: 3,
	}
}

// Analyzer compares human and agent attention
type Analyzer struct {
	classifier *Classifier
	config     AnalyzerConfig
}

// NewAnalyzer creates an analyzer
func NewAnalyzer(classifier *Classifier, config AnalyzerConfig) *Analyzer {
	return &Analyzer{
		classifier: classifier,
		config:     config,
	}
}

// Analyze produces a divergence report
func (a *Analyzer) Analyze(beats []model.Beat, clusters []model.Cluster) *DivergenceReport {
	now := time.Now()
	windowStart := now.Add(-a.config.Window)

	// Filter to window
	var windowBeats []model.Beat
	for _, b := range beats {
		if !b.CreatedAt.Before(windowStart) {
			windowBeats = append(windowBeats, b)
		}
	}

	human, agent := a.classifier.ClassifyAll(windowBeats)

	// Build cluster ID to name map
	clusterNames := make(map[string]string)
	for _, c := range clusters {
		clusterNames[c.ID] = c.Name
	}

	// Count topics per origin
	humanTopics := countTopics(human, clusters)
	agentTopics := countTopics(agent, clusters)

	report := &DivergenceReport{
		Window:     a.config.Window,
		HumanBeats: len(human),
		AgentBeats: len(agent),
	}

	// Categorize topics
	allTopics := mergeKeys(humanTopics, agentTopics)
	for topic := range allTopics {
		hCount := humanTopics[topic]
		aCount := agentTopics[topic]

		item := DivergenceItem{
			Topic:      topic,
			HumanCount: hCount,
			AgentCount: aCount,
		}

		if aCount > 0 {
			item.Ratio = float64(hCount) / float64(aCount)
		} else if hCount > 0 {
			item.Ratio = float64(hCount + 1) // Indicate human-only
		}

		if hCount > 0 && aCount == 0 {
			report.HumanOnly = append(report.HumanOnly, item)
		} else if aCount > 0 && hCount == 0 {
			report.AgentOnly = append(report.AgentOnly, item)
		} else if hCount > 0 && aCount > 0 {
			report.Amplified = append(report.Amplified, item)
		}
	}

	// Sort by counts
	sort.Slice(report.HumanOnly, func(i, j int) bool {
		return report.HumanOnly[i].HumanCount > report.HumanOnly[j].HumanCount
	})
	sort.Slice(report.AgentOnly, func(i, j int) bool {
		return report.AgentOnly[i].AgentCount > report.AgentOnly[j].AgentCount
	})
	sort.Slice(report.Amplified, func(i, j int) bool {
		return report.Amplified[i].HumanCount+report.Amplified[i].AgentCount >
			report.Amplified[j].HumanCount+report.Amplified[j].AgentCount
	})

	// Detect blind spots
	report.BlindSpots = DetectBlindSpots(humanTopics, agentTopics, BlindSpotConfig{
		MinAgentCaptures: a.config.MinAgentForBlind,
	})

	return report
}

func countTopics(beats []model.Beat, clusters []model.Cluster) map[string]int {
	counts := make(map[string]int)

	// Build beat to cluster map
	beatCluster := make(map[string]string)
	for _, c := range clusters {
		for _, bid := range c.BeatIDs {
			beatCluster[bid] = c.Name
		}
	}

	for _, b := range beats {
		if cluster, ok := beatCluster[b.ID]; ok {
			counts[cluster]++
		}
	}

	return counts
}

func mergeKeys(a, b map[string]int) map[string]bool {
	result := make(map[string]bool)
	for k := range a {
		result[k] = true
	}
	for k := range b {
		result[k] = true
	}
	return result
}
