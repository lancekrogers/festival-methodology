package status

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// resolveStatusPath resolves the target path for status commands.
// If pathArg is empty, uses current working directory.
// If pathArg is relative to a festivals/ root (e.g., "active/my-festival"),
// it resolves from the festivals root.
func resolveStatusPath(pathArg string) (string, error) {
	if pathArg == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.IO("getting current directory", err)
		}
		return cwd, nil
	}

	// Try as absolute or relative path first
	absPath, err := filepath.Abs(pathArg)
	if err != nil {
		return "", errors.Wrap(err, "resolving path").WithField("path", pathArg)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err == nil {
		return absPath, nil
	}

	// Try resolving relative to festivals/ root
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting current directory", err)
	}

	// Find festivals root and try pathArg relative to it
	festivalsRoot := findFestivalsRoot(cwd)
	if festivalsRoot != "" {
		candidatePath := filepath.Join(festivalsRoot, pathArg)
		if _, err := os.Stat(candidatePath); err == nil {
			return candidatePath, nil
		}
	}

	return "", errors.NotFound("path").WithField("path", pathArg)
}

// findFestivalsRoot walks up from startPath looking for a festivals/ directory.
func findFestivalsRoot(startPath string) string {
	current := startPath
	for {
		// Check if current is festivals/ or contains festivals/
		if filepath.Base(current) == "festivals" {
			return current
		}
		festivalsDir := filepath.Join(current, "festivals")
		if info, err := os.Stat(festivalsDir); err == nil && info.IsDir() {
			return festivalsDir
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

// isValidStatus checks if status is valid for the given entity type.
func isValidStatus(entityType EntityType, status string) bool {
	validStatuses, ok := ValidStatuses[entityType]
	if !ok {
		return false
	}
	for _, valid := range validStatuses {
		if valid == status {
			return true
		}
	}
	return false
}

// hasNumericPrefix checks if a directory name starts with digits.
func hasNumericPrefix(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// filterPhasesByStatus filters phases to only those matching the given status.
// If status is empty, returns all phases.
func filterPhasesByStatus(phases []*PhaseInfo, status string) []*PhaseInfo {
	if status == "" {
		return phases
	}

	var filtered []*PhaseInfo
	for _, phase := range phases {
		if phase.Status == status {
			filtered = append(filtered, phase)
		}
	}
	return filtered
}

// filterSequencesByStatus filters sequences to only those matching the given status.
// If status is empty, returns all sequences.
func filterSequencesByStatus(sequences []*SequenceInfo, status string) []*SequenceInfo {
	if status == "" {
		return sequences
	}

	var filtered []*SequenceInfo
	for _, seq := range sequences {
		if seq.Status == status {
			filtered = append(filtered, seq)
		}
	}
	return filtered
}

// filterTasksByStatus filters tasks to only those matching the given status.
// If status is empty, returns all tasks.
func filterTasksByStatus(tasks []*TaskInfo, status string) []*TaskInfo {
	if status == "" {
		return tasks
	}

	var filtered []*TaskInfo
	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}
