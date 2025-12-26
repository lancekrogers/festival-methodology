package validation

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateCompletenessCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "completeness [festival-path]",
		Short: "Validate required files exist",
		Long: `Validate that all required files exist:

  • FESTIVAL_OVERVIEW.md (required)
  • PHASE_GOAL.md in each phase (required)
  • SEQUENCE_GOAL.md in each sequence (required)
  • FESTIVAL_RULES.md (recommended)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateCompleteness(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func runValidateCompleteness(opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_completeness",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateCompletenessChecks(festivalPath, result)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printValidationSection(display, "COMPLETENESS", result.Issues)
	return nil
}

func validateCompletenessChecks(festivalPath string, result *ValidationResult) {
	// Check FESTIVAL_OVERVIEW.md
	overviewPath := filepath.Join(festivalPath, "FESTIVAL_OVERVIEW.md")
	if !validateFileExists(overviewPath) {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelError,
			Code:    CodeMissingFile,
			Path:    overviewPath,
			Message: "FESTIVAL_OVERVIEW.md is required",
			Fix:     "Create FESTIVAL_OVERVIEW.md with project goals and success criteria",
		})
	}

	// Check FESTIVAL_RULES.md (warning, not error)
	rulesPath := filepath.Join(festivalPath, "FESTIVAL_RULES.md")
	if !validateFileExists(rulesPath) {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelWarning,
			Code:    CodeMissingFile,
			Path:    rulesPath,
			Message: "FESTIVAL_RULES.md is recommended",
		})
	}

	ctx := context.Background()
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(ctx, festivalPath)

	for _, phase := range phases {
		// Check PHASE_GOAL.md
		phaseGoalPath := filepath.Join(phase.Path, "PHASE_GOAL.md")
		if !validateFileExists(phaseGoalPath) {
			result.Issues = append(result.Issues, ValidationIssue{
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
			if !validateFileExists(seqGoalPath) {
				result.Issues = append(result.Issues, ValidationIssue{
					Level:   LevelError,
					Code:    CodeMissingGoal,
					Path:    seqGoalPath,
					Message: fmt.Sprintf("SEQUENCE_GOAL.md required in %s", seq.FullName),
					Fix:     fmt.Sprintf("fest create sequence --name %q --json", seq.Name),
				})
			}
		}
	}
}
