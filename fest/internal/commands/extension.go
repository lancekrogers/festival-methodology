package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/lancekrogers/festival-methodology/fest/internal/extensions"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewExtensionCommand creates the extension command group
func NewExtensionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "extension",
		Short: "Manage methodology extensions",
		Long: `Manage and view methodology extension packs.

Extensions are loaded from three sources with the following precedence:
1. Project-local: .festival/extensions/ (highest priority)
2. User config: ~/.config/fest/active/festivals/.festival/extensions/
3. Built-in: ~/.config/fest/festivals/.festival/extensions/ (lowest priority)

Higher priority sources override lower ones when extensions have the same name.`,
	}

	cmd.AddCommand(newExtensionListCommand())
	cmd.AddCommand(newExtensionInfoCommand())

	return cmd
}

func newExtensionListCommand() *cobra.Command {
	var jsonOutput bool
	var source string
	var extType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all loaded extensions",
		Long:  `List all methodology extensions loaded from project, user, and built-in sources.`,
		Example: `  fest extension list
  fest extension list --source user
  fest extension list --type workflow
  fest extension list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtensionList(jsonOutput, source, extType)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().StringVar(&source, "source", "", "filter by source (project, user, built-in)")
	cmd.Flags().StringVar(&extType, "type", "", "filter by type (workflow, template, agent)")

	return cmd
}

func runExtensionList(jsonOutput bool, source, extType string) error {
	display := ui.New(noColor, verbose)

	// Get festival root if available
	cwd, _ := os.Getwd()
	festivalRoot, _ := gates.FindFestivalRoot(cwd)

	// Load extensions
	loader := extensions.NewExtensionLoader()
	if err := loader.LoadAll(festivalRoot); err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	// Get extensions with optional filtering
	var exts []*extensions.Extension
	if source != "" {
		exts = loader.ListBySource(source)
	} else if extType != "" {
		exts = loader.ListByType(extType)
	} else {
		exts = loader.List()
	}

	// Sort by name
	sort.Slice(exts, func(i, j int) bool {
		return exts[i].Name < exts[j].Name
	})

	if jsonOutput {
		output := map[string]any{
			"extensions": exts,
			"count":      len(exts),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	if len(exts) == 0 {
		display.Info("No extensions found")
		return nil
	}

	display.Info("Loaded extensions (%d):", len(exts))
	fmt.Println()

	for _, ext := range exts {
		fmt.Printf("  %s", ext.Name)
		if ext.Version != "" {
			fmt.Printf(" (v%s)", ext.Version)
		}
		fmt.Printf(" [%s]\n", ext.Source)

		if ext.Description != "" && verbose {
			fmt.Printf("    %s\n", ext.Description)
		}
		if ext.Type != "" {
			fmt.Printf("    Type: %s\n", ext.Type)
		}
	}

	return nil
}

func newExtensionInfoCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info <name>",
		Short: "Show extension details",
		Long:  `Show detailed information about a specific extension.`,
		Example: `  fest extension info ai-review
  fest extension info workflow-patterns --json`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExtensionInfo(args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")

	return cmd
}

func runExtensionInfo(name string, jsonOutput bool) error {
	display := ui.New(noColor, verbose)

	// Get festival root if available
	cwd, _ := os.Getwd()
	festivalRoot, _ := gates.FindFestivalRoot(cwd)

	// Load extensions
	loader := extensions.NewExtensionLoader()
	if err := loader.LoadAll(festivalRoot); err != nil {
		return fmt.Errorf("failed to load extensions: %w", err)
	}

	ext := loader.Get(name)
	if ext == nil {
		if jsonOutput {
			output := map[string]any{
				"error": fmt.Sprintf("extension '%s' not found", name),
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(output)
		}
		return fmt.Errorf("extension '%s' not found", name)
	}

	if jsonOutput {
		// List files
		files, _ := ext.ListFiles()
		output := map[string]any{
			"extension": ext,
			"files":     files,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	display.Info("Extension: %s", ext.Name)
	if ext.Version != "" {
		display.Info("Version: %s", ext.Version)
	}
	if ext.Description != "" {
		display.Info("Description: %s", ext.Description)
	}
	if ext.Author != "" {
		display.Info("Author: %s", ext.Author)
	}
	if ext.Type != "" {
		display.Info("Type: %s", ext.Type)
	}
	if len(ext.Tags) > 0 {
		display.Info("Tags: %v", ext.Tags)
	}
	display.Info("Source: %s", ext.Source)
	display.Info("Path: %s", ext.Path)

	// List files
	files, err := ext.ListFiles()
	if err == nil && len(files) > 0 {
		fmt.Println()
		display.Info("Files (%d):", len(files))
		for _, f := range files {
			fmt.Printf("  %s\n", f)
		}
	}

	return nil
}
