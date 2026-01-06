package components

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	"github.com/charmbracelet/lipgloss"
)

var (
	facetTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			MarginBottom(1)

	facetItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	facetSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#73F59F"))

	facetCountStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))
)

type FacetSidebar struct {
	width  int
	height int

	channelCounts map[model.Channel]int
	sourceCounts  map[model.Source]int

	selectedChannel *model.Channel
	selectedSource  *model.Source

	focusChannel bool
	cursorPos    int
}

func NewFacetSidebar(width, height int) *FacetSidebar {
	return &FacetSidebar{
		width:         width,
		height:        height,
		channelCounts: make(map[model.Channel]int),
		sourceCounts:  make(map[model.Source]int),
		focusChannel:  true,
		cursorPos:     0,
	}
}

func (f *FacetSidebar) SetSize(width, height int) {
	f.width = width
	f.height = height
}

func (f *FacetSidebar) UpdateCounts(beats []model.EnrichedBeat) {
	f.channelCounts = make(map[model.Channel]int)
	f.sourceCounts = make(map[model.Source]int)

	for _, eb := range beats {
		f.channelCounts[eb.Taxonomy.Channel]++
		f.sourceCounts[eb.Taxonomy.Source]++
	}
}

func (f *FacetSidebar) SelectedChannel() *model.Channel {
	return f.selectedChannel
}

func (f *FacetSidebar) SelectedSource() *model.Source {
	return f.selectedSource
}

func (f *FacetSidebar) SelectChannelByNumber(n int) {
	channels := model.AllChannels()
	if n >= 1 && n <= len(channels) {
		ch := channels[n-1]
		if f.selectedChannel != nil && *f.selectedChannel == ch {
			f.selectedChannel = nil
		} else {
			f.selectedChannel = &ch
		}
	}
}

func (f *FacetSidebar) ClearFilters() {
	f.selectedChannel = nil
	f.selectedSource = nil
}

func (f *FacetSidebar) CursorUp() {
	if f.cursorPos > 0 {
		f.cursorPos--
	}
}

func (f *FacetSidebar) CursorDown() {
	maxPos := len(model.AllChannels()) + len(model.AllSources()) + 1
	if f.cursorPos < maxPos {
		f.cursorPos++
	}
}

func (f *FacetSidebar) ToggleSelection() {
	channels := model.AllChannels()
	sources := model.AllSources()

	if f.cursorPos == 0 {
		f.selectedChannel = nil
	} else if f.cursorPos <= len(channels) {
		ch := channels[f.cursorPos-1]
		if f.selectedChannel != nil && *f.selectedChannel == ch {
			f.selectedChannel = nil
		} else {
			f.selectedChannel = &ch
		}
	} else if f.cursorPos == len(channels)+1 {
		f.selectedSource = nil
	} else {
		idx := f.cursorPos - len(channels) - 2
		if idx >= 0 && idx < len(sources) {
			src := sources[idx]
			if f.selectedSource != nil && *f.selectedSource == src {
				f.selectedSource = nil
			} else {
				f.selectedSource = &src
			}
		}
	}
}

func (f *FacetSidebar) TotalCount() int {
	total := 0
	for _, c := range f.channelCounts {
		total += c
	}
	return total
}

func (f *FacetSidebar) View() string {
	var sb strings.Builder

	sb.WriteString(facetTitleStyle.Render("Channel"))
	sb.WriteString("\n")

	total := f.TotalCount()
	allSelected := f.selectedChannel == nil
	allLine := f.renderFacetItem("All", total, allSelected, f.cursorPos == 0)
	sb.WriteString(allLine)
	sb.WriteString("\n")

	channels := model.AllChannels()
	for i, ch := range channels {
		count := f.channelCounts[ch]
		selected := f.selectedChannel != nil && *f.selectedChannel == ch
		focused := f.cursorPos == i+1
		line := f.renderFacetItem(fmt.Sprintf("%d %s", i+1, ch.String()), count, selected, focused)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(facetTitleStyle.Render("Source"))
	sb.WriteString("\n")

	allSourceSelected := f.selectedSource == nil
	sourceAllFocused := f.cursorPos == len(channels)+1
	sb.WriteString(f.renderFacetItem("All", total, allSourceSelected, sourceAllFocused))
	sb.WriteString("\n")

	sources := model.AllSources()
	for i, src := range sources {
		count := f.sourceCounts[src]
		selected := f.selectedSource != nil && *f.selectedSource == src
		focused := f.cursorPos == len(channels)+2+i
		line := f.renderFacetItem(src.String(), count, selected, focused)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Width(f.width).
		MaxHeight(f.height).
		Padding(0, 1).
		Render(sb.String())
}

func (f *FacetSidebar) renderFacetItem(label string, count int, selected, focused bool) string {
	bullet := "○"
	if selected {
		bullet = "●"
	}

	countStr := facetCountStyle.Render(fmt.Sprintf("(%d)", count))

	var line string
	if selected {
		line = facetSelectedStyle.Render(fmt.Sprintf("%s %s", bullet, label))
	} else {
		line = facetItemStyle.Render(fmt.Sprintf("%s %s", bullet, label))
	}

	result := fmt.Sprintf("%s %s", line, countStr)

	if focused {
		result = lipgloss.NewStyle().
			Background(lipgloss.Color("#383838")).
			Render(result)
	}

	return result
}

func FilterByFacets(beats []model.EnrichedBeat, channel *model.Channel, source *model.Source) []model.EnrichedBeat {
	if channel == nil && source == nil {
		return beats
	}

	var filtered []model.EnrichedBeat
	for _, eb := range beats {
		if channel != nil && eb.Taxonomy.Channel != *channel {
			continue
		}
		if source != nil && eb.Taxonomy.Source != *source {
			continue
		}
		filtered = append(filtered, eb)
	}
	return filtered
}
