package navigation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsInsideFestival(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure
	festivalPath := filepath.Join(tmpDir, "festivals", "active", "my-fest")
	festivalPhase := filepath.Join(festivalPath, "001_PLAN")
	festivalTask := filepath.Join(festivalPhase, "tasks")
	projectPath := filepath.Join(tmpDir, "projects", "my-project")
	projectSubdir := filepath.Join(projectPath, "src", "components")

	// Create all directories
	for _, d := range []string{festivalTask, projectSubdir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"festival root", festivalPath, true},
		{"festival phase", festivalPhase, true},
		{"festival nested task", festivalTask, true},
		{"project root", projectPath, false},
		{"project subdir", projectSubdir, false},
		{"festivals parent dir", filepath.Join(tmpDir, "festivals"), true},
		{"random dir", tmpDir, false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isInsideFestival(tc.path)
			if result != tc.expected {
				t.Errorf("isInsideFestival(%q) = %v, want %v", tc.path, result, tc.expected)
			}
		})
	}
}

func TestCollectFestivals(t *testing.T) {
	tmpDir := t.TempDir()
	festivalsDir := filepath.Join(tmpDir, "festivals")

	// Create some festivals
	activeFests := []string{"fest-a", "fest-b"}
	plannedFests := []string{"fest-c"}

	for _, f := range activeFests {
		if err := os.MkdirAll(filepath.Join(festivalsDir, "active", f), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, f := range plannedFests {
		if err := os.MkdirAll(filepath.Join(festivalsDir, "planned", f), 0755); err != nil {
			t.Fatal(err)
		}
	}

	festivals, err := collectFestivals(festivalsDir)
	if err != nil {
		t.Fatalf("collectFestivals() error = %v", err)
	}

	if len(festivals) != 3 {
		t.Errorf("collectFestivals() returned %d festivals, want 3", len(festivals))
	}

	// Verify festivals are categorized correctly
	activeCount := 0
	plannedCount := 0
	for _, f := range festivals {
		switch f.status {
		case "active":
			activeCount++
		case "planned":
			plannedCount++
		}
	}

	if activeCount != 2 {
		t.Errorf("Expected 2 active festivals, got %d", activeCount)
	}
	if plannedCount != 1 {
		t.Errorf("Expected 1 planned festival, got %d", plannedCount)
	}
}

func TestResolveFestivalPath(t *testing.T) {
	tmpDir := t.TempDir()
	festivalsDir := filepath.Join(tmpDir, "festivals")

	// Create festivals in different statuses
	activeFest := filepath.Join(festivalsDir, "active", "my-active-fest")
	plannedFest := filepath.Join(festivalsDir, "planned", "my-planned-fest")
	completedFest := filepath.Join(festivalsDir, "completed", "my-completed-fest")

	for _, f := range []string{activeFest, plannedFest, completedFest} {
		if err := os.MkdirAll(f, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		expected string
	}{
		{"my-active-fest", activeFest},
		{"my-planned-fest", plannedFest},
		{"my-completed-fest", completedFest},
		{"nonexistent", ""},
	}

	for _, tc := range tests {
		result := resolveFestivalPath(festivalsDir, tc.name)
		if result != tc.expected {
			t.Errorf("resolveFestivalPath(%q) = %q, want %q", tc.name, result, tc.expected)
		}
	}
}
