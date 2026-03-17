package thermal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// LegacyBeat represents a beat with _legacy_context field
type LegacyBeat struct {
	ID            string          `json:"id"`
	Content       string          `json:"content"`
	LegacyContext *LegacyContext  `json:"_legacy_context,omitempty"`
}

// LegacyContext represents the migration context
type LegacyContext struct {
	WALDDirectory string `json:"wald_directory"`
	MigratedFrom  string `json:"migrated_from"`
}

// DirectorySuggestion contains suggestions for a directory based on legacy context
type DirectorySuggestion struct {
	Path          string              `json:"path"`
	BeatCount     int                 `json:"beat_count"`
	Clusters      []ClusterSuggestion `json:"clusters"`
	Topics        []string            `json:"topics"`
	Cooperators   []CooperatorMention `json:"cooperators"`
}

// ClusterSuggestion is a cluster to claim
type ClusterSuggestion struct {
	Name      string `json:"name"`
	BeatCount int    `json:"beat_count"`
}

// CooperatorMention tracks cooperator mentions in beats
type CooperatorMention struct {
	Slug      string `json:"slug"`
	BeatCount int    `json:"beat_count"`
}

// SuggestClaimsOutput is the full output of suggest-claims
type SuggestClaimsOutput struct {
	Suggestions []DirectorySuggestion `json:"suggestions"`
	TotalBeats  int                   `json:"total_beats"`
}

// SuggestClaimsFromLegacy analyzes beats with _legacy_context and suggests claims
func SuggestClaimsFromLegacy(werkRoot string, clusters []model.Cluster) (*SuggestClaimsOutput, error) {
	// Load beats from global store
	globalStore := filepath.Join(werkRoot, ".beats", "beats.jsonl")
	beats, err := loadLegacyBeats(globalStore)
	if err != nil {
		return nil, fmt.Errorf("failed to load beats: %w", err)
	}

	// Load WALD for cooperators
	wald, err := LoadWALD(werkRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load WALD.yaml: %w", err)
	}

	// Extract known cooperator slugs
	cooperatorSlugs := extractCooperatorSlugs(wald)

	// Group beats by legacy directory
	beatsByDir := make(map[string][]LegacyBeat)
	for _, beat := range beats {
		if beat.LegacyContext != nil && beat.LegacyContext.WALDDirectory != "" {
			dir := beat.LegacyContext.WALDDirectory
			beatsByDir[dir] = append(beatsByDir[dir], beat)
		}
	}

	// Build beat ID to cluster mapping
	beatToCluster := make(map[string][]model.Cluster)
	for _, cluster := range clusters {
		for _, beatID := range cluster.BeatIDs {
			beatToCluster[beatID] = append(beatToCluster[beatID], cluster)
		}
	}

	// Generate suggestions for each directory
	var suggestions []DirectorySuggestion
	for dir, dirBeats := range beatsByDir {
		suggestion := DirectorySuggestion{
			Path:      dir,
			BeatCount: len(dirBeats),
		}

		// Find clusters these beats belong to
		clusterCounts := make(map[string]int)
		clusterNames := make(map[string]string)
		for _, beat := range dirBeats {
			for _, cluster := range beatToCluster[beat.ID] {
				clusterCounts[cluster.ID]++
				clusterNames[cluster.ID] = cluster.Name
			}
		}

		// Sort clusters by beat count
		type clusterCount struct {
			id    string
			name  string
			count int
		}
		var sortedClusters []clusterCount
		for id, count := range clusterCounts {
			sortedClusters = append(sortedClusters, clusterCount{id, clusterNames[id], count})
		}
		sort.Slice(sortedClusters, func(i, j int) bool {
			return sortedClusters[i].count > sortedClusters[j].count
		})

		for _, cc := range sortedClusters {
			suggestion.Clusters = append(suggestion.Clusters, ClusterSuggestion{
				Name:      cc.name,
				BeatCount: cc.count,
			})
		}

		// Extract common topics from beat content
		suggestion.Topics = extractCommonTopics(dirBeats)

		// Find cooperator mentions
		coopCounts := make(map[string]int)
		for _, beat := range dirBeats {
			contentLower := strings.ToLower(beat.Content)
			for slug := range cooperatorSlugs {
				if strings.Contains(contentLower, strings.ToLower(slug)) ||
					strings.Contains(contentLower, strings.ToLower(slugToDisplayName(slug))) {
					coopCounts[slug]++
				}
			}
		}

		for slug, count := range coopCounts {
			suggestion.Cooperators = append(suggestion.Cooperators, CooperatorMention{
				Slug:      slug,
				BeatCount: count,
			})
		}
		sort.Slice(suggestion.Cooperators, func(i, j int) bool {
			return suggestion.Cooperators[i].BeatCount > suggestion.Cooperators[j].BeatCount
		})

		suggestions = append(suggestions, suggestion)
	}

	// Sort suggestions by beat count
	sort.Slice(suggestions, func(i, j int) bool {
		return suggestions[i].BeatCount > suggestions[j].BeatCount
	})

	return &SuggestClaimsOutput{
		Suggestions: suggestions,
		TotalBeats:  len(beats),
	}, nil
}

func loadLegacyBeats(path string) ([]LegacyBeat, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var beats []LegacyBeat
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var beat LegacyBeat
		if err := json.Unmarshal([]byte(line), &beat); err != nil {
			continue // Skip malformed lines
		}
		beats = append(beats, beat)
	}
	return beats, scanner.Err()
}

func extractCooperatorSlugs(wald *WALDFile) map[string]bool {
	slugs := make(map[string]bool)
	for _, dir := range wald.Directories {
		if strings.HasPrefix(dir.Path, "cooperators/") {
			slug := strings.TrimPrefix(dir.Path, "cooperators/")
			slugs[slug] = true
		}
	}
	return slugs
}

func slugToDisplayName(slug string) string {
	parts := strings.Split(slug, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func extractCommonTopics(beats []LegacyBeat) []string {
	wordCounts := make(map[string]int)
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"to": true, "of": true, "in": true, "for": true, "on": true,
		"with": true, "at": true, "by": true, "from": true, "as": true,
		"this": true, "that": true, "it": true, "its": true, "i": true,
		"we": true, "you": true, "they": true, "he": true, "she": true,
		"about": true, "will": true, "can": true, "would": true, "could": true,
		"has": true, "have": true, "had": true, "not": true, "but": true,
	}

	for _, beat := range beats {
		words := strings.Fields(strings.ToLower(beat.Content))
		seen := make(map[string]bool)
		for _, word := range words {
			word = strings.Trim(word, ".,!?;:'\"()[]{}")
			if len(word) < 4 || stopWords[word] || seen[word] {
				continue
			}
			seen[word] = true
			wordCounts[word]++
		}
	}

	// Find words appearing in at least 20% of beats
	threshold := len(beats) / 5
	if threshold < 2 {
		threshold = 2
	}

	type wordCount struct {
		word  string
		count int
	}
	var sorted []wordCount
	for word, count := range wordCounts {
		if count >= threshold {
			sorted = append(sorted, wordCount{word, count})
		}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	var topics []string
	for i, wc := range sorted {
		if i >= 5 {
			break
		}
		topics = append(topics, wc.word)
	}
	return topics
}

// ApplySuggestedClaims applies the suggestions to WALD.yaml
func ApplySuggestedClaims(werkRoot string, suggestions []DirectorySuggestion) error {
	wald, err := LoadWALD(werkRoot)
	if err != nil {
		return err
	}

	// Build suggestion map
	suggestionMap := make(map[string]DirectorySuggestion)
	for _, s := range suggestions {
		suggestionMap[s.Path] = s
	}

	// Update directory claims
	for i := range wald.Directories {
		dir := &wald.Directories[i]
		suggestion, ok := suggestionMap[dir.Path]
		if !ok {
			continue
		}

		// Add cluster claims
		existingClusters := make(map[string]bool)
		for _, c := range dir.Claims.Clusters {
			existingClusters[c] = true
		}
		for _, cluster := range suggestion.Clusters {
			if !existingClusters[cluster.Name] {
				dir.Claims.Clusters = append(dir.Claims.Clusters, cluster.Name)
			}
		}

		// Add topic claims
		existingTopics := make(map[string]bool)
		for _, t := range dir.Claims.Topics {
			existingTopics[t] = true
		}
		for _, topic := range suggestion.Topics {
			if !existingTopics[topic] {
				dir.Claims.Topics = append(dir.Claims.Topics, topic)
			}
		}

		// Add cooperator claims
		existingCoops := make(map[string]bool)
		for _, c := range dir.Claims.Cooperators {
			existingCoops[c] = true
		}
		for _, coop := range suggestion.Cooperators {
			if !existingCoops[coop.Slug] {
				dir.Claims.Cooperators = append(dir.Claims.Cooperators, coop.Slug)
			}
		}
	}

	return SaveWALD(werkRoot, wald)
}

// PrintSuggestClaimsHuman prints suggestions in human-readable format
func PrintSuggestClaimsHuman(output *SuggestClaimsOutput) {
	fmt.Println("Suggested claims based on legacy beat contexts:")
	fmt.Println()

	for _, s := range output.Suggestions {
		if len(s.Clusters) == 0 && len(s.Topics) == 0 {
			continue
		}
		fmt.Printf("%s (%d beats captured here):\n", s.Path, s.BeatCount)

		if len(s.Clusters) > 0 {
			fmt.Println("  clusters:")
			for _, c := range s.Clusters {
				fmt.Printf("    - \"%s\" (%d beats)\n", c.Name, c.BeatCount)
			}
		}

		if len(s.Topics) > 0 {
			fmt.Printf("  topics: %v\n", s.Topics)
		}

		if len(s.Cooperators) > 0 {
			var coopStrs []string
			for _, c := range s.Cooperators {
				coopStrs = append(coopStrs, fmt.Sprintf("\"%s\" (%d)", c.Slug, c.BeatCount))
			}
			fmt.Printf("  cooperators: %s\n", strings.Join(coopStrs, ", "))
		}

		fmt.Println()
	}
}
