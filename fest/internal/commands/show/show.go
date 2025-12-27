// Package show implements the fest show command for displaying festival information.
package show

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

type showOptions struct {
	json bool
}

// NewShowCommand creates the show command with all subcommands.
func NewShowCommand() *cobra.Command {
	opts := &showOptions{}

	cmd := &cobra.Command{
		Use:   "show [status|festival-name]",
		Short: "Display festival information",
		Long: `Display festival information by status or show details of a specific festival.

When run inside a festival directory, shows the current festival's details.
When run with a status argument, lists all festivals with that status.

SUBCOMMANDS:
  fest show              Show current festival (detect from cwd)
  fest show active       List festivals in active/ directory
  fest show planned      List festivals in planned/ directory
  fest show completed    List festivals in completed/ directory
  fest show dungeon      List festivals in dungeon/ directory
  fest show all          List all festivals grouped by status
  fest show <name>       Show details of a specific festival by name`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return runShowCurrent(cmd.Context(), opts)
			}
			return runShow(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	// Add subcommands for status directories
	cmd.AddCommand(newShowActiveCommand(opts))
	cmd.AddCommand(newShowPlannedCommand(opts))
	cmd.AddCommand(newShowCompletedCommand(opts))
	cmd.AddCommand(newShowDungeonCommand(opts))
	cmd.AddCommand(newShowAllCommand(opts))

	return cmd
}

func newShowActiveCommand(opts *showOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "active",
		Short: "List festivals in active/ directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowStatus(cmd.Context(), "active", opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	return cmd
}

func newShowPlannedCommand(opts *showOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "planned",
		Short: "List festivals in planned/ directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowStatus(cmd.Context(), "planned", opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	return cmd
}

func newShowCompletedCommand(opts *showOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completed",
		Short: "List festivals in completed/ directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowStatus(cmd.Context(), "completed", opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	return cmd
}

func newShowDungeonCommand(opts *showOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dungeon",
		Short: "List festivals in dungeon/ directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowStatus(cmd.Context(), "dungeon", opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	return cmd
}

func newShowAllCommand(opts *showOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "all",
		Short: "List all festivals grouped by status",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShowAll(cmd.Context(), opts)
		},
	}
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	return cmd
}

func runShowCurrent(_ context.Context, opts *showOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festival, err := DetectCurrentFestival(cwd)
	if err != nil {
		if errors.Is(err, errors.ErrCodeNotFound) {
			if opts.json {
				return emitShowErrorJSON("not in a festival directory")
			}
			return errors.NotFound("festival").WithOp("show").
				WithField("hint", "navigate to a festival directory or specify a festival name")
		}
		return err
	}

	if opts.json {
		return emitFestivalJSON(festival)
	}
	return emitFestivalText(festival)
}

func runShow(_ context.Context, target string, opts *showOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Find festivals directory
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals directory")
	}
	if festivalsDir == "" {
		return errors.NotFound("festivals directory")
	}

	// Try to find festival by name in any status directory
	festival, err := FindFestivalByName(festivalsDir, target)
	if err != nil {
		if opts.json {
			return emitShowErrorJSON(fmt.Sprintf("festival '%s' not found", target))
		}
		return err
	}

	if opts.json {
		return emitFestivalJSON(festival)
	}
	return emitFestivalText(festival)
}

func runShowStatus(_ context.Context, status string, opts *showOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals directory")
	}
	if festivalsDir == "" {
		return errors.NotFound("festivals directory")
	}

	festivals, err := ListFestivalsByStatus(festivalsDir, status)
	if err != nil {
		return err
	}

	if opts.json {
		return emitFestivalListJSON(status, festivals)
	}
	return emitFestivalListText(status, festivals)
}

func runShowAll(_ context.Context, opts *showOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals directory")
	}
	if festivalsDir == "" {
		return errors.NotFound("festivals directory")
	}

	allFestivals := make(map[string][]*FestivalInfo)
	statusOrder := []string{"active", "planned", "completed", "dungeon"}

	for _, status := range statusOrder {
		festivals, err := ListFestivalsByStatus(festivalsDir, status)
		if err != nil {
			continue // Skip empty or inaccessible directories
		}
		allFestivals[status] = festivals
	}

	if opts.json {
		return emitAllFestivalsJSON(allFestivals, statusOrder)
	}
	return emitAllFestivalsText(allFestivals, statusOrder)
}

func emitShowErrorJSON(message string) error {
	result := map[string]interface{}{
		"error": message,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

func emitFestivalJSON(festival *FestivalInfo) error {
	data, err := json.MarshalIndent(festival, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling festival to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitFestivalText(festival *FestivalInfo) error {
	verbose := shared.IsVerbose()
	fmt.Println(FormatFestivalDetails(festival, verbose))
	return nil
}

func emitFestivalListJSON(status string, festivals []*FestivalInfo) error {
	result := map[string]interface{}{
		"status":    status,
		"count":     len(festivals),
		"festivals": festivals,
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling festival list to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitFestivalListText(status string, festivals []*FestivalInfo) error {
	fmt.Println(FormatFestivalList(status, festivals))
	return nil
}

func emitAllFestivalsJSON(allFestivals map[string][]*FestivalInfo, statusOrder []string) error {
	result := make(map[string]interface{})
	total := 0
	for _, status := range statusOrder {
		festivals := allFestivals[status]
		result[status] = map[string]interface{}{
			"count":     len(festivals),
			"festivals": festivals,
		}
		total += len(festivals)
	}
	result["total"] = total

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling all festivals to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitAllFestivalsText(allFestivals map[string][]*FestivalInfo, statusOrder []string) error {
	fmt.Println(FormatAllFestivals(allFestivals, statusOrder))
	return nil
}
