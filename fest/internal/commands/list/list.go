// Package list implements the fest list command for listing festivals by status.
package list

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
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
			return runList(status, opts)
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

func runList(filterStatus string, opts *listOptions) error {
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
		return listByStatus(festivalsDir, filterStatus, opts)
	}

	// List all statuses
	return listAll(festivalsDir, opts)
}

func listByStatus(festivalsDir, status string, opts *listOptions) error {
	festivals, err := show.ListFestivalsByStatus(festivalsDir, status)
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

	if len(festivals) == 0 {
		fmt.Printf("No %s festivals found.\n", status)
		return nil
	}

	fmt.Printf("%s (%d)\n", strings.ToUpper(status), len(festivals))
	fmt.Println(strings.Repeat("─", 50))
	for _, f := range festivals {
		printFestival(f)
	}

	return nil
}

func listAll(festivalsDir string, opts *listOptions) error {
	result := make(map[string]interface{})
	var totalCount int

	for _, status := range validStatuses {
		festivals, err := show.ListFestivalsByStatus(festivalsDir, status)
		if err != nil {
			continue
		}
		if len(festivals) > 0 || opts.all {
			result[status] = festivalsToMap(festivals)
			totalCount += len(festivals)
		}
	}

	if opts.json {
		result["total"] = totalCount
		return outputJSON(result)
	}

	if totalCount == 0 {
		fmt.Println("No festivals found.")
		fmt.Println("\nCreate a festival with: fest create festival")
		return nil
	}

	for _, status := range validStatuses {
		festivals, _ := show.ListFestivalsByStatus(festivalsDir, status)
		if len(festivals) == 0 && !opts.all {
			continue
		}

		fmt.Printf("\n%s (%d)\n", strings.ToUpper(status), len(festivals))
		fmt.Println(strings.Repeat("─", 50))

		if len(festivals) == 0 {
			fmt.Println("  (none)")
		} else {
			for _, f := range festivals {
				printFestival(f)
			}
		}
	}

	fmt.Printf("\nTotal: %d festivals\n", totalCount)
	return nil
}

func printFestival(f *show.FestivalInfo) {
	progress := ""
	if f.Stats != nil && f.Stats.Progress > 0 {
		progress = fmt.Sprintf(" (%.0f%%)", f.Stats.Progress)
	}
	fmt.Printf("  • %s%s\n", f.Name, progress)
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
