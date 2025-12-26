package festival

import (
	"context"
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

// CreateSequenceOptions holds options for the create sequence command.
type CreateSequenceOptions struct {
	After      int
	Name       string
	Path       string
	VarsFile   string
	JSONOutput bool
	NoPrompt   bool
}

type createSequenceResult struct {
	OK       bool                   `json:"ok"`
	Action   string                 `json:"action"`
	Sequence map[string]interface{} `json:"sequence,omitempty"`
	Created  []string               `json:"created,omitempty"`
	Renumber []string               `json:"renumbered,omitempty"`
	Errors   []map[string]any       `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
}

// NewCreateSequenceCommand adds 'create sequence'
func NewCreateSequenceCommand() *cobra.Command {
	opts := &CreateSequenceOptions{}
	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Insert a new sequence and render its goal file",
		Long: `Create a new sequence directory with SEQUENCE_GOAL.md.

IMPORTANT: After creating a sequence, you must also create TASK FILES.
The SEQUENCE_GOAL.md defines WHAT to achieve, but AI agents need task
files to know HOW to execute. See 'fest understand tasks'.

TEMPLATE VARIABLES (automatically set):
  {{ sequence_name }}        Name of the sequence
  {{ sequence_number }}      Sequential number (01, 02, ...)
  {{ sequence_id }}          Full ID (e.g., "01_api_endpoints")
  {{ parent_phase_id }}      Parent phase ID

EXAMPLES:
  # Create sequence in current phase
  fest create sequence --name "api endpoints" --json

  # Create sequence at specific position
  fest create sequence --name "frontend" --after 2 --json

NEXT STEPS after creating a sequence:
  # Add task files (required for implementation sequences)
  fest create task --name "design" --json
  fest create task --name "implement" --json

  # Add quality gates
  fest gates apply --approve

Run 'fest validate tasks' to verify task files exist.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreateSequenceTUI()
			}
			if strings.TrimSpace(opts.Name) == "" {
				return fmt.Errorf("--name is required (or run without flags to open TUI)")
			}
			return RunCreateSequence(cmd.Context(), opts)
		},
	}
	cmd.Flags().IntVar(&opts.After, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Sequence name (required)")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Path to phase directory (directory containing numbered sequences)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().BoolVar(&opts.NoPrompt, "no-prompt", false, "Skip interactive prompts")
	return cmd
}

// RunCreateSequence executes the create sequence command logic.
func RunCreateSequence(ctx context.Context, opts *CreateSequenceOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateSequenceError(opts, err)
	}

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitCreateSequenceError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Insert sequence
	ren := festival.NewRenumberer(festival.RenumberOptions{AutoApprove: true, Quiet: true})
	if err := ren.InsertSequence(ctx, absPath, opts.After, opts.Name); err != nil {
		return emitCreateSequenceError(opts, fmt.Errorf("failed to insert sequence: %w", err))
	}

	// Compute new sequence id
	newNumber := opts.After + 1
	seqID := tpl.FormatSequenceID(newNumber, opts.Name)
	seqDir := filepath.Join(absPath, seqID)

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreateSequenceError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Build template context for sequence
	tmplCtx := tpl.NewContext()
	tmplCtx.SetSequence(newNumber, opts.Name)
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// Render or copy SEQUENCE_GOAL template
	catalog, _ := tpl.LoadCatalog(ctx, tmplRoot)
	mgr := tpl.NewManager()
	var content string
	var renderErr error
	if catalog != nil {
		content, renderErr = mgr.RenderByID(ctx, catalog, "SEQUENCE_GOAL", tmplCtx)
	}
	if renderErr != nil || content == "" {
		// Fall back to default filename
		tpath := filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(ctx, tpath)
			if err != nil {
				return emitCreateSequenceError(opts, fmt.Errorf("failed to load sequence goal template: %w", err))
			}
			// Render if it appears templated; else copy
			if strings.Contains(t.Content, "{{") {
				out, err := mgr.Render(t, tmplCtx)
				if err != nil {
					return emitCreateSequenceError(opts, fmt.Errorf("failed to render sequence goal: %w", err))
				}
				content = out
			} else {
				content = t.Content
			}
		}
	}

	// Ensure dir and write file
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		return emitCreateSequenceError(opts, fmt.Errorf("failed to create sequence dir: %w", err))
	}
	goalPath := filepath.Join(seqDir, "SEQUENCE_GOAL.md")
	if content != "" {
		if err := os.WriteFile(goalPath, []byte(content), 0644); err != nil {
			return emitCreateSequenceError(opts, fmt.Errorf("failed to write sequence goal: %w", err))
		}
	}

	if opts.JSONOutput {
		result := createSequenceResult{
			OK:     true,
			Action: "create_sequence",
			Sequence: map[string]interface{}{
				"number": newNumber,
				"id":     seqID,
				"name":   opts.Name,
			},
			Created:  []string{goalPath},
			Renumber: []string{},
			Warnings: []string{
				"Sequences need task files for AI execution. Goals define WHAT, tasks define HOW.",
				"Create tasks with: fest create task --name \"...\"",
				"Learn more: fest understand tasks",
			},
		}
		return emitCreateSequenceJSON(opts, result)
	}

	display.Success("Created sequence: %s", seqID)
	display.Info("  └── %s", "SEQUENCE_GOAL.md")
	fmt.Println()

	// Show education message
	display.Warning("Sequences need task files to be executable.")
	fmt.Println("   SEQUENCE_GOAL.md defines WHAT to accomplish.")
	fmt.Println("   Task files (01_*.md, 02_*.md) define HOW to do it.")
	fmt.Println()
	fmt.Println("   💡 Run 'fest understand tasks' to learn more about task structure.")
	fmt.Println()

	// Blocking prompt (skip if --no-prompt or --json)
	if !opts.NoPrompt && !opts.JSONOutput {
		if display.Confirm("Create task files now?") {
			fmt.Println()
			fmt.Println("   To create tasks, run:")
			fmt.Println("     fest create task --name \"your_task_name\"")
			fmt.Println()
			fmt.Println("   Or start the interactive TUI:")
			fmt.Println("     fest create task")
		}
	}

	return nil
}

func emitCreateSequenceError(opts *CreateSequenceOptions, err error) error {
	if opts.JSONOutput {
		_ = emitCreateSequenceJSON(opts, createSequenceResult{
			OK:     false,
			Action: "create_sequence",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitCreateSequenceJSON(opts *CreateSequenceOptions, res createSequenceResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
