package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/cluster"
	"github.com/bierlingm/beats_viewer/pkg/loader"
	"github.com/bierlingm/beats_viewer/pkg/model"
	"github.com/bierlingm/beats_viewer/pkg/ripeness"
	"github.com/bierlingm/beats_viewer/pkg/timeline"
	"github.com/bierlingm/beats_viewer/pkg/ui"
	"github.com/bierlingm/beats_viewer/pkg/ui/views"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.2.0"

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
			robotClusters()
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
		case "--rebuild-cache":
			rebuildCache()
			return
		case "--capture":
			runCapture()
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
	fmt.Print(`btv - beats_viewer v0.2 TUI

USAGE:
  btv [options]

OPTIONS:
  --root <path>       Root directory for beats discovery (default: current dir)
  --rebuild-cache     Force rebuild of btv cache
  -v, --version       Show version
  -h, --help          Show this help

ROBOT COMMANDS (JSON output for AI agents):
  --robot-help                  Show robot command schemas
  --robot-list                  List all beats as JSON
  --robot-search                Search beats (reads JSON from stdin)
  --robot-show <beat-id>        Show single beat as JSON
  --robot-taxonomy-stats        Channel/source distribution
  --robot-ripeness <beat-id>    Get ripeness score breakdown
  --robot-ripe                  List ripest beats
  --robot-stale                 List stale beats with reasons
  --robot-entities              List all extracted entities
  --robot-timeline              Timeline data by zoom level
  --robot-clusters              List theme clusters

ENVIRONMENT:
  BEATS_ROOT        Override default root directory

KEYBINDINGS (in TUI):
  j/k       Navigate up/down
  /         Search
  f/e       Toggle facet/entity sidebar
  t/C/S     Timeline/Cluster/Stale review views
  R         Sort by ripeness
  p/a       Cycle projects / All projects
  y/Y       Copy ID/content
  ?         Help
  q         Quit
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
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
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
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	limit := 10
	threshold := 0.0
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
	}

	var filtered []model.EnrichedBeat
	for _, eb := range enriched {
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
	enriched, _, err := getEnrichedBeats()
	if err != nil {
		fatalJSON("error", err.Error())
	}

	stale := views.FindStaleBeats(enriched)

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
