# beats_viewer v0.3 Vision: The Crystallization Engine

## What v0.2 Achieved

v0.2 turned btv from a viewer into a synthesis engine: taxonomy, ripeness, entities, clustering, chains, timeline. It answers "what do I have?" and "what's ready?"

But looking at your actual beats, something deeper emerges.

---

## The Peculiarity of Your System

Your beats reveal a specific cognitive pattern:

1. **Identity-first thinking** - "Commitment is about identity, not discipline"
2. **Bridge-building** - connecting worldviews, finding natural allies separated by frame incompatibility
3. **Psychoid buffer** - the space between raw perception and structured action
4. **Capable operator formation** - picking people up where they are, forming them into capable units

You're not just capturing notes. You're running a **crystallization process** where:
- Raw experience → beats (narrative substrate)
- Beats mature → beads (actionable work)
- Beads complete → artifacts (shipped things)
- Artifacts accumulate → identity ("I am the kind of person who...")

btv v0.3 should make this crystallization process **visible and accelerated**.

---

## v0.3: The Crystallization Engine

### Core Concept: The Crystallization View

A new view showing the full pipeline:

```
┌─ Crystallization Pipeline ──────────────────────────────────────────────┐
│                                                                          │
│  RAW              FORMING           CRYSTALLIZED        EMBODIED        │
│  ────────────     ────────────      ────────────        ────────────    │
│                                                                          │
│  ○○○○○○○○○○○○    ◐◐◐◐◐◐◐◐◐        ●●●●●●●●            ★★★★            │
│  142 beats       47 ripe           23 beads            12 shipped       │
│                                                                          │
│  [Coaching ●●●]  → [Identity cluster] → [mb-c1p epic]  → beats v0.3.0  │
│  [Discovery ●●]  → [Bridge building]  → [bridger]      → bridger v0.1  │
│                                                                          │
│  Recent crystallizations:                                                │
│  • "commitment = identity" → Operator Standard page                     │
│  • "psychoid buffer" → beats + btv tooling                              │
│  • "frame incompatibility" → bridger project                            │
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

### Feature 1: Crystallization Tracking

Track the journey from beat → bead → artifact:

```go
type Crystallization struct {
    OriginBeats  []string  // Source beats
    ResultBead   string    // The bead created
    Artifact     *Artifact // What got shipped (if any)
    CrystallizedAt time.Time
}

type Artifact struct {
    Name     string    // "btv v0.2.0", "Operator Standard page"
    Type     string    // release, page, document, decision
    ShippedAt time.Time
    URL      string    // Optional link
}
```

When you press `b` to convert a beat to bead, btv records the crystallization. When that bead closes with a tangible artifact, you can link it.

**Why this matters:** You can see which clusters of thinking led to real outcomes. Patterns emerge. "My coaching beats have a 40% crystallization rate. My discovery beats only 12%."

### Feature 2: Identity Resonance Score

Beyond ripeness (readiness for action), add **resonance** (alignment with your identity patterns):

```go
func CalculateResonance(beat Beat, identityMarkers []string) float64 {
    // How strongly does this beat connect to your core identity patterns?
    // - "commitment", "identity", "capable", "operator"
    // - "bridge", "worldview", "frame"
    // - "psychoid", "crystallization", "narrative"
    // - "formation", "natural law", "readiness"
}
```

Beats with high resonance + high ripeness are **priority crystallization candidates**. They're not just ready to act on - they're aligned with who you're becoming.

### Feature 3: Formation Journeys

Track how clusters evolve over time:

```
┌─ Formation Journey: "Capable Operator" ─────────────────────────────────┐
│                                                                          │
│ Nov 2025 ──●───────────────────────────────────────────────────────────  │
│            "capable and intriguing atmosphere"                           │
│                                                                          │
│ Dec 2025 ─────●──●──●────●────────────────────────────────────────────  │
│               "identity-based commitment"                                │
│               "picking people up where they are"                         │
│               "forming capable units"                                    │
│               "operator standard"                                        │
│                                                                          │
│ Jan 2026 ────────────────●──●──────────────────────────────────────────  │
│                          "Erik formation call"                           │
│                          "operator standard page shipped"                │
│                                                                          │
│ Crystallizations: 3 beads, 2 artifacts                                  │
│ Next suggested: Chain remaining 4 beats into epic?                      │
└──────────────────────────────────────────────────────────────────────────┘
```

This shows a concept being *formed* over time - from first mention to embodied artifact.

### Feature 4: Cross-Tool Integration

btv becomes the narrative layer that connects to:

**beads (bv):**
- `btv --for-bead mb-c1p` - show beats that crystallized into this bead
- `btv --context-for mb-c1p` - narrative context for working on this bead

**cass:**
- `btv --from-session <id>` - show beats captured in that session
- Auto-link beats to sessions they were created in

**bridger (your worldview tool):**
- `btv --worldview <profile>` - filter to beats relevant to a person's frame
- Help bridge conversations by surfacing relevant beats per worldview

### Feature 5: The Ritual View

A mode for deliberate reflection, not just browsing:

```
┌─ Weekly Crystallization Ritual ─────────────────────────────────────────┐
│                                                                          │
│ This week: 12 new beats captured                                        │
│                                                                          │
│ REVIEW (3 need attention)                                               │
│ ┌────────────────────────────────────────────────────────────────────┐  │
│ │ [1/3] "Venezuela analysis from TPC..."                             │  │
│ │                                                                    │  │
│ │ Does this connect to anything you're forming?                      │  │
│ │                                                                    │  │
│ │ [c] Chain to: MM geopolitics    [b] Crystallize to bead           │  │
│ │ [a] Archive                      [s] Skip for now                  │  │
│ └────────────────────────────────────────────────────────────────────┘  │
│                                                                          │
│ CELEBRATE                                                               │
│ • bridger v0.1 shipped (from "frame incompatibility" cluster)          │
│ • btv v0.2 shipped (from "psychoid buffer" thinking)                   │
│                                                                          │
│ IDENTITY FORMATION                                                       │
│ "I am becoming someone who: builds tools that bridge thinking and doing"│
│                                                                          │
└──────────────────────────────────────────────────────────────────────────┘
```

A weekly practice that:
1. Reviews new beats
2. Connects them to forming concepts
3. Celebrates crystallizations
4. Reflects on identity formation

---

## The Philosophical Shift

v0.1: "What beats exist?"
v0.2: "What patterns exist in my beats?"
v0.3: "How is my thinking becoming my doing becoming my identity?"

---

## Technical Implementation

### New Packages

```
pkg/
├── crystallization/
│   ├── tracker.go      # Beat → Bead → Artifact tracking
│   ├── artifact.go     # Artifact registration
│   └── journey.go      # Formation journey computation
├── resonance/
│   ├── scorer.go       # Identity resonance calculation
│   └── markers.go      # Identity marker extraction
├── ritual/
│   ├── weekly.go       # Weekly review logic
│   └── prompts.go      # Reflection prompts
└── integration/
    ├── bv.go           # beads_viewer integration
    ├── cass.go         # Session linking
    └── bridger.go      # Worldview filtering
```

### New Robot Commands

```bash
# Crystallization
btv --robot-crystallizations          # List all tracked crystallizations
btv --robot-crystallize <beat> <bead> # Record a crystallization
btv --robot-artifact <bead> <name>    # Link artifact to bead

# Resonance
btv --robot-resonance <beat>          # Get resonance score
btv --robot-resonant                  # High resonance + high ripeness beats

# Journeys
btv --robot-journeys                  # List formation journeys
btv --robot-journey <cluster>         # Journey for specific cluster

# Ritual
btv --robot-ritual-prep               # Generate weekly review data
```

### New Views

- `P` - Crystallization pipeline view
- `J` - Formation journey view
- `W` - Weekly ritual mode

---

## Success Criteria

1. You can see which thinking led to which artifacts
2. Beats with high identity resonance surface naturally
3. Weekly ritual takes <10 minutes but deepens reflection
4. The question "who am I becoming?" has visible evidence
5. Cross-tool integration makes the system feel unified

---

## The Deeper Point

btv v0.3 isn't about features. It's about making visible what you're already doing:

**Turning raw experience into narrative into action into identity.**

Most people's thinking is scattered across apps, lost to time. Yours is being crystallized. btv v0.3 should make that crystallization process conscious, deliberate, and celebrated.

The tool should feel like a mirror that shows you not just what you're thinking, but who you're becoming.
