// Package progress implements the fest progress command for tracking execution progress.
package progress

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

const (
	// ProgressBarWidth defines the number of characters in the progress bar
	ProgressBarWidth = 20
)

type progressOptions struct {
	json       bool
	update     string
	complete   bool
	blocker    string
	clear      bool
	taskID     string
	taskPath   string
	phase      string
	sequence   string
	festival   string
	inProgress bool
}

var taskFilenamePattern = regexp.MustCompile(`^\d{2}[\._].*\.md$`)

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
  fest progress --task <id> --clear          Clear blocker
  fest progress --path <task_path> --complete
  fest progress --phase <phase> --sequence <seq> --task <id> --complete

Task IDs can be festival-relative paths (e.g. 002_FOUNDATION/01_project_scaffold/01_design.md)
or absolute paths. Use --path or --phase/--sequence to disambiguate duplicates.
Use --festival to run outside a festival directory.`,
		Example: `  fest progress                          # Show overall progress
  fest progress --task 01_setup.md --update 75%
  fest progress --path 002_FOUNDATION/01_project_scaffold/01_design.md --complete
  fest progress --phase 002_FOUNDATION --sequence 01_project_scaffold --task 01_design.md --complete
  fest progress --festival festivals/active/guild-chat-GC0001 --task 01_setup.md --update 75%
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
	cmd.Flags().StringVar(&opts.taskPath, "path", "", "task path (festival-relative or absolute)")
	cmd.Flags().StringVar(&opts.phase, "phase", "", "phase directory name for task path")
	cmd.Flags().StringVar(&opts.sequence, "sequence", "", "sequence directory name for task path")
	cmd.Flags().StringVar(&opts.festival, "festival", "", "festival root path (directory containing fest.yaml)")
	cmd.Flags().BoolVar(&opts.inProgress, "in-progress", false, "mark task as in progress")

	return cmd
}

func runProgress(opts *progressOptions) error {
	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	if err := validateTaskOptions(opts); err != nil {
		return err
	}

	// Resolve festival path from --festival flag, navigation links, or current directory
	festivalPath := opts.festival
	if festivalPath != "" && !filepath.IsAbs(festivalPath) {
		festivalPath = filepath.Join(cwd, festivalPath)
	}

	// Use shared helper to resolve festival path (supports linked festivals)
	resolvedFestivalPath, err := shared.ResolveFestivalPath(cwd, festivalPath)
	if err != nil {
		return errors.Wrap(err, "detecting festival location")
	}

	targetPath := resolvedFestivalPath
	if opts.taskPath != "" {
		resolvedTaskPath, err := resolveTaskPath(opts.taskPath, resolvedFestivalPath, cwd)
		if err != nil {
			return err
		}
		opts.taskPath = resolvedTaskPath
		targetPath = resolvedTaskPath
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(targetPath)
	if err != nil {
		return errors.Wrap(err, "detecting festival location")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival").
			WithField("hint", "run from inside a festival directory")
	}

	// Create progress manager
	mgr, err := progress.NewManager(ctx, loc.Festival.Path)
	if err != nil {
		return errors.Wrap(err, "initializing progress manager")
	}

	// Handle task updates
	if opts.taskID != "" || opts.taskPath != "" {
		return handleTaskUpdate(ctx, mgr, loc.Festival.Path, opts)
	}

	// Show progress overview
	return showProgressOverview(ctx, mgr, loc, opts)
}

func handleTaskUpdate(ctx context.Context, mgr *progress.Manager, festivalPath string, opts *progressOptions) error {
	taskID, err := resolveTaskID(festivalPath, opts)
	if err != nil {
		return err
	}

	// Handle blocker report
	if opts.blocker != "" {
		if err := mgr.ReportBlocker(ctx, taskID, opts.blocker); err != nil {
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
			task, _ := mgr.GetTaskProgress(taskID)
			task = ensureTaskProgress(taskID, task, &progress.TaskProgress{
				Status:         progress.StatusBlocked,
				Progress:       0,
				BlockerMessage: opts.blocker,
			})
			printTaskProgress("Task Blocked", task)
		}
		return nil
	}

	// Handle clear blocker
	if opts.clear {
		if err := mgr.ClearBlocker(ctx, taskID); err != nil {
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
			task, _ := mgr.GetTaskProgress(taskID)
			task = ensureTaskProgress(taskID, task, &progress.TaskProgress{
				Status:   progress.StatusInProgress,
				Progress: 0,
			})
			printTaskProgress("Blocker Cleared", task)
		}
		return nil
	}

	// Handle complete
	if opts.complete {
		if err := mgr.MarkComplete(ctx, taskID); err != nil {
			return err
		}
		if opts.json {
			task, _ := mgr.GetTaskProgress(taskID)
			timeSpent := 0
			if task != nil {
				timeSpent = task.TimeSpentMinutes
			}
			result := map[string]interface{}{
				"success":            true,
				"task":               taskID,
				"status":             progress.StatusCompleted,
				"time_spent_minutes": timeSpent,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			task, _ := mgr.GetTaskProgress(taskID)
			task = ensureTaskProgress(taskID, task, &progress.TaskProgress{
				Status:   progress.StatusCompleted,
				Progress: 100,
			})
			printTaskProgress("Task Completed", task)
		}
		return nil
	}

	// Handle in-progress
	if opts.inProgress {
		if err := mgr.MarkInProgress(ctx, taskID); err != nil {
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
			task, _ := mgr.GetTaskProgress(taskID)
			task = ensureTaskProgress(taskID, task, &progress.TaskProgress{
				Status:   progress.StatusInProgress,
				Progress: 0,
			})
			printTaskProgress("Task In Progress", task)
		}
		return nil
	}

	// Handle progress update
	if opts.update != "" {
		pct, err := parsePercentage(opts.update)
		if err != nil {
			return err
		}
		if err := mgr.UpdateProgress(ctx, taskID, pct); err != nil {
			return err
		}
		if opts.json {
			task, _ := mgr.GetTaskProgress(taskID)
			status := statusForProgress(pct)
			if task != nil {
				status = task.Status
			}
			result := map[string]interface{}{
				"success":  true,
				"task":     taskID,
				"progress": pct,
				"status":   status,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			task, _ := mgr.GetTaskProgress(taskID)
			task = ensureTaskProgress(taskID, task, &progress.TaskProgress{
				Status:   statusForProgress(pct),
				Progress: pct,
			})
			printTaskProgress("Progress Updated", task)
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
			printTaskProgress("Task Progress", &progress.TaskProgress{
				TaskID:   taskID,
				Status:   progress.StatusPending,
				Progress: 0,
			})
		}
		return nil
	}

	if opts.json {
		data, _ := json.MarshalIndent(task, "", "  ")
		fmt.Println(string(data))
	} else {
		printTaskProgress("Task Progress", task)
	}

	return nil
}

func printTaskProgress(title string, task *progress.TaskProgress) {
	if task == nil {
		return
	}
	fmt.Println(ui.H1(title))
	fmt.Printf("%s %s\n", ui.Label("Task"), ui.Value(task.TaskID, ui.TaskColor))
	fmt.Printf("%s %s\n", ui.Label("Status"), ui.GetStateStyle(task.Status).Render(task.Status))
	fmt.Printf("%s %s\n", ui.Label("Progress"), ui.Value(fmt.Sprintf("%d%%", task.Progress)))
	if task.BlockerMessage != "" {
		fmt.Printf("%s %s\n", ui.Label("Blocker"), ui.Warning(task.BlockerMessage))
	}
	if task.TimeSpentMinutes > 0 {
		fmt.Printf("%s %s\n", ui.Label("Time"), ui.Value(fmt.Sprintf("%d min", task.TimeSpentMinutes)))
	}
}

func ensureTaskProgress(taskID string, task *progress.TaskProgress, fallback *progress.TaskProgress) *progress.TaskProgress {
	if task != nil {
		return task
	}
	if fallback == nil {
		fallback = &progress.TaskProgress{
			Status:   progress.StatusPending,
			Progress: 0,
		}
	}
	fallback.TaskID = taskID
	return fallback
}

func statusForProgress(progressPct int) string {
	switch {
	case progressPct >= 100:
		return progress.StatusCompleted
	case progressPct > 0:
		return progress.StatusInProgress
	default:
		return progress.StatusPending
	}
}

func showProgressOverview(ctx context.Context, mgr *progress.Manager, loc *show.LocationInfo, opts *progressOptions) error {
	// Determine scope based on location
	switch loc.Type {
	case "sequence":
		return showSequenceProgress(ctx, mgr, loc, opts)
	case "phase":
		return showPhaseProgress(ctx, mgr, loc, opts)
	case "festival", "task":
		return showFestivalProgress(ctx, mgr, loc, opts)
	default:
		return showFestivalProgress(ctx, mgr, loc, opts)
	}
}

func showFestivalProgress(ctx context.Context, mgr *progress.Manager, loc *show.LocationInfo, opts *progressOptions) error {
	festProgress, err := mgr.GetFestivalProgress(ctx, loc.Festival.Path)
	if err != nil {
		return errors.Wrap(err, "calculating progress")
	}

	if opts.json {
		data, _ := json.MarshalIndent(festProgress, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Println(ui.H1("Festival Progress"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(festProgress.FestivalName, ui.FestivalColor))
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	// Overall progress bar
	overall := festProgress.Overall
	fmt.Printf("\n%s %s %s %s\n",
		ui.Label("Overall"),
		renderProgressBar(overall.Percentage),
		ui.Value(fmt.Sprintf("%d%%", overall.Percentage)),
		ui.Dim(fmt.Sprintf("(%d/%d tasks)", overall.Completed, overall.Total)))

	if overall.Blocked > 0 {
		fmt.Printf("%s %s\n",
			ui.StateIcon("blocked"),
			ui.Value(fmt.Sprintf("%d task(s) blocked", overall.Blocked), ui.WarningColor))
	}

	if overall.TimeSpentMin > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Total time"),
			ui.Value(fmt.Sprintf("%d min", overall.TimeSpentMin)))
	}

	// Phase breakdown
	if len(festProgress.Phases) > 0 {
		fmt.Printf("\n%s\n", ui.H2("Phases"))
		fmt.Println(ui.Dim(strings.Repeat("─", 60)))
		for _, phase := range festProgress.Phases {
			state := "pending"
			if phase.Progress.Completed == phase.Progress.Total && phase.Progress.Total > 0 {
				state = "completed"
			} else if phase.Progress.InProgress > 0 || phase.Progress.Completed > 0 {
				state = "in_progress"
			}
			fmt.Printf("%s %s %s %s\n",
				ui.StateIcon(state),
				ui.Value(phase.PhaseName, ui.PhaseColor),
				ui.Dim(fmt.Sprintf("%3d%%", phase.Progress.Percentage)),
				ui.Dim(fmt.Sprintf("(%d/%d)", phase.Progress.Completed, phase.Progress.Total)))
		}
	}

	// Show blockers if any
	if len(overall.Blockers) > 0 {
		fmt.Printf("\n%s\n", ui.H2("Blockers"))
		fmt.Println(ui.Dim(strings.Repeat("─", 60)))
		for _, blocker := range overall.Blockers {
			fmt.Printf("%s %s %s\n",
				ui.StateIcon("blocked"),
				ui.Value(blocker.TaskID, ui.TaskColor),
				ui.Dim(blocker.BlockerMessage))
		}
	}

	return nil
}

func showPhaseProgress(ctx context.Context, mgr *progress.Manager, loc *show.LocationInfo, opts *progressOptions) error {
	phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
	phaseProgress, err := mgr.GetPhaseProgress(ctx, phasePath)
	if err != nil {
		return errors.Wrap(err, "calculating phase progress")
	}

	if opts.json {
		data, _ := json.MarshalIndent(phaseProgress, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Println(ui.H2("Phase Progress"))
	fmt.Printf("%s %s\n", ui.Label("Phase"), ui.Value(phaseProgress.PhaseName, ui.PhaseColor))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(loc.Festival.Name, ui.FestivalColor))
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	// Phase progress bar
	prog := phaseProgress.Progress
	fmt.Printf("\n%s %s %s %s\n",
		ui.Label("Progress"),
		renderProgressBar(prog.Percentage),
		ui.Value(fmt.Sprintf("%d%%", prog.Percentage)),
		ui.Dim(fmt.Sprintf("(%d/%d tasks)", prog.Completed, prog.Total)))

	if prog.InProgress > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("In progress"),
			ui.Value(fmt.Sprintf("%d", prog.InProgress), ui.InProgressColor))
	}

	if prog.Blocked > 0 {
		fmt.Printf("%s %s\n",
			ui.StateIcon("blocked"),
			ui.Value(fmt.Sprintf("%d task(s) blocked", prog.Blocked), ui.WarningColor))
	}

	if prog.TimeSpentMin > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Time spent"),
			ui.Value(fmt.Sprintf("%d min", prog.TimeSpentMin)))
	}

	// Show blockers if any
	if len(prog.Blockers) > 0 {
		fmt.Printf("\n%s\n", ui.H3("Blockers"))
		fmt.Println(ui.Dim(strings.Repeat("─", 60)))
		for _, blocker := range prog.Blockers {
			fmt.Printf("%s %s %s\n",
				ui.StateIcon("blocked"),
				ui.Value(blocker.TaskID, ui.TaskColor),
				ui.Dim(blocker.BlockerMessage))
		}
	}

	return nil
}

func showSequenceProgress(ctx context.Context, mgr *progress.Manager, loc *show.LocationInfo, opts *progressOptions) error {
	seqPath := filepath.Join(loc.Festival.Path, loc.Phase, loc.Sequence)
	seqProgress, err := mgr.GetSequenceProgress(ctx, seqPath)
	if err != nil {
		return errors.Wrap(err, "calculating sequence progress")
	}

	if opts.json {
		data, _ := json.MarshalIndent(seqProgress, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	// Human-readable output
	fmt.Println(ui.H2("Sequence Progress"))
	fmt.Printf("%s %s\n", ui.Label("Sequence"), ui.Value(seqProgress.SequenceName, ui.SequenceColor))
	fmt.Printf("%s %s\n", ui.Label("Phase"), ui.Value(loc.Phase, ui.PhaseColor))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(loc.Festival.Name, ui.FestivalColor))
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	// Sequence progress bar
	prog := seqProgress.Progress
	fmt.Printf("\n%s %s %s %s\n",
		ui.Label("Progress"),
		renderProgressBar(prog.Percentage),
		ui.Value(fmt.Sprintf("%d%%", prog.Percentage)),
		ui.Dim(fmt.Sprintf("(%d/%d tasks)", prog.Completed, prog.Total)))

	if prog.InProgress > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("In progress"),
			ui.Value(fmt.Sprintf("%d", prog.InProgress), ui.InProgressColor))
	}

	if prog.Pending > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Pending"),
			ui.Value(fmt.Sprintf("%d", prog.Pending), ui.PendingColor))
	}

	if prog.Blocked > 0 {
		fmt.Printf("%s %s\n",
			ui.StateIcon("blocked"),
			ui.Value(fmt.Sprintf("%d task(s) blocked", prog.Blocked), ui.WarningColor))
	}

	if prog.TimeSpentMin > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Time spent"),
			ui.Value(fmt.Sprintf("%d min", prog.TimeSpentMin)))
	}

	// Show blockers if any
	if len(prog.Blockers) > 0 {
		fmt.Printf("\n%s\n", ui.H3("Blockers"))
		fmt.Println(ui.Dim(strings.Repeat("─", 60)))
		for _, blocker := range prog.Blockers {
			fmt.Printf("%s %s %s\n",
				ui.StateIcon("blocked"),
				ui.Value(blocker.TaskID, ui.TaskColor),
				ui.Dim(blocker.BlockerMessage))
		}
	}

	return nil
}

func renderProgressBar(percentage int) string {
	opts := ui.DefaultProgressBarOptions()
	opts.Current = percentage
	opts.Total = 100
	opts.Width = ProgressBarWidth
	opts.ShowPercentage = false
	opts.ShowFraction = false
	opts.FilledColor = ui.SuccessColor
	opts.EmptyColor = ui.BorderColor
	return ui.RenderProgressBar(opts)
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

func validateTaskOptions(opts *progressOptions) error {
	if opts.taskPath != "" && (opts.taskID != "" || opts.phase != "" || opts.sequence != "") {
		return errors.Validation("use --path or --task/--phase/--sequence, not both")
	}

	if (opts.phase != "" || opts.sequence != "") && opts.taskID == "" {
		return errors.Validation("--phase/--sequence require --task")
	}

	if (opts.phase == "") != (opts.sequence == "") {
		return errors.Validation("both --phase and --sequence must be provided together")
	}

	return nil
}

func resolveTaskPath(pathArg, festivalPath, cwd string) (string, error) {
	if pathArg == "" {
		return "", errors.Validation("task path required")
	}

	if filepath.IsAbs(pathArg) {
		return filepath.Clean(pathArg), nil
	}

	if festivalPath != "" {
		return filepath.Clean(filepath.Join(festivalPath, pathArg)), nil
	}

	return filepath.Clean(filepath.Join(cwd, pathArg)), nil
}

func resolveTaskID(festivalPath string, opts *progressOptions) (string, error) {
	if opts.taskPath != "" {
		return progress.NormalizeTaskID(festivalPath, opts.taskPath)
	}

	taskID := strings.TrimSpace(opts.taskID)
	if taskID == "" {
		return "", errors.Validation("task ID required")
	}

	if opts.phase != "" && opts.sequence != "" {
		taskID = ensureMarkdownFilename(taskID)
		taskPath := filepath.Join(opts.phase, opts.sequence, taskID)
		return progress.NormalizeTaskID(festivalPath, taskPath)
	}

	normalized, err := progress.NormalizeTaskID(festivalPath, taskID)
	if err != nil {
		return "", err
	}

	if !strings.Contains(taskID, "/") && !strings.Contains(taskID, "\\") && !filepath.IsAbs(taskID) {
		matches, err := findTaskMatches(festivalPath, taskID)
		if err != nil {
			return "", err
		}

		if len(matches) > 1 {
			return "", errors.Validation("task ID is ambiguous; provide a full task path or use --phase/--sequence").
				WithField("task", taskID).
				WithField("matches", strings.Join(matches, ", "))
		}
		if len(matches) == 1 {
			return matches[0], nil
		}
	}

	return normalized, nil
}

func ensureMarkdownFilename(name string) string {
	if strings.HasSuffix(name, ".md") {
		return name
	}
	return name + ".md"
}

func findTaskMatches(festivalPath, taskID string) ([]string, error) {
	if festivalPath == "" {
		return nil, errors.Validation("festival path required")
	}

	exact := taskID
	withExt := taskID
	if !strings.HasSuffix(taskID, ".md") {
		withExt = taskID + ".md"
	}

	var matches []string
	err := filepath.WalkDir(festivalPath, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if entry.Name() == ".fest" || entry.Name() == "results" {
				return filepath.SkipDir
			}
			return nil
		}

		name := entry.Name()
		if !isTaskFile(name) {
			return nil
		}

		if name != exact && name != withExt {
			return nil
		}

		rel, err := filepath.Rel(festivalPath, path)
		if err != nil {
			return err
		}

		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}

		matches = append(matches, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, errors.IO("walking festival", err).WithField("path", festivalPath)
	}

	sort.Strings(matches)
	return matches, nil
}

func isTaskFile(name string) bool {
	return taskFilenamePattern.MatchString(name)
}
