package thermal

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

const emergenceCacheFileName = "emergence.json"

// SuggestedClaims contains suggested cluster and topic claims
type SuggestedClaims struct {
	Clusters []string `json:"clusters"`
	Topics   []string `json:"topics"`
}

// PartiallyClaimed represents a cluster claimed by one dir but relevant to others
type PartiallyClaimed struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	ClaimedBy      []string `json:"claimed_by"`
	AlsoRelevantTo []string `json:"also_relevant_to"`
	Suggestion     string   `json:"suggestion"`
}

// ClaimSuggestion suggests additional claims for existing directories
type ClaimSuggestion struct {
	Directory       string          `json:"directory"`
	SuggestedClaims SuggestedClaims `json:"suggested_claims"`
	Reason          string          `json:"reason"`
}

// EmergenceOutput is the full output of emergence detection
type EmergenceOutput struct {
	ComputedAt       time.Time          `json:"computed_at"`
	Clusters         []EmergentCluster  `json:"unclaimed_clusters"`
	PartiallyClaimed []PartiallyClaimed `json:"partially_claimed"`
	ClaimSuggestions []ClaimSuggestion  `json:"claim_suggestions"`
	NotReady         []PendingCluster   `json:"not_ready"`
}

// EmergentCluster represents a cluster ready for crystallization
type EmergentCluster struct {
	ClusterID            string          `json:"id"`
	ClusterName          string          `json:"name"`
	BeatCount            int             `json:"beat_count"`
	Ripeness             float64         `json:"ripeness"`
	Temperature          float64         `json:"temperature"`
	Keywords             []string        `json:"keywords"`
	Coverage             CoverageInfo    `json:"coverage,omitempty"`
	CrystallizationReady bool            `json:"crystallization_ready"`
	SuggestedPath        string          `json:"suggested_path"`
	SuggestedPurpose     string          `json:"suggested_purpose,omitempty"`
	SuggestedClaims      SuggestedClaims `json:"suggested_claims"`
	SampleBeats          []BeatSample    `json:"sample_beats"`
}

// CoverageInfo tracks how many beats are covered by existing directories
type CoverageInfo struct {
	TotalBeats        int     `json:"total_beats"`
	CoveredByExisting int     `json:"covered_by_existing"`
	Uncovered         int     `json:"uncovered"`
	UncoveredRatio    float64 `json:"uncovered_ratio"`
}

// BeatSample is a beat preview for emergence output
type BeatSample struct {
	ID      string `json:"id"`
	Preview string `json:"preview"`
}

// PendingCluster represents a cluster not yet ready for emergence
type PendingCluster struct {
	ClusterID   string  `json:"cluster_id"`
	ClusterName string  `json:"cluster_name"`
	Reason      string  `json:"reason"`
	BeatCount   int     `json:"beat_count,omitempty"`
	Required    int     `json:"required,omitempty"`
	Ripeness    float64 `json:"ripeness,omitempty"`
}

// DetectEmergence identifies clusters ready for crystallization into WALD directories
func DetectEmergence(clusters []model.Cluster, directories []WALDDirectory, beats []Beat, config *Config) *EmergenceOutput {
	if config == nil {
		config = DefaultConfig()
	}

	beatMap := make(map[string]Beat)
	for _, b := range beats {
		beatMap[b.ID] = b
	}

	dirPaths := make(map[string]bool)
	for _, d := range directories {
		dirPaths[d.Path] = true
	}

	output := &EmergenceOutput{
		ComputedAt:       time.Now(),
		Clusters:         []EmergentCluster{},
		PartiallyClaimed: []PartiallyClaimed{},
		ClaimSuggestions: []ClaimSuggestion{},
		NotReady:         []PendingCluster{},
	}

	// Track which directories claim which clusters (by beat overlap)
	clusterToDirs := make(map[string][]string)
	dirToClusterOverlap := make(map[string]map[string]int) // dir -> cluster -> beat count

	for _, cluster := range clusters {
		for _, beatID := range cluster.BeatIDs {
			beat, ok := beatMap[beatID]
			if !ok {
				continue
			}
			if beat.Context != nil && beat.Context.WALDDirectory != "" {
				dir := beat.Context.WALDDirectory
				if dirPaths[dir] {
					// Track cluster->dir mapping
					found := false
					for _, d := range clusterToDirs[cluster.ID] {
						if d == dir {
							found = true
							break
						}
					}
					if !found {
						clusterToDirs[cluster.ID] = append(clusterToDirs[cluster.ID], dir)
					}
					// Track overlap counts
					if dirToClusterOverlap[dir] == nil {
						dirToClusterOverlap[dir] = make(map[string]int)
					}
					dirToClusterOverlap[dir][cluster.ID]++
				}
			}
		}
	}

	for _, cluster := range clusters {
		beatCount := len(cluster.BeatIDs)

		// Check beat count threshold
		if beatCount < config.Emergence.MinBeatCount {
			output.NotReady = append(output.NotReady, PendingCluster{
				ClusterID:   cluster.ID,
				ClusterName: cluster.Name,
				Reason:      "below_beat_threshold",
				BeatCount:   beatCount,
				Required:    config.Emergence.MinBeatCount,
			})
			continue
		}

		// Check ripeness threshold
		if cluster.RipenessScore < config.Emergence.MinRipeness {
			output.NotReady = append(output.NotReady, PendingCluster{
				ClusterID:   cluster.ID,
				ClusterName: cluster.Name,
				Reason:      "below_ripeness_threshold",
				Ripeness:    cluster.RipenessScore,
			})
			continue
		}

		// Check coverage by existing directories
		coveredCount := 0
		clusterBeats := []Beat{}
		coveringDirs := make(map[string]int)
		for _, beatID := range cluster.BeatIDs {
			beat, ok := beatMap[beatID]
			if !ok {
				continue
			}
			clusterBeats = append(clusterBeats, beat)
			if beat.Context != nil && beat.Context.WALDDirectory != "" {
				if dirPaths[beat.Context.WALDDirectory] {
					coveredCount++
					coveringDirs[beat.Context.WALDDirectory]++
				}
			}
		}

		uncoveredRatio := 1.0
		if beatCount > 0 {
			uncoveredRatio = 1.0 - float64(coveredCount)/float64(beatCount)
		}

		// Check if partially claimed (claimed by some dirs but relevant to others)
		claimedBy := clusterToDirs[cluster.ID]
		if len(claimedBy) > 0 && uncoveredRatio > 0.3 {
			// Find other directories that might be relevant based on keywords
			relevantDirs := findRelevantDirectories(cluster, directories, claimedBy)
			if len(relevantDirs) > 0 {
				suggestion := "Consider adding claim to " + relevantDirs[0]
				output.PartiallyClaimed = append(output.PartiallyClaimed, PartiallyClaimed{
					ID:             cluster.ID,
					Name:           cluster.Name,
					ClaimedBy:      claimedBy,
					AlsoRelevantTo: relevantDirs,
					Suggestion:     suggestion,
				})
			}
		}

		if uncoveredRatio < config.Emergence.MinUncoveredRatio {
			output.NotReady = append(output.NotReady, PendingCluster{
				ClusterID:   cluster.ID,
				ClusterName: cluster.Name,
				Reason:      "too_much_coverage",
				BeatCount:   beatCount,
			})
			continue
		}

		// Cluster is ready for crystallization (unclaimed)
		temperature := computeClusterTemperature(clusterBeats, config)
		suggestedPath := SuggestPath(cluster, clusterBeats, directories)
		sampleBeats := getSampleBeats(clusterBeats, 5)

		// Build suggested claims
		suggestedClaims := SuggestedClaims{
			Clusters: []string{cluster.Name},
			Topics:   extractTopics(cluster),
		}

		output.Clusters = append(output.Clusters, EmergentCluster{
			ClusterID:            cluster.ID,
			ClusterName:          cluster.Name,
			BeatCount:            beatCount,
			Ripeness:             cluster.RipenessScore,
			Temperature:          temperature,
			Keywords:             cluster.Keywords,
			CrystallizationReady: true,
			SuggestedPath:        suggestedPath,
			SuggestedClaims:      suggestedClaims,
			SampleBeats:          sampleBeats,
		})
	}

	// Generate claim suggestions for existing directories
	for dir, clusterOverlaps := range dirToClusterOverlap {
		for clusterID, overlapCount := range clusterOverlaps {
			if overlapCount >= 5 { // Significant overlap threshold
				// Find the cluster
				var cluster *model.Cluster
				for i := range clusters {
					if clusters[i].ID == clusterID {
						cluster = &clusters[i]
						break
					}
				}
				if cluster == nil {
					continue
				}

				// Check if this directory already claims this cluster
				alreadyClaimed := false
				for _, claimed := range clusterToDirs[clusterID] {
					if claimed == dir {
						alreadyClaimed = true
						break
					}
				}
				if alreadyClaimed {
					continue
				}

				output.ClaimSuggestions = append(output.ClaimSuggestions, ClaimSuggestion{
					Directory: dir,
					SuggestedClaims: SuggestedClaims{
						Clusters: []string{cluster.Name},
						Topics:   []string{},
					},
					Reason: fmt.Sprintf("%d beats overlap with this cluster", overlapCount),
				})
			}
		}
	}

	return output
}

// findRelevantDirectories finds directories that might be relevant to a cluster based on keywords
func findRelevantDirectories(cluster model.Cluster, directories []WALDDirectory, exclude []string) []string {
	excludeSet := make(map[string]bool)
	for _, e := range exclude {
		excludeSet[e] = true
	}

	var relevant []string
	for _, dir := range directories {
		if excludeSet[dir.Path] {
			continue
		}
		dirText := strings.ToLower(dir.Path + " " + dir.Purpose)
		for _, keyword := range cluster.Keywords {
			if len(keyword) >= 3 && strings.Contains(dirText, strings.ToLower(keyword)) {
				relevant = append(relevant, dir.Path)
				break
			}
		}
	}
	return relevant
}

// extractTopics extracts topic suggestions from cluster keywords
func extractTopics(cluster model.Cluster) []string {
	if len(cluster.Keywords) == 0 {
		return []string{}
	}
	// Return top 2-3 keywords as topics
	topics := cluster.Keywords
	if len(topics) > 3 {
		topics = topics[:3]
	}
	return topics
}

func computeClusterTemperature(beats []Beat, config *Config) float64 {
	if len(beats) == 0 {
		return 0
	}

	now := time.Now()
	totalWeight := 0.0

	for _, beat := range beats {
		ageDays := now.Sub(beat.CreatedAt).Hours() / 24
		weight := math.Exp(-config.Temperature.RecencyDecayLambda * ageDays)
		totalWeight += weight
	}

	maxPossible := float64(len(beats))
	if maxPossible == 0 {
		return 0
	}

	return math.Min(1.0, totalWeight/maxPossible)
}

// SuggestPath finds the most common parent directory and suggests a new path
func SuggestPath(cluster model.Cluster, beats []Beat, directories []WALDDirectory) string {
	directoryCounts := make(map[string]int)

	for _, beat := range beats {
		if beat.Context != nil && beat.Context.WALDDirectory != "" {
			parent := filepath.Dir(beat.Context.WALDDirectory)
			if parent != "." && parent != "/" {
				directoryCounts[parent]++
			}
		}
	}

	slug := slugify(cluster.Name)

	if len(directoryCounts) > 0 {
		var maxParent string
		maxCount := 0
		for parent, count := range directoryCounts {
			if count > maxCount {
				maxCount = count
				maxParent = parent
			}
		}
		return filepath.Join(maxParent, slug)
	}

	return filepath.Join("projects", slug)
}

// SuggestPurpose generates a purpose description based on cluster keywords
func SuggestPurpose(cluster model.Cluster, beats []Beat) string {
	if len(cluster.Keywords) > 0 {
		keywords := cluster.Keywords
		if len(keywords) > 3 {
			keywords = keywords[:3]
		}
		return "Work related to " + strings.Join(keywords, ", ")
	}

	return "Emerging focus area: " + cluster.Name
}

func getSampleBeats(beats []Beat, limit int) []BeatSample {
	sortedBeats := make([]Beat, len(beats))
	copy(sortedBeats, beats)
	sort.Slice(sortedBeats, func(i, j int) bool {
		return sortedBeats[i].CreatedAt.After(sortedBeats[j].CreatedAt)
	})

	samples := []BeatSample{}
	for i, beat := range sortedBeats {
		if i >= limit {
			break
		}
		preview := beat.Content
		if len(preview) > 80 {
			preview = preview[:80] + "..."
		}
		samples = append(samples, BeatSample{
			ID:      beat.ID,
			Preview: preview,
		})
	}

	return samples
}

func slugify(name string) string {
	name = strings.ToLower(name)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	name = reg.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if len(name) > 50 {
		name = name[:50]
		if idx := strings.LastIndex(name, "-"); idx > 30 {
			name = name[:idx]
		}
	}
	return name
}

// SaveEmergenceCache writes the emergence output to .wald/emergence.json
func SaveEmergenceCache(werkRoot string, output *EmergenceOutput) error {
	waldDir := filepath.Join(werkRoot, ".wald")
	if err := os.MkdirAll(waldDir, 0755); err != nil {
		return err
	}

	cachePath := filepath.Join(waldDir, emergenceCacheFileName)
	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cachePath, data, 0644)
}

// LoadEmergenceCache reads the emergence output from .wald/emergence.json
func LoadEmergenceCache(werkRoot string) (*EmergenceOutput, error) {
	cachePath := filepath.Join(werkRoot, ".wald", emergenceCacheFileName)
	data, err := os.ReadFile(cachePath)
	if err != nil {
		return nil, err
	}

	var output EmergenceOutput
	if err := json.Unmarshal(data, &output); err != nil {
		return nil, err
	}

	return &output, nil
}
