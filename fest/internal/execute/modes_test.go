package execute

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Mode != ModeHumanGuided {
		t.Errorf("Mode = %q, want %q", cfg.Mode, ModeHumanGuided)
	}

	if !cfg.AutoDemote {
		t.Error("AutoDemote should be true by default")
	}

	if !cfg.ReviewAtGates {
		t.Error("ReviewAtGates should be true by default")
	}
}

func TestShouldDemote(t *testing.T) {
	cfg := DefaultConfig()

	tests := []struct {
		name     string
		task     *TaskInfo
		expected bool
	}{
		{
			name:     "low autonomy task",
			task:     &TaskInfo{AutonomyLevel: AutonomyLow},
			expected: true,
		},
		{
			name:     "high autonomy task",
			task:     &TaskInfo{AutonomyLevel: AutonomyHigh},
			expected: false,
		},
		{
			name:     "medium autonomy task",
			task:     &TaskInfo{AutonomyLevel: AutonomyMedium},
			expected: false,
		},
		{
			name:     "gate task with review enabled",
			task:     &TaskInfo{AutonomyLevel: AutonomyHigh, IsGate: true},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := ShouldDemote(tc.task, cfg)
			if result != tc.expected {
				t.Errorf("ShouldDemote() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestShouldDemote_AutoDemoteDisabled(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AutoDemote = false

	task := &TaskInfo{AutonomyLevel: AutonomyLow}
	if ShouldDemote(task, cfg) {
		t.Error("ShouldDemote() should return false when AutoDemote is disabled")
	}
}

func TestDetectBestMode(t *testing.T) {
	tests := []struct {
		name     string
		tasks    []*TaskInfo
		expected ExecutionMode
	}{
		{
			name:     "no tasks",
			tasks:    nil,
			expected: ModeHumanGuided,
		},
		{
			name: "all high autonomy",
			tasks: []*TaskInfo{
				{AutonomyLevel: AutonomyHigh},
				{AutonomyLevel: AutonomyHigh},
				{AutonomyLevel: AutonomyHigh},
			},
			expected: ModeSemiAutonomous,
		},
		{
			name: "mostly low autonomy",
			tasks: []*TaskInfo{
				{AutonomyLevel: AutonomyLow},
				{AutonomyLevel: AutonomyLow},
				{AutonomyLevel: AutonomyHigh},
			},
			expected: ModeHumanGuided,
		},
		{
			name: "mixed but mostly high",
			tasks: []*TaskInfo{
				{AutonomyLevel: AutonomyHigh},
				{AutonomyLevel: AutonomyHigh},
				{AutonomyLevel: AutonomyHigh},
				{AutonomyLevel: AutonomyMedium},
				{AutonomyLevel: AutonomyLow},
			},
			expected: ModeSemiAutonomous,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := DetectBestMode(tc.tasks)
			if result != tc.expected {
				t.Errorf("DetectBestMode() = %q, want %q", result, tc.expected)
			}
		})
	}
}

func TestLoopState_ShouldContinue(t *testing.T) {
	tests := []struct {
		name     string
		state    *LoopState
		expected bool
	}{
		{
			name: "semi-autonomous with remaining tasks",
			state: &LoopState{
				Mode:           ModeSemiAutonomous,
				TasksRemaining: 5,
				Config:         DefaultConfig(),
			},
			expected: true,
		},
		{
			name: "human-guided mode",
			state: &LoopState{
				Mode:           ModeHumanGuided,
				TasksRemaining: 5,
				Config:         DefaultConfig(),
			},
			expected: false,
		},
		{
			name: "no remaining tasks",
			state: &LoopState{
				Mode:           ModeSemiAutonomous,
				TasksRemaining: 0,
				Config:         DefaultConfig(),
			},
			expected: false,
		},
		{
			name: "max tasks reached",
			state: &LoopState{
				Mode:           ModeSemiAutonomous,
				TasksRemaining: 5,
				TasksCompleted: 3,
				Config:         &ExecutionConfig{MaxTasks: 3},
			},
			expected: false,
		},
		{
			name: "handoff requiring user",
			state: &LoopState{
				Mode:           ModeSemiAutonomous,
				TasksRemaining: 5,
				Config:         DefaultConfig(),
				LastResult: &ExecutionResult{
					Handoff: &HandoffEvent{RequiresUser: true},
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.state.ShouldContinue()
			if result != tc.expected {
				t.Errorf("ShouldContinue() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestCreateHandoff(t *testing.T) {
	event := CreateHandoff(
		ModeSemiAutonomous,
		ModeHumanGuided,
		HandoffLowAutonomy,
		"01_task.md",
		"Task requires human input",
	)

	if event.FromMode != ModeSemiAutonomous {
		t.Errorf("FromMode = %q, want %q", event.FromMode, ModeSemiAutonomous)
	}

	if event.ToMode != ModeHumanGuided {
		t.Errorf("ToMode = %q, want %q", event.ToMode, ModeHumanGuided)
	}

	if event.Reason != HandoffLowAutonomy {
		t.Errorf("Reason = %q, want %q", event.Reason, HandoffLowAutonomy)
	}

	if !event.RequiresUser {
		t.Error("RequiresUser should be true for low autonomy handoff")
	}
}

func TestCreateHandoff_NoUserRequired(t *testing.T) {
	event := CreateHandoff(
		ModeSemiAutonomous,
		ModeHumanGuided,
		HandoffComplete,
		"",
		"All tasks completed",
	)

	if event.RequiresUser {
		t.Error("RequiresUser should be false for completion handoff")
	}
}

func TestExecutionConfig_Defaults(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.ClaimTimeout != 5*time.Minute {
		t.Errorf("ClaimTimeout = %v, want %v", cfg.ClaimTimeout, 5*time.Minute)
	}

	if cfg.MaxTasks != 0 {
		t.Errorf("MaxTasks = %d, want 0 (unlimited)", cfg.MaxTasks)
	}

	if cfg.DryRun {
		t.Error("DryRun should be false by default")
	}
}
