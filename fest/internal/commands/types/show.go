package types

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/types"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var (
		level        string
		jsonOutput   bool
		showTemplate bool
	)

	cmd := &cobra.Command{
		Use:   "show <type-name>",
		Short: "Show details about a template type",
		Long: `Display detailed information about a specific template type.

Shows the type's level, description, number of markers, template files,
and example usage.

Examples:
  fest types show feature                   # Show feature type details
  fest types show implementation --level phase  # Show phase-level implementation
  fest types show simple --level task --json    # JSON output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(cmd.Context(), args[0], level, jsonOutput, showTemplate)
		},
	}

	cmd.Flags().StringVarP(&level, "level", "l", "", "Filter by level (disambiguate if same name at multiple levels)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().BoolVarP(&showTemplate, "template", "t", false, "Show raw template content")

	return cmd
}

func runShow(ctx context.Context, typeName, levelFilter string, jsonOutput, showTemplate bool) error {
	registry := types.NewRegistry()

	// Discover templates with marker counting
	opts := types.DiscoverOptions{
		BuiltInDir:   getBuiltInTemplatesDir(),
		CustomDir:    getCustomTemplatesDir(),
		CountMarkers: true,
	}

	if err := registry.Discover(ctx, opts); err != nil {
		return err
	}

	// Find the type
	typeInfo, err := findType(registry, typeName, levelFilter)
	if err != nil {
		return err
	}

	if jsonOutput {
		return outputTypeJSON(typeInfo)
	}

	return outputTypeText(typeInfo, showTemplate)
}

func findType(registry *types.Registry, name, levelFilter string) (*types.TypeInfo, error) {
	var matches []*types.TypeInfo

	// Search by level if specified
	if levelFilter != "" {
		level := types.Level(levelFilter)
		if info := registry.FindType(level, name); info != nil {
			return info, nil
		}
		return nil, typeNotFoundError(registry, name, levelFilter)
	}

	// Search all levels
	for _, level := range types.AllLevels() {
		if info := registry.FindType(level, name); info != nil {
			matches = append(matches, info)
		}
	}

	if len(matches) == 0 {
		return nil, typeNotFoundError(registry, name, "")
	}

	if len(matches) > 1 {
		levels := make([]string, len(matches))
		for i, m := range matches {
			levels[i] = string(m.Level)
		}
		return nil, errors.Validation(
			fmt.Sprintf("type '%s' exists at multiple levels: %s", name, strings.Join(levels, ", ")),
		).WithField("hint", "use --level to specify which one")
	}

	return matches[0], nil
}

func typeNotFoundError(registry *types.Registry, name, level string) error {
	// Find similar types for suggestions
	suggestions := findSimilarTypes(registry, name)

	errMsg := fmt.Sprintf("type '%s' not found", name)
	if level != "" {
		errMsg = fmt.Sprintf("type '%s' not found at level '%s'", name, level)
	}

	err := errors.NotFound(errMsg)
	if len(suggestions) > 0 {
		err = err.WithField("similar", strings.Join(suggestions, ", "))
	}
	return err
}

func findSimilarTypes(registry *types.Registry, name string) []string {
	var suggestions []string
	nameLower := strings.ToLower(name)

	for _, t := range registry.AllTypes() {
		typeLower := strings.ToLower(t.Name)
		// Check for substring match or prefix match
		if strings.Contains(typeLower, nameLower) || strings.Contains(nameLower, typeLower) {
			suggestions = append(suggestions, t.QualifiedName())
		}
	}

	// Limit suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions
}

func outputTypeJSON(t *types.TypeInfo) error {
	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTypeText(t *types.TypeInfo, showTemplate bool) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// Header
	fmt.Printf("Type: %s\n", t.Name)
	fmt.Printf("Level: %s\n", t.Level)

	if t.Description != "" {
		fmt.Printf("Description: %s\n", t.Description)
	}

	fmt.Println()

	// Details
	if t.Markers > 0 {
		fmt.Printf("Markers: ~%d\n", t.Markers)
	}

	fmt.Printf("Default: %s\n", boolToYesNo(t.IsDefault))

	if t.IsCustom {
		fmt.Printf("Custom: Yes\n")
	}

	fmt.Println()

	// Example usage
	fmt.Printf("Example Usage:\n")
	fmt.Printf("  fest create %s --type %s\n", t.Level, t.Name)
	fmt.Println()

	// Template files
	if len(t.Templates) > 0 {
		fmt.Printf("Template Files:\n")
		for _, tmpl := range t.Templates {
			fmt.Printf("  - %s\n", tmpl)
		}
	}

	// Source directory
	if t.Source != "" {
		fmt.Printf("\nSource: %s\n", t.Source)
	}

	// Show template content if requested
	if showTemplate && len(t.Templates) > 0 && t.Source != "" {
		fmt.Println()
		fmt.Println("─── Template Content ───")
		for _, tmpl := range t.Templates {
			path := filepath.Join(t.Source, tmpl)
			content, err := os.ReadFile(path)
			if err != nil {
				display.Warning("Could not read template: %s", tmpl)
				continue
			}
			fmt.Printf("\n%s:\n", tmpl)
			fmt.Println(string(content))
		}
	}

	return nil
}

func boolToYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
