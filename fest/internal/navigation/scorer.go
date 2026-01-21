// Package navigation provides festival-project linking and navigation state management.
package navigation

import (
	"strings"
	"unicode"
)

// Scoring constants define the weight of different match types.
const (
	// ScoreExactMatch is awarded for exact string match.
	ScoreExactMatch = 100
	// ScorePrefixMatch is awarded when query is a prefix of target.
	ScorePrefixMatch = 75
	// ScoreWordBoundary is awarded for matches at word boundaries.
	ScoreWordBoundary = 50
	// ScoreCamelCase is awarded for matching at camelCase boundaries.
	ScoreCamelCase = 40
	// ScoreConsecutive is awarded for consecutive character matches.
	ScoreConsecutive = 25
	// ScoreSubstringMatch is the base score for any character match.
	ScoreSubstringMatch = 10
)

// Score calculates match score for query against target.
// Returns the score and positions of matched characters.
// A score of 0 indicates no match.
func Score(query, target string) (int, []int) {
	if query == "" {
		return 0, nil
	}

	queryLower := strings.ToLower(query)
	targetLower := strings.ToLower(target)

	// Exact match
	if queryLower == targetLower {
		return ScoreExactMatch, allPositions(len(target))
	}

	// Prefix match
	if strings.HasPrefix(targetLower, queryLower) {
		return ScorePrefixMatch, prefixPositions(len(query))
	}

	// Contains match (substring)
	if idx := strings.Index(targetLower, queryLower); idx >= 0 {
		score := ScoreSubstringMatch * len(query)
		// Bonus for word boundary
		if isWordBoundary(target, idx) {
			score += ScoreWordBoundary
		}
		return score, substringPositions(idx, len(query))
	}

	// Fuzzy match - find query characters in order
	return fuzzyMatch(queryLower, targetLower, target)
}

// fuzzyMatch attempts to find all query characters in target in order.
func fuzzyMatch(queryLower, targetLower, target string) (int, []int) {
	positions := make([]int, 0, len(queryLower))
	score := 0
	lastPos := -1

	qi := 0
	for ti := 0; ti < len(targetLower) && qi < len(queryLower); ti++ {
		if targetLower[ti] == queryLower[qi] {
			positions = append(positions, ti)

			// Bonus for consecutive matches (but not for the first character)
			if lastPos >= 0 && ti == lastPos+1 {
				score += ScoreConsecutive
			}

			// Bonus for word boundary (after separator)
			if isWordBoundary(target, ti) {
				score += ScoreWordBoundary
			}

			// Bonus for camelCase boundary
			if isCamelCaseBoundary(target, ti) {
				score += ScoreCamelCase
			}

			score += ScoreSubstringMatch
			lastPos = ti
			qi++
		}
	}

	// All query chars must match
	if qi != len(queryLower) {
		return 0, nil
	}

	return score, positions
}

// isWordBoundary checks if position is at a word boundary.
func isWordBoundary(s string, pos int) bool {
	if pos == 0 {
		return true
	}
	if pos >= len(s) {
		return false
	}

	prev := rune(s[pos-1])
	return prev == '-' || prev == '_' || prev == '/' || prev == '.' || prev == ' '
}

// isCamelCaseBoundary checks if position is at a camelCase boundary.
func isCamelCaseBoundary(s string, pos int) bool {
	if pos == 0 || pos >= len(s) {
		return false
	}

	// Check if this character is uppercase and previous is lowercase
	curr := rune(s[pos])
	prev := rune(s[pos-1])

	return unicode.IsUpper(curr) && unicode.IsLower(prev)
}

// allPositions returns positions for all characters.
func allPositions(length int) []int {
	positions := make([]int, length)
	for i := range positions {
		positions[i] = i
	}
	return positions
}

// prefixPositions returns positions for a prefix match.
func prefixPositions(length int) []int {
	positions := make([]int, length)
	for i := range positions {
		positions[i] = i
	}
	return positions
}

// substringPositions returns positions for a substring match.
func substringPositions(start, length int) []int {
	positions := make([]int, length)
	for i := range positions {
		positions[i] = start + i
	}
	return positions
}
