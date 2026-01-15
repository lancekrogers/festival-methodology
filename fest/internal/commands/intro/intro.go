// Package intro provides the fest intro command for getting started with fest CLI.
package intro

import (
	"fmt"

	introdocs "github.com/lancekrogers/festival-methodology/fest/docs/intro"
	"github.com/spf13/cobra"
)

// NewIntroCommand creates the intro command for getting started.
func NewIntroCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "intro",
		Short: "Getting started guide for fest CLI and common workflows",
		Long: `Display a getting started guide for AI agents using the fest CLI.

This command provides essential information for quickly becoming productive
with Festival Methodology and the fest CLI commands.

SUBCOMMANDS:
  fest intro             Show the getting started guide
  fest intro workflows   Show common workflow patterns

After reading the intro, explore deeper with:
  fest understand methodology    Core principles and philosophy
  fest understand structure      3-level hierarchy explained
  fest understand tasks          When to create task files`,
		Run: func(cmd *cobra.Command, args []string) {
			printIntro()
		},
	}

	cmd.AddCommand(newWorkflowsCommand())

	return cmd
}

func newWorkflowsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "workflows",
		Short: "Common fest workflow patterns",
		Long: `Display common workflow patterns for using fest CLI effectively.

Covers workflows for:
  - Starting a new project
  - Executing existing festivals
  - Adding structure to festivals
  - Checking and fixing structure
  - Navigation shortcuts
  - Machine-readable output`,
		Run: func(cmd *cobra.Command, args []string) {
			printWorkflows()
		},
	}
}

func printIntro() {
	fmt.Print(introdocs.Load("intro.txt"))
}

func printWorkflows() {
	fmt.Print(introdocs.Load("workflows.txt"))
}
