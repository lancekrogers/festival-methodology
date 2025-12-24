package validation

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/validator"
	"github.com/spf13/cobra"
)

func newValidateOrderingCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "ordering [festival-path]",
		Short: "Validate sequential numbering (gap detection)",
		Long: `Validate that festival elements are sequentially numbered without gaps.

This checks:
  • Phases are sequential: 001, 002, 003... (no gaps, must start at 001)
  • Sequences within each phase: 01, 02, 03... (no gaps, must start at 01)
  • Tasks within each sequence: 01, 02, 03... (no gaps, must start at 01)

Parallel work (same number) is allowed if items are CONSECUTIVE:
  VALID:   01_task_a.md, 01_task_b.md, 02_task_c.md
  INVALID: 01_task_a.md, 02_task_b.md, 01_task_c.md

Gaps break agent execution order - agents rely on sequential numbering
to determine which phase/sequence/task to work on next.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateOrdering(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func runValidateOrdering(opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_ordering",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateOrderingChecks(festivalPath, result)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		if !result.Valid {
			emitValidateJSON(result)
			return fmt.Errorf("ordering validation failed with %d issue(s)", len(result.Issues))
		}
		return emitValidateJSON(result)
	}

	printValidationSection(display, "ORDERING (Gap Detection)", result.Issues)

	if !result.Valid {
		return fmt.Errorf("ordering validation failed with %d issue(s)", len(result.Issues))
	}
	return nil
}

func validateOrderingChecks(festivalPath string, result *ValidationResult) {
	v := validator.NewOrderingValidator()
	issues, err := v.Validate(context.Background(), festivalPath)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelError,
			Code:    "ordering_error",
			Path:    festivalPath,
			Message: fmt.Sprintf("Failed to validate ordering: %v", err),
		})
		return
	}

	// Convert validator.Issue to ValidationIssue
	for _, issue := range issues {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:       issue.Level,
			Code:        issue.Code,
			Path:        issue.Path,
			Message:     issue.Message,
			Fix:         issue.Fix,
			AutoFixable: issue.AutoFixable,
		})
	}
}
