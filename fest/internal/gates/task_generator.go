// Package gates provides quality gate task generation.
package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

// TaskGenerator generates quality gate task files in sequences.
type TaskGenerator struct {
	templateRoot string
	catalog      *tpl.Catalog
	manager      *tpl.Manager
}

// NewTaskGenerator creates a task generator with the given template root.
func NewTaskGenerator(templateRoot string) (*TaskGenerator, error) {
	catalog, _ := tpl.LoadCatalog(templateRoot)

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
func (g *TaskGenerator) GenerateForSequence(
	ctx context.Context,
	sequencePath string,
	gates []GateTask,
	opts GenerateOptions,
) ([]GenerateResult, []string, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("context cancelled: %w", err)
	}

	var results []GenerateResult
	var warnings []string

	// Get existing tasks in sequence
	entries, err := os.ReadDir(sequencePath)
	if err != nil {
		return nil, nil, fmt.Errorf("reading sequence directory: %w", err)
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
			content := g.renderGateContent(gate, taskNum)

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
func (g *TaskGenerator) renderGateContent(gate GateTask, taskNum int) string {
	// Build context
	tmplCtx := tpl.NewContext()
	tmplCtx.SetTask(taskNum, gate.ID)
	if gate.Customizations != nil {
		for k, v := range gate.Customizations {
			tmplCtx.SetCustom(k, v)
		}
	}

	// Try catalog first
	var content string
	if g.catalog != nil {
		content, _ = g.manager.RenderByID(g.catalog, gate.Template, tmplCtx)
	}

	// Fallback to direct file load
	if content == "" && g.templateRoot != "" {
		tpath := filepath.Join(g.templateRoot, gate.Template+".md")
		if _, err := os.Stat(tpath); err == nil {
			loader := tpl.NewLoader()
			t, err := loader.Load(tpath)
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
	return "", fmt.Errorf("no festival root found")
}

// FindImplementationSequences finds all implementation sequences in a festival.
func FindImplementationSequences(festivalRoot string, excludePatterns []string) ([]string, error) {
	var sequences []string

	entries, err := os.ReadDir(festivalRoot)
	if err != nil {
		return nil, fmt.Errorf("reading festival root: %w", err)
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
}

// FindSequencesWithInfo finds sequences and returns detailed info.
func FindSequencesWithInfo(festivalRoot string, excludePatterns []string) ([]SequenceInfo, error) {
	var sequences []SequenceInfo

	entries, err := os.ReadDir(festivalRoot)
	if err != nil {
		return nil, fmt.Errorf("reading festival root: %w", err)
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
			})
		}
	}

	return sequences, nil
}
