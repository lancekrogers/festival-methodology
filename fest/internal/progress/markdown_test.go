package progress

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTaskStatus(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "all checkboxes completed",
			content: `# Task: Test Task

## Definition of Done
- [x] First item completed
- [x] Second item completed
- [x] Third item completed
`,
			expected: StatusCompleted,
		},
		{
			name: "some checkboxes completed",
			content: `# Task: Test Task

## Definition of Done
- [x] First item completed
- [ ] Second item pending
- [x] Third item completed
`,
			expected: StatusInProgress,
		},
		{
			name: "no checkboxes completed",
			content: `# Task: Test Task

## Definition of Done
- [ ] First item pending
- [ ] Second item pending
- [ ] Third item pending
`,
			expected: StatusPending,
		},
		{
			name: "uppercase X checkbox",
			content: `# Task: Test Task

## Definition of Done
- [X] First item completed
- [X] Second item completed
`,
			expected: StatusCompleted,
		},
		{
			name: "requirements section fallback",
			content: `# Task: Test Task

## Requirements
- [x] Requirement 1 met
- [x] Requirement 2 met
`,
			expected: StatusCompleted,
		},
		{
			name: "acceptance criteria section",
			content: `# Task: Test Task

## Acceptance Criteria
- [x] Criteria 1 met
- [ ] Criteria 2 pending
`,
			expected: StatusInProgress,
		},
		{
			name: "emoji checkbox completed",
			content: `# Task: Test Task

## Definition of Done
- [✅] First item completed
- [✅] Second item completed
`,
			expected: StatusCompleted,
		},
		{
			name: "emoji checkbox in progress",
			content: `# Task: Test Task

## Definition of Done
- [✅] First item completed
- [🚧] Second item in progress
`,
			expected: StatusInProgress,
		},
		{
			name: "no checkboxes in file",
			content: `# Task: Test Task

This is just some text without any checkboxes.
`,
			expected: StatusPending,
		},
		{
			name: "checkboxes outside target sections",
			content: `# Task: Test Task

Some intro text with a checkbox:
- [x] This is in the body

## Notes
- [x] This is in notes section
`,
			expected: StatusCompleted, // Falls back to counting all checkboxes
		},
		{
			name: "mixed checkbox formats",
			content: `# Task: Test Task

## Definition of Done
- [x] Lowercase x completed
- [X] Uppercase X completed
- [ ] Still pending
`,
			expected: StatusInProgress,
		},
		{
			name: "asterisk list markers",
			content: `# Task: Test Task

## Definition of Done
* [x] First item
* [x] Second item
`,
			expected: StatusCompleted,
		},
		{
			name: "indented checkboxes",
			content: `# Task: Test Task

## Definition of Done
  - [x] Indented checkbox 1
    - [x] Nested checkbox
  - [x] Indented checkbox 2
`,
			expected: StatusCompleted,
		},
		{
			name: "deliverables section",
			content: `# Task: Test Task

## Deliverables
- [x] Deliverable 1
- [x] Deliverable 2
`,
			expected: StatusCompleted,
		},
		{
			name: "checklist section",
			content: `# Task: Test Task

## Checklist
- [x] Item 1
- [ ] Item 2
`,
			expected: StatusInProgress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			taskPath := filepath.Join(tmpDir, "test_task.md")
			if err := os.WriteFile(taskPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			// Test
			result := ParseTaskStatus(taskPath)
			if result != tt.expected {
				t.Errorf("ParseTaskStatus() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseTaskStatus_NonexistentFile(t *testing.T) {
	result := ParseTaskStatus("/nonexistent/path/to/file.md")
	if result != StatusPending {
		t.Errorf("ParseTaskStatus() for nonexistent file = %q, want %q", result, StatusPending)
	}
}

func TestParseHeader(t *testing.T) {
	tests := []struct {
		line          string
		expectedLevel int
		expectedText  string
	}{
		{"# Header 1", 1, "Header 1"},
		{"## Header 2", 2, "Header 2"},
		{"### Header 3", 3, "Header 3"},
		{"#### Header 4", 4, "Header 4"},
		{"##### Header 5", 5, "Header 5"},
		{"###### Header 6", 6, "Header 6"},
		{"####### Too many hashes", 0, ""},
		{"Not a header", 0, ""},
		{"  ## Indented header", 2, "Indented header"},
		{"## Definition of Done", 2, "Definition of Done"},
	}

	for _, tt := range tests {
		t.Run(tt.line, func(t *testing.T) {
			level, text := parseHeader(tt.line)
			if level != tt.expectedLevel || text != tt.expectedText {
				t.Errorf("parseHeader(%q) = (%d, %q), want (%d, %q)",
					tt.line, level, text, tt.expectedLevel, tt.expectedText)
			}
		})
	}
}

func TestStatusFromCounts(t *testing.T) {
	tests := []struct {
		name     string
		counts   CheckboxCounts
		expected string
	}{
		{
			name:     "all completed",
			counts:   CheckboxCounts{Checked: 5, Unchecked: 0, Total: 5},
			expected: StatusCompleted,
		},
		{
			name:     "some completed",
			counts:   CheckboxCounts{Checked: 3, Unchecked: 2, Total: 5},
			expected: StatusInProgress,
		},
		{
			name:     "none completed",
			counts:   CheckboxCounts{Checked: 0, Unchecked: 5, Total: 5},
			expected: StatusPending,
		},
		{
			name:     "empty counts",
			counts:   CheckboxCounts{Checked: 0, Unchecked: 0, Total: 0},
			expected: StatusPending,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := statusFromCounts(tt.counts)
			if result != tt.expected {
				t.Errorf("statusFromCounts(%+v) = %q, want %q", tt.counts, result, tt.expected)
			}
		})
	}
}

func TestAddCheckboxCounts(t *testing.T) {
	tests := []struct {
		name     string
		line     string
		expected CheckboxCounts
	}{
		{
			name:     "checked lowercase x",
			line:     "- [x] Completed item",
			expected: CheckboxCounts{Checked: 1, Unchecked: 0, Total: 1},
		},
		{
			name:     "checked uppercase X",
			line:     "- [X] Completed item",
			expected: CheckboxCounts{Checked: 1, Unchecked: 0, Total: 1},
		},
		{
			name:     "unchecked",
			line:     "- [ ] Pending item",
			expected: CheckboxCounts{Checked: 0, Unchecked: 1, Total: 1},
		},
		{
			name:     "emoji completed",
			line:     "- [✅] Completed item",
			expected: CheckboxCounts{Checked: 1, Unchecked: 0, Total: 1},
		},
		{
			name:     "emoji in progress",
			line:     "- [🚧] In progress item",
			expected: CheckboxCounts{Checked: 0, Unchecked: 1, Total: 1},
		},
		{
			name:     "no checkbox",
			line:     "Just regular text",
			expected: CheckboxCounts{Checked: 0, Unchecked: 0, Total: 0},
		},
		{
			name:     "asterisk marker checked",
			line:     "* [x] Completed item",
			expected: CheckboxCounts{Checked: 1, Unchecked: 0, Total: 1},
		},
		{
			name:     "indented checkbox",
			line:     "  - [x] Indented completed",
			expected: CheckboxCounts{Checked: 1, Unchecked: 0, Total: 1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counts := CheckboxCounts{}
			addCheckboxCounts(&counts, tt.line)
			if counts != tt.expected {
				t.Errorf("addCheckboxCounts() = %+v, want %+v", counts, tt.expected)
			}
		})
	}
}
