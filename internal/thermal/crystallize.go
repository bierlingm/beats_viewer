package thermal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// CrystallizeInput contains parameters for crystallization
type CrystallizeInput struct {
	ClusterID       string `json:"cluster_id"`
	PathOverride    string `json:"path_override,omitempty"`
	PurposeOverride string `json:"purpose_override,omitempty"`
	Confirm         bool   `json:"confirm"`
}

// CrystallizeResult contains the outcome of crystallization
type CrystallizeResult struct {
	Created struct {
		Directory string `json:"directory"`
		AgentsMd  string `json:"agents_md"`
	} `json:"created"`
	ClaimsAdded struct {
		Clusters []string `json:"clusters"`
		Topics   []string `json:"topics"`
	} `json:"claims_added"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// CreatedInfo describes what was created during crystallization
type CreatedInfo struct {
	Directory      string `json:"directory"`
	AgentsMd       string `json:"agents_md"`
	WALDEntryAdded bool   `json:"wald_entry_added"`
}

// NewDirectoryInfo describes the new directory entry
type NewDirectoryInfo struct {
	Path          string  `json:"path"`
	Purpose       string  `json:"purpose"`
	State         string  `json:"state"`
	StateSource   string  `json:"state_source"`
	OriginCluster string  `json:"origin_cluster"`
	Temperature   float64 `json:"temperature"`
}

// PreviewCrystallize shows what would be created without doing it
func PreviewCrystallize(werkRoot string, cluster *EmergentCluster, beats []Beat) (*CrystallizeResult, error) {
	if cluster == nil {
		return nil, errors.New("cluster is nil")
	}

	if !cluster.CrystallizationReady {
		return nil, errors.New("cluster is not ready for crystallization")
	}

	path := cluster.SuggestedPath

	// Extract topics from sample beats
	topics := extractTopicsFromEmergent(cluster)

	result := &CrystallizeResult{
		Message: "Directory created with claim. Beats remain in global store.",
	}
	result.Created.Directory = filepath.Join(werkRoot, path)
	result.Created.AgentsMd = filepath.Join(werkRoot, path, "AGENTS.md")
	result.ClaimsAdded.Clusters = []string{cluster.ClusterName}
	result.ClaimsAdded.Topics = topics

	return result, nil
}

// extractTopicsFromEmergent extracts topic keywords from emergent cluster
func extractTopicsFromEmergent(cluster *EmergentCluster) []string {
	var topics []string
	seen := make(map[string]bool)

	// Extract from sample beats previews
	for _, sample := range cluster.SampleBeats {
		words := strings.Fields(strings.ToLower(sample.Preview))
		for _, word := range words {
			word = strings.Trim(word, ".,!?:;\"'()[]")
			if len(word) > 3 && !seen[word] {
				seen[word] = true
				topics = append(topics, word)
			}
			if len(topics) >= 5 {
				break
			}
		}
		if len(topics) >= 5 {
			break
		}
	}

	return topics
}

// Crystallize performs the actual crystallization
func Crystallize(werkRoot string, input CrystallizeInput, cluster *EmergentCluster, beats []Beat) (*CrystallizeResult, error) {
	if !input.Confirm {
		return nil, errors.New("crystallization requires confirm=true")
	}

	if cluster == nil {
		return nil, errors.New("cluster is nil")
	}

	if !cluster.CrystallizationReady {
		return nil, errors.New("cluster is not ready for crystallization")
	}

	path := cluster.SuggestedPath
	if input.PathOverride != "" {
		path = input.PathOverride
	}

	purpose := cluster.SuggestedPurpose
	if input.PurposeOverride != "" {
		purpose = input.PurposeOverride
	}

	// Extract topics
	topics := extractTopicsFromEmergent(cluster)

	result := &CrystallizeResult{
		Message: "Directory created with claim. Beats remain in global store.",
	}
	result.ClaimsAdded.Clusters = []string{cluster.ClusterName}
	result.ClaimsAdded.Topics = topics

	// Create directory
	fullPath := filepath.Join(werkRoot, path)
	if err := CreateDirectory(werkRoot, path); err != nil {
		result.Error = "failed to create directory: " + err.Error()
		return result, err
	}
	result.Created.Directory = fullPath

	// Generate and write AGENTS.md
	agentsMd := GenerateAgentsMd(cluster, beats, path, purpose)
	agentsMdPath := filepath.Join(fullPath, "AGENTS.md")
	if err := os.WriteFile(agentsMdPath, []byte(agentsMd), 0644); err != nil {
		result.Error = "failed to write AGENTS.md: " + err.Error()
		return result, err
	}
	result.Created.AgentsMd = agentsMdPath

	// Update WALD.yaml with claims
	newDir := NewDirectoryInfo{
		Path:          path,
		Purpose:       purpose,
		State:         "active",
		StateSource:   "crystallized",
		OriginCluster: cluster.ClusterID,
		Temperature:   cluster.Temperature,
	}
	if err := UpdateWALDWithClaims(werkRoot, newDir, cluster.ClusterName, topics); err != nil {
		result.Error = "failed to update WALD.yaml: " + err.Error()
		return result, err
	}

	// NOTE: Do NOT update beat contexts - beats stay in global store

	return result, nil
}

// CreateDirectory creates a directory at the specified path
func CreateDirectory(werkRoot, path string) error {
	fullPath := filepath.Join(werkRoot, path)
	return os.MkdirAll(fullPath, 0755)
}

// GenerateAgentsMd creates the AGENTS.md content for a crystallized directory
func GenerateAgentsMd(cluster *EmergentCluster, beats []Beat, path, purpose string) string {
	var buf strings.Builder

	dirName := filepath.Base(path)

	buf.WriteString(fmt.Sprintf("# %s\n\n", dirName))
	buf.WriteString(fmt.Sprintf("%s\n\n", purpose))

	buf.WriteString("## Origin\n\n")
	buf.WriteString(fmt.Sprintf("Crystallized from beat cluster `%s` on %s.\n\n",
		cluster.ClusterName, time.Now().Format("2006-01-02")))
	buf.WriteString("This directory emerged from patterns in articulable noticing — repeated attention\n")
	buf.WriteString("to these themes coalesced into structure.\n\n")

	buf.WriteString("## Founding Beats\n\n")

	sampleLimit := 5
	if len(cluster.SampleBeats) < sampleLimit {
		sampleLimit = len(cluster.SampleBeats)
	}

	for i := 0; i < sampleLimit; i++ {
		sample := cluster.SampleBeats[i]
		buf.WriteString(fmt.Sprintf("- %s: %s\n", sample.ID, sample.Preview))
	}

	remaining := len(beats) - sampleLimit
	if remaining > 0 {
		buf.WriteString(fmt.Sprintf("\n[... %d more beats ...]\n", remaining))
	}

	buf.WriteString("\n## Context\n\n")

	parentDir := filepath.Dir(path)
	if parentDir == "." {
		parentDir = "(root)"
	}
	buf.WriteString(fmt.Sprintf("Parent gravity: %s\n", parentDir))
	buf.WriteString(fmt.Sprintf("Origin cluster: %s\n", cluster.ClusterID))
	buf.WriteString(fmt.Sprintf("Initial temperature: %.2f\n", cluster.Temperature))

	return buf.String()
}

// UpdateWALDWithNewDirectory adds a new directory entry to WALD.yaml
func UpdateWALDWithNewDirectory(werkRoot string, newDir NewDirectoryInfo) error {
	return UpdateWALDWithClaims(werkRoot, newDir, "", nil)
}

// UpdateWALDWithClaims adds a new directory entry to WALD.yaml with claims
func UpdateWALDWithClaims(werkRoot string, newDir NewDirectoryInfo, clusterName string, topics []string) error {
	waldPath := filepath.Join(werkRoot, "WALD.yaml")

	data, err := os.ReadFile(waldPath)
	if err != nil {
		return fmt.Errorf("failed to read WALD.yaml: %w", err)
	}

	var wald WALDFile
	if err := yaml.Unmarshal(data, &wald); err != nil {
		return fmt.Errorf("failed to parse WALD.yaml: %w", err)
	}

	newEntry := WALDDirectory{
		Path:          newDir.Path,
		Purpose:       newDir.Purpose,
		State:         newDir.State,
		StateSource:   newDir.StateSource,
		OriginCluster: newDir.OriginCluster,
	}

	// Add claims if provided
	if clusterName != "" || len(topics) > 0 {
		newEntry.Claims = DirectoryClaims{
			Clusters: []string{clusterName},
			Topics:   topics,
		}
	}

	wald.Directories = append(wald.Directories, newEntry)
	wald.GeneratedAt = time.Now().Format(time.RFC3339)

	newData, err := yaml.Marshal(&wald)
	if err != nil {
		return fmt.Errorf("failed to marshal WALD.yaml: %w", err)
	}

	if err := os.WriteFile(waldPath, newData, 0644); err != nil {
		return fmt.Errorf("failed to write WALD.yaml: %w", err)
	}

	return nil
}

// UpdateBeatContexts updates the wald_directory in beat contexts
func UpdateBeatContexts(werkRoot string, beatIDs []string, newPath string) (int, error) {
	updated := 0

	beatIDSet := make(map[string]bool)
	for _, id := range beatIDs {
		beatIDSet[id] = true
	}

	// Find all .beats directories
	var beatsDirs []string
	err := filepath.Walk(werkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".beats" {
			beatsDirs = append(beatsDirs, path)
		}
		return nil
	})
	if err != nil {
		return updated, err
	}

	// Process each .beats directory
	for _, beatsDir := range beatsDirs {
		jsonlPath := filepath.Join(beatsDir, "beats.jsonl")
		if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
			continue
		}

		data, err := os.ReadFile(jsonlPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		modified := false
		var newLines []string

		for _, line := range lines {
			if line == "" {
				newLines = append(newLines, line)
				continue
			}

			// Check if this line contains a beat we need to update
			needsUpdate := false
			for id := range beatIDSet {
				if strings.Contains(line, `"id":"`+id+`"`) {
					needsUpdate = true
					break
				}
			}

			if needsUpdate {
				// Update wald_directory in context
				if strings.Contains(line, `"wald_directory":`) {
					// Replace existing wald_directory
					line = replaceWALDDirectory(line, newPath)
				} else if strings.Contains(line, `"context":{`) {
					// Add wald_directory to existing context
					line = strings.Replace(line, `"context":{`, `"context":{"wald_directory":"`+newPath+`",`, 1)
				} else if strings.Contains(line, `"context":null`) || !strings.Contains(line, `"context":`) {
					// Add context with wald_directory
					line = strings.Replace(line, `}`, `,"context":{"wald_directory":"`+newPath+`"}}`, 1)
				}
				modified = true
				updated++
			}

			newLines = append(newLines, line)
		}

		if modified {
			os.WriteFile(jsonlPath, []byte(strings.Join(newLines, "\n")), 0644)
		}
	}

	return updated, nil
}

func replaceWALDDirectory(line, newPath string) string {
	// Simple replacement - find "wald_directory":"..." and replace the value
	start := strings.Index(line, `"wald_directory":"`)
	if start == -1 {
		return line
	}
	start += len(`"wald_directory":"`)
	end := strings.Index(line[start:], `"`)
	if end == -1 {
		return line
	}
	return line[:start] + newPath + line[start+end:]
}

// FindClusterByID finds a cluster by ID in emergence output
func FindClusterByID(emergence *EmergenceOutput, clusterID string) *EmergentCluster {
	if emergence == nil {
		return nil
	}
	for i := range emergence.Clusters {
		if emergence.Clusters[i].ClusterID == clusterID {
			return &emergence.Clusters[i]
		}
	}
	return nil
}
