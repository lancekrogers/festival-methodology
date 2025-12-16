package commands

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsPhaseShortcut(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"01", true},
		{"001", true},
		{"2", true},
		{"12", true},
		{"123", true},
		{"", false},
		{"1234", false},     // Too long
		{"abc", false},      // Not numeric
		{"1a", false},       // Mixed
		{"a1", false},       // Mixed
		{"001_PLAN", false}, // Full phase name
	}

	for _, tc := range tests {
		result := isPhaseShortcut(tc.input)
		if result != tc.expected {
			t.Errorf("isPhaseShortcut(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestIsSequenceShortcut(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"01", true},
		{"12", true},
		{"", false},
		{"123", false},      // Too long for sequence (max 2 digits)
		{"abc", false},      // Not numeric
		{"01_setup", false}, // Full sequence name
		{"1a", false},       // Mixed
	}

	for _, tc := range tests {
		result := isSequenceShortcut(tc.input)
		if result != tc.expected {
			t.Errorf("isSequenceShortcut(%q) = %v, want %v", tc.input, result, tc.expected)
		}
	}
}

func TestResolvePhaseShortcut(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure with phases
	festivalsDir := filepath.Join(tmpDir, "festivals")
	activeDir := filepath.Join(festivalsDir, "active")
	plannedDir := filepath.Join(festivalsDir, "planned")

	// Create some phases
	phases := []string{
		filepath.Join(activeDir, "my-festival", "001_PLAN"),
		filepath.Join(activeDir, "my-festival", "002_IMPLEMENT"),
		filepath.Join(activeDir, "my-festival", "003_REVIEW"),
		filepath.Join(plannedDir, "another", "001_DISCOVERY"),
	}

	for _, p := range phases {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		shortcut string
		expected string
		wantErr  bool
	}{
		{"1", filepath.Join(activeDir, "my-festival", "001_PLAN"), false},
		{"01", filepath.Join(activeDir, "my-festival", "001_PLAN"), false},
		{"001", filepath.Join(activeDir, "my-festival", "001_PLAN"), false},
		{"2", filepath.Join(activeDir, "my-festival", "002_IMPLEMENT"), false},
		{"3", filepath.Join(activeDir, "my-festival", "003_REVIEW"), false},
		{"999", "", true}, // Non-existent phase
	}

	for _, tc := range tests {
		result, err := resolvePhaseShortcut(tc.shortcut, filepath.Join(activeDir, "my-festival"))
		if tc.wantErr {
			if err == nil {
				t.Errorf("resolvePhaseShortcut(%q) expected error, got nil", tc.shortcut)
			}
		} else {
			if err != nil {
				t.Errorf("resolvePhaseShortcut(%q) unexpected error: %v", tc.shortcut, err)
			} else if result != tc.expected {
				t.Errorf("resolvePhaseShortcut(%q) = %q, want %q", tc.shortcut, result, tc.expected)
			}
		}
	}
}

func TestResolveSequenceShortcut(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phase with sequences
	phaseDir := filepath.Join(tmpDir, "001_PLAN")
	sequences := []string{
		filepath.Join(phaseDir, "01_requirements"),
		filepath.Join(phaseDir, "02_design"),
		filepath.Join(phaseDir, "03_review"),
	}

	for _, s := range sequences {
		if err := os.MkdirAll(s, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		shortcut string
		expected string
		wantErr  bool
	}{
		{"1", filepath.Join(phaseDir, "01_requirements"), false},
		{"01", filepath.Join(phaseDir, "01_requirements"), false},
		{"2", filepath.Join(phaseDir, "02_design"), false},
		{"3", filepath.Join(phaseDir, "03_review"), false},
		{"99", "", true}, // Non-existent sequence
	}

	for _, tc := range tests {
		result, err := resolveSequenceShortcut(tc.shortcut, phaseDir)
		if tc.wantErr {
			if err == nil {
				t.Errorf("resolveSequenceShortcut(%q) expected error, got nil", tc.shortcut)
			}
		} else {
			if err != nil {
				t.Errorf("resolveSequenceShortcut(%q) unexpected error: %v", tc.shortcut, err)
			} else if result != tc.expected {
				t.Errorf("resolveSequenceShortcut(%q) = %q, want %q", tc.shortcut, result, tc.expected)
			}
		}
	}
}

func TestResolveGoTarget(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure - phases should be in active/
	festivalsDir := filepath.Join(tmpDir, "festivals")
	activeDir := filepath.Join(festivalsDir, "active")

	// Create phases directly in active/ (not nested in another festival dir)
	phase1 := filepath.Join(activeDir, "001_PLAN")
	seq1 := filepath.Join(phase1, "01_requirements")

	for _, d := range []string{seq1, filepath.Join(festivalsDir, "custom", "path")} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		target   string
		expected string
		wantErr  bool
	}{
		// Phase shortcuts - resolvePhaseShortcut searches active/, planned/, completed/
		{"1", phase1, false},
		{"001", phase1, false},

		// Relative paths
		{"custom/path", filepath.Join(festivalsDir, "custom", "path"), false},
		{"active", activeDir, false},

		// Non-existent
		{"nonexistent", "", true},
	}

	for _, tc := range tests {
		result, err := resolveGoTarget(tc.target, festivalsDir)
		if tc.wantErr {
			if err == nil {
				t.Errorf("resolveGoTarget(%q) expected error, got nil", tc.target)
			}
		} else {
			if err != nil {
				t.Errorf("resolveGoTarget(%q) unexpected error: %v", tc.target, err)
			} else if result != tc.expected {
				t.Errorf("resolveGoTarget(%q) = %q, want %q", tc.target, result, tc.expected)
			}
		}
	}
}

func TestResolveGoTargetWithSequence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals structure with phase and sequence
	// Phases should be in active/ directly
	festivalsDir := filepath.Join(tmpDir, "festivals")
	activeDir := filepath.Join(festivalsDir, "active")
	phase1 := filepath.Join(activeDir, "001_PLAN")
	seq1 := filepath.Join(phase1, "01_requirements")
	seq2 := filepath.Join(phase1, "02_design")

	for _, d := range []string{seq1, seq2} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		target   string
		expected string
		wantErr  bool
	}{
		// Phase/sequence shortcuts
		{"1/1", seq1, false},
		{"1/01", seq1, false},
		{"001/01", seq1, false},
		{"1/2", seq2, false},

		// Invalid sequence
		{"1/99", "", true},
	}

	for _, tc := range tests {
		result, err := resolveGoTarget(tc.target, festivalsDir)
		if tc.wantErr {
			if err == nil {
				t.Errorf("resolveGoTarget(%q) expected error, got nil", tc.target)
			}
		} else {
			if err != nil {
				t.Errorf("resolveGoTarget(%q) unexpected error: %v", tc.target, err)
			} else if result != tc.expected {
				t.Errorf("resolveGoTarget(%q) = %q, want %q", tc.target, result, tc.expected)
			}
		}
	}
}
