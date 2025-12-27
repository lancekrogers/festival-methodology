package frontmatter

import (
	"testing"
	"time"
)

func TestHasFrontmatter(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected bool
	}{
		{
			name: "with frontmatter",
			content: `---
fest_type: task
---

# Content`,
			expected: true,
		},
		{
			name:     "without frontmatter",
			content:  "# Just a heading",
			expected: false,
		},
		{
			name: "with whitespace before",
			content: `

---
fest_type: task
---`,
			expected: true,
		},
		{
			name:     "empty content",
			content:  "",
			expected: false,
		},
		{
			name:     "just dashes",
			content:  "--- some text",
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := HasFrontmatter([]byte(tc.content))
			if got != tc.expected {
				t.Errorf("HasFrontmatter() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantFM         bool
		expectedType   Type
		expectedID     string
		remainingStart string
	}{
		{
			name: "valid frontmatter",
			content: `---
fest_type: task
fest_id: 01_test
fest_status: pending
fest_created: 2025-01-01T00:00:00Z
---

# Task

Content here`,
			wantFM:         true,
			expectedType:   TypeTask,
			expectedID:     "01_test",
			remainingStart: "\n# Task", // Remaining includes leading newline after frontmatter
		},
		{
			name:           "no frontmatter",
			content:        "# Just content",
			wantFM:         false,
			remainingStart: "# Just content",
		},
		{
			name: "frontmatter with all fields",
			content: `---
fest_type: festival
fest_id: test-fest
fest_name: "Test Festival"
fest_status: active
fest_priority: high
fest_created: 2025-01-01T00:00:00Z
fest_tags: [v1, test]
---

# Festival`,
			wantFM:         true,
			expectedType:   TypeFestival,
			expectedID:     "test-fest",
			remainingStart: "\n# Festival", // Remaining includes leading newline after frontmatter
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fm, remaining, err := Parse([]byte(tc.content))
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if tc.wantFM {
				if fm == nil {
					t.Fatal("Parse() returned nil frontmatter")
				}
				if fm.Type != tc.expectedType {
					t.Errorf("Type = %q, want %q", fm.Type, tc.expectedType)
				}
				if fm.ID != tc.expectedID {
					t.Errorf("ID = %q, want %q", fm.ID, tc.expectedID)
				}
			} else {
				if fm != nil {
					t.Errorf("Parse() returned frontmatter, want nil")
				}
			}

			if len(tc.remainingStart) > 0 {
				remainStr := string(remaining)
				if len(remainStr) < len(tc.remainingStart) ||
					remainStr[:len(tc.remainingStart)] != tc.remainingStart {
					t.Errorf("Remaining starts with %q, want %q", remainStr[:min(50, len(remainStr))], tc.remainingStart)
				}
			}
		})
	}
}

func TestParse_Error(t *testing.T) {
	// Unclosed frontmatter
	content := `---
fest_type: task
fest_id: test

# Content without closing delimiter`

	_, _, err := Parse([]byte(content))
	if err == nil {
		t.Error("Parse() should return error for unclosed frontmatter")
	}
}

func TestExtract(t *testing.T) {
	content := `---
fest_type: task
fest_id: 01_test
---

# Content`

	fmBytes, remaining, err := Extract([]byte(content))
	if err != nil {
		t.Fatalf("Extract() error = %v", err)
	}

	if fmBytes == nil {
		t.Error("Extract() returned nil frontmatter bytes")
	}

	fmStr := string(fmBytes)
	if !containsStr(fmStr, "fest_type") {
		t.Error("Frontmatter bytes should contain 'fest_type'")
	}

	remainStr := string(remaining)
	if !containsStr(remainStr, "# Content") {
		t.Error("Remaining should contain '# Content'")
	}
}

func TestStripFrontmatter(t *testing.T) {
	content := `---
fest_type: task
---

# Content`

	stripped, err := StripFrontmatter([]byte(content))
	if err != nil {
		t.Fatalf("StripFrontmatter() error = %v", err)
	}

	if HasFrontmatter(stripped) {
		t.Error("Stripped content should not have frontmatter")
	}

	if !containsStr(string(stripped), "# Content") {
		t.Error("Stripped content should contain '# Content'")
	}
}

func TestParseFile(t *testing.T) {
	content := `---
fest_type: task
fest_id: test
fest_status: pending
fest_created: 2025-01-01T00:00:00Z
---

# Task Content`

	fm, remaining, err := ParseFile([]byte(content))
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if fm == nil {
		t.Fatal("ParseFile() returned nil frontmatter")
	}

	if fm.Type != TypeTask {
		t.Errorf("Type = %q, want 'task'", fm.Type)
	}

	if !containsStr(remaining, "# Task Content") {
		t.Error("Remaining should contain '# Task Content'")
	}
}

func TestParse_WithTimestamps(t *testing.T) {
	content := `---
fest_type: task
fest_id: test
fest_status: pending
fest_created: 2025-01-15T10:30:00Z
fest_updated: 2025-01-20T15:45:00Z
---

Content`

	fm, _, err := Parse([]byte(content))
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if fm.Created.IsZero() {
		t.Error("Created should not be zero")
	}

	expectedYear := 2025
	if fm.Created.Year() != expectedYear {
		t.Errorf("Created year = %d, want %d", fm.Created.Year(), expectedYear)
	}

	if fm.Updated.IsZero() {
		t.Error("Updated should not be zero")
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func TestReplace(t *testing.T) {
	original := `---
fest_type: task
fest_id: old
fest_status: pending
fest_created: 2025-01-01T00:00:00Z
---

# Content`

	newFM := &Frontmatter{
		Type:    TypeTask,
		ID:      "new",
		Status:  StatusCompleted,
		Created: time.Now(),
	}

	result, err := Replace([]byte(original), newFM)
	if err != nil {
		t.Fatalf("Replace() error = %v", err)
	}

	fm, _, err := Parse(result)
	if err != nil {
		t.Fatalf("Parse() after Replace error = %v", err)
	}

	if fm.ID != "new" {
		t.Errorf("ID = %q, want 'new'", fm.ID)
	}

	if fm.Status != StatusCompleted {
		t.Errorf("Status = %q, want 'completed'", fm.Status)
	}
}
