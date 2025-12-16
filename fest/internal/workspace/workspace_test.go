package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMarkerPath(t *testing.T) {
	festivalsDir := "/path/to/festivals"
	expected := "/path/to/festivals/.festival/.workspace"
	result := MarkerPath(festivalsDir)
	if result != expected {
		t.Errorf("MarkerPath(%q) = %q, want %q", festivalsDir, result, expected)
	}
}

func TestHasMarker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure without marker
	festivalsDir := filepath.Join(tmpDir, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		t.Fatal(err)
	}

	// Test: no marker exists
	if HasMarker(festivalsDir) {
		t.Error("HasMarker returned true when marker does not exist")
	}

	// Create marker file
	markerPath := filepath.Join(dotFestival, ".workspace")
	if err := os.WriteFile(markerPath, []byte("{}"), 0644); err != nil {
		t.Fatal(err)
	}

	// Test: marker exists
	if !HasMarker(festivalsDir) {
		t.Error("HasMarker returned false when marker exists")
	}
}

func TestReadMarker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure with marker
	festivalsDir := filepath.Join(tmpDir, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		t.Fatal(err)
	}

	// Create marker file with valid JSON
	marker := Marker{
		Workspace:  "test-workspace",
		Registered: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}
	data, _ := json.Marshal(marker)
	markerPath := filepath.Join(dotFestival, ".workspace")
	if err := os.WriteFile(markerPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Test: read marker
	result, err := ReadMarker(festivalsDir)
	if err != nil {
		t.Fatalf("ReadMarker failed: %v", err)
	}
	if result.Workspace != "test-workspace" {
		t.Errorf("Workspace = %q, want %q", result.Workspace, "test-workspace")
	}

	// Test: read non-existent marker
	emptyDir := filepath.Join(tmpDir, "empty")
	if err := os.MkdirAll(emptyDir, 0755); err != nil {
		t.Fatal(err)
	}
	_, err = ReadMarker(emptyDir)
	if err == nil {
		t.Error("ReadMarker should fail for non-existent marker")
	}
}

func TestRegisterFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory structure: parent/festivals/
	parentDir := filepath.Join(tmpDir, "my-project")
	festivalsDir := filepath.Join(parentDir, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		t.Fatal(err)
	}

	// Register
	if err := RegisterFestivals(festivalsDir); err != nil {
		t.Fatalf("RegisterFestivals failed: %v", err)
	}

	// Verify marker was created
	if !HasMarker(festivalsDir) {
		t.Error("Marker should exist after registration")
	}

	// Verify workspace name is derived from parent directory
	marker, err := ReadMarker(festivalsDir)
	if err != nil {
		t.Fatalf("Failed to read marker: %v", err)
	}
	if marker.Workspace != "my-project" {
		t.Errorf("Workspace = %q, want %q", marker.Workspace, "my-project")
	}
}

func TestUnregisterFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create and register festivals
	parentDir := filepath.Join(tmpDir, "my-project")
	festivalsDir := filepath.Join(parentDir, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		t.Fatal(err)
	}
	if err := RegisterFestivals(festivalsDir); err != nil {
		t.Fatal(err)
	}

	// Verify marker exists
	if !HasMarker(festivalsDir) {
		t.Fatal("Marker should exist before unregistration")
	}

	// Unregister
	if err := UnregisterFestivals(festivalsDir); err != nil {
		t.Fatalf("UnregisterFestivals failed: %v", err)
	}

	// Verify marker was removed
	if HasMarker(festivalsDir) {
		t.Error("Marker should not exist after unregistration")
	}

	// Unregistering again should not error
	if err := UnregisterFestivals(festivalsDir); err != nil {
		t.Errorf("Unregistering non-existent marker should not error: %v", err)
	}
}

func TestFindMarkedFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure:
	// tmpDir/
	//   outer-project/
	//     festivals/          <- has marker
	//       .festival/
	//     inner-project/
	//       festivals/        <- no marker
	//         .festival/
	//       src/
	//         code/

	outerProject := filepath.Join(tmpDir, "outer-project")
	outerFestivals := filepath.Join(outerProject, "festivals")
	outerDotFestival := filepath.Join(outerFestivals, ".festival")

	innerProject := filepath.Join(outerProject, "inner-project")
	innerFestivals := filepath.Join(innerProject, "festivals")
	innerDotFestival := filepath.Join(innerFestivals, ".festival")

	codeDir := filepath.Join(innerProject, "src", "code")

	// Create all directories
	for _, dir := range []string{outerDotFestival, innerDotFestival, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Register only outer festivals
	if err := RegisterFestivals(outerFestivals); err != nil {
		t.Fatal(err)
	}

	// Test: from codeDir, should find outer festivals (skips inner because no marker)
	result, err := FindMarkedFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindMarkedFestivals failed: %v", err)
	}
	if result != outerFestivals {
		t.Errorf("FindMarkedFestivals = %q, want %q", result, outerFestivals)
	}

	// Test: from inner project root
	result, err = FindMarkedFestivals(innerProject)
	if err != nil {
		t.Fatalf("FindMarkedFestivals failed: %v", err)
	}
	if result != outerFestivals {
		t.Errorf("FindMarkedFestivals = %q, want %q", result, outerFestivals)
	}

	// Test: from outer festivals directly
	result, err = FindMarkedFestivals(outerFestivals)
	if err != nil {
		t.Fatalf("FindMarkedFestivals failed: %v", err)
	}
	if result != outerFestivals {
		t.Errorf("FindMarkedFestivals = %q, want %q", result, outerFestivals)
	}
}

func TestFindMarkedFestivalsNoMarker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals without marker
	festivalsDir := filepath.Join(tmpDir, "project", "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		t.Fatal(err)
	}

	// Test: no marker anywhere
	result, err := FindMarkedFestivals(festivalsDir)
	if err != nil {
		t.Fatalf("FindMarkedFestivals failed: %v", err)
	}
	if result != "" {
		t.Errorf("FindMarkedFestivals should return empty string when no marker found, got %q", result)
	}
}

func TestFindAllMarkedFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure with multiple markers:
	// tmpDir/
	//   level1/
	//     festivals/          <- has marker
	//     level2/
	//       festivals/        <- has marker
	//       level3/
	//         code/

	level1 := filepath.Join(tmpDir, "level1")
	level1Festivals := filepath.Join(level1, "festivals")
	level1DotFestival := filepath.Join(level1Festivals, ".festival")

	level2 := filepath.Join(level1, "level2")
	level2Festivals := filepath.Join(level2, "festivals")
	level2DotFestival := filepath.Join(level2Festivals, ".festival")

	codeDir := filepath.Join(level2, "level3", "code")

	// Create all directories
	for _, dir := range []string{level1DotFestival, level2DotFestival, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Register both
	if err := RegisterFestivals(level1Festivals); err != nil {
		t.Fatal(err)
	}
	if err := RegisterFestivals(level2Festivals); err != nil {
		t.Fatal(err)
	}

	// Test: from codeDir, should find both markers (level2 first, then level1)
	results, err := FindAllMarkedFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindAllMarkedFestivals failed: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("FindAllMarkedFestivals returned %d results, want 2", len(results))
	}
	// First should be level2 (closest)
	if len(results) > 0 && results[0] != level2Festivals {
		t.Errorf("First result = %q, want %q", results[0], level2Festivals)
	}
	// Second should be level1
	if len(results) > 1 && results[1] != level1Festivals {
		t.Errorf("Second result = %q, want %q", results[1], level1Festivals)
	}
}

func TestFindNearestFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure without markers
	project := filepath.Join(tmpDir, "project")
	festivalsDir := filepath.Join(project, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	codeDir := filepath.Join(project, "src", "code")

	for _, dir := range []string{dotFestival, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Test: finds nearest festivals even without marker
	result, err := FindNearestFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindNearestFestivals failed: %v", err)
	}
	if result != festivalsDir {
		t.Errorf("FindNearestFestivals = %q, want %q", result, festivalsDir)
	}

	// Test: from festivals directory itself
	result, err = FindNearestFestivals(festivalsDir)
	if err != nil {
		t.Fatalf("FindNearestFestivals failed: %v", err)
	}
	if result != festivalsDir {
		t.Errorf("FindNearestFestivals = %q, want %q", result, festivalsDir)
	}
}

func TestFindNearestFestivalsRequiresDotFestival(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals directory WITHOUT .festival subdirectory
	project := filepath.Join(tmpDir, "project")
	festivalsDir := filepath.Join(project, "festivals")
	codeDir := filepath.Join(project, "src")

	for _, dir := range []string{festivalsDir, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Test: should NOT find festivals without .festival subdirectory
	result, err := FindNearestFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindNearestFestivals failed: %v", err)
	}
	if result != "" {
		t.Errorf("FindNearestFestivals should return empty when .festival missing, got %q", result)
	}
}

func TestFindFestivals(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure with marker
	project := filepath.Join(tmpDir, "project")
	festivalsDir := filepath.Join(project, "festivals")
	dotFestival := filepath.Join(festivalsDir, ".festival")
	codeDir := filepath.Join(project, "src")

	for _, dir := range []string{dotFestival, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Test without marker: should use fallback
	result, err := FindFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindFestivals failed: %v", err)
	}
	if result != festivalsDir {
		t.Errorf("FindFestivals (no marker) = %q, want %q", result, festivalsDir)
	}

	// Register and test with marker
	if err := RegisterFestivals(festivalsDir); err != nil {
		t.Fatal(err)
	}

	result, err = FindFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindFestivals failed: %v", err)
	}
	if result != festivalsDir {
		t.Errorf("FindFestivals (with marker) = %q, want %q", result, festivalsDir)
	}
}

func TestFindFestivalsPrefersMark(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure where only outer has marker
	// tmpDir/
	//   outer/
	//     festivals/          <- has marker
	//       .festival/
	//     inner/
	//       festivals/        <- no marker (like source repo)
	//         .festival/
	//       code/

	outer := filepath.Join(tmpDir, "outer")
	outerFestivals := filepath.Join(outer, "festivals")
	outerDotFestival := filepath.Join(outerFestivals, ".festival")

	inner := filepath.Join(outer, "inner")
	innerFestivals := filepath.Join(inner, "festivals")
	innerDotFestival := filepath.Join(innerFestivals, ".festival")

	codeDir := filepath.Join(inner, "code")

	for _, dir := range []string{outerDotFestival, innerDotFestival, codeDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Register only outer
	if err := RegisterFestivals(outerFestivals); err != nil {
		t.Fatal(err)
	}

	// Test: from code dir, should find outer (marked) not inner (unmarked)
	result, err := FindFestivals(codeDir)
	if err != nil {
		t.Fatalf("FindFestivals failed: %v", err)
	}
	if result != outerFestivals {
		t.Errorf("FindFestivals should prefer marked directory, got %q, want %q", result, outerFestivals)
	}
}
