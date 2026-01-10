package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/markers"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

// MarkerOptions holds options for marker processing.
type MarkerOptions struct {
	FilePath    string // Path to the file to process
	Markers     string // Inline JSON string with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing entirely
	DryRun      bool   // Don't write file, just report markers
	JSONOutput  bool   // Output as JSON
}

// MarkerResult holds the result of marker processing for JSON output.
type MarkerResult struct {
	Markers       []map[string]interface{} `json:"markers,omitempty"`
	Filled        int                      `json:"filled,omitempty"`
	Total         int                      `json:"total,omitempty"`
	UnfilledHints []string                 `json:"unfilled_hints,omitempty"`
}

// ProcessMarkers handles post-creation marker completion.
// Returns a MarkerResult for JSON output integration.
// Always counts markers even when SkipMarkers is true, to provide visibility.
func ProcessMarkers(ctx context.Context, opts MarkerOptions) (*MarkerResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("ProcessMarkers")
	}

	// Parse markers from file - always count for visibility
	foundMarkers, err := markers.ParseFile(ctx, opts.FilePath)
	if err != nil {
		return nil, errors.Wrap(err, "parsing markers").WithField("path", opts.FilePath)
	}

	if len(foundMarkers) == 0 {
		return &MarkerResult{Total: 0}, nil
	}

	// If skip-markers is set, return count only (don't fill)
	if opts.SkipMarkers {
		return &MarkerResult{
			Total:         len(foundMarkers),
			Filled:        0,
			UnfilledHints: markers.ExtractHints(foundMarkers),
		}, nil
	}

	// Dry-run mode: just report markers without filling
	if opts.DryRun {
		return &MarkerResult{
			Markers: markers.MarkersToJSON(foundMarkers),
			Total:   len(foundMarkers),
		}, nil
	}

	// Load input from inline JSON or file
	var input map[string]string

	if opts.Markers != "" {
		input, err = markers.ParseJSON(opts.Markers)
		if err != nil {
			return nil, errors.Wrap(err, "parsing --markers JSON")
		}
	} else if opts.MarkersFile != "" {
		input, err = markers.ReadJSONFile(ctx, opts.MarkersFile)
		if err != nil {
			return nil, err
		}
	}

	// If no input provided, nothing to fill
	if input == nil {
		return &MarkerResult{
			Total:         len(foundMarkers),
			Filled:        0,
			UnfilledHints: markers.ExtractHints(foundMarkers),
		}, nil
	}

	// Apply input to markers
	values := markers.ApplyInput(foundMarkers, input)

	// Replace markers in file
	if err := markers.ReplaceInFile(ctx, opts.FilePath, values); err != nil {
		return nil, err
	}

	// Compute result
	result := markers.ComputeResult(opts.FilePath, values)

	return &MarkerResult{
		Total:         result.TotalMarkers,
		Filled:        result.FilledMarkers,
		UnfilledHints: result.UnfilledHints,
	}, nil
}

// PrintDryRunMarkers outputs markers in the requested format for --dry-run.
func PrintDryRunMarkers(result *MarkerResult, jsonOutput bool) error {
	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(map[string]interface{}{
			"ok":      true,
			"action":  "dry_run",
			"markers": result.Markers,
			"count":   result.Total,
		})
	}

	// Human-readable output
	if result.Total == 0 {
		return nil
	}

	fmt.Println()
	fmt.Println(ui.H2("Replace Markers in Template"))
	for i, m := range result.Markers {
		hint, _ := m["hint"].(string)
		line, _ := m["line"].(int)
		fmt.Printf("  %s %s\n", ui.Value(fmt.Sprintf("%d.", i+1)), ui.Warning(fmt.Sprintf("[line %d] %s", line, hint)))
	}
	fmt.Println()
	fmt.Println(ui.Info("Use --markers '{\"hint\": \"value\", ...}' to fill these markers."))

	return nil
}
