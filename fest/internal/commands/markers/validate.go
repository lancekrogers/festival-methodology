package markers

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/markers"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// ValidateResult contains the validation results
type ValidateResult struct {
	Valid          bool     `json:"valid"`
	MissingMarkers []string `json:"missing_markers,omitempty"`
	EmptyMarkers   []string `json:"empty_markers,omitempty"`
	UnknownMarkers []string `json:"unknown_markers,omitempty"`
	Suggestions    []string `json:"suggestions,omitempty"`
	TotalExpected  int      `json:"total_expected"`
	TotalProvided  int      `json:"total_provided"`
	TotalFilled    int      `json:"total_filled"`
}

type validateOptions struct {
	file     string
	template string
	source   string
	strict   bool
}

// newValidateCommand creates the validate subcommand
func newValidateCommand(opts *markersOptions) *cobra.Command {
	validateOpts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate marker JSON against template",
		Long: `Validate a marker JSON or YAML file against a template to detect issues.

This command checks for:
  - Missing required markers (present in template but not in file)
  - Empty marker values
  - Unknown markers (possible typos)

In strict mode (--strict), unknown markers cause validation to fail.

Examples:
  # Validate against built-in template
  fest markers validate --file markers.json --template task-simple

  # Validate against existing file
  fest markers validate --file markers.json --source PHASE_GOAL.md

  # Strict mode - fail on unknown markers
  fest markers validate --file markers.json --template task --strict`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runValidate(validateOpts, opts)
		},
	}

	cmd.Flags().StringVar(&validateOpts.file, "file", "", "Marker file to validate (required)")
	cmd.Flags().StringVar(&validateOpts.template, "template", "", "Template to validate against")
	cmd.Flags().StringVar(&validateOpts.source, "source", "", "Source file to validate against")
	cmd.Flags().BoolVar(&validateOpts.strict, "strict", false, "Fail on unknown markers")

	_ = cmd.MarkFlagRequired("file")

	return cmd
}

func runValidate(validateOpts *validateOptions, opts *markersOptions) error {
	// Validate inputs
	if validateOpts.file == "" {
		return errors.Validation("--file is required")
	}
	if validateOpts.template == "" && validateOpts.source == "" {
		return errors.Validation("must specify either --template or --source")
	}
	if validateOpts.template != "" && validateOpts.source != "" {
		return errors.Validation("cannot specify both --template and --source")
	}

	// Load marker file
	providedMarkers, err := loadMarkerFile(validateOpts.file)
	if err != nil {
		return err
	}

	// Get expected markers from template/source
	var expectedHints []string
	if validateOpts.template != "" {
		_, content, err := resolveTemplate(validateOpts.template)
		if err != nil {
			return err
		}
		parsedMarkers := markers.Parse(string(content))
		expectedHints = deduplicateHints(markers.ExtractHints(parsedMarkers))
	} else {
		content, err := os.ReadFile(validateOpts.source)
		if err != nil {
			return errors.Wrap(err, "reading source file").WithField("path", validateOpts.source)
		}
		parsedMarkers := markers.Parse(string(content))
		expectedHints = deduplicateHints(markers.ExtractHints(parsedMarkers))
	}

	// Validate
	result := validateMarkers(providedMarkers, expectedHints, validateOpts.strict)

	// Output
	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}

	return outputValidateHuman(result, validateOpts.strict)
}

func loadMarkerFile(path string) (map[string]string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.Wrap(err, "reading marker file").WithField("path", path)
	}

	// Try to parse as scaffold output format first
	var scaffold ScaffoldOutput
	if err := json.Unmarshal(content, &scaffold); err == nil && scaffold.Markers != nil {
		return scaffold.Markers, nil
	}

	// Try YAML scaffold format
	if err := yaml.Unmarshal(content, &scaffold); err == nil && scaffold.Markers != nil {
		return scaffold.Markers, nil
	}

	// Try as plain map (JSON)
	var plainMap map[string]string
	if err := json.Unmarshal(content, &plainMap); err == nil {
		return plainMap, nil
	}

	// Try as plain map (YAML)
	if err := yaml.Unmarshal(content, &plainMap); err == nil {
		return plainMap, nil
	}

	return nil, errors.Validation("could not parse marker file as JSON or YAML")
}

func validateMarkers(provided map[string]string, expected []string, strict bool) *ValidateResult {
	result := &ValidateResult{
		Valid:         true,
		TotalExpected: len(expected),
		TotalProvided: len(provided),
	}

	// Build set of expected hints
	expectedSet := make(map[string]bool)
	for _, hint := range expected {
		expectedSet[hint] = true
	}

	// Check for missing and empty markers
	for _, hint := range expected {
		value, exists := provided[hint]
		if !exists {
			result.MissingMarkers = append(result.MissingMarkers, hint)
			result.Valid = false
		} else if strings.TrimSpace(value) == "" {
			result.EmptyMarkers = append(result.EmptyMarkers, hint)
		} else {
			result.TotalFilled++
		}
	}

	// Check for unknown markers (potential typos)
	for hint := range provided {
		if hint == "_meta" {
			continue // Skip metadata field
		}
		if !expectedSet[hint] {
			result.UnknownMarkers = append(result.UnknownMarkers, hint)
			// Try to suggest a correction
			if suggestion := findSimilarHint(hint, expected); suggestion != "" {
				result.Suggestions = append(result.Suggestions, fmt.Sprintf("%q → %q", hint, suggestion))
			}
			if strict {
				result.Valid = false
			}
		}
	}

	// Sort results for consistent output
	sort.Strings(result.MissingMarkers)
	sort.Strings(result.EmptyMarkers)
	sort.Strings(result.UnknownMarkers)
	sort.Strings(result.Suggestions)

	return result
}

// findSimilarHint finds a similar hint using simple string matching
func findSimilarHint(unknown string, expected []string) string {
	unknownLower := strings.ToLower(unknown)
	bestMatch := ""
	bestScore := 0

	for _, hint := range expected {
		hintLower := strings.ToLower(hint)

		// Simple similarity score based on common substrings
		score := 0

		// Check if one contains the other
		if strings.Contains(unknownLower, hintLower) || strings.Contains(hintLower, unknownLower) {
			score += 50
		}

		// Count common words
		unknownWords := strings.Fields(unknownLower)
		hintWords := strings.Fields(hintLower)
		for _, uw := range unknownWords {
			for _, hw := range hintWords {
				if uw == hw {
					score += 10
				} else if strings.HasPrefix(uw, hw) || strings.HasPrefix(hw, uw) {
					score += 5
				}
			}
		}

		// Check first/last characters match
		if len(unknownLower) > 0 && len(hintLower) > 0 {
			if unknownLower[0] == hintLower[0] {
				score += 5
			}
			if unknownLower[len(unknownLower)-1] == hintLower[len(hintLower)-1] {
				score += 5
			}
		}

		if score > bestScore && score >= 20 { // Minimum threshold
			bestScore = score
			bestMatch = hint
		}
	}

	return bestMatch
}

func outputValidateHuman(result *ValidateResult, strict bool) error {
	display := ui.New(false, false)

	fmt.Println(ui.H1("Marker Validation"))
	fmt.Printf("%s %d expected, %d provided, %d filled\n\n",
		ui.Label("Summary"),
		result.TotalExpected,
		result.TotalProvided,
		result.TotalFilled)

	hasIssues := false

	// Missing markers
	if len(result.MissingMarkers) > 0 {
		hasIssues = true
		display.Error("Missing markers (%d):", len(result.MissingMarkers))
		for _, m := range result.MissingMarkers {
			fmt.Printf("  • %s\n", m)
		}
		fmt.Println()
	}

	// Empty markers
	if len(result.EmptyMarkers) > 0 {
		display.Warning("Empty markers (%d):", len(result.EmptyMarkers))
		for _, m := range result.EmptyMarkers {
			fmt.Printf("  • %s\n", m)
		}
		fmt.Println()
	}

	// Unknown markers
	if len(result.UnknownMarkers) > 0 {
		if strict {
			hasIssues = true
			display.Error("Unknown markers (%d) - strict mode:", len(result.UnknownMarkers))
		} else {
			display.Warning("Unknown markers (%d) - possible typos:", len(result.UnknownMarkers))
		}
		for _, m := range result.UnknownMarkers {
			fmt.Printf("  • %s\n", m)
		}
		if len(result.Suggestions) > 0 {
			fmt.Println("\n  Suggestions:")
			for _, s := range result.Suggestions {
				fmt.Printf("    %s\n", s)
			}
		}
		fmt.Println()
	}

	// Final status
	if result.Valid {
		if len(result.EmptyMarkers) == 0 && len(result.UnknownMarkers) == 0 {
			display.Success("All markers valid and filled!")
		} else {
			display.Success("Validation passed (with warnings)")
		}
	} else {
		display.Error("Validation failed")
		if !strict && hasIssues {
			display.Info("Use --strict to fail on unknown markers")
		}
		return errors.Validation("marker validation failed")
	}

	return nil
}
