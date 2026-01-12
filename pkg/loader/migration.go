package loader

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/attention"
	"github.com/bierlingm/beats_viewer/pkg/entity"
	"github.com/bierlingm/beats_viewer/pkg/model"
	"github.com/bierlingm/beats_viewer/pkg/ripeness"
	"github.com/bierlingm/beats_viewer/pkg/taxonomy"
)

// MigrateToV02 performs first-run enrichment to build the cache
func MigrateToV02(beatsDir string, progressFn func(step string, current, total int)) (*model.Cache, error) {
	progress := func(step string, current, total int) {
		if progressFn != nil {
			progressFn(step, current, total)
		}
	}

	progress("Loading beats", 0, 0)
	beats, err := LoadBeats(beatsDir)
	if err != nil {
		return nil, fmt.Errorf("loading beats: %w", err)
	}

	sourceHash, err := ComputeSourceHash(beatsDir)
	if err != nil {
		return nil, fmt.Errorf("computing source hash: %w", err)
	}

	cache := model.NewCache()
	cache.SourceHash = sourceHash
	cache.GeneratedAt = time.Now()

	progress("Classifying taxonomies", 0, len(beats))
	for i, beat := range beats {
		cache.Taxonomies[beat.ID] = taxonomy.Classify(beat)
		progress("Classifying taxonomies", i+1, len(beats))
	}

	progress("Extracting entities", 0, len(beats))
	cache.Entities, cache.EntityIndex = entity.ExtractAll(beats)
	progress("Extracting entities", len(beats), len(beats))

	progress("Calculating ripeness", 0, len(beats))
	cache.ViewStats = make(map[string]model.ViewStat)
	for _, beat := range beats {
		cache.ViewStats[beat.ID] = model.ViewStat{}
	}
	cache.Ripeness = ripeness.CalculateAll(beats, cache.ViewStats)
	progress("Calculating ripeness", len(beats), len(beats))

	cache.Clusters = []model.Cluster{}
	cache.Chains = []model.Chain{}
	cache.EmbeddingsAvailable = false

	progress("Saving cache", 0, 1)
	if err := SaveCache(beatsDir, cache); err != nil {
		return nil, fmt.Errorf("saving cache: %w", err)
	}
	progress("Saving cache", 1, 1)

	return cache, nil
}

// EnsureCache loads existing cache or migrates to create one
func EnsureCache(beatsDir string, progressFn func(step string, current, total int)) (*model.Cache, error) {
	cache, needsRebuild, err := LoadOrCreateCache(beatsDir)
	if err != nil {
		return nil, err
	}

	if !needsRebuild && cache != nil {
		return cache, nil
	}

	return MigrateToV02(beatsDir, progressFn)
}

// LoadEnrichedBeats loads beats with their computed fields from cache
func LoadEnrichedBeats(beatsDir string, progressFn func(step string, current, total int)) ([]model.EnrichedBeat, *model.Cache, error) {
	beats, err := LoadBeats(beatsDir)
	if err != nil {
		return nil, nil, fmt.Errorf("loading beats: %w", err)
	}

	cache, err := EnsureCache(beatsDir, progressFn)
	if err != nil {
		return nil, nil, fmt.Errorf("ensuring cache: %w", err)
	}

	clusterIndex := make(map[string]string)
	for _, cluster := range cache.Clusters {
		for _, beatID := range cluster.BeatIDs {
			clusterIndex[beatID] = cluster.ID
		}
	}

	chainIndex := make(map[string][]string)
	for _, chain := range cache.Chains {
		for _, beatID := range chain.BeatIDs {
			chainIndex[beatID] = append(chainIndex[beatID], chain.ID)
		}
	}

	entityIdx := entity.NewIndex(cache.Entities, cache.EntityIndex)

	var enriched []model.EnrichedBeat
	for _, beat := range beats {
		eb := model.EnrichedBeat{
			Beat:          beat,
			Taxonomy:      cache.Taxonomies[beat.ID],
			RipenessScore: cache.Ripeness[beat.ID],
			ClusterID:     clusterIndex[beat.ID],
			ChainIDs:      chainIndex[beat.ID],
		}

		if viewStat, ok := cache.ViewStats[beat.ID]; ok {
			eb.ViewCount = viewStat.ViewCount
			eb.LastViewedAt = viewStat.LastViewedAt
		}

		for _, e := range entityIdx.GetForBeat(beat.ID) {
			eb.ExtractedEntities = append(eb.ExtractedEntities, *e)
		}

		enriched = append(enriched, eb)
	}

	return enriched, cache, nil
}

// RefreshCache rebuilds the cache regardless of validity
func RefreshCache(beatsDir string, progressFn func(step string, current, total int)) (*model.Cache, error) {
	return MigrateToV03(beatsDir, progressFn)
}

// MigrateToV03 builds v0.3 cache with attention analysis
func MigrateToV03(beatsDir string, progressFn func(step string, current, total int)) (*model.Cache, error) {
	progress := func(step string, current, total int) {
		if progressFn != nil {
			progressFn(step, current, total)
		}
	}

	progress("Loading beats", 0, 0)
	beats, err := LoadBeats(beatsDir)
	if err != nil {
		return nil, fmt.Errorf("loading beats: %w", err)
	}

	sourceHash, err := ComputeSourceHash(beatsDir)
	if err != nil {
		return nil, fmt.Errorf("computing source hash: %w", err)
	}

	cache := model.NewCache()
	cache.SourceHash = sourceHash
	cache.GeneratedAt = time.Now()

	// v0.2 fields
	progress("Classifying taxonomies", 0, len(beats))
	for i, beat := range beats {
		cache.Taxonomies[beat.ID] = taxonomy.Classify(beat)
		progress("Classifying taxonomies", i+1, len(beats))
	}

	progress("Extracting entities", 0, len(beats))
	cache.Entities, cache.EntityIndex = entity.ExtractAll(beats)
	progress("Extracting entities", len(beats), len(beats))

	progress("Calculating ripeness", 0, len(beats))
	cache.ViewStats = make(map[string]model.ViewStat)
	for _, beat := range beats {
		cache.ViewStats[beat.ID] = model.ViewStat{}
	}
	cache.Ripeness = ripeness.CalculateAll(beats, cache.ViewStats)
	progress("Calculating ripeness", len(beats), len(beats))

	// Preserve existing clusters if available
	existingCache, _ := LoadCache(beatsDir)
	if existingCache != nil && len(existingCache.Clusters) > 0 {
		cache.Clusters = existingCache.Clusters
		cache.EmbeddingsAvailable = existingCache.EmbeddingsAvailable
		cache.Chains = existingCache.Chains
	} else {
		cache.Clusters = []model.Cluster{}
		cache.Chains = []model.Chain{}
		cache.EmbeddingsAvailable = false
	}

	// v0.3 attention analysis
	progress("Computing attention", 0, 1)
	attState := ComputeAttentionState(beats, cache.Clusters, cache.Ripeness)
	if attJSON, err := json.Marshal(attState); err == nil {
		cache.AttentionStateJSON = attJSON
	}
	progress("Computing attention", 1, 1)

	progress("Saving cache", 0, 1)
	if err := SaveCache(beatsDir, cache); err != nil {
		return nil, fmt.Errorf("saving cache: %w", err)
	}
	progress("Saving cache", 1, 1)

	return cache, nil
}

// ComputeAttentionState computes all attention analysis
func ComputeAttentionState(beats []model.Beat, clusters []model.Cluster, ripeness map[string]float64) *attention.AttentionState {
	state := attention.NewAttentionState()

	// Compute heartbeat
	state.Heartbeat = attention.ComputeHeartbeat(beats, attention.DefaultHeartbeatConfig())

	// Compute activations
	state.Activations = attention.DetectActivations(beats, clusters, attention.DefaultActivationConfig())

	// Compute drift
	state.DriftReport = attention.ComputeDrift(beats, clusters, attention.DefaultDriftConfig())

	// Compute dormancy
	state.Dormant = attention.DetectDormancy(beats, clusters, ripeness, attention.DefaultDormancyConfig())

	// Compute emergence
	state.Emergent = attention.DetectEmergence(beats, clusters, attention.DefaultEmergenceConfig())

	// Compute orientation (depends on activations and drift)
	state.Orientation = attention.ComputeOrientation(
		beats, clusters, state.Activations, state.DriftReport, ripeness,
		attention.DefaultOrientationConfig(),
	)

	return state
}

// GetAttentionState deserializes attention state from cache
func GetAttentionState(cache *model.Cache) (*attention.AttentionState, error) {
	if cache == nil || len(cache.AttentionStateJSON) == 0 {
		return nil, nil
	}

	var state attention.AttentionState
	if err := json.Unmarshal(cache.AttentionStateJSON, &state); err != nil {
		return nil, fmt.Errorf("unmarshaling attention state: %w", err)
	}

	return &state, nil
}

// RefreshAttentionState recomputes attention state for existing cache
func RefreshAttentionState(beatsDir string, cache *model.Cache) error {
	beats, err := LoadBeats(beatsDir)
	if err != nil {
		return fmt.Errorf("loading beats: %w", err)
	}

	state := ComputeAttentionState(beats, cache.Clusters, cache.Ripeness)
	if attJSON, err := json.Marshal(state); err == nil {
		cache.AttentionStateJSON = attJSON
	}

	return SaveCache(beatsDir, cache)
}
