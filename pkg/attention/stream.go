package attention

import (
	"strings"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// Channel represents the capture channel (how the beat was captured)
type Channel string

const (
	ChannelCLI     Channel = "cli"
	ChannelAgent   Channel = "agent"
	ChannelUnknown Channel = "unknown"
)

// Source represents the capture source (what triggered the capture)
type Source string

// AttentionStream models captures as a temporal stream
type AttentionStream struct {
	Beats     []model.Beat
	StartTime time.Time
	EndTime   time.Time
	Window    time.Duration
	FlowRate  float64        // beats per day
	ByCluster map[string]int // cluster ID -> beat count
	ByChannel map[Channel]int
	BySource  map[Source]int
}

// NewAttentionStream creates a stream from beats within the given window
func NewAttentionStream(beats []model.Beat, clusters []model.Cluster, window time.Duration) *AttentionStream {
	now := time.Now()
	windowBeats := BeatsInWindow(beats, window)

	s := &AttentionStream{
		Beats:     windowBeats,
		StartTime: now.Add(-window),
		EndTime:   now,
		Window:    window,
		ByCluster: make(map[string]int),
		ByChannel: make(map[Channel]int),
		BySource:  make(map[Source]int),
	}

	// Build beat ID to cluster mapping
	beatToCluster := make(map[string]string)
	for _, c := range clusters {
		for _, bid := range c.BeatIDs {
			beatToCluster[bid] = c.ID
		}
	}

	// Compute aggregations
	for _, beat := range windowBeats {
		// By cluster
		if cid, ok := beatToCluster[beat.ID]; ok {
			s.ByCluster[cid]++
		}

		// By channel
		ch := ClassifyChannel(beat)
		s.ByChannel[ch]++

		// By source
		src := ClassifySource(beat)
		s.BySource[src]++
	}

	s.FlowRate = s.FlowRatePerDay()
	return s
}

// FlowRatePerDay computes beats per day in the stream
func (s *AttentionStream) FlowRatePerDay() float64 {
	if s.Window <= 0 {
		return 0
	}
	days := s.Window.Hours() / 24
	if days <= 0 {
		return 0
	}
	return float64(len(s.Beats)) / days
}

// BeatsInWindow returns beats within a specific time window from now
func BeatsInWindow(beats []model.Beat, window time.Duration) []model.Beat {
	now := time.Now()
	cutoff := now.Add(-window)

	var result []model.Beat
	for _, b := range beats {
		if !b.CreatedAt.Before(cutoff) {
			result = append(result, b)
		}
	}
	return result
}

// BeatsInTimeRange returns beats within a specific time range
func BeatsInTimeRange(beats []model.Beat, start, end time.Time) []model.Beat {
	var result []model.Beat
	for _, b := range beats {
		if !b.CreatedAt.Before(start) && !b.CreatedAt.After(end) {
			result = append(result, b)
		}
	}
	return result
}

// BeatsInCluster returns beats that belong to a specific cluster
func BeatsInCluster(beats []model.Beat, cluster model.Cluster) []model.Beat {
	beatSet := make(map[string]bool)
	for _, bid := range cluster.BeatIDs {
		beatSet[bid] = true
	}

	var result []model.Beat
	for _, b := range beats {
		if beatSet[b.ID] {
			result = append(result, b)
		}
	}
	return result
}

// ClassifyChannel determines the channel for a beat based on impetus
func ClassifyChannel(beat model.Beat) Channel {
	label := strings.ToLower(beat.Impetus.Label)
	raw := strings.ToLower(beat.Impetus.Raw)

	agentPatterns := []string{"droid", "agent", "factory", "session"}
	for _, pattern := range agentPatterns {
		if strings.Contains(label, pattern) || strings.Contains(raw, pattern) {
			return ChannelAgent
		}
	}

	// Check meta for agent indicators
	for _, v := range beat.Impetus.Meta {
		vLower := strings.ToLower(v)
		for _, pattern := range agentPatterns {
			if strings.Contains(vLower, pattern) {
				return ChannelAgent
			}
		}
	}

	if label != "" {
		return ChannelCLI
	}

	return ChannelUnknown
}

// ClassifySource extracts the source from a beat's impetus
func ClassifySource(beat model.Beat) Source {
	if beat.Impetus.Label != "" {
		return Source(beat.Impetus.Label)
	}
	return Source("unknown")
}

// CountByCluster returns the number of beats per cluster for given beats
func CountByCluster(beats []model.Beat, clusters []model.Cluster) map[string]int {
	beatToCluster := make(map[string]string)
	for _, c := range clusters {
		for _, bid := range c.BeatIDs {
			beatToCluster[bid] = c.ID
		}
	}

	counts := make(map[string]int)
	for _, b := range beats {
		if cid, ok := beatToCluster[b.ID]; ok {
			counts[cid]++
		}
	}
	return counts
}
