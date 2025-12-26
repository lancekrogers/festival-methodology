package validator

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
)

// ValidateQualityGates checks that implementation sequences have quality gate tasks
func ValidateQualityGates(ctx context.Context, festivalPath string) ([]Issue, error) {
	issues := []Issue{}

	parser := festival.NewParser()
	phases, err := parser.ParsePhases(ctx, festivalPath)
	if err != nil {
		return issues, fmt.Errorf("parse phases: %w", err)
	}

	policy := gates.DefaultPolicy()
	gateIDs := map[string]bool{}
	for _, gate := range policy.GetEnabledTasks() {
		gateIDs[gate.ID] = true
	}

	// Phase name patterns that don't require quality gates
	nonImplPhasePatterns := []string{"PLAN", "DESIGN", "REVIEW", "UAT", "FINALIZE", "DOCS", "RESEARCH"}

	for _, phase := range phases {
		if isNonImplementationPhase(phase.Name, nonImplPhasePatterns) {
			continue
		}
		sequences, err := parser.ParseSequences(ctx, phase.Path)
		if err != nil {
			return issues, fmt.Errorf("parse sequences: %w", err)
		}
		for _, seq := range sequences {
			if isExcludedSequence(seq.Name, policy.ExcludePatterns) {
				continue
			}

			tasks, err := parser.ParseTasks(ctx, seq.Path)
			if err != nil {
				return issues, fmt.Errorf("parse tasks: %w", err)
			}
			hasGates := false
			for _, task := range tasks {
				for gateID := range gateIDs {
					if strings.Contains(strings.ToLower(task.Name), strings.ReplaceAll(gateID, "_", "")) ||
						strings.Contains(task.Name, gateID) {
						hasGates = true
						break
					}
				}
				if hasGates {
					break
				}
			}
			if !hasGates && len(tasks) > 0 {
				rel, _ := filepath.Rel(festivalPath, seq.Path)
				issues = append(issues, Issue{
					Level:       LevelError,
					Code:        CodeMissingQualityGate,
					Path:        rel,
					Message:     "Implementation sequence missing quality gates",
					Fix:         fmt.Sprintf("fest gates apply --sequence %s --approve", rel),
					AutoFixable: true,
				})
			}
		}
	}

	return issues, nil
}

// isNonImplementationPhase checks if a phase is for planning/review rather than implementation
func isNonImplementationPhase(phaseName string, patterns []string) bool {
	upper := strings.ToUpper(phaseName)
	for _, p := range patterns {
		if strings.Contains(upper, p) {
			return true
		}
	}
	return false
}
