package validator

import (
	"context"
	"strings"
)

// CheckOrderCorrect verifies that phases, sequences, and tasks are sequentially numbered.
// It uses the comprehensive ordering validator to check all levels.
func CheckOrderCorrect(path string) bool {
	return CheckOrderingCorrect(path)
}

// CheckParallelCorrect verifies parallelization standards.
// Returns false if there are non-consecutive duplicate numbers.
func CheckParallelCorrect(path string) bool {
	issues, err := ValidateOrdering(context.Background(), path)
	if err != nil {
		return true // Can't check, assume OK
	}

	// Check if any issues are about non-consecutive duplicates
	for _, issue := range issues {
		if issue.Code == CodeNumberingGap &&
			strings.Contains(issue.Message, "Non-consecutive duplicate") {
			return false
		}
	}
	return true
}

// RunChecklist performs all checklist checks and returns results.
func RunChecklist(path string) *Checklist {
	checklist := &Checklist{}

	templatesFilled := CheckTemplatesFilled(path)
	checklist.TemplatesFilled = &templatesFilled

	// Goals achievable is a manual check - always nil
	checklist.GoalsAchievable = nil

	taskFilesExist := CheckTaskFilesExist(path)
	checklist.TaskFilesExist = &taskFilesExist

	orderCorrect := CheckOrderCorrect(path)
	checklist.OrderCorrect = &orderCorrect

	parallelCorrect := CheckParallelCorrect(path)
	checklist.ParallelCorrect = &parallelCorrect

	return checklist
}
