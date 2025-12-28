package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewParser(t *testing.T) {
	opts := ParseOptions{
		IncludeContent: true,
		Format:         "json",
	}
	p := NewParser(opts)
	if p == nil {
		t.Fatal("NewParser returned nil")
	}
}

func TestParseFestival(t *testing.T) {
	// Create test festival structure
	tmpDir := t.TempDir()
	festDir := filepath.Join(tmpDir, "test-festival")
	phaseDir := filepath.Join(festDir, "001_Planning")
	seqDir := filepath.Join(phaseDir, "01_setup")

	os.MkdirAll(seqDir, 0755)

	// Create PHASE_GOAL.md
	phaseGoal := `# Phase Goal

## Objective
Complete the planning phase.

## Success Criteria
- [ ] Define requirements
- [ ] Create timeline
`
	os.WriteFile(filepath.Join(phaseDir, "PHASE_GOAL.md"), []byte(phaseGoal), 0644)

	// Create SEQUENCE_GOAL.md
	seqGoal := "# Sequence Goal\n\n## Objective\nSetup tasks"
	os.WriteFile(filepath.Join(seqDir, "SEQUENCE_GOAL.md"), []byte(seqGoal), 0644)

	// Create task file
	task := `# First Task

Task description here.
`
	os.WriteFile(filepath.Join(seqDir, "01_first_task.md"), []byte(task), 0644)

	// Parse
	p := NewParser(ParseOptions{InferMissing: true})
	festival, err := p.ParseFestival(festDir)
	if err != nil {
		t.Fatalf("ParseFestival error: %v", err)
	}

	if festival.Type != "festival" {
		t.Errorf("Type = %q, want festival", festival.Type)
	}

	if len(festival.Phases) != 1 {
		t.Fatalf("Phases count = %d, want 1", len(festival.Phases))
	}

	phase := festival.Phases[0]
	if phase.Type != "phase" {
		t.Errorf("Phase type = %q, want phase", phase.Type)
	}
	if phase.Order != 1 {
		t.Errorf("Phase order = %d, want 1", phase.Order)
	}

	if len(phase.Sequences) != 1 {
		t.Fatalf("Sequences count = %d, want 1", len(phase.Sequences))
	}

	seq := phase.Sequences[0]
	if seq.Type != "sequence" {
		t.Errorf("Sequence type = %q, want sequence", seq.Type)
	}

	if len(seq.Tasks) != 1 {
		t.Fatalf("Tasks count = %d, want 1", len(seq.Tasks))
	}

	task0 := seq.Tasks[0]
	if task0.Type != "task" {
		t.Errorf("Task type = %q, want task", task0.Type)
	}
	if task0.Order != 1 {
		t.Errorf("Task order = %d, want 1", task0.Order)
	}
}

func TestParseGoalContent(t *testing.T) {
	content := `# Goal

## Objective
This is the objective text.

## Success Criteria
- [ ] First criterion
- [ ] Second criterion
- [x] Completed criterion

## Context
Some context here.
`
	goal := parseGoalContent(content)

	if goal.Objective != "This is the objective text." {
		t.Errorf("Objective = %q, want 'This is the objective text.'", goal.Objective)
	}

	if len(goal.SuccessCriteria) != 3 {
		t.Errorf("SuccessCriteria count = %d, want 3", len(goal.SuccessCriteria))
	}

	if goal.Context != "Some context here." {
		t.Errorf("Context = %q, want 'Some context here.'", goal.Context)
	}
}

func TestExtractOrder(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"001_Planning", 1},
		{"01_setup", 1},
		{"42_answer", 42},
		{"123_test", 123},
		{"no_number", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractOrder(tc.name)
			if got != tc.expected {
				t.Errorf("extractOrder(%q) = %d, want %d", tc.name, got, tc.expected)
			}
		})
	}
}

func TestExtractName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"001_Planning_Phase", "Planning Phase"},
		{"01_setup", "setup"},
		{"first_task", "first task"},
		{"123_with_numbers", "with numbers"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := extractName(tc.input)
			if got != tc.expected {
				t.Errorf("extractName(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}

func TestIsGateFile(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		{"quality_gate_review.md", true},
		{"gate_testing.md", true},
		{"review_gate.md", true},
		{"testing_and_verify.md", true},
		{"regular_task.md", false},
		{"01_implementation.md", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isGateFile(tc.name)
			if got != tc.expected {
				t.Errorf("isGateFile(%q) = %v, want %v", tc.name, got, tc.expected)
			}
		})
	}
}

func TestFlattenByType(t *testing.T) {
	festival := &ParsedFestival{
		ParsedEntity: ParsedEntity{Type: "festival", ID: "test"},
		Phases: []ParsedPhase{
			{
				ParsedEntity: ParsedEntity{Type: "phase", ID: "phase1"},
				Sequences: []ParsedSequence{
					{
						ParsedEntity: ParsedEntity{Type: "sequence", ID: "seq1"},
						Tasks: []ParsedTask{
							{ParsedEntity: ParsedEntity{Type: "task", ID: "task1"}},
							{ParsedEntity: ParsedEntity{Type: "gate", ID: "gate1"}},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		entityType string
		expected   int
	}{
		{"festival", 1},
		{"phase", 1},
		{"sequence", 1},
		{"task", 1},
		{"gate", 1},
	}

	for _, tc := range tests {
		t.Run(tc.entityType, func(t *testing.T) {
			entities := FlattenByType(festival, tc.entityType)
			if len(entities) != tc.expected {
				t.Errorf("FlattenByType(%q) count = %d, want %d", tc.entityType, len(entities), tc.expected)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	entity := &ParsedEntity{
		Type:   "task",
		ID:     "test",
		Name:   "Test Task",
		Status: "pending",
	}

	// Test JSON
	jsonData, err := Format(entity, "json", false)
	if err != nil {
		t.Fatalf("Format JSON error: %v", err)
	}
	if len(jsonData) == 0 {
		t.Error("JSON output should not be empty")
	}

	// Test YAML
	yamlData, err := Format(entity, "yaml", false)
	if err != nil {
		t.Fatalf("Format YAML error: %v", err)
	}
	if len(yamlData) == 0 {
		t.Error("YAML output should not be empty")
	}
}
