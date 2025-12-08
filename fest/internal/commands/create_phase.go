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

type createPhaseOptions struct {
	after      int
	name       string
	phaseType  string
	path       string
	varsFile   string
	jsonOutput bool
}

type createPhaseResult struct {
	OK       bool                   `json:"ok"`
	Action   string                 `json:"action"`
	Phase    map[string]interface{} `json:"phase,omitempty"`
	Created  []string               `json:"created,omitempty"`
	Renumber []string               `json:"renumbered,omitempty"`
	Errors   []map[string]any       `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
}

// NewCreatePhaseCommand adds 'create phase'
func NewCreatePhaseCommand() *cobra.Command {
    opts := &createPhaseOptions{}
    cmd := &cobra.Command{
        Use:   "phase",
        Short: "Insert a new phase and render its goal file",
        RunE: func(cmd *cobra.Command, args []string) error {
            if cmd.Flags().NFlag() == 0 {
                return StartCreatePhaseTUI()
            }
            if strings.TrimSpace(opts.name) == "" {
                return fmt.Errorf("--name is required (or run without flags to open TUI)")
            }
            return runCreatePhase(opts)
        },
    }
    cmd.Flags().IntVar(&opts.after, "after", 0, "Insert after this number (0 inserts at beginning)")
    cmd.Flags().StringVar(&opts.name, "name", "", "Phase name (required)")
    cmd.Flags().StringVar(&opts.phaseType, "type", "planning", "Phase type (planning|implementation|review|deployment)")
    cmd.Flags().StringVar(&opts.path, "path", ".", "Path to festival root (directory containing numbered phases)")
    cmd.Flags().StringVar(&opts.varsFile, "vars-file", "", "JSON vars for rendering")
    cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit JSON output")
    return cmd
}

func runCreatePhase(opts *createPhaseOptions) error {
	display := ui.New(noColor, verbose)
	cwd, _ := os.Getwd()
	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreatePhaseError(opts, err)
	}

	absPath, err := filepath.Abs(opts.path)
	if err != nil {
		return emitCreatePhaseError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Insert phase
	ren := festival.NewRenumberer(festival.RenumberOptions{})
	if err := ren.InsertPhase(absPath, opts.after, opts.name); err != nil {
		return emitCreatePhaseError(opts, fmt.Errorf("failed to insert phase: %w", err))
	}

	// Compute new phase id
	newNumber := opts.after + 1
	phaseID := tpl.FormatPhaseID(newNumber, opts.name)
	phaseDir := filepath.Join(absPath, phaseID)

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.varsFile) != "" {
		v, err := loadVarsFile(opts.varsFile)
		if err != nil {
			return emitCreatePhaseError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Build context for phase
	ctx := tpl.NewContext()
	ctx.SetPhase(newNumber, opts.name, opts.phaseType)
	for k, v := range vars {
		ctx.SetCustom(k, v)
	}

	// Render or copy PHASE_GOAL template
	// Try IDs first via catalog
	catalog, _ := tpl.LoadCatalog(tmplRoot)
	mgr := tpl.NewManager()
	var content string
	var renderErr error
	if catalog != nil {
		content, renderErr = mgr.RenderByID(catalog, "PHASE_GOAL", ctx)
	}
	if renderErr != nil || content == "" {
		// Fall back to default filename
		tpath := filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(tpath)
			if err != nil {
				return emitCreatePhaseError(opts, fmt.Errorf("failed to load phase goal template: %w", err))
			}
			// Render if it appears templated; else copy
			if strings.Contains(t.Content, "{{") {
				out, err := mgr.Render(t, ctx)
				if err != nil {
					return emitCreatePhaseError(opts, fmt.Errorf("failed to render phase goal: %w", err))
				}
				content = out
			} else {
				content = t.Content
			}
		}
	}

	// Ensure dir and write file
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		return emitCreatePhaseError(opts, fmt.Errorf("failed to create phase dir: %w", err))
	}
	goalPath := filepath.Join(phaseDir, "PHASE_GOAL.md")
	if content != "" {
		if err := os.WriteFile(goalPath, []byte(content), 0644); err != nil {
			return emitCreatePhaseError(opts, fmt.Errorf("failed to write phase goal: %w", err))
		}
	}

	if opts.jsonOutput {
		return emitCreatePhaseJSON(opts, createPhaseResult{
			OK:     true,
			Action: "create_phase",
			Phase: map[string]interface{}{
				"number": newNumber,
				"id":     phaseID,
				"name":   opts.name,
				"type":   opts.phaseType,
			},
			Created:  []string{goalPath},
			Renumber: []string{},
		})
	}

	display.Success("Created phase: %s", phaseID)
	display.Info("  • %s", goalPath)
	return nil
}

func emitCreatePhaseError(opts *createPhaseOptions, err error) error {
	if opts.jsonOutput {
		_ = emitCreatePhaseJSON(opts, createPhaseResult{
			OK:     false,
			Action: "create_phase",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitCreatePhaseJSON(opts *createPhaseOptions, res createPhaseResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
