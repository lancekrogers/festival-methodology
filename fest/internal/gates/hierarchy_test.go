package gates

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestPolicyLevel_Constants(t *testing.T) {
	// Verify all policy levels are defined
	levels := []PolicyLevel{
		PolicyLevelBuiltin,
		PolicyLevelGlobal,
		PolicyLevelFestival,
		PolicyLevelPhase,
		PolicyLevelSequence,
	}

	expected := []string{"builtin", "global", "festival", "phase", "sequence"}
	for i, level := range levels {
		if string(level) != expected[i] {
			t.Errorf("PolicyLevel %d = %q, want %q", i, level, expected[i])
		}
	}
}

func TestPolicySource(t *testing.T) {
	source := PolicySource{
		Level: PolicyLevelPhase,
		Path:  "/path/to/policy.yml",
		Name:  "custom",
	}

	if source.Level != PolicyLevelPhase {
		t.Errorf("PolicySource.Level = %q, want %q", source.Level, PolicyLevelPhase)
	}
	if source.Path != "/path/to/policy.yml" {
		t.Errorf("PolicySource.Path = %q, want %q", source.Path, "/path/to/policy.yml")
	}
	if source.Name != "custom" {
		t.Errorf("PolicySource.Name = %q, want %q", source.Name, "custom")
	}
}

func TestGateTask_NewFields(t *testing.T) {
	source := &PolicySource{Level: PolicyLevelGlobal}
	task := GateTask{
		ID:       "test",
		Template: "TEST",
		Source:   source,
		Removed:  true,
	}

	if task.Source != source {
		t.Error("GateTask.Source not set correctly")
	}
	if !task.Removed {
		t.Error("GateTask.Removed not set correctly")
	}
}

func TestGatePolicy_ShouldInherit(t *testing.T) {
	tests := []struct {
		name     string
		inherit  *bool
		expected bool
	}{
		{"nil inherits", nil, true},
		{"true inherits", boolPtr(true), true},
		{"false does not inherit", boolPtr(false), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &GatePolicy{Inherit: tt.inherit}
			if got := policy.ShouldInherit(); got != tt.expected {
				t.Errorf("ShouldInherit() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewHierarchicalLoader(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hierarchy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test with valid path
	loader, err := NewHierarchicalLoader(tmpDir, nil)
	if err != nil {
		t.Fatalf("NewHierarchicalLoader error: %v", err)
	}
	if loader.FestivalsRoot() != tmpDir {
		t.Errorf("FestivalsRoot = %q, want %q", loader.FestivalsRoot(), tmpDir)
	}

	// Test with empty path
	_, err = NewHierarchicalLoader("", nil)
	if err == nil {
		t.Error("NewHierarchicalLoader with empty path should error")
	}

	// Test with nonexistent path
	_, err = NewHierarchicalLoader("/nonexistent/path", nil)
	if err == nil {
		t.Error("NewHierarchicalLoader with nonexistent path should error")
	}
}

func TestHierarchicalLoader_LoadForFestival(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hierarchy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create festival directory
	festivalPath := filepath.Join(tmpDir, "test-festival")
	os.MkdirAll(festivalPath, 0755)

	loader, err := NewHierarchicalLoader(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	effective, err := loader.LoadForFestival(ctx, festivalPath)
	if err != nil {
		t.Fatalf("LoadForFestival error: %v", err)
	}

	// Should have default gates
	if len(effective.Gates) != 4 {
		t.Errorf("Expected 4 default gates, got %d", len(effective.Gates))
	}
	if effective.Level != PolicyLevelBuiltin {
		t.Errorf("Expected level %q, got %q", PolicyLevelBuiltin, effective.Level)
	}
}

func TestHierarchicalLoader_ContextCancellation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "hierarchy-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	loader, err := NewHierarchicalLoader(tmpDir, nil)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err = loader.LoadForFestival(ctx, tmpDir)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestEffectivePolicy_GetActiveGates(t *testing.T) {
	effective := &EffectivePolicy{
		Gates: []GateTask{
			{ID: "active1", Removed: false},
			{ID: "removed", Removed: true},
			{ID: "active2", Removed: false},
		},
	}

	active := effective.GetActiveGates()
	if len(active) != 2 {
		t.Errorf("GetActiveGates returned %d gates, want 2", len(active))
	}
	if active[0].ID != "active1" || active[1].ID != "active2" {
		t.Error("GetActiveGates returned wrong gates")
	}
}

func boolPtr(b bool) *bool {
	return &b
}
