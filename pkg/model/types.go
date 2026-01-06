package model

import (
	"encoding/json"
	"time"
)

type ImpetusMeta map[string]string

type Impetus struct {
	Label string      `json:"label"`
	Raw   string      `json:"raw,omitempty"`
	Meta  ImpetusMeta `json:"meta,omitempty"`
}

type Beat struct {
	ID          string    `json:"id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Impetus     Impetus   `json:"impetus"`
	Content     string    `json:"content"`
	Entities    []string  `json:"entities,omitempty"`
	References  []string  `json:"references,omitempty"`
	LinkedBeads []string  `json:"linked_beads,omitempty"`
}

type Project struct {
	Name      string
	Path      string
	BeatCount int
}

func (b Beat) ContentPreview(maxLen int) string {
	content := b.Content
	if len(content) > maxLen {
		return content[:maxLen-3] + "..."
	}
	return content
}

func (b Beat) ImpetusLabel() string {
	if b.Impetus.Label != "" {
		return b.Impetus.Label
	}
	return "unknown"
}

func (b Beat) ToJSON() ([]byte, error) {
	return json.Marshal(b)
}

type BeatListItem struct {
	ID             string    `json:"id"`
	ContentPreview string    `json:"content_preview"`
	ImpetusLabel   string    `json:"impetus_label"`
	Project        string    `json:"project"`
	CreatedAt      time.Time `json:"created_at"`
}

func (b Beat) ToListItem(project string, previewLen int) BeatListItem {
	return BeatListItem{
		ID:             b.ID,
		ContentPreview: b.ContentPreview(previewLen),
		ImpetusLabel:   b.ImpetusLabel(),
		Project:        project,
		CreatedAt:      b.CreatedAt,
	}
}

type RobotListResponse struct {
	Beats         []BeatListItem `json:"beats"`
	Total         int            `json:"total"`
	ProjectFilter *string        `json:"project_filter"`
}

type RobotSearchResponse struct {
	Results      []BeatListItem `json:"results"`
	Query        string         `json:"query"`
	TotalMatches int            `json:"total_matches"`
}

type RobotHelpCommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Input       string `json:"input,omitempty"`
	Output      string `json:"output"`
}

type RobotHelpResponse struct {
	Version  string             `json:"version"`
	Commands []RobotHelpCommand `json:"commands"`
}
