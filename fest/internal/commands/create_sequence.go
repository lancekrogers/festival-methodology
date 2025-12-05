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

type createSequenceOptions struct {
	after      int
	name       string
	path       string
	varsFile   string
	jsonOutput bool
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
	opts := &createSequenceOptions{}
	cmd := &cobra.Command{
		Use:   "sequence",
		Short: "Insert a new sequence and render its goal file",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCreateSequence(opts)
		},
	}
	cmd.Flags().IntVar(&opts.after, "after", 0, "Insert after this number (0 inserts at beginning)")
	cmd.Flags().StringVar(&opts.name, "name", "", "Sequence name (required)")
	cmd.Flags().StringVar(&opts.path, "path", ".", "Path to phase directory (directory containing numbered sequences)")
	cmd.Flags().StringVar(&opts.varsFile, "vars-file", "", "JSON vars for rendering")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Emit JSON output")
	cmd.MarkFlagRequired("name")
	return cmd
}

func runCreateSequence(opts *createSequenceOptions) error {
	display := ui.New(noColor, verbose)
	cwd, _ := os.Getwd()

	// Resolve template root
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateSequenceError(opts, err)
	}

	absPath, err := filepath.Abs(opts.path)
	if err != nil {
		return emitCreateSequenceError(opts, fmt.Errorf("invalid path: %w", err))
	}

	// Insert sequence
	ren := festival.NewRenumberer(festival.RenumberOptions{})
	if err := ren.InsertSequence(absPath, opts.after, opts.name); err != nil {
		return emitCreateSequenceError(opts, fmt.Errorf("failed to insert sequence: %w", err))
	}

	// Compute new sequence id
	newNumber := opts.after + 1
	seqID := tpl.FormatSequenceID(newNumber, opts.name)
	seqDir := filepath.Join(absPath, seqID)

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.varsFile) != "" {
		v, err := loadVarsFile(opts.varsFile)
		if err != nil {
			return emitCreateSequenceError(opts, fmt.Errorf("failed to read vars-file: %w", err))
		}
		vars = v
	}

	// Build context for sequence
	ctx := tpl.NewContext()
	ctx.SetSequence(newNumber, opts.name)
	for k, v := range vars {
		ctx.SetCustom(k, v)
	}

	// Render or copy SEQUENCE_GOAL template
	catalog, _ := tpl.LoadCatalog(tmplRoot)
	mgr := tpl.NewManager()
	var content string
	var renderErr error
	if catalog != nil {
		content, renderErr = mgr.RenderByID(catalog, "SEQUENCE_GOAL", ctx)
	}
	if renderErr != nil || content == "" {
		// Fall back to default filename
		tpath := filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(tpath)
			if err != nil {
				return emitCreateSequenceError(opts, fmt.Errorf("failed to load sequence goal template: %w", err))
			}
			// Render if it appears templated; else copy
			if strings.Contains(t.Content, "{{") {
				out, err := mgr.Render(t, ctx)
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

	if opts.jsonOutput {
		return emitCreateSequenceJSON(opts, createSequenceResult{
			OK:     true,
			Action: "create_sequence",
			Sequence: map[string]interface{}{
				"number": newNumber,
				"id":     seqID,
				"name":   opts.name,
			},
			Created:  []string{goalPath},
			Renumber: []string{},
		})
	}

	display.Success("Created sequence: %s", seqID)
	display.Info("  • %s", goalPath)
	return nil
}

func emitCreateSequenceError(opts *createSequenceOptions, err error) error {
	if opts.jsonOutput {
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

func emitCreateSequenceJSON(opts *createSequenceOptions, res createSequenceResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
