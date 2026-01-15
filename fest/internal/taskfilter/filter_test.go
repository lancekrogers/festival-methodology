package taskfilter

import (
	"testing"
)

func TestClassifyFile(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     FileType
	}{
		// Goal files
		{"sequence goal", "SEQUENCE_GOAL.md", FileTypeGoal},
		{"phase goal", "PHASE_GOAL.md", FileTypeGoal},
		{"festival goal", "FESTIVAL_GOAL.md", FileTypeGoal},
		{"festival overview", "FESTIVAL_OVERVIEW.md", FileTypeGoal},
		{"todo", "TODO.md", FileTypeGoal},
		{"context", "CONTEXT.md", FileTypeGoal},
		{"festival rules", "FESTIVAL_RULES.md", FileTypeGoal},

		// Regular tasks
		{"simple task", "01_design.md", FileTypeTask},
		{"numbered task", "02_implement.md", FileTypeTask},
		{"task with multiple words", "03_test_the_feature.md", FileTypeTask},
		{"decimal task", "01.5_hotfix.md", FileTypeTask},
		{"high number task", "99_final_cleanup.md", FileTypeTask},

		// Quality gate files
		{"testing gate", "04_testing_and_verify.md", FileTypeGate},
		{"code review gate", "05_code_review.md", FileTypeGate},
		{"review results gate", "06_review_results_iterate.md", FileTypeGate},
		{"commit gate", "07_commit.md", FileTypeGate},
		{"generic gate", "08_quality_gate.md", FileTypeGate},
		{"uppercase gate", "04_TESTING_AND_VERIFY.md", FileTypeGate},

		// Unknown/ignored files
		{"readme", "README.md", FileTypeUnknown},
		{"random file", "notes.md", FileTypeUnknown},
		{"non-markdown", "01_design.txt", FileTypeUnknown},
		{"no extension", "01_design", FileTypeUnknown},
		{"single digit prefix", "1_task.md", FileTypeUnknown},
		{"three digit prefix file", "001_task.md", FileTypeUnknown},
		{"hidden file", ".01_hidden.md", FileTypeUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFile(tt.filename)
			if got != tt.want {
				t.Errorf("ClassifyFile(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsTask(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Should be tracked as tasks (includes gates)
		{"regular task", "01_design.md", true},
		{"gate file", "04_testing_and_verify.md", true},
		{"commit gate", "07_commit.md", true},

		// Should not be tracked
		{"goal file", "SEQUENCE_GOAL.md", false},
		{"readme", "README.md", false},
		{"random file", "notes.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTask(tt.filename)
			if got != tt.want {
				t.Errorf("IsTask(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsTaskOnly(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Regular tasks only
		{"regular task", "01_design.md", true},
		{"implement task", "02_implement.md", true},

		// Gates should NOT match
		{"gate file", "04_testing_and_verify.md", false},
		{"commit gate", "07_commit.md", false},

		// Goals should NOT match
		{"goal file", "SEQUENCE_GOAL.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTaskOnly(tt.filename)
			if got != tt.want {
				t.Errorf("IsTaskOnly(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsGate(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Gate files
		{"testing gate", "04_testing_and_verify.md", true},
		{"code review", "05_code_review.md", true},
		{"review results", "06_review_results_iterate.md", true},
		{"commit gate", "07_commit.md", true},
		{"quality gate", "08_quality_gate.md", true},

		// Non-gate files
		{"regular task", "01_design.md", false},
		{"goal file", "SEQUENCE_GOAL.md", false},
		{"readme", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGate(tt.filename)
			if got != tt.want {
				t.Errorf("IsGate(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsGoal(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Goal files
		{"sequence goal", "SEQUENCE_GOAL.md", true},
		{"phase goal", "PHASE_GOAL.md", true},
		{"festival goal", "FESTIVAL_GOAL.md", true},
		{"todo", "TODO.md", true},

		// Non-goal files
		{"regular task", "01_design.md", false},
		{"gate", "04_testing_and_verify.md", false},
		{"readme", "README.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsGoal(tt.filename)
			if got != tt.want {
				t.Errorf("IsGoal(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsPhaseDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"valid phase", "001_PLANNING", true},
		{"valid phase 2", "002_IMPLEMENTATION", true},
		{"valid phase 3", "999_FINAL", true},
		{"invalid two digits", "01_setup", false},
		{"invalid four digits", "0001_setup", false},
		{"no underscore", "001planning", false},
		{"no digits", "planning", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsPhaseDir(tt.dir)
			if got != tt.want {
				t.Errorf("IsPhaseDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestIsSequenceDir(t *testing.T) {
	tests := []struct {
		name string
		dir  string
		want bool
	}{
		{"valid sequence", "01_setup", true},
		{"valid sequence 2", "02_core_feature", true},
		{"valid sequence 99", "99_cleanup", true},
		{"invalid three digits", "001_setup", false},
		{"invalid single digit", "1_setup", false},
		{"no underscore", "01setup", false},
		{"no digits", "setup", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSequenceDir(tt.dir)
			if got != tt.want {
				t.Errorf("IsSequenceDir(%q) = %v, want %v", tt.dir, got, tt.want)
			}
		})
	}
}

func TestShouldTrack(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		// Should be tracked
		{"task", "01_design.md", true},
		{"gate", "04_testing_and_verify.md", true},

		// Should not be tracked
		{"goal", "SEQUENCE_GOAL.md", false},
		{"readme", "README.md", false},
		{"random", "notes.md", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldTrack(tt.filename)
			if got != tt.want {
				t.Errorf("ShouldTrack(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestFileTypeString(t *testing.T) {
	tests := []struct {
		ft   FileType
		want string
	}{
		{FileTypeTask, "task"},
		{FileTypeGate, "gate"},
		{FileTypeGoal, "goal"},
		{FileTypeUnknown, "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.ft.String()
			if got != tt.want {
				t.Errorf("FileType(%d).String() = %q, want %q", tt.ft, got, tt.want)
			}
		})
	}
}

func TestClassifyPath(t *testing.T) {
	tests := []struct {
		name string
		path string
		want FileInfo
	}{
		{
			name: "task with path",
			path: "/festivals/active/my-fest/001_PHASE/01_seq/01_design.md",
			want: FileInfo{
				Name:    "01_design.md",
				Path:    "/festivals/active/my-fest/001_PHASE/01_seq/01_design.md",
				Type:    FileTypeTask,
				IsTask:  true,
				IsGate:  false,
				IsGoal:  false,
				Tracked: true,
			},
		},
		{
			name: "gate with path",
			path: "/festivals/active/my-fest/001_PHASE/01_seq/04_testing_and_verify.md",
			want: FileInfo{
				Name:    "04_testing_and_verify.md",
				Path:    "/festivals/active/my-fest/001_PHASE/01_seq/04_testing_and_verify.md",
				Type:    FileTypeGate,
				IsTask:  false,
				IsGate:  true,
				IsGoal:  false,
				Tracked: true,
			},
		},
		{
			name: "goal with path",
			path: "/festivals/active/my-fest/001_PHASE/01_seq/SEQUENCE_GOAL.md",
			want: FileInfo{
				Name:    "SEQUENCE_GOAL.md",
				Path:    "/festivals/active/my-fest/001_PHASE/01_seq/SEQUENCE_GOAL.md",
				Type:    FileTypeGoal,
				IsTask:  false,
				IsGate:  false,
				IsGoal:  true,
				Tracked: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyPath(tt.path)
			if got.Name != tt.want.Name {
				t.Errorf("ClassifyPath(%q).Name = %q, want %q", tt.path, got.Name, tt.want.Name)
			}
			if got.Path != tt.want.Path {
				t.Errorf("ClassifyPath(%q).Path = %q, want %q", tt.path, got.Path, tt.want.Path)
			}
			if got.Type != tt.want.Type {
				t.Errorf("ClassifyPath(%q).Type = %v, want %v", tt.path, got.Type, tt.want.Type)
			}
			if got.IsTask != tt.want.IsTask {
				t.Errorf("ClassifyPath(%q).IsTask = %v, want %v", tt.path, got.IsTask, tt.want.IsTask)
			}
			if got.IsGate != tt.want.IsGate {
				t.Errorf("ClassifyPath(%q).IsGate = %v, want %v", tt.path, got.IsGate, tt.want.IsGate)
			}
			if got.IsGoal != tt.want.IsGoal {
				t.Errorf("ClassifyPath(%q).IsGoal = %v, want %v", tt.path, got.IsGoal, tt.want.IsGoal)
			}
			if got.Tracked != tt.want.Tracked {
				t.Errorf("ClassifyPath(%q).Tracked = %v, want %v", tt.path, got.Tracked, tt.want.Tracked)
			}
		})
	}
}

// TestEdgeCases tests various edge cases for robustness
func TestEdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		wantType FileType
	}{
		{"empty string", "", FileTypeUnknown},
		{"just extension", ".md", FileTypeUnknown},
		{"unicode task", "01_デザイン.md", FileTypeTask},
		{"spaces in name", "01_my task.md", FileTypeTask},
		{"multiple dots", "01.2.3_task.md", FileTypeTask},             // Regex allows this pattern
		{"review in task name", "01_review_feature.md", FileTypeTask}, // review not at gate position
		{"commit in task name", "01_commit_changes.md", FileTypeTask}, // only exact "commit" is a gate
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClassifyFile(tt.filename)
			if got != tt.wantType {
				t.Errorf("ClassifyFile(%q) = %v, want %v", tt.filename, got, tt.wantType)
			}
		})
	}
}
