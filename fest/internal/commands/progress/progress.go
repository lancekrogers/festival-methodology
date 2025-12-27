// Package progress implements the fest progress command for tracking execution progress.
package progress

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/spf13/cobra"
)

type progressOptions struct {
	json       bool
	update     string
	complete   bool
	blocker    string
	clear      bool
	taskID     string
	inProgress bool
}

// NewProgressCommand creates the progress command
func NewProgressCommand() *cobra.Command {
	opts := &progressOptions{}

	cmd := &cobra.Command{
		Use:   "progress",
		Short: "Track and display festival execution progress",
		Long: `Track and display progress for festival execution.

When run without flags, shows an overview of festival progress.
Use flags to update task progress, report blockers, or mark tasks complete.

PROGRESS OVERVIEW:
  fest progress              Show festival progress summary
  fest progress --json       Output progress in JSON format

TASK UPDATES:
  fest progress --task <id> --update 50%     Update task progress
  fest progress --task <id> --complete       Mark task as complete
  fest progress --task <id> --in-progress    Mark task as in progress
  fest progress --task <id> --blocker "msg"  Report a blocker
  fest progress --task <id> --clear          Clear blocker`,
		Example: `  fest progress                          # Show overall progress
  fest progress --task 01_setup.md --update 75%
  fest progress --task 01_setup.md --complete
  fest progress --task 02_impl.md --blocker "Waiting on API spec"
  fest progress --task 02_impl.md --clear`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runProgress(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.update, "update", "", "update progress percentage (e.g., 50%)")
	cmd.Flags().BoolVar(&opts.complete, "complete", false, "mark task as complete")
	cmd.Flags().StringVar(&opts.blocker, "blocker", "", "report a blocker with message")
	cmd.Flags().BoolVar(&opts.clear, "clear", false, "clear blocker for task")
	cmd.Flags().StringVar(&opts.taskID, "task", "", "task ID to update")
	cmd.Flags().BoolVar(&opts.inProgress, "in-progress", false, "mark task as in progress")

	return cmd
}

func runProgress(opts *progressOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return errors.Wrap(err, "detecting festival location")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival").
			WithField("hint", "run from inside a festival directory")
	}

	// Create progress manager
	mgr, err := progress.NewManager(loc.Festival.Path)
	if err != nil {
		return errors.Wrap(err, "initializing progress manager")
	}

	// Handle task updates
	if opts.taskID != "" {
		return handleTaskUpdate(mgr, opts)
	}

	// Show progress overview
	return showProgressOverview(mgr, loc, opts)
}

func handleTaskUpdate(mgr *progress.Manager, opts *progressOptions) error {
	taskID := opts.taskID

	// Handle blocker report
	if opts.blocker != "" {
		if err := mgr.ReportBlocker(taskID, opts.blocker); err != nil {
			return err
		}
		if opts.json {
			result := map[string]interface{}{
				"success": true,
				"task":    taskID,
				"status":  progress.StatusBlocked,
				"blocker": opts.blocker,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Blocker reported for %s\n", taskID)
		}
		return nil
	}

	// Handle clear blocker
	if opts.clear {
		if err := mgr.ClearBlocker(taskID); err != nil {
			return err
		}
		if opts.json {
			result := map[string]interface{}{
				"success": true,
				"task":    taskID,
				"cleared": true,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Blocker cleared for %s\n", taskID)
		}
		return nil
	}

	// Handle complete
	if opts.complete {
		if err := mgr.MarkComplete(taskID); err != nil {
			return err
		}
		if opts.json {
			task, _ := mgr.GetTaskProgress(taskID)
			result := map[string]interface{}{
				"success":            true,
				"task":               taskID,
				"status":             progress.StatusCompleted,
				"time_spent_minutes": task.TimeSpentMinutes,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			task, _ := mgr.GetTaskProgress(taskID)
			fmt.Printf("Task %s marked complete (time: %d min)\n", taskID, task.TimeSpentMinutes)
		}
		return nil
	}

	// Handle in-progress
	if opts.inProgress {
		if err := mgr.MarkInProgress(taskID); err != nil {
			return err
		}
		if opts.json {
			result := map[string]interface{}{
				"success": true,
				"task":    taskID,
				"status":  progress.StatusInProgress,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Task %s marked in progress\n", taskID)
		}
		return nil
	}

	// Handle progress update
	if opts.update != "" {
		pct, err := parsePercentage(opts.update)
		if err != nil {
			return err
		}
		if err := mgr.UpdateProgress(taskID, pct); err != nil {
			return err
		}
		if opts.json {
			task, _ := mgr.GetTaskProgress(taskID)
			result := map[string]interface{}{
				"success":  true,
				"task":     taskID,
				"progress": pct,
				"status":   task.Status,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Task %s progress updated to %d%%\n", taskID, pct)
		}
		return nil
	}

	// No update flags, show task progress
	task, exists := mgr.GetTaskProgress(taskID)
	if !exists {
		if opts.json {
			result := map[string]interface{}{
				"task":     taskID,
				"progress": 0,
				"status":   progress.StatusPending,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Task %s: pending (0%%)\n", taskID)
		}
		return nil
	}

	if opts.json {
		data, _ := json.MarshalIndent(task, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Task: %s\n", task.TaskID)
		fmt.Printf("Status: %s\n", task.Status)
		fmt.Printf("Progress: %d%%\n", task.Progress)
		if task.BlockerMessage != "" {
			fmt.Printf("Blocker: %s\n", task.BlockerMessage)
		}
		if task.TimeSpentMinutes > 0 {
			fmt.Printf("Time: %d min\n", task.TimeSpentMinutes)
		}
	}

	return nil
}

func showProgressOverview(mgr *progress.Manager, loc *show.LocationInfo, opts *progressOptions) error {
	festProgress, err := mgr.GetFestivalProgress(loc.Festival.Path)
	if err != nil {
		return errors.Wrap(err, "calculating progress")
	}

	if opts.json {
		data, _ := json.MarshalIndent(festProgress, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Printf("FESTIVAL PROGRESS: %s\n", festProgress.FestivalName)
	fmt.Println(strings.Repeat("=", 60))

	// Overall progress bar
	overall := festProgress.Overall
	fmt.Printf("\nOverall: %s %d%% (%d/%d tasks)\n",
		progressBar(overall.Percentage),
		overall.Percentage,
		overall.Completed,
		overall.Total)

	if overall.Blocked > 0 {
		fmt.Printf("⚠️  %d task(s) blocked\n", overall.Blocked)
	}

	if overall.TimeSpentMin > 0 {
		fmt.Printf("⏱️  Total time: %d min\n", overall.TimeSpentMin)
	}

	// Phase breakdown
	if len(festProgress.Phases) > 0 {
		fmt.Println("\nPHASES")
		fmt.Println(strings.Repeat("-", 60))
		for _, phase := range festProgress.Phases {
			status := "○"
			if phase.Progress.Completed == phase.Progress.Total && phase.Progress.Total > 0 {
				status = "✓"
			} else if phase.Progress.InProgress > 0 || phase.Progress.Completed > 0 {
				status = "●"
			}
			fmt.Printf("%s %-20s %3d%% (%d/%d)\n",
				status,
				phase.PhaseName,
				phase.Progress.Percentage,
				phase.Progress.Completed,
				phase.Progress.Total)
		}
	}

	// Show blockers if any
	if len(overall.Blockers) > 0 {
		fmt.Println("\nBLOCKERS")
		fmt.Println(strings.Repeat("-", 60))
		for _, blocker := range overall.Blockers {
			fmt.Printf("🚫 %s: %s\n", blocker.TaskID, blocker.BlockerMessage)
		}
	}

	return nil
}

func progressBar(percentage int) string {
	filled := percentage / 5  // 20 char bar
	empty := 20 - filled
	return "[" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "]"
}

func parsePercentage(s string) (int, error) {
	s = strings.TrimSuffix(s, "%")
	pct, err := strconv.Atoi(s)
	if err != nil {
		return 0, errors.Validation("invalid percentage").WithField("value", s)
	}
	if pct < 0 || pct > 100 {
		return 0, errors.Validation("percentage must be 0-100").WithField("value", pct)
	}
	return pct, nil
}
