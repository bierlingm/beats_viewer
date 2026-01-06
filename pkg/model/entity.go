package model

// EntityType represents the type of extracted entity
type EntityType int

const (
	EntityPerson EntityType = iota
	EntityTool
	EntityConcept
	EntityProject
	EntityOrganization
)

func (e EntityType) String() string {
	switch e {
	case EntityPerson:
		return "Person"
	case EntityTool:
		return "Tool"
	case EntityConcept:
		return "Concept"
	case EntityProject:
		return "Project"
	case EntityOrganization:
		return "Organization"
	default:
		return "Unknown"
	}
}

// AllEntityTypes returns all valid entity types
func AllEntityTypes() []EntityType {
	return []EntityType{
		EntityPerson,
		EntityTool,
		EntityConcept,
		EntityProject,
		EntityOrganization,
	}
}

// Entity represents an extracted entity from beats
type Entity struct {
	Name    string     `json:"name"`
	Type    EntityType `json:"type"`
	BeatIDs []string   `json:"beat_ids"`
}
