package ui

import (
	"fmt"
	"io"

	"beats_viewer/pkg/model"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type BeatItem struct {
	beat    model.Beat
	project string
}

func (i BeatItem) Title() string {
	return i.beat.ID
}

func (i BeatItem) Description() string {
	return i.beat.ContentPreview(60)
}

func (i BeatItem) FilterValue() string {
	return i.beat.Content + " " + i.beat.Impetus.Label + " " + i.beat.ID
}

func (i BeatItem) Beat() model.Beat {
	return i.beat
}

func (i BeatItem) Project() string {
	return i.project
}

type BeatDelegate struct {
	width int
}

func NewBeatDelegate() BeatDelegate {
	return BeatDelegate{width: 80}
}

func (d BeatDelegate) SetWidth(w int) BeatDelegate {
	d.width = w
	return d
}

func (d BeatDelegate) Height() int {
	return 2
}

func (d BeatDelegate) Spacing() int {
	return 0
}

func (d BeatDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d BeatDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	bi, ok := item.(BeatItem)
	if !ok {
		return
	}

	beat := bi.beat
	isSelected := index == m.Index()

	idWidth := 20
	impetusWidth := 20
	contentWidth := d.width - idWidth - impetusWidth - 6

	if contentWidth < 20 {
		contentWidth = 20
	}

	id := Truncate(beat.ID, idWidth)
	impetus := Truncate(beat.ImpetusLabel(), impetusWidth)
	content := Truncate(beat.Content, contentWidth)

	line1 := fmt.Sprintf("%-*s  %-*s", idWidth, id, impetusWidth, impetus)
	line2 := fmt.Sprintf("  %s", content)

	if isSelected {
		line1 = SelectedStyle.Render(line1)
		line2 = SelectedStyle.Render(line2)
	} else {
		line1 = IDStyle.Render(Truncate(beat.ID, idWidth)) + "  " + ImpetusStyle.Render(Truncate(beat.ImpetusLabel(), impetusWidth))
		line2 = "  " + ContentStyle.Render(content)
	}

	fmt.Fprintf(w, "%s\n%s\n", line1, line2)
}

func BeatsToItems(beats []model.Beat, beatToProject map[string]string) []list.Item {
	items := make([]list.Item, len(beats))
	for i, b := range beats {
		project := ""
		if beatToProject != nil {
			project = beatToProject[b.ID]
		}
		items[i] = BeatItem{beat: b, project: project}
	}
	return items
}
