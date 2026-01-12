package markers

import (
	"os"
	"path/filepath"
	"testing"
)

func TestExtractPrimaryGoal(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name: "standard format with heading",
			content: `# Festival Goal: Test Festival

## Primary Goal

Build a comprehensive testing framework for the guild system.

## Context
More content here...`,
			expected: "Build a comprehensive testing framework for the guild system.",
		},
		{
			name: "inline format",
			content: `# Festival Goal

**Primary Goal:** Implement agent marker workflow improvements

## Background
Other content...`,
			expected: "Implement agent marker workflow improvements",
		},
		{
			name: "with bold markers",
			content: `## Primary Goal

**Create a robust testing infrastructure** that validates all core functionality.`,
			expected: "Create a robust testing infrastructure that validates all core functionality.",
		},
		{
			name: "with template markers",
			content: `## Primary Goal

[REPLACE: One sentence describing what this festival accomplishes]

## Context`,
			expected: "",
		},
		{
			name: "multiline goal",
			content: `## Primary Goal

This festival aims to improve the marker workflow
by implementing a next command that shows one file at a time.

## Success Criteria`,
			expected: "This festival aims to improve the marker workflow by implementing a next command that shows one file at a time.",
		},
		{
			name: "no primary goal section",
			content: `# Festival Goal

## Objective

Some objective text here.`,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test_goal.md")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			if err != nil {
				t.Fatalf("Failed to create temp file: %v", err)
			}

			result, err := extractPrimaryGoal(tmpFile)
			if err != nil {
				t.Fatalf("extractPrimaryGoal() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("extractPrimaryGoal() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestExtractPrimaryGoal_MissingFile(t *testing.T) {
	_, err := extractPrimaryGoal("/nonexistent/file.md")
	if err == nil {
		t.Error("Expected error for missing file, got nil")
	}
}

func TestBuildMarkerFileContext(t *testing.T) {
	// Create temporary festival structure
	tmpDir := t.TempDir()

	// Create festival goal
	festivalGoalContent := `# Festival Goal: test-festival

## Primary Goal

Improve fest CLI for better agent integration.`

	err := os.WriteFile(filepath.Join(tmpDir, "FESTIVAL_GOAL.md"), []byte(festivalGoalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create festival goal: %v", err)
	}

	// Create phase directory and goal
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	err = os.MkdirAll(phaseDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create phase dir: %v", err)
	}

	phaseGoalContent := `# Phase Goal: 001_PLANNING

## Primary Goal

Plan and design the marker workflow improvements.`

	err = os.WriteFile(filepath.Join(phaseDir, "PHASE_GOAL.md"), []byte(phaseGoalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create phase goal: %v", err)
	}

	// Create sequence directory and goal
	seqDir := filepath.Join(phaseDir, "01_design")
	err = os.MkdirAll(seqDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sequence dir: %v", err)
	}

	seqGoalContent := `# Sequence Goal: 01_design

## Primary Goal

Design the output format and behavior of fest markers next.`

	err = os.WriteFile(filepath.Join(seqDir, "SEQUENCE_GOAL.md"), []byte(seqGoalContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create sequence goal: %v", err)
	}

	tests := []struct {
		name              string
		filePath          string
		expectFestival    bool
		expectPhase       bool
		expectSequence    bool
		expectedFestGoal  string
		expectedPhaseGoal string
		expectedSeqGoal   string
	}{
		{
			name:             "festival level file",
			filePath:         filepath.Join(tmpDir, "FESTIVAL_OVERVIEW.md"),
			expectFestival:   true,
			expectPhase:      false,
			expectSequence:   false,
			expectedFestGoal: "Improve fest CLI for better agent integration.",
		},
		{
			name:              "phase level file",
			filePath:          filepath.Join(phaseDir, "PHASE_OVERVIEW.md"),
			expectFestival:    true,
			expectPhase:       true,
			expectSequence:    false,
			expectedFestGoal:  "Improve fest CLI for better agent integration.",
			expectedPhaseGoal: "Plan and design the marker workflow improvements.",
		},
		{
			name:              "sequence level file",
			filePath:          filepath.Join(seqDir, "SEQUENCE_OVERVIEW.md"),
			expectFestival:    true,
			expectPhase:       true,
			expectSequence:    true,
			expectedFestGoal:  "Improve fest CLI for better agent integration.",
			expectedPhaseGoal: "Plan and design the marker workflow improvements.",
			expectedSeqGoal:   "Design the output format and behavior of fest markers next.",
		},
		{
			name:              "task file",
			filePath:          filepath.Join(seqDir, "01_task.md"),
			expectFestival:    true,
			expectPhase:       true,
			expectSequence:    true,
			expectedFestGoal:  "Improve fest CLI for better agent integration.",
			expectedPhaseGoal: "Plan and design the marker workflow improvements.",
			expectedSeqGoal:   "Design the output format and behavior of fest markers next.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := buildMarkerFileContext(tmpDir, tt.filePath)
			if err != nil {
				t.Fatalf("buildMarkerFileContext() error = %v", err)
			}

			if ctx.Festival == nil && tt.expectFestival {
				t.Error("Expected festival context, got nil")
			} else if ctx.Festival != nil {
				expectedName := filepath.Base(tmpDir)
				if ctx.Festival.Name != expectedName {
					t.Errorf("Festival name = %q, want %q", ctx.Festival.Name, expectedName)
				}
				if tt.expectedFestGoal != "" && ctx.Festival.Goal != tt.expectedFestGoal {
					t.Errorf("Festival goal = %q, want %q", ctx.Festival.Goal, tt.expectedFestGoal)
				}
			}

			if tt.expectPhase && ctx.Phase == nil {
				t.Error("Expected phase context, got nil")
			} else if !tt.expectPhase && ctx.Phase != nil {
				t.Error("Did not expect phase context, but got one")
			} else if ctx.Phase != nil {
				if ctx.Phase.Name != "001_PLANNING" {
					t.Errorf("Phase name = %q, want %q", ctx.Phase.Name, "001_PLANNING")
				}
				if tt.expectedPhaseGoal != "" && ctx.Phase.Goal != tt.expectedPhaseGoal {
					t.Errorf("Phase goal = %q, want %q", ctx.Phase.Goal, tt.expectedPhaseGoal)
				}
			}

			if tt.expectSequence && ctx.Sequence == nil {
				t.Error("Expected sequence context, got nil")
			} else if !tt.expectSequence && ctx.Sequence != nil {
				t.Error("Did not expect sequence context, but got one")
			} else if ctx.Sequence != nil {
				if ctx.Sequence.Name != "01_design" {
					t.Errorf("Sequence name = %q, want %q", ctx.Sequence.Name, "01_design")
				}
				if tt.expectedSeqGoal != "" && ctx.Sequence.Goal != tt.expectedSeqGoal {
					t.Errorf("Sequence goal = %q, want %q", ctx.Sequence.Goal, tt.expectedSeqGoal)
				}
			}
		})
	}
}

func TestIsNumberedDir(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"numbered with underscore", "001_PLANNING", true},
		{"numbered with hyphen", "01-design", true},
		{"single digit", "1_file", true},
		{"not numbered", "FESTIVAL_GOAL", false},
		{"starts with dot", ".hidden", false},
		{"starts with underscore", "_private", false},
		{"empty string", "", false},
		{"single char", "1", false}, // Too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isNumberedDir(tt.input)
			if result != tt.expected {
				t.Errorf("isNumberedDir(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
