// Package wizard provides interactive guidance commands for the fest CLI.
package wizard

import "github.com/spf13/cobra"

// NewWizardCommand creates the parent wizard command.
func NewWizardCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "wizard",
		Short: "Interactive guidance and assistance for festival creation",
		Long: `The wizard command provides interactive guidance for festival creation.

SUBCOMMANDS:
  fill <file>    Interactively fill REPLACE markers in a file

EXAMPLES:
  # Fill markers in a specific file
  fest wizard fill PHASE_GOAL.md

  # Fill markers in all files in current directory
  fest wizard fill .

  # Preview changes without writing
  fest wizard fill FESTIVAL_GOAL.md --dry-run

The wizard subcommands help automate tedious tasks and guide users
through the festival creation process step by step.`,
	}

	// Add subcommands
	cmd.AddCommand(NewFillCommand())

	return cmd
}
