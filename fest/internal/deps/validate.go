package deps

import (
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ValidationResult holds the results of dependency validation
type ValidationResult struct {
	Valid    bool              `json:"valid"`
	Errors   []ValidationError `json:"errors,omitempty"`
	Warnings []ValidationError `json:"warnings,omitempty"`
	Graph    *Graph            `json:"graph,omitempty"`
}

// Validate checks all dependency declarations in a festival
func Validate(festivalPath string) (*ValidationResult, error) {
	resolver := NewResolver(festivalPath)
	graph, err := resolver.ResolveFestival()
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve dependencies")
	}

	result := &ValidationResult{
		Valid: true,
		Graph: graph,
	}

	// Check for cycles
	_, err = graph.TopologicalSort()
	if err != nil {
		if cycleErr, ok := err.(*CycleError); ok {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "CYCLE_DETECTED",
				Message:  fmt.Sprintf("Circular dependency detected: %v", cycleErr.Cycle),
				Severity: "error",
			})
			result.Valid = false
		}
	}

	// Check for missing dependencies
	for _, task := range graph.Tasks {
		for _, depRef := range task.Dependencies {
			depTask := resolver.resolveTaskReference(task, depRef)
			if depTask == nil {
				result.Errors = append(result.Errors, ValidationError{
					TaskID:   task.ID,
					Code:     "MISSING_DEPENDENCY",
					Message:  fmt.Sprintf("Task %s declares dependency on %q which does not exist", task.Name, depRef),
					Severity: "error",
				})
				result.Valid = false
			}
		}

		for _, depRef := range task.SoftDeps {
			depTask := resolver.resolveTaskReference(task, depRef)
			if depTask == nil {
				result.Warnings = append(result.Warnings, ValidationError{
					TaskID:   task.ID,
					Code:     "MISSING_SOFT_DEPENDENCY",
					Message:  fmt.Sprintf("Task %s declares soft dependency on %q which does not exist", task.Name, depRef),
					Severity: "warning",
				})
			}
		}
	}

	// Check for implicit dependency violations
	// (e.g., task 03 modifying files that task 01 reads)
	result.Warnings = append(result.Warnings, checkImplicitViolations(graph)...)

	return result, nil
}

// ValidateSequence validates dependencies within a single sequence
func ValidateSequence(seqPath string) (*ValidationResult, error) {
	festivalPath := findFestivalPath(seqPath)
	resolver := NewResolver(festivalPath)
	graph, err := resolver.ResolveSequence(seqPath)
	if err != nil {
		return nil, errors.Wrap(err, "failed to resolve sequence dependencies")
	}

	result := &ValidationResult{
		Valid: true,
		Graph: graph,
	}

	// Check for cycles
	_, err = graph.TopologicalSort()
	if err != nil {
		if cycleErr, ok := err.(*CycleError); ok {
			result.Errors = append(result.Errors, ValidationError{
				Code:     "CYCLE_DETECTED",
				Message:  fmt.Sprintf("Circular dependency detected: %v", cycleErr.Cycle),
				Severity: "error",
			})
			result.Valid = false
		}
	}

	return result, nil
}

// checkImplicitViolations checks for potential implicit dependency violations
func checkImplicitViolations(graph *Graph) []ValidationError {
	var warnings []ValidationError

	// Group tasks by sequence
	bySequence := make(map[string][]*Task)
	for _, task := range graph.Tasks {
		bySequence[task.SequencePath] = append(bySequence[task.SequencePath], task)
	}

	// Check each sequence
	for seqPath, tasks := range bySequence {
		// Check for gaps in numbering
		if len(tasks) > 0 {
			numbers := make(map[int]bool)
			maxNum := 0
			for _, task := range tasks {
				numbers[task.Number] = true
				if task.Number > maxNum {
					maxNum = task.Number
				}
			}

			// Check for gaps (e.g., 01, 02, 04 missing 03)
			for i := 1; i < maxNum; i++ {
				if !numbers[i] {
					warnings = append(warnings, ValidationError{
						Code:     "NUMBERING_GAP",
						Message:  fmt.Sprintf("Sequence %s has a gap in task numbering at position %d", seqPath, i),
						Severity: "warning",
					})
				}
			}
		}
	}

	return warnings
}

// findFestivalPath walks up from seqPath to find the festival root
func findFestivalPath(seqPath string) string {
	// Go up two directories (sequence -> phase -> festival)
	phasePath := parentDir(seqPath)
	festivalPath := parentDir(phasePath)

	// Verify it looks like a festival
	if _, err := os.Stat(festivalPath); err == nil {
		return festivalPath
	}

	return seqPath
}

// parentDir returns the parent directory
func parentDir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return path
}
