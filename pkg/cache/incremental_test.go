package cache

import (
	"testing"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

func makeBeat(id, content string) model.Beat {
	return model.Beat{
		ID:        id,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestComputeDiff_EmptyCache(t *testing.T) {
	beats := []model.Beat{makeBeat("b1", "content1")}

	diff := ComputeDiff(beats, nil)

	if len(diff.Added) != 1 || diff.Added[0] != "b1" {
		t.Errorf("expected 1 added beat, got %v", diff.Added)
	}
	if len(diff.Modified) != 0 {
		t.Errorf("expected 0 modified, got %v", diff.Modified)
	}
	if len(diff.Removed) != 0 {
		t.Errorf("expected 0 removed, got %v", diff.Removed)
	}
}

func TestComputeDiff_NoChanges(t *testing.T) {
	b := makeBeat("b1", "content1")
	beats := []model.Beat{b}
	cache := &model.Cache{
		BeatHashes: map[string]string{
			"b1": model.HashBeat(b),
		},
	}

	diff := ComputeDiff(beats, cache)

	if !diff.IsEmpty() {
		t.Errorf("expected empty diff, got added=%v mod=%v rem=%v", diff.Added, diff.Modified, diff.Removed)
	}
}

func TestComputeDiff_AddedBeat(t *testing.T) {
	b1 := makeBeat("b1", "content1")
	b2 := makeBeat("b2", "content2")
	beats := []model.Beat{b1, b2}
	cache := &model.Cache{
		BeatHashes: map[string]string{
			"b1": model.HashBeat(b1),
		},
	}

	diff := ComputeDiff(beats, cache)

	if len(diff.Added) != 1 || diff.Added[0] != "b2" {
		t.Errorf("expected b2 added, got %v", diff.Added)
	}
}

func TestComputeDiff_ModifiedBeat(t *testing.T) {
	b := makeBeat("b1", "content1")
	beats := []model.Beat{b}
	cache := &model.Cache{
		BeatHashes: map[string]string{
			"b1": "oldhash",
		},
	}

	diff := ComputeDiff(beats, cache)

	if len(diff.Modified) != 1 || diff.Modified[0] != "b1" {
		t.Errorf("expected b1 modified, got %v", diff.Modified)
	}
}

func TestComputeDiff_RemovedBeat(t *testing.T) {
	beats := []model.Beat{}
	cache := &model.Cache{
		BeatHashes: map[string]string{
			"b1": "somehash",
		},
	}

	diff := ComputeDiff(beats, cache)

	if len(diff.Removed) != 1 || diff.Removed[0] != "b1" {
		t.Errorf("expected b1 removed, got %v", diff.Removed)
	}
}

func TestNeedsFullRebuild_NilCache(t *testing.T) {
	beats := []model.Beat{makeBeat("b1", "c")}
	if !NeedsFullRebuild(nil, beats) {
		t.Error("expected full rebuild for nil cache")
	}
}

func TestNeedsFullRebuild_VersionMismatch(t *testing.T) {
	beats := []model.Beat{makeBeat("b1", "c")}
	cache := &model.Cache{
		Version:    "0.0.1",
		BeatHashes: map[string]string{"b1": "x"},
	}
	if !NeedsFullRebuild(cache, beats) {
		t.Error("expected full rebuild for version mismatch")
	}
}

func TestNeedsFullRebuild_NoBeatHashes(t *testing.T) {
	beats := []model.Beat{makeBeat("b1", "c")}
	cache := &model.Cache{
		Version: model.CacheVersion,
	}
	if !NeedsFullRebuild(cache, beats) {
		t.Error("expected full rebuild when no beat hashes")
	}
}

func TestNeedsFullRebuild_SmallChange(t *testing.T) {
	b1 := makeBeat("b1", "c1")
	b2 := makeBeat("b2", "c2")
	beats := []model.Beat{b1, b2, makeBeat("b3", "c3")}
	cache := &model.Cache{
		Version: model.CacheVersion,
		BeatHashes: map[string]string{
			"b1": model.HashBeat(b1),
			"b2": model.HashBeat(b2),
		},
	}
	if NeedsFullRebuild(cache, beats) {
		t.Error("expected incremental update for small change (1/3 < 50%)")
	}
}

func TestApplyDiff_Updates(t *testing.T) {
	cache := &model.Cache{
		BeatHashes: map[string]string{"b1": "old"},
		Taxonomies: map[string]model.Taxonomy{"b1": {}},
		Ripeness:   map[string]float64{"b1": 0.5},
		ViewStats:  map[string]model.ViewStat{"b1": {}},
	}
	b2 := makeBeat("b2", "new")
	beats := []model.Beat{b2}
	diff := DiffResult{
		Added:   []string{"b2"},
		Removed: []string{"b1"},
	}

	called := make(map[string]bool)
	ApplyDiff(cache, diff, beats, func(id string, _ *model.Beat) {
		called[id] = true
	})

	if !called["b2"] {
		t.Error("expected updateFn called for b2")
	}
	if _, ok := cache.BeatHashes["b1"]; ok {
		t.Error("expected b1 removed from BeatHashes")
	}
	if _, ok := cache.BeatHashes["b2"]; !ok {
		t.Error("expected b2 added to BeatHashes")
	}
}
