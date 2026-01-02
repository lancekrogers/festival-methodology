package wizard

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/markers"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui/theme"
	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
)

// FillOptions holds options for the fill command.
type FillOptions struct {
	Path        string
	DryRun      bool
	JSONOutput  bool
	Interactive bool // true = TUI mode, false = text prompts (for agents)
}

type fillResult struct {
	OK      bool             `json:"ok"`
	Action  string           `json:"action"`
	File    string           `json:"file"`
	Filled  int              `json:"filled"`
	Total   int              `json:"total"`
	Skipped int              `json:"skipped"`
	Changes []fillChange     `json:"changes,omitempty"`
	Errors  []map[string]any `json:"errors,omitempty"`
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
			// Detect TTY for interactive mode (TUI for humans, text for agents)
			opts.Interactive = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
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
		// Recursively find all markdown files in the directory tree
		err := filepath.WalkDir(absPath, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			// Skip hidden directories, .festival directory, and gates/ (template files)
			if d.IsDir() {
				name := d.Name()
				if strings.HasPrefix(name, ".") || name == ".festival" || name == "gates" {
					return filepath.SkipDir
				}
				return nil
			}
			// Collect markdown files
			if strings.HasSuffix(d.Name(), ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return emitFillError(opts, errors.Wrap(err, "walking directory").WithField("path", absPath))
		}
		if len(files) == 0 {
			return emitFillError(opts, errors.NotFound("no markdown files in directory"))
		}
	} else {
		files = []string{absPath}
	}

	// Route based on TTY mode
	if opts.Interactive {
		return runVimFill(ctx, opts, files, absPath)
	}

	// Non-TTY mode: sequential prompts for agents
	return runAgentFill(ctx, opts, display, files, absPath)
}

// runVimFill opens the configured editor with all files containing markers.
func runVimFill(ctx context.Context, opts *FillOptions, files []string, rootPath string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("runVimFill")
	}

	display := ui.New(false, false)

	// Filter to only files with markers
	filesWithMarkers := make([]string, 0, len(files))
	for _, f := range files {
		ms, err := markers.ParseFile(ctx, f)
		if err != nil {
			// Log parse errors but continue - file may still be editable
			display.Warning("Could not parse %s: %v", filepath.Base(f), err)
			continue
		}
		if len(ms) > 0 {
			filesWithMarkers = append(filesWithMarkers, f)
		}
	}

	if len(filesWithMarkers) == 0 {
		display.Info("No REPLACE markers found")
		return nil
	}

	// Show summary before opening
	display.Info("Found %d files with REPLACE markers:", len(filesWithMarkers))
	for _, f := range filesWithMarkers {
		relPath, err := filepath.Rel(rootPath, f)
		if err != nil || relPath == "" {
			relPath = filepath.Base(f)
		}
		fmt.Printf("  • %s\n", relPath)
	}
	fmt.Println()
	display.Info("Opening in editor... Use :wqa to save all and quit")
	fmt.Println()

	// Get editor: config > $EDITOR > "vim"
	editor := getEditor(ctx)

	// Open editor with all files as vertical splits (-O flag)
	args := append([]string{"-O"}, filesWithMarkers...)
	cmd := exec.CommandContext(ctx, editor, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run editor - don't treat exit codes as errors since editors like vim
	// return non-zero for normal operations like :q without saving
	_ = cmd.Run()

	return nil
}

// getEditor returns the configured editor from config, $EDITOR, or "vim" as fallback.
func getEditor(ctx context.Context) string {
	// Try config first
	cfg, err := config.Load(ctx, "")
	if err == nil && cfg.Behavior.Editor != "" {
		return cfg.Behavior.Editor
	}

	// Fall back to $EDITOR
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}

	// Default to vim
	return "vim"
}

// runAgentFill handles sequential prompt-based filling for non-TTY (agent) mode.
func runAgentFill(ctx context.Context, opts *FillOptions, display *ui.UI, files []string, rootPath string) error {
	totalFilled := 0
	totalMarkers := 0

	for _, filePath := range files {
		filled, total, err := processFile(ctx, opts, display, rootPath, filePath)
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
func processFile(ctx context.Context, opts *FillOptions, display *ui.UI, rootPath, filePath string) (int, int, error) {
	if err := ctx.Err(); err != nil {
		return 0, 0, errors.Wrap(err, "context cancelled").WithOp("processFile")
	}

	// Calculate display name (relative path from root, or base name if same)
	displayName := filepath.Base(filePath)
	if rootPath != filePath {
		if rel, err := filepath.Rel(rootPath, filePath); err == nil {
			displayName = rel
		}
	}

	// Parse markers from file
	fileMarkers, err := markers.ParseFile(ctx, filePath)
	if err != nil {
		return 0, 0, emitFillError(opts, errors.Wrap(err, "parsing markers").WithField("file", filePath))
	}

	if len(fileMarkers) == 0 {
		if !opts.JSONOutput {
			display.Info("No REPLACE markers found in %s", displayName)
		}
		return 0, 0, nil
	}

	if !opts.JSONOutput {
		fmt.Println()
		display.Info("Found %d markers in %s", len(fileMarkers), displayName)
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
		display.Success("Filled %d/%d markers in %s", filled, total, displayName)
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
		return promptSelect(ctx, opts, m)
	}

	return promptInput(ctx, opts, m)
}

// promptSelect creates a select form for pipe-separated options.
func promptSelect(ctx context.Context, opts *FillOptions, m markers.Marker) (string, bool, error) {
	options := strings.Split(m.Hint, "|")
	for i, opt := range options {
		options[i] = strings.TrimSpace(opt)
	}

	// Non-interactive mode: text-based prompt for agents
	if !opts.Interactive {
		fmt.Printf("Line %d - Select from options:\n", m.LineNumber)
		for i, opt := range options {
			fmt.Printf("  [%d] %s\n", i+1, opt)
		}
		fmt.Printf("  [0] (skip)\n")
		fmt.Print("Enter number: ")

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return "", false, errors.Wrap(err, "reading input")
		}

		input = strings.TrimSpace(input)
		if input == "" || input == "0" {
			return "", true, nil
		}

		// Parse numeric selection
		var idx int
		if _, err := fmt.Sscanf(input, "%d", &idx); err == nil && idx >= 1 && idx <= len(options) {
			return options[idx-1], false, nil
		}

		// Try direct match
		for _, opt := range options {
			if strings.EqualFold(input, opt) {
				return opt, false, nil
			}
		}

		return "", true, nil // Skip if invalid input
	}

	// Interactive mode: TUI
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
	).WithTheme(theme.FestTheme())

	if err := theme.RunForm(ctx, form); err != nil {
		if theme.IsCancelled(err) {
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
func promptInput(ctx context.Context, opts *FillOptions, m markers.Marker) (string, bool, error) {
	// Non-interactive mode: text-based prompt for agents
	if !opts.Interactive {
		fmt.Printf("Line %d: %s\n", m.LineNumber, m.Hint)
		fmt.Print("Enter value (or press Enter to skip): ")

		reader := bufio.NewReader(os.Stdin)
		value, err := reader.ReadString('\n')
		if err != nil {
			return "", false, errors.Wrap(err, "reading input")
		}

		value = strings.TrimSpace(value)
		if value == "" {
			return "", true, nil
		}

		return value, false, nil
	}

	// Interactive mode: TUI
	var value string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(fmt.Sprintf("Line %d: %s", m.LineNumber, m.Hint)).
				Placeholder("(press Enter to skip)").
				Value(&value),
		),
	).WithTheme(theme.FestTheme())

	if err := theme.RunForm(ctx, form); err != nil {
		if theme.IsCancelled(err) {
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
