package main

import (
	"encoding/json"
	"fmt"
	"os"

	"beats_viewer/pkg/loader"
	"beats_viewer/pkg/model"
	"beats_viewer/pkg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

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
		}
	}

	rootPath := loader.GetDefaultRoot()
	if envRoot := os.Getenv("BEATS_ROOT"); envRoot != "" {
		rootPath = envRoot
	}

	if len(os.Args) > 1 && os.Args[1] == "--root" && len(os.Args) > 2 {
		rootPath = os.Args[2]
	}

	m := ui.NewModel(rootPath)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Print(`btv - beats_viewer TUI

USAGE:
  btv [options]

OPTIONS:
  --root <path>     Root directory for beats discovery (default: ~/werk)
  -v, --version     Show version
  -h, --help        Show this help

ROBOT COMMANDS (JSON output for AI agents):
  --robot-help                  Show robot command schemas
  --robot-list                  List all beats as JSON
  --robot-search                Search beats (reads JSON from stdin)
  --robot-show <beat-id>        Show single beat as JSON

ENVIRONMENT:
  BEATS_ROOT        Override default root directory

KEYBINDINGS (in TUI):
  j/k       Navigate up/down
  /         Search
  p         Cycle projects
  a         Toggle all-projects
  y/Y       Copy ID/content
  ?         Help
  q         Quit
`)
}

func robotHelp() {
	resp := model.RobotHelpResponse{
		Version: version,
		Commands: []model.RobotHelpCommand{
			{
				Name:        "--robot-list",
				Description: "List all beats with optional project filter",
				Input:       "none (use --project flag)",
				Output:      `{"beats": [...], "total": N, "project_filter": "name"|null}`,
			},
			{
				Name:        "--robot-search",
				Description: "Search beats by content/impetus",
				Input:       `{"query": "search term", "all_projects": true|false}`,
				Output:      `{"results": [...], "query": "...", "total_matches": N}`,
			},
			{
				Name:        "--robot-show",
				Description: "Get full details of a single beat",
				Input:       "beat ID as argument",
				Output:      "Full beat JSON object",
			},
			{
				Name:        "--robot-help",
				Description: "Show this help",
				Output:      "This JSON schema",
			},
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
