package views

import (
	"fmt"
	"strings"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/attention"
	"github.com/bierlingm/beats_viewer/pkg/divergence"
	"github.com/bierlingm/beats_viewer/pkg/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	dashSectionTitle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	dashSectionSelectedTitle = lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#FAFAFA")).
					Background(lipgloss.Color("#7D56F4"))

	dashContentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"}).
				PaddingLeft(2)

	dashMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	dashHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	dashWarningStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F39C12"))

	dashBarFilled = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	dashBarEmpty = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#383838"))
)

// DashboardSection represents a navigable section
type DashboardSection int

const (
	SectionActivating DashboardSection = iota
	SectionDrift
	SectionCrystallizing
	SectionEmerging
	SectionDivergence
	SectionDormant
	sectionCount
)

func (s DashboardSection) String() string {
	switch s {
	case SectionActivating:
		return "NOW ACTIVATING"
	case SectionDrift:
		return "ATTENTION DRIFT"
	case SectionCrystallizing:
		return "CRYSTALLIZING"
	case SectionEmerging:
		return "EMERGING"
	case SectionDivergence:
		return "AGENT DIVERGENCE"
	case SectionDormant:
		return "DORMANT"
	default:
		return ""
	}
}

// DashboardView displays attention state
type DashboardView struct {
	width            int
	height           int
	focused          DashboardSection
	expanded         bool
	state            *attention.AttentionState
	crystallizations []model.CrystallizationInference
	divergence       *divergence.DivergenceReport
	scrollOffset     int
}

// NewDashboardView creates a new dashboard
func NewDashboardView(width, height int) *DashboardView {
	return &DashboardView{
		width:   width,
		height:  height,
		focused: SectionActivating,
	}
}

// SetState updates the attention state
func (d *DashboardView) SetState(state *attention.AttentionState) {
	d.state = state
}

// SetCrystallizations updates crystallization data
func (d *DashboardView) SetCrystallizations(crysts []model.CrystallizationInference) {
	d.crystallizations = crysts
}

// SetDivergence updates divergence data
func (d *DashboardView) SetDivergence(div *divergence.DivergenceReport) {
	d.divergence = div
}

// Update handles input
func (d *DashboardView) Update(msg tea.Msg) (*DashboardView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if d.focused < sectionCount-1 {
				d.focused++
			}
		case "k", "up":
			if d.focused > 0 {
				d.focused--
			}
		case "enter":
			d.expanded = !d.expanded
		}
	}
	return d, nil
}

// View renders the dashboard
func (d *DashboardView) View() string {
	var sections []string

	sections = append(sections, d.renderSection(SectionActivating, d.renderActivating()))
	sections = append(sections, d.renderSection(SectionDrift, d.renderDrift()))
	sections = append(sections, d.renderSection(SectionCrystallizing, d.renderCrystallizing()))
	sections = append(sections, d.renderSection(SectionEmerging, d.renderEmerging()))
	sections = append(sections, d.renderSection(SectionDivergence, d.renderDivergence()))
	sections = append(sections, d.renderSection(SectionDormant, d.renderDormant()))

	content := strings.Join(sections, "\n")

	title := dashSectionTitle.Render("─ Attention Dashboard ─")
	help := dashMutedStyle.Render("A:dashboard L:list T:timeline C:clusters D:drift H:heartbeat ?:help")

	return fmt.Sprintf("%s\n%s\n%s", title, content, help)
}

func (d *DashboardView) renderSection(section DashboardSection, content string) string {
	titleStyle := dashSectionTitle
	if d.focused == section {
		titleStyle = dashSectionSelectedTitle
	}

	marker := "▌"
	if d.focused == section {
		marker = "▶"
	}

	title := fmt.Sprintf("%s %s", marker, section.String())
	header := titleStyle.Render(title)

	if content == "" {
		content = dashMutedStyle.Render("  (none)")
	}

	return fmt.Sprintf("%s\n%s", header, content)
}

func (d *DashboardView) renderActivating() string {
	if d.state == nil || len(d.state.Activations) == 0 {
		return ""
	}

	var lines []string
	limit := 3
	if d.expanded && d.focused == SectionActivating {
		limit = len(d.state.Activations)
	}

	for i, a := range d.state.Activations {
		if i >= limit {
			break
		}
		bar := d.renderBar(a.BeatCount, 10)
		line := fmt.Sprintf("│ %q %s %d beats in %s",
			truncate(a.ClusterName, 30), bar, a.BeatCount, formatDuration(a.Window))
		lines = append(lines, dashContentStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (d *DashboardView) renderDrift() string {
	if d.state == nil || d.state.DriftReport == nil {
		return ""
	}

	dr := d.state.DriftReport
	var lines []string

	// Rising
	for i, item := range dr.Rising {
		if i >= 2 {
			break
		}
		line := fmt.Sprintf("│ ↑ %s +%d", truncate(item.ClusterName, 20), item.CurrentCount-item.PriorCount)
		lines = append(lines, dashHighlightStyle.Render(line))
	}

	// Fading
	for i, item := range dr.Fading {
		if i >= 2 {
			break
		}
		line := fmt.Sprintf("│ ↓ %s %.0f%%", truncate(item.ClusterName, 20), item.ChangePercent)
		lines = append(lines, dashWarningStyle.Render(line))
	}

	if len(lines) == 0 {
		return ""
	}

	return strings.Join(lines, "    ")
}

func (d *DashboardView) renderCrystallizing() string {
	if len(d.crystallizations) == 0 {
		return ""
	}

	var lines []string
	limit := 2
	if d.expanded && d.focused == SectionCrystallizing {
		limit = len(d.crystallizations)
	}

	for i, c := range d.crystallizations {
		if i >= limit {
			break
		}
		line := fmt.Sprintf("│ • %d beats → %s          confidence: %.0f%%",
			len(c.BeatIDs), truncate(c.BeadTitle, 25), c.Confidence*100)
		lines = append(lines, dashContentStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (d *DashboardView) renderEmerging() string {
	if d.state == nil || len(d.state.Emergent) == 0 {
		return ""
	}

	var lines []string
	for i, e := range d.state.Emergent {
		if i >= 2 {
			break
		}
		// Use CommonTerms as theme description
		theme := e.Signal
		if len(e.CommonTerms) > 0 {
			theme = strings.Join(e.CommonTerms[:min(3, len(e.CommonTerms))], ", ")
		}
		line := fmt.Sprintf("│ ? %q - %d beats, no cluster yet",
			truncate(theme, 30), e.BeatCount)
		lines = append(lines, dashContentStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (d *DashboardView) renderDivergence() string {
	if d.divergence == nil {
		return ""
	}

	var lines []string

	// Blind spots
	if len(d.divergence.BlindSpots) > 0 {
		spot := d.divergence.BlindSpots[0]
		var agentCount int
		for _, item := range d.divergence.AgentOnly {
			if item.Topic == spot {
				agentCount = item.AgentCount
				break
			}
		}
		line := fmt.Sprintf("│ You're missing: %q (agents: %d, you: 0)", spot, agentCount)
		lines = append(lines, dashWarningStyle.Render(line))
	}

	// Human-only highlights
	if len(d.divergence.HumanOnly) > 0 {
		item := d.divergence.HumanOnly[0]
		line := fmt.Sprintf("│ You notice more: %q (you: %d, agents: 0)", item.Topic, item.HumanCount)
		lines = append(lines, dashHighlightStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (d *DashboardView) renderDormant() string {
	if d.state == nil || len(d.state.Dormant) == 0 {
		return ""
	}

	var lines []string
	for i, dorm := range d.state.Dormant {
		if i >= 2 {
			break
		}
		line := fmt.Sprintf("│ • %q - %d ripe beats, %d days inactive",
			truncate(dorm.ClusterName, 25), dorm.RipeBeatCount, dorm.InactiveDays)
		lines = append(lines, dashContentStyle.Render(line))
	}

	return strings.Join(lines, "\n")
}

func (d *DashboardView) renderBar(value, max int) string {
	filled := value
	if filled > max {
		filled = max
	}
	empty := max - filled

	bar := dashBarFilled.Render(strings.Repeat("█", filled)) +
		dashBarEmpty.Render(strings.Repeat("░", empty))
	return bar
}

// SetSize updates dimensions
func (d *DashboardView) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// SelectedSection returns the currently focused section
func (d *DashboardView) SelectedSection() DashboardSection {
	return d.focused
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatDuration(dur time.Duration) string {
	hours := int(dur.Hours())
	if hours < 24 {
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%dd", hours/24)
}
