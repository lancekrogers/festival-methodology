package status

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// resolveFestivalFromPath resolves a festival name or path from anywhere in the workspace.
// It searches: cwd (if path is relative), festivals/active/, festivals/planned/,
// festivals/completed/, and festivals/dungeon/.
// Returns the absolute path to the festival directory if found.
func resolveFestivalFromPath(cwd, pathArg string) (string, error) {
	// 1. Check if pathArg is an absolute path
	if filepath.IsAbs(pathArg) {
		if isValidFestivalDir(pathArg) {
			return pathArg, nil
		}
		return "", errors.NotFound("festival").WithField("path", pathArg)
	}

	// 2. Check if pathArg is relative to cwd
	relPath := filepath.Join(cwd, pathArg)
	if isValidFestivalDir(relPath) {
		return relPath, nil
	}

	// 3. Find festivals root and search all status directories
	festivalsRoot := findFestivalsRoot(cwd)
	if festivalsRoot == "" {
		return "", errors.NotFound("festivals directory").
			WithField("hint", "navigate to a workspace with festivals/ directory")
	}

	// Search in all status directories
	statusDirs := []string{"active", "planned", "completed", "dungeon"}
	for _, status := range statusDirs {
		// Try direct path: festivals/<status>/<pathArg>
		candidatePath := filepath.Join(festivalsRoot, status, pathArg)
		if isValidFestivalDir(candidatePath) {
			return candidatePath, nil
		}
	}

	// 4. Check if pathArg includes status prefix (e.g., "active/my-festival")
	candidatePath := filepath.Join(festivalsRoot, pathArg)
	if isValidFestivalDir(candidatePath) {
		return candidatePath, nil
	}

	return "", errors.NotFound("festival").
		WithField("name", pathArg).
		WithField("hint", "festival not found in active, planned, completed, or dungeon")
}

// isValidFestivalDir checks if a directory is a valid festival root.
func isValidFestivalDir(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		return false
	}

	// Check for festival markers: FESTIVAL_GOAL.md, FESTIVAL_OVERVIEW.md, or fest.yaml
	markers := []string{"FESTIVAL_GOAL.md", "FESTIVAL_OVERVIEW.md", "fest.yaml"}
	for _, marker := range markers {
		markerPath := filepath.Join(dir, marker)
		if _, err := os.Stat(markerPath); err == nil {
			return true
		}
	}
	return false
}

// detectEntityType determines what type of entity a path points to.
// Returns EntityFestival, EntityPhase, EntitySequence, or EntityTask.
func detectEntityType(path string) EntityType {
	info, err := os.Stat(path)
	if err != nil {
		return ""
	}

	// If it's a file, it's a task (markdown file)
	if !info.IsDir() {
		return EntityTask
	}

	// Check for festival markers
	if isValidFestivalDir(path) {
		return EntityFestival
	}

	// Check for phase marker (PHASE_GOAL.md)
	if _, err := os.Stat(filepath.Join(path, "PHASE_GOAL.md")); err == nil {
		return EntityPhase
	}

	// Check for sequence marker (SEQUENCE_GOAL.md)
	if _, err := os.Stat(filepath.Join(path, "SEQUENCE_GOAL.md")); err == nil {
		return EntitySequence
	}

	// Default to unknown (could be a regular directory)
	return ""
}

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
