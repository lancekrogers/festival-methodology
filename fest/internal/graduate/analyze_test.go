package graduate

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestAnalyzer_Analyze(t *testing.T) {
	// Create test planning phase
	tmpDir := t.TempDir()
	planningDir := filepath.Join(tmpDir, "001_PLANNING")

	// Create directory structure
	dirs := []string{
		filepath.Join(planningDir, "requirements"),
		filepath.Join(planningDir, "architecture"),
		filepath.Join(planningDir, "decisions"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create dir %s: %v", dir, err)
		}
	}

	// Create documents
	docs := map[string]string{
		filepath.Join(planningDir, "requirements", "auth.md"):      "# Authentication Requirements\n\nUser auth flow.",
		filepath.Join(planningDir, "requirements", "api.md"):       "# API Requirements\n\nREST API design.",
		filepath.Join(planningDir, "architecture", "overview.md"):  "# Architecture Overview\n\nMicroservices.",
		filepath.Join(planningDir, "decisions", "ADR-001-db.md"):   "# ADR-001: Use PostgreSQL\n\n## Status\nAccepted\n\n## Context\nWe need a reliable database.",
		filepath.Join(planningDir, "decisions", "ADR-002-auth.md"): "# ADR-002: Use JWT\n\n## Status\nProposed\n\n## Decision\nUse JWT for authentication.",
	}
	for path, content := range docs {
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to write %s: %v", path, err)
		}
	}

	// Create planning summary
	summaryContent := `# Planning Summary

**Primary Goal:** Build a secure API gateway

## Key Decisions
- Use PostgreSQL for persistence
- Use JWT for authentication
- Implement rate limiting

## Proposed Sequences
1. core_infrastructure
2. authentication
3. api_gateway
`
	if err := os.WriteFile(filepath.Join(planningDir, "PLANNING_SUMMARY.md"), []byte(summaryContent), 0644); err != nil {
		t.Fatalf("failed to write planning summary: %v", err)
	}

	// Run analyzer
	analyzer := NewAnalyzer(tmpDir)
	ctx := context.Background()

	source, err := analyzer.Analyze(ctx, planningDir)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	// Verify results
	if source.PhaseName != "001_PLANNING" {
		t.Errorf("PhaseName = %q, want %q", source.PhaseName, "001_PLANNING")
	}

	// Check topic directories (decisions is also a topic dir)
	if len(source.TopicDirs) != 3 {
		t.Errorf("TopicDirs count = %d, want 3", len(source.TopicDirs))
	}

	// Check total docs (5 docs across all topics: 2 requirements, 1 architecture, 2 decisions)
	if source.TotalDocs != 5 {
		t.Errorf("TotalDocs = %d, want 5", source.TotalDocs)
	}

	// Check decisions
	if len(source.Decisions) != 2 {
		t.Errorf("Decisions count = %d, want 2", len(source.Decisions))
	} else {
		// Verify first decision
		found := false
		for _, d := range source.Decisions {
			if d.ID == "ADR-001" {
				found = true
				if d.Status != "accepted" {
					t.Errorf("ADR-001 status = %q, want %q", d.Status, "accepted")
				}
			}
		}
		if !found {
			t.Error("ADR-001 not found in decisions")
		}
	}

	// Check summary
	if source.Summary == nil {
		t.Fatal("Summary is nil")
	}
	if source.Summary.Goal != "Build a secure API gateway" {
		t.Errorf("Summary.Goal = %q, want %q", source.Summary.Goal, "Build a secure API gateway")
	}
	if len(source.Summary.KeyDecisions) != 3 {
		t.Errorf("Summary.KeyDecisions count = %d, want 3", len(source.Summary.KeyDecisions))
	}
	if len(source.Summary.ProposedSequences) != 3 {
		t.Errorf("Summary.ProposedSequences count = %d, want 3", len(source.Summary.ProposedSequences))
	}
}

func TestAnalyzer_Analyze_NotFound(t *testing.T) {
	analyzer := NewAnalyzer("/tmp")
	ctx := context.Background()

	_, err := analyzer.Analyze(ctx, "/nonexistent/path")
	if err == nil {
		t.Error("Analyze() expected error for nonexistent path")
	}
}

func TestAnalyzer_Analyze_NotDirectory(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(tmpFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewAnalyzer("/tmp")
	ctx := context.Background()

	_, err := analyzer.Analyze(ctx, tmpFile)
	if err == nil {
		t.Error("Analyze() expected error for file path")
	}
}

func TestAnalyzer_Analyze_EmptyPhase(t *testing.T) {
	tmpDir := t.TempDir()
	planningDir := filepath.Join(tmpDir, "001_PLANNING")
	if err := os.MkdirAll(planningDir, 0755); err != nil {
		t.Fatal(err)
	}

	analyzer := NewAnalyzer(tmpDir)
	ctx := context.Background()

	source, err := analyzer.Analyze(ctx, planningDir)
	if err != nil {
		t.Fatalf("Analyze() error = %v", err)
	}

	if len(source.TopicDirs) != 0 {
		t.Errorf("TopicDirs count = %d, want 0", len(source.TopicDirs))
	}
	if source.TotalDocs != 0 {
		t.Errorf("TotalDocs = %d, want 0", source.TotalDocs)
	}
	if source.Summary != nil {
		t.Error("Summary should be nil for empty phase")
	}
}

func TestAnalyzer_Analyze_ContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	analyzer := NewAnalyzer("/tmp")
	_, err := analyzer.Analyze(ctx, "/tmp")
	if err == nil {
		t.Error("Analyze() expected error for canceled context")
	}
}

func TestAnalyzer_FindPlanningPhase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phases
	phases := []string{
		"001_PLANNING",
		"002_IMPLEMENTATION",
		"003_REVIEW",
	}
	for _, p := range phases {
		if err := os.MkdirAll(filepath.Join(tmpDir, p), 0755); err != nil {
			t.Fatal(err)
		}
	}

	analyzer := NewAnalyzer(tmpDir)
	ctx := context.Background()

	path, err := analyzer.FindPlanningPhase(ctx)
	if err != nil {
		t.Fatalf("FindPlanningPhase() error = %v", err)
	}

	if filepath.Base(path) != "001_PLANNING" {
		t.Errorf("FindPlanningPhase() = %q, want planning phase", path)
	}
}

func TestAnalyzer_FindPlanningPhase_Design(t *testing.T) {
	tmpDir := t.TempDir()

	// Create design phase instead of planning
	if err := os.MkdirAll(filepath.Join(tmpDir, "001_DESIGN"), 0755); err != nil {
		t.Fatal(err)
	}

	analyzer := NewAnalyzer(tmpDir)
	ctx := context.Background()

	path, err := analyzer.FindPlanningPhase(ctx)
	if err != nil {
		t.Fatalf("FindPlanningPhase() error = %v", err)
	}

	if filepath.Base(path) != "001_DESIGN" {
		t.Errorf("FindPlanningPhase() = %q, want design phase", path)
	}
}

func TestAnalyzer_FindPlanningPhase_NotFound(t *testing.T) {
	tmpDir := t.TempDir()

	// Create only implementation phases
	if err := os.MkdirAll(filepath.Join(tmpDir, "001_IMPLEMENTATION"), 0755); err != nil {
		t.Fatal(err)
	}

	analyzer := NewAnalyzer(tmpDir)
	ctx := context.Background()

	_, err := analyzer.FindPlanningPhase(ctx)
	if err == nil {
		t.Error("FindPlanningPhase() expected error when no planning phase")
	}
}

func TestIsGoalFile(t *testing.T) {
	tests := []struct {
		filename string
		want     bool
	}{
		{"PHASE_GOAL.md", true},
		{"SEQUENCE_GOAL.md", true},
		{"FESTIVAL_GOAL.md", true},
		{"phase_goal.md", true},           // Case insensitive
		{"sequence_goal.md", true},        // Case insensitive
		{"MY_CUSTOM_GOAL.md", true},       // Pattern: *_GOAL.md
		{"requirements.md", false},        // Regular document
		{"auth_design.md", false},         // Regular document
		{"PLANNING_SUMMARY.md", false},    // Not a goal file
		{"GOAL_setup.md", false},          // GOAL at start, not end
		{"database_migration.md", false},  // Regular document
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			if got := isGoalFile(tt.filename); got != tt.want {
				t.Errorf("isGoalFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFindDocuments_FiltersGoalFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create mix of documents and goal files
	files := []string{
		"auth_requirements.md",      // Should include
		"api_design.md",             // Should include
		"SEQUENCE_GOAL.md",          // Should exclude
		"PHASE_GOAL.md",             // Should exclude
		"CUSTOM_GOAL.md",            // Should exclude (matches *_GOAL.md)
		"database_schema.md",        // Should include
	}
	for _, f := range files {
		if err := os.WriteFile(filepath.Join(tmpDir, f), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	analyzer := NewAnalyzer(tmpDir)
	docs, err := analyzer.findDocuments(tmpDir)
	if err != nil {
		t.Fatalf("findDocuments() error = %v", err)
	}

	// Should have 3 documents (excluding 3 goal files)
	if len(docs) != 3 {
		t.Errorf("findDocuments() returned %d docs, want 3 (excluding goal files)", len(docs))
		t.Logf("Got docs: %v", docs)
	}

	// Verify no goal files included
	for _, doc := range docs {
		if isGoalFile(doc) {
			t.Errorf("findDocuments() included goal file: %s", doc)
		}
	}
}

func TestParseDecisionFile(t *testing.T) {
	tmpDir := t.TempDir()

	adrContent := `# ADR-003: Use Kubernetes

## Status
Accepted

## Context
We need container orchestration for production.

## Decision
Use Kubernetes for container orchestration.

## Consequences
- Need K8s expertise
- Higher complexity
`
	adrPath := filepath.Join(tmpDir, "ADR-003-k8s.md")
	if err := os.WriteFile(adrPath, []byte(adrContent), 0644); err != nil {
		t.Fatal(err)
	}

	analyzer := NewAnalyzer(tmpDir)
	decision, err := analyzer.parseDecisionFile(adrPath)
	if err != nil {
		t.Fatalf("parseDecisionFile() error = %v", err)
	}

	if decision.ID != "ADR-003" {
		t.Errorf("ID = %q, want %q", decision.ID, "ADR-003")
	}
	if decision.Title != "ADR-003: Use Kubernetes" {
		t.Errorf("Title = %q, want %q", decision.Title, "ADR-003: Use Kubernetes")
	}
	if decision.Status != "accepted" {
		t.Errorf("Status = %q, want %q", decision.Status, "accepted")
	}
	if decision.Summary != "We need container orchestration for production." {
		t.Errorf("Summary = %q, want context line", decision.Summary)
	}
}
