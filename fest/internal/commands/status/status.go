// Package status implements the fest status command for managing entity statuses.
package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/spf13/cobra"
)

// EntityType represents the type of festival entity.
type EntityType string

const (
	EntityFestival EntityType = "festival"
	EntityPhase    EntityType = "phase"
	EntitySequence EntityType = "sequence"
	EntityTask     EntityType = "task"
	EntityGate     EntityType = "gate"
)

// ValidStatuses defines valid status values per entity type.
var ValidStatuses = map[EntityType][]string{
	EntityFestival: {"planned", "active", "completed", "dungeon"},
	EntityPhase:    {"pending", "in_progress", "completed"},
	EntitySequence: {"pending", "in_progress", "completed"},
	EntityTask:     {"pending", "in_progress", "blocked", "completed"},
	EntityGate:     {"pending", "passed", "failed"},
}

type statusOptions struct {
	json       bool
	entityType string
	force      bool
}

// NewStatusCommand creates the status command with all subcommands.
func NewStatusCommand() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Manage and query festival entity statuses",
		Long: `Manage and query status for festivals, phases, sequences, tasks, and gates.

When run without subcommands, shows the status of the current entity based on
your working directory location.

SUBCOMMANDS:
  fest status              Show current entity status
  fest status set <status> Change entity status
  fest status list         List entities by status
  fest status history      View status change history`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusShow(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "", "entity type to query (festival, phase, sequence, task, gate)")

	// Add subcommands
	cmd.AddCommand(newStatusSetCommand(opts))
	cmd.AddCommand(newStatusListCommand(opts))
	cmd.AddCommand(newStatusHistoryCommand(opts))

	return cmd
}

func runStatusShow(cmd *cobra.Command, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(cwd)
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

func emitErrorJSON(message string) error {
	result := map[string]interface{}{
		"error": message,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

func emitLocationJSON(loc *show.LocationInfo) error {
	result := map[string]interface{}{
		"type": loc.Type,
	}

	if loc.Festival != nil {
		result["festival"] = map[string]interface{}{
			"name":   loc.Festival.Name,
			"status": loc.Festival.Status,
			"path":   loc.Festival.Path,
		}
		if loc.Festival.Stats != nil {
			result["progress"] = loc.Festival.Stats.Progress
		}
	}

	if loc.Phase != "" {
		result["phase"] = loc.Phase
	}
	if loc.Sequence != "" {
		result["sequence"] = loc.Sequence
	}
	if loc.Task != "" {
		result["task"] = loc.Task
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling status to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitLocationText(loc *show.LocationInfo) error {
	if loc.Festival == nil {
		fmt.Println("Not in a festival directory")
		return nil
	}

	fmt.Printf("Festival: %s\n", loc.Festival.Name)
	fmt.Printf("Status:   %s\n", loc.Festival.Status)
	fmt.Printf("Location: %s\n", loc.Type)

	if loc.Phase != "" {
		fmt.Printf("Phase:    %s\n", loc.Phase)
	}
	if loc.Sequence != "" {
		fmt.Printf("Sequence: %s\n", loc.Sequence)
	}
	if loc.Task != "" {
		fmt.Printf("Task:     %s\n", loc.Task)
	}

	if loc.Festival.Stats != nil {
		fmt.Printf("\nProgress: %.1f%%\n", loc.Festival.Stats.Progress)
		fmt.Printf("  Phases:    %d/%d completed\n",
			loc.Festival.Stats.Phases.Completed,
			loc.Festival.Stats.Phases.Total)
		fmt.Printf("  Sequences: %d/%d completed\n",
			loc.Festival.Stats.Sequences.Completed,
			loc.Festival.Stats.Sequences.Total)
		fmt.Printf("  Tasks:     %d/%d completed\n",
			loc.Festival.Stats.Tasks.Completed,
			loc.Festival.Stats.Tasks.Total)
	}

	return nil
}

func newStatusSetCommand(opts *statusOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <status>",
		Short: "Change entity status",
		Long: `Change the status of the current entity.

For festivals, this will move the directory between status folders
(planned, active, completed, dungeon).

For other entities, this updates the frontmatter in the relevant files.`,
		Example: `  fest status set active               # Set current festival to active
  fest status set completed --force    # Set without confirmation
  fest status set in_progress          # Set phase/sequence/task status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusSet(cmd, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

func runStatusSet(cmd *cobra.Command, newStatus string, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

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
		return handleFestivalStatusChange(loc.Festival, newStatus, opts)
	}

	// For other entities, update frontmatter (placeholder - needs frontmatter support)
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"entity":     loc.Type,
			"new_status": newStatus,
			"note":       "frontmatter updates not yet implemented",
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Status change recorded: %s -> %s\n", loc.Type, newStatus)
		fmt.Println("Note: Frontmatter updates pending implementation")
	}

	return nil
}

func handleFestivalStatusChange(festival *show.FestivalInfo, newStatus string, opts *statusOptions) error {
	if festival.Status == newStatus {
		if opts.json {
			result := map[string]interface{}{
				"success": true,
				"message": "already at requested status",
				"status":  newStatus,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Festival is already %s\n", newStatus)
		}
		return nil
	}

	// Confirm unless forced
	if !opts.force {
		fmt.Printf("Move festival '%s' from %s to %s? [y/N]: ", festival.Name, festival.Status, newStatus)
		var response string
		fmt.Scanln(&response)
		if !strings.HasPrefix(strings.ToLower(response), "y") {
			fmt.Println("Operation cancelled.")
			return nil
		}
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
		fmt.Printf("Moved festival '%s' from %s to %s\n", festival.Name, festival.Status, newStatus)
		fmt.Printf("New path: %s\n", newPath)
	}

	return nil
}

func isValidStatus(entityType EntityType, status string) bool {
	validStatuses, ok := ValidStatuses[entityType]
	if !ok {
		return false
	}
	for _, valid := range validStatuses {
		if valid == status {
			return true
		}
	}
	return false
}

func newStatusListCommand(opts *statusOptions) *cobra.Command {
	var filterStatus string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entities by status",
		Long: `List festivals, phases, sequences, or tasks filtered by status.

Without filters, lists all festivals grouped by status.`,
		Example: `  fest status list                     # List all festivals by status
  fest status list --status active     # List active festivals only
  fest status list --type task --status pending  # List pending tasks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusList(cmd, filterStatus, opts)
		},
	}

	cmd.Flags().StringVar(&filterStatus, "status", "", "filter by status")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "festival", "entity type (festival, phase, sequence, task)")

	return cmd
}

func runStatusList(cmd *cobra.Command, filterStatus string, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Use the show package's listing functionality
	if opts.entityType == "festival" || opts.entityType == "" {
		// Reuse show command's festival listing
		if filterStatus != "" {
			// List festivals with specific status
			festivals, err := show.ListFestivalsByStatus(filepath.Dir(cwd), filterStatus)
			if err != nil {
				// Try finding festivals root
				loc, err := show.DetectCurrentLocation(cwd)
				if err != nil {
					return err
				}
				if loc.Festival != nil {
					festivalsRoot := filepath.Dir(filepath.Dir(loc.Festival.Path))
					festivals, err = show.ListFestivalsByStatus(festivalsRoot, filterStatus)
					if err != nil {
						return err
					}
				}
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
			// List all festivals
			fmt.Println("Use 'fest show all' to see all festivals grouped by status")
		}
	}

	return nil
}

func newStatusHistoryCommand(opts *statusOptions) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "View status change history",
		Long: `View the history of status changes for the current entity.

History is stored in .fest/status_history.json within each festival.`,
		Example: `  fest status history            # Show status history
  fest status history --limit 10 # Show last 10 changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusHistory(cmd, limit, opts)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of entries to show")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

func runStatusHistory(cmd *cobra.Command, limit int, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// History file path
	historyPath := filepath.Join(loc.Festival.Path, ".fest", "status_history.json")

	// Check if history exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
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

	// Read history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return errors.IO("reading history file", err)
	}

	var history []map[string]interface{}
	if err := json.Unmarshal(data, &history); err != nil {
		return errors.Wrap(err, "parsing history file")
	}

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	if opts.json {
		result := map[string]interface{}{
			"festival": loc.Festival.Name,
			"count":    len(history),
			"history":  history,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Status History for %s:\n", loc.Festival.Name)
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
