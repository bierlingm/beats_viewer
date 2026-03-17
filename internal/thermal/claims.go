package thermal

import (
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// DirectoryClaims represents the claims a directory makes
type DirectoryClaims struct {
	Clusters    []string `yaml:"clusters" json:"clusters"`
	Topics      []string `yaml:"topics" json:"topics"`
	Keywords    []string `yaml:"keywords" json:"keywords"`
	Cooperators []string `yaml:"cooperators" json:"cooperators"`
}

// ClaimMatch represents a single match between a beat and a claim
type ClaimMatch struct {
	Type      string  `json:"type"`       // cluster, topic, keyword, cooperator
	Value     string  `json:"value"`      // the matched claim value
	Relevance float64 `json:"relevance"`  // 0.0-1.0 relevance score
	MatchInfo string  `json:"match_info"` // explanation of match
}

// ClaimMatchResult contains all matches for a beat
type ClaimMatchResult struct {
	Beat       string       `json:"beat_id"`
	Directory  string       `json:"directory"`
	Matches    []ClaimMatch `json:"matches"`
	TotalScore float64      `json:"total_score"`
}

// MatchBeatToClaims matches a beat against a directory's claims
// Returns matches with relevance scores per match type
func MatchBeatToClaims(beat Beat, dir WALDDirectory, clusters []model.Cluster, config *Config) ClaimMatchResult {
	result := ClaimMatchResult{
		Beat:      beat.ID,
		Directory: dir.Path,
		Matches:   []ClaimMatch{},
	}

	// Match cluster membership
	for _, clusterName := range dir.Claims.Clusters {
		for _, cluster := range clusters {
			if strings.EqualFold(cluster.Name, clusterName) || cluster.ID == clusterName {
				for _, beatID := range cluster.BeatIDs {
					if beatID == beat.ID {
						result.Matches = append(result.Matches, ClaimMatch{
							Type:      "cluster",
							Value:     clusterName,
							Relevance: 1.0,
							MatchInfo: "beat is member of claimed cluster",
						})
						break
					}
				}
			}
		}
	}

	// Match topics (case-insensitive substring)
	contentLower := strings.ToLower(beat.Content)
	for _, topic := range dir.Claims.Topics {
		topicLower := strings.ToLower(topic)
		if strings.Contains(contentLower, topicLower) {
			count := strings.Count(contentLower, topicLower)
			idx := strings.Index(contentLower, topicLower)
			relevance := 0.5 + 0.3*float64(count)/10.0
			if idx < 100 {
				relevance += 0.2
			}
			if relevance > 1.0 {
				relevance = 1.0
			}
			result.Matches = append(result.Matches, ClaimMatch{
				Type:      "topic",
				Value:     topic,
				Relevance: relevance,
				MatchInfo: "topic found in content",
			})
		}
	}

	// Match keywords based on config mode
	matchMode := "fuzzy"
	if config != nil && config.Claims.KeywordMatchMode != "" {
		matchMode = config.Claims.KeywordMatchMode
	}

	for _, keyword := range dir.Claims.Keywords {
		keywordLower := strings.ToLower(keyword)
		switch matchMode {
		case "exact":
			if containsWord(contentLower, keywordLower) {
				result.Matches = append(result.Matches, ClaimMatch{
					Type:      "keyword",
					Value:     keyword,
					Relevance: 1.0,
					MatchInfo: "exact word match",
				})
			}
		case "fuzzy", "semantic":
			if strings.Contains(contentLower, keywordLower) {
				result.Matches = append(result.Matches, ClaimMatch{
					Type:      "keyword",
					Value:     keyword,
					Relevance: 0.8,
					MatchInfo: "substring match",
				})
			}
		}
	}

	// Match cooperators
	for _, cooperator := range dir.Claims.Cooperators {
		cooperatorLower := strings.ToLower(cooperator)
		if strings.Contains(contentLower, cooperatorLower) {
			result.Matches = append(result.Matches, ClaimMatch{
				Type:      "cooperator",
				Value:     cooperator,
				Relevance: 1.0,
				MatchInfo: "cooperator mentioned in content",
			})
		}
	}

	// Calculate total score
	for _, m := range result.Matches {
		result.TotalScore += m.Relevance
	}

	return result
}

// GetClaimedClusters returns clusters claimed by a directory
func GetClaimedClusters(dir WALDDirectory, allClusters []model.Cluster) []model.Cluster {
	var claimed []model.Cluster
	for _, clusterName := range dir.Claims.Clusters {
		for _, cluster := range allClusters {
			if strings.EqualFold(cluster.Name, clusterName) || cluster.ID == clusterName {
				claimed = append(claimed, cluster)
				break
			}
		}
	}
	return claimed
}

// containsWord checks if content contains keyword as a whole word
func containsWord(content, word string) bool {
	words := strings.Fields(content)
	for _, w := range words {
		w = strings.Trim(w, ".,!?;:'\"()[]{}")
		if w == word {
			return true
		}
	}
	return false
}
