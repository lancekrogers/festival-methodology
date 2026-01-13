package graduate

import "time"

// PlanningSource describes a planning phase to graduate from.
type PlanningSource struct {
	Path       string           `json:"path"`
	PhaseName  string           `json:"phase_name"`
	TopicDirs  []TopicDirectory `json:"topic_dirs"`
	Decisions  []Decision       `json:"decisions,omitempty"`
	Summary    *PlanningSummary `json:"summary,omitempty"`
	TotalDocs  int              `json:"total_docs"`
	AnalyzedAt time.Time        `json:"analyzed_at"`
}

// TopicDirectory represents a topic directory in the planning phase.
type TopicDirectory struct {
	Name      string   `json:"name"`
	Path      string   `json:"path"`
	Documents []string `json:"documents"`
	DocCount  int      `json:"doc_count"`
}

// Decision represents an Architecture Decision Record.
type Decision struct {
	ID       string `json:"id"` // e.g., "ADR-001"
	Title    string `json:"title"`
	Status   string `json:"status"` // accepted, proposed, deprecated
	FilePath string `json:"file_path"`
	Summary  string `json:"summary,omitempty"`
}

// PlanningSummary from PLANNING_SUMMARY.md.
type PlanningSummary struct {
	Goal              string   `json:"goal"`
	KeyDecisions      []string `json:"key_decisions"`
	ProposedSequences []string `json:"proposed_sequences,omitempty"`
	FilePath          string   `json:"file_path"`
}

// GraduationPlan is the complete plan for transitioning to implementation.
type GraduationPlan struct {
	Source     PlanningSource       `json:"source"`
	Target     ImplementationTarget `json:"target"`
	PhaseGoal  GeneratedContent     `json:"phase_goal"`
	Sequences  []ProposedSequence   `json:"sequences"`
	Warnings   []string             `json:"warnings,omitempty"`
	Confidence float64              `json:"confidence"`
}

// ImplementationTarget describes the target implementation phase.
type ImplementationTarget struct {
	Path      string `json:"path"`
	PhaseName string `json:"phase_name"`
	Number    int    `json:"number"`
}

// ProposedSequence is a sequence to be created.
type ProposedSequence struct {
	Number      int              `json:"number"`
	Name        string           `json:"name"`
	FullName    string           `json:"full_name"` // e.g., "01_core_setup"
	Goal        GeneratedContent `json:"goal"`
	SourceTopic string           `json:"source_topic"`
	Tasks       []ProposedTask   `json:"tasks"`
}

// ProposedTask is a task to be created.
type ProposedTask struct {
	Number       int      `json:"number"`
	Name         string   `json:"name"`
	FullName     string   `json:"full_name"` // e.g., "01_setup.md"
	Objective    string   `json:"objective"`
	SourceDocs   []string `json:"source_docs"`
	Dependencies []string `json:"dependencies,omitempty"`
}

// GeneratedContent holds generated markdown content.
type GeneratedContent struct {
	Title    string   `json:"title"`
	Goal     string   `json:"goal"`
	Sections []string `json:"sections,omitempty"`
}
