package navigation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestScore(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		target    string
		wantScore int
		wantMatch bool
	}{
		{
			name:      "exact match",
			query:     "test",
			target:    "test",
			wantScore: ScoreExactMatch,
			wantMatch: true,
		},
		{
			name:      "exact match case insensitive",
			query:     "TEST",
			target:    "test",
			wantScore: ScoreExactMatch,
			wantMatch: true,
		},
		{
			name:      "prefix match",
			query:     "impl",
			target:    "implement",
			wantScore: ScorePrefixMatch,
			wantMatch: true,
		},
		{
			name:      "substring match",
			query:     "ment",
			target:    "implement",
			wantScore: ScoreSubstringMatch * 4, // 4 chars matched
			wantMatch: true,
		},
		{
			name:      "substring at word boundary",
			query:     "api",
			target:    "01_api",
			wantScore: ScoreSubstringMatch*3 + ScoreWordBoundary, // 3 chars + word boundary bonus
			wantMatch: true,
		},
		{
			name:      "fuzzy match",
			query:     "it",
			target:    "implement",
			wantScore: ScoreWordBoundary + ScoreSubstringMatch + ScoreSubstringMatch, // 'i' at word boundary + 't' later
			wantMatch: true,
		},
		{
			name:      "fuzzy match with consecutive bonus",
			query:     "imp",
			target:    "implement",
			wantScore: ScorePrefixMatch, // prefix match takes precedence
			wantMatch: true,
		},
		{
			name:      "no match",
			query:     "xyz",
			target:    "test",
			wantScore: 0,
			wantMatch: false,
		},
		{
			name:      "empty query",
			query:     "",
			target:    "test",
			wantScore: 0,
			wantMatch: false,
		},
		{
			name:      "camelCase boundary match",
			query:     "FN",
			target:    "festivalName",
			wantScore: ScoreWordBoundary + ScoreSubstringMatch + ScoreCamelCase + ScoreSubstringMatch, // f at word boundary (start), N at camelCase
			wantMatch: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score, indices := Score(tt.query, tt.target)
			if tt.wantMatch {
				assert.Greater(t, score, 0, "expected a match")
				assert.NotEmpty(t, indices, "expected match positions")
				assert.Equal(t, tt.wantScore, score)
			} else {
				assert.Equal(t, 0, score, "expected no match")
				assert.Nil(t, indices, "expected no positions")
			}
		})
	}
}

func TestIsWordBoundary(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		pos    int
		want   bool
	}{
		{"start of string", "test", 0, true},
		{"after dash", "foo-bar", 4, true},
		{"after underscore", "foo_bar", 4, true},
		{"after slash", "foo/bar", 4, true},
		{"after dot", "foo.bar", 4, true},
		{"after space", "foo bar", 4, true},
		{"middle of word", "foobar", 3, false},
		{"end of string", "test", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isWordBoundary(tt.s, tt.pos)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestIsCamelCaseBoundary(t *testing.T) {
	tests := []struct {
		name string
		s    string
		pos  int
		want bool
	}{
		{"start of string", "Test", 0, false},
		{"camelCase boundary", "festivalName", 8, true},    // N is uppercase after lowercase
		{"PascalCase boundary", "FestivalName", 8, true},   // N is uppercase after lowercase
		{"all lowercase", "festival", 4, false},
		{"all uppercase", "FESTIVAL", 4, false},
		{"end of string", "test", 4, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCamelCaseBoundary(tt.s, tt.pos)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllPositions(t *testing.T) {
	positions := allPositions(5)
	assert.Equal(t, []int{0, 1, 2, 3, 4}, positions)

	positions = allPositions(0)
	assert.Equal(t, []int{}, positions)
}

func TestPrefixPositions(t *testing.T) {
	positions := prefixPositions(3)
	assert.Equal(t, []int{0, 1, 2}, positions)
}

func TestSubstringPositions(t *testing.T) {
	positions := substringPositions(3, 4)
	assert.Equal(t, []int{3, 4, 5, 6}, positions)
}

func TestScorePriority(t *testing.T) {
	// Verify exact match beats everything
	exactScore, _ := Score("test", "test")
	prefixScore, _ := Score("test", "testing")
	assert.Greater(t, exactScore, prefixScore, "exact match should score higher than prefix")

	// Verify prefix match beats plain substring match
	plainSubstringScore, _ := Score("est", "atest")
	prefixShortScore, _ := Score("est", "esting")
	assert.Greater(t, prefixShortScore, plainSubstringScore, "prefix match should score higher than plain substring")

	// Verify word boundary substring beats plain substring
	wordBoundaryScore, _ := Score("test", "_test")
	plainSubstringScore2, _ := Score("test", "atest")
	assert.Greater(t, wordBoundaryScore, plainSubstringScore2, "word boundary substring should score higher than plain substring")

	// Note: word boundary complete match can score higher than prefix
	// because ScoreSubstringMatch*4 + ScoreWordBoundary(50) = 90 > ScorePrefixMatch(75)
	// This is intentional - a complete word at a boundary is often more relevant
	assert.Greater(t, wordBoundaryScore, prefixScore, "word boundary complete match scores higher than prefix")

	// Verify exact > word boundary complete > prefix > plain substring
	assert.Greater(t, exactScore, wordBoundaryScore)
	assert.Greater(t, prefixScore, plainSubstringScore)
}
