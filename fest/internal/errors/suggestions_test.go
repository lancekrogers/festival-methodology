package errors

import (
	"testing"
)

func TestLevenshteinDistance(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		{"identical strings", "hello", "hello", 0},
		{"empty strings", "", "", 0},
		{"one empty", "hello", "", 5},
		{"other empty", "", "world", 5},
		{"single substitution", "hello", "hallo", 1},
		{"single insertion", "hello", "helllo", 1},
		{"single deletion", "hello", "helo", 1},
		{"multiple edits", "sitting", "kitten", 3},
		{"case sensitivity", "Hello", "hello", 1},
		{"completely different", "abc", "xyz", 3},
		{"transposition", "ab", "ba", 2}, // Note: not Damerau-Levenshtein
		{"unicode chars", "café", "cafe", 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := LevenshteinDistance(tc.a, tc.b)
			if got != tc.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d",
					tc.a, tc.b, got, tc.expected)
			}
			// Test symmetry
			reversed := LevenshteinDistance(tc.b, tc.a)
			if reversed != tc.expected {
				t.Errorf("LevenshteinDistance(%q, %q) = %d, want %d (symmetry check)",
					tc.b, tc.a, reversed, tc.expected)
			}
		})
	}
}

func TestSuggestSimilar(t *testing.T) {
	candidates := []string{"active", "planned", "completed", "dungeon"}

	tests := []struct {
		name        string
		input       string
		maxDistance int
		expected    []string
	}{
		{"exact match", "active", 2, []string{"active"}},
		{"typo - actve", "actve", 2, []string{"active"}},
		{"typo - planed", "planed", 2, []string{"planned"}},
		{"typo - complted", "complted", 2, []string{"completed"}},
		{"case insensitive", "ACTIVE", 2, []string{"active"}},
		{"too far", "unknown", 2, nil},
		{"empty input", "", 2, nil},
		{"empty candidates", "test", 2, nil},
		{"multiple matches", "plan", 3, []string{"planned"}}, // distance 3
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var cands []string
			if tc.name != "empty candidates" {
				cands = candidates
			}
			got := SuggestSimilar(tc.input, cands, tc.maxDistance)
			if len(got) != len(tc.expected) {
				t.Errorf("SuggestSimilar(%q) = %v, want %v",
					tc.input, got, tc.expected)
				return
			}
			for i, v := range got {
				if v != tc.expected[i] {
					t.Errorf("SuggestSimilar(%q)[%d] = %q, want %q",
						tc.input, i, v, tc.expected[i])
				}
			}
		})
	}
}

func TestSuggestSimilar_SortOrder(t *testing.T) {
	candidates := []string{"plan", "plant", "plane", "planet"}

	// All within distance 2 of "plan", should be sorted by distance
	got := SuggestSimilar("plan", candidates, 3)

	// "plan" should be first (distance 0)
	if len(got) == 0 || got[0] != "plan" {
		t.Errorf("expected 'plan' first, got %v", got)
	}
}

func TestDidYouMean(t *testing.T) {
	candidates := []string{"active", "planned", "completed"}

	tests := []struct {
		name        string
		input       string
		wantNil     bool
		wantMessage string
	}{
		{"close match", "actve", false, `unknown value "actve"`},
		{"no match", "xyz123", true, ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := DidYouMean(tc.input, candidates)
			if tc.wantNil {
				if err != nil {
					t.Errorf("DidYouMean(%q) = %v, want nil", tc.input, err)
				}
				return
			}
			if err == nil {
				t.Errorf("DidYouMean(%q) = nil, want error", tc.input)
				return
			}
			if err.Message != tc.wantMessage {
				t.Errorf("DidYouMean(%q).Message = %q, want %q",
					tc.input, err.Message, tc.wantMessage)
			}
			if err.Hint == "" {
				t.Errorf("DidYouMean(%q).Hint is empty, want suggestion", tc.input)
			}
		})
	}
}

func TestFormatSuggestions(t *testing.T) {
	tests := []struct {
		name        string
		suggestions []string
		expected    string
	}{
		{"empty", nil, ""},
		{"single", []string{"foo"}, `Did you mean "foo"?`},
		{"two", []string{"foo", "bar"}, `Did you mean "foo" or "bar"?`},
		{"three", []string{"foo", "bar", "baz"}, `Did you mean "foo", "bar", or one of 1 others?`},
		{"many", []string{"a", "b", "c", "d", "e"}, `Did you mean "a", "b", or one of 3 others?`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := formatSuggestions(tc.suggestions)
			if got != tc.expected {
				t.Errorf("formatSuggestions(%v) = %q, want %q",
					tc.suggestions, got, tc.expected)
			}
		})
	}
}

func TestValidateWithSuggestions(t *testing.T) {
	validStatuses := []string{"pending", "in_progress", "completed", "blocked"}

	tests := []struct {
		name      string
		input     string
		wantError bool
		wantHint  bool
	}{
		{"valid exact", "pending", false, false},
		{"valid case insensitive", "PENDING", false, false},
		{"invalid with suggestion", "pendng", true, true},
		{"invalid no suggestion", "xyz123", true, true}, // Will have "Valid values" hint
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateWithSuggestions(tc.input, validStatuses, "status")
			if tc.wantError {
				if err == nil {
					t.Errorf("ValidateWithSuggestions(%q) = nil, want error", tc.input)
					return
				}
				festErr, ok := err.(*Error)
				if !ok {
					t.Errorf("expected *Error, got %T", err)
					return
				}
				if tc.wantHint && festErr.Hint == "" {
					t.Errorf("ValidateWithSuggestions(%q) hint is empty, want hint", tc.input)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateWithSuggestions(%q) = %v, want nil", tc.input, err)
				}
			}
		})
	}
}
