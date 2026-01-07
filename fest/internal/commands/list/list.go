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
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// Valid status values
var validStatuses = []string{"active", "planned", "completed", "dungeon"}

type listOptions struct {
	json bool
	all  bool
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
If no status is provided, lists all festivals grouped by status.`,
		Example: `  fest list              # List all festivals grouped by status
  fest list active       # List only active festivals
  fest list planned      # List only planned festivals
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
	cmd.Flags().BoolVar(&opts.all, "all", false, "include empty status categories")

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

	if opts.json {
		return outputJSON(map[string]interface{}{
			"status":    status,
			"count":     len(festivals),
			"festivals": festivalsToMap(festivals),
		})
	}

	fmt.Print(show.FormatFestivalList(status, festivals))

	return nil
}

func listAll(ctx context.Context, festivalsDir string, opts *listOptions) error {
	result := make(map[string]interface{})
	var totalCount int
	allFestivals := make(map[string][]*show.FestivalInfo)
	statusOrder := make([]string, 0, len(validStatuses))

	for _, status := range validStatuses {
		festivals, err := show.ListFestivalsByStatus(ctx, festivalsDir, status)
		if err != nil {
			continue
		}
		if len(festivals) > 0 || opts.all {
			allFestivals[status] = festivals
			statusOrder = append(statusOrder, status)
			result[status] = festivalsToMap(festivals)
			totalCount += len(festivals)
		}
	}

	if opts.json {
		result["total"] = totalCount
		return outputJSON(result)
	}

	if totalCount == 0 {
		fmt.Println(ui.Warning("No festivals found."))
		fmt.Println(ui.Info("Create a festival with: fest create festival"))
		return nil
	}

	fmt.Print(show.FormatAllFestivals(allFestivals, statusOrder))
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
