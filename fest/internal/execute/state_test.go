package execute

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStateManager(t *testing.T) {
	sm := NewStateManager("/tmp/test-festival")
	if sm == nil {
		t.Fatal("NewStateManager returned nil")
	}

	expected := "/tmp/test-festival/.fest/execution_state.yaml"
	if sm.statePath != expected {
		t.Errorf("statePath = %q, want %q", sm.statePath, expected)
	}
}

func TestStateManager_LoadNew(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	state, err := sm.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if state == nil {
		t.Fatal("Load() returned nil state")
	}

	if state.FestivalPath != tmpDir {
		t.Errorf("FestivalPath = %q, want %q", state.FestivalPath, tmpDir)
	}

	if state.TaskStatuses == nil {
		t.Error("TaskStatuses should not be nil")
	}
}

func TestStateManager_SaveAndLoad(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)

	// Load to initialize
	_, err := sm.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Set some values
	sm.SetTaskStatus("task-1", StatusCompleted)
	sm.SetCurrentPosition(1, 2, 3)
	sm.SetMode(ModeSemiAutonomous)

	// Save
	if err := sm.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	statePath := filepath.Join(tmpDir, ".fest", "execution_state.yaml")
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		t.Error("State file was not created")
	}

	// Load again with a new manager
	sm2 := NewStateManager(tmpDir)
	state, err := sm2.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error on reload = %v", err)
	}

	if state.Mode != ModeSemiAutonomous {
		t.Errorf("Mode = %q, want %q", state.Mode, ModeSemiAutonomous)
	}

	if state.CurrentPhase != 1 {
		t.Errorf("CurrentPhase = %d, want 1", state.CurrentPhase)
	}

	status := sm2.GetTaskStatus("task-1")
	if status != StatusCompleted {
		t.Errorf("GetTaskStatus(task-1) = %q, want %q", status, StatusCompleted)
	}
}

func TestStateManager_GetTaskStatus_Default(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)
	sm.Load(ctx)

	status := sm.GetTaskStatus("nonexistent")
	if status != StatusPending {
		t.Errorf("GetTaskStatus(nonexistent) = %q, want %q", status, StatusPending)
	}
}

func TestStateManager_Clear(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	sm := NewStateManager(tmpDir)
	sm.Load(ctx)
	sm.SetTaskStatus("task-1", StatusCompleted)
	sm.Save(ctx)

	if !sm.Exists() {
		t.Error("Expected state to exist after save")
	}

	if err := sm.Clear(ctx); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	if sm.Exists() {
		t.Error("Expected state to not exist after clear")
	}
}

func TestExecutionState_Progress(t *testing.T) {
	tests := []struct {
		name     string
		state    *ExecutionState
		expected float64
	}{
		{
			name: "50% complete",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 5,
			},
			expected: 50.0,
		},
		{
			name: "no tasks",
			state: &ExecutionState{
				TotalTasks: 0,
			},
			expected: 0.0,
		},
		{
			name: "100% complete",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 10,
			},
			expected: 100.0,
		},
		{
			name: "with skipped",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 5,
				SkippedTasks:   3,
			},
			expected: 80.0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.state.Progress()
			if got != tc.expected {
				t.Errorf("Progress() = %f, want %f", got, tc.expected)
			}
		})
	}
}

func TestExecutionState_IsComplete(t *testing.T) {
	tests := []struct {
		name     string
		state    *ExecutionState
		expected bool
	}{
		{
			name: "complete",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 10,
			},
			expected: true,
		},
		{
			name: "incomplete",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 5,
			},
			expected: false,
		},
		{
			name: "with skipped and failed",
			state: &ExecutionState{
				TotalTasks:     10,
				CompletedTasks: 5,
				SkippedTasks:   3,
				FailedTasks:    2,
			},
			expected: true,
		},
		{
			name: "no tasks",
			state: &ExecutionState{
				TotalTasks: 0,
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.state.IsComplete()
			if got != tc.expected {
				t.Errorf("IsComplete() = %v, want %v", got, tc.expected)
			}
		})
	}
}

func TestExecutionState_PendingTasks(t *testing.T) {
	state := &ExecutionState{
		TotalTasks:     10,
		CompletedTasks: 3,
		SkippedTasks:   2,
		FailedTasks:    1,
	}

	pending := state.PendingTasks()
	if pending != 4 {
		t.Errorf("PendingTasks() = %d, want 4", pending)
	}
}
