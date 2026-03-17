package thermal

import (
	"path/filepath"
	"strings"
)

// MatchResult contains the result of beat-to-directory matching.
type MatchResult struct {
	Matched    bool    `json:"matched"`
	Method     string  `json:"method,omitempty"`     // explicit_context, capture_path, semantic
	Confidence float64 `json:"confidence,omitempty"` // 0-1
}

// BeatContext represents the context field of a beat.
type BeatContext struct {
	CapturePath     string  `json:"capture_path"`
	WALDDirectory   string  `json:"wald_directory,omitempty"`
	InferenceMethod string  `json:"inference_method,omitempty"`
	Confidence      float64 `json:"confidence,omitempty"`
}

// WALDDirectory represents a directory entry from WALD.yaml.
type WALDDirectory struct {
	Path          string          `yaml:"path" json:"path"`
	Purpose       string          `yaml:"purpose" json:"purpose"`
	State         string          `yaml:"state" json:"state"`
	Entry         string          `yaml:"entry" json:"entry,omitempty"`
	Gravity       string          `yaml:"gravity" json:"gravity,omitempty"`
	Preserve      bool            `yaml:"preserve" json:"preserve,omitempty"`
	StateSource   string          `yaml:"state_source" json:"state_source,omitempty"`
	OriginCluster string          `yaml:"origin_cluster" json:"origin_cluster,omitempty"`
	Claims        DirectoryClaims `yaml:"claims" json:"claims,omitempty"`
}

// BeatBelongsToDirectory determines if a beat belongs to a WALD directory.
// Returns match result with confidence score.
//
// Matching methods (in priority order):
// 1. Explicit context match - beat.context.wald_directory == directory.path
// 2. Capture path prefix - beat.context.capture_path starts with directory path
// 3. Semantic similarity (future) - content similarity to directory purpose
func BeatBelongsToDirectory(beatContext *BeatContext, directory WALDDirectory, werkRoot string) MatchResult {
	if beatContext == nil {
		return MatchResult{Matched: false}
	}

	// Method 1: Explicit context match
	if beatContext.WALDDirectory != "" {
		// Exact match
		if beatContext.WALDDirectory == directory.Path {
			return MatchResult{
				Matched:    true,
				Method:     "explicit_context",
				Confidence: 1.0,
			}
		}
		// Check if beat's WALD directory is under this directory
		if strings.HasPrefix(beatContext.WALDDirectory, directory.Path+"/") {
			return MatchResult{
				Matched:    true,
				Method:     "explicit_context",
				Confidence: 0.9,
			}
		}
	}

	// Method 2: Capture path is within directory
	if beatContext.CapturePath != "" && werkRoot != "" {
		dirAbsPath := filepath.Join(werkRoot, directory.Path)

		// Check if capture path starts with directory path
		if strings.HasPrefix(beatContext.CapturePath, dirAbsPath+"/") ||
			beatContext.CapturePath == dirAbsPath {
			return MatchResult{
				Matched:    true,
				Method:     "capture_path",
				Confidence: 1.0,
			}
		}
	}

	// Method 3: Semantic similarity (placeholder - requires embeddings)
	// This would compare beat content embeddings to directory purpose embeddings
	// For now, we don't have embeddings integrated here

	return MatchResult{Matched: false}
}

// FindMatchingDirectory finds the best matching WALD directory for a beat context.
// Returns the matching directory path and confidence, or empty if no match.
func FindMatchingDirectory(beatContext *BeatContext, directories []WALDDirectory, werkRoot string) (string, float64) {
	if beatContext == nil {
		return "", 0
	}

	// If beat has explicit WALD directory, use it if it exists in directories
	if beatContext.WALDDirectory != "" {
		for _, dir := range directories {
			if dir.Path == beatContext.WALDDirectory {
				return dir.Path, 1.0
			}
		}
		// Try to find parent directory
		for _, dir := range directories {
			if strings.HasPrefix(beatContext.WALDDirectory, dir.Path+"/") {
				return dir.Path, 0.9
			}
		}
	}

	// Fall back to capture path matching
	var bestMatch string
	var bestConfidence float64
	var bestLength int

	for _, dir := range directories {
		result := BeatBelongsToDirectory(beatContext, dir, werkRoot)
		if result.Matched {
			// Prefer longer path matches (more specific)
			if len(dir.Path) > bestLength {
				bestMatch = dir.Path
				bestConfidence = result.Confidence
				bestLength = len(dir.Path)
			}
		}
	}

	return bestMatch, bestConfidence
}

// GetGravityMultiplier returns the multiplier for a gravity level.
func GetGravityMultiplier(gravity string, config *Config) float64 {
	if config == nil {
		// Default values
		switch gravity {
		case "high":
			return 1.5
		case "low":
			return 0.7
		default:
			return 1.0
		}
	}

	switch gravity {
	case "high":
		return config.Gravity.High
	case "low":
		return config.Gravity.Low
	default:
		return config.Gravity.Normal
	}
}

// Config represents the .wald/config.yaml configuration.
type Config struct {
	Version     int               `yaml:"version"`
	Temperature TemperatureConfig `yaml:"temperature"`
	Clustering  ClusteringConfig  `yaml:"clustering"`
	Claims      ClaimsConfig      `yaml:"claims"`
	Entities    EntitiesConfig    `yaml:"entities"`
	Gravity     GravityConfig     `yaml:"gravity"`
	Thresholds  ThresholdsConfig  `yaml:"thresholds"`
	Trend       TrendConfig       `yaml:"trend"`
	Emergence   EmergenceConfig   `yaml:"emergence"`
	Context     ContextConfig     `yaml:"context"`
	Display     DisplayConfig     `yaml:"display"`
	Sync        SyncConfig        `yaml:"sync"`
}

type TemperatureConfig struct {
	WindowDays              int     `yaml:"window_days"`
	RecencyDecayLambda      float64 `yaml:"recency_decay_lambda"`
	MinBeats                int     `yaml:"min_beats"`
	RecomputeIntervalHours  int     `yaml:"recompute_interval_hours"`
}

type GravityConfig struct {
	High   float64 `yaml:"high"`
	Normal float64 `yaml:"normal"`
	Low    float64 `yaml:"low"`
}

type ThresholdsConfig struct {
	Hot  float64 `yaml:"hot"`
	Warm float64 `yaml:"warm"`
	Cool float64 `yaml:"cool"`
}

type TrendConfig struct {
	SignificanceThreshold  float64 `yaml:"significance_threshold"`
	ComparisonWindowDays   int     `yaml:"comparison_window_days"`
}

type EmergenceConfig struct {
	MinBeatCount       int     `yaml:"min_beat_count"`
	MinRipeness        float64 `yaml:"min_ripeness"`
	MinUncoveredRatio  float64 `yaml:"min_uncovered_ratio"`
}

type ContextConfig struct {
	MaxResults                   int     `yaml:"max_results"`
	ClusterOverlapThreshold      float64 `yaml:"cluster_overlap_threshold"`
	SemanticSimilarityThreshold  float64 `yaml:"semantic_similarity_threshold"`
}

type DisplayConfig struct {
	ShowTemperatureInRouteHere bool `yaml:"show_temperature_in_route_here"`
	TemperaturePrecision       int  `yaml:"temperature_precision"`
	ShowTrendArrows            bool `yaml:"show_trend_arrows"`
	ShowApertureNote           bool `yaml:"show_aperture_note"`
}

type SyncConfig struct {
	BackupBeforeWrite bool   `yaml:"backup_before_write"`
	BackupDir         string `yaml:"backup_dir"`
	MaxBackups        int    `yaml:"max_backups"`
}

type ClusteringConfig struct {
	MinClusterSize         int     `yaml:"min_cluster_size"`
	SimilarityThreshold    float64 `yaml:"similarity_threshold"`
	RecomputeIntervalHours int     `yaml:"recompute_interval_hours"`
}

type ClaimsConfig struct {
	TopicMatchThreshold float64 `yaml:"topic_match_threshold"`
	KeywordMatchMode    string  `yaml:"keyword_match_mode"`
}

type EntitiesConfig struct {
	ExtractPeople   bool    `yaml:"extract_people"`
	ExtractProjects bool    `yaml:"extract_projects"`
	ExtractURLs     bool    `yaml:"extract_urls"`
	ExtractTopics   bool    `yaml:"extract_topics"`
	MinConfidence   float64 `yaml:"min_confidence"`
}

// DefaultConfig returns sensible defaults.
func DefaultConfig() *Config {
	return &Config{
		Version: 1,
		Temperature: TemperatureConfig{
			WindowDays:             30,
			RecencyDecayLambda:     0.1,
			MinBeats:               1,
			RecomputeIntervalHours: 1,
		},
		Gravity: GravityConfig{
			High:   1.5,
			Normal: 1.0,
			Low:    0.7,
		},
		Thresholds: ThresholdsConfig{
			Hot:  0.7,
			Warm: 0.4,
			Cool: 0.15,
		},
		Trend: TrendConfig{
			SignificanceThreshold: 0.05,
			ComparisonWindowDays:  7,
		},
		Emergence: EmergenceConfig{
			MinBeatCount:      10,
			MinRipeness:       0.4,
			MinUncoveredRatio: 0.6,
		},
		Context: ContextConfig{
			MaxResults:                  10,
			ClusterOverlapThreshold:     0.3,
			SemanticSimilarityThreshold: 0.6,
		},
		Display: DisplayConfig{
			ShowTemperatureInRouteHere: true,
			TemperaturePrecision:       2,
			ShowTrendArrows:            true,
			ShowApertureNote:           true,
		},
		Sync: SyncConfig{
			BackupBeforeWrite: true,
			BackupDir:         ".wald/backups",
			MaxBackups:        10,
		},
		Clustering: ClusteringConfig{
			MinClusterSize:         3,
			SimilarityThreshold:    0.65,
			RecomputeIntervalHours: 24,
		},
		Claims: ClaimsConfig{
			TopicMatchThreshold: 0.7,
			KeywordMatchMode:    "fuzzy",
		},
		Entities: EntitiesConfig{
			ExtractPeople:   true,
			ExtractProjects: true,
			ExtractURLs:     true,
			ExtractTopics:   true,
			MinConfidence:   0.7,
		},
	}
}
