// Package progress provides progress tracking for festival execution.
package progress

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	// ProgressFileName is the name of the progress file
	ProgressFileName = "progress.yaml"

	// ProgressDir is the directory within festival for progress data
	ProgressDir = ".fest"

	// Status values
	StatusPending    = "pending"
	StatusInProgress = "in_progress"
	StatusBlocked    = "blocked"
	StatusCompleted  = "completed"
)

// TaskProgress represents the progress state of a single task
type TaskProgress struct {
	TaskID           string     `yaml:"task_id"`
	Status           string     `yaml:"status"`
	Progress         int        `yaml:"progress"`
	StartedAt        *time.Time `yaml:"started_at,omitempty"`
	CompletedAt      *time.Time `yaml:"completed_at,omitempty"`
	TimeSpentMinutes int        `yaml:"time_spent_minutes,omitempty"`
	BlockerMessage   string     `yaml:"blocker_message,omitempty"`
	BlockedAt        *time.Time `yaml:"blocked_at,omitempty"`
}

// FestivalTimeMetrics tracks festival-level time metrics separate from task-level tracking.
// This enables displaying both "how long agents worked" (TotalWorkMinutes) and
// "how long the festival has existed" (lifecycle from CreatedAt to CompletedAt).
//
// Note: This is distinct from FestivalProgress in aggregate.go which tracks
// task counts and completion status. FestivalTimeMetrics focuses on time data.
//
// Example usage:
//
//	ftm := &FestivalTimeMetrics{
//	    CreatedAt:        time.Now(),
//	    TotalWorkMinutes: 0,
//	}
//	// Later, when festival completes:
//	now := time.Now()
//	ftm.CompletedAt = &now
//	ftm.LifecycleDuration = int(now.Sub(ftm.CreatedAt).Hours() / 24)
type FestivalTimeMetrics struct {
	// CreatedAt is when the festival was created (from fest.yaml or directory creation time)
	CreatedAt time.Time `yaml:"created_at"`
	// CompletedAt is when the last task was completed (nil if festival is ongoing)
	CompletedAt *time.Time `yaml:"completed_at,omitempty"`
	// LifecycleDuration is the number of days from creation to completion (0 if ongoing)
	LifecycleDuration int `yaml:"lifecycle_duration_days,omitempty"`
	// TotalWorkMinutes is the sum of all task TimeSpentMinutes (agent work time)
	TotalWorkMinutes int `yaml:"total_work_minutes"`
}

// FestivalProgressData is the persisted progress state for a festival
type FestivalProgressData struct {
	Festival    string                   `yaml:"festival"`
	UpdatedAt   time.Time                `yaml:"updated_at"`
	TimeMetrics *FestivalTimeMetrics     `yaml:"time_metrics,omitempty"`
	Tasks       map[string]*TaskProgress `yaml:"tasks"`
}

// Store manages progress persistence
type Store struct {
	festivalPath string
	data         *FestivalProgressData
}

// NewStore creates a new progress store for a festival
func NewStore(festivalPath string) *Store {
	return &Store{
		festivalPath: festivalPath,
	}
}

// progressFilePath returns the path to the progress file
func (s *Store) progressFilePath() string {
	return filepath.Join(s.festivalPath, ProgressDir, ProgressFileName)
}

// Load loads progress data from disk
func (s *Store) Load(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	progressPath := s.progressFilePath()

	// Check if file exists
	if _, err := os.Stat(progressPath); os.IsNotExist(err) {
		// Create empty progress data with initialized time metrics
		now := time.Now().UTC()
		s.data = &FestivalProgressData{
			Festival:  filepath.Base(s.festivalPath),
			UpdatedAt: now,
			TimeMetrics: &FestivalTimeMetrics{
				CreatedAt: now,
			},
			Tasks: make(map[string]*TaskProgress),
		}
		return nil
	}

	data, err := os.ReadFile(progressPath)
	if err != nil {
		return errors.IO("reading progress file", err).WithField("path", progressPath)
	}

	var progressData FestivalProgressData
	if err := yaml.Unmarshal(data, &progressData); err != nil {
		return errors.Parse("parsing progress file", err).WithField("path", progressPath)
	}

	// Initialize maps if nil
	if progressData.Tasks == nil {
		progressData.Tasks = make(map[string]*TaskProgress)
	}

	// Initialize TimeMetrics for legacy files that don't have it
	if progressData.TimeMetrics == nil {
		progressData.TimeMetrics = &FestivalTimeMetrics{
			CreatedAt: progressData.UpdatedAt, // Use first known timestamp as fallback
		}
	}

	s.data = &progressData
	return nil
}

// Save writes progress data to disk
func (s *Store) Save(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if s.data == nil {
		return errors.Validation("no progress data to save")
	}

	s.data.UpdatedAt = time.Now().UTC()

	data, err := yaml.Marshal(s.data)
	if err != nil {
		return errors.Wrap(err, "marshaling progress data")
	}

	progressPath := s.progressFilePath()

	// Ensure directory exists
	progressDir := filepath.Dir(progressPath)
	if err := os.MkdirAll(progressDir, 0755); err != nil {
		return errors.IO("creating progress directory", err).WithField("path", progressDir)
	}

	if err := os.WriteFile(progressPath, data, 0644); err != nil {
		return errors.IO("writing progress file", err).WithField("path", progressPath)
	}

	return nil
}

// Data returns the current progress data
func (s *Store) Data() *FestivalProgressData {
	return s.data
}

// GetTask retrieves progress for a specific task
func (s *Store) GetTask(taskID string) (*TaskProgress, bool) {
	if s.data == nil || s.data.Tasks == nil {
		return nil, false
	}
	task, ok := s.data.Tasks[taskID]
	return task, ok
}

// SetTask updates or creates progress for a task
func (s *Store) SetTask(task *TaskProgress) {
	if s.data == nil {
		s.data = &FestivalProgressData{
			Festival:  filepath.Base(s.festivalPath),
			UpdatedAt: time.Now().UTC(),
			Tasks:     make(map[string]*TaskProgress),
		}
	}
	s.data.Tasks[task.TaskID] = task
}

// AllTasks returns all task progress entries
func (s *Store) AllTasks() map[string]*TaskProgress {
	if s.data == nil {
		return nil
	}
	return s.data.Tasks
}

// GetTimeMetrics returns the festival time metrics (may be nil for legacy data)
func (s *Store) GetTimeMetrics() *FestivalTimeMetrics {
	if s.data == nil {
		return nil
	}
	return s.data.TimeMetrics
}

// SetTimeMetrics sets the festival time metrics
func (s *Store) SetTimeMetrics(metrics *FestivalTimeMetrics) {
	if s.data == nil {
		s.data = &FestivalProgressData{
			Festival:  filepath.Base(s.festivalPath),
			UpdatedAt: time.Now().UTC(),
			Tasks:     make(map[string]*TaskProgress),
		}
	}
	s.data.TimeMetrics = metrics
}

// EnsureTimeMetrics ensures TimeMetrics is initialized, creating it if nil
func (s *Store) EnsureTimeMetrics() *FestivalTimeMetrics {
	if s.data == nil {
		s.data = &FestivalProgressData{
			Festival:  filepath.Base(s.festivalPath),
			UpdatedAt: time.Now().UTC(),
			Tasks:     make(map[string]*TaskProgress),
		}
	}
	if s.data.TimeMetrics == nil {
		s.data.TimeMetrics = &FestivalTimeMetrics{
			CreatedAt: time.Now().UTC(),
		}
	}
	return s.data.TimeMetrics
}

// MarkFestivalCompleted sets the completion timestamp and calculates lifecycle duration
func (s *Store) MarkFestivalCompleted() {
	metrics := s.EnsureTimeMetrics()
	now := time.Now().UTC()
	metrics.CompletedAt = &now
	// Calculate lifecycle duration in days
	if !metrics.CreatedAt.IsZero() {
		metrics.LifecycleDuration = int(now.Sub(metrics.CreatedAt).Hours() / 24)
	}
}

// UpdateTotalWorkMinutes recalculates TotalWorkMinutes from all tasks
func (s *Store) UpdateTotalWorkMinutes() {
	if s.data == nil {
		return
	}
	total := 0
	for _, task := range s.data.Tasks {
		total += task.TimeSpentMinutes
	}
	metrics := s.EnsureTimeMetrics()
	metrics.TotalWorkMinutes = total
}

// LazyPopulateTimeData infers time data for completed tasks that don't have it.
// This provides a seamless experience for legacy festivals - time data appears
// automatically on first access without requiring users to run migration.
func (s *Store) LazyPopulateTimeData(festivalPath string) bool {
	if s.data == nil || len(s.data.Tasks) == 0 {
		return false
	}

	modified := false
	for _, task := range s.data.Tasks {
		if task.Status != StatusCompleted {
			continue
		}

		// Skip tasks that already have time data
		if task.TimeSpentMinutes > 0 {
			continue
		}

		// Build full task path and infer time
		taskPath := filepath.Join(festivalPath, task.TaskID)
		if InferTaskTime(taskPath, task) {
			modified = true
		}
	}

	if modified {
		s.UpdateTotalWorkMinutes()
	}

	return modified
}

// IsFestivalComplete returns true if all tracked tasks are completed
func (s *Store) IsFestivalComplete() bool {
	if s.data == nil || len(s.data.Tasks) == 0 {
		return false // Empty festivals are not complete
	}
	for _, task := range s.data.Tasks {
		if task.Status != StatusCompleted {
			return false
		}
	}
	return true
}

// CheckAndSetCompletion checks if festival is complete and marks it if so
// This is idempotent - calling multiple times will not reset CompletedAt
func (s *Store) CheckAndSetCompletion() bool {
	if !s.IsFestivalComplete() {
		return false
	}
	metrics := s.GetTimeMetrics()
	if metrics != nil && metrics.CompletedAt != nil {
		return false // Already marked complete
	}
	s.MarkFestivalCompleted()
	s.UpdateTotalWorkMinutes()
	return true
}

// CalculateLifecycleDuration computes days from creation to completion.
// Returns -1 for ongoing festivals (not yet completed).
// No arbitrary caps - if festival took 365 days, returns 365.
func (m *FestivalTimeMetrics) CalculateLifecycleDuration() int {
	if m == nil || m.CompletedAt == nil {
		return -1 // ongoing
	}

	duration := m.CompletedAt.Sub(m.CreatedAt)
	days := int(duration.Hours() / 24)

	return days
}

// GetLifecycleDuration returns the lifecycle duration, recalculating if needed
func (m *FestivalTimeMetrics) GetLifecycleDuration() int {
	if m == nil {
		return -1
	}
	if m.CompletedAt != nil {
		m.LifecycleDuration = m.CalculateLifecycleDuration()
	}
	return m.LifecycleDuration
}

// FormatLifecycleDuration formats lifecycle duration for display
func FormatLifecycleDuration(days int) string {
	if days < 0 {
		return "ongoing"
	}
	if days == 0 {
		return "< 1 day"
	}
	if days == 1 {
		return "1 day"
	}
	return fmt.Sprintf("%d days", days)
}

// GetCurrentDuration calculates days since festival creation (for ongoing festivals)
func (m *FestivalTimeMetrics) GetCurrentDuration() int {
	if m == nil {
		return 0
	}
	duration := time.Since(m.CreatedAt)
	return int(duration.Hours() / 24)
}

// FormatDurationWithStatus formats duration with status indicator for ongoing festivals
// For completed festivals, shows "X days"
// For ongoing festivals, shows "X days (ongoing)" or "< 1 day (ongoing)"
func FormatDurationWithStatus(metrics *FestivalTimeMetrics) string {
	if metrics == nil {
		return "unknown"
	}

	// Completed festival - use stored lifecycle duration
	if metrics.CompletedAt != nil {
		days := metrics.GetLifecycleDuration()
		return FormatLifecycleDuration(days)
	}

	// Ongoing festival - calculate current duration
	days := metrics.GetCurrentDuration()
	if days == 0 {
		return "< 1 day (ongoing)"
	}
	if days == 1 {
		return "1 day (ongoing)"
	}
	return fmt.Sprintf("%d days (ongoing)", days)
}
