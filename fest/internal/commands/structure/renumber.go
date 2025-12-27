package structure

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/spf13/cobra"
)

type renumberOptions struct {
	dryRun    bool
	backup    bool
	startFrom int
	verbose   bool
}

// NewRenumberCommand creates the renumber command
func NewRenumberCommand() *cobra.Command {
	opts := &renumberOptions{}

	cmd := &cobra.Command{
		Use:   "renumber",
		Short: "Renumber festival elements",
		Long: `Renumber phases, sequences, or tasks in a festival structure.
		
This command helps maintain proper numbering when elements are added,
removed, or reordered in the festival hierarchy.`,
	}

	// Add persistent flags for all subcommands
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", true, "preview changes without applying them")
	cmd.PersistentFlags().BoolVar(&opts.backup, "backup", false, "create backup before renumbering")
	cmd.PersistentFlags().IntVar(&opts.startFrom, "start", 1, "starting number for renumbering")
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
	cmd.AddCommand(newRenumberPhaseCommand(opts))
	cmd.AddCommand(newRenumberSequenceCommand(opts))
	cmd.AddCommand(newRenumberTaskCommand(opts))

	return cmd
}

// newRenumberPhaseCommand creates the phase renumbering subcommand
func newRenumberPhaseCommand(opts *renumberOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "phase [festival-dir]",
		Short: "Renumber phases in a festival",
		Long: `Renumber all phases starting from the specified number (default: 1).
		
Phases are numbered with 3 digits (001, 002, 003, etc.).`,
		Example: `  fest renumber phase                    # Renumber phases in current directory (dry-run preview)
  fest renumber phase ./my-festival      # Renumber phases in specified directory
  fest renumber phase --start 2          # Start numbering from 002
  fest renumber phase --skip-dry-run     # Skip preview and apply changes immediately`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalDir := "."
			if len(args) > 0 {
				festivalDir = args[0]
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(festivalDir)
			if err != nil {
				return errors.Wrap(err, "resolving path").WithField("path", festivalDir)
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform renumbering
			return renumberer.RenumberPhases(cmd.Context(), absPath, opts.startFrom)
		},
	}
}

// newRenumberSequenceCommand creates the sequence renumbering subcommand
func newRenumberSequenceCommand(opts *renumberOptions) *cobra.Command {
	var phaseFlag string

	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Renumber sequences within a phase",
		Long: `Renumber all sequences in a phase starting from the specified number (default: 1).

Sequences are numbered with 2 digits (01, 02, 03, etc.).

If --phase is omitted and you're inside a phase directory, it will use the current phase.`,
		Example: `  fest renumber sequence                            # Use current phase (if inside one)
  fest renumber sequence --phase 1                  # Numeric shortcut for phase 001_*
  fest renumber sequence --phase 001_PLAN           # Renumber sequences in phase
  fest renumber sequence --phase ./003_IMPLEMENT    # Use path to phase
  fest renumber sequence --phase 001_PLAN --start 2 # Start from 02`,
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

			var phaseDir string
			if phaseFlag != "" {
				// Resolve phase flag (supports numeric shortcuts)
				if ctx.FestivalDir == "" {
					return errors.Validation("not inside a festival directory")
				}
				phaseDir, err = festival.ResolvePhase(phaseFlag, ctx.FestivalDir)
				if err != nil {
					return err
				}
			} else if ctx.PhaseDir != "" {
				// Use current phase from context
				phaseDir = ctx.PhaseDir
			} else {
				return errors.Validation("--phase required (not inside a phase directory)")
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform renumbering
			return renumberer.RenumberSequences(cmd.Context(), phaseDir, opts.startFrom)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")

	return cmd
}

// newRenumberTaskCommand creates the task renumbering subcommand
func newRenumberTaskCommand(opts *renumberOptions) *cobra.Command {
	var phaseFlag string
	var sequenceFlag string

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Renumber tasks within a sequence",
		Long: `Renumber all tasks in a sequence starting from the specified number (default: 1).

Tasks are numbered with 2 digits (01, 02, 03, etc.).
Parallel tasks (multiple tasks with the same number) are preserved.

If --sequence is omitted and you're inside a sequence directory, it will use the current sequence.
Use --phase to specify the phase when using numeric sequence shortcuts.`,
		Example: `  fest renumber task                              # Use current sequence (if inside one)
  fest renumber task --sequence 1                 # Numeric shortcut for sequence 01_*
  fest renumber task --phase 1 --sequence 2       # Phase 001_*, sequence 02_*
  fest renumber task --sequence 01_requirements   # Use sequence name
  fest renumber task --sequence ./path/to/seq     # Use path to sequence`,
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

			var sequenceDir string
			if sequenceFlag != "" {
				// Need to resolve sequence, which requires phase
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
				// Use current sequence from context
				sequenceDir = ctx.SequenceDir
			} else {
				return errors.Validation("--sequence required (not inside a sequence directory)")
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform renumbering
			return renumberer.RenumberTasks(cmd.Context(), sequenceDir, opts.startFrom)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")
	cmd.Flags().StringVar(&sequenceFlag, "sequence", "", "sequence directory (numeric shortcut, name, or path)")

	return cmd
}
