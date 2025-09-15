package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/festival-methodology/fest/internal/festival"
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
				phases, err := parser.ParsePhases(".")
				if err != nil {
					return fmt.Errorf("failed to parse phases: %w", err)
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
					return fmt.Errorf("phase %03d not found", num)
				}
			} else {
				// Use as path
				targetPath = target
			}
			
			// Convert to absolute path
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
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
				Verbose: opts.verbose || verbose,
			})
			
			// Perform removal
			return renumberer.RemoveElement(absPath)
		},
	}
}

// newRemoveSequenceCommand creates the sequence removal subcommand
func newRemoveSequenceCommand(opts *removeOptions) *cobra.Command {
	var phaseDir string
	
	cmd := &cobra.Command{
		Use:   "sequence [sequence-number|sequence-name]",
		Short: "Remove a sequence and renumber subsequent sequences",
		Long: `Remove a sequence by number or name and automatically renumber all following sequences.
		
Warning: This will permanently delete the sequence and all its contents!`,
		Example: `  fest remove sequence --phase 001_PLAN 2  # Remove sequence 02
  fest remove sequence --phase 001_PLAN 02_architecture`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if phaseDir == "" {
				return fmt.Errorf("--phase flag is required")
			}
			
			target := args[0]
			
			// Convert phase to absolute path
			phaseAbs, err := filepath.Abs(phaseDir)
			if err != nil {
				return fmt.Errorf("failed to resolve phase path: %w", err)
			}
			
			// Determine target path
			var targetPath string
			if num, err := parseSequenceNumber(target); err == nil {
				// Find sequence by number
				parser := festival.NewParser()
				sequences, err := parser.ParseSequences(phaseAbs)
				if err != nil {
					return fmt.Errorf("failed to parse sequences: %w", err)
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
					return fmt.Errorf("sequence %02d not found in %s", num, phaseDir)
				}
			} else {
				// Use as name/path
				targetPath = filepath.Join(phaseAbs, target)
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
				Verbose: opts.verbose || verbose,
			})
			
			// Perform removal
			return renumberer.RemoveElement(targetPath)
		},
	}
	
	cmd.Flags().StringVar(&phaseDir, "phase", "", "phase containing the sequence")
	cmd.MarkFlagRequired("phase")
	
	return cmd
}

// newRemoveTaskCommand creates the task removal subcommand
func newRemoveTaskCommand(opts *removeOptions) *cobra.Command {
	var sequenceDir string
	
	cmd := &cobra.Command{
		Use:   "task [task-number|task-name]",
		Short: "Remove a task and renumber subsequent tasks",
		Long: `Remove a task by number or name and automatically renumber all following tasks.
		
Warning: This will permanently delete the task file!`,
		Example: `  fest remove task --sequence 001_PLAN/01_requirements 2
  fest remove task --sequence ./path/to/sequence 02_validate.md`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if sequenceDir == "" {
				return fmt.Errorf("--sequence flag is required")
			}
			
			target := args[0]
			
			// Convert sequence to absolute path
			seqAbs, err := filepath.Abs(sequenceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve sequence path: %w", err)
			}
			
			// Determine target path
			var targetPath string
			if num, err := parseTaskNumber(target); err == nil {
				// Find task by number
				parser := festival.NewParser()
				tasks, err := parser.ParseTasks(seqAbs)
				if err != nil {
					return fmt.Errorf("failed to parse tasks: %w", err)
				}
				
				// Handle potential parallel tasks
				var matches []festival.FestivalElement
				for _, task := range tasks {
					if task.Number == num {
						matches = append(matches, task)
					}
				}
				
				if len(matches) == 0 {
					return fmt.Errorf("task %02d not found in %s", num, sequenceDir)
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
						return fmt.Errorf("invalid selection")
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
				targetPath = filepath.Join(seqAbs, target)
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
				Verbose: opts.verbose || verbose,
			})
			
			// Perform removal
			return renumberer.RemoveElement(targetPath)
		},
	}
	
	cmd.Flags().StringVar(&sequenceDir, "sequence", "", "sequence containing the task")
	cmd.MarkFlagRequired("sequence")
	
	return cmd
}

// Helper functions to parse numbers
func parsePhaseNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, fmt.Errorf("not a number")
}

func parseSequenceNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, fmt.Errorf("not a number")
}

func parseTaskNumber(s string) (int, error) {
	var num int
	if _, err := fmt.Sscanf(s, "%d", &num); err == nil {
		return num, nil
	}
	return 0, fmt.Errorf("not a number")
}