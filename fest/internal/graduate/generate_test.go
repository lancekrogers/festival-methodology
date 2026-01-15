package graduate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
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
		{"setup_database", 0},              // "setup" is first
		{"init_config", 1},                 // "init" is second
		{"database_schema", 3},             // "database" is in list
		{"random_task", len(TaskPriority)}, // Not in list, returns len(TaskPriority)
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

func TestScanExistingPhases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create some phase directories
	phases := []string{
		"001_PLANNING",
		"002_IMPLEMENTATION",
		"004_REVIEW",
		".hidden",   // Should be skipped
		"not_phase", // Should be skipped (no number)
	}
	for _, p := range phases {
		if err := os.MkdirAll(filepath.Join(tmpDir, p), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", p, err)
		}
	}

	// Also create a file (should be skipped)
	if err := os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	generator := NewGenerator(tmpDir)
	existing := generator.scanExistingPhases()

	// Should find 3 phases (001, 002, 004)
	if len(existing) != 3 {
		t.Errorf("scanExistingPhases() found %d phases, want 3", len(existing))
	}

	// Check specific phases
	if existing[1] != "001_PLANNING" {
		t.Errorf("Phase 1 = %q, want %q", existing[1], "001_PLANNING")
	}
	if existing[2] != "002_IMPLEMENTATION" {
		t.Errorf("Phase 2 = %q, want %q", existing[2], "002_IMPLEMENTATION")
	}
	if existing[4] != "004_REVIEW" {
		t.Errorf("Phase 4 = %q, want %q", existing[4], "004_REVIEW")
	}

	// Verify hidden and non-numbered dirs are skipped
	if _, exists := existing[0]; exists {
		t.Error("Phase 0 should not exist (hidden or non-numbered)")
	}
}

func TestFindNextAvailablePhaseNumber(t *testing.T) {
	generator := NewGenerator("/test")

	tests := []struct {
		name     string
		existing map[int]string
		start    int
		expected int
	}{
		{
			name:     "no collision",
			existing: map[int]string{1: "001_PLANNING"},
			start:    2,
			expected: 2,
		},
		{
			name:     "simple collision",
			existing: map[int]string{1: "001_PLANNING", 2: "002_IMPLEMENTATION"},
			start:    2,
			expected: 3,
		},
		{
			name:     "multiple consecutive collisions",
			existing: map[int]string{1: "001", 2: "002", 3: "003", 4: "004"},
			start:    2,
			expected: 5,
		},
		{
			name:     "gap in numbers",
			existing: map[int]string{1: "001", 2: "002", 4: "004"},
			start:    2,
			expected: 3, // Should find the gap at 3
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generator.findNextAvailablePhaseNumber(tt.existing, tt.start)
			if result != tt.expected {
				t.Errorf("findNextAvailablePhaseNumber() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGenerator_Generate_PhaseCollision(t *testing.T) {
	tmpDir := t.TempDir()

	// Create existing phases that cause collision
	phases := []string{
		"001_PLANNING",
		"002_EXISTING_PHASE", // This will collide with target
	}
	for _, p := range phases {
		if err := os.MkdirAll(filepath.Join(tmpDir, p), 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", p, err)
		}
	}

	source := &PlanningSource{
		Path:       filepath.Join(tmpDir, "001_PLANNING"),
		PhaseName:  "001_PLANNING",
		TopicDirs:  []TopicDirectory{},
		TotalDocs:  0,
		AnalyzedAt: time.Now(),
	}

	generator := NewGenerator(tmpDir)
	plan, err := generator.Generate(context.Background(), source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Target should be adjusted to 003
	if plan.Target.Number != 3 {
		t.Errorf("Target.Number = %d, want 3 (avoiding collision)", plan.Target.Number)
	}
	if plan.Target.PhaseName != "003_IMPLEMENTATION" {
		t.Errorf("Target.PhaseName = %q, want %q", plan.Target.PhaseName, "003_IMPLEMENTATION")
	}

	// Collision should be detected
	if !plan.Target.CollisionDetected {
		t.Error("CollisionDetected = false, want true")
	}
	if plan.Target.OriginalNumber != 2 {
		t.Errorf("OriginalNumber = %d, want 2", plan.Target.OriginalNumber)
	}
	if plan.Target.ExistingPhase != "002_EXISTING_PHASE" {
		t.Errorf("ExistingPhase = %q, want %q", plan.Target.ExistingPhase, "002_EXISTING_PHASE")
	}

	// Warning should be present
	hasCollisionWarning := false
	for _, w := range plan.Warnings {
		if strings.Contains(w, "already exists") && strings.Contains(w, "002") {
			hasCollisionWarning = true
			break
		}
	}
	if !hasCollisionWarning {
		t.Errorf("Expected collision warning in Warnings: %v", plan.Warnings)
	}
}

func TestGenerator_Generate_NoCollision(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only planning phase
	if err := os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755); err != nil {
		t.Fatal(err)
	}

	source := &PlanningSource{
		Path:       filepath.Join(tmpDir, "001_PLANNING"),
		PhaseName:  "001_PLANNING",
		TopicDirs:  []TopicDirectory{},
		TotalDocs:  0,
		AnalyzedAt: time.Now(),
	}

	generator := NewGenerator(tmpDir)
	plan, err := generator.Generate(context.Background(), source)
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}

	// Target should be 002 with no collision
	if plan.Target.Number != 2 {
		t.Errorf("Target.Number = %d, want 2", plan.Target.Number)
	}
	if plan.Target.CollisionDetected {
		t.Error("CollisionDetected = true, want false (no collision)")
	}
}

func TestExtractObjectiveFromDoc(t *testing.T) {
	tmpDir := t.TempDir()
	generator := NewGenerator(tmpDir)

	tests := []struct {
		name     string
		content  string
		taskName string
		expected string
	}{
		{
			name: "objective_section",
			content: `# Auth Requirements

## Objective

Implement user authentication with OAuth2 support.

## Details
More content here.
`,
			taskName: "auth_requirements",
			expected: "Implement user authentication with OAuth2 support.",
		},
		{
			name: "goal_section",
			content: `# API Design

## Goal

Build a RESTful API for user management.
`,
			taskName: "api_design",
			expected: "Build a RESTful API for user management.",
		},
		{
			name: "inline_objective",
			content: `# Database Schema

**Objective:** Create database schema for users and sessions.

More details follow.
`,
			taskName: "database_schema",
			expected: "Create database schema for users and sessions.",
		},
		{
			name: "title_with_colon",
			content: `# Authentication: Implement JWT token validation

Some content.
`,
			taskName: "auth",
			expected: "Implement JWT token validation",
		},
		{
			name: "descriptive_title",
			content: `# User Registration Flow Implementation

Detailed planning for user registration.
`,
			taskName: "registration",
			expected: "Implement user registration flow implementation",
		},
		{
			name: "fallback_generic",
			content: `# Short

No useful content.
`,
			taskName: "short_doc",
			expected: "Implement short_doc as specified in planning",
		},
		{
			name: "requirements_bullet",
			content: `# Feature X

## Requirements

- Support multiple authentication providers
- Handle token refresh automatically
`,
			taskName: "feature_x",
			expected: "Support multiple authentication providers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with content
			docPath := filepath.Join(tmpDir, tt.name+".md")
			if err := os.WriteFile(docPath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			result := generator.extractObjectiveFromDoc(docPath, tt.taskName)
			if result != tt.expected {
				t.Errorf("extractObjectiveFromDoc() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestCleanObjective(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"bold_text", "**bold text**", "Bold text"},
		{"code", "`code`", "Code"},
		{"spaces", "  spaces  ", "Spaces"},
		{"lowercase", "lowercase start", "Lowercase start"},
		{"truncate", strings.Repeat("x", 200), "X" + strings.Repeat("x", 146) + "..."}, // First x gets capitalized
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cleanObjective(tt.input)
			if result != tt.expected {
				t.Errorf("cleanObjective(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGenerateTasks_WithObjectiveExtraction(t *testing.T) {
	tmpDir := t.TempDir()

	// Create topic directory with documents
	topicDir := filepath.Join(tmpDir, "requirements")
	if err := os.MkdirAll(topicDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create document with extractable objective
	docContent := `# Authentication

## Objective

Implement secure user authentication with multi-factor support.
`
	if err := os.WriteFile(filepath.Join(topicDir, "auth.md"), []byte(docContent), 0644); err != nil {
		t.Fatal(err)
	}

	generator := NewGenerator(tmpDir)
	topic := TopicDirectory{
		Name:      "requirements",
		Path:      topicDir,
		Documents: []string{"auth.md"},
		DocCount:  1,
	}

	tasks := generator.generateTasks(topic)

	if len(tasks) != 1 {
		t.Fatalf("generateTasks() returned %d tasks, want 1", len(tasks))
	}

	// Should have extracted the objective, not generic text
	expected := "Implement secure user authentication with multi-factor support."
	if tasks[0].Objective != expected {
		t.Errorf("Task objective = %q, want %q", tasks[0].Objective, expected)
	}
}
