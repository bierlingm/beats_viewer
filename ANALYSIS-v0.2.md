# beats_viewer v0.2 Analysis & Vision

## Current State (v0.1)

**What exists:** 1,391 LOC Go TUI that browses 240 beats across 5 projects with search, project filtering, and robot commands.

**What's actually being used:**
- 133/240 beats (55%) have generic "Manual entry" impetus - zero semantic signal
- Only 14 beats (6%) have linked_beads - the bridge to action is barely used
- Impetus labels are freeform, no taxonomy - can't filter/group meaningfully
- Content is plain text blobs - no structure, no entities extracted

**Honest assessment:** v0.1 is a JSONL viewer with vim bindings. It doesn't understand beats.

---

## The Core Problem

Beats are supposed to be **"the psychoid buffer between discovery and action"** - narrative substrate that eventually crystallizes into beads. But btv treats them as flat text to scroll through.

**What's missing:**
1. No way to see patterns emerge across beats
2. No way to identify which beats are "ripe" for conversion to beads
3. No semantic understanding of content
4. No temporal awareness (what's fresh vs. stale)
5. No relationship mapping between beats

---

## World-Class Vision for v0.2

### Principle: btv should be a **synthesis engine**, not a viewer

The magic happens when you can:
1. See themes emerge from chaos
2. Identify connections you didn't know existed
3. Surface beats that are ready to become work
4. Track the narrative arc of your thinking

---

## Proposed Features (Ranked by Impact)

### Tier 1: Transformative

#### 1. **Impetus Taxonomy + Faceted Navigation**
Instead of freeform labels, establish a two-level taxonomy:
```
Channel: coaching | research | discovery | development | reflection
Source: twitter | github | web | conversation | session | book
```
- Auto-classify existing "Manual entry" beats on first run
- Faceted sidebar: click coaching → see all coaching beats
- Visual: colored badges by channel

#### 2. **Theme Clustering (Local AI)**
Use Ollama embeddings to cluster beats by semantic similarity:
```
┌─ Cluster: "Identity & Commitment" (12 beats) ──────────────┐
│ • Commitment is about identity, not discipline (coaching)   │
│ • "I am the kind of person who..." framing (coaching)       │
│ • Identity-based habit formation research (web)             │
└────────────────────────────────────────────────────────────┘
```
- No cloud API needed - runs locally
- Show clusters as collapsible groups
- Suggest cluster names based on content

#### 3. **Ripeness Score: "Ready to Become a Bead"**
Calculate which beats are candidates for action:
```go
ripeness := (age_days * 0.2) + 
            (related_beats * 0.3) + 
            (contains_action_language * 0.3) +
            (revisit_count * 0.2)
```
- Beats with high ripeness bubble to top
- Visual indicator: 🟢 ripe / 🟡 maturing / ⚪ fresh
- One-key action: `b` to create bead from beat

#### 4. **Timeline View**
Toggle between list and timeline:
```
Dec 2025 ──●──●────●───●●●──────●─────●●────●──●──
           │  │    │   ╰─coaching cluster
           │  ╰─discovery
           ╰─milestone
```
- See density of activity
- Spot gaps and bursts
- Filter by impetus shows different patterns

### Tier 2: Significant

#### 5. **Entity Extraction**
Auto-extract and display entities from beat content:
- People: Nick, Claude, DHH
- Tools: Supabase, Ollama, beads
- Concepts: commitment, identity, narrative substrate
- Filter by entity: "Show all beats mentioning Supabase"

#### 6. **Beat Chains / Threads**
Connect related beats explicitly:
- `c` to chain current beat to another
- Show chains in detail view: "Part of: AI Memory Research (5 beats)"
- Chains suggest when they're ready to become an epic

#### 7. **Quick Capture Integration**
`btv --capture` mode: minimal UI for rapid beat entry
- Single text field, auto-impetus detection
- Keyboard-first: type → enter → done
- Integrates with Alfred/Raycast

#### 8. **Stale Beat Review**
Surface beats that haven't been touched or linked:
- "42 beats older than 30 days with no links"
- Batch actions: archive, link, convert to bead
- Prevent beat graveyard accumulation

### Tier 3: Polish

#### 9. **Split Horizontal/Vertical Toggle**
`v` to toggle layout orientation based on terminal shape

#### 10. **Beat Templates**
Quick-add with structure:
```
btv --add-coaching    # Pre-fills impetus, prompts for insight
btv --add-discovery   # Pre-fills channel=web, asks for URL
```

#### 11. **Export: Beat → Markdown Brief**
Select multiple beats → export as themed markdown document
- For sharing insights with humans
- Structured output for docs/notes

#### 12. **Keyboard Macro Recording**
Record and replay navigation patterns for repetitive workflows

---

## Technical Architecture for v0.2

```
btv v0.2
├── pkg/
│   ├── model/           # Beat, Project, Cluster, Entity types
│   ├── loader/          # JSONL loading, migration
│   ├── taxonomy/        # Impetus classification, auto-tagging
│   ├── cluster/         # Ollama embeddings, k-means clustering
│   ├── ripeness/        # Scoring algorithm
│   ├── entity/          # NER extraction (local, regex + heuristics)
│   ├── chain/           # Beat relationships
│   └── ui/
│       ├── views/       # list, timeline, cluster, detail
│       ├── components/  # facets, search, statusbar
│       └── actions/     # capture, convert, chain
├── cmd/btv/
│   └── main.go          # CLI with subcommands
└── migrations/          # One-time data enrichment scripts
```

### New Dependencies
- `github.com/ollama/ollama` - Local embeddings (optional, graceful fallback)
- No cloud APIs, no API keys, fully offline-capable

---

## Migration Path

**v0.1 → v0.2 first run:**
1. Scan all beats
2. Auto-classify "Manual entry" impetus based on content heuristics
3. Extract entities
4. Generate initial embeddings (if Ollama available)
5. Calculate ripeness scores
6. Store enriched data in `.beats/btv-cache.json`

---

## Success Metrics for v0.2

1. **< 5 seconds** to find thematically related beats
2. **One keypress** to convert ripe beat to bead
3. **Zero "Manual entry"** impetus labels after migration
4. **Cluster view** surfaces connections user didn't consciously make
5. **Timeline** reveals thinking patterns over time

---

## What Makes This World-Class

Most "note viewers" are glorified grep. btv v0.2 would be:

1. **Synthesis-first** - Surfaces patterns, not just text
2. **Action-oriented** - Knows when thinking is ready to become doing
3. **Temporally aware** - Understands freshness and maturation
4. **Semantically intelligent** - Groups by meaning, not just keywords
5. **Agent-native** - Robot commands for AI orchestration
6. **Privacy-preserving** - All AI runs locally via Ollama

The vision: **btv is where raw thinking becomes structured action.**

---

## Recommended v0.2 Scope

Given effort/impact, prioritize:

1. ✅ Impetus taxonomy + auto-classification (high impact, medium effort)
2. ✅ Ripeness scoring + ripe-first sorting (high impact, low effort)  
3. ✅ Entity extraction (medium impact, medium effort)
4. ✅ Timeline view (high impact, medium effort)
5. 🔄 Theme clustering (very high impact, high effort - needs Ollama)

Ship 1-4 as v0.2.0, add clustering as v0.2.1 when Ollama integration is solid.
