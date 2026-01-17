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
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreatePhaseOptions holds options for the create phase command.
type CreatePhaseOptions struct {
	After       int
	Name        string
	PhaseType   string
	Path        string
	VarsFile    string
	Markers     string // Inline JSON with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing
	DryRun      bool   // Show markers without creating file
	JSONOutput  bool
	AgentMode   bool // Strict mode for AI agents
}

type createPhaseResult struct {
	OK            bool                     `json:"ok"`
	Action        string                   `json:"action"`
	Phase         map[string]interface{}   `json:"phase,omitempty"`
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

// selectPhaseTemplate returns the appropriate template ID and filename for a given phase type.
// Returns (templateID, templateFilename, error) tuple.
// Phase-type templates are stored in phases/{phase_type}/GOAL.md
// Returns error for unknown phase types (no fallback - phase type is required).
func selectPhaseTemplate(phaseType string) (string, string, error) {
	pt := strings.ToLower(phaseType)
	switch pt {
	case "planning", "implementation", "research", "review", "non_coding_action":
		return fmt.Sprintf("phase-goal-%s", pt), filepath.Join("phases", pt, "GOAL.md"), nil
	default:
		return "", "", fmt.Errorf("unknown phase type %q: must be one of planning, implementation, research, review, non_coding_action", phaseType)
	}
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
	cmd.Flags().IntVar(&opts.After, "after", -1, "Insert after this phase number (-1 or omit to append at end)")
	cmd.Flags().StringVar(&opts.Name, "name", "", "Phase name (required)")
	cmd.Flags().StringVar(&opts.PhaseType, "type", "planning", "Phase type (planning|implementation|review|deployment|research)")
	cmd.Flags().StringVar(&opts.Path, "path", ".", "Path to festival root (directory containing numbered phases)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().StringVar(&opts.Markers, "markers", "", "JSON string with REPLACE marker hint→value mappings")
	cmd.Flags().StringVar(&opts.MarkersFile, "markers-file", "", "JSON file with REPLACE marker hint→value mappings")
	cmd.Flags().BoolVar(&opts.SkipMarkers, "skip-markers", false, "Skip REPLACE marker processing")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show template markers without creating file")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().BoolVar(&opts.AgentMode, "agent", false, "Strict mode: require markers, auto-validate, block on errors, JSON output")
	return cmd
}

// RunCreatePhase executes the create phase command logic.
func RunCreatePhase(ctx context.Context, opts *CreatePhaseOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreatePhase")
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
		return emitCreatePhaseError(opts, err)
	}

	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitCreatePhaseError(opts, errors.Wrap(err, "resolving path").WithField("path", opts.Path))
	}

	// Auto-detect last phase number when --after is not specified (default -1)
	if opts.After == -1 {
		parser := festival.NewParser()
		phases, parseErr := parser.ParsePhases(ctx, absPath)
		if parseErr == nil && len(phases) > 0 {
			maxNum := 0
			for _, p := range phases {
				if p.Number > maxNum {
					maxNum = p.Number
				}
			}
			opts.After = maxNum
		} else {
			// No existing phases or parse error - insert at beginning
			opts.After = 0
		}
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
	templateID, templateFilename, phaseTypeErr := selectPhaseTemplate(opts.PhaseType)
	if phaseTypeErr != nil {
		return emitCreatePhaseError(opts, errors.Validation(phaseTypeErr.Error()).WithField("phase_type", opts.PhaseType))
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

	// If no template content was found, create a minimal placeholder
	if content == "" {
		content = fmt.Sprintf("# Phase Goal: %s\n\n**Phase:** %03d | **Type:** %s | **Status:** Planning\n\n## Objective\n\n[REPLACE: Describe the phase objective]\n\n## Success Criteria\n\n- [ ] [REPLACE: Criterion 1]\n- [ ] [REPLACE: Criterion 2]\n", opts.Name, newNumber, opts.PhaseType)
	}

	// Ensure content has proper phase type frontmatter
	// Strip any template metadata frontmatter and add phase frontmatter with type
	content = ensurePhaseTypeFrontmatter(content, opts.PhaseType)

	var markersFilled, markersTotal int
	if content != "" {
		if err := os.WriteFile(goalPath, []byte(content), 0644); err != nil {
			return emitCreatePhaseError(opts, errors.IO("writing phase goal", err).WithField("path", goalPath))
		}

		// Process REPLACE markers in the created file
		markerResult, err := ProcessMarkers(ctx, MarkerOptions{
			FilePath:    goalPath,
			Markers:     opts.Markers,
			MarkersFile: opts.MarkersFile,
			SkipMarkers: effectiveSkipMarkers,
			DryRun:      opts.DryRun,
			JSONOutput:  opts.JSONOutput,
		})
		if err != nil {
			return emitCreatePhaseError(opts, errors.Wrap(err, "processing markers"))
		}

		// For dry-run, output markers and exit
		if opts.DryRun && markerResult != nil {
			if err := PrintDryRunMarkers(markerResult, opts.JSONOutput); err != nil {
				return emitCreatePhaseError(opts, err)
			}
			return nil
		}

		if markerResult != nil {
			markersFilled = markerResult.Filled
			markersTotal = markerResult.Total
		}
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
				return emitCreatePhaseError(opts, errors.Validation("validation errors detected - fix issues before proceeding"))
			}
		}
	}

	if opts.JSONOutput {
		remainingMarkers := markersTotal - markersFilled
		warnings := []string{}
		if remainingMarkers > 0 {
			warnings = append(warnings,
				fmt.Sprintf("CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers),
				"Run 'fest wizard fill PHASE_GOAL.md' to fill markers interactively",
			)
		}
		warnings = append(warnings, "Next: Create sequences with 'fest create sequence --name SEQUENCE_NAME'")

		// Add discovery commands for agents
		suggestions := []string{
			"fest status        - View festival progress",
			"fest next          - Find what to work on next",
			"fest show plan     - View the execution plan",
			"fest validate      - Check completion status",
		}

		return emitCreatePhaseJSON(opts, createPhaseResult{
			OK:     true,
			Action: "create_phase",
			Phase: map[string]interface{}{
				"number": newNumber,
				"id":     phaseID,
				"name":   opts.Name,
				"type":   opts.PhaseType,
			},
			Created:       []string{goalPath},
			Renumber:      []string{},
			MarkersFilled: markersFilled,
			MarkersTotal:  markersTotal,
			Validation:    validationResult,
			Warnings:      warnings,
			Suggestions:   suggestions,
		})
	}

	// Show marker warning FIRST (before success message) for visibility
	remainingMarkers := markersTotal - markersFilled
	if remainingMarkers > 0 {
		fmt.Println()
		display.Error("🚫 CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers)
		display.Info("   Run 'fest wizard fill PHASE_GOAL.md' to fill markers interactively")
		display.Info("   Or edit PHASE_GOAL.md directly to replace [REPLACE: ...] markers")
		fmt.Println()
	}

	display.Success("Created phase: %s", phaseID)
	display.Info("  └── %s", "PHASE_GOAL.md")

	fmt.Println()
	fmt.Println(ui.H2("Next Steps"))
	fmt.Printf("  %s\n", ui.Info(fmt.Sprintf("1. cd %s", phaseDir)))
	if remainingMarkers > 0 {
		fmt.Printf("  %s\n", ui.Info("2. Edit PHASE_GOAL.md to define phase objectives"))
	}
	if opts.PhaseType == "research" {
		fmt.Printf("  %s\n", ui.Info("3. Create subdirectories for research topics"))
		fmt.Printf("  %s\n", ui.Info("4. fest research create --type investigation --title \"topic\""))
	} else {
		fmt.Printf("  %s\n", ui.Info("3. fest create sequence --name \"requirements\""))
		fmt.Printf("  %s\n", ui.Info("4. fest validate (check completion status)"))
	}
	fmt.Println()
	fmt.Println(ui.H2("Discover More Commands"))
	fmt.Printf("  %s %s\n", ui.Value("fest status"), ui.Dim("View festival progress"))
	fmt.Printf("  %s %s\n", ui.Value("fest next"), ui.Dim("Find what to work on next"))
	fmt.Printf("  %s %s\n", ui.Value("fest show plan"), ui.Dim("View the execution plan"))
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

// ensurePhaseTypeFrontmatter ensures the content has proper YAML frontmatter with the phase type.
// If content already has frontmatter (template metadata), it strips it and adds phase frontmatter.
// If content has no frontmatter, it prepends phase frontmatter.
func ensurePhaseTypeFrontmatter(content, phaseType string) string {
	// Phase frontmatter using the fest_ prefix convention for proper parsing
	phaseFrontmatter := fmt.Sprintf("---\nfest_phase_type: %s\n---\n\n", phaseType)

	// Check if content starts with frontmatter
	if strings.HasPrefix(content, "---") {
		// Find the closing --- to strip template metadata frontmatter
		rest := content[3:] // skip opening ---
		endIdx := strings.Index(rest, "---")
		if endIdx != -1 {
			// Skip past the closing --- and any following newlines
			afterFrontmatter := rest[endIdx+3:]
			afterFrontmatter = strings.TrimLeft(afterFrontmatter, "\n\r")
			return phaseFrontmatter + afterFrontmatter
		}
	}

	// No existing frontmatter, just prepend
	return phaseFrontmatter + content
}
