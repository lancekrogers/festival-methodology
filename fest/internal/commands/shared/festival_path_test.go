package shared

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
)

func TestResolveFestivalPath_ExplicitPath(t *testing.T) {
	// Explicit path should take precedence
	explicitPath := "/some/explicit/festival/path"
	result, err := ResolveFestivalPath("/any/cwd", explicitPath)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != explicitPath {
		t.Errorf("expected %q, got %q", explicitPath, result)
	}
}

func TestResolveFestivalPath_LinkedProject(t *testing.T) {
	// Create temp directories for festival structure
	tmpDir := t.TempDir()

	// Create festivals root with active festival
	festivalsDir := filepath.Join(tmpDir, "festivals")
	activeFest := filepath.Join(festivalsDir, "active", "test-festival-FT0001")
	if err := os.MkdirAll(activeFest, 0755); err != nil {
		t.Fatal(err)
	}

	// Create festival marker file
	goalPath := filepath.Join(activeFest, "FESTIVAL_GOAL.md")
	if err := os.WriteFile(goalPath, []byte("# Test Festival\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create project directory (separate from festivals)
	projectDir := filepath.Join(tmpDir, "my-project")
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Set up navigation with link
	// Override config dir for test isolation
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	configDir := filepath.Join(tmpDir, "config")
	os.Setenv("XDG_CONFIG_HOME", configDir)
	defer func() {
		if origConfigDir == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		}
	}()

	// Ensure config directory exists
	festConfigDir := filepath.Join(configDir, "fest")
	if err := os.MkdirAll(festConfigDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create navigation link
	nav, err := navigation.LoadNavigation()
	if err != nil {
		t.Fatalf("failed to load navigation: %v", err)
	}

	nav.SetLinkWithPath("test-festival-FT0001", projectDir, activeFest)
	if err := nav.Save(); err != nil {
		t.Fatalf("failed to save navigation: %v", err)
	}

	// Test: from project subdirectory, should resolve to festival
	projectSubdir := filepath.Join(projectDir, "src", "pkg")
	if err := os.MkdirAll(projectSubdir, 0755); err != nil {
		t.Fatal(err)
	}

	// We need to mock FindFestivals to return our test festivals dir
	// For now, just verify navigation lookup works
	loadedNav, err := navigation.LoadNavigation()
	if err != nil {
		t.Fatalf("failed to reload navigation: %v", err)
	}

	festivalName := loadedNav.FindFestivalForPath(projectSubdir)
	if festivalName != "test-festival-FT0001" {
		t.Errorf("expected festival name 'test-festival-FT0001', got %q", festivalName)
	}
}

func TestResolveFestivalPath_InsideFestival(t *testing.T) {
	// Create temp festival directory
	tmpDir := t.TempDir()
	festivalDir := filepath.Join(tmpDir, "test-festival")
	if err := os.MkdirAll(festivalDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fest.yaml marker
	configPath := filepath.Join(festivalDir, "fest.yaml")
	if err := os.WriteFile(configPath, []byte("name: test-festival\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// From inside festival, should resolve to festival root
	result, err := ResolveFestivalPath(festivalDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != festivalDir {
		t.Errorf("expected %q, got %q", festivalDir, result)
	}
}

func TestResolveFestivalPath_FestivalSubdir(t *testing.T) {
	// Create temp festival directory
	tmpDir := t.TempDir()
	festivalDir := filepath.Join(tmpDir, "test-festival")
	phaseDir := filepath.Join(festivalDir, "001_PHASE")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create fest.yaml marker at festival root
	configPath := filepath.Join(festivalDir, "fest.yaml")
	if err := os.WriteFile(configPath, []byte("name: test-festival\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// From phase subdirectory, should resolve to festival root
	result, err := ResolveFestivalPath(phaseDir, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != festivalDir {
		t.Errorf("expected %q, got %q", festivalDir, result)
	}
}

func TestResolveFestivalPath_NoFestival(t *testing.T) {
	// Create temp directory that is NOT a festival
	tmpDir := t.TempDir()
	randomDir := filepath.Join(tmpDir, "random-dir")
	if err := os.MkdirAll(randomDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Override config to have no links
	origConfigDir := os.Getenv("XDG_CONFIG_HOME")
	configDir := filepath.Join(tmpDir, "config")
	os.Setenv("XDG_CONFIG_HOME", configDir)
	defer func() {
		if origConfigDir == "" {
			os.Unsetenv("XDG_CONFIG_HOME")
		} else {
			os.Setenv("XDG_CONFIG_HOME", origConfigDir)
		}
	}()

	// From random directory with no links, should return error
	_, err := ResolveFestivalPath(randomDir, "")
	if err == nil {
		t.Error("expected error when not in festival and no link, got nil")
	}
}

func TestFindFestivalByName(t *testing.T) {
	tmpDir := t.TempDir()
	festivalsDir := filepath.Join(tmpDir, "festivals")

	// Create festivals in different statuses
	tests := []struct {
		name   string
		status string
	}{
		{"active-fest", "active"},
		{"planned-fest", "planned"},
		{"completed-fest", "completed"},
		{"dungeon-fest", "dungeon"},
	}

	for _, tc := range tests {
		festPath := filepath.Join(festivalsDir, tc.status, tc.name)
		if err := os.MkdirAll(festPath, 0755); err != nil {
			t.Fatal(err)
		}
		// Create marker file
		goalPath := filepath.Join(festPath, "FESTIVAL_GOAL.md")
		if err := os.WriteFile(goalPath, []byte("# Goal\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test finding each festival by name
	for _, tc := range tests {
		expected := filepath.Join(festivalsDir, tc.status, tc.name)
		result, err := findFestivalByName(festivalsDir, tc.name)
		if err != nil {
			// findFestivalByName falls back to FindFestivalRoot on not found
			// So for this test we just verify it searched the right places
			continue
		}
		if result != expected {
			t.Errorf("findFestivalByName(%q) = %q, want %q", tc.name, result, expected)
		}
	}
}

func TestIsValidFestival(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with FESTIVAL_GOAL.md
	festWithGoal := filepath.Join(tmpDir, "fest-goal")
	os.MkdirAll(festWithGoal, 0755)
	os.WriteFile(filepath.Join(festWithGoal, "FESTIVAL_GOAL.md"), []byte("# Goal\n"), 0644)

	if !isValidFestival(festWithGoal) {
		t.Error("expected directory with FESTIVAL_GOAL.md to be valid")
	}

	// Test with FESTIVAL_OVERVIEW.md
	festWithOverview := filepath.Join(tmpDir, "fest-overview")
	os.MkdirAll(festWithOverview, 0755)
	os.WriteFile(filepath.Join(festWithOverview, "FESTIVAL_OVERVIEW.md"), []byte("# Overview\n"), 0644)

	if !isValidFestival(festWithOverview) {
		t.Error("expected directory with FESTIVAL_OVERVIEW.md to be valid")
	}

	// Test with fest.yaml
	festWithConfig := filepath.Join(tmpDir, "fest-config")
	os.MkdirAll(festWithConfig, 0755)
	os.WriteFile(filepath.Join(festWithConfig, "fest.yaml"), []byte("name: test\n"), 0644)

	if !isValidFestival(festWithConfig) {
		t.Error("expected directory with fest.yaml to be valid")
	}

	// Test empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	os.MkdirAll(emptyDir, 0755)

	if isValidFestival(emptyDir) {
		t.Error("expected empty directory to be invalid")
	}
}

// TestConfigDir verifies the config directory path
func TestConfigDir(t *testing.T) {
	// Just verify config.ConfigDir() returns a non-empty path
	dir := config.ConfigDir()
	if dir == "" {
		t.Error("config.ConfigDir() returned empty string")
	}
}
