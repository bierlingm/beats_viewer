package thermal

import (
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"

	"gopkg.in/yaml.v3"
)

// StateChange represents a directory whose state will change.
type StateChange struct {
	Path          string  `json:"path"`
	CurrentState  string  `json:"current_state"`
	InferredState string  `json:"inferred_state"`
	OldState      string  `json:"old_state"`
	NewState      string  `json:"new_state"`
	Temperature   float64 `json:"temperature"`
	Reason        string  `json:"reason"`
	Preserve      bool    `json:"preserve"`
	WillChange    bool    `json:"will_change"`
}

// PreservedDirectory represents a directory that won't change due to preserve flag.
type PreservedDirectory struct {
	Path          string  `json:"path"`
	CurrentState  string  `json:"current_state"`
	InferredState string  `json:"inferred_state"`
	Temperature   float64 `json:"temperature"`
	Preserve      bool    `json:"preserve"`
	WillChange    bool    `json:"will_change"`
	Reason        string  `json:"reason"`
}

// SyncPreviewSummary contains counts for the preview.
type SyncPreviewSummary struct {
	DirectoriesChanging  int `json:"directories_changing"`
	DirectoriesPreserved int `json:"directories_preserved"`
	ClustersReady        int `json:"clusters_ready"`
}

// ClusterSyncInfo represents cluster temperature info for sync output.
type ClusterSyncInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Temperature float64  `json:"temperature"`
	BeatCount   int      `json:"beat_count"`
	ClaimedBy   []string `json:"claimed_by"`
	State       string   `json:"state"`
}

// UnclaimedCluster represents a hot cluster without directory claims.
type UnclaimedCluster struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Temperature   float64 `json:"temperature"`
	BeatCount     int     `json:"beat_count"`
	SuggestedPath string  `json:"suggested_path"`
}

// SyncPreview is the output of --robot-sync-preview.
type SyncPreview struct {
	ComputedAt     time.Time                        `json:"computed_at"`
	Directories    map[string]*DirectoryTemperature `json:"directories"`
	StateChanges   []StateChange                    `json:"state_changes"`
	Preserved      []PreservedDirectory             `json:"preserved"`
	EmergenceReady []EmergentCluster                `json:"emergence_ready"`
	Summary        SyncPreviewSummary               `json:"summary"`

	ClusterTemperatures  []ClusterSyncInfo  `json:"cluster_temperatures"`
	UnclaimedHotClusters []UnclaimedCluster `json:"unclaimed_hot_clusters"`
	ClaimSuggestions     []ClaimSuggestion  `json:"claim_suggestions"`
}

// AppliedChange represents a single change that was applied.
type AppliedChange struct {
	Path   string `json:"path"`
	Change string `json:"change"`
}

// SyncApplyResult is the output of --robot-sync-apply.
type SyncApplyResult struct {
	AppliedAt             time.Time       `json:"applied_at"`
	BackupPath            string          `json:"backup_path"`
	ChangesApplied        []AppliedChange `json:"changes_applied"`
	WALDYamlUpdated       bool            `json:"wald_yaml_updated"`
	TemperatureCacheUpdated bool          `json:"temperature_cache_updated"`
	Errors                []string        `json:"errors"`
}

// EmergenceResult wraps emergence detection output for sync operations.
type EmergenceResult = EmergenceOutput

// GenerateSyncPreview generates a preview of state changes.
func GenerateSyncPreview(tempOutput *TemperatureOutput, wald *WALDFile, emergence *EmergenceResult, config *Config) *SyncPreview {
	if config == nil {
		config = DefaultConfig()
	}

	preview := &SyncPreview{
		ComputedAt:           time.Now(),
		Directories:          tempOutput.Directories,
		StateChanges:         []StateChange{},
		Preserved:            []PreservedDirectory{},
		EmergenceReady:       []EmergentCluster{},
		ClusterTemperatures:  []ClusterSyncInfo{},
		UnclaimedHotClusters: []UnclaimedCluster{},
		ClaimSuggestions:     []ClaimSuggestion{},
	}

	// Build map of current declared states
	declaredStates := make(map[string]string)
	preserveFlags := make(map[string]bool)
	for _, dir := range wald.Directories {
		declaredStates[dir.Path] = dir.State
		preserveFlags[dir.Path] = dir.Preserve
	}

	// Build cluster temperatures (PRIMARY)
	if tempOutput.Clusters != nil {
		for _, ct := range tempOutput.Clusters {
			state := inferClusterState(ct.Temperature, config)
			claimedBy := ct.ClaimedBy
			if claimedBy == nil {
				claimedBy = []string{}
			}
			preview.ClusterTemperatures = append(preview.ClusterTemperatures, ClusterSyncInfo{
				ID:          ct.ID,
				Name:        ct.Name,
				Temperature: ct.Temperature,
				BeatCount:   ct.BeatCount,
				ClaimedBy:   claimedBy,
				State:       state,
			})

			// Track unclaimed hot clusters
			if ct.Unclaimed && ct.Temperature >= config.Thresholds.Hot*0.6 {
				suggestedPath := suggestPathForCluster(ct.Name)
				preview.UnclaimedHotClusters = append(preview.UnclaimedHotClusters, UnclaimedCluster{
					ID:            ct.ID,
					Name:          ct.Name,
					Temperature:   ct.Temperature,
					BeatCount:     ct.BeatCount,
					SuggestedPath: suggestedPath,
				})
			}
		}

		// Sort cluster temperatures by temperature descending
		sort.Slice(preview.ClusterTemperatures, func(i, j int) bool {
			return preview.ClusterTemperatures[i].Temperature > preview.ClusterTemperatures[j].Temperature
		})
		sort.Slice(preview.UnclaimedHotClusters, func(i, j int) bool {
			return preview.UnclaimedHotClusters[i].Temperature > preview.UnclaimedHotClusters[j].Temperature
		})
	}

	// Compare inferred vs declared states for directories
	for path, dirTemp := range tempOutput.Directories {
		declared := declaredStates[path]
		inferred := dirTemp.StateInferred
		preserve := preserveFlags[path]

		// Generate reason based on claimed cluster heat
		reason := generateStateReasonFromClusters(dirTemp, tempOutput.Clusters, config)

		if preserve {
			// Preserved directories don't change
			if declared != inferred {
				preview.Preserved = append(preview.Preserved, PreservedDirectory{
					Path:          path,
					CurrentState:  declared,
					InferredState: inferred,
					Temperature:   dirTemp.Temperature,
					Preserve:      true,
					WillChange:    false,
					Reason:        "preserve flag set",
				})
			}
		} else if declared != inferred {
			// State will change
			preview.StateChanges = append(preview.StateChanges, StateChange{
				Path:          path,
				CurrentState:  declared,
				InferredState: inferred,
				OldState:      declared,
				NewState:      inferred,
				Temperature:   dirTemp.Temperature,
				Reason:        reason,
				Preserve:      false,
				WillChange:    true,
			})
		}
	}

	// Sort state changes by path for consistent output
	sort.Slice(preview.StateChanges, func(i, j int) bool {
		return preview.StateChanges[i].Path < preview.StateChanges[j].Path
	})
	sort.Slice(preview.Preserved, func(i, j int) bool {
		return preview.Preserved[i].Path < preview.Preserved[j].Path
	})

	// Add emergence-ready clusters
	if emergence != nil {
		preview.EmergenceReady = emergence.Clusters
	}

	// Build summary
	preview.Summary = SyncPreviewSummary{
		DirectoriesChanging:  len(preview.StateChanges),
		DirectoriesPreserved: len(preview.Preserved),
		ClustersReady:        len(preview.EmergenceReady),
	}

	return preview
}

// inferClusterState infers state from cluster temperature.
func inferClusterState(temp float64, config *Config) string {
	if temp >= config.Thresholds.Hot {
		return "hot"
	} else if temp >= config.Thresholds.Warm {
		return "warm"
	} else if temp >= config.Thresholds.Cool {
		return "cool"
	}
	return "cold"
}

// suggestPathForCluster suggests a directory path for an unclaimed cluster.
func suggestPathForCluster(clusterName string) string {
	// Simple heuristic: convert name to path format
	name := clusterName
	// Replace common separators
	for _, sep := range []string{" & ", " and ", ", "} {
		if idx := indexString(name, sep); idx >= 0 {
			name = name[:idx]
		}
	}
	// Convert to lowercase, replace spaces with hyphens
	result := ""
	for _, c := range name {
		if c >= 'A' && c <= 'Z' {
			result += string(c + 32) // lowercase
		} else if c == ' ' {
			result += "-"
		} else if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' {
			result += string(c)
		}
	}
	return "projects/" + result
}

// indexString returns index of substring, -1 if not found.
func indexString(s, sub string) int {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// generateStateReasonFromClusters creates reason based on claimed cluster heat.
func generateStateReasonFromClusters(dirTemp *DirectoryTemperature, clusters map[string]*ClusterTemperature, config *Config) string {
	if config == nil {
		config = DefaultConfig()
	}

	if len(dirTemp.ClaimedClusters) == 0 {
		return "no claimed clusters"
	}

	// Find hottest claimed cluster
	maxTemp := 0.0
	hottestName := ""
	for _, clusterID := range dirTemp.ClaimedClusters {
		if ct, ok := clusters[clusterID]; ok {
			if ct.Temperature > maxTemp {
				maxTemp = ct.Temperature
				hottestName = ct.Name
			}
		}
	}

	if hottestName == "" {
		return "derived from " + itoa(len(dirTemp.ClaimedClusters)) + " clusters"
	}

	state := inferClusterState(maxTemp, config)
	return "hottest cluster \"" + hottestName + "\" is " + state
}

// generateStateReason creates a human-readable reason for state inference.
func generateStateReason(dirTemp *DirectoryTemperature, config *Config) string {
	if config == nil {
		config = DefaultConfig()
	}

	windowDays := config.Temperature.WindowDays

	if dirTemp.ClaimedBeatCount == 0 {
		return "no beats in " + string(rune(windowDays)) + " days"
	}

	switch dirTemp.StateInferred {
	case "hot":
		return "high activity with " + itoa(dirTemp.ClaimedBeatCount) + " beats"
	case "warm":
		return "moderate activity with " + itoa(dirTemp.ClaimedBeatCount) + " beats"
	case "cool":
		return "low activity with " + itoa(dirTemp.ClaimedBeatCount) + " beats"
	case "cold":
		return "minimal activity with " + itoa(dirTemp.ClaimedBeatCount) + " beats"
	default:
		return "temperature: " + ftoa(dirTemp.Temperature)
	}
}

// itoa converts int to string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	result := ""
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = string(rune('0'+n%10)) + result
		n /= 10
	}
	if negative {
		result = "-" + result
	}
	return result
}

// ftoa converts float64 to string with 2 decimal places.
func ftoa(f float64) string {
	intPart := int(f)
	fracPart := int((f - float64(intPart)) * 100)
	if fracPart < 0 {
		fracPart = -fracPart
	}
	frac := itoa(fracPart)
	if len(frac) == 1 {
		frac = "0" + frac
	}
	return itoa(intPart) + "." + frac
}

// ApplySyncChanges applies the state changes to WALD.yaml.
func ApplySyncChanges(preview *SyncPreview, werkRoot string, config *Config) (*SyncApplyResult, error) {
	result := &SyncApplyResult{
		AppliedAt:      time.Now(),
		ChangesApplied: []AppliedChange{},
		Errors:         []string{},
	}

	if len(preview.StateChanges) == 0 {
		return result, nil
	}

	if config == nil {
		config = DefaultConfig()
	}

	// Create backup
	if config.Sync.BackupBeforeWrite {
		backupPath, err := createBackup(werkRoot, config)
		if err != nil {
			result.Errors = append(result.Errors, "backup failed: "+err.Error())
			return result, err
		}
		result.BackupPath = backupPath
	}

	// Load current WALD.yaml
	waldPath := filepath.Join(werkRoot, "WALD.yaml")
	data, err := os.ReadFile(waldPath)
	if err != nil {
		result.Errors = append(result.Errors, "failed to read WALD.yaml: "+err.Error())
		return result, err
	}

	var wald WALDFile
	if err := yaml.Unmarshal(data, &wald); err != nil {
		result.Errors = append(result.Errors, "failed to parse WALD.yaml: "+err.Error())
		return result, err
	}

	// Build change map
	changeMap := make(map[string]StateChange)
	for _, change := range preview.StateChanges {
		changeMap[change.Path] = change
	}

	// Apply changes
	for i := range wald.Directories {
		if change, ok := changeMap[wald.Directories[i].Path]; ok {
			oldState := wald.Directories[i].State
			wald.Directories[i].State = change.InferredState
			wald.Directories[i].StateSource = "inferred"

			result.ChangesApplied = append(result.ChangesApplied, AppliedChange{
				Path:   change.Path,
				Change: "state: " + oldState + " → " + change.InferredState,
			})
		}
	}

	// Update generated_at
	wald.GeneratedAt = time.Now().Format(time.RFC3339)

	// Write updated WALD.yaml
	newData, err := yaml.Marshal(&wald)
	if err != nil {
		result.Errors = append(result.Errors, "failed to marshal WALD.yaml: "+err.Error())
		// Restore from backup
		if result.BackupPath != "" {
			restoreBackup(result.BackupPath, waldPath)
		}
		return result, err
	}

	if err := os.WriteFile(waldPath, newData, 0644); err != nil {
		result.Errors = append(result.Errors, "failed to write WALD.yaml: "+err.Error())
		// Restore from backup
		if result.BackupPath != "" {
			restoreBackup(result.BackupPath, waldPath)
		}
		return result, err
	}

	result.WALDYamlUpdated = true

	// Update temperature cache
	if err := SaveTemperatureCache(werkRoot, nil); err == nil {
		result.TemperatureCacheUpdated = true
	}

	return result, nil
}

// ApplySync is an alias for ApplySyncChanges with default config.
func ApplySync(werkRoot string, preview *SyncPreview) *SyncApplyResult {
	result, _ := ApplySyncChanges(preview, werkRoot, nil)
	return result
}

// createBackup creates a backup of WALD.yaml.
func createBackup(werkRoot string, config *Config) (string, error) {
	backupDir := filepath.Join(werkRoot, config.Sync.BackupDir)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02T15:04:05Z")
	backupPath := filepath.Join(backupDir, "WALD.yaml."+timestamp)

	srcPath := filepath.Join(werkRoot, "WALD.yaml")
	if err := copyFile(srcPath, backupPath); err != nil {
		return "", err
	}

	// Cleanup old backups
	cleanupOldBackups(backupDir, config.Sync.MaxBackups)

	return backupPath, nil
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// restoreBackup restores WALD.yaml from backup.
func restoreBackup(backupPath, waldPath string) error {
	return copyFile(backupPath, waldPath)
}

// cleanupOldBackups removes old backups keeping only maxBackups.
func cleanupOldBackups(backupDir string, maxBackups int) {
	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return
	}

	// Filter WALD.yaml backups
	var backups []string
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 10 && entry.Name()[:10] == "WALD.yaml." {
			backups = append(backups, filepath.Join(backupDir, entry.Name()))
		}
	}

	// Sort by name (timestamp) descending
	sort.Sort(sort.Reverse(sort.StringSlice(backups)))

	// Remove old backups
	for i := maxBackups; i < len(backups); i++ {
		os.Remove(backups[i])
	}
}
