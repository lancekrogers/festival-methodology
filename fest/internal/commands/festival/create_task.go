package festival

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreateTaskOptions holds options for the create task command.
type CreateTaskOptions struct {
	After      int
	Names      []string
	Path       string
	VarsFile   string
	JSONOutput bool
}

type createTaskResult struct {
	OK       bool                   `json:"ok"`
	Action   string                 `json:"action"`
	Task     map[string]interface{} `json:"task,omitempty"`
	Created  []string               `json:"created,omitempty"`
	Renumber []string               `json:"renumbered,omitempty"`
	Errors   []map[string]any       `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
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

Run 'fest understand tasks' for detailed guidance on task file creation.
Run 'fest validate tasks' to verify task files exist in implementation sequences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreateTaskTUI()
			}
			if len(opts.Names) == 0 {
				return fmt.Errorf("--name is required (or run without flags to open TUI)")
			}
			// Validate all names are non-empty
			for _, name := range opts.Names {
				if strings.TrimSpace(name) == "" {
					return fmt.Errorf("task names cannot be empty")
				}
			}
			return RunCreateTask(opts)
		},
	}
	cmd.Flags().IntVar(&opts.After, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringSliceVar(&opts.Names, "name", nil, "Task name(s) - can be specified multiple times for batch creation")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Path to sequence directory (directory containing numbered task files)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	return cmd
}

// RunCreateTask executes the create task command logic.
func RunCreateTask(opts *CreateTaskOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateTaskError(opts, err)
	}

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitCreateTaskError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Load vars once for all tasks
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreateTaskError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Load template catalog once
	catalog, _ := tpl.LoadCatalog(tmplRoot)
	mgr := tpl.NewManager()
	loader := tpl.NewLoader()

	// Track all created tasks for output
	var createdTasks []map[string]interface{}
	var createdPaths []string
	currentAfter := opts.After

	// Create each task sequentially
	for _, name := range opts.Names {
		// Insert task at current position
		ren := festival.NewRenumberer(festival.RenumberOptions{AutoApprove: true, Quiet: true})
		if err := ren.InsertTask(absPath, currentAfter, name); err != nil {
			return emitCreateTaskError(opts, fmt.Errorf("failed to insert task %q: %w", name, err))
		}

		// Compute new task id
		newNumber := currentAfter + 1
		taskID := tpl.FormatTaskID(newNumber, name)
		taskPath := filepath.Join(absPath, taskID)

		// Build context for this task
		ctx := tpl.NewContext()
		ctx.SetTask(newNumber, name)
		for k, v := range vars {
			ctx.SetCustom(k, v)
		}

		// Render or copy TASK template
		var content string
		var renderErr error
		if catalog != nil {
			content, renderErr = mgr.RenderByID(catalog, "TASK", ctx)
		}
		if renderErr != nil || content == "" {
			// Fall back to default filename
			tpath := filepath.Join(tmplRoot, "TASK_TEMPLATE.md")
			if _, err := os.Stat(tpath); err == nil {
				t, err := loader.Load(tpath)
				if err != nil {
					return emitCreateTaskError(opts, fmt.Errorf("failed to load task template: %w", err))
				}
				// Render if it appears templated; else copy
				if strings.Contains(t.Content, "{{") {
					out, err := mgr.Render(t, ctx)
					if err != nil {
						return emitCreateTaskError(opts, fmt.Errorf("failed to render task: %w", err))
					}
					content = out
				} else {
					content = t.Content
				}
			}
		}

		// Write task file (the file was created by InsertTask, but we need to write content)
		if content != "" {
			if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
				return emitCreateTaskError(opts, fmt.Errorf("failed to write task: %w", err))
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

	// Output results
	if opts.JSONOutput {
		result := createTaskResult{
			OK:       true,
			Action:   "create_task",
			Created:  createdPaths,
			Renumber: []string{},
		}
		// For single task, use Task field for backward compatibility
		if len(createdTasks) == 1 {
			result.Task = createdTasks[0]
		}
		return emitCreateTaskJSON(opts, result)
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
	fmt.Println()
	fmt.Println("   Next steps:")
	fmt.Println("   • Add more tasks: fest create task --name \"next_step\"")
	fmt.Println("   • Add quality gates: fest gates apply --approve")
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
