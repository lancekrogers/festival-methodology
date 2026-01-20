package progress

import (
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ResolveTaskTime returns the time spent on a task.
// Uses explicit API data if available, otherwise infers from file timestamps.
// This enables time tracking for tasks completed via markdown edits.
func ResolveTaskTime(store *Store, festivalPath, taskPath string) int {
	task, ok := ResolveTaskProgress(store, festivalPath, taskPath)
	if !ok {
		return 0
	}

	// Explicit time data takes precedence
	if task.TimeSpentMinutes > 0 {
		return task.TimeSpentMinutes
	}

	// Infer time from file timestamps if needed
	fullPath := resolveFullPath(festivalPath, taskPath)
	if InferTaskTime(fullPath, task) {
		return task.TimeSpentMinutes
	}

	return 0
}

// resolveFullPath returns the absolute path for a task.
func resolveFullPath(festivalPath, taskPath string) string {
	if filepath.IsAbs(taskPath) {
		return taskPath
	}
	return filepath.Join(festivalPath, taskPath)
}

// NormalizeTaskID normalizes a user-provided task ID to a canonical key.
// If the ID is a path, it is converted to a festival-relative path.
// If the ID is a bare filename, it is returned as-is for legacy support.
func NormalizeTaskID(festivalPath, taskID string) (string, error) {
	if taskID == "" {
		return "", errors.Validation("task ID required")
	}

	if filepath.IsAbs(taskID) {
		return taskKeyFromPath(festivalPath, taskID)
	}

	if strings.Contains(taskID, "/") || strings.Contains(taskID, "\\") {
		absPath := filepath.Join(festivalPath, taskID)
		return taskKeyFromPath(festivalPath, absPath)
	}

	return taskID, nil
}

// ResolveTaskProgress finds progress data for a task path, preferring
// festival-relative keys and falling back to legacy filename keys.
// Time inference is applied lazily if the task has no explicit time data.
func ResolveTaskProgress(store *Store, festivalPath, taskPath string) (*TaskProgress, bool) {
	if store == nil || taskPath == "" {
		return nil, false
	}

	var task *TaskProgress
	var ok bool

	if key, err := taskKeyFromPath(festivalPath, taskPath); err == nil {
		task, ok = store.GetTask(key)
	}
	if !ok {
		task, ok = store.GetTask(filepath.Base(taskPath))
	}

	if ok && task != nil {
		// Apply time inference lazily if task needs it
		fullPath := resolveFullPath(festivalPath, taskPath)
		InferTaskTime(fullPath, task)
	}

	return task, ok
}

// ResolveTaskStatus returns the status for a task path, with YAML progress store
// as the primary source of truth. Markdown checkboxes are only used as a fallback
// when no YAML record exists.
func ResolveTaskStatus(store *Store, festivalPath, taskPath string) string {
	// Check YAML first - this is the source of truth
	if task, ok := ResolveTaskProgress(store, festivalPath, taskPath); ok {
		return task.Status
	}

	// Fall back to markdown if no YAML record exists
	return ParseTaskStatus(taskPath)
}

func taskKeyFromPath(festivalPath, taskPath string) (string, error) {
	rel, err := filepath.Rel(festivalPath, taskPath)
	if err != nil {
		return "", errors.IO("resolving task path", err).
			WithField("festival_path", festivalPath).
			WithField("task_path", taskPath)
	}

	if rel == "." || strings.HasPrefix(rel, "..") {
		return "", errors.Validation("task path is outside festival").
			WithField("festival_path", festivalPath).
			WithField("task_path", taskPath)
	}

	return filepath.ToSlash(rel), nil
}
