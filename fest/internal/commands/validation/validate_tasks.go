package validation

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateTasksCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "tasks [festival-path]",
		Short: "Validate task files exist (CRITICAL)",
		Long: `Validate that implementation sequences have TASK FILES.

THIS IS THE MOST COMMON MISTAKE: Creating sequences with only
SEQUENCE_GOAL.md but no task files.

  Goals define WHAT to achieve.
  Tasks define HOW to execute.

AI agents EXECUTE TASK FILES. Without them, agents know the objective
but don't know what specific work to perform.

CORRECT STRUCTURE:
  02_api/
  ├── SEQUENCE_GOAL.md          ← Defines objective
  ├── 01_design_endpoints.md    ← Task: design work
  ├── 02_implement_crud.md      ← Task: implementation
  └── 03_testing_and_verify.md  ← Quality gate

INCORRECT STRUCTURE (common mistake):
  02_api/
  └── SEQUENCE_GOAL.md          ← No task files!`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateTasks(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func runValidateTasks(ctx context.Context, opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	if ctx == nil {
		ctx = context.Background()
	}

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_tasks",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateTaskFilesChecks(ctx, festivalPath, result)

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

	printTaskValidationResult(display, result)
	return nil
}

func validateTaskFilesChecks(ctx context.Context, festivalPath string, result *ValidationResult) {
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(ctx, festivalPath)
	policy := gates.DefaultPolicy()

	for _, phase := range phases {
		sequences, _ := parser.ParseSequences(ctx, phase.Path)
		for _, seq := range sequences {
			// Check if this is an implementation sequence
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			// Check for task files
			tasks, _ := parser.ParseTasks(ctx, seq.Path)
			if len(tasks) == 0 {
				relPath, _ := filepath.Rel(festivalPath, seq.Path)
				result.Issues = append(result.Issues, ValidationIssue{
					Level:       LevelError,
					Code:        CodeMissingTaskFiles,
					Path:        relPath,
					Message:     "Implementation sequence has SEQUENCE_GOAL.md but no task files",
					Fix:         fmt.Sprintf("fest create task --name \"design\" --path %s --json", relPath),
					AutoFixable: false,
				})
			}
		}
	}
}

func printTaskValidationSection(display *ui.UI, issues []ValidationIssue) {
	printSectionHeader("Task Files", issues)
	fmt.Println(ui.Dim("Critical for AI execution"))

	if len(issues) == 0 {
		display.Success("All implementation sequences have task files")
		return
	}

	display.Error("Implementation sequences need task files, not just goals")
	fmt.Println()
	fmt.Println(ui.Info("Goals define what to achieve; tasks define how to execute."))
	fmt.Println(ui.Info("AI agents execute task files."))
	fmt.Println()
	fmt.Println(ui.H3("Sequences without tasks"))

	for _, issue := range issues {
		fmt.Printf("  - %s\n", ui.Dim(issue.Path))
	}

	fmt.Println()
	fmt.Println(ui.H3("For each sequence, create task files"))
	fmt.Println("  fest create task --name \"design\" --json")
	fmt.Println("  fest create task --name \"implement\" --json")
	fmt.Println("  fest create task --name \"test\" --json")
}

func printTaskValidationResult(display *ui.UI, result *ValidationResult) {
	printTaskValidationSection(display, result.Issues)
}
