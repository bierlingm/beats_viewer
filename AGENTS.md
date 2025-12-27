# beats_viewer

TUI browser for beats - similar to beads_viewer for beads.

## Overview

Interactive terminal UI for browsing, searching, and managing beats across projects.

## Planned Features

- List/browse beats with filtering
- Full-text search within TUI
- Cross-project navigation
- Beat detail view with linked beads
- Theme/brief generation from selected beats
- Robot commands for AI agent integration

## Architecture

Based on bubbletea TUI framework (like beads_viewer).

## Build & Install

```bash
go build -o beats_viewer .
cp beats_viewer ~/.local/bin/
ln -sf ~/.local/bin/beats_viewer ~/.local/bin/btv  # short alias
```

## Dependencies

- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/bubbles
