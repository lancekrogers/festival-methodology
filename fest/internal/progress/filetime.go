package progress

import (
	"os"
	"sync"
	"time"
)

// GetFileModTime returns the modification time of a file.
// Returns zero time if file doesn't exist or can't be stat'd.
// This is the simple non-cached version for most use cases.
func GetFileModTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

// FileTimeCache provides cached access to file modification times.
// Useful when repeatedly checking the same files during a single operation.
type FileTimeCache struct {
	cache map[string]time.Time
	mu    sync.RWMutex
}

// NewFileTimeCache creates a new FileTimeCache instance.
func NewFileTimeCache() *FileTimeCache {
	return &FileTimeCache{
		cache: make(map[string]time.Time),
	}
}

// GetModTime returns the modification time of a file, using cache if available.
// Returns zero time if file doesn't exist or can't be stat'd.
func (c *FileTimeCache) GetModTime(path string) time.Time {
	// Check cache first
	c.mu.RLock()
	if t, ok := c.cache[path]; ok {
		c.mu.RUnlock()
		return t
	}
	c.mu.RUnlock()

	// Get from filesystem
	t := GetFileModTime(path)

	// Cache the result (including zero time for missing files)
	c.mu.Lock()
	c.cache[path] = t
	c.mu.Unlock()

	return t
}

// Invalidate removes a path from the cache, forcing a re-stat on next access.
func (c *FileTimeCache) Invalidate(path string) {
	c.mu.Lock()
	delete(c.cache, path)
	c.mu.Unlock()
}

// Clear removes all entries from the cache.
func (c *FileTimeCache) Clear() {
	c.mu.Lock()
	c.cache = make(map[string]time.Time)
	c.mu.Unlock()
}

// Size returns the number of cached entries.
func (c *FileTimeCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// InferTaskTime populates time fields from file timestamps
// if no explicit API data exists. Returns true if inference was applied.
//
// This function respects explicit data: if TimeSpentMinutes > 0,
// the task was tracked via API and should not be modified.
//
// For tasks with no time data, it uses file ModTime to infer:
// - CompletedAt: set from file modification time if nil and task is completed
// - TimeSpentMinutes: calculated from StartedAt to CompletedAt span
//
// IMPORTANT: No arbitrary duration caps are applied. If a task
// took 3 days according to timestamps, 3 days is reported.
func InferTaskTime(taskPath string, task *TaskProgress) bool {
	// Skip if explicit time data exists
	if task.TimeSpentMinutes > 0 {
		return false
	}

	modTime := GetFileModTime(taskPath)
	if modTime.IsZero() {
		return false
	}

	applied := false

	// Use ModTime as completion time if task is completed but has no timestamp
	if task.CompletedAt == nil && task.Status == StatusCompleted {
		task.CompletedAt = &modTime
		applied = true
	}

	// If we have a completed timestamp but no started timestamp,
	// use file creation as a rough estimate (fall back to same time)
	if task.StartedAt == nil && task.CompletedAt != nil {
		// Use ModTime as StartedAt too (results in 0 time, but at least timestamps exist)
		// In practice, this means "we don't know when work started"
		startTime := *task.CompletedAt
		task.StartedAt = &startTime
		applied = true
	}

	// Calculate time spent from the span
	if task.StartedAt != nil && task.CompletedAt != nil && task.TimeSpentMinutes == 0 {
		minutes := int(task.CompletedAt.Sub(*task.StartedAt).Minutes())
		if minutes > 0 {
			task.TimeSpentMinutes = minutes
			applied = true
		}
	}

	return applied
}

// NeedsTimeInference returns true if a task lacks time tracking data
// and would benefit from file-based inference.
func NeedsTimeInference(task *TaskProgress) bool {
	if task == nil {
		return false
	}
	// Has explicit time data - no inference needed
	if task.TimeSpentMinutes > 0 {
		return false
	}
	// Only completed tasks can have time inferred from completion timestamp
	if task.Status != StatusCompleted {
		return false
	}
	// Missing completion timestamp - needs inference
	if task.CompletedAt == nil {
		return true
	}
	return false
}
