package attention

import (
	"testing"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

func TestDetectBursts(t *testing.T) {
	now := time.Now()
	config := DefaultActivationConfig()
	config.BurstThreshold = 3

	// Create a cluster with 4 beats in last 72h (should trigger burst)
	beats := []model.Beat{
		{ID: "1", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "2", CreatedAt: now.Add(-12 * time.Hour)},
		{ID: "3", CreatedAt: now.Add(-24 * time.Hour)},
		{ID: "4", CreatedAt: now.Add(-48 * time.Hour)},
	}

	clusters := []model.Cluster{
		{ID: "c1", Name: "Test Cluster", BeatIDs: []string{"1", "2", "3", "4"}},
	}

	activations := DetectBursts(beats, clusters, config)

	if len(activations) != 1 {
		t.Errorf("expected 1 burst activation, got %d", len(activations))
	}

	if len(activations) > 0 {
		a := activations[0]
		if a.Type != ActivationBurst {
			t.Errorf("expected burst type, got %v", a.Type)
		}
		if a.BeatCount != 4 {
			t.Errorf("expected 4 beats, got %d", a.BeatCount)
		}
	}
}

func TestDetectReactivations(t *testing.T) {
	now := time.Now()
	config := DefaultActivationConfig()
	config.DormancyPeriod = 14 * 24 * time.Hour

	// Create beats with a gap (dormant for 20 days, then new activity)
	beats := []model.Beat{
		{ID: "new", CreatedAt: now.Add(-1 * time.Hour)},              // New activity
		{ID: "old1", CreatedAt: now.Add(-25 * 24 * time.Hour)},       // Before dormancy period
		{ID: "old2", CreatedAt: now.Add(-30 * 24 * time.Hour)},       // Before dormancy period
	}

	clusters := []model.Cluster{
		{ID: "c1", Name: "Reactivating Cluster", BeatIDs: []string{"new", "old1", "old2"}},
	}

	activations := DetectReactivations(beats, clusters, config)

	if len(activations) != 1 {
		t.Errorf("expected 1 reactivation, got %d", len(activations))
	}

	if len(activations) > 0 && activations[0].Type != ActivationReactivation {
		t.Errorf("expected reactivation type, got %v", activations[0].Type)
	}
}

func TestDetectActivations_NoFalsePositives(t *testing.T) {
	now := time.Now()
	config := DefaultActivationConfig()

	// Only 2 beats in window (below threshold)
	beats := []model.Beat{
		{ID: "1", CreatedAt: now.Add(-12 * time.Hour)},
		{ID: "2", CreatedAt: now.Add(-24 * time.Hour)},
	}

	clusters := []model.Cluster{
		{ID: "c1", Name: "Quiet Cluster", BeatIDs: []string{"1", "2"}},
	}

	activations := DetectActivations(beats, clusters, config)

	if len(activations) != 0 {
		t.Errorf("expected no activations for quiet cluster, got %d", len(activations))
	}
}

func TestActivationType_String(t *testing.T) {
	tests := []struct {
		at   ActivationType
		want string
	}{
		{ActivationBurst, "burst"},
		{ActivationReactivation, "reactivation"},
		{ActivationEmergent, "emergent"},
	}

	for _, tt := range tests {
		if got := tt.at.String(); got != tt.want {
			t.Errorf("ActivationType.String() = %v, want %v", got, tt.want)
		}
	}
}
