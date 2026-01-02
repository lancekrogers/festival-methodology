// Package status implements the fest status command for managing entity statuses.
package status

// PhaseInfo holds information about a phase.
type PhaseInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Status    string       `json:"status"`
	TaskStats StatusCounts `json:"task_stats,omitempty"`
}

// SequenceInfo holds information about a sequence.
type SequenceInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	PhaseName string       `json:"phase_name"`
	Status    string       `json:"status"`
	TaskStats StatusCounts `json:"task_stats,omitempty"`
}

// TaskInfo holds information about a task.
type TaskInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PhaseName    string `json:"phase_name"`
	SequenceName string `json:"sequence_name"`
	Status       string `json:"status"`
}

// StatusCounts tracks entity completion.
type StatusCounts struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
	Blocked    int `json:"blocked,omitempty"`
}
