package types

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/types"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
	var (
		level      string
		jsonOutput bool
		showAll    bool
		customOnly bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available template types",
		Long: `List all template types available at each festival level.

Types are discovered from:
  - Built-in templates (~/.config/fest/templates/)
  - Custom templates (.festival/templates/ in a festival)

Examples:
  fest types list                  # List all types grouped by level
  fest types list --level task     # List task-level types only
  fest types list --custom         # Show only custom types
  fest types list --all            # Include marker counts
  fest types list --json           # Machine-readable output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd.Context(), level, jsonOutput, showAll, customOnly)
		},
	}

	cmd.Flags().StringVarP(&level, "level", "l", "", "Filter by level (festival, phase, sequence, task)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&showAll, "all", "a", false, "Show additional details including marker counts")
	cmd.Flags().BoolVarP(&customOnly, "custom", "c", false, "Show only custom (user-defined) types")

	return cmd
}

func runList(ctx context.Context, levelFilter string, jsonOutput, showAll, customOnly bool) error {
	registry := types.NewRegistry()

	// Discover templates
	opts := types.DiscoverOptions{
		BuiltInDir:   getBuiltInTemplatesDir(),
		CustomDir:    getCustomTemplatesDir(),
		CountMarkers: showAll,
	}

	if err := registry.Discover(ctx, opts); err != nil {
		return err
	}

	// Filter by level if specified
	var filteredTypes []types.TypeInfo
	if levelFilter != "" {
		level := types.Level(levelFilter)
		filteredTypes = registry.TypesForLevel(level)
	} else {
		filteredTypes = registry.AllTypes()
	}

	// Filter by custom if specified
	if customOnly {
		filteredTypes = filterCustomTypes(filteredTypes)
	}

	if jsonOutput {
		return outputJSON(filteredTypes)
	}

	return outputText(registry, levelFilter, showAll, customOnly)
}

func filterCustomTypes(typeInfos []types.TypeInfo) []types.TypeInfo {
	result := []types.TypeInfo{}
	for _, t := range typeInfos {
		if t.IsCustom {
			result = append(result, t)
		}
	}
	return result
}

func outputJSON(typeInfos []types.TypeInfo) error {
	data, err := json.MarshalIndent(typeInfos, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputText(registry *types.Registry, levelFilter string, showAll, customOnly bool) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	if registry.TypeCount() == 0 {
		display.Info("No template types found.")
		display.Info("Tip: Run 'fest system sync' to download built-in templates.")
		return nil
	}

	if levelFilter != "" {
		level := types.Level(levelFilter)
		typeInfos := registry.TypesForLevel(level)
		if customOnly {
			typeInfos = filterCustomTypes(typeInfos)
		}
		if len(typeInfos) == 0 {
			display.Info("No %s types found.", levelFilter)
			if customOnly {
				display.Info("Tip: Create custom templates in .festival/templates/")
			}
			return nil
		}
		fmt.Printf("%s Types:\n\n", capitalize(levelFilter))
		printTypes(typeInfos, showAll)
	} else {
		// Print by level
		foundAny := false
		for _, level := range types.AllLevels() {
			typeInfos := registry.TypesForLevel(level)
			if customOnly {
				typeInfos = filterCustomTypes(typeInfos)
			}
			if len(typeInfos) == 0 {
				continue
			}
			foundAny = true
			fmt.Printf("%s Types:\n", capitalize(string(level)))
			printTypes(typeInfos, showAll)
			fmt.Println()
		}
		if !foundAny && customOnly {
			display.Info("No custom types found.")
			display.Info("Tip: Create custom templates in .festival/templates/")
		}
	}

	return nil
}

func printTypes(typeInfos []types.TypeInfo, showAll bool) {
	for _, t := range typeInfos {
		suffix := ""
		if t.IsDefault {
			suffix = " (default)"
		} else if t.IsCustom {
			suffix = " (custom)"
		}

		if showAll && t.Markers > 0 {
			fmt.Printf("  %-20s %d markers%s\n", t.Name, t.Markers, suffix)
		} else {
			fmt.Printf("  %s%s\n", t.Name, suffix)
		}
	}
}

func capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func getBuiltInTemplatesDir() string {
	// Check environment variable first
	if dir := os.Getenv("FEST_TEMPLATES_DIR"); dir != "" {
		return dir
	}
	// Default to standard config location
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "fest", "templates")
}

func getCustomTemplatesDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Look for .festival/templates in current festival
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return ""
	}

	return filepath.Join(festivalPath, ".festival", "templates")
}
