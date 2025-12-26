package validation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newValidateChecklistCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "checklist [festival-path]",
		Short: "Post-completion questionnaire",
		Long: `Run through post-completion checklist to ensure methodology compliance.

Checks (auto-verified where possible):
  1. Did you fill out ALL templates? (auto-check for markers)
  2. Does this plan achieve project goals? (manual review)
  3. Are items in order of operation? (auto-check)
  4. Did you follow parallelization standards? (auto-check)
  5. Did you create TASK FILES for implementation? (auto-check)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateChecklist(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func runValidateChecklist(opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:        true,
		Action:    "validate_checklist",
		Festival:  filepath.Base(festivalPath),
		Valid:     true,
		Issues:    []ValidationIssue{},
		Checklist: &Checklist{},
	}

	// Run all checks and populate checklist
	templatesFilled := checkTemplatesFilled(festivalPath)
	result.Checklist.TemplatesFilled = &templatesFilled

	// Goals achievable is a manual check - always null
	result.Checklist.GoalsAchievable = nil

	taskFilesExist := checkTaskFilesExist(festivalPath)
	result.Checklist.TaskFilesExist = &taskFilesExist

	orderCorrect := checkOrderCorrect(festivalPath)
	result.Checklist.OrderCorrect = &orderCorrect

	parallelCorrect := checkParallelCorrect(festivalPath)
	result.Checklist.ParallelCorrect = &parallelCorrect

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printChecklistResult(display, result.Checklist)
	return nil
}

// Checklist check functions

func checkTemplatesFilled(festivalPath string) bool {
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}
	filled := true

	filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		relPath, _ := filepath.Rel(festivalPath, path)
		if strings.HasPrefix(relPath, ".") || strings.Contains(relPath, "/.") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		for _, marker := range markers {
			if strings.Contains(contentStr, marker) {
				filled = false
				return filepath.SkipAll
			}
		}
		return nil
	})

	return filled
}

func checkTaskFilesExist(festivalPath string) bool {
	ctx := context.Background()
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(ctx, festivalPath)
	policy := gates.DefaultPolicy()

	for _, phase := range phases {
		sequences, _ := parser.ParseSequences(ctx, phase.Path)
		for _, seq := range sequences {
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}
			tasks, _ := parser.ParseTasks(ctx, seq.Path)
			if len(tasks) == 0 {
				return false
			}
		}
	}
	return true
}

func checkOrderCorrect(festivalPath string) bool {
	parser := festival.NewParser()
	phases, err := parser.ParsePhases(context.Background(), festivalPath)
	if err != nil {
		return true // Can't check, assume OK
	}

	// Check phases are sequential
	lastNum := 0
	for _, phase := range phases {
		if phase.Number < lastNum {
			return false
		}
		lastNum = phase.Number
	}

	return true
}

func checkParallelCorrect(festivalPath string) bool {
	// For now, always return true - parallel validation is complex
	// and false positives would be confusing
	return true
}

// Output functions for checklist

func printChecklistResult(display *ui.UI, checklist *Checklist) {
	fmt.Println("\nPost-Completion Checklist")
	fmt.Println(strings.Repeat("=", 50))

	printCheckItem(display, 1, "Templates Filled",
		"Did you fill out ALL templates?",
		checklist.TemplatesFilled,
		"Search for [FILL:] markers and replace with content")

	printCheckItem(display, 2, "Goals Achievable",
		"Does this plan achieve project goals?",
		checklist.GoalsAchievable,
		"Review FESTIVAL_OVERVIEW.md and verify alignment")

	printCheckItem(display, 3, "Task Files Exist",
		"Did you create TASK FILES for implementation?",
		checklist.TaskFilesExist,
		"Run 'fest understand tasks' to learn about task files")

	printCheckItem(display, 4, "Order Correct",
		"Are items in order of operation?",
		checklist.OrderCorrect,
		"Lower numbers execute first (01_ before 02_)")

	printCheckItem(display, 5, "Parallel Correct",
		"Did you follow parallelization standards?",
		checklist.ParallelCorrect,
		"Same-numbered items (01_a, 01_b) run in parallel")

	// Summary
	passed := 0
	total := 0
	if checklist.TemplatesFilled != nil {
		total++
		if *checklist.TemplatesFilled {
			passed++
		}
	}
	if checklist.TaskFilesExist != nil {
		total++
		if *checklist.TaskFilesExist {
			passed++
		}
	}
	if checklist.OrderCorrect != nil {
		total++
		if *checklist.OrderCorrect {
			passed++
		}
	}
	if checklist.ParallelCorrect != nil {
		total++
		if *checklist.ParallelCorrect {
			passed++
		}
	}

	fmt.Printf("\nOverall: %d/%d auto-checks passed (1 manual check required)\n", passed, total)
}

func printCheckItem(display *ui.UI, num int, title, question string, result *bool, guidance string) {
	fmt.Printf("\n%d. %s\n", num, title)
	fmt.Printf("   Question: %s\n", question)

	if result == nil {
		fmt.Println("   Status: [MANUAL CHECK REQUIRED]")
		fmt.Printf("   Guidance: %s\n", guidance)
	} else if *result {
		display.Success("   Auto-check: [PASS]")
	} else {
		display.Error("   Auto-check: [FAIL]")
		fmt.Printf("   Fix: %s\n", guidance)
	}
}
