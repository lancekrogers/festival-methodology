// Package execute provides execution mode management for festival automation.
package execute

import (
	"time"
)

// ExecutionMode defines how a festival is executed
type ExecutionMode string

const (
	// ModeHumanGuided is the default mode where humans navigate and invoke agents per task
	ModeHumanGuided ExecutionMode = "human_guided"

	// ModeSemiAutonomous allows agents to loop through tasks with automatic handoffs
	ModeSemiAutonomous ExecutionMode = "semi_autonomous"

	// ModeOrchestrated enables multi-agent coordination with task claiming (post-v1)
	ModeOrchestrated ExecutionMode = "orchestrated"
)

// AutonomyLevel represents how autonomous a task can be
type AutonomyLevel string

const (
	AutonomyHigh   AutonomyLevel = "high"
	AutonomyMedium AutonomyLevel = "medium"
	AutonomyLow    AutonomyLevel = "low"
)

// ExecutionConfig holds configuration for execution
type ExecutionConfig struct {
	Mode          ExecutionMode `json:"mode"`
	AutoDemote    bool          `json:"auto_demote"`     // Auto-switch to human-guided on low autonomy
	ReviewAtGates bool          `json:"review_at_gates"` // Pause at quality gates for review
	ClaimTimeout  time.Duration `json:"claim_timeout"`   // For orchestrated mode
	MaxTasks      int           `json:"max_tasks"`       // Maximum tasks to execute in a loop (0 = unlimited)
	DryRun        bool          `json:"dry_run"`         // Preview without executing
}

// DefaultConfig returns the default execution configuration
func DefaultConfig() *ExecutionConfig {
	return &ExecutionConfig{
		Mode:          ModeHumanGuided,
		AutoDemote:    true,
		ReviewAtGates: true,
		ClaimTimeout:  5 * time.Minute,
		MaxTasks:      0,
		DryRun:        false,
	}
}

// TaskInfo holds information about a task for execution
type TaskInfo struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Path          string        `json:"path"`
	AutonomyLevel AutonomyLevel `json:"autonomy_level"`
	Dependencies  []string      `json:"dependencies,omitempty"`
	Status        string        `json:"status"`
	IsGate        bool          `json:"is_gate"`
}

// HandoffReason explains why execution mode changed
type HandoffReason string

const (
	HandoffLowAutonomy   HandoffReason = "low_autonomy"
	HandoffBlocker       HandoffReason = "blocker"
	HandoffError         HandoffReason = "error"
	HandoffGateReview    HandoffReason = "gate_review"
	HandoffUserRequested HandoffReason = "user_requested"
	HandoffComplete      HandoffReason = "complete"
)

// HandoffEvent describes a mode transition
type HandoffEvent struct {
	FromMode     ExecutionMode `json:"from_mode"`
	ToMode       ExecutionMode `json:"to_mode"`
	Reason       HandoffReason `json:"reason"`
	TaskID       string        `json:"task_id,omitempty"`
	Message      string        `json:"message"`
	RequiresUser bool          `json:"requires_user"`
}

// ExecutionResult holds the result of task execution
type ExecutionResult struct {
	TaskID       string        `json:"task_id"`
	Success      bool          `json:"success"`
	Message      string        `json:"message"`
	TimeSpentSec int           `json:"time_spent_sec"`
	Handoff      *HandoffEvent `json:"handoff,omitempty"`
}

// LoopState tracks the state of semi-autonomous execution
type LoopState struct {
	Mode           ExecutionMode    `json:"mode"`
	TasksCompleted int              `json:"tasks_completed"`
	TasksRemaining int              `json:"tasks_remaining"`
	CurrentTask    *TaskInfo        `json:"current_task,omitempty"`
	LastResult     *ExecutionResult `json:"last_result,omitempty"`
	Handoffs       []*HandoffEvent  `json:"handoffs,omitempty"`
	StartedAt      time.Time        `json:"started_at"`
	Config         *ExecutionConfig `json:"config"`
}

// ShouldContinue determines if the execution loop should continue
func (s *LoopState) ShouldContinue() bool {
	if s.Mode != ModeSemiAutonomous {
		return false
	}

	if s.TasksRemaining == 0 {
		return false
	}

	if s.Config.MaxTasks > 0 && s.TasksCompleted >= s.Config.MaxTasks {
		return false
	}

	// Check if last result triggered a handoff
	if s.LastResult != nil && s.LastResult.Handoff != nil && s.LastResult.Handoff.RequiresUser {
		return false
	}

	return true
}

// ShouldDemote checks if we should demote to human-guided mode
func ShouldDemote(task *TaskInfo, config *ExecutionConfig) bool {
	if !config.AutoDemote {
		return false
	}

	// Demote for low autonomy tasks
	if task.AutonomyLevel == AutonomyLow {
		return true
	}

	// Optionally review at gates
	if task.IsGate && config.ReviewAtGates {
		return true
	}

	return false
}

// DetectBestMode analyzes tasks and recommends an execution mode
func DetectBestMode(tasks []*TaskInfo) ExecutionMode {
	if len(tasks) == 0 {
		return ModeHumanGuided
	}

	lowAutonomyCount := 0
	for _, task := range tasks {
		if task.AutonomyLevel == AutonomyLow {
			lowAutonomyCount++
		}
	}

	// If more than 30% low autonomy, use human-guided
	lowPercent := float64(lowAutonomyCount) / float64(len(tasks)) * 100
	if lowPercent > 30 {
		return ModeHumanGuided
	}

	return ModeSemiAutonomous
}

// CreateHandoff creates a handoff event for mode transition
func CreateHandoff(from, to ExecutionMode, reason HandoffReason, taskID, message string) *HandoffEvent {
	requiresUser := false
	switch reason {
	case HandoffLowAutonomy, HandoffBlocker, HandoffError, HandoffGateReview:
		requiresUser = true
	}

	return &HandoffEvent{
		FromMode:     from,
		ToMode:       to,
		Reason:       reason,
		TaskID:       taskID,
		Message:      message,
		RequiresUser: requiresUser,
	}
}
