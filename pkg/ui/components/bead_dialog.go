package components

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/crystallize"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dialogTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	dialogLabelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888"))

	dialogValueStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	dialogHighlightStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#73F59F"))
)

type BeadDialog struct {
	width     int
	height    int
	beatID    string
	bead      crystallize.BeadSuggestion
	Confirmed bool
	Cancelled bool
}

func NewBeadDialog(bead crystallize.BeadSuggestion, beatID string) *BeadDialog {
	return &BeadDialog{
		width:  60,
		height: 20,
		beatID: beatID,
		bead:   bead,
	}
}

func (d *BeadDialog) SetSize(width, height int) {
	d.width = width
	d.height = height
}

func (d BeadDialog) Update(msg tea.Msg) (BeadDialog, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "y":
			d.Confirmed = true
		case "esc", "n", "q":
			d.Cancelled = true
		}
	}
	return d, nil
}

func (d BeadDialog) View() string {
	var sb strings.Builder

	sb.WriteString(dialogTitleStyle.Render("Create Bead from Beat"))
	sb.WriteString("\n\n")

	sb.WriteString(dialogLabelStyle.Render("Beat: "))
	sb.WriteString(dialogValueStyle.Render(d.beatID))
	sb.WriteString("\n\n")

	sb.WriteString(dialogLabelStyle.Render("Title: "))
	sb.WriteString(dialogHighlightStyle.Render(d.bead.Title))
	sb.WriteString("\n\n")

	sb.WriteString(dialogLabelStyle.Render("Type: "))
	sb.WriteString(dialogValueStyle.Render(d.bead.Type))
	sb.WriteString("\n")

	sb.WriteString(dialogLabelStyle.Render("Priority: "))
	sb.WriteString(dialogValueStyle.Render(fmt.Sprintf("P%d", d.bead.Priority)))
	sb.WriteString("\n\n")

	if d.bead.Description != "" {
		desc := d.bead.Description
		if len(desc) > 100 {
			desc = desc[:97] + "..."
		}
		sb.WriteString(dialogLabelStyle.Render("Description: "))
		sb.WriteString(dialogValueStyle.Render(desc))
		sb.WriteString("\n\n")
	}

	sb.WriteString("\n")
	sb.WriteString(dialogLabelStyle.Render("y/enter: create  n/esc: cancel"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(d.width).
		MaxHeight(d.height)

	return boxStyle.Render(sb.String())
}

func (d BeadDialog) BeadID() string {
	return d.beatID
}

func (d BeadDialog) Bead() crystallize.BeadSuggestion {
	return d.bead
}
