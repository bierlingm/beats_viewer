package divergence

import "sort"

// BlindSpotConfig configures blind spot detection
type BlindSpotConfig struct {
	MinAgentCaptures int
}

// DefaultBlindSpotConfig returns defaults
func DefaultBlindSpotConfig() BlindSpotConfig {
	return BlindSpotConfig{
		MinAgentCaptures: 3,
	}
}

// DetectBlindSpots finds topics agents notice but human doesn't
func DetectBlindSpots(
	humanTopics map[string]int,
	agentTopics map[string]int,
	config BlindSpotConfig,
) []string {
	type blindSpot struct {
		topic string
		count int
	}

	var spots []blindSpot

	for topic, agentCount := range agentTopics {
		humanCount := humanTopics[topic]
		if humanCount == 0 && agentCount >= config.MinAgentCaptures {
			spots = append(spots, blindSpot{
				topic: topic,
				count: agentCount,
			})
		}
	}

	// Sort by agent count descending
	sort.Slice(spots, func(i, j int) bool {
		return spots[i].count > spots[j].count
	})

	result := make([]string, len(spots))
	for i, s := range spots {
		result[i] = s.topic
	}

	return result
}
