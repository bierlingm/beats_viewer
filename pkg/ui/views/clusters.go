package views

import (
	"fmt"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	clusterTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	clusterExpandedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	clusterCollapsedStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	clusterSelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Background(lipgloss.Color("#383838"))

	clusterBeatStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262")).
				PaddingLeft(4)

	ripenessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F39C12"))
)

type ClusterView struct {
	clusters     []model.Cluster
	beatContents map[string]string
	width        int
	height       int

	cursorPos    int
	expanded     map[int]bool
	scrollOffset int
}

func NewClusterView(width, height int) *ClusterView {
	return &ClusterView{
		width:        width,
		height:       height,
		expanded:     make(map[int]bool),
		beatContents: make(map[string]string),
	}
}

func (cv *ClusterView) SetSize(width, height int) {
	cv.width = width
	cv.height = height
}

func (cv *ClusterView) SetClusters(clusters []model.Cluster) {
	cv.clusters = clusters
	cv.cursorPos = 0
	cv.scrollOffset = 0
}

func (cv *ClusterView) SetBeatContents(beats []model.EnrichedBeat) {
	cv.beatContents = make(map[string]string)
	for _, b := range beats {
		preview := b.Content
		if len(preview) > 60 {
			preview = preview[:57] + "..."
		}
		cv.beatContents[b.ID] = preview
	}
}

func (cv *ClusterView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			cv.cursorDown()
		case "k", "up":
			cv.cursorUp()
		case "enter":
			cv.toggleExpand()
		}
	}
	return nil
}

func (cv *ClusterView) cursorDown() {
	maxPos := len(cv.clusters) - 1
	if cv.cursorPos < maxPos {
		cv.cursorPos++
		cv.ensureVisible()
	}
}

func (cv *ClusterView) cursorUp() {
	if cv.cursorPos > 0 {
		cv.cursorPos--
		cv.ensureVisible()
	}
}

func (cv *ClusterView) toggleExpand() {
	cv.expanded[cv.cursorPos] = !cv.expanded[cv.cursorPos]
}

func (cv *ClusterView) ensureVisible() {
	visibleHeight := cv.height - 4
	if cv.cursorPos < cv.scrollOffset {
		cv.scrollOffset = cv.cursorPos
	} else if cv.cursorPos >= cv.scrollOffset+visibleHeight {
		cv.scrollOffset = cv.cursorPos - visibleHeight + 1
	}
}

func (cv *ClusterView) SelectedCluster() *model.Cluster {
	if cv.cursorPos >= 0 && cv.cursorPos < len(cv.clusters) {
		return &cv.clusters[cv.cursorPos]
	}
	return nil
}

func (cv *ClusterView) View() string {
	if len(cv.clusters) == 0 {
		msg := "No clusters available.\n\n"
		msg += "Clustering requires Ollama with nomic-embed-text model.\n"
		msg += "Install Ollama and run: ollama pull nomic-embed-text"
		return lipgloss.NewStyle().
			Width(cv.width).
			Height(cv.height).
			Padding(1).
			Foreground(lipgloss.Color("#626262")).
			Render(msg)
	}

	var lines []string

	lines = append(lines, clusterTitleStyle.Render(fmt.Sprintf("Theme Clusters (%d)", len(cv.clusters))))
	lines = append(lines, "")

	for i, cluster := range cv.clusters {
		isSelected := i == cv.cursorPos
		isExpanded := cv.expanded[i]

		arrow := "▶"
		if isExpanded {
			arrow = "▼"
		}

		ripenessEmoji := model.RipenessEmoji(cluster.RipenessScore)
		ripenessStr := ripenessStyle.Render(fmt.Sprintf("%.2f", cluster.RipenessScore))

		header := fmt.Sprintf("%s %s (%d beats) %s %s",
			arrow, cluster.Name, len(cluster.BeatIDs), ripenessEmoji, ripenessStr)

		if isExpanded {
			header = clusterExpandedStyle.Render(header)
		} else {
			header = clusterCollapsedStyle.Render(header)
		}

		if isSelected {
			header = clusterSelectedStyle.Render(header)
		}

		lines = append(lines, header)

		if isExpanded {
			maxBeats := 5
			for j, beatID := range cluster.BeatIDs {
				if j >= maxBeats {
					lines = append(lines, clusterBeatStyle.Render(
						fmt.Sprintf("  +%d more...", len(cluster.BeatIDs)-maxBeats)))
					break
				}
				preview := cv.beatContents[beatID]
				if preview == "" {
					preview = beatID
				}
				lines = append(lines, clusterBeatStyle.Render("│ "+preview))
			}
			if len(cluster.Keywords) > 0 {
				kw := "Keywords: " + strings.Join(cluster.Keywords, ", ")
				lines = append(lines, clusterBeatStyle.Render(kw))
			}
			lines = append(lines, "")
		}
	}

	lines = append(lines, "")
	lines = append(lines, lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		"j/k Navigate  Enter Expand  n Rename  m Merge  b Create bead"))

	visibleHeight := cv.height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	start := cv.scrollOffset
	end := start + visibleHeight
	if start > len(lines) {
		start = len(lines)
	}
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]

	return lipgloss.NewStyle().
		Width(cv.width).
		Height(cv.height).
		Render(strings.Join(visibleLines, "\n"))
}
