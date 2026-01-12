package divergence

import (
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// BeatOrigin represents whether human or agent captured
type BeatOrigin int

const (
	OriginHuman BeatOrigin = iota
	OriginAgent
	OriginUnknown
)

func (o BeatOrigin) String() string {
	switch o {
	case OriginHuman:
		return "human"
	case OriginAgent:
		return "agent"
	default:
		return "unknown"
	}
}

// ClassifierConfig holds classification settings
type ClassifierConfig struct {
	AgentPatterns []string
}

// DefaultClassifierConfig returns default settings
func DefaultClassifierConfig() ClassifierConfig {
	return ClassifierConfig{
		AgentPatterns: []string{"droid-session", "factory", "agent", "claude", "droid"},
	}
}

// Classifier classifies beats by origin
type Classifier struct {
	config ClassifierConfig
}

// NewClassifier creates a classifier with config
func NewClassifier(config ClassifierConfig) *Classifier {
	return &Classifier{config: config}
}

// Classify returns the origin of a beat
func (c *Classifier) Classify(beat model.Beat) BeatOrigin {
	// Check label and raw fields of Impetus
	impetusText := strings.ToLower(beat.Impetus.Label + " " + beat.Impetus.Raw)

	for _, pattern := range c.config.AgentPatterns {
		if strings.Contains(impetusText, strings.ToLower(pattern)) {
			return OriginAgent
		}
	}

	// Check metadata if available
	if beat.Impetus.Label == "" && beat.Impetus.Raw == "" {
		return OriginUnknown
	}

	return OriginHuman
}

// ClassifyAll classifies all beats, returns maps by origin
func (c *Classifier) ClassifyAll(beats []model.Beat) (human, agent []model.Beat) {
	for _, b := range beats {
		switch c.Classify(b) {
		case OriginAgent:
			agent = append(agent, b)
		case OriginHuman:
			human = append(human, b)
		default:
			human = append(human, b) // Default unknown to human
		}
	}
	return
}
