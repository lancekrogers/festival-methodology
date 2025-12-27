package context

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestFestival(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()

	// Create festival structure
	festivalPath := filepath.Join(tmpDir, "test-festival")
	phase1Path := filepath.Join(festivalPath, "001_Research")
	seq1Path := filepath.Join(phase1Path, "01_analysis")

	dirs := []string{festivalPath, phase1Path, seq1Path}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}

	// Create FESTIVAL_GOAL.md
	festivalGoal := `# Test Festival

## Objective

Test festival for context building.

## Success Criteria

- [ ] Context loading works
- [ ] JSON output works
`
	if err := os.WriteFile(filepath.Join(festivalPath, "FESTIVAL_GOAL.md"), []byte(festivalGoal), 0644); err != nil {
		t.Fatalf("Failed to write FESTIVAL_GOAL.md: %v", err)
	}

	// Create PHASE_GOAL.md
	phaseGoal := `# Phase Goal: Research

## Objective

Complete research tasks.
`
	if err := os.WriteFile(filepath.Join(phase1Path, "PHASE_GOAL.md"), []byte(phaseGoal), 0644); err != nil {
		t.Fatalf("Failed to write PHASE_GOAL.md: %v", err)
	}

	// Create SEQUENCE_GOAL.md
	sequenceGoal := `# Sequence Goal: Analysis

## Objective

Perform initial analysis.
`
	if err := os.WriteFile(filepath.Join(seq1Path, "SEQUENCE_GOAL.md"), []byte(sequenceGoal), 0644); err != nil {
		t.Fatalf("Failed to write SEQUENCE_GOAL.md: %v", err)
	}

	// Create task file
	taskContent := `# Task: 01_analyze

> **Task Number**: 01 | **Parallel Execution**: No | **Dependencies**: None | **Autonomy Level**: high

## Objective

Analyze the requirements.

## Deliverables

- requirements.md
- analysis.md
`
	if err := os.WriteFile(filepath.Join(seq1Path, "01_analyze.md"), []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to write task file: %v", err)
	}

	return festivalPath
}

func TestBuilder_Build_FestivalLevel(t *testing.T) {
	festivalPath := setupTestFestival(t)

	// Save and restore working directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(festivalPath)

	builder := NewBuilder(festivalPath, DepthStandard)
	output, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if output.Location == nil {
		t.Fatal("Location should not be nil")
	}

	if output.Festival == nil {
		t.Fatal("Festival should not be nil")
	}

	if output.Festival.PhaseCount != 1 {
		t.Errorf("PhaseCount = %d, want 1", output.Festival.PhaseCount)
	}

	if output.Festival.Goal == nil {
		t.Fatal("Festival Goal should not be nil")
	}
}

func TestBuilder_Build_PhaseLevel(t *testing.T) {
	festivalPath := setupTestFestival(t)
	phasePath := filepath.Join(festivalPath, "001_Research")

	// Resolve to absolute paths with symlink evaluation (for macOS /tmp -> /private/tmp)
	absFestival, _ := filepath.EvalSymlinks(festivalPath)
	absPhase, _ := filepath.EvalSymlinks(phasePath)

	// Save and restore working directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(absPhase); err != nil {
		t.Fatalf("Failed to chdir to %s: %v", absPhase, err)
	}

	builder := NewBuilder(absFestival, DepthStandard)
	output, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if output.Location.Level != "phase" {
		t.Errorf("Level = %q, want %q", output.Location.Level, "phase")
	}

	if output.Phase == nil {
		t.Fatal("Phase should not be nil")
	}

	if output.Phase.SequenceCount != 1 {
		t.Errorf("SequenceCount = %d, want 1", output.Phase.SequenceCount)
	}
}

func TestBuilder_Build_SequenceLevel(t *testing.T) {
	festivalPath := setupTestFestival(t)
	seqPath := filepath.Join(festivalPath, "001_Research", "01_analysis")

	// Resolve to absolute paths with symlink evaluation (for macOS /tmp -> /private/tmp)
	absFestival, _ := filepath.EvalSymlinks(festivalPath)
	absSeq, _ := filepath.EvalSymlinks(seqPath)

	// Save and restore working directory
	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	if err := os.Chdir(absSeq); err != nil {
		t.Fatalf("Failed to chdir to %s: %v", absSeq, err)
	}

	builder := NewBuilder(absFestival, DepthStandard)
	output, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if output.Location.Level != "sequence" {
		t.Errorf("Level = %q, want %q", output.Location.Level, "sequence")
	}

	if output.Sequence == nil {
		t.Fatal("Sequence should not be nil")
	}

	if output.Sequence.TaskCount != 1 {
		t.Errorf("TaskCount = %d, want 1", output.Sequence.TaskCount)
	}
}

func TestBuilder_DepthMinimal(t *testing.T) {
	festivalPath := setupTestFestival(t)

	// Create FESTIVAL_RULES.md to test it's NOT included in minimal
	rulesContent := `# FESTIVAL_RULES

## Testing

### Always Test

Write tests for all code.
`
	if err := os.WriteFile(filepath.Join(festivalPath, "FESTIVAL_RULES.md"), []byte(rulesContent), 0644); err != nil {
		t.Fatalf("Failed to write FESTIVAL_RULES.md: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(festivalPath)

	builder := NewBuilder(festivalPath, DepthMinimal)
	output, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Rules should NOT be loaded for minimal depth
	if len(output.Rules) > 0 {
		t.Errorf("Rules should be empty for minimal depth, got %d", len(output.Rules))
	}
}

func TestBuilder_DepthFull(t *testing.T) {
	festivalPath := setupTestFestival(t)

	// Create FESTIVAL_RULES.md
	rulesContent := `# FESTIVAL_RULES

## Error Handling

### Always Wrap Errors

Use proper error wrapping.
`
	if err := os.WriteFile(filepath.Join(festivalPath, "FESTIVAL_RULES.md"), []byte(rulesContent), 0644); err != nil {
		t.Fatalf("Failed to write FESTIVAL_RULES.md: %v", err)
	}

	// Create CONTEXT.md with decisions
	contextContent := `# CONTEXT

### 2024-12-01: Use Go

We decided to use Go for implementation.
`
	if err := os.WriteFile(filepath.Join(festivalPath, "CONTEXT.md"), []byte(contextContent), 0644); err != nil {
		t.Fatalf("Failed to write CONTEXT.md: %v", err)
	}

	origDir, _ := os.Getwd()
	defer os.Chdir(origDir)
	os.Chdir(festivalPath)

	builder := NewBuilder(festivalPath, DepthFull)
	output, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Rules should be loaded for full depth
	if len(output.Rules) == 0 {
		t.Error("Rules should be loaded for full depth")
	}

	// Decisions should be loaded for full depth
	if len(output.Decisions) == 0 {
		t.Error("Decisions should be loaded for full depth")
	}
}

func TestIsPhaseDir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"numbered phase", "001_Research", true},
		{"two digit phase", "01_Implementation", true},
		{"hidden dir", ".git", false},
		{"underscore prefix", "_templates", false},
		{"short name", "ab", false},
		{"no number", "Research", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isPhaseDir(tc.input)
			if got != tc.expected {
				t.Errorf("isPhaseDir(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsTaskFile(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"numbered task", "01_task.md", true},
		{"two digit task", "01_analyze.md", true},
		{"goal file", "PHASE_GOAL.md", false},
		{"festival goal", "FESTIVAL_GOAL.md", false},
		{"not markdown", "01_task.txt", false},
		{"no number", "task.md", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isTaskFile(tc.input)
			if got != tc.expected {
				t.Errorf("isTaskFile(%q) = %v, want %v", tc.input, got, tc.expected)
			}
		})
	}
}

func TestExtractTaskNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"01_task.md", 1},
		{"12_analyze.md", 12},
		{"003_implement.md", 3},
		{"task.md", 0},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := extractTaskNumber(tc.input)
			if got != tc.expected {
				t.Errorf("extractTaskNumber(%q) = %d, want %d", tc.input, got, tc.expected)
			}
		})
	}
}
