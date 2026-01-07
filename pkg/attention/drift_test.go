package attention

import (
	"testing"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

func TestCategorizeDrift(t *testing.T) {
	tests := []struct {
		name      string
		current   int
		prior     int
		threshold float64
		want      DriftDirection
	}{
		{"emerged", 5, 0, 0.2, DriftEmerged},
		{"vanished", 0, 5, 0.2, DriftVanished},
		{"rising", 10, 5, 0.2, DriftRising},
		{"fading", 3, 10, 0.2, DriftFading},
		{"stable_same", 5, 5, 0.2, DriftStable},
		{"stable_small_change", 6, 5, 0.2, DriftStable},
		{"both_zero", 0, 0, 0.2, DriftStable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CategorizeDrift(tt.current, tt.prior, tt.threshold)
			if got != tt.want {
				t.Errorf("CategorizeDrift(%d, %d, %f) = %v, want %v",
					tt.current, tt.prior, tt.threshold, got, tt.want)
			}
		})
	}
}

func TestDriftDirection_Symbol(t *testing.T) {
	tests := []struct {
		d    DriftDirection
		want string
	}{
		{DriftRising, "↑"},
		{DriftStable, "→"},
		{DriftFading, "↓"},
		{DriftEmerged, "+"},
		{DriftVanished, "×"},
	}

	for _, tt := range tests {
		if got := tt.d.Symbol(); got != tt.want {
			t.Errorf("DriftDirection.Symbol() = %v, want %v", got, tt.want)
		}
	}
}

func TestComputeDrift(t *testing.T) {
	now := time.Now()
	config := DefaultDriftConfig()
	config.Window = 30 * 24 * time.Hour

	// Create beats with different patterns:
	// Cluster 1: 5 beats current, 2 beats prior (rising)
	// Cluster 2: 1 beat current, 5 beats prior (fading)
	// Cluster 3: only current beats (emerged)
	beats := []model.Beat{
		// Cluster 1 - current window
		{ID: "c1-1", CreatedAt: now.Add(-5 * 24 * time.Hour)},
		{ID: "c1-2", CreatedAt: now.Add(-10 * 24 * time.Hour)},
		{ID: "c1-3", CreatedAt: now.Add(-15 * 24 * time.Hour)},
		{ID: "c1-4", CreatedAt: now.Add(-20 * 24 * time.Hour)},
		{ID: "c1-5", CreatedAt: now.Add(-25 * 24 * time.Hour)},
		// Cluster 1 - prior window
		{ID: "c1-6", CreatedAt: now.Add(-35 * 24 * time.Hour)},
		{ID: "c1-7", CreatedAt: now.Add(-40 * 24 * time.Hour)},

		// Cluster 2 - current (fading)
		{ID: "c2-1", CreatedAt: now.Add(-10 * 24 * time.Hour)},
		// Cluster 2 - prior
		{ID: "c2-2", CreatedAt: now.Add(-35 * 24 * time.Hour)},
		{ID: "c2-3", CreatedAt: now.Add(-40 * 24 * time.Hour)},
		{ID: "c2-4", CreatedAt: now.Add(-45 * 24 * time.Hour)},
		{ID: "c2-5", CreatedAt: now.Add(-50 * 24 * time.Hour)},
		{ID: "c2-6", CreatedAt: now.Add(-55 * 24 * time.Hour)},

		// Cluster 3 - emerged (only current)
		{ID: "c3-1", CreatedAt: now.Add(-5 * 24 * time.Hour)},
		{ID: "c3-2", CreatedAt: now.Add(-10 * 24 * time.Hour)},
	}

	clusters := []model.Cluster{
		{ID: "c1", Name: "Rising", BeatIDs: []string{"c1-1", "c1-2", "c1-3", "c1-4", "c1-5", "c1-6", "c1-7"}},
		{ID: "c2", Name: "Fading", BeatIDs: []string{"c2-1", "c2-2", "c2-3", "c2-4", "c2-5", "c2-6"}},
		{ID: "c3", Name: "Emerged", BeatIDs: []string{"c3-1", "c3-2"}},
	}

	report := ComputeDrift(beats, clusters, config)

	if len(report.Rising) == 0 {
		t.Error("expected at least one rising cluster")
	}

	if len(report.Fading) == 0 {
		t.Error("expected at least one fading cluster")
	}

	if len(report.Emerged) == 0 {
		t.Error("expected at least one emerged cluster")
	}
}

func TestDriftReport_TotalChanges(t *testing.T) {
	report := &DriftReport{
		Rising:   []DriftItem{{}, {}},
		Fading:   []DriftItem{{}},
		Emerged:  []DriftItem{{}},
		Vanished: []DriftItem{{}},
		Stable:   []DriftItem{{}, {}, {}},
	}

	if got := report.TotalChanges(); got != 5 {
		t.Errorf("TotalChanges() = %d, want 5", got)
	}
}
