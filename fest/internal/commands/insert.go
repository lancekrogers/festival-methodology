package commands

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/festival-methodology/fest/internal/festival"
	"github.com/spf13/cobra"
)

type insertOptions struct {
	after   int
	name    string
	dryRun  bool
	backup  bool
	verbose bool
}

// NewInsertCommand creates the insert command
func NewInsertCommand() *cobra.Command {
	opts := &insertOptions{}
	
	cmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert new festival elements",
		Long: `Insert a new phase, sequence, or task and renumber subsequent elements.
		
This command creates a new element and automatically renumbers all
following elements to maintain proper ordering.`,
	}
	
	// Add persistent flags
	cmd.PersistentFlags().BoolVar(&opts.dryRun, "dry-run", true, "preview changes without applying them")
	cmd.PersistentFlags().BoolVar(&opts.backup, "backup", false, "create backup before changes")
	cmd.PersistentFlags().BoolVar(&opts.verbose, "verbose", false, "show detailed output")
	
	// Add subcommands
	cmd.AddCommand(newInsertPhaseCommand(opts))
	cmd.AddCommand(newInsertSequenceCommand(opts))
	cmd.AddCommand(newInsertTaskCommand(opts))
	
	return cmd
}

// newInsertPhaseCommand creates the phase insertion subcommand
func newInsertPhaseCommand(opts *insertOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phase [festival-dir]",
		Short: "Insert a new phase",
		Long: `Insert a new phase after the specified number and renumber subsequent phases.
		
The new phase will be created with the proper 3-digit numbering format.`,
		Example: `  fest insert phase --after 1 --name "DESIGN_REVIEW"
  fest insert phase ./my-festival --after 2 --name "TESTING"
  fest insert phase --after 0 --name "REQUIREMENTS" # Insert at beginning`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.name == "" {
				return fmt.Errorf("--name flag is required")
			}
			
			festivalDir := "."
			if len(args) > 0 {
				festivalDir = args[0]
			}
			
			// Convert to absolute path
			absPath, err := filepath.Abs(festivalDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			
			// Validate name (no spaces, special chars)
			if strings.Contains(opts.name, " ") {
				opts.name = strings.ReplaceAll(opts.name, " ", "_")
			}
			
			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || verbose,
			})
			
			// Perform insertion
			return renumberer.InsertPhase(absPath, opts.after, opts.name)
		},
	}
	
	cmd.Flags().IntVar(&opts.after, "after", 0, "insert after this phase number (0 for beginning)")
	cmd.Flags().StringVar(&opts.name, "name", "", "name of the new phase")
	cmd.MarkFlagRequired("name")
	
	return cmd
}

// newInsertSequenceCommand creates the sequence insertion subcommand
func newInsertSequenceCommand(opts *insertOptions) *cobra.Command {
	var phaseDir string
	
	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Insert a new sequence within a phase",
		Long: `Insert a new sequence after the specified number and renumber subsequent sequences.
		
The new sequence will be created with the proper 2-digit numbering format.`,
		Example: `  fest insert sequence --phase 001_PLAN --after 1 --name "validation"
  fest insert sequence --phase ./003_IMPLEMENT --after 0 --name "setup"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.name == "" {
				return fmt.Errorf("--name flag is required")
			}
			if phaseDir == "" {
				return fmt.Errorf("--phase flag is required")
			}
			
			// Convert to absolute path
			absPath, err := filepath.Abs(phaseDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			
			// Validate name
			if strings.Contains(opts.name, " ") {
				opts.name = strings.ReplaceAll(opts.name, " ", "_")
			}
			
			// Create renumberer
			renumberer := festival.NewRenumberer(festival.RenumberOptions{
				DryRun:  opts.dryRun,
				Backup:  opts.backup,
				Verbose: opts.verbose || verbose,
			})
			
			// Perform insertion
			return renumberer.InsertSequence(absPath, opts.after, opts.name)
		},
	}
	
	cmd.Flags().StringVar(&phaseDir, "phase", "", "phase directory to insert sequence in")
	cmd.Flags().IntVar(&opts.after, "after", 0, "insert after this sequence number (0 for beginning)")
	cmd.Flags().StringVar(&opts.name, "name", "", "name of the new sequence")
	cmd.MarkFlagRequired("phase")
	cmd.MarkFlagRequired("name")
	
	return cmd
}

// newInsertTaskCommand creates the task insertion subcommand
func newInsertTaskCommand(opts *insertOptions) *cobra.Command {
	var sequenceDir string
	
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Insert a new task within a sequence",
		Long: `Insert a new task after the specified number and renumber subsequent tasks.
		
The new task will be created with the proper 2-digit numbering format.
Note: Tasks are markdown files, so .md extension will be added automatically.`,
		Example: `  fest insert task --sequence 001_PLAN/01_requirements --after 1 --name "validate_input"
  fest insert task --sequence ./path/to/sequence --after 0 --name "setup"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.name == "" {
				return fmt.Errorf("--name flag is required")
			}
			if sequenceDir == "" {
				return fmt.Errorf("--sequence flag is required")
			}
			
			// Convert to absolute path
			absPath, err := filepath.Abs(sequenceDir)
			if err != nil {
				return fmt.Errorf("failed to resolve path: %w", err)
			}
			
			// Validate and clean name
			opts.name = strings.TrimSuffix(opts.name, ".md") // Remove .md if provided
			if strings.Contains(opts.name, " ") {
				opts.name = strings.ReplaceAll(opts.name, " ", "_")
			}
			
			// For tasks, we need to handle file creation differently
			// This would be implemented in the renumberer
			fmt.Printf("Note: Task insertion would create: %s_%s.md\n", 
				festival.FormatNumber(opts.after+1, festival.TaskType), opts.name)
			
			// For now, return a message since task insertion needs file creation
			if !opts.dryRun {
				return fmt.Errorf("task insertion not yet fully implemented")
			}
			
			// Show what would happen
			parser := festival.NewParser()
			tasks, err := parser.ParseTasks(absPath)
			if err != nil {
				return fmt.Errorf("failed to parse tasks: %w", err)
			}
			
			fmt.Println("\nTask insertion preview:")
			fmt.Printf("  ✓ Create: %02d_%s.md\n", opts.after+1, opts.name)
			for _, task := range tasks {
				if task.Number > opts.after {
					fmt.Printf("  → Rename: %s → %02d_%s\n", 
						task.FullName, task.Number+1, task.Name+".md")
				}
			}
			
			return nil
		},
	}
	
	cmd.Flags().StringVar(&sequenceDir, "sequence", "", "sequence directory to insert task in")
	cmd.Flags().IntVar(&opts.after, "after", 0, "insert after this task number (0 for beginning)")
	cmd.Flags().StringVar(&opts.name, "name", "", "name of the new task (without .md extension)")
	cmd.MarkFlagRequired("sequence")
	cmd.MarkFlagRequired("name")
	
	return cmd
}