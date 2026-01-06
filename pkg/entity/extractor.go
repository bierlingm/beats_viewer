package entity

import (
	"regexp"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

var capitalizedNamePattern = regexp.MustCompile(`\b([A-Z][a-z]+(?:\s+[A-Z][a-z]+)?)\b`)

// Extract extracts entities from a single beat
func Extract(beat model.Beat) []model.Entity {
	var entities []model.Entity
	content := beat.Content
	contentLower := strings.ToLower(content)

	seen := make(map[string]bool)

	for entityType, dict := range EntityDictionaries {
		for _, name := range dict {
			if strings.Contains(contentLower, strings.ToLower(name)) {
				key := strings.ToLower(name) + "-" + entityType.String()
				if !seen[key] {
					seen[key] = true
					entities = append(entities, model.Entity{
						Name:    name,
						Type:    entityType,
						BeatIDs: []string{beat.ID},
					})
				}
			}
		}
	}

	matches := capitalizedNamePattern.FindAllString(content, -1)
	for _, match := range matches {
		if isCommonWord(match) {
			continue
		}
		key := strings.ToLower(match) + "-person"
		if !seen[key] {
			seen[key] = true
			entities = append(entities, model.Entity{
				Name:    match,
				Type:    model.EntityPerson,
				BeatIDs: []string{beat.ID},
			})
		}
	}

	return entities
}

// ExtractAll extracts entities from all beats and builds an index
func ExtractAll(beats []model.Beat) ([]model.Entity, map[string][]string) {
	entityMap := make(map[string]*model.Entity)
	entityIndex := make(map[string][]string)

	for _, beat := range beats {
		extracted := Extract(beat)
		for _, e := range extracted {
			key := strings.ToLower(e.Name) + "-" + e.Type.String()
			if existing, ok := entityMap[key]; ok {
				existing.BeatIDs = append(existing.BeatIDs, beat.ID)
			} else {
				entity := e
				entityMap[key] = &entity
			}
			entityIndex[e.Name] = append(entityIndex[e.Name], beat.ID)
		}
	}

	var entities []model.Entity
	for _, e := range entityMap {
		entities = append(entities, *e)
	}

	return entities, entityIndex
}

var commonWords = map[string]bool{
	"The": true, "This": true, "That": true, "These": true, "Those": true,
	"What": true, "When": true, "Where": true, "Which": true, "Who": true,
	"How": true, "Why": true, "Some": true, "Many": true, "Most": true,
	"Such": true, "Each": true, "Every": true, "Both": true, "All": true,
	"Any": true, "Other": true, "Another": true, "First": true, "Last": true,
	"New": true, "Old": true, "Good": true, "Great": true, "Best": true,
	"Just": true, "Only": true, "Also": true, "Even": true, "Still": true,
	"Now": true, "Here": true, "There": true, "Today": true, "Tomorrow": true,
	"Monday": true, "Tuesday": true, "Wednesday": true, "Thursday": true,
	"Friday": true, "Saturday": true, "Sunday": true,
	"January": true, "February": true, "March": true, "April": true,
	"May": true, "June": true, "July": true, "August": true,
	"September": true, "October": true, "November": true, "December": true,
}

func isCommonWord(word string) bool {
	return commonWords[word]
}

// GetEntitiesByType filters entities by type
func GetEntitiesByType(entities []model.Entity, entityType model.EntityType) []model.Entity {
	var result []model.Entity
	for _, e := range entities {
		if e.Type == entityType {
			result = append(result, e)
		}
	}
	return result
}
