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

// NewTaskDefaultsShowCommand creates the "show" subcommand
func NewTaskDefaultsShowCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current fest.yaml configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTaskDefaultsShow(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output JSON")

	return cmd
}

func runTaskDefaultsShow(jsonOutput bool) error {
	display := ui.New(noColor, verbose)

	cwd, _ := os.Getwd()
	festivalRoot, err := findFestivalRoot(cwd)
	if err != nil {
		if jsonOutput {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{
				"ok":     false,
				"action": "task_defaults_show",
				"error":  "not in a festival directory",
			})
		}
		return fmt.Errorf("not in a festival directory")
	}

	cfg, err := config.LoadFestivalConfig(festivalRoot)
	if err != nil {
		return err
	}

	if jsonOutput {
		result := map[string]any{
			"ok":     true,
			"action": "task_defaults_show",
			"config": cfg,
			"path":   filepath.Join(festivalRoot, config.FestivalConfigFileName),
			"exists": config.FestivalConfigExists(festivalRoot),
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	display.Info("Festival config: %s", filepath.Join(festivalRoot, config.FestivalConfigFileName))
	display.Info("Config exists: %v", config.FestivalConfigExists(festivalRoot))
	display.Info("")
	display.Info("Quality Gates:")
	display.Info("  Enabled: %v", cfg.QualityGates.Enabled)
	display.Info("  Auto-append: %v", cfg.QualityGates.AutoAppend)
	display.Info("  Tasks:")
	for _, task := range cfg.QualityGates.Tasks {
		status := "enabled"
		if !task.Enabled {
			status = "disabled"
		}
		display.Info("    - %s (%s): %s", task.ID, task.Template, status)
	}
	display.Info("")
	display.Info("Excluded patterns: %v", cfg.ExcludedPatterns)

	return nil
}
