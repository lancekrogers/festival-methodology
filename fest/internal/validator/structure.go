package validator

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// isResearchPhase checks if a phase is a research phase based on naming or type.
// Research phases use freeform subdirectory structure and don't require numbered sequences/tasks.
func isResearchPhase(phaseName string) bool {
	normalized := strings.ToUpper(phaseName)
	return strings.Contains(normalized, "RESEARCH")
}

// ValidateStructure checks naming conventions and hierarchy only.
// Required file presence is handled by the CompletenessValidator.
func ValidateStructure(festivalPath string) ([]Issue, error) {
	issues := []Issue{}

	parser := festival.NewParser()
	phases, err := parser.ParsePhases(festivalPath)
	if err != nil {
		return issues, fmt.Errorf("parse phases: %w", err)
	}

	phaseRe := regexp.MustCompile(`^\d{3}_(.+)$`)
	seqRe := regexp.MustCompile(`^\d{2}_(.+)$`)
	taskRe := regexp.MustCompile(`^\d{2}_(.+)\.md$`)

	for _, phase := range phases {
		// Phase directory must be NNN_UPPERCASE
		if m := phaseRe.FindStringSubmatch(phase.FullName); m != nil {
			namePart := m[1]
			if namePart != strings.ToUpper(namePart) {
				issues = append(issues, Issue{
					Level:   LevelWarning,
					Code:    CodeNamingConvention,
					Path:    phase.Path,
					Message: fmt.Sprintf("Phase name should be UPPERCASE: %s", phase.FullName),
					Fix:     fmt.Sprintf("rename to %03d_%s", phase.Number, strings.ToUpper(namePart)),
				})
			}
		}

		// Skip sequence/task validation for research phases
		// Research phases use freeform subdirectory structure
		if isResearchPhase(phase.Name) {
			continue
		}

		sequences, err := parser.ParseSequences(phase.Path)
		if err != nil {
			return issues, fmt.Errorf("parse sequences: %w", err)
		}
		for _, seq := range sequences {
			// Sequence directory must be NN_lowercase
			if m := seqRe.FindStringSubmatch(seq.FullName); m != nil {
				namePart := m[1]
				if namePart != strings.ToLower(namePart) {
					issues = append(issues, Issue{
						Level:   LevelWarning,
						Code:    CodeNamingConvention,
						Path:    seq.Path,
						Message: fmt.Sprintf("Sequence name should be lowercase: %s", seq.FullName),
						Fix:     fmt.Sprintf("rename to %02d_%s", seq.Number, strings.ToLower(namePart)),
					})
				}
			}

			// Tasks: NN_lowercase.md (skip SEQUENCE_GOAL.md)
			entries, _ := filepath.Glob(filepath.Join(seq.Path, "*.md"))
			for _, f := range entries {
				base := filepath.Base(f)
				if base == "SEQUENCE_GOAL.md" {
					continue
				}
				if m := taskRe.FindStringSubmatch(base); m != nil {
					namePart := m[1]
					if namePart != strings.ToLower(namePart) {
						issues = append(issues, Issue{
							Level:   LevelWarning,
							Code:    CodeNamingConvention,
							Path:    f,
							Message: fmt.Sprintf("Task filename should be lowercase: %s", base),
						})
					}
				} else {
					// Non-conforming task filename
					issues = append(issues, Issue{
						Level:   LevelWarning,
						Code:    CodeNamingConvention,
						Path:    f,
						Message: fmt.Sprintf("Task filename should match NN_name.md: %s", base),
					})
				}
			}
		}
	}

	return issues, nil
}
