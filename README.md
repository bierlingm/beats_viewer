# btv - beats_viewer

[![CI](https://github.com/bierlingm/beats_viewer/actions/workflows/ci.yml/badge.svg)](https://github.com/bierlingm/beats_viewer/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/bierlingm/beats_viewer)](https://goreportcard.com/report/github.com/bierlingm/beats_viewer)

TUI for browsing [beats](https://github.com/bierlingm/beats) - a narrative synthesis engine that surfaces patterns in your thinking.

## Relationship to beats

**[beats](https://github.com/bierlingm/beats)** is the CLI for capturing insights, discoveries, and reflections. It writes to `.beats/beats.jsonl`.

**btv** is the TUI for browsing, analyzing, and acting on those beats. It adds:
- Taxonomy classification (Channel + Source)
- Ripeness scoring (which beats are ready for action?)
- Entity extraction (people, tools, concepts)
- Timeline visualization
- Theme clustering (requires Ollama)
- Stale beat review

Use `beats` to capture. Use `btv` to synthesize.

## What's New in v0.4

### Project Filtering
Filter beats by project in TUI and robot commands:
```bash
btv --robot-list --project modernminuteman.net
btv --robot-ripe --project .
```

TUI keybindings:
- `p` - Cycle through projects
- `a` - Toggle all-projects mode

### Session Filtering
```bash
btv --robot-list --session abc123
```

### Crystallization Suggestions
Surface ripe beats ready to become beads:
```bash
btv --robot-crystallize
btv --robot-crystallize --project .
```

TUI: Press `B` on a ripe beat to create a bead.

### Semantic Similarity
Find conceptually similar beats:
```bash
btv --robot-similar beat-20260113-001
```

### Incremental Cache
Cache updates incrementally (<100ms for single beat changes).
Force full rebuild with `--rebuild-cache`.

## Installation

**With Go:**
```bash
go install github.com/bierlingm/beats_viewer/cmd/btv@latest
```

**From source:**
```bash
git clone https://github.com/bierlingm/beats_viewer
cd beats_viewer && go build -o btv ./cmd/btv
```

## Quick Start

```bash
# Navigate to a directory with .beats/
cd ~/my-project

# Launch TUI
btv

# Or specify root directory
btv --root ~/notes
```

## Keybindings

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `/` | Search |
| `p` | Cycle through projects |
| `a` | Toggle all-projects mode |
| `B` | Create bead from beat (if ripe >= 0.5) |
| `R` | Sort by ripeness |
| `y/Y` | Copy beat ID / content |
| `tab` | Switch focus list/detail |
| `?` | Help |
| `q` | Quit |

## Views

### List View (default)
Browse beats with ripeness indicators:
- ⚪ Fresh (< 0.3)
- 🟡 Maturing (0.3-0.6)
- 🟢 Ripe (0.6-0.8) - ready for action
- 🔴 Overripe (> 0.8) - act or archive

### Timeline View (`t`)
Visualize beat density over time. Navigate with arrow keys, zoom with `z`.

### Cluster View (`C`)
Theme groupings via semantic clustering. Requires [Ollama](https://ollama.ai) with `nomic-embed-text` model.

### Stale Review (`S`)
Process beats needing attention. Actions: Keep, Archive, Convert to bead, Add to chain, Delete.

## Robot Commands

For AI agent integration, all output JSON:

```bash
btv --robot-list                  # List all beats
btv --robot-show <beat-id>        # Show beat details
btv --robot-stale                 # List stale beats with reasons
btv --robot-ripeness <beat-id>    # Get ripeness breakdown
btv --robot-ripe                  # List ripest beats
btv --robot-taxonomy-stats        # Channel/source distribution
btv --robot-entities              # List extracted entities
btv --robot-timeline              # Timeline data
btv --robot-clusters              # Theme clusters
btv --rebuild-cache               # Force cache rebuild
```

## Configuration

| Env Variable | Description |
|--------------|-------------|
| `BEATS_ROOT` | Root directory for beats discovery |

## Responsive Layout

btv adapts to terminal width:
- **Compact** (< 60 cols): Single column
- **Normal** (100+): List + detail split
- **Wide** (140+): + one sidebar
- **UltraWide** (180+): All panels

## License

MIT - see [LICENSE](LICENSE)
