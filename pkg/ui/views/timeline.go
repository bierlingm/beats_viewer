package views

import (
	"github.com/bierlingm/beats_viewer/pkg/model"
	"github.com/bierlingm/beats_viewer/pkg/timeline"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TimelineView struct {
	renderer *timeline.TimelineRenderer
	data     *timeline.TimelineData
	beats    []model.EnrichedBeat
	width    int
	height   int
}

func NewTimelineView(width, height int) *TimelineView {
	return &TimelineView{
		renderer: timeline.NewTimelineRenderer(width, height),
		width:    width,
		height:   height,
	}
}

func (tv *TimelineView) SetSize(width, height int) {
	tv.width = width
	tv.height = height
	tv.renderer.SetSize(width, height)
}

func (tv *TimelineView) SetBeats(beats []model.EnrichedBeat) {
	tv.beats = beats
	tv.data = timeline.BuildTimeline(beats, timeline.ZoomMonth)
	tv.renderer.SetData(tv.data)
}

func (tv *TimelineView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			tv.renderer.CursorLeft()
		case "right", "l":
			tv.renderer.CursorRight()
		case "z":
			newZoom := tv.renderer.CycleZoom()
			tv.data = timeline.BuildTimeline(tv.beats, newZoom)
			tv.renderer.SetData(tv.data)
		case "c":
			tv.renderer.ToggleColors()
		}
	}
	return nil
}

func (tv *TimelineView) SelectedBeatIDs() []string {
	if bucket := tv.renderer.SelectedBucket(); bucket != nil {
		return bucket.BeatIDs
	}
	return nil
}

func (tv *TimelineView) View() string {
	return lipgloss.NewStyle().
		Width(tv.width).
		Height(tv.height).
		Render(tv.renderer.Render())
}
