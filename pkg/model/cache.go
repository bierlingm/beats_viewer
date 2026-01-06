package model

import "time"

// ViewStat tracks viewing statistics for a beat
type ViewStat struct {
	ViewCount    int        `json:"view_count"`
	LastViewedAt *time.Time `json:"last_viewed_at,omitempty"`
}

// Cache stores computed v0.2 data alongside beats.jsonl
type Cache struct {
	Version     string    `json:"version"`
	GeneratedAt time.Time `json:"generated_at"`
	SourceHash  string    `json:"source_hash"`

	Taxonomies  map[string]Taxonomy `json:"taxonomies"`
	Entities    []Entity            `json:"entities"`
	EntityIndex map[string][]string `json:"entity_index"`
	Ripeness    map[string]float64  `json:"ripeness"`
	Clusters    []Cluster           `json:"clusters"`
	Chains      []Chain             `json:"chains"`
	ViewStats   map[string]ViewStat `json:"view_stats"`

	EmbeddingsAvailable bool `json:"embeddings_available"`
}

const CacheVersion = "0.2.0"
const CacheFileName = "btv-cache.json"

// NewCache creates a new empty cache
func NewCache() *Cache {
	return &Cache{
		Version:     CacheVersion,
		GeneratedAt: time.Now(),
		Taxonomies:  make(map[string]Taxonomy),
		Entities:    []Entity{},
		EntityIndex: make(map[string][]string),
		Ripeness:    make(map[string]float64),
		Clusters:    []Cluster{},
		Chains:      []Chain{},
		ViewStats:   make(map[string]ViewStat),
	}
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
