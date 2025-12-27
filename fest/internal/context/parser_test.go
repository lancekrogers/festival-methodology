package context

import (
	"testing"
	"time"
)

func TestParseGoalFile_Basic(t *testing.T) {
	content := []byte(`# FESTIVAL_GOAL

## Objective

Build a CLI tool for managing festivals.

## Success Criteria

- [ ] Create command implemented
- [x] Validate command implemented
- [ ] Status command implemented
`)

	goal, err := ParseGoalFile(content)
	if err != nil {
		t.Fatalf("ParseGoalFile() error = %v", err)
	}

	if goal.Title != "FESTIVAL_GOAL" {
		t.Errorf("Title = %q, want %q", goal.Title, "FESTIVAL_GOAL")
	}

	if goal.Objective == "" {
		t.Error("Objective should not be empty")
	}

	if len(goal.SuccessCriteria) != 3 {
		t.Errorf("SuccessCriteria count = %d, want 3", len(goal.SuccessCriteria))
	}
}

func TestParseGoalFile_WithFrontmatter(t *testing.T) {
	content := []byte(`---
status: in_progress
priority: high
---

# Phase Goal

## Objective

Complete the implementation phase.
`)

	goal, err := ParseGoalFile(content)
	if err != nil {
		t.Fatalf("ParseGoalFile() error = %v", err)
	}

	if goal.Status != "in_progress" {
		t.Errorf("Status = %q, want %q", goal.Status, "in_progress")
	}

	if goal.Priority != "high" {
		t.Errorf("Priority = %q, want %q", goal.Priority, "high")
	}
}

func TestParseRulesFile(t *testing.T) {
	content := []byte(`# FESTIVAL_RULES

## Error Handling

### Always Wrap Errors

Use gerror.Wrap() for all error handling.

### Include Context

Error messages should include relevant context.

## Testing

- **Unit Tests Required**: All new code must have unit tests.
- **Coverage Target**: Aim for 80% coverage minimum.
`)

	rules, err := ParseRulesFile(content)
	if err != nil {
		t.Fatalf("ParseRulesFile() error = %v", err)
	}

	if len(rules) == 0 {
		t.Fatal("Rules should not be empty")
	}

	// Check for error handling rules
	var foundErrorHandling bool
	for _, rule := range rules {
		if rule.Category == "Error Handling" {
			foundErrorHandling = true
			break
		}
	}

	if !foundErrorHandling {
		t.Error("Should have found Error Handling category")
	}
}

func TestParseContextFile(t *testing.T) {
	content := []byte(`# CONTEXT

## Decisions

### 2024-12-01: Use Cobra for CLI

We decided to use Cobra for the CLI framework.

- Rationale: Well-established, good documentation
- Impact: Consistent command structure

### 2024-12-15: Add JSON output

Added --json flag to all commands.

- Rationale: Machine-readable output for automation
- Impact: All commands support JSON
`)

	decisions, err := ParseContextFile(content)
	if err != nil {
		t.Fatalf("ParseContextFile() error = %v", err)
	}

	if len(decisions) == 0 {
		t.Fatal("Decisions should not be empty")
	}

	// Verify first decision
	found := false
	for _, d := range decisions {
		if d.Summary == "Use Cobra for CLI" {
			found = true
			if d.Date.Year() != 2024 || d.Date.Month() != time.December || d.Date.Day() != 1 {
				t.Errorf("Date = %v, want 2024-12-01", d.Date)
			}
			break
		}
	}

	if !found {
		t.Error("Should have found 'Use Cobra for CLI' decision")
	}
}

func TestParseTaskFile(t *testing.T) {
	content := []byte(`# Task: 03_fest_context

> **Task Number**: 03 | **Parallel Execution**: Yes | **Dependencies**: None | **Autonomy Level**: high

## Objective

Implement the fest context command.

## Deliverables

- internal/context/builder.go
- internal/context/parser.go
- internal/context/output.go
`)

	task, err := ParseTaskFile(content)
	if err != nil {
		t.Fatalf("ParseTaskFile() error = %v", err)
	}

	if task.Name != "03_fest_context" {
		t.Errorf("Name = %q, want %q", task.Name, "03_fest_context")
	}

	if task.AutonomyLevel != "high" {
		t.Errorf("AutonomyLevel = %q, want %q", task.AutonomyLevel, "high")
	}

	if !task.ParallelAllowed {
		t.Error("ParallelAllowed should be true")
	}

	if len(task.Deliverables) != 3 {
		t.Errorf("Deliverables count = %d, want 3", len(task.Deliverables))
	}
}

func TestExtractSection(t *testing.T) {
	tests := []struct {
		name        string
		text        string
		sectionName string
		prefix      string
		want        string
	}{
		{
			name:        "basic section",
			text:        "# Title\n\n## Objective\n\nDo something great.\n\n## Next",
			sectionName: "Objective",
			prefix:      "##",
			want:        "Do something great.",
		},
		{
			name:        "section not found",
			text:        "# Title\n\n## Other\n\nContent",
			sectionName: "Objective",
			prefix:      "##",
			want:        "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractSection(tc.text, tc.sectionName, tc.prefix)
			if got != tc.want {
				t.Errorf("extractSection() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestExtractChecklist(t *testing.T) {
	text := `## Tasks

- [ ] First item
- [x] Second item (done)
- [ ] Third item

Some other text.
`

	items := extractChecklist(text)
	if len(items) != 3 {
		t.Errorf("extractChecklist() = %d items, want 3", len(items))
	}

	if items[0] != "First item" {
		t.Errorf("items[0] = %q, want %q", items[0], "First item")
	}
}

func TestExtractFrontmatterField(t *testing.T) {
	text := `---
status: complete
priority: high
author: test
---

# Content
`

	status := extractFrontmatterField(text, "status")
	if status != "complete" {
		t.Errorf("status = %q, want %q", status, "complete")
	}

	priority := extractFrontmatterField(text, "priority")
	if priority != "high" {
		t.Errorf("priority = %q, want %q", priority, "high")
	}

	missing := extractFrontmatterField(text, "nonexistent")
	if missing != "" {
		t.Errorf("missing field = %q, want empty", missing)
	}
}

func TestExtractBulletList(t *testing.T) {
	text := `Some intro text.

- First bullet
- Second bullet
* Third with asterisk

More text.
`

	items := extractBulletList(text)
	if len(items) != 3 {
		t.Errorf("extractBulletList() = %d items, want 3", len(items))
	}
}
