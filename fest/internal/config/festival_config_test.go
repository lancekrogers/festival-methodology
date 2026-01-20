package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultFestivalConfig(t *testing.T) {
	cfg := DefaultFestivalConfig()

	if cfg.Version != "1.0" {
		t.Errorf("expected version 1.0, got %s", cfg.Version)
	}

	if !cfg.QualityGates.Enabled {
		t.Error("expected quality gates to be enabled by default")
	}

	if len(cfg.QualityGates.Tasks) != 4 {
		t.Errorf("expected 4 default quality gate tasks, got %d", len(cfg.QualityGates.Tasks))
	}

	// Check task IDs (IDs match template filenames)
	expectedIDs := []string{"QUALITY_GATE_TESTING", "QUALITY_GATE_REVIEW", "QUALITY_GATE_ITERATE", "QUALITY_GATE_COMMIT"}
	for i, task := range cfg.QualityGates.Tasks {
		if task.ID != expectedIDs[i] {
			t.Errorf("expected task ID %s, got %s", expectedIDs[i], task.ID)
		}
		if !task.Enabled {
			t.Errorf("expected task %s to be enabled", task.ID)
		}
	}

	// Check PhaseGates is populated
	if cfg.QualityGates.PhaseGates == nil {
		t.Error("expected PhaseGates to be populated")
	}

	// Check implementation gates are in correct order
	implGates := cfg.QualityGates.PhaseGates["implementation"]
	if len(implGates) != 4 {
		t.Errorf("expected 4 implementation gates, got %d", len(implGates))
	}
	for i, gate := range implGates {
		if gate.ID != expectedIDs[i] {
			t.Errorf("expected implementation gate ID %s at index %d, got %s", expectedIDs[i], i, gate.ID)
		}
	}
}

func TestLoadFestivalConfig_Default(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Load from non-existent file should return defaults
	cfg, err := LoadFestivalConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != "1.0" {
		t.Errorf("expected default version, got %s", cfg.Version)
	}
}

func TestLoadFestivalConfig_FromFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test config file
	configContent := `version: "2.0"
quality_gates:
  enabled: true
  auto_append: false
  tasks:
    - id: custom_test
      template: CUSTOM_TEMPLATE
      enabled: true
excluded_patterns:

- "*_docs"
`
	configPath := filepath.Join(tmpDir, FestivalConfigFileName)
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	cfg, err := LoadFestivalConfig(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Version != "2.0" {
		t.Errorf("expected version 2.0, got %s", cfg.Version)
	}

	if cfg.QualityGates.AutoAppend {
		t.Error("expected auto_append to be false")
	}

	if len(cfg.QualityGates.Tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(cfg.QualityGates.Tasks))
	}

	if cfg.QualityGates.Tasks[0].ID != "custom_test" {
		t.Errorf("expected task ID custom_test, got %s", cfg.QualityGates.Tasks[0].ID)
	}
}

func TestSaveFestivalConfig(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &FestivalConfig{
		Version: "1.0",
		QualityGates: QualityGatesConfig{
			Enabled: true,
			Tasks: []QualityGateTask{
				{
					ID:       "test_task",
					Template: "TEST_TEMPLATE",
					Enabled:  true,
				},
			},
		},
	}

	if err := SaveFestivalConfig(tmpDir, cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, FestivalConfigFileName)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Error("config file was not created")
	}

	// Load and verify
	loaded, err := LoadFestivalConfig(tmpDir)
	if err != nil {
		t.Fatalf("failed to load saved config: %v", err)
	}

	if loaded.Version != cfg.Version {
		t.Errorf("version mismatch: expected %s, got %s", cfg.Version, loaded.Version)
	}
}

func TestIsSequenceExcluded(t *testing.T) {
	cfg := &FestivalConfig{
		ExcludedPatterns: []string{
			"*_planning",
			"*_requirements",
			"01_research",
		},
	}

	tests := []struct {
		name     string
		expected bool
	}{
		{"01_planning", true},
		{"02_requirements", true},
		{"01_research", true},
		{"01_implementation", false},
		{"02_api", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cfg.IsSequenceExcluded(tt.name)
			if result != tt.expected {
				t.Errorf("IsSequenceExcluded(%s) = %v, want %v", tt.name, result, tt.expected)
			}
		})
	}
}

func TestGetEnabledTasks(t *testing.T) {
	cfg := &FestivalConfig{
		QualityGates: QualityGatesConfig{
			Tasks: []QualityGateTask{
				{ID: "task1", Enabled: true},
				{ID: "task2", Enabled: false},
				{ID: "task3", Enabled: true},
			},
		},
	}

	enabled := cfg.GetEnabledTasks()

	if len(enabled) != 2 {
		t.Errorf("expected 2 enabled tasks, got %d", len(enabled))
	}

	if enabled[0].ID != "task1" || enabled[1].ID != "task3" {
		t.Error("unexpected enabled tasks returned")
	}
}

func TestFestivalConfigExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Should not exist initially
	if FestivalConfigExists(tmpDir) {
		t.Error("expected config to not exist")
	}

	// Create config file
	configPath := filepath.Join(tmpDir, FestivalConfigFileName)
	if err := os.WriteFile(configPath, []byte("version: 1.0"), 0644); err != nil {
		t.Fatalf("failed to create config: %v", err)
	}

	// Should exist now
	if !FestivalConfigExists(tmpDir) {
		t.Error("expected config to exist")
	}
}
