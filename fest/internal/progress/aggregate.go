package progress

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

var (
	// phasePattern matches phase directory names (e.g., "001_Phase_Name")
	phasePattern = regexp.MustCompile(`^\d{3}_`)
	// sequencePattern matches sequence directory names (e.g., "01_Sequence_Name")
	sequencePattern = regexp.MustCompile(`^\d{2}_`)
	// taskPattern matches task file names (e.g., "01_task.md" or "01.5_task.md")
	taskPattern = regexp.MustCompile(`^\d{2}[\._].*\.md$`)
)

// AggregateProgress holds aggregated progress stats
type AggregateProgress struct {
	Total        int             `json:"total"`
	Completed    int             `json:"completed"`
	InProgress   int             `json:"in_progress"`
	Blocked      int             `json:"blocked"`
	Pending      int             `json:"pending"`
	Percentage   int             `json:"percentage"`
	Blockers     []*TaskProgress `json:"blockers,omitempty"`
	TimeSpentMin int             `json:"time_spent_minutes"`
}

// PhaseProgress holds progress for a phase
type PhaseProgress struct {
	PhaseID   string             `json:"phase_id"`
	PhaseName string             `json:"phase_name"`
	Progress  *AggregateProgress `json:"progress"`
}

// SequenceProgress holds progress for a sequence
type SequenceProgress struct {
	SequenceID   string             `json:"sequence_id"`
	SequenceName string             `json:"sequence_name"`
	Progress     *AggregateProgress `json:"progress"`
}

// FestivalProgress holds complete festival progress
type FestivalProgress struct {
	FestivalName string             `json:"festival_name"`
	Overall      *AggregateProgress `json:"overall"`
	Phases       []*PhaseProgress   `json:"phases,omitempty"`
}

// isTask checks if a filename looks like a task file
func isTask(name string) bool {
	// Task files match pattern: NN_name.md or NN.N_name.md
	return taskPattern.MatchString(name) && !strings.HasPrefix(name, "SEQUENCE")
}

// GetFestivalProgress calculates overall festival progress
func (m *Manager) GetFestivalProgress(ctx context.Context, festivalPath string) (*FestivalProgress, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	festivalName := filepath.Base(festivalPath)
	overall := &AggregateProgress{}
	var phases []*PhaseProgress

	// Walk the festival directory
	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, err
	}

	// Find and process phases
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a phase directory (starts with NNN_)
		if !phasePattern.MatchString(entry.Name()) {
			continue
		}

		phasePath := filepath.Join(festivalPath, entry.Name())
		phaseProgress, err := m.GetPhaseProgress(ctx, phasePath)
		if err != nil {
			continue
		}

		phases = append(phases, phaseProgress)

		// Aggregate to overall
		overall.Total += phaseProgress.Progress.Total
		overall.Completed += phaseProgress.Progress.Completed
		overall.InProgress += phaseProgress.Progress.InProgress
		overall.Blocked += phaseProgress.Progress.Blocked
		overall.Pending += phaseProgress.Progress.Pending
		overall.TimeSpentMin += phaseProgress.Progress.TimeSpentMin
		overall.Blockers = append(overall.Blockers, phaseProgress.Progress.Blockers...)
	}

	// Calculate percentage
	if overall.Total > 0 {
		overall.Percentage = (overall.Completed * 100) / overall.Total
	}

	return &FestivalProgress{
		FestivalName: festivalName,
		Overall:      overall,
		Phases:       phases,
	}, nil
}

// GetPhaseProgress calculates progress for a phase
func (m *Manager) GetPhaseProgress(ctx context.Context, phasePath string) (*PhaseProgress, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	phaseName := filepath.Base(phasePath)
	aggregate := &AggregateProgress{}

	// Walk the phase directory
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, err
	}

	// Find and process sequences
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a sequence directory (starts with NN_)
		if !sequencePattern.MatchString(entry.Name()) {
			continue
		}

		seqPath := filepath.Join(phasePath, entry.Name())
		seqProgress, err := m.GetSequenceProgress(ctx, seqPath)
		if err != nil {
			continue
		}

		// Aggregate
		aggregate.Total += seqProgress.Progress.Total
		aggregate.Completed += seqProgress.Progress.Completed
		aggregate.InProgress += seqProgress.Progress.InProgress
		aggregate.Blocked += seqProgress.Progress.Blocked
		aggregate.Pending += seqProgress.Progress.Pending
		aggregate.TimeSpentMin += seqProgress.Progress.TimeSpentMin
		aggregate.Blockers = append(aggregate.Blockers, seqProgress.Progress.Blockers...)
	}

	// Calculate percentage
	if aggregate.Total > 0 {
		aggregate.Percentage = (aggregate.Completed * 100) / aggregate.Total
	}

	return &PhaseProgress{
		PhaseID:   phaseName,
		PhaseName: phaseName,
		Progress:  aggregate,
	}, nil
}

// GetSequenceProgress calculates progress for a sequence
func (m *Manager) GetSequenceProgress(ctx context.Context, seqPath string) (*SequenceProgress, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	seqName := filepath.Base(seqPath)
	aggregate := &AggregateProgress{}

	// Walk the sequence directory
	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, err
	}

	// Find and count tasks
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !isTask(entry.Name()) {
			continue
		}

		aggregate.Total++
		taskPath := filepath.Join(seqPath, entry.Name())

		// Use ResolveTaskStatus which prioritizes markdown checkboxes as source of truth
		status := ResolveTaskStatus(m.store, m.store.festivalPath, taskPath)

		switch status {
		case StatusCompleted:
			aggregate.Completed++
		case StatusInProgress:
			aggregate.InProgress++
		case StatusBlocked:
			aggregate.Blocked++
			// Get task from YAML for blocker details if available
			if task, exists := ResolveTaskProgress(m.store, m.store.festivalPath, taskPath); exists && task.Status == StatusBlocked {
				aggregate.Blockers = append(aggregate.Blockers, task)
			}
		default:
			aggregate.Pending++
		}

		// Get time tracking from YAML if available (markdown doesn't track time)
		if task, exists := ResolveTaskProgress(m.store, m.store.festivalPath, taskPath); exists {
			aggregate.TimeSpentMin += task.TimeSpentMinutes
		}
	}

	// Calculate percentage
	if aggregate.Total > 0 {
		aggregate.Percentage = (aggregate.Completed * 100) / aggregate.Total
	}

	return &SequenceProgress{
		SequenceID:   seqName,
		SequenceName: seqName,
		Progress:     aggregate,
	}, nil
}
