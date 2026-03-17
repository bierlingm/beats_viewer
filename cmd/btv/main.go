package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bierlingm/beats_viewer/internal/thermal"
	"github.com/fsnotify/fsnotify"
	"github.com/bierlingm/beats_viewer/pkg/attention"
	"github.com/bierlingm/beats_viewer/pkg/cluster"
	"github.com/bierlingm/beats_viewer/pkg/crystallize"
	"github.com/bierlingm/beats_viewer/pkg/divergence"
	"github.com/bierlingm/beats_viewer/pkg/loader"
	"github.com/bierlingm/beats_viewer/pkg/model"
	"github.com/bierlingm/beats_viewer/pkg/ripeness"
	"github.com/bierlingm/beats_viewer/pkg/timeline"
	"github.com/bierlingm/beats_viewer/pkg/ui"
	"github.com/bierlingm/beats_viewer/pkg/ui/views"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.4.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("btv %s\n", version)
			return
		case "-h", "--help":
			printHelp()
			return
		case "--robot-help":
			robotHelp()
			return
		case "--robot-list":
			robotList()
			return
		case "--robot-search":
			robotSearch()
			return
		case "--robot-show":
			if len(os.Args) < 3 {
				fatal("--robot-show requires a beat ID")
			}
			robotShow(os.Args[2])
			return
		case "--robot-taxonomy-stats":
			robotTaxonomyStats()
			return
		case "--robot-ripeness":
			if len(os.Args) < 3 {
				fatal("--robot-ripeness requires a beat ID")
			}
			robotRipeness(os.Args[2])
			return
		case "--robot-ripe":
			robotRipe()
			return
		case "--robot-entities":
			robotEntities()
			return
		case "--robot-entity-beats":
			if len(os.Args) < 3 {
				fatal("--robot-entity-beats requires entity name")
			}
			robotEntityBeats(os.Args[2])
			return
		case "--robot-timeline":
			robotTimeline()
			return
		case "--robot-gaps":
			robotGaps()
			return
		case "--robot-cluster":
			robotCluster()
			return
		case "--robot-clusters":
			robotClustersWithTemperature()
			return
		case "--robot-similar":
			if len(os.Args) < 3 {
				fatal("--robot-similar requires a beat ID")
			}
			robotSimilar(os.Args[2])
			return
		case "--robot-chains":
			robotChains()
			return
		case "--robot-create-chain":
			robotCreateChain()
			return
		case "--robot-chain-add":
			robotChainAdd()
			return
		case "--robot-stale":
			robotStale()
			return
		case "--robot-crystallize":
			robotCrystallizeCluster()
			return
		case "--robot-crystallize-suggestions":
			robotCrystallizeSuggestions()
			return
		case "--rebuild-cache":
			rebuildCache()
			return
		case "--capture":
			runCapture()
			return
		case "--robot-attention":
			robotAttention()
			return
		case "--robot-activating":
			robotActivating()
			return
		case "--robot-drift":
			robotDrift()
			return
		case "--robot-orientation":
			robotOrientation()
			return
		case "--robot-heartbeat":
			robotHeartbeat()
			return
		case "--robot-crystallizations":
			robotCrystallizations()
			return
		case "--robot-crystallization":
			if len(os.Args) < 3 {
				fatal("--robot-crystallization requires a bead ID")
			}
			robotCrystallization(os.Args[2])
			return
		case "--robot-infer":
			robotInfer()
			return
		case "--robot-divergence":
			robotDivergence()
			return
		case "--robot-blindspots":
			robotBlindspots()
			return
		case "--robot-agent-beats":
			robotAgentBeats()
			return
		case "--robot-alerts":
			robotAlerts()
			return
		case "--robot-dismiss":
			if len(os.Args) < 3 {
				fatal("--robot-dismiss requires an alert ID")
			}
			robotDismissAlert(os.Args[2])
			return
		case "--robot-temperature":
			robotTemperature()
			return
		case "--robot-watch-temperature":
			robotWatchTemperature()
			return
		case "--robot-wald":
			robotWALD()
			return
		case "--robot-temporal":
			robotTemporal()
			return
		case "--robot-cooperators":
			robotCooperators()
			return
		case "--robot-emergence":
			robotEmergence()
			return
		case "crystallize":
			runCrystallizeCmd()
			return
		case "claim":
			runClaim()
			return
		case "unclaim":
			runUnclaim()
			return
		case "--robot-claim":
			robotClaim()
			return
		case "--robot-unclaim":
			robotUnclaim()
			return
		case "suggest-claims":
			runSuggestClaims()
			return
		case "--robot-suggest-claims":
			robotSuggestClaims()
			return
		case "--robot-sync-preview":
			robotSyncPreview()
			return
		case "--robot-sync-apply":
			robotSyncApply()
			return
		case "sync":
			dryRun := false
			force := false
			for _, arg := range os.Args[2:] {
				if arg == "--dry-run" {
					dryRun = true
				}
				if arg == "--force" {
					force = true
				}
			}
			runSync(dryRun, force)
			return
		}
	}

	rootPath := loader.GetDefaultRoot()
	if envRoot := os.Getenv("BEATS_ROOT"); envRoot != "" {
		rootPath = envRoot
	}

	if len(os.Args) > 1 && os.Args[1] == "--root" && len(os.Args) > 2 {
		rootPath = os.Args[2]
	}

	fmt.Fprintf(os.Stderr, "Loading beats...\n")
	
	m := ui.NewModelV2(rootPath)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`btv - beats_viewer v0.3 - The Attention Engine

USAGE:  btv [options]

OPTIONS:
  --root <path>       Root directory (default: current)
  --rebuild-cache     Force rebuild cache
  -v, --version       Show version
  -h, --help          Show this help

VIEWS & KEYBINDINGS:
  A    Attention Dashboard (default) - what's activating now
  L    List view - all beats
  D    Drift view - attention shifts over time
  H    Heartbeat view - capture rhythm visualization
  C    Cluster view - theme groupings
  j/k  Navigate    /  Search    ?  Help    q  Quit

CONCEPTS:
  Activation     Burst of related captures (3+ in 72h)
  Drift          Topics rising/fading over time
  Crystallization  Inferred beat->bead connections
  Divergence     Human vs agent attention gaps
  Dormancy       Ripe clusters without recent activity

ROBOT COMMANDS (JSON for agents):
  Attention:
    --robot-attention           Full attention state
    --robot-activating          Current activations
    --robot-drift --days N      Drift report (default 30)
    --robot-orientation         Where attention points
    --robot-heartbeat --days N  Rhythm data (default 90)
  Inference:
    --robot-crystallizations    All inferred connections
    --robot-crystallization ID  Beats for specific bead
    --robot-infer               Force inference run
  Divergence:
    --robot-divergence          Human vs agent comparison
    --robot-blindspots          Agent-only topics
    --robot-agent-beats         Agent-captured beats
  Alerts:
    --robot-alerts [--unseen]   Current alerts
    --robot-dismiss <id>        Mark alert seen
  Core (v0.2):
    --robot-list/show/search    Beat operations
    --robot-ripe/stale          Ripeness queries
    --robot-clusters/entities   Analysis data

Config: .beats/btv-config.yaml (thresholds, windows, patterns)
`)
}

func robotHelp() {
	resp := model.RobotHelpResponse{
		Version: version,
		Commands: []model.RobotHelpCommand{
			{Name: "--robot-list", Description: "List beats with filters", Input: "--channel/--source/--sort/--limit flags", Output: "beats array"},
			{Name: "--robot-search", Description: "Search by content/impetus", Input: `{"query": "...", "max_results": N}`, Output: "results array"},
			{Name: "--robot-show", Description: "Get beat details", Input: "beat ID", Output: "beat object"},
			{Name: "--robot-taxonomy-stats", Description: "Channel/source distribution", Output: "channels/sources counts"},
			{Name: "--robot-ripeness", Description: "Get ripeness score+factors", Input: "beat ID", Output: "score breakdown"},
			{Name: "--robot-ripe", Description: "List ripest beats", Input: "--limit/--threshold flags", Output: "beats sorted by ripeness"},
			{Name: "--robot-entities", Description: "List all entities", Output: "people/tools/concepts arrays"},
			{Name: "--robot-entity-beats", Description: "Beats containing entity", Input: "entity name", Output: "beats array"},
			{Name: "--robot-timeline", Description: "Timeline bucket data", Input: "--zoom/--start/--end flags", Output: "buckets array"},
			{Name: "--robot-gaps", Description: "Activity gaps", Input: "--threshold flag", Output: "gaps array"},
			{Name: "--robot-cluster", Description: "Generate/refresh clusters", Input: "--k flag", Output: "clusters array"},
			{Name: "--robot-clusters", Description: "List current clusters", Output: "clusters array"},
			{Name: "--robot-similar", Description: "Find similar beats", Input: "beat ID, --limit flag", Output: "similar beats array"},
			{Name: "--robot-chains", Description: "List chains", Output: "chains array"},
			{Name: "--robot-create-chain", Description: "Create chain", Input: `{"name": "...", "beat_ids": [...]}`, Output: "chain object"},
			{Name: "--robot-chain-add", Description: "Add beat to chain", Input: `{"chain_id": "...", "beat_id": "..."}`, Output: "success"},
			{Name: "--robot-stale", Description: "List stale beats with reasons", Output: "stale beats with reasons and suggested actions"},
			{Name: "--rebuild-cache", Description: "Force rebuild cache", Input: "--project flag", Output: "cache stats"},
		},
	}
	outputJSON(resp)
}

func robotList() {
	rootPath := loader.GetDefaultRoot()

	var projectFilter *string
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			p := os.Args[i+1]
			projectFilter = &p
		}
		if arg == "--root" && i+1 < len(os.Args) {
			rootPath = os.Args[i+1]
		}
	}

	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	if projectFilter != nil {
		var filtered []model.Beat
		for _, b := range beats {
			if beatToProject[b.ID] == *projectFilter {
				filtered = append(filtered, b)
			}
		}
		beats = filtered
	}

	items := make([]model.BeatListItem, len(beats))
	for i, b := range beats {
		items[i] = b.ToListItem(beatToProject[b.ID], 80)
	}

	resp := model.RobotListResponse{
		Beats:         items,
		Total:         len(items),
		ProjectFilter: projectFilter,
	}
	outputJSON(resp)
}

func robotSearch() {
	var input struct {
		Query       string `json:"query"`
		AllProjects bool   `json:"all_projects"`
		MaxResults  int    `json:"max_results"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	if input.MaxResults == 0 {
		input.MaxResults = 50
	}

	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	results := loader.SearchBeats(beats, input.Query)

	if len(results) > input.MaxResults {
		results = results[:input.MaxResults]
	}

	items := make([]model.BeatListItem, len(results))
	for i, b := range results {
		items[i] = b.ToListItem(beatToProject[b.ID], 80)
	}

	resp := model.RobotSearchResponse{
		Results:      items,
		Query:        input.Query,
		TotalMatches: len(items),
	}
	outputJSON(resp)
}

func robotShow(beatID string) {
	rootPath := loader.GetDefaultRoot()
	beats, _, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	beat := loader.FindBeatByID(beats, beatID)
	if beat == nil {
		fatalJSON("error", "beat not found: "+beatID)
	}

	outputJSON(beat)
}

func outputJSON(v interface{}) {
	// Marshal to map first for session context injection
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON encoding error: %v\n", err)
		os.Exit(1)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		// If not a map, output as-is
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.Encode(v)
		return
	}

	// Add session context if FACTORY_SESSION_ID is set
	if sessionID := os.Getenv("FACTORY_SESSION_ID"); sessionID != "" {
		workspace := os.Getenv("FACTORY_WORKSPACE")
		if workspace == "" {
			workspace, _ = os.Getwd()
		}
		result := map[string]interface{}{
			"session_id":        sessionID,
			"session_workspace": workspace,
		}
		for k, val := range m {
			result[k] = val
		}
		m = result
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		fmt.Fprintf(os.Stderr, "JSON encoding error: %v\n", err)
		os.Exit(1)
	}
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	os.Exit(1)
}

func fatalJSON(key, msg string) {
	fmt.Fprintf(os.Stdout, `{"%s": "%s"}`+"\n", key, msg)
	os.Exit(1)
}

func getEnrichedBeats() ([]model.EnrichedBeat, *model.Cache, error) {
	rootPath := loader.GetDefaultRoot()
	projects, err := loader.DiscoverProjects(rootPath)
	if err != nil || len(projects) == 0 {
		return nil, nil, fmt.Errorf("no projects found")
	}
	return loader.LoadEnrichedBeats(projects[0].Path, nil)
}

func robotTaxonomyStats() {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	channels := make(map[string]int)
	sources := make(map[string]int)

	for _, eb := range enriched {
		tax := cache.Taxonomies[eb.ID]
		channels[tax.Channel.String()]++
		sources[tax.Source.String()]++
	}

	resp := map[string]interface{}{
		"channels": channels,
		"sources":  sources,
		"total":    len(enriched),
	}
	outputJSON(resp)
}

func robotRipeness(beatID string) {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var beats []model.Beat
	for _, eb := range enriched {
		beats = append(beats, eb.Beat)
	}

	var target *model.Beat
	for _, b := range beats {
		if b.ID == beatID {
			target = &b
			break
		}
	}
	if target == nil {
		fatalJSON("error", "beat not found: "+beatID)
	}

	viewStat := cache.ViewStats[beatID]
	breakdown := ripeness.CalculateWithBreakdown(*target, beats, viewStat)

	resp := map[string]interface{}{
		"beat_id": beatID,
		"score":   breakdown.Total,
		"tier":    model.RipenessTier(breakdown.Total),
		"factors": map[string]float64{
			"age":          breakdown.Age,
			"revisit":      breakdown.Revisit,
			"connection":   breakdown.Connection,
			"action":       breakdown.Action,
			"completeness": breakdown.Completeness,
		},
	}
	outputJSON(resp)
}

func robotRipe() {
	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	limit := 10
	threshold := 0.0
	var projectFilter string
	for i, arg := range os.Args {
		if arg == "--limit" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				limit = n
			}
		}
		if arg == "--threshold" && i+1 < len(os.Args) {
			if f, err := strconv.ParseFloat(os.Args[i+1], 64); err == nil {
				threshold = f
			}
		}
		if arg == "--project" && i+1 < len(os.Args) {
			projectFilter = os.Args[i+1]
		}
	}

	beats = loader.FilterBeatsByProject(beats, beatToProject, projectFilter)

	projects, _ := loader.DiscoverProjects(rootPath)
	var enriched []model.EnrichedBeat
	if len(projects) > 0 {
		enriched, _, _ = loader.LoadEnrichedBeats(projects[0].Path, nil)
	}

	// Filter enriched beats to match project filter
	beatSet := make(map[string]bool)
	for _, b := range beats {
		beatSet[b.ID] = true
	}
	var filteredEnriched []model.EnrichedBeat
	for _, eb := range enriched {
		if beatSet[eb.ID] && eb.RipenessScore >= threshold {
			filteredEnriched = append(filteredEnriched, eb)
		}
	}

	var filtered []model.EnrichedBeat
	for _, eb := range filteredEnriched {
		if eb.RipenessScore >= threshold {
			filtered = append(filtered, eb)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].RipenessScore > filtered[j].RipenessScore
	})

	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	var results []map[string]interface{}
	for _, eb := range filtered {
		results = append(results, map[string]interface{}{
			"id":       eb.ID,
			"ripeness": eb.RipenessScore,
			"tier":     model.RipenessTier(eb.RipenessScore),
			"preview":  eb.ContentPreview(80),
		})
	}

	outputJSON(map[string]interface{}{"beats": results, "count": len(results)})
}

func robotEntities() {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var people, tools, concepts []map[string]interface{}
	for _, e := range cache.Entities {
		item := map[string]interface{}{
			"name":       e.Name,
			"beat_count": len(e.BeatIDs),
		}
		switch e.Type {
		case model.EntityPerson:
			people = append(people, item)
		case model.EntityTool:
			tools = append(tools, item)
		case model.EntityConcept:
			concepts = append(concepts, item)
		}
	}

	outputJSON(map[string]interface{}{
		"people":   people,
		"tools":    tools,
		"concepts": concepts,
	})
}

func robotEntityBeats(entityName string) {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	beatIDs := cache.EntityIndex[entityName]
	if len(beatIDs) == 0 {
		outputJSON(map[string]interface{}{"beats": []interface{}{}, "entity": entityName})
		return
	}

	idSet := make(map[string]bool)
	for _, id := range beatIDs {
		idSet[id] = true
	}

	var results []map[string]interface{}
	for _, eb := range enriched {
		if idSet[eb.ID] {
			results = append(results, map[string]interface{}{
				"id":      eb.ID,
				"preview": eb.ContentPreview(80),
			})
		}
	}

	outputJSON(map[string]interface{}{"beats": results, "entity": entityName, "count": len(results)})
}

func robotTimeline() {
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	zoom := timeline.ZoomMonth
	for i, arg := range os.Args {
		if arg == "--zoom" && i+1 < len(os.Args) {
			switch os.Args[i+1] {
			case "day":
				zoom = timeline.ZoomDay
			case "week":
				zoom = timeline.ZoomWeek
			case "month":
				zoom = timeline.ZoomMonth
			case "quarter":
				zoom = timeline.ZoomQuarter
			}
		}
	}

	data := timeline.BuildTimeline(enriched, zoom)

	var buckets []map[string]interface{}
	for _, b := range data.Buckets {
		byChannel := make(map[string]int)
		for ch, count := range b.ByChannel {
			byChannel[ch.String()] = count
		}
		buckets = append(buckets, map[string]interface{}{
			"date":       b.Date.Format("2006-01-02"),
			"count":      b.BeatCount,
			"by_channel": byChannel,
		})
	}

	outputJSON(map[string]interface{}{
		"buckets":    buckets,
		"zoom_level": zoom.String(),
		"start":      data.Start.Format("2006-01-02"),
		"end":        data.End.Format("2006-01-02"),
	})
}

func robotGaps() {
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	threshold := 7 * 24 * time.Hour
	for i, arg := range os.Args {
		if arg == "--threshold" && i+1 < len(os.Args) {
			if days, err := strconv.Atoi(os.Args[i+1]); err == nil {
				threshold = time.Duration(days) * 24 * time.Hour
			}
		}
	}

	data := timeline.BuildTimeline(enriched, timeline.ZoomDay)
	gaps := data.FindGaps(threshold)

	var result []map[string]interface{}
	for _, g := range gaps {
		result = append(result, map[string]interface{}{
			"start":    g.Start.Format("2006-01-02"),
			"end":      g.End.Format("2006-01-02"),
			"days":     int(g.End.Sub(g.Start).Hours() / 24),
		})
	}

	outputJSON(map[string]interface{}{"gaps": result, "threshold_days": int(threshold.Hours() / 24)})
}

func robotCluster() {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	k := 8
	for i, arg := range os.Args {
		if arg == "--k" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				k = n
			}
		}
	}

	engine := cluster.NewEngine()
	if !engine.IsAvailable() {
		outputJSON(map[string]interface{}{
			"error":   "ollama not available",
			"message": "Install Ollama and run: ollama pull nomic-embed-text",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	clusters, err := engine.GenerateClusters(ctx, enriched, k)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	cache.Clusters = clusters
	cache.EmbeddingsAvailable = true

	// Persist clusters to cache
	rootPath := loader.GetDefaultRoot()
	projects, _ := loader.DiscoverProjects(rootPath)
	if len(projects) > 0 {
		loader.SaveCache(projects[0].Path, cache)
	}

	var result []map[string]interface{}
	for _, c := range clusters {
		result = append(result, map[string]interface{}{
			"id":         c.ID,
			"name":       c.Name,
			"beat_count": len(c.BeatIDs),
			"keywords":   c.Keywords,
			"ripeness":   c.RipenessScore,
		})
	}

	outputJSON(map[string]interface{}{"clusters": result, "count": len(result)})
}

func robotClusters() {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var result []map[string]interface{}
	for _, c := range cache.Clusters {
		result = append(result, map[string]interface{}{
			"id":         c.ID,
			"name":       c.Name,
			"beat_count": len(c.BeatIDs),
			"keywords":   c.Keywords,
			"ripeness":   c.RipenessScore,
		})
	}

	outputJSON(map[string]interface{}{
		"clusters":             result,
		"count":                len(result),
		"embeddings_available": cache.EmbeddingsAvailable,
	})
}

func robotClustersWithTemperature() {
	// Parse flags
	unclaimedOnly := false
	minTemperature := 0.0
	for i, arg := range os.Args {
		if arg == "--unclaimed" {
			unclaimedOnly = true
		}
		if arg == "--min-temperature" && i+1 < len(os.Args) {
			if f, err := strconv.ParseFloat(os.Args[i+1], 64); err == nil {
				minTemperature = f
			}
		}
	}

	// Load clusters from cache
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load WALD and config
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load config: "+err.Error())
	}

	// Load beats
	thermalBeats := loadThermalBeats(werkRoot)

	// Compute temperatures
	tempOutput := thermal.ComputeTemperature(thermalBeats, cache.Clusters, wald.Directories, config, werkRoot)

	// Build cluster to directory mapping (which directories claim which clusters)
	clusterToDirs := make(map[string][]string)
	for _, dir := range wald.Directories {
		for _, c := range cache.Clusters {
			dirText := strings.ToLower(dir.Path + " " + dir.Purpose)
			for _, keyword := range c.Keywords {
				if strings.Contains(dirText, strings.ToLower(keyword)) {
					clusterToDirs[c.ID] = append(clusterToDirs[c.ID], dir.Path)
					break
				}
			}
		}
	}

	// Build enriched beat map for temperature computation
	enrichedBeats, _, _ := getEnrichedBeats()
	beatMap := make(map[string]model.EnrichedBeat)
	for _, eb := range enrichedBeats {
		beatMap[eb.ID] = eb
	}

	// Build cluster output
	type ClusterOutput struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Keywords    []string `json:"keywords"`
		BeatCount   int      `json:"beat_count"`
		Temperature float64  `json:"temperature"`
		Ripeness    float64  `json:"ripeness"`
		Trend       string   `json:"trend"`
		ClaimedBy   []string `json:"claimed_by"`
		Unclaimed   bool     `json:"unclaimed"`
	}

	var clusters []ClusterOutput
	for _, c := range cache.Clusters {
		clusterTemp := computeClusterTemperature(c.BeatIDs, beatMap)

		trend := "stable"
		if tempOutput != nil {
			for _, dirPath := range clusterToDirs[c.ID] {
				if dirTemp, ok := tempOutput.Directories[dirPath]; ok && dirTemp.Trend != "" {
					trend = dirTemp.Trend
					break
				}
			}
		}

		claimedBy := clusterToDirs[c.ID]
		if claimedBy == nil {
			claimedBy = []string{}
		}
		unclaimed := len(claimedBy) == 0

		if unclaimedOnly && !unclaimed {
			continue
		}
		if clusterTemp < minTemperature {
			continue
		}

		clusters = append(clusters, ClusterOutput{
			ID:          c.ID,
			Name:        c.Name,
			Keywords:    c.Keywords,
			BeatCount:   len(c.BeatIDs),
			Temperature: clusterTemp,
			Ripeness:    c.RipenessScore,
			Trend:       trend,
			ClaimedBy:   claimedBy,
			Unclaimed:   unclaimed,
		})
	}

	sort.Slice(clusters, func(i, j int) bool {
		return clusters[i].Temperature > clusters[j].Temperature
	})

	outputJSON(map[string]interface{}{
		"clusters": clusters,
		"count":    len(clusters),
		"filters": map[string]interface{}{
			"unclaimed_only":  unclaimedOnly,
			"min_temperature": minTemperature,
		},
	})
}

func robotSimilar(beatID string) {
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var target *model.EnrichedBeat
	for _, eb := range enriched {
		if eb.ID == beatID {
			target = &eb
			break
		}
	}
	if target == nil {
		fatalJSON("error", "beat not found: "+beatID)
	}

	limit := 5
	for i, arg := range os.Args {
		if arg == "--limit" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				limit = n
			}
		}
	}

	engine := cluster.NewEngine()
	if !engine.IsAvailable() {
		outputJSON(map[string]interface{}{
			"error":   "ollama not available",
			"message": "Install Ollama for similarity search",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	similar, err := engine.FindSimilar(ctx, *target, enriched, limit)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var result []map[string]interface{}
	for _, eb := range similar {
		result = append(result, map[string]interface{}{
			"id":      eb.ID,
			"preview": eb.ContentPreview(80),
		})
	}

	outputJSON(map[string]interface{}{"similar": result, "source_beat": beatID})
}

func robotChains() {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var result []map[string]interface{}
	for _, c := range cache.Chains {
		result = append(result, map[string]interface{}{
			"id":         c.ID,
			"name":       c.Name,
			"beat_count": len(c.BeatIDs),
			"ripeness":   c.RipenessScore,
		})
	}

	outputJSON(map[string]interface{}{"chains": result, "count": len(result)})
}

func robotCreateChain() {
	var input struct {
		Name    string   `json:"name"`
		BeatIDs []string `json:"beat_ids"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	if input.Name == "" {
		fatalJSON("error", "chain name required")
	}

	chain := model.Chain{
		ID:        fmt.Sprintf("chain-%d", time.Now().UnixNano()),
		Name:      input.Name,
		BeatIDs:   input.BeatIDs,
		CreatedAt: time.Now(),
	}

	outputJSON(map[string]interface{}{
		"chain":   chain,
		"message": "Chain created (note: not persisted without cache save)",
	})
}

func robotChainAdd() {
	var input struct {
		ChainID string `json:"chain_id"`
		BeatID  string `json:"beat_id"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"chain_id": input.ChainID,
		"beat_id":  input.BeatID,
		"message":  "Beat added to chain (note: not persisted without cache save)",
	})
}

type StaleReason struct {
	Code       string `json:"code"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion"`
}

func getStaleReasons(eb model.EnrichedBeat) []StaleReason {
	var reasons []StaleReason
	ageDays := int(time.Since(eb.CreatedAt).Hours() / 24)

	if ageDays > 60 {
		reasons = append(reasons, StaleReason{
			Code:       "very_old",
			Message:    fmt.Sprintf("Beat is %d days old", ageDays),
			Suggestion: "Review for relevance, archive if outdated",
		})
	} else if ageDays > 30 {
		reasons = append(reasons, StaleReason{
			Code:       "old",
			Message:    fmt.Sprintf("Beat is %d days old", ageDays),
			Suggestion: "Consider converting to bead or archiving",
		})
	}

	if eb.ViewCount == 0 {
		reasons = append(reasons, StaleReason{
			Code:       "never_viewed",
			Message:    "Never viewed in btv",
			Suggestion: "Review content, may contain forgotten insight",
		})
	} else if eb.LastViewedAt != nil {
		daysSinceView := int(time.Since(*eb.LastViewedAt).Hours() / 24)
		if daysSinceView > 14 {
			reasons = append(reasons, StaleReason{
				Code:       "not_recently_viewed",
				Message:    fmt.Sprintf("Not viewed in %d days", daysSinceView),
				Suggestion: "Revisit to assess current relevance",
			})
		}
	}

	if len(eb.LinkedBeads) == 0 {
		reasons = append(reasons, StaleReason{
			Code:       "no_linked_beads",
			Message:    "Not linked to any beads",
			Suggestion: "Convert to bead if actionable",
		})
	}

	if len(eb.ChainIDs) == 0 {
		reasons = append(reasons, StaleReason{
			Code:       "not_in_chain",
			Message:    "Not part of any thought chain",
			Suggestion: "Add to chain if related to other beats",
		})
	}

	return reasons
}

func robotStale() {
	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var projectFilter string
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			projectFilter = os.Args[i+1]
		}
	}

	beats = loader.FilterBeatsByProject(beats, beatToProject, projectFilter)

	projects, _ := loader.DiscoverProjects(rootPath)
	var enriched []model.EnrichedBeat
	if len(projects) > 0 {
		enriched, _, _ = loader.LoadEnrichedBeats(projects[0].Path, nil)
	}

	// Filter enriched beats to match project filter
	beatSet := make(map[string]bool)
	for _, b := range beats {
		beatSet[b.ID] = true
	}
	var filteredEnriched []model.EnrichedBeat
	for _, eb := range enriched {
		if beatSet[eb.ID] {
			filteredEnriched = append(filteredEnriched, eb)
		}
	}

	stale := views.FindStaleBeats(filteredEnriched)

	var result []map[string]interface{}
	for _, eb := range stale {
		ageDays := int(time.Since(eb.CreatedAt).Hours() / 24)
		reasons := getStaleReasons(eb)

		primarySuggestion := "Review and take action"
		if len(reasons) > 0 {
			primarySuggestion = reasons[0].Suggestion
		}

		result = append(result, map[string]interface{}{
			"id":                eb.ID,
			"age_days":          ageDays,
			"view_count":        eb.ViewCount,
			"preview":           eb.ContentPreview(80),
			"reasons":           reasons,
			"suggested_action":  primarySuggestion,
		})
	}

	outputJSON(map[string]interface{}{"stale_beats": result, "count": len(result)})
}

func ensureAttentionState() (*model.Cache, error) {
	rootPath := loader.GetDefaultRoot()
	projects, err := loader.DiscoverProjects(rootPath)
	if err != nil || len(projects) == 0 {
		return nil, fmt.Errorf("no projects found")
	}

	projectPath := projects[0].Path
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			for _, p := range projects {
				if p.Name == os.Args[i+1] {
					projectPath = p.Path
					break
				}
			}
		}
	}

	// Load or create cache
	cache, err := loader.EnsureCache(projectPath, func(step string, current, total int) {
		if current > 0 && total > 0 {
			fmt.Fprintf(os.Stderr, "%s: %d/%d\r", step, current, total)
		}
	})
	if err != nil {
		return nil, fmt.Errorf("ensuring cache: %w", err)
	}

	// Load beats for potential cluster generation and attention computation
	beats, err := loader.LoadBeats(projectPath)
	if err != nil {
		return nil, fmt.Errorf("loading beats: %w", err)
	}

	// Note if clusters are empty - don't auto-generate as it's slow and may OOM
	if len(cache.Clusters) == 0 && len(beats) > 0 {
		fmt.Fprintf(os.Stderr, "Note: Clusters empty. Run 'btv --robot-cluster' for full attention analysis.\n")
	}

	// Check if attention state exists and is recent
	if cache.AttentionStateJSON == nil || len(cache.AttentionStateJSON) == 0 {
		// Compute and save attention state
		state := loader.ComputeAttentionState(beats, cache.Clusters, cache.Ripeness)
		if attJSON, err := json.Marshal(state); err == nil {
			cache.AttentionStateJSON = attJSON
			if err := loader.SaveCache(projectPath, cache); err != nil {
				return nil, fmt.Errorf("saving cache: %w", err)
			}
		}
	}

	return cache, nil
}

func rebuildCache() {
	rootPath := loader.GetDefaultRoot()
	for i, arg := range os.Args {
		if arg == "--root" && i+1 < len(os.Args) {
			rootPath = os.Args[i+1]
		}
	}

	projects, err := loader.DiscoverProjects(rootPath)
	if err != nil || len(projects) == 0 {
		fatalJSON("error", "no projects found")
	}

	var projectPath string
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			for _, p := range projects {
				if p.Name == os.Args[i+1] {
					projectPath = p.Path
					break
				}
			}
		}
	}
	if projectPath == "" {
		projectPath = projects[0].Path
	}

	fmt.Fprintf(os.Stderr, "Rebuilding cache for: %s\n", projectPath)

	progressFn := func(step string, current, total int) {
		if total > 0 {
			fmt.Fprintf(os.Stderr, "\r%s: %d/%d", step, current, total)
		} else {
			fmt.Fprintf(os.Stderr, "\r%s...", step)
		}
	}

	cache, err := loader.RefreshCache(projectPath, progressFn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n")
		fatalJSON("error", err.Error())
	}
	fmt.Fprintf(os.Stderr, "\n")

	outputJSON(map[string]interface{}{
		"success":      true,
		"version":      cache.Version,
		"generated_at": cache.GeneratedAt,
		"source_hash":  cache.SourceHash,
		"beats_count":  len(cache.Taxonomies),
		"entities":     len(cache.Entities),
	})
}

func runCapture() {
	fmt.Println("Quick capture mode - implement with minimal TUI")
	fmt.Println("For now, use: bt add \"your insight\"")
}

// ExtendedActivation adds temperature and related_directories to an activation.
type ExtendedActivation struct {
	ClusterID          string   `json:"cluster_id"`
	ClusterName        string   `json:"cluster_name"`
	Type               string   `json:"type"`
	BeatCount          int      `json:"beat_count"`
	WindowDays         int      `json:"window_days"`
	PriorActivity      int      `json:"prior_activity"`
	Temperature        float64  `json:"temperature"`
	RelatedDirectories []string `json:"related_directories"`
	SampleBeats        []string `json:"sample_beats"`
}

func robotAttention() {
	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var projectFilter string
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			projectFilter = os.Args[i+1]
		}
	}

	beats = loader.FilterBeatsByProject(beats, beatToProject, projectFilter)
	beatSet := make(map[string]bool)
	for _, b := range beats {
		beatSet[b.ID] = true
	}

	// Ensure attention state is computed and available
	cache, err := ensureAttentionState()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	state, err := loader.GetAttentionState(cache)
	if err != nil {
		fatalJSON("error", err.Error())
	}
	if state == nil {
		// This should never happen now, but keep as fallback
		outputJSON(map[string]interface{}{"error": "no attention state computed"})
		return
	}

	// Load enriched beats for additional context
	enrichedBeats, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	// Filter activations by project if specified
	if projectFilter != "" && state.Activations != nil {
		var filtered []attention.Activation
		for _, act := range state.Activations {
			var filteredBeats []string
			for _, bid := range act.Beats {
				if beatSet[bid] {
					filteredBeats = append(filteredBeats, bid)
				}
			}
			if len(filteredBeats) > 0 {
				act.Beats = filteredBeats
				filtered = append(filtered, act)
			}
		}
		state.Activations = filtered
	}

	// Load WALD directories for related_directories computation
	werkRoot := thermal.FindWerkRoot()
	var waldDirs []thermal.WALDDirectory
	if werkRoot != "" {
		wald, err := thermal.LoadWALD(werkRoot)
		if err == nil {
			waldDirs = wald.Directories
		}
	}

	// Build beat ID to enriched beat map for context lookup
	beatMap := make(map[string]model.EnrichedBeat)
	for _, eb := range enrichedBeats {
		beatMap[eb.ID] = eb
	}

	// Build extended activations with temperature and related_directories
	var extendedActivations []ExtendedActivation
	for _, act := range state.Activations {
		// Compute temperature for this cluster based on recency-weighted beats
		clusterTemp := computeClusterTemperature(act.Beats, beatMap)

		// Find related directories
		relatedDirs := findRelatedDirectories(act.ClusterName, waldDirs, werkRoot)

		// Get sample beats (up to 3)
		sampleBeats := act.Beats
		if len(sampleBeats) > 3 {
			sampleBeats = sampleBeats[:3]
		}

		extendedActivations = append(extendedActivations, ExtendedActivation{
			ClusterID:          act.ClusterID,
			ClusterName:        act.ClusterName,
			Type:               act.Type.String(),
			BeatCount:          act.BeatCount,
			WindowDays:         int(act.Window.Hours() / 24),
			PriorActivity:      act.PriorActivity,
			Temperature:        clusterTemp,
			RelatedDirectories: relatedDirs,
			SampleBeats:        sampleBeats,
		})
	}

	// Output extended state
	output := map[string]interface{}{
		"computed_at": state.ComputedAt,
		"activations": extendedActivations,
	}
	if state.DriftReport != nil {
		output["drift_report"] = state.DriftReport
	}
	if state.Orientation != nil {
		output["orientation"] = state.Orientation
	}
	if state.Heartbeat != nil {
		output["heartbeat"] = state.Heartbeat
	}

	outputJSON(output)
}

// runCrystallizeCmd handles the `btv crystallize <cluster-id>` command
func runCrystallizeCmd() {
	// Parse flags
	var pathOverride, purposeOverride string
	var dryRun bool
	var clusterID string

	args := os.Args[2:] // skip "btv" and "crystallize"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 < len(args) {
				pathOverride = args[i+1]
				i++
			}
		case "--purpose":
			if i+1 < len(args) {
				purposeOverride = args[i+1]
				i++
			}
		case "--dry-run":
			dryRun = true
		default:
			if !strings.HasPrefix(args[i], "-") && clusterID == "" {
				clusterID = args[i]
			}
		}
	}

	if clusterID == "" {
		fatal("crystallize requires cluster ID argument")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatal("WALD.yaml not found - not in a werk directory")
	}

	// Load emergence cache
	emergence, err := thermal.LoadEmergenceCache(werkRoot)
	if err != nil {
		fmt.Println("Cluster not found in cache, computing emergence...")
		emergence = recomputeEmergence(werkRoot)
	}

	cluster := thermal.FindClusterByID(emergence, clusterID)
	if cluster == nil && emergence != nil {
		// Try recomputing
		fmt.Println("Cluster not found, recomputing emergence...")
		emergence = recomputeEmergence(werkRoot)
		cluster = thermal.FindClusterByID(emergence, clusterID)
	}

	if cluster == nil {
		fatal("Cluster not found: " + clusterID)
	}

	// Apply overrides
	if pathOverride != "" {
		cluster.SuggestedPath = pathOverride
	}
	if purposeOverride != "" {
		cluster.SuggestedPurpose = purposeOverride
	}

	// Show preview
	showCrystallizePreview(cluster, werkRoot)

	if dryRun {
		fmt.Println("\n[dry-run mode - no changes made]")
		return
	}

	// Confirmation
	if !promptYesNo("Proceed?") {
		fmt.Println("Aborted.")
		return
	}

	// Execute crystallization
	beats := loadThermalBeats(werkRoot)
	input := thermal.CrystallizeInput{
		ClusterID:       clusterID,
		PathOverride:    pathOverride,
		PurposeOverride: purposeOverride,
		Confirm:         true,
	}

	result, err := thermal.Crystallize(werkRoot, input, cluster, beats)
	if err != nil {
		fatal(err.Error())
	}

	fmt.Printf("\n✓ Created %s\n", result.Created.Directory)
	fmt.Printf("✓ Generated %s\n", result.Created.AgentsMd)
	fmt.Printf("✓ %s\n", result.Message)
}

func showCrystallizePreview(cluster *thermal.EmergentCluster, werkRoot string) {
	fmt.Println("Crystallize Cluster")
	fmt.Println("═══════════════════")
	fmt.Println()
	fmt.Printf("Cluster: %s\n", cluster.ClusterName)
	fmt.Printf("Temperature: %.2f\n", cluster.Temperature)
	fmt.Printf("Beats: %d\n", cluster.BeatCount)
	if len(cluster.SampleBeats) > 0 {
		var keywords []string
		for _, sample := range cluster.SampleBeats {
			keywords = append(keywords, sample.Preview)
		}
		if len(keywords) > 3 {
			keywords = keywords[:3]
		}
		fmt.Printf("Keywords: %s\n", strings.Join(keywords, ", "))
	}
	fmt.Println()

	fmt.Println("This will create:")
	fmt.Printf("  Directory: %s/\n", cluster.SuggestedPath)
	fmt.Println("  AGENTS.md with purpose derived from cluster")
	fmt.Println("  WALD.yaml entry with claims:")
	fmt.Printf("    clusters: [\"%s\"]\n", cluster.ClusterName)
	fmt.Printf("    topics: [extracted from cluster]\n")
	fmt.Println()
	fmt.Println("The directory will claim this cluster.")
	fmt.Println("Beats remain in the global store.")
	fmt.Println()
}

func promptYesNo(prompt string) bool {
	fmt.Printf("%s [y/N] ", prompt)
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func recomputeEmergence(werkRoot string) *thermal.EmergenceOutput {
	// Load WALD and config
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		return nil
	}
	config, _ := thermal.LoadConfig(werkRoot)

	// Load clusters from cache
	_, cache, err := getEnrichedBeats()
	if err != nil || cache == nil {
		return nil
	}

	// Load beats
	beats := loadThermalBeats(werkRoot)

	// Detect emergence
	emergence := thermal.DetectEmergence(cache.Clusters, wald.Directories, beats, config)
	thermal.SaveEmergenceCache(werkRoot, emergence)
	return emergence
}

func loadThermalBeats(werkRoot string) []thermal.Beat {
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		return nil
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}
	return thermalBeats
}

// robotCrystallizeCluster handles --robot-crystallize (stdin JSON input)
func robotCrystallizeCluster() {
	var input thermal.CrystallizeInput
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid input: "+err.Error())
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load emergence data to find cluster
	emergence, err := thermal.LoadEmergenceCache(werkRoot)
	if err != nil {
		emergence = recomputeEmergence(werkRoot)
	}

	cluster := thermal.FindClusterByID(emergence, input.ClusterID)
	if cluster == nil {
		fatalJSON("error", "cluster not found: "+input.ClusterID)
	}

	// Load beats
	beats := loadThermalBeats(werkRoot)

	if !input.Confirm {
		// Preview mode
		result, _ := thermal.PreviewCrystallize(werkRoot, cluster, beats)
		outputJSON(result)
		return
	}

	// Do crystallization
	result, err := thermal.Crystallize(werkRoot, input, cluster, beats)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	outputJSON(result)
}
func robotSyncPreview() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load config: "+err.Error())
	}

	// Load all beats
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	tempOutput, _ := thermal.ComputeTemperatureWithCache(thermalBeats, clusters, wald.Directories, config, werkRoot)

	// Detect emergence (optional)
	var emergence *thermal.EmergenceResult
	if cache != nil {
		emergence = thermal.DetectEmergence(cache.Clusters, wald.Directories, thermalBeats, config)
	}

	// Generate preview
	preview := thermal.GenerateSyncPreview(tempOutput, wald, emergence, config)

	outputJSON(preview)
}

func robotSyncApply() {
	// Read confirmation from stdin
	var input struct {
		Confirm bool `json:"confirm"`
	}
	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	if !input.Confirm {
		fatalJSON("error", "confirmation required - send {\"confirm\": true}")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load config: "+err.Error())
	}

	// Load all beats
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	tempOutput, _ := thermal.ComputeTemperatureWithCache(thermalBeats, clusters, wald.Directories, config, werkRoot)

	// Detect emergence (optional)
	var emergence *thermal.EmergenceResult
	if cache != nil {
		emergence = thermal.DetectEmergence(cache.Clusters, wald.Directories, thermalBeats, config)
	}

	// Generate preview first
	preview := thermal.GenerateSyncPreview(tempOutput, wald, emergence, config)

	if len(preview.StateChanges) == 0 {
		outputJSON(map[string]interface{}{
			"applied_at":               time.Now(),
			"backup_path":              "",
			"changes_applied":          []interface{}{},
			"wald_yaml_updated":        false,
			"temperature_cache_updated": false,
			"errors":                   []string{},
			"message":                  "no changes to apply",
		})
		return
	}

	// Apply changes
	result, err := thermal.ApplySyncChanges(preview, werkRoot, config)
	if err != nil {
		fatalJSON("error", "failed to apply changes: "+err.Error())
	}

	outputJSON(result)
}
func runSync(dryRun, force bool) {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fmt.Fprintln(os.Stderr, "Error: WALD.yaml not found - not in a werk directory")
		os.Exit(1)
	}

	// Print header
	fmt.Println("Thermal WALD Sync")
	fmt.Println("═════════════════")
	fmt.Println()

	// Show computing status
	fmt.Println("Computing temperatures...")

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading WALD.yaml: %v\n", err)
		os.Exit(1)
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Load all beats
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error discovering projects: %v\n", err)
		os.Exit(1)
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	fmt.Printf("  Window: %d days\n", config.Temperature.WindowDays)
	fmt.Printf("  Beats stores: %d\n", len(projects))
	fmt.Printf("  Beats analyzed: %d\n", len(thermalBeats))
	fmt.Printf("  Directories: %d\n", len(wald.Directories))
	fmt.Println()

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	tempOutput, err := thermal.ComputeTemperatureWithCache(thermalBeats, clusters, wald.Directories, config, werkRoot)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Warning: %v\n", err)
	}

	// Generate preview
	preview := thermal.GenerateSyncPreview(tempOutput, wald, nil, config)

	// Display cluster temperatures (PRIMARY)
	showClusterTemperatures(preview.ClusterTemperatures)

	// Display directory temperatures (derived from claims)
	showDirectoryTemperatures(tempOutput)

	// Display unclaimed hot clusters
	showUnclaimedClusters(preview.UnclaimedHotClusters)

	// Display claim suggestions
	showClaimSuggestions(preview.ClaimSuggestions)

	// Display state changes
	showStateChanges(preview.StateChanges)

	// Display preserved
	showPreserved(preview.Preserved)

	// Display emergent structures
	showEmergence(preview.EmergenceReady)

	if dryRun {
		fmt.Println("\n[dry-run mode - no changes made]")
		return
	}

	if len(preview.StateChanges) == 0 {
		fmt.Println("\nNo changes to apply.")
		return
	}

	// Confirmation prompt (unless --force)
	if !force {
		response := promptConfirmation()
		if response == "review" {
			showDetailedReview(preview)
			response = promptConfirmation()
		}
		if response != "y" && response != "yes" {
			fmt.Println("Aborted.")
			return
		}
	}

	// Apply changes
	result, err := thermal.ApplySyncChanges(preview, werkRoot, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(result.Errors) > 0 {
		for _, errMsg := range result.Errors {
			fmt.Fprintf(os.Stderr, "Error: %s\n", errMsg)
		}
		os.Exit(1)
	}

	fmt.Printf("\nBackup created: %s\n", result.BackupPath)
	fmt.Printf("Applied %d changes.\n", len(result.ChangesApplied))
}

func temperatureBar(temp float64) string {
	filled := int(temp * 10)
	if filled > 10 {
		filled = 10
	}
	if filled < 0 {
		filled = 0
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", 10-filled)
}

func trendArrow(trend string) string {
	switch trend {
	case "rising":
		return "↑"
	case "falling":
		return "↓"
	default:
		return "→"
	}
}

func showClusterTemperatures(clusters []thermal.ClusterSyncInfo) {
	if len(clusters) == 0 {
		return
	}

	fmt.Println("Cluster Temperatures")
	fmt.Println("────────────────────")

	limit := 10
	if len(clusters) < limit {
		limit = len(clusters)
	}

	for i := 0; i < limit; i++ {
		c := clusters[i]
		bar := temperatureBar(c.Temperature)
		claimedStr := "UNCLAIMED"
		if len(c.ClaimedBy) > 0 {
			claimedStr = "claimed by: " + c.ClaimedBy[0]
			if len(c.ClaimedBy) > 1 {
				claimedStr += fmt.Sprintf(" +%d", len(c.ClaimedBy)-1)
			}
		}
		fmt.Printf("  %-30s %.2f %s %-5s (%d beats, %s)\n",
			truncatePath(c.Name, 30),
			c.Temperature,
			bar,
			c.State,
			c.BeatCount,
			claimedStr)
	}

	if len(clusters) > limit {
		fmt.Printf("  ... and %d more clusters\n", len(clusters)-limit)
	}
	fmt.Println()
}

func showDirectoryTemperatures(tempOutput *thermal.TemperatureOutput) {
	fmt.Println("Directory Temperatures (derived from claims)")
	fmt.Println("────────────────────────────────────────────")

	// Sort directories by temperature
	type dirEntry struct {
		path string
		temp *thermal.DirectoryTemperature
	}
	var entries []dirEntry
	for path, temp := range tempOutput.Directories {
		entries = append(entries, dirEntry{path, temp})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].temp.Temperature > entries[j].temp.Temperature
	})

	// Show top 10
	limit := 10
	if len(entries) < limit {
		limit = len(entries)
	}

	for i := 0; i < limit; i++ {
		e := entries[i]
		state := e.temp.StateInferred
		bar := temperatureBar(e.temp.Temperature)
		clusterInfo := fmt.Sprintf("claims %d clusters, %d beats", len(e.temp.ClaimedClusters), e.temp.ClaimedBeatCount)
		fmt.Printf("  %-30s %.2f %s %-5s (%s)\n",
			truncatePath(e.path, 30),
			e.temp.Temperature,
			bar,
			state,
			clusterInfo)
	}

	if len(entries) > limit {
		fmt.Printf("  ... and %d more directories\n", len(entries)-limit)
	}
	fmt.Println()
}

func showUnclaimedClusters(unclaimed []thermal.UnclaimedCluster) {
	if len(unclaimed) == 0 {
		return
	}

	fmt.Println("Unclaimed Hot Clusters")
	fmt.Println("──────────────────────")

	for _, c := range unclaimed {
		fmt.Printf("  ⚡ %s (%.2f, %d beats)\n", c.Name, c.Temperature, c.BeatCount)
		if c.SuggestedPath != "" {
			fmt.Printf("     Suggested: %s\n", c.SuggestedPath)
		}
		fmt.Println("     Or add claim to existing directory")
	}
	fmt.Println()
}

func showClaimSuggestions(suggestions []thermal.ClaimSuggestion) {
	if len(suggestions) == 0 {
		return
	}

	fmt.Println("Suggested Claim Updates")
	fmt.Println("───────────────────────")

	for _, s := range suggestions {
		fmt.Printf("  %s could also claim:\n", s.Directory)
		for _, cluster := range s.SuggestedClaims.Clusters {
			fmt.Printf("    - cluster \"%s\"\n", cluster)
		}
		if s.Reason != "" {
			fmt.Printf("    (%s)\n", s.Reason)
		}
	}
	fmt.Println()
}

func showStateChanges(changes []thermal.StateChange) {
	if len(changes) == 0 {
		return
	}

	fmt.Println("State Changes Detected")
	fmt.Println("──────────────────────")

	for _, change := range changes {
		fmt.Printf("  %s:\n", change.Path)
		fmt.Printf("    %s → %s (temp: %.2f", change.CurrentState, change.InferredState, change.Temperature)
		if change.Reason != "" {
			fmt.Printf(", %s", change.Reason)
		}
		fmt.Println(")")
	}
	fmt.Println()
}

func showPreserved(preserved []thermal.PreservedDirectory) {
	if len(preserved) == 0 {
		return
	}

	fmt.Println("Preserved (no state change)")
	fmt.Println("───────────────────────────")

	for _, p := range preserved {
		fmt.Printf("  %s (%s, temp: %.2f)\n", p.Path, p.Reason, p.Temperature)
	}
	fmt.Println()
}

func showEmergence(candidates []thermal.EmergentCluster) {
	if len(candidates) == 0 {
		return
	}

	fmt.Println("Emergent Structures")
	fmt.Println("───────────────────")

	for _, c := range candidates {
		fmt.Printf("  ⚡ %s\n", c.ClusterID)
		fmt.Printf("     ripeness: %.2f, temperature: %.2f\n", c.Ripeness, c.Temperature)
		if c.SuggestedPath != "" {
			fmt.Printf("     Suggested: %s\n", c.SuggestedPath)
		}
		fmt.Printf("     Run: btv crystallize %s\n", c.ClusterID)
	}
	fmt.Println()
}

func promptConfirmation() string {
	fmt.Print("\nApply changes to WALD.yaml? [y/N/review] ")
	var response string
	fmt.Scanln(&response)
	return strings.ToLower(response)
}

func showDetailedReview(preview *thermal.SyncPreview) {
	fmt.Println("\n=== Detailed Review ===")
	fmt.Println()

	for _, change := range preview.StateChanges {
		fmt.Printf("Directory: %s\n", change.Path)
		fmt.Printf("  Current state: %s\n", change.CurrentState)
		fmt.Printf("  New state: %s\n", change.InferredState)
		fmt.Printf("  Temperature: %.2f\n", change.Temperature)
		if change.Reason != "" {
			fmt.Printf("  Reason: %s\n", change.Reason)
		}
		fmt.Println()
	}
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	return "..." + path[len(path)-maxLen+3:]
}
func robotEmergence() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Parse --threshold flag
	var thresholdOverride *float64
	for i, arg := range os.Args {
		if arg == "--threshold" && i+1 < len(os.Args) {
			if f, err := strconv.ParseFloat(os.Args[i+1], 64); err == nil {
				thresholdOverride = &f
			}
		}
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load config: "+err.Error())
	}

	// Load clusters from cache
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	// Load all beats as thermal.Beat
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Build emergence config - use full Config from thermal
	emergenceConfig := config
	if emergenceConfig == nil {
		emergenceConfig = thermal.DefaultConfig()
	}
	if thresholdOverride != nil {
		emergenceConfig.Emergence.MinRipeness = *thresholdOverride
	}

	// Detect emergence
	result := thermal.DetectEmergence(cache.Clusters, wald.Directories, thermalBeats, emergenceConfig)

	outputJSON(result)
}
// computeClusterTemperature computes recency-weighted temperature for a cluster's beats.
func computeClusterTemperature(beatIDs []string, beatMap map[string]model.EnrichedBeat) float64 {
	if len(beatIDs) == 0 {
		return 0.0
	}

	now := time.Now()
	lambda := 0.1 // decay rate
	totalWeight := 0.0

	for _, bid := range beatIDs {
		eb, ok := beatMap[bid]
		if !ok {
			continue
		}
		ageDays := now.Sub(eb.CreatedAt).Hours() / 24
		weight := math.Exp(-lambda * ageDays)
		totalWeight += weight
	}

	// Temperature is based on average recency weight, scaled to 0-1
	avgWeight := totalWeight / float64(len(beatIDs))
	if avgWeight > 1.0 {
		avgWeight = 1.0
	}
	return math.Round(avgWeight*100) / 100
}

// findRelatedDirectories finds WALD directories that overlap with a cluster's keywords.
func findRelatedDirectories(clusterName string, waldDirs []thermal.WALDDirectory, werkRoot string) []string {
	if len(waldDirs) == 0 || werkRoot == "" {
		return []string{}
	}

	// Extract keywords from cluster name
	clusterKeywords := strings.Fields(strings.ToLower(clusterName))

	// Score directories by keyword overlap
	type dirScore struct {
		path  string
		score int
	}
	var scored []dirScore

	for _, dir := range waldDirs {
		score := 0
		dirText := strings.ToLower(dir.Path + " " + dir.Purpose)

		for _, keyword := range clusterKeywords {
			if len(keyword) < 3 {
				continue // Skip short words
			}
			if strings.Contains(dirText, keyword) {
				score++
			}
		}

		if score > 0 {
			scored = append(scored, dirScore{dir.Path, score})
		}
	}

	// Sort by score descending
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// Return top matching directories (limit to 5)
	var result []string
	for i, ds := range scored {
		if i >= 5 {
			break
		}
		result = append(result, ds.path)
	}
	return result
}

func robotActivating() {
	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var projectFilter string
	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			projectFilter = os.Args[i+1]
		}
	}

	beats = loader.FilterBeatsByProject(beats, beatToProject, projectFilter)
	beatSet := make(map[string]bool)
	for _, b := range beats {
		beatSet[b.ID] = true
	}

	// Ensure attention state is computed and available
	cache, err := ensureAttentionState()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	state, err := loader.GetAttentionState(cache)
	if err != nil {
		fatalJSON("error", err.Error())
	}
	if state == nil || len(state.Activations) == 0 {
		// Return empty but valid response
		outputJSON(map[string]interface{}{"activations": []interface{}{}, "count": 0})
		return
	}

	// Filter activations by project if specified
	activations := state.Activations
	if projectFilter != "" {
		var filtered []attention.Activation
		for _, act := range activations {
			var filteredBeats []string
			for _, bid := range act.Beats {
				if beatSet[bid] {
					filteredBeats = append(filteredBeats, bid)
				}
			}
			if len(filteredBeats) > 0 {
				act.Beats = filteredBeats
				filtered = append(filtered, act)
			}
		}
		activations = filtered
	}

	outputJSON(map[string]interface{}{
		"activations": activations,
		"count":       len(activations),
		"computed_at": state.ComputedAt,
	})
}

func robotDrift() {
	days := 30
	for i, arg := range os.Args {
		if arg == "--days" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				days = n
			}
		}
	}

	// Ensure attention state is computed and available
	cache, err := ensureAttentionState()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	state, err := loader.GetAttentionState(cache)
	if err != nil {
		fatalJSON("error", err.Error())
	}
	if state == nil || state.DriftReport == nil {
		// This should never happen now, but keep as fallback
		outputJSON(map[string]interface{}{"error": "no drift data available"})
		return
	}

	report := state.DriftReport
	outputJSON(map[string]interface{}{
		"drift_report":    report,
		"requested_days":  days,
		"window_days":     int(report.Window.Hours() / 24),
	})
}

func robotOrientation() {
	// Ensure attention state is computed and available
	cache, err := ensureAttentionState()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	state, err := loader.GetAttentionState(cache)
	if err != nil {
		fatalJSON("error", err.Error())
	}
	if state == nil || state.Orientation == nil {
		// This should never happen now, but keep as fallback
		outputJSON(map[string]interface{}{"error": "no orientation data available"})
		return
	}

	outputJSON(state.Orientation)
}

func robotHeartbeat() {
	days := 90
	for i, arg := range os.Args {
		if arg == "--days" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				days = n
			}
		}
	}

	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	state, err := loader.GetAttentionState(cache)
	if err != nil {
		fatalJSON("error", err.Error())
	}
	if state == nil || state.Heartbeat == nil {
		outputJSON(map[string]interface{}{"error": "no heartbeat data available"})
		return
	}

	hb := state.Heartbeat
	outputJSON(map[string]interface{}{
		"heartbeat":       hb,
		"requested_days":  days,
		"window_days":     int(hb.Window.Hours() / 24),
	})
}

func robotCrystallizations() {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	outputJSON(map[string]interface{}{
		"crystallizations": cache.Crystallizations,
		"count":            len(cache.Crystallizations),
	})
}

func robotCrystallization(beadID string) {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	for _, c := range cache.Crystallizations {
		if c.BeadID == beadID {
			outputJSON(c)
			return
		}
	}

	outputJSON(map[string]interface{}{"error": "no crystallization found for bead: " + beadID})
}

func robotInfer() {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var beats []model.Beat
	for _, eb := range enriched {
		beats = append(beats, eb.Beat)
	}

	state := loader.ComputeAttentionState(beats, cache.Clusters, cache.Ripeness)

	outputJSON(map[string]interface{}{
		"attention_state":  state,
		"crystallizations": cache.Crystallizations,
		"computed_at":      time.Now(),
	})
}

func robotDivergence() {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var beats []model.Beat
	for _, eb := range enriched {
		beats = append(beats, eb.Beat)
	}

	classifier := divergence.NewClassifier(divergence.DefaultClassifierConfig())
	analyzer := divergence.NewAnalyzer(classifier, divergence.DefaultAnalyzerConfig())
	report := analyzer.Analyze(beats, cache.Clusters)

	outputJSON(report)
}

func robotBlindspots() {
	enriched, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var beats []model.Beat
	for _, eb := range enriched {
		beats = append(beats, eb.Beat)
	}

	classifier := divergence.NewClassifier(divergence.DefaultClassifierConfig())
	analyzer := divergence.NewAnalyzer(classifier, divergence.DefaultAnalyzerConfig())
	report := analyzer.Analyze(beats, cache.Clusters)

	outputJSON(map[string]interface{}{
		"blind_spots": report.BlindSpots,
		"agent_only":  report.AgentOnly,
		"count":       len(report.BlindSpots),
	})
}

func robotAgentBeats() {
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	classifier := divergence.NewClassifier(divergence.DefaultClassifierConfig())

	var agentBeats []map[string]interface{}
	for _, eb := range enriched {
		origin := classifier.Classify(eb.Beat)
		if origin == divergence.OriginAgent {
			agentBeats = append(agentBeats, map[string]interface{}{
				"id":      eb.ID,
				"preview": eb.ContentPreview(80),
			})
		}
	}

	outputJSON(map[string]interface{}{
		"agent_beats": agentBeats,
		"count":       len(agentBeats),
	})
}

func robotAlerts() {
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	unseenOnly := false
	for _, arg := range os.Args {
		if arg == "--unseen" {
			unseenOnly = true
		}
	}

	var alerts []model.Alert
	for _, a := range cache.Alerts {
		if unseenOnly && a.SeenAt != nil {
			continue
		}
		alerts = append(alerts, a)
	}

	outputJSON(map[string]interface{}{
		"alerts":      alerts,
		"count":       len(alerts),
		"unseen_only": unseenOnly,
	})
}

func robotDismissAlert(alertID string) {
	rootPath := loader.GetDefaultRoot()
	projects, err := loader.DiscoverProjects(rootPath)
	if err != nil || len(projects) == 0 {
		fatalJSON("error", "no projects found")
	}

	_, cache, err := loader.LoadEnrichedBeats(projects[0].Path, nil)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	found := false
	now := time.Now()
	for i := range cache.Alerts {
		if cache.Alerts[i].ID == alertID {
			cache.Alerts[i].SeenAt = &now
			found = true
			break
		}
	}

	if !found {
		fatalJSON("error", "alert not found: "+alertID)
	}

	if err := loader.SaveCache(projects[0].Path, cache); err != nil {
		fatalJSON("error", "failed to save cache: "+err.Error())
	}

	outputJSON(map[string]interface{}{
		"success":  true,
		"alert_id": alertID,
		"seen_at":  now,
	})
}

func robotCrystallizeSuggestions() {
	rootPath := loader.GetDefaultRoot()
	beats, beatToProject, err := loader.LoadAllBeats(rootPath)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	var projectFilter *string
	minRipeness := 0.5
	maxSuggestions := 10

	for i, arg := range os.Args {
		if arg == "--project" && i+1 < len(os.Args) {
			p := os.Args[i+1]
			projectFilter = &p
		}
		if arg == "--min-ripeness" && i+1 < len(os.Args) {
			if f, err := strconv.ParseFloat(os.Args[i+1], 64); err == nil {
				minRipeness = f
			}
		}
		if arg == "--limit" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				maxSuggestions = n
			}
		}
	}

	if projectFilter != nil {
		beats = loader.FilterBeatsByProject(beats, beatToProject, *projectFilter)
	}

	projects, _ := loader.DiscoverProjects(rootPath)
	var enriched []model.EnrichedBeat
	var cache *model.Cache
	if len(projects) > 0 {
		enriched, cache, _ = loader.LoadEnrichedBeats(projects[0].Path, nil)
	}

	// Filter enriched beats to match project filter
	beatSet := make(map[string]bool)
	for _, b := range beats {
		beatSet[b.ID] = true
	}
	var filteredEnriched []model.EnrichedBeat
	for _, eb := range enriched {
		if beatSet[eb.ID] {
			filteredEnriched = append(filteredEnriched, eb)
		}
	}

	// Count total ripe beats
	totalRipe := 0
	for _, eb := range filteredEnriched {
		if eb.RipenessScore >= minRipeness && len(eb.LinkedBeads) == 0 {
			totalRipe++
		}
	}

	opts := crystallize.SuggestOptions{
		MinRipeness:    minRipeness,
		MaxSuggestions: maxSuggestions,
		IncludeRelated: true,
	}

	suggestions := crystallize.GenerateSuggestions(filteredEnriched, cache, opts)

	outputJSON(crystallize.SuggestionsResponse{
		Suggestions:   suggestions,
		TotalRipe:     totalRipe,
		ProjectFilter: projectFilter,
	})
}

func robotTemperature() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load .wald/config.yaml: "+err.Error())
	}

	// Load all beats from all .beats directories under werk
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var beatsStores []string
	var thermalBeats []thermal.Beat

	for _, proj := range projects {
		beatsStores = append(beatsStores, proj.Path)
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			// Extract context if present in raw beat data
			// Note: model.Beat doesn't have Context field yet, so we parse from capture path
			// For now, infer from project path
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath) // Remove .beats
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures with cache (for trends)
	output, err := thermal.ComputeTemperatureWithCache(thermalBeats, clusters, wald.Directories, config, werkRoot)
	if err != nil {
		// Log error but continue - cache save failure shouldn't block output
		fmt.Fprintf(os.Stderr, "Warning: failed to save temperature cache: %v\n", err)
	}
	output.BeatsStores = beatsStores

	// Convert clusters to output format (clusters are PRIMARY)
	type ClusterOutput struct {
		ID          string   `json:"id"`
		Name        string   `json:"name"`
		Temperature float64  `json:"temperature"`
		BeatCount   int      `json:"beat_count"`
		Ripeness    float64  `json:"ripeness"`
		Trend       string   `json:"trend"`
		ClaimedBy   []string `json:"claimed_by"`
	}

	var clustersOutput []ClusterOutput
	for _, ct := range output.Clusters {
		clustersOutput = append(clustersOutput, ClusterOutput{
			ID:          ct.ID,
			Name:        ct.Name,
			Temperature: ct.Temperature,
			BeatCount:   ct.BeatCount,
			Ripeness:    ct.Ripeness,
			Trend:       ct.Trend,
			ClaimedBy:   ct.ClaimedBy,
		})
	}

	// Convert directories to output format (directories are DERIVED from clusters)
	type DirectoryOutput struct {
		Path             string   `json:"path"`
		Temperature      float64  `json:"temperature"`
		Gravity          string   `json:"gravity"`
		ClaimedClusters  []string `json:"claimed_clusters"`
		ClaimedBeatCount int      `json:"claimed_beat_count"`
		DominantCluster  string   `json:"dominant_cluster,omitempty"`
		Trend            string   `json:"trend,omitempty"`
	}

	var directories []DirectoryOutput
	for _, dir := range wald.Directories {
		temp, ok := output.Directories[dir.Path]
		if !ok {
			continue
		}
		// Compute dominant cluster (hottest claimed cluster)
		dominantCluster := ""
		maxTemp := 0.0
		for _, clusterID := range temp.ClaimedClusters {
			if ct, ok := output.Clusters[clusterID]; ok {
				if ct.Temperature > maxTemp {
					maxTemp = ct.Temperature
					dominantCluster = clusterID
				}
			}
		}
		directories = append(directories, DirectoryOutput{
			Path:             dir.Path,
			Temperature:      temp.Temperature,
			Gravity:          temp.Gravity,
			ClaimedClusters:  temp.ClaimedClusters,
			ClaimedBeatCount: temp.ClaimedBeatCount,
			DominantCluster:  dominantCluster,
			Trend:            temp.Trend,
		})
	}

	// Build cooperators output
	var cooperators []DirectoryOutput
	for _, dir := range wald.Directories {
		if !strings.HasPrefix(dir.Path, "cooperators/") {
			continue
		}
		temp, ok := output.Directories[dir.Path]
		if !ok {
			continue
		}
		// Compute dominant cluster for cooperator
		dominantCluster := ""
		maxTemp := 0.0
		for _, clusterID := range temp.ClaimedClusters {
			if ct, ok := output.Clusters[clusterID]; ok {
				if ct.Temperature > maxTemp {
					maxTemp = ct.Temperature
					dominantCluster = clusterID
				}
			}
		}
		cooperators = append(cooperators, DirectoryOutput{
			Path:             dir.Path,
			Temperature:      temp.Temperature,
			Gravity:          temp.Gravity,
			ClaimedClusters:  temp.ClaimedClusters,
			ClaimedBeatCount: temp.ClaimedBeatCount,
			DominantCluster:  dominantCluster,
			Trend:            temp.Trend,
		})
	}

	// Output final JSON - clusters FIRST (they are PRIMARY)
	result := map[string]interface{}{
		"computed_at":    output.ComputedAt,
		"beats_analyzed": output.BeatsAnalyzed,
		"aperture":       output.Aperture,
		"clusters":       clustersOutput,
		"directories":    directories,
		"cooperators":    cooperators,
		"aperture_note":  output.ApertureNote,
	}

	outputJSON(result)
}

func robotWALD() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Parse flags
	var stateFilter, gravityFilter string
	for i, arg := range os.Args {
		if arg == "--state" && i+1 < len(os.Args) {
			stateFilter = os.Args[i+1]
		}
		if arg == "--gravity" && i+1 < len(os.Args) {
			gravityFilter = os.Args[i+1]
		}
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load .wald/config.yaml: "+err.Error())
	}

	// Load beats for temperature computation
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	tempOutput := thermal.ComputeTemperature(thermalBeats, clusters, wald.Directories, config, werkRoot)

	// Build directory list
	type DirectoryInfo struct {
		Path        string   `json:"path"`
		Purpose     string   `json:"purpose"`
		Entry       string   `json:"entry,omitempty"`
		State       string   `json:"state"`
		Temperature float64  `json:"temperature"`
		Gravity     string   `json:"gravity"`
		Children    []string `json:"children"`
	}

	var allDirs []DirectoryInfo
	for _, dir := range wald.Directories {
		temp := tempOutput.Directories[dir.Path]
		if temp == nil {
			continue
		}

		gravity := dir.Gravity
		if gravity == "" {
			gravity = "normal"
		}

		info := DirectoryInfo{
			Path:        dir.Path,
			Purpose:     dir.Purpose,
			Entry:       dir.Entry,
			State:       temp.StateInferred,
			Temperature: temp.Temperature,
			Gravity:     gravity,
			Children:    []string{},
		}

		// Find children
		for _, other := range wald.Directories {
			if other.Path != dir.Path && strings.HasPrefix(other.Path, dir.Path+"/") {
				info.Children = append(info.Children, other.Path)
			}
		}

		allDirs = append(allDirs, info)
	}

	// Apply filters
	var filtered []DirectoryInfo
	for _, d := range allDirs {
		if stateFilter != "" && stateFilter != "all" && d.State != stateFilter {
			continue
		}
		if gravityFilter != "" && gravityFilter != "all" && d.Gravity != gravityFilter {
			continue
		}
		filtered = append(filtered, d)
	}

	// Build filter_applied
	filterApplied := map[string]string{
		"state":   stateFilter,
		"gravity": gravityFilter,
	}
	if stateFilter == "" {
		filterApplied["state"] = "all"
	}
	if gravityFilter == "" {
		filterApplied["gravity"] = "all"
	}

	outputJSON(map[string]interface{}{
		"directories":       filtered,
		"filter_applied":    filterApplied,
		"total_matching":    len(filtered),
		"total_directories": len(allDirs),
	})
}

func robotCooperators() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Parse --min-temperature flag
	minTemperature := 0.0
	for i, arg := range os.Args {
		if arg == "--min-temperature" && i+1 < len(os.Args) {
			if f, err := strconv.ParseFloat(os.Args[i+1], 64); err == nil {
				minTemperature = f
			}
		}
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load .wald/config.yaml: "+err.Error())
	}

	// Load all beats from all .beats directories
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	output := thermal.ComputeTemperature(thermalBeats, clusters, wald.Directories, config, werkRoot)

	// Build cooperators output
	type CooperatorOutput struct {
		Path               string   `json:"path"`
		Temperature        float64  `json:"temperature"`
		State              string   `json:"state"`
		BeatCount          int      `json:"beat_count"`
		LastBeat           string   `json:"last_beat,omitempty"`
		LastBeatID         string   `json:"last_beat_id,omitempty"`
		RelatedDirectories []string `json:"related_directories"`
		Trend              string   `json:"trend,omitempty"`
	}

	type InactiveCooperator struct {
		Path      string `json:"path"`
		LastBeat  string `json:"last_beat,omitempty"`
		DaysSince int    `json:"days_since"`
	}

	var activeCooperators []CooperatorOutput
	var inactiveCooperators []InactiveCooperator

	now := time.Now()
	windowDays := 30
	if config != nil {
		windowDays = config.Temperature.WindowDays
	}

	// Find cooperator directories
	for _, dir := range wald.Directories {
		if !strings.HasPrefix(dir.Path, "cooperators/") {
			continue
		}

		temp, ok := output.Directories[dir.Path]
		if !ok {
			continue
		}

		// Find last beat and related directories
		var lastBeatTime time.Time
		var lastBeatID string
		relatedDirs := make(map[string]bool)

		for _, beat := range thermalBeats {
			if beat.Context == nil {
				continue
			}
			// Check if beat belongs to this cooperator directory
			if beat.Context.WALDDirectory == dir.Path || strings.HasPrefix(beat.Context.WALDDirectory, dir.Path+"/") {
				if beat.CreatedAt.After(lastBeatTime) {
					lastBeatTime = beat.CreatedAt
					lastBeatID = beat.ID
				}
			}
			// Also collect beats that mention this cooperator elsewhere (related directories)
			if beat.Context.WALDDirectory != dir.Path && !strings.HasPrefix(beat.Context.WALDDirectory, dir.Path+"/") {
				// Check if beat content mentions cooperator name
				cooperatorName := strings.TrimPrefix(dir.Path, "cooperators/")
				if strings.Contains(strings.ToLower(beat.Content), strings.ToLower(cooperatorName)) {
					relatedDirs[beat.Context.WALDDirectory] = true
					if beat.CreatedAt.After(lastBeatTime) {
						lastBeatTime = beat.CreatedAt
						lastBeatID = beat.ID
					}
				}
			}
		}

		// Convert related dirs to slice
		var relatedDirsList []string
		for d := range relatedDirs {
			relatedDirsList = append(relatedDirsList, d)
		}
		sort.Strings(relatedDirsList)

		// Check if inactive (no beats in window)
		daysSinceLastBeat := int(now.Sub(lastBeatTime).Hours() / 24)
		if lastBeatTime.IsZero() {
			daysSinceLastBeat = -1 // Never had beats
		}

		if temp.ClaimedBeatCount == 0 || (daysSinceLastBeat > windowDays && daysSinceLastBeat != -1) {
			inactive := InactiveCooperator{
				Path:      dir.Path,
				DaysSince: daysSinceLastBeat,
			}
			if !lastBeatTime.IsZero() {
				inactive.LastBeat = lastBeatTime.Format(time.RFC3339)
			}
			inactiveCooperators = append(inactiveCooperators, inactive)
		} else if temp.Temperature >= minTemperature {
			coop := CooperatorOutput{
				Path:               dir.Path,
				Temperature:        temp.Temperature,
				State:              temp.StateInferred,
				BeatCount:          temp.ClaimedBeatCount,
				RelatedDirectories: relatedDirsList,
				Trend:              temp.Trend,
			}
			if !lastBeatTime.IsZero() {
				coop.LastBeat = lastBeatTime.Format(time.RFC3339)
				coop.LastBeatID = lastBeatID
			}
			activeCooperators = append(activeCooperators, coop)
		}
	}

	// Sort active by temperature descending
	sort.Slice(activeCooperators, func(i, j int) bool {
		return activeCooperators[i].Temperature > activeCooperators[j].Temperature
	})

	outputJSON(map[string]interface{}{
		"cooperators": activeCooperators,
		"inactive":    inactiveCooperators,
	})
}
func robotTemporal() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Parse flags
	days := 7
	granularity := "day"
	for i, arg := range os.Args {
		if arg == "--days" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil {
				days = n
			}
		}
		if arg == "--granularity" && i+1 < len(os.Args) {
			g := os.Args[i+1]
			if g == "day" || g == "week" {
				granularity = g
			}
		}
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load WALD.yaml: "+err.Error())
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to load config: "+err.Error())
	}

	// Load all beats
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		fatalJSON("error", "failed to discover beats projects: "+err.Error())
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	now := time.Now()
	endDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	startDate := endDate.AddDate(0, 0, -days+1)

	// Build time buckets
	type TimeBucket struct {
		Date time.Time
		End  time.Time
	}
	var buckets []TimeBucket

	if granularity == "day" {
		for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
			buckets = append(buckets, TimeBucket{Date: d, End: d.AddDate(0, 0, 1)})
		}
	} else {
		// Week granularity - start from Monday
		d := startDate
		for d.Weekday() != time.Monday {
			d = d.AddDate(0, 0, -1)
		}
		for !d.After(endDate) {
			weekEnd := d.AddDate(0, 0, 7)
			buckets = append(buckets, TimeBucket{Date: d, End: weekEnd})
			d = weekEnd
		}
	}

	// Match beats to directories
	beatsByDir := make(map[string][]thermal.Beat)
	for _, beat := range thermalBeats {
		dirPath, _ := thermal.FindMatchingDirectory(beat.Context, wald.Directories, werkRoot)
		if dirPath != "" {
			beatsByDir[dirPath] = append(beatsByDir[dirPath], beat)
		}
	}

	// Build series per directory
	type SeriesPoint struct {
		Date        string  `json:"date"`
		Beats       int     `json:"beats"`
		Temperature float64 `json:"temperature"`
	}
	series := make(map[string][]SeriesPoint)

	// For each directory, compute time series
	for dirPath := range beatsByDir {
		var points []SeriesPoint
		for _, bucket := range buckets {
			// Count beats in this bucket
			beatCount := 0
			for _, beat := range beatsByDir[dirPath] {
				if !beat.CreatedAt.Before(bucket.Date) && beat.CreatedAt.Before(bucket.End) {
					beatCount++
				}
			}

			// Compute temperature up to this bucket's end
			var rawScore float64
			for _, beat := range beatsByDir[dirPath] {
				if beat.CreatedAt.Before(bucket.End) {
					ageDays := bucket.End.Sub(beat.CreatedAt).Hours() / 24
					if ageDays <= float64(config.Temperature.WindowDays) {
						weight := 1.0
						if config.Temperature.RecencyDecayLambda > 0 {
							weight = math.Exp(-config.Temperature.RecencyDecayLambda * ageDays)
						}
						rawScore += weight
					}
				}
			}

			dateStr := bucket.Date.Format("2006-01-02")
			points = append(points, SeriesPoint{
				Date:        dateStr,
				Beats:       beatCount,
				Temperature: rawScore, // Will normalize later
			})
		}
		series[dirPath] = points
	}

	// Normalize temperatures across all directories per bucket
	for i := range buckets {
		maxTemp := 0.0
		for _, points := range series {
			if i < len(points) && points[i].Temperature > maxTemp {
				maxTemp = points[i].Temperature
			}
		}
		if maxTemp > 0 {
			for dirPath := range series {
				if i < len(series[dirPath]) {
					series[dirPath][i].Temperature = series[dirPath][i].Temperature / maxTemp
					// Round to 2 decimal places
					series[dirPath][i].Temperature = math.Round(series[dirPath][i].Temperature*100) / 100
				}
			}
		}
	}

	// Build daily totals
	type DailyTotal struct {
		Date              string `json:"date"`
		TotalBeats        int    `json:"total_beats"`
		ActiveDirectories int    `json:"active_directories"`
	}
	var dailyTotals []DailyTotal

	for i, bucket := range buckets {
		total := 0
		active := 0
		for _, points := range series {
			if i < len(points) {
				total += points[i].Beats
				if points[i].Beats > 0 {
					active++
				}
			}
		}
		dailyTotals = append(dailyTotals, DailyTotal{
			Date:              bucket.Date.Format("2006-01-02"),
			TotalBeats:        total,
			ActiveDirectories: active,
		})
	}

	// Detect patterns
	type Patterns struct {
		MostConsistent string   `json:"most_consistent"`
		MostVariable   string   `json:"most_variable"`
		Emerging       []string `json:"emerging"`
		Fading         []string `json:"fading"`
	}
	patterns := Patterns{
		Emerging: []string{},
		Fading:   []string{},
	}

	// Calculate variance for each directory
	type DirStats struct {
		Path     string
		Variance float64
		FirstSum int
		LastSum  int
	}
	var dirStats []DirStats

	for dirPath, points := range series {
		if len(points) < 2 {
			continue
		}

		// Calculate mean and variance
		sum := 0.0
		for _, p := range points {
			sum += float64(p.Beats)
		}
		mean := sum / float64(len(points))

		variance := 0.0
		for _, p := range points {
			diff := float64(p.Beats) - mean
			variance += diff * diff
		}
		variance /= float64(len(points))

		// Calculate first half vs last half for emerging/fading
		mid := len(points) / 2
		firstSum := 0
		lastSum := 0
		for i, p := range points {
			if i < mid {
				firstSum += p.Beats
			} else {
				lastSum += p.Beats
			}
		}

		dirStats = append(dirStats, DirStats{
			Path:     dirPath,
			Variance: variance,
			FirstSum: firstSum,
			LastSum:  lastSum,
		})
	}

	// Find most consistent (lowest variance) and most variable (highest variance)
	if len(dirStats) > 0 {
		sort.Slice(dirStats, func(i, j int) bool {
			return dirStats[i].Variance < dirStats[j].Variance
		})
		patterns.MostConsistent = dirStats[0].Path
		patterns.MostVariable = dirStats[len(dirStats)-1].Path

		// Emerging: last half > first half
		// Fading: first half > last half
		for _, ds := range dirStats {
			if ds.LastSum > ds.FirstSum && ds.LastSum > 0 {
				patterns.Emerging = append(patterns.Emerging, ds.Path)
			} else if ds.FirstSum > ds.LastSum && ds.FirstSum > 0 {
				patterns.Fading = append(patterns.Fading, ds.Path)
			}
		}
	}

	// Build output
	output := map[string]interface{}{
		"window": map[string]interface{}{
			"start": startDate.Format("2006-01-02"),
			"end":   endDate.Format("2006-01-02"),
			"days":  days,
		},
		"granularity":  granularity,
		"series":       series,
		"daily_totals": dailyTotals,
		"patterns":     patterns,
	}

	outputJSON(output)
}
func robotWatchTemperature() {
	// Parse --interval flag (default 60 seconds)
	interval := 60
	for i, arg := range os.Args {
		if arg == "--interval" && i+1 < len(os.Args) {
			if n, err := strconv.Atoi(os.Args[i+1]); err == nil && n > 0 {
				interval = n
			}
		}
	}

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Emit initial temperature
	emitTemperature()

	for {
		select {
		case <-ticker.C:
			emitTemperature()
		case <-sigChan:
			return
		}
	}
}

func emitTemperature() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		outputJSONL(map[string]interface{}{
			"type":      "error",
			"timestamp": time.Now().Format(time.RFC3339),
			"error":     "WALD.yaml not found - not in a werk directory",
		})
		return
	}

	// Load WALD.yaml
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		outputJSONL(map[string]interface{}{
			"type":      "error",
			"timestamp": time.Now().Format(time.RFC3339),
			"error":     "failed to load WALD.yaml: " + err.Error(),
		})
		return
	}

	// Load config
	config, err := thermal.LoadConfig(werkRoot)
	if err != nil {
		outputJSONL(map[string]interface{}{
			"type":      "error",
			"timestamp": time.Now().Format(time.RFC3339),
			"error":     "failed to load .wald/config.yaml: " + err.Error(),
		})
		return
	}

	// Load all beats
	projects, err := loader.DiscoverProjects(werkRoot)
	if err != nil {
		outputJSONL(map[string]interface{}{
			"type":      "error",
			"timestamp": time.Now().Format(time.RFC3339),
			"error":     "failed to discover beats projects: " + err.Error(),
		})
		return
	}

	var thermalBeats []thermal.Beat
	for _, proj := range projects {
		beats, err := loader.LoadBeats(proj.Path)
		if err != nil {
			continue
		}
		for _, b := range beats {
			tb := thermal.Beat{
				ID:        b.ID,
				CreatedAt: b.CreatedAt,
				Content:   b.Content,
			}
			relPath, _ := filepath.Rel(werkRoot, proj.Path)
			relPath = filepath.Dir(relPath)
			if relPath != "." && relPath != "" {
				tb.Context = &thermal.BeatContext{
					CapturePath:     proj.Path,
					WALDDirectory:   relPath,
					InferenceMethod: "capture_location",
					Confidence:      1.0,
				}
			}
			thermalBeats = append(thermalBeats, tb)
		}
	}

	// Load clusters
	_, cache, _ := getEnrichedBeats()
	var clusters []model.Cluster
	if cache != nil {
		clusters = cache.Clusters
	}

	// Compute temperatures
	output, _ := thermal.ComputeTemperatureWithCache(thermalBeats, clusters, wald.Directories, config, werkRoot)

	// Build directory list
	type DirectoryOutput struct {
		Path          string  `json:"path"`
		Temperature   float64 `json:"temperature"`
		StateInferred string  `json:"state_inferred"`
		StateDeclared string  `json:"state_declared,omitempty"`
		Gravity       string  `json:"gravity"`
		RecentBeats   int     `json:"recent_beats"`
		Trend         string  `json:"trend,omitempty"`
	}

	var directories []DirectoryOutput
	for _, dir := range wald.Directories {
		temp, ok := output.Directories[dir.Path]
		if !ok {
			continue
		}
		directories = append(directories, DirectoryOutput{
			Path:          dir.Path,
			Temperature:   temp.Temperature,
			StateInferred: temp.StateInferred,
			StateDeclared: dir.State,
			Gravity:       temp.Gravity,
			RecentBeats:   temp.ClaimedBeatCount,
			Trend:         temp.Trend,
		})
	}

	// Sort by temperature descending
	sort.Slice(directories, func(i, j int) bool {
		return directories[i].Temperature > directories[j].Temperature
	})

	// Emit JSONL
	outputJSONL(map[string]interface{}{
		"type":        "temperature_update",
		"timestamp":   time.Now().Format(time.RFC3339),
		"directories": directories,
	})
}

func outputJSONL(v interface{}) {
	json.NewEncoder(os.Stdout).Encode(v)
}

func robotWatchBeats() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fatalJSON("error", "failed to create watcher: "+err.Error())
	}
	defer watcher.Close()

	// Track last known line counts for each beats.jsonl file
	lineCountCache := make(map[string]int)

	// Find all .beats directories and watch beats.jsonl files
	var watchedFiles []string
	filepath.Walk(werkRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && info.Name() == ".beats" {
			beatsFile := filepath.Join(path, "beats.jsonl")
			if _, err := os.Stat(beatsFile); err == nil {
				watcher.Add(beatsFile)
				watchedFiles = append(watchedFiles, beatsFile)
				lineCountCache[beatsFile] = countLines(beatsFile)
			}
		}
		return nil
	})

	if len(watchedFiles) == 0 {
		fatalJSON("error", "no beats.jsonl files found under "+werkRoot)
	}

	// Load WALD for temperature impact
	wald, _ := thermal.LoadWALD(werkRoot)
	config, _ := thermal.LoadConfig(werkRoot)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	enc := json.NewEncoder(os.Stdout)

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				// File was written to - check for new beats
				filePath := event.Name
				oldCount := lineCountCache[filePath]
				newCount := countLines(filePath)

				if newCount > oldCount {
					// New beats added - emit them
					newBeats := readNewBeats(filePath, oldCount, newCount)
					lineCountCache[filePath] = newCount

					// Compute context from file path
					relPath, _ := filepath.Rel(werkRoot, filepath.Dir(filePath))
					relPath = filepath.Dir(relPath) // Remove .beats
					if relPath == "." {
						relPath = ""
					}

					for _, beat := range newBeats {
						// Compute temperature impact
						var impactedDirs []string
						if wald != nil {
							for _, dir := range wald.Directories {
								if relPath != "" && (dir.Path == relPath || strings.HasPrefix(dir.Path, relPath+"/") || strings.HasPrefix(relPath, dir.Path+"/")) {
									impactedDirs = append(impactedDirs, dir.Path)
								}
							}
						}

						output := map[string]interface{}{
							"type": "beat_added",
							"beat": beat,
							"context": map[string]interface{}{
								"source_file":    filePath,
								"wald_directory": relPath,
							},
							"temperature_impact": map[string]interface{}{
								"directories": impactedDirs,
								"window_days": getWindowDays(config),
							},
						}
						enc.Encode(output)
					}
				}
			}
		case err := <-watcher.Errors:
			errOutput := map[string]interface{}{
				"type":  "error",
				"error": err.Error(),
			}
			enc.Encode(errOutput)
		case <-sigChan:
			return
		}
	}
}

func countLines(filePath string) int {
	file, err := os.Open(filePath)
	if err != nil {
		return 0
	}
	defer file.Close()

	count := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		count++
	}
	return count
}

func readNewBeats(filePath string, startLine, endLine int) []map[string]interface{} {
	file, err := os.Open(filePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var beats []map[string]interface{}
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		if lineNum >= startLine && lineNum < endLine {
			var beat map[string]interface{}
			if err := json.Unmarshal(scanner.Bytes(), &beat); err == nil {
				beats = append(beats, beat)
			}
		}
		lineNum++
	}
	return beats
}

func getWindowDays(config *thermal.Config) int {
	if config != nil {
		return config.Temperature.WindowDays
	}
	return 30
}

// runClaim handles `btv claim PATH --cluster/--topic/--keyword/--cooperator NAME`
func runClaim() {
	args := os.Args[2:]
	if len(args) < 3 {
		fatal("claim requires PATH and one of --cluster/--topic/--keyword/--cooperator NAME")
	}

	path := args[0]
	var claimType, claimValue string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--cluster":
			if i+1 < len(args) {
				claimType, claimValue = "cluster", args[i+1]
				i++
			}
		case "--topic":
			if i+1 < len(args) {
				claimType, claimValue = "topic", args[i+1]
				i++
			}
		case "--keyword":
			if i+1 < len(args) {
				claimType, claimValue = "keyword", args[i+1]
				i++
			}
		case "--cooperator":
			if i+1 < len(args) {
				claimType, claimValue = "cooperator", args[i+1]
				i++
			}
		}
	}

	if claimType == "" || claimValue == "" {
		fatal("claim requires one of --cluster/--topic/--keyword/--cooperator NAME")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatal("WALD.yaml not found - not in a werk directory")
	}

	claims, err := modifyClaim(werkRoot, path, claimType, claimValue, true)
	if err != nil {
		fatal(err.Error())
	}

	fmt.Printf("Added %s claim '%s' to %s\n", claimType, claimValue, path)
	fmt.Printf("Current claims: clusters=%v topics=%v keywords=%v cooperators=%v\n",
		claims.Clusters, claims.Topics, claims.Keywords, claims.Cooperators)
}

// runUnclaim handles `btv unclaim PATH --cluster/--topic/--keyword/--cooperator NAME`
func runUnclaim() {
	args := os.Args[2:]
	if len(args) < 3 {
		fatal("unclaim requires PATH and one of --cluster/--topic/--keyword/--cooperator NAME")
	}

	path := args[0]
	var claimType, claimValue string

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--cluster":
			if i+1 < len(args) {
				claimType, claimValue = "cluster", args[i+1]
				i++
			}
		case "--topic":
			if i+1 < len(args) {
				claimType, claimValue = "topic", args[i+1]
				i++
			}
		case "--keyword":
			if i+1 < len(args) {
				claimType, claimValue = "keyword", args[i+1]
				i++
			}
		case "--cooperator":
			if i+1 < len(args) {
				claimType, claimValue = "cooperator", args[i+1]
				i++
			}
		}
	}

	if claimType == "" || claimValue == "" {
		fatal("unclaim requires one of --cluster/--topic/--keyword/--cooperator NAME")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatal("WALD.yaml not found - not in a werk directory")
	}

	claims, err := modifyClaim(werkRoot, path, claimType, claimValue, false)
	if err != nil {
		fatal(err.Error())
	}

	fmt.Printf("Removed %s claim '%s' from %s\n", claimType, claimValue, path)
	fmt.Printf("Current claims: clusters=%v topics=%v keywords=%v cooperators=%v\n",
		claims.Clusters, claims.Topics, claims.Keywords, claims.Cooperators)
}

// robotClaim handles `echo '{"path":"...","cluster":"..."}' | btv --robot-claim`
func robotClaim() {
	var input struct {
		Path       string `json:"path"`
		Cluster    string `json:"cluster,omitempty"`
		Topic      string `json:"topic,omitempty"`
		Keyword    string `json:"keyword,omitempty"`
		Cooperator string `json:"cooperator,omitempty"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	if input.Path == "" {
		fatalJSON("error", "path is required")
	}

	var claimType, claimValue string
	switch {
	case input.Cluster != "":
		claimType, claimValue = "cluster", input.Cluster
	case input.Topic != "":
		claimType, claimValue = "topic", input.Topic
	case input.Keyword != "":
		claimType, claimValue = "keyword", input.Keyword
	case input.Cooperator != "":
		claimType, claimValue = "cooperator", input.Cooperator
	default:
		fatalJSON("error", "one of cluster/topic/keyword/cooperator is required")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	claims, err := modifyClaim(werkRoot, input.Path, claimType, claimValue, true)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	outputJSON(map[string]interface{}{
		"success":   true,
		"directory": input.Path,
		"action":    "claim",
		"type":      claimType,
		"value":     claimValue,
		"claims": map[string]interface{}{
			"clusters":    claims.Clusters,
			"topics":      claims.Topics,
			"keywords":    claims.Keywords,
			"cooperators": claims.Cooperators,
		},
	})
}

// robotUnclaim handles `echo '{"path":"...","topic":"..."}' | btv --robot-unclaim`
func robotUnclaim() {
	var input struct {
		Path       string `json:"path"`
		Cluster    string `json:"cluster,omitempty"`
		Topic      string `json:"topic,omitempty"`
		Keyword    string `json:"keyword,omitempty"`
		Cooperator string `json:"cooperator,omitempty"`
	}

	if err := json.NewDecoder(os.Stdin).Decode(&input); err != nil {
		fatalJSON("error", "invalid JSON input: "+err.Error())
	}

	if input.Path == "" {
		fatalJSON("error", "path is required")
	}

	var claimType, claimValue string
	switch {
	case input.Cluster != "":
		claimType, claimValue = "cluster", input.Cluster
	case input.Topic != "":
		claimType, claimValue = "topic", input.Topic
	case input.Keyword != "":
		claimType, claimValue = "keyword", input.Keyword
	case input.Cooperator != "":
		claimType, claimValue = "cooperator", input.Cooperator
	default:
		fatalJSON("error", "one of cluster/topic/keyword/cooperator is required")
	}

	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	claims, err := modifyClaim(werkRoot, input.Path, claimType, claimValue, false)
	if err != nil {
		fatalJSON("error", err.Error())
	}

	outputJSON(map[string]interface{}{
		"success":   true,
		"directory": input.Path,
		"action":    "unclaim",
		"type":      claimType,
		"value":     claimValue,
		"claims": map[string]interface{}{
			"clusters":    claims.Clusters,
			"topics":      claims.Topics,
			"keywords":    claims.Keywords,
			"cooperators": claims.Cooperators,
		},
	})
}

// modifyClaim adds or removes a claim from a directory's WALD.yaml entry
func modifyClaim(werkRoot, path, claimType, claimValue string, add bool) (*thermal.DirectoryClaims, error) {
	wald, err := thermal.LoadWALD(werkRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to load WALD.yaml: %w", err)
	}

	var dirIndex = -1
	for i, dir := range wald.Directories {
		if dir.Path == path {
			dirIndex = i
			break
		}
	}

	if dirIndex == -1 {
		return nil, fmt.Errorf("directory not found in WALD.yaml: %s", path)
	}

	claims := &wald.Directories[dirIndex].Claims

	modifySlice := func(slice []string, value string, add bool) []string {
		if add {
			for _, v := range slice {
				if v == value {
					return slice
				}
			}
			return append(slice, value)
		}
		var result []string
		for _, v := range slice {
			if v != value {
				result = append(result, v)
			}
		}
		return result
	}

	switch claimType {
	case "cluster":
		claims.Clusters = modifySlice(claims.Clusters, claimValue, add)
	case "topic":
		claims.Topics = modifySlice(claims.Topics, claimValue, add)
	case "keyword":
		claims.Keywords = modifySlice(claims.Keywords, claimValue, add)
	case "cooperator":
		claims.Cooperators = modifySlice(claims.Cooperators, claimValue, add)
	default:
		return nil, fmt.Errorf("unknown claim type: %s", claimType)
	}

	if err := thermal.SaveWALD(werkRoot, wald); err != nil {
		return nil, fmt.Errorf("failed to save WALD.yaml: %w", err)
	}

	return claims, nil
}

// runSuggestClaims handles the `btv suggest-claims` command
func runSuggestClaims() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatal("WALD.yaml not found - not in a werk directory")
	}

	// Parse flags
	robotMode := false
	applyMode := false
	for _, arg := range os.Args[2:] {
		if arg == "--robot" {
			robotMode = true
		}
		if arg == "--apply" {
			applyMode = true
		}
	}

	// Load clusters from cache
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatal("failed to load beats: " + err.Error())
	}

	output, err := thermal.SuggestClaimsFromLegacy(werkRoot, cache.Clusters)
	if err != nil {
		fatal("failed to generate suggestions: " + err.Error())
	}

	if robotMode {
		outputJSON(output)
		return
	}

	// Human output
	thermal.PrintSuggestClaimsHuman(output)

	if len(output.Suggestions) == 0 {
		fmt.Println("No suggestions found. No beats with legacy context.")
		return
	}

	if applyMode {
		// Apply without prompting
		if err := thermal.ApplySuggestedClaims(werkRoot, output.Suggestions); err != nil {
			fatal("failed to apply suggestions: " + err.Error())
		}
		fmt.Println("Applied suggested claims to WALD.yaml")
		return
	}

	// Prompt for confirmation
	fmt.Print("Apply suggested claims? [y/N/review] ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(response)

	if response == "review" {
		outputJSON(output)
		fmt.Print("\nApply suggested claims? [y/N] ")
		fmt.Scanln(&response)
		response = strings.ToLower(response)
	}

	if response == "y" || response == "yes" {
		if err := thermal.ApplySuggestedClaims(werkRoot, output.Suggestions); err != nil {
			fatal("failed to apply suggestions: " + err.Error())
		}
		fmt.Println("Applied suggested claims to WALD.yaml")
	} else {
		fmt.Println("Aborted.")
	}
}

// robotSuggestClaims handles --robot-suggest-claims
func robotSuggestClaims() {
	werkRoot := thermal.FindWerkRoot()
	if werkRoot == "" {
		fatalJSON("error", "WALD.yaml not found - not in a werk directory")
	}

	// Load clusters from cache
	_, cache, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", "failed to load beats: "+err.Error())
	}

	output, err := thermal.SuggestClaimsFromLegacy(werkRoot, cache.Clusters)
	if err != nil {
		fatalJSON("error", "failed to generate suggestions: "+err.Error())
	}

	outputJSON(output)
}
