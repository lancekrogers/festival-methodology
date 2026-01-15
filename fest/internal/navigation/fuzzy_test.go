package navigation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFuzzyFinder_Find(t *testing.T) {
	tests := []struct {
		name     string
		targets  []FuzzyTarget
		pattern  string
		wantLen  int
		wantTop  string // expected top match name
	}{
		{
			name: "exact match",
			targets: []FuzzyTarget{
				{Name: "fest-improvements-FI0002", Path: "/path/to/fest"},
				{Name: "other-festival", Path: "/path/to/other"},
			},
			pattern: "fest-improvements",
			wantLen: 1,
			wantTop: "fest-improvements-FI0002",
		},
		{
			name: "fuzzy match",
			targets: []FuzzyTarget{
				{Name: "001_CRITICAL_BUGS_AND_TRACKING", Path: "/path/001"},
				{Name: "002_IMPLEMENT", Path: "/path/002"},
				{Name: "003_FOUNDATION", Path: "/path/003"},
			},
			pattern: "impl",
			wantLen: 1,
			wantTop: "002_IMPLEMENT",
		},
		{
			name: "multi-word pattern",
			targets: []FuzzyTarget{
				{Name: "002_IMPLEMENT/01_api", Path: "/path/001"},
				{Name: "002_IMPLEMENT/02_service", Path: "/path/002"},
				{Name: "003_FOUNDATION/01_api", Path: "/path/003"},
			},
			pattern: "impl api",
			wantLen: 1,
			wantTop: "002_IMPLEMENT/01_api",
		},
		{
			name: "no match",
			targets: []FuzzyTarget{
				{Name: "one", Path: "/path/one"},
				{Name: "two", Path: "/path/two"},
			},
			pattern: "xyz",
			wantLen: 0,
		},
		{
			name:    "empty targets",
			targets: []FuzzyTarget{},
			pattern: "test",
			wantLen: 0,
		},
		{
			name: "empty pattern",
			targets: []FuzzyTarget{
				{Name: "test", Path: "/path"},
			},
			pattern: "",
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			finder := NewFuzzyFinder(tt.targets)
			matches := finder.Find(tt.pattern)

			assert.Len(t, matches, tt.wantLen)
			if tt.wantLen > 0 && tt.wantTop != "" {
				assert.Equal(t, tt.wantTop, matches[0].Name)
			}
		})
	}
}

func TestIsUnambiguous(t *testing.T) {
	tests := []struct {
		name    string
		matches []FuzzyMatch
		want    bool
	}{
		{
			name:    "empty matches",
			matches: []FuzzyMatch{},
			want:    true,
		},
		{
			name: "single match",
			matches: []FuzzyMatch{
				{Name: "test", Score: 100},
			},
			want: true,
		},
		{
			name: "clearly better top match",
			matches: []FuzzyMatch{
				{Name: "test1", Score: 100},
				{Name: "test2", Score: 50},
			},
			want: true,
		},
		{
			name: "ambiguous matches",
			matches: []FuzzyMatch{
				{Name: "test1", Score: 100},
				{Name: "test2", Score: 95},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnambiguous(tt.matches)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsPhaseDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"001_PHASE", true},
		{"002_test", true},
		{"999_last", true},
		{"01_sequence", false}, // Only 2 digits
		{"1_short", false},     // Only 1 digit
		{"abc_name", false},    // No digits
		{"00", false},          // Too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPhaseDir(tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsSequenceDir(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"01_sequence", true},
		{"02_test", true},
		{"99_last", true},
		{"001_phase", false}, // 3 digits
		{"1_short", false},   // Only 1 digit
		{"ab_name", false},   // No digits
		{"0", false},         // Too short
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isSequenceDir(tt.name)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFormatMatchList(t *testing.T) {
	matches := []FuzzyMatch{
		{Name: "one"},
		{Name: "two"},
		{Name: "three"},
		{Name: "four"},
		{Name: "five"},
	}

	// Test with limit
	got := FormatMatchList(matches, 3)
	assert.Equal(t, []string{"one", "two", "three"}, got)

	// Test with 0 limit (returns all - no limit)
	got = FormatMatchList(matches, 0)
	assert.Len(t, got, 5)

	// Test with negative limit (returns all - no limit)
	got = FormatMatchList(matches, -1)
	assert.Len(t, got, 5)

	// Test with limit > len
	got = FormatMatchList(matches, 10)
	assert.Len(t, got, 5)
}
