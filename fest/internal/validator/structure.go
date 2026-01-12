package validator

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// isFreeformPhase checks if a phase uses freeform structure (not numbered sequences/tasks).
// Research and planning phases use freeform subdirectory structure because:
// - Research explores topics in any order
// - Planning naturally works backward from goals to dependencies
func isFreeformPhase(phaseName string) bool {
	normalized := strings.ToUpper(phaseName)
	return strings.Contains(normalized, "RESEARCH") ||
		strings.Contains(normalized, "PLANNING") ||
		strings.Contains(normalized, "DESIGN")
}

// ValidateStructure checks naming conventions and hierarchy only.
// Required file presence is handled by the CompletenessValidator.
func ValidateStructure(ctx context.Context, festivalPath string) ([]Issue, error) {
	issues := []Issue{}

	parser := festival.NewParser()
	phases, err := parser.ParsePhases(ctx, festivalPath)
	if err != nil {
		return issues, errors.Wrap(err, "parsing phases").WithField("path", festivalPath)
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

		// Skip sequence/task validation for freeform phases (research, planning, design)
		// Freeform phases use flexible subdirectory structure
		if isFreeformPhase(phase.Name) {
			continue
		}

		sequences, err := parser.ParseSequences(ctx, phase.Path)
		if err != nil {
			return issues, errors.Wrap(err, "parsing sequences").WithField("phase", phase.Name)
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
