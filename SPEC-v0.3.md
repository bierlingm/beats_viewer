# beats_viewer v0.3 Specification: The Attention Engine

## Philosophy

v0.1 asked: "What beats exist?"
v0.2 asked: "What patterns exist in my beats?"
v0.3 asks: **"What is my attention doing, and what does that reveal?"**

The fundamental insight: **the act of capturing a beat is itself data.** Something crossed your threshold from noise to "this matters." The pattern of what you notice - not the content of the notes - is the primary signal.

Beats are data points. Attention patterns are the signal.

---

## Core Concepts

### Attention Stream

Your captures over time form a stream. The stream has:
- **Flow rate**: How much are you capturing? (bursts, droughts, steady)
- **Composition**: What topics, sources, channels?
- **Direction**: What's growing, shrinking, emerging, fading?

### Activation

A burst of related captures indicates something is **activating** in your thinking. Not a deliberate project - something your unconscious is orienting toward. Activation detection:
- 3+ beats in a cluster within 72 hours
- New beats in a dormant cluster
- Multiple sources converging on same topic

### Drift

Your attention shifts over time. Drift detection compares windows:
- What topics are you capturing more than last month?
- What topics are fading?
- What new territories are emerging?

Drift is not good or bad. It's information about where you're headed.

### Emergence

New clusters forming that don't fit existing patterns. Something new is entering your attention that hasn't crystallized into a category yet. Early signal of new directions.

### Dormancy

Clusters with ripe beats but no recent activity. Either:
- Integration complete (archive)
- Attention moved elsewhere (ok)
- Crystallization overdue (surface it)

### Crystallization Inference

Don't track beat→bead connections manually. **Infer them:**
- Content similarity between beat and bead description
- Temporal proximity (bead created after related beats)
- Entity overlap (same people/tools/concepts)
- Session correlation

Confidence scores, not binary. "This bead seems connected to these beats (78% confidence)."

### Agent Divergence

Agents capture beats too (from droid sessions). Compare patterns:
- What do agents notice that you don't? (blind spots)
- What do you notice that agents don't? (human-only perception)
- Where do you align? (amplification)

The gap between human and agent attention is itself meaningful.

---

## Features

### 1. Attention Dashboard (Primary View)

The new default view when launching btv. Not a list of beats - a view of your attention patterns.

```
┌─ Attention Dashboard ───────────────────────────────────────────────────┐
│                                                                          │
│ ▌NOW ACTIVATING                                                         │
│ │ "Tool Infrastructure" ████████░░ 4 beats in 72h                       │
│ │ "Identity/Commitment" ███░░░░░░░ reactivated after 2 weeks            │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ ▌ATTENTION DRIFT (30 days)                                              │
│ │ ↑ Tool Infrastructure    +12 beats   [crystallizing]                  │
│ │ ↑ Identity/Commitment    +5 beats                                     │
│ │ → MM Concepts            stable                                       │
│ │ ↓ Geopolitics            -60%                                         │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ ▌CRYSTALLIZING (inferred)                                               │
│ │ • "psychoid buffer" (6 beats) → btv v0.2          confidence: 84%    │
│ │ • "frame incompatibility" (4 beats) → bridger     confidence: 71%    │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ ▌EMERGING                                                               │
│ │ ? "Human-AI collaboration" - 3 beats, no cluster yet                  │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ ▌AGENT DIVERGENCE                                                       │
│ │ Agents captured: 23 beats this week                                   │
│ │ You're missing: "session memory patterns" (agents: 3, you: 0)        │
│ │ You notice more: "identity framing" (you: 4, agents: 0)              │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ ▌DORMANT (may need attention)                                           │
│ │ • "Book project" - 4 ripe beats, 38 days inactive                    │
│ └───────────────────────────────────────────────────────────────────────│
│                                                                          │
│ A:dashboard  L:list  T:timeline  C:clusters  ?:help  q:quit             │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2. Activation Detection

Real-time detection of what's activating:

```go
type Activation struct {
    ClusterID     string
    ClusterName   string
    BeatCount     int           // Beats in detection window
    Window        time.Duration // e.g., 72 hours
    Beats         []string      // The triggering beat IDs
    PriorActivity int           // Beats in prior equivalent window
    Type          ActivationType
}

type ActivationType int
const (
    ActivationBurst      // Sudden increase in existing cluster
    ActivationReactivation // Dormant cluster waking up
    ActivationEmergent   // New pattern forming
)
```

Detection algorithm:
- Window: 72 hours (configurable)
- Threshold: 3+ beats in window OR 2x prior window rate
- Reactivation: Any beat in cluster dormant >14 days

### 3. Drift Analysis

Compare attention patterns across time windows:

```go
type DriftReport struct {
    Window      time.Duration // e.g., 30 days
    PriorWindow time.Duration // Comparison period
    
    Rising      []DriftItem   // Topics with increased capture
    Stable      []DriftItem   // Topics with similar capture
    Fading      []DriftItem   // Topics with decreased capture
    Emerged     []DriftItem   // New topics not in prior window
    Vanished    []DriftItem   // Topics in prior window with zero now
}

type DriftItem struct {
    ClusterID    string
    ClusterName  string
    CurrentCount int
    PriorCount   int
    ChangePercent float64
    Direction    DriftDirection // Rising, Stable, Fading, Emerged, Vanished
}
```

### 4. Crystallization Inference

Automatically detect when beats became beads:

```go
type CrystallizationInference struct {
    BeatIDs       []string
    BeadID        string
    BeadTitle     string
    Confidence    float64
    Signals       []InferenceSignal
    InferredAt    time.Time
}

type InferenceSignal struct {
    Type   SignalType // ContentSimilarity, TemporalProximity, EntityOverlap, SessionCorrelation
    Score  float64
    Detail string
}
```

Inference algorithm:
1. For each bead, compute content similarity with all beats (TF-IDF or embedding distance)
2. Weight by temporal proximity (beats captured shortly before bead creation)
3. Boost for entity overlap (same people/tools/concepts)
4. Boost for session correlation (same droid session)
5. Threshold: confidence > 0.6 to surface

This requires integration with beads (read .beads/issues.jsonl).

### 5. Agent Divergence Analysis

Compare human captures vs. agent captures:

```go
type DivergenceReport struct {
    Window         time.Duration
    HumanBeats     int
    AgentBeats     int
    
    HumanOnly      []DivergenceItem // Topics human captures, agents don't
    AgentOnly      []DivergenceItem // Topics agents capture, human doesn't
    Amplified      []DivergenceItem // Both capture, high overlap
    
    BlindSpots     []string // Agent-only topics human might want to notice
}

type DivergenceItem struct {
    Topic        string // Cluster or entity
    HumanCount   int
    AgentCount   int
    Ratio        float64
}
```

Agent beats identified by:
- Impetus containing "droid-session", "factory", "agent"
- Or explicit flag in beat metadata

### 6. Orientation Summary

Computed summary of where attention is pointing:

```go
type OrientationSummary struct {
    ComputedAt    time.Time
    Window        time.Duration
    
    TopTopics     []OrientationItem // Weighted by recency + volume
    Growing       []OrientationItem // Positive drift
    Crystallizing []OrientationItem // High ripeness + recent activity
    Emerging      []OrientationItem // New clusters forming
    
    Narrative     string // One-sentence summary (optional, AI-generated)
}

type OrientationItem struct {
    Topic    string
    Weight   float64 // 0.0-1.0, normalized
    Trend    string  // ↑ ↓ →
    Signal   string  // Why this is notable
}
```

### 7. Heartbeat Visualization

Temporal rhythm of captures:

```
┌─ Capture Rhythm (90 days) ──────────────────────────────────────────────┐
│                                                                          │
│ ███░░░███████░░░░░░██░░░░░░░░████████████████░░░░░███████░░░█████████   │
│ Oct         Nov              Dec                  Jan                    │
│    ^            ^                ^                    ^                  │
│    Gap       Burst           Quiet             Current burst            │
│                                                                          │
│ Average: 2.3 beats/day    Current: 4.1 beats/day (↑78%)                │
│ Longest gap: 8 days (Nov 12-20)                                         │
└──────────────────────────────────────────────────────────────────────────┘
```

Shows:
- Density of captures over time
- Bursts and gaps
- Current rate vs. average
- Anomalies

### 8. Alert System

Threshold-based surfacing (not requiring ritual):

```go
type Alert struct {
    ID        string
    Type      AlertType
    Severity  AlertSeverity // Info, Notable, Urgent
    Title     string
    Detail    string
    Actions   []AlertAction
    CreatedAt time.Time
    SeenAt    *time.Time
}

type AlertType int
const (
    AlertActivation         // Something activating
    AlertEmergence          // New pattern forming
    AlertDormancy           // Ripe cluster going stale
    AlertCrystallization    // Inferred beat→bead connection
    AlertDivergence         // Notable human/agent gap
    AlertDriftAnomaly       // Unusual attention shift
)

type AlertAction struct {
    Key   string // Keyboard shortcut
    Label string
    Cmd   string // What happens
}
```

Alerts appear:
- Banner in TUI when launching
- Available via `--robot-alerts`
- Optional: desktop notification integration

Thresholds (configurable):
- Activation: 3+ beats in 72h
- Emergence: 3+ beats not matching existing clusters
- Dormancy: Ripe cluster with no activity for 30 days
- Divergence: Agent captures 3+ on topic human has 0

---

## Architecture

```
pkg/
├── attention/
│   ├── stream.go        # Attention stream modeling
│   ├── activation.go    # Burst/reactivation detection
│   ├── drift.go         # Attention drift over time
│   ├── emergence.go     # New pattern detection
│   ├── dormancy.go      # Stale cluster detection
│   ├── orientation.go   # Orientation summary
│   └── heartbeat.go     # Temporal rhythm analysis
│
├── inference/
│   ├── crystallization.go  # Beat→bead connection inference
│   ├── similarity.go       # Content similarity computation
│   └── beads.go            # Beads integration (read issues.jsonl)
│
├── divergence/
│   ├── analyzer.go      # Human vs. agent comparison
│   ├── classifier.go    # Classify beat as human/agent
│   └── blindspot.go     # Identify blind spots
│
├── alert/
│   ├── detector.go      # Threshold detection
│   ├── store.go         # Alert persistence
│   └── config.go        # Threshold configuration
│
└── ui/
    └── views/
        ├── dashboard.go     # Attention dashboard
        ├── heartbeat.go     # Rhythm visualization
        └── drift.go         # Drift detail view
```

### Cache Extensions

```go
// Added to existing Cache structure
type CacheV3 struct {
    // ... existing v0.2 fields ...
    
    // v0.3 additions
    AttentionState    AttentionState    `json:"attention_state"`
    Crystallizations  []CrystallizationInference `json:"crystallizations"`
    Alerts            []Alert           `json:"alerts"`
    LastDivergenceAt  time.Time         `json:"last_divergence_at"`
}

type AttentionState struct {
    ComputedAt     time.Time
    Activations    []Activation
    DriftReport    DriftReport
    Orientation    OrientationSummary
    Divergence     DivergenceReport
}
```

### Beads Integration

btv needs to read beads data for crystallization inference:

```go
// pkg/inference/beads.go
func LoadBeads(beadsDir string) ([]Bead, error) {
    // Read .beads/issues.jsonl
    // Parse into minimal Bead struct (ID, Title, Description, CreatedAt, ClosedAt)
    // Used for crystallization inference only
}
```

This is read-only. btv never writes to beads.

---

## UI Changes

### New Default: Attention Dashboard

`btv` now opens to Attention Dashboard, not list view.

Keybindings:
- `A` - Attention Dashboard (new default)
- `L` - List view (was default)
- `T` - Timeline view
- `C` - Cluster view
- `D` - Drift detail view (new)
- `H` - Heartbeat/rhythm view (new)

### Alert Banner

When alerts exist, show banner at top:

```
┌─ 2 alerts ───────────────────────────────────────────────────────────────┐
│ ! "Tool Infrastructure" activating (4 beats in 72h)                     │
│ ? New pattern emerging: "Human-AI collaboration"                         │
│                                                             [a] view all │
└──────────────────────────────────────────────────────────────────────────┘
```

Dismiss with `x`, view all with `a`.

### Dashboard Navigation

In dashboard, sections are navigable:
- `j/k` moves between sections
- `Enter` expands section or drills into detail
- `→` on a cluster goes to cluster view filtered to it

---

## Robot Commands

### Attention

```bash
btv --robot-attention              # Full attention state
btv --robot-activating             # Currently activating clusters
btv --robot-drift --days 30        # Drift report for period
btv --robot-orientation            # Where attention is pointing
btv --robot-heartbeat --days 90    # Capture rhythm data
```

### Inference

```bash
btv --robot-crystallizations       # All inferred crystallizations
btv --robot-crystallization <bead> # Beats that likely became this bead
btv --robot-infer                  # Run crystallization inference now
```

### Divergence

```bash
btv --robot-divergence             # Human vs. agent comparison
btv --robot-blindspots             # What agents notice that you don't
btv --robot-agent-beats            # List agent-captured beats
```

### Alerts

```bash
btv --robot-alerts                 # Current alerts
btv --robot-alerts --unseen        # Only unseen alerts
btv --robot-dismiss <alert-id>     # Mark alert seen
```

---

## Configuration

New config options in `.beats/btv-config.yaml`:

```yaml
attention:
  activation_window: 72h        # Window for burst detection
  activation_threshold: 3       # Beats in window to trigger
  drift_window: 30d             # Default drift comparison period
  dormancy_threshold: 30d       # When to flag dormant clusters

alerts:
  enabled: true
  desktop_notifications: false  # Future: OS notifications
  
  thresholds:
    activation: true
    emergence: true
    dormancy: true
    divergence: true
    crystallization: true

inference:
  crystallization_confidence: 0.6  # Minimum to surface
  similarity_method: tfidf         # or "embedding" if Ollama available

divergence:
  agent_impetus_patterns:
    - "droid-session"
    - "factory"
    - "agent"
```

---

## Implementation Phases

### Phase 1: Attention Core (Days 1-3)
- pkg/attention/ - stream, activation, drift, dormancy
- Cache extensions for attention state
- Basic computation on startup

### Phase 2: Dashboard UI (Days 4-5)
- Attention Dashboard view
- Section navigation
- Alert banner

### Phase 3: Inference (Days 6-7)
- pkg/inference/ - crystallization, similarity
- Beads integration (read-only)
- Confidence scoring

### Phase 4: Divergence (Days 8-9)
- pkg/divergence/ - analyzer, classifier
- Human/agent classification
- Blind spot detection

### Phase 5: Alerts & Polish (Days 10-12)
- pkg/alert/ - detector, store, config
- Robot commands
- Heartbeat visualization
- Configuration system

---

## Success Criteria

### Quantitative
1. Dashboard loads in <500ms with 300 beats
2. Activation detection catches 90%+ of actual bursts (manual review)
3. Crystallization inference has <20% false positive rate
4. Alert volume: 1-5 per week (not noisy, not silent)

### Qualitative
1. Opening btv tells you something you didn't consciously know
2. Activations match your felt sense of "what's alive right now"
3. Drift report matches your sense of where attention has moved
4. Crystallization inferences feel accurate ("yes, that bead came from those beats")
5. Agent divergence reveals genuine blind spots

### The Mirror Test
When you look at the Attention Dashboard, does it show you something true about where you're headed that you couldn't see directly?

---

## What This Is Not

- Not a productivity tracker
- Not a goal-setting system
- Not a manual journaling prompt
- Not requiring ritual or discipline

It's an **attention microscope** - making visible what's already happening below conscious awareness.

---

## The Deeper Point

Your captures are not notes. They're **votes** - each beat is a vote for "this matters." The pattern of votes reveals orientation before you consciously choose it.

v0.3 makes that pattern visible. Not so you can optimize it. So you can see where you're already going and decide if that's where you want to be.

The tool doesn't tell you who to become. It shows you who you're becoming.
