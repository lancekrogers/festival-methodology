package context

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Builder constructs context output for a given location
type Builder struct {
	festivalPath string
	depth        Depth
	metadata     *config.FestivalMetadata // Loaded from fest.yaml
}

// NewBuilder creates a new context builder
func NewBuilder(festivalPath string, depth Depth) *Builder {
	return &Builder{
		festivalPath: festivalPath,
		depth:        depth,
	}
}

// formatNodeReference creates a node reference string from festival ID and location.
// Format: ID:P###.S##.T## (e.g., GU0001:P002.S01.T03)
func formatNodeReference(festivalID string, phase, sequence, task int) string {
	if festivalID == "" {
		return ""
	}
	return fmt.Sprintf("%s:P%03d.S%02d.T%02d", festivalID, phase, sequence, task)
}

// Build constructs the complete context output for the current location
func (b *Builder) Build() (*ContextOutput, error) {
	output := &ContextOutput{
		Depth: b.depth,
	}

	// Load festival metadata from fest.yaml (optional - legacy festivals may not have it)
	b.loadMetadata()

	// Determine location
	location, err := b.determineLocation()
	if err != nil {
		return nil, err
	}
	output.Location = location

	// Load festival context (always included)
	festivalCtx, err := b.loadFestivalContext()
	if err != nil {
		return nil, err
	}
	output.Festival = festivalCtx

	// Load phase context if in a phase
	if location.PhasePath != "" {
		phaseCtx, err := b.loadPhaseContext(location.PhasePath)
		if err != nil {
			return nil, err
		}
		output.Phase = phaseCtx
	}

	// Load sequence context if in a sequence
	if location.SequencePath != "" {
		seqCtx, err := b.loadSequenceContext(location.SequencePath)
		if err != nil {
			return nil, err
		}
		output.Sequence = seqCtx
	}

	// Load task context if in a task or task specified
	if location.TaskPath != "" {
		taskCtx, err := b.loadTaskContext(location.TaskPath)
		if err != nil {
			return nil, err
		}
		output.Task = taskCtx
	}

	// Populate FestivalID and CurrentRef fields
	b.populateNodeReference(output, location)

	// Load rules for standard and full depth
	if b.depth == DepthStandard || b.depth == DepthFull {
		rules, err := b.loadRules()
		if err == nil { // Rules are optional
			output.Rules = b.filterRules(rules, output.Task)
		}

		// Load decisions
		decisions, err := b.loadDecisions()
		if err == nil { // Decisions are optional
			output.Decisions = b.filterDecisions(decisions)
		}
	}

	// Load dependency outputs for full depth
	if b.depth == DepthFull && output.Task != nil && len(output.Task.Dependencies) > 0 {
		depOutputs := b.loadDependencyOutputs(output.Task.Dependencies, location.SequencePath)
		output.DependencyOutputs = depOutputs
	}

	return output, nil
}

// loadMetadata loads festival metadata from fest.yaml
func (b *Builder) loadMetadata() {
	festConfig, err := config.LoadFestivalConfig(b.festivalPath)
	if err != nil {
		return // Metadata is optional
	}
	if festConfig.Metadata.ID != "" {
		b.metadata = &festConfig.Metadata
	}
}

// populateNodeReference sets FestivalID and CurrentRef on the output
func (b *Builder) populateNodeReference(output *ContextOutput, location *Location) {
	// If no metadata, leave fields as nil (JSON null)
	if b.metadata == nil || b.metadata.ID == "" {
		return
	}

	// Set FestivalID
	id := b.metadata.ID
	output.FestivalID = &id

	// Extract phase, sequence, and task numbers from location
	phaseNum := extractLocationNumber(location.PhaseName)
	seqNum := extractLocationNumber(location.SequenceName)
	taskNum := 0
	if output.Task != nil {
		taskNum = output.Task.TaskNumber
	}

	// Format and set CurrentRef
	ref := formatNodeReference(id, phaseNum, seqNum, taskNum)
	output.CurrentRef = &ref
}

// extractLocationNumber extracts the leading number from a location name (e.g., "001_Research" -> 1)
func extractLocationNumber(name string) int {
	num := 0
	for _, c := range name {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	return num
}

// BuildForTask constructs context for a specific task
func (b *Builder) BuildForTask(taskName string) (*ContextOutput, error) {
	// Find the task file
	taskPath, err := b.findTaskByName(taskName)
	if err != nil {
		return nil, err
	}

	// Create a builder at the task location (share metadata)
	taskBuilder := &Builder{
		festivalPath: b.festivalPath,
		depth:        b.depth,
		metadata:     b.metadata,
	}

	// Load metadata if not already loaded
	if taskBuilder.metadata == nil {
		taskBuilder.loadMetadata()
	}

	output, err := taskBuilder.Build()
	if err != nil {
		return nil, err
	}

	// Ensure the task context is loaded
	taskCtx, err := b.loadTaskContext(taskPath)
	if err != nil {
		return nil, err
	}
	output.Task = taskCtx
	output.Location.TaskPath = taskPath
	output.Location.TaskName = taskName
	output.Location.Level = "task"

	// Recalculate CurrentRef now that we have the task number
	taskBuilder.populateNodeReference(output, output.Location)

	return output, nil
}

// determineLocation determines the current location in the festival hierarchy
func (b *Builder) determineLocation() (*Location, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	location := &Location{
		FestivalPath: b.festivalPath,
		FestivalName: filepath.Base(b.festivalPath),
		Level:        "festival",
	}

	// Check if we're in a phase
	rel, err := filepath.Rel(b.festivalPath, cwd)
	if err != nil || strings.HasPrefix(rel, "..") {
		// Not inside festival, return festival level
		return location, nil
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "." {
		return location, nil
	}

	// First part should be phase
	phaseName := parts[0]
	if isPhaseDir(phaseName) {
		location.PhasePath = filepath.Join(b.festivalPath, phaseName)
		location.PhaseName = phaseName
		location.Level = "phase"
	}

	// Second part should be sequence
	if len(parts) >= 2 {
		seqName := parts[1]
		if isSequenceDir(seqName) {
			location.SequencePath = filepath.Join(b.festivalPath, phaseName, seqName)
			location.SequenceName = seqName
			location.Level = "sequence"
		}
	}

	// Check if there's a task file in the current directory
	if location.SequencePath != "" {
		entries, err := os.ReadDir(cwd)
		if err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && isTaskFile(entry.Name()) {
					// Not setting task path here - user should use --task flag
					break
				}
			}
		}
	}

	return location, nil
}

// loadFestivalContext loads the festival-level context
func (b *Builder) loadFestivalContext() (*FestivalContext, error) {
	ctx := &FestivalContext{
		Name: filepath.Base(b.festivalPath),
		Path: b.festivalPath,
	}

	// Load FESTIVAL_GOAL.md or fest.yaml
	goalPath := filepath.Join(b.festivalPath, "FESTIVAL_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		goal, _ := ParseGoalFile(content)
		ctx.Goal = goal
	}

	// Count phases
	entries, err := os.ReadDir(b.festivalPath)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() && isPhaseDir(entry.Name()) {
				ctx.PhaseCount++
			}
		}
	}

	return ctx, nil
}

// loadPhaseContext loads phase-level context
func (b *Builder) loadPhaseContext(phasePath string) (*PhaseContext, error) {
	ctx := &PhaseContext{
		Name: filepath.Base(phasePath),
		Path: phasePath,
	}

	// Load PHASE_GOAL.md
	goalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		goal, _ := ParseGoalFile(content)
		ctx.Goal = goal
	}

	// Detect phase type and freeform status
	ctx.PhaseType, ctx.IsFreeform = detectPhaseType(ctx.Name)

	// Count sequences or list topic directories based on phase type
	entries, err := os.ReadDir(phasePath)
	if err == nil {
		if ctx.IsFreeform {
			// For freeform phases, list topic directories
			for _, entry := range entries {
				if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
					ctx.TopicDirs = append(ctx.TopicDirs, entry.Name())
				}
			}
		} else {
			// For structured phases, count sequences
			for _, entry := range entries {
				if entry.IsDir() && isSequenceDir(entry.Name()) {
					ctx.SequenceCount++
				}
			}
		}
	}

	return ctx, nil
}

// detectPhaseType determines the phase type and whether it uses freeform structure.
// Returns (phaseType, isFreeform)
func detectPhaseType(phaseName string) (string, bool) {
	name := strings.ToLower(phaseName)

	switch {
	case strings.Contains(name, "planning"):
		return "planning", true
	case strings.Contains(name, "research"):
		return "research", true
	case strings.Contains(name, "design"):
		return "research", true // Design phases use research-like structure
	case strings.Contains(name, "implementation"):
		return "implementation", false
	case strings.Contains(name, "review"), strings.Contains(name, "uat"):
		return "review", false
	default:
		return "", false
	}
}

// loadSequenceContext loads sequence-level context
func (b *Builder) loadSequenceContext(seqPath string) (*SequenceContext, error) {
	ctx := &SequenceContext{
		Name: filepath.Base(seqPath),
		Path: seqPath,
	}

	// Load SEQUENCE_GOAL.md
	goalPath := filepath.Join(seqPath, "SEQUENCE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		goal, _ := ParseGoalFile(content)
		ctx.Goal = goal
	}

	// Count tasks
	entries, err := os.ReadDir(seqPath)
	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && isTaskFile(entry.Name()) {
				ctx.TaskCount++
			}
		}
	}

	return ctx, nil
}

// loadTaskContext loads task-level context
func (b *Builder) loadTaskContext(taskPath string) (*TaskContext, error) {
	content, err := os.ReadFile(taskPath)
	if err != nil {
		return nil, err
	}

	ctx, err := ParseTaskFile(content)
	if err != nil {
		return nil, err
	}

	ctx.Path = taskPath
	ctx.TaskNumber = extractTaskNumber(filepath.Base(taskPath))

	return ctx, nil
}

// loadRules loads rules from FESTIVAL_RULES.md
func (b *Builder) loadRules() ([]Rule, error) {
	rulesPath := filepath.Join(b.festivalPath, "FESTIVAL_RULES.md")
	content, err := os.ReadFile(rulesPath)
	if err != nil {
		return nil, err
	}

	return ParseRulesFile(content)
}

// filterRules filters rules based on relevance
func (b *Builder) filterRules(rules []Rule, task *TaskContext) []Rule {
	// Always include these categories
	alwaysInclude := map[string]bool{
		"error_handling":      true,
		"testing":             true,
		"context_propagation": true,
		"coding_standards":    true,
		"error handling":      true,
		"coding standards":    true,
		"context propagation": true,
	}

	var filtered []Rule
	for _, rule := range rules {
		category := strings.ToLower(rule.Category)
		if alwaysInclude[category] {
			filtered = append(filtered, rule)
			continue
		}

		// Include task-specific rules if we have task context
		if task != nil && b.isRuleRelevantToTask(rule, task) {
			filtered = append(filtered, rule)
		}
	}

	return filtered
}

// isRuleRelevantToTask checks if a rule is relevant to the current task
func (b *Builder) isRuleRelevantToTask(rule Rule, task *TaskContext) bool {
	taskNameLower := strings.ToLower(task.Name)
	categoryLower := strings.ToLower(rule.Category)
	titleLower := strings.ToLower(rule.Title)

	// Check if task name matches rule category or title
	for _, keyword := range []string{"test", "api", "validation", "security", "performance"} {
		if strings.Contains(taskNameLower, keyword) {
			if strings.Contains(categoryLower, keyword) || strings.Contains(titleLower, keyword) {
				return true
			}
		}
	}

	return false
}

// loadDecisions loads decisions from CONTEXT.md
func (b *Builder) loadDecisions() ([]Decision, error) {
	contextPath := filepath.Join(b.festivalPath, "CONTEXT.md")
	content, err := os.ReadFile(contextPath)
	if err != nil {
		return nil, err
	}

	return ParseContextFile(content)
}

// filterDecisions filters and sorts decisions
func (b *Builder) filterDecisions(decisions []Decision) []Decision {
	// Sort by date, most recent first
	sort.Slice(decisions, func(i, j int) bool {
		return decisions[i].Date.After(decisions[j].Date)
	})

	// For standard depth, limit to 5 most recent
	if b.depth == DepthStandard && len(decisions) > 5 {
		return decisions[:5]
	}

	return decisions
}

// loadDependencyOutputs loads outputs from dependency tasks
func (b *Builder) loadDependencyOutputs(deps []string, seqPath string) []DepOutput {
	var outputs []DepOutput

	for _, dep := range deps {
		depOutput := DepOutput{
			TaskID:   dep,
			TaskName: dep,
		}

		// Try to find the task file and extract its outputs
		taskPath := filepath.Join(seqPath, dep+".md")
		if content, err := os.ReadFile(taskPath); err == nil {
			text := string(content)
			// Look for outputs or deliverables section
			outputSection := extractSection(text, "Outputs", "##")
			if outputSection == "" {
				outputSection = extractSection(text, "Deliverables", "##")
			}
			if outputSection != "" {
				depOutput.Outputs = extractBulletList(outputSection)
			}
		}

		outputs = append(outputs, depOutput)
	}

	return outputs
}

// findTaskByName finds a task file by name
func (b *Builder) findTaskByName(taskName string) (string, error) {
	var taskPath string

	// Walk the festival looking for the task
	err := filepath.Walk(b.festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil
		}

		// Check if filename matches
		base := filepath.Base(path)
		if strings.Contains(base, taskName) && isTaskFile(base) {
			taskPath = path
			return filepath.SkipAll
		}

		return nil
	})

	if err != nil {
		return "", err
	}

	if taskPath == "" {
		return "", errors.NotFound("task")
	}

	return taskPath, nil
}

// Helper functions

func isPhaseDir(name string) bool {
	// Phases typically start with a number or have specific patterns
	if len(name) < 3 {
		return false
	}
	// Skip hidden directories and special files
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	// Check for numbered directories like "001_Research" or "01_Implementation"
	return name[0] >= '0' && name[0] <= '9'
}

func isSequenceDir(name string) bool {
	// Similar to phase, but inside a phase
	if len(name) < 2 {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	// Check for numbered directories
	return name[0] >= '0' && name[0] <= '9'
}

func isTaskFile(name string) bool {
	// Task files are numbered markdown files
	if !strings.HasSuffix(name, ".md") {
		return false
	}
	// Skip goal files
	if strings.Contains(strings.ToUpper(name), "GOAL") {
		return false
	}
	// Check for numbered files like "01_task.md"
	return len(name) > 3 && name[0] >= '0' && name[0] <= '9'
}

func extractTaskNumber(filename string) int {
	// Extract number from filename like "01_task.md"
	num := 0
	for _, c := range filename {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	return num
}
