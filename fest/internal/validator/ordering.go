package validator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// OrderingValidator validates sequential numbering at all levels.
type OrderingValidator struct{}

// NewOrderingValidator creates a new ordering validator.
func NewOrderingValidator() *OrderingValidator {
	return &OrderingValidator{}
}

// Validate checks for numbering gaps in the festival structure.
func (v *OrderingValidator) Validate(ctx context.Context, festivalPath string) ([]Issue, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return ValidateOrdering(ctx, festivalPath)
}

// ValidateOrdering is the main entry point for ordering validation.
func ValidateOrdering(ctx context.Context, festivalPath string) ([]Issue, error) {
	issues := []Issue{}
	parser := festival.NewParser()

	// Validate phase ordering
	phases, err := parser.ParsePhases(festivalPath)
	if err != nil {
		return issues, fmt.Errorf("parse phases: %w", err)
	}

	phaseIssues := validateElementOrdering(phases, festivalPath, "phase")
	issues = append(issues, phaseIssues...)

	// Validate sequences within each phase
	for _, phase := range phases {
		if err := ctx.Err(); err != nil {
			return issues, err
		}

		sequences, err := parser.ParseSequences(phase.Path)
		if err != nil {
			continue // Skip phases that can't be parsed
		}

		seqIssues := validateElementOrdering(sequences, phase.Path, "sequence")
		issues = append(issues, seqIssues...)

		// Validate tasks within each sequence
		for _, seq := range sequences {
			if err := ctx.Err(); err != nil {
				return issues, err
			}

			tasks, err := parser.ParseTasks(seq.Path)
			if err != nil {
				continue
			}

			taskIssues := validateElementOrdering(tasks, seq.Path, "task")
			issues = append(issues, taskIssues...)
		}
	}

	return issues, nil
}

// validateElementOrdering checks for gaps in a list of elements.
// Elements are pre-sorted by number from the parser.
func validateElementOrdering(elements []festival.FestivalElement, parentPath, elementType string) []Issue {
	issues := []Issue{}

	if len(elements) == 0 {
		return issues // No elements, no gaps possible
	}

	// Group elements by number (to detect parallel/duplicate items)
	numberGroups := groupByNumber(elements)

	// Get sorted unique numbers
	uniqueNumbers := getSortedUniqueNumbers(numberGroups)

	if len(uniqueNumbers) == 0 {
		return issues
	}

	// Check first element starts at 1
	if uniqueNumbers[0] != 1 {
		firstElem := numberGroups[uniqueNumbers[0]][0]
		numFormat := formatNumber(uniqueNumbers[0], elementType)
		issues = append(issues, Issue{
			Level:   LevelError,
			Code:    CodeNumberingGap,
			Path:    filepath.Base(firstElem.Path),
			Message: fmt.Sprintf("%s numbering must start at %s, found %s", elementType, formatNumber(1, elementType), numFormat),
			Fix:     fmt.Sprintf("Use 'fest renumber' to fix, or create %s %s", elementType, formatNumber(1, elementType)),
		})
	}

	if len(uniqueNumbers) <= 1 {
		return issues // Only one unique number, no gaps possible (parallel items)
	}

	// Check for gaps between consecutive unique numbers
	for i := 1; i < len(uniqueNumbers); i++ {
		prev := uniqueNumbers[i-1]
		curr := uniqueNumbers[i]

		// Expected next number is prev + 1
		if curr != prev+1 {
			missing := getMissingNumbers(prev, curr)
			currElem := numberGroups[curr][0]

			var msg string
			if len(missing) == 1 {
				msg = fmt.Sprintf("%s numbering gap: %s follows %s (missing %s)",
					elementType, formatNumber(curr, elementType), formatNumber(prev, elementType), formatNumber(missing[0], elementType))
			} else if len(missing) <= 5 {
				// Show all missing numbers if 5 or fewer
				missingStrs := make([]string, len(missing))
				for i, m := range missing {
					missingStrs[i] = formatNumber(m, elementType)
				}
				msg = fmt.Sprintf("%s numbering gap: %s follows %s (missing %v)",
					elementType, formatNumber(curr, elementType), formatNumber(prev, elementType), missingStrs)
			} else {
				// Show range for large gaps
				msg = fmt.Sprintf("%s numbering gap: %s follows %s (missing %s-%s, %d total)",
					elementType, formatNumber(curr, elementType), formatNumber(prev, elementType),
					formatNumber(missing[0], elementType), formatNumber(missing[len(missing)-1], elementType), len(missing))
			}

			issues = append(issues, Issue{
				Level:   LevelError,
				Code:    CodeNumberingGap,
				Path:    filepath.Base(currElem.Path),
				Message: msg,
				Fix:     fmt.Sprintf("Use 'fest renumber' to fix gaps, or create missing %s(s)", elementType),
			})
		}
	}

	// Check for non-consecutive duplicates (invalid parallel work)
	issues = append(issues, validateConsecutiveDuplicates(elements, elementType)...)

	return issues
}

// formatNumber formats a number based on element type.
func formatNumber(n int, elementType string) string {
	if elementType == "phase" {
		return fmt.Sprintf("%03d", n)
	}
	return fmt.Sprintf("%02d", n)
}

// groupByNumber groups elements by their number.
func groupByNumber(elements []festival.FestivalElement) map[int][]festival.FestivalElement {
	groups := make(map[int][]festival.FestivalElement)
	for _, elem := range elements {
		groups[elem.Number] = append(groups[elem.Number], elem)
	}
	return groups
}

// getSortedUniqueNumbers returns sorted unique numbers from a group map.
func getSortedUniqueNumbers(groups map[int][]festival.FestivalElement) []int {
	numbers := make([]int, 0, len(groups))
	for num := range groups {
		numbers = append(numbers, num)
	}
	sort.Ints(numbers)
	return numbers
}

// getMissingNumbers returns the numbers between start (exclusive) and end (exclusive).
func getMissingNumbers(start, end int) []int {
	missing := make([]int, 0, end-start-1)
	for n := start + 1; n < end; n++ {
		missing = append(missing, n)
	}
	return missing
}

// validateConsecutiveDuplicates checks that duplicate numbers appear consecutively in filesystem order.
// Invalid: 01_a, 02_b, 01_c (01 appears non-consecutively)
// Valid: 01_a, 01_b, 02_c (01s are consecutive)
// Note: elements are sorted by number, but filesystem order may differ.
// We need to check the original filesystem order for this validation.
func validateConsecutiveDuplicates(elements []festival.FestivalElement, elementType string) []Issue {
	issues := []Issue{}

	if len(elements) == 0 {
		return issues
	}

	// Get the parent directory from the first element
	parentDir := filepath.Dir(elements[0].Path)

	// Read directory entries in filesystem order (alphabetical by name)
	entries, err := os.ReadDir(parentDir)
	if err != nil {
		return issues // Can't read, skip check
	}

	// Build a map of element paths to numbers
	pathToNumber := make(map[string]int)
	for _, elem := range elements {
		pathToNumber[elem.Path] = elem.Number
	}

	// Check filesystem order for non-consecutive duplicates
	seen := make(map[int]bool)
	lastNumber := -1

	for _, entry := range entries {
		entryPath := filepath.Join(parentDir, entry.Name())
		num, exists := pathToNumber[entryPath]
		if !exists {
			continue // Not one of our elements (e.g., SEQUENCE_GOAL.md)
		}

		if num != lastNumber {
			// New number encountered
			if seen[num] {
				// This number was seen before but not in the previous position
				issues = append(issues, Issue{
					Level:   LevelError,
					Code:    CodeNumberingGap,
					Path:    entry.Name(),
					Message: fmt.Sprintf("Non-consecutive duplicate: %s %s appears after different number (must be consecutive for parallel work)",
						elementType, formatNumber(num, elementType)),
					Fix: "Parallel items must have consecutive identical numbers",
				})
			}
			seen[num] = true
		}
		lastNumber = num
	}

	return issues
}

// CheckOrderingCorrect is a boolean helper for the checklist.
// Returns true if no numbering gaps exist.
func CheckOrderingCorrect(festivalPath string) bool {
	issues, err := ValidateOrdering(context.Background(), festivalPath)
	if err != nil {
		return true // Can't check, assume OK
	}
	return len(issues) == 0
}
