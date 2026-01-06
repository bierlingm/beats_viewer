package ui

import (
	"fmt"
	"strings"

	"beats_viewer/pkg/model"

	"github.com/charmbracelet/bubbles/viewport"
	"github.com/charmbracelet/lipgloss"
)

type DetailView struct {
	viewport viewport.Model
	beat     *model.Beat
	project  string
	width    int
	height   int
}

func NewDetailView(width, height int) DetailView {
	vp := viewport.New(width, height)
	vp.Style = lipgloss.NewStyle().Padding(1)
	return DetailView{
		viewport: vp,
		width:    width,
		height:   height,
	}
}

func (d *DetailView) SetSize(width, height int) {
	d.width = width
	d.height = height
	d.viewport.Width = width
	d.viewport.Height = height
	if d.beat != nil {
		d.viewport.SetContent(d.renderContent())
	}
}

func (d *DetailView) SetBeat(beat *model.Beat, project string) {
	d.beat = beat
	d.project = project
	if beat != nil {
		d.viewport.SetContent(d.renderContent())
		d.viewport.GotoTop()
	}
}

func (d *DetailView) renderContent() string {
	if d.beat == nil {
		return "No beat selected"
	}

	var sb strings.Builder

	sb.WriteString(DetailLabelStyle.Render("ID: "))
	sb.WriteString(DetailValueStyle.Render(d.beat.ID))
	sb.WriteString("\n")

	if d.project != "" {
		sb.WriteString(DetailLabelStyle.Render("Project: "))
		sb.WriteString(ProjectBadgeStyle.Render(d.project))
		sb.WriteString("\n")
	}

	sb.WriteString(DetailLabelStyle.Render("Created: "))
	sb.WriteString(DetailValueStyle.Render(d.beat.CreatedAt.Format("2006-01-02 15:04")))
	sb.WriteString("\n")

	if !d.beat.UpdatedAt.Equal(d.beat.CreatedAt) {
		sb.WriteString(DetailLabelStyle.Render("Updated: "))
		sb.WriteString(DetailValueStyle.Render(d.beat.UpdatedAt.Format("2006-01-02 15:04")))
		sb.WriteString("\n")
	}

	sb.WriteString(DetailLabelStyle.Render("Impetus: "))
	sb.WriteString(ImpetusStyle.Render(d.beat.ImpetusLabel()))
	sb.WriteString("\n")

	if d.beat.Impetus.Raw != "" {
		sb.WriteString(DetailLabelStyle.Render("Raw: "))
		sb.WriteString(DetailValueStyle.Render(d.beat.Impetus.Raw))
		sb.WriteString("\n")
	}

	if len(d.beat.Impetus.Meta) > 0 {
		sb.WriteString(DetailLabelStyle.Render("Meta: "))
		for k, v := range d.beat.Impetus.Meta {
			sb.WriteString(fmt.Sprintf("%s=%s ", k, v))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(DetailLabelStyle.Render("Content:"))
	sb.WriteString("\n")
	sb.WriteString(ContentStyle.Render(d.beat.Content))
	sb.WriteString("\n")

	if len(d.beat.Entities) > 0 {
		sb.WriteString("\n")
		sb.WriteString(DetailLabelStyle.Render("Entities: "))
		sb.WriteString(DetailValueStyle.Render(strings.Join(d.beat.Entities, ", ")))
		sb.WriteString("\n")
	}

	if len(d.beat.References) > 0 {
		sb.WriteString(DetailLabelStyle.Render("References: "))
		sb.WriteString(DetailValueStyle.Render(strings.Join(d.beat.References, ", ")))
		sb.WriteString("\n")
	}

	if len(d.beat.LinkedBeads) > 0 {
		sb.WriteString("\n")
		sb.WriteString(DetailLabelStyle.Render("Linked Beads: "))
		sb.WriteString(DetailValueStyle.Render(strings.Join(d.beat.LinkedBeads, ", ")))
		sb.WriteString("\n")
	} else {
		sb.WriteString("\n")
		sb.WriteString(SubtitleStyle.Render("No linked beads"))
		sb.WriteString("\n")
	}

	return sb.String()
}

func (d *DetailView) View() string {
	return d.viewport.View()
}

func (d *DetailView) ScrollDown() {
	d.viewport.LineDown(1)
}

func (d *DetailView) ScrollUp() {
	d.viewport.LineUp(1)
}

func (d *DetailView) PageDown() {
	d.viewport.HalfViewDown()
}

func (d *DetailView) PageUp() {
	d.viewport.HalfViewUp()
}

func (d *DetailView) GotoTop() {
	d.viewport.GotoTop()
}

func (d *DetailView) GotoBottom() {
	d.viewport.GotoBottom()
}
