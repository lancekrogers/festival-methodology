package types

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/types"
)

func TestNewTypesCommand(t *testing.T) {
	cmd := NewTypesCommand()
	if cmd.Use != "types" {
		t.Errorf("Use = %q, want %q", cmd.Use, "types")
	}

	// Check subcommands exist
	listCmd, _, err := cmd.Find([]string{"list"})
	if err != nil {
		t.Errorf("list subcommand not found: %v", err)
	}
	if listCmd == nil {
		t.Error("list subcommand is nil")
	}

	showCmd, _, err := cmd.Find([]string{"show"})
	if err != nil {
		t.Errorf("show subcommand not found: %v", err)
	}
	if showCmd == nil {
		t.Error("show subcommand is nil")
	}
}

func TestFindType(t *testing.T) {
	registry := types.NewRegistry()
	registry.Festival = append(registry.Festival, types.TypeInfo{
		Name:  "feature",
		Level: types.LevelFestival,
	})
	registry.Phase = append(registry.Phase, types.TypeInfo{
		Name:  "implementation",
		Level: types.LevelPhase,
	})
	registry.Task = append(registry.Task, types.TypeInfo{
		Name:  "simple",
		Level: types.LevelTask,
	})

	tests := []struct {
		name        string
		typeName    string
		level       string
		wantLevel   types.Level
		wantErr     bool
		errContains string
	}{
		{
			name:      "find by name only",
			typeName:  "feature",
			level:     "",
			wantLevel: types.LevelFestival,
			wantErr:   false,
		},
		{
			name:      "find by name and level",
			typeName:  "implementation",
			level:     "phase",
			wantLevel: types.LevelPhase,
			wantErr:   false,
		},
		{
			name:        "not found",
			typeName:    "nonexistent",
			level:       "",
			wantErr:     true,
			errContains: "not found",
		},
		{
			name:        "not found at level",
			typeName:    "feature",
			level:       "task",
			wantErr:     true,
			errContains: "not found at level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := findType(registry, tt.typeName, tt.level)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("error = %q, want to contain %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Level != tt.wantLevel {
				t.Errorf("Level = %v, want %v", got.Level, tt.wantLevel)
			}
		})
	}
}

func TestFindTypeAmbiguous(t *testing.T) {
	registry := types.NewRegistry()
	// Add same name at different levels
	registry.Festival = append(registry.Festival, types.TypeInfo{
		Name:  "goal",
		Level: types.LevelFestival,
	})
	registry.Phase = append(registry.Phase, types.TypeInfo{
		Name:  "goal",
		Level: types.LevelPhase,
	})

	_, err := findType(registry, "goal", "")
	if err == nil {
		t.Error("expected error for ambiguous type")
	}
	if !strings.Contains(err.Error(), "multiple levels") {
		t.Errorf("error should mention multiple levels, got: %v", err)
	}
}

func TestFindSimilarTypes(t *testing.T) {
	registry := types.NewRegistry()
	registry.Festival = append(registry.Festival, types.TypeInfo{
		Name:  "feature",
		Level: types.LevelFestival,
	})
	registry.Phase = append(registry.Phase, types.TypeInfo{
		Name:  "implementation",
		Level: types.LevelPhase,
	})
	registry.Task = append(registry.Task, types.TypeInfo{
		Name:  "simple",
		Level: types.LevelTask,
	})

	tests := []struct {
		name         string
		searchFor    string
		wantContains string
	}{
		{
			name:         "partial match",
			searchFor:    "feat",
			wantContains: "festival/feature",
		},
		{
			name:         "impl prefix",
			searchFor:    "impl",
			wantContains: "phase/implementation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := findSimilarTypes(registry, tt.searchFor)
			found := false
			for _, s := range suggestions {
				if strings.Contains(s, tt.wantContains) {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("suggestions %v should contain %q", suggestions, tt.wantContains)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"festival", "Festival"},
		{"phase", "Phase"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := capitalize(tt.input)
			if got != tt.want {
				t.Errorf("capitalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBoolToYesNo(t *testing.T) {
	if got := boolToYesNo(true); got != "Yes" {
		t.Errorf("boolToYesNo(true) = %q, want %q", got, "Yes")
	}
	if got := boolToYesNo(false); got != "No" {
		t.Errorf("boolToYesNo(false) = %q, want %q", got, "No")
	}
}

func TestRunListWithTempDir(t *testing.T) {
	// Create temp directory with test templates
	tmpDir := t.TempDir()
	templates := map[string]string{
		"FESTIVAL_GOAL_TEMPLATE.md": "# Festival Goal",
		"TASK_TEMPLATE.md":          "# Task",
	}

	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}
	}

	// Mock the built-in dir function temporarily
	origFunc := getBuiltInTemplatesDir
	defer func() { _ = origFunc }()

	// Test that we can create a registry and discover types
	registry := types.NewRegistry()
	err := registry.Discover(context.Background(), types.DiscoverOptions{
		BuiltInDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	if len(registry.Festival) < 1 {
		t.Errorf("Expected at least 1 festival type")
	}
	if len(registry.Task) < 1 {
		t.Errorf("Expected at least 1 task type")
	}
}

func TestFilterCustomTypes(t *testing.T) {
	typeInfos := []types.TypeInfo{
		{Name: "feature", Level: types.LevelFestival, IsCustom: false},
		{Name: "custom_fest", Level: types.LevelFestival, IsCustom: true},
		{Name: "standard", Level: types.LevelTask, IsCustom: false},
		{Name: "my_task", Level: types.LevelTask, IsCustom: true},
	}

	filtered := filterCustomTypes(typeInfos)

	if len(filtered) != 2 {
		t.Errorf("Expected 2 custom types, got %d", len(filtered))
	}

	// Verify only custom types remain
	for _, ti := range filtered {
		if !ti.IsCustom {
			t.Errorf("Non-custom type %s should have been filtered out", ti.Name)
		}
	}
}

func TestOutputTypeText(t *testing.T) {
	typeInfo := &types.TypeInfo{
		Name:        "feature",
		Level:       types.LevelFestival,
		Description: "A feature implementation",
		Markers:     50,
		IsDefault:   false,
		IsCustom:    false,
		Templates:   []string{"FESTIVAL_GOAL_FEATURE.md"},
		Source:      "/test/templates",
	}

	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputTypeText(typeInfo, false)

	_ = w.Close()
	os.Stdout = old

	if err != nil {
		t.Fatalf("outputTypeText error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	output := buf.String()

	// Check expected content
	checks := []string{
		"Type: feature",
		"Level: festival",
		"Markers: ~50",
		"fest create festival --type feature",
	}

	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("output missing %q", check)
		}
	}
}
