package execute

import (
	"testing"
)

func TestNewPlanBuilder(t *testing.T) {
	pb := NewPlanBuilder("/tmp/test-festival")
	if pb == nil {
		t.Fatal("NewPlanBuilder returned nil")
	}
	if pb.festivalPath != "/tmp/test-festival" {
		t.Errorf("festivalPath = %q, want /tmp/test-festival", pb.festivalPath)
	}
}

func TestExtractPhaseNumber(t *testing.T) {
	tests := []struct {
		name string
		want int
	}{
		{"01_Planning", 1},
		{"001_Research", 1},
		{"12_Implementation", 12},
		{"00_Init", 0},
		{"Planning", 0},
		{"", 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := extractPhaseNumber(tc.name)
			if got != tc.want {
				t.Errorf("extractPhaseNumber(%q) = %d, want %d", tc.name, got, tc.want)
			}
		})
	}
}

func TestExecutionSummary(t *testing.T) {
	summary := &ExecutionSummary{
		TotalPhases:    3,
		TotalSequences: 6,
		TotalTasks:     18,
		TotalSteps:     12,
		ParallelGroups: 4,
		QualityGates:   2,
	}

	if summary.TotalPhases != 3 {
		t.Errorf("TotalPhases = %d, want 3", summary.TotalPhases)
	}
	if summary.TotalTasks != 18 {
		t.Errorf("TotalTasks = %d, want 18", summary.TotalTasks)
	}
}

func TestStepGroup(t *testing.T) {
	step := &StepGroup{
		Number: 1,
		Type:   "parallel",
		Tasks: []*PlanTask{
			{ID: "task-1", Name: "Task 1", Number: 1},
			{ID: "task-2", Name: "Task 2", Number: 1},
		},
		Parallel: true,
	}

	if step.Number != 1 {
		t.Errorf("Number = %d, want 1", step.Number)
	}
	if !step.Parallel {
		t.Error("Expected parallel to be true")
	}
	if len(step.Tasks) != 2 {
		t.Errorf("Tasks count = %d, want 2", len(step.Tasks))
	}
}

func TestPhaseExecution(t *testing.T) {
	phase := &PhaseExecution{
		Name:       "01_Planning",
		Path:       "/test/01_Planning",
		Number:     1,
		Sequences:  nil,
		TotalTasks: 5,
		Status:     "pending",
	}

	if phase.Name != "01_Planning" {
		t.Errorf("Name = %q, want 01_Planning", phase.Name)
	}
	if phase.TotalTasks != 5 {
		t.Errorf("TotalTasks = %d, want 5", phase.TotalTasks)
	}
}

func TestSequenceExecution(t *testing.T) {
	seq := &SequenceExecution{
		Name:       "01_core",
		Path:       "/test/01_Planning/01_core",
		Number:     1,
		Steps:      nil,
		TotalTasks: 3,
		Status:     "in_progress",
	}

	if seq.Name != "01_core" {
		t.Errorf("Name = %q, want 01_core", seq.Name)
	}
	if seq.Status != "in_progress" {
		t.Errorf("Status = %q, want in_progress", seq.Status)
	}
}

func TestPlanTask(t *testing.T) {
	task := &PlanTask{
		ID:            "/test/01.md",
		Name:          "01_task",
		Path:          "/test/01_task.md",
		Number:        1,
		AutonomyLevel: "high",
		Dependencies:  []string{"00_init"},
		Status:        "pending",
	}

	if task.Name != "01_task" {
		t.Errorf("Name = %q, want 01_task", task.Name)
	}
	if task.AutonomyLevel != "high" {
		t.Errorf("AutonomyLevel = %q, want high", task.AutonomyLevel)
	}
}

func TestQualityGateInfo(t *testing.T) {
	gate := &QualityGateInfo{
		PhaseName:   "01_Planning",
		Type:        "phase_transition",
		Description: "Must review planning before implementation",
		Criteria:    []string{"All tasks complete", "All tests pass"},
		Passed:      false,
	}

	if gate.PhaseName != "01_Planning" {
		t.Errorf("PhaseName = %q, want 01_Planning", gate.PhaseName)
	}
	if len(gate.Criteria) != 2 {
		t.Errorf("Criteria count = %d, want 2", len(gate.Criteria))
	}
}
