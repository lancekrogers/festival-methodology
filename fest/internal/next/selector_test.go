package next

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewSelector(t *testing.T) {
	s := NewSelector("/tmp/test-festival")
	if s == nil {
		t.Fatal("NewSelector returned nil")
	}
	if s.festivalPath != "/tmp/test-festival" {
		t.Errorf("festivalPath = %q, want /tmp/test-festival", s.festivalPath)
	}
}

func TestSelector_determineLocation(t *testing.T) {
	festivalPath := "/festivals/my-festival"
	s := NewSelector(festivalPath)

	tests := []struct {
		name        string
		currentPath string
		wantPhase   string
		wantSeq     string
	}{
		{
			name:        "at festival root",
			currentPath: festivalPath,
			wantPhase:   "",
			wantSeq:     "",
		},
		{
			name:        "in phase directory",
			currentPath: filepath.Join(festivalPath, "01_Planning"),
			wantPhase:   filepath.Join(festivalPath, "01_Planning"),
			wantSeq:     "",
		},
		{
			name:        "in sequence directory",
			currentPath: filepath.Join(festivalPath, "01_Planning", "01_core"),
			wantPhase:   filepath.Join(festivalPath, "01_Planning"),
			wantSeq:     filepath.Join(festivalPath, "01_Planning", "01_core"),
		},
		{
			name:        "outside festival",
			currentPath: "/other/path",
			wantPhase:   "",
			wantSeq:     "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			loc := s.determineLocation(tc.currentPath)

			if loc.FestivalPath != festivalPath {
				t.Errorf("FestivalPath = %q, want %q", loc.FestivalPath, festivalPath)
			}
			if loc.PhasePath != tc.wantPhase {
				t.Errorf("PhasePath = %q, want %q", loc.PhasePath, tc.wantPhase)
			}
			if loc.SequencePath != tc.wantSeq {
				t.Errorf("SequencePath = %q, want %q", loc.SequencePath, tc.wantSeq)
			}
		})
	}
}

func TestIsNumberedDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"01_test", true},
		{"1_test", true}, // starts with digit
		{"00_test", true},
		{"99_test", true},
		{"abc_test", false},
		{".hidden", false},
		{"_private", false},
		{"", false},
		{"a", false},
		{"1", false}, // too short
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isNumberedDir(tc.name)
			if got != tc.want {
				t.Errorf("isNumberedDir(%q) = %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}

func TestNextTaskResult_Complete(t *testing.T) {
	result := &NextTaskResult{
		FestivalComplete: true,
		Reason:           "All tasks complete",
		Location: &LocationInfo{
			FestivalPath: "/test",
		},
	}

	// Test text format
	text := FormatText(result)
	if text == "" {
		t.Error("FormatText returned empty string")
	}
	if !contains(text, "Complete") {
		t.Error("Expected 'Complete' in output")
	}

	// Test short format
	short := FormatShort(result)
	if short != "Festival complete" {
		t.Errorf("FormatShort = %q, want 'Festival complete'", short)
	}
}

func TestNextTaskResult_BlockingGate(t *testing.T) {
	result := &NextTaskResult{
		BlockingGate: &GateInfo{
			Phase:       "01_Planning",
			GateType:    "phase_transition",
			Description: "Quality gate required",
		},
		Location: &LocationInfo{
			FestivalPath: "/test",
		},
	}

	text := FormatText(result)
	if !contains(text, "Quality Gate") {
		t.Error("Expected 'Quality Gate' in output")
	}

	short := FormatShort(result)
	if !contains(short, "Blocked") {
		t.Error("Expected 'Blocked' in short output")
	}
}

func TestNextTaskResult_WithTask(t *testing.T) {
	result := &NextTaskResult{
		Task: &TaskInfo{
			Name:         "01_test_task",
			Path:         "/test/01_Planning/01_core/01_test_task.md",
			Number:       1,
			SequenceName: "01_core",
			PhaseName:    "01_Planning",
			Status:       "pending",
		},
		Reason: "Next task in sequence",
		Location: &LocationInfo{
			FestivalPath: "/test",
		},
	}

	text := FormatText(result)
	if !contains(text, "01_test_task") {
		t.Error("Expected task name in output")
	}

	short := FormatShort(result)
	if short != result.Task.Path {
		t.Errorf("FormatShort = %q, want %q", short, result.Task.Path)
	}

	cd := FormatCD(result)
	if cd != "/test/01_Planning/01_core" {
		t.Errorf("FormatCD = %q, want directory path", cd)
	}
}

func TestFormatJSON(t *testing.T) {
	result := &NextTaskResult{
		Task: &TaskInfo{
			Name:   "test_task",
			Number: 1,
		},
		Reason: "test",
		Location: &LocationInfo{
			FestivalPath: "/test",
		},
	}

	json, err := FormatJSON(result)
	if err != nil {
		t.Fatalf("FormatJSON error: %v", err)
	}

	if !contains(json, "test_task") {
		t.Error("Expected task name in JSON output")
	}
}

func TestProgressStats(t *testing.T) {
	stats := &ProgressStats{
		TotalTasks:      10,
		CompletedTasks:  5,
		InProgressTasks: 2,
		PendingTasks:    3,
		PercentComplete: 50.0,
	}

	if stats.TotalTasks != 10 {
		t.Errorf("TotalTasks = %d, want 10", stats.TotalTasks)
	}
	if stats.PercentComplete != 50.0 {
		t.Errorf("PercentComplete = %f, want 50.0", stats.PercentComplete)
	}
}

func TestSelector_FindNext_Integration(t *testing.T) {
	// Skip if no test festival available
	testFestival := os.Getenv("TEST_FESTIVAL_PATH")
	if testFestival == "" {
		t.Skip("TEST_FESTIVAL_PATH not set")
	}

	s := NewSelector(testFestival)
	result, err := s.FindNext(testFestival)
	if err != nil {
		t.Fatalf("FindNext error: %v", err)
	}

	if result == nil {
		t.Fatal("FindNext returned nil result")
	}

	// Should have either a task, blocking gate, or be complete
	hasResult := result.Task != nil || result.BlockingGate != nil || result.FestivalComplete
	if !hasResult {
		t.Logf("No task available: %s", result.Reason)
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
