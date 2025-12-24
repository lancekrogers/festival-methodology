package commands

import (
	"github.com/spf13/cobra"
)

// NewTaskCommand creates the "task" parent command
func NewTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task management commands",
		Long:  `Commands for managing tasks including quality gate defaults.`,
	}

	// Add subcommands
	cmd.AddCommand(NewTaskDefaultsCommand())

	return cmd
}

// NewTaskDefaultsCommand creates the "defaults" subcommand
func NewTaskDefaultsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defaults",
		Short: "Manage quality gate default tasks",
		Long: `Manage quality gate default tasks for festivals.

Quality gates are standard tasks (testing, code review, iterate) that should
appear at the end of every implementation sequence.

Use fest.yaml in your festival root to customize which quality gates are included.`,
	}

	// Add subcommands
	cmd.AddCommand(NewTaskDefaultsSyncCommand())
	cmd.AddCommand(NewTaskDefaultsAddCommand())
	cmd.AddCommand(NewTaskDefaultsShowCommand())
	cmd.AddCommand(NewTaskDefaultsInitCommand())

	return cmd
}
