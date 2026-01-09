package validation

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateQualityGatesCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "quality-gates [festival-path]",
		Short: "Validate quality gates exist",
		Long: `Validate that implementation sequences have quality gate tasks.

Quality gates are required for implementation sequences:
  • XX_testing_and_verify.md
  • XX_code_review.md
  • XX_review_results_iterate.md
  • XX_commit.md

Use --fix to automatically add missing quality gates.
Planning sequences (*_planning, *_research, etc.) are excluded.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateQualityGates(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Automatically add missing quality gates")

	return cmd
}

func runValidateQualityGates(ctx context.Context, opts *validateOptions) error {
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
		Action:   "validate_quality_gates",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateQualityGatesChecks(ctx, festivalPath, result, opts.fix)

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

	printValidationSection(display, "QUALITY GATES", result.Issues)
	return nil
}

func validateQualityGatesChecks(ctx context.Context, festivalPath string, result *ValidationResult, autoFix bool) {
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(ctx, festivalPath)
	policy := gates.DefaultPolicy()

	gateIDs := map[string]bool{}
	for _, gate := range policy.GetEnabledTasks() {
		gateIDs[gate.ID] = true
	}

	// Phase name patterns that don't require quality gates
	nonImplPhasePatterns := []string{"PLAN", "DESIGN", "REVIEW", "UAT", "FINALIZE", "DOCS", "RESEARCH"}

	for _, phase := range phases {
		// Skip non-implementation phases entirely
		if isNonImplementationPhase(phase.Name, nonImplPhasePatterns) {
			continue
		}

		sequences, _ := parser.ParseSequences(ctx, phase.Path)
		for _, seq := range sequences {
			// Check if this is an implementation sequence
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			// Check for quality gate tasks
			tasks, _ := parser.ParseTasks(ctx, seq.Path)
			hasGates := false
			for _, task := range tasks {
				// Check if task name contains any gate ID
				for gateID := range gateIDs {
					if strings.Contains(strings.ToLower(task.Name), strings.ReplaceAll(gateID, "_", "")) ||
						strings.Contains(task.Name, gateID) {
						hasGates = true
						break
					}
				}
			}

			if !hasGates && len(tasks) > 0 {
				relPath, _ := filepath.Rel(festivalPath, seq.Path)
				result.Issues = append(result.Issues, ValidationIssue{
					Level:       LevelError,
					Code:        CodeMissingQualityGate,
					Path:        relPath,
					Message:     "Implementation sequence missing quality gates",
					Fix:         "fest gates apply --sequence " + relPath + " --approve",
					AutoFixable: true,
				})

				if autoFix {
					// TODO: Call fest gates apply
					result.FixesApplied = append(result.FixesApplied, FixApplied{
						Code:   CodeMissingQualityGate,
						Path:   relPath,
						Action: "Quality gates would be added (--fix not yet implemented)",
					})
				}
			}
		}
	}
}

// isNonImplementationPhase checks if a phase is for planning/review rather than implementation
func isNonImplementationPhase(phaseName string, patterns []string) bool {
	upperName := strings.ToUpper(phaseName)
	for _, pattern := range patterns {
		if strings.Contains(upperName, pattern) {
			return true
		}
	}
	return false
}

// isExcludedSequence checks if a sequence name matches exclusion patterns
func isExcludedSequence(name string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple glob matching
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(name, suffix) {
				return true
			}
		} else if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		} else if name == pattern {
			return true
		}
	}
	return false
}
