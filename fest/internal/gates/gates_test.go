package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultPolicy(t *testing.T) {
	policy := DefaultPolicy()

	if policy.Version != 1 {
		t.Errorf("DefaultPolicy version = %d, want 1", policy.Version)
	}
	if policy.Name != DefaultPolicyName {
		t.Errorf("DefaultPolicy name = %q, want %q", policy.Name, DefaultPolicyName)
	}
	if len(policy.Append) != 4 {
		t.Errorf("DefaultPolicy has %d tasks, want 4", len(policy.Append))
	}

	// Check default task IDs
	expectedIDs := []string{"testing_and_verify", "code_review", "review_results_iterate", "commit"}
	for i, expected := range expectedIDs {
		if policy.Append[i].ID != expected {
			t.Errorf("DefaultPolicy task %d ID = %q, want %q", i, policy.Append[i].ID, expected)
		}
	}
}

func TestGetEnabledTasks(t *testing.T) {
	policy := &GatePolicy{
		Append: []GateTask{
			{ID: "task1", Enabled: true},
			{ID: "task2", Enabled: false},
			{ID: "task3", Enabled: true},
		},
	}

	enabled := policy.GetEnabledTasks()

	if len(enabled) != 2 {
		t.Errorf("GetEnabledTasks returned %d tasks, want 2", len(enabled))
	}
	if enabled[0].ID != "task1" || enabled[1].ID != "task3" {
		t.Error("GetEnabledTasks returned wrong tasks")
	}
}

func TestPolicyClone(t *testing.T) {
	original := DefaultPolicy()
	original.Append[0].Customizations = map[string]any{"key": "value"}

	clone := original.Clone()

	// Modify clone
	clone.Name = "modified"
	clone.Append[0].ID = "modified_id"
	clone.Append[0].Customizations["key"] = "modified"

	// Verify original unchanged
	if original.Name == "modified" {
		t.Error("Clone modified original policy name")
	}
	if original.Append[0].ID == "modified_id" {
		t.Error("Clone modified original task ID")
	}
	if original.Append[0].Customizations["key"] == "modified" {
		t.Error("Clone modified original customizations")
	}
}

func TestApplyPhaseOverride_Add(t *testing.T) {
	policy := &GatePolicy{
		Append: []GateTask{
			{ID: "task1", Enabled: true},
			{ID: "task2", Enabled: true},
		},
	}

	override := &PhaseOverride{
		Ops: []GateOperation{
			{
				Add: &GateAddOp{
					Task:  GateTask{ID: "new_task", Enabled: true},
					After: "task1",
				},
			},
		},
	}

	result := ApplyPhaseOverride(policy, override)

	if len(result.Append) != 3 {
		t.Errorf("ApplyPhaseOverride resulted in %d tasks, want 3", len(result.Append))
	}
	if result.Append[1].ID != "new_task" {
		t.Errorf("New task not inserted after task1, got %q at index 1", result.Append[1].ID)
	}
}

func TestApplyPhaseOverride_Remove(t *testing.T) {
	policy := &GatePolicy{
		Append: []GateTask{
			{ID: "task1", Enabled: true},
			{ID: "task2", Enabled: true},
			{ID: "task3", Enabled: true},
		},
	}

	override := &PhaseOverride{
		Ops: []GateOperation{
			{
				Remove: &GateRemoveOp{ID: "task2"},
			},
		},
	}

	result := ApplyPhaseOverride(policy, override)

	if len(result.Append) != 2 {
		t.Errorf("ApplyPhaseOverride resulted in %d tasks, want 2", len(result.Append))
	}
	for _, task := range result.Append {
		if task.ID == "task2" {
			t.Error("task2 was not removed")
		}
	}
}

func TestApplyPhaseOverride_Nil(t *testing.T) {
	policy := DefaultPolicy()
	original := len(policy.Append)

	result := ApplyPhaseOverride(policy, nil)

	if len(result.Append) != original {
		t.Error("ApplyPhaseOverride with nil override modified policy")
	}
}

func TestLoadSavePolicy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gates-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	policy := &GatePolicy{
		Version: 1,
		Name:    "test",
		Append: []GateTask{
			{ID: "test_task", Template: "TEST_TEMPLATE", Enabled: true},
		},
		ExcludePatterns: []string{"*_docs"},
	}

	policyPath := filepath.Join(tmpDir, "test.yml")

	// Save
	if err := SavePolicy(policyPath, policy); err != nil {
		t.Fatalf("SavePolicy error: %v", err)
	}

	// Load
	loaded, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadPolicy error: %v", err)
	}

	if loaded.Name != policy.Name {
		t.Errorf("Loaded policy name = %q, want %q", loaded.Name, policy.Name)
	}
	if len(loaded.Append) != 1 {
		t.Errorf("Loaded policy has %d tasks, want 1", len(loaded.Append))
	}
	if loaded.Append[0].ID != "test_task" {
		t.Errorf("Loaded task ID = %q, want %q", loaded.Append[0].ID, "test_task")
	}
}

func TestLoadPhaseOverride(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gates-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test no override file
	override, err := LoadPhaseOverride(tmpDir)
	if err != nil {
		t.Errorf("LoadPhaseOverride with no file: unexpected error %v", err)
	}
	if override != nil {
		t.Error("LoadPhaseOverride with no file: expected nil")
	}

	// Create override file
	overrideContent := `ops:
  - add:
      task:
        id: security_review
        template: SECURITY_REVIEW
        enabled: true
      after: code_review
`
	overridePath := filepath.Join(tmpDir, PhaseOverrideFileName)
	if err := os.WriteFile(overridePath, []byte(overrideContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Load
	override, err = LoadPhaseOverride(tmpDir)
	if err != nil {
		t.Fatalf("LoadPhaseOverride error: %v", err)
	}
	if override == nil {
		t.Fatal("LoadPhaseOverride returned nil")
	}
	if len(override.Ops) != 1 {
		t.Errorf("LoadPhaseOverride has %d ops, want 1", len(override.Ops))
	}
	if override.Ops[0].Add == nil {
		t.Error("LoadPhaseOverride first op is not an add")
	}
	if override.Ops[0].Add.Task.ID != "security_review" {
		t.Errorf("Override task ID = %q, want %q", override.Ops[0].Add.Task.ID, "security_review")
	}
}

func TestAddMarkers(t *testing.T) {
	content := "# Task Title\n\nSome content here."
	result := AddMarkers(content, "test_gate")

	if result[:3] != "---" {
		t.Error("AddMarkers did not add frontmatter delimiter")
	}
	if result[len(result)-len(content):] != content {
		t.Error("AddMarkers modified original content")
	}
}

func TestAddMarkersToExistingFrontmatter(t *testing.T) {
	content := `---
template_id: test
---

# Task Title
`
	result := AddMarkers(content, "test_gate")

	// Should contain both original and new markers
	if result[:3] != "---" {
		t.Error("AddMarkers removed frontmatter")
	}
}

func TestIsManaged(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "markers-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Test unmanaged file
	unmanagedContent := "# Just content"
	unmanagedPath := filepath.Join(tmpDir, "unmanaged.md")
	os.WriteFile(unmanagedPath, []byte(unmanagedContent), 0644)

	if IsManaged(unmanagedPath) {
		t.Error("IsManaged returned true for unmanaged file")
	}

	// Test managed file
	managedContent := `---
fest_managed: true
fest_gate_id: test
---

# Content
`
	managedPath := filepath.Join(tmpDir, "managed.md")
	os.WriteFile(managedPath, []byte(managedContent), 0644)

	if !IsManaged(managedPath) {
		t.Error("IsManaged returned false for managed file")
	}
}

func TestGetGateID(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "markers-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	content := `---
fest_managed: true
fest_gate_id: my_gate_id
---

# Content
`
	filePath := filepath.Join(tmpDir, "test.md")
	os.WriteFile(filePath, []byte(content), 0644)

	gateID := GetGateID(filePath)
	if gateID != "my_gate_id" {
		t.Errorf("GetGateID = %q, want %q", gateID, "my_gate_id")
	}
}

func TestStripMarkers(t *testing.T) {
	content := `---
fest_managed: true
fest_gate_id: test
other_field: keep
---

# Content
`
	result := StripMarkers(content)

	// Should still have frontmatter with other_field
	if result[:3] != "---" {
		t.Error("StripMarkers removed all frontmatter")
	}
	if !contains(result, "other_field") {
		t.Error("StripMarkers removed non-marker fields")
	}
	if contains(result, "fest_managed") {
		t.Error("StripMarkers did not remove fest_managed")
	}
	if contains(result, "fest_gate_id") {
		t.Error("StripMarkers did not remove fest_gate_id")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestDetectPhaseType(t *testing.T) {
	tests := []struct {
		phaseName string
		expected  string
	}{
		// Planning phases
		{"001_PLANNING", "planning"},
		{"002_Plan", "planning"},
		{"planning_phase", "planning"},
		// Research phases
		{"001_RESEARCH", "research"},
		{"002_Discovery", "research"},
		{"003_DESIGN", "research"},
		// Implementation phases
		{"001_IMPLEMENTATION", "implementation"},
		{"002_Implement", "implementation"},
		{"003_DEVELOP", "implementation"},
		{"004_Build", "implementation"},
		{"005_FOUNDATION", "implementation"},
		{"006_CRITICAL_BUGS", "implementation"},
		// Review phases
		{"001_REVIEW", "review"},
		{"002_QA", "review"},
		{"003_UAT", "review"},
		// Deployment phases
		{"001_DEPLOYMENT", "deployment"},
		{"002_Deploy", "deployment"},
		{"003_Release", "deployment"},
		// Unknown defaults to implementation
		{"001_UNKNOWN", "implementation"},
		{"random_name", "implementation"},
	}

	for _, tc := range tests {
		t.Run(tc.phaseName, func(t *testing.T) {
			result := DetectPhaseType(tc.phaseName)
			if result != tc.expected {
				t.Errorf("DetectPhaseType(%q) = %q, want %q", tc.phaseName, result, tc.expected)
			}
		})
	}
}

func TestGetGatesForPhaseType(t *testing.T) {
	tests := []struct {
		phaseType   string
		expectedLen int
		expectedID  string // First gate ID to verify
	}{
		{"implementation", 4, "testing_and_verify"},
		{"planning", 3, "planning_review"},
		{"research", 3, "research_review"},
		{"review", 2, "review_checklist"},
		{"deployment", 0, ""},                // Deployment has no gates
		{"unknown", 4, "testing_and_verify"}, // Unknown defaults to implementation
	}

	for _, tc := range tests {
		t.Run(tc.phaseType, func(t *testing.T) {
			gates := GetGatesForPhaseType(tc.phaseType)
			if len(gates) != tc.expectedLen {
				t.Errorf("GetGatesForPhaseType(%q) returned %d gates, want %d", tc.phaseType, len(gates), tc.expectedLen)
			}
			if tc.expectedLen > 0 && gates[0].ID != tc.expectedID {
				t.Errorf("GetGatesForPhaseType(%q) first gate ID = %q, want %q", tc.phaseType, gates[0].ID, tc.expectedID)
			}
		})
	}
}

func TestImplementationGates(t *testing.T) {
	gates := ImplementationGates()
	if len(gates) != 4 {
		t.Errorf("ImplementationGates() returned %d gates, want 4", len(gates))
	}
	expectedIDs := []string{"testing_and_verify", "code_review", "review_results_iterate", "commit"}
	for i, expected := range expectedIDs {
		if gates[i].ID != expected {
			t.Errorf("ImplementationGates()[%d].ID = %q, want %q", i, gates[i].ID, expected)
		}
	}
}

func TestPlanningGates(t *testing.T) {
	gates := PlanningGates()
	if len(gates) != 3 {
		t.Errorf("PlanningGates() returned %d gates, want 3", len(gates))
	}
	expectedIDs := []string{"planning_review", "decision_validation", "planning_summary"}
	for i, expected := range expectedIDs {
		if gates[i].ID != expected {
			t.Errorf("PlanningGates()[%d].ID = %q, want %q", i, gates[i].ID, expected)
		}
	}
}

func TestResearchGates(t *testing.T) {
	gates := ResearchGates()
	if len(gates) != 3 {
		t.Errorf("ResearchGates() returned %d gates, want 3", len(gates))
	}
	expectedIDs := []string{"research_review", "findings_synthesis", "research_summary"}
	for i, expected := range expectedIDs {
		if gates[i].ID != expected {
			t.Errorf("ResearchGates()[%d].ID = %q, want %q", i, gates[i].ID, expected)
		}
	}
}

func TestReviewGates(t *testing.T) {
	gates := ReviewGates()
	if len(gates) != 2 {
		t.Errorf("ReviewGates() returned %d gates, want 2", len(gates))
	}
	expectedIDs := []string{"review_checklist", "signoff"}
	for i, expected := range expectedIDs {
		if gates[i].ID != expected {
			t.Errorf("ReviewGates()[%d].ID = %q, want %q", i, gates[i].ID, expected)
		}
	}
}
