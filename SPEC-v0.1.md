# beats_viewer v0.1 Specification

## Vision

beats_viewer (btv) is the TUI companion to beats (bt), providing interactive browsing, filtering, and synthesis across the narrative layer that feeds beads. Where bv provides graph-aware triage for actionable work, btv provides narrative-aware discovery for pre-actionable insights.

**Core insight:** Beats are the psychoid buffer between discovery and action. btv helps surface patterns, connections, and themes that are ready to crystallize into beads.

## Scope: v0.1 (MVP)

Focus on **read-only exploration** across projects. No editing, no synthesis, no AI integration in v0.1.

### Must Have

1. **Cross-project beat browsing**
   - List beats from current project or all projects (`bt projects` equivalent)
   - Display: content preview, impetus, created date, project name
   - Scrollable list with vim-style navigation (j/k/gg/G)

2. **Full-text search within TUI**
   - Search bar (/) that filters displayed beats in real-time
   - Uses existing `bt search` infrastructure
   - Cross-project search toggle

3. **Beat detail view**
   - Full content display in viewport
   - Metadata: id, created_at, updated_at, impetus (label + meta)
   - Linked beads (if any)
   - References/entities (if present in beat)

4. **Filtering**
   - By project (p to cycle, or picker)
   - By impetus label
   - By date range (optional stretch)

5. **Robot commands**
   - `--robot-list` - JSON list of beats (supports project filter)
   - `--robot-search` - JSON search results
   - `--robot-show <id>` - JSON single beat detail

### Won't Have (v0.1)

- Beat creation/editing (use `bt add`)
- AI-powered synthesis/briefing (future: `--robot-brief`)
- Beat-to-bead mapping suggestions
- Theme extraction
- Graph visualization of beat relationships

## Data Model

### Beat (from beats.jsonl)

```json
{
  "id": "beat-20251204-001",
  "created_at": "2025-12-04T18:48:43Z",
  "updated_at": "2025-12-04T18:48:43Z",
  "impetus": {
    "label": "Nick coaching call #7",
    "raw": "optional raw source text",
    "meta": {
      "channel": "coaching",
      "counterparty": "Nick"
    }
  },
  "content": "Commitment is fundamentally about identity...",
  "entities": ["Nick", "identity"],
  "references": [],
  "linked_beads": ["bv-42"]
}
```

### Project

```go
type Project struct {
    Name      string // directory name
    Path      string // absolute path to .beats/
    BeatCount int
}
```

## Architecture

```
btv
├── cmd/btv/
│   └── main.go              # CLI entry, robot flags, TUI launch
├── pkg/
│   ├── loader/
│   │   └── loader.go        # Load beats from .beats/beats.jsonl
│   ├── model/
│   │   └── types.go         # Beat, Project types
│   ├── search/
│   │   └── search.go        # FTS wrapper (delegates to bt --robot-search)
│   └── ui/
│       ├── model.go         # Bubbletea model
│       ├── list.go          # Beat list component
│       ├── detail.go        # Beat detail viewport
│       ├── search.go        # Search input component
│       └── styles.go        # Lipgloss styles
├── go.mod
└── go.sum
```

### Dependencies

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - List, viewport, textinput components

No external AI/embedding dependencies in v0.1.

## UI Layout

```
┌─────────────────────────────────────────────────────────────────────┐
│ btv v0.1.0                            [werk] 255 beats   /: search │
├─────────────────────────────────────────────────────────────────────┤
│ ▸ beat-20251204-001  Nick coaching     Commitment is fundament...   │
│   beat-20251204-002  Nick coaching     AI agents need persiste...   │
│   beat-20251227-001  droid-session     Set up GitHub Actions w...   │
│   ...                                                               │
├─────────────────────────────────────────────────────────────────────┤
│ DETAIL                                                              │
│ ───────────────────────────────────────────────────────────────────│
│ ID: beat-20251204-001                                               │
│ Created: 2025-12-04 18:48                                           │
│ Impetus: Nick coaching call #7                                      │
│                                                                     │
│ Commitment is fundamentally about identity, not discipline.         │
│ When I frame something as 'I am the kind of person who does X',    │
│ it bypasses willpower entirely.                                     │
│                                                                     │
│ Linked beads: (none)                                                │
├─────────────────────────────────────────────────────────────────────┤
│ j/k:nav  /:search  p:project  enter:detail  q:quit  ?:help         │
└─────────────────────────────────────────────────────────────────────┘
```

### Adaptive Layout

- **Narrow (<100 cols):** List only, detail on Enter (full screen)
- **Normal (100-140):** Split view, list + detail side by side
- **Wide (>140):** Split view with more content visible

## Keybindings

| Key | Action |
|-----|--------|
| `j/↓` | Next beat |
| `k/↑` | Previous beat |
| `gg` | First beat |
| `G` | Last beat |
| `Enter` | Toggle detail focus / expand |
| `/` | Focus search input |
| `Esc` | Clear search / exit detail |
| `p` | Cycle project filter |
| `P` | Open project picker |
| `a` | Toggle all-projects mode |
| `r` | Refresh beats |
| `y` | Yank beat ID to clipboard |
| `Y` | Yank beat content to clipboard |
| `?` | Help overlay |
| `q` | Quit |

## Robot Commands

### --robot-list

```bash
btv --robot-list [--project <name>] [--limit N]
```

Output:
```json
{
  "beats": [
    {
      "id": "beat-20251204-001",
      "content_preview": "Commitment is fundamentally...",
      "impetus_label": "Nick coaching call #7",
      "project": "werk",
      "created_at": "2025-12-04T18:48:43Z"
    }
  ],
  "total": 255,
  "project_filter": null
}
```

### --robot-search

```bash
echo '{"query":"coaching","all_projects":true}' | btv --robot-search
```

Output:
```json
{
  "results": [...],
  "query": "coaching",
  "total_matches": 12
}
```

### --robot-show

```bash
btv --robot-show beat-20251204-001
```

Output: Full beat JSON

### --robot-help

```bash
btv --robot-help
```

Output: JSON schema of all robot commands (following bv pattern)

## Implementation Plan

### Phase 1: Core loader + types (Day 1)
- [ ] `pkg/model/types.go` - Beat, Project structs
- [ ] `pkg/loader/loader.go` - Load from beats.jsonl, discover projects
- [ ] Unit tests for loader

### Phase 2: Basic TUI (Day 2)
- [ ] `pkg/ui/model.go` - Bubbletea model with focus states
- [ ] `pkg/ui/list.go` - Beat list using bubbles/list
- [ ] `pkg/ui/styles.go` - Lipgloss theming
- [ ] Basic navigation (j/k/gg/G)

### Phase 3: Detail view + search (Day 3)
- [ ] `pkg/ui/detail.go` - Beat detail viewport
- [ ] `pkg/ui/search.go` - Search input integration
- [ ] Split view layout

### Phase 4: Project filtering + polish (Day 4)
- [ ] Project picker/cycler
- [ ] All-projects toggle
- [ ] Help overlay
- [ ] Clipboard yank

### Phase 5: Robot commands (Day 5)
- [ ] `cmd/btv/main.go` - CLI with robot flags
- [ ] JSON output formatting
- [ ] Integration tests

## Success Criteria

1. `btv` launches and displays beats from current project
2. `/coaching` filters to matching beats instantly
3. `p` cycles through 5 projects smoothly
4. `btv --robot-list` returns valid JSON
5. Performance: <100ms startup with 255 beats

## Future (v0.2+)

- **Theme synthesis:** Select beats → generate brief (AI)
- **Beat-to-bead bridge:** Suggest which beats are ready to become beads
- **Graph view:** Visualize entity/reference connections
- **Impetus taxonomy:** Browse by impetus hierarchy
- **Timeline view:** Beats on temporal axis
- **Session correlation:** Link beats to droid sessions via cass
