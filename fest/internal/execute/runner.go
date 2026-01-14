package execute

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/templates/agent"
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
	plan, err := r.planBuilder.BuildPlan(ctx)
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
	data := struct {
		Header         string
		FestivalLine   string
		PathLine       string
		SummarySection string
		PhasesSection  string
	}{
		Header:         ui.H1("Execution Plan (Dry Run)"),
		FestivalLine:   buildLabelValue("Festival", ui.Value(filepath.Base(r.festivalPath), ui.FestivalColor)),
		PathLine:       buildLabelValue("Path", ui.Dim(r.festivalPath)),
		SummarySection: buildDryRunSummary(r),
		PhasesSection:  buildDryRunPhases(r),
	}

	var buf bytes.Buffer
	agent.MustGet("execute/dry_run").Execute(&buf, data)
	return buf.String()
}

// buildLabelValue creates a label-value pair string
func buildLabelValue(label, value string) string {
	var sb strings.Builder
	ui.WriteLabelValue(&sb, label, value)
	return strings.TrimSuffix(sb.String(), "\n")
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

	// Build sequence path for context files
	sequencePath := seq.Path
	if sequencePath == "" && phase.Path != "" {
		sequencePath = filepath.Join(phase.Path, seq.Name)
	}
	if sequencePath == "" && len(step.Tasks) > 0 {
		sequencePath = filepath.Dir(step.Tasks[0].Path)
	}

	data := struct {
		Header            string
		ProgressLine      string
		PositionSection   string
		TasksSection      string
		ActionInstruction string
		ProgressCmd       string
		ContextSection    string
	}{
		Header:            ui.H1("Agent Execution Instructions"),
		ProgressLine:      buildAgentProgressLine(r),
		PositionSection:   buildAgentPositionSection(phase, seq, step),
		TasksSection:      buildAgentTasksSection(r, step),
		ActionInstruction: ui.Info("Read the task file and follow the instructions laid out exactly."),
		ProgressCmd:       buildProgressCommand(step),
		ContextSection:    buildAgentContextSection(r.festivalPath, phase.Path, sequencePath),
	}

	var buf bytes.Buffer
	agent.MustGet("execute/instructions").Execute(&buf, data)
	return buf.String(), nil
}

func buildDryRunSummary(r *Runner) string {
	var sb strings.Builder
	sb.WriteString(ui.H2("Summary"))
	sb.WriteString("\n")
	ui.WriteLabelValue(&sb, "Phases", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalPhases)))
	ui.WriteLabelValue(&sb, "Sequences", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalSequences)))
	ui.WriteLabelValue(&sb, "Tasks", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalTasks)))
	ui.WriteLabelValue(&sb, "Execution steps", ui.Value(fmt.Sprintf("%d", r.plan.Summary.TotalSteps)))
	ui.WriteLabelValue(&sb, "Parallel groups", ui.Value(fmt.Sprintf("%d", r.plan.Summary.ParallelGroups)))
	ui.WriteLabelValue(&sb, "Quality gates", ui.Value(fmt.Sprintf("%d", r.plan.Summary.QualityGates)))
	return sb.String()
}

func buildDryRunPhases(r *Runner) string {
	var sb strings.Builder
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
	return sb.String()
}

func formatExecutionComplete() string {
	data := struct {
		Header  string
		Message string
	}{
		Header:  ui.H1("Execution Complete"),
		Message: ui.Success("All tasks have been executed."),
	}

	var buf bytes.Buffer
	agent.MustGet("execute/complete").Execute(&buf, data)
	return buf.String()
}

func buildAgentProgressLine(r *Runner) string {
	state := r.stateManager.State()
	progress := state.Progress()
	var sb strings.Builder
	ui.WriteLabelValue(&sb, "Progress", ui.Value(fmt.Sprintf("%.1f%% (%d/%d tasks)", progress, state.CompletedTasks, state.TotalTasks)))
	return sb.String()
}

func buildAgentPositionSection(phase *PhaseExecution, seq *SequenceExecution, step *StepGroup) string {
	var sb strings.Builder
	sb.WriteString(ui.H2("Current Position"))
	sb.WriteString("\n")
	ui.WriteLabelValue(&sb, "Phase", ui.Value(phase.Name, ui.PhaseColor))
	ui.WriteLabelValue(&sb, "Sequence", ui.Value(seq.Name, ui.SequenceColor))
	ui.WriteLabelValue(&sb, "Step", ui.Value(fmt.Sprintf("%d", step.Number)))
	return sb.String()
}

func buildAgentTasksSection(r *Runner, step *StepGroup) string {
	var sb strings.Builder
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
	return sb.String()
}

func buildAgentContextSection(festivalPath, phasePath, seqPath string) string {
	var sb strings.Builder
	sb.WriteString(ui.H2("Context Files"))
	sb.WriteString("\n")
	festivalGoal := filepath.Join(festivalPath, "FESTIVAL_GOAL.md")
	phaseGoal := filepath.Join(phasePath, "PHASE_GOAL.md")
	sequenceGoal := filepath.Join(seqPath, "SEQUENCE_GOAL.md")
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(festivalGoal)))
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(phaseGoal)))
	sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(sequenceGoal)))
	return sb.String()
}

// buildProgressCommand builds the fest progress command for marking a task complete
func buildProgressCommand(step *StepGroup) string {
	if len(step.Tasks) == 0 {
		return ""
	}
	// Use the first task to build the progress command
	task := step.Tasks[0]
	return ui.Value(fmt.Sprintf("fest progress --task %s --complete", task.Path))
}

// FormatJSON returns the plan as JSON-serializable structure
func (r *Runner) FormatJSON() *ExecutionPlan {
	return r.plan
}
