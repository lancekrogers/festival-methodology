package status

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/tui"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// runStatusShow handles the status show command.
func runStatusShow(ctx context.Context, cmd *cobra.Command, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, opts.path)
	if err != nil {
		if opts.json {
			return emitErrorJSON("not in a festival directory")
		}
		return errors.Wrap(err, "not inside a festival")
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(ctx, festivalPath)
	if err != nil {
		if opts.json {
			return emitErrorJSON("not in a festival directory")
		}
		return err
	}

	if opts.json {
		return emitLocationJSON(loc)
	}
	return emitLocationText(ctx, loc)
}

// runStatusSet handles the status set command.
func runStatusSet(ctx context.Context, cmd *cobra.Command, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// Check if a level-specific flag was provided
	if opts.task != "" {
		return handleTaskStatusSet(ctx, display, cwd, newStatus, opts)
	}
	if opts.sequence != "" {
		return handleSequenceStatusSet(ctx, display, cwd, newStatus, opts)
	}
	if opts.phase != "" {
		return handlePhaseStatusSet(ctx, display, cwd, newStatus, opts)
	}

	// Handle --path flag: detect entity type and route accordingly
	if opts.path != "" {
		return handlePathBasedStatusSet(ctx, display, cwd, newStatus, opts)
	}

	// No level flag - use original logic (festival level or context-aware)
	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")

	// Handle case when not inside a festival or interactive mode requested
	if err != nil || opts.interactive {
		// Interactive selection mode
		selectedFestival, selectErr := selectFestivalForStatus(ctx, cwd, newStatus)
		if selectErr != nil {
			return selectErr
		}
		if selectedFestival == nil {
			// User cancelled
			display.Info("Selection cancelled.")
			return nil
		}

		// Use selected festival
		return applyStatusToFestival(ctx, display, selectedFestival, newStatus, opts)
	}

	// Detect current location
	// Try cwd first (when inside festival), then fall back to festivalPath (when linked)
	var loc *show.LocationInfo
	loc, err = show.DetectCurrentLocation(ctx, cwd)
	if err != nil || loc.Festival == nil {
		// We might be in a linked project directory
		// Fall back to festival root detection
		loc, err = show.DetectCurrentLocation(ctx, festivalPath)
		if err != nil {
			return err
		}
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// Context-aware routing based on detected location
	switch loc.Type {
	case "task":
		// In a task context - require explicit --task flag
		// Task status is too granular for auto-detect
		return showContextHint(display, opts, loc, newStatus, "task")

	case "sequence":
		// Auto-detect sequence status
		if !isValidStatus(EntitySequence, newStatus) {
			validOptions := ValidStatuses[EntitySequence]
			return errors.Validation("invalid status for sequence").
				WithField("status", newStatus).
				WithField("valid_options", strings.Join(validOptions, ", "))
		}
		// Route to sequence handler
		opts.sequence = loc.Sequence
		return handleSequenceStatusSet(ctx, display, cwd, newStatus, opts)

	case "phase":
		// Auto-detect phase status
		if !isValidStatus(EntityPhase, newStatus) {
			validOptions := ValidStatuses[EntityPhase]
			return errors.Validation("invalid status for phase").
				WithField("status", newStatus).
				WithField("valid_options", strings.Join(validOptions, ", "))
		}
		// Route to phase handler
		opts.phase = loc.Phase
		return handlePhaseStatusSet(ctx, display, cwd, newStatus, opts)

	case "festival":
		// At festival root - validate festival status
		if !isValidStatus(EntityFestival, newStatus) {
			validOptions := ValidStatuses[EntityFestival]
			return errors.Validation("invalid status for festival").
				WithField("status", newStatus).
				WithField("valid_options", strings.Join(validOptions, ", "))
		}
		return handleFestivalStatusChange(ctx, display, loc.Festival, newStatus, opts)

	default:
		// Unknown context - show help
		return errors.Validation("unknown context").
			WithField("type", loc.Type).
			WithField("hint", "use --phase, --sequence, or --task to specify level")
	}
}

// showContextHint shows a hint when in task context but no flag provided.
func showContextHint(display *ui.UI, opts *statusOptions, loc *show.LocationInfo, newStatus, contextType string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":      false,
			"context_type": contextType,
			"hint":         "use --task flag to set task status explicitly",
			"current_task": loc.Task,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
		return nil
	}

	fmt.Println(ui.Warning("Context Detection"))
	fmt.Printf("%s %s\n", ui.Label("Current location"), ui.Dim(contextType))
	if loc.Task != "" {
		fmt.Printf("%s %s\n", ui.Label("Task"), ui.Dim(loc.Task))
	}
	fmt.Println()
	fmt.Println(ui.Info("Task status requires explicit targeting:"))
	fmt.Printf("  fest status set --task %s %s\n", loc.Task, newStatus)
	fmt.Println()
	fmt.Println(ui.Dim("Or to set a higher level:"))
	fmt.Printf("  fest status set --sequence %s %s  # sequence status\n", loc.Sequence, newStatus)
	fmt.Printf("  fest status set --phase %s %s       # phase status\n", loc.Phase, newStatus)

	return nil
}

// selectFestivalForStatus opens an interactive selector for choosing a festival.
func selectFestivalForStatus(ctx context.Context, cwd, newStatus string) (*show.FestivalInfo, error) {
	// Find festivals directory
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return nil, errors.Wrap(err, "finding festivals directory")
	}
	if festivalsDir == "" {
		return nil, errors.NotFound("festivals directory").
			WithField("hint", "navigate to a workspace with festivals/ directory")
	}

	// Configure selector with appropriate title
	config := tui.DefaultSelectorConfig()
	config.Title = "Select Festival to Change Status"
	config.Description = fmt.Sprintf("Choose festival to set to '%s'", newStatus)

	// Create and run selector
	selector := tui.NewFestivalSelector(festivalsDir, config)
	result, err := selector.Run(ctx)
	if err != nil {
		return nil, err
	}

	if result.Cancelled {
		return nil, nil // User cancelled - not an error
	}

	return result.Selected, nil
}

// applyStatusToFestival applies a status change to a festival.
func applyStatusToFestival(ctx context.Context, display *ui.UI, festival *show.FestivalInfo, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Validate status for festivals
	if !isValidStatus(EntityFestival, newStatus) {
		validOptions := ValidStatuses[EntityFestival]
		return errors.Validation("invalid status").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Apply the status change
	return handleFestivalStatusChange(ctx, display, festival, newStatus, opts)
}

// emitStatusSetPlaceholder outputs a placeholder message for non-festival status changes.
func emitStatusSetPlaceholder(display *ui.UI, opts *statusOptions, entityType, newStatus string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"entity":     entityType,
			"new_status": newStatus,
			"note":       "frontmatter updates not yet implemented",
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Println(ui.H1("Status Updated"))
		fmt.Printf("%s %s\n", ui.Label("Entity"), ui.Value(string(entityType)))
		fmt.Printf("%s %s\n", ui.Label("Status"), ui.GetStateStyle(newStatus).Render(newStatus))
		fmt.Println(ui.Dim("Frontmatter updates pending implementation"))
	}
	return nil
}

// handleFestivalStatusChange handles changing a festival's status by moving its directory.
func handleFestivalStatusChange(ctx context.Context, display *ui.UI, festival *show.FestivalInfo, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if festival.Status == newStatus {
		return emitAlreadyAtStatus(display, opts, newStatus)
	}

	// Confirm unless forced
	if !opts.force {
		if !confirmFestivalMove(display, festival, newStatus) {
			display.Info("Operation cancelled.")
			return nil
		}
	}

	// Execute the move
	return executeFestivalMove(ctx, festival, newStatus, opts)
}

// emitAlreadyAtStatus outputs a message when festival is already at the requested status.
func emitAlreadyAtStatus(display *ui.UI, opts *statusOptions, status string) error {
	if opts.json {
		result := map[string]interface{}{
			"success": true,
			"message": "already at requested status",
			"status":  status,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Printf("%s %s\n", ui.Info("Already at status"), ui.GetStateStyle(status).Render(status))
	}
	return nil
}

// confirmFestivalMove prompts the user to confirm a festival move operation.
func confirmFestivalMove(display *ui.UI, festival *show.FestivalInfo, newStatus string) bool {
	prompt := fmt.Sprintf("Move festival %s from %s to %s?",
		ui.Value(festival.Name, ui.FestivalColor),
		ui.GetStateStyle(festival.Status).Render(festival.Status),
		ui.GetStateStyle(newStatus).Render(newStatus))
	return display.Confirm("%s", prompt)
}

// executeFestivalMove performs the actual directory move for a festival status change.
func executeFestivalMove(ctx context.Context, festival *show.FestivalInfo, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Calculate new path
	festivalsRoot := filepath.Dir(filepath.Dir(festival.Path))
	newPath := filepath.Join(festivalsRoot, newStatus, festival.Name)

	// Check if destination exists
	if _, err := os.Stat(newPath); err == nil {
		return errors.Validation("destination already exists").WithField("path", newPath)
	}

	// Create destination directory if needed
	destDir := filepath.Join(festivalsRoot, newStatus)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.IO("creating destination directory", err)
	}

	// Move the directory
	if err := os.Rename(festival.Path, newPath); err != nil {
		return errors.IO("moving festival directory", err)
	}

	return emitFestivalMoveSuccess(opts, festival, newStatus, newPath)
}

// emitFestivalMoveSuccess outputs success message after moving a festival.
func emitFestivalMoveSuccess(opts *statusOptions, festival *show.FestivalInfo, newStatus, newPath string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"festival":   festival.Name,
			"old_status": festival.Status,
			"new_status": newStatus,
			"old_path":   festival.Path,
			"new_path":   newPath,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Println(ui.Success("✓ Festival status updated"))
		fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(festival.Name, ui.FestivalColor))
		fmt.Printf("%s %s\n", ui.Label("From"), ui.GetStateStyle(festival.Status).Render(festival.Status))
		fmt.Printf("%s %s\n", ui.Label("To"), ui.GetStateStyle(newStatus).Render(newStatus))
		fmt.Printf("%s %s\n", ui.Label("Path"), ui.Dim(newPath))
	}
	return nil
}

// runFestivalListing handles listing festivals.
func runFestivalListing(ctx context.Context, festivalsRoot, filterStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if filterStatus != "" {
		festivals, err := show.ListFestivalsByStatus(ctx, festivalsRoot, filterStatus)
		if err != nil {
			return err
		}

		if opts.json {
			result := map[string]interface{}{
				"status":    filterStatus,
				"count":     len(festivals),
				"festivals": festivals,
			}
			if err := shared.EncodeJSON(os.Stdout, result); err != nil {
				return errors.Wrap(err, "encoding JSON output")
			}
		} else {
			fmt.Println(show.FormatFestivalList(filterStatus, festivals))
		}
	} else {
		fmt.Println("Use 'fest show all' to see all festivals grouped by status")
	}

	return nil
}

// runPhaseListing handles listing phases in a festival.
func runPhaseListing(ctx context.Context, loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	phases, err := collectPhasesForListing(ctx, loc)
	if err != nil {
		return err
	}

	// Filter by status
	phases = filterPhasesByStatus(phases, filterStatus)

	// Handle empty results
	if len(phases) == 0 {
		return emitEmptyPhasesResult(opts, filterStatus)
	}

	// Emit output
	if opts.json {
		return emitPhasesJSON(phases, filterStatus)
	}
	return emitPhasesText(phases, filterStatus)
}

// collectPhasesForListing collects phases based on the current location context.
func collectPhasesForListing(ctx context.Context, loc *show.LocationInfo) ([]*PhaseInfo, error) {
	if loc.Type == "festival" {
		// List all phases in festival
		return collectPhases(ctx, loc.Festival.Path)
	}

	// In phase/sequence/task - list just current phase
	phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
	phase, err := collectPhaseInfo(ctx, phasePath, loc.Phase)
	if err != nil {
		return nil, err
	}
	return []*PhaseInfo{phase}, nil
}

// emitEmptyPhasesResult outputs a message when no phases are found.
func emitEmptyPhasesResult(opts *statusOptions, filterStatus string) error {
	if opts.json {
		return emitEmptyJSON("phase", filterStatus)
	}
	message := "No phases found"
	if filterStatus != "" {
		message = fmt.Sprintf("No phases found with status '%s'", filterStatus)
	}
	fmt.Println(ui.Info(message))
	return nil
}

// runSequenceListing handles listing sequences in a festival or phase.
func runSequenceListing(ctx context.Context, loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	sequences, err := collectSequencesForListing(ctx, loc)
	if err != nil {
		return err
	}

	// Filter by status
	sequences = filterSequencesByStatus(sequences, filterStatus)

	// Handle empty results
	if len(sequences) == 0 {
		return emitEmptySequencesResult(opts, filterStatus)
	}

	// Emit output
	if opts.json {
		return emitSequencesJSON(sequences, filterStatus)
	}
	return emitSequencesText(sequences, filterStatus)
}

// collectSequencesForListing collects sequences based on the current location context.
func collectSequencesForListing(ctx context.Context, loc *show.LocationInfo) ([]*SequenceInfo, error) {
	store := progressStoreForFestival(ctx, loc.Festival.Path)
	switch loc.Type {
	case "festival":
		// List all sequences in festival
		return collectSequencesFromFestival(ctx, loc.Festival.Path)
	case "phase":
		// List sequences in current phase
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		return collectSequences(ctx, phasePath, loc.Phase, store, loc.Festival.Path)
	default:
		// In sequence or task - list current sequence
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		seqPath := filepath.Join(phasePath, loc.Sequence)
		seq, err := collectSequenceInfo(ctx, seqPath, loc.Phase, loc.Sequence, store, loc.Festival.Path)
		if err != nil {
			return nil, err
		}
		return []*SequenceInfo{seq}, nil
	}
}

// emitEmptySequencesResult outputs a message when no sequences are found.
func emitEmptySequencesResult(opts *statusOptions, filterStatus string) error {
	if opts.json {
		return emitEmptyJSON("sequence", filterStatus)
	}
	message := "No sequences found"
	if filterStatus != "" {
		message = fmt.Sprintf("No sequences found with status '%s'", filterStatus)
	}
	fmt.Println(ui.Info(message))
	return nil
}

// runTaskListing handles listing tasks in a festival, phase, or sequence.
func runTaskListing(ctx context.Context, loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	tasks, err := collectTasksForListing(ctx, loc)
	if err != nil {
		return err
	}

	// Filter by status
	tasks = filterTasksByStatus(tasks, filterStatus)

	// Handle empty results
	if len(tasks) == 0 {
		return emitEmptyTasksResult(opts, filterStatus)
	}

	// Emit output
	if opts.json {
		return emitTasksJSON(tasks, filterStatus)
	}
	return emitTasksText(tasks, filterStatus)
}

// collectTasksForListing collects tasks based on the current location context.
func collectTasksForListing(ctx context.Context, loc *show.LocationInfo) ([]*TaskInfo, error) {
	store := progressStoreForFestival(ctx, loc.Festival.Path)
	switch loc.Type {
	case "festival":
		// List all tasks in festival
		return collectTasksFromFestival(ctx, loc.Festival.Path)
	case "phase":
		// List tasks in current phase
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		return collectTasksFromPhase(ctx, phasePath, loc.Phase, store, loc.Festival.Path)
	default:
		// In sequence or task - list tasks in current sequence
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		seqPath := filepath.Join(phasePath, loc.Sequence)
		return collectTasks(ctx, seqPath, loc.Phase, loc.Sequence, store, loc.Festival.Path)
	}
}

// emitEmptyTasksResult outputs a message when no tasks are found.
func emitEmptyTasksResult(opts *statusOptions, filterStatus string) error {
	if opts.json {
		return emitEmptyJSON("task", filterStatus)
	}
	message := "No tasks found"
	if filterStatus != "" {
		message = fmt.Sprintf("No tasks found with status '%s'", filterStatus)
	}
	fmt.Println(ui.Info(message))
	return nil
}

// runStatusList is the main handler for the status list command.
func runStatusList(ctx context.Context, cmd *cobra.Command, filterStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return handleStatusListOutsideFestival(ctx, cwd, filterStatus, opts)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(ctx, festivalPath)
	if err != nil {
		return handleStatusListOutsideFestival(ctx, cwd, filterStatus, opts)
	}

	// Validate entity type and status
	entityType := EntityType(opts.entityType)
	if filterStatus != "" && !isValidStatus(entityType, filterStatus) {
		validOptions := ValidStatuses[entityType]
		return errors.Validation("invalid status for entity type").
			WithField("status", filterStatus).
			WithField("type", opts.entityType).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Route based on entity type
	return routeStatusListByType(ctx, loc, filterStatus, opts)
}

// handleStatusListOutsideFestival handles status list when not in a festival directory.
func handleStatusListOutsideFestival(ctx context.Context, cwd, filterStatus string, opts *statusOptions) error {
	festivalsDir := findFestivalsRoot(cwd)
	if festivalsDir == "" {
		return errors.NotFound("festival or festivals directory").
			WithField("hint", "navigate to a festival directory to list phases/sequences/tasks")
	}
	if opts.entityType == "festival" || opts.entityType == "" {
		return runFestivalListing(ctx, festivalsDir, filterStatus, opts)
	}
	return errors.NotFound("festival").
		WithField("hint", "navigate to a festival directory to list phases/sequences/tasks")
}

// routeStatusListByType routes the status list command based on entity type.
func routeStatusListByType(ctx context.Context, loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	switch opts.entityType {
	case "festival", "":
		festivalsRoot := filepath.Dir(filepath.Dir(loc.Festival.Path))
		return runFestivalListing(ctx, festivalsRoot, filterStatus, opts)
	case "phase":
		return runPhaseListing(ctx, loc, filterStatus, opts)
	case "sequence":
		return runSequenceListing(ctx, loc, filterStatus, opts)
	case "task":
		return runTaskListing(ctx, loc, filterStatus, opts)
	default:
		return errors.Validation("invalid entity type").
			WithField("type", opts.entityType).
			WithField("valid_types", "festival, phase, sequence, task")
	}
}

// runStatusHistory handles the status history command.
func runStatusHistory(ctx context.Context, cmd *cobra.Command, limit int, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	loc, err := show.DetectCurrentLocation(ctx, festivalPath)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// Load and emit history
	history, err := loadStatusHistory(ctx, loc.Festival.Path)
	if err != nil {
		return err
	}

	return emitStatusHistory(opts, loc.Festival.Name, history, limit)
}

// loadStatusHistory loads the status history from a festival's history file.
func loadStatusHistory(ctx context.Context, festivalPath string) ([]map[string]interface{}, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	historyPath := filepath.Join(festivalPath, ".fest", "status_history.json")

	// Check if history exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return nil, nil // No history file - not an error
	}

	// Read history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, errors.IO("reading history file", err)
	}

	var history []map[string]interface{}
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, errors.Wrap(err, "parsing history file")
	}

	return history, nil
}

// emitStatusHistory outputs the status history in the appropriate format.
func emitStatusHistory(opts *statusOptions, festivalName string, history []map[string]interface{}, limit int) error {
	// Handle no history case
	if history == nil {
		if opts.json {
			result := map[string]interface{}{
				"history": []interface{}{},
				"message": "no status history found",
			}
			if err := shared.EncodeJSON(os.Stdout, result); err != nil {
				return errors.Wrap(err, "encoding JSON output")
			}
		} else {
			fmt.Println("No status history found for this festival.")
		}
		return nil
	}

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	if opts.json {
		result := map[string]interface{}{
			"festival": festivalName,
			"count":    len(history),
			"history":  history,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Printf("Status History for %s:\n", festivalName)
		fmt.Println(strings.Repeat("-", 50))
		for _, entry := range history {
			fmt.Printf("%s: %s -> %s\n",
				entry["timestamp"],
				entry["from_status"],
				entry["to_status"])
			if note, ok := entry["note"].(string); ok && note != "" {
				fmt.Printf("  Note: %s\n", note)
			}
		}
	}

	return nil
}

// handlePathBasedStatusSet handles --path flag by detecting entity type and routing accordingly.
// This allows setting festival status from anywhere in the workspace by passing the festival name.
func handlePathBasedStatusSet(ctx context.Context, display *ui.UI, cwd, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// First, try to resolve as a festival
	festivalPath, err := resolveFestivalFromPath(cwd, opts.path)
	if err == nil {
		// Successfully resolved as a festival - detect entity type to confirm
		entityType := detectEntityType(festivalPath)

		switch entityType {
		case EntityFestival:
			// Validate festival status
			if !isValidStatus(EntityFestival, newStatus) {
				validOptions := ValidStatuses[EntityFestival]
				return errors.Validation("invalid status for festival").
					WithField("status", newStatus).
					WithField("valid_options", strings.Join(validOptions, ", "))
			}

			// Get festival info using DetectCurrentLocation
			loc, locErr := show.DetectCurrentLocation(ctx, festivalPath)
			if locErr != nil {
				return errors.Wrap(locErr, "detecting festival info")
			}
			if loc.Festival == nil {
				return errors.NotFound("festival info").WithField("path", festivalPath)
			}

			return handleFestivalStatusChange(ctx, display, loc.Festival, newStatus, opts)

		case EntityPhase:
			// Path points to a phase - set phase name and route to phase handler
			opts.phase = filepath.Base(festivalPath)
			// Get festival root (parent of phase)
			festivalRoot := filepath.Dir(festivalPath)
			return handlePhaseStatusSetWithPath(ctx, display, festivalRoot, newStatus, opts)

		case EntitySequence:
			// Path points to a sequence - more complex routing needed
			// For now, fall through to task handling which can resolve sequences
			break

		case EntityTask:
			// Path points to a task file - route to task handler
			return handleTaskStatusSet(ctx, display, cwd, newStatus, opts)
		}
	}

	// Path didn't resolve as a festival - try as a task path
	// This maintains backward compatibility for task-level --path usage
	return handleTaskStatusSet(ctx, display, cwd, newStatus, opts)
}

// handlePhaseStatusSetWithPath handles phase status when we already know the festival path.
func handlePhaseStatusSetWithPath(ctx context.Context, display *ui.UI, festivalPath, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Validate status for phases
	if !isValidStatus(EntityPhase, newStatus) {
		validOptions := ValidStatuses[EntityPhase]
		return errors.Validation("invalid status for phase").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Find the phase directory
	phasePath, phaseName, err := resolvePhase(festivalPath, opts.phase)
	if err != nil {
		return err
	}

	// Phase status is stored in PHASE_GOAL.md frontmatter
	_ = phasePath
	return emitPhaseStatusPlaceholder(display, opts, phaseName, newStatus)
}

// handleTaskStatusSet handles setting status for a specific task.
func handleTaskStatusSet(ctx context.Context, display *ui.UI, cwd, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Validate status for tasks
	if !isValidStatus(EntityTask, newStatus) {
		validOptions := ValidStatuses[EntityTask]
		return errors.Validation("invalid status for task").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Resolve festival path
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival").
			WithField("hint", "navigate to a festival directory or use 'fest link'")
	}

	// Determine task ID from flag
	taskID := opts.task
	if opts.path != "" {
		taskID = opts.path
	}

	// Normalize task ID - resolve relative to festival
	taskID, err = resolveTaskID(festivalPath, cwd, taskID)
	if err != nil {
		return err
	}

	// Create progress manager
	mgr, err := progress.NewManager(ctx, festivalPath)
	if err != nil {
		return errors.Wrap(err, "creating progress manager")
	}

	// Get current status for display
	currentTask, exists := mgr.GetTaskProgress(taskID)
	currentStatus := "pending"
	if exists {
		currentStatus = string(currentTask.Status)
	}

	// Check if already at target status
	if currentStatus == newStatus {
		return emitTaskStatusAlready(display, opts, taskID, newStatus)
	}

	// Apply the status change based on target status
	switch newStatus {
	case "pending":
		// Reset to pending - set progress to 0
		if err := mgr.UpdateProgress(ctx, taskID, 0); err != nil {
			return errors.Wrap(err, "resetting task status")
		}
	case "in_progress":
		if err := mgr.MarkInProgress(ctx, taskID); err != nil {
			return errors.Wrap(err, "marking task in progress")
		}
	case "blocked":
		// For blocked, we need a message - use generic if not provided
		if err := mgr.ReportBlocker(ctx, taskID, "Blocked via status set"); err != nil {
			return errors.Wrap(err, "marking task blocked")
		}
	case "completed":
		if err := mgr.MarkComplete(ctx, taskID); err != nil {
			return errors.Wrap(err, "marking task complete")
		}
	}

	return emitTaskStatusSuccess(display, opts, taskID, currentStatus, newStatus)
}

// resolveTaskID normalizes a task identifier to a festival-relative path.
func resolveTaskID(festivalPath, cwd, taskInput string) (string, error) {
	// If it's already a full path within the festival, extract relative part
	if strings.HasPrefix(taskInput, festivalPath) {
		return strings.TrimPrefix(taskInput, festivalPath+"/"), nil
	}

	// If it's a relative path starting with ./ or ../
	if strings.HasPrefix(taskInput, "./") || strings.HasPrefix(taskInput, "../") {
		absPath := filepath.Join(cwd, taskInput)
		if strings.HasPrefix(absPath, festivalPath) {
			return strings.TrimPrefix(absPath, festivalPath+"/"), nil
		}
		return "", errors.Validation("path is outside festival").
			WithField("path", taskInput).
			WithField("festival", festivalPath)
	}

	// If it looks like a phase/sequence/task path (e.g., 001/01/01_task.md)
	if strings.Contains(taskInput, "/") || strings.HasSuffix(taskInput, ".md") {
		// Verify it exists
		fullPath := filepath.Join(festivalPath, taskInput)
		if _, err := os.Stat(fullPath); err == nil {
			return taskInput, nil
		}
	}

	// Try to find in current directory context
	// If cwd is within festival, try appending task name
	if strings.HasPrefix(cwd, festivalPath) {
		relCwd := strings.TrimPrefix(cwd, festivalPath+"/")
		testPath := filepath.Join(relCwd, taskInput)
		fullPath := filepath.Join(festivalPath, testPath)
		if _, err := os.Stat(fullPath); err == nil {
			return testPath, nil
		}
	}

	// Finally, try searching for the task
	return findTaskByName(festivalPath, taskInput)
}

// findTaskByName searches for a task file by name within a festival.
func findTaskByName(festivalPath, taskName string) (string, error) {
	var matches []string

	err := filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip hidden directories
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Check if this matches the task name
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".md") {
			if info.Name() == taskName || strings.Contains(info.Name(), taskName) {
				relPath := strings.TrimPrefix(path, festivalPath+"/")
				matches = append(matches, relPath)
			}
		}

		return nil
	})
	if err != nil {
		return "", errors.Wrap(err, "searching for task")
	}

	if len(matches) == 0 {
		return "", errors.NotFound("task").
			WithField("name", taskName).
			WithField("hint", "use full path like '001/01/01_task.md'")
	}

	if len(matches) > 1 {
		return "", errors.Validation("ambiguous task name").
			WithField("name", taskName).
			WithField("matches", strings.Join(matches, ", ")).
			WithField("hint", "use full path to disambiguate")
	}

	return matches[0], nil
}

// emitTaskStatusAlready outputs message when task is already at the requested status.
func emitTaskStatusAlready(display *ui.UI, opts *statusOptions, taskID, status string) error {
	if opts.json {
		result := map[string]interface{}{
			"success": true,
			"message": "task already at requested status",
			"task":    taskID,
			"status":  status,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Printf("%s %s\n", ui.Info("Task already at status"), ui.GetStateStyle(status).Render(status))
		fmt.Printf("%s %s\n", ui.Label("Task"), ui.Dim(taskID))
	}
	return nil
}

// emitTaskStatusSuccess outputs success message after changing task status.
func emitTaskStatusSuccess(display *ui.UI, opts *statusOptions, taskID, oldStatus, newStatus string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"task":       taskID,
			"old_status": oldStatus,
			"new_status": newStatus,
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Println(ui.Success("✓ Task status updated"))
		fmt.Printf("%s %s\n", ui.Label("Task"), ui.Dim(taskID))
		fmt.Printf("%s %s\n", ui.Label("From"), ui.GetStateStyle(oldStatus).Render(oldStatus))
		fmt.Printf("%s %s\n", ui.Label("To"), ui.GetStateStyle(newStatus).Render(newStatus))
	}
	return nil
}

// handlePhaseStatusSet handles setting status for a specific phase.
func handlePhaseStatusSet(ctx context.Context, display *ui.UI, cwd, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Validate status for phases
	if !isValidStatus(EntityPhase, newStatus) {
		validOptions := ValidStatuses[EntityPhase]
		return errors.Validation("invalid status for phase").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Resolve festival path
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival").
			WithField("hint", "navigate to a festival directory or use 'fest link'")
	}

	// Find the phase directory
	phasePath, phaseName, err := resolvePhase(festivalPath, opts.phase)
	if err != nil {
		return err
	}

	// Phase status is stored in PHASE_GOAL.md frontmatter
	// For now, emit a placeholder until frontmatter editing is implemented
	_ = phasePath
	return emitPhaseStatusPlaceholder(display, opts, phaseName, newStatus)
}

// resolvePhase finds a phase directory by name or number.
func resolvePhase(festivalPath, phaseInput string) (string, string, error) {
	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return "", "", errors.IO("reading festival directory", err)
	}

	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden and metadata directories
		if strings.HasPrefix(name, ".") {
			continue
		}
		// Check for phase match (by prefix number or full name)
		if strings.HasPrefix(name, phaseInput) || name == phaseInput {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return "", "", errors.NotFound("phase").
			WithField("input", phaseInput).
			WithField("hint", "use phase number like '001' or full name like '001_CRITICAL'")
	}

	if len(matches) > 1 {
		return "", "", errors.Validation("ambiguous phase").
			WithField("input", phaseInput).
			WithField("matches", strings.Join(matches, ", "))
	}

	return filepath.Join(festivalPath, matches[0]), matches[0], nil
}

// emitPhaseStatusPlaceholder outputs a placeholder for phase status changes.
func emitPhaseStatusPlaceholder(display *ui.UI, opts *statusOptions, phaseName, newStatus string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"phase":      phaseName,
			"new_status": newStatus,
			"note":       "phase status in frontmatter - implementation pending",
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Println(ui.H1("Phase Status"))
		fmt.Printf("%s %s\n", ui.Label("Phase"), ui.Value(phaseName, ui.PhaseColor))
		fmt.Printf("%s %s\n", ui.Label("Target"), ui.GetStateStyle(newStatus).Render(newStatus))
		fmt.Println(ui.Dim("Phase status stored in PHASE_GOAL.md frontmatter"))
		fmt.Println(ui.Dim("Frontmatter editing pending implementation"))
	}
	return nil
}

// handleSequenceStatusSet handles setting status for a specific sequence.
func handleSequenceStatusSet(ctx context.Context, display *ui.UI, cwd, newStatus string, opts *statusOptions) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Validate status for sequences
	if !isValidStatus(EntitySequence, newStatus) {
		validOptions := ValidStatuses[EntitySequence]
		return errors.Validation("invalid status for sequence").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Resolve festival path
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival").
			WithField("hint", "navigate to a festival directory or use 'fest link'")
	}

	// Find the sequence directory
	seqPath, seqName, err := resolveSequence(festivalPath, cwd, opts.sequence)
	if err != nil {
		return err
	}

	// Sequence status is stored in SEQUENCE_GOAL.md frontmatter
	// For now, emit a placeholder until frontmatter editing is implemented
	_ = seqPath
	return emitSequenceStatusPlaceholder(display, opts, seqName, newStatus)
}

// resolveSequence finds a sequence directory by name or path.
func resolveSequence(festivalPath, cwd, seqInput string) (string, string, error) {
	// If input contains a slash, treat as phase/sequence path
	if strings.Contains(seqInput, "/") {
		parts := strings.SplitN(seqInput, "/", 2)
		phasePath, phaseName, err := resolvePhase(festivalPath, parts[0])
		if err != nil {
			return "", "", err
		}
		seqPath, seqName, err := findSequenceInPhase(phasePath, parts[1])
		if err != nil {
			return "", "", err
		}
		return seqPath, phaseName + "/" + seqName, nil
	}

	// Otherwise, search in current phase context or all phases
	// First check if we're in a phase directory
	if strings.HasPrefix(cwd, festivalPath) {
		relPath := strings.TrimPrefix(cwd, festivalPath+"/")
		parts := strings.Split(relPath, "/")
		if len(parts) >= 1 {
			// Try to find sequence in current phase
			phasePath := filepath.Join(festivalPath, parts[0])
			seqPath, seqName, err := findSequenceInPhase(phasePath, seqInput)
			if err == nil {
				return seqPath, parts[0] + "/" + seqName, nil
			}
		}
	}

	// Search all phases for the sequence
	return findSequenceGlobally(festivalPath, seqInput)
}

// findSequenceInPhase finds a sequence within a specific phase.
func findSequenceInPhase(phasePath, seqInput string) (string, string, error) {
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return "", "", errors.IO("reading phase directory", err)
	}

	var matches []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, ".") {
			continue
		}
		if strings.HasPrefix(name, seqInput) || name == seqInput {
			matches = append(matches, name)
		}
	}

	if len(matches) == 0 {
		return "", "", errors.NotFound("sequence in phase").
			WithField("input", seqInput)
	}

	if len(matches) > 1 {
		return "", "", errors.Validation("ambiguous sequence").
			WithField("input", seqInput).
			WithField("matches", strings.Join(matches, ", "))
	}

	return filepath.Join(phasePath, matches[0]), matches[0], nil
}

// findSequenceGlobally searches all phases for a sequence.
func findSequenceGlobally(festivalPath, seqInput string) (string, string, error) {
	phases, err := os.ReadDir(festivalPath)
	if err != nil {
		return "", "", errors.IO("reading festival directory", err)
	}

	var matches []string
	for _, phase := range phases {
		if !phase.IsDir() || strings.HasPrefix(phase.Name(), ".") {
			continue
		}
		phasePath := filepath.Join(festivalPath, phase.Name())
		sequences, err := os.ReadDir(phasePath)
		if err != nil {
			continue
		}
		for _, seq := range sequences {
			if !seq.IsDir() || strings.HasPrefix(seq.Name(), ".") {
				continue
			}
			if strings.HasPrefix(seq.Name(), seqInput) || seq.Name() == seqInput {
				matches = append(matches, phase.Name()+"/"+seq.Name())
			}
		}
	}

	if len(matches) == 0 {
		return "", "", errors.NotFound("sequence").
			WithField("input", seqInput).
			WithField("hint", "use phase/sequence format like '001/01_api_design'")
	}

	if len(matches) > 1 {
		return "", "", errors.Validation("ambiguous sequence").
			WithField("input", seqInput).
			WithField("matches", strings.Join(matches, ", ")).
			WithField("hint", "use phase/sequence format to disambiguate")
	}

	parts := strings.SplitN(matches[0], "/", 2)
	return filepath.Join(festivalPath, parts[0], parts[1]), matches[0], nil
}

// emitSequenceStatusPlaceholder outputs a placeholder for sequence status changes.
func emitSequenceStatusPlaceholder(display *ui.UI, opts *statusOptions, seqName, newStatus string) error {
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"sequence":   seqName,
			"new_status": newStatus,
			"note":       "sequence status in frontmatter - implementation pending",
		}
		if err := shared.EncodeJSON(os.Stdout, result); err != nil {
			return errors.Wrap(err, "encoding JSON output")
		}
	} else {
		fmt.Println(ui.H1("Sequence Status"))
		fmt.Printf("%s %s\n", ui.Label("Sequence"), ui.Value(seqName, ui.SequenceColor))
		fmt.Printf("%s %s\n", ui.Label("Target"), ui.GetStateStyle(newStatus).Render(newStatus))
		fmt.Println(ui.Dim("Sequence status stored in SEQUENCE_GOAL.md frontmatter"))
		fmt.Println(ui.Dim("Frontmatter editing pending implementation"))
	}
	return nil
}
