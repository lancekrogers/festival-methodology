package markers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []Marker
	}{
		{
			name:     "empty content",
			content:  "",
			expected: nil,
		},
		{
			name:    "single marker",
			content: "[REPLACE: Project Name]",
			expected: []Marker{
				{FullMatch: "[REPLACE: Project Name]", Hint: "Project Name", LineNumber: 1, StartOffset: 0, EndOffset: 23},
			},
		},
		{
			name:    "multiple markers same line",
			content: "[REPLACE: A] and [REPLACE: B]",
			expected: []Marker{
				{FullMatch: "[REPLACE: A]", Hint: "A", LineNumber: 1, StartOffset: 0, EndOffset: 12},
				{FullMatch: "[REPLACE: B]", Hint: "B", LineNumber: 1, StartOffset: 17, EndOffset: 29},
			},
		},
		{
			name:    "markers on different lines",
			content: "[REPLACE: First]\n[REPLACE: Second]",
			expected: []Marker{
				{FullMatch: "[REPLACE: First]", Hint: "First", LineNumber: 1, StartOffset: 0, EndOffset: 16},
				{FullMatch: "[REPLACE: Second]", Hint: "Second", LineNumber: 2, StartOffset: 17, EndOffset: 34},
			},
		},
		{
			name:    "marker with options",
			content: "[REPLACE: Yes/No]",
			expected: []Marker{
				{FullMatch: "[REPLACE: Yes/No]", Hint: "Yes/No", LineNumber: 1, StartOffset: 0, EndOffset: 17},
			},
		},
		{
			name:    "marker with pipe options",
			content: "[REPLACE: high|medium|low]",
			expected: []Marker{
				{FullMatch: "[REPLACE: high|medium|low]", Hint: "high|medium|low", LineNumber: 1, StartOffset: 0, EndOffset: 26},
			},
		},
		{
			name:    "marker with long description",
			content: "[REPLACE: One clear sentence describing what will be accomplished]",
			expected: []Marker{
				{FullMatch: "[REPLACE: One clear sentence describing what will be accomplished]", Hint: "One clear sentence describing what will be accomplished", LineNumber: 1, StartOffset: 0, EndOffset: 66},
			},
		},
		{
			name:    "marker with extra whitespace",
			content: "[REPLACE:   spaced  hint   ]",
			expected: []Marker{
				{FullMatch: "[REPLACE:   spaced  hint   ]", Hint: "spaced  hint", LineNumber: 1, StartOffset: 0, EndOffset: 28},
			},
		},
		{
			name:     "no markers",
			content:  "This is regular text without any markers.",
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.content)

			if len(result) != len(tt.expected) {
				t.Fatalf("expected %d markers, got %d", len(tt.expected), len(result))
			}

			for i, exp := range tt.expected {
				got := result[i]
				if got.FullMatch != exp.FullMatch {
					t.Errorf("marker %d: expected FullMatch %q, got %q", i, exp.FullMatch, got.FullMatch)
				}
				if got.Hint != exp.Hint {
					t.Errorf("marker %d: expected Hint %q, got %q", i, exp.Hint, got.Hint)
				}
				if got.LineNumber != exp.LineNumber {
					t.Errorf("marker %d: expected LineNumber %d, got %d", i, exp.LineNumber, got.LineNumber)
				}
			}
		})
	}
}

func TestParseFile(t *testing.T) {
	ctx := context.Background()

	// Create temp file
	dir := t.TempDir()
	filePath := filepath.Join(dir, "test.md")

	content := `# Task: [REPLACE: Task Name]

## Objective
[REPLACE: Brief description]

**Status**: [REPLACE: pending/in_progress/completed]
`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	markers, err := ParseFile(ctx, filePath)
	if err != nil {
		t.Fatalf("ParseFile failed: %v", err)
	}

	if len(markers) != 3 {
		t.Fatalf("expected 3 markers, got %d", len(markers))
	}

	expectedHints := []string{"Task Name", "Brief description", "pending/in_progress/completed"}
	for i, hint := range expectedHints {
		if markers[i].Hint != hint {
			t.Errorf("marker %d: expected hint %q, got %q", i, hint, markers[i].Hint)
		}
	}
}

func TestReplace(t *testing.T) {
	content := "Hello [REPLACE: name], welcome to [REPLACE: place]!"

	markers := Parse(content)
	if len(markers) != 2 {
		t.Fatalf("expected 2 markers, got %d", len(markers))
	}

	values := []MarkerValue{
		{Marker: markers[0], Value: "Alice"},
		{Marker: markers[1], Value: "Wonderland"},
	}

	result := Replace(content, values)
	expected := "Hello Alice, welcome to Wonderland!"

	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestApplyInput(t *testing.T) {
	markers := []Marker{
		{FullMatch: "[REPLACE: A]", Hint: "A"},
		{FullMatch: "[REPLACE: B]", Hint: "B"},
		{FullMatch: "[REPLACE: C]", Hint: "C"},
	}

	input := map[string]string{
		"A": "Value A",
		"C": "Value C",
		// B is missing
	}

	values := ApplyInput(markers, input)

	if len(values) != 3 {
		t.Fatalf("expected 3 values, got %d", len(values))
	}

	if values[0].Value != "Value A" {
		t.Errorf("expected 'Value A', got %q", values[0].Value)
	}
	if values[1].Value != "[REPLACE: B]" {
		t.Errorf("expected original marker for B, got %q", values[1].Value)
	}
	if values[2].Value != "Value C" {
		t.Errorf("expected 'Value C', got %q", values[2].Value)
	}
}

func TestComputeResult(t *testing.T) {
	markers := []Marker{
		{FullMatch: "[REPLACE: A]", Hint: "A"},
		{FullMatch: "[REPLACE: B]", Hint: "B"},
		{FullMatch: "[REPLACE: C]", Hint: "C"},
	}

	values := []MarkerValue{
		{Marker: markers[0], Value: "Filled A"},
		{Marker: markers[1], Value: "[REPLACE: B]"}, // Unfilled
		{Marker: markers[2], Value: "Filled C"},
	}

	result := ComputeResult("/path/to/file.md", values)

	if result.TotalMarkers != 3 {
		t.Errorf("expected TotalMarkers 3, got %d", result.TotalMarkers)
	}
	if result.FilledMarkers != 2 {
		t.Errorf("expected FilledMarkers 2, got %d", result.FilledMarkers)
	}
	if len(result.UnfilledHints) != 1 || result.UnfilledHints[0] != "B" {
		t.Errorf("expected UnfilledHints [B], got %v", result.UnfilledHints)
	}
}

func TestParseJSON(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		want    map[string]string
	}{
		{
			name:    "empty string",
			json:    "",
			wantErr: false,
			want:    nil,
		},
		{
			name:    "valid json",
			json:    `{"A": "Value A", "B": "Value B"}`,
			wantErr: false,
			want:    map[string]string{"A": "Value A", "B": "Value B"},
		},
		{
			name:    "invalid json",
			json:    `{invalid}`,
			wantErr: true,
			want:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSON(tt.json)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if tt.want == nil && got != nil {
					t.Errorf("expected nil, got %v", got)
				}
				for k, v := range tt.want {
					if got[k] != v {
						t.Errorf("key %q: expected %q, got %q", k, v, got[k])
					}
				}
			}
		})
	}
}

func TestExtractHints(t *testing.T) {
	markers := []Marker{
		{Hint: "First"},
		{Hint: "Second"},
		{Hint: "Third"},
	}

	hints := ExtractHints(markers)

	if len(hints) != 3 {
		t.Fatalf("expected 3 hints, got %d", len(hints))
	}
	if hints[0] != "First" || hints[1] != "Second" || hints[2] != "Third" {
		t.Errorf("unexpected hints: %v", hints)
	}
}

func TestMarkersToJSON(t *testing.T) {
	markers := []Marker{
		{Hint: "Name", LineNumber: 1},
		{Hint: "Description", LineNumber: 5},
	}

	result := MarkersToJSON(markers)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	if result[0]["hint"] != "Name" {
		t.Errorf("expected hint 'Name', got %v", result[0]["hint"])
	}
	if result[0]["line"] != 1 {
		t.Errorf("expected line 1, got %v", result[0]["line"])
	}
}
