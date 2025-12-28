package execute

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// ExecutionState persists execution progress
type ExecutionState struct {
	FestivalPath   string            `yaml:"festival_path" json:"festival_path"`
	CurrentPhase   int               `yaml:"current_phase" json:"current_phase"`
	CurrentSeq     int               `yaml:"current_sequence" json:"current_sequence"`
	CurrentStep    int               `yaml:"current_step" json:"current_step"`
	Mode           ExecutionMode     `yaml:"mode" json:"mode"`
	TaskStatuses   map[string]string `yaml:"task_statuses" json:"task_statuses"`
	StartedAt      time.Time         `yaml:"started_at" json:"started_at"`
	LastUpdated    time.Time         `yaml:"last_updated" json:"last_updated"`
	TotalTasks     int               `yaml:"total_tasks" json:"total_tasks"`
	CompletedTasks int               `yaml:"completed_tasks" json:"completed_tasks"`
	SkippedTasks   int               `yaml:"skipped_tasks" json:"skipped_tasks"`
	FailedTasks    int               `yaml:"failed_tasks" json:"failed_tasks"`
}

// TaskStatus represents the status of a task
const (
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusCompleted  = "completed"
	StatusSkipped    = "skipped"
	StatusFailed     = "failed"
)

// StateManager handles execution state persistence
type StateManager struct {
	festivalPath string
	statePath    string
	state        *ExecutionState
}

// NewStateManager creates a new state manager
func NewStateManager(festivalPath string) *StateManager {
	statePath := filepath.Join(festivalPath, ".fest", "execution_state.yaml")
	return &StateManager{
		festivalPath: festivalPath,
		statePath:    statePath,
	}
}

// Load loads the execution state from disk
func (m *StateManager) Load() (*ExecutionState, error) {
	data, err := os.ReadFile(m.statePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Return new state
			m.state = &ExecutionState{
				FestivalPath: m.festivalPath,
				TaskStatuses: make(map[string]string),
				StartedAt:    time.Now(),
				Mode:         ModeHumanGuided,
			}
			return m.state, nil
		}
		return nil, fmt.Errorf("failed to read state: %w", err)
	}

	var state ExecutionState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state: %w", err)
	}

	if state.TaskStatuses == nil {
		state.TaskStatuses = make(map[string]string)
	}

	m.state = &state
	return m.state, nil
}

// Save persists the execution state to disk
func (m *StateManager) Save() error {
	if m.state == nil {
		return fmt.Errorf("no state to save")
	}

	m.state.LastUpdated = time.Now()

	// Ensure .fest directory exists
	festDir := filepath.Dir(m.statePath)
	if err := os.MkdirAll(festDir, 0o755); err != nil {
		return fmt.Errorf("failed to create .fest directory: %w", err)
	}

	data, err := yaml.Marshal(m.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	if err := os.WriteFile(m.statePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write state: %w", err)
	}

	return nil
}

// State returns the current state
func (m *StateManager) State() *ExecutionState {
	return m.state
}

// SetTaskStatus updates the status of a task
func (m *StateManager) SetTaskStatus(taskID, status string) {
	if m.state == nil {
		m.state = &ExecutionState{
			TaskStatuses: make(map[string]string),
		}
	}

	oldStatus := m.state.TaskStatuses[taskID]
	m.state.TaskStatuses[taskID] = status

	// Update counts
	m.updateCounts(oldStatus, status)
}

// GetTaskStatus returns the status of a task
func (m *StateManager) GetTaskStatus(taskID string) string {
	if m.state == nil || m.state.TaskStatuses == nil {
		return StatusPending
	}

	status, ok := m.state.TaskStatuses[taskID]
	if !ok {
		return StatusPending
	}
	return status
}

// SetCurrentPosition updates the current execution position
func (m *StateManager) SetCurrentPosition(phase, seq, step int) {
	if m.state == nil {
		m.state = &ExecutionState{
			TaskStatuses: make(map[string]string),
		}
	}

	m.state.CurrentPhase = phase
	m.state.CurrentSeq = seq
	m.state.CurrentStep = step
}

// SetMode updates the execution mode
func (m *StateManager) SetMode(mode ExecutionMode) {
	if m.state == nil {
		m.state = &ExecutionState{
			TaskStatuses: make(map[string]string),
		}
	}
	m.state.Mode = mode
}

// Clear removes the execution state file
func (m *StateManager) Clear() error {
	if err := os.Remove(m.statePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove state file: %w", err)
	}
	m.state = nil
	return nil
}

// Exists checks if state file exists
func (m *StateManager) Exists() bool {
	_, err := os.Stat(m.statePath)
	return err == nil
}

// updateCounts adjusts task counts based on status change
func (m *StateManager) updateCounts(oldStatus, newStatus string) {
	// Decrement old status count
	switch oldStatus {
	case StatusCompleted:
		m.state.CompletedTasks--
	case StatusSkipped:
		m.state.SkippedTasks--
	case StatusFailed:
		m.state.FailedTasks--
	}

	// Increment new status count
	switch newStatus {
	case StatusCompleted:
		m.state.CompletedTasks++
	case StatusSkipped:
		m.state.SkippedTasks++
	case StatusFailed:
		m.state.FailedTasks++
	}
}

// Progress returns execution progress as a percentage
func (s *ExecutionState) Progress() float64 {
	if s.TotalTasks == 0 {
		return 0
	}
	completed := s.CompletedTasks + s.SkippedTasks
	return float64(completed) / float64(s.TotalTasks) * 100
}

// IsComplete returns true if all tasks are done
func (s *ExecutionState) IsComplete() bool {
	if s.TotalTasks == 0 {
		return false
	}
	done := s.CompletedTasks + s.SkippedTasks + s.FailedTasks
	return done >= s.TotalTasks
}

// PendingTasks returns count of pending tasks
func (s *ExecutionState) PendingTasks() int {
	done := s.CompletedTasks + s.SkippedTasks + s.FailedTasks
	pending := s.TotalTasks - done
	if pending < 0 {
		return 0
	}
	return pending
}
