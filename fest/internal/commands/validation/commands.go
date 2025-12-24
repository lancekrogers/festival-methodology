package validation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
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
		Short: "Validate festival methodology compliance",
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
			return runValidateAll(opts)
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
			return "", fmt.Errorf("invalid path: %w", err)
		}

		// Check if path exists
		if !validateDirExists(absPath) {
			return "", fmt.Errorf("path does not exist: %s", absPath)
		}

		// Check if it's a festival directory
		if !isFestivalDir(absPath) {
			return "", fmt.Errorf("path is not a festival directory: %s\n  (expected FESTIVAL_OVERVIEW.md or phase directories like 001_PHASE_NAME)", absPath)
		}

		return absPath, nil
	}

	// Try current directory
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
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
	return "", fmt.Errorf("not inside a festival directory\n  Run from inside a festival, or provide a path: fest validate /path/to/festival")
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
func runValidateAll(opts *validateOptions) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

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
	validateStructureChecks(festivalPath, result)
	validateCompletenessChecks(festivalPath, result)
	validateTaskFilesChecks(festivalPath, result)
	validateQualityGatesChecks(festivalPath, result, opts.fix)
	validateTemplateMarkers(festivalPath, result)
	validateOrderingChecks(festivalPath, result)

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
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}

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

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		contentStr := string(content)
		for _, marker := range markers {
			if strings.Contains(contentStr, marker) {
				result.Issues = append(result.Issues, ValidationIssue{
					Level:   LevelWarning,
					Code:    CodeUnfilledTemplate,
					Path:    relPath,
					Message: fmt.Sprintf("File contains unfilled template marker: %s", marker),
					Fix:     "Edit file and replace template markers with actual content",
				})
				break
			}
		}

		return nil
	})
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
			"Run 'fest task defaults sync --approve' to add quality gates")
	}
	if hasUnfilledTemplates {
		result.Suggestions = append(result.Suggestions,
			"Edit files with [FILL:] markers and replace with actual content")
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

func printValidationResult(display *ui.UI, festivalPath string, result *ValidationResult) {
	fmt.Printf("\nFestival Validation: %s\n", result.Festival)
	fmt.Println(strings.Repeat("=", 50))

	// Group issues by category
	structureIssues := filterIssuesByCode(result.Issues, CodeNamingConvention)
	completenessIssues := filterIssuesByCode(result.Issues, CodeMissingFile, CodeMissingGoal)
	taskIssues := filterIssuesByCode(result.Issues, CodeMissingTaskFiles)
	gateIssues := filterIssuesByCode(result.Issues, CodeMissingQualityGate)
	templateIssues := filterIssuesByCode(result.Issues, CodeUnfilledTemplate)
	orderingIssues := filterIssuesByCode(result.Issues, CodeNumberingGap)

	printValidationSection(display, "STRUCTURE", structureIssues)
	printValidationSection(display, "COMPLETENESS", completenessIssues)
	printTaskValidationSection(display, taskIssues)
	printValidationSection(display, "QUALITY GATES", gateIssues)
	printValidationSection(display, "TEMPLATES", templateIssues)
	printValidationSection(display, "ORDERING", orderingIssues)

	// Score and summary
	fmt.Printf("\nScore: %d/100", result.Score)
	if result.Valid {
		fmt.Println(" - Festival structure is valid")
	} else {
		fmt.Println(" - Festival structure needs attention")
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Println("\nSuggestions:")
		for _, s := range result.Suggestions {
			fmt.Printf("  • %s\n", s)
		}
	}
}

func printValidationSection(display *ui.UI, title string, issues []ValidationIssue) {
	fmt.Printf("\n%s\n", title)

	if len(issues) == 0 {
		display.Success("[OK] All checks passed")
		return
	}

	for _, issue := range issues {
		switch issue.Level {
		case LevelError:
			display.Error("[ERROR] %s", issue.Message)
		case LevelWarning:
			display.Warning("[WARN] %s", issue.Message)
		case LevelInfo:
			display.Info("[INFO] %s", issue.Message)
		}
		if issue.Path != "" {
			fmt.Printf("        Path: %s\n", issue.Path)
		}
		if issue.Fix != "" {
			fmt.Printf("        FIX: %s\n", issue.Fix)
		}
	}
}

func filterIssuesByCode(issues []ValidationIssue, codes ...string) []ValidationIssue {
	codeSet := make(map[string]bool)
	for _, c := range codes {
		codeSet[c] = true
	}

	var filtered []ValidationIssue
	for _, issue := range issues {
		if codeSet[issue.Code] {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// countFileMarkers counts unfilled template markers in a file
func countFileMarkers(path string) int {
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}
	count := 0

	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, marker := range markers {
			count += strings.Count(line, marker)
		}
	}

	return count
}
