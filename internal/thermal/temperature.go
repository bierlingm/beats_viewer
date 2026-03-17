package thermal

import (
	"math"
	"os"
	"path/filepath"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
	"gopkg.in/yaml.v3"
)

// Beat represents a beat with context for temperature computation.
type Beat struct {
	ID        string     `json:"id"`
	CreatedAt time.Time  `json:"created_at"`
	Content   string     `json:"content"`
	Context   *BeatContext `json:"context,omitempty"`
}

// ClusterTemperature contains computed temperature data for a cluster
type ClusterTemperature struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Temperature float64  `json:"temperature"`
	RawScore    float64  `json:"raw_score"`
	BeatCount   int      `json:"beat_count"`
	Ripeness    float64  `json:"ripeness"`
	Trend       string   `json:"trend,omitempty"`
	TrendDelta  float64  `json:"trend_delta,omitempty"`
	ClaimedBy   []string `json:"claimed_by"`
	Unclaimed   bool     `json:"unclaimed"`
	BeatIDs     []string `json:"beat_ids"`
}

// DirectoryTemperature contains computed temperature data for a directory.
type DirectoryTemperature struct {
	Path              string   `json:"path"`
	Temperature       float64  `json:"temperature"`
	RawScore          float64  `json:"raw_score"`
	Gravity           string   `json:"gravity"`
	GravityMultiplier float64  `json:"gravity_multiplier"`
	StateInferred     string   `json:"state_inferred"`
	Trend             string   `json:"trend,omitempty"`
	TrendDelta        float64  `json:"trend_delta,omitempty"`
	DominantCluster   string   `json:"dominant_cluster,omitempty"`
	ClaimedClusters   []string `json:"claimed_clusters"`
	ClaimedBeatCount  int      `json:"claimed_beat_count"`
}

// TemperatureOutput is the full output of temperature computation.
type TemperatureOutput struct {
	ComputedAt     time.Time                        `json:"computed_at"`
	WindowDays     int                              `json:"window_days"`
	BeatsAnalyzed  int                              `json:"beats_analyzed"`
	BeatsStores    []string                         `json:"beats_stores"`
	Aperture       string                           `json:"aperture"`
	Clusters       map[string]*ClusterTemperature   `json:"clusters"`
	Directories    map[string]*DirectoryTemperature `json:"directories"`
	Cooperators    map[string]*DirectoryTemperature `json:"cooperators,omitempty"`
	Normalization  NormalizationInfo                `json:"normalization"`
	ApertureNote   string                           `json:"aperture_note,omitempty"`
}

type NormalizationInfo struct {
	MaxRawScore float64 `json:"max_raw_score"`
	Method      string  `json:"method"`
}

// ComputeClusterTemperatures computes temperatures for all clusters
// Algorithm: recency-weighted beat score × ripeness factor
// Clusters are the PRIMARY unit of heat
func ComputeClusterTemperatures(clusters []model.Cluster, beats []Beat, directories []WALDDirectory, config *Config) map[string]*ClusterTemperature {
	if config == nil {
		config = DefaultConfig()
	}

	now := time.Now()
	windowStart := now.AddDate(0, 0, -config.Temperature.WindowDays)

	// Build beat lookup map
	beatMap := make(map[string]Beat)
	for _, beat := range beats {
		beatMap[beat.ID] = beat
	}

	result := make(map[string]*ClusterTemperature)

	for _, cluster := range clusters {
		ct := &ClusterTemperature{
			ID:        cluster.ID,
			Name:      cluster.Name,
			Ripeness:  cluster.RipenessScore,
			BeatIDs:   cluster.BeatIDs,
			ClaimedBy: []string{},
			Unclaimed: true,
		}

		// Compute recency-weighted score
		score := 0.0
		beatCount := 0
		for _, beatID := range cluster.BeatIDs {
			beat, ok := beatMap[beatID]
			if !ok {
				continue
			}
			// Skip beats outside window
			if beat.CreatedAt.Before(windowStart) {
				continue
			}
			ageDays := now.Sub(beat.CreatedAt).Hours() / 24
			weight := math.Exp(-config.Temperature.RecencyDecayLambda * ageDays)
			score += weight
			beatCount++
		}

		// Factor in ripeness
		score *= (1 + cluster.RipenessScore*0.5)

		ct.RawScore = score
		ct.BeatCount = beatCount
		result[cluster.ID] = ct
	}

	// Normalize temperatures
	maxScore := 0.0
	for _, ct := range result {
		if ct.RawScore > maxScore {
			maxScore = ct.RawScore
		}
	}
	for _, ct := range result {
		if maxScore > 0 {
			ct.Temperature = math.Min(1.0, ct.RawScore/maxScore)
		}
	}

	return result
}

// ComputeTemperature computes temperatures for all WALD directories.
func ComputeTemperature(beats []Beat, clusters []model.Cluster, directories []WALDDirectory, config *Config, werkRoot string) *TemperatureOutput {
	if config == nil {
		config = DefaultConfig()
	}

	now := time.Now()

	output := &TemperatureOutput{
		ComputedAt:    now,
		WindowDays:    config.Temperature.WindowDays,
		BeatsAnalyzed: len(beats),
		Aperture:      "articulable_noticing",
		Clusters:      make(map[string]*ClusterTemperature),
		Directories:   make(map[string]*DirectoryTemperature),
		Cooperators:   make(map[string]*DirectoryTemperature),
		ApertureNote:  "Shows articulable noticing only. Deep work, flow states, and embodied practice are not captured.",
	}

	// Compute cluster temperatures first (clusters are PRIMARY unit of heat)
	output.Clusters = ComputeClusterTemperatures(clusters, beats, directories, config)

	// Compute directory temperatures from claimed cluster temperatures
	for _, dir := range directories {
		gravity := dir.Gravity
		if gravity == "" {
			gravity = "normal"
		}

		gravityMult := GetGravityMultiplier(gravity, config)
		claimedClusters := GetClaimedClusters(dir, clusters)

		var claimedClusterIDs []string
		claimedBeatCount := 0
		weightedSum := 0.0
		totalWeight := 0.0

		for _, cluster := range claimedClusters {
			clusterTemp, ok := output.Clusters[cluster.ID]
			if !ok {
				continue
			}
			claimedClusterIDs = append(claimedClusterIDs, cluster.ID)
			claimedBeatCount += clusterTemp.BeatCount

			// Weight by log of cluster's beat count
			weight := math.Log(1 + float64(clusterTemp.BeatCount))
			weightedSum += clusterTemp.Temperature * weight
			totalWeight += weight
		}

		baseTemp := 0.0
		if totalWeight > 0 {
			baseTemp = weightedSum / totalWeight
		}

		// Apply gravity modifier
		temperature := math.Min(1.0, baseTemp*gravityMult)

		output.Directories[dir.Path] = &DirectoryTemperature{
			Path:              dir.Path,
			Temperature:       temperature,
			RawScore:          baseTemp,
			Gravity:           gravity,
			GravityMultiplier: gravityMult,
			StateInferred:     InferState(temperature, config),
			ClaimedClusters:   claimedClusterIDs,
			ClaimedBeatCount:  claimedBeatCount,
		}
	}

	// Find max score for normalization info
	maxScore := 0.0
	for _, dirTemp := range output.Directories {
		if dirTemp.RawScore > maxScore {
			maxScore = dirTemp.RawScore
		}
	}

	output.Normalization = NormalizationInfo{
		MaxRawScore: maxScore,
		Method:      "weighted_average_claimed_clusters",
	}

	return output
}

// InferState determines the state based on temperature thresholds.
func InferState(temperature float64, config *Config) string {
	if config == nil {
		config = DefaultConfig()
	}

	if temperature >= config.Thresholds.Hot {
		return "hot"
	}
	if temperature >= config.Thresholds.Warm {
		return "warm"
	}
	if temperature >= config.Thresholds.Cool {
		return "cool"
	}
	return "cold"
}

// ComputeTrend determines the temperature trend.
func ComputeTrend(current, previous float64, config *Config) (string, float64) {
	if config == nil {
		config = DefaultConfig()
	}

	delta := current - previous
	if delta > config.Trend.SignificanceThreshold {
		return "rising", delta
	}
	if delta < -config.Trend.SignificanceThreshold {
		return "falling", delta
	}
	return "stable", delta
}

// LoadConfig loads the thermal config from .wald/config.yaml.
func LoadConfig(werkRoot string) (*Config, error) {
	configPath := filepath.Join(werkRoot, ".wald", "config.yaml")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultConfig(), nil
		}
		return nil, err
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// LoadWALD loads WALD.yaml and returns directories.
type WALDFile struct {
	Version       int                 `yaml:"version"`
	GeneratedAt   string              `yaml:"generated_at"`
	ThermalConfig *ThermalWALDConfig  `yaml:"thermal_config,omitempty"`
	Roots         map[string]RootInfo `yaml:"roots"`
	Directories   []WALDDirectory     `yaml:"directories"`
}

type ThermalWALDConfig struct {
	Enabled          bool    `yaml:"enabled"`
	WindowDays       int     `yaml:"window_days"`
	GravityBoostHigh float64 `yaml:"gravity_boost_high"`
	GravityBoostLow  float64 `yaml:"gravity_boost_low"`
	Thresholds       struct {
		Hot  float64 `yaml:"hot"`
		Warm float64 `yaml:"warm"`
		Cool float64 `yaml:"cool"`
	} `yaml:"thresholds"`
}

type RootInfo struct {
	Purpose string `yaml:"purpose"`
	State   string `yaml:"state"`
}

func LoadWALD(werkRoot string) (*WALDFile, error) {
	waldPath := filepath.Join(werkRoot, "WALD.yaml")
	data, err := os.ReadFile(waldPath)
	if err != nil {
		return nil, err
	}

	var wald WALDFile
	if err := yaml.Unmarshal(data, &wald); err != nil {
		return nil, err
	}

	return &wald, nil
}

// SaveWALD writes the WALD structure back to WALD.yaml
func SaveWALD(werkRoot string, wald *WALDFile) error {
	waldPath := filepath.Join(werkRoot, "WALD.yaml")
	data, err := yaml.Marshal(wald)
	if err != nil {
		return err
	}
	return os.WriteFile(waldPath, data, 0644)
}

// FindWerkRoot finds the werk root by looking for WALD.yaml.
func FindWerkRoot() string {
	// Check BEATS_ROOT first
	if root := os.Getenv("BEATS_ROOT"); root != "" {
		if _, err := os.Stat(filepath.Join(root, "WALD.yaml")); err == nil {
			return root
		}
	}

	// Walk up from cwd
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, "WALD.yaml")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Try home/werk
	home, _ := os.UserHomeDir()
	werkPath := filepath.Join(home, "werk")
	if _, err := os.Stat(filepath.Join(werkPath, "WALD.yaml")); err == nil {
		return werkPath
	}

	return ""
}
