package parser

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// FormatJSON formats the entity as JSON
func FormatJSON(entity interface{}, compact bool) ([]byte, error) {
	if compact {
		return json.Marshal(entity)
	}
	return json.MarshalIndent(entity, "", "  ")
}

// FormatYAML formats the entity as YAML
func FormatYAML(entity interface{}) ([]byte, error) {
	return yaml.Marshal(entity)
}

// Format outputs in the specified format
func Format(entity interface{}, format string, compact bool) ([]byte, error) {
	switch format {
	case "yaml":
		return FormatYAML(entity)
	default:
		return FormatJSON(entity, compact)
	}
}

// FlattenedResult holds flattened entities for filtered queries
type FlattenedResult struct {
	Query    string         `json:"query" yaml:"query"`
	Count    int            `json:"count" yaml:"count"`
	Entities []ParsedEntity `json:"entities" yaml:"entities"`
}

// FlattenTasks extracts all tasks from a festival into a flat list
func FlattenTasks(festival *ParsedFestival) []ParsedEntity {
	var entities []ParsedEntity

	for _, phase := range festival.Phases {
		for _, seq := range phase.Sequences {
			for _, task := range seq.Tasks {
				entities = append(entities, task.ParsedEntity)
			}
		}
	}

	return entities
}

// FlattenByType extracts all entities of a specific type
func FlattenByType(festival *ParsedFestival, entityType string) []ParsedEntity {
	var entities []ParsedEntity

	switch entityType {
	case "festival":
		entities = append(entities, festival.ParsedEntity)
	case "phase":
		for _, phase := range festival.Phases {
			entities = append(entities, phase.ParsedEntity)
		}
	case "sequence":
		for _, phase := range festival.Phases {
			for _, seq := range phase.Sequences {
				entities = append(entities, seq.ParsedEntity)
			}
		}
	case "task":
		for _, phase := range festival.Phases {
			for _, seq := range phase.Sequences {
				for _, task := range seq.Tasks {
					if task.Type == "task" || task.Type == "" {
						entities = append(entities, task.ParsedEntity)
					}
				}
			}
		}
	case "gate":
		for _, phase := range festival.Phases {
			for _, seq := range phase.Sequences {
				for _, task := range seq.Tasks {
					if task.Type == "gate" {
						entities = append(entities, task.ParsedEntity)
					}
				}
			}
		}
	}

	return entities
}

// FilterByStatus filters entities by status
func FilterByStatus(entities []ParsedEntity, status string) []ParsedEntity {
	var filtered []ParsedEntity
	for _, e := range entities {
		if e.Status == status {
			filtered = append(filtered, e)
		}
	}
	return filtered
}

// AllFestivalsResult holds multiple parsed festivals
type AllFestivalsResult struct {
	Festivals []ParsedFestival  `json:"festivals" yaml:"festivals"`
	Summary   *WorkspaceSummary `json:"summary" yaml:"summary"`
}

// WorkspaceSummary holds workspace-level statistics
type WorkspaceSummary struct {
	FestivalCount int `json:"festival_count" yaml:"festival_count"`
	PhaseCount    int `json:"phase_count" yaml:"phase_count"`
	TaskCount     int `json:"task_count" yaml:"task_count"`
}
