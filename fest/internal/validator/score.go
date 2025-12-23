package validator

// CalculateScore computes a validation score based on issues.
func CalculateScore(result *Result) int {
	if len(result.Issues) == 0 {
		return 100
	}

	errorCount := 0
	warningCount := 0

	for _, issue := range result.Issues {
		switch issue.Level {
		case LevelError:
			errorCount++
		case LevelWarning:
			warningCount++
		}
	}

	// Base score of 100, minus 15 per error, minus 5 per warning
	score := 100 - (errorCount * 15) - (warningCount * 5)
	if score < 0 {
		score = 0
	}

	return score
}

// AddSuggestions adds helpful suggestions based on issues found.
func AddSuggestions(result *Result) {
	hasMissingTasks := false
	hasMissingGates := false
	hasUnfilledTemplates := false

	for _, issue := range result.Issues {
		switch issue.Code {
		case CodeMissingTaskFiles:
			hasMissingTasks = true
		case CodeMissingQualityGate:
			hasMissingGates = true
		case CodeUnfilledTemplate:
			hasUnfilledTemplates = true
		}
	}

	if hasMissingTasks {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest understand tasks' to learn about task file creation")
	}
	if hasMissingGates {
		result.Suggestions = append(result.Suggestions,
			"Run 'fest task defaults sync --approve' to add quality gates")
	}
	if hasUnfilledTemplates {
		result.Suggestions = append(result.Suggestions,
			"Edit files with [FILL:] markers and replace with actual content")
	}
}
