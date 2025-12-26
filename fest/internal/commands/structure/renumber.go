package structure

import (
	"fmt"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
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
				return fmt.Errorf("failed to resolve path: %w", err)
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
	var phaseDir string

	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Renumber sequences within a phase",
		Long: `Renumber all sequences in a phase starting from the specified number (default: 1).
		
Sequences are numbered with 2 digits (01, 02, 03, etc.).`,
		Example: `  fest renumber sequence --phase 001_PLAN           # Renumber sequences in phase
  fest renumber sequence --phase ./003_IMPLEMENT    # Use path to phase
  fest renumber sequence --phase 001_PLAN --start 2 # Start from 02`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if phaseDir == "" {
				return fmt.Errorf("--phase flag is required")
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(phaseDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform renumbering
			return renumberer.RenumberSequences(cmd.Context(), absPath, opts.startFrom)
		},
	}

	cmd.Flags().StringVar(&phaseDir, "phase", "", "phase directory to renumber sequences in")
	cmd.MarkFlagRequired("phase")

	return cmd
}

// newRenumberTaskCommand creates the task renumbering subcommand
func newRenumberTaskCommand(opts *renumberOptions) *cobra.Command {
	var sequenceDir string

	cmd := &cobra.Command{
		Use:   "task",
		Short: "Renumber tasks within a sequence",
		Long: `Renumber all tasks in a sequence starting from the specified number (default: 1).
		
Tasks are numbered with 2 digits (01, 02, 03, etc.).
Parallel tasks (multiple tasks with the same number) are preserved.`,
		Example: `  fest renumber task --sequence 001_PLAN/01_requirements
  fest renumber task --sequence ./path/to/sequence --start 2`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if sequenceDir == "" {
				return fmt.Errorf("--sequence flag is required")
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(sequenceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform renumbering
			return renumberer.RenumberTasks(cmd.Context(), absPath, opts.startFrom)
		},
	}

	cmd.Flags().StringVar(&sequenceDir, "sequence", "", "sequence directory to renumber tasks in")
	cmd.MarkFlagRequired("sequence")

	return cmd
}
