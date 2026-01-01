// Package context provides context aggregation for AI agents working on festival tasks.
package context

import "time"

// Depth controls how much context is included
type Depth string

const (
	DepthMinimal  Depth = "minimal"  // Immediate goals, dependencies, autonomy
	DepthStandard Depth = "standard" // + Rules, recent decisions
	DepthFull     Depth = "full"     // + All decisions, dependency outputs
)

// Location represents the current position in the festival hierarchy
type Location struct {
	FestivalPath string `json:"festival_path"`
	FestivalName string `json:"festival_name"`
	PhasePath    string `json:"phase_path,omitempty"`
	PhaseName    string `json:"phase_name,omitempty"`
	SequencePath string `json:"sequence_path,omitempty"`
	SequenceName string `json:"sequence_name,omitempty"`
	TaskPath     string `json:"task_path,omitempty"`
	TaskName     string `json:"task_name,omitempty"`
	Level        string `json:"level"` // festival, phase, sequence, task
}

// GoalContext holds parsed information from a goal file
type GoalContext struct {
	Title           string   `json:"title"`
	Objective       string   `json:"objective"`
	SuccessCriteria []string `json:"success_criteria,omitempty"`
	Status          string   `json:"status,omitempty"`
	Priority        string   `json:"priority,omitempty"`
}

// FestivalContext holds festival-level context
type FestivalContext struct {
	Name        string       `json:"name"`
	Path        string       `json:"path"`
	Goal        *GoalContext `json:"goal,omitempty"`
	PhaseCount  int          `json:"phase_count"`
	ActivePhase string       `json:"active_phase,omitempty"`
}

// PhaseContext holds phase-level context
type PhaseContext struct {
	Name          string       `json:"name"`
	Path          string       `json:"path"`
	Goal          *GoalContext `json:"goal,omitempty"`
	PhaseType     string       `json:"phase_type,omitempty"`
	SequenceCount int          `json:"sequence_count"`
}

// SequenceContext holds sequence-level context
type SequenceContext struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Goal      *GoalContext `json:"goal,omitempty"`
	TaskCount int          `json:"task_count"`
}

// TaskContext holds task-level context
type TaskContext struct {
	Name            string   `json:"name"`
	Path            string   `json:"path"`
	Objective       string   `json:"objective,omitempty"`
	TaskNumber      int      `json:"task_number"`
	AutonomyLevel   string   `json:"autonomy_level,omitempty"`
	ParallelAllowed bool     `json:"parallel_allowed"`
	Dependencies    []string `json:"dependencies,omitempty"`
	Deliverables    []string `json:"deliverables,omitempty"`
}

// Rule represents a festival rule
type Rule struct {
	Category    string `json:"category"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority,omitempty"`
}

// Decision represents a recorded decision from CONTEXT.md
type Decision struct {
	Date      time.Time `json:"date"`
	Summary   string    `json:"summary"`
	Rationale string    `json:"rationale,omitempty"`
	Impact    string    `json:"impact,omitempty"`
	Category  string    `json:"category,omitempty"`
}

// DepOutput represents output from a completed dependency task
type DepOutput struct {
	TaskID      string   `json:"task_id"`
	TaskName    string   `json:"task_name"`
	Outputs     []string `json:"outputs"`
	CompletedAt string   `json:"completed_at,omitempty"`
}

// ContextOutput is the complete context response
type ContextOutput struct {
	Location          *Location        `json:"location"`
	FestivalID        *string          `json:"festival_id"` // Unique festival ID from metadata (null if legacy)
	CurrentRef        *string          `json:"current_ref"` // Full node reference (e.g., GU0001:P002.S01.T03)
	Festival          *FestivalContext `json:"festival,omitempty"`
	Phase             *PhaseContext    `json:"phase,omitempty"`
	Sequence          *SequenceContext `json:"sequence,omitempty"`
	Task              *TaskContext     `json:"task,omitempty"`
	Rules             []Rule           `json:"rules,omitempty"`
	Decisions         []Decision       `json:"decisions,omitempty"`
	DependencyOutputs []DepOutput      `json:"dependency_outputs,omitempty"`
	Depth             Depth            `json:"depth"`
}
