package id

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractInitials(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		// Standard cases
		{"two words", "guild usable", "GU"},
		{"three words", "my big project", "MB"},
		{"four words", "very long project name", "VL"},

		// Single word - use first two letters
		{"single word short", "api", "AP"},
		{"single word long", "onboarding", "ON"},

		// Hyphenated names
		{"with hyphen", "my-big-project", "MB"},
		{"double hyphen", "my--project", "MP"},

		// Skip common words
		{"skip the", "the big project", "BP"},
		{"skip a", "a new feature", "NF"},
		{"skip an", "an awesome tool", "AT"},
		{"skip of", "tower of hanoi", "TH"},
		{"skip for", "tools for testing", "TT"},

		// Case handling
		{"all caps", "MY PROJECT", "MP"},
		{"mixed case no space", "MyProject", "MY"}, // treated as single word
		{"all lowercase", "myproject", "MY"},

		// Edge cases
		{"empty string", "", "XX"},
		{"whitespace only", "   ", "XX"},
		{"single char", "x", "XX"},
		{"numbers only", "123", "XX"},
		{"special chars", "!@#$%", "XX"},

		// Unicode
		{"with unicode", "café tools", "CT"},

		// Real festival names
		{"fest-node-ids", "fest-node-ids", "FN"},
		{"guild-usable", "guild-usable", "GU"},
		{"cli-ux-improvements", "cli-ux-improvements", "CU"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractInitials(tt.input)
			if got != tt.expected {
				t.Errorf("ExtractInitials(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestExtractInitials_AlwaysTwoChars(t *testing.T) {
	// Verify all outputs are exactly 2 uppercase characters
	inputs := []string{
		"x", "xy", "xyz", "a b c d e f",
		"the", "a", "an", // common words only
	}

	for _, input := range inputs {
		got := ExtractInitials(input)
		if len(got) != 2 {
			t.Errorf("ExtractInitials(%q) = %q, length = %d, want 2", input, got, len(got))
		}
		if got[0] < 'A' || got[0] > 'Z' || got[1] < 'A' || got[1] > 'Z' {
			t.Errorf("ExtractInitials(%q) = %q, not uppercase letters", input, got)
		}
	}
}

func TestFindNextCounter(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create mock festival directories
	statuses := []string{"planned", "active", "completed/2025-01"}
	for _, status := range statuses {
		if err := os.MkdirAll(filepath.Join(tmpDir, status), 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
	}

	tests := []struct {
		name          string
		prefix        string
		existingDirs  map[string][]string // status -> dirs
		expectedCount int
	}{
		{
			name:          "no existing",
			prefix:        "GU",
			existingDirs:  map[string][]string{},
			expectedCount: 1,
		},
		{
			name:   "one existing",
			prefix: "GU",
			existingDirs: map[string][]string{
				"active": {"guild-usable-GU0001"},
			},
			expectedCount: 2,
		},
		{
			name:   "multiple existing same prefix",
			prefix: "GU",
			existingDirs: map[string][]string{
				"active":  {"guild-usable-GU0001", "guild-ui-GU0002"},
				"planned": {"guild-v2-GU0003"},
			},
			expectedCount: 4,
		},
		{
			name:   "different prefixes",
			prefix: "FN",
			existingDirs: map[string][]string{
				"active": {"guild-usable-GU0001", "fest-node-FN0001"},
			},
			expectedCount: 2,
		},
		{
			name:   "gap in counter",
			prefix: "GU",
			existingDirs: map[string][]string{
				"active": {"guild-usable-GU0001", "guild-v3-GU0005"},
			},
			expectedCount: 6, // Next after highest
		},
		{
			name:   "in completed subdirectory",
			prefix: "GU",
			existingDirs: map[string][]string{
				"completed/2025-01": {"guild-old-GU0010"},
			},
			expectedCount: 11,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean and recreate
			for _, status := range statuses {
				statusDir := filepath.Join(tmpDir, status)
				entries, _ := os.ReadDir(statusDir)
				for _, e := range entries {
					os.RemoveAll(filepath.Join(statusDir, e.Name()))
				}
			}

			// Create test directories
			for status, dirs := range tt.existingDirs {
				for _, dir := range dirs {
					path := filepath.Join(tmpDir, status, dir)
					if err := os.MkdirAll(path, 0755); err != nil {
						t.Fatalf("Failed to create %s: %v", path, err)
					}
				}
			}

			got, err := FindNextCounter(tmpDir, tt.prefix)
			if err != nil {
				t.Fatalf("FindNextCounter() error = %v", err)
			}
			if got != tt.expectedCount {
				t.Errorf("FindNextCounter(%q, %q) = %d, want %d",
					tmpDir, tt.prefix, got, tt.expectedCount)
			}
		})
	}
}

func TestGenerateID(t *testing.T) {
	tmpDir := t.TempDir()

	// Create status directories
	for _, status := range []string{"planned", "active", "completed", "dungeon"} {
		os.MkdirAll(filepath.Join(tmpDir, status), 0755)
	}

	tests := []struct {
		name         string
		festivalName string
		existingDirs map[string][]string
		expectedID   string
	}{
		{
			name:         "first festival",
			festivalName: "guild-usable",
			existingDirs: map[string][]string{},
			expectedID:   "GU0001",
		},
		{
			name:         "second with same prefix",
			festivalName: "guild-ui",
			existingDirs: map[string][]string{
				"active": {"guild-usable-GU0001"},
			},
			expectedID: "GU0002",
		},
		{
			name:         "different prefix",
			festivalName: "fest-node-ids",
			existingDirs: map[string][]string{
				"active": {"guild-usable-GU0001"},
			},
			expectedID: "FN0001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean existing
			for _, status := range []string{"planned", "active", "completed", "dungeon"} {
				statusDir := filepath.Join(tmpDir, status)
				entries, _ := os.ReadDir(statusDir)
				for _, e := range entries {
					os.RemoveAll(filepath.Join(statusDir, e.Name()))
				}
			}

			// Create existing directories
			for status, dirs := range tt.existingDirs {
				for _, dir := range dirs {
					os.MkdirAll(filepath.Join(tmpDir, status, dir), 0755)
				}
			}

			got, err := GenerateID(tt.festivalName, tmpDir)
			if err != nil {
				t.Fatalf("GenerateID() error = %v", err)
			}
			if got != tt.expectedID {
				t.Errorf("GenerateID(%q, %q) = %q, want %q",
					tt.festivalName, tmpDir, got, tt.expectedID)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		prefix   string
		counter  int
		expected string
	}{
		{"GU", 1, "GU0001"},
		{"FN", 42, "FN0042"},
		{"AB", 999, "AB0999"},
		{"XY", 9999, "XY9999"},
		{"ZZ", 10000, "ZZ10000"}, // overflow case
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatID(tt.prefix, tt.counter)
			if got != tt.expected {
				t.Errorf("FormatID(%q, %d) = %q, want %q",
					tt.prefix, tt.counter, got, tt.expected)
			}
		})
	}
}

func TestParseID(t *testing.T) {
	tests := []struct {
		id            string
		expectPrefix  string
		expectCounter int
		expectErr     bool
	}{
		{"GU0001", "GU", 1, false},
		{"FN0042", "FN", 42, false},
		{"AB9999", "AB", 9999, false},
		{"invalid", "", 0, true},
		{"G0001", "", 0, true},  // prefix too short
		{"GU001", "", 0, true},  // counter too short
		{"GU000a", "", 0, true}, // non-numeric counter
		{"", "", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			prefix, counter, err := ParseID(tt.id)
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseID(%q) error = %v, wantErr %v", tt.id, err, tt.expectErr)
				return
			}
			if !tt.expectErr {
				if prefix != tt.expectPrefix {
					t.Errorf("ParseID(%q) prefix = %q, want %q", tt.id, prefix, tt.expectPrefix)
				}
				if counter != tt.expectCounter {
					t.Errorf("ParseID(%q) counter = %d, want %d", tt.id, counter, tt.expectCounter)
				}
			}
		})
	}
}

func TestExtractIDFromDirName(t *testing.T) {
	tests := []struct {
		dirName    string
		expectedID string
		expectErr  bool
	}{
		{"guild-usable-GU0001", "GU0001", false},
		{"fest-node-ids-FN0042", "FN0042", false},
		{"my-project-AB9999", "AB9999", false},
		{"no-id-here", "", true},
		{"invalid-XX", "", true},
		{"", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.dirName, func(t *testing.T) {
			got, err := ExtractIDFromDirName(tt.dirName)
			if (err != nil) != tt.expectErr {
				t.Errorf("ExtractIDFromDirName(%q) error = %v, wantErr %v", tt.dirName, err, tt.expectErr)
				return
			}
			if got != tt.expectedID {
				t.Errorf("ExtractIDFromDirName(%q) = %q, want %q", tt.dirName, got, tt.expectedID)
			}
		})
	}
}
