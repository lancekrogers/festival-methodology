package status

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/taskfilter"
)

// collectPhases collects all phases from a festival directory.
func collectPhases(ctx context.Context, festivalPath string) ([]*PhaseInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	var phases []*PhaseInfo
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if !entry.IsDir() {
			continue
		}

		// Check if it's a phase directory (numeric prefix)
		if !hasNumericPrefix(entry.Name()) {
			continue
		}

		phaseDir := filepath.Join(festivalPath, entry.Name())
		phase, err := collectPhaseInfo(ctx, phaseDir, entry.Name())
		if err != nil {
			// Log warning but continue - partial results are acceptable
			continue
		}

		phases = append(phases, phase)
	}

	return phases, nil
}

// collectPhaseInfo collects information about a single phase.
func collectPhaseInfo(ctx context.Context, phasePath, phaseName string) (*PhaseInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	// Calculate phase stats using show package
	festStats, err := show.CalculateFestivalStats(ctx, phasePath)
	var taskStats StatusCounts
	if err == nil && festStats != nil {
		taskStats = StatusCounts{
			Total:      festStats.Tasks.Total,
			Completed:  festStats.Tasks.Completed,
			InProgress: festStats.Tasks.InProgress,
			Pending:    festStats.Tasks.Pending,
			Blocked:    festStats.Tasks.Blocked,
		}
	}

	// Determine phase status
	status := "pending"
	if taskStats.Total > 0 {
		if taskStats.Completed == taskStats.Total {
			status = "completed"
		} else if taskStats.InProgress > 0 || taskStats.Completed > 0 {
			status = "in_progress"
		}
	}

	return &PhaseInfo{
		Name:      phaseName,
		Path:      phasePath,
		Status:    status,
		TaskStats: taskStats,
	}, nil
}

// collectSequencesFromFestival collects all sequences across all phases in a festival.
func collectSequencesFromFestival(ctx context.Context, festivalPath string) ([]*SequenceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	store := progressStoreForFestival(ctx, festivalPath)

	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	var allSequences []*SequenceInfo
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if !entry.IsDir() || !hasNumericPrefix(entry.Name()) {
			continue
		}

		phaseDir := filepath.Join(festivalPath, entry.Name())
		sequences, err := collectSequences(ctx, phaseDir, entry.Name(), store, festivalPath)
		if err != nil {
			// Log warning but continue - partial results are acceptable
			continue
		}

		allSequences = append(allSequences, sequences...)
	}

	return allSequences, nil
}

// collectSequences collects all sequences from a phase directory.
func collectSequences(ctx context.Context, phasePath, phaseName string, store *progress.Store, festivalRoot string) ([]*SequenceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, errors.IO("reading phase directory", err).WithField("path", phasePath)
	}

	var sequences []*SequenceInfo
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if !entry.IsDir() || !hasNumericPrefix(entry.Name()) {
			continue
		}

		seqDir := filepath.Join(phasePath, entry.Name())
		seq, err := collectSequenceInfo(ctx, seqDir, phaseName, entry.Name(), store, festivalRoot)
		if err != nil {
			// Log warning but continue - partial results are acceptable
			continue
		}

		sequences = append(sequences, seq)
	}

	return sequences, nil
}

// collectSequenceInfo collects information about a single sequence.
func collectSequenceInfo(ctx context.Context, seqPath, phaseName, seqName string, store *progress.Store, festivalRoot string) (*SequenceInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	// Count tasks in sequence
	taskStats, err := countSequenceTasks(ctx, seqPath, store, festivalRoot)
	if err != nil {
		taskStats = StatusCounts{}
	}

	// Determine sequence status
	status := "pending"
	if taskStats.Total > 0 {
		if taskStats.Completed == taskStats.Total {
			status = "completed"
		} else if taskStats.InProgress > 0 || taskStats.Completed > 0 {
			status = "in_progress"
		}
	}

	return &SequenceInfo{
		Name:      seqName,
		Path:      seqPath,
		PhaseName: phaseName,
		Status:    status,
		TaskStats: taskStats,
	}, nil
}

// countSequenceTasks counts tasks in a sequence directory.
func countSequenceTasks(ctx context.Context, seqDir string, store *progress.Store, festivalRoot string) (StatusCounts, error) {
	counts := StatusCounts{}

	if err := ctx.Err(); err != nil {
		return counts, errors.Wrap(err, "context cancelled")
	}

	entries, err := os.ReadDir(seqDir)
	if err != nil {
		return counts, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := entry.Name()
		// Use shared taskfilter to determine if this file should be tracked
		// This includes both regular tasks AND quality gates for unified counting
		if !taskfilter.ShouldTrack(name) {
			continue
		}

		counts.Total++
		taskPath := filepath.Join(seqDir, name)
		status := progress.ResolveTaskStatus(store, festivalRoot, taskPath)

		switch status {
		case "completed":
			counts.Completed++
		case "in_progress":
			counts.InProgress++
		case "blocked":
			counts.Blocked++
		default:
			counts.Pending++
		}
	}

	return counts, nil
}

// collectTasksFromFestival collects all tasks across all sequences in a festival.
func collectTasksFromFestival(ctx context.Context, festivalPath string) ([]*TaskInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	store := progressStoreForFestival(ctx, festivalPath)

	phases, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	var allTasks []*TaskInfo
	for _, phaseEntry := range phases {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if !phaseEntry.IsDir() || !hasNumericPrefix(phaseEntry.Name()) {
			continue
		}

		phaseDir := filepath.Join(festivalPath, phaseEntry.Name())
		tasks, err := collectTasksFromPhase(ctx, phaseDir, phaseEntry.Name(), store, festivalPath)
		if err != nil {
			// Log warning but continue - partial results are acceptable
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// collectTasksFromPhase collects all tasks from all sequences in a phase.
func collectTasksFromPhase(ctx context.Context, phasePath, phaseName string, store *progress.Store, festivalRoot string) ([]*TaskInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	sequences, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, errors.IO("reading phase directory", err).WithField("path", phasePath)
	}

	var allTasks []*TaskInfo
	for _, seqEntry := range sequences {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if !seqEntry.IsDir() || !hasNumericPrefix(seqEntry.Name()) {
			continue
		}

		seqDir := filepath.Join(phasePath, seqEntry.Name())
		tasks, err := collectTasks(ctx, seqDir, phaseName, seqEntry.Name(), store, festivalRoot)
		if err != nil {
			// Log warning but continue - partial results are acceptable
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// collectTasks collects all tasks from a sequence directory.
func collectTasks(ctx context.Context, seqPath, phaseName, seqName string, store *progress.Store, festivalRoot string) ([]*TaskInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, errors.IO("reading sequence directory", err).WithField("path", seqPath)
	}

	var tasks []*TaskInfo
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := entry.Name()
		// Use shared taskfilter to determine if this file should be tracked
		// This includes both regular tasks AND quality gates for unified counting
		if !taskfilter.ShouldTrack(name) {
			continue
		}

		taskPath := filepath.Join(seqPath, name)
		status := progress.ResolveTaskStatus(store, festivalRoot, taskPath)

		tasks = append(tasks, &TaskInfo{
			Name:         strings.TrimSuffix(name, ".md"),
			Path:         taskPath,
			PhaseName:    phaseName,
			SequenceName: seqName,
			Status:       status,
		})
	}

	return tasks, nil
}

func progressStoreForFestival(ctx context.Context, festivalPath string) *progress.Store {
	mgr, err := progress.NewManager(ctx, festivalPath)
	if err != nil {
		return nil
	}
	return mgr.Store()
}
