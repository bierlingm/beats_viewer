package ui

import (
	"fmt"

	"beats_viewer/pkg/loader"
	"beats_viewer/pkg/model"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	SplitViewThreshold = 80
)

type focus int

const (
	focusList focus = iota
	focusDetail
	focusSearch
	focusHelp
	focusProjectPicker
)

type Model struct {
	beats         []model.Beat
	beatToProject map[string]string
	projects      []model.Project
	currentProj   int
	allProjects   bool

	list       list.Model
	detail     DetailView
	search     SearchInput
	focus      focus
	showHelp   bool

	width  int
	height int

	statusMsg  string
	rootPath   string
}

func NewModel(rootPath string) Model {
	delegate := NewBeatDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return Model{
		list:        l,
		detail:      NewDetailView(40, 20),
		search:      NewSearchInput(),
		focus:       focusList,
		rootPath:    rootPath,
		allProjects: false,
		currentProj: -1,
	}
}

type beatsLoadedMsg struct {
	beats         []model.Beat
	beatToProject map[string]string
	projects      []model.Project
	err           error
}

func (m Model) loadBeatsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := loader.DiscoverProjects(m.rootPath)
		if err != nil {
			return beatsLoadedMsg{err: err}
		}

		if m.allProjects || m.currentProj < 0 {
			beats, beatToProject, err := loader.LoadAllBeats(m.rootPath)
			return beatsLoadedMsg{
				beats:         beats,
				beatToProject: beatToProject,
				projects:      projects,
				err:           err,
			}
		}

		if m.currentProj >= 0 && m.currentProj < len(projects) {
			beats, err := loader.LoadBeats(projects[m.currentProj].Path)
			beatToProject := make(map[string]string)
			for _, b := range beats {
				beatToProject[b.ID] = projects[m.currentProj].Name
			}
			return beatsLoadedMsg{
				beats:         beats,
				beatToProject: beatToProject,
				projects:      projects,
				err:           err,
			}
		}

		return beatsLoadedMsg{projects: projects}
	}
}

func (m Model) Init() tea.Cmd {
	return m.loadBeatsCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateLayout()
		return m, nil

	case beatsLoadedMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", msg.err)
			return m, nil
		}
		m.beats = msg.beats
		m.beatToProject = msg.beatToProject
		m.projects = msg.projects
		m.updateList()
		if len(m.beats) > 0 {
			m.updateSelectedBeat()
		}
		m.statusMsg = fmt.Sprintf("Loaded %d beats from %d projects", len(m.beats), len(m.projects))
		return m, nil

	case tea.KeyMsg:
		if m.focus == focusSearch && m.search.IsActive() {
			switch msg.String() {
			case "enter", "esc":
				m.search.Blur()
				m.focus = focusList
				m.filterBeats()
				return m, nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.filterBeats()
				return m, cmd
			}
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit

		case "/":
			m.focus = focusSearch
			return m, m.search.Focus()

		case "esc":
			if m.search.Query() != "" {
				m.search.Clear()
				m.filterBeats()
			}
			m.focus = focusList
			return m, nil

		case "?":
			m.showHelp = true
			return m, nil

		case "p":
			m.cycleProject()
			return m, m.loadBeatsCmd()

		case "a":
			m.allProjects = !m.allProjects
			if m.allProjects {
				m.currentProj = -1
			}
			return m, m.loadBeatsCmd()

		case "r":
			return m, m.loadBeatsCmd()

		case "y":
			if item, ok := m.list.SelectedItem().(BeatItem); ok {
				clipboard.WriteAll(item.Beat().ID)
				m.statusMsg = "Copied beat ID to clipboard"
			}
			return m, nil

		case "Y":
			if item, ok := m.list.SelectedItem().(BeatItem); ok {
				clipboard.WriteAll(item.Beat().Content)
				m.statusMsg = "Copied beat content to clipboard"
			}
			return m, nil

		case "enter", "tab":
			if m.focus == focusList {
				m.focus = focusDetail
			} else {
				m.focus = focusList
			}
			return m, nil

		case "j", "down":
			if m.focus == focusDetail {
				m.detail.ScrollDown()
			} else {
				m.list.CursorDown()
				m.updateSelectedBeat()
			}
			return m, nil

		case "k", "up":
			if m.focus == focusDetail {
				m.detail.ScrollUp()
			} else {
				m.list.CursorUp()
				m.updateSelectedBeat()
			}
			return m, nil

		case "g":
			if m.focus == focusDetail {
				m.detail.GotoTop()
			} else {
				m.list.Select(0)
				m.updateSelectedBeat()
			}
			return m, nil

		case "G":
			if m.focus == focusDetail {
				m.detail.GotoBottom()
			} else {
				m.list.Select(len(m.list.Items()) - 1)
				m.updateSelectedBeat()
			}
			return m, nil

		case "ctrl+d":
			if m.focus == focusDetail {
				m.detail.PageDown()
			}
			return m, nil

		case "ctrl+u":
			if m.focus == focusDetail {
				m.detail.PageUp()
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	prevIdx := m.list.Index()
	m.list, cmd = m.list.Update(msg)
	if m.list.Index() != prevIdx {
		m.updateSelectedBeat()
	}
	return m, cmd
}

func (m *Model) updateLayout() {
	headerHeight := 2
	footerHeight := 1
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 5 {
		contentHeight = 5
	}

	if m.width >= SplitViewThreshold {
		listWidth := m.width / 2
		detailWidth := m.width - listWidth - 2
		m.list.SetSize(listWidth, contentHeight)
		m.detail.SetSize(detailWidth, contentHeight-2)
		delegate := NewBeatDelegate().SetWidth(listWidth - 2)
		m.list.SetDelegate(delegate)
	} else {
		m.list.SetSize(m.width, contentHeight)
		m.detail.SetSize(m.width, contentHeight)
		delegate := NewBeatDelegate().SetWidth(m.width - 2)
		m.list.SetDelegate(delegate)
	}

	m.search.SetWidth(m.width / 3)
}

func (m *Model) updateList() {
	items := BeatsToItems(m.beats, m.beatToProject)
	m.list.SetItems(items)
}

func (m *Model) filterBeats() {
	query := m.search.Query()
	filtered := loader.SearchBeats(m.beats, query)
	items := BeatsToItems(filtered, m.beatToProject)
	m.list.SetItems(items)
	if len(items) > 0 {
		m.list.Select(0)
		m.updateSelectedBeat()
	}
}

func (m *Model) updateSelectedBeat() {
	if item, ok := m.list.SelectedItem().(BeatItem); ok {
		beat := item.Beat()
		m.detail.SetBeat(&beat, item.Project())
	}
}

func (m *Model) cycleProject() {
	if len(m.projects) == 0 {
		return
	}
	m.allProjects = false
	m.currentProj = (m.currentProj + 1) % len(m.projects)
}

func (m Model) View() string {
	if m.showHelp {
		return m.renderHelp()
	}

	header := m.renderHeader()
	footer := m.renderFooter()

	contentHeight := m.height - 3
	if contentHeight < 5 {
		contentHeight = 5
	}

	var content string
	if m.width >= SplitViewThreshold {
		listView := m.list.View()
		detailView := m.detail.View()

		listStyle := lipgloss.NewStyle().Width(m.width / 2).MaxHeight(contentHeight)
		detailStyle := lipgloss.NewStyle().Width(m.width - m.width/2 - 2).MaxHeight(contentHeight).BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#383838"))

		if m.focus == focusList {
			listStyle = listStyle.BorderForeground(lipgloss.Color("#7D56F4"))
		}
		if m.focus == focusDetail {
			detailStyle = detailStyle.BorderForeground(lipgloss.Color("#7D56F4"))
		}

		content = lipgloss.JoinHorizontal(lipgloss.Top,
			listStyle.Render(listView),
			detailStyle.Render(detailView),
		)
	} else {
		if m.focus == focusDetail {
			content = m.detail.View()
		} else {
			content = m.list.View()
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m Model) renderHeader() string {
	title := TitleStyle.Render("btv")

	var projectInfo string
	if m.allProjects || m.currentProj < 0 {
		projectInfo = ProjectBadgeStyle.Render("all projects")
	} else if m.currentProj >= 0 && m.currentProj < len(m.projects) {
		projectInfo = ProjectBadgeStyle.Render(m.projects[m.currentProj].Name)
	}

	beatCount := SubtitleStyle.Render(fmt.Sprintf("%d beats", len(m.list.Items())))

	searchView := m.search.View()

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", projectInfo, " ", beatCount)
	right := searchView

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, left, lipgloss.NewStyle().Width(gap).Render(""), right) + "\n"
}

func (m Model) renderFooter() string {
	focusIndicator := ""
	if m.focus == focusDetail {
		focusIndicator = StatusBarStyle.Render(" DETAIL ") + " "
	}

	help := "j/k:nav /:search p:proj a:all tab:focus y:copy ?:help q:quit"
	if m.width < 80 {
		help = "j/k /search p a tab y ? q"
	}

	status := ""
	if m.statusMsg != "" {
		maxStatusLen := m.width - lipgloss.Width(help) - lipgloss.Width(focusIndicator) - 4
		if maxStatusLen > 10 {
			status = Truncate(m.statusMsg, maxStatusLen)
		}
	}

	left := focusIndicator + HelpStyle.Render(help)
	if status != "" {
		return lipgloss.JoinHorizontal(lipgloss.Center, left, "  ", SubtitleStyle.Render(status))
	}
	return left
}

func (m Model) renderHelp() string {
	helpText := `beats_viewer (btv) - TUI for browsing beats

NAVIGATION
  j/↓         Next beat
  k/↑         Previous beat
  ←/→         Page through list
  g           Go to first beat
  G           Go to last beat
  Tab/Enter   Switch focus between list and detail
  Ctrl+d/u    Page down/up in detail view

SEARCH & FILTER
  /           Start search (filters beats in real-time)
  Esc         Clear search / cancel
  p           Cycle through projects
  a           Toggle all-projects mode
  r           Refresh beats from disk

COPY TO CLIPBOARD
  y           Copy beat ID (e.g. "beat-20251204-001")
  Y           Copy full beat content

OTHER
  ?           Toggle this help
  q/Ctrl+c    Quit

Press any key to close...`
	return lipgloss.NewStyle().Padding(1, 2).Render(helpText)
}
