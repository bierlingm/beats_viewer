package components

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	"github.com/charmbracelet/lipgloss"
)

var (
	projectTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		MarginBottom(1)

	projectItemStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	projectSelectedStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#73F59F"))

	projectCountStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262"))
)

type ProjectPicker struct {
	width    int
	height   int
	projects []model.Project
	cursor   int
	selected *string // nil = all projects, "." = current dir
	visible  bool
}

func NewProjectPicker(width, height int) *ProjectPicker {
	return &ProjectPicker{
		width:    width,
		height:   height,
		cursor:   0,
		visible:  false,
	}
}

func (p *ProjectPicker) SetSize(width, height int) {
	p.width = width
	p.height = height
}

func (p *ProjectPicker) SetProjects(projects []model.Project) {
	p.projects = projects
}

func (p *ProjectPicker) Show() {
	p.visible = true
}

func (p *ProjectPicker) Hide() {
	p.visible = false
}

func (p *ProjectPicker) IsVisible() bool {
	return p.visible
}

func (p *ProjectPicker) Selected() *string {
	return p.selected
}

func (p *ProjectPicker) SelectedName() string {
	if p.selected == nil {
		return ""
	}
	return *p.selected
}

func (p *ProjectPicker) ClearFilter() {
	p.selected = nil
}

func (p *ProjectPicker) CursorUp() {
	if p.cursor > 0 {
		p.cursor--
	}
}

func (p *ProjectPicker) CursorDown() {
	// +2 for "All" and "Current Dir"
	maxPos := len(p.projects) + 1
	if p.cursor < maxPos {
		p.cursor++
	}
}

func (p *ProjectPicker) Select() {
	if p.cursor == 0 {
		// All projects
		p.selected = nil
	} else if p.cursor == 1 {
		// Current directory
		dot := "."
		p.selected = &dot
	} else {
		idx := p.cursor - 2
		if idx >= 0 && idx < len(p.projects) {
			name := p.projects[idx].Name
			p.selected = &name
		}
	}
	p.visible = false
}

func (p *ProjectPicker) View() string {
	if !p.visible {
		return ""
	}

	var sb strings.Builder

	sb.WriteString(projectTitleStyle.Render("Select Project"))
	sb.WriteString("\n\n")

	// All projects option
	allSelected := p.selected == nil
	sb.WriteString(p.renderItem("All Projects", -1, allSelected, p.cursor == 0))
	sb.WriteString("\n")

	// Current directory option
	dotSelected := p.selected != nil && *p.selected == "."
	sb.WriteString(p.renderItem("Current Directory", -1, dotSelected, p.cursor == 1))
	sb.WriteString("\n")

	sb.WriteString("\n")

	// Individual projects
	for i, proj := range p.projects {
		selected := p.selected != nil && *p.selected == proj.Name
		focused := p.cursor == i+2
		displayName := proj.Name
		if len(displayName) > 25 {
			displayName = displayName[:22] + "..."
		}
		sb.WriteString(p.renderItem(displayName, proj.BeatCount, selected, focused))
		sb.WriteString("\n")
	}

	sb.WriteString("\n")
	sb.WriteString(projectCountStyle.Render("j/k: navigate  enter: select  esc: cancel"))

	boxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(p.width).
		MaxHeight(p.height)

	return boxStyle.Render(sb.String())
}

func (p *ProjectPicker) renderItem(label string, count int, selected, focused bool) string {
	bullet := "○"
	if selected {
		bullet = "●"
	}

	var line string
	if selected {
		line = projectSelectedStyle.Render(fmt.Sprintf("%s %s", bullet, label))
	} else {
		line = projectItemStyle.Render(fmt.Sprintf("%s %s", bullet, label))
	}

	if count >= 0 {
		countStr := projectCountStyle.Render(fmt.Sprintf("(%d)", count))
		line = fmt.Sprintf("%s %s", line, countStr)
	}

	if focused {
		line = lipgloss.NewStyle().
			Background(lipgloss.Color("#383838")).
			Width(p.width - 6).
			Render(line)
	}

	return line
}

// FilterBeatsByProject filters beats by project name
// If projectFilter is nil, returns all beats
// If projectFilter is ".", returns beats from current working directory project
func FilterBeatsByProject(beats []model.EnrichedBeat, beatToProject map[string]string, projectFilter *string, currentDir string) []model.EnrichedBeat {
	if projectFilter == nil {
		return beats
	}

	targetProject := *projectFilter
	if targetProject == "." {
		// Use the current directory name as project
		targetProject = filepath.Base(currentDir)
	}

	var filtered []model.EnrichedBeat
	for _, b := range beats {
		if beatToProject[b.ID] == targetProject {
			filtered = append(filtered, b)
		}
	}
	return filtered
}
