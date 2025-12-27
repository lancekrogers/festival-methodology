// Package festival provides unified resolution for festival elements.
package festival

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Context holds the current context within a festival hierarchy.
type Context struct {
	FestivalsRoot string // Path to festivals/ directory
	FestivalDir   string // Path to current festival (has FESTIVAL_GOAL.md)
	PhaseDir      string // Path to current phase (NNN_* directory)
	SequenceDir   string // Path to current sequence (NN_* directory)
}

var (
	phasePattern    = regexp.MustCompile(`^\d{3}_`)
	sequencePattern = regexp.MustCompile(`^\d{2}_`)
)

// DetectContext determines the current context from the working directory.
// It walks up the directory tree to find festival markers and identifies
// the current phase/sequence based on directory naming patterns.
func DetectContext(cwd string) (*Context, error) {
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return nil, errors.IO("getting absolute path", err)
	}

	ctx := &Context{}

	// Walk up to find festivals root and festival markers
	dir := absPath
	var pathComponents []string
	for {
		// Check for festivals/ directory
		if filepath.Base(dir) == "festivals" {
			ctx.FestivalsRoot = dir
			break
		}

		// Check for festival marker (FESTIVAL_GOAL.md or fest.yaml)
		if isFestivalRoot(dir) {
			ctx.FestivalDir = dir
		}

		pathComponents = append([]string{filepath.Base(dir)}, pathComponents...)
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root without finding festivals/
			break
		}
		dir = parent
	}

	// If we found a festivals root but no festival, check if cwd is in a status directory
	if ctx.FestivalsRoot != "" && ctx.FestivalDir == "" {
		// Try to find festival from relative path
		relPath, err := filepath.Rel(ctx.FestivalsRoot, absPath)
		if err == nil {
			parts := strings.Split(relPath, string(filepath.Separator))
			// Expected: status/festival-name or status/festival-name/...
			if len(parts) >= 2 {
				potentialFestival := filepath.Join(ctx.FestivalsRoot, parts[0], parts[1])
				if isFestivalRoot(potentialFestival) {
					ctx.FestivalDir = potentialFestival
				}
			}
		}
	}

	// Detect phase and sequence from current path
	if ctx.FestivalDir != "" {
		relPath, err := filepath.Rel(ctx.FestivalDir, absPath)
		if err == nil && relPath != "." {
			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) >= 1 && phasePattern.MatchString(parts[0]) {
				ctx.PhaseDir = filepath.Join(ctx.FestivalDir, parts[0])
			}
			if len(parts) >= 2 && sequencePattern.MatchString(parts[1]) {
				ctx.SequenceDir = filepath.Join(ctx.FestivalDir, parts[0], parts[1])
			}
		}
	}

	return ctx, nil
}

// isFestivalRoot checks if a directory is a festival root.
func isFestivalRoot(dir string) bool {
	markers := []string{"FESTIVAL_GOAL.md", "FESTIVAL_OVERVIEW.md", "fest.yaml"}
	for _, marker := range markers {
		if _, err := os.Stat(filepath.Join(dir, marker)); err == nil {
			return true
		}
	}
	return false
}

// ResolvePhase resolves a phase reference to a directory path.
// Input can be:
// - Numeric shortcut: "1", "01", "001" -> resolves to "001_*" directory
// - Full path: "/path/to/001_PLAN" -> used directly
// - Directory name: "001_PLAN" -> searched for in festival
func ResolvePhase(input string, festivalDir string) (string, error) {
	if input == "" {
		return "", errors.Validation("phase input is required")
	}

	// If it's a full path, use it directly
	if filepath.IsAbs(input) {
		if info, err := os.Stat(input); err == nil && info.IsDir() {
			return input, nil
		}
		return "", errors.NotFound("phase").WithField("path", input)
	}

	// If it looks like a numeric shortcut
	if isNumericShortcut(input, 3) {
		return resolveNumericPhase(input, festivalDir)
	}

	// Try as relative path within festival
	fullPath := filepath.Join(festivalDir, input)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return fullPath, nil
	}

	// Search for matching directory
	return searchPhaseDir(input, festivalDir)
}

// ResolveSequence resolves a sequence reference to a directory path.
// Input can be:
// - Numeric shortcut: "1", "01" -> resolves to "01_*" directory
// - Full path: "/path/to/01_setup" -> used directly
// - Directory name: "01_setup" -> searched for in phase
func ResolveSequence(input string, phaseDir string) (string, error) {
	if input == "" {
		return "", errors.Validation("sequence input is required")
	}

	// If it's a full path, use it directly
	if filepath.IsAbs(input) {
		if info, err := os.Stat(input); err == nil && info.IsDir() {
			return input, nil
		}
		return "", errors.NotFound("sequence").WithField("path", input)
	}

	// If it looks like a numeric shortcut
	if isNumericShortcut(input, 2) {
		return resolveNumericSequence(input, phaseDir)
	}

	// Try as relative path within phase
	fullPath := filepath.Join(phaseDir, input)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return fullPath, nil
	}

	return "", errors.NotFound("sequence").
		WithField("input", input).
		WithField("phase", filepath.Base(phaseDir))
}

// isNumericShortcut checks if input is a pure numeric string of maxDigits or less.
func isNumericShortcut(s string, maxDigits int) bool {
	if len(s) == 0 || len(s) > maxDigits {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// resolveNumericPhase resolves a numeric shortcut to a phase directory.
func resolveNumericPhase(shortcut string, festivalDir string) (string, error) {
	// Pad to 3 digits
	n := 0
	fmt.Sscanf(shortcut, "%d", &n)
	padded := fmt.Sprintf("%03d", n)

	// Search for matching directory
	entries, err := os.ReadDir(festivalDir)
	if err != nil {
		return "", errors.IO("reading festival directory", err).WithField("path", festivalDir)
	}

	var available []string
	for _, entry := range entries {
		if entry.IsDir() && phasePattern.MatchString(entry.Name()) {
			if strings.HasPrefix(entry.Name(), padded+"_") {
				return filepath.Join(festivalDir, entry.Name()), nil
			}
			available = append(available, entry.Name())
		}
	}

	return "", errors.NotFound("phase").
		WithField("shortcut", shortcut).
		WithField("looking_for", padded+"_*").
		WithField("available", strings.Join(available, ", "))
}

// resolveNumericSequence resolves a numeric shortcut to a sequence directory.
func resolveNumericSequence(shortcut string, phaseDir string) (string, error) {
	// Pad to 2 digits
	n := 0
	fmt.Sscanf(shortcut, "%d", &n)
	padded := fmt.Sprintf("%02d", n)

	// Search for matching directory
	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return "", errors.IO("reading phase directory", err).WithField("path", phaseDir)
	}

	var available []string
	for _, entry := range entries {
		if entry.IsDir() && sequencePattern.MatchString(entry.Name()) {
			if strings.HasPrefix(entry.Name(), padded+"_") {
				return filepath.Join(phaseDir, entry.Name()), nil
			}
			available = append(available, entry.Name())
		}
	}

	return "", errors.NotFound("sequence").
		WithField("shortcut", shortcut).
		WithField("looking_for", padded+"_*").
		WithField("phase", filepath.Base(phaseDir)).
		WithField("available", strings.Join(available, ", "))
}

// searchPhaseDir searches for a phase directory by name.
func searchPhaseDir(name string, festivalDir string) (string, error) {
	entries, err := os.ReadDir(festivalDir)
	if err != nil {
		return "", errors.IO("reading festival directory", err).WithField("path", festivalDir)
	}

	var available []string
	for _, entry := range entries {
		if entry.IsDir() && phasePattern.MatchString(entry.Name()) {
			available = append(available, entry.Name())
			// Check for exact match or suffix match
			if entry.Name() == name || strings.HasSuffix(entry.Name(), "_"+name) {
				return filepath.Join(festivalDir, entry.Name()), nil
			}
		}
	}

	return "", errors.NotFound("phase").
		WithField("name", name).
		WithField("available", strings.Join(available, ", "))
}

// ListPhases returns all phase directories in a festival.
func ListPhases(festivalDir string) ([]string, error) {
	entries, err := os.ReadDir(festivalDir)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalDir)
	}

	var phases []string
	for _, entry := range entries {
		if entry.IsDir() && phasePattern.MatchString(entry.Name()) {
			phases = append(phases, entry.Name())
		}
	}
	return phases, nil
}

// ListSequences returns all sequence directories in a phase.
func ListSequences(phaseDir string) ([]string, error) {
	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return nil, errors.IO("reading phase directory", err).WithField("path", phaseDir)
	}

	var sequences []string
	for _, entry := range entries {
		if entry.IsDir() && sequencePattern.MatchString(entry.Name()) {
			sequences = append(sequences, entry.Name())
		}
	}
	return sequences, nil
}
