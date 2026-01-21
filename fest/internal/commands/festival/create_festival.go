package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/registry"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreateFestivalOptions holds options for the create festival command.
type CreateFestivalOptions struct {
	Name        string
	Goal        string
	Tags        string
	Project     string // Project directory path
	VarsFile    string
	Markers     string // Inline JSON with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing
	DryRun      bool   // Show markers without creating file
	JSONOutput  bool
	Dest        string // "active" or "planned"
	AgentMode   bool   // Strict mode for AI agents
}

type createFestivalResult struct {
	OK             bool                     `json:"ok"`
	Action         string                   `json:"action"`
	Festival       map[string]string        `json:"festival,omitempty"`
	Created        []string                 `json:"created,omitempty"`
	GatesDirectory string                   `json:"gates_directory,omitempty"`
	FestYAML       string                   `json:"fest_yaml,omitempty"`
	GateTemplates  []string                 `json:"gate_templates,omitempty"`
	ProjectPath    string                   `json:"project_path,omitempty"`
	ProjectLinked  bool                     `json:"project_linked,omitempty"`
	Markers        []map[string]interface{} `json:"markers,omitempty"`
	MarkersFilled  int                      `json:"markers_filled,omitempty"`
	MarkersTotal   int                      `json:"markers_total,omitempty"`
	Validation     *ValidationSummary       `json:"validation,omitempty"`
	Errors         []map[string]any         `json:"errors,omitempty"`
	Warnings       []string                 `json:"warnings,omitempty"`
	Extra          map[string]interface{}   `json:"extra,omitempty"`
}

// NewCreateFestivalCommand adds 'create festival'
func NewCreateFestivalCommand() *cobra.Command {
	opts := &CreateFestivalOptions{}
	cmd := &cobra.Command{
		Use:   "festival",
		Short: "Create a new festival scaffold under festivals/(active|planned)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags were provided, open TUI for this flow
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreateFestivalTUI(cmd.Context())
			}
			// Otherwise, require name and proceed
			if strings.TrimSpace(opts.Name) == "" {
				return errors.Validation("--name is required (or run without flags to open TUI)")
			}
			return RunCreateFestival(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Festival name (required)")
	cmd.Flags().StringVar(&opts.Goal, "goal", "", "Festival goal")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVarP(&opts.Project, "project", "p", "", "Project directory path (auto-links to festival)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON file with variables")
	cmd.Flags().StringVar(&opts.Markers, "markers", "", "JSON string with REPLACE marker hint→value mappings")
	cmd.Flags().StringVar(&opts.MarkersFile, "markers-file", "", "JSON file with REPLACE marker hint→value mappings")
	cmd.Flags().BoolVar(&opts.SkipMarkers, "skip-markers", false, "Skip REPLACE marker processing")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show template markers without creating file")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&opts.Dest, "dest", "active", "Destination under festivals/: active or planned")
	cmd.Flags().BoolVar(&opts.AgentMode, "agent", false, "Strict mode: require markers, auto-validate, block on errors, JSON output")
	return cmd
}

// RunCreateFestival executes the create festival command logic.
func RunCreateFestival(ctx context.Context, opts *CreateFestivalOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreateFestival")
	}

	// Agent mode implies JSON output
	if opts.AgentMode {
		opts.JSONOutput = true
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve festivals root and template root
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}

	// Load effective agent config (workspace config only for new festival)
	agentCfg := LoadEffectiveAgentConfig(festivalsRoot, "")

	// Determine effective skip-markers behavior
	effectiveSkipMarkers := config.EffectiveSkipMarkers(agentCfg, opts.AgentMode, opts.SkipMarkers)

	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "reading vars-file").WithField("path", opts.VarsFile))
		}
		vars = v
	}

	// Build template context
	tmplCtx := tpl.NewContext()
	tmplCtx.SetFestival(opts.Name, opts.Goal, parseTags(opts.Tags))
	tmplCtx.ComputeStructureVariables()
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// Destination
	slug := Slugify(opts.Name)
	destCategory := strings.ToLower(strings.TrimSpace(opts.Dest))
	if destCategory != "planned" && destCategory != "active" {
		destCategory = "active"
	}

	// Generate unique festival ID
	festivalID, err := id.GenerateID(opts.Name, festivalsRoot)
	if err != nil {
		return emitCreateFestivalError(opts, errors.Wrap(err, "generating festival ID").WithField("name", opts.Name))
	}

	// Create directory with ID suffix: {slug}-{ID}
	dirName := fmt.Sprintf("%s-%s", slug, festivalID)
	destDir := filepath.Join(festivalsRoot, destCategory, dirName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return emitCreateFestivalError(opts, errors.IO("creating festival directory", err).WithField("path", destDir))
	}

	// Render/copy core files
	mgr := tpl.NewManager()
	created := []string{}

	core := []struct{ Template, Out string }{
		{"festival/OVERVIEW.md", "FESTIVAL_OVERVIEW.md"},
		{"festival/GOAL.md", "FESTIVAL_GOAL.md"},
		{"festival/RULES.md", "FESTIVAL_RULES.md"},
		{"festival/TODO.md", "TODO.md"},
	}

	for _, c := range core {
		tpath := filepath.Join(tmplRoot, c.Template)
		if _, err := os.Stat(tpath); err != nil {
			// Skip missing template silently; report warning via non-JSON path
			continue
		}
		// Load and decide copy vs render
		loader := tpl.NewLoader()
		t, err := loader.Load(ctx, tpath)
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "loading template").WithField("template", c.Template))
		}
		outPath := filepath.Join(destDir, c.Out)
		// If template appears to require variables, render; else copy
		requires := t.Metadata != nil && len(t.Metadata.RequiredVariables) > 0
		var content string
		if requires || strings.Contains(t.Content, "{{") {
			out, err := mgr.Render(t, tmplCtx)
			if err != nil {
				return emitCreateFestivalError(opts, errors.Wrap(err, "rendering template").WithField("template", c.Template))
			}
			content = out
		} else {
			content = t.Content
		}

		// Inject festival frontmatter for FESTIVAL_GOAL.md if not already present
		if c.Out == "FESTIVAL_GOAL.md" && !strings.HasPrefix(strings.TrimSpace(content), "---") {
			fm := frontmatter.NewFrontmatter(frontmatter.TypeFestival, festivalID, opts.Name)
			fm.Status = frontmatter.StatusPlanned
			contentWithFM, fmErr := frontmatter.InjectString(content, fm)
			if fmErr != nil {
				return emitCreateFestivalError(opts, errors.Wrap(fmErr, "injecting frontmatter"))
			}
			content = contentWithFM
		}

		if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
			return emitCreateFestivalError(opts, errors.IO("writing file", err).WithField("path", outPath))
		}
		created = append(created, outPath)
	}

	// Create gates/ directory at festival root with templates organized by phase type
	// Structure: gates/{phase_type}/ contains gate templates for that phase type
	gatesDir := filepath.Join(destDir, "gates")
	srcPhasesDir := filepath.Join(tmplRoot, "phases")

	// Phase types that should have gate subdirectories
	phaseTypes := []string{"planning", "implementation", "research", "review", "non_coding_action"}

	copiedGates := []string{}
	for _, phaseType := range phaseTypes {
		srcGatesDir := filepath.Join(srcPhasesDir, phaseType, "gates")

		// Check if source gates directory exists for this phase type
		if _, err := os.Stat(srcGatesDir); os.IsNotExist(err) {
			continue
		}

		// Create destination gates/{phase_type}/ directory
		destGatesDir := filepath.Join(gatesDir, phaseType)
		if err := os.MkdirAll(destGatesDir, 0755); err != nil {
			return emitCreateFestivalError(opts, errors.IO("creating gates directory", err).WithField("path", destGatesDir))
		}

		// Copy all .md files from source gates to destination
		gateEntries, err := os.ReadDir(srcGatesDir)
		if err != nil {
			return emitCreateFestivalError(opts, errors.IO("reading gates directory", err).WithField("path", srcGatesDir))
		}

		for _, gateEntry := range gateEntries {
			if gateEntry.IsDir() || !strings.HasSuffix(gateEntry.Name(), ".md") {
				continue
			}

			srcPath := filepath.Join(srcGatesDir, gateEntry.Name())
			destPath := filepath.Join(destGatesDir, gateEntry.Name())

			content, err := os.ReadFile(srcPath)
			if err != nil {
				return emitCreateFestivalError(opts, errors.IO("reading gate template", err).WithField("path", srcPath))
			}

			if err := os.WriteFile(destPath, content, 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing gate template", err).WithField("path", destPath))
			}

			copiedGates = append(copiedGates, destPath)
			created = append(created, destPath)
		}
	}

	// Generate fest.yaml with default gates configuration and metadata
	festConfig := DefaultFestivalGatesConfig()

	// Populate metadata section
	now := time.Now().UTC()
	festConfig.Metadata = config.FestivalMetadata{
		ID:        festivalID,
		UUID:      uuid.New().String(),
		Name:      opts.Name,
		CreatedAt: now,
		StatusHistory: []config.StatusChange{
			{
				Status:    destCategory,
				Timestamp: now,
				Path:      destDir,
				Notes:     "Festival created",
			},
		},
	}

	// Handle project path if specified
	var resolvedProjectPath string
	var projectLinked bool
	if opts.Project != "" {
		workspaceRoot := filepath.Dir(festivalsRoot)
		resolved, err := ResolveProjectPath(opts.Project, workspaceRoot)
		if err != nil {
			if !opts.JSONOutput {
				display.Warning("Could not resolve project path: %v", err)
			}
		} else {
			resolvedProjectPath = resolved
			festConfig.ProjectPath = resolved

			// Validate path exists (warning only, don't fail)
			if validateErr := ValidateProjectPath(resolved); validateErr != nil {
				if !opts.JSONOutput {
					display.Warning("Project path doesn't exist yet: %s", resolved)
					display.Info("Link will be created when path exists")
				}
			} else {
				// Auto-link if path exists
				nav, navErr := navigation.LoadNavigation()
				if navErr == nil {
					nav.SetLinkWithPath(dirName, resolved, destDir)
					if saveErr := nav.Save(); saveErr == nil {
						projectLinked = true
						if !opts.JSONOutput {
							display.Success("Linked to project: %s", resolved)
						}
					}
				}
			}
		}
	}

	festConfigPath := filepath.Join(destDir, config.FestivalConfigFileName)
	if err := config.SaveFestivalConfig(destDir, festConfig); err != nil {
		return emitCreateFestivalError(opts, errors.Wrap(err, "writing fest.yaml").WithField("path", festConfigPath))
	}
	created = append(created, festConfigPath)

	// Update ID registry
	regPath := registry.GetRegistryPath(festivalsRoot)
	reg, regErr := registry.Load(ctx, regPath)
	if regErr == nil {
		regEntry := registry.RegistryEntry{
			ID:        festivalID,
			Name:      opts.Name,
			Status:    destCategory,
			Path:      destDir,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if addErr := reg.Add(ctx, regEntry); addErr == nil {
			_ = reg.Save(ctx) // Non-blocking - registry is optional
		}
	}

	// Process REPLACE markers in all created files
	var totalMarkersFilled, totalMarkersCount int
	var allMarkers []map[string]interface{}

	for _, filePath := range created {
		markerResult, err := ProcessMarkers(ctx, MarkerOptions{
			FilePath:    filePath,
			Markers:     opts.Markers,
			MarkersFile: opts.MarkersFile,
			SkipMarkers: effectiveSkipMarkers,
			DryRun:      opts.DryRun,
			JSONOutput:  opts.JSONOutput,
		})
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "processing markers"))
		}

		if markerResult != nil {
			totalMarkersFilled += markerResult.Filled
			totalMarkersCount += markerResult.Total
			allMarkers = append(allMarkers, markerResult.Markers...)
		}
	}

	// For dry-run, output all markers and exit
	if opts.DryRun && totalMarkersCount > 0 {
		result := &MarkerResult{
			Markers: allMarkers,
			Total:   totalMarkersCount,
		}
		if err := PrintDryRunMarkers(result, opts.JSONOutput); err != nil {
			return emitCreateFestivalError(opts, err)
		}
		return nil
	}

	// Run post-create validation if configured
	var validationResult *ValidationSummary
	shouldValidate := config.ShouldValidate(agentCfg, opts.AgentMode)
	if shouldValidate {
		validationResult, err = RunPostCreateValidation(ctx, destDir)
		if err != nil {
			// Don't fail on validation errors, just report
			if !opts.JSONOutput {
				display.Warning("Validation failed: %v", err)
			}
		}

		// Block on errors if configured
		if validationResult != nil && !validationResult.OK {
			if config.ShouldBlockOnErrors(agentCfg, opts.AgentMode) {
				return emitCreateFestivalError(opts, errors.Validation("validation errors detected - fix issues before proceeding"))
			}
		}
	}

	if opts.JSONOutput {
		remainingMarkers := totalMarkersCount - totalMarkersFilled
		warnings := []string{}
		if remainingMarkers > 0 {
			warnings = append(warnings,
				fmt.Sprintf("CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers),
				"Run 'fest validate' to see which files need editing",
				"Run 'fest wizard fill FESTIVAL_GOAL.md' to fill markers interactively",
			)
		}
		warnings = append(warnings, "Next: Create phases with 'fest create phase --name PHASE_NAME'")

		return emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     true,
			Action: "create_festival",
			Festival: map[string]string{
				"name":      opts.Name,
				"slug":      slug,
				"dest":      destCategory,
				"id":        festivalID,
				"directory": dirName,
			},
			Created:        created,
			GatesDirectory: gatesDir,
			FestYAML:       festConfigPath,
			GateTemplates:  copiedGates,
			ProjectPath:    resolvedProjectPath,
			ProjectLinked:  projectLinked,
			MarkersFilled:  totalMarkersFilled,
			MarkersTotal:   totalMarkersCount,
			Validation:     validationResult,
			Warnings:       warnings,
		})
	}

	// Show marker warning FIRST (before success message) for visibility
	remainingMarkers := totalMarkersCount - totalMarkersFilled
	if remainingMarkers > 0 {
		fmt.Println()
		display.Error("🚫 CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers)
		display.Info("   Run 'fest validate' to see which files need editing")
		display.Info("   Run 'fest wizard fill FESTIVAL_GOAL.md' to fill markers interactively")
		fmt.Println()
	}

	display.Success("Created festival: %s (%s)", dirName, destCategory)
	display.Info("  ID: %s", festivalID)
	for _, p := range created {
		display.Info("  • %s", p)
	}

	// Report gates setup
	if len(copiedGates) > 0 {
		display.Success("Created gates/ directory with %d templates organized by phase type", len(copiedGates))
		display.Info("  Quality gates configured in fest.yaml")
	}

	// Report project path if set
	if resolvedProjectPath != "" {
		if projectLinked {
			display.Success("Project path: %s (linked)", resolvedProjectPath)
		} else {
			display.Info("Project path: %s (not linked - path doesn't exist yet)", resolvedProjectPath)
		}
	}

	fmt.Println()
	fmt.Println(ui.H2("Next Steps"))
	fmt.Printf("  %s\n", ui.Info(fmt.Sprintf("1. cd %s", destDir)))
	if remainingMarkers > 0 {
		fmt.Printf("  %s\n", ui.Info("2. Edit FESTIVAL_GOAL.md to define your objectives"))
		fmt.Printf("  %s\n", ui.Info("3. fest create phase --name \"PLAN\" --after 0"))
		fmt.Printf("  %s\n", ui.Info("4. fest validate (check completion status)"))
	} else {
		fmt.Printf("  %s\n", ui.Info("2. fest create phase --name \"PLAN\" --after 0"))
		fmt.Printf("  %s\n", ui.Info("3. fest create phase --name \"IMPLEMENT\" --after 1"))
		fmt.Printf("  %s\n", ui.Info("4. After creating tasks: fest gates apply --approve"))
	}
	return nil
}

func parseTags(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := []string{}
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func emitCreateFestivalError(opts *CreateFestivalOptions, err error) error {
	if opts.JSONOutput {
		_ = emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     false,
			Action: "create_festival",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitCreateFestivalJSON(opts *CreateFestivalOptions, res createFestivalResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// Slugify converts a string to a URL-safe slug.
func Slugify(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "festival"
	}
	return slug
}
