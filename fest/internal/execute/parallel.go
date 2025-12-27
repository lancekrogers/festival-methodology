package execute

import (
	"context"
	"fmt"
	"sync"
)

// ParallelExecutor handles parallel task execution
type ParallelExecutor struct {
	maxParallel int
}

// NewParallelExecutor creates a new parallel executor
func NewParallelExecutor(maxParallel int) *ParallelExecutor {
	if maxParallel <= 0 {
		maxParallel = 1
	}
	return &ParallelExecutor{
		maxParallel: maxParallel,
	}
}

// TaskExecutor is a function that executes a single task
type TaskExecutor func(ctx context.Context, task *PlanTask) error

// Execute runs tasks in parallel up to maxParallel limit
func (p *ParallelExecutor) Execute(ctx context.Context, tasks []*PlanTask, executor TaskExecutor) []error {
	if len(tasks) == 0 {
		return nil
	}

	// Use a semaphore to limit parallelism
	sem := make(chan struct{}, p.maxParallel)
	var wg sync.WaitGroup

	errors := make([]error, len(tasks))
	var mu sync.Mutex

	for i, task := range tasks {
		// Check context before starting each task
		select {
		case <-ctx.Done():
			mu.Lock()
			errors[i] = ctx.Err()
			mu.Unlock()
			continue
		case sem <- struct{}{}:
			// Got semaphore, proceed
		}

		wg.Add(1)
		go func(idx int, t *PlanTask) {
			defer wg.Done()
			defer func() { <-sem }()

			// Execute the task
			err := executor(ctx, t)

			mu.Lock()
			errors[idx] = err
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()

	// Return nil if no errors
	hasErrors := false
	for _, err := range errors {
		if err != nil {
			hasErrors = true
			break
		}
	}

	if !hasErrors {
		return nil
	}

	return errors
}

// ExecuteStep runs all tasks in a step group with parallel support
func (p *ParallelExecutor) ExecuteStep(ctx context.Context, step *StepGroup, executor TaskExecutor) []error {
	if !step.Parallel {
		// Execute sequentially
		var errors []error
		for _, task := range step.Tasks {
			select {
			case <-ctx.Done():
				errors = append(errors, ctx.Err())
				return errors
			default:
			}

			if err := executor(ctx, task); err != nil {
				errors = append(errors, err)
			}
		}
		if len(errors) == 0 {
			return nil
		}
		return errors
	}

	// Execute in parallel
	return p.Execute(ctx, step.Tasks, executor)
}

// ParallelResult holds the result of parallel execution
type ParallelResult struct {
	Task    *PlanTask
	Success bool
	Error   error
}

// ExecuteWithResults runs tasks in parallel and returns detailed results
func (p *ParallelExecutor) ExecuteWithResults(ctx context.Context, tasks []*PlanTask, executor TaskExecutor) []*ParallelResult {
	if len(tasks) == 0 {
		return nil
	}

	results := make([]*ParallelResult, len(tasks))
	sem := make(chan struct{}, p.maxParallel)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, task := range tasks {
		select {
		case <-ctx.Done():
			mu.Lock()
			results[i] = &ParallelResult{
				Task:    task,
				Success: false,
				Error:   ctx.Err(),
			}
			mu.Unlock()
			continue
		case sem <- struct{}{}:
		}

		wg.Add(1)
		go func(idx int, t *PlanTask) {
			defer wg.Done()
			defer func() { <-sem }()

			err := executor(ctx, t)

			mu.Lock()
			results[idx] = &ParallelResult{
				Task:    t,
				Success: err == nil,
				Error:   err,
			}
			mu.Unlock()
		}(i, task)
	}

	wg.Wait()
	return results
}

// GroupParallelTasks groups tasks that can be executed in parallel
func GroupParallelTasks(tasks []*PlanTask) [][]*PlanTask {
	// Group by task number
	byNumber := make(map[int][]*PlanTask)
	for _, task := range tasks {
		byNumber[task.Number] = append(byNumber[task.Number], task)
	}

	// Convert to sorted groups
	var groups [][]*PlanTask
	var numbers []int
	for num := range byNumber {
		numbers = append(numbers, num)
	}

	// Sort numbers
	for i := 0; i < len(numbers); i++ {
		for j := i + 1; j < len(numbers); j++ {
			if numbers[i] > numbers[j] {
				numbers[i], numbers[j] = numbers[j], numbers[i]
			}
		}
	}

	for _, num := range numbers {
		groups = append(groups, byNumber[num])
	}

	return groups
}

// ValidateParallelSafety checks if tasks can safely run in parallel
func ValidateParallelSafety(tasks []*PlanTask) error {
	// Check for conflicting dependencies
	taskIDs := make(map[string]bool)
	for _, task := range tasks {
		taskIDs[task.ID] = true
	}

	for _, task := range tasks {
		for _, dep := range task.Dependencies {
			if taskIDs[dep] {
				return fmt.Errorf("task %q depends on task %q which is in the same parallel group",
					task.Name, dep)
			}
		}
	}

	return nil
}
