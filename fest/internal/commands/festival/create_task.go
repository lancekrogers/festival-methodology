package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreateTaskOptions holds options for the create task command.
type CreateTaskOptions struct {
	After       int
	Names       []string
	Path        string
	VarsFile    string
	Markers     string // Inline JSON with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing
	DryRun      bool   // Show markers without creating file
	JSONOutput  bool
	AgentMode   bool // Strict mode for AI agents
}

type createTaskResult struct {
	OK            bool                     `json:"ok"`
	Action        string                   `json:"action"`
	Task          map[string]interface{}   `json:"task,omitempty"`
	Created       []string                 `json:"created,omitempty"`
	Renumber      []string                 `json:"renumbered,omitempty"`
	Markers       []map[string]interface{} `json:"markers,omitempty"`
	MarkersFilled int                      `json:"markers_filled,omitempty"`
	MarkersTotal  int                      `json:"markers_total,omitempty"`
	Validation    *ValidationSummary       `json:"validation,omitempty"`
	Errors        []map[string]any         `json:"errors,omitempty"`
	Warnings      []string                 `json:"warnings,omitempty"`
	Suggestions   []string                 `json:"suggestions,omitempty"`
}

// NewCreateTaskCommand adds 'create task'
func NewCreateTaskCommand() *cobra.Command {
	opts := &CreateTaskOptions{}
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Insert a new task file in a sequence",
		Long: `Create new task file(s) with automatic numbering and template rendering.

IMPORTANT: AI agents execute TASK FILES, not goals. If your sequences only
have SEQUENCE_GOAL.md without task files, agents won't know HOW to execute.

BATCH CREATION: Use multiple --name flags to create sequential tasks at once.
This avoids the common mistake of all tasks getting numbered 01_.

TEMPLATE VARIABLES (automatically set from --name):
  {{ task_name }}            Name of the task
  {{ task_number }}          Sequential number (01, 02, ...)
  {{ task_id }}              Full filename (e.g., "01_design.md")
  {{ parent_sequence_id }}   Parent sequence ID
  {{ parent_phase_id }}      Parent phase ID
  {{ full_path }}            Complete path from festival root

EXAMPLES:
  # Create single task in current sequence
  fest create task --name "design endpoints" --json

  # Create multiple tasks at once (sequential numbering)
  fest create task --name "requirements" --name "design" --name "implement"
  # Creates: 01_requirements.md, 02_design.md, 03_implement.md

  # Create tasks starting after position 2
  fest create task --after 2 --name "new step" --name "another step"
  # Creates: 03_new_step.md, 04_another_step.md

  # Create task in specific sequence
  fest create task --name "setup" --path ./002_IMPLEMENT/01_api --json

MARKER FILLING (for AI agents):
  # Fill all REPLACE markers in one command
  fest create task --name "setup" --markers '{"Brief description": "Auth middleware", "Yes/No": "Yes"}'

  # Preview template markers first (dry-run)
  fest create task --name "setup" --dry-run --json

  # Skip marker filling (leave REPLACE tags)
  fest create task --name "setup" --skip-markers

Run 'fest understand tasks' for detailed guidance on task file creation.
Run 'fest validate tasks' to verify task files exist in implementation sequences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreateTaskTUI(cmd.Context())
			}
			if len(opts.Names) == 0 {
				return errors.Validation("--name is required (or run without flags to open TUI)")
			}
			// Validate all names are non-empty
			for _, name := range opts.Names {
				if strings.TrimSpace(name) == "" {
					return errors.Validation("task names cannot be empty")
				}
			}
			return RunCreateTask(cmd.Context(), opts)
		},
	}
	cmd.Flags().IntVar(&opts.After, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringSliceVar(&opts.Names, "name", nil, "Task name(s) - can be specified multiple times for batch creation")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Path to sequence directory (directory containing numbered task files)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().StringVar(&opts.Markers, "markers", "", "JSON string with REPLACE marker hint→value mappings")
	cmd.Flags().StringVar(&opts.MarkersFile, "markers-file", "", "JSON file with REPLACE marker hint→value mappings")
	cmd.Flags().BoolVar(&opts.SkipMarkers, "skip-markers", false, "Skip REPLACE marker processing")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show template markers without creating file")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().BoolVar(&opts.AgentMode, "agent", false, "Strict mode: require markers, auto-validate, block on errors, JSON output")
	return cmd
}

// RunCreateTask executes the create task command logic.
func RunCreateTask(ctx context.Context, opts *CreateTaskOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreateTask")
	}

	// Agent mode implies JSON output
	if opts.AgentMode {
		opts.JSONOutput = true
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve paths for config loading
	festivalsRoot := ResolveFestivalsRoot(cwd)
	festivalPath := ResolveFestivalPath(cwd)

	// Load effective agent config (workspace + festival merged)
	agentCfg := LoadEffectiveAgentConfig(festivalsRoot, festivalPath)

	// Determine effective skip-markers behavior
	effectiveSkipMarkers := config.EffectiveSkipMarkers(agentCfg, opts.AgentMode, opts.SkipMarkers)

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateTaskError(opts, err)
	}

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitCreateTaskError(opts, errors.Wrap(err, "resolving path").WithField("path", opts.Path))
	}

	// Load vars once for all tasks
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreateTaskError(opts, errors.Wrap(err, "reading vars-file").WithField("path", opts.VarsFile))
		}
		vars = v
	}

	// Load template catalog once
	catalog, _ := tpl.LoadCatalog(ctx, tmplRoot)
	mgr := tpl.NewManager()
	loader := tpl.NewLoader()

	// Track all created tasks for output
	var createdTasks []map[string]interface{}
	var createdPaths []string
	var totalMarkersFilled, totalMarkersCount int
	currentAfter := opts.After

	// Create each task sequentially
	for _, name := range opts.Names {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return errors.Wrap(ctxErr, "context cancelled").WithOp("RunCreateTask")
		}

		// Insert task at current position
		ren := festival.NewRenumberer(festival.RenumberOptions{AutoApprove: true, Quiet: true})
		if err := ren.InsertTask(ctx, absPath, currentAfter, name); err != nil {
			return emitCreateTaskError(opts, errors.Wrap(err, "inserting task").WithField("name", name))
		}

		// Compute new task id
		newNumber := currentAfter + 1
		taskID := tpl.FormatTaskID(newNumber, name)
		taskPath := filepath.Join(absPath, taskID)

		// Build template context for this task
		tmplCtx := tpl.NewContext()
		tmplCtx.SetTask(newNumber, name)
		tmplCtx.ComputeStructureVariables()
		for k, v := range vars {
			tmplCtx.SetCustom(k, v)
		}

		// Render or copy TASK template
		var content string
		var renderErr error
		if catalog != nil {
			content, renderErr = mgr.RenderByID(ctx, catalog, "TASK", tmplCtx)
		}
		if renderErr != nil || content == "" {
			// Fall back to default filename
			tpath := filepath.Join(tmplRoot, "tasks", "TASK.md")
			if _, err := os.Stat(tpath); err == nil {
				t, err := loader.Load(ctx, tpath)
				if err != nil {
					return emitCreateTaskError(opts, errors.Wrap(err, "loading task template"))
				}
				// Render if it appears templated; else copy
				if strings.Contains(t.Content, "{{") {
					out, err := mgr.Render(t, tmplCtx)
					if err != nil {
						return emitCreateTaskError(opts, errors.Wrap(err, "rendering task"))
					}
					content = out
				} else {
					content = t.Content
				}
			}
		}

		// If no template content was found, create a minimal placeholder
		if content == "" {
			content = fmt.Sprintf("# Task: %s\n\n> **Task Number**: %02d | **Status:** Pending\n\n## Objective\n\n[REPLACE: Describe the task objective]\n\n## Steps\n\n1. [REPLACE: Step 1]\n2. [REPLACE: Step 2]\n\n## Definition of Done\n\n- [ ] [REPLACE: Completion criterion]\n", name, newNumber)
		}

		// Write task file (the file was created by InsertTask, but we need to write content)
		if content != "" {
			// Inject frontmatter if content doesn't already have it
			if !strings.HasPrefix(strings.TrimSpace(content), "---") {
				parentSequenceID := filepath.Base(absPath)
				fm := frontmatter.NewTaskFrontmatter(taskID, name, parentSequenceID, newNumber, frontmatter.AutonomyMedium)
				contentWithFM, fmErr := frontmatter.InjectString(content, fm)
				if fmErr != nil {
					return emitCreateTaskError(opts, errors.Wrap(fmErr, "injecting frontmatter"))
				}
				content = contentWithFM
			}

			if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
				return emitCreateTaskError(opts, errors.IO("writing task", err).WithField("path", taskPath))
			}

			// Process REPLACE markers in the created file
			markerResult, err := ProcessMarkers(ctx, MarkerOptions{
				FilePath:    taskPath,
				Markers:     opts.Markers,
				MarkersFile: opts.MarkersFile,
				SkipMarkers: effectiveSkipMarkers,
				DryRun:      opts.DryRun,
				JSONOutput:  opts.JSONOutput,
			})
			if err != nil {
				return emitCreateTaskError(opts, errors.Wrap(err, "processing markers"))
			}

			// For dry-run, output markers and exit
			if opts.DryRun && markerResult != nil {
				if err := PrintDryRunMarkers(markerResult, opts.JSONOutput); err != nil {
					return emitCreateTaskError(opts, err)
				}
				return nil
			}

			// Track marker results for reporting
			if markerResult != nil && markerResult.Total > 0 {
				totalMarkersFilled += markerResult.Filled
				totalMarkersCount += markerResult.Total
			}
		}

		// Track created task
		createdTasks = append(createdTasks, map[string]interface{}{
			"number": newNumber,
			"id":     taskID,
			"name":   name,
		})
		createdPaths = append(createdPaths, taskPath)

		// Increment position for next task
		currentAfter = newNumber
	}

	// Run post-create validation if configured
	var validationResult *ValidationSummary
	shouldValidate := config.ShouldValidate(agentCfg, opts.AgentMode)
	if shouldValidate && festivalPath != "" {
		validationResult, err = RunPostCreateValidation(ctx, festivalPath)
		if err != nil {
			// Don't fail on validation errors, just report
			if !opts.JSONOutput {
				display.Warning("Validation failed: %v", err)
			}
		}

		// Block on errors if configured
		if validationResult != nil && !validationResult.OK {
			if config.ShouldBlockOnErrors(agentCfg, opts.AgentMode) {
				return emitCreateTaskError(opts, errors.Validation("validation errors detected - fix issues before proceeding"))
			}
		}
	}

	// Output results
	remainingMarkers := totalMarkersCount - totalMarkersFilled

	if opts.JSONOutput {
		warnings := []string{}
		if remainingMarkers > 0 {
			warnings = append(warnings,
				fmt.Sprintf("CRITICAL: %d unfilled markers - task cannot be executed until resolved", remainingMarkers),
				"Edit task file directly to replace [REPLACE: ...] markers",
			)
		}
		warnings = append(warnings, "Edit task file to define implementation steps")

		// Add discovery commands for agents
		suggestions := []string{
			"fest status        - View festival progress",
			"fest next          - Find what to work on next",
			"fest show plan     - View the execution plan",
			"fest validate      - Check completion status",
			"fest progress      - See detailed progress",
		}

		result := createTaskResult{
			OK:            true,
			Action:        "create_task",
			Created:       createdPaths,
			Renumber:      []string{},
			MarkersFilled: totalMarkersFilled,
			MarkersTotal:  totalMarkersCount,
			Validation:    validationResult,
			Warnings:      warnings,
			Suggestions:   suggestions,
		}
		// For single task, use Task field for backward compatibility
		if len(createdTasks) == 1 {
			result.Task = createdTasks[0]
		}
		return emitCreateTaskJSON(opts, result)
	}

	// Show marker warning FIRST (before success message) for visibility
	if remainingMarkers > 0 {
		fmt.Println()
		display.Error("🚫 CRITICAL: %d unfilled markers - task cannot be executed until resolved", remainingMarkers)
		display.Info("   Edit task file(s) directly to replace [REPLACE: ...] markers")
		fmt.Println()
	}

	// Human-readable output
	if len(createdTasks) == 1 {
		display.Success("Created task: %s", createdTasks[0]["id"])
		display.Info("  └── %s", createdPaths[0])
	} else {
		display.Success("Created %d tasks:", len(createdTasks))
		for _, task := range createdTasks {
			display.Info("  └── %s", task["id"])
		}
	}

	// Report validation results
	if validationResult != nil {
		if validationResult.OK {
			display.Success("Validation passed (score: %d)", validationResult.Score)
		} else {
			display.Warning("Validation issues found (score: %d, errors: %d, warnings: %d)",
				validationResult.Score, validationResult.Errors, validationResult.Warnings)
			for _, issue := range validationResult.Issues {
				display.Info("  • [%s] %s: %s", issue.Level, issue.Path, issue.Message)
			}
		}
	}

	fmt.Println()
	fmt.Println(ui.H2("Next Steps"))
	if remainingMarkers > 0 {
		fmt.Printf("  %s\n", ui.Info("1. Edit task file to define implementation steps"))
		fmt.Printf("  %s\n", ui.Info("2. fest create task --name \"next_step\" (add more tasks)"))
	} else {
		fmt.Printf("  %s\n", ui.Info("• Add more tasks: fest create task --name \"next_step\""))
	}
	fmt.Printf("  %s\n", ui.Info("• Add quality gates: fest gates apply --approve"))
	fmt.Printf("  %s\n", ui.Info("• Validate progress: fest validate"))
	fmt.Println()
	fmt.Println(ui.H2("Discover More Commands"))
	fmt.Printf("  %s %s\n", ui.Value("fest status"), ui.Dim("View festival progress"))
	fmt.Printf("  %s %s\n", ui.Value("fest next"), ui.Dim("Find what to work on next"))
	fmt.Printf("  %s %s\n", ui.Value("fest show plan"), ui.Dim("View the execution plan"))
	return nil
}

func emitCreateTaskError(opts *CreateTaskOptions, err error) error {
	if opts.JSONOutput {
		_ = emitCreateTaskJSON(opts, createTaskResult{
			OK:     false,
			Action: "create_task",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitCreateTaskJSON(opts *CreateTaskOptions, res createTaskResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
