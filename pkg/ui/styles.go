package ui

import "github.com/charmbracelet/lipgloss"

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	muted     = lipgloss.AdaptiveColor{Light: "#9B9B9B", Dark: "#626262"}

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlight)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	SelectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(highlight)

	NormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	ImpetusStyle = lipgloss.NewStyle().
			Foreground(special).
			Italic(true)

	IDStyle = lipgloss.NewStyle().
		Foreground(muted)

	DateStyle = lipgloss.NewStyle().
			Foreground(muted).
			Width(10)

	ContentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(subtle)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Background(subtle)

	HelpStyle = lipgloss.NewStyle().
			Foreground(muted).
			Padding(0, 1)

	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#FFFDF5", Dark: "#FFFDF5"}).
			Background(lipgloss.AdaptiveColor{Light: "#6124DF", Dark: "#6124DF"}).
			Padding(0, 1)

	DetailLabelStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(highlight)

	DetailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	SearchPromptStyle = lipgloss.NewStyle().
				Foreground(special)

	ProjectBadgeStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FAFAFA")).
				Background(highlight).
				Padding(0, 1)
)

func Truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
