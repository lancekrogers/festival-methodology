package validation

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateStructureCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "structure [festival-path]",
		Short: "Validate naming conventions and hierarchy",
		Long: `Validate that festival structure follows naming conventions:

  • Phases: NNN_PHASE_NAME (3-digit prefix, UPPERCASE)
  • Sequences: NN_sequence_name (2-digit prefix, lowercase)
  • Tasks: NN_task_name.md (2-digit prefix, lowercase, .md extension)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateStructure(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func runValidateStructure(opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_structure",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateStructureChecks(festivalPath, result)

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

	printValidationSection(display, "STRUCTURE", result.Issues)
	return nil
}

func validateStructureChecks(festivalPath string, result *ValidationResult) {
	ctx := context.Background()
	parser := festival.NewParser()

	// Parse festival structure
	phases, err := parser.ParsePhases(ctx, festivalPath)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelError,
			Code:    CodeNamingConvention,
			Path:    festivalPath,
			Message: fmt.Sprintf("Failed to parse phases: %v", err),
		})
		return
	}

	// Check phase naming (should be UPPERCASE after number)
	phaseUpperPattern := regexp.MustCompile(`^\d{3}_[A-Z][A-Z0-9_]*$`)
	for _, phase := range phases {
		if !phaseUpperPattern.MatchString(phase.FullName) {
			result.Issues = append(result.Issues, ValidationIssue{
				Level:   LevelError,
				Code:    CodeNamingConvention,
				Path:    phase.Path,
				Message: fmt.Sprintf("Phase name should be UPPERCASE: %s", phase.FullName),
				Fix:     "Rename phase directory to use UPPERCASE (e.g., 001_PHASE_NAME)",
			})
		}

		// Check sequences
		sequences, _ := parser.ParseSequences(ctx, phase.Path)
		seqLowerPattern := regexp.MustCompile(`^\d{2}_[a-z][a-z0-9_]*$`)
		for _, seq := range sequences {
			if !seqLowerPattern.MatchString(seq.FullName) {
				result.Issues = append(result.Issues, ValidationIssue{
					Level:   LevelError,
					Code:    CodeNamingConvention,
					Path:    seq.Path,
					Message: fmt.Sprintf("Sequence name should be lowercase: %s", seq.FullName),
					Fix:     "Rename sequence directory to use lowercase (e.g., 01_sequence_name)",
				})
			}

			// Check tasks
			tasks, _ := parser.ParseTasks(ctx, seq.Path)
			taskLowerPattern := regexp.MustCompile(`^\d{2}_[a-z][a-z0-9_]*\.md$`)
			for _, task := range tasks {
				if !taskLowerPattern.MatchString(task.FullName) {
					result.Issues = append(result.Issues, ValidationIssue{
						Level:   LevelError,
						Code:    CodeNamingConvention,
						Path:    task.Path,
						Message: fmt.Sprintf("Task name should be lowercase: %s", task.FullName),
						Fix:     "Rename task file to use lowercase (e.g., 01_task_name.md)",
					})
				}
			}
		}
	}
}
