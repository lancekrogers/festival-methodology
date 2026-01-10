package validation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// Validation issue levels
const (
	LevelError   = "error"
	LevelWarning = "warning"
	LevelInfo    = "info"
)

// Validation issue codes
const (
	CodeMissingFile        = "missing_file"
	CodeMissingTaskFiles   = "missing_task_files"
	CodeMissingQualityGate = "missing_quality_gates"
	CodeNamingConvention   = "naming_convention"
	CodeUnfilledTemplate   = "unfilled_template"
	CodeMissingGoal        = "missing_goal"
	CodeNumberingGap       = "numbering_gap"
)

// ValidationIssue represents a single validation problem
type ValidationIssue struct {
	Level       string `json:"level"`
	Code        string `json:"code"`
	Path        string `json:"path"`
	Message     string `json:"message"`
	Fix         string `json:"fix,omitempty"`
	AutoFixable bool   `json:"auto_fixable"`
}

// Checklist represents post-completion checklist results
type Checklist struct {
	TemplatesFilled *bool `json:"templates_filled"`
	GoalsAchievable *bool `json:"goals_achievable"` // null = manual check required
	TaskFilesExist  *bool `json:"task_files_exist"`
	OrderCorrect    *bool `json:"order_correct"`
	ParallelCorrect *bool `json:"parallel_correct"`
}

// FixApplied represents a fix that was automatically applied
type FixApplied struct {
	Code   string `json:"code"`
	Path   string `json:"path"`
	Action string `json:"action"`
}

// MarkerInfo tracks template marker details
type MarkerInfo struct {
	TotalFiles       int      `json:"total_files"`
	TotalCount       int      `json:"total_count"`
	FilesWithMarkers []string `json:"files,omitempty"`
}

// ValidationResult represents the complete validation result
type ValidationResult struct {
	OK           bool              `json:"ok"`
	Action       string            `json:"action"`
	Festival     string            `json:"festival"`
	Valid        bool              `json:"valid"`
	Score        int               `json:"score"`
	Issues       []ValidationIssue `json:"issues,omitempty"`
	Warnings     []string          `json:"warnings,omitempty"`
	Checklist    *Checklist        `json:"checklist,omitempty"`
	FixesApplied []FixApplied      `json:"fixes_applied,omitempty"`
	Suggestions  []string          `json:"suggestions,omitempty"`
	MarkerInfo   *MarkerInfo       `json:"marker_info,omitempty"`
}

type validateOptions struct {
	path       string
	jsonOutput bool
	fix        bool
}

// NewValidateCommand creates the validate command group
func NewValidateCommand() *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "validate [festival-path]",
		Short: "Check festival structure - find missing task files and issues",
		Long: `Validate that a festival follows the methodology correctly.

Unlike 'fest index validate' which checks index-to-filesystem sync,
this command validates METHODOLOGY COMPLIANCE:

  • Required files exist (FESTIVAL_OVERVIEW.md, goal files)
  • Implementation sequences have TASK FILES (not just goals)
  • Quality gates are present in implementation sequences
  • Naming conventions are followed
  • Templates have been filled out (no [FILL:] markers)

AI agents execute TASK FILES, not goals. If your sequences only have
SEQUENCE_GOAL.md without task files, agents won't know HOW to execute.

Use --fix to automatically apply safe fixes (like adding quality gates).`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateAll(cmd.Context(), opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Automatically apply safe fixes")

	// Add subcommands
	cmd.AddCommand(newValidateStructureCmd(opts))
	cmd.AddCommand(newValidateCompletenessCmd(opts))
	cmd.AddCommand(newValidateTasksCmd(opts))
	cmd.AddCommand(newValidateQualityGatesCmd(opts))
	cmd.AddCommand(newValidateChecklistCmd(opts))
	cmd.AddCommand(newValidateOrderingCmd(opts))

	return cmd
}

// resolveFestivalPath resolves the festival root directory
func resolveFestivalPath(pathArg string) (string, error) {
	if pathArg != "" {
		absPath, err := filepath.Abs(pathArg)
		if err != nil {
			return "", errors.Wrap(err, "resolving path").WithField("path", pathArg)
		}

		// Check if path exists
		if !validateDirExists(absPath) {
			return "", errors.NotFound("path").WithField("path", absPath)
		}

		// Check if it's a festival directory
		if !isFestivalDir(absPath) {
			return "", errors.Validation("path is not a festival directory (expected FESTIVAL_OVERVIEW.md or phase directories like 001_PHASE_NAME)").WithField("path", absPath)
		}

		return absPath, nil
	}

	// Try current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting working directory", err)
	}

	// Check if we're in a festival directory
	if isFestivalDir(cwd) {
		return cwd, nil
	}

	// Try to find festivals/ root and look for active festivals
	root, err := tpl.FindFestivalsRoot(cwd)
	if err == nil {
		// Check if we're inside an active festival
		rel, _ := filepath.Rel(root, cwd)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) >= 2 && parts[0] == "active" {
			festivalPath := filepath.Join(root, parts[0], parts[1])
			if isFestivalDir(festivalPath) {
				return festivalPath, nil
			}
		}
	}

	// Not in a festival - provide helpful error
	return "", errors.Validation("not inside a festival directory - run from inside a festival, or provide a path: fest validate /path/to/festival")
}

// isFestivalDir checks if a directory appears to be a festival root
func isFestivalDir(path string) bool {
	// Check for FESTIVAL_OVERVIEW.md or phases
	if validateFileExists(filepath.Join(path, "FESTIVAL_OVERVIEW.md")) {
		return true
	}

	// Check for phase directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	phasePattern := regexp.MustCompile(`^\d{3}_`)
	for _, entry := range entries {
		if entry.IsDir() && phasePattern.MatchString(entry.Name()) {
			return true
		}
	}

	return false
}

func validateFileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func validateDirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// runValidateAll runs all validation checks
func runValidateAll(ctx context.Context, opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	if ctx == nil {
		ctx = context.Background()
	}

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	// Run all validation checks
	validateStructureChecks(ctx, festivalPath, result)
	validateCompletenessChecks(ctx, festivalPath, result)
	validateTaskFilesChecks(ctx, festivalPath, result)
	validateQualityGatesChecks(ctx, festivalPath, result, opts.fix)
	validateTemplateMarkers(festivalPath, result)
	validateOrderingChecks(ctx, festivalPath, result)

	// Calculate score
	result.Score = calculateScore(result)

	// Add suggestions based on issues
	addSuggestions(result)

	// Determine overall validity
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	// Human-readable output
	printValidationResult(display, festivalPath, result)
	return nil
}

func validateTemplateMarkers(festivalPath string, result *ValidationResult) {
	// Scan for unfilled template markers
	markers := []string{"[FILL:", "[REPLACE:", "[GUIDANCE:", "{{ "}

	markerInfo := &MarkerInfo{
		FilesWithMarkers: []string{},
	}

	filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
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
		// Skip gates/ directory - these are intentional template files
		if strings.HasPrefix(relPath, "gates/") || strings.HasPrefix(relPath, "gates"+string(filepath.Separator)) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		fileMarkerCount := 0
		foundMarkers := make(map[string]bool)

		// Count all markers in this file
		for _, marker := range markers {
			count := strings.Count(contentStr, marker)
			if count > 0 {
				fileMarkerCount += count
				foundMarkers[marker] = true
			}
		}

		if fileMarkerCount > 0 {
			markerInfo.TotalCount += fileMarkerCount
			markerInfo.TotalFiles++
			markerInfo.FilesWithMarkers = append(markerInfo.FilesWithMarkers, relPath)

			// Create a single issue per file listing all marker types found
			markerTypes := []string{}
			for marker := range foundMarkers {
				markerTypes = append(markerTypes, marker)
			}

			result.Issues = append(result.Issues, ValidationIssue{
				Level:   LevelError,
				Code:    CodeUnfilledTemplate,
				Path:    relPath,
				Message: fmt.Sprintf("File contains %d unfilled template markers (%s)", fileMarkerCount, strings.Join(markerTypes, ", ")),
				Fix:     "Edit file and replace template markers with actual content",
			})
		}

		return nil
	})

	// Only set MarkerInfo if markers were found
	if markerInfo.TotalCount > 0 {
		result.MarkerInfo = markerInfo
	}
}

// Score calculation

func calculateScore(result *ValidationResult) int {
	if len(result.Issues) == 0 {
		return 100
	}

	errorCount := 0
	warningCount := 0

	for _, issue := range result.Issues {
		switch issue.Level {
		case LevelError:
			errorCount++
		case LevelWarning:
			warningCount++
		}
	}

	// Base score of 100, minus 15 per error, minus 5 per warning
	score := 100 - (errorCount * 15) - (warningCount * 5)
	if score < 0 {
		score = 0
	}

	return score
}

func addSuggestions(result *ValidationResult) {
	hasMissingTasks := false
	hasMissingGates := false
	hasUnfilledTemplates := false
	hasNumberingGaps := false

	for _, issue := range result.Issues {
		switch issue.Code {
		case CodeMissingTaskFiles:
			hasMissingTasks = true
		case CodeMissingQualityGate:
			hasMissingGates = true
		case CodeUnfilledTemplate:
			hasUnfilledTemplates = true
		case CodeNumberingGap:
			hasNumberingGaps = true
		}
	}

	if hasMissingTasks {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest understand tasks' to learn about task file creation")
	}
	if hasMissingGates {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest gates apply --approve' to add quality gates")
	}
	if hasUnfilledTemplates {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest markers list' to see all unfilled template markers",
			"Use 'fest markers fill' or edit files manually to replace markers with actual content")
	}
	if hasNumberingGaps {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest renumber' to automatically fix numbering gaps")
	}
}

// Output functions

func emitValidateJSON(result *ValidationResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func emitValidateError(opts *validateOptions, err error) error {
	if opts.jsonOutput {
		result := &ValidationResult{
			OK:     false,
			Action: "validate",
			Valid:  false,
			Issues: []ValidationIssue{{
				Level:   LevelError,
				Code:    "error",
				Message: err.Error(),
			}},
		}
		return emitValidateJSON(result)
	}
	return err
}
