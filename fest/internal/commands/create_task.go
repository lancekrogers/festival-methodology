package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type createTaskOptions struct {
	after      int
	name       string
	path       string
	varsFile   string
	jsonOutput bool
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
	opts := &createTaskOptions{}
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Insert a new task file in a sequence",
		Long: `Create a new task file with automatic numbering and template rendering.

IMPORTANT: AI agents execute TASK FILES, not goals. If your sequences only
have SEQUENCE_GOAL.md without task files, agents won't know HOW to execute.

TEMPLATE VARIABLES (automatically set from --name):
  {{ task_name }}            Name of the task
  {{ task_number }}          Sequential number (01, 02, ...)
  {{ task_id }}              Full filename (e.g., "01_design.md")
  {{ parent_sequence_id }}   Parent sequence ID
  {{ parent_phase_id }}      Parent phase ID
  {{ full_path }}            Complete path from festival root

EXAMPLES:
  # Create task in current sequence
  fest create task --name "design endpoints" --json

  # Create task at specific position
  fest create task --name "validation" --after 2 --json

  # Create task in specific sequence
  fest create task --name "setup" --path ./002_IMPLEMENT/01_api --json

Run 'fest understand tasks' for detailed guidance on task file creation.
Run 'fest validate tasks' to verify task files exist in implementation sequences.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 {
				return StartCreateTaskTUI()
			}
			if strings.TrimSpace(opts.name) == "" {
				return fmt.Errorf("--name is required (or run without flags to open TUI)")
			}
			return runCreateTask(opts)
		},
	}
	cmd.Flags().IntVar(&opts.after, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringVar(&opts.name, "name", "", "Task name (required)")
	cmd.Flags().StringVar(&opts.path, "path", ".", "Path to sequence directory (directory containing numbered task files)")
	cmd.Flags().StringVar(&opts.varsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit JSON output")
	return cmd
}

func runCreateTask(opts *createTaskOptions) error {
	display := ui.New(noColor, verbose)
	cwd, _ := os.Getwd()

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateTaskError(opts, err)
	}

	absPath, err := filepath.Abs(opts.path)
	if err != nil {
		return emitCreateTaskError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Insert task
	ren := festival.NewRenumberer(festival.RenumberOptions{AutoApprove: true, Quiet: true})
	if err := ren.InsertTask(absPath, opts.after, opts.name); err != nil {
		return emitCreateTaskError(opts, fmt.Errorf("failed to insert task: %w", err))
	}

	// Compute new task id
	newNumber := opts.after + 1
	taskID := tpl.FormatTaskID(newNumber, opts.name)
	taskPath := filepath.Join(absPath, taskID)

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.varsFile) != "" {
		v, err := loadVarsFile(opts.varsFile)
		if err != nil {
			return emitCreateTaskError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Build context for task
	ctx := tpl.NewContext()
	ctx.SetTask(newNumber, opts.name)
	for k, v := range vars {
		ctx.SetCustom(k, v)
	}

	// Render or copy TASK template
	catalog, _ := tpl.LoadCatalog(tmplRoot)
	mgr := tpl.NewManager()
	var content string
	var renderErr error
	if catalog != nil {
		content, renderErr = mgr.RenderByID(catalog, "TASK", ctx)
	}
	if renderErr != nil || content == "" {
		// Fall back to default filename
		tpath := filepath.Join(tmplRoot, "TASK_TEMPLATE.md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
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

	if opts.jsonOutput {
		return emitCreateTaskJSON(opts, createTaskResult{
			OK:     true,
			Action: "create_task",
			Task: map[string]interface{}{
				"number": newNumber,
				"id":     taskID,
				"name":   opts.name,
			},
			Created:  []string{taskPath},
			Renumber: []string{},
		})
	}

	display.Success("Created task: %s", taskID)
	display.Info("  • %s", taskPath)
	return nil
}

func emitCreateTaskError(opts *createTaskOptions, err error) error {
	if opts.jsonOutput {
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

func emitCreateTaskJSON(opts *createTaskOptions, res createTaskResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
