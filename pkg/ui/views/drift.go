package views

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/attention"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	driftTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	driftSectionStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#626262"))

	driftRisingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	driftFadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#F39C12"))

	driftStableStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	driftEmergedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#3498DB"))

	driftVanishedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#E74C3C"))
)

// DriftSection represents a section in drift view
type DriftSection int

const (
	DriftSectionRising DriftSection = iota
	DriftSectionStable
	DriftSectionFading
	DriftSectionEmerged
	DriftSectionVanished
	driftSectionCount
)

// DriftView displays full drift analysis
type DriftView struct {
	width        int
	height       int
	report       *attention.DriftReport
	focused      DriftSection
	scrollOffset int
}

// NewDriftView creates a new drift view
func NewDriftView(width, height int) *DriftView {
	return &DriftView{
		width:   width,
		height:  height,
		focused: DriftSectionRising,
	}
}

// SetReport updates the drift report
func (d *DriftView) SetReport(report *attention.DriftReport) {
	d.report = report
}

// Update handles input
func (d *DriftView) Update(msg tea.Msg) (*DriftView, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			if d.focused < driftSectionCount-1 {
				d.focused++
			}
		case "k", "up":
			if d.focused > 0 {
				d.focused--
			}
		}
	}
	return d, nil
}

// View renders the drift view
func (d *DriftView) View() string {
	if d.report == nil {
		return driftSectionStyle.Render("No drift data available")
	}

	var sb strings.Builder

	// Title
	windowDays := int(d.report.Window.Hours() / 24)
	title := driftTitleStyle.Render(fmt.Sprintf("─ Attention Drift (%d days) ─", windowDays))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// Rising
	sb.WriteString(d.renderSection(DriftSectionRising, "↑ RISING", d.report.Rising, driftRisingStyle))
	sb.WriteString("\n")

	// Stable
	sb.WriteString(d.renderSection(DriftSectionStable, "→ STABLE", d.report.Stable, driftStableStyle))
	sb.WriteString("\n")

	// Fading
	sb.WriteString(d.renderSection(DriftSectionFading, "↓ FADING", d.report.Fading, driftFadingStyle))
	sb.WriteString("\n")

	// Emerged
	sb.WriteString(d.renderSection(DriftSectionEmerged, "★ EMERGED", d.report.Emerged, driftEmergedStyle))
	sb.WriteString("\n")

	// Vanished
	sb.WriteString(d.renderSection(DriftSectionVanished, "✕ VANISHED", d.report.Vanished, driftVanishedStyle))

	return sb.String()
}

func (d *DriftView) renderSection(section DriftSection, title string, items []attention.DriftItem, style lipgloss.Style) string {
	var sb strings.Builder

	headerStyle := driftSectionStyle
	if d.focused == section {
		headerStyle = driftTitleStyle
	}

	sb.WriteString(headerStyle.Render(title))
	sb.WriteString("\n")

	if len(items) == 0 {
		sb.WriteString(driftSectionStyle.Render("  (none)"))
		sb.WriteString("\n")
		return sb.String()
	}

	limit := 5
	if d.focused == section {
		limit = 10
	}

	for i, item := range items {
		if i >= limit {
			remaining := len(items) - limit
			sb.WriteString(driftSectionStyle.Render(fmt.Sprintf("  ... and %d more", remaining)))
			sb.WriteString("\n")
			break
		}

		var line string
		switch item.Direction {
		case attention.DriftRising:
			line = fmt.Sprintf("  %s +%d (%d → %d)",
				truncateDrift(item.ClusterName, 25),
				item.CurrentCount-item.PriorCount,
				item.PriorCount,
				item.CurrentCount)
		case attention.DriftFading:
			line = fmt.Sprintf("  %s %.0f%% (%d → %d)",
				truncateDrift(item.ClusterName, 25),
				item.ChangePercent,
				item.PriorCount,
				item.CurrentCount)
		case attention.DriftEmerged:
			line = fmt.Sprintf("  %s (new: %d beats)",
				truncateDrift(item.ClusterName, 25),
				item.CurrentCount)
		case attention.DriftVanished:
			line = fmt.Sprintf("  %s (was: %d beats)",
				truncateDrift(item.ClusterName, 25),
				item.PriorCount)
		default:
			line = fmt.Sprintf("  %s (%d beats)",
				truncateDrift(item.ClusterName, 25),
				item.CurrentCount)
		}

		sb.WriteString(style.Render(line))
		sb.WriteString("\n")
	}

	return sb.String()
}

// SetSize updates dimensions
func (d *DriftView) SetSize(w, h int) {
	d.width = w
	d.height = h
}

// SelectedSection returns the currently focused section
func (d *DriftView) SelectedSection() DriftSection {
	return d.focused
}

func truncateDrift(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
