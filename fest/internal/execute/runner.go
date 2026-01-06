package execute

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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

	sb.WriteString(ui.H1("Execution Plan (Dry Run)"))
	sb.WriteString("\n\n")

	writeDryRunSummary(&sb, r)
	writeDryRunPhases(&sb, r)

	return sb.String()
}

// FormatAgentInstructions generates agent-friendly execution instructions
func (r *Runner) FormatAgentInstructions() (string, error) {
	step, seq, phase, err := r.GetNextStep()
	if err != nil {
		return "", err
	}

	if step == nil {
		return formatExecutionComplete(), nil
	}

	var sb strings.Builder

	sb.WriteString(ui.H1("Agent Execution Instructions"))
	sb.WriteString("\n\n")

	writeAgentProgress(&sb, r)
	writeAgentCurrentPosition(&sb, phase, seq, step)
	writeAgentTasks(&sb, r, step)
	sequencePath := seq.Path
	if sequencePath == "" && phase.Path != "" {
		sequencePath = filepath.Join(phase.Path, seq.Name)
	}
	if sequencePath == "" && len(step.Tasks) > 0 {
		sequencePath = filepath.Dir(step.Tasks[0].Path)
	}
	writeAgentContextFiles(&sb, r.festivalPath, phase.Path, sequencePath)
	writeAgentCompletionCommands(&sb, step)
	writeAgentNextStepCommand(&sb)

	return sb.String(), nil
}

func writeLabelValue(sb *strings.Builder, label, value string) {
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label(label), value))
}

func writeDryRunSummary(sb *strings.Builder, r *Runner) {
	writeLabelValue(sb, "Festival", ui.Value(filepath.Base(r.festivalPath), ui.FestivalColor))
	writeLabelValue(sb, "Path", ui.Dim(r.festivalPath))
	sb.WriteString("\n")

	sb.WriteString(ui.H2("Summary"))
	sb.WriteString("\n")
	writeLabelValue(sb, "Phases", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalPhases)))
	writeLabelValue(sb, "Sequences", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalSequences)))
	writeLabelValue(sb, "Tasks", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalTasks)))
	writeLabelValue(sb, "Execution steps", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalSteps)))
	writeLabelValue(sb, "Parallel groups", ui.Value(fmt.Sprintf("%d", r.plan.Summary.ParallelGroups)))
	writeLabelValue(sb, "Quality gates", ui.Value(fmt.Sprintf("%d", r.plan.Summary.QualityGates)))
}

func writeDryRunPhases(sb *strings.Builder, r *Runner) {
	stepNum := 0
	for _, phase := range r.plan.Phases {
		sb.WriteString("\n")
		sb.WriteString(ui.H2(fmt.Sprintf("Phase %s", phase.Name)))
		sb.WriteString("\n")

		for _, seq := range phase.Sequences {
			sb.WriteString(ui.H3(fmt.Sprintf("Sequence %s", seq.Name)))

			for _, step := range seq.Steps {
				stepNum++
				stepType := "sequential"
				if step.Parallel {
					stepType = "parallel"
				}

				sb.WriteString(fmt.Sprintf("\n%s %s\n", ui.Label(fmt.Sprintf("Step %d", stepNum)), ui.Dim(fmt.Sprintf("(%s)", stepType))))
				for _, task := range step.Tasks {
					autonomy := ""
					if task.AutonomyLevel != "" {
						autonomy = fmt.Sprintf(" %s", ui.Dim(fmt.Sprintf("[%s]", task.AutonomyLevel)))
					}
					sb.WriteString(fmt.Sprintf("  - %s%s\n", ui.Value(task.Name, ui.TaskColor), autonomy))
				}
			}
			sb.WriteString("\n\n")
		}

		if phase.QualityGate != nil {
			sb.WriteString(ui.Warning("Quality gate required before next phase"))
			sb.WriteString("\n")
		}
	}
}

func formatExecutionComplete() string {
	var sb strings.Builder
	sb.WriteString(ui.H1("Execution Complete"))
	sb.WriteString("\n\n")
	sb.WriteString(ui.Success("All tasks have been executed."))
	sb.WriteString("\n")
	return sb.String()
}

func writeAgentProgress(sb *strings.Builder, r *Runner) {
	state := r.stateManager.State()
	progress := state.Progress()
	writeLabelValue(sb, "Progress", ui.Value(fmt.Sprintf("%.1f%% (%d/%d tasks)", progress, state.CompletedTasks, state.TotalTasks)))
	sb.WriteString("\n")
}

func writeAgentCurrentPosition(sb *strings.Builder, phase *PhaseExecution, seq *SequenceExecution, step *StepGroup) {
	sb.WriteString(ui.H2("Current Position"))
	sb.WriteString("\n")
	writeLabelValue(sb, "Phase", ui.Value(phase.Name, ui.PhaseColor))
	writeLabelValue(sb, "Sequence", ui.Value(seq.Name, ui.SequenceColor))
	writeLabelValue(sb, "Step", ui.Value(fmt.Sprintf("%d", step.Number)))
	sb.WriteString("\n")
}

func writeAgentTasks(sb *strings.Builder, r *Runner, step *StepGroup) {
	sb.WriteString(ui.H2("Tasks to Execute"))
	sb.WriteString("\n")
	for _, task := range step.Tasks {
		status := r.stateManager.GetTaskStatus(task.ID)
		statusIcon := ui.StateIcon(status)
		sb.WriteString(fmt.Sprintf("  %s %s\n", statusIcon, ui.Value(task.Name, ui.TaskColor)))
		sb.WriteString(fmt.Sprintf("    %s %s\n", ui.Label("Path"), ui.Dim(task.Path)))
		if task.AutonomyLevel != "" {
			sb.WriteString(fmt.Sprintf("    %s %s\n", ui.Label("Autonomy"), ui.Value(task.AutonomyLevel)))
		}
	}
	sb.WriteString("\n")
}

func writeAgentContextFiles(sb *strings.Builder, festivalPath, phasePath, seqPath string) {
	sb.WriteString(ui.H2("Context Files"))
	sb.WriteString("\n")
	festivalGoal := "FESTIVAL_GOAL.md"
	if festivalPath != "" {
		festivalGoal = fmt.Sprintf("%s/%s", festivalPath, festivalGoal)
	}
	phaseGoal := "PHASE_GOAL.md"
	if phasePath != "" {
		phaseGoal = fmt.Sprintf("%s/%s", phasePath, phaseGoal)
	}
	sequenceGoal := "SEQUENCE_GOAL.md"
	if seqPath != "" {
		sequenceGoal = fmt.Sprintf("%s/%s", seqPath, sequenceGoal)
	}
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(festivalGoal)))
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(phaseGoal)))
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(sequenceGoal)))
	sb.WriteString("\n")
}

func writeAgentCompletionCommands(sb *strings.Builder, step *StepGroup) {
	sb.WriteString(ui.H2("Completion Commands"))
	sb.WriteString("\n")
	for _, task := range step.Tasks {
		sb.WriteString(fmt.Sprintf("  %s\n", ui.Value(fmt.Sprintf("fest status set %s completed", task.Name))))
	}
	sb.WriteString("\n")
}

func writeAgentNextStepCommand(sb *strings.Builder) {
	sb.WriteString(ui.H2("Next Step Command"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  %s\n", ui.Value("fest execute --agent --continue")))
}

// FormatJSON returns the plan as JSON-serializable structure
func (r *Runner) FormatJSON() *ExecutionPlan {
	return r.plan
}
