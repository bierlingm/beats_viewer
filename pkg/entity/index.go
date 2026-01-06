package entity

import (
	"sort"
	"strings"

	"github.com/bierlingm/beats_viewer/pkg/model"
)

// Index provides fast entity lookups
type Index struct {
	entities    []model.Entity
	byName      map[string]*model.Entity
	byType      map[model.EntityType][]*model.Entity
	byBeatID    map[string][]*model.Entity
	entityIndex map[string][]string
}

// NewIndex creates a new entity index from entities
func NewIndex(entities []model.Entity, entityIndex map[string][]string) *Index {
	idx := &Index{
		entities:    entities,
		byName:      make(map[string]*model.Entity),
		byType:      make(map[model.EntityType][]*model.Entity),
		byBeatID:    make(map[string][]*model.Entity),
		entityIndex: entityIndex,
	}

	for i := range entities {
		e := &entities[i]
		key := strings.ToLower(e.Name)
		idx.byName[key] = e
		idx.byType[e.Type] = append(idx.byType[e.Type], e)
		for _, beatID := range e.BeatIDs {
			idx.byBeatID[beatID] = append(idx.byBeatID[beatID], e)
		}
	}

	return idx
}

// GetByName returns an entity by name (case-insensitive)
func (idx *Index) GetByName(name string) *model.Entity {
	return idx.byName[strings.ToLower(name)]
}

// GetByType returns all entities of a given type
func (idx *Index) GetByType(entityType model.EntityType) []*model.Entity {
	return idx.byType[entityType]
}

// GetForBeat returns all entities in a specific beat
func (idx *Index) GetForBeat(beatID string) []*model.Entity {
	return idx.byBeatID[beatID]
}

// GetBeatIDsForEntity returns beat IDs containing an entity
func (idx *Index) GetBeatIDsForEntity(name string) []string {
	return idx.entityIndex[name]
}

// AllEntities returns all entities
func (idx *Index) AllEntities() []model.Entity {
	return idx.entities
}

// TopEntitiesByType returns the most common entities of each type
func (idx *Index) TopEntitiesByType(limit int) map[model.EntityType][]model.Entity {
	result := make(map[model.EntityType][]model.Entity)

	for _, entityType := range model.AllEntityTypes() {
		entities := idx.byType[entityType]
		sorted := make([]*model.Entity, len(entities))
		copy(sorted, entities)

		sort.Slice(sorted, func(i, j int) bool {
			return len(sorted[i].BeatIDs) > len(sorted[j].BeatIDs)
		})

		count := limit
		if count > len(sorted) {
			count = len(sorted)
		}

		for i := 0; i < count; i++ {
			result[entityType] = append(result[entityType], *sorted[i])
		}
	}

	return result
}
