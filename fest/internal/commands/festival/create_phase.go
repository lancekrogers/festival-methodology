package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreatePhaseOptions holds options for the create phase command.
type CreatePhaseOptions struct {
	After      int
	Name       string
	PhaseType  string
	Path       string
	VarsFile   string
	JSONOutput bool
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
	opts := &CreatePhaseOptions{}
	cmd := &cobra.Command{
		Use:   "phase",
		Short: "Insert a new phase and render its goal file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreatePhaseTUI(cmd.Context())
			}
			if strings.TrimSpace(opts.Name) == "" {
				return errors.Validation("--name is required (or run without flags to open TUI)")
			}
			return RunCreatePhase(cmd.Context(), opts)
		},
	}
	cmd.Flags().IntVar(&opts.After, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Phase name (required)")
	cmd.Flags().StringVar(&opts.PhaseType, "type", "planning", "Phase type (planning|implementation|review|deployment|research)")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Path to festival root (directory containing numbered phases)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	return cmd
}

// RunCreatePhase executes the create phase command logic.
func RunCreatePhase(ctx context.Context, opts *CreatePhaseOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreatePhase")
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()
	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreatePhaseError(opts, err)
	}

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitCreatePhaseError(opts, errors.Wrap(err, "resolving path").WithField("path", opts.Path))
	}

	// Insert phase
	ren := festival.NewRenumberer(festival.RenumberOptions{AutoApprove: true, Quiet: true})
	if err := ren.InsertPhase(ctx, absPath, opts.After, opts.Name); err != nil {
		return emitCreatePhaseError(opts, errors.Wrap(err, "inserting phase"))
	}

	// Compute new phase id
	newNumber := opts.After + 1
	phaseID := tpl.FormatPhaseID(newNumber, opts.Name)
	phaseDir := filepath.Join(absPath, phaseID)

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreatePhaseError(opts, errors.Wrap(err, "reading vars-file").WithField("path", opts.VarsFile))
		}
		vars = v
	}

	// Build template context for phase
	tmplCtx := tpl.NewContext()
	tmplCtx.SetPhase(newNumber, opts.Name, opts.PhaseType)
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// Render or copy PHASE_GOAL template
	// Try IDs first via catalog
	catalog, _ := tpl.LoadCatalog(ctx, tmplRoot)
	mgr := tpl.NewManager()
	var content string
	var renderErr error

	// Select template based on phase type
	templateID := "PHASE_GOAL"
	templateFilename := "PHASE_GOAL_TEMPLATE.md"
	if opts.PhaseType == "research" {
		templateID = "RESEARCH_PHASE_GOAL"
		templateFilename = "RESEARCH_PHASE_GOAL_TEMPLATE.md"
	}

	if catalog != nil {
		content, renderErr = mgr.RenderByID(ctx, catalog, templateID, tmplCtx)
	}
	if renderErr != nil || content == "" {
		// Fall back to default filename
		tpath := filepath.Join(tmplRoot, templateFilename)
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(ctx, tpath)
			if err != nil {
				return emitCreatePhaseError(opts, errors.Wrap(err, "loading phase goal template").WithField("template", templateFilename))
			}
			// Render if it appears templated; else copy
			if strings.Contains(t.Content, "{{") {
				out, err := mgr.Render(t, tmplCtx)
				if err != nil {
					return emitCreatePhaseError(opts, errors.Wrap(err, "rendering phase goal"))
				}
				content = out
			} else {
				content = t.Content
			}
		}
	}

	// Ensure dir and write file
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		return emitCreatePhaseError(opts, errors.IO("creating phase dir", err).WithField("path", phaseDir))
	}
	goalPath := filepath.Join(phaseDir, "PHASE_GOAL.md")
	if content != "" {
		if err := os.WriteFile(goalPath, []byte(content), 0644); err != nil {
			return emitCreatePhaseError(opts, errors.IO("writing phase goal", err).WithField("path", goalPath))
		}
	}

	if opts.JSONOutput {
		return emitCreatePhaseJSON(opts, createPhaseResult{
			OK:     true,
			Action: "create_phase",
			Phase: map[string]interface{}{
				"number": newNumber,
				"id":     phaseID,
				"name":   opts.Name,
				"type":   opts.PhaseType,
			},
			Created:  []string{goalPath},
			Renumber: []string{},
			Warnings: []string{
				"Next: Create sequences with 'fest create sequence --name SEQUENCE_NAME'",
			},
		})
	}

	display.Success("Created phase: %s", phaseID)
	display.Info("  └── %s", "PHASE_GOAL.md")
	fmt.Println()
	fmt.Println("   Next steps:")
	fmt.Println("   1. cd", phaseDir)
	if opts.PhaseType == "research" {
		fmt.Println("   2. Create subdirectories for research topics")
		fmt.Println("   3. fest research create --type investigation --title \"topic\"")
	} else {
		fmt.Println("   2. fest create sequence --name \"requirements\"")
		fmt.Println("   3. fest create sequence --name \"implementation\"")
	}
	return nil
}

func emitCreatePhaseError(opts *CreatePhaseOptions, err error) error {
	if opts.JSONOutput {
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

func emitCreatePhaseJSON(opts *CreatePhaseOptions, res createPhaseResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
