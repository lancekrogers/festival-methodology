package progress

import (
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

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
func ResolveTaskProgress(store *Store, festivalPath, taskPath string) (*TaskProgress, bool) {
	if store == nil || taskPath == "" {
		return nil, false
	}

	if key, err := taskKeyFromPath(festivalPath, taskPath); err == nil {
		if task, ok := store.GetTask(key); ok {
			return task, true
		}
	}

	return store.GetTask(filepath.Base(taskPath))
}

// ResolveTaskStatus returns the status for a task path, using progress data
// when available and falling back to markdown checkbox parsing.
func ResolveTaskStatus(store *Store, festivalPath, taskPath string) string {
	if task, ok := ResolveTaskProgress(store, festivalPath, taskPath); ok {
		if task.Status != "" {
			return task.Status
		}
	}

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
