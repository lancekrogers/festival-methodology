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
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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
	loc, err := show.DetectCurrentLocation(festivalPath)
	if err != nil {
		if opts.json {
			return emitErrorJSON("not in a festival directory")
		}
		return err
	}

	if opts.json {
		return emitLocationJSON(loc)
	}
	return emitLocationText(loc)
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

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(festivalPath)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// Determine entity type and validate status
	entityType := EntityType(loc.Type)
	if !isValidStatus(entityType, newStatus) {
		validOptions := ValidStatuses[entityType]
		return errors.Validation("invalid status").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Handle festival status changes (directory moves)
	if loc.Type == "festival" {
		return handleFestivalStatusChange(ctx, display, loc.Festival, newStatus, opts)
	}

	// For other entities, update frontmatter (placeholder - needs frontmatter support)
	return emitStatusSetPlaceholder(display, opts, loc.Type, newStatus)
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
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
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
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
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
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
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
		festivals, err := show.ListFestivalsByStatus(festivalsRoot, filterStatus)
		if err != nil {
			return err
		}

		if opts.json {
			result := map[string]interface{}{
				"status":    filterStatus,
				"count":     len(festivals),
				"festivals": festivals,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
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
	loc, err := show.DetectCurrentLocation(festivalPath)
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

	loc, err := show.DetectCurrentLocation(festivalPath)
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
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
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
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
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
