package attention

import (
	"testing"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

func TestBeatsInWindow(t *testing.T) {
	now := time.Now()
	beats := []model.Beat{
		{ID: "old", CreatedAt: now.Add(-48 * time.Hour)},
		{ID: "recent", CreatedAt: now.Add(-12 * time.Hour)},
		{ID: "newest", CreatedAt: now.Add(-1 * time.Hour)},
	}

	// 24h window should return 2 beats
	result := BeatsInWindow(beats, 24*time.Hour)
	if len(result) != 2 {
		t.Errorf("expected 2 beats in 24h window, got %d", len(result))
	}

	// 72h window should return all 3
	result = BeatsInWindow(beats, 72*time.Hour)
	if len(result) != 3 {
		t.Errorf("expected 3 beats in 72h window, got %d", len(result))
	}

	// 6h window should return 1
	result = BeatsInWindow(beats, 6*time.Hour)
	if len(result) != 1 {
		t.Errorf("expected 1 beat in 6h window, got %d", len(result))
	}
}

func TestClassifyChannel(t *testing.T) {
	tests := []struct {
		name    string
		beat    model.Beat
		want    Channel
	}{
		{
			name: "CLI capture",
			beat: model.Beat{Impetus: model.Impetus{Label: "reading"}},
			want: ChannelCLI,
		},
		{
			name: "Agent capture from label",
			beat: model.Beat{Impetus: model.Impetus{Label: "droid-session"}},
			want: ChannelAgent,
		},
		{
			name: "Agent capture from raw",
			beat: model.Beat{Impetus: model.Impetus{Label: "coding", Raw: "factory agent"}},
			want: ChannelAgent,
		},
		{
			name: "Unknown capture",
			beat: model.Beat{Impetus: model.Impetus{}},
			want: ChannelUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyChannel(tt.beat)
			if got != tt.want {
				t.Errorf("ClassifyChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAttentionStream(t *testing.T) {
	now := time.Now()
	beats := []model.Beat{
		{ID: "1", CreatedAt: now.Add(-12 * time.Hour), Impetus: model.Impetus{Label: "reading"}},
		{ID: "2", CreatedAt: now.Add(-24 * time.Hour), Impetus: model.Impetus{Label: "coding"}},
		{ID: "3", CreatedAt: now.Add(-36 * time.Hour), Impetus: model.Impetus{Label: "reading"}},
	}

	clusters := []model.Cluster{
		{ID: "c1", BeatIDs: []string{"1", "3"}},
		{ID: "c2", BeatIDs: []string{"2"}},
	}

	stream := NewAttentionStream(beats, clusters, 48*time.Hour)

	if len(stream.Beats) != 3 {
		t.Errorf("expected 3 beats, got %d", len(stream.Beats))
	}

	if stream.ByCluster["c1"] != 2 {
		t.Errorf("expected 2 beats in c1, got %d", stream.ByCluster["c1"])
	}

	if stream.FlowRate <= 0 {
		t.Errorf("expected positive flow rate, got %f", stream.FlowRate)
	}
}

func TestBeatsInCluster(t *testing.T) {
	beats := []model.Beat{
		{ID: "1"},
		{ID: "2"},
		{ID: "3"},
	}

	cluster := model.Cluster{BeatIDs: []string{"1", "3"}}

	result := BeatsInCluster(beats, cluster)
	if len(result) != 2 {
		t.Errorf("expected 2 beats in cluster, got %d", len(result))
	}
}
