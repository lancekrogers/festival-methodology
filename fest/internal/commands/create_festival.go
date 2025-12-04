package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type createFestivalOptions struct {
	name       string
	goal       string
	tags       string
	varsFile   string
	jsonOutput bool
}

type createFestivalResult struct {
	OK       bool                   `json:"ok"`
	Action   string                 `json:"action"`
	Festival map[string]string      `json:"festival,omitempty"`
	Created  []string               `json:"created,omitempty"`
	Errors   []map[string]any       `json:"errors,omitempty"`
	Warnings []string               `json:"warnings,omitempty"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// NewCreateFestivalCommand adds 'create festival'
func NewCreateFestivalCommand() *cobra.Command {
	opts := &createFestivalOptions{}
	cmd := &cobra.Command{
		Use:   "festival",
		Short: "Create a new festival scaffold under festivals/active",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateFestival(opts)
		},
	}
	cmd.Flags().StringVar(&opts.name, "name", "", "Festival name (required)")
	cmd.Flags().StringVar(&opts.goal, "goal", "", "Festival goal")
	cmd.Flags().StringVar(&opts.tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&opts.varsFile, "vars-file", "", "JSON file with variables")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit JSON output")
	cmd.MarkFlagRequired("name")
	return cmd
}

func runCreateFestival(opts *createFestivalOptions) error {
	display := ui.New(noColor, verbose)
	cwd, _ := os.Getwd()

	// Resolve workspace root and template root
	root, err := tpl.FindWorkspaceRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, fmt.Errorf("%w", err))
	}
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.varsFile) != "" {
		v, err := loadVarsFile(opts.varsFile)
		if err != nil {
			return emitCreateFestivalError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Build context
	ctx := tpl.NewContext()
	ctx.SetFestival(opts.name, opts.goal, parseTags(opts.tags))
	for k, v := range vars {
		ctx.SetCustom(k, v)
	}

	// Destination
	slug := slugify(opts.name)
	destDir := filepath.Join(root, "festivals", "active", slug)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return emitCreateFestivalError(opts, fmt.Errorf("failed to create festival directory: %w", err))
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
		t, err := loader.Load(tpath)
		if err != nil {
			return emitCreateFestivalError(opts, fmt.Errorf("failed to load template %s: %w", c.Template, err))
		}
		outPath := filepath.Join(destDir, c.Out)
		// If template appears to require variables, render; else copy
		requires := t.Metadata != nil && len(t.Metadata.RequiredVariables) > 0
		if requires || strings.Contains(t.Content, "{{") {
			out, err := mgr.Render(t, ctx)
			if err != nil {
				return emitCreateFestivalError(opts, fmt.Errorf("failed to render %s: %w", c.Template, err))
			}
			if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
				return emitCreateFestivalError(opts, fmt.Errorf("failed to write %s: %w", outPath, err))
			}
		} else {
			if err := os.WriteFile(outPath, []byte(t.Content), 0644); err != nil {
				return emitCreateFestivalError(opts, fmt.Errorf("failed to write %s: %w", outPath, err))
			}
		}
		created = append(created, outPath)
	}

	if opts.jsonOutput {
		return emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     true,
			Action: "create_festival",
			Festival: map[string]string{
				"name": opts.name,
				"slug": slug,
			},
			Created: created,
		})
	}

	display.Success("Created festival: %s", slug)
	for _, p := range created {
		display.Info("  • %s", p)
	}
	return nil
}

func slugify(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "festival"
	}
	return slug
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

func emitCreateFestivalError(opts *createFestivalOptions, err error) error {
	if opts.jsonOutput {
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

func emitCreateFestivalJSON(opts *createFestivalOptions, res createFestivalResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
