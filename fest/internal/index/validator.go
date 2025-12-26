package index

import (
	"fmt"
	"os"
	"path/filepath"
)

// ValidationResult represents the result of index validation
type ValidationResult struct {
	Valid       bool              `json:"valid"`
	Errors      []ValidationError `json:"errors,omitempty"`
	Warnings    []string          `json:"warnings,omitempty"`
	MissingInFS []string          `json:"missing_in_fs,omitempty"` // In index but not on disk
	ExtraInFS   []string          `json:"extra_in_fs,omitempty"`   // On disk but not in index
}

// ValidationError represents a validation error
type ValidationError struct {
	Type    string `json:"type"`
	Path    string `json:"path"`
	Message string `json:"message"`
}

// IndexValidator validates festival indices against filesystem
type IndexValidator struct {
	festivalRoot string
	index        *FestivalIndex
}

// NewIndexValidator creates a new index validator
func NewIndexValidator(festivalRoot string, index *FestivalIndex) *IndexValidator {
	return &IndexValidator{
		festivalRoot: festivalRoot,
		index:        index,
	}
}

// Validate validates the index against the filesystem
func (v *IndexValidator) Validate() *ValidationResult {
	result := &ValidationResult{
		Valid:       true,
		Errors:      []ValidationError{},
		Warnings:    []string{},
		MissingInFS: []string{},
		ExtraInFS:   []string{},
	}

	// Build set of indexed paths
	indexedPaths := make(map[string]bool)

	// Validate all indexed entries exist
	for _, phase := range v.index.Phases {
		phasePath := filepath.Join(v.festivalRoot, phase.Path)

		if !pathExists(phasePath) {
			result.MissingInFS = append(result.MissingInFS, phase.Path)
			result.Errors = append(result.Errors, ValidationError{
				Type:    "missing_phase",
				Path:    phase.Path,
				Message: fmt.Sprintf("Phase directory not found: %s", phase.Path),
			})
			result.Valid = false
		}
		indexedPaths[phase.Path] = true

		// Validate goal file if specified
		if phase.GoalFile != "" {
			goalPath := filepath.Join(phasePath, phase.GoalFile)
			if !pathExists(goalPath) {
				result.Warnings = append(result.Warnings, fmt.Sprintf("Phase goal file missing: %s", filepath.Join(phase.Path, phase.GoalFile)))
			}
		}

		for _, seq := range phase.Sequences {
			seqPath := filepath.Join(v.festivalRoot, seq.Path)

			if !pathExists(seqPath) {
				result.MissingInFS = append(result.MissingInFS, seq.Path)
				result.Errors = append(result.Errors, ValidationError{
					Type:    "missing_sequence",
					Path:    seq.Path,
					Message: fmt.Sprintf("Sequence directory not found: %s", seq.Path),
				})
				result.Valid = false
			}
			indexedPaths[seq.Path] = true

			// Validate goal file if specified
			if seq.GoalFile != "" {
				goalPath := filepath.Join(seqPath, seq.GoalFile)
				if !pathExists(goalPath) {
					result.Warnings = append(result.Warnings, fmt.Sprintf("Sequence goal file missing: %s", filepath.Join(seq.Path, seq.GoalFile)))
				}
			}

			for _, task := range seq.Tasks {
				taskPath := filepath.Join(v.festivalRoot, task.Path)

				if !pathExists(taskPath) {
					result.MissingInFS = append(result.MissingInFS, task.Path)
					result.Errors = append(result.Errors, ValidationError{
						Type:    "missing_task",
						Path:    task.Path,
						Message: fmt.Sprintf("Task file not found: %s", task.Path),
					})
					result.Valid = false
				}
				indexedPaths[task.Path] = true
			}
		}
	}

	// Find files on disk but not in index
	v.findExtraFiles(indexedPaths, result)

	return result
}

// findExtraFiles finds files on disk that are not in the index
func (v *IndexValidator) findExtraFiles(indexed map[string]bool, result *ValidationResult) {
	filepath.Walk(v.festivalRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		relPath, err := filepath.Rel(v.festivalRoot, path)
		if err != nil {
			return nil
		}

		// Skip root directory
		if relPath == "." {
			return nil
		}

		// Skip hidden directories and files (check safely)
		if len(relPath) > 0 && relPath[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-markdown files and directories
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".md" {
			return nil
		}

		// Skip goal files and other metadata
		base := filepath.Base(path)
		if base == "PHASE_GOAL.md" || base == "SEQUENCE_GOAL.md" ||
			base == "FESTIVAL_OVERVIEW.md" || base == "FESTIVAL_GOAL.md" {
			return nil
		}

		// Check if in index
		if !indexed[relPath] {
			result.ExtraInFS = append(result.ExtraInFS, relPath)
		}

		return nil
	})
}

// ValidateFromFile loads and validates an index from a file
func ValidateFromFile(festivalRoot string, indexPath string) (*ValidationResult, error) {
	index, err := LoadIndex(indexPath)
	if err != nil {
		return nil, err
	}

	validator := NewIndexValidator(festivalRoot, index)
	return validator.Validate(), nil
}

// pathExists checks if a path exists
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
