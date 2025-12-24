package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewTaskDefaultsInitCommand creates the "init" subcommand
func NewTaskDefaultsInitCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create a default fest.yaml file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskDefaultsInit(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output JSON")

	return cmd
}

func runTaskDefaultsInit(jsonOutput bool) error {
	display := ui.New(noColor, verbose)

	cwd, _ := os.Getwd()
	festivalRoot, err := findFestivalRoot(cwd)
	if err != nil {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"ok":     false,
				"action": "task_defaults_init",
				"error":  "not in a festival directory",
			})
		}
		return fmt.Errorf("not in a festival directory")
	}

	configPath := filepath.Join(festivalRoot, config.FestivalConfigFileName)

	// Check if already exists
	if config.FestivalConfigExists(festivalRoot) {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"ok":     false,
				"action": "task_defaults_init",
				"error":  "fest.yaml already exists",
				"path":   configPath,
			})
		}
		return fmt.Errorf("fest.yaml already exists at %s", configPath)
	}

	// Create default config
	cfg := config.DefaultFestivalConfig()
	if err := config.SaveFestivalConfig(festivalRoot, cfg); err != nil {
		return err
	}

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]any{
			"ok":      true,
			"action":  "task_defaults_init",
			"created": configPath,
		})
	}

	display.Success("Created fest.yaml at %s", configPath)
	display.Info("Edit this file to customize quality gate behavior.")

	return nil
}
