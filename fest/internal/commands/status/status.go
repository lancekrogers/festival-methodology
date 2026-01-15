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
	json        bool
	entityType  string
	force       bool
	path        string
	interactive bool // force interactive selection

	// Level targeting flags for status set
	phase    string // --phase flag: target phase by name/number
	sequence string // --sequence flag: target sequence by name/number
	task     string // --task flag: target task by filename/path
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

CONTEXT-AWARE BEHAVIOR:
When no explicit level flag is provided, the command auto-detects the
appropriate level based on your current directory:

  Festival root  → Sets festival status (planned/active/completed/dungeon)
  Phase directory → Sets phase status (pending/in_progress/completed)
  Sequence directory → Sets sequence status (pending/in_progress/completed)
  Task directory → Shows hint (task status requires explicit --task flag)

For festivals, this will move the directory between status folders.
If not inside a festival, an interactive selector will be shown.

EXPLICIT TARGETING:
Use flags to override auto-detection:
  --phase    Target a specific phase
  --sequence Target a specific sequence
  --task     Target a specific task
  --path     Target by explicit file path

These flags are mutually exclusive - only one level can be targeted at a time.`,
		Example: `  fest status set active               # Set current festival to active
  fest status set active -i            # Force interactive selection
  fest status set completed --force    # Set without confirmation
  fest status set in_progress          # Set phase/sequence/task status

  # Level-specific status setting:
  fest status set --phase 001_CRITICAL completed
  fest status set --phase 001 in_progress
  fest status set --sequence 01_api_design completed
  fest status set --sequence 002/01 pending
  fest status set --task 01_analyze.md in_progress
  fest status set --task 001/01/02_impl.md completed
  fest status set --path ./002/01/task.md blocked`,
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
	cmd.Flags().BoolVarP(&opts.interactive, "interactive", "i", false, "force interactive festival selection")

	// Level targeting flags
	cmd.Flags().StringVar(&opts.phase, "phase", "", "target phase by name or number (e.g., '001_CRITICAL' or '001')")
	cmd.Flags().StringVar(&opts.sequence, "sequence", "", "target sequence by name (e.g., '01_api_design' or '002/01')")
	cmd.Flags().StringVar(&opts.task, "task", "", "target task by filename or path (e.g., '01_analyze.md' or '001/01/02_impl.md')")
	cmd.Flags().StringVar(&opts.path, "path", "", "explicit file path for status change")

	// Mark level flags as mutually exclusive
	cmd.MarkFlagsMutuallyExclusive("phase", "sequence", "task", "path")

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
