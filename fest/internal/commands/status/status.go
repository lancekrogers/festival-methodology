// Package status implements the fest status command for managing entity statuses.
package status

import (
	"context"

	"github.com/spf13/cobra"
)

// EntityType represents the type of festival entity.
type EntityType string

const (
	EntityFestival EntityType = "festival"
	EntityPhase    EntityType = "phase"
	EntitySequence EntityType = "sequence"
	EntityTask     EntityType = "task"
	EntityGate     EntityType = "gate"
)

// ValidStatuses defines valid status values per entity type.
var ValidStatuses = map[EntityType][]string{
	EntityFestival: {"planned", "active", "completed", "dungeon"},
	EntityPhase:    {"pending", "in_progress", "completed"},
	EntitySequence: {"pending", "in_progress", "completed"},
	EntityTask:     {"pending", "in_progress", "blocked", "completed"},
	EntityGate:     {"pending", "passed", "failed"},
}

// statusOptions holds options shared across status subcommands.
type statusOptions struct {
	json       bool
	entityType string
	force      bool
	path       string
}

// NewStatusCommand creates the status command with all subcommands.
func NewStatusCommand() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status [path]",
		Short: "Manage and query festival entity statuses",
		Long: `Manage and query status for festivals, phases, sequences, tasks, and gates.

When run without arguments, shows the status of the current entity based on
your working directory location.

EXAMPLES:
  fest status                                  # Status from current directory
  fest status ./festivals/active/my-festival   # Status for specific path
  fest status active/my-festival               # Relative to festivals/ root

SUBCOMMANDS:
  fest status              Show current entity status
  fest status set <status> Change entity status
  fest status list         List entities by status
  fest status history      View status change history`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runStatusShow(ctx, cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "", "entity type to query (festival, phase, sequence, task, gate)")

	// Add subcommands
	cmd.AddCommand(newStatusSetCommand(opts))
	cmd.AddCommand(newStatusListCommand(opts))
	cmd.AddCommand(newStatusHistoryCommand(opts))

	return cmd
}

// newStatusSetCommand creates the status set subcommand.
func newStatusSetCommand(opts *statusOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <status>",
		Short: "Change entity status",
		Long: `Change the status of the current entity.

For festivals, this will move the directory between status folders
(planned, active, completed, dungeon).

For other entities, this updates the frontmatter in the relevant files.`,
		Example: `  fest status set active               # Set current festival to active
  fest status set completed --force    # Set without confirmation
  fest status set in_progress          # Set phase/sequence/task status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runStatusSet(ctx, cmd, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

// newStatusListCommand creates the status list subcommand.
func newStatusListCommand(opts *statusOptions) *cobra.Command {
	var filterStatus string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entities by status",
		Long: `List festivals, phases, sequences, or tasks filtered by status.

Without filters, lists all festivals grouped by status.`,
		Example: `  fest status list                     # List all festivals by status
  fest status list --status active     # List active festivals only
  fest status list --type task --status pending  # List pending tasks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runStatusList(ctx, cmd, filterStatus, opts)
		},
	}

	cmd.Flags().StringVar(&filterStatus, "status", "", "filter by status")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "festival", "entity type (festival, phase, sequence, task)")

	return cmd
}

// newStatusHistoryCommand creates the status history subcommand.
func newStatusHistoryCommand(opts *statusOptions) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "View status change history",
		Long: `View the history of status changes for the current entity.

History is stored in .fest/status_history.json within each festival.`,
		Example: `  fest status history            # Show status history
  fest status history --limit 10 # Show last 10 changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runStatusHistory(ctx, cmd, limit, opts)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of entries to show")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}
