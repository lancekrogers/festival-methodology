package show

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHasNumericPrefix(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"001_PLAN", true},
		{"01_setup", true},
		{"1_task", true},
		{"001", false},    // No underscore
		{"abc_test", false},
		{"a01_test", false},
		{"", false},
		{"_001", false},
	}

	for _, tc := range tests {
		result := hasNumericPrefix(tc.input)
		if result != tc.expected {
			t.Errorf("hasNumericPrefix(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsGateFile(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"01_quality_gate.md", true},
		{"01_testing_gate.md", true},
		{"01_code_review.md", true},
		{"01_verify_build.md", true},
		{"01_iterate_feedback.md", true},
		{"01_implementation.md", false},
		{"01_task.md", false},
		{"SEQUENCE_GOAL.md", false},
	}

	for _, tc := range tests {
		result := isGateFile(tc.input)
		if result != tc.expected {
			t.Errorf("isGateFile(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsValidFestival(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid festival with FESTIVAL_GOAL.md
	validFestival1 := filepath.Join(tmpDir, "valid1")
	if err := os.MkdirAll(validFestival1, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validFestival1, FestivalGoalFile), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a valid festival with fest.yaml
	validFestival2 := filepath.Join(tmpDir, "valid2")
	if err := os.MkdirAll(validFestival2, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(validFestival2, FestivalConfigFile), []byte("name: test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create an invalid directory
	invalidDir := filepath.Join(tmpDir, "invalid")
	if err := os.MkdirAll(invalidDir, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		dir      string
		expected bool
	}{
		{validFestival1, true},
		{validFestival2, true},
		{invalidDir, false},
		{filepath.Join(tmpDir, "nonexistent"), false},
	}

	for _, tc := range tests {
		result := isValidFestival(tc.dir)
		if result != tc.expected {
			t.Errorf("isValidFestival(%q) = %v, want %v", tc.dir, result, tc.expected)
		}
	}
}

func TestDetectCurrentFestival(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure
	festivalDir := filepath.Join(tmpDir, "my-festival")
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	seqDir := filepath.Join(phaseDir, "01_setup")

	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(festivalDir, FestivalGoalFile), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(phaseDir, PhaseGoalFile), []byte("# Phase"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seqDir, SequenceGoalFile), []byte("# Sequence"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		startDir string
		wantName string
		wantErr  bool
	}{
		{festivalDir, "my-festival", false},
		{phaseDir, "my-festival", false},
		{seqDir, "my-festival", false},
		{tmpDir, "", true}, // Not in a festival
	}

	for _, tc := range tests {
		result, err := DetectCurrentFestival(tc.startDir)
		if tc.wantErr {
			if err == nil {
				t.Errorf("DetectCurrentFestival(%q) expected error, got nil", tc.startDir)
			}
		} else {
			if err != nil {
				t.Errorf("DetectCurrentFestival(%q) unexpected error: %v", tc.startDir, err)
			} else if result.Name != tc.wantName {
				t.Errorf("DetectCurrentFestival(%q) name = %q, want %q", tc.startDir, result.Name, tc.wantName)
			}
		}
	}
}

func TestDetectCurrentLocation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure
	festivalDir := filepath.Join(tmpDir, "my-festival")
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	seqDir := filepath.Join(phaseDir, "01_setup")

	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(festivalDir, FestivalGoalFile), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(phaseDir, PhaseGoalFile), []byte("# Phase"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seqDir, SequenceGoalFile), []byte("# Sequence"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		startDir     string
		wantType     string
		wantPhase    string
		wantSequence string
	}{
		{festivalDir, "festival", "", ""},
		{phaseDir, "phase", "001_PLAN", ""},
		{seqDir, "sequence", "001_PLAN", "01_setup"},
	}

	for _, tc := range tests {
		result, err := DetectCurrentLocation(tc.startDir)
		if err != nil {
			t.Errorf("DetectCurrentLocation(%q) unexpected error: %v", tc.startDir, err)
			continue
		}
		if result.Type != tc.wantType {
			t.Errorf("DetectCurrentLocation(%q) type = %q, want %q", tc.startDir, result.Type, tc.wantType)
		}
		if result.Phase != tc.wantPhase {
			t.Errorf("DetectCurrentLocation(%q) phase = %q, want %q", tc.startDir, result.Phase, tc.wantPhase)
		}
		if result.Sequence != tc.wantSequence {
			t.Errorf("DetectCurrentLocation(%q) sequence = %q, want %q", tc.startDir, result.Sequence, tc.wantSequence)
		}
	}
}

func TestCalculateFestivalStats(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure with phases, sequences, and tasks
	festivalDir := filepath.Join(tmpDir, "my-festival")
	phase1 := filepath.Join(festivalDir, "001_PLAN")
	seq1 := filepath.Join(phase1, "01_setup")

	if err := os.MkdirAll(seq1, 0755); err != nil {
		t.Fatal(err)
	}

	// Create goal files
	if err := os.WriteFile(filepath.Join(festivalDir, FestivalGoalFile), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(phase1, PhaseGoalFile), []byte("# Phase"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seq1, SequenceGoalFile), []byte("# Sequence"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create tasks
	if err := os.WriteFile(filepath.Join(seq1, "01_task1.md"), []byte("# Task 1"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seq1, "02_task2.md"), []byte("# Task 2"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(seq1, "03_quality_gate.md"), []byte("# Gate"), 0644); err != nil {
		t.Fatal(err)
	}

	stats, err := CalculateFestivalStats(festivalDir)
	if err != nil {
		t.Fatalf("CalculateFestivalStats() unexpected error: %v", err)
	}

	if stats.Phases.Total != 1 {
		t.Errorf("Phases.Total = %d, want 1", stats.Phases.Total)
	}
	if stats.Sequences.Total != 1 {
		t.Errorf("Sequences.Total = %d, want 1", stats.Sequences.Total)
	}
	if stats.Tasks.Total != 2 {
		t.Errorf("Tasks.Total = %d, want 2", stats.Tasks.Total)
	}
	if stats.Gates.Total != 1 {
		t.Errorf("Gates.Total = %d, want 1", stats.Gates.Total)
	}
}

func TestListFestivalsByStatus(t *testing.T) {
	tmpDir := t.TempDir()

	// Create status directories with festivals
	activeDir := filepath.Join(tmpDir, "active")
	plannedDir := filepath.Join(tmpDir, "planned")

	festival1 := filepath.Join(activeDir, "fest1")
	festival2 := filepath.Join(activeDir, "fest2")
	festival3 := filepath.Join(plannedDir, "fest3")

	for _, d := range []string{festival1, festival2, festival3} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
		// Add festival goal file to make it valid
		if err := os.WriteFile(filepath.Join(d, FestivalGoalFile), []byte("# Test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Test active festivals
	active, err := ListFestivalsByStatus(tmpDir, "active")
	if err != nil {
		t.Fatalf("ListFestivalsByStatus(active) unexpected error: %v", err)
	}
	if len(active) != 2 {
		t.Errorf("ListFestivalsByStatus(active) returned %d festivals, want 2", len(active))
	}

	// Test planned festivals
	planned, err := ListFestivalsByStatus(tmpDir, "planned")
	if err != nil {
		t.Fatalf("ListFestivalsByStatus(planned) unexpected error: %v", err)
	}
	if len(planned) != 1 {
		t.Errorf("ListFestivalsByStatus(planned) returned %d festivals, want 1", len(planned))
	}

	// Test non-existent status
	empty, err := ListFestivalsByStatus(tmpDir, "completed")
	if err != nil {
		t.Fatalf("ListFestivalsByStatus(completed) unexpected error: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("ListFestivalsByStatus(completed) returned %d festivals, want 0", len(empty))
	}
}

func TestFormatFestivalDetails(t *testing.T) {
	festival := &FestivalInfo{
		ID:     "test-fest",
		Name:   "test-fest",
		Status: "active",
		Path:   "/path/to/test-fest",
		Stats: &FestivalStats{
			Phases: StatusCounts{Total: 3, Completed: 1, InProgress: 1, Pending: 1},
			Tasks:  StatusCounts{Total: 10, Completed: 5, InProgress: 2, Pending: 3},
		},
	}
	festival.Stats.Progress = 50.0

	output := FormatFestivalDetails(festival, false)

	// Check that key elements are present
	if !contains(output, "test-fest") {
		t.Error("Output should contain festival name")
	}
	if !contains(output, "active") {
		t.Error("Output should contain status")
	}
	if !contains(output, "50.0%") {
		t.Error("Output should contain progress percentage")
	}
}

func TestFormatFestivalList(t *testing.T) {
	festivals := []*FestivalInfo{
		{Name: "fest1", Stats: &FestivalStats{Progress: 25}},
		{Name: "fest2", Stats: &FestivalStats{Progress: 75}},
	}

	output := FormatFestivalList("active", festivals)

	if !contains(output, "Active Festivals (2)") {
		t.Error("Output should contain header with count")
	}
	if !contains(output, "fest1") {
		t.Error("Output should contain festival names")
	}
	if !contains(output, "[25%]") {
		t.Error("Output should contain progress")
	}
}

func TestFormatFestivalListEmpty(t *testing.T) {
	output := FormatFestivalList("completed", []*FestivalInfo{})

	if !contains(output, "Completed Festivals (0)") {
		t.Error("Output should indicate zero festivals")
	}
	if !contains(output, "(none)") {
		t.Error("Output should indicate no festivals")
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || contains(s[1:], substr)))
}
