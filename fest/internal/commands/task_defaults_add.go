package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// Add command options
type taskDefaultsAddOptions struct {
	sequence   string
	dryRun     bool
	approve    bool
	jsonOutput bool
}

// NewTaskDefaultsAddCommand creates the "add" subcommand
func NewTaskDefaultsAddCommand() *cobra.Command {
	opts := &taskDefaultsAddOptions{}

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add quality gate tasks to a specific sequence",
		Long: `Add quality gate tasks to a specific sequence.

This adds the configured quality gate tasks (testing, code review, iterate)
to the end of the specified sequence.`,
		Example: `  # Preview what would be added
  fest task defaults add --sequence ./002_IMPLEMENT/01_api

  # Apply changes
  fest task defaults add --sequence ./002_IMPLEMENT/01_api --approve`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.sequence == "" {
				return fmt.Errorf("--sequence is required")
			}
			if !opts.approve {
				opts.dryRun = true
			}
			return runTaskDefaultsAdd(opts)
		},
	}

	cmd.Flags().StringVar(&opts.sequence, "sequence", "", "Path to target sequence (required)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", true, "Preview changes (default)")
	cmd.Flags().BoolVar(&opts.approve, "approve", false, "Apply changes")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output JSON")

	return cmd
}

func runTaskDefaultsAdd(opts *taskDefaultsAddOptions) error {
	display := ui.New(noColor, verbose)

	// Resolve sequence path
	absPath, err := filepath.Abs(opts.sequence)
	if err != nil {
		return fmt.Errorf("invalid sequence path: %w", err)
	}

	// Find festival root
	festivalRoot, err := findFestivalRoot(absPath)
	if err != nil {
		return fmt.Errorf("not in a festival directory: %w", err)
	}

	// Load config
	cfg, err := config.LoadFestivalConfig(festivalRoot)
	if err != nil {
		return fmt.Errorf("failed to load festival config: %w", err)
	}

	// Get template root
	tmplRoot, err := tpl.LocalTemplateRoot(festivalRoot)
	if err != nil {
		return fmt.Errorf("failed to find template root: %w", err)
	}

	// Sync this sequence only
	syncOpts := &taskDefaultsSyncOptions{
		dryRun:     opts.dryRun,
		jsonOutput: opts.jsonOutput,
	}

	changes, warnings, err := syncSequenceDefaults(absPath, cfg.GetEnabledTasks(), tmplRoot, syncOpts)
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		result := taskDefaultsSyncResult{
			OK:       true,
			Action:   "task_defaults_add",
			DryRun:   opts.dryRun,
			Changes:  changes,
			Warnings: warnings,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	if opts.dryRun {
		display.Info("Dry-run mode (use --approve to apply)")
	}

	for _, c := range changes {
		switch c.Type {
		case "create":
			display.Success("  + %s", c.Path)
		case "skip":
			display.Warning("  ~ Skipped %s (%s)", c.Path, c.Reason)
		}
	}

	return nil
}
