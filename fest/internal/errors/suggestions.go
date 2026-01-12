// Package errors provides structured error types for fest CLI.
package errors

import (
	"fmt"
	"sort"
	"strings"
)

// LevenshteinDistance calculates the edit distance between two strings.
// This is the minimum number of single-character edits (insertions, deletions,
// or substitutions) required to change one string into the other.
func LevenshteinDistance(a, b string) int {
	if a == b {
		return 0
	}
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Convert to runes for proper Unicode handling
	aRunes := []rune(a)
	bRunes := []rune(b)

	// Create matrix
	lenA, lenB := len(aRunes), len(bRunes)
	matrix := make([][]int, lenA+1)
	for i := range matrix {
		matrix[i] = make([]int, lenB+1)
		matrix[i][0] = i
	}
	for j := 0; j <= lenB; j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= lenA; i++ {
		for j := 1; j <= lenB; j++ {
			cost := 1
			if aRunes[i-1] == bRunes[j-1] {
				cost = 0
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[lenA][lenB]
}

// suggestion represents a candidate with its edit distance.
type suggestion struct {
	value    string
	distance int
}

// SuggestSimilar returns candidates within maxDistance of the input string,
// sorted by distance (closest first). Returns empty slice if no matches found.
func SuggestSimilar(input string, candidates []string, maxDistance int) []string {
	if len(candidates) == 0 || input == "" {
		return nil
	}

	// Normalize input for comparison
	normalizedInput := strings.ToLower(input)

	var suggestions []suggestion
	for _, candidate := range candidates {
		normalizedCandidate := strings.ToLower(candidate)
		distance := LevenshteinDistance(normalizedInput, normalizedCandidate)
		if distance <= maxDistance {
			suggestions = append(suggestions, suggestion{
				value:    candidate,
				distance: distance,
			})
		}
	}

	if len(suggestions) == 0 {
		return nil
	}

	// Sort by distance, then alphabetically
	sort.Slice(suggestions, func(i, j int) bool {
		if suggestions[i].distance != suggestions[j].distance {
			return suggestions[i].distance < suggestions[j].distance
		}
		return suggestions[i].value < suggestions[j].value
	})

	// Extract sorted values
	result := make([]string, len(suggestions))
	for i, s := range suggestions {
		result[i] = s.value
	}
	return result
}

// DidYouMean creates an error with a "did you mean?" suggestion.
// Returns nil if no similar candidates are found within the default distance.
func DidYouMean(input string, candidates []string) *Error {
	// Default max distance is 2 (typical typo threshold)
	similar := SuggestSimilar(input, candidates, 2)
	if len(similar) == 0 {
		return nil
	}

	hint := formatSuggestions(similar)
	return Validation(fmt.Sprintf("unknown value %q", input)).
		WithHint(hint).
		WithField("input", input).
		WithField("suggestions", similar)
}

// formatSuggestions formats a list of suggestions as a "Did you mean?" hint.
func formatSuggestions(suggestions []string) string {
	switch len(suggestions) {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf("Did you mean %q?", suggestions[0])
	case 2:
		return fmt.Sprintf("Did you mean %q or %q?", suggestions[0], suggestions[1])
	default:
		// Show first two plus count of others
		return fmt.Sprintf("Did you mean %q, %q, or one of %d others?",
			suggestions[0], suggestions[1], len(suggestions)-2)
	}
}

// ValidateWithSuggestions validates input against a list of valid values.
// Returns nil if input is valid, otherwise returns an error with suggestions.
func ValidateWithSuggestions(input string, validValues []string, valueType string) error {
	// Check for exact match (case-insensitive)
	normalizedInput := strings.ToLower(input)
	for _, valid := range validValues {
		if strings.ToLower(valid) == normalizedInput {
			return nil // Valid
		}
	}

	// Not valid - check for similar values
	similar := SuggestSimilar(input, validValues, 2)
	if len(similar) > 0 {
		hint := formatSuggestions(similar)
		return Validation(fmt.Sprintf("invalid %s: %q", valueType, input)).
			WithHint(hint).
			WithField("input", input).
			WithField("valid_values", validValues).
			WithField("suggestions", similar)
	}

	// No suggestions found
	return Validation(fmt.Sprintf("invalid %s: %q", valueType, input)).
		WithHintf("Valid values: %s", strings.Join(validValues, ", ")).
		WithField("input", input).
		WithField("valid_values", validValues)
}
