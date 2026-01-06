# beats_viewer v0.2 Specification

## Executive Summary

beats_viewer v0.2 transforms from a JSONL browser into a **narrative synthesis engine**. It understands beats semantically, surfaces patterns across your thinking, identifies when insights are ready to become action, and provides temporal awareness of your intellectual activity.

**Core thesis:** Beats are the psychoid buffer between discovery and action. btv should accelerate the crystallization of raw thinking into structured work.

---

## Design Principles

1. **Synthesis over browsing** - Surface patterns, don't just display text
2. **Action-oriented** - Every view answers "what should I do with this?"
3. **Temporally aware** - Fresh vs. mature vs. stale beats behave differently
4. **Semantically intelligent** - Group by meaning, not just string matching
5. **Zero-config intelligence** - Works immediately, improves with Ollama
6. **Agent-native** - Every feature has a robot command equivalent
7. **Privacy-first** - All AI runs locally, no cloud dependencies

---

## Data Model Extensions

### Beat (Extended)

```go
type Beat struct {
    // Existing v0.1 fields
    ID          string    `json:"id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Impetus     Impetus   `json:"impetus"`
    Content     string    `json:"content"`
    Entities    []string  `json:"entities,omitempty"`
    References  []string  `json:"references,omitempty"`
    LinkedBeads []string  `json:"linked_beads,omitempty"`
    
    // v0.2 computed fields (stored in cache, not source JSONL)
    Taxonomy      Taxonomy        `json:"-"` // Classified impetus
    ExtractedEntities []Entity   `json:"-"` // Auto-extracted
    RipenessScore float64        `json:"-"` // 0.0-1.0
    ClusterID     string         `json:"-"` // Theme cluster assignment
    ChainIDs      []string       `json:"-"` // Linked beat chains
    ViewCount     int            `json:"-"` // Times viewed in btv
    LastViewedAt  *time.Time     `json:"-"` // Recency tracking
}
```

### Taxonomy (New)

```go
type Taxonomy struct {
    Channel  Channel  `json:"channel"`  // Primary classification
    Source   Source   `json:"source"`   // Origin type
    Confidence float64 `json:"confidence"` // Classification confidence
}

type Channel int
const (
    ChannelUnknown Channel = iota
    ChannelCoaching     // Insights from coaching/mentoring
    ChannelResearch     // Deliberate investigation
    ChannelDiscovery    // Serendipitous finding
    ChannelDevelopment  // Building/coding insight
    ChannelReflection   // Personal synthesis
    ChannelReference    // Saved for later use
    ChannelMilestone    // Achievement/completion
)

type Source int
const (
    SourceUnknown Source = iota
    SourceConversation  // Human dialogue
    SourceWeb           // Browser discovery
    SourceTwitter       // X/Twitter
    SourceGitHub        // Code/issues/discussions
    SourceBook          // Reading
    SourceSession       // Agent/droid session
    SourceInternal      // Self-generated
)
```

### Entity (New)

```go
type Entity struct {
    Name     string     `json:"name"`
    Type     EntityType `json:"type"`
    BeatIDs  []string   `json:"beat_ids"` // Beats containing this entity
}

type EntityType int
const (
    EntityPerson      // Nick, DHH, etc.
    EntityTool        // Supabase, Ollama, etc.
    EntityConcept     // commitment, identity, etc.
    EntityProject     // runcible, modern-minuteman, etc.
    EntityOrganization // Factory, Anthropic, etc.
)
```

### Cluster (New)

```go
type Cluster struct {
    ID          string   `json:"id"`
    Name        string   `json:"name"`        // Auto-generated or user-set
    BeatIDs     []string `json:"beat_ids"`
    Centroid    []float64 `json:"centroid"`   // Embedding centroid
    Keywords    []string `json:"keywords"`    // Representative terms
    CreatedAt   time.Time `json:"created_at"`
}
```

### Chain (New)

```go
type Chain struct {
    ID          string    `json:"id"`
    Name        string    `json:"name"`
    BeatIDs     []string  `json:"beat_ids"`   // Ordered sequence
    CreatedAt   time.Time `json:"created_at"`
    RipenessScore float64 `json:"ripeness"`   // Aggregate ripeness
}
```

### Cache Structure

```go
// Stored in .beats/btv-cache.json
type Cache struct {
    Version       string              `json:"version"`
    GeneratedAt   time.Time           `json:"generated_at"`
    SourceHash    string              `json:"source_hash"` // Hash of beats.jsonl
    
    Taxonomies    map[string]Taxonomy `json:"taxonomies"`    // beat_id -> taxonomy
    Entities      []Entity            `json:"entities"`
    EntityIndex   map[string][]string `json:"entity_index"`  // entity_name -> beat_ids
    Ripeness      map[string]float64  `json:"ripeness"`      // beat_id -> score
    Clusters      []Cluster           `json:"clusters"`
    Chains        []Chain             `json:"chains"`
    ViewStats     map[string]ViewStat `json:"view_stats"`    // beat_id -> stats
    
    EmbeddingsAvailable bool `json:"embeddings_available"`
}

type ViewStat struct {
    ViewCount    int        `json:"view_count"`
    LastViewedAt *time.Time `json:"last_viewed_at"`
}
```

---

## Feature Specifications

### 1. Impetus Taxonomy & Auto-Classification

#### Problem
55% of beats (133/240) have generic "Manual entry" impetus, providing zero semantic signal for filtering or understanding.

#### Solution
Implement two-level taxonomy (Channel + Source) with automatic classification of unstructured impetus labels.

#### Classification Algorithm

```go
func ClassifyBeat(beat Beat) Taxonomy {
    // 1. Check explicit impetus label patterns
    label := strings.ToLower(beat.Impetus.Label)
    
    // Channel detection
    channel := detectChannel(label, beat.Content)
    source := detectSource(label, beat.Impetus.Meta)
    
    // Confidence based on signal strength
    confidence := calculateConfidence(label, beat.Content)
    
    return Taxonomy{Channel: channel, Source: source, Confidence: confidence}
}

func detectChannel(label, content string) Channel {
    patterns := map[Channel][]string{
        ChannelCoaching:    {"coaching", "nick", "mentor", "insight from"},
        ChannelResearch:    {"research", "study", "paper", "investigation"},
        ChannelDiscovery:   {"discovery", "found", "discovered", "stumbled"},
        ChannelDevelopment: {"development", "built", "implemented", "code"},
        ChannelReflection:  {"reflection", "thinking", "realized", "synthesis"},
        ChannelReference:   {"reference", "bookmark", "save", "purchase"},
        ChannelMilestone:   {"milestone", "complete", "shipped", "published"},
    }
    // Score each channel, return highest
}

func detectSource(label string, meta map[string]string) Source {
    // Check meta["channel"] first (explicit)
    if ch, ok := meta["channel"]; ok {
        // Map to Source enum
    }
    
    // Pattern match label
    patterns := map[Source][]string{
        SourceTwitter:     {"twitter", "x discovery", "tweet"},
        SourceGitHub:      {"github", "repo", "issue", "pr"},
        SourceWeb:         {"web", "article", "blog", "site"},
        SourceConversation: {"coaching", "call", "chat", "conversation"},
        SourceBook:        {"book", "reading", "chapter"},
        SourceSession:     {"session", "droid", "factory", "agent"},
    }
}
```

#### UI: Faceted Navigation

```
┌─ Channel ──────────────┐  ┌─ Source ─────────────────┐
│ ● All (240)            │  │ ● All                    │
│ ○ Coaching (23)        │  │ ○ Conversation (45)      │
│ ○ Research (18)        │  │ ○ Web (67)               │
│ ○ Discovery (89)       │  │ ○ Twitter (23)           │
│ ○ Development (42)     │  │ ○ GitHub (34)            │
│ ○ Reflection (31)      │  │ ○ Session (41)           │
│ ○ Reference (22)       │  │ ○ Internal (30)          │
│ ○ Milestone (15)       │  └──────────────────────────┘
└────────────────────────┘
```

#### Keybindings
- `f` - Toggle facet sidebar
- `1-7` - Quick filter by channel
- `!` - Clear all filters

#### Robot Commands

```bash
# List with taxonomy filter
btv --robot-list --channel coaching --source conversation

# Get taxonomy distribution
btv --robot-taxonomy-stats
# {"channels": {"coaching": 23, ...}, "sources": {...}}

# Reclassify a beat manually
echo '{"beat_id": "beat-123", "channel": "research", "source": "web"}' | btv --robot-set-taxonomy
```

---

### 2. Ripeness Scoring

#### Problem
No way to identify which beats are "mature" enough to become actionable work (beads).

#### Solution
Calculate ripeness score (0.0-1.0) based on multiple signals indicating a beat is ready for action.

#### Scoring Algorithm

```go
func CalculateRipeness(beat Beat, allBeats []Beat, viewStats ViewStat) float64 {
    var score float64
    
    // 1. Age factor (0-0.2): Older beats have had time to prove relevance
    ageDays := time.Since(beat.CreatedAt).Hours() / 24
    ageFactor := math.Min(ageDays / 30, 1.0) * 0.2
    
    // 2. Revisit factor (0-0.25): Beats you keep coming back to matter
    revisitFactor := math.Min(float64(viewStats.ViewCount) / 5, 1.0) * 0.25
    
    // 3. Connection factor (0-0.25): Related to other beats or beads
    connections := len(beat.LinkedBeads) + countRelatedBeats(beat, allBeats)
    connectionFactor := math.Min(float64(connections) / 3, 1.0) * 0.25
    
    // 4. Action language factor (0-0.2): Contains actionable phrasing
    actionFactor := detectActionLanguage(beat.Content) * 0.2
    
    // 5. Completeness factor (0-0.1): Has entities, good impetus, etc.
    completenessFactor := calculateCompleteness(beat) * 0.1
    
    score = ageFactor + revisitFactor + connectionFactor + actionFactor + completenessFactor
    
    return math.Min(score, 1.0)
}

func detectActionLanguage(content string) float64 {
    actionPatterns := []string{
        "should", "need to", "must", "will", "plan to",
        "implement", "build", "create", "fix", "add",
        "todo", "action", "next step", "follow up",
    }
    matches := 0
    content = strings.ToLower(content)
    for _, pattern := range actionPatterns {
        if strings.Contains(content, pattern) {
            matches++
        }
    }
    return math.Min(float64(matches) / 3, 1.0)
}
```

#### Ripeness Tiers

```go
const (
    RipenessFresh    = 0.0  // < 0.3: Recently added, still forming
    RipenessMaturing = 0.3  // 0.3-0.6: Gaining connections and revisits
    RipenessRipe     = 0.6  // 0.6-0.8: Strong candidate for action
    RipenessOverripe = 0.8  // > 0.8: Urgent - act or archive
)
```

#### UI: Ripeness Indicators

```
┌─ Beats ─────────────────────────────────────────────────────┐
│ 🟢 beat-20251204-001  Coaching    Commitment is about...    │  <- Ripe (0.72)
│ 🟢 beat-20251211-003  Research    AI memory persistence...  │  <- Ripe (0.68)
│ 🟡 beat-20251218-002  Discovery   Workspace reorg...        │  <- Maturing (0.45)
│ ⚪ beat-20260105-001  Session     Supabase keepalive...     │  <- Fresh (0.12)
└─────────────────────────────────────────────────────────────┘
```

#### Keybindings
- `R` - Sort by ripeness (highest first)
- `b` - Convert selected beat to bead (opens bd create with content)

#### Robot Commands

```bash
# Get ripest beats
btv --robot-list --sort ripeness --limit 10

# Get ripeness for specific beat
btv --robot-ripeness beat-20251204-001
# {"beat_id": "...", "score": 0.72, "factors": {"age": 0.18, "revisit": 0.20, ...}}
```

---

### 3. Entity Extraction

#### Problem
Beats reference people, tools, and concepts but there's no way to navigate by these entities.

#### Solution
Auto-extract entities from beat content using pattern matching and heuristics (no ML required for v0.2).

#### Extraction Algorithm

```go
func ExtractEntities(beat Beat) []Entity {
    var entities []Entity
    content := beat.Content
    
    // 1. People: Capitalized names, known patterns
    people := extractPeople(content)
    for _, name := range people {
        entities = append(entities, Entity{Name: name, Type: EntityPerson})
    }
    
    // 2. Tools: Known tool dictionary + CamelCase/lowercase patterns
    tools := extractTools(content)
    for _, tool := range tools {
        entities = append(entities, Entity{Name: tool, Type: EntityTool})
    }
    
    // 3. Concepts: Quoted phrases, repeated nouns
    concepts := extractConcepts(content)
    for _, concept := range concepts {
        entities = append(entities, Entity{Name: concept, Type: EntityConcept})
    }
    
    return entities
}

// Known entities dictionary (expandable)
var knownPeople = []string{"Nick", "DHH", "Claude", "Moritz"}
var knownTools = []string{
    "Supabase", "Ollama", "GitHub", "Cloudflare", "Vercel",
    "beads", "beats", "bv", "btv", "bd", "Factory", "Droid",
}
var knownConcepts = []string{
    "commitment", "identity", "narrative substrate", "psychoid buffer",
}

func extractPeople(content string) []string {
    var found []string
    
    // Check known people
    for _, person := range knownPeople {
        if strings.Contains(content, person) {
            found = append(found, person)
        }
    }
    
    // Pattern: "with [Name]", "from [Name]", "[Name] said"
    patterns := []string{
        `(?:with|from|by)\s+([A-Z][a-z]+)`,
        `([A-Z][a-z]+)\s+(?:said|mentioned|noted|suggested)`,
    }
    // Apply regex patterns...
    
    return unique(found)
}
```

#### UI: Entity Navigation

```
┌─ Entities ──────────────────────────┐
│ People                              │
│   Nick (8)  Claude (12)  DHH (3)   │
│                                     │
│ Tools                               │
│   Supabase (15)  beads (23)        │
│   Ollama (7)  Factory (19)         │
│                                     │
│ Concepts                            │
│   identity (6)  commitment (4)     │
│   narrative substrate (3)          │
└─────────────────────────────────────┘
```

#### Keybindings
- `e` - Toggle entity sidebar
- `E` - Entity search (fuzzy find entity, filter to its beats)

#### Robot Commands

```bash
# List all entities
btv --robot-entities
# {"people": [...], "tools": [...], "concepts": [...]}

# Filter beats by entity
btv --robot-list --entity "Nick"
btv --robot-list --entity-type person
```

---

### 4. Timeline View

#### Problem
No temporal awareness - can't see patterns in when thinking happened.

#### Solution
Toggle-able timeline view showing beat density over time with interactive navigation.

#### Timeline Rendering

```
┌─ Timeline ─────────────────────────────────────────────────────────────┐
│ 2025                                                                    │
│ Nov  ─────●─●──────●●●───●────────●──────────────────────────────────  │
│ Dec  ──●●●●●●●●●●●──●●●●●────●●●●●●●●●●●●──●●────●●●●●●●●●●●●●●●●●●●── │
│ 2026                                                                    │
│ Jan  ─●●●●●●─────────────────────────────────────────────────────────  │
│      ^                                                                  │
│      [6 beats on Jan 5-6]                                              │
│                                                                         │
│ ← → Navigate    Enter: Expand day    c: Channel colors    z: Zoom      │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Timeline Features

1. **Density visualization** - More dots = more beats
2. **Channel coloring** - Different colors for coaching vs discovery vs development
3. **Zoom levels** - Day / Week / Month / Quarter
4. **Interactive selection** - Navigate to date, see beats from that period
5. **Gap detection** - Highlight periods of inactivity

#### Data Structure

```go
type TimelineData struct {
    Start     time.Time
    End       time.Time
    Buckets   []TimelineBucket
    ZoomLevel ZoomLevel
}

type TimelineBucket struct {
    Date      time.Time
    BeatCount int
    Beats     []string // Beat IDs
    ByChannel map[Channel]int
}

type ZoomLevel int
const (
    ZoomDay ZoomLevel = iota
    ZoomWeek
    ZoomMonth
    ZoomQuarter
)
```

#### Keybindings
- `t` - Toggle timeline view
- `←/→` - Navigate timeline
- `z` - Cycle zoom level
- `Enter` - Select date, show beats in list
- `c` - Toggle channel coloring

#### Robot Commands

```bash
# Get timeline data
btv --robot-timeline --zoom month --start 2025-11-01 --end 2026-01-31
# {"buckets": [{"date": "2025-11", "count": 45, "by_channel": {...}}, ...]}

# Get beats for date range
btv --robot-list --after 2025-12-01 --before 2025-12-31
```

---

### 5. Theme Clustering (Ollama Integration)

#### Problem
Beats on related topics are scattered; no way to see thematic groupings.

#### Solution
Use local Ollama embeddings to cluster beats by semantic similarity.

#### Architecture

```go
type ClusterEngine struct {
    ollama    *ollama.Client
    available bool
    cache     *EmbeddingCache
}

func (c *ClusterEngine) GenerateClusters(beats []Beat, k int) ([]Cluster, error) {
    if !c.available {
        return nil, ErrOllamaUnavailable
    }
    
    // 1. Generate embeddings for all beats
    embeddings := make([][]float64, len(beats))
    for i, beat := range beats {
        emb, err := c.getEmbedding(beat)
        if err != nil {
            continue
        }
        embeddings[i] = emb
    }
    
    // 2. K-means clustering
    clusters := kmeans(embeddings, k)
    
    // 3. Generate cluster names from keywords
    for i := range clusters {
        clusters[i].Name = generateClusterName(clusters[i].BeatIDs, beats)
        clusters[i].Keywords = extractKeywords(clusters[i].BeatIDs, beats)
    }
    
    return clusters, nil
}

func (c *ClusterEngine) getEmbedding(beat Beat) ([]float64, error) {
    // Check cache first
    if emb, ok := c.cache.Get(beat.ID); ok {
        return emb, nil
    }
    
    // Generate via Ollama
    resp, err := c.ollama.Embeddings(context.Background(), &ollama.EmbeddingRequest{
        Model:  "nomic-embed-text",
        Prompt: beat.Content,
    })
    if err != nil {
        return nil, err
    }
    
    c.cache.Set(beat.ID, resp.Embedding)
    return resp.Embedding, nil
}
```

#### Graceful Degradation

When Ollama is unavailable:
1. Skip clustering entirely
2. Show message: "Install Ollama for theme clustering"
3. All other features work normally

#### UI: Cluster View

```
┌─ Theme Clusters ───────────────────────────────────────────────────────┐
│                                                                         │
│ ▼ Identity & Commitment (12 beats)                      Ripeness: 0.67 │
│   │ Commitment is fundamentally about identity...                       │
│   │ "I am the kind of person who..." framing                           │
│   │ Identity-based habit formation research                            │
│   └ [+9 more]                                                          │
│                                                                         │
│ ▶ AI Agent Memory (8 beats)                             Ripeness: 0.54 │
│ ▶ Tool Infrastructure (15 beats)                        Ripeness: 0.41 │
│ ▶ Modern Minuteman Concepts (11 beats)                  Ripeness: 0.38 │
│                                                                         │
│ [Unclustered: 23 beats]                                                │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Keybindings
- `C` - Toggle cluster view
- `Enter` - Expand/collapse cluster
- `n` - Rename cluster
- `m` - Merge selected clusters
- `b` - Create bead from entire cluster (epic)

#### Robot Commands

```bash
# Generate/refresh clusters
btv --robot-cluster --k 8

# Get cluster data
btv --robot-clusters
# {"clusters": [{"id": "...", "name": "...", "beat_ids": [...], "ripeness": 0.67}]}

# Get similar beats to a given beat
btv --robot-similar beat-20251204-001 --limit 5
```

---

### 6. Beat Chains

#### Problem
Related beats are not explicitly connected; thinking threads are implicit.

#### Solution
Allow explicit chaining of beats into sequences representing a line of thought.

#### Chain Operations

```go
func (s *Store) CreateChain(name string, beatIDs []string) (*Chain, error) {
    chain := &Chain{
        ID:        generateChainID(),
        Name:      name,
        BeatIDs:   beatIDs,
        CreatedAt: time.Now(),
    }
    
    // Calculate aggregate ripeness
    chain.RipenessScore = s.calculateChainRipeness(beatIDs)
    
    return chain, s.saveChain(chain)
}

func (s *Store) AddToChain(chainID, beatID string) error {
    chain, err := s.getChain(chainID)
    if err != nil {
        return err
    }
    
    chain.BeatIDs = append(chain.BeatIDs, beatID)
    chain.RipenessScore = s.calculateChainRipeness(chain.BeatIDs)
    
    return s.saveChain(chain)
}
```

#### UI: Chain Indicator in Detail View

```
┌─ Detail ────────────────────────────────────────────────────────────────┐
│ ID: beat-20251204-001                                                   │
│ Created: 2025-12-04 18:48                                               │
│ Impetus: Coaching / Conversation                                        │
│ Ripeness: 🟢 0.72                                                       │
│                                                                         │
│ ┌─ Part of Chain: "Identity Research" (5 beats) ──────────────────────┐│
│ │ 1. [this] Commitment is about identity                              ││
│ │ 2. "I am the kind of person..." framing                             ││
│ │ 3. Identity-based habit formation                                   ││
│ │ 4. Atomic Habits identity chapter                                   ││
│ │ 5. Application to MM operator standard                              ││
│ └─────────────────────────────────────────────────────────────────────┘│
│                                                                         │
│ Content:                                                                │
│ Commitment is fundamentally about identity, not discipline...          │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Keybindings
- `c` - Add beat to chain (prompts for chain name or existing)
- `x` - Remove beat from chain
- `[` / `]` - Navigate within chain (prev/next beat in chain)

#### Robot Commands

```bash
# Create chain
echo '{"name": "Identity Research", "beat_ids": ["beat-1", "beat-2"]}' | btv --robot-create-chain

# List chains
btv --robot-chains
# {"chains": [{"id": "...", "name": "...", "beat_count": 5, "ripeness": 0.67}]}

# Add to chain
echo '{"chain_id": "chain-1", "beat_id": "beat-3"}' | btv --robot-chain-add
```

---

### 7. Quick Capture Mode

#### Problem
Adding beats requires leaving context; friction prevents capture.

#### Solution
Minimal capture interface invokable from anywhere.

#### Implementation

```bash
# Minimal capture - opens single-line prompt
btv --capture

# With pre-filled impetus
btv --capture --channel coaching --source conversation

# Pipe content
echo "Quick insight about X" | btv --capture --channel reflection
```

#### UI: Capture Mode

```
┌─ btv capture ───────────────────────────────────────────────────────────┐
│                                                                         │
│ Channel: [Reflection ▼]    Source: [Internal ▼]                        │
│                                                                         │
│ ┌─────────────────────────────────────────────────────────────────────┐│
│ │ Your insight here...                                                ││
│ │                                                                     ││
│ └─────────────────────────────────────────────────────────────────────┘│
│                                                                         │
│ Enter: Save    Tab: Next field    Esc: Cancel                          │
└─────────────────────────────────────────────────────────────────────────┘
```

---

### 8. Stale Beat Review

#### Problem
Old beats accumulate without review; the archive becomes a graveyard.

#### Solution
Dedicated review mode for beats needing attention.

#### Staleness Detection

```go
func IsStale(beat Beat, viewStats ViewStat) bool {
    // Stale if:
    // 1. Older than 30 days AND
    // 2. Never viewed or not viewed in 14 days AND
    // 3. No linked beads AND
    // 4. Not part of a chain
    
    age := time.Since(beat.CreatedAt)
    if age < 30 * 24 * time.Hour {
        return false
    }
    
    if viewStats.ViewCount > 0 {
        if viewStats.LastViewedAt != nil {
            if time.Since(*viewStats.LastViewedAt) < 14 * 24 * time.Hour {
                return false
            }
        }
    }
    
    if len(beat.LinkedBeads) > 0 {
        return false
    }
    
    // Check chains...
    
    return true
}
```

#### UI: Review Mode

```
┌─ Stale Beat Review (42 beats need attention) ───────────────────────────┐
│                                                                         │
│ beat-20251104-003 (62 days old, never viewed)                          │
│ Channel: Discovery    Source: Web                                       │
│                                                                         │
│ "Some old insight that may or may not still be relevant..."            │
│                                                                         │
│ Actions:                                                                │
│   [k] Keep - Mark as reviewed                                          │
│   [a] Archive - Move to archive                                        │
│   [b] Convert - Create bead from this                                  │
│   [c] Chain - Add to existing chain                                    │
│   [d] Delete - Remove permanently                                      │
│                                                                         │
│ Progress: 3/42    Skip: →    Quit: q                                   │
└─────────────────────────────────────────────────────────────────────────┘
```

#### Keybindings
- `S` - Enter stale review mode
- `k/a/b/c/d` - Actions in review mode
- `→` - Skip without action

---

## View Architecture

### View Modes

```go
type ViewMode int
const (
    ViewList ViewMode = iota      // Default list view
    ViewTimeline                   // Temporal visualization
    ViewClusters                   // Theme groupings
    ViewReview                     // Stale beat review
    ViewCapture                    // Quick capture mode
)
```

### Layout System

```go
type Layout struct {
    Mode        ViewMode
    Width       int
    Height      int
    
    // Panels
    ShowFacets  bool    // Left sidebar
    ShowDetail  bool    // Right panel
    ShowEntities bool   // Entity sidebar
    
    // Focus
    Focus       FocusArea
}

type FocusArea int
const (
    FocusList FocusArea = iota
    FocusDetail
    FocusFacets
    FocusEntities
    FocusTimeline
    FocusCapture
)
```

### Responsive Breakpoints

```go
const (
    WidthCompact    = 60   // Single column, no sidebars
    WidthNormal     = 100  // List + detail
    WidthWide       = 140  // List + detail + one sidebar
    WidthUltraWide  = 180  // All panels
)
```

---

## Package Architecture

```
beats_viewer/
├── cmd/btv/
│   └── main.go                 # CLI entry, subcommands
├── pkg/
│   ├── model/
│   │   ├── beat.go            # Beat type + methods
│   │   ├── taxonomy.go        # Channel, Source enums
│   │   ├── entity.go          # Entity types
│   │   ├── cluster.go         # Cluster type
│   │   ├── chain.go           # Chain type
│   │   └── cache.go           # Cache structure
│   ├── loader/
│   │   ├── loader.go          # JSONL loading
│   │   ├── cache.go           # Cache read/write
│   │   └── migration.go       # v0.1 -> v0.2 migration
│   ├── taxonomy/
│   │   ├── classifier.go      # Auto-classification
│   │   └── patterns.go        # Detection patterns
│   ├── entity/
│   │   ├── extractor.go       # Entity extraction
│   │   ├── dictionary.go      # Known entities
│   │   └── index.go           # Entity index
│   ├── ripeness/
│   │   ├── scorer.go          # Ripeness calculation
│   │   └── factors.go         # Individual factors
│   ├── cluster/
│   │   ├── engine.go          # Clustering logic
│   │   ├── ollama.go          # Ollama client wrapper
│   │   ├── kmeans.go          # K-means implementation
│   │   └── naming.go          # Cluster name generation
│   ├── chain/
│   │   ├── store.go           # Chain CRUD
│   │   └── ripeness.go        # Chain ripeness
│   ├── timeline/
│   │   ├── data.go            # Timeline data structures
│   │   └── render.go          # ASCII rendering
│   └── ui/
│       ├── model.go           # Main bubbletea model
│       ├── views/
│       │   ├── list.go        # List view
│       │   ├── detail.go      # Detail view
│       │   ├── timeline.go    # Timeline view
│       │   ├── clusters.go    # Cluster view
│       │   ├── review.go      # Stale review
│       │   └── capture.go     # Quick capture
│       ├── components/
│       │   ├── facets.go      # Facet sidebar
│       │   ├── entities.go    # Entity sidebar
│       │   ├── search.go      # Search input
│       │   ├── statusbar.go   # Bottom status
│       │   └── help.go        # Help overlay
│       ├── styles.go          # Lipgloss styles
│       └── keybindings.go     # Key handling
├── migrations/
│   └── v0.2.go                # Data migration script
└── go.mod
```

---

## Robot Command Reference

### Discovery & Listing

| Command | Description |
|---------|-------------|
| `--robot-list` | List beats with filters |
| `--robot-show <id>` | Get single beat details |
| `--robot-search` | Full-text search |
| `--robot-help` | Command reference |

### Taxonomy

| Command | Description |
|---------|-------------|
| `--robot-taxonomy-stats` | Distribution by channel/source |
| `--robot-set-taxonomy` | Manually classify a beat |
| `--robot-reclassify` | Re-run auto-classification |

### Ripeness

| Command | Description |
|---------|-------------|
| `--robot-ripeness <id>` | Get ripeness score + factors |
| `--robot-ripe` | List beats above ripeness threshold |

### Entities

| Command | Description |
|---------|-------------|
| `--robot-entities` | List all extracted entities |
| `--robot-entity-beats <name>` | Beats containing entity |

### Timeline

| Command | Description |
|---------|-------------|
| `--robot-timeline` | Get timeline data |
| `--robot-gaps` | Identify activity gaps |

### Clusters

| Command | Description |
|---------|-------------|
| `--robot-cluster` | Generate clusters |
| `--robot-clusters` | List current clusters |
| `--robot-similar <id>` | Find similar beats |

### Chains

| Command | Description |
|---------|-------------|
| `--robot-chains` | List chains |
| `--robot-create-chain` | Create new chain |
| `--robot-chain-add` | Add beat to chain |

### Maintenance

| Command | Description |
|---------|-------------|
| `--robot-stale` | List stale beats |
| `--robot-cache-rebuild` | Rebuild entire cache |

---

## Migration: v0.1 → v0.2

### First Run Behavior

```go
func Migrate(beatsDir string) error {
    // 1. Create cache file
    cache := &Cache{
        Version:     "0.2.0",
        GeneratedAt: time.Now(),
    }
    
    // 2. Load all beats
    beats, err := loader.LoadBeats(beatsDir)
    if err != nil {
        return err
    }
    
    // 3. Auto-classify taxonomies
    cache.Taxonomies = make(map[string]Taxonomy)
    for _, beat := range beats {
        cache.Taxonomies[beat.ID] = taxonomy.Classify(beat)
    }
    
    // 4. Extract entities
    cache.Entities, cache.EntityIndex = entity.ExtractAll(beats)
    
    // 5. Calculate ripeness scores
    cache.Ripeness = make(map[string]float64)
    for _, beat := range beats {
        cache.Ripeness[beat.ID] = ripeness.Calculate(beat, beats, ViewStat{})
    }
    
    // 6. Initialize empty structures
    cache.Clusters = []Cluster{}
    cache.Chains = []Chain{}
    cache.ViewStats = make(map[string]ViewStat)
    
    // 7. Check Ollama availability
    cache.EmbeddingsAvailable = cluster.CheckOllama()
    
    // 8. Generate clusters if Ollama available
    if cache.EmbeddingsAvailable {
        cache.Clusters, _ = cluster.Generate(beats, 8)
    }
    
    // 9. Save cache
    return saveCache(beatsDir, cache)
}
```

### Cache Invalidation

Cache is rebuilt when:
1. `beats.jsonl` hash changes (beats added/modified)
2. User explicitly requests (`btv --rebuild-cache`)
3. Cache version mismatches binary version

---

## Success Criteria

### Quantitative

1. **Classification coverage**: 0% "Manual entry" after migration (all classified)
2. **Ripeness utility**: Top 10 ripest beats contain 80%+ of actionable items
3. **Entity coverage**: 90%+ of beats have at least one extracted entity
4. **Cluster coherence**: Manual review confirms 80%+ of clusters are thematically sensible
5. **Performance**: <200ms startup with 500 beats, <50ms navigation

### Qualitative

1. User can answer "what should I work on?" in <10 seconds
2. User discovers non-obvious connections between beats
3. Timeline reveals patterns in thinking activity
4. Stale beat review prevents archive accumulation
5. Quick capture removes friction from beat creation

---

## Implementation Phases

### Phase 1: Foundation (Days 1-2)
- Extended data model types
- Cache structure and persistence
- Migration script
- Taxonomy classifier

### Phase 2: Core Intelligence (Days 3-4)
- Ripeness scorer
- Entity extractor
- View stats tracking
- Cache rebuild logic

### Phase 3: UI Overhaul (Days 5-7)
- Faceted navigation
- Timeline view
- Ripeness indicators
- Entity sidebar

### Phase 4: Clustering (Days 8-9)
- Ollama integration
- Embedding cache
- K-means clustering
- Cluster UI

### Phase 5: Chains & Review (Days 10-11)
- Chain CRUD
- Chain UI
- Stale beat detection
- Review mode

### Phase 6: Polish (Days 12-14)
- Quick capture mode
- All robot commands
- Responsive layouts
- Help system update
- Testing & edge cases

---

## Dependencies

### Required
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/charmbracelet/bubbles` - Components
- `github.com/atotto/clipboard` - Clipboard access

### Optional
- `github.com/ollama/ollama` - Local embeddings (graceful degradation if unavailable)

### Development
- `github.com/stretchr/testify` - Testing assertions

---

## Appendix: Keybinding Reference

### Global
| Key | Action |
|-----|--------|
| `q` | Quit |
| `?` | Help |
| `/` | Search |
| `Esc` | Cancel/back |
| `Tab` | Cycle focus |

### Navigation
| Key | Action |
|-----|--------|
| `j/k` | Up/down |
| `g/G` | First/last |
| `←/→` | Page/navigate |
| `Enter` | Select/expand |

### Views
| Key | Action |
|-----|--------|
| `t` | Timeline view |
| `C` | Cluster view |
| `S` | Stale review |
| `f` | Toggle facets |
| `e` | Toggle entities |

### Filtering
| Key | Action |
|-----|--------|
| `1-7` | Filter by channel |
| `!` | Clear filters |
| `R` | Sort by ripeness |
| `E` | Entity search |
| `p` | Cycle projects |
| `a` | All projects |

### Actions
| Key | Action |
|-----|--------|
| `y` | Copy ID |
| `Y` | Copy content |
| `b` | Create bead |
| `c` | Add to chain |
| `r` | Refresh |
