// Package taskfilter provides unified file classification for festival progress tracking.
// This package is the single source of truth for determining what counts as a task,
// gate, or other festival file type.
package taskfilter

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// taskPattern matches task file names: NN_name.md or NN.N_name.md
	// Examples: 01_design.md, 02_implement.md, 01.5_hotfix.md
	taskPattern = regexp.MustCompile(`^\d{2}[\._].*\.md$`)

	// phasePattern matches phase directory names: NNN_Name
	// Examples: 001_PLANNING, 002_IMPLEMENTATION
	phasePattern = regexp.MustCompile(`^\d{3}_`)

	// sequencePattern matches sequence directory names: NN_Name
	// Examples: 01_setup, 02_core_feature
	sequencePattern = regexp.MustCompile(`^\d{2}_`)
)

// FileType represents the type of festival file
type FileType int

const (
	// FileTypeUnknown is an unrecognized file type
	FileTypeUnknown FileType = iota
	// FileTypeTask is a regular task file
	FileTypeTask
	// FileTypeGate is a quality gate file (testing, review, etc.)
	FileTypeGate
	// FileTypeGoal is a goal document (SEQUENCE_GOAL.md, PHASE_GOAL.md, etc.)
	FileTypeGoal
)

// String returns the string representation of the FileType
func (ft FileType) String() string {
	switch ft {
	case FileTypeTask:
		return "task"
	case FileTypeGate:
		return "gate"
	case FileTypeGoal:
		return "goal"
	default:
		return "unknown"
	}
}

// goalFiles contains the known goal file names
var goalFiles = map[string]bool{
	"SEQUENCE_GOAL.md":     true,
	"PHASE_GOAL.md":        true,
	"FESTIVAL_GOAL.md":     true,
	"FESTIVAL_OVERVIEW.md": true,
	"TODO.md":              true,
	"CONTEXT.md":           true,
	"FESTIVAL_RULES.md":    true,
}

// gatePatterns contains patterns that identify quality gate files.
// These patterns are matched against the filename with the numeric prefix removed.
// For example, "04_testing_and_verify.md" becomes "testing_and_verify"
var gatePatterns = []string{
	"gate",                   // Matches *_gate.md, *_quality_gate.md
	"testing_and_verify",     // Matches *_testing_and_verify.md
	"code_review",            // Matches *_code_review.md
	"review_results_iterate", // Matches *_review_results_iterate.md
}

// gateExactMatches contains exact matches for gate files (after prefix removal)
var gateExactMatches = map[string]bool{
	"commit": true, // Only "NN_commit.md" is a gate, not "NN_commit_changes.md"
}

// ClassifyFile determines the type of a festival file based on its name.
// This is the canonical classification logic used by all progress tracking commands.
func ClassifyFile(filename string) FileType {
	// Check if it's a goal file first
	if goalFiles[filename] {
		return FileTypeGoal
	}

	// Must be a markdown file with task pattern to be considered
	if !taskPattern.MatchString(filename) {
		return FileTypeUnknown
	}

	// Extract the name part after the numeric prefix (e.g., "04_testing_and_verify.md" -> "testing_and_verify")
	namePart := extractNamePart(filename)
	lower := strings.ToLower(namePart)

	// Check exact matches first (like "commit")
	if gateExactMatches[lower] {
		return FileTypeGate
	}

	// Check if it's a gate file using pattern matching
	for _, pattern := range gatePatterns {
		if strings.Contains(lower, pattern) {
			return FileTypeGate
		}
	}

	// It's a regular task
	return FileTypeTask
}

// extractNamePart extracts the name portion of a task filename, removing prefix and extension.
// "04_testing_and_verify.md" -> "testing_and_verify"
// "01.5_hotfix.md" -> "hotfix"
func extractNamePart(filename string) string {
	// Remove .md extension
	name := strings.TrimSuffix(filename, ".md")

	// Find the first underscore after digits
	for i, c := range name {
		if c == '_' {
			if i+1 < len(name) {
				return name[i+1:]
			}
			return ""
		}
	}
	return name
}

// IsTask returns true if the file should be counted as a task for progress tracking.
// By default, this returns true for both tasks AND gates since gates are work items.
// Use ClassifyFile if you need to distinguish between tasks and gates.
func IsTask(filename string) bool {
	ft := ClassifyFile(filename)
	return ft == FileTypeTask || ft == FileTypeGate
}

// IsTaskOnly returns true if the file is a regular task (not a gate).
func IsTaskOnly(filename string) bool {
	return ClassifyFile(filename) == FileTypeTask
}

// IsGate returns true if the file is a quality gate file.
func IsGate(filename string) bool {
	return ClassifyFile(filename) == FileTypeGate
}

// IsGoal returns true if the file is a goal document.
func IsGoal(filename string) bool {
	return ClassifyFile(filename) == FileTypeGoal
}

// IsPhaseDir returns true if the directory name matches the phase naming pattern.
func IsPhaseDir(name string) bool {
	return phasePattern.MatchString(name)
}

// IsSequenceDir returns true if the directory name matches the sequence naming pattern.
func IsSequenceDir(name string) bool {
	return sequencePattern.MatchString(name)
}

// ShouldTrack returns true if the file should be included in progress tracking.
// This includes both tasks and gates but excludes goals and unknown files.
func ShouldTrack(filename string) bool {
	ft := ClassifyFile(filename)
	return ft == FileTypeTask || ft == FileTypeGate
}

// FilterOption controls how files are filtered for progress calculation
type FilterOption struct {
	// IncludeGates determines whether gates are counted as part of task totals
	// When true (default): tasks + gates are counted together
	// When false: only regular tasks are counted, gates are separate
	IncludeGates bool
}

// DefaultFilterOption returns the default filter options
func DefaultFilterOption() FilterOption {
	return FilterOption{
		IncludeGates: true,
	}
}

// FileInfo contains classified information about a festival file
type FileInfo struct {
	Name    string
	Path    string
	Type    FileType
	IsTask  bool
	IsGate  bool
	IsGoal  bool
	Tracked bool
}

// ClassifyPath classifies a file given its full path
func ClassifyPath(path string) FileInfo {
	filename := filepath.Base(path)
	ft := ClassifyFile(filename)
	return FileInfo{
		Name:    filename,
		Path:    path,
		Type:    ft,
		IsTask:  ft == FileTypeTask,
		IsGate:  ft == FileTypeGate,
		IsGoal:  ft == FileTypeGoal,
		Tracked: ft == FileTypeTask || ft == FileTypeGate,
	}
}
