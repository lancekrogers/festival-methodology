package validator

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
)

// ValidateTasks ensures implementation sequences have task files (not just goals)
func ValidateTasks(festivalPath string) ([]Issue, error) {
	issues := []Issue{}

	parser := festival.NewParser()
	phases, err := parser.ParsePhases(festivalPath)
	if err != nil {
		return issues, fmt.Errorf("parse phases: %w", err)
	}

	policy := gates.DefaultPolicy()

	for _, phase := range phases {
		// Skip research phases - they use freeform structure
		if isResearchPhaseForTasks(phase.Name) {
			continue
		}

		sequences, err := parser.ParseSequences(phase.Path)
		if err != nil {
			return issues, fmt.Errorf("parse sequences: %w", err)
		}
		for _, seq := range sequences {
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			tasks, err := parser.ParseTasks(seq.Path)
			if err != nil {
				return issues, fmt.Errorf("parse tasks: %w", err)
			}
			if len(tasks) == 0 {
				rel, _ := filepath.Rel(festivalPath, seq.Path)
				issues = append(issues, Issue{
					Level:       LevelError,
					Code:        CodeMissingTaskFiles,
					Path:        rel,
					Message:     "Implementation sequence has SEQUENCE_GOAL.md but no task files",
					Fix:         fmt.Sprintf("fest create task --name \"design\" --path %s --json", rel),
					AutoFixable: false,
				})
			}
		}
	}

	return issues, nil
}

// isExcludedSequence checks if a sequence name matches exclusion patterns
func isExcludedSequence(name string, patterns []string) bool {
	for _, pattern := range patterns {
		if len(pattern) == 0 {
			continue
		}
		if pattern[0] == '*' {
			suffix := pattern[1:]
			if len(name) >= len(suffix) && name[len(name)-len(suffix):] == suffix {
				return true
			}
		} else if pattern[len(pattern)-1] == '*' {
			prefix := pattern[:len(pattern)-1]
			if len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				return true
			}
		} else if name == pattern {
			return true
		}
	}
	return false
}

// CheckTaskFilesExist returns true if all implementation sequences have at least one task file.
func CheckTaskFilesExist(path string) bool {
	issues, err := ValidateTasks(path)
	if err != nil {
		return true
	}
	return len(issues) == 0
}

// isResearchPhaseForTasks checks if a phase is a research phase.
// Research phases use freeform subdirectory structure and don't require numbered tasks.
func isResearchPhaseForTasks(phaseName string) bool {
	normalized := strings.ToUpper(phaseName)
	return strings.Contains(normalized, "RESEARCH")
}
