package progress

import (
	"os"
	"path/filepath"
	"testing"
)

// boolPtr is a helper to create a pointer to a bool
func boolPtr(b bool) *bool {
	return &b
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantTracking *bool
		wantNil     bool
	}{
		{
			name:         "tracking false",
			content:      "---\ntracking: false\n---\n# Title",
			wantTracking: boolPtr(false),
			wantNil:      false,
		},
		{
			name:         "tracking true",
			content:      "---\ntracking: true\n---\n# Title",
			wantTracking: boolPtr(true),
			wantNil:      false,
		},
		{
			name:         "no frontmatter",
			content:      "# Title\nContent",
			wantTracking: nil,
			wantNil:      true,
		},
		{
			name:         "empty frontmatter",
			content:      "---\n---\n# Title",
			wantTracking: nil,
			wantNil:      false,
		},
		{
			name:         "frontmatter without tracking",
			content:      "---\ntitle: Test\nauthor: Me\n---\n# Title",
			wantTracking: nil,
			wantNil:      false,
		},
		{
			name:         "empty file",
			content:      "",
			wantTracking: nil,
			wantNil:      true,
		},
		{
			name:         "only opening delimiter",
			content:      "---\nsome content without closing",
			wantTracking: nil,
			wantNil:      true,
		},
		{
			name:         "malformed yaml in frontmatter",
			content:      "---\n{invalid yaml: [\n---\n# Title",
			wantTracking: nil,
			wantNil:      true, // Malformed YAML returns nil (fail-safe)
		},
		{
			name:         "frontmatter with other fields and tracking",
			content:      "---\ntitle: My Doc\ntracking: false\nauthor: Test\n---\n# Content",
			wantTracking: boolPtr(false),
			wantNil:      false,
		},
		{
			name:         "whitespace around delimiters",
			content:      "---  \ntracking: true\n  ---\n# Title",
			wantTracking: boolPtr(true),
			wantNil:      false, // Whitespace is trimmed, valid frontmatter
		},
		{
			name:         "content starting with dashes but not frontmatter",
			content:      "---- not frontmatter\nContent",
			wantTracking: nil,
			wantNil:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Parse frontmatter
			got, err := ParseFrontmatter(tmpFile)
			if err != nil {
				t.Errorf("ParseFrontmatter() error = %v", err)
				return
			}

			// Check if nil as expected
			if tt.wantNil && got != nil {
				t.Errorf("ParseFrontmatter() = %v, want nil", got)
				return
			}
			if !tt.wantNil && got == nil {
				t.Errorf("ParseFrontmatter() = nil, want non-nil")
				return
			}
			if tt.wantNil {
				return // Both are nil, test passes
			}

			// Compare tracking values
			if tt.wantTracking == nil && got.Tracking != nil {
				t.Errorf("ParseFrontmatter().Tracking = %v, want nil", *got.Tracking)
				return
			}
			if tt.wantTracking != nil && got.Tracking == nil {
				t.Errorf("ParseFrontmatter().Tracking = nil, want %v", *tt.wantTracking)
				return
			}
			if tt.wantTracking != nil && got.Tracking != nil {
				if *tt.wantTracking != *got.Tracking {
					t.Errorf("ParseFrontmatter().Tracking = %v, want %v", *got.Tracking, *tt.wantTracking)
				}
			}
		})
	}
}

func TestParseFrontmatter_FileNotFound(t *testing.T) {
	_, err := ParseFrontmatter("/nonexistent/path/file.md")
	if err == nil {
		t.Error("ParseFrontmatter() expected error for nonexistent file, got nil")
	}
}

func TestIsTracked(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "tracking false - should not be tracked",
			content: "---\ntracking: false\n---\n# Title",
			want:    false,
		},
		{
			name:    "tracking true - should be tracked",
			content: "---\ntracking: true\n---\n# Title",
			want:    true,
		},
		{
			name:    "no frontmatter - default tracked",
			content: "# Title\nContent",
			want:    true,
		},
		{
			name:    "empty frontmatter - default tracked",
			content: "---\n---\n# Title",
			want:    true,
		},
		{
			name:    "frontmatter without tracking field - default tracked",
			content: "---\ntitle: Test\n---\n# Content",
			want:    true,
		},
		{
			name:    "malformed yaml - fail-safe to tracked",
			content: "---\n{invalid yaml\n---\n# Title",
			want:    true,
		},
		{
			name:    "empty file - default tracked",
			content: "",
			want:    true,
		},
		{
			name:    "fest frontmatter with tracking false",
			content: "---\nfest_type: task\nfest_id: test\ntracking: false\n---\n# Task",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(tmpFile, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create temp file: %v", err)
			}

			// Check IsTracked
			got := IsTracked(tmpFile)
			if got != tt.want {
				t.Errorf("IsTracked() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTracked_FileNotFound(t *testing.T) {
	// Should fail-safe to tracked when file not found
	got := IsTracked("/nonexistent/path/file.md")
	if got != true {
		t.Errorf("IsTracked() for nonexistent file = %v, want true (fail-safe)", got)
	}
}

func TestIsTracked_NonMarkdownFile(t *testing.T) {
	// Test with a binary-like file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.bin")
	if err := os.WriteFile(tmpFile, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	// Should handle gracefully and default to tracked
	got := IsTracked(tmpFile)
	if got != true {
		t.Errorf("IsTracked() for binary file = %v, want true", got)
	}
}
