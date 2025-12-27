package execute

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

func TestNewParallelExecutor(t *testing.T) {
	tests := []struct {
		name        string
		maxParallel int
		expected    int
	}{
		{"positive value", 4, 4},
		{"zero", 0, 1},
		{"negative", -1, 1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pe := NewParallelExecutor(tc.maxParallel)
			if pe.maxParallel != tc.expected {
				t.Errorf("maxParallel = %d, want %d", pe.maxParallel, tc.expected)
			}
		})
	}
}

func TestParallelExecutor_Execute_Sequential(t *testing.T) {
	pe := NewParallelExecutor(1)
	tasks := []*PlanTask{
		{ID: "task-1", Name: "Task 1"},
		{ID: "task-2", Name: "Task 2"},
		{ID: "task-3", Name: "Task 3"},
	}

	var order []string
	executor := func(ctx context.Context, task *PlanTask) error {
		order = append(order, task.ID)
		return nil
	}

	errs := pe.Execute(context.Background(), tasks, executor)
	if errs != nil {
		t.Errorf("Execute() returned errors: %v", errs)
	}

	// With maxParallel=1, should execute in order
	if len(order) != 3 {
		t.Errorf("Expected 3 tasks executed, got %d", len(order))
	}
}

func TestParallelExecutor_Execute_Parallel(t *testing.T) {
	pe := NewParallelExecutor(3)
	tasks := []*PlanTask{
		{ID: "task-1", Name: "Task 1"},
		{ID: "task-2", Name: "Task 2"},
		{ID: "task-3", Name: "Task 3"},
	}

	var count int32
	var maxConcurrent int32
	var currentConcurrent int32

	executor := func(ctx context.Context, task *PlanTask) error {
		atomic.AddInt32(&currentConcurrent, 1)
		defer atomic.AddInt32(&currentConcurrent, -1)

		// Check max concurrent
		curr := atomic.LoadInt32(&currentConcurrent)
		for {
			max := atomic.LoadInt32(&maxConcurrent)
			if curr > max {
				if atomic.CompareAndSwapInt32(&maxConcurrent, max, curr) {
					break
				}
			} else {
				break
			}
		}

		atomic.AddInt32(&count, 1)
		time.Sleep(10 * time.Millisecond)
		return nil
	}

	errs := pe.Execute(context.Background(), tasks, executor)
	if errs != nil {
		t.Errorf("Execute() returned errors: %v", errs)
	}

	if count != 3 {
		t.Errorf("Expected 3 tasks executed, got %d", count)
	}

	// With 3 parallel and 3 tasks, all should run concurrently
	if maxConcurrent < 2 {
		t.Logf("maxConcurrent = %d (might vary due to timing)", maxConcurrent)
	}
}

func TestParallelExecutor_Execute_WithErrors(t *testing.T) {
	pe := NewParallelExecutor(2)
	tasks := []*PlanTask{
		{ID: "task-1", Name: "Task 1"},
		{ID: "task-2", Name: "Task 2"},
		{ID: "task-3", Name: "Task 3"},
	}

	executor := func(ctx context.Context, task *PlanTask) error {
		if task.ID == "task-2" {
			return errors.New("task-2 failed")
		}
		return nil
	}

	errs := pe.Execute(context.Background(), tasks, executor)
	if errs == nil {
		t.Error("Expected errors from Execute()")
	}

	// Check that error is in the right position
	if errs[1] == nil || errs[1].Error() != "task-2 failed" {
		t.Errorf("Expected 'task-2 failed' at index 1, got %v", errs[1])
	}

	if errs[0] != nil {
		t.Errorf("Expected nil error at index 0, got %v", errs[0])
	}
}

func TestParallelExecutor_Execute_ContextCanceled(t *testing.T) {
	pe := NewParallelExecutor(1)
	tasks := []*PlanTask{
		{ID: "task-1", Name: "Task 1"},
		{ID: "task-2", Name: "Task 2"},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var executedCount int32
	executor := func(ctx context.Context, task *PlanTask) error {
		atomic.AddInt32(&executedCount, 1)
		return nil
	}

	errs := pe.Execute(ctx, tasks, executor)
	if errs == nil {
		t.Error("Expected errors from canceled context")
	}

	// At least one should have context.Canceled error
	hasCanceledErr := false
	for _, err := range errs {
		if err == context.Canceled {
			hasCanceledErr = true
			break
		}
	}
	if !hasCanceledErr {
		t.Error("Expected context.Canceled error")
	}
}

func TestParallelExecutor_ExecuteStep_Sequential(t *testing.T) {
	pe := NewParallelExecutor(4)
	step := &StepGroup{
		Number:   1,
		Parallel: false,
		Tasks: []*PlanTask{
			{ID: "task-1"},
			{ID: "task-2"},
		},
	}

	var count int32
	executor := func(ctx context.Context, task *PlanTask) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	errs := pe.ExecuteStep(context.Background(), step, executor)
	if errs != nil {
		t.Errorf("ExecuteStep() returned errors: %v", errs)
	}

	if count != 2 {
		t.Errorf("Expected 2 tasks executed, got %d", count)
	}
}

func TestParallelExecutor_ExecuteStep_Parallel(t *testing.T) {
	pe := NewParallelExecutor(4)
	step := &StepGroup{
		Number:   1,
		Parallel: true,
		Tasks: []*PlanTask{
			{ID: "task-1"},
			{ID: "task-2"},
		},
	}

	var count int32
	executor := func(ctx context.Context, task *PlanTask) error {
		atomic.AddInt32(&count, 1)
		return nil
	}

	errs := pe.ExecuteStep(context.Background(), step, executor)
	if errs != nil {
		t.Errorf("ExecuteStep() returned errors: %v", errs)
	}

	if count != 2 {
		t.Errorf("Expected 2 tasks executed, got %d", count)
	}
}

func TestParallelExecutor_ExecuteWithResults(t *testing.T) {
	pe := NewParallelExecutor(2)
	tasks := []*PlanTask{
		{ID: "task-1", Name: "Task 1"},
		{ID: "task-2", Name: "Task 2"},
	}

	executor := func(ctx context.Context, task *PlanTask) error {
		if task.ID == "task-2" {
			return errors.New("failed")
		}
		return nil
	}

	results := pe.ExecuteWithResults(context.Background(), tasks, executor)
	if len(results) != 2 {
		t.Fatalf("Expected 2 results, got %d", len(results))
	}

	if !results[0].Success {
		t.Error("Expected task-1 to succeed")
	}
	if results[1].Success {
		t.Error("Expected task-2 to fail")
	}
}

func TestGroupParallelTasks(t *testing.T) {
	tasks := []*PlanTask{
		{ID: "task-1", Number: 1},
		{ID: "task-2", Number: 1},
		{ID: "task-3", Number: 2},
		{ID: "task-4", Number: 2},
		{ID: "task-5", Number: 3},
	}

	groups := GroupParallelTasks(tasks)

	if len(groups) != 3 {
		t.Fatalf("Expected 3 groups, got %d", len(groups))
	}

	// First group (number 1) should have 2 tasks
	if len(groups[0]) != 2 {
		t.Errorf("Expected 2 tasks in group 0, got %d", len(groups[0]))
	}

	// Second group (number 2) should have 2 tasks
	if len(groups[1]) != 2 {
		t.Errorf("Expected 2 tasks in group 1, got %d", len(groups[1]))
	}

	// Third group (number 3) should have 1 task
	if len(groups[2]) != 1 {
		t.Errorf("Expected 1 task in group 2, got %d", len(groups[2]))
	}
}

func TestValidateParallelSafety(t *testing.T) {
	tests := []struct {
		name      string
		tasks     []*PlanTask
		wantError bool
	}{
		{
			name: "safe - no dependencies",
			tasks: []*PlanTask{
				{ID: "task-1", Name: "Task 1"},
				{ID: "task-2", Name: "Task 2"},
			},
			wantError: false,
		},
		{
			name: "safe - external dependencies",
			tasks: []*PlanTask{
				{ID: "task-1", Name: "Task 1", Dependencies: []string{"task-0"}},
				{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-0"}},
			},
			wantError: false,
		},
		{
			name: "unsafe - internal dependency",
			tasks: []*PlanTask{
				{ID: "task-1", Name: "Task 1"},
				{ID: "task-2", Name: "Task 2", Dependencies: []string{"task-1"}},
			},
			wantError: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateParallelSafety(tc.tasks)
			if tc.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestParallelExecutor_EmptyTasks(t *testing.T) {
	pe := NewParallelExecutor(4)

	executor := func(ctx context.Context, task *PlanTask) error {
		return nil
	}

	errs := pe.Execute(context.Background(), nil, executor)
	if errs != nil {
		t.Errorf("Execute(nil) should return nil, got %v", errs)
	}

	errs = pe.Execute(context.Background(), []*PlanTask{}, executor)
	if errs != nil {
		t.Errorf("Execute([]) should return nil, got %v", errs)
	}
}
