// Package list implements the fest list command for listing festivals by status.
package list

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// Valid status values
var validStatuses = []string{"active", "planned", "completed", "dungeon"}

// Default statuses shown without --all flag
var defaultStatuses = []string{"active", "planned"}

type listOptions struct {
	json     bool
	all      bool
	progress bool
}

// NewListCommand creates the list command for listing festivals by status.
func NewListCommand() *cobra.Command {
	opts := &listOptions{}

	cmd := &cobra.Command{
		Use:   "list [status]",
		Short: "List festivals by status",
		Long: `List festivals filtered by status.

Works from anywhere - finds the festivals workspace automatically.

STATUS can be: active, planned, completed, dungeon

By default, shows only active and planned festivals.
Use --all to include completed and dungeon festivals.`,
		Example: `  fest list              # List active and planned festivals
  fest list --all        # List all festivals (including completed/dungeon)
  fest list active       # List only active festivals
  fest list completed    # List completed festivals
  fest list --json       # Output in JSON format`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			status := ""
			if len(args) > 0 {
				status = strings.ToLower(args[0])
				if !isValidStatus(status) {
					return errors.Validation("invalid status").
						WithField("status", status).
						WithField("valid", strings.Join(validStatuses, ", "))
				}
			}
			return runList(cmd.Context(), status, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.all, "all", false, "include completed and dungeon festivals")
	cmd.Flags().BoolVar(&opts.progress, "progress", false, "show detailed progress for each festival")

	return cmd
}

func isValidStatus(status string) bool {
	for _, v := range validStatuses {
		if v == status {
			return true
		}
	}
	return false
}

func runList(ctx context.Context, filterStatus string, opts *listOptions) error {
	// Find festivals workspace from anywhere
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil || festivalsDir == "" {
		return errors.NotFound("festivals workspace").
			WithField("hint", "run from a project with a festivals/ directory or use 'fest init --register'")
	}

	if filterStatus != "" {
		// List single status
		return listByStatus(ctx, festivalsDir, filterStatus, opts)
	}

	// List all statuses
	return listAll(ctx, festivalsDir, opts)
}

func listByStatus(ctx context.Context, festivalsDir, status string, opts *listOptions) error {
	festivals, err := show.ListFestivalsByStatus(ctx, festivalsDir, status)
	if err != nil {
		return err
	}

	// Fetch detailed progress if requested
	var progressMap map[string]*progress.FestivalProgress
	if opts.progress {
		progressMap = fetchProgressForFestivals(ctx, festivals)
	}

	if opts.json {
		return outputJSON(map[string]interface{}{
			"status":    status,
			"count":     len(festivals),
			"festivals": festivalsToMapWithProgress(festivals, progressMap),
		})
	}

	if opts.progress {
		fmt.Print(show.FormatFestivalListWithProgress(status, festivals, progressMap))
	} else {
		fmt.Print(show.FormatFestivalList(status, festivals))
	}

	return nil
}

func listAll(ctx context.Context, festivalsDir string, opts *listOptions) error {
	result := make(map[string]interface{})
	var totalCount int
	allFestivals := make(map[string][]*show.FestivalInfo)

	// Use all statuses if --all flag, otherwise just active/planned
	statuses := defaultStatuses
	if opts.all {
		statuses = validStatuses
	}

	statusOrder := make([]string, 0, len(statuses))
	var allFestivalsList []*show.FestivalInfo

	for _, status := range statuses {
		festivals, err := show.ListFestivalsByStatus(ctx, festivalsDir, status)
		if err != nil {
			continue
		}
		if len(festivals) > 0 {
			allFestivals[status] = festivals
			statusOrder = append(statusOrder, status)
			totalCount += len(festivals)
			allFestivalsList = append(allFestivalsList, festivals...)
		}
	}

	// Fetch detailed progress if requested
	var progressMap map[string]*progress.FestivalProgress
	if opts.progress {
		progressMap = fetchProgressForFestivals(ctx, allFestivalsList)
	}

	if opts.json {
		for status, festivals := range allFestivals {
			result[status] = festivalsToMapWithProgress(festivals, progressMap)
		}
		result["total"] = totalCount
		return outputJSON(result)
	}

	if totalCount == 0 {
		fmt.Println(ui.Warning("No festivals found."))
		fmt.Println(ui.Info("Create a festival with: fest create festival"))
		return nil
	}

	if opts.progress {
		fmt.Print(show.FormatAllFestivalsWithProgress(allFestivals, statusOrder, progressMap))
	} else {
		fmt.Print(show.FormatAllFestivals(allFestivals, statusOrder))
	}
	return nil
}

func festivalsToMap(festivals []*show.FestivalInfo) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(festivals))
	for _, f := range festivals {
		m := map[string]interface{}{
			"name":   f.Name,
			"path":   f.Path,
			"status": f.Status,
		}
		if f.Stats != nil {
			m["progress"] = f.Stats.Progress
		}
		result = append(result, m)
	}
	return result
}

func outputJSON(data interface{}) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

// fetchProgressForFestivals fetches detailed progress for each festival.
// Returns a map from festival path to progress data.
// Silently skips festivals where progress cannot be fetched.
func fetchProgressForFestivals(ctx context.Context, festivals []*show.FestivalInfo) map[string]*progress.FestivalProgress {
	progressMap := make(map[string]*progress.FestivalProgress)
	for _, f := range festivals {
		mgr, err := progress.NewManager(ctx, f.Path)
		if err != nil {
			continue // Silently skip
		}
		prog, err := mgr.GetFestivalProgress(ctx, f.Path)
		if err != nil {
			continue // Silently skip
		}
		progressMap[f.Path] = prog
	}
	return progressMap
}

// festivalsToMapWithProgress converts festivals to map with optional detailed progress.
func festivalsToMapWithProgress(festivals []*show.FestivalInfo, progressMap map[string]*progress.FestivalProgress) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(festivals))
	for _, f := range festivals {
		m := map[string]interface{}{
			"name":   f.Name,
			"path":   f.Path,
			"status": f.Status,
		}
		if f.Stats != nil {
			m["progress"] = f.Stats.Progress
		}
		// Add detailed progress if available
		if progressMap != nil {
			if prog, ok := progressMap[f.Path]; ok && prog != nil && prog.Overall != nil {
				m["tasks"] = map[string]interface{}{
					"total":       prog.Overall.Total,
					"completed":   prog.Overall.Completed,
					"in_progress": prog.Overall.InProgress,
					"blocked":     prog.Overall.Blocked,
					"pending":     prog.Overall.Pending,
				}
				m["time_spent_minutes"] = prog.Overall.TimeSpentMin
			}
		}
		result = append(result, m)
	}
	return result
}
