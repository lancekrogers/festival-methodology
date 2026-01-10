package graduate

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestGenerator_Generate(t *testing.T) {
	source := &PlanningSource{
		Path:      "/test/001_PLANNING",
		PhaseName: "001_PLANNING",
		TopicDirs: []TopicDirectory{
			{
				Name:      "requirements",
				Path:      "/test/001_PLANNING/requirements",
				Documents: []string{"auth.md", "api.md"},
				DocCount:  2,
			},
			{
				Name:      "architecture",
				Path:      "/test/001_PLANNING/architecture",
				Documents: []string{"overview.md"},
				DocCount:  1,
			},
		},
		Decisions: []Decision{
			{ID: "ADR-001", Title: "Use PostgreSQL", Status: "accepted"},
		},
		Summary: &PlanningSummary{
			Goal:         "Build user authentication",
			KeyDecisions: []string{"Use PostgreSQL", "Use JWT"},
		},
		TotalDocs:  3,
		AnalyzedAt: time.Now(),
	}

	generator := NewGenerator("/test")
	ctx := context.Background()

	plan, err := generator.Generate(ctx, source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check target phase
	if plan.Target.Number != 2 {
		t.Errorf("Target.Number = %d, want 2", plan.Target.Number)
	}
	if plan.Target.PhaseName != "002_IMPLEMENTATION" {
		t.Errorf("Target.PhaseName = %q, want %q", plan.Target.PhaseName, "002_IMPLEMENTATION")
	}

	// Check phase goal
	if plan.PhaseGoal.Goal != "Implement: Build user authentication" {
		t.Errorf("PhaseGoal.Goal = %q, want goal with summary", plan.PhaseGoal.Goal)
	}

	// Check sequences - requirements and architecture should merge into "core"
	if len(plan.Sequences) != 1 {
		t.Errorf("Sequences count = %d, want 1 (merged core)", len(plan.Sequences))
	}
	if len(plan.Sequences) > 0 && plan.Sequences[0].Name != "core" {
		t.Errorf("First sequence name = %q, want %q", plan.Sequences[0].Name, "core")
	}

	// Check confidence (should be high with summary and decisions)
	if plan.Confidence < 0.8 {
		t.Errorf("Confidence = %f, want >= 0.8", plan.Confidence)
	}
}

func TestGenerator_Generate_NoSummary(t *testing.T) {
	source := &PlanningSource{
		Path:      "/test/001_PLANNING",
		PhaseName: "001_PLANNING",
		TopicDirs: []TopicDirectory{
			{
				Name:      "requirements",
				Documents: []string{"auth.md"},
				DocCount:  1,
			},
		},
		TotalDocs:  1,
		AnalyzedAt: time.Now(),
	}

	generator := NewGenerator("/test")
	ctx := context.Background()

	plan, err := generator.Generate(ctx, source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check default goal when no summary
	if plan.PhaseGoal.Goal != "Implement the system as planned" {
		t.Errorf("PhaseGoal.Goal = %q, want default goal", plan.PhaseGoal.Goal)
	}

	// Check reduced confidence
	if plan.Confidence > 0.8 {
		t.Errorf("Confidence = %f, want < 0.8 (no summary, no decisions, few docs)", plan.Confidence)
	}
}

func TestGenerator_Generate_UnmappedTopic(t *testing.T) {
	source := &PlanningSource{
		Path:      "/test/001_PLANNING",
		PhaseName: "001_PLANNING",
		TopicDirs: []TopicDirectory{
			{
				Name:      "custom_topic",
				Documents: []string{"doc.md"},
				DocCount:  1,
			},
		},
		TotalDocs:  1,
		AnalyzedAt: time.Now(),
	}

	generator := NewGenerator("/test")
	ctx := context.Background()

	plan, err := generator.Generate(ctx, source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check that warning is generated
	if len(plan.Warnings) == 0 {
		t.Error("Expected warning for unmapped topic")
	}

	// Check sequence is created with topic name
	if len(plan.Sequences) != 1 {
		t.Fatalf("Sequences count = %d, want 1", len(plan.Sequences))
	}
	if plan.Sequences[0].Name != "custom_topic" {
		t.Errorf("Sequence name = %q, want %q", plan.Sequences[0].Name, "custom_topic")
	}
}

func TestGenerator_Generate_ContextCanceled(t *testing.T) {
	source := &PlanningSource{
		PhaseName: "001_PLANNING",
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generator := NewGenerator("/test")
	_, err := generator.Generate(ctx, source)
	if err == nil {
		t.Error("Generate() expected error for canceled context")
	}
}

func TestExtractPhaseNumber(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"001_PLANNING", 1},
		{"002_IMPLEMENTATION", 2},
		{"010_REVIEW", 10},
		{"PLANNING", 1}, // Default
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractPhaseNumber(tt.name)
			if result != tt.expected {
				t.Errorf("extractPhaseNumber(%q) = %d, want %d", tt.name, result, tt.expected)
			}
		})
	}
}

func TestGetTaskPriority(t *testing.T) {
	tests := []struct {
		name     string
		expected int
	}{
		{"setup_database", 0},                 // "setup" is first
		{"init_config", 1},                    // "init" is second
		{"database_schema", 3},                // "database" is in list
		{"random_task", len(TaskPriority)},    // Not in list, returns len(TaskPriority)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getTaskPriority(tt.name)
			if result != tt.expected {
				t.Errorf("getTaskPriority(%q) = %d, want %d", tt.name, result, tt.expected)
			}
		})
	}
}

func TestGenerator_SequenceNumbering(t *testing.T) {
	source := &PlanningSource{
		Path:      "/test/001_PLANNING",
		PhaseName: "001_PLANNING",
		TopicDirs: []TopicDirectory{
			{Name: "api", Documents: []string{"endpoints.md"}, DocCount: 1},
			{Name: "database", Documents: []string{"schema.md"}, DocCount: 1},
		},
		TotalDocs: 2,
	}

	generator := NewGenerator("/test")
	plan, err := generator.Generate(context.Background(), source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Check sequences are numbered correctly
	for i, seq := range plan.Sequences {
		expectedNum := i + 1
		if seq.Number != expectedNum {
			t.Errorf("Sequence %d number = %d, want %d", i, seq.Number, expectedNum)
		}
		expectedFullName := fmt.Sprintf("%02d_%s", expectedNum, seq.Name)
		if seq.FullName != expectedFullName {
			t.Errorf("Sequence %d FullName = %q, want %q", i, seq.FullName, expectedFullName)
		}
	}
}

func TestGenerator_TaskNumbering(t *testing.T) {
	source := &PlanningSource{
		Path:      "/test/001_PLANNING",
		PhaseName: "001_PLANNING",
		TopicDirs: []TopicDirectory{
			{
				Name:      "requirements",
				Documents: []string{"auth.md", "api.md", "database.md"},
				DocCount:  3,
			},
		},
		TotalDocs: 3,
	}

	generator := NewGenerator("/test")
	plan, err := generator.Generate(context.Background(), source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	if len(plan.Sequences) == 0 {
		t.Fatal("Expected at least one sequence")
	}

	// Check tasks are numbered correctly
	for i, task := range plan.Sequences[0].Tasks {
		expectedNum := i + 1
		if task.Number != expectedNum {
			t.Errorf("Task %d number = %d, want %d", i, task.Number, expectedNum)
		}
		if !strings.HasPrefix(task.FullName, fmt.Sprintf("%02d_", expectedNum)) {
			t.Errorf("Task %d FullName = %q, doesn't start with %02d_", i, task.FullName, expectedNum)
		}
	}
}
