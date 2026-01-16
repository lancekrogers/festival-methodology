// Package gates provides quality gate task generation.
package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

// TaskGenerator generates quality gate task files in sequences.
type TaskGenerator struct {
	templateRoot string
	catalog      *tpl.Catalog
	manager      *tpl.Manager
}

// NewTaskGenerator creates a task generator with the given template root.
func NewTaskGenerator(ctx context.Context, templateRoot string) (*TaskGenerator, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("NewTaskGenerator")
	}

	catalog, _ := tpl.LoadCatalog(ctx, templateRoot)

	return &TaskGenerator{
		templateRoot: templateRoot,
		catalog:      catalog,
		manager:      tpl.NewManager(),
	}, nil
}

// GenerateOptions controls task file generation.
type GenerateOptions struct {
	DryRun  bool // Preview without creating files
	Force   bool // Overwrite existing files
	Verbose bool // Include verbose output
}

// GenerateResult represents the result of generating a task file.
type GenerateResult struct {
	Type     string `json:"type"`     // "create", "skip", "exists"
	Path     string `json:"path"`     // Full path to task file
	Template string `json:"template"` // Template used
	Reason   string `json:"reason"`   // Reason for skip
	TaskID   string `json:"task_id"`  // Gate task ID
}

// GenerateSummary provides statistics about generation.
type GenerateSummary struct {
	TotalSequences   int `json:"total_sequences"`
	SequencesUpdated int `json:"sequences_updated"`
	FilesCreated     int `json:"files_created"`
	FilesSkipped     int `json:"files_skipped"`
}

// GenerateForSequence generates task files for gates in a single sequence.
// festivalPath is optional - if provided, it's used to resolve gates/ prefixed templates.
func (g *TaskGenerator) GenerateForSequence(
	ctx context.Context,
	sequencePath string,
	gates []GateTask,
	opts GenerateOptions,
	festivalPath ...string,
) ([]GenerateResult, []string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, errors.Wrap(err, "context cancelled").
			WithOp("TaskGenerator.GenerateForSequence")
	}

	var results []GenerateResult
	var warnings []string

	// Get existing tasks in sequence
	entries, err := os.ReadDir(sequencePath)
	if err != nil {
		return nil, nil, errors.IO("reading sequence directory", err).
			WithField("path", sequencePath)
	}

	// Find highest task number and existing task IDs
	maxNum := 0
	existingTasks := make(map[string]bool)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if entry.Name() == "SEQUENCE_GOAL.md" {
			continue
		}
		existingTasks[entry.Name()] = true
		num := festival.ParseTaskNumber(entry.Name())
		if num > maxNum {
			maxNum = num
		}
	}

	// Generate each gate task
	for i, gate := range gates {
		if !gate.Enabled {
			continue
		}

		taskNum := maxNum + i + 1
		taskFileName := tpl.FormatTaskID(taskNum, gate.ID)
		taskPath := filepath.Join(sequencePath, taskFileName)

		// Check if task already exists (by ID pattern)
		taskExists := false
		for existingName := range existingTasks {
			if strings.Contains(existingName, gate.ID) {
				taskExists = true
				break
			}
		}

		if taskExists {
			results = append(results, GenerateResult{
				Type:   "exists",
				Path:   taskPath,
				TaskID: gate.ID,
				Reason: "task_already_exists",
			})
			continue
		}

		// Check if file exists (could be renamed)
		if _, err := os.Stat(taskPath); err == nil {
			if !opts.Force {
				results = append(results, GenerateResult{
					Type:   "skip",
					Path:   taskPath,
					TaskID: gate.ID,
					Reason: "file_exists",
				})
				continue
			}
		}

		// Create the task
		if !opts.DryRun {
			var festPath string
			if len(festivalPath) > 0 {
				festPath = festivalPath[0]
			}
			content := g.renderGateContent(ctx, gate, taskNum, festPath)

			if err := os.WriteFile(taskPath, []byte(content), 0644); err != nil {
				warnings = append(warnings, fmt.Sprintf("Failed to write %s: %v", taskPath, err))
				continue
			}
		}

		results = append(results, GenerateResult{
			Type:     "create",
			Path:     taskPath,
			Template: gate.Template,
			TaskID:   gate.ID,
		})
	}

	return results, warnings, nil
}

// renderGateContent renders the content for a gate task file.
// festivalPath is used to resolve gates/ prefixed templates.
func (g *TaskGenerator) renderGateContent(ctx context.Context, gate GateTask, taskNum int, festivalPath string) string {
	// Build context
	tmplCtx := tpl.NewContext()
	tmplCtx.SetTask(taskNum, gate.ID)
	if gate.Customizations != nil {
		for k, v := range gate.Customizations {
			tmplCtx.SetCustom(k, v)
		}
	}

	var content string

	// Handle "gates/TEMPLATE_NAME" format - resolve from festival's gates/ directory first
	if strings.HasPrefix(gate.Template, "gates/") && festivalPath != "" {
		templateName := strings.TrimPrefix(gate.Template, "gates/")
		gatesPath := filepath.Join(festivalPath, "gates", templateName+".md")

		if _, err := os.Stat(gatesPath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(ctx, gatesPath)
			if err == nil {
				if strings.Contains(t.Content, "{{") {
					content, _ = g.manager.Render(t, tmplCtx)
				} else {
					content = t.Content
				}
			}
		}
	}

	// Try catalog if not found in festival gates/
	if content == "" && g.catalog != nil {
		// For gates/ prefix, try with the stripped template name
		templateID := gate.Template
		if strings.HasPrefix(templateID, "gates/") {
			templateID = strings.TrimPrefix(templateID, "gates/")
		}
		content, _ = g.manager.RenderByID(ctx, g.catalog, templateID, tmplCtx)
	}

	// Fallback to direct file load from template root
	if content == "" && g.templateRoot != "" {
		templateName := gate.Template
		if strings.HasPrefix(templateName, "gates/") {
			templateName = strings.TrimPrefix(templateName, "gates/")
		}
		tpath := filepath.Join(g.templateRoot, "gates", templateName+".md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(ctx, tpath)
			if err == nil {
				if strings.Contains(t.Content, "{{") {
					content, _ = g.manager.Render(t, tmplCtx)
				} else {
					content = t.Content
				}
			}
		}
	}

	// Use default content if template not found
	if content == "" {
		content = generateDefaultGateContent(gate)
	}

	return content
}

// generateDefaultGateContent creates default content for a gate task.
func generateDefaultGateContent(gate GateTask) string {
	name := gate.Name
	if name == "" {
		name = strings.ReplaceAll(gate.ID, "_", " ")
		name = strings.Title(name)
	}

	return fmt.Sprintf(`# Task: %s

## Objective

%s

## Requirements

- [ ] Complete this quality gate task
- [ ] Verify all requirements are met
- [ ] Document any findings

## Definition of Done

- [ ] Task completed successfully
- [ ] All checks pass
- [ ] Ready to proceed

## Notes

[Add notes here]
`, name, name)
}

// FindFestivalRoot finds the festival root directory from a starting path.
func FindFestivalRoot(startPath string) (string, error) {
	path := startPath
	for {
		// Check for festival markers
		if _, err := os.Stat(filepath.Join(path, "FESTIVAL_OVERVIEW.md")); err == nil {
			return path, nil
		}
		if _, err := os.Stat(filepath.Join(path, "fest.yaml")); err == nil {
			return path, nil
		}
		if _, err := os.Stat(filepath.Join(path, "FESTIVAL_GOAL.md")); err == nil {
			return path, nil
		}

		// Move up
		parent := filepath.Dir(path)
		if parent == path {
			break
		}
		path = parent
	}
	return "", errors.NotFound("festival root").
		WithField("start_path", startPath)
}

// FindImplementationSequences finds all implementation sequences in a festival.
func FindImplementationSequences(festivalRoot string, excludePatterns []string) ([]string, error) {
	var sequences []string

	entries, err := os.ReadDir(festivalRoot)
	if err != nil {
		return nil, errors.IO("reading festival root", err).
			WithField("path", festivalRoot)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Check if it's a phase (starts with number)
		if !festival.IsPhase(entry.Name()) {
			continue
		}

		phasePath := filepath.Join(festivalRoot, entry.Name())

		// Walk through sequences in phase
		seqEntries, err := os.ReadDir(phasePath)
		if err != nil {
			continue
		}

		for _, seqEntry := range seqEntries {
			if !seqEntry.IsDir() {
				continue
			}

			// Check if it's a sequence (starts with number)
			if !festival.IsSequence(seqEntry.Name()) {
				continue
			}

			// Check excluded patterns
			if isSequenceExcluded(seqEntry.Name(), excludePatterns) {
				continue
			}

			sequences = append(sequences, filepath.Join(phasePath, seqEntry.Name()))
		}
	}

	return sequences, nil
}

// isSequenceExcluded checks if a sequence matches any excluded pattern.
func isSequenceExcluded(sequenceName string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, sequenceName)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}
	return false
}

// SequenceInfo contains information about a sequence for generation.
type SequenceInfo struct {
	Path      string // Full path to sequence directory
	PhasePath string // Path to parent phase
	Name      string // Sequence directory name
	PhaseType string // Phase type: "implementation", "planning", "research", "review", "action"
	PhaseName string // Name of the parent phase directory
}

// DetectPhaseType determines the phase type from the phase directory.
// Priority order:
// 1. Read from PHASE_GOAL.md frontmatter (fest_phase_type field)
// 2. Fall back to inferring from directory name
// Returns: "planning", "implementation", "research", "review", "non_coding_action"
// Returns empty string if type cannot be determined (error case).
func DetectPhaseType(phasePath string) string {
	// First try to read from PHASE_GOAL.md frontmatter
	goalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		if fm, _, err := frontmatter.Parse(content); err == nil && fm != nil {
			if fm.PhaseType != "" {
				// Map frontmatter.PhaseType to our internal naming
				return mapPhaseType(string(fm.PhaseType))
			}
		}
	}

	// Fall back to directory name inference
	phaseName := filepath.Base(phasePath)
	return inferPhaseTypeFromName(phaseName)
}

// mapPhaseType normalizes phase type values to internal naming.
func mapPhaseType(phaseType string) string {
	switch strings.ToLower(phaseType) {
	case "planning", "plan":
		return "planning"
	case "implementation", "implement", "build":
		return "implementation"
	case "research", "discovery":
		return "research"
	case "review", "qa":
		return "review"
	case "deployment", "deploy", "action", "non_coding_action":
		return "non_coding_action"
	default:
		return phaseType
	}
}

// inferPhaseTypeFromName infers phase type from directory name.
// Returns empty string if type cannot be determined.
func inferPhaseTypeFromName(phaseName string) string {
	lower := strings.ToLower(phaseName)

	switch {
	case strings.Contains(lower, "planning") || strings.Contains(lower, "plan"):
		return "planning"
	case strings.Contains(lower, "research") || strings.Contains(lower, "discovery"):
		return "research"
	case strings.Contains(lower, "design"):
		return "research" // Design phases use research-like structure
	case strings.Contains(lower, "review") || strings.Contains(lower, "qa") || strings.Contains(lower, "uat"):
		return "review"
	// Action phases: deployment, configuration, publishing, migrations, operational tasks
	case strings.Contains(lower, "deployment") || strings.Contains(lower, "deploy") ||
		strings.Contains(lower, "release") || strings.Contains(lower, "action") ||
		strings.Contains(lower, "operation") || strings.Contains(lower, "config") ||
		strings.Contains(lower, "publish") || strings.Contains(lower, "migrat"):
		return "non_coding_action"
	case strings.Contains(lower, "implementation") || strings.Contains(lower, "implement") ||
		strings.Contains(lower, "develop") || strings.Contains(lower, "build") ||
		strings.Contains(lower, "foundation") || strings.Contains(lower, "critical"):
		return "implementation"
	default:
		return "" // No default - require explicit type
	}
}

// FindSequencesWithInfo finds sequences and returns detailed info.
func FindSequencesWithInfo(festivalRoot string, excludePatterns []string) ([]SequenceInfo, error) {
	var sequences []SequenceInfo

	entries, err := os.ReadDir(festivalRoot)
	if err != nil {
		return nil, errors.IO("reading festival root", err).
			WithField("path", festivalRoot)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		if !festival.IsPhase(entry.Name()) {
			continue
		}

		phasePath := filepath.Join(festivalRoot, entry.Name())

		seqEntries, err := os.ReadDir(phasePath)
		if err != nil {
			continue
		}

		// Detect phase type from phase path (checks frontmatter first)
		phaseName := entry.Name()
		phaseType := DetectPhaseType(phasePath)

		for _, seqEntry := range seqEntries {
			if !seqEntry.IsDir() {
				continue
			}

			if !festival.IsSequence(seqEntry.Name()) {
				continue
			}

			if isSequenceExcluded(seqEntry.Name(), excludePatterns) {
				continue
			}

			sequences = append(sequences, SequenceInfo{
				Path:      filepath.Join(phasePath, seqEntry.Name()),
				PhasePath: phasePath,
				Name:      seqEntry.Name(),
				PhaseType: phaseType,
				PhaseName: phaseName,
			})
		}
	}

	return sequences, nil
}

// DiscoverGatesForPhaseType reads gate templates from the festival's gates/{phase_type}/ directory.
// Returns gate tasks constructed from the .md files found in that directory.
// Returns an error if the directory doesn't exist or contains no gates.
func DiscoverGatesForPhaseType(festivalPath, phaseType string) ([]GateTask, error) {
	gatesDir := filepath.Join(festivalPath, "gates", phaseType)

	// Check if directory exists
	if _, err := os.Stat(gatesDir); os.IsNotExist(err) {
		return nil, errors.NotFound("gates directory for phase type").
			WithField("phase_type", phaseType).
			WithField("path", gatesDir).
			WithField("hint", fmt.Sprintf("Create gates directory: %s", gatesDir))
	}

	// Read all .md files from the directory
	entries, err := os.ReadDir(gatesDir)
	if err != nil {
		return nil, errors.IO("reading gates directory", err).
			WithField("path", gatesDir)
	}

	var gates []GateTask
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		// Extract gate ID from filename (e.g., "testing.md" -> "testing")
		baseName := strings.TrimSuffix(entry.Name(), ".md")
		gateID := strings.ReplaceAll(baseName, "-", "_")

		// Build template path that matches the new structure
		templatePath := fmt.Sprintf("gates/%s/%s", phaseType, baseName)

		// Try to read gate name from frontmatter if available
		gateName := strings.Title(strings.ReplaceAll(baseName, "_", " "))
		filePath := filepath.Join(gatesDir, entry.Name())
		if content, err := os.ReadFile(filePath); err == nil {
			if fm, _, err := frontmatter.Parse(content); err == nil && fm != nil && fm.Name != "" {
				gateName = fm.Name
			}
		}

		gates = append(gates, GateTask{
			ID:       gateID,
			Template: templatePath,
			Name:     gateName,
			Enabled:  true,
		})
	}

	if len(gates) == 0 {
		return nil, errors.Validation("no gate templates found in directory").
			WithField("phase_type", phaseType).
			WithField("path", gatesDir)
	}

	return gates, nil
}
