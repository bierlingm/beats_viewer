package ui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type SearchInput struct {
	input   textinput.Model
	active  bool
	query   string
	width   int
}

func NewSearchInput() SearchInput {
	ti := textinput.New()
	ti.Placeholder = "Search beats..."
	ti.Prompt = SearchPromptStyle.Render("/") + " "
	ti.CharLimit = 100
	ti.Width = 40

	return SearchInput{
		input:  ti,
		active: false,
	}
}

func (s *SearchInput) SetWidth(w int) {
	s.width = w
	s.input.Width = w - 4
}

func (s *SearchInput) Focus() tea.Cmd {
	s.active = true
	return s.input.Focus()
}

func (s *SearchInput) Blur() {
	s.active = false
	s.input.Blur()
}

func (s *SearchInput) IsActive() bool {
	return s.active
}

func (s *SearchInput) Query() string {
	return s.query
}

func (s *SearchInput) SetQuery(q string) {
	s.query = q
	s.input.SetValue(q)
}

func (s *SearchInput) Clear() {
	s.query = ""
	s.input.SetValue("")
}

func (s *SearchInput) Update(msg tea.Msg) (SearchInput, tea.Cmd) {
	if !s.active {
		return *s, nil
	}

	var cmd tea.Cmd
	s.input, cmd = s.input.Update(msg)
	s.query = s.input.Value()

	return *s, cmd
}

func (s *SearchInput) View() string {
	if s.active {
		return s.input.View()
	}
	if s.query != "" {
		return SearchPromptStyle.Render("/") + " " + s.query
	}
	return SearchPromptStyle.Render("/") + " " + SubtitleStyle.Render("search")
}
