package understand

import (
	"fmt"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

// NewUnderstandCommand creates the understand command group with subcommands
func NewUnderstandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "understand",
		Short: "Learn methodology FIRST - run before executing festival tasks",
		Long: `Learn about Festival Methodology - a goal-oriented project management
system designed for AI agent development workflows.

START HERE if you're new to Festival Methodology:
  fest understand methodology    Core principles and philosophy
  fest understand structure      3-level hierarchy explained
  fest understand tasks          CRITICAL: When to create task files

QUICK REFERENCE:
  fest understand checklist      Validation checklist before starting
  fest understand rules          Naming conventions for automation
  fest understand workflow       Just-in-time reading pattern

The understand command helps you grasp WHEN and WHY to use specific
approaches. For command syntax, use --help on any command.

Content is pulled from your local .festival/ directory when available,
ensuring you see the current methodology design and any customizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			dotFestival := findDotFestivalDir()
			printOverview(dotFestival)
		},
	}

	// Add subcommands - ordered by importance
	// CRITICAL commands first
	cmd.AddCommand(newUnderstandTasksCmd())     // Most common mistake
	cmd.AddCommand(newUnderstandStructureCmd()) // Core hierarchy
	cmd.AddCommand(newUnderstandRulesCmd())     // Mandatory rules
	cmd.AddCommand(newUnderstandChecklistCmd()) // Quick validation

	// Learning commands
	cmd.AddCommand(newUnderstandMethodologyCmd())
	cmd.AddCommand(newUnderstandContextCmd())  // Session memory - CREATE FIRST
	cmd.AddCommand(newUnderstandNodeIDsCmd())  // Node reference system for traceability
	cmd.AddCommand(newUnderstandWorkflowCmd())
	cmd.AddCommand(newUnderstandTemplatesCmd())
	cmd.AddCommand(newUnderstandResourcesCmd())

	// Extension/plugin discovery
	cmd.AddCommand(newUnderstandGatesCmd())
	cmd.AddCommand(newUnderstandPluginsCmd())
	cmd.AddCommand(newUnderstandExtensionsCmd())

	return cmd
}

func printOverview(dotFestival string) {
	// Use embedded overview content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("overview.txt"))

	if dotFestival != "" {
		fmt.Printf("\nSource: %s\n", dotFestival)
	} else {
		fmt.Println("\nNote: No .festival/ directory found. Showing default content.")
		fmt.Println("      Run from a festivals/ tree to see your methodology resources.")
	}
}
