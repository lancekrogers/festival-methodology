package navigation

import (
	"testing"
)

func TestIsValidShortcutName(t *testing.T) {
	tests := []struct {
		name     string
		expected bool
	}{
		// Valid names
		{"n", true},
		{"api", true},
		{"my_shortcut", true},
		{"A1", true},
		{"test123", true},
		{"a", true},
		{"UPPER", true},
		{"Mix_Case123", true},
		{"12345", true},
		{"a_b_c_d_e", true},

		// Valid at max length (20 chars)
		{"12345678901234567890", true},

		// Invalid names
		{"", false},
		{"foo bar", false},               // space
		{"with-dash", false},             // dash
		{"foo.bar", false},               // dot
		{"foo/bar", false},               // slash
		{"foo:bar", false},               // colon
		{"foo@bar", false},               // at
		{"123456789012345678901", false}, // too long (21 chars)
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := isValidShortcutName(tc.name)
			if result != tc.expected {
				t.Errorf("isValidShortcutName(%q) = %v, want %v", tc.name, result, tc.expected)
			}
		})
	}
}
