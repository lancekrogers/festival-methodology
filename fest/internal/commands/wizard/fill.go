package wizard

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/markers"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// FillOptions holds options for the fill command.
type FillOptions struct {
	Path       string
	DryRun     bool
	JSONOutput bool
}

type fillResult struct {
	OK       bool              `json:"ok"`
	Action   string            `json:"action"`
	File     string            `json:"file"`
	Filled   int               `json:"filled"`
	Total    int               `json:"total"`
	Skipped  int               `json:"skipped"`
	Changes  []fillChange      `json:"changes,omitempty"`
	Errors   []map[string]any  `json:"errors,omitempty"`
}

type fillChange struct {
	Hint  string `json:"hint"`
	Value string `json:"value"`
	Line  int    `json:"line"`
}

// NewFillCommand creates the wizard fill command.
func NewFillCommand() *cobra.Command {
	opts := &FillOptions{}
	cmd := &cobra.Command{
		Use:   "fill [file-or-directory]",
		Short: "Interactively fill REPLACE markers in festival files",
		Long: `Interactively fill [REPLACE: hint] markers in festival files.

This command scans for REPLACE markers in the specified file (or all .md files
in the directory) and presents an interactive form to fill each one.

MARKER SYNTAX:
  [REPLACE: hint text]           Regular text input
  [REPLACE: Yes|No]              Select from options (pipe-separated)
  [REPLACE: Option A|Option B]   Multiple choice selection

EXAMPLES:
  # Fill markers in a specific file
  fest wizard fill FESTIVAL_GOAL.md

  # Fill markers in current directory
  fest wizard fill .

  # Preview without writing changes
  fest wizard fill PHASE_GOAL.md --dry-run

  # Output results as JSON
  fest wizard fill PHASE_GOAL.md --json

The fill wizard transforms tedious manual editing into a guided experience,
ensuring all template markers are properly completed.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts.Path = "."
			if len(args) > 0 {
				opts.Path = args[0]
			}
			return RunFill(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Preview changes without writing")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Output results as JSON")

	return cmd
}

// RunFill executes the fill command logic.
func RunFill(ctx context.Context, opts *FillOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunFill")
	}

	display := ui.New(false, false)

	// Resolve the path
	absPath, err := filepath.Abs(opts.Path)
	if err != nil {
		return emitFillError(opts, errors.Wrap(err, "resolving path").WithField("path", opts.Path))
	}

	// Check if path is file or directory
	info, err := os.Stat(absPath)
	if err != nil {
		return emitFillError(opts, errors.Wrap(err, "accessing path").WithField("path", absPath))
	}

	var files []string
	if info.IsDir() {
		// Find all markdown files in the directory
		entries, err := os.ReadDir(absPath)
		if err != nil {
			return emitFillError(opts, errors.Wrap(err, "reading directory").WithField("path", absPath))
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				files = append(files, filepath.Join(absPath, entry.Name()))
			}
		}
		if len(files) == 0 {
			return emitFillError(opts, errors.NotFound("no markdown files in directory"))
		}
	} else {
		files = []string{absPath}
	}

	// Process each file
	totalFilled := 0
	totalMarkers := 0

	for _, filePath := range files {
		filled, total, err := processFile(ctx, opts, display, filePath)
		if err != nil {
			return err
		}
		totalFilled += filled
		totalMarkers += total
	}

	// Final summary
	if !opts.JSONOutput && len(files) > 1 {
		fmt.Println()
		display.Success("Processed %d files: filled %d/%d markers", len(files), totalFilled, totalMarkers)
	}

	return nil
}

// processFile handles marker filling for a single file.
func processFile(ctx context.Context, opts *FillOptions, display *ui.UI, filePath string) (int, int, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, errors.Wrap(err, "context cancelled").WithOp("processFile")
	}

	// Parse markers from file
	fileMarkers, err := markers.ParseFile(ctx, filePath)
	if err != nil {
		return 0, 0, emitFillError(opts, errors.Wrap(err, "parsing markers").WithField("file", filePath))
	}

	if len(fileMarkers) == 0 {
		if !opts.JSONOutput {
			display.Info("No REPLACE markers found in %s", filepath.Base(filePath))
		}
		return 0, 0, nil
	}

	if !opts.JSONOutput {
		fmt.Println()
		display.Info("Found %d markers in %s", len(fileMarkers), filepath.Base(filePath))
		fmt.Println()
	}

	// Build input map for values
	input := make(map[string]string)
	changes := []fillChange{}
	skipped := 0

	// Process each marker
	for _, m := range fileMarkers {
		value, skip, err := promptForMarker(ctx, opts, m)
		if err != nil {
			return 0, 0, err
		}

		if skip {
			skipped++
			continue
		}

		if value != "" {
			input[m.Hint] = value
			changes = append(changes, fillChange{
				Hint:  m.Hint,
				Value: value,
				Line:  m.LineNumber,
			})
		}
	}

	filled := len(changes)
	total := len(fileMarkers)

	// Apply changes unless dry-run
	if !opts.DryRun && filled > 0 {
		values := markers.ApplyInput(fileMarkers, input)
		if err := markers.ReplaceInFile(ctx, filePath, values); err != nil {
			return 0, 0, emitFillError(opts, errors.Wrap(err, "writing file").WithField("file", filePath))
		}
	}

	// Output results
	if opts.JSONOutput {
		return filled, total, emitFillJSON(opts, fillResult{
			OK:      true,
			Action:  "fill",
			File:    filePath,
			Filled:  filled,
			Total:   total,
			Skipped: skipped,
			Changes: changes,
		})
	}

	// Human-readable output
	fmt.Println()
	if opts.DryRun {
		display.Warning("Dry run - no changes written")
		if filled > 0 {
			fmt.Println()
			display.Info("Would fill %d/%d markers:", filled, total)
			for _, c := range changes {
				fmt.Printf("  Line %d: %s → %s\n", c.Line, c.Hint, truncate(c.Value, 40))
			}
		}
	} else if filled > 0 {
		display.Success("Filled %d/%d markers in %s", filled, total, filepath.Base(filePath))
	} else {
		display.Info("No markers filled (all skipped or empty)")
	}

	return filled, total, nil
}

// promptForMarker presents an interactive prompt for a single marker.
func promptForMarker(ctx context.Context, opts *FillOptions, m markers.Marker) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, errors.Wrap(err, "context cancelled")
	}

	// Check if hint contains pipe-separated options
	if strings.Contains(m.Hint, "|") {
		return promptSelect(m)
	}

	return promptInput(m)
}

// promptSelect creates a select form for pipe-separated options.
func promptSelect(m markers.Marker) (string, bool, error) {
	options := strings.Split(m.Hint, "|")
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	// Add skip option
	options = append(options, "(skip)")

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Line %d", m.LineNumber)).
				Description(m.Hint).
				Options(toOptions(options)...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return "", true, nil
		}
		return "", false, errors.Wrap(err, "form error")
	}

	if selected == "(skip)" {
		return "", true, nil
	}

	return selected, false, nil
}

// promptInput creates a text input form for regular hints.
func promptInput(m markers.Marker) (string, bool, error) {
	var value string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Line %d: %s", m.LineNumber, m.Hint)).
				Placeholder("(press Enter to skip)").
				Value(&value),
		),
	)

	if err := form.Run(); err != nil {
		if err == huh.ErrUserAborted {
			return "", true, nil
		}
		return "", false, errors.Wrap(err, "form error")
	}

	if strings.TrimSpace(value) == "" {
		return "", true, nil
	}

	return value, false, nil
}

// toOptions converts string slice to huh.Option slice.
func toOptions(items []string) []huh.Option[string] {
	opts := make([]huh.Option[string], len(items))
	for i, item := range items {
		opts[i] = huh.NewOption(item, item)
	}
	return opts
}

// truncate shortens a string if it exceeds max length.
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

// emitFillError handles error output in JSON or text format.
func emitFillError(opts *FillOptions, err error) error {
	if opts.JSONOutput {
		_ = emitFillJSON(opts, fillResult{
			OK:     false,
			Action: "fill",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

// emitFillJSON outputs the result as JSON.
func emitFillJSON(opts *FillOptions, res fillResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}
