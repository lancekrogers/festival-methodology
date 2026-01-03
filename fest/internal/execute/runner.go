package execute

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Runner orchestrates festival execution
type Runner struct {
	festivalPath string
	config       *ExecutionConfig
	planBuilder  *PlanBuilder
	stateManager *StateManager
	plan         *ExecutionPlan
}

// NewRunner creates a new execution runner
func NewRunner(festivalPath string, config *ExecutionConfig) *Runner {
	if config == nil {
		config = DefaultConfig()
	}

	return &Runner{
		festivalPath: festivalPath,
		config:       config,
		planBuilder:  NewPlanBuilder(festivalPath),
		stateManager: NewStateManager(festivalPath),
	}
}

// Initialize loads or creates execution state and builds plan
func (r *Runner) Initialize(ctx context.Context) error {
	// Load or create state
	_, err := r.stateManager.Load(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to load state")
	}

	// Build execution plan (includes syncing from progress system)
	plan, err := r.planBuilder.BuildPlan()
	if err != nil {
		return errors.Wrap(err, "failed to build plan")
	}
	r.plan = plan

	// Sync execution state from plan's task statuses
	// (Plan task statuses come from progress system YAML)
	for _, phase := range plan.Phases {
		for _, seq := range phase.Sequences {
			for _, step := range seq.Steps {
				for _, task := range step.Tasks {
					// Only update if task has a non-pending status in the plan
					if task.Status != "" && task.Status != StatusPending {
						r.stateManager.SetTaskStatus(task.ID, task.Status)
					}
				}
			}
		}
	}

	// Update state with total tasks
	r.stateManager.State().TotalTasks = plan.Summary.TotalTasks

	return nil
}

// GetPlan returns the execution plan
func (r *Runner) GetPlan() *ExecutionPlan {
	return r.plan
}

// GetState returns the current execution state
func (r *Runner) GetState() *ExecutionState {
	return r.stateManager.State()
}

// GetNextStep returns the next step to execute
func (r *Runner) GetNextStep() (*StepGroup, *SequenceExecution, *PhaseExecution, error) {
	state := r.stateManager.State()

	for _, phase := range r.plan.Phases {
		if phase.Number < state.CurrentPhase {
			continue
		}

		for _, seq := range phase.Sequences {
			if phase.Number == state.CurrentPhase && seq.Number < state.CurrentSeq {
				continue
			}

			for _, step := range seq.Steps {
				if phase.Number == state.CurrentPhase &&
					seq.Number == state.CurrentSeq &&
					step.Number < state.CurrentStep {
					continue
				}

				// Check if any tasks in this step are pending
				hasPending := false
				for _, task := range step.Tasks {
					status := r.stateManager.GetTaskStatus(task.ID)
					if status == StatusPending || status == StatusInProgress {
						hasPending = true
						break
					}
				}

				if hasPending {
					return step, seq, phase, nil
				}
			}
		}
	}

	return nil, nil, nil, nil
}

// MarkTaskComplete marks a task as completed
func (r *Runner) MarkTaskComplete(ctx context.Context, taskID string) error {
	r.stateManager.SetTaskStatus(taskID, StatusCompleted)
	return r.stateManager.Save(ctx)
}

// MarkTaskSkipped marks a task as skipped
func (r *Runner) MarkTaskSkipped(ctx context.Context, taskID string) error {
	r.stateManager.SetTaskStatus(taskID, StatusSkipped)
	return r.stateManager.Save(ctx)
}

// MarkTaskFailed marks a task as failed
func (r *Runner) MarkTaskFailed(ctx context.Context, taskID string) error {
	r.stateManager.SetTaskStatus(taskID, StatusFailed)
	return r.stateManager.Save(ctx)
}

// AdvancePosition moves to the next position
func (r *Runner) AdvancePosition(ctx context.Context, phase, seq, step int) error {
	r.stateManager.SetCurrentPosition(phase, seq, step)
	return r.stateManager.Save(ctx)
}

// Reset clears execution state
func (r *Runner) Reset(ctx context.Context) error {
	return r.stateManager.Clear(ctx)
}

// FormatDryRun generates a dry-run output showing the execution plan
func (r *Runner) FormatDryRun() string {
	var sb strings.Builder

	sb.WriteString("═══════════════════════════════════════════════════════════\n")
	sb.WriteString("                    EXECUTION PLAN (DRY RUN)                \n")
	sb.WriteString("═══════════════════════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("Festival: %s\n\n", filepath.Base(r.festivalPath)))

	sb.WriteString("SUMMARY\n")
	sb.WriteString("───────\n")
	sb.WriteString(fmt.Sprintf("  Phases:          %d\n", r.plan.Summary.TotalPhases))
	sb.WriteString(fmt.Sprintf("  Sequences:       %d\n", r.plan.Summary.TotalSequences))
	sb.WriteString(fmt.Sprintf("  Tasks:           %d\n", r.plan.Summary.TotalTasks))
	sb.WriteString(fmt.Sprintf("  Execution Steps: %d\n", r.plan.Summary.TotalSteps))
	sb.WriteString(fmt.Sprintf("  Parallel Groups: %d\n", r.plan.Summary.ParallelGroups))
	sb.WriteString(fmt.Sprintf("  Quality Gates:   %d\n", r.plan.Summary.QualityGates))
	sb.WriteString("\n")

	stepNum := 0
	for _, phase := range r.plan.Phases {
		sb.WriteString(fmt.Sprintf("═══ PHASE: %s ═══\n\n", phase.Name))

		for _, seq := range phase.Sequences {
			sb.WriteString(fmt.Sprintf("  ─── Sequence: %s ───\n", seq.Name))

			for _, step := range seq.Steps {
				stepNum++
				stepType := "sequential"
				if step.Parallel {
					stepType = "parallel"
				}

				sb.WriteString(fmt.Sprintf("\n    Step %d (%s):\n", stepNum, stepType))
				for _, task := range step.Tasks {
					autonomy := ""
					if task.AutonomyLevel != "" {
						autonomy = fmt.Sprintf(" [%s]", task.AutonomyLevel)
					}
					sb.WriteString(fmt.Sprintf("      • %s%s\n", task.Name, autonomy))
				}
			}
			sb.WriteString("\n")
		}

		if phase.QualityGate != nil {
			sb.WriteString("  ⚠️  QUALITY GATE before next phase\n\n")
		}
	}

	return sb.String()
}

// FormatAgentInstructions generates agent-friendly execution instructions
func (r *Runner) FormatAgentInstructions() (string, error) {
	step, seq, phase, err := r.GetNextStep()
	if err != nil {
		return "", err
	}

	if step == nil {
		return "EXECUTION COMPLETE\n\nAll tasks have been executed.\n", nil
	}

	var sb strings.Builder

	sb.WriteString("AGENT EXECUTION INSTRUCTIONS\n")
	sb.WriteString("════════════════════════════\n\n")

	state := r.stateManager.State()
	progress := state.Progress()

	sb.WriteString(fmt.Sprintf("PROGRESS: %.1f%% (%d/%d tasks)\n\n",
		progress, state.CompletedTasks, state.TotalTasks))

	sb.WriteString(fmt.Sprintf("CURRENT PHASE: %s\n", phase.Name))
	sb.WriteString(fmt.Sprintf("CURRENT SEQUENCE: %s\n", seq.Name))
	sb.WriteString(fmt.Sprintf("CURRENT STEP: %d\n\n", step.Number))

	sb.WriteString("TASKS TO EXECUTE:\n")
	for _, task := range step.Tasks {
		status := r.stateManager.GetTaskStatus(task.ID)
		statusIcon := "○"
		if status == StatusCompleted {
			statusIcon = "✓"
		} else if status == StatusInProgress {
			statusIcon = "►"
		}
		sb.WriteString(fmt.Sprintf("  %s %s\n", statusIcon, task.Name))
		sb.WriteString(fmt.Sprintf("    Path: %s\n", task.Path))
		if task.AutonomyLevel != "" {
			sb.WriteString(fmt.Sprintf("    Autonomy: %s\n", task.AutonomyLevel))
		}
	}

	sb.WriteString("\nCONTEXT FILES:\n")
	sb.WriteString(fmt.Sprintf("  • %s/FESTIVAL_GOAL.md\n", r.festivalPath))
	sb.WriteString(fmt.Sprintf("  • %s/PHASE_GOAL.md\n", phase.Path))
	sb.WriteString(fmt.Sprintf("  • %s/SEQUENCE_GOAL.md\n", seq.Path))

	sb.WriteString("\nCOMPLETION COMMANDS:\n")
	for _, task := range step.Tasks {
		sb.WriteString(fmt.Sprintf("  fest status set %s completed\n", task.Name))
	}

	sb.WriteString("\nNEXT STEP COMMAND:\n")
	sb.WriteString("  fest execute --agent --continue\n")

	return sb.String(), nil
}

// FormatJSON returns the plan as JSON-serializable structure
func (r *Runner) FormatJSON() *ExecutionPlan {
	return r.plan
}
