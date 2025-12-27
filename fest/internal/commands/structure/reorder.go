package structure

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/spf13/cobra"
)

type reorderOptions struct {
	dryRun  bool
	backup  bool
	force   bool
	verbose bool
}

// NewReorderCommand creates the reorder command
func NewReorderCommand() *cobra.Command {
	opts := &reorderOptions{}

	cmd := &cobra.Command{
		Use:   "reorder",
		Short: "Reorder festival elements",
		Long: `Reorder phases, sequences, or tasks by moving an element from one position to another.

This command moves an element to a new position and shifts other elements
accordingly to maintain proper ordering.`,
	}

	// Add persistent flags for all subcommands
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", true, "preview changes without applying them")
	cmd.PersistentFlags().BoolVar(&opts.backup, "backup", false, "create backup before reordering")
	cmd.PersistentFlags().BoolVar(&opts.force, "force", false, "skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "show detailed output")

	// Allow skipping dry-run mode entirely
	var skipDryRun bool
	cmd.PersistentFlags().BoolVar(&skipDryRun, "skip-dry-run", false, "skip preview and apply changes immediately")

	// Override dryRun if skip-dry-run is set
	cmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if skipDryRun {
			opts.dryRun = false
		}
	}

	// Add subcommands
	cmd.AddCommand(newReorderPhaseCommand(opts))
	cmd.AddCommand(newReorderSequenceCommand(opts))
	cmd.AddCommand(newReorderTaskCommand(opts))

	return cmd
}

// newReorderPhaseCommand creates the phase reordering subcommand
func newReorderPhaseCommand(opts *reorderOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "phase <from> <to> [festival-dir]",
		Short: "Reorder phases in a festival",
		Long: `Move a phase from one position to another within a festival.

Elements between the source and destination positions are shifted accordingly.
For example, moving phase 3 to position 1 will shift phases 1 and 2 down.`,
		Example: `  fest reorder phase 3 1                    # Move phase 003 to position 001 (dry-run preview)
  fest reorder phase 1 3 ./my-festival      # Move phase 001 to position 003
  fest reorder phase 4 2 --skip-dry-run     # Apply immediately without preview`,
		Args: cobra.RangeArgs(2, 3),
		RunE: func(cmd *cobra.Command, args []string) error {
			from, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.Validation("invalid source position").WithField("position", args[0])
			}

			to, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Validation("invalid destination position").WithField("position", args[1])
			}

			festivalDir := "."
			if len(args) > 2 {
				festivalDir = args[2]
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(festivalDir)
			if err != nil {
				return errors.Wrap(err, "resolving path").WithField("path", festivalDir)
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:      opts.dryRun,
				Backup:      opts.backup,
				Verbose:     opts.verbose || shared.IsVerbose(),
				AutoApprove: opts.force,
			})

			// Perform reordering
			return renumberer.ReorderPhase(cmd.Context(), absPath, from, to)
		},
	}
}

// newReorderSequenceCommand creates the sequence reordering subcommand
func newReorderSequenceCommand(opts *reorderOptions) *cobra.Command {
	var phaseFlag string

	cmd := &cobra.Command{
		Use:   "sequence <from> <to>",
		Short: "Reorder sequences within a phase",
		Long: `Move a sequence from one position to another within a phase.

Elements between the source and destination positions are shifted accordingly.

If --phase is omitted and you're inside a phase directory, it will use the current phase.`,
		Example: `  fest reorder sequence 3 1                            # Use current phase (if inside one)
  fest reorder sequence --phase 1 3 1                  # Numeric shortcut for phase 001_*
  fest reorder sequence --phase 001_PLAN 3 1           # Move sequence 03 to position 01
  fest reorder sequence --phase ./003_IMPLEMENT 1 4    # Move sequence 01 to position 04`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return errors.IO("getting current directory", err)
			}

			// Detect context
			ctx, err := festival.DetectContext(cwd)
			if err != nil {
				return errors.Wrap(err, "detecting context")
			}

			from, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.Validation("invalid source position").WithField("position", args[0])
			}

			to, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Validation("invalid destination position").WithField("position", args[1])
			}

			var phaseDir string
			if phaseFlag != "" {
				if ctx.FestivalDir == "" {
					return errors.Validation("not inside a festival directory")
				}
				phaseDir, err = festival.ResolvePhase(phaseFlag, ctx.FestivalDir)
				if err != nil {
					return err
				}
			} else if ctx.PhaseDir != "" {
				phaseDir = ctx.PhaseDir
			} else {
				return errors.Validation("--phase required (not inside a phase directory)")
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:      opts.dryRun,
				Backup:      opts.backup,
				Verbose:     opts.verbose || shared.IsVerbose(),
				AutoApprove: opts.force,
			})

			// Perform reordering
			return renumberer.ReorderSequence(cmd.Context(), phaseDir, from, to)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")

	return cmd
}

// newReorderTaskCommand creates the task reordering subcommand
func newReorderTaskCommand(opts *reorderOptions) *cobra.Command {
	var phaseFlag string
	var sequenceFlag string

	cmd := &cobra.Command{
		Use:   "task <from> <to>",
		Short: "Reorder tasks within a sequence",
		Long: `Move a task from one position to another within a sequence.

Elements between the source and destination positions are shifted accordingly.
Parallel tasks (multiple tasks with the same number) are moved together.

If --sequence is omitted and you're inside a sequence directory, it will use the current sequence.`,
		Example: `  fest reorder task 3 1                               # Use current sequence (if inside one)
  fest reorder task --sequence 1 3 1                  # Numeric shortcut for sequence 01_*
  fest reorder task --phase 1 --sequence 2 3 1        # Phase 001_*, sequence 02_*
  fest reorder task --sequence ./path/to/sequence 1 5`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return errors.IO("getting current directory", err)
			}

			// Detect context
			ctx, err := festival.DetectContext(cwd)
			if err != nil {
				return errors.Wrap(err, "detecting context")
			}

			from, err := strconv.Atoi(args[0])
			if err != nil {
				return errors.Validation("invalid source position").WithField("position", args[0])
			}

			to, err := strconv.Atoi(args[1])
			if err != nil {
				return errors.Validation("invalid destination position").WithField("position", args[1])
			}

			var sequenceDir string
			if sequenceFlag != "" {
				var phaseDir string
				if phaseFlag != "" {
					if ctx.FestivalDir == "" {
						return errors.Validation("not inside a festival directory")
					}
					phaseDir, err = festival.ResolvePhase(phaseFlag, ctx.FestivalDir)
					if err != nil {
						return err
					}
				} else if ctx.PhaseDir != "" {
					phaseDir = ctx.PhaseDir
				} else {
					return errors.Validation("--phase required when using --sequence with numeric shortcut")
				}

				sequenceDir, err = festival.ResolveSequence(sequenceFlag, phaseDir)
				if err != nil {
					return err
				}
			} else if ctx.SequenceDir != "" {
				sequenceDir = ctx.SequenceDir
			} else {
				return errors.Validation("--sequence required (not inside a sequence directory)")
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:      opts.dryRun,
				Backup:      opts.backup,
				Verbose:     opts.verbose || shared.IsVerbose(),
				AutoApprove: opts.force,
			})

			// Perform reordering
			return renumberer.ReorderTask(cmd.Context(), sequenceDir, from, to)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")
	cmd.Flags().StringVar(&sequenceFlag, "sequence", "", "sequence directory (numeric shortcut, name, or path)")

	return cmd
}
