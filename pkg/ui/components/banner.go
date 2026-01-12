package components

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	bannerBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#F39C12")).
			Padding(0, 1)

	bannerTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F39C12"))

	bannerContent = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	bannerMuted = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	alertIconUrgent = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#E74C3C")).
			Render("!")

	alertIconInfo = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3498DB")).
			Render("?")
)

// DismissAlertMsg is sent when user dismisses an alert
type DismissAlertMsg struct {
	AlertID string
}

// ViewAllAlertsMsg is sent when user wants to see all alerts
type ViewAllAlertsMsg struct{}

// AlertBanner shows unseen alerts
type AlertBanner struct {
	alerts   []model.Alert
	width    int
	expanded bool
}

// NewAlertBanner creates a banner
func NewAlertBanner(width int) *AlertBanner {
	return &AlertBanner{
		width: width,
	}
}

// SetAlerts updates the alerts to display (unseen only)
func (b *AlertBanner) SetAlerts(alerts []model.Alert) {
	// Filter to unseen only
	var unseen []model.Alert
	for _, a := range alerts {
		if a.SeenAt == nil {
			unseen = append(unseen, a)
		}
	}
	b.alerts = unseen
}

// Update handles input (x=dismiss, a=view all)
func (b *AlertBanner) Update(msg tea.Msg) (*AlertBanner, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "x":
			if len(b.alerts) > 0 {
				return b, func() tea.Msg {
					return DismissAlertMsg{AlertID: b.alerts[0].ID}
				}
			}
		case "a":
			return b, func() tea.Msg {
				return ViewAllAlertsMsg{}
			}
		}
	}
	return b, nil
}

// View renders the banner (empty string if no alerts)
func (b *AlertBanner) View() string {
	if len(b.alerts) == 0 {
		return ""
	}

	title := bannerTitle.Render(fmt.Sprintf("─ %d alerts ─", len(b.alerts)))

	var lines []string
	limit := 3
	if len(b.alerts) < limit {
		limit = len(b.alerts)
	}

	for i := 0; i < limit; i++ {
		alert := b.alerts[i]
		icon := alertIconInfo
		if alert.Severity == model.AlertUrgent {
			icon = alertIconUrgent
		}

		line := fmt.Sprintf("%s %s", icon, truncateStr(alert.Title, b.width-10))
		lines = append(lines, bannerContent.Render(line))
	}

	help := bannerMuted.Render("[x]dismiss [a]all")

	content := strings.Join(lines, "\n")
	inner := fmt.Sprintf("%s\n%s\n%s", title, content, help)

	return bannerBorder.Width(b.width - 4).Render(inner)
}

// SetWidth updates width
func (b *AlertBanner) SetWidth(w int) {
	b.width = w
}

// HasAlerts returns true if there are unseen alerts
func (b *AlertBanner) HasAlerts() bool {
	return len(b.alerts) > 0
}

// Count returns number of unseen alerts
func (b *AlertBanner) Count() int {
	return len(b.alerts)
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
