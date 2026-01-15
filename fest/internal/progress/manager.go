package progress

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// getTaskFileModTime returns the modification time of the task file.
// Falls back to current time if file cannot be stat'd.
func getTaskFileModTime(festivalPath, taskID string) time.Time {
	taskPath := filepath.Join(festivalPath, taskID)
	info, err := os.Stat(taskPath)
	if err != nil {
		return time.Now().UTC()
	}
	return info.ModTime().UTC()
}

// Manager handles progress operations for a festival
type Manager struct {
	store *Store
}

// NewManager creates a new progress manager
func NewManager(ctx context.Context, festivalPath string) (*Manager, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	store := NewStore(festivalPath)
	if err := store.Load(ctx); err != nil {
		return nil, errors.Wrap(err, "loading progress data")
	}
	return &Manager{store: store}, nil
}

// UpdateProgress updates the progress percentage for a task
func (m *Manager) UpdateProgress(ctx context.Context, taskID string, progress int) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if progress < 0 || progress > 100 {
		return errors.Validation("progress must be between 0 and 100").
			WithField("progress", progress)
	}

	task, exists := m.store.GetTask(taskID)
	if !exists {
		task = &TaskProgress{
			TaskID: taskID,
			Status: StatusPending,
		}
	}

	// Start tracking time on first progress update
	// Use file modification time as estimate if task wasn't explicitly started
	if task.StartedAt == nil {
		modTime := getTaskFileModTime(m.store.festivalPath, taskID)
		task.StartedAt = &modTime
	}

	// If progress > 0, mark as in progress
	if progress > 0 && task.Status == StatusPending {
		task.Status = StatusInProgress
	}

	// If progress is 100, mark as completed
	if progress == 100 {
		task.Status = StatusCompleted
		now := time.Now().UTC()
		task.CompletedAt = &now

		// Calculate time spent
		if task.StartedAt != nil {
			task.TimeSpentMinutes = int(now.Sub(*task.StartedAt).Minutes())
		}
	}

	task.Progress = progress

	// Clear blocker if task is progressing
	if progress > 0 && task.BlockerMessage != "" {
		task.BlockerMessage = ""
		task.BlockedAt = nil
	}

	m.store.SetTask(task)
	return m.store.Save(ctx)
}

// MarkComplete marks a task as complete
func (m *Manager) MarkComplete(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	task, exists := m.store.GetTask(taskID)
	if !exists {
		task = &TaskProgress{
			TaskID: taskID,
		}
	}

	now := time.Now().UTC()

	// Set start time if not already set - use file modification time as estimate
	// This provides reasonable time tracking when tasks are completed directly
	// without first being marked "in progress"
	if task.StartedAt == nil {
		modTime := getTaskFileModTime(m.store.festivalPath, taskID)
		task.StartedAt = &modTime
	}

	task.Status = StatusCompleted
	task.Progress = 100
	task.CompletedAt = &now
	task.TimeSpentMinutes = int(now.Sub(*task.StartedAt).Minutes())

	// Clear any blocker
	task.BlockerMessage = ""
	task.BlockedAt = nil

	m.store.SetTask(task)
	return m.store.Save(ctx)
}

// MarkInProgress marks a task as in progress
func (m *Manager) MarkInProgress(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	task, exists := m.store.GetTask(taskID)
	if !exists {
		task = &TaskProgress{
			TaskID: taskID,
		}
	}

	// Set start time if not already set
	// For MarkInProgress, use current time (user explicitly starting work)
	if task.StartedAt == nil {
		now := time.Now().UTC()
		task.StartedAt = &now
	}

	task.Status = StatusInProgress

	m.store.SetTask(task)
	return m.store.Save(ctx)
}

// ReportBlocker reports a blocker for a task
func (m *Manager) ReportBlocker(ctx context.Context, taskID, message string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if message == "" {
		return errors.Validation("blocker message required")
	}

	task, exists := m.store.GetTask(taskID)
	if !exists {
		task = &TaskProgress{
			TaskID: taskID,
			Status: StatusPending,
		}
	}

	now := time.Now().UTC()
	task.Status = StatusBlocked
	task.BlockerMessage = message
	task.BlockedAt = &now

	// Start tracking time if not already
	if task.StartedAt == nil {
		task.StartedAt = &now
	}

	m.store.SetTask(task)
	return m.store.Save(ctx)
}

// ClearBlocker clears a blocker for a task
func (m *Manager) ClearBlocker(ctx context.Context, taskID string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	task, exists := m.store.GetTask(taskID)
	if !exists {
		return errors.NotFound("task").WithField("taskID", taskID)
	}

	if task.BlockerMessage == "" {
		return nil // No blocker to clear
	}

	task.BlockerMessage = ""
	task.BlockedAt = nil

	// Return to in_progress if was blocked
	if task.Status == StatusBlocked {
		task.Status = StatusInProgress
	}

	m.store.SetTask(task)
	return m.store.Save(ctx)
}

// GetTaskProgress retrieves progress for a specific task
func (m *Manager) GetTaskProgress(taskID string) (*TaskProgress, bool) {
	return m.store.GetTask(taskID)
}

// AllTaskProgress returns all task progress entries
func (m *Manager) AllTaskProgress() map[string]*TaskProgress {
	return m.store.AllTasks()
}

// Store returns the underlying store for advanced operations
func (m *Manager) Store() *Store {
	return m.store
}
