package gates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	gatescore "github.com/lancekrogers/festival-methodology/fest/internal/gates"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	phase      string // --phase: target phase
	sequence   string // --sequence: target sequence
	dryRun     bool   // --dry-run (default: true)
	approve    bool   // --approve: actually remove
	jsonOutput bool   // --json: output as JSON
}

type removeResult struct {
	OK       bool              `json:"ok"`
	Action   string            `json:"action"`
	DryRun   bool              `json:"dry_run"`
	Removed  []removedFileInfo `json:"removed,omitempty"`
	Summary  removeSummary     `json:"summary"`
	Warnings []string          `json:"warnings,omitempty"`
}

type removedFileInfo struct {
	Path     string `json:"path"`
	GateID   string `json:"gate_id,omitempty"`
	Sequence string `json:"sequence"`
}

type removeSummary struct {
	TotalSequences   int `json:"total_sequences"`
	SequencesUpdated int `json:"sequences_updated"`
	FilesRemoved     int `json:"files_removed"`
}

func newGatesRemoveCmd() *cobra.Command {
	opts := &removeOptions{}

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove quality gate files from sequences",
		Long: `Remove quality gate task files from implementation sequences.

Only files with fest_managed: true marker in frontmatter are removed.
This safely removes only gate files, not regular task files.

By default, runs in dry-run mode showing what would be removed.
Use --approve to actually remove the files.`,
		Example: `  # Preview what would be removed (dry-run is default)
  fest gates remove

  # Remove all gates from all sequences
  fest gates remove --approve

  # Remove gates from specific phase
  fest gates remove --phase 001_IMPLEMENTATION --approve

  # Remove gates from specific sequence
  fest gates remove --sequence 001_IMPL/01_core --approve

  # JSON output for automation
  fest gates remove --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.approve {
				opts.dryRun = false
			}
			return runGatesRemove(cmd.Context(), cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.phase, "phase", "", "Remove from specific phase")
	cmd.Flags().StringVar(&opts.sequence, "sequence", "", "Remove from specific sequence (format: phase/sequence)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", true, "Preview changes without removing (default)")
	cmd.Flags().BoolVar(&opts.approve, "approve", false, "Actually remove files")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output JSON")

	return cmd
}

func runGatesRemove(ctx context.Context, cmd *cobra.Command, opts *removeOptions) error {
	if err := ctx.Err(); err != nil {
		return emitRemoveError(opts, errors.Wrap(err, "context cancelled").WithOp("runGatesRemove"))
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	cwd, err := os.Getwd()
	if err != nil {
		return emitRemoveError(opts, errors.IO("getting working directory", err))
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitRemoveError(opts, errors.Wrap(err, "finding festivals root").WithOp("runGatesRemove"))
	}

	// Resolve paths
	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, opts.phase, opts.sequence)
	if err != nil {
		return emitRemoveError(opts, errors.Wrap(err, "resolving paths").WithOp("runGatesRemove"))
	}

	// Find sequences to process
	var sequences []gatescore.SequenceInfo
	if sequencePath != "" {
		// Single sequence
		sequences = []gatescore.SequenceInfo{{
			Path:      sequencePath,
			PhasePath: phasePath,
			Name:      filepath.Base(sequencePath),
		}}
	} else {
		// Find all sequences in festival
		allSeqs, err := gatescore.FindSequencesWithInfo(festivalPath, nil)
		if err != nil {
			return emitRemoveError(opts, errors.Wrap(err, "finding sequences").WithOp("runGatesRemove"))
		}

		// Filter by phase if specified
		if phasePath != "" {
			for _, seq := range allSeqs {
				if seq.PhasePath == phasePath {
					sequences = append(sequences, seq)
				}
			}
		} else {
			sequences = allSeqs
		}
	}

	if len(sequences) == 0 {
		return emitRemoveResult(opts, removeResult{
			OK:       true,
			Action:   "gates_remove",
			DryRun:   opts.dryRun,
			Warnings: []string{"No sequences found"},
		})
	}

	// Scan for managed gate files
	var toRemove []removedFileInfo
	var warnings []string
	summary := removeSummary{TotalSequences: len(sequences)}

	for _, seq := range sequences {
		files, err := findManagedGateFiles(seq.Path)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Sequence %s: %v", seq.Name, err))
			continue
		}

		if len(files) > 0 {
			summary.SequencesUpdated++
			for _, f := range files {
				toRemove = append(toRemove, removedFileInfo{
					Path:     f.Path,
					GateID:   f.GateID,
					Sequence: seq.Name,
				})
			}
		}
	}

	if len(toRemove) == 0 {
		return emitRemoveResult(opts, removeResult{
			OK:      true,
			Action:  "gates_remove",
			DryRun:  opts.dryRun,
			Summary: summary,
			Removed: []removedFileInfo{},
		})
	}

	// Actually remove if not dry-run
	if !opts.dryRun {
		for _, f := range toRemove {
			if err := os.Remove(f.Path); err != nil {
				warnings = append(warnings, fmt.Sprintf("Failed to remove %s: %v", f.Path, err))
				continue
			}
			summary.FilesRemoved++
		}
	} else {
		summary.FilesRemoved = len(toRemove)
	}

	result := removeResult{
		OK:       true,
		Action:   "gates_remove",
		DryRun:   opts.dryRun,
		Removed:  toRemove,
		Summary:  summary,
		Warnings: warnings,
	}

	if opts.jsonOutput {
		return emitRemoveResult(opts, result)
	}

	// Human-readable output
	out := cmd.OutOrStdout()

	if opts.dryRun {
		display.Info("Dry-run mode (use --approve to remove files)")
	}

	display.Info("Found %d sequences, %d contain gate files", summary.TotalSequences, summary.SequencesUpdated)

	for _, f := range toRemove {
		relPath := f.Path
		if rel, err := filepath.Rel(festivalPath, f.Path); err == nil {
			relPath = rel
		}
		if opts.dryRun {
			display.Warning("  - %s", relPath)
		} else {
			display.Success("  ✓ Removed %s", relPath)
		}
	}

	for _, w := range warnings {
		display.Warning("  Warning: %s", w)
	}

	fmt.Fprintln(out)
	if opts.dryRun {
		display.Info("Summary: %d files would be removed", summary.FilesRemoved)
		display.Info("Run with --approve to remove these files")
	} else {
		display.Info("Summary: %d files removed", summary.FilesRemoved)
	}

	return nil
}

// managedFileInfo holds info about a managed gate file
type managedFileInfo struct {
	Path   string
	GateID string
}

// findManagedGateFiles scans a sequence directory for files with gate markers.
// Looks for both fest_managed: true (newer) and fest_type: gate (legacy).
func findManagedGateFiles(sequencePath string) ([]managedFileInfo, error) {
	var managed []managedFileInfo

	entries, err := os.ReadDir(sequencePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}

		filePath := filepath.Join(sequencePath, name)

		// Check if file has fest_managed marker (newer style)
		if gatescore.IsManaged(filePath) {
			gateID := gatescore.GetGateID(filePath)
			managed = append(managed, managedFileInfo{
				Path:   filePath,
				GateID: gateID,
			})
			continue
		}

		// Also check for fest_type: gate (legacy/current style)
		if isGateType(filePath) {
			managed = append(managed, managedFileInfo{
				Path:   filePath,
				GateID: getGateTypeFromFile(filePath),
			})
		}
	}

	return managed, nil
}

// isGateType checks if a file has fest_type: gate in frontmatter
func isGateType(filePath string) bool {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return false
	}
	// Quick check for fest_type: gate in frontmatter
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				break // End of frontmatter
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(trimmed, "fest_type:") {
			return strings.Contains(trimmed, "gate")
		}
	}
	return false
}

// getGateTypeFromFile extracts fest_gate_type from frontmatter
func getGateTypeFromFile(filePath string) string {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}
	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(trimmed, "fest_gate_type:") {
			parts := strings.SplitN(trimmed, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

func emitRemoveResult(opts *removeOptions, result removeResult) error {
	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	return nil
}

func emitRemoveError(opts *removeOptions, err error) error {
	if opts.jsonOutput {
		result := removeResult{
			OK:       false,
			Action:   "gates_remove",
			DryRun:   opts.dryRun,
			Warnings: []string{err.Error()},
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(result)
	}
	return err
}
