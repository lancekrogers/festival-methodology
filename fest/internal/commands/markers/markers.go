package markers

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// stripInlineCode removes text inside backticks from a line.
// This prevents markers inside inline code examples from being counted.
func stripInlineCode(line string) string {
	result := strings.Builder{}
	inBacktick := false
	for _, ch := range line {
		if ch == '`' {
			inBacktick = !inBacktick
			continue
		}
		if !inBacktick {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

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

// FilePriority defines sorting priority for file types
type FilePriority int

const (
	PriorityGoal FilePriority = iota
	PriorityOverview
	PriorityTodo
	PriorityOther
)

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
	cmd.AddCommand(newNextCommand(opts))
	cmd.AddCommand(newScaffoldCommand(opts))
	cmd.AddCommand(newValidateCommand(opts))

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
		defer func() { _ = file.Close() }()

		scanner := bufio.NewScanner(file)
		lineNum := 0
		inCodeBlock := false

		for scanner.Scan() {
			lineNum++
			line := scanner.Text()

			// Toggle fence state on ``` lines (handles ```go, ```yaml, etc.)
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}

			// Skip markers inside code blocks - they're documentation examples
			if inCodeBlock {
				continue
			}

			// Strip inline code (backticks) before checking for markers
			lineWithoutCode := stripInlineCode(line)

			for _, marker := range markers {
				if strings.Contains(lineWithoutCode, marker) {
					// Extract marker content from original line
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

	fmt.Println(ui.H1("Festival Markers"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(filepath.Base(festivalPath), ui.FestivalColor))

	if summary.TotalMarkers == 0 {
		display.Success("No unfilled template markers found!")
		return nil
	}

	display.Warning("Found %d markers in %d files", summary.TotalMarkers, summary.TotalFiles)

	// Group by file
	byFile := make(map[string][]MarkerOccurrence)
	for _, occ := range summary.Occurrences {
		byFile[occ.File] = append(byFile[occ.File], occ)
	}

	// Print grouped by file
	for file, occurrences := range byFile {
		fmt.Printf("\n%s\n", ui.H2(fmt.Sprintf("%s (%d markers)", file, len(occurrences))))
		for _, occ := range occurrences {
			fmt.Printf("  %s %s %s\n",
				ui.Label("Line"),
				ui.Value(fmt.Sprintf("%d", occ.Line)),
				ui.Warning(occ.Content),
			)
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

	fmt.Println(ui.H1("Marker Count"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(filepath.Base(festivalPath), ui.FestivalColor))

	if summary.TotalMarkers == 0 {
		display.Success("No unfilled template markers found!")
		return nil
	}

	fmt.Printf("%s %s\n", ui.Label("Total Markers"), ui.Value(fmt.Sprintf("%d", summary.TotalMarkers)))
	fmt.Printf("%s %s\n", ui.Label("Files Affected"), ui.Value(fmt.Sprintf("%d", summary.TotalFiles)))

	if len(summary.ByType) > 0 {
		fmt.Println()
		fmt.Println(ui.H3("By Type"))
		for markerType, count := range summary.ByType {
			fmt.Printf("  %s %s\n", ui.Label(markerType), ui.Value(fmt.Sprintf("%d", count)))
		}
	}

	fmt.Println()
	display.Info("Run 'fest markers list' to see detailed locations")

	return nil
}

// getFilePriority returns the priority of a file based on its name
func getFilePriority(filename string) FilePriority {
	upper := strings.ToUpper(filename)
	switch {
	case strings.Contains(upper, "GOAL"):
		return PriorityGoal
	case strings.Contains(upper, "OVERVIEW"):
		return PriorityOverview
	case strings.Contains(upper, "TODO"):
		return PriorityTodo
	default:
		return PriorityOther
	}
}

// getPathDepth returns the number of directory levels in a path
func getPathDepth(path string) int {
	path = filepath.Clean(path)
	if path == "." {
		return 0
	}
	return len(strings.Split(path, string(filepath.Separator)))
}

// extractNumericPrefix extracts leading numbers from a filename/dirname
// e.g., "001_PLANNING" -> 1, "02_task.md" -> 2
func extractNumericPrefix(name string) int {
	// Remove directory path
	base := filepath.Base(name)

	// Extract leading digits
	var numStr string
	for _, r := range base {
		if r >= '0' && r <= '9' {
			numStr += string(r)
		} else if r == '_' || r == '-' {
			break
		} else {
			break
		}
	}

	if numStr == "" {
		return 9999 // Sort unnumbered items last
	}

	var num int
	_, err := fmt.Sscanf(numStr, "%d", &num)
	if err != nil {
		return 9999
	}
	return num
}

// sortMarkerFiles sorts files by hierarchy (festival -> phase -> sequence -> task)
func sortMarkerFiles(files []string) []string {
	filesCopy := make([]string, len(files))
	copy(filesCopy, files)

	sort.Slice(filesCopy, func(i, j int) bool {
		pathI, pathJ := filesCopy[i], filesCopy[j]

		// Split paths into components
		partsI := strings.Split(filepath.Clean(pathI), string(filepath.Separator))
		partsJ := strings.Split(filepath.Clean(pathJ), string(filepath.Separator))

		// Compare each path component level by level (directory hierarchy first)
		minLen := len(partsI)
		if len(partsJ) < minLen {
			minLen = len(partsJ)
		}

		for level := 0; level < minLen-1; level++ {
			compI, compJ := partsI[level], partsJ[level]
			if compI != compJ {
				// Compare numeric prefixes
				numI, numJ := extractNumericPrefix(compI), extractNumericPrefix(compJ)
				if numI != numJ {
					return numI < numJ
				}
				// Same number, alphabetical
				return compI < compJ
			}
		}

		// Same directory path up to this point
		// If one path is deeper, it comes after (parent directories process their direct children first)
		if len(partsI) != len(partsJ) {
			return len(partsI) < len(partsJ)
		}

		// Same depth and same directory path, compare filenames by priority
		fileI, fileJ := partsI[len(partsI)-1], partsJ[len(partsJ)-1]
		priI, priJ := getFilePriority(fileI), getFilePriority(fileJ)
		if priI != priJ {
			return priI < priJ
		}

		// Same priority, sort by numeric prefix
		numI, numJ := extractNumericPrefix(fileI), extractNumericPrefix(fileJ)
		if numI != numJ {
			return numI < numJ
		}

		// Fallback: alphabetical
		return pathI < pathJ
	})

	return filesCopy
}

// scanMarkersOrdered scans and returns sorted files with markers
func scanMarkersOrdered(festivalPath string) ([]string, *MarkerSummary, error) {
	summary, err := scanMarkers(festivalPath)
	if err != nil {
		return nil, nil, err
	}

	// Get unique files with markers
	filesWithMarkers := make(map[string]bool)
	for _, occ := range summary.Occurrences {
		filesWithMarkers[occ.File] = true
	}

	// Convert to slice
	files := make([]string, 0, len(filesWithMarkers))
	for file := range filesWithMarkers {
		files = append(files, file)
	}

	// Sort files by hierarchy
	sortedFiles := sortMarkerFiles(files)

	return sortedFiles, summary, nil
}

// newNextCommand creates the next subcommand
func newNextCommand(opts *markersOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "next",
		Short: "Show the next file with unfilled markers",
		Long: `Show the next file that needs markers filled, with context hierarchy.

Files are presented in priority order:
1. Festival-level files (FESTIVAL_GOAL.md, FESTIVAL_OVERVIEW.md, TODO.md)
2. Phase-level files (PHASE_GOAL.md for each phase)
3. Sequence-level files (SEQUENCE_GOAL.md for each sequence)
4. Task files (within each sequence)

The context hierarchy is shown to help understand how the file relates to
the overall festival goals.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runNext(opts)
		},
	}
}

// runNext executes the next command
func runNext(opts *markersOptions) error {
	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return err
	}

	// Scan and sort markers
	sortedFiles, summary, err := scanMarkersOrdered(festivalPath)
	if err != nil {
		return err
	}

	// Handle no markers case
	if len(sortedFiles) == 0 {
		if opts.jsonOutput {
			output := map[string]interface{}{
				"complete": true,
				"message":  "All markers have been filled",
				"progress": map[string]int{
					"files_remaining": 0,
					"total_markers":   0,
				},
			}
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(output)
		}
		display := ui.New(shared.IsNoColor(), shared.IsVerbose())
		display.Success("All markers have been filled!")
		fmt.Println("\nNo unfilled template markers found in this festival.")
		return nil
	}

	// Get next file with context
	nextFile, err := getNextFileWithContext(festivalPath, sortedFiles, summary)
	if err != nil {
		return err
	}

	// Output based on format
	if opts.jsonOutput {
		return outputNextJSON(nextFile, summary)
	}

	return outputNextHuman(nextFile, summary)
}

// FileInfo contains file path and position information for JSON output
type FileInfo struct {
	Path       string `json:"path"`
	FullPath   string `json:"full_path"`
	Position   int    `json:"position"`
	TotalFiles int    `json:"total_files"`
}

// ProgressInfo contains progress tracking information for JSON output
type ProgressInfo struct {
	FilesRemaining int `json:"files_remaining"`
	FilesCompleted int `json:"files_completed"`
	TotalMarkers   int `json:"total_markers"`
	MarkersInFile  int `json:"markers_in_file"`
}

// NextFileOutput is the JSON structure for fest markers next --json
type NextFileOutput struct {
	File     FileInfo           `json:"file"`
	Context  *MarkerFileContext `json:"context"`
	Markers  []MarkerOccurrence `json:"markers"`
	Progress ProgressInfo       `json:"progress"`
}

// outputNextJSON outputs the next file information as JSON
func outputNextJSON(file *FileWithContext, summary *MarkerSummary) error {
	output := NextFileOutput{
		File: FileInfo{
			Path:       file.Path,
			FullPath:   file.FullPath,
			Position:   file.Position,
			TotalFiles: file.TotalFiles,
		},
		Context: file.Context,
		Markers: file.Markers,
		Progress: ProgressInfo{
			FilesRemaining: file.TotalFiles,
			FilesCompleted: 0, // Would need tracking to calculate
			TotalMarkers:   summary.TotalMarkers,
			MarkersInFile:  len(file.Markers),
		},
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

// outputNextHuman outputs the next file information in human-readable format
func outputNextHuman(file *FileWithContext, summary *MarkerSummary) error {
	// Header
	fmt.Println(strings.Repeat("=", 80))
	fmt.Printf("NEXT FILE: %s (%d of %d files with markers)\n",
		file.Path, file.Position, file.TotalFiles)
	fmt.Println(strings.Repeat("=", 80))

	// Context hierarchy
	fmt.Println("\nCONTEXT HIERARCHY:")
	if file.Context.Festival != nil {
		fmt.Printf("  Festival: %s\n", file.Context.Festival.Name)
		if file.Context.Festival.Goal != "" {
			fmt.Printf("  Goal: %s\n", file.Context.Festival.Goal)
		}
	}
	if file.Context.Phase != nil {
		fmt.Printf("  Phase: %s\n", file.Context.Phase.Name)
		if file.Context.Phase.Goal != "" {
			fmt.Printf("  Phase Goal: %s\n", file.Context.Phase.Goal)
		}
	}
	if file.Context.Sequence != nil {
		fmt.Printf("  Sequence: %s\n", file.Context.Sequence.Name)
		if file.Context.Sequence.Goal != "" {
			fmt.Printf("  Sequence Goal: %s\n", file.Context.Sequence.Goal)
		}
	}

	// File info
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("FILE: %s\n", file.FullPath)
	fmt.Printf("MARKERS: %d unfilled\n", len(file.Markers))
	fmt.Println(strings.Repeat("-", 80))

	// Markers list
	fmt.Println("\n## Markers to Fill:")
	for _, marker := range file.Markers {
		fmt.Printf("Line %d: %s\n", marker.Line, marker.Content)
	}

	// Progress
	fmt.Println()
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("PROGRESS: %d/%d files remaining | %d total markers\n",
		file.TotalFiles, file.TotalFiles, summary.TotalMarkers)
	fmt.Println(strings.Repeat("-", 80))

	// Hint
	fmt.Println("\nAfter filling markers, run: fest markers next")

	return nil
}
