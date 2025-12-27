package structure

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	dryRun  bool
	backup  bool
	force   bool
	verbose bool
}

// NewRemoveCommand creates the remove command
func NewRemoveCommand() *cobra.Command {
	opts := &removeOptions{}

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove festival elements and renumber",
		Long: `Remove a phase, sequence, or task and automatically renumber subsequent elements.
		
This command safely removes elements and maintains proper numbering
for all following elements in the hierarchy.`,
	}

	// Add persistent flags
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", true, "preview changes without applying them")
	cmd.PersistentFlags().BoolVar(&opts.backup, "backup", false, "create backup before removal")
	cmd.PersistentFlags().BoolVar(&opts.force, "force", false, "skip confirmation prompts")
	cmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "show detailed output")

	// Add subcommands
	cmd.AddCommand(newRemovePhaseCommand(opts))
	cmd.AddCommand(newRemoveSequenceCommand(opts))
	cmd.AddCommand(newRemoveTaskCommand(opts))

	return cmd
}

// newRemovePhaseCommand creates the phase removal subcommand
func newRemovePhaseCommand(opts *removeOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "phase [phase-number|phase-path]",
		Short: "Remove a phase and renumber subsequent phases",
		Long: `Remove a phase by number or path and automatically renumber all following phases.
		
Warning: This will permanently delete the phase and all its contents!`,
		Example: `  fest remove phase 2                    # Remove phase 002
  fest remove phase 002_DEFINE_INTERFACES # Remove by directory name
  fest remove phase ./002_DEFINE          # Remove by path`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := args[0]

			// Determine if target is a number or path
			var targetPath string
			if num, err := parsePhaseNumber(target); err == nil {
				// Find phase by number
				parser := festival.NewParser()
				phases, err := parser.ParsePhases(cmd.Context(), ".")
				if err != nil {
					return errors.Wrap(err, "parsing phases").WithOp("removePhase")
				}

				found := false
				for _, phase := range phases {
					if phase.Number == num {
						targetPath = phase.Path
						found = true
						break
					}
				}

				if !found {
					return errors.NotFound("phase").WithField("number", num)
				}
			} else {
				// Use as path
				targetPath = target
			}

			// Convert to absolute path
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				return errors.Wrap(err, "resolving path").WithField("path", targetPath)
			}

			// Confirm removal if not forced
			if !opts.force && !opts.dryRun {
				fmt.Printf("Warning: This will permanently delete %s and all its contents!\n", filepath.Base(absPath))
				fmt.Print("Are you sure? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if !strings.HasPrefix(strings.ToLower(response), "y") {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform removal
			return renumberer.RemoveElement(cmd.Context(), absPath)
		},
	}
}

// newRemoveSequenceCommand creates the sequence removal subcommand
func newRemoveSequenceCommand(opts *removeOptions) *cobra.Command {
	var phaseFlag string

	cmd := &cobra.Command{
		Use:   "sequence [sequence-number|sequence-name]",
		Short: "Remove a sequence and renumber subsequent sequences",
		Long: `Remove a sequence by number or name and automatically renumber all following sequences.

Warning: This will permanently delete the sequence and all its contents!

If --phase is omitted and you're inside a phase directory, it will use the current phase.`,
		Example: `  fest remove sequence 2                   # Use current phase (if inside one)
  fest remove sequence --phase 1 2          # Numeric shortcut for phase 001_*
  fest remove sequence --phase 001_PLAN 2   # Remove sequence 02
  fest remove sequence --phase 001_PLAN 02_architecture`,
		Args: cobra.ExactArgs(1),
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

			target := args[0]

			// Determine target path
			var targetPath string
			if num, err := parseSequenceNumber(target); err == nil {
				// Find sequence by number
				parser := festival.NewParser()
				sequences, err := parser.ParseSequences(cmd.Context(), phaseDir)
				if err != nil {
					return errors.Wrap(err, "parsing sequences").WithOp("removeSequence")
				}

				found := false
				for _, seq := range sequences {
					if seq.Number == num {
						targetPath = seq.Path
						found = true
						break
					}
				}

				if !found {
					return errors.NotFound("sequence").WithField("number", num).WithField("phase", filepath.Base(phaseDir))
				}
			} else {
				// Use as name/path
				targetPath = filepath.Join(phaseDir, target)
			}

			// Confirm removal if not forced
			if !opts.force && !opts.dryRun {
				fmt.Printf("Warning: This will permanently delete %s and all its contents!\n", filepath.Base(targetPath))
				fmt.Print("Are you sure? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if !strings.HasPrefix(strings.ToLower(response), "y") {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform removal
			return renumberer.RemoveElement(cmd.Context(), targetPath)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")

	return cmd
}

// newRemoveTaskCommand creates the task removal subcommand
func newRemoveTaskCommand(opts *removeOptions) *cobra.Command {
	var phaseFlag string
	var sequenceFlag string

	cmd := &cobra.Command{
		Use:   "task [task-number|task-name]",
		Short: "Remove a task and renumber subsequent tasks",
		Long: `Remove a task by number or name and automatically renumber all following tasks.

Warning: This will permanently delete the task file!

If --sequence is omitted and you're inside a sequence directory, it will use the current sequence.`,
		Example: `  fest remove task 2                              # Use current sequence (if inside one)
  fest remove task --sequence 1 2                 # Numeric shortcut for sequence 01_*
  fest remove task --phase 1 --sequence 2 3       # Phase 001_*, sequence 02_*
  fest remove task --sequence ./path/to/seq 02_validate.md`,
		Args: cobra.ExactArgs(1),
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

			target := args[0]

			// Determine target path
			var targetPath string
			if num, err := parseTaskNumber(target); err == nil {
				// Find task by number
				parser := festival.NewParser()
				tasks, err := parser.ParseTasks(cmd.Context(), sequenceDir)
				if err != nil {
					return errors.Wrap(err, "parsing tasks").WithOp("removeTask")
				}

				// Handle potential parallel tasks
				var matches []festival.FestivalElement
				for _, task := range tasks {
					if task.Number == num {
						matches = append(matches, task)
					}
				}

				if len(matches) == 0 {
					return errors.NotFound("task").WithField("number", num).WithField("sequence", filepath.Base(sequenceDir))
				} else if len(matches) > 1 {
					// Multiple tasks with same number
					fmt.Println("Multiple tasks found with that number:")
					for i, task := range matches {
						fmt.Printf("  [%d] %s\n", i+1, task.FullName)
					}
					fmt.Print("Select task to remove (number): ")
					var choice int
					fmt.Scanln(&choice)
					if choice < 1 || choice > len(matches) {
						return errors.Validation("invalid selection")
					}
					targetPath = matches[choice-1].Path
				} else {
					targetPath = matches[0].Path
				}
			} else {
				// Use as name/path
				if !strings.HasSuffix(target, ".md") {
					target += ".md"
				}
				targetPath = filepath.Join(sequenceDir, target)
			}

			// Confirm removal if not forced
			if !opts.force && !opts.dryRun {
				fmt.Printf("Warning: This will permanently delete %s!\n", filepath.Base(targetPath))
				fmt.Print("Are you sure? [y/N]: ")
				var response string
				fmt.Scanln(&response)
				if !strings.HasPrefix(strings.ToLower(response), "y") {
					fmt.Println("Operation cancelled.")
					return nil
				}
			}

			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || shared.IsVerbose(),
			})

			// Perform removal
			return renumberer.RemoveElement(cmd.Context(), targetPath)
		},
	}

	cmd.Flags().StringVar(&phaseFlag, "phase", "", "phase directory (numeric shortcut, name, or path)")
	cmd.Flags().StringVar(&sequenceFlag, "sequence", "", "sequence directory (numeric shortcut, name, or path)")

	return cmd
}

// Helper functions to parse numbers
func parsePhaseNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, errors.Validation("not a number").WithField("input", s)
}

func parseSequenceNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, errors.Validation("not a number").WithField("input", s)
}

func parseTaskNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, errors.Validation("not a number").WithField("input", s)
}
