package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/index"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

// NewIndexCommand creates the index command group
func NewIndexCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "index",
		Short: "Manage festival indices",
		Long: `Generate and validate festival indices for Guild integration.

The index file (.festival/index.json) provides a machine-readable representation
of the festival structure, including phases, sequences, and tasks.`,
	}

	cmd.AddCommand(newIndexWriteCommand())
	cmd.AddCommand(newIndexValidateCommand())
	cmd.AddCommand(newIndexShowCommand())

	return cmd
}

func newIndexWriteCommand() *cobra.Command {
	var outputPath string
	var prettyPrint bool

	cmd := &cobra.Command{
		Use:   "write [festival-path]",
		Short: "Generate festival index",
		Long: `Generate a festival index from the filesystem structure.

The index is written to .festival/index.json within the festival directory.
Use --output to write to a different location.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			writer := index.NewIndexWriter(festivalRoot)
			idx, err := writer.Generate()
			if err != nil {
				return fmt.Errorf("failed to generate index: %w", err)
			}

			// Determine output path
			if outputPath == "" {
				outputPath = filepath.Join(festivalRoot, ".festival", index.IndexFileName)
			}

			// Ensure parent directory exists
			if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
				return fmt.Errorf("failed to create output directory: %w", err)
			}

			// Save the index
			if err := idx.Save(outputPath); err != nil {
				return fmt.Errorf("failed to save index: %w", err)
			}

			summary := idx.Summary()
			fmt.Printf("Index written to: %s\n", outputPath)
			fmt.Printf("  Festival: %s\n", idx.FestivalID)
			fmt.Printf("  Phases: %d\n", summary.PhaseCount)
			fmt.Printf("  Sequences: %d\n", summary.SequenceCount)
			fmt.Printf("  Tasks: %d\n", summary.TaskCount)
			if summary.ManagedCount > 0 {
				fmt.Printf("  Managed gates: %d\n", summary.ManagedCount)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Output path (default: .festival/index.json)")
	cmd.Flags().BoolVar(&prettyPrint, "pretty", true, "Pretty print JSON output")

	return cmd
}

func newIndexValidateCommand() *cobra.Command {
	var indexPath string

	cmd := &cobra.Command{
		Use:   "validate [festival-path]",
		Short: "Validate festival index against filesystem",
		Long: `Validate that the festival index matches the actual filesystem structure.

Reports:
- Entries in index that don't exist on disk (missing)
- Files on disk that aren't in the index (extra)
- Missing goal files (warnings)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			// Determine index path
			if indexPath == "" {
				indexPath = filepath.Join(festivalRoot, ".festival", index.IndexFileName)
			}

			// Check if index exists
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				return fmt.Errorf("index file not found: %s\nRun 'fest index write' to generate it", indexPath)
			}

			result, err := index.ValidateFromFile(festivalRoot, indexPath)
			if err != nil {
				return fmt.Errorf("failed to validate index: %w", err)
			}

			// Print results
			if result.Valid && len(result.Warnings) == 0 && len(result.ExtraInFS) == 0 {
				fmt.Println("Index is valid and synchronized with filesystem.")
				return nil
			}

			if len(result.Errors) > 0 {
				fmt.Println("Errors:")
				for _, e := range result.Errors {
					fmt.Printf("  [%s] %s: %s\n", e.Type, e.Path, e.Message)
				}
			}

			if len(result.Warnings) > 0 {
				fmt.Println("\nWarnings:")
				for _, w := range result.Warnings {
					fmt.Printf("  %s\n", w)
				}
			}

			if len(result.ExtraInFS) > 0 {
				fmt.Println("\nFiles not in index:")
				for _, f := range result.ExtraInFS {
					fmt.Printf("  %s\n", f)
				}
			}

			if !result.Valid {
				return fmt.Errorf("index validation failed")
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&indexPath, "index", "i", "", "Path to index file (default: .festival/index.json)")

	return cmd
}

func newIndexShowCommand() *cobra.Command {
	var showJSON bool

	cmd := &cobra.Command{
		Use:   "show [festival-path]",
		Short: "Show festival index contents",
		Long:  `Display the contents of the festival index file.`,
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			festivalRoot, err := resolveFestivalRoot(args)
			if err != nil {
				return err
			}

			indexPath := filepath.Join(festivalRoot, ".festival", index.IndexFileName)

			// Check if index exists
			if _, err := os.Stat(indexPath); os.IsNotExist(err) {
				return fmt.Errorf("index file not found: %s\nRun 'fest index write' to generate it", indexPath)
			}

			idx, err := index.LoadIndex(indexPath)
			if err != nil {
				return fmt.Errorf("failed to load index: %w", err)
			}

			if showJSON {
				data, err := json.MarshalIndent(idx, "", "  ")
				if err != nil {
					return fmt.Errorf("failed to format index: %w", err)
				}
				fmt.Println(string(data))
				return nil
			}

			// Human-readable output
			fmt.Printf("Festival: %s\n", idx.FestivalID)
			fmt.Printf("Spec Version: %d\n", idx.FestSpec)
			fmt.Printf("Generated: %s\n\n", idx.GeneratedAt.Format("2006-01-02 15:04:05"))

			for _, phase := range idx.Phases {
				fmt.Printf("Phase: %s\n", phase.PhaseID)
				if phase.GoalFile != "" {
					fmt.Printf("  Goal: %s\n", phase.GoalFile)
				}

				for _, seq := range phase.Sequences {
					fmt.Printf("  Sequence: %s\n", seq.SequenceID)
					if seq.GoalFile != "" {
						fmt.Printf("    Goal: %s\n", seq.GoalFile)
					}

					for _, task := range seq.Tasks {
						if task.Managed {
							fmt.Printf("    [M] %s (gate: %s)\n", task.TaskID, task.GateID)
						} else {
							fmt.Printf("    [ ] %s\n", task.TaskID)
						}
					}
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&showJSON, "json", false, "Output as JSON")

	return cmd
}

// resolveFestivalRoot determines the festival root directory
func resolveFestivalRoot(args []string) (string, error) {
	if len(args) > 0 {
		absPath, err := filepath.Abs(args[0])
		if err != nil {
			return "", fmt.Errorf("invalid path: %w", err)
		}
		return absPath, nil
	}

	// Try to find festival root from current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Look for festivals/ parent or a festival indicator
	root, err := tpl.FindFestivalsRoot(cwd)
	if err == nil {
		// If we're in a festivals/ tree, try to find the specific festival
		rel, _ := filepath.Rel(root, cwd)
		if rel != "." && rel != "" {
			// We might be inside a festival - find its root
			parts := filepath.SplitList(rel)
			if len(parts) > 0 {
				return filepath.Join(root, parts[0]), nil
			}
		}
	}

	// Default to current directory
	return cwd, nil
}
