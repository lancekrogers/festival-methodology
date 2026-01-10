package markers

import (
	"reflect"
	"testing"
)

func TestSortMarkerFiles(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name: "festival hierarchy sorting",
			input: []string{
				"001_PLANNING/01_requirements/01_task.md",
				"FESTIVAL_OVERVIEW.md",
				"001_PLANNING/PHASE_GOAL.md",
				"FESTIVAL_GOAL.md",
				"001_PLANNING/01_requirements/SEQUENCE_GOAL.md",
				"TODO.md",
				"002_IMPLEMENT/PHASE_GOAL.md",
			},
			expected: []string{
				"FESTIVAL_GOAL.md",
				"FESTIVAL_OVERVIEW.md",
				"TODO.md",
				"001_PLANNING/PHASE_GOAL.md",
				"001_PLANNING/01_requirements/SEQUENCE_GOAL.md",
				"001_PLANNING/01_requirements/01_task.md",
				"002_IMPLEMENT/PHASE_GOAL.md",
			},
		},
		{
			name: "numeric prefix ordering",
			input: []string{
				"002_PHASE/03_seq/02_task.md",
				"001_PHASE/01_seq/01_task.md",
				"001_PHASE/02_seq/01_task.md",
				"002_PHASE/01_seq/01_task.md",
			},
			expected: []string{
				"001_PHASE/01_seq/01_task.md",
				"001_PHASE/02_seq/01_task.md",
				"002_PHASE/01_seq/01_task.md",
				"002_PHASE/03_seq/02_task.md",
			},
		},
		{
			name: "file priority within same directory",
			input: []string{
				"001_PHASE/03_random.md",
				"001_PHASE/PHASE_OVERVIEW.md",
				"001_PHASE/01_task.md",
				"001_PHASE/PHASE_GOAL.md",
			},
			expected: []string{
				"001_PHASE/PHASE_GOAL.md",
				"001_PHASE/PHASE_OVERVIEW.md",
				"001_PHASE/01_task.md",
				"001_PHASE/03_random.md",
			},
		},
		{
			name: "depth-based sorting",
			input: []string{
				"001_PHASE/01_seq/01_task.md",
				"FESTIVAL_GOAL.md",
				"001_PHASE/PHASE_GOAL.md",
			},
			expected: []string{
				"FESTIVAL_GOAL.md",
				"001_PHASE/PHASE_GOAL.md",
				"001_PHASE/01_seq/01_task.md",
			},
		},
		{
			name:     "empty slice",
			input:    []string{},
			expected: []string{},
		},
		{
			name: "single file",
			input: []string{
				"FESTIVAL_GOAL.md",
			},
			expected: []string{
				"FESTIVAL_GOAL.md",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortMarkerFiles(tt.input)
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("sortMarkerFiles() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestGetFilePriority(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		expected FilePriority
	}{
		{"GOAL file", "FESTIVAL_GOAL.md", PriorityGoal},
		{"goal lowercase", "phase_goal.md", PriorityGoal},
		{"OVERVIEW file", "PHASE_OVERVIEW.md", PriorityOverview},
		{"TODO file", "TODO.md", PriorityTodo},
		{"todo lowercase", "todo.md", PriorityTodo},
		{"regular task", "01_task.md", PriorityOther},
		{"random file", "notes.md", PriorityOther},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFilePriority(tt.filename)
			if result != tt.expected {
				t.Errorf("getFilePriority(%s) = %v, want %v", tt.filename, result, tt.expected)
			}
		})
	}
}

func TestGetPathDepth(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected int
	}{
		{"current directory", ".", 0},
		{"single file", "FESTIVAL_GOAL.md", 1},
		{"phase file", "001_PHASE/PHASE_GOAL.md", 2},
		{"sequence file", "001_PHASE/01_seq/SEQUENCE_GOAL.md", 3},
		{"task file", "001_PHASE/01_seq/01_task.md", 3},
		{"empty path", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathDepth(tt.path)
			if result != tt.expected {
				t.Errorf("getPathDepth(%s) = %d, want %d", tt.path, result, tt.expected)
			}
		})
	}
}

func TestExtractNumericPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"three digits", "001_PLANNING", 1},
		{"two digits", "02_task.md", 2},
		{"single digit", "1_file.md", 1},
		{"no prefix", "FESTIVAL_GOAL.md", 9999},
		{"no prefix with underscore", "_file.md", 9999},
		{"path with numeric dir", "001_PHASE/GOAL.md", 9999}, // filepath.Base extracts GOAL.md, which has no prefix
		{"multiple numbers", "123_test", 123},
		{"hyphen separator", "001-task.md", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNumericPrefix(tt.input)
			if result != tt.expected {
				t.Errorf("extractNumericPrefix(%s) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}
