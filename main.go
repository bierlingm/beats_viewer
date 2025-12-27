package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const version = "0.1.0"

type model struct {
	message string
}

func initialModel() model {
	return model{
		message: "beats_viewer v" + version + " - TUI coming soon",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Quit
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, tea.Quit
}

func (m model) View() string {
	return m.message + "\n"
}

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-v" || os.Args[1] == "--version") {
		fmt.Println("beats_viewer", version)
		return
	}

	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
