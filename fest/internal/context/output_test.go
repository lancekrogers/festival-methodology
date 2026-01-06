package context

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestFormatter_FormatText_IncludesKeyData(t *testing.T) {
	ctx := &ContextOutput{
		Location: &Location{
			FestivalName: "test-fest",
			Level:        "task",
			PhaseName:    "001_FOUNDATION",
			SequenceName: "01_sequence",
			TaskName:     "01_task",
		},
		Depth: DepthStandard,
		Festival: &FestivalContext{
			Goal: &GoalContext{
				Title:     "Test Festival",
				Objective: "Festival objective",
			},
			PhaseCount: 2,
		},
		Phase: &PhaseContext{
			Name:      "001_FOUNDATION",
			PhaseType: "planning",
			Goal: &GoalContext{
				Objective: "Phase objective",
			},
			SequenceCount: 3,
		},
		Sequence: &SequenceContext{
			Name: "01_sequence",
			Goal: &GoalContext{
				Objective: "Sequence objective",
			},
			TaskCount: 4,
		},
		Task: &TaskContext{
			Name:            "01_task",
			TaskNumber:      1,
			AutonomyLevel:   "high",
			ParallelAllowed: true,
			Objective:       "Task objective",
			Dependencies:    []string{"01_prev"},
			Deliverables:    []string{"Deliverable A"},
		},
		Rules: []Rule{
			{
				Category:    "Engineering",
				Title:       "Follow standards",
				Description: "Use established conventions",
			},
		},
		Decisions: []Decision{
			{
				Date:      time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				Summary:   "Decision summary",
				Rationale: "Decision rationale",
			},
		},
		DependencyOutputs: []DepOutput{
			{
				TaskName: "01_prev",
				Outputs:  []string{"Output 1"},
			},
		},
	}

	formatter := NewFormatter(false)
	output := formatter.FormatText(ctx)

	for _, snippet := range []string{
		"CONTEXT",
		"Location",
		"Festival",
		"Phase",
		"Sequence",
		"Task",
		"Rules",
		"Decisions",
		"Dependency Outputs",
		"test-fest",
		"001_FOUNDATION",
		"01_sequence",
		"01_task",
		"planning",
		"Festival objective",
		"Phase objective",
		"Sequence objective",
		"Task objective",
		"Engineering",
		"Follow standards",
		"Decision summary",
		"2025-01-02",
		"01_prev",
		"Output 1",
		"standard",
	} {
		if !contains(output, snippet) {
			t.Errorf("expected output to contain %q", snippet)
		}
	}
}

func contains(s, substr string) bool {
	if substr == "" {
		return false
	}
	return strings.Contains(stripANSI(s), substr)
}

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripANSI(s string) string {
	return ansiRegexp.ReplaceAllString(s, "")
}
