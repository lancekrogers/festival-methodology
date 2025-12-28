package festival

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure
	festivalsDir := filepath.Join(tmpDir, "festivals")
	activeDir := filepath.Join(festivalsDir, "active")
	festivalDir := filepath.Join(activeDir, "my-festival")
	phase1 := filepath.Join(festivalDir, "001_PLAN")
	seq1 := filepath.Join(phase1, "01_requirements")

	// Create directories
	if err := os.MkdirAll(seq1, 0755); err != nil {
		t.Fatal(err)
	}

	// Create festival marker
	if err := os.WriteFile(filepath.Join(festivalDir, "FESTIVAL_GOAL.md"), []byte("# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		startDir     string
		wantFestival string
		wantPhase    string
		wantSequence string
	}{
		{
			name:         "at festival root",
			startDir:     festivalDir,
			wantFestival: festivalDir,
			wantPhase:    "",
			wantSequence: "",
		},
		{
			name:         "inside phase",
			startDir:     phase1,
			wantFestival: festivalDir,
			wantPhase:    phase1,
			wantSequence: "",
		},
		{
			name:         "inside sequence",
			startDir:     seq1,
			wantFestival: festivalDir,
			wantPhase:    phase1,
			wantSequence: seq1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := DetectContext(tc.startDir)
			if err != nil {
				t.Fatalf("DetectContext() error: %v", err)
			}
			if ctx.FestivalDir != tc.wantFestival {
				t.Errorf("FestivalDir = %q, want %q", ctx.FestivalDir, tc.wantFestival)
			}
			if ctx.PhaseDir != tc.wantPhase {
				t.Errorf("PhaseDir = %q, want %q", ctx.PhaseDir, tc.wantPhase)
			}
			if ctx.SequenceDir != tc.wantSequence {
				t.Errorf("SequenceDir = %q, want %q", ctx.SequenceDir, tc.wantSequence)
			}
		})
	}
}

func TestResolvePhase(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival with phases
	festivalDir := tmpDir
	phase1 := filepath.Join(festivalDir, "001_PLANNING")
	phase2 := filepath.Join(festivalDir, "002_IMPLEMENT")

	for _, d := range []string{phase1, phase2} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"1", phase1, false},
		{"01", phase1, false},
		{"001", phase1, false},
		{"2", phase2, false},
		{"002", phase2, false},
		{phase1, phase1, false}, // Full path
		{"001_PLANNING", phase1, false},
		{"99", "", true}, // Not found
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ResolvePhase(tc.input, festivalDir)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ResolvePhase(%q) expected error, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ResolvePhase(%q) unexpected error: %v", tc.input, err)
				} else if result != tc.expected {
					t.Errorf("ResolvePhase(%q) = %q, want %q", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestResolveSequence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phase with sequences
	phaseDir := tmpDir
	seq1 := filepath.Join(phaseDir, "01_requirements")
	seq2 := filepath.Join(phaseDir, "02_design")

	for _, d := range []string{seq1, seq2} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		input    string
		expected string
		wantErr  bool
	}{
		{"1", seq1, false},
		{"01", seq1, false},
		{"2", seq2, false},
		{"02", seq2, false},
		{seq1, seq1, false}, // Full path
		{"01_requirements", seq1, false},
		{"99", "", true}, // Not found
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			result, err := ResolveSequence(tc.input, phaseDir)
			if tc.wantErr {
				if err == nil {
					t.Errorf("ResolveSequence(%q) expected error, got nil", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ResolveSequence(%q) unexpected error: %v", tc.input, err)
				} else if result != tc.expected {
					t.Errorf("ResolveSequence(%q) = %q, want %q", tc.input, result, tc.expected)
				}
			}
		})
	}
}

func TestIsNumericShortcut(t *testing.T) {
	tests := []struct {
		input     string
		maxDigits int
		expected  bool
	}{
		{"1", 3, true},
		{"01", 3, true},
		{"001", 3, true},
		{"1234", 3, false}, // Too long
		{"abc", 3, false},  // Not numeric
		{"1a", 2, false},   // Mixed
		{"", 3, false},     // Empty
		{"1", 2, true},
		{"12", 2, true},
		{"123", 2, false}, // Too long for 2 digits
	}

	for _, tc := range tests {
		result := isNumericShortcut(tc.input, tc.maxDigits)
		if result != tc.expected {
			t.Errorf("isNumericShortcut(%q, %d) = %v, want %v", tc.input, tc.maxDigits, result, tc.expected)
		}
	}
}

func TestListPhases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phases
	phases := []string{"001_PLAN", "002_IMPLEMENT", "003_REVIEW"}
	for _, p := range phases {
		if err := os.MkdirAll(filepath.Join(tmpDir, p), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create non-phase directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "not_a_phase"), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := ListPhases(tmpDir)
	if err != nil {
		t.Fatalf("ListPhases() error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("ListPhases() returned %d phases, want 3", len(result))
	}
}

func TestListSequences(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sequences
	sequences := []string{"01_setup", "02_build", "03_test"}
	for _, s := range sequences {
		if err := os.MkdirAll(filepath.Join(tmpDir, s), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Create non-sequence directory
	if err := os.MkdirAll(filepath.Join(tmpDir, "not_a_sequence"), 0755); err != nil {
		t.Fatal(err)
	}

	result, err := ListSequences(tmpDir)
	if err != nil {
		t.Fatalf("ListSequences() error: %v", err)
	}

	if len(result) != 3 {
		t.Errorf("ListSequences() returned %d sequences, want 3", len(result))
	}
}
