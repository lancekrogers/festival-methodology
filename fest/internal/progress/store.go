// Package progress provides progress tracking for festival execution.
package progress

import (
	"context"
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

// FestivalProgressData is the persisted progress state for a festival
type FestivalProgressData struct {
	Festival  string                   `yaml:"festival"`
	UpdatedAt time.Time                `yaml:"updated_at"`
	Tasks     map[string]*TaskProgress `yaml:"tasks"`
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
		// Create empty progress data
		s.data = &FestivalProgressData{
			Festival:  filepath.Base(s.festivalPath),
			UpdatedAt: time.Now().UTC(),
			Tasks:     make(map[string]*TaskProgress),
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
