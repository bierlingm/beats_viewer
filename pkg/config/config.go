package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Attention  AttentionConfig  `yaml:"attention"`
	Alerts     AlertsConfig     `yaml:"alerts"`
	Inference  InferenceConfig  `yaml:"inference"`
	Divergence DivergenceConfig `yaml:"divergence"`
}

type AttentionConfig struct {
	ActivationWindow    Duration `yaml:"activation_window"`
	ActivationThreshold int      `yaml:"activation_threshold"`
	DriftWindow         Duration `yaml:"drift_window"`
	DormancyThreshold   Duration `yaml:"dormancy_threshold"`
}

type AlertsConfig struct {
	Enabled bool `yaml:"enabled"`
}

type InferenceConfig struct {
	CrystallizationConfidence float64 `yaml:"crystallization_confidence"`
}

type DivergenceConfig struct {
	AgentImpetusPatterns []string `yaml:"agent_impetus_patterns"`
}

// Duration wraps time.Duration to support "30d" format in YAML
type Duration time.Duration

func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	s := value.Value

	// Handle day notation (e.g., "30d")
	if strings.HasSuffix(s, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return err
		}
		*d = Duration(time.Duration(days) * 24 * time.Hour)
		return nil
	}

	// Standard Go duration parsing
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = Duration(dur)
	return nil
}

func (d Duration) MarshalYAML() (interface{}, error) {
	return time.Duration(d).String(), nil
}

// DefaultConfig returns sensible defaults
func DefaultConfig() Config {
	return Config{
		Attention: AttentionConfig{
			ActivationWindow:    Duration(72 * time.Hour),
			ActivationThreshold: 3,
			DriftWindow:         Duration(30 * 24 * time.Hour),
			DormancyThreshold:   Duration(30 * 24 * time.Hour),
		},
		Alerts: AlertsConfig{
			Enabled: true,
		},
		Inference: InferenceConfig{
			CrystallizationConfidence: 0.6,
		},
		Divergence: DivergenceConfig{
			AgentImpetusPatterns: []string{"droid-session", "factory", "agent"},
		},
	}
}

// Load reads config from .beats/btv-config.yaml if exists, otherwise returns defaults
func Load(beatsDir string) (*Config, error) {
	cfg := DefaultConfig()

	configPath := filepath.Join(beatsDir, "btv-config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &cfg, nil
		}
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
