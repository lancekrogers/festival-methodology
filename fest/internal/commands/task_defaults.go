package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewTaskCommand creates the "task" parent command
func NewTaskCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Task management commands",
		Long:  `Commands for managing tasks including quality gate defaults.`,
	}

	// Add subcommands
	cmd.AddCommand(NewTaskDefaultsCommand())

	return cmd
}

// NewTaskDefaultsCommand creates the "defaults" subcommand
func NewTaskDefaultsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "defaults",
		Short: "Manage quality gate default tasks",
		Long: `Manage quality gate default tasks for festivals.

Quality gates are standard tasks (testing, code review, iterate) that should
appear at the end of every implementation sequence.

Use fest.yaml in your festival root to customize which quality gates are included.`,
	}

	// Add subcommands
	cmd.AddCommand(NewTaskDefaultsSyncCommand())
	cmd.AddCommand(NewTaskDefaultsAddCommand())
	cmd.AddCommand(NewTaskDefaultsShowCommand())
	cmd.AddCommand(NewTaskDefaultsInitCommand())

	return cmd
}

// Sync command options
type taskDefaultsSyncOptions struct {
	path        string
	dryRun      bool
	approve     bool
	interactive bool
	force       bool
	jsonOutput  bool
	verboseFlag bool
}

// Sync result structure
type taskDefaultsSyncResult struct {
	OK      bool              `json:"ok"`
	Action  string            `json:"action"`
	DryRun  bool              `json:"dry_run"`
	Changes []syncChange      `json:"changes,omitempty"`
	Summary syncSummary       `json:"summary,omitempty"`
	Errors  []map[string]any  `json:"errors,omitempty"`
	Warnings []string         `json:"warnings,omitempty"`
}

type syncChange struct {
	Type     string `json:"type"`
	Path     string `json:"path"`
	Template string `json:"template,omitempty"`
	Reason   string `json:"reason,omitempty"`
}

type syncSummary struct {
	TotalSequences   int `json:"total_sequences"`
	SequencesUpdated int `json:"sequences_updated"`
	FilesCreated     int `json:"files_created"`
	FilesSkipped     int `json:"files_skipped"`
}

// NewTaskDefaultsSyncCommand creates the "sync" subcommand
func NewTaskDefaultsSyncCommand() *cobra.Command {
	opts := &taskDefaultsSyncOptions{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "Sync quality gate tasks to all sequences",
		Long: `Sync quality gate tasks to all implementation sequences in a festival.

By default, this runs in dry-run mode showing what would change.
Use --approve to actually apply the changes.

Quality gates are only added to sequences not matching excluded_patterns
in fest.yaml (default: *_planning, *_research, *_requirements).

Modified files are detected and skipped unless --force is used.`,
		Example: `  # Preview changes (dry-run is default)
  fest task defaults sync

  # Apply changes
  fest task defaults sync --approve

  # Force overwrite modified files
  fest task defaults sync --approve --force

  # JSON output for automation
  fest task defaults sync --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			// Default to dry-run unless approve is set
			if !opts.approve {
				opts.dryRun = true
			}
			return runTaskDefaultsSync(opts)
		},
	}

	cmd.Flags().StringVar(&opts.path, "path", ".", "Festival root path")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", true, "Preview changes without applying (default)")
	cmd.Flags().BoolVar(&opts.approve, "approve", false, "Apply changes")
	cmd.Flags().BoolVar(&opts.interactive, "interactive", false, "Prompt for each change")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite modified files")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output JSON")
	cmd.Flags().BoolVar(&opts.verboseFlag, "verbose", false, "Verbose logging")

	return cmd
}

func runTaskDefaultsSync(opts *taskDefaultsSyncOptions) error {
	display := ui.New(noColor, verbose || opts.verboseFlag)

	// Resolve festival path
	absPath, err := filepath.Abs(opts.path)
	if err != nil {
		return emitTaskDefaultsSyncError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Find festival root
	festivalRoot, err := findFestivalRoot(absPath)
	if err != nil {
		return emitTaskDefaultsSyncError(opts, fmt.Errorf("not in a festival directory: %w", err))
	}

	// Load festival config
	cfg, err := config.LoadFestivalConfig(festivalRoot)
	if err != nil {
		return emitTaskDefaultsSyncError(opts, fmt.Errorf("failed to load festival config: %w", err))
	}

	if !cfg.QualityGates.Enabled {
		return emitTaskDefaultsSyncResult(opts, taskDefaultsSyncResult{
			OK:       true,
			Action:   "task_defaults_sync",
			DryRun:   opts.dryRun,
			Warnings: []string{"Quality gates are disabled in fest.yaml"},
		})
	}

	// Find all sequences
	sequences, err := findImplementationSequences(festivalRoot, cfg)
	if err != nil {
		return emitTaskDefaultsSyncError(opts, fmt.Errorf("failed to find sequences: %w", err))
	}

	// Get template root
	tmplRoot, err := tpl.LocalTemplateRoot(festivalRoot)
	if err != nil {
		return emitTaskDefaultsSyncError(opts, fmt.Errorf("failed to find template root: %w", err))
	}

	// Process each sequence
	var changes []syncChange
	var warnings []string
	summary := syncSummary{TotalSequences: len(sequences)}

	enabledTasks := cfg.GetEnabledTasks()

	for _, seqPath := range sequences {
		seqChanges, seqWarnings, err := syncSequenceDefaults(seqPath, enabledTasks, tmplRoot, opts)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Sequence %s: %v", seqPath, err))
			continue
		}

		if len(seqChanges) > 0 {
			summary.SequencesUpdated++
		}

		for _, c := range seqChanges {
			if c.Type == "create" {
				summary.FilesCreated++
			} else if c.Type == "skip" {
				summary.FilesSkipped++
			}
		}

		changes = append(changes, seqChanges...)
		warnings = append(warnings, seqWarnings...)
	}

	// Output result
	result := taskDefaultsSyncResult{
		OK:       true,
		Action:   "task_defaults_sync",
		DryRun:   opts.dryRun,
		Changes:  changes,
		Summary:  summary,
		Warnings: warnings,
	}

	if opts.jsonOutput {
		return emitTaskDefaultsSyncResult(opts, result)
	}

	// Human-readable output
	if opts.dryRun {
		display.Info("Dry-run mode (use --approve to apply changes)")
	}

	display.Info("Found %d sequences, %d will be updated", summary.TotalSequences, summary.SequencesUpdated)

	for _, c := range changes {
		switch c.Type {
		case "create":
			display.Success("  + %s", c.Path)
		case "skip":
			display.Warning("  ~ Skipped %s (%s)", c.Path, c.Reason)
		case "exists":
			if verbose {
				display.Info("  = %s (already exists)", c.Path)
			}
		}
	}

	if len(warnings) > 0 {
		for _, w := range warnings {
			display.Warning("  Warning: %s", w)
		}
	}

	display.Info("")
	display.Info("Summary: %d files created, %d skipped", summary.FilesCreated, summary.FilesSkipped)

	return nil
}

func syncSequenceDefaults(seqPath string, tasks []config.QualityGateTask, tmplRoot string, opts *taskDefaultsSyncOptions) ([]syncChange, []string, error) {
	var changes []syncChange
	var warnings []string

	// Get existing tasks in sequence
	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, nil, err
	}

	// Find highest task number
	maxNum := 0
	existingTasks := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if entry.Name() == "SEQUENCE_GOAL.md" {
			continue
		}
		existingTasks[entry.Name()] = true
		num := festival.ParseTaskNumber(entry.Name())
		if num > maxNum {
			maxNum = num
		}
	}

	// Load catalog for template rendering
	catalog, _ := tpl.LoadCatalog(tmplRoot)
	mgr := tpl.NewManager()

	// Add each quality gate task
	for i, task := range tasks {
		taskNum := maxNum + i + 1
		taskFileName := tpl.FormatTaskID(taskNum, task.ID) // e.g., "04_testing_and_verify.md"
		taskPath := filepath.Join(seqPath, taskFileName)

		// Check if task already exists (by ID pattern)
		taskExists := false
		for existingName := range existingTasks {
			if strings.Contains(existingName, task.ID) {
				taskExists = true
				break
			}
		}

		if taskExists {
			changes = append(changes, syncChange{
				Type:   "exists",
				Path:   taskPath,
				Reason: "task_already_exists",
			})
			continue
		}

		// Check if file exists (could be renamed)
		if _, err := os.Stat(taskPath); err == nil {
			// File exists, check if modified
			// TODO: Implement checksum tracking
			if !opts.force {
				changes = append(changes, syncChange{
					Type:   "skip",
					Path:   taskPath,
					Reason: "file_exists",
				})
				continue
			}
		}

		// Create the task
		if !opts.dryRun {
			// Build context
			ctx := tpl.NewContext()
			ctx.SetTask(taskNum, task.ID)
			if task.Customizations != nil {
				for k, v := range task.Customizations {
					ctx.SetCustom(k, v)
				}
			}

			// Render template
			var content string
			if catalog != nil {
				content, _ = mgr.RenderByID(catalog, task.Template, ctx)
			}

			// Fallback to direct file load
			if content == "" {
				tpath := filepath.Join(tmplRoot, task.Template+".md")
				if _, err := os.Stat(tpath); err == nil {
					loader := tpl.NewLoader()
					t, err := loader.Load(tpath)
					if err == nil {
						if strings.Contains(t.Content, "{{") {
							content, _ = mgr.Render(t, ctx)
						} else {
							content = t.Content
						}
					}
				}
			}

			// Use default content if template not found
			if content == "" {
				content = generateDefaultQualityGateContent(task)
			}

			// Write file
			if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
				warnings = append(warnings, fmt.Sprintf("Failed to write %s: %v", taskPath, err))
				continue
			}
		}

		changes = append(changes, syncChange{
			Type:     "create",
			Path:     taskPath,
			Template: task.Template,
		})
	}

	return changes, warnings, nil
}

func generateDefaultQualityGateContent(task config.QualityGateTask) string {
	name := task.Name
	if name == "" {
		name = strings.ReplaceAll(task.ID, "_", " ")
		name = strings.Title(name)
	}

	return fmt.Sprintf(`# Task: %s

## Objective

%s

## Requirements

- [ ] Complete this quality gate task
- [ ] Verify all requirements are met
- [ ] Document any findings

## Definition of Done

- [ ] Task completed successfully
- [ ] All checks pass
- [ ] Ready to proceed

## Notes

[Add notes here]
`, name, name)
}

func findFestivalRoot(startPath string) (string, error) {
	path := startPath
	for {
		// Check for festival markers (FESTIVAL_OVERVIEW.md or fest.yaml)
		if _, err := os.Stat(filepath.Join(path, "FESTIVAL_OVERVIEW.md")); err == nil {
			return path, nil
		}
		if _, err := os.Stat(filepath.Join(path, "fest.yaml")); err == nil {
			return path, nil
		}
		if _, err := os.Stat(filepath.Join(path, "FESTIVAL_GOAL.md")); err == nil {
			return path, nil
		}

		// Move up
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "", fmt.Errorf("no festival root found")
}

func findImplementationSequences(festivalRoot string, cfg *config.FestivalConfig) ([]string, error) {
	var sequences []string

	// Walk through phases
	entries, err := os.ReadDir(festivalRoot)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a phase (starts with number)
		if !festival.IsPhase(entry.Name()) {
			continue
		}

		phasePath := filepath.Join(festivalRoot, entry.Name())

		// Walk through sequences in phase
		seqEntries, err := os.ReadDir(phasePath)
		if err != nil {
			continue
		}

		for _, seqEntry := range seqEntries {
			if !seqEntry.IsDir() {
				continue
			}

			// Check if it's a sequence (starts with number)
			if !festival.IsSequence(seqEntry.Name()) {
				continue
			}

			// Check excluded patterns
			if cfg.IsSequenceExcluded(seqEntry.Name()) {
				continue
			}

			sequences = append(sequences, filepath.Join(phasePath, seqEntry.Name()))
		}
	}

	return sequences, nil
}

func emitTaskDefaultsSyncError(opts *taskDefaultsSyncOptions, err error) error {
	if opts.jsonOutput {
		return emitTaskDefaultsSyncResult(opts, taskDefaultsSyncResult{
			OK:     false,
			Action: "task_defaults_sync",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
	}
	return err
}

func emitTaskDefaultsSyncResult(opts *taskDefaultsSyncOptions, result taskDefaultsSyncResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

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
				"ok":      false,
				"action":  "task_defaults_init",
				"error":   "fest.yaml already exists",
				"path":    configPath,
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
