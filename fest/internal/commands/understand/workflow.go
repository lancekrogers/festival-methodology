package understand

import (
	"fmt"
	"path/filepath"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

func newUnderstandWorkflowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "workflow",
		Short: "Just-in-time reading and execution patterns",
		Long:  `Learn the just-in-time approach to reading templates and documentation, preserving context for actual work.`,
		Run: func(cmd *cobra.Command, args []string) {
			printWorkflow(findDotFestivalDir())
		},
	}
}

func newUnderstandRulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "MANDATORY structure rules for automation",
		Long:  `Learn the RIGID structure requirements that enable Festival automation: naming conventions, required files, quality gates, and parallel execution.`,
		Run: func(cmd *cobra.Command, args []string) {
			printRules()
		},
	}
}

func newUnderstandTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "Template variables that save tokens",
		Long:  `Learn how to pass variables to fest create commands to generate pre-filled documents, minimizing post-creation editing and saving tokens.`,
		Run: func(cmd *cobra.Command, args []string) {
			printTemplates()
		},
	}
}

func newUnderstandChecklistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checklist",
		Short: "Quick festival validation checklist",
		Long: `Show a quick checklist for validating your festival structure.

This is a quick reference. For full validation, run 'fest validate checklist'.

Checklist:
  1. FESTIVAL_OVERVIEW.md exists and is filled out
  2. Each phase has PHASE_GOAL.md
  3. Each sequence has SEQUENCE_GOAL.md
  4. Implementation sequences have TASK FILES (not just goals!)
  5. Quality gates present in implementation sequences
  6. No unfilled template markers ([FILL:], {{ }})`,
		Run: func(cmd *cobra.Command, args []string) {
			printChecklist()
		},
	}
}

func printWorkflow(dotFestival string) {
	// Hybrid: try .festival/ supplements first, fall back to embedded defaults
	if dotFestival != "" {
		readmePath := filepath.Join(dotFestival, "README.md")
		hasContent := false

		// Check if .festival/ has workflow content
		if content := extractSection(readmePath, "When to Read What", "### Never Do This"); content != "" {
			fmt.Print("\nFestival Workflow - Just-in-Time Reading\n")
			fmt.Println("========================================")
			fmt.Println("\nThe just-in-time approach preserves context window for actual work.")
			fmt.Println("\nWhen to Read What (from .festival/)")
			fmt.Println("------------------------------------")
			fmt.Println(content)
			hasContent = true

			if never := extractSection(readmePath, "### Never Do This", "### Always Do This"); never != "" {
				fmt.Println("\nNEVER Do This")
				fmt.Println("-------------")
				fmt.Println(never)
			}

			if always := extractSection(readmePath, "### Always Do This", "## Quick Navigation"); always != "" {
				fmt.Println("\nALWAYS Do This")
				fmt.Println("--------------")
				fmt.Println(always)
			}
		}

		if hasContent {
			fmt.Printf("\nSource: %s\n", dotFestival)
			return
		}
	}

	// Default: use embedded content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("workflow.txt"))
}

func printRules() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("rules.txt"))
}

func printTemplates() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("templates.txt"))
}

func printChecklist() {
	fmt.Print(`
Festival Validation Checklist
=============================

Before executing a festival, verify:

  ✓ Festival Level
    □ FESTIVAL_OVERVIEW.md exists and defines goals
    □ FESTIVAL_RULES.md exists with quality standards

  ✓ Phase Level
    □ Each phase has PHASE_GOAL.md
    □ Phases are numbered correctly (001_, 002_, ...)

  ✓ Sequence Level
    □ Each sequence has SEQUENCE_GOAL.md
    □ Sequences are numbered correctly (01_, 02_, ...)

  ✗ CRITICAL: Task Files
    □ Implementation sequences have TASK FILES
    □ Not just SEQUENCE_GOAL.md - actual task files!
    □ Goals define WHAT; tasks define HOW
    □ AI agents execute TASK FILES

  ✓ Quality Gates
    □ Implementation sequences end with quality gates
    □ XX_testing_and_verify.md
    □ XX_code_review.md
    □ XX_review_results_iterate.md

  ✓ Templates
    □ No [FILL:] markers remaining
    □ No {{ }} template syntax in final docs


Quick Validation Commands
-------------------------

  fest validate                # Full validation
  fest validate tasks          # Check task files exist
  fest validate checklist      # Detailed checklist with auto-checks

`)
}
