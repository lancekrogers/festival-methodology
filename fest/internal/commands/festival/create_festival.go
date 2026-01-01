package festival

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// Embedded default gate template content - used as fallback when templates don't exist
var defaultGateTemplates = map[string]string{
	"QUALITY_GATE_TESTING.md": `# Task: Testing and Verification

**Task Number:** 01 | **Parallel Group:** None | **Dependencies:** All implementation tasks | **Autonomy:** medium

## Objective

Verify all functionality implemented in this sequence works correctly through comprehensive testing.

## Requirements

- [ ] All unit tests pass
- [ ] Integration tests verify main workflows
- [ ] Manual testing confirms user stories work as expected
- [ ] Error cases are handled correctly
- [ ] Edge cases are addressed

## Test Categories

### Unit Tests

[REPLACE: Run your project's test command]

**Verify:**

- [ ] All new/modified code has test coverage
- [ ] Tests are meaningful (not just coverage padding)
- [ ] Test names describe what they verify

### Integration Tests

[REPLACE: Run your project's integration test command]

**Verify:**

- [ ] Components work together correctly
- [ ] External integrations function properly
- [ ] Data flows correctly through the system

### Manual Verification

Walk through each requirement from the sequence:

1. [ ] **Requirement 1**: [Describe manual test steps and expected result]
2. [ ] **Requirement 2**: [Describe manual test steps and expected result]
3. [ ] **Requirement 3**: [Describe manual test steps and expected result]

## Coverage Requirements

- Minimum coverage: [REPLACE: coverage threshold, e.g., 80%] for new code

[REPLACE: Run your project's coverage command]

## Error Handling Verification

- [ ] Invalid inputs are rejected gracefully
- [ ] Error messages are clear and actionable
- [ ] Errors don't expose sensitive information
- [ ] Recovery paths work correctly

## Definition of Done

- [ ] All automated tests pass
- [ ] Manual verification complete
- [ ] Coverage meets requirements
- [ ] Error handling verified
- [ ] No regressions introduced

## Notes

Document any test gaps, flaky tests, or areas needing future attention here.

---

**Test Results Summary:**

- Unit tests: [ ] Pass / [ ] Fail
- Integration tests: [ ] Pass / [ ] Fail
- Manual tests: [ ] Pass / [ ] Fail
- Coverage: ____%
`,
	"QUALITY_GATE_REVIEW.md": `# Task: Code Review

**Task Number:** 02 | **Parallel Group:** None | **Dependencies:** Testing and Verification | **Autonomy:** low

## Objective

Review all code changes in this sequence for quality, correctness, and adherence to project standards.

## Review Checklist

### Code Quality

- [ ] Code is readable and well-organized
- [ ] Functions/methods are focused (single responsibility)
- [ ] No unnecessary complexity
- [ ] Naming is clear and consistent
- [ ] Comments explain "why" not "what"

### Architecture & Design

- [ ] Changes align with project architecture
- [ ] No unnecessary coupling introduced
- [ ] Dependencies are appropriate
- [ ] Interfaces are clean and focused
- [ ] No code duplication

### Standards Compliance

[REPLACE: Run your project's lint command]

- [ ] Linting passes without warnings
- [ ] Formatting is consistent
- [ ] Project conventions are followed

### Error Handling

- [ ] Errors are handled appropriately
- [ ] Error messages are helpful
- [ ] No panic/crash scenarios
- [ ] Resources are properly cleaned up

### Security Considerations

- [ ] No secrets in code
- [ ] Input validation present
- [ ] No SQL injection risks
- [ ] No XSS vulnerabilities
- [ ] Proper authentication/authorization

### Performance

- [ ] No obvious performance issues
- [ ] Queries are efficient
- [ ] No memory leaks
- [ ] Appropriate caching used

### Testing

- [ ] Tests are meaningful
- [ ] Edge cases covered
- [ ] Test data is appropriate
- [ ] Mocks used correctly

## Review Process

1. **Read the sequence goal** - Understand what was being built
2. **Review file by file** - Check each modified file
3. **Run the code** - Verify functionality works
4. **Document findings** - Note issues and suggestions

## Findings

### Critical Issues (Must Fix)

1. [ ] [Issue description and recommendation]

### Suggestions (Should Consider)

1. [ ] [Suggestion and rationale]

### Positive Observations

- [Note good patterns or practices observed]

## Definition of Done

- [ ] All files reviewed
- [ ] Linting passes
- [ ] No critical issues remaining
- [ ] Suggestions documented
- [ ] Knowledge shared with team (if applicable)

## Review Summary

**Reviewer:** [Name/Agent]
**Date:** [Date]
**Verdict:** [ ] Approved / [ ] Needs Changes

**Notes:**
[Summary of the review and any outstanding concerns]
`,
	"QUALITY_GATE_ITERATE.md": `# Task: Review Results and Iterate

**Task Number:** 03 | **Parallel Group:** None | **Dependencies:** Code Review | **Autonomy:** medium

## Objective

Address all findings from code review and testing, iterate until the sequence meets quality standards.

## Review Findings to Address

### From Testing

| Finding | Priority | Status | Notes |
|---------|----------|--------|-------|
| [Finding 1] | [High/Medium/Low] | [ ] Fixed | |
| [Finding 2] | [High/Medium/Low] | [ ] Fixed | |

### From Code Review

| Finding | Priority | Status | Notes |
|---------|----------|--------|-------|
| [Finding 1] | [High/Medium/Low] | [ ] Fixed | |
| [Finding 2] | [High/Medium/Low] | [ ] Fixed | |

## Iteration Process

### Round 1

**Changes Made:**

- [ ] [Change 1 description]
- [ ] [Change 2 description]

**Verification:**

- [ ] Tests re-run and pass
- [ ] Linting passes
- [ ] Changes reviewed

### Round 2 (if needed)

**Changes Made:**

- [ ] [Change 1 description]

**Verification:**

- [ ] Tests re-run and pass
- [ ] Linting passes
- [ ] Changes reviewed

## Final Verification

After all iterations:

- [ ] All critical findings addressed
- [ ] All tests pass
- [ ] Linting passes
- [ ] Code review approved
- [ ] Sequence objectives met

## Lessons Learned

Document patterns or issues to avoid in future sequences:

### What Went Well

- [Positive observation]

### What Could Improve

- [Area for improvement]

### Process Improvements

- [Suggestion for future work]

## Definition of Done

- [ ] All critical findings fixed
- [ ] All tests pass
- [ ] Linting passes
- [ ] Code review approval received
- [ ] Lessons learned documented
- [ ] Ready to proceed to next sequence

## Sign-Off

**Sequence Complete:** [ ] Yes / [ ] No

**Final Status:**

- Tests: [ ] All Pass
- Review: [ ] Approved
- Quality: [ ] Meets Standards

**Notes:**
[Any final notes or observations about this sequence]

---

**Next Steps:**
[Identify what follows - next sequence, phase completion, etc.]
`,
}

// CreateFestivalOptions holds options for the create festival command.
type CreateFestivalOptions struct {
	Name        string
	Goal        string
	Tags        string
	VarsFile    string
	Markers     string // Inline JSON with hint→value mappings
	MarkersFile string // JSON file path with hint→value mappings
	SkipMarkers bool   // Skip marker processing
	DryRun      bool   // Show markers without creating file
	JSONOutput  bool
	Dest        string // "active" or "planned"
	AgentMode   bool   // Strict mode for AI agents
}

type createFestivalResult struct {
	OK             bool                     `json:"ok"`
	Action         string                   `json:"action"`
	Festival       map[string]string        `json:"festival,omitempty"`
	Created        []string                 `json:"created,omitempty"`
	GatesDirectory string                   `json:"gates_directory,omitempty"`
	FestYAML       string                   `json:"fest_yaml,omitempty"`
	GateTemplates  []string                 `json:"gate_templates,omitempty"`
	Markers        []map[string]interface{} `json:"markers,omitempty"`
	MarkersFilled  int                      `json:"markers_filled,omitempty"`
	MarkersTotal   int                      `json:"markers_total,omitempty"`
	Validation     *ValidationSummary       `json:"validation,omitempty"`
	Errors         []map[string]any         `json:"errors,omitempty"`
	Warnings       []string                 `json:"warnings,omitempty"`
	Extra          map[string]interface{}   `json:"extra,omitempty"`
}

// NewCreateFestivalCommand adds 'create festival'
func NewCreateFestivalCommand() *cobra.Command {
	opts := &CreateFestivalOptions{}
	cmd := &cobra.Command{
		Use:   "festival",
		Short: "Create a new festival scaffold under festivals/(active|planned)",
		RunE: func(cmd *cobra.Command, args []string) error {
			// If no flags were provided, open TUI for this flow
			if cmd.Flags().NFlag() == 0 {
				return shared.StartCreateFestivalTUI(cmd.Context())
			}
			// Otherwise, require name and proceed
			if strings.TrimSpace(opts.Name) == "" {
				return errors.Validation("--name is required (or run without flags to open TUI)")
			}
			return RunCreateFestival(cmd.Context(), opts)
		},
	}
	cmd.Flags().StringVar(&opts.Name, "name", "", "Festival name (required)")
	cmd.Flags().StringVar(&opts.Goal, "goal", "", "Festival goal")
	cmd.Flags().StringVar(&opts.Tags, "tags", "", "Comma-separated tags")
	cmd.Flags().StringVar(&opts.VarsFile, "vars-file", "", "JSON file with variables")
	cmd.Flags().StringVar(&opts.Markers, "markers", "", "JSON string with REPLACE marker hint→value mappings")
	cmd.Flags().StringVar(&opts.MarkersFile, "markers-file", "", "JSON file with REPLACE marker hint→value mappings")
	cmd.Flags().BoolVar(&opts.SkipMarkers, "skip-markers", false, "Skip REPLACE marker processing")
	cmd.Flags().BoolVar(&opts.DryRun, "dry-run", false, "Show template markers without creating file")
	cmd.Flags().BoolVar(&opts.JSONOutput, "json", false, "Emit JSON output")
	cmd.Flags().StringVar(&opts.Dest, "dest", "active", "Destination under festivals/: active or planned")
	cmd.Flags().BoolVar(&opts.AgentMode, "agent", false, "Strict mode: require markers, auto-validate, block on errors, JSON output")
	return cmd
}

// RunCreateFestival executes the create festival command logic.
func RunCreateFestival(ctx context.Context, opts *CreateFestivalOptions) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("RunCreateFestival")
	}

	// Agent mode implies JSON output
	if opts.AgentMode {
		opts.JSONOutput = true
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Resolve festivals root and template root
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}

	// Load effective agent config (workspace config only for new festival)
	agentCfg := LoadEffectiveAgentConfig(festivalsRoot, "")

	// Determine effective skip-markers behavior
	effectiveSkipMarkers := config.EffectiveSkipMarkers(agentCfg, opts.AgentMode, opts.SkipMarkers)

	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return emitCreateFestivalError(opts, err)
	}

	// Load vars
	vars := map[string]interface{}{}
	if strings.TrimSpace(opts.VarsFile) != "" {
		v, err := loadVarsFile(opts.VarsFile)
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "reading vars-file").WithField("path", opts.VarsFile))
		}
		vars = v
	}

	// Build template context
	tmplCtx := tpl.NewContext()
	tmplCtx.SetFestival(opts.Name, opts.Goal, parseTags(opts.Tags))
	for k, v := range vars {
		tmplCtx.SetCustom(k, v)
	}

	// Destination
	slug := Slugify(opts.Name)
	destCategory := strings.ToLower(strings.TrimSpace(opts.Dest))
	if destCategory != "planned" && destCategory != "active" {
		destCategory = "active"
	}

	// Generate unique festival ID
	festivalID, err := id.GenerateID(opts.Name, festivalsRoot)
	if err != nil {
		return emitCreateFestivalError(opts, errors.Wrap(err, "generating festival ID").WithField("name", opts.Name))
	}

	// Create directory with ID suffix: {slug}_{ID}
	dirName := fmt.Sprintf("%s_%s", slug, festivalID)
	destDir := filepath.Join(festivalsRoot, destCategory, dirName)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return emitCreateFestivalError(opts, errors.IO("creating festival directory", err).WithField("path", destDir))
	}

	// Render/copy core files
	mgr := tpl.NewManager()
	created := []string{}

	core := []struct{ Template, Out string }{
		{"FESTIVAL_OVERVIEW_TEMPLATE.md", "FESTIVAL_OVERVIEW.md"},
		{"FESTIVAL_GOAL_TEMPLATE.md", "FESTIVAL_GOAL.md"},
		{"FESTIVAL_RULES_TEMPLATE.md", "FESTIVAL_RULES.md"},
		{"FESTIVAL_TODO_TEMPLATE.md", "TODO.md"},
	}

	for _, c := range core {
		tpath := filepath.Join(tmplRoot, c.Template)
		if _, err := os.Stat(tpath); err != nil {
			// Skip missing template silently; report warning via non-JSON path
			continue
		}
		// Load and decide copy vs render
		loader := tpl.NewLoader()
		t, err := loader.Load(ctx, tpath)
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "loading template").WithField("template", c.Template))
		}
		outPath := filepath.Join(destDir, c.Out)
		// If template appears to require variables, render; else copy
		requires := t.Metadata != nil && len(t.Metadata.RequiredVariables) > 0
		if requires || strings.Contains(t.Content, "{{") {
			out, err := mgr.Render(t, tmplCtx)
			if err != nil {
				return emitCreateFestivalError(opts, errors.Wrap(err, "rendering template").WithField("template", c.Template))
			}
			if err := os.WriteFile(outPath, []byte(out), 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing file", err).WithField("path", outPath))
			}
		} else {
			if err := os.WriteFile(outPath, []byte(t.Content), 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing file", err).WithField("path", outPath))
			}
		}
		created = append(created, outPath)
	}

	// Create gates directory and copy gate templates
	gatesDir := filepath.Join(destDir, "gates")
	if err := os.MkdirAll(gatesDir, 0755); err != nil {
		return emitCreateFestivalError(opts, errors.IO("creating gates directory", err).WithField("path", gatesDir))
	}

	gateTemplates := []string{
		"QUALITY_GATE_TESTING.md",
		"QUALITY_GATE_REVIEW.md",
		"QUALITY_GATE_ITERATE.md",
	}

	copiedGates := []string{}
	for _, gt := range gateTemplates {
		srcPath := filepath.Join(tmplRoot, gt)
		if _, err := os.Stat(srcPath); err != nil {
			// Skip if template doesn't exist - will try fallback below
			continue
		}
		content, err := os.ReadFile(srcPath)
		if err != nil {
			return emitCreateFestivalError(opts, errors.IO("reading gate template", err).WithField("path", srcPath))
		}
		outPath := filepath.Join(gatesDir, gt)
		if err := os.WriteFile(outPath, content, 0644); err != nil {
			return emitCreateFestivalError(opts, errors.IO("writing gate template", err).WithField("path", outPath))
		}
		copiedGates = append(copiedGates, outPath)
		created = append(created, outPath)
	}

	// Fallback: Use embedded templates if no templates were copied from file system
	if len(copiedGates) == 0 {
		for filename, content := range defaultGateTemplates {
			outPath := filepath.Join(gatesDir, filename)
			if err := os.WriteFile(outPath, []byte(content), 0644); err != nil {
				return emitCreateFestivalError(opts, errors.IO("writing embedded gate template", err).WithField("path", outPath))
			}
			copiedGates = append(copiedGates, outPath)
			created = append(created, outPath)
		}
	}

	// Generate fest.yaml with default gates configuration and metadata
	festConfig := defaultFestivalGatesConfig()

	// Populate metadata section
	now := time.Now().UTC()
	festConfig.Metadata = config.FestivalMetadata{
		ID:        festivalID,
		UUID:      uuid.New().String(),
		Name:      opts.Name,
		CreatedAt: now,
		StatusHistory: []config.StatusChange{
			{
				Status:    destCategory,
				Timestamp: now,
				Path:      destDir,
				Notes:     "Festival created",
			},
		},
	}

	festConfigPath := filepath.Join(destDir, config.FestivalConfigFileName)
	if err := config.SaveFestivalConfig(destDir, festConfig); err != nil {
		return emitCreateFestivalError(opts, errors.Wrap(err, "writing fest.yaml").WithField("path", festConfigPath))
	}
	created = append(created, festConfigPath)

	// Process REPLACE markers in all created files
	var totalMarkersFilled, totalMarkersCount int
	var allMarkers []map[string]interface{}

	for _, filePath := range created {
		markerResult, err := ProcessMarkers(ctx, MarkerOptions{
			FilePath:    filePath,
			Markers:     opts.Markers,
			MarkersFile: opts.MarkersFile,
			SkipMarkers: effectiveSkipMarkers,
			DryRun:      opts.DryRun,
			JSONOutput:  opts.JSONOutput,
		})
		if err != nil {
			return emitCreateFestivalError(opts, errors.Wrap(err, "processing markers"))
		}

		if markerResult != nil {
			totalMarkersFilled += markerResult.Filled
			totalMarkersCount += markerResult.Total
			allMarkers = append(allMarkers, markerResult.Markers...)
		}
	}

	// For dry-run, output all markers and exit
	if opts.DryRun && totalMarkersCount > 0 {
		result := &MarkerResult{
			Markers: allMarkers,
			Total:   totalMarkersCount,
		}
		if err := PrintDryRunMarkers(result, opts.JSONOutput); err != nil {
			return emitCreateFestivalError(opts, err)
		}
		return nil
	}

	// Run post-create validation if configured
	var validationResult *ValidationSummary
	shouldValidate := config.ShouldValidate(agentCfg, opts.AgentMode)
	if shouldValidate {
		validationResult, err = RunPostCreateValidation(ctx, destDir)
		if err != nil {
			// Don't fail on validation errors, just report
			if !opts.JSONOutput {
				display.Warning("Validation failed: %v", err)
			}
		}

		// Block on errors if configured
		if validationResult != nil && !validationResult.OK {
			if config.ShouldBlockOnErrors(agentCfg, opts.AgentMode) {
				return emitCreateFestivalError(opts, errors.Validation("validation errors detected - fix issues before proceeding"))
			}
		}
	}

	if opts.JSONOutput {
		remainingMarkers := totalMarkersCount - totalMarkersFilled
		warnings := []string{}
		if remainingMarkers > 0 {
			warnings = append(warnings,
				fmt.Sprintf("CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers),
				"Run 'fest validate' to see which files need editing",
				"Run 'fest wizard fill FESTIVAL_GOAL.md' to fill markers interactively",
			)
		}
		warnings = append(warnings, "Next: Create phases with 'fest create phase --name PHASE_NAME'")

		return emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     true,
			Action: "create_festival",
			Festival: map[string]string{
				"name":      opts.Name,
				"slug":      slug,
				"dest":      destCategory,
				"id":        festivalID,
				"directory": dirName,
			},
			Created:        created,
			GatesDirectory: gatesDir,
			FestYAML:       festConfigPath,
			GateTemplates:  copiedGates,
			MarkersFilled:  totalMarkersFilled,
			MarkersTotal:   totalMarkersCount,
			Validation:     validationResult,
			Warnings:       warnings,
		})
	}

	// Show marker warning FIRST (before success message) for visibility
	remainingMarkers := totalMarkersCount - totalMarkersFilled
	if remainingMarkers > 0 {
		fmt.Println()
		display.Error("🚫 CRITICAL: %d unfilled markers - festival cannot be executed until resolved", remainingMarkers)
		display.Info("   Run 'fest validate' to see which files need editing")
		display.Info("   Run 'fest wizard fill FESTIVAL_GOAL.md' to fill markers interactively")
		fmt.Println()
	}

	display.Success("Created festival: %s (%s)", dirName, destCategory)
	display.Info("  ID: %s", festivalID)
	for _, p := range created {
		display.Info("  • %s", p)
	}

	// Report gates setup
	if len(copiedGates) > 0 {
		display.Success("Created gates/ directory with %d default templates", len(copiedGates))
		display.Info("  Quality gates configured in fest.yaml")
	}

	fmt.Println()
	fmt.Println("   Next steps:")
	fmt.Println("   1. cd", destDir)
	if remainingMarkers > 0 {
		fmt.Println("   2. Edit FESTIVAL_GOAL.md to define your objectives")
		fmt.Println("   3. fest create phase --name \"PLAN\" --after 0")
		fmt.Println("   4. fest validate  (check completion status)")
	} else {
		fmt.Println("   2. fest create phase --name \"PLAN\" --after 0")
		fmt.Println("   3. fest create phase --name \"IMPLEMENT\" --after 1")
		fmt.Println("   4. After creating tasks: fest gates apply --approve")
	}
	return nil
}

func parseTags(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := []string{}
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func emitCreateFestivalError(opts *CreateFestivalOptions, err error) error {
	if opts.JSONOutput {
		_ = emitCreateFestivalJSON(opts, createFestivalResult{
			OK:     false,
			Action: "create_festival",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
		return nil
	}
	return err
}

func emitCreateFestivalJSON(opts *CreateFestivalOptions, res createFestivalResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(res)
}

// Slugify converts a string to a URL-safe slug.
func Slugify(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "festival"
	}
	return slug
}

// defaultFestivalGatesConfig creates a festival config with gates/ prefixed template paths.
// This is used when creating a new festival to set up default quality gates.
func defaultFestivalGatesConfig() *config.FestivalConfig {
	return &config.FestivalConfig{
		Version: "1.0",
		QualityGates: config.QualityGatesConfig{
			Enabled:    true,
			AutoAppend: true,
			Tasks: []config.QualityGateTask{
				{
					ID:       "testing_and_verify",
					Template: "gates/QUALITY_GATE_TESTING",
					Name:     "Testing and Verification",
					Enabled:  true,
				},
				{
					ID:       "code_review",
					Template: "gates/QUALITY_GATE_REVIEW",
					Name:     "Code Review",
					Enabled:  true,
				},
				{
					ID:       "review_results_iterate",
					Template: "gates/QUALITY_GATE_ITERATE",
					Name:     "Review Results and Iterate",
					Enabled:  true,
				},
			},
		},
		ExcludedPatterns: []string{
			"*_planning",
			"*_research",
			"*_requirements",
			"*_docs",
		},
		Templates: config.TemplatePrefs{
			TaskDefault:  "TASK_TEMPLATE_SIMPLE",
			PreferSimple: true,
		},
		Tracking: config.TrackingConfig{
			Enabled:      true,
			ChecksumFile: ".festival-checksums.json",
		},
	}
}
