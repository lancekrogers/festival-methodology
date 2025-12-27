// Package frontmatter provides YAML frontmatter support for festival documents.
package frontmatter

import (
	"time"
)

// Type represents the document type
type Type string

const (
	TypeFestival Type = "festival"
	TypePhase    Type = "phase"
	TypeSequence Type = "sequence"
	TypeTask     Type = "task"
	TypeGate     Type = "gate"
)

// Status represents document status
type Status string

// Festival statuses
const (
	StatusPlanned   Status = "planned"
	StatusActive    Status = "active"
	StatusCompleted Status = "completed"
	StatusDungeon   Status = "dungeon"
)

// Task statuses
const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusBlocked    Status = "blocked"
	// StatusCompleted also applies to tasks
)

// Gate statuses
const (
	StatusPassed Status = "passed"
	StatusFailed Status = "failed"
	// StatusPending also applies to gates
)

// GateType represents the type of quality gate
type GateType string

const (
	GateTesting     GateType = "testing"
	GateReview      GateType = "review"
	GateIterate     GateType = "iterate"
	GateSecurity    GateType = "security"
	GatePerformance GateType = "performance"
)

// Autonomy represents task autonomy level
type Autonomy string

const (
	AutonomyHigh   Autonomy = "high"
	AutonomyMedium Autonomy = "medium"
	AutonomyLow    Autonomy = "low"
)

// Priority represents festival priority
type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

// Frontmatter represents the YAML frontmatter in festival documents
type Frontmatter struct {
	Type     Type      `yaml:"fest_type" json:"fest_type"`
	ID       string    `yaml:"fest_id" json:"fest_id"`
	Name     string    `yaml:"fest_name,omitempty" json:"fest_name,omitempty"`
	Parent   string    `yaml:"fest_parent,omitempty" json:"fest_parent,omitempty"`
	Order    int       `yaml:"fest_order,omitempty" json:"fest_order,omitempty"`
	Status   Status    `yaml:"fest_status" json:"fest_status"`
	Priority Priority  `yaml:"fest_priority,omitempty" json:"fest_priority,omitempty"`
	Autonomy Autonomy  `yaml:"fest_autonomy,omitempty" json:"fest_autonomy,omitempty"`
	GateType GateType  `yaml:"fest_gate_type,omitempty" json:"fest_gate_type,omitempty"`
	Managed  bool      `yaml:"fest_managed,omitempty" json:"fest_managed,omitempty"`
	Tags     []string  `yaml:"fest_tags,omitempty" json:"fest_tags,omitempty"`
	Created  time.Time `yaml:"fest_created" json:"fest_created"`
	Updated  time.Time `yaml:"fest_updated,omitempty" json:"fest_updated,omitempty"`
}

// Validate checks if the frontmatter is valid
func (f *Frontmatter) Validate() []string {
	var errors []string

	if f.Type == "" {
		errors = append(errors, "fest_type is required")
	}

	if f.ID == "" {
		errors = append(errors, "fest_id is required")
	}

	if f.Status == "" {
		errors = append(errors, "fest_status is required")
	}

	if f.Created.IsZero() {
		errors = append(errors, "fest_created is required")
	}

	// Type-specific validation
	switch f.Type {
	case TypeFestival:
		// Festival doesn't need parent
	case TypePhase, TypeSequence, TypeTask, TypeGate:
		if f.Parent == "" {
			errors = append(errors, "fest_parent is required for "+string(f.Type))
		}
		if f.Order == 0 {
			errors = append(errors, "fest_order is required for "+string(f.Type))
		}
	}

	// Validate status values
	if !isValidStatus(f.Type, f.Status) {
		errors = append(errors, "invalid status value: "+string(f.Status))
	}

	return errors
}

// isValidStatus checks if status is valid for the document type
func isValidStatus(docType Type, status Status) bool {
	switch docType {
	case TypeFestival:
		return status == StatusPlanned || status == StatusActive ||
			status == StatusCompleted || status == StatusDungeon
	case TypePhase, TypeSequence:
		return status == StatusPending || status == StatusInProgress ||
			status == StatusCompleted
	case TypeTask:
		return status == StatusPending || status == StatusInProgress ||
			status == StatusBlocked || status == StatusCompleted
	case TypeGate:
		return status == StatusPending || status == StatusPassed ||
			status == StatusFailed
	}
	return false
}

// DefaultStatus returns the default status for a document type
func DefaultStatus(docType Type) Status {
	switch docType {
	case TypeFestival:
		return StatusPlanned
	case TypePhase, TypeSequence, TypeTask, TypeGate:
		return StatusPending
	}
	return StatusPending
}

// NewFrontmatter creates a new frontmatter with defaults
func NewFrontmatter(docType Type, id, name string) *Frontmatter {
	return &Frontmatter{
		Type:    docType,
		ID:      id,
		Name:    name,
		Status:  DefaultStatus(docType),
		Created: time.Now(),
	}
}
