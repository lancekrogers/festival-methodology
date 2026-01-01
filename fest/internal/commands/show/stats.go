package show

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
)

// FestivalInfo holds information about a festival.
type FestivalInfo struct {
	ID         string         `json:"id"`                    // Directory-based ID (e.g., "my-project_GU0001")
	MetadataID string         `json:"metadata_id,omitempty"` // ID from fest.yaml metadata (e.g., "GU0001")
	Name       string         `json:"name"`
	Status     string         `json:"status"`
	Priority   string         `json:"priority,omitempty"`
	Path       string         `json:"path"`
	Stats      *FestivalStats `json:"stats,omitempty"`
}

// FestivalStats holds statistical information about a festival's progress.
type FestivalStats struct {
	Phases    StatusCounts `json:"phases"`
	Sequences StatusCounts `json:"sequences"`
	Tasks     StatusCounts `json:"tasks"`
	Gates     GateCounts   `json:"gates,omitempty"`
	Progress  float64      `json:"progress"` // 0-100 percentage
}

// StatusCounts holds counts by status.
type StatusCounts struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
	Blocked    int `json:"blocked,omitempty"`
}

// GateCounts holds counts specific to quality gates.
type GateCounts struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// CalculateFestivalStats calculates statistics for a festival.
func CalculateFestivalStats(festivalDir string) (*FestivalStats, error) {
	stats := &FestivalStats{}

	// Find and count phases
	entries, err := os.ReadDir(festivalDir)
	if err != nil {
		return nil, errors.IO("reading festival directory", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a phase directory (has PHASE_GOAL.md or numeric prefix)
		phaseDir := filepath.Join(festivalDir, entry.Name())
		if !isPhaseDir(phaseDir) && !hasNumericPrefix(entry.Name()) {
			continue
		}

		stats.Phases.Total++

		// Count sequences within the phase
		phaseStats, err := calculatePhaseStats(phaseDir)
		if err != nil {
			continue
		}

		stats.Sequences.Total += phaseStats.Sequences.Total
		stats.Sequences.Completed += phaseStats.Sequences.Completed
		stats.Sequences.InProgress += phaseStats.Sequences.InProgress
		stats.Sequences.Pending += phaseStats.Sequences.Pending

		stats.Tasks.Total += phaseStats.Tasks.Total
		stats.Tasks.Completed += phaseStats.Tasks.Completed
		stats.Tasks.InProgress += phaseStats.Tasks.InProgress
		stats.Tasks.Pending += phaseStats.Tasks.Pending
		stats.Tasks.Blocked += phaseStats.Tasks.Blocked

		stats.Gates.Total += phaseStats.Gates.Total
		stats.Gates.Passed += phaseStats.Gates.Passed
		stats.Gates.Failed += phaseStats.Gates.Failed

		// Determine phase status based on its tasks
		if phaseStats.Tasks.Total == 0 {
			stats.Phases.Pending++
		} else if phaseStats.Tasks.Completed == phaseStats.Tasks.Total {
			stats.Phases.Completed++
		} else if phaseStats.Tasks.InProgress > 0 || phaseStats.Tasks.Completed > 0 {
			stats.Phases.InProgress++
		} else {
			stats.Phases.Pending++
		}
	}

	// Calculate progress percentage
	if stats.Tasks.Total > 0 {
		stats.Progress = float64(stats.Tasks.Completed) / float64(stats.Tasks.Total) * 100
	}

	return stats, nil
}

func calculatePhaseStats(phaseDir string) (*FestivalStats, error) {
	stats := &FestivalStats{}

	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return stats, nil
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a sequence directory
		seqDir := filepath.Join(phaseDir, entry.Name())
		if !isSequenceDir(seqDir) && !hasNumericPrefix(entry.Name()) {
			continue
		}

		stats.Sequences.Total++

		// Count tasks within the sequence
		seqStats, err := calculateSequenceStats(seqDir)
		if err != nil {
			continue
		}

		stats.Tasks.Total += seqStats.Tasks.Total
		stats.Tasks.Completed += seqStats.Tasks.Completed
		stats.Tasks.InProgress += seqStats.Tasks.InProgress
		stats.Tasks.Pending += seqStats.Tasks.Pending
		stats.Tasks.Blocked += seqStats.Tasks.Blocked

		stats.Gates.Total += seqStats.Gates.Total
		stats.Gates.Passed += seqStats.Gates.Passed
		stats.Gates.Failed += seqStats.Gates.Failed

		// Determine sequence status based on its tasks
		if seqStats.Tasks.Total == 0 {
			stats.Sequences.Pending++
		} else if seqStats.Tasks.Completed == seqStats.Tasks.Total {
			stats.Sequences.Completed++
		} else if seqStats.Tasks.InProgress > 0 || seqStats.Tasks.Completed > 0 {
			stats.Sequences.InProgress++
		} else {
			stats.Sequences.Pending++
		}
	}

	return stats, nil
}

func calculateSequenceStats(seqDir string) (*FestivalStats, error) {
	stats := &FestivalStats{}

	entries, err := os.ReadDir(seqDir)
	if err != nil {
		return stats, nil
	}

	for _, entry := range entries {
		// Skip directories and non-markdown files
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		// Skip goal files
		if name == SequenceGoalFile || name == PhaseGoalFile || name == FestivalGoalFile {
			continue
		}

		// Check if it's a gate file
		isGate := isGateFile(name)

		if isGate {
			stats.Gates.Total++
			// Default to pending for gates
			// TODO: Parse file to determine actual status
		} else {
			stats.Tasks.Total++
			// Parse markdown file for checkbox status
			taskPath := filepath.Join(seqDir, entry.Name())
			status := progress.ParseTaskStatus(taskPath)
			switch status {
			case progress.StatusCompleted:
				stats.Tasks.Completed++
			case progress.StatusInProgress:
				stats.Tasks.InProgress++
			case progress.StatusBlocked:
				stats.Tasks.Blocked++
			default:
				stats.Tasks.Pending++
			}
		}
	}

	return stats, nil
}

func hasNumericPrefix(name string) bool {
	if len(name) < 3 {
		return false
	}
	// Check for numeric prefix followed by underscore (e.g., 001_, 01_, 1_)
	// Look for underscore after 1-3 digits
	digitCount := 0
	for i := 0; i < len(name) && i < 4; i++ {
		if name[i] == '_' && digitCount > 0 {
			return true
		}
		if name[i] >= '0' && name[i] <= '9' {
			digitCount++
		} else if name[i] != '_' {
			return false
		}
	}
	return false
}

func isGateFile(name string) bool {
	lower := strings.ToLower(name)
	return strings.Contains(lower, "_gate") ||
		strings.Contains(lower, "_testing") ||
		strings.Contains(lower, "_review") ||
		strings.Contains(lower, "_verify") ||
		strings.Contains(lower, "_iterate")
}
