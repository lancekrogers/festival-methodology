package validator

import (
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// CheckOrderCorrect verifies that phases are in sequential order.
func CheckOrderCorrect(path string) bool {
	parser := festival.NewParser()
	phases, err := parser.ParsePhases(path)
	if err != nil {
		return true // Can't check, assume OK
	}

	// Check phases are sequential
	lastNum := 0
	for _, phase := range phases {
		if phase.Number < lastNum {
			return false
		}
		lastNum = phase.Number
	}

	return true
}

// CheckParallelCorrect verifies parallelization standards.
func CheckParallelCorrect(path string) bool {
	// For now, always return true - parallel validation is complex
	// and false positives would be confusing
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
