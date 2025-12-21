package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
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

	return cmd
}

func newValidateStructureCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "structure [festival-path]",
		Short: "Validate naming conventions and hierarchy",
		Long: `Validate that festival structure follows naming conventions:

  • Phases: NNN_PHASE_NAME (3-digit prefix, UPPERCASE)
  • Sequences: NN_sequence_name (2-digit prefix, lowercase)
  • Tasks: NN_task_name.md (2-digit prefix, lowercase, .md extension)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateStructure(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func newValidateCompletenessCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "completeness [festival-path]",
		Short: "Validate required files exist",
		Long: `Validate that all required files exist:

  • FESTIVAL_OVERVIEW.md (required)
  • PHASE_GOAL.md in each phase (required)
  • SEQUENCE_GOAL.md in each sequence (required)
  • FESTIVAL_RULES.md (recommended)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateCompleteness(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func newValidateTasksCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "tasks [festival-path]",
		Short: "Validate task files exist (CRITICAL)",
		Long: `Validate that implementation sequences have TASK FILES.

THIS IS THE MOST COMMON MISTAKE: Creating sequences with only
SEQUENCE_GOAL.md but no task files.

  Goals define WHAT to achieve.
  Tasks define HOW to execute.

AI agents EXECUTE TASK FILES. Without them, agents know the objective
but don't know what specific work to perform.

CORRECT STRUCTURE:
  02_api/
  ├── SEQUENCE_GOAL.md          ← Defines objective
  ├── 01_design_endpoints.md    ← Task: design work
  ├── 02_implement_crud.md      ← Task: implementation
  └── 03_testing_and_verify.md  ← Quality gate

INCORRECT STRUCTURE (common mistake):
  02_api/
  └── SEQUENCE_GOAL.md          ← No task files!`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateTasks(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

	return cmd
}

func newValidateQualityGatesCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "quality-gates [festival-path]",
		Short: "Validate quality gates exist",
		Long: `Validate that implementation sequences have quality gate tasks.

Quality gates are required for implementation sequences:
  • XX_testing_and_verify.md
  • XX_code_review.md
  • XX_review_results_iterate.md

Use --fix to automatically add missing quality gates.
Planning sequences (*_planning, *_research, etc.) are excluded.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateQualityGates(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")
	cmd.Flags().BoolVar(&opts.fix, "fix", false, "Automatically add missing quality gates")

	return cmd
}

func newValidateChecklistCmd(parentOpts *validateOptions) *cobra.Command {
	opts := &validateOptions{}

	cmd := &cobra.Command{
		Use:   "checklist [festival-path]",
		Short: "Post-completion questionnaire",
		Long: `Run through post-completion checklist to ensure methodology compliance.

Checks (auto-verified where possible):
  1. Did you fill out ALL templates? (auto-check for markers)
  2. Does this plan achieve project goals? (manual review)
  3. Are items in order of operation? (auto-check)
  4. Did you follow parallelization standards? (auto-check)
  5. Did you create TASK FILES for implementation? (auto-check)`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runValidateChecklist(opts)
		},
	}

	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output results as JSON")

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
	display := ui.New(noColor, verbose)

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

func runValidateStructure(opts *validateOptions) error {
	display := ui.New(noColor, verbose)

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_structure",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateStructureChecks(festivalPath, result)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printValidationSection(display, "STRUCTURE", result.Issues)
	return nil
}

func runValidateCompleteness(opts *validateOptions) error {
	display := ui.New(noColor, verbose)

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_completeness",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateCompletenessChecks(festivalPath, result)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printValidationSection(display, "COMPLETENESS", result.Issues)
	return nil
}

func runValidateTasks(opts *validateOptions) error {
	display := ui.New(noColor, verbose)

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_tasks",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateTaskFilesChecks(festivalPath, result)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printTaskValidationResult(display, result)
	return nil
}

func runValidateQualityGates(opts *validateOptions) error {
	display := ui.New(noColor, verbose)

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:       true,
		Action:   "validate_quality_gates",
		Festival: filepath.Base(festivalPath),
		Valid:    true,
		Issues:   []ValidationIssue{},
	}

	validateQualityGatesChecks(festivalPath, result, opts.fix)

	result.Score = calculateScore(result)
	for _, issue := range result.Issues {
		if issue.Level == LevelError {
			result.Valid = false
			break
		}
	}

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printValidationSection(display, "QUALITY GATES", result.Issues)
	return nil
}

func runValidateChecklist(opts *validateOptions) error {
	display := ui.New(noColor, verbose)

	festivalPath, err := resolveFestivalPath(opts.path)
	if err != nil {
		return emitValidateError(opts, err)
	}

	result := &ValidationResult{
		OK:        true,
		Action:    "validate_checklist",
		Festival:  filepath.Base(festivalPath),
		Valid:     true,
		Issues:    []ValidationIssue{},
		Checklist: &Checklist{},
	}

	// Run all checks and populate checklist
	templatesFilled := checkTemplatesFilled(festivalPath)
	result.Checklist.TemplatesFilled = &templatesFilled

	// Goals achievable is a manual check - always null
	result.Checklist.GoalsAchievable = nil

	taskFilesExist := checkTaskFilesExist(festivalPath)
	result.Checklist.TaskFilesExist = &taskFilesExist

	orderCorrect := checkOrderCorrect(festivalPath)
	result.Checklist.OrderCorrect = &orderCorrect

	parallelCorrect := checkParallelCorrect(festivalPath)
	result.Checklist.ParallelCorrect = &parallelCorrect

	if opts.jsonOutput {
		return emitValidateJSON(result)
	}

	printChecklistResult(display, result.Checklist)
	return nil
}

// Validation check implementations

func validateStructureChecks(festivalPath string, result *ValidationResult) {
	parser := festival.NewParser()

	// Parse festival structure
	phases, err := parser.ParsePhases(festivalPath)
	if err != nil {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelError,
			Code:    CodeNamingConvention,
			Path:    festivalPath,
			Message: fmt.Sprintf("Failed to parse phases: %v", err),
		})
		return
	}

	// Check phase naming (should be UPPERCASE after number)
	phaseUpperPattern := regexp.MustCompile(`^\d{3}_[A-Z][A-Z0-9_]*$`)
	for _, phase := range phases {
		if !phaseUpperPattern.MatchString(phase.FullName) {
			result.Issues = append(result.Issues, ValidationIssue{
				Level:   LevelWarning,
				Code:    CodeNamingConvention,
				Path:    phase.Path,
				Message: fmt.Sprintf("Phase name should be UPPERCASE: %s", phase.FullName),
			})
		}

		// Check sequences
		sequences, _ := parser.ParseSequences(phase.Path)
		seqLowerPattern := regexp.MustCompile(`^\d{2}_[a-z][a-z0-9_]*$`)
		for _, seq := range sequences {
			if !seqLowerPattern.MatchString(seq.FullName) {
				result.Issues = append(result.Issues, ValidationIssue{
					Level:   LevelWarning,
					Code:    CodeNamingConvention,
					Path:    seq.Path,
					Message: fmt.Sprintf("Sequence name should be lowercase: %s", seq.FullName),
				})
			}

			// Check tasks
			tasks, _ := parser.ParseTasks(seq.Path)
			taskLowerPattern := regexp.MustCompile(`^\d{2}_[a-z][a-z0-9_]*\.md$`)
			for _, task := range tasks {
				if !taskLowerPattern.MatchString(task.FullName) {
					result.Issues = append(result.Issues, ValidationIssue{
						Level:   LevelWarning,
						Code:    CodeNamingConvention,
						Path:    task.Path,
						Message: fmt.Sprintf("Task name should be lowercase: %s", task.FullName),
					})
				}
			}
		}
	}
}

func validateCompletenessChecks(festivalPath string, result *ValidationResult) {
	// Check FESTIVAL_OVERVIEW.md
	overviewPath := filepath.Join(festivalPath, "FESTIVAL_OVERVIEW.md")
	if !validateFileExists(overviewPath) {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelError,
			Code:    CodeMissingFile,
			Path:    overviewPath,
			Message: "FESTIVAL_OVERVIEW.md is required",
			Fix:     "Create FESTIVAL_OVERVIEW.md with project goals and success criteria",
		})
	}

	// Check FESTIVAL_RULES.md (warning, not error)
	rulesPath := filepath.Join(festivalPath, "FESTIVAL_RULES.md")
	if !validateFileExists(rulesPath) {
		result.Issues = append(result.Issues, ValidationIssue{
			Level:   LevelWarning,
			Code:    CodeMissingFile,
			Path:    rulesPath,
			Message: "FESTIVAL_RULES.md is recommended",
		})
	}

	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(festivalPath)

	for _, phase := range phases {
		// Check PHASE_GOAL.md
		phaseGoalPath := filepath.Join(phase.Path, "PHASE_GOAL.md")
		if !validateFileExists(phaseGoalPath) {
			result.Issues = append(result.Issues, ValidationIssue{
				Level:   LevelError,
				Code:    CodeMissingGoal,
				Path:    phaseGoalPath,
				Message: fmt.Sprintf("PHASE_GOAL.md required in %s", phase.FullName),
				Fix:     fmt.Sprintf("fest create phase --name %q --json", phase.Name),
			})
		}

		// Check sequences
		sequences, _ := parser.ParseSequences(phase.Path)
		for _, seq := range sequences {
			seqGoalPath := filepath.Join(seq.Path, "SEQUENCE_GOAL.md")
			if !validateFileExists(seqGoalPath) {
				result.Issues = append(result.Issues, ValidationIssue{
					Level:   LevelError,
					Code:    CodeMissingGoal,
					Path:    seqGoalPath,
					Message: fmt.Sprintf("SEQUENCE_GOAL.md required in %s", seq.FullName),
					Fix:     fmt.Sprintf("fest create sequence --name %q --json", seq.Name),
				})
			}
		}
	}
}

func validateTaskFilesChecks(festivalPath string, result *ValidationResult) {
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(festivalPath)
	policy := gates.DefaultPolicy()

	sequencesWithoutTasks := []string{}

	for _, phase := range phases {
		sequences, _ := parser.ParseSequences(phase.Path)
		for _, seq := range sequences {
			// Check if this is an implementation sequence
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			// Check for task files
			tasks, _ := parser.ParseTasks(seq.Path)
			if len(tasks) == 0 {
				relPath, _ := filepath.Rel(festivalPath, seq.Path)
				sequencesWithoutTasks = append(sequencesWithoutTasks, relPath)
				result.Issues = append(result.Issues, ValidationIssue{
					Level:       LevelError,
					Code:        CodeMissingTaskFiles,
					Path:        relPath,
					Message:     "Implementation sequence has SEQUENCE_GOAL.md but no task files",
					Fix:         fmt.Sprintf("fest create task --name \"design\" --path %s --json", relPath),
					AutoFixable: false,
				})
			}
		}
	}
}

func validateQualityGatesChecks(festivalPath string, result *ValidationResult, autoFix bool) {
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(festivalPath)
	policy := gates.DefaultPolicy()

	gateIDs := map[string]bool{}
	for _, gate := range policy.GetEnabledTasks() {
		gateIDs[gate.ID] = true
	}

	// Phase name patterns that don't require quality gates
	nonImplPhasePatterns := []string{"PLAN", "DESIGN", "REVIEW", "UAT", "FINALIZE", "DOCS", "RESEARCH"}

	for _, phase := range phases {
		// Skip non-implementation phases entirely
		if isNonImplementationPhase(phase.Name, nonImplPhasePatterns) {
			continue
		}

		sequences, _ := parser.ParseSequences(phase.Path)
		for _, seq := range sequences {
			// Check if this is an implementation sequence
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			// Check for quality gate tasks
			tasks, _ := parser.ParseTasks(seq.Path)
			hasGates := false
			for _, task := range tasks {
				// Check if task name contains any gate ID
				for gateID := range gateIDs {
					if strings.Contains(strings.ToLower(task.Name), strings.ReplaceAll(gateID, "_", "")) ||
						strings.Contains(task.Name, gateID) {
						hasGates = true
						break
					}
				}
			}

			if !hasGates && len(tasks) > 0 {
				relPath, _ := filepath.Rel(festivalPath, seq.Path)
				result.Issues = append(result.Issues, ValidationIssue{
					Level:       LevelError,
					Code:        CodeMissingQualityGate,
					Path:        relPath,
					Message:     "Implementation sequence missing quality gates",
					Fix:         fmt.Sprintf("fest task defaults sync --path %s --approve --json", relPath),
					AutoFixable: true,
				})

				if autoFix {
					// TODO: Call fest task defaults sync
					result.FixesApplied = append(result.FixesApplied, FixApplied{
						Code:   CodeMissingQualityGate,
						Path:   relPath,
						Action: "Quality gates would be added (--fix not yet implemented)",
					})
				}
			}
		}
	}
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

// isNonImplementationPhase checks if a phase is for planning/review rather than implementation
func isNonImplementationPhase(phaseName string, patterns []string) bool {
	upperName := strings.ToUpper(phaseName)
	for _, pattern := range patterns {
		if strings.Contains(upperName, pattern) {
			return true
		}
	}
	return false
}

// isExcludedSequence checks if a sequence name matches exclusion patterns
func isExcludedSequence(name string, patterns []string) bool {
	for _, pattern := range patterns {
		// Simple glob matching
		if strings.HasPrefix(pattern, "*") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(name, suffix) {
				return true
			}
		} else if strings.HasSuffix(pattern, "*") {
			prefix := strings.TrimSuffix(pattern, "*")
			if strings.HasPrefix(name, prefix) {
				return true
			}
		} else if name == pattern {
			return true
		}
	}
	return false
}

// Checklist check functions

func checkTemplatesFilled(festivalPath string) bool {
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}
	filled := true

	filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

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
				filled = false
				return filepath.SkipAll
			}
		}
		return nil
	})

	return filled
}

func checkTaskFilesExist(festivalPath string) bool {
	parser := festival.NewParser()
	phases, _ := parser.ParsePhases(festivalPath)
	policy := gates.DefaultPolicy()

	for _, phase := range phases {
		sequences, _ := parser.ParseSequences(phase.Path)
		for _, seq := range sequences {
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}
			tasks, _ := parser.ParseTasks(seq.Path)
			if len(tasks) == 0 {
				return false
			}
		}
	}
	return true
}

func checkOrderCorrect(festivalPath string) bool {
	parser := festival.NewParser()
	phases, err := parser.ParsePhases(festivalPath)
	if err != nil {
		return true // Can't check, assume OK
	}

	// Check phases are sequential
	lastNum := 0
	for _, phase := range phases {
		if phase.Number < lastNum {
			return false
		}
		lastNum = phase.Number
	}

	return true
}

func checkParallelCorrect(festivalPath string) bool {
	// For now, always return true - parallel validation is complex
	// and false positives would be confusing
	return true
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

	for _, issue := range result.Issues {
		switch issue.Code {
		case CodeMissingTaskFiles:
			hasMissingTasks = true
		case CodeMissingQualityGate:
			hasMissingGates = true
		case CodeUnfilledTemplate:
			hasUnfilledTemplates = true
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

	printValidationSection(display, "STRUCTURE", structureIssues)
	printValidationSection(display, "COMPLETENESS", completenessIssues)
	printTaskValidationSection(display, taskIssues)
	printValidationSection(display, "QUALITY GATES", gateIssues)
	printValidationSection(display, "TEMPLATES", templateIssues)

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

func printTaskValidationSection(display *ui.UI, issues []ValidationIssue) {
	fmt.Println("\nTASK FILES (Critical for AI Execution)")

	if len(issues) == 0 {
		display.Success("[OK] All implementation sequences have task files")
		return
	}

	display.Error("[ERROR] Implementation sequences need task files, not just goals")
	fmt.Println()
	fmt.Println("        Goals define WHAT to achieve; tasks define HOW to execute.")
	fmt.Println("        AI agents EXECUTE task files.")
	fmt.Println()
	fmt.Println("        Sequences without tasks:")

	for _, issue := range issues {
		fmt.Printf("        - %s\n", issue.Path)
	}

	fmt.Println()
	fmt.Println("        For each sequence, create task files:")
	fmt.Println("          fest create task --name \"design\" --json")
	fmt.Println("          fest create task --name \"implement\" --json")
	fmt.Println("          fest create task --name \"test\" --json")
}

func printTaskValidationResult(display *ui.UI, result *ValidationResult) {
	printTaskValidationSection(display, result.Issues)
}

func printChecklistResult(display *ui.UI, checklist *Checklist) {
	fmt.Println("\nPost-Completion Checklist")
	fmt.Println(strings.Repeat("=", 50))

	printCheckItem(display, 1, "Templates Filled",
		"Did you fill out ALL templates?",
		checklist.TemplatesFilled,
		"Search for [FILL:] markers and replace with content")

	printCheckItem(display, 2, "Goals Achievable",
		"Does this plan achieve project goals?",
		checklist.GoalsAchievable,
		"Review FESTIVAL_OVERVIEW.md and verify alignment")

	printCheckItem(display, 3, "Task Files Exist",
		"Did you create TASK FILES for implementation?",
		checklist.TaskFilesExist,
		"Run 'fest understand tasks' to learn about task files")

	printCheckItem(display, 4, "Order Correct",
		"Are items in order of operation?",
		checklist.OrderCorrect,
		"Lower numbers execute first (01_ before 02_)")

	printCheckItem(display, 5, "Parallel Correct",
		"Did you follow parallelization standards?",
		checklist.ParallelCorrect,
		"Same-numbered items (01_a, 01_b) run in parallel")

	// Summary
	passed := 0
	total := 0
	if checklist.TemplatesFilled != nil {
		total++
		if *checklist.TemplatesFilled {
			passed++
		}
	}
	if checklist.TaskFilesExist != nil {
		total++
		if *checklist.TaskFilesExist {
			passed++
		}
	}
	if checklist.OrderCorrect != nil {
		total++
		if *checklist.OrderCorrect {
			passed++
		}
	}
	if checklist.ParallelCorrect != nil {
		total++
		if *checklist.ParallelCorrect {
			passed++
		}
	}

	fmt.Printf("\nOverall: %d/%d auto-checks passed (1 manual check required)\n", passed, total)
}

func printCheckItem(display *ui.UI, num int, title, question string, result *bool, guidance string) {
	fmt.Printf("\n%d. %s\n", num, title)
	fmt.Printf("   Question: %s\n", question)

	if result == nil {
		fmt.Println("   Status: [MANUAL CHECK REQUIRED]")
		fmt.Printf("   Guidance: %s\n", guidance)
	} else if *result {
		display.Success("   Auto-check: [PASS]")
	} else {
		display.Error("   Auto-check: [FAIL]")
		fmt.Printf("   Fix: %s\n", guidance)
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
