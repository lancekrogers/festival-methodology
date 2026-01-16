package types

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParseTemplateFilename(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantLevel Level
		wantName  string
		wantDef   bool
	}{
		{
			name:      "festival goal template",
			filename:  "FESTIVAL_GOAL_TEMPLATE.md",
			wantLevel: LevelFestival,
			wantName:  "goal",
			wantDef:   true,
		},
		{
			name:      "phase goal implementation",
			filename:  "PHASE_GOAL_IMPLEMENTATION_TEMPLATE.md",
			wantLevel: LevelPhase,
			wantName:  "implementation",
			wantDef:   false,
		},
		{
			name:      "task template simple",
			filename:  "TASK_TEMPLATE_SIMPLE.md",
			wantLevel: LevelTask,
			wantName:  "simple",
			wantDef:   false,
		},
		{
			name:      "task template default",
			filename:  "TASK_TEMPLATE.md",
			wantLevel: LevelTask,
			wantName:  "template",
			wantDef:   true,
		},
		{
			name:      "sequence goal template",
			filename:  "SEQUENCE_GOAL_TEMPLATE.md",
			wantLevel: LevelSequence,
			wantName:  "goal",
			wantDef:   true,
		},
		{
			name:      "quality gate review",
			filename:  "QUALITY_GATE_REVIEW.md",
			wantLevel: LevelTask,
			wantName:  "gate/review",
			wantDef:   false,
		},
		{
			name:      "research analysis",
			filename:  "RESEARCH_ANALYSIS_TEMPLATE.md",
			wantLevel: LevelTask,
			wantName:  "research/analysis",
			wantDef:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := parseTemplateFilename(tt.filename, "/test", false)
			if info == nil {
				t.Fatal("parseTemplateFilename returned nil")
			}
			if info.Level != tt.wantLevel {
				t.Errorf("Level = %v, want %v", info.Level, tt.wantLevel)
			}
			if info.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", info.Name, tt.wantName)
			}
			if info.IsDefault != tt.wantDef {
				t.Errorf("IsDefault = %v, want %v", info.IsDefault, tt.wantDef)
			}
		})
	}
}

func TestParseTemplateFilenameUnknown(t *testing.T) {
	// Unknown prefixes should return nil
	unknowns := []string{
		"README.md",
		"NOTES.md",
		"random_file.md",
	}

	for _, fn := range unknowns {
		t.Run(fn, func(t *testing.T) {
			info := parseTemplateFilename(fn, "/test", false)
			if info != nil {
				t.Errorf("Expected nil for %s, got %+v", fn, info)
			}
		})
	}
}

func TestRegistryDiscover(t *testing.T) {
	// Create temp directory with test templates
	tmpDir := t.TempDir()

	// Create test templates
	templates := map[string]string{
		"FESTIVAL_GOAL_TEMPLATE.md":    "# Festival Goal\n[REPLACE: description]",
		"PHASE_GOAL_TEMPLATE.md":       "# Phase Goal\n[REPLACE: objective]",
		"SEQUENCE_GOAL_TEMPLATE.md":    "# Sequence Goal",
		"TASK_TEMPLATE.md":             "# Task\n[REPLACE: task_name]",
		"TASK_TEMPLATE_SIMPLE.md":      "# Simple Task",
		"QUALITY_GATE_TESTING.md":      "# Testing Gate\n[REPLACE: criteria]",
	}

	for name, content := range templates {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test template: %v", err)
		}
	}

	// Discover types
	reg := NewRegistry()
	err := reg.Discover(context.Background(), DiscoverOptions{
		BuiltInDir:   tmpDir,
		CountMarkers: true,
	})
	if err != nil {
		t.Fatalf("Discover failed: %v", err)
	}

	// Verify counts
	if len(reg.Festival) < 1 {
		t.Errorf("Expected at least 1 festival type, got %d", len(reg.Festival))
	}
	if len(reg.Phase) < 1 {
		t.Errorf("Expected at least 1 phase type, got %d", len(reg.Phase))
	}
	if len(reg.Sequence) < 1 {
		t.Errorf("Expected at least 1 sequence type, got %d", len(reg.Sequence))
	}
	if len(reg.Task) < 1 {
		t.Errorf("Expected at least 1 task type, got %d", len(reg.Task))
	}

	// Verify marker counting
	for _, ft := range reg.Festival {
		if ft.Name == "goal" && ft.Markers < 1 {
			t.Errorf("Festival goal should have at least 1 marker, got %d", ft.Markers)
		}
	}
}

func TestRegistryFindType(t *testing.T) {
	reg := NewRegistry()
	reg.Festival = []TypeInfo{
		{Name: "goal", Level: LevelFestival},
		{Name: "feature", Level: LevelFestival},
	}
	reg.Task = []TypeInfo{
		{Name: "simple", Level: LevelTask},
	}

	// Find existing type
	found := reg.FindType(LevelFestival, "goal")
	if found == nil {
		t.Error("Expected to find festival/goal type")
	}

	// Find non-existing type
	notFound := reg.FindType(LevelFestival, "nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent type")
	}

	// Find task type
	taskFound := reg.FindType(LevelTask, "simple")
	if taskFound == nil {
		t.Error("Expected to find task/simple type")
	}
}

func TestRegistryAllTypes(t *testing.T) {
	reg := NewRegistry()
	reg.Festival = []TypeInfo{{Name: "f1"}}
	reg.Phase = []TypeInfo{{Name: "p1"}, {Name: "p2"}}
	reg.Sequence = []TypeInfo{{Name: "s1"}}
	reg.Task = []TypeInfo{{Name: "t1"}, {Name: "t2"}, {Name: "t3"}}

	all := reg.AllTypes()
	if len(all) != 7 {
		t.Errorf("Expected 7 total types, got %d", len(all))
	}
}

func TestTypeInfoString(t *testing.T) {
	tests := []struct {
		info TypeInfo
		want string
	}{
		{
			info: TypeInfo{Name: "goal", Level: LevelFestival},
			want: "festival/goal",
		},
		{
			info: TypeInfo{Name: "goal", Level: LevelFestival, IsDefault: true},
			want: "festival/goal (default)",
		},
		{
			info: TypeInfo{Name: "custom", Level: LevelTask, IsCustom: true},
			want: "task/custom (custom)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.info.String()
			if got != tt.want {
				t.Errorf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeInfoQualifiedName(t *testing.T) {
	info := TypeInfo{Name: "implementation", Level: LevelPhase}
	want := "phase/implementation"
	if got := info.QualifiedName(); got != want {
		t.Errorf("QualifiedName() = %q, want %q", got, want)
	}
}

func TestAllLevels(t *testing.T) {
	levels := AllLevels()
	if len(levels) != 4 {
		t.Errorf("Expected 4 levels, got %d", len(levels))
	}
	expected := []Level{LevelFestival, LevelPhase, LevelSequence, LevelTask}
	for i, l := range expected {
		if levels[i] != l {
			t.Errorf("Level %d = %v, want %v", i, levels[i], l)
		}
	}
}

func TestCountMarkersInFile(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")

	content := `# Test
[REPLACE: marker1]
Some text
[REPLACE: marker2] and [REPLACE: marker3]
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	count, err := countMarkersInFile(testFile)
	if err != nil {
		t.Fatalf("countMarkersInFile error: %v", err)
	}
	if count != 3 {
		t.Errorf("Expected 3 markers, got %d", count)
	}
}

func TestRegistryCustomOverridesBuiltIn(t *testing.T) {
	reg := NewRegistry()

	// Add built-in type
	reg.addType(&TypeInfo{
		Name:     "goal",
		Level:    LevelFestival,
		IsCustom: false,
		Markers:  5,
	})

	// Add custom type with same name
	reg.addType(&TypeInfo{
		Name:     "goal",
		Level:    LevelFestival,
		IsCustom: true,
		Markers:  10,
	})

	// Should have only one type, marked as custom
	if len(reg.Festival) != 1 {
		t.Errorf("Expected 1 festival type, got %d", len(reg.Festival))
	}
	if !reg.Festival[0].IsCustom {
		t.Error("Type should be marked as custom")
	}
	if reg.Festival[0].Markers != 10 {
		t.Errorf("Markers should be updated to 10, got %d", reg.Festival[0].Markers)
	}
}
