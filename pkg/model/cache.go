package model

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"
)

// ViewStat tracks viewing statistics for a beat
type ViewStat struct {
	ViewCount    int        `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
}

// AlertType represents the type of alert
type AlertType int

const (
	AlertActivation      AlertType = iota // Something activating
	AlertEmergence                         // New pattern forming
	AlertDormancy                          // Ripe cluster going stale
	AlertCrystallization                   // Inferred beat→bead connection
	AlertDivergence                        // Notable human/agent gap
	AlertDriftAnomaly                      // Unusual attention shift
)

// AlertSeverity represents alert urgency
type AlertSeverity int

const (
	AlertInfo    AlertSeverity = iota
	AlertNotable
	AlertUrgent
)

// Alert represents a surfaced attention event
type Alert struct {
	ID        string        `json:"id"`
	Type      AlertType     `json:"type"`
	Severity  AlertSeverity `json:"severity"`
	Title     string        `json:"title"`
	Detail    string        `json:"detail"`
	ClusterID string        `json:"cluster_id,omitempty"`
	CreatedAt time.Time     `json:"created_at"`
	SeenAt    *time.Time    `json:"seen_at,omitempty"`
}

// CrystallizationInference represents an inferred beat→bead connection
type CrystallizationInference struct {
	BeatIDs    []string  `json:"beat_ids"`
	BeadID     string    `json:"bead_id"`
	BeadTitle  string    `json:"bead_title"`
	Confidence float64   `json:"confidence"`
	InferredAt time.Time `json:"inferred_at"`
}

// Cache stores computed v0.3 data alongside beats.jsonl
type Cache struct {
	Version     string    `json:"version"`
	GeneratedAt time.Time `json:"generated_at"`
	SourceHash  string    `json:"source_hash"`

	// v0.3.1 incremental update tracking
	BeatHashes map[string]string `json:"beat_hashes,omitempty"` // beatID -> content hash

	// v0.2 fields
	Taxonomies  map[string]Taxonomy `json:"taxonomies"`
	Entities    []Entity            `json:"entities"`
	EntityIndex map[string][]string `json:"entity_index"`
	Ripeness    map[string]float64  `json:"ripeness"`
	Clusters    []Cluster           `json:"clusters"`
	Chains      []Chain             `json:"chains"`
	ViewStats   map[string]ViewStat `json:"view_stats"`

	EmbeddingsAvailable bool `json:"embeddings_available"`

	// v0.3 additions
	AttentionStateJSON []byte                     `json:"attention_state,omitempty"`
	Crystallizations   []CrystallizationInference `json:"crystallizations,omitempty"`
	Alerts             []Alert                    `json:"alerts,omitempty"`
}

const CacheVersion = "0.4.0"
const CacheFileName = "btv-cache.json"

// NewCache creates a new empty cache
func NewCache() *Cache {
	return &Cache{
		Version:          CacheVersion,
		GeneratedAt:      time.Now(),
		BeatHashes:       make(map[string]string),
		Taxonomies:       make(map[string]Taxonomy),
		Entities:         []Entity{},
		EntityIndex:      make(map[string][]string),
		Ripeness:         make(map[string]float64),
		Clusters:         []Cluster{},
		Chains:           []Chain{},
		ViewStats:        make(map[string]ViewStat),
		Crystallizations: []CrystallizationInference{},
		Alerts:           []Alert{},
	}
}

// AddAlert adds an alert to the cache
func (c *Cache) AddAlert(alert Alert) {
	c.Alerts = append(c.Alerts, alert)
}

// UnseenAlerts returns alerts that haven't been seen
func (c *Cache) UnseenAlerts() []Alert {
	var unseen []Alert
	for _, a := range c.Alerts {
		if a.SeenAt == nil {
			unseen = append(unseen, a)
		}
	}
	return unseen
}

// MarkAlertSeen marks an alert as seen
func (c *Cache) MarkAlertSeen(id string) {
	now := time.Now()
	for i := range c.Alerts {
		if c.Alerts[i].ID == id {
			c.Alerts[i].SeenAt = &now
			break
		}
	}
}

// HashBeat computes a content hash for a beat for change detection
func HashBeat(b Beat) string {
	data, _ := json.Marshal(b)
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])[:16]
}

// EnrichedBeat holds a beat with its computed fields from cache
type EnrichedBeat struct {
	Beat
	Taxonomy          Taxonomy  `json:"-"`
	ExtractedEntities []Entity  `json:"-"`
	RipenessScore     float64   `json:"-"`
	ClusterID         string    `json:"-"`
	ChainIDs          []string  `json:"-"`
	ViewCount         int       `json:"-"`
	LastViewedAt      *time.Time `json:"-"`
}

// RipenessTier returns the ripeness tier for a score
func RipenessTier(score float64) string {
	switch {
	case score >= 0.8:
		return "Overripe"
	case score >= 0.6:
		return "Ripe"
	case score >= 0.3:
		return "Maturing"
	default:
		return "Fresh"
	}
}

// RipenessEmoji returns an emoji indicator for a ripeness score
func RipenessEmoji(score float64) string {
	switch {
	case score >= 0.8:
		return "🔴"
	case score >= 0.6:
		return "🟢"
	case score >= 0.3:
		return "🟡"
	default:
		return "⚪"
	}
}
