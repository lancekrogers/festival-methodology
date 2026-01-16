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

// PhaseType represents the type of work in a phase
type PhaseType string

const (
	PhaseTypePlanning       PhaseType = "planning"
	PhaseTypeImplementation PhaseType = "implementation"
	PhaseTypeResearch       PhaseType = "research"
	PhaseTypeReview         PhaseType = "review"
	PhaseTypeDeployment     PhaseType = "deployment"
)

// FestivalType represents the scope/complexity of a festival
type FestivalType string

const (
	FestivalTypeBugfix        FestivalType = "bugfix"        // Quick bug fixes, minimal structure
	FestivalTypeFeature       FestivalType = "feature"       // Single feature implementation
	FestivalTypeProject       FestivalType = "project"       // Multi-component work
	FestivalTypeComprehensive FestivalType = "comprehensive" // Major initiative with full planning
)

// SequenceType represents the complexity of a sequence
type SequenceType string

const (
	SequenceTypeMinimal  SequenceType = "minimal"  // Simple sequences with few tasks
	SequenceTypeStandard SequenceType = "standard" // Normal sequences
	SequenceTypeDetailed SequenceType = "detailed" // Complex sequences with dependencies
)

// TaskType represents task template selection
type TaskType string

const (
	TaskTypeMinimal  TaskType = "minimal"  // Ultra-simple tasks
	TaskTypeSimple   TaskType = "simple"   // Standard tasks
	TaskTypeDetailed TaskType = "detailed" // Full documentation
)

// Complexity represents task complexity level for routing
type Complexity string

const (
	ComplexityLow      Complexity = "low"
	ComplexityMedium   Complexity = "medium"
	ComplexityHigh     Complexity = "high"
	ComplexityCritical Complexity = "critical"
)

// Frontmatter represents the YAML frontmatter in festival documents
type Frontmatter struct {
	// Core fields
	Type     Type      `yaml:"fest_type" json:"fest_type"`
	ID       string    `yaml:"fest_id" json:"fest_id"`
	Ref      string    `yaml:"fest_ref,omitempty" json:"fest_ref,omitempty"` // Unique short-hash ID for commit tracing
	Name     string    `yaml:"fest_name,omitempty" json:"fest_name,omitempty"`
	Parent   string    `yaml:"fest_parent,omitempty" json:"fest_parent,omitempty"`
	Order    int       `yaml:"fest_order,omitempty" json:"fest_order,omitempty"`
	Status   Status    `yaml:"fest_status" json:"fest_status"`
	Priority Priority  `yaml:"fest_priority,omitempty" json:"fest_priority,omitempty"`
	Autonomy Autonomy  `yaml:"fest_autonomy,omitempty" json:"fest_autonomy,omitempty"`
	GateType GateType  `yaml:"fest_gate_type,omitempty" json:"fest_gate_type,omitempty"`
	Managed  bool      `yaml:"fest_managed,omitempty" json:"fest_managed,omitempty"` // Deprecated: use Tracking
	Tags     []string  `yaml:"fest_tags,omitempty" json:"fest_tags,omitempty"`
	Created  time.Time `yaml:"fest_created" json:"fest_created"`
	Updated  time.Time `yaml:"fest_updated,omitempty" json:"fest_updated,omitempty"`

	// Type-awareness fields
	FestivalType FestivalType `yaml:"fest_festival_type,omitempty" json:"fest_festival_type,omitempty"`
	PhaseType    PhaseType    `yaml:"fest_phase_type,omitempty" json:"fest_phase_type,omitempty"`
	SequenceType SequenceType `yaml:"fest_sequence_type,omitempty" json:"fest_sequence_type,omitempty"`
	TaskType     TaskType     `yaml:"fest_task_type,omitempty" json:"fest_task_type,omitempty"`
	Tracking     *bool        `yaml:"fest_tracking,omitempty" json:"fest_tracking,omitempty"` // Pointer to distinguish unset from false
	Version      string       `yaml:"fest_version,omitempty" json:"fest_version,omitempty"`

	// Task-specific fields
	Dependencies  []string `yaml:"fest_dependencies,omitempty" json:"fest_dependencies,omitempty"`
	ParallelGroup string   `yaml:"fest_parallel_group,omitempty" json:"fest_parallel_group,omitempty"`

	// Future fields for task routing (reserved)
	Agent           string     `yaml:"fest_agent,omitempty" json:"fest_agent,omitempty"`
	Complexity      Complexity `yaml:"fest_complexity,omitempty" json:"fest_complexity,omitempty"`
	EstimatedTokens int        `yaml:"fest_estimated_tokens,omitempty" json:"fest_estimated_tokens,omitempty"`
	RequiresHuman   bool       `yaml:"fest_requires_human,omitempty" json:"fest_requires_human,omitempty"`
	RequiresContext bool       `yaml:"fest_requires_context,omitempty" json:"fest_requires_context,omitempty"`
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
	tracking := true
	return &Frontmatter{
		Type:     docType,
		ID:       id,
		Name:     name,
		Status:   DefaultStatus(docType),
		Tracking: &tracking,
		Created:  time.Now(),
	}
}

// NewPhaseFrontmatter creates frontmatter for a phase document
func NewPhaseFrontmatter(id, name, parent string, order int, phaseType PhaseType) *Frontmatter {
	tracking := true
	return &Frontmatter{
		Type:      TypePhase,
		ID:        id,
		Name:      name,
		Parent:    parent,
		Order:     order,
		PhaseType: phaseType,
		Status:    StatusPending,
		Tracking:  &tracking,
		Created:   time.Now(),
	}
}

// NewSequenceFrontmatter creates frontmatter for a sequence document
func NewSequenceFrontmatter(id, name, parent string, order int) *Frontmatter {
	tracking := true
	return &Frontmatter{
		Type:     TypeSequence,
		ID:       id,
		Name:     name,
		Parent:   parent,
		Order:    order,
		Status:   StatusPending,
		Tracking: &tracking,
		Created:  time.Now(),
	}
}

// NewTaskFrontmatter creates frontmatter for a task document
func NewTaskFrontmatter(id, name, parent string, order int, autonomy Autonomy) *Frontmatter {
	tracking := true
	return &Frontmatter{
		Type:     TypeTask,
		ID:       id,
		Name:     name,
		Parent:   parent,
		Order:    order,
		Autonomy: autonomy,
		Status:   StatusPending,
		Tracking: &tracking,
		Created:  time.Now(),
	}
}

// NewGateFrontmatter creates frontmatter for a quality gate document
func NewGateFrontmatter(id, name, parent string, order int, gateType GateType) *Frontmatter {
	tracking := true
	return &Frontmatter{
		Type:     TypeGate,
		ID:       id,
		Name:     name,
		Parent:   parent,
		Order:    order,
		GateType: gateType,
		Status:   StatusPending,
		Tracking: &tracking,
		Created:  time.Now(),
	}
}

// IsTracked returns whether the document should be tracked in progress
func (f *Frontmatter) IsTracked() bool {
	if f.Tracking != nil {
		return *f.Tracking
	}
	// Legacy fallback: check Managed field
	return !f.Managed
}

// DefaultPhaseType returns the default phase type
func DefaultPhaseType() PhaseType {
	return PhaseTypeImplementation
}

// DefaultComplexity returns the default complexity
func DefaultComplexity() Complexity {
	return ComplexityMedium
}
