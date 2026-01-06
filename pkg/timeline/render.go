package timeline

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	"github.com/charmbracelet/lipgloss"
)

var (
	timelineTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	timelineLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				Width(6)

	timelineDotStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	timelineSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#383838"))

	channelColors = map[model.Channel]lipgloss.Color{
		model.ChannelCoaching:    lipgloss.Color("#FF6B6B"),
		model.ChannelResearch:    lipgloss.Color("#4ECDC4"),
		model.ChannelDiscovery:   lipgloss.Color("#45B7D1"),
		model.ChannelDevelopment: lipgloss.Color("#96CEB4"),
		model.ChannelReflection:  lipgloss.Color("#FFEAA7"),
		model.ChannelReference:   lipgloss.Color("#DDA0DD"),
		model.ChannelMilestone:   lipgloss.Color("#F39C12"),
	}
)

type TimelineRenderer struct {
	data          *TimelineData
	width         int
	height        int
	cursorPos     int
	scrollOffset  int
	showColors    bool
}

func NewTimelineRenderer(width, height int) *TimelineRenderer {
	return &TimelineRenderer{
		width:      width,
		height:     height,
		showColors: false,
	}
}

func (tr *TimelineRenderer) SetData(data *TimelineData) {
	tr.data = data
	tr.cursorPos = 0
	tr.scrollOffset = 0
}

func (tr *TimelineRenderer) SetSize(width, height int) {
	tr.width = width
	tr.height = height
}

func (tr *TimelineRenderer) CursorLeft() {
	if tr.cursorPos > 0 {
		tr.cursorPos--
		tr.ensureVisible()
	}
}

func (tr *TimelineRenderer) CursorRight() {
	if tr.data != nil && tr.cursorPos < len(tr.data.Buckets)-1 {
		tr.cursorPos++
		tr.ensureVisible()
	}
}

func (tr *TimelineRenderer) ensureVisible() {
	visibleWidth := tr.width - 10
	if tr.cursorPos < tr.scrollOffset {
		tr.scrollOffset = tr.cursorPos
	} else if tr.cursorPos >= tr.scrollOffset+visibleWidth {
		tr.scrollOffset = tr.cursorPos - visibleWidth + 1
	}
}

func (tr *TimelineRenderer) ToggleColors() {
	tr.showColors = !tr.showColors
}

func (tr *TimelineRenderer) CycleZoom() ZoomLevel {
	if tr.data == nil {
		return ZoomMonth
	}
	next := (tr.data.ZoomLevel + 1) % 4
	return next
}

func (tr *TimelineRenderer) SelectedBucket() *TimelineBucket {
	if tr.data == nil {
		return nil
	}
	return tr.data.GetBucketAt(tr.cursorPos)
}

func (tr *TimelineRenderer) Render() string {
	if tr.data == nil || len(tr.data.Buckets) == 0 {
		return timelineTitleStyle.Render("No timeline data")
	}

	var sb strings.Builder

	title := fmt.Sprintf("Timeline (%s) - %d buckets", tr.data.ZoomLevel.String(), len(tr.data.Buckets))
	sb.WriteString(timelineTitleStyle.Render(title))
	sb.WriteString("\n\n")

	maxCount := tr.data.MaxBeatCount()
	if maxCount == 0 {
		maxCount = 1
	}

	bucketsByMonth := groupBucketsByMonth(tr.data.Buckets)

	visibleHeight := tr.height - 6
	monthCount := 0

	for _, monthData := range bucketsByMonth {
		if monthCount >= visibleHeight {
			break
		}

		label := timelineLabelStyle.Render(monthData.Label)
		dots := tr.renderDots(monthData.Buckets, maxCount)

		line := fmt.Sprintf("%s %s", label, dots)
		sb.WriteString(line)
		sb.WriteString("\n")
		monthCount++
	}

	sb.WriteString("\n")

	if bucket := tr.SelectedBucket(); bucket != nil {
		info := fmt.Sprintf("Selected: %s (%d beats)", bucket.Date.Format("2006-01-02"), bucket.BeatCount)
		sb.WriteString(timelineSelectedStyle.Render(info))
		sb.WriteString("\n")
	}

	help := "←/→ Navigate  z Zoom  c Colors  Enter Select"
	sb.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(help))

	return sb.String()
}

func (tr *TimelineRenderer) renderDots(buckets []TimelineBucket, maxCount int) string {
	var dots strings.Builder

	for i, bucket := range buckets {
		globalIdx := tr.findGlobalIndex(bucket)
		isSelected := globalIdx == tr.cursorPos

		intensity := float64(bucket.BeatCount) / float64(maxCount)
		dot := densityChar(intensity)

		var color lipgloss.Color
		if tr.showColors && len(bucket.ByChannel) > 0 {
			color = dominantChannelColor(bucket.ByChannel)
		} else {
			color = lipgloss.Color("#73F59F")
		}

		style := lipgloss.NewStyle().Foreground(color)
		if isSelected {
			style = style.Bold(true).Background(lipgloss.Color("#7D56F4"))
		}

		dots.WriteString(style.Render(dot))

		if i < len(buckets)-1 {
			dots.WriteString("─")
		}
	}

	return dots.String()
}

func (tr *TimelineRenderer) findGlobalIndex(bucket TimelineBucket) int {
	if tr.data == nil {
		return -1
	}
	for i, b := range tr.data.Buckets {
		if b.Date.Equal(bucket.Date) {
			return i
		}
	}
	return -1
}

func densityChar(intensity float64) string {
	switch {
	case intensity >= 0.8:
		return "●"
	case intensity >= 0.5:
		return "◉"
	case intensity >= 0.2:
		return "○"
	default:
		return "·"
	}
}

func dominantChannelColor(byChannel map[model.Channel]int) lipgloss.Color {
	var maxChannel model.Channel
	maxCount := 0

	for ch, count := range byChannel {
		if count > maxCount {
			maxCount = count
			maxChannel = ch
		}
	}

	if color, ok := channelColors[maxChannel]; ok {
		return color
	}
	return lipgloss.Color("#73F59F")
}

type monthGroup struct {
	Label   string
	Buckets []TimelineBucket
}

func groupBucketsByMonth(buckets []TimelineBucket) []monthGroup {
	groups := make(map[string]*monthGroup)
	var order []string

	for _, b := range buckets {
		key := b.Date.Format("2006-01")
		if _, ok := groups[key]; !ok {
			groups[key] = &monthGroup{Label: b.Date.Format("Jan 06")}
			order = append(order, key)
		}
		groups[key].Buckets = append(groups[key].Buckets, b)
	}

	var result []monthGroup
	for _, key := range order {
		result = append(result, *groups[key])
	}
	return result
}
