// Package id provides festival ID generation and parsing utilities.
// Festival IDs follow the format XX0001 where XX is a 2-letter prefix
// derived from the festival name and 0001 is a 4-digit counter.
package id

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

// Common words to skip when extracting initials
var commonWords = map[string]bool{
	"the": true, "a": true, "an": true,
	"of": true, "for": true, "and": true,
	"to": true, "in": true, "on": true,
}

// idPattern matches festival IDs in directory names (e.g., _GU0001)
var idPattern = regexp.MustCompile(`_([A-Z]{2})(\d{4,})$`)

// StatusDirectories are the directories that contain festivals
var StatusDirectories = []string{"planned", "active", "completed", "dungeon"}

// ExtractInitials extracts a 2-letter uppercase prefix from a festival name.
// For multi-word names: first letter of first two significant words.
// For single-word names: first two letters.
// If extraction fails, returns "XX" as a fallback.
func ExtractInitials(name string) string {
	// Normalize: replace hyphens with spaces, lowercase
	normalized := strings.ToLower(strings.ReplaceAll(name, "-", " "))

	// Split into words
	words := strings.Fields(normalized)

	// Filter out common words and non-alphabetic entries
	var significant []string
	for _, word := range words {
		// Skip common words
		if commonWords[word] {
			continue
		}
		// Keep only if starts with a letter
		if len(word) > 0 && unicode.IsLetter(rune(word[0])) {
			significant = append(significant, word)
		}
	}

	// Build initials
	var initials strings.Builder

	switch len(significant) {
	case 0:
		// No significant words found
		return "XX"
	case 1:
		// Single word: take first two letters
		word := significant[0]
		for _, r := range word {
			if unicode.IsLetter(r) {
				initials.WriteRune(unicode.ToUpper(r))
				if initials.Len() >= 2 {
					break
				}
			}
		}
	default:
		// Multiple words: first letter of first two
		for i := 0; i < 2 && i < len(significant); i++ {
			for _, r := range significant[i] {
				if unicode.IsLetter(r) {
					initials.WriteRune(unicode.ToUpper(r))
					break
				}
			}
		}
	}

	// Pad if needed
	result := initials.String()
	if len(result) < 2 {
		result = result + strings.Repeat("X", 2-len(result))
	}

	return result[:2]
}

// FormatID creates a festival ID string from prefix and counter.
// Format: XX0001 (2 letters + 4 digits, zero-padded)
func FormatID(prefix string, counter int) string {
	if counter > 9999 {
		// Overflow: don't zero-pad beyond 4 digits
		return fmt.Sprintf("%s%d", prefix, counter)
	}
	return fmt.Sprintf("%s%04d", prefix, counter)
}

// ParseID extracts the prefix and counter from a festival ID.
// Returns an error if the ID format is invalid.
func ParseID(id string) (prefix string, counter int, err error) {
	if len(id) < 6 {
		return "", 0, fmt.Errorf("invalid ID format: too short")
	}

	prefix = id[:2]
	if !isValidPrefix(prefix) {
		return "", 0, fmt.Errorf("invalid ID format: prefix must be two uppercase letters")
	}

	counterStr := id[2:]
	if len(counterStr) < 4 {
		return "", 0, fmt.Errorf("invalid ID format: counter too short")
	}

	counter, err = strconv.Atoi(counterStr)
	if err != nil {
		return "", 0, fmt.Errorf("invalid ID format: counter must be numeric")
	}

	return prefix, counter, nil
}

// isValidPrefix checks if a prefix is two uppercase letters
func isValidPrefix(prefix string) bool {
	if len(prefix) != 2 {
		return false
	}
	for _, r := range prefix {
		if r < 'A' || r > 'Z' {
			return false
		}
	}
	return true
}

// ExtractIDFromDirName extracts the festival ID from a directory name.
// Directory names follow the pattern: name_XX0001
func ExtractIDFromDirName(dirName string) (string, error) {
	if dirName == "" {
		return "", fmt.Errorf("empty directory name")
	}

	matches := idPattern.FindStringSubmatch(dirName)
	if matches == nil {
		return "", fmt.Errorf("no valid ID found in directory name: %s", dirName)
	}

	// matches[0] is full match, matches[1] is prefix, matches[2] is counter
	return matches[1] + matches[2], nil
}

// FindNextCounter scans all festival directories to find the next available
// counter for the given prefix. Returns the next counter value (max + 1).
func FindNextCounter(festivalsRoot string, prefix string) (int, error) {
	maxCounter := 0

	// Scan each status directory
	for _, status := range StatusDirectories {
		statusPath := filepath.Join(festivalsRoot, status)

		// Skip if directory doesn't exist
		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		// Walk the directory tree to find all festivals
		err := filepath.WalkDir(statusPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil // Skip directories we can't read
			}

			if !d.IsDir() {
				return nil
			}

			// Try to extract ID from directory name
			id, err := ExtractIDFromDirName(d.Name())
			if err != nil {
				return nil // Not a festival directory with ID
			}

			// Parse the ID
			idPrefix, counter, err := ParseID(id)
			if err != nil {
				return nil
			}

			// Only count matching prefixes
			if idPrefix == prefix && counter > maxCounter {
				maxCounter = counter
			}

			return nil
		})

		if err != nil {
			// Continue scanning other directories even if one fails
			continue
		}
	}

	return maxCounter + 1, nil
}

// GenerateID creates a unique festival ID for the given name.
// It extracts initials from the name and finds the next available counter
// by scanning all festival directories.
func GenerateID(name string, festivalsRoot string) (string, error) {
	initials := ExtractInitials(name)

	counter, err := FindNextCounter(festivalsRoot, initials)
	if err != nil {
		return "", fmt.Errorf("failed to find next counter: %w", err)
	}

	return FormatID(initials, counter), nil
}

// taskRefPattern matches task references like FEST-123456 or [FEST-123456]
var taskRefPattern = regexp.MustCompile(`FEST-(\d{6})`)

// Validate checks if a string is a valid task reference format.
// Valid format: FEST-xxxxxx where x is a digit
func Validate(ref string) bool {
	return taskRefPattern.MatchString(ref)
}

// ExtractFromMessage extracts all task references from a commit message.
// Looks for patterns like FEST-123456 or [FEST-123456]
func ExtractFromMessage(message string) []string {
	matches := taskRefPattern.FindAllString(message, -1)
	if matches == nil {
		return []string{}
	}
	return matches
}
