// Package types provides type discovery for festival templates.
// It scans template directories to discover available template types
// at each festival level (festival, phase, sequence, task).
package types

// Level represents a festival hierarchy level.
type Level string

const (
	// LevelFestival is the top-level festival entity.
	LevelFestival Level = "festival"
	// LevelPhase is a major division within a festival.
	LevelPhase Level = "phase"
	// LevelSequence is a group of related tasks within a phase.
	LevelSequence Level = "sequence"
	// LevelTask is an individual work item within a sequence.
	LevelTask Level = "task"
)

// AllLevels returns all festival hierarchy levels in order.
func AllLevels() []Level {
	return []Level{LevelFestival, LevelPhase, LevelSequence, LevelTask}
}

// TypeInfo contains metadata about a discovered template type.
type TypeInfo struct {
	// Name is the type name extracted from the template filename.
	// Examples: "goal", "implementation", "simple", "research"
	Name string `json:"name"`

	// Level indicates which hierarchy level this type applies to.
	// One of: festival, phase, sequence, task
	Level Level `json:"level"`

	// Description is a brief description of the type's purpose.
	Description string `json:"description,omitempty"`

	// Markers is the count of REPLACE markers in the template.
	Markers int `json:"markers"`

	// IsDefault indicates if this is the default type for its level.
	// Default types have no suffix (e.g., FESTIVAL_GOAL_TEMPLATE.md).
	IsDefault bool `json:"is_default"`

	// IsCustom indicates if this type came from a local custom template.
	// Custom templates are in .festival/templates/ within a festival.
	IsCustom bool `json:"is_custom"`

	// Templates lists the template files that define this type.
	// May include multiple files (e.g., goal and overview templates).
	Templates []string `json:"template_files,omitempty"`

	// Example shows a brief usage example for this type.
	Example string `json:"example,omitempty"`

	// Source is the path to the source directory for this type.
	Source string `json:"source,omitempty"`
}

// String returns a human-readable representation of the type.
func (t *TypeInfo) String() string {
	suffix := ""
	if t.IsDefault {
		suffix = " (default)"
	}
	if t.IsCustom {
		suffix = " (custom)"
	}
	return string(t.Level) + "/" + t.Name + suffix
}

// QualifiedName returns the fully qualified type name (level/name).
func (t *TypeInfo) QualifiedName() string {
	return string(t.Level) + "/" + t.Name
}
