package components

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"

	"github.com/charmbracelet/lipgloss"
)

var (
	entityTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4")).
				MarginTop(1)

	entityItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#1a1a1a", Dark: "#dddddd"})

	entitySelectedStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#73F59F"))
)

type EntityItem struct {
	Name    string
	Type    model.EntityType
	Count   int
	BeatIDs []string
}

type EntitySidebar struct {
	width  int
	height int

	people   []EntityItem
	tools    []EntityItem
	concepts []EntityItem

	selectedEntity *string
	cursorPos      int
	scrollOffset   int

	expandPeople   bool
	expandTools    bool
	expandConcepts bool
}

func NewEntitySidebar(width, height int) *EntitySidebar {
	return &EntitySidebar{
		width:          width,
		height:         height,
		expandPeople:   true,
		expandTools:    true,
		expandConcepts: true,
	}
}

func (e *EntitySidebar) SetSize(width, height int) {
	e.width = width
	e.height = height
}

func (e *EntitySidebar) UpdateEntities(entities []model.Entity) {
	entityMap := make(map[string]*EntityItem)

	for _, ent := range entities {
		key := fmt.Sprintf("%s-%d", ent.Name, ent.Type)
		if existing, ok := entityMap[key]; ok {
			existing.Count = len(ent.BeatIDs)
			existing.BeatIDs = ent.BeatIDs
		} else {
			entityMap[key] = &EntityItem{
				Name:    ent.Name,
				Type:    ent.Type,
				Count:   len(ent.BeatIDs),
				BeatIDs: ent.BeatIDs,
			}
		}
	}

	e.people = nil
	e.tools = nil
	e.concepts = nil

	for _, item := range entityMap {
		switch item.Type {
		case model.EntityPerson:
			e.people = append(e.people, *item)
		case model.EntityTool:
			e.tools = append(e.tools, *item)
		case model.EntityConcept:
			e.concepts = append(e.concepts, *item)
		}
	}

	sortByCount := func(items []EntityItem) {
		sort.Slice(items, func(i, j int) bool {
			return items[i].Count > items[j].Count
		})
	}

	sortByCount(e.people)
	sortByCount(e.tools)
	sortByCount(e.concepts)
}

func (e *EntitySidebar) SelectedEntity() *string {
	return e.selectedEntity
}

func (e *EntitySidebar) ClearSelection() {
	e.selectedEntity = nil
}

func (e *EntitySidebar) CursorUp() {
	if e.cursorPos > 0 {
		e.cursorPos--
		e.ensureVisible()
	}
}

func (e *EntitySidebar) CursorDown() {
	maxPos := e.totalItems() - 1
	if e.cursorPos < maxPos {
		e.cursorPos++
		e.ensureVisible()
	}
}

func (e *EntitySidebar) totalItems() int {
	count := 3 // section headers
	if e.expandPeople {
		count += len(e.people)
	}
	if e.expandTools {
		count += len(e.tools)
	}
	if e.expandConcepts {
		count += len(e.concepts)
	}
	return count
}

func (e *EntitySidebar) ensureVisible() {
	visibleHeight := e.height - 4
	if e.cursorPos < e.scrollOffset {
		e.scrollOffset = e.cursorPos
	} else if e.cursorPos >= e.scrollOffset+visibleHeight {
		e.scrollOffset = e.cursorPos - visibleHeight + 1
	}
}

func (e *EntitySidebar) ToggleSelection() {
	item := e.itemAtCursor()
	if item == nil {
		return
	}

	if e.selectedEntity != nil && *e.selectedEntity == item.Name {
		e.selectedEntity = nil
	} else {
		e.selectedEntity = &item.Name
	}
}

func (e *EntitySidebar) ToggleSection() {
	pos := 0

	if e.cursorPos == pos {
		e.expandPeople = !e.expandPeople
		return
	}
	pos++
	if e.expandPeople {
		pos += len(e.people)
	}

	if e.cursorPos == pos {
		e.expandTools = !e.expandTools
		return
	}
	pos++
	if e.expandTools {
		pos += len(e.tools)
	}

	if e.cursorPos == pos {
		e.expandConcepts = !e.expandConcepts
		return
	}
}

func (e *EntitySidebar) itemAtCursor() *EntityItem {
	pos := 0

	// People header
	if e.cursorPos == pos {
		return nil
	}
	pos++

	if e.expandPeople {
		for i := range e.people {
			if e.cursorPos == pos {
				return &e.people[i]
			}
			pos++
		}
	}

	// Tools header
	if e.cursorPos == pos {
		return nil
	}
	pos++

	if e.expandTools {
		for i := range e.tools {
			if e.cursorPos == pos {
				return &e.tools[i]
			}
			pos++
		}
	}

	// Concepts header
	if e.cursorPos == pos {
		return nil
	}
	pos++

	if e.expandConcepts {
		for i := range e.concepts {
			if e.cursorPos == pos {
				return &e.concepts[i]
			}
			pos++
		}
	}

	return nil
}

func (e *EntitySidebar) View() string {
	var lines []string
	pos := 0

	// People section
	arrow := "▶"
	if e.expandPeople {
		arrow = "▼"
	}
	header := entityTitleStyle.Render(fmt.Sprintf("%s People (%d)", arrow, len(e.people)))
	if e.cursorPos == pos {
		header = lipgloss.NewStyle().Background(lipgloss.Color("#383838")).Render(header)
	}
	lines = append(lines, header)
	pos++

	if e.expandPeople {
		for _, item := range e.people {
			line := e.renderEntityItem(item, e.cursorPos == pos)
			lines = append(lines, line)
			pos++
		}
	}

	// Tools section
	arrow = "▶"
	if e.expandTools {
		arrow = "▼"
	}
	header = entityTitleStyle.Render(fmt.Sprintf("%s Tools (%d)", arrow, len(e.tools)))
	if e.cursorPos == pos {
		header = lipgloss.NewStyle().Background(lipgloss.Color("#383838")).Render(header)
	}
	lines = append(lines, header)
	pos++

	if e.expandTools {
		for _, item := range e.tools {
			line := e.renderEntityItem(item, e.cursorPos == pos)
			lines = append(lines, line)
			pos++
		}
	}

	// Concepts section
	arrow = "▶"
	if e.expandConcepts {
		arrow = "▼"
	}
	header = entityTitleStyle.Render(fmt.Sprintf("%s Concepts (%d)", arrow, len(e.concepts)))
	if e.cursorPos == pos {
		header = lipgloss.NewStyle().Background(lipgloss.Color("#383838")).Render(header)
	}
	lines = append(lines, header)
	pos++

	if e.expandConcepts {
		for _, item := range e.concepts {
			line := e.renderEntityItem(item, e.cursorPos == pos)
			lines = append(lines, line)
			pos++
		}
	}

	// Apply scroll offset
	visibleHeight := e.height - 2
	if visibleHeight < 1 {
		visibleHeight = 1
	}

	start := e.scrollOffset
	end := start + visibleHeight
	if start > len(lines) {
		start = len(lines)
	}
	if end > len(lines) {
		end = len(lines)
	}

	visibleLines := lines[start:end]

	return lipgloss.NewStyle().
		Width(e.width).
		MaxHeight(e.height).
		Padding(0, 1).
		Render(strings.Join(visibleLines, "\n"))
}

func (e *EntitySidebar) renderEntityItem(item EntityItem, focused bool) string {
	selected := e.selectedEntity != nil && *e.selectedEntity == item.Name

	countStr := facetCountStyle.Render(fmt.Sprintf("(%d)", item.Count))

	var line string
	if selected {
		line = entitySelectedStyle.Render(fmt.Sprintf("  %s", item.Name))
	} else {
		line = entityItemStyle.Render(fmt.Sprintf("  %s", item.Name))
	}

	result := fmt.Sprintf("%s %s", line, countStr)

	if focused {
		result = lipgloss.NewStyle().
			Background(lipgloss.Color("#383838")).
			Render(result)
	}

	return result
}

func FilterByEntity(beats []model.EnrichedBeat, entityName *string, entityIndex map[string][]string) []model.EnrichedBeat {
	if entityName == nil {
		return beats
	}

	beatIDs := entityIndex[*entityName]
	if len(beatIDs) == 0 {
		return nil
	}

	beatIDSet := make(map[string]bool)
	for _, id := range beatIDs {
		beatIDSet[id] = true
	}

	var filtered []model.EnrichedBeat
	for _, eb := range beats {
		if beatIDSet[eb.ID] {
			filtered = append(filtered, eb)
		}
	}
	return filtered
}
