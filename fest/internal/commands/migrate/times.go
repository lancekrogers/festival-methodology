// Package migrate provides the fest migrate command for document migrations.
package migrate

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type timesOptions struct {
	path    string
	dryRun  bool
	verbose bool
}

// timesMigrationResult tracks results for a single festival time migration
type timesMigrationResult struct {
	path      string
	status    string // "migrated", "skipped", "error"
	taskCount int
	err       error
}

// NewTimesCommand creates the migrate times subcommand
func NewTimesCommand() *cobra.Command {
	opts := &timesOptions{}

	cmd := &cobra.Command{
		Use:   "times [path]",
		Short: "Populate time tracking data from file modification times",
		Long: `Retroactively populate time tracking data for existing festivals.

This command walks through festivals and uses file modification times to
infer task completion times for tasks that don't have explicit time data.

The migration:
- Finds all festivals in the specified path (or current directory)
- For each completed task without time data, infers time from file stats
- Updates progress.yaml with the inferred times
- Calculates total work time for the festival

Use --dry-run to preview changes without modifying files.`,
		Example: `  fest migrate times                    # Migrate current festival
  fest migrate times festivals/         # Migrate all festivals in directory
  fest migrate times --dry-run          # Preview changes
  fest migrate times --verbose          # Show detailed progress`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runTimesMigration(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "preview changes without modifying files")
	cmd.Flags().BoolVar(&opts.verbose, "verbose", false, "show detailed progress")

	return cmd
}

func runTimesMigration(ctx context.Context, opts *timesOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine starting path
	startPath := opts.path
	if startPath == "" {
		var err error
		startPath, err = os.Getwd()
		if err != nil {
			return errors.IO("getting current directory", err)
		}
	}

	// Make absolute
	if !filepath.IsAbs(startPath) {
		cwd, err := os.Getwd()
		if err != nil {
			return errors.IO("getting current directory", err)
		}
		startPath = filepath.Join(cwd, startPath)
	}

	// Find all festivals
	festivals, err := findFestivals(startPath)
	if err != nil {
		return err
	}

	if len(festivals) == 0 {
		fmt.Println(ui.Warning("No festivals found in " + startPath))
		return nil
	}

	if opts.dryRun {
		fmt.Println(ui.H1("Migrate Times (Dry Run)"))
	} else {
		fmt.Println(ui.H1("Migrate Times"))
	}

	// Process each festival
	var migrated, skipped, errored int
	for _, festPath := range festivals {
		result := migrateFestivalTimes(ctx, festPath, opts)

		switch result.status {
		case "migrated":
			migrated++
			if opts.verbose {
				fmt.Printf("%s %s: %d tasks updated\n",
					ui.StateIcon("completed"),
					ui.Value(filepath.Base(festPath)),
					result.taskCount)
			}
		case "skipped":
			skipped++
			if opts.verbose {
				fmt.Printf("%s %s: skipped (already has time data)\n",
					ui.StateIcon("pending"),
					ui.Value(filepath.Base(festPath)))
			}
		case "error":
			errored++
			fmt.Printf("%s %s: %s\n",
				ui.StateIcon("blocked"),
				ui.Value(filepath.Base(festPath)),
				ui.Warning(result.err.Error()))
		}
	}

	// Print summary
	fmt.Println()
	fmt.Printf("%s Migrated: %d, Skipped: %d, Errors: %d\n",
		ui.Label("Summary"),
		migrated, skipped, errored)

	if opts.dryRun {
		fmt.Println(ui.Dim("Run without --dry-run to apply changes"))
	}

	return nil
}

func findFestivals(root string) ([]string, error) {
	var festivals []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip directories we can't read
		}

		// Skip .fest directories
		if d.IsDir() && d.Name() == ".fest" {
			return filepath.SkipDir
		}

		// Look for fest.yaml files
		if !d.IsDir() && d.Name() == "fest.yaml" {
			festivals = append(festivals, filepath.Dir(path))
		}

		return nil
	})

	if err != nil {
		return nil, errors.IO("walking directory tree", err).WithField("root", root)
	}

	return festivals, nil
}

func migrateFestivalTimes(ctx context.Context, festPath string, opts *timesOptions) timesMigrationResult {
	result := timesMigrationResult{path: festPath}

	// Load progress store
	store := progress.NewStore(festPath)
	if err := store.Load(ctx); err != nil {
		result.status = "error"
		result.err = err
		return result
	}

	// Check if festival already has significant time data
	metrics := store.GetTimeMetrics()
	if metrics != nil && metrics.TotalWorkMinutes > 0 {
		result.status = "skipped"
		return result
	}

	// Process each task
	tasks := store.AllTasks()
	tasksUpdated := 0

	for _, task := range tasks {
		if task.Status != progress.StatusCompleted {
			continue
		}

		// Skip tasks that already have time data
		if task.TimeSpentMinutes > 0 {
			continue
		}

		// Build full task path
		taskPath := filepath.Join(festPath, task.TaskID)

		// Infer time from file
		if progress.InferTaskTime(taskPath, task) {
			tasksUpdated++
		}
	}

	if tasksUpdated == 0 {
		result.status = "skipped"
		return result
	}

	// Update total work minutes
	store.UpdateTotalWorkMinutes()

	// Save if not dry run
	if !opts.dryRun {
		if err := store.Save(ctx); err != nil {
			result.status = "error"
			result.err = err
			return result
		}
	}

	result.status = "migrated"
	result.taskCount = tasksUpdated
	return result
}
