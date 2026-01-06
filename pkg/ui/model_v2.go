package ui

import (
	"fmt"
	"io"
	"os/exec"
	"sort"
	"time"

	"github.com/bierlingm/beats_viewer/pkg/chain"
	"github.com/bierlingm/beats_viewer/pkg/cluster"
	"github.com/bierlingm/beats_viewer/pkg/loader"
	"github.com/bierlingm/beats_viewer/pkg/model"
	"github.com/bierlingm/beats_viewer/pkg/ui/components"
	"github.com/bierlingm/beats_viewer/pkg/ui/views"

	"github.com/atotto/clipboard"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ViewMode int

const (
	ViewList ViewMode = iota
	ViewTimeline
	ViewClusters
	ViewReview
	ViewCapture
)

type ModelV2 struct {
	beats         []model.Beat
	enrichedBeats []model.EnrichedBeat
	filteredBeats []model.EnrichedBeat
	cache         *model.Cache
	beatToProject map[string]string
	projects      []model.Project

	currentProj int
	allProjects bool

	list         list.Model
	detail       DetailView
	search       SearchInput
	facets       *components.FacetSidebar
	entities     *components.EntitySidebar
	timelineView *views.TimelineView
	clusterView  *views.ClusterView
	reviewView   *views.StaleReviewView
	captureView  *views.CaptureView

	chainStore    *chain.Store
	clusterEngine *cluster.Engine

	focus    focus
	viewMode ViewMode
	showHelp bool

	showFacets   bool
	showEntities bool

	sortByRipeness bool

	width  int
	height int

	statusMsg string
	rootPath  string
}

func NewModelV2(rootPath string) ModelV2 {
	delegate := NewBeatDelegate()
	l := list.New([]list.Item{}, delegate, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()

	return ModelV2{
		list:          l,
		detail:        NewDetailView(40, 20),
		search:        NewSearchInput(),
		facets:        components.NewFacetSidebar(25, 20),
		entities:      components.NewEntitySidebar(25, 20),
		timelineView:  views.NewTimelineView(80, 20),
		clusterView:   views.NewClusterView(80, 20),
		reviewView:    views.NewStaleReviewView(80, 20),
		captureView:   views.NewCaptureView(60, 15),
		chainStore:    chain.NewStore(),
		clusterEngine: cluster.NewEngine(),
		focus:         focusList,
		viewMode:      ViewList,
		rootPath:      rootPath,
		allProjects:   false,
		currentProj:   -1,
	}
}

func (m ModelV2) Init() tea.Cmd {
	return m.loadBeatsCmd()
}

func (m ModelV2) loadBeatsCmd() tea.Cmd {
	return func() tea.Msg {
		projects, err := loader.DiscoverProjects(m.rootPath)
		if err != nil {
			return beatsLoadedMsg{err: err}
		}

		if m.allProjects || m.currentProj < 0 {
			beats, beatToProject, err := loader.LoadAllBeats(m.rootPath)
			if err != nil {
				return beatsLoadedMsg{err: err, projects: projects}
			}

			var enrichedBeats []model.EnrichedBeat
			var cache *model.Cache
			if len(projects) > 0 {
				enrichedBeats, cache, _ = loader.LoadEnrichedBeats(projects[0].Path, nil)
			}

			return beatsLoadedMsg{
				beats:         beats,
				enrichedBeats: enrichedBeats,
				cache:         cache,
				beatToProject: beatToProject,
				projects:      projects,
				err:           err,
			}
		}

		if m.currentProj >= 0 && m.currentProj < len(projects) {
			enrichedBeats, cache, err := loader.LoadEnrichedBeats(projects[m.currentProj].Path, nil)
			beatToProject := make(map[string]string)
			var beats []model.Beat
			for _, eb := range enrichedBeats {
				beats = append(beats, eb.Beat)
				beatToProject[eb.ID] = projects[m.currentProj].Name
			}
			return beatsLoadedMsg{
				beats:         beats,
				enrichedBeats: enrichedBeats,
				cache:         cache,
				beatToProject: beatToProject,
				projects:      projects,
				err:           err,
			}
		}

		return beatsLoadedMsg{projects: projects}
	}
}

func (m ModelV2) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		m.enrichedBeats = msg.enrichedBeats
		m.filteredBeats = msg.enrichedBeats
		m.cache = msg.cache
		m.beatToProject = msg.beatToProject
		m.projects = msg.projects

		if m.cache != nil {
			m.chainStore.LoadFromCache(m.cache.Chains)
			m.facets.UpdateCounts(m.enrichedBeats)
			m.entities.UpdateEntities(m.cache.Entities)
			m.timelineView.SetBeats(m.enrichedBeats)
			m.clusterView.SetClusters(m.cache.Clusters)
			m.clusterView.SetBeatContents(m.enrichedBeats)
		}

		m.updateList()
		if len(m.beats) > 0 {
			m.updateSelectedBeat()
		}

		cacheStatus := ""
		if m.cache != nil {
			cacheStatus = " (cache loaded)"
		}
		m.statusMsg = fmt.Sprintf("Loaded %d beats from %d projects%s", len(m.beats), len(m.projects), cacheStatus)
		return m, nil

	case tea.KeyMsg:
		if m.viewMode == ViewCapture {
			cmd := m.captureView.Update(msg)
			if m.captureView.IsSubmitted() || m.captureView.IsCancelled() {
				m.viewMode = ViewList
				m.captureView.Reset()
			}
			return m, cmd
		}

		if m.viewMode == ViewReview {
			action, cmd := m.reviewView.Update(msg)
			if msg.String() == "q" || m.reviewView.IsComplete() {
				m.viewMode = ViewList
				_ = action
			}
			return m, cmd
		}

		if m.focus == focusSearch && m.search.IsActive() {
			switch msg.String() {
			case "enter", "esc":
				m.search.Blur()
				m.focus = focusList
				m.applyFilters()
				return m, nil
			default:
				var cmd tea.Cmd
				m.search, cmd = m.search.Update(msg)
				m.applyFilters()
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
				m.applyFilters()
			}
			if m.viewMode != ViewList {
				m.viewMode = ViewList
			}
			m.focus = focusList
			return m, nil

		case "?":
			m.showHelp = true
			return m, nil

		case "f":
			m.showFacets = !m.showFacets
			m.updateLayout()
			return m, nil

		case "e":
			m.showEntities = !m.showEntities
			m.updateLayout()
			return m, nil

		case "t":
			if m.viewMode == ViewTimeline {
				m.viewMode = ViewList
			} else {
				m.viewMode = ViewTimeline
			}
			return m, nil

		case "C":
			if m.viewMode == ViewClusters {
				m.viewMode = ViewList
			} else {
				m.viewMode = ViewClusters
			}
			return m, nil

		case "S":
			staleBeats := views.FindStaleBeats(m.enrichedBeats)
			m.reviewView.SetStaleBeats(staleBeats)
			m.viewMode = ViewReview
			return m, nil

		case "R":
			m.sortByRipeness = !m.sortByRipeness
			m.applyFilters()
			if m.sortByRipeness {
				m.statusMsg = "Sorted by ripeness"
			} else {
				m.statusMsg = "Sorted by date"
			}
			return m, nil

		case "1", "2", "3", "4", "5", "6", "7":
			n := int(msg.String()[0] - '0')
			m.facets.SelectChannelByNumber(n)
			m.applyFilters()
			return m, nil

		case "!":
			m.facets.ClearFilters()
			m.entities.ClearSelection()
			m.applyFilters()
			m.statusMsg = "Filters cleared"
			return m, nil

		case "b":
			if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
				m.convertToBead(item.beat)
			}
			return m, nil

		case "c":
			if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
				m.addToChain(item.beat.ID)
			}
			return m, nil

		case "[":
			m.navigateChainPrev()
			return m, nil

		case "]":
			m.navigateChainNext()
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
			if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
				clipboard.WriteAll(item.beat.ID)
				m.statusMsg = "Copied beat ID"
			}
			return m, nil

		case "Y":
			if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
				clipboard.WriteAll(item.beat.Content)
				m.statusMsg = "Copied content"
			}
			return m, nil

		case "enter", "tab":
			if m.viewMode == ViewTimeline {
				if beatIDs := m.timelineView.SelectedBeatIDs(); len(beatIDs) > 0 {
					m.filterToBeatIDs(beatIDs)
					m.viewMode = ViewList
				}
				return m, nil
			}
			if m.focus == focusList {
				m.focus = focusDetail
			} else {
				m.focus = focusList
			}
			m.recordView()
			return m, nil

		case "j", "down":
			if m.viewMode == ViewTimeline {
				return m, nil
			}
			if m.viewMode == ViewClusters {
				m.clusterView.Update(msg)
				return m, nil
			}
			if m.focus == focusDetail {
				m.detail.ScrollDown()
			} else {
				m.list.CursorDown()
				m.updateSelectedBeat()
			}
			return m, nil

		case "k", "up":
			if m.viewMode == ViewTimeline {
				return m, nil
			}
			if m.viewMode == ViewClusters {
				m.clusterView.Update(msg)
				return m, nil
			}
			if m.focus == focusDetail {
				m.detail.ScrollUp()
			} else {
				m.list.CursorUp()
				m.updateSelectedBeat()
			}
			return m, nil

		case "left", "h":
			if m.viewMode == ViewTimeline {
				m.timelineView.Update(msg)
			}
			return m, nil

		case "right", "l":
			if m.viewMode == ViewTimeline {
				m.timelineView.Update(msg)
			}
			return m, nil

		case "z":
			if m.viewMode == ViewTimeline {
				m.timelineView.Update(msg)
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

func (m *ModelV2) updateLayout() {
	headerHeight := 2
	footerHeight := 1
	contentHeight := m.height - headerHeight - footerHeight
	if contentHeight < 5 {
		contentHeight = 5
	}

	// Responsive sidebar visibility based on breakpoints
	leftSidebarWidth := 0
	rightSidebarWidth := 0

	// Auto-show/hide sidebars based on width breakpoints
	canShowFacets := m.width >= WidthNormal
	canShowEntities := m.width >= WidthWide
	canShowBoth := m.width >= WidthUltraWide

	if m.showFacets && canShowFacets {
		leftSidebarWidth = 26
	}
	if m.showEntities && canShowEntities {
		// Only show if we have room, or if ultra-wide
		if canShowBoth || !m.showFacets {
			rightSidebarWidth = 26
		}
	}

	mainWidth := m.width - leftSidebarWidth - rightSidebarWidth
	if mainWidth < WidthCompact {
		mainWidth = WidthCompact
	}

	// Compact mode: single column only
	if m.width < WidthCompact {
		m.list.SetSize(m.width, contentHeight)
		m.detail.SetSize(m.width, contentHeight)
		delegate := NewEnrichedBeatDelegate().SetWidth(m.width - 4)
		m.list.SetDelegate(delegate)
	} else if m.width >= SplitViewThreshold {
		listWidth := mainWidth / 2
		detailWidth := mainWidth - listWidth - 2
		m.list.SetSize(listWidth, contentHeight)
		m.detail.SetSize(detailWidth, contentHeight-2)
		delegate := NewEnrichedBeatDelegate().SetWidth(listWidth - 2)
		m.list.SetDelegate(delegate)
	} else {
		m.list.SetSize(mainWidth, contentHeight)
		m.detail.SetSize(mainWidth, contentHeight)
		delegate := NewEnrichedBeatDelegate().SetWidth(mainWidth - 2)
		m.list.SetDelegate(delegate)
	}

	m.facets.SetSize(leftSidebarWidth, contentHeight)
	m.entities.SetSize(rightSidebarWidth, contentHeight)
	m.timelineView.SetSize(mainWidth, contentHeight)
	m.clusterView.SetSize(mainWidth, contentHeight)
	m.reviewView.SetSize(mainWidth, contentHeight)
	m.captureView.SetSize(mainWidth-10, contentHeight-5)
	m.search.SetWidth(m.width / 3)
}

func (m *ModelV2) updateList() {
	items := EnrichedBeatsToItems(m.filteredBeats, m.beatToProject)
	m.list.SetItems(items)
}

func (m *ModelV2) applyFilters() {
	filtered := m.enrichedBeats

	if q := m.search.Query(); q != "" {
		var searchFiltered []model.EnrichedBeat
		for _, eb := range filtered {
			if containsIgnoreCase(eb.Content, q) || containsIgnoreCase(eb.ImpetusLabel(), q) || containsIgnoreCase(eb.ID, q) {
				searchFiltered = append(searchFiltered, eb)
			}
		}
		filtered = searchFiltered
	}

	filtered = components.FilterByFacets(filtered, m.facets.SelectedChannel(), m.facets.SelectedSource())

	if m.cache != nil {
		filtered = components.FilterByEntity(filtered, m.entities.SelectedEntity(), m.cache.EntityIndex)
	}

	if m.sortByRipeness {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].RipenessScore > filtered[j].RipenessScore
		})
	} else {
		sort.Slice(filtered, func(i, j int) bool {
			return filtered[i].CreatedAt.After(filtered[j].CreatedAt)
		})
	}

	m.filteredBeats = filtered
	m.updateList()
	if len(filtered) > 0 {
		m.list.Select(0)
		m.updateSelectedBeat()
	}
}

func (m *ModelV2) filterToBeatIDs(beatIDs []string) {
	idSet := make(map[string]bool)
	for _, id := range beatIDs {
		idSet[id] = true
	}

	var filtered []model.EnrichedBeat
	for _, eb := range m.enrichedBeats {
		if idSet[eb.ID] {
			filtered = append(filtered, eb)
		}
	}

	m.filteredBeats = filtered
	m.updateList()
}

func (m *ModelV2) updateSelectedBeat() {
	if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
		beat := item.beat.Beat
		m.detail.SetBeat(&beat, item.project)
	}
}

func (m *ModelV2) recordView() {
	if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
		if m.cache != nil {
			stat := m.cache.ViewStats[item.beat.ID]
			stat.ViewCount++
			now := time.Now()
			stat.LastViewedAt = &now
			m.cache.ViewStats[item.beat.ID] = stat
		}
	}
}

func (m *ModelV2) cycleProject() {
	if len(m.projects) == 0 {
		return
	}
	m.allProjects = false
	m.currentProj = (m.currentProj + 1) % len(m.projects)
}

func (m *ModelV2) convertToBead(beat model.EnrichedBeat) {
	content := beat.Content
	if len(content) > 200 {
		content = content[:200] + "..."
	}
	title := fmt.Sprintf("Beat: %s", beat.ID)

	cmd := exec.Command("bd", "create", title, "-d", content)
	if err := cmd.Start(); err != nil {
		m.statusMsg = fmt.Sprintf("Error: %v", err)
	} else {
		m.statusMsg = "Opening bd create..."
	}
}

func (m *ModelV2) addToChain(beatID string) {
	chains := m.chainStore.List()
	if len(chains) == 0 {
		_, err := m.chainStore.Create("New Chain", []string{beatID})
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = "Created new chain"
		}
	} else {
		err := m.chainStore.AddBeat(chains[0].ID, beatID)
		if err != nil {
			m.statusMsg = fmt.Sprintf("Error: %v", err)
		} else {
			m.statusMsg = fmt.Sprintf("Added to chain: %s", chains[0].Name)
		}
	}
}

func (m *ModelV2) navigateChainPrev() {
	if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
		chains := m.chainStore.GetChainsForBeat(item.beat.ID)
		if len(chains) > 0 {
			prev, _ := m.chainStore.GetAdjacentBeats(chains[0].ID, item.beat.ID)
			if prev != "" {
				m.selectBeatByID(prev)
			}
		}
	}
}

func (m *ModelV2) navigateChainNext() {
	if item, ok := m.list.SelectedItem().(EnrichedBeatItem); ok {
		chains := m.chainStore.GetChainsForBeat(item.beat.ID)
		if len(chains) > 0 {
			_, next := m.chainStore.GetAdjacentBeats(chains[0].ID, item.beat.ID)
			if next != "" {
				m.selectBeatByID(next)
			}
		}
	}
}

func (m *ModelV2) selectBeatByID(beatID string) {
	items := m.list.Items()
	for i, item := range items {
		if bi, ok := item.(EnrichedBeatItem); ok && bi.beat.ID == beatID {
			m.list.Select(i)
			m.updateSelectedBeat()
			return
		}
	}
}

func (m ModelV2) View() string {
	if m.showHelp {
		return m.renderHelpV2()
	}

	header := m.renderHeaderV2()
	footer := m.renderFooterV2()

	contentHeight := m.height - 3
	if contentHeight < 5 {
		contentHeight = 5
	}

	var mainContent string

	switch m.viewMode {
	case ViewTimeline:
		mainContent = m.timelineView.View()
	case ViewClusters:
		mainContent = m.clusterView.View()
	case ViewReview:
		mainContent = m.reviewView.View()
	case ViewCapture:
		mainContent = m.captureView.View()
	default:
		mainContent = m.renderListView(contentHeight)
	}

	var content string
	if m.showFacets && m.width >= WidthNormal {
		content = lipgloss.JoinHorizontal(lipgloss.Top, m.facets.View(), mainContent)
	} else {
		content = mainContent
	}

	canShowEntities := m.width >= WidthWide
	canShowBoth := m.width >= WidthUltraWide
	if m.showEntities && canShowEntities && (canShowBoth || !m.showFacets) {
		content = lipgloss.JoinHorizontal(lipgloss.Top, content, m.entities.View())
	}

	return lipgloss.JoinVertical(lipgloss.Left, header, content, footer)
}

func (m ModelV2) renderListView(contentHeight int) string {
	leftSidebarWidth := 0
	rightSidebarWidth := 0
	if m.showFacets && m.width >= 100 {
		leftSidebarWidth = 26
	}
	if m.showEntities && m.width >= 140 {
		rightSidebarWidth = 26
	}
	mainWidth := m.width - leftSidebarWidth - rightSidebarWidth

	if mainWidth >= SplitViewThreshold {
		listView := m.list.View()
		detailView := m.detail.View()

		listWidth := mainWidth / 2
		detailWidth := mainWidth - listWidth - 2

		listStyle := lipgloss.NewStyle().Width(listWidth).MaxHeight(contentHeight)
		detailStyle := lipgloss.NewStyle().Width(detailWidth).MaxHeight(contentHeight).
			BorderLeft(true).BorderStyle(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color("#383838"))

		if m.focus == focusList {
			listStyle = listStyle.BorderForeground(lipgloss.Color("#7D56F4"))
		}
		if m.focus == focusDetail {
			detailStyle = detailStyle.BorderForeground(lipgloss.Color("#7D56F4"))
		}

		return lipgloss.JoinHorizontal(lipgloss.Top,
			listStyle.Render(listView),
			detailStyle.Render(detailView),
		)
	}

	if m.focus == focusDetail {
		return m.detail.View()
	}
	return m.list.View()
}

func (m ModelV2) renderHeaderV2() string {
	title := TitleStyle.Render("btv")

	var projectInfo string
	if m.allProjects || m.currentProj < 0 {
		projectInfo = ProjectBadgeStyle.Render("all projects")
	} else if m.currentProj >= 0 && m.currentProj < len(m.projects) {
		projectInfo = ProjectBadgeStyle.Render(m.projects[m.currentProj].Name)
	}

	beatCount := SubtitleStyle.Render(fmt.Sprintf("%d beats", len(m.list.Items())))

	viewIndicator := ""
	switch m.viewMode {
	case ViewTimeline:
		viewIndicator = StatusBarStyle.Render(" TIMELINE ")
	case ViewClusters:
		viewIndicator = StatusBarStyle.Render(" CLUSTERS ")
	case ViewReview:
		viewIndicator = StatusBarStyle.Render(" REVIEW ")
	}

	searchView := m.search.View()

	left := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", projectInfo, " ", beatCount, " ", viewIndicator)
	right := searchView

	gap := m.width - lipgloss.Width(left) - lipgloss.Width(right) - 2
	if gap < 0 {
		gap = 0
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, left, lipgloss.NewStyle().Width(gap).Render(""), right) + "\n"
}

func (m ModelV2) renderFooterV2() string {
	focusIndicator := ""
	if m.focus == focusDetail {
		focusIndicator = StatusBarStyle.Render(" DETAIL ") + " "
	}

	help := "j/k:nav /:search f:facets e:entities t:timeline C:clusters R:ripeness ?:help q:quit"
	if m.width < 100 {
		help = "j/k / f e t C R ? q"
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

func (m ModelV2) renderHelpV2() string {
	// Fits on 80x24 terminal
	helpText := `btv v0.2 - beats_viewer

GLOBAL                        NAVIGATION
  q       Quit                  j/k     Up/down
  ?       This help             g/G     First/last
  Esc     Cancel/back           Enter   Select/expand
  /       Search                Tab     Cycle focus
  r       Refresh               [/]     Chain prev/next

VIEWS                         FILTERING
  t       Timeline              1-7     Channel filter
  C       Clusters              !       Clear filters
  S       Stale review          R       Sort by ripeness
  f       Facet sidebar         E       Entity search
  e       Entity sidebar

ACTIONS                       PROJECT
  y       Copy beat ID          p       Cycle projects
  Y       Copy content          a       All projects
  b       Create bead
  c       Add to chain

LAYOUT: Compact(<60) Normal(100) Wide(140) UltraWide(180)
Sidebars auto-show/hide based on terminal width.

Ripeness: ⚪Fresh 🟡Maturing 🟢Ripe 🔴Overripe

                    Press any key to close`
	return lipgloss.NewStyle().Padding(1, 2).Render(helpText)
}

func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsIgnoreCaseImpl(s, substr)))
}

func containsIgnoreCaseImpl(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if equalFoldAt(s, i, substr) {
			return true
		}
	}
	return false
}

func equalFoldAt(s string, start int, substr string) bool {
	for j := 0; j < len(substr); j++ {
		c1 := s[start+j]
		c2 := substr[j]
		if c1 != c2 && toLower(c1) != toLower(c2) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + 32
	}
	return c
}

type EnrichedBeatItem struct {
	beat    model.EnrichedBeat
	project string
}

func (i EnrichedBeatItem) Title() string {
	return i.beat.ID
}

func (i EnrichedBeatItem) Description() string {
	return i.beat.ContentPreview(60)
}

func (i EnrichedBeatItem) FilterValue() string {
	return i.beat.Content + " " + i.beat.Impetus.Label + " " + i.beat.ID
}

func EnrichedBeatsToItems(beats []model.EnrichedBeat, beatToProject map[string]string) []list.Item {
	items := make([]list.Item, len(beats))
	for i, b := range beats {
		project := ""
		if beatToProject != nil {
			project = beatToProject[b.ID]
		}
		items[i] = EnrichedBeatItem{beat: b, project: project}
	}
	return items
}

type EnrichedBeatDelegate struct {
	width int
}

func NewEnrichedBeatDelegate() EnrichedBeatDelegate {
	return EnrichedBeatDelegate{width: 80}
}

func (d EnrichedBeatDelegate) SetWidth(w int) EnrichedBeatDelegate {
	d.width = w
	return d
}

func (d EnrichedBeatDelegate) Height() int {
	return 2
}

func (d EnrichedBeatDelegate) Spacing() int {
	return 0
}

func (d EnrichedBeatDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd {
	return nil
}

func (d EnrichedBeatDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	bi, ok := item.(EnrichedBeatItem)
	if !ok {
		return
	}

	beat := bi.beat
	isSelected := index == m.Index()

	ripenessEmoji := model.RipenessEmoji(beat.RipenessScore)
	channelStr := beat.Taxonomy.Channel.String()
	if len(channelStr) > 12 {
		channelStr = channelStr[:12]
	}

	idWidth := 20
	channelWidth := 12
	contentWidth := d.width - idWidth - channelWidth - 8

	if contentWidth < 20 {
		contentWidth = 20
	}

	id := Truncate(beat.ID, idWidth)
	content := Truncate(beat.Content, contentWidth)

	line1 := fmt.Sprintf("%s %-*s  %-*s", ripenessEmoji, idWidth, id, channelWidth, channelStr)
	line2 := fmt.Sprintf("     %s", content)

	if isSelected {
		line1 = SelectedStyle.Render(line1)
		line2 = SelectedStyle.Render(line2)
	} else {
		line1 = IDStyle.Render(Truncate(beat.ID, idWidth)) + "  " + ImpetusStyle.Render(channelStr)
		line1 = ripenessEmoji + " " + line1
		line2 = "     " + ContentStyle.Render(content)
	}

	fmt.Fprintf(w, "%s\n%s\n", line1, line2)
}
