package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/fileops"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// ApplyOptions holds options for the apply command.
type ApplyOptions struct {
	TemplateID   string
	TemplatePath string
	DestPath     string
	VarsFile     string
	JSONOutput   bool
}

type applyResult struct {
	OK          bool                   `json:"ok"`
	Action      string                 `json:"action"`
	Template    map[string]string      `json:"template,omitempty"`
	Destination string                 `json:"destination,omitempty"`
	Mode        string                 `json:"mode,omitempty"`
	Errors      []map[string]any       `json:"errors,omitempty"`
	Warnings    []string               `json:"warnings,omitempty"`
	Extra       map[string]interface{} `json:"extra,omitempty"`
}

// NewApplyCommand creates the 'apply' command (copy-first single template)
func NewApplyCommand() *cobra.Command {
	opts := &ApplyOptions{}

	cmd := &cobra.Command{
		Use:   "apply",
		Short: "Apply a local template to a destination file (copy or render)",
		Long:  "Apply a local template (from .festival/templates) to a destination file. Copy if no variables provided; render if variables exist.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return RunApply(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.TemplateID, "template-id", "", "Template ID or alias (from frontmatter)")
	cmd.Flags().StringVar(&opts.TemplatePath, "template-path", "", "Path to template file (relative to .festival/templates or absolute)")
	cmd.Flags().StringVar(&opts.DestPath, "to", "", "Destination file path (required)")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON file providing variables for rendering")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")

	cmd.MarkFlagRequired("to")

	return cmd
}

// RunApply executes the apply command logic.
func RunApply(ctx context.Context, opts *ApplyOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve local template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("%w", err))
	}

	// Load vars if provided
	var vars map[string]interface{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitApplyError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	} else {
		vars = map[string]interface{}{}
	}

	// Locate template path
	tpath := opts.TemplatePath
	tmplID := opts.TemplateID
	if tpath == "" && tmplID == "" {
		return emitApplyError(opts, fmt.Errorf("must provide --template-id or --template-path"))
	}

	if tpath == "" {
		// resolve by ID using catalog
		catalog, err := tpl.LoadCatalog(ctx, tmplRoot)
		if err != nil {
			return emitApplyError(opts, fmt.Errorf("failed to load template catalog: %w", err))
		}
		if p, ok := catalog.Resolve(tmplID); ok {
			tpath = p
		} else {
			return emitApplyError(opts, fmt.Errorf("unknown template id: %s", tmplID))
		}
	} else {
		// If relative, make it relative to template root
		if !filepath.IsAbs(tpath) {
			tpath = filepath.Join(tmplRoot, tpath)
		}
	}

	// Ensure destination parent exists
	if err := os.MkdirAll(filepath.Dir(opts.DestPath), 0755); err != nil {
		return emitApplyError(opts, fmt.Errorf("failed to create destination directory: %w", err))
	}

	// Decide copy vs render
	mgr := tpl.NewManager()
	loader := tpl.NewLoader()
	tmpl, err := loader.Load(ctx, tpath)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("failed to load template: %w", err))
	}

	mode := "copy"
	// Build template context
	tmplCtx := tpl.NewContext()
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// If template has required variables or contains '{{', render; else copy
	requiresVars := tmpl.Metadata != nil && len(tmpl.Metadata.RequiredVariables) > 0
	containsDelims := strings.Contains(tmpl.Content, "{{")
	if requiresVars || containsDelims {
		mode = "render"
		// Validate missing vars
		if tmpl.Metadata != nil && len(tmpl.Metadata.RequiredVariables) > 0 {
			missing := []string{}
			for _, req := range tmpl.Metadata.RequiredVariables {
				if _, ok := tmplCtx.Get(req); !ok {
					missing = append(missing, req)
				}
			}
			if len(missing) > 0 {
				return emitApplyJSON(opts, applyResult{
					OK:     false,
					Action: "apply",
					Errors: []map[string]any{{
						"code":    "missing_vars",
						"message": "missing required variables",
						"details": map[string]any{"required": missing},
					}},
				})
			}
		}
		out, err := mgr.Render(tmpl, tmplCtx)
		if err != nil {
			return emitApplyError(opts, fmt.Errorf("failed to render: %w", err))
		}
		if err := os.WriteFile(opts.DestPath, []byte(out), 0644); err != nil {
			return emitApplyError(opts, fmt.Errorf("failed to write destination: %w", err))
		}
	} else {
		if err := fileops.CopyFile(ctx, tpath, opts.DestPath); err != nil {
			return emitApplyError(opts, fmt.Errorf("failed to copy: %w", err))
		}
	}

	if opts.JSONOutput {
		return emitApplyJSON(opts, applyResult{
			OK:     true,
			Action: "apply",
			Template: map[string]string{
				"id":   tmplID,
				"path": tpath,
			},
			Destination: opts.DestPath,
			Mode:        mode,
		})
	}

	display.Success("Applied template → %s (%s)", opts.DestPath, mode)
	return nil
}

func emitApplyError(opts *ApplyOptions, err error) error {
	if opts.JSONOutput {
		_ = emitApplyJSON(opts, applyResult{
			OK:     false,
			Action: "apply",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitApplyJSON(opts *ApplyOptions, res applyResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// loadVarsFile reads a JSON file and returns a map
func loadVarsFile(path string) (map[string]interface{}, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, err
	}
	return m, nil
}
