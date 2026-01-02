package markers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// MarkerOccurrence represents a single marker instance
type MarkerOccurrence struct {
	File       string `json:"file"`
	Line       int    `json:"line"`
	MarkerType string `json:"marker_type"`
	Content    string `json:"content"`
}

// MarkerSummary represents marker statistics
type MarkerSummary struct {
	TotalMarkers int                `json:"total_markers"`
	TotalFiles   int                `json:"total_files"`
	ByType       map[string]int     `json:"by_type"`
	Occurrences  []MarkerOccurrence `json:"occurrences,omitempty"`
}

type markersOptions struct {
	path       string
	jsonOutput bool
}

// NewMarkersCommand creates the markers command group
func NewMarkersCommand() *cobra.Command {
	opts := &markersOptions{}

	cmd := &cobra.Command{
		Use:   "markers",
		Short: "Manage template markers in festival files",
		Long: `View and manage unfilled template markers in festival files.

Template markers are placeholders that must be replaced with actual content:
  [FILL: description]    - Write actual content
  [REPLACE: guidance]    - Replace with your content
  [GUIDANCE: hint]       - Remove and write real content
  {{ placeholder }}      - Fill in the value

Use subcommands to list markers or fill them interactively.`,
	}

	cmd.PersistentFlags().StringVar(&opts.path, "path", "", "Festival path (default: current directory)")
	cmd.PersistentFlags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	// Add subcommands
	cmd.AddCommand(newListCommand(opts))
	cmd.AddCommand(newCountCommand(opts))

	return cmd
}

// newListCommand creates the list subcommand
func newListCommand(opts *markersOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all unfilled template markers",
		Long:  `Scan festival files and list all unfilled template markers with their locations.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(opts)
		},
	}
}

// newCountCommand creates the count subcommand
func newCountCommand(opts *markersOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "count",
		Short: "Count unfilled template markers",
		Long:  `Show a summary count of unfilled template markers by type.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(opts)
		},
	}
}

// resolveFestivalPath resolves the festival root directory
func resolveFestivalPath(pathArg string) (string, error) {
	if pathArg != "" {
		absPath, err := filepath.Abs(pathArg)
		if err != nil {
			return "", errors.Wrap(err, "resolving path").WithField("path", pathArg)
		}
		return absPath, nil
	}

	// Try current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting working directory", err)
	}

	// Try to find festivals/ root
	root, err := tpl.FindFestivalsRoot(cwd)
	if err == nil {
		// Check if we're inside an active festival
		rel, _ := filepath.Rel(root, cwd)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) >= 2 && parts[0] == "active" {
			festivalPath := filepath.Join(root, parts[0], parts[1])
			return festivalPath, nil
		}
	}

	// Use current directory
	return cwd, nil
}

// scanMarkers scans festival files for template markers
func scanMarkers(festivalPath string) (*MarkerSummary, error) {
	markers := []string{"[FILL:", "[REPLACE:", "[GUIDANCE:", "{{ "}

	summary := &MarkerSummary{
		ByType:      make(map[string]int),
		Occurrences: []MarkerOccurrence{},
	}

	filesWithMarkers := make(map[string]bool)

	err := filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Skip hidden directories
		relPath, _ := filepath.Rel(festivalPath, path)
		if strings.HasPrefix(relPath, ".") || strings.Contains(relPath, "/.") {
			return nil
		}
		// Skip gates/ directory - these are template files with intentional markers
		if strings.HasPrefix(relPath, "gates/") || strings.HasPrefix(relPath, "gates"+string(filepath.Separator)) {
			return nil
		}

		// Read file and scan for markers
		file, err := os.Open(path)
		if err != nil {
			return nil
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineNum := 0

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			for _, marker := range markers {
				if strings.Contains(line, marker) {
					// Extract marker content
					startIdx := strings.Index(line, marker)
					endMarker := "]"
					if marker == "{{ " {
						endMarker = "}}"
					}
					endIdx := strings.Index(line[startIdx:], endMarker)
					markerContent := ""
					if endIdx != -1 {
						markerContent = line[startIdx : startIdx+endIdx+len(endMarker)]
					} else {
						markerContent = line[startIdx:]
					}

					// Track occurrence
					summary.Occurrences = append(summary.Occurrences, MarkerOccurrence{
						File:       relPath,
						Line:       lineNum,
						MarkerType: marker,
						Content:    strings.TrimSpace(markerContent),
					})

					summary.ByType[marker]++
					summary.TotalMarkers++
					filesWithMarkers[relPath] = true
				}
			}
		}

		return scanner.Err()
	})

	if err != nil {
		return nil, err
	}

	summary.TotalFiles = len(filesWithMarkers)
	return summary, nil
}

// runList executes the list command
func runList(opts *markersOptions) error {
	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return err
	}

	summary, err := scanMarkers(festivalPath)
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(summary)
	}

	// Human-readable output
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	fmt.Printf("Festival Markers: %s\n", filepath.Base(festivalPath))
	fmt.Println(strings.Repeat("=", 50))

	if summary.TotalMarkers == 0 {
		display.Success("\nNo unfilled template markers found!")
		return nil
	}

	display.Warning("\nFound %d markers in %d files\n", summary.TotalMarkers, summary.TotalFiles)

	// Group by file
	byFile := make(map[string][]MarkerOccurrence)
	for _, occ := range summary.Occurrences {
		byFile[occ.File] = append(byFile[occ.File], occ)
	}

	// Print grouped by file
	for file, occurrences := range byFile {
		fmt.Printf("\n%s (%d markers):\n", file, len(occurrences))
		for _, occ := range occurrences {
			fmt.Printf("  Line %d: %s\n", occ.Line, occ.Content)
		}
	}

	fmt.Println()
	display.Info("Run 'fest validate' to see validation status")
	display.Info("Edit files manually to replace markers with actual content")

	return nil
}

// runCount executes the count command
func runCount(opts *markersOptions) error {
	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return err
	}

	summary, err := scanMarkers(festivalPath)
	if err != nil {
		return err
	}

	if opts.jsonOutput {
		// Only output summary stats, not full occurrences
		countSummary := &MarkerSummary{
			TotalMarkers: summary.TotalMarkers,
			TotalFiles:   summary.TotalFiles,
			ByType:       summary.ByType,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(countSummary)
	}

	// Human-readable output
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	fmt.Printf("Marker Count: %s\n", filepath.Base(festivalPath))
	fmt.Println(strings.Repeat("=", 50))

	if summary.TotalMarkers == 0 {
		display.Success("\nNo unfilled template markers found!")
		return nil
	}

	fmt.Printf("\nTotal Markers: %d\n", summary.TotalMarkers)
	fmt.Printf("Files Affected: %d\n", summary.TotalFiles)

	if len(summary.ByType) > 0 {
		fmt.Println("\nBy Type:")
		for markerType, count := range summary.ByType {
			fmt.Printf("  %s: %d\n", markerType, count)
		}
	}

	fmt.Println()
	display.Info("Run 'fest markers list' to see detailed locations")

	return nil
}
