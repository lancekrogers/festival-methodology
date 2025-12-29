package progress

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
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
	matched, _ := regexp.MatchString(`^\d{2}[\._].*\.md$`, name)
	return matched && !strings.HasPrefix(name, "SEQUENCE")
}

// GetFestivalProgress calculates overall festival progress
func (m *Manager) GetFestivalProgress(festivalPath string) (*FestivalProgress, error) {
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
		matched, _ := regexp.MatchString(`^\d{3}_`, entry.Name())
		if !matched {
			continue
		}

		phasePath := filepath.Join(festivalPath, entry.Name())
		phaseProgress, err := m.GetPhaseProgress(phasePath)
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
func (m *Manager) GetPhaseProgress(phasePath string) (*PhaseProgress, error) {
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
		matched, _ := regexp.MatchString(`^\d{2}_`, entry.Name())
		if !matched {
			continue
		}

		seqPath := filepath.Join(phasePath, entry.Name())
		seqProgress, err := m.GetSequenceProgress(seqPath)
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
func (m *Manager) GetSequenceProgress(seqPath string) (*SequenceProgress, error) {
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

		taskID := entry.Name()
		aggregate.Total++

		// Check if we have progress data for this task
		task, exists := m.store.GetTask(taskID)
		if !exists {
			// No YAML data - parse markdown file for checkbox status
			taskPath := filepath.Join(seqPath, entry.Name())
			status := ParseTaskStatus(taskPath)
			switch status {
			case StatusCompleted:
				aggregate.Completed++
			case StatusInProgress:
				aggregate.InProgress++
			case StatusBlocked:
				aggregate.Blocked++
			default:
				aggregate.Pending++
			}
			continue
		}

		switch task.Status {
		case StatusCompleted:
			aggregate.Completed++
		case StatusInProgress:
			aggregate.InProgress++
		case StatusBlocked:
			aggregate.Blocked++
			aggregate.Blockers = append(aggregate.Blockers, task)
		default:
			aggregate.Pending++
		}

		aggregate.TimeSpentMin += task.TimeSpentMinutes
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
