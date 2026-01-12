package views

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/attention"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	hbTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4"))

	hbBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4"))

	hbMutedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	hbHighlightStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	hbWarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12"))
)

// HeartbeatView displays capture rhythm visualization
type HeartbeatView struct {
	width     int
	height    int
	heartbeat *attention.Heartbeat
}

// NewHeartbeatView creates a new heartbeat view
func NewHeartbeatView(width, height int) *HeartbeatView {
	return &HeartbeatView{
		width:  width,
		height: height,
	}
}

// SetHeartbeat updates the heartbeat data
func (h *HeartbeatView) SetHeartbeat(hb *attention.Heartbeat) {
	h.heartbeat = hb
}

// Update handles input
func (h *HeartbeatView) Update(msg tea.Msg) (*HeartbeatView, tea.Cmd) {
	return h, nil
}

// View renders the heartbeat visualization
func (h *HeartbeatView) View() string {
	if h.heartbeat == nil {
		return hbMutedStyle.Render("No heartbeat data available")
	}

	hb := h.heartbeat
	var sb strings.Builder

	// Title
	title := hbTitleStyle.Render(fmt.Sprintf("─ Capture Rhythm (%d days) ─", int(hb.Window.Hours()/24)))
	sb.WriteString(title)
	sb.WriteString("\n\n")

	// ASCII bar chart
	barWidth := h.width - 10
	if barWidth > 80 {
		barWidth = 80
	}

	// Find max for scaling
	maxCount := 0
	for _, d := range hb.DailyBuckets {
		if d.Count > maxCount {
			maxCount = d.Count
		}
	}

	// Render bars
	if len(hb.DailyBuckets) > 0 && maxCount > 0 {
		// Compress to fit width
		step := len(hb.DailyBuckets) / barWidth
		if step < 1 {
			step = 1
		}

		var bars strings.Builder
		for i := 0; i < len(hb.DailyBuckets); i += step {
			count := hb.DailyBuckets[i].Count
			height := (count * 8) / maxCount
			if count > 0 && height == 0 {
				height = 1
			}

			char := "░"
			if height > 0 {
				chars := []string{"░", "▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
				if height >= len(chars) {
					height = len(chars) - 1
				}
				char = chars[height]
			}
			bars.WriteString(char)
		}
		sb.WriteString(hbBarStyle.Render(bars.String()))
		sb.WriteString("\n")

		// Month labels
		if len(hb.DailyBuckets) > 0 {
			first := hb.DailyBuckets[0].Date
			last := hb.DailyBuckets[len(hb.DailyBuckets)-1].Date
			sb.WriteString(hbMutedStyle.Render(fmt.Sprintf("%s          %s",
				first.Format("Jan"), last.Format("Jan"))))
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")

	// Statistics
	avgRate := hb.DailyAverage

	currentRate := hb.CurrentRate
	changePercent := 0.0
	if avgRate > 0 {
		changePercent = ((currentRate - avgRate) / avgRate) * 100
	}

	changeIndicator := "→"
	changeStyle := hbMutedStyle
	if changePercent > 20 {
		changeIndicator = "↑"
		changeStyle = hbHighlightStyle
	} else if changePercent < -20 {
		changeIndicator = "↓"
		changeStyle = hbWarningStyle
	}

	stats := fmt.Sprintf("Average: %.1f beats/day    Current: %.1f beats/day (%s%.0f%%)",
		avgRate, currentRate, changeIndicator, changePercent)
	sb.WriteString(changeStyle.Render(stats))
	sb.WriteString("\n")

	// Longest gap
	if hb.LongestGap.Duration > 0 {
		gapDays := int(hb.LongestGap.Duration.Hours() / 24)
		gapInfo := fmt.Sprintf("Longest gap: %d days", gapDays)
		sb.WriteString(hbMutedStyle.Render(gapInfo))
		sb.WriteString("\n")
	}

	// Bursts
	if len(hb.Bursts) > 0 {
		sb.WriteString("\n")
		sb.WriteString(hbHighlightStyle.Render("Bursts:"))
		sb.WriteString("\n")
		for i, burst := range hb.Bursts {
			if i >= 3 {
				break
			}
			burstInfo := fmt.Sprintf("  • %s: %d beats", burst.Start.Format("Jan 2"), burst.Total)
			sb.WriteString(hbMutedStyle.Render(burstInfo))
			sb.WriteString("\n")
		}
	}

	// Gaps
	if len(hb.Gaps) > 0 {
		sb.WriteString("\n")
		sb.WriteString(hbWarningStyle.Render("Gaps:"))
		sb.WriteString("\n")
		for i, gap := range hb.Gaps {
			if i >= 3 {
				break
			}
			gapDays := int(gap.Duration.Hours() / 24)
			gapInfo := fmt.Sprintf("  • %s - %s (%d days)",
				gap.Start.Format("Jan 2"), gap.End.Format("Jan 2"), gapDays)
			sb.WriteString(hbMutedStyle.Render(gapInfo))
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// SetSize updates dimensions
func (h *HeartbeatView) SetSize(w, hgt int) {
	h.width = w
	h.height = hgt
}
