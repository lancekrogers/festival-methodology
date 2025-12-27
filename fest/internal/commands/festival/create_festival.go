package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// CreateFestivalOptions holds options for the create festival command.
type CreateFestivalOptions struct {
	Name        string
	Goal        string
	Tags        string
	VarsFile    string
	Markers     string // Inline JSON with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing
	DryRun      bool   // Show markers without creating file
	JSONOutput  bool
	Dest        string // "active" or "planned"
}

type createFestivalResult struct {
	OK            bool                     `json:"ok"`
	Action        string                   `json:"action"`
	Festival      map[string]string        `json:"festival,omitempty"`
	Created       []string                 `json:"created,omitempty"`
	Markers       []map[string]interface{} `json:"markers,omitempty"`
	MarkersFilled int                      `json:"markers_filled,omitempty"`
	MarkersTotal  int                      `json:"markers_total,omitempty"`
	Errors        []map[string]any         `json:"errors,omitempty"`
	Warnings      []string                 `json:"warnings,omitempty"`
	Extra         map[string]interface{}   `json:"extra,omitempty"`
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
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON file with variables")
	cmd.Flags().StringVar(&opts.Markers, "markers", "", "JSON string with REPLACE marker hint→value mappings")
	cmd.Flags().StringVar(&opts.MarkersFile, "markers-file", "", "JSON file with REPLACE marker hint→value mappings")
	cmd.Flags().BoolVar(&opts.SkipMarkers, "skip-markers", false, "Skip REPLACE marker processing")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show template markers without creating file")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&opts.Dest, "dest", "active", "Destination under festivals/: active or planned")
	return cmd
}

// RunCreateFestival executes the create festival command logic.
func RunCreateFestival(ctx context.Context, opts *CreateFestivalOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreateFestival")
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve festivals root and template root
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}
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
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// Destination
	slug := Slugify(opts.Name)
	destCategory := strings.ToLower(strings.TrimSpace(opts.Dest))
	if destCategory != "planned" && destCategory != "active" {
		destCategory = "active"
	}
	destDir := filepath.Join(festivalsRoot, destCategory, slug)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return emitCreateFestivalError(opts, errors.IO("creating festival directory", err).WithField("path", destDir))
	}

	// Render/copy core files
	mgr := tpl.NewManager()
	created := []string{}

	core := []struct{ Template, Out string }{
		{"FESTIVAL_OVERVIEW_TEMPLATE.md", "FESTIVAL_OVERVIEW.md"},
		{"FESTIVAL_GOAL_TEMPLATE.md", "FESTIVAL_GOAL.md"},
		{"FESTIVAL_RULES_TEMPLATE.md", "FESTIVAL_RULES.md"},
		{"FESTIVAL_TODO_TEMPLATE.md", "TODO.md"},
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
		if requires || strings.Contains(t.Content, "{{") {
			out, err := mgr.Render(t, tmplCtx)
			if err != nil {
				return emitCreateFestivalError(opts, errors.Wrap(err, "rendering template").WithField("template", c.Template))
			}
			if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing file", err).WithField("path", outPath))
			}
		} else {
			if err := os.WriteFile(outPath, []byte(t.Content), 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing file", err).WithField("path", outPath))
			}
		}
		created = append(created, outPath)
	}

	// Process REPLACE markers in all created files
	var totalMarkersFilled, totalMarkersCount int
	var allMarkers []map[string]interface{}

	for _, filePath := range created {
		markerResult, err := ProcessMarkers(ctx, MarkerOptions{
			FilePath:    filePath,
			Markers:     opts.Markers,
			MarkersFile: opts.MarkersFile,
			SkipMarkers: opts.SkipMarkers,
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

	if opts.JSONOutput {
		return emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     true,
			Action: "create_festival",
			Festival: map[string]string{
				"name": opts.Name,
				"slug": slug,
				"dest": destCategory,
			},
			Created:       created,
			MarkersFilled: totalMarkersFilled,
			MarkersTotal:  totalMarkersCount,
			Warnings: []string{
				"Next: Create phases with 'fest create phase --name PHASE_NAME'",
			},
		})
	}

	display.Success("Created festival: %s (%s)", slug, destCategory)
	for _, p := range created {
		display.Info("  • %s", p)
	}

	// Report marker filling status
	if totalMarkersCount > 0 {
		if totalMarkersFilled == totalMarkersCount {
			display.Success("Filled %d/%d REPLACE markers", totalMarkersFilled, totalMarkersCount)
		} else {
			display.Warning("Filled %d/%d REPLACE markers (%d remaining)",
				totalMarkersFilled, totalMarkersCount, totalMarkersCount-totalMarkersFilled)
		}
	}

	fmt.Println()
	fmt.Println("   Next steps:")
	fmt.Println("   1. cd", destDir)
	fmt.Println("   2. fest create phase --name \"PLAN\"")
	fmt.Println("   3. fest create phase --name \"IMPLEMENT\"")
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
