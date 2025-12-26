package validator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// CompletenessValidator validates that required files exist.
type CompletenessValidator struct{}

// NewCompletenessValidator creates a new completeness validator.
func NewCompletenessValidator() *CompletenessValidator {
	return &CompletenessValidator{}
}

// Validate checks that all required files exist.
func (v *CompletenessValidator) Validate(ctx context.Context, path string) ([]Issue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var issues []Issue

	// Check FESTIVAL_OVERVIEW.md
	overviewPath := filepath.Join(path, "FESTIVAL_OVERVIEW.md")
	if !fileExists(overviewPath) {
		issues = append(issues, Issue{
			Level:   LevelError,
			Code:    CodeMissingFile,
			Path:    overviewPath,
			Message: "FESTIVAL_OVERVIEW.md is required",
			Fix:     "Create FESTIVAL_OVERVIEW.md with project goals and success criteria",
		})
	}

	// Check FESTIVAL_RULES.md (warning, not error)
	rulesPath := filepath.Join(path, "FESTIVAL_RULES.md")
	if !fileExists(rulesPath) {
		issues = append(issues, Issue{
			Level:   LevelWarning,
			Code:    CodeMissingFile,
			Path:    rulesPath,
			Message: "FESTIVAL_RULES.md is recommended",
		})
	}

	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(ctx, path)

	for _, phase := range phases {
		if err := ctx.Err(); err != nil {
			return issues, err
		}

		// Check PHASE_GOAL.md
		phaseGoalPath := filepath.Join(phase.Path, "PHASE_GOAL.md")
		if !fileExists(phaseGoalPath) {
			issues = append(issues, Issue{
				Level:   LevelError,
				Code:    CodeMissingGoal,
				Path:    phaseGoalPath,
				Message: fmt.Sprintf("PHASE_GOAL.md required in %s", phase.FullName),
				Fix:     fmt.Sprintf("fest create phase --name %q --json", phase.Name),
			})
		}

		// Check sequences
		sequences, _ := parser.ParseSequences(ctx, phase.Path)
		for _, seq := range sequences {
			seqGoalPath := filepath.Join(seq.Path, "SEQUENCE_GOAL.md")
			if !fileExists(seqGoalPath) {
				issues = append(issues, Issue{
					Level:   LevelError,
					Code:    CodeMissingGoal,
					Path:    seqGoalPath,
					Message: fmt.Sprintf("SEQUENCE_GOAL.md required in %s", seq.FullName),
					Fix:     fmt.Sprintf("fest create sequence --name %q --json", seq.Name),
				})
			}
		}
	}

	return issues, nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
