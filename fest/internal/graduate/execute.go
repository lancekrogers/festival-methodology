package graduate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Executor creates files based on a graduation plan.
type Executor struct {
	festivalPath string
}

// NewExecutor creates a new plan executor.
func NewExecutor(festivalPath string) *Executor {
	return &Executor{festivalPath: festivalPath}
}

// Execute creates the implementation structure from a plan.
func (e *Executor) Execute(ctx context.Context, plan *GraduationPlan) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Create phase directory
	if err := os.MkdirAll(plan.Target.Path, 0755); err != nil {
		return fmt.Errorf("creating phase directory: %w", err)
	}

	// Create PHASE_GOAL.md
	if err := e.writePhaseGoal(plan); err != nil {
		return fmt.Errorf("writing PHASE_GOAL.md: %w", err)
	}

	// Create sequences
	for _, seq := range plan.Sequences {
		if err := ctx.Err(); err != nil {
			return err
		}

		seqPath := filepath.Join(plan.Target.Path, seq.FullName)
		if err := os.MkdirAll(seqPath, 0755); err != nil {
			return fmt.Errorf("creating sequence directory: %w", err)
		}

		// Create SEQUENCE_GOAL.md
		if err := e.writeSequenceGoal(seqPath, &seq); err != nil {
			return fmt.Errorf("writing SEQUENCE_GOAL.md: %w", err)
		}

		// Create task stubs
		for _, task := range seq.Tasks {
			taskPath := filepath.Join(seqPath, task.FullName)
			if err := e.writeTaskStub(taskPath, &task); err != nil {
				return fmt.Errorf("writing task %s: %w", task.FullName, err)
			}
		}
	}

	return nil
}

func (e *Executor) writePhaseGoal(plan *GraduationPlan) error {
	content := fmt.Sprintf(`# %s

**Status:** Not Started | **Graduated From:** %s

## Phase Objective

**Primary Goal:** %s

%s

## Sequences

%s
`,
		plan.PhaseGoal.Title,
		plan.Source.PhaseName,
		plan.PhaseGoal.Goal,
		strings.Join(plan.PhaseGoal.Sections, "\n\n"),
		formatSequenceList(plan.Sequences),
	)

	return os.WriteFile(filepath.Join(plan.Target.Path, "PHASE_GOAL.md"), []byte(content), 0644)
}

func (e *Executor) writeSequenceGoal(seqPath string, seq *ProposedSequence) error {
	// Note: Sequence metadata (status, order) is in frontmatter, not in markdown
	content := fmt.Sprintf(`# %s

## Sequence Objective

**Primary Goal:** %s

**Source Topic:** %s

## Tasks

%s
`,
		seq.Goal.Title,
		seq.Goal.Goal,
		seq.SourceTopic,
		formatTaskList(seq.Tasks),
	)

	return os.WriteFile(filepath.Join(seqPath, "SEQUENCE_GOAL.md"), []byte(content), 0644)
}

func (e *Executor) writeTaskStub(taskPath string, task *ProposedTask) error {
	sourceRef := ""
	if len(task.SourceDocs) > 0 {
		sourceRef = fmt.Sprintf("\n**Planning Reference:** `%s`\n", task.SourceDocs[0])
	}

	content := fmt.Sprintf(`# Task: %s

> **Task Number**: %02d | **Parallel Execution**: No | **Dependencies**: None | **Autonomy Level**: medium

## Objective

%s
%s
## Requirements

- [ ] Implement functionality as specified in planning
- [ ] Add tests
- [ ] Update documentation

## Implementation Steps

### 1. Review Planning Documents

Review the source planning documents for requirements and decisions.

### 2. Implement

[Fill in implementation steps]

### 3. Test

[Fill in test steps]

## Completion Checklist

- [ ] Implementation complete
- [ ] Tests pass
- [ ] Documentation updated
- [ ] Self-review completed
`,
		task.Name,
		task.Number,
		task.Objective,
		sourceRef,
	)

	return os.WriteFile(taskPath, []byte(content), 0644)
}

func formatSequenceList(sequences []ProposedSequence) string {
	var lines []string
	for _, seq := range sequences {
		lines = append(lines, fmt.Sprintf("- **%s**: %s (%d tasks)",
			seq.FullName, seq.Goal.Goal, len(seq.Tasks)))
	}
	return strings.Join(lines, "\n")
}

func formatTaskList(tasks []ProposedTask) string {
	var lines []string
	for _, task := range tasks {
		lines = append(lines, fmt.Sprintf("- %s: %s", task.FullName, task.Objective))
	}
	return strings.Join(lines, "\n")
}
