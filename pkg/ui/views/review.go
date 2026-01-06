package views

import (
	"fmt"
	"strings"
	"time"

	"beats_viewer/pkg/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	reviewTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF6B6B"))

	reviewBeatStyle = lipgloss.NewStyle().
			Padding(1).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#383838"))

	reviewActionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#73F59F"))

	reviewProgressStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#626262"))
)

type ReviewAction int

const (
	ReviewKeep ReviewAction = iota
	ReviewArchive
	ReviewConvert
	ReviewChain
	ReviewDelete
	ReviewSkip
)

type StaleReviewView struct {
	staleBeats   []model.EnrichedBeat
	currentIndex int
	width        int
	height       int
	completed    int
	actions      map[string]ReviewAction
}

func NewStaleReviewView(width, height int) *StaleReviewView {
	return &StaleReviewView{
		width:   width,
		height:  height,
		actions: make(map[string]ReviewAction),
	}
}

func (rv *StaleReviewView) SetSize(width, height int) {
	rv.width = width
	rv.height = height
}

func (rv *StaleReviewView) SetStaleBeats(beats []model.EnrichedBeat) {
	rv.staleBeats = beats
	rv.currentIndex = 0
	rv.completed = 0
	rv.actions = make(map[string]ReviewAction)
}

func (rv *StaleReviewView) CurrentBeat() *model.EnrichedBeat {
	if rv.currentIndex >= 0 && rv.currentIndex < len(rv.staleBeats) {
		return &rv.staleBeats[rv.currentIndex]
	}
	return nil
}

func (rv *StaleReviewView) IsComplete() bool {
	return rv.currentIndex >= len(rv.staleBeats)
}

func (rv *StaleReviewView) Progress() (current, total int) {
	return rv.completed, len(rv.staleBeats)
}

func (rv *StaleReviewView) Update(msg tea.Msg) (ReviewAction, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "k":
			return rv.recordAction(ReviewKeep), nil
		case "a":
			return rv.recordAction(ReviewArchive), nil
		case "b":
			return rv.recordAction(ReviewConvert), nil
		case "c":
			return rv.recordAction(ReviewChain), nil
		case "d":
			return rv.recordAction(ReviewDelete), nil
		case "right", "n":
			return rv.skip(), nil
		}
	}
	return ReviewSkip, nil
}

func (rv *StaleReviewView) recordAction(action ReviewAction) ReviewAction {
	if beat := rv.CurrentBeat(); beat != nil {
		rv.actions[beat.ID] = action
		rv.completed++
		rv.currentIndex++
	}
	return action
}

func (rv *StaleReviewView) skip() ReviewAction {
	if rv.currentIndex < len(rv.staleBeats) {
		rv.currentIndex++
	}
	return ReviewSkip
}

func (rv *StaleReviewView) GetActions() map[string]ReviewAction {
	return rv.actions
}

func (rv *StaleReviewView) View() string {
	if len(rv.staleBeats) == 0 {
		return lipgloss.NewStyle().
			Width(rv.width).
			Height(rv.height).
			Padding(2).
			Render("No stale beats to review! All caught up.")
	}

	if rv.IsComplete() {
		return lipgloss.NewStyle().
			Width(rv.width).
			Height(rv.height).
			Padding(2).
			Render(fmt.Sprintf("Review complete! Processed %d beats.", rv.completed))
	}

	var sb strings.Builder

	sb.WriteString(reviewTitleStyle.Render(
		fmt.Sprintf("Stale Beat Review (%d beats need attention)", len(rv.staleBeats))))
	sb.WriteString("\n\n")

	beat := rv.CurrentBeat()
	if beat == nil {
		return sb.String()
	}

	age := time.Since(beat.CreatedAt)
	ageDays := int(age.Hours() / 24)

	viewInfo := "never viewed"
	if beat.ViewCount > 0 {
		viewInfo = fmt.Sprintf("viewed %d times", beat.ViewCount)
	}

	header := fmt.Sprintf("%s (%d days old, %s)", beat.ID, ageDays, viewInfo)
	sb.WriteString(lipgloss.NewStyle().Bold(true).Render(header))
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("Channel: %s    Source: %s\n",
		beat.Taxonomy.Channel.String(), beat.Taxonomy.Source.String()))
	sb.WriteString("\n")

	contentPreview := beat.Content
	maxContent := rv.height - 15
	if maxContent < 50 {
		maxContent = 50
	}
	if len(contentPreview) > maxContent {
		contentPreview = contentPreview[:maxContent] + "..."
	}

	sb.WriteString(reviewBeatStyle.Width(rv.width - 4).Render(contentPreview))
	sb.WriteString("\n\n")

	sb.WriteString("Actions:\n")
	sb.WriteString(reviewActionStyle.Render("  [k] Keep") + " - Mark as reviewed\n")
	sb.WriteString(reviewActionStyle.Render("  [a] Archive") + " - Move to archive\n")
	sb.WriteString(reviewActionStyle.Render("  [b] Convert") + " - Create bead from this\n")
	sb.WriteString(reviewActionStyle.Render("  [c] Chain") + " - Add to existing chain\n")
	sb.WriteString(reviewActionStyle.Render("  [d] Delete") + " - Remove permanently\n")
	sb.WriteString("\n")

	progress := reviewProgressStyle.Render(
		fmt.Sprintf("Progress: %d/%d    Skip: →/n    Quit: q", rv.completed, len(rv.staleBeats)))
	sb.WriteString(progress)

	return lipgloss.NewStyle().
		Width(rv.width).
		Height(rv.height).
		Render(sb.String())
}

func IsStale(beat model.EnrichedBeat) bool {
	age := time.Since(beat.CreatedAt)
	if age < 30*24*time.Hour {
		return false
	}

	if beat.ViewCount > 0 && beat.LastViewedAt != nil {
		if time.Since(*beat.LastViewedAt) < 14*24*time.Hour {
			return false
		}
	}

	if len(beat.LinkedBeads) > 0 {
		return false
	}

	if len(beat.ChainIDs) > 0 {
		return false
	}

	return true
}

func FindStaleBeats(beats []model.EnrichedBeat) []model.EnrichedBeat {
	var stale []model.EnrichedBeat
	for _, b := range beats {
		if IsStale(b) {
			stale = append(stale, b)
		}
	}
	return stale
}
