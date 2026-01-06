package taxonomy

import (
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// Classify auto-classifies a beat by Channel and Source
func Classify(beat model.Beat) model.Taxonomy {
	label := strings.ToLower(beat.Impetus.Label)
	content := strings.ToLower(beat.Content)

	channel := detectChannel(label, content)
	source := detectSource(label, beat.Impetus.Meta)
	confidence := calculateConfidence(label, content, channel, source)

	return model.Taxonomy{
		Channel:    channel,
		Source:     source,
		Confidence: confidence,
	}
}

func detectChannel(label, content string) model.Channel {
	scores := make(map[model.Channel]int)

	for ch, patterns := range ChannelPatterns {
		for _, pattern := range patterns {
			if strings.Contains(label, pattern) {
				scores[ch] += 3
			}
			if strings.Contains(content, pattern) {
				scores[ch] += 1
			}
		}
	}

	best := model.ChannelUnknown
	bestScore := 0
	for ch, score := range scores {
		if score > bestScore {
			best = ch
			bestScore = score
		}
	}

	if bestScore == 0 {
		return model.ChannelDiscovery
	}

	return best
}

func detectSource(label string, meta model.ImpetusMeta) model.Source {
	if ch, ok := meta["channel"]; ok {
		if source, found := MetaChannelMap[strings.ToLower(ch)]; found {
			return source
		}
	}

	scores := make(map[model.Source]int)
	for src, patterns := range SourcePatterns {
		for _, pattern := range patterns {
			if strings.Contains(label, pattern) {
				scores[src] += 2
			}
		}
	}

	best := model.SourceUnknown
	bestScore := 0
	for src, score := range scores {
		if score > bestScore {
			best = src
			bestScore = score
		}
	}

	if bestScore == 0 {
		return model.SourceInternal
	}

	return best
}

func calculateConfidence(label, content string, ch model.Channel, src model.Source) float64 {
	confidence := 0.3

	if label != "" && label != "manual entry" {
		confidence += 0.3
	}

	if ch != model.ChannelUnknown {
		confidence += 0.2
	}

	if src != model.SourceUnknown {
		confidence += 0.2
	}

	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// ClassifyAll classifies all beats and returns a map of beat ID to taxonomy
func ClassifyAll(beats []model.Beat) map[string]model.Taxonomy {
	result := make(map[string]model.Taxonomy)
	for _, beat := range beats {
		result[beat.ID] = Classify(beat)
	}
	return result
}
