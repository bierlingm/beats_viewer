package loader

import (
	"fmt"
	"time"

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
	return MigrateToV02(beatsDir, progressFn)
}
