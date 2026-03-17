package cache

import (
	"github.com/bierlingm/beats_viewer/pkg/model"
)

// DiffResult represents changes between beats and cached state
type DiffResult struct {
	Added    []string // Beat IDs that are new
	Modified []string // Beat IDs that changed
	Removed  []string // Beat IDs that no longer exist
}

// IsEmpty returns true if there are no changes
func (d DiffResult) IsEmpty() bool {
	return len(d.Added) == 0 && len(d.Modified) == 0 && len(d.Removed) == 0
}

// TotalChanges returns the number of changed beats
func (d DiffResult) TotalChanges() int {
	return len(d.Added) + len(d.Modified) + len(d.Removed)
}

// ComputeDiff compares beats against cached hashes to find changes
func ComputeDiff(beats []model.Beat, cache *model.Cache) DiffResult {
	result := DiffResult{}

	if cache == nil || cache.BeatHashes == nil {
		for _, b := range beats {
			result.Added = append(result.Added, b.ID)
		}
		return result
	}

	currentIDs := make(map[string]bool)
	for _, b := range beats {
		currentIDs[b.ID] = true
		hash := model.HashBeat(b)

		if cachedHash, exists := cache.BeatHashes[b.ID]; !exists {
			result.Added = append(result.Added, b.ID)
		} else if cachedHash != hash {
			result.Modified = append(result.Modified, b.ID)
		}
	}

	for id := range cache.BeatHashes {
		if !currentIDs[id] {
			result.Removed = append(result.Removed, id)
		}
	}

	return result
}

// NeedsFullRebuild determines if incremental update is viable
func NeedsFullRebuild(cache *model.Cache, beats []model.Beat) bool {
	if cache == nil {
		return true
	}

	if cache.Version != model.CacheVersion {
		return true
	}

	if cache.BeatHashes == nil || len(cache.BeatHashes) == 0 {
		return true
	}

	diff := ComputeDiff(beats, cache)
	total := len(beats)
	if total == 0 {
		return false
	}

	changeRatio := float64(diff.TotalChanges()) / float64(total)
	return changeRatio > 0.5
}

// ApplyDiff updates cache incrementally for changed beats
func ApplyDiff(cache *model.Cache, diff DiffResult, beats []model.Beat, updateFn func(beatID string, beat *model.Beat)) {
	beatMap := make(map[string]*model.Beat)
	for i := range beats {
		beatMap[beats[i].ID] = &beats[i]
	}

	if cache.BeatHashes == nil {
		cache.BeatHashes = make(map[string]string)
	}

	for _, id := range diff.Added {
		if b := beatMap[id]; b != nil {
			cache.BeatHashes[id] = model.HashBeat(*b)
			updateFn(id, b)
		}
	}

	for _, id := range diff.Modified {
		if b := beatMap[id]; b != nil {
			cache.BeatHashes[id] = model.HashBeat(*b)
			updateFn(id, b)
		}
	}

	for _, id := range diff.Removed {
		delete(cache.BeatHashes, id)
		delete(cache.Taxonomies, id)
		delete(cache.Ripeness, id)
		delete(cache.ViewStats, id)
	}
}
