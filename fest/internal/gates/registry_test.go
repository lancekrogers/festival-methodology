package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewPolicyRegistry(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry, err := NewPolicyRegistry(tmpDir, "")
	if err != nil {
		t.Fatalf("NewPolicyRegistry error: %v", err)
	}

	// Should have built-in policies
	names := registry.List()
	if len(names) != 3 {
		t.Errorf("Expected 3 built-in policies, got %d", len(names))
	}

	// Check specific policies
	for _, name := range []string{"default", "strict", "lightweight"} {
		info, ok := registry.Get(name)
		if !ok {
			t.Errorf("Built-in policy %q not found", name)
		}
		if info.Source != "built-in" {
			t.Errorf("Policy %q source = %q, want %q", name, info.Source, "built-in")
		}
	}
}

func TestPolicyRegistry_GetPolicy(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry, _ := NewPolicyRegistry(tmpDir, "")

	// Test default policy
	policy, err := registry.GetPolicy("default")
	if err != nil {
		t.Fatalf("GetPolicy(default) error: %v", err)
	}
	if policy.Name != "default" {
		t.Errorf("Policy name = %q, want %q", policy.Name, "default")
	}
	if len(policy.Append) != 3 {
		t.Errorf("Default policy has %d gates, want 3", len(policy.Append))
	}

	// Test strict policy
	policy, err = registry.GetPolicy("strict")
	if err != nil {
		t.Fatalf("GetPolicy(strict) error: %v", err)
	}
	if len(policy.Append) != 5 {
		t.Errorf("Strict policy has %d gates, want 5", len(policy.Append))
	}

	// Test lightweight policy
	policy, err = registry.GetPolicy("lightweight")
	if err != nil {
		t.Fatalf("GetPolicy(lightweight) error: %v", err)
	}
	if len(policy.Append) != 1 {
		t.Errorf("Lightweight policy has %d gates, want 1", len(policy.Append))
	}

	// Test nonexistent policy
	_, err = registry.GetPolicy("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent policy")
	}
}

func TestPolicyRegistry_ScanDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create policy directory
	policyDir := filepath.Join(tmpDir, ".festival", "gates", "policies")
	os.MkdirAll(policyDir, 0755)

	// Create custom policy file
	customPolicy := `version: 1
name: custom
description: Custom test policy
append:
  - id: custom_gate
    template: CUSTOM
    enabled: true
`
	os.WriteFile(filepath.Join(policyDir, "custom.yml"), []byte(customPolicy), 0644)

	registry, _ := NewPolicyRegistry(tmpDir, "")

	// Should have custom policy
	info, ok := registry.Get("custom")
	if !ok {
		t.Fatal("Custom policy not found")
	}
	if info.Source != "global" {
		t.Errorf("Custom policy source = %q, want %q", info.Source, "global")
	}
	if info.Description != "Custom test policy" {
		t.Errorf("Description = %q, want %q", info.Description, "Custom test policy")
	}

	// Verify we can load it
	policy, err := registry.GetPolicy("custom")
	if err != nil {
		t.Fatalf("GetPolicy(custom) error: %v", err)
	}
	if len(policy.Append) != 1 {
		t.Errorf("Custom policy has %d gates, want 1", len(policy.Append))
	}
}

func TestPolicyRegistry_ListInfo(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry, _ := NewPolicyRegistry(tmpDir, "")

	infos := registry.ListInfo()
	if len(infos) != 3 {
		t.Errorf("ListInfo returned %d policies, want 3", len(infos))
	}

	// Check sorting
	if infos[0].Name != "default" || infos[1].Name != "lightweight" || infos[2].Name != "strict" {
		t.Error("ListInfo not sorted alphabetically")
	}
}

func TestPolicyRegistry_Register(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	registry, _ := NewPolicyRegistry(tmpDir, "")

	// Register new policy
	registry.Register("test", &PolicyInfo{
		Name:        "test",
		Description: "Test policy",
		Source:      "test",
	})

	info, ok := registry.Get("test")
	if !ok {
		t.Fatal("Registered policy not found")
	}
	if info.Description != "Test policy" {
		t.Errorf("Description = %q, want %q", info.Description, "Test policy")
	}
}

func TestPolicyRegistry_Refresh(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "registry-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create policy directory
	policyDir := filepath.Join(tmpDir, ".festival", "gates", "policies")
	os.MkdirAll(policyDir, 0755)

	registry, _ := NewPolicyRegistry(tmpDir, "")

	// Initially 3 built-in policies
	if len(registry.List()) != 3 {
		t.Fatalf("Expected 3 policies, got %d", len(registry.List()))
	}

	// Add a policy file
	customPolicy := `version: 1
name: new
description: New policy
append: []
`
	os.WriteFile(filepath.Join(policyDir, "new.yml"), []byte(customPolicy), 0644)

	// Refresh
	registry.Refresh()

	// Should now have 4 policies
	if len(registry.List()) != 4 {
		t.Errorf("After refresh, expected 4 policies, got %d", len(registry.List()))
	}
}

func TestStrictPolicy(t *testing.T) {
	policy := StrictPolicy()

	if policy.Name != "strict" {
		t.Errorf("Name = %q, want %q", policy.Name, "strict")
	}
	if len(policy.Append) != 5 {
		t.Errorf("Strict policy has %d gates, want 5", len(policy.Append))
	}

	// Check gate IDs
	expectedIDs := []string{
		"testing_and_verify",
		"code_review",
		"security_audit",
		"performance_check",
		"review_results_iterate",
	}
	for i, expected := range expectedIDs {
		if policy.Append[i].ID != expected {
			t.Errorf("Gate %d ID = %q, want %q", i, policy.Append[i].ID, expected)
		}
	}
}

func TestLightweightPolicy(t *testing.T) {
	policy := LightweightPolicy()

	if policy.Name != "lightweight" {
		t.Errorf("Name = %q, want %q", policy.Name, "lightweight")
	}
	if len(policy.Append) != 1 {
		t.Errorf("Lightweight policy has %d gates, want 1", len(policy.Append))
	}
	if policy.Append[0].ID != "code_review" {
		t.Errorf("Gate ID = %q, want %q", policy.Append[0].ID, "code_review")
	}
}
