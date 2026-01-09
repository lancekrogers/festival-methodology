package execute

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/lancekrogers/festival-methodology/fest/internal/deps"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
)

// ExecutionPlan represents the complete plan for executing a festival
type ExecutionPlan struct {
	FestivalPath string            `json:"festival_path"`
	Phases       []*PhaseExecution `json:"phases"`
	Summary      *ExecutionSummary `json:"summary"`
}

// PhaseExecution represents execution plan for a phase
type PhaseExecution struct {
	Name        string               `json:"name"`
	Path        string               `json:"path"`
	Number      int                  `json:"number"`
	Sequences   []*SequenceExecution `json:"sequences"`
	QualityGate *QualityGateInfo     `json:"quality_gate,omitempty"`
	TotalTasks  int                  `json:"total_tasks"`
	Status      string               `json:"status"`
}

// SequenceExecution represents execution plan for a sequence
type SequenceExecution struct {
	Name       string       `json:"name"`
	Path       string       `json:"path"`
	Number     int          `json:"number"`
	Steps      []*StepGroup `json:"steps"`
	TotalTasks int          `json:"total_tasks"`
	Status     string       `json:"status"`
}

// StepGroup represents a group of tasks to execute together
type StepGroup struct {
	Number   int         `json:"number"`
	Type     string      `json:"type"` // "parallel" or "sequential"
	Tasks    []*PlanTask `json:"tasks"`
	Parallel bool        `json:"parallel"`
}

// PlanTask represents a task in the execution plan
type PlanTask struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	Number        int      `json:"number"`
	AutonomyLevel string   `json:"autonomy_level"`
	Dependencies  []string `json:"dependencies,omitempty"`
	Status        string   `json:"status"`
}

// QualityGateInfo describes a quality gate
type QualityGateInfo struct {
	PhaseName   string   `json:"phase_name"`
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Criteria    []string `json:"criteria,omitempty"`
	Passed      bool     `json:"passed"`
}

// ExecutionSummary provides summary statistics for the plan
type ExecutionSummary struct {
	TotalPhases    int    `json:"total_phases"`
	TotalSequences int    `json:"total_sequences"`
	TotalTasks     int    `json:"total_tasks"`
	TotalSteps     int    `json:"total_steps"`
	ParallelGroups int    `json:"parallel_groups"`
	QualityGates   int    `json:"quality_gates"`
	EstimatedTime  string `json:"estimated_time,omitempty"`
}

// PlanBuilder builds execution plans from festival structure
type PlanBuilder struct {
	festivalPath string
	resolver     *deps.Resolver
}

// NewPlanBuilder creates a new execution plan builder
func NewPlanBuilder(festivalPath string) *PlanBuilder {
	return &PlanBuilder{
		festivalPath: festivalPath,
		resolver:     deps.NewResolver(festivalPath),
	}
}

// BuildPlan creates the complete execution plan for the festival
func (b *PlanBuilder) BuildPlan(ctx context.Context) (*ExecutionPlan, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	graph, err := b.resolver.ResolveFestival()
	if err != nil {
		return nil, err
	}

	// Update task statuses from progress system (YAML source of truth)
	if err := b.updateTaskStatusesFromProgress(ctx, graph); err != nil {
		return nil, err
	}

	// Group tasks by phase and sequence
	byPhase := b.groupTasksByPhase(graph.Tasks)

	// Build phase execution plans
	var phases []*PhaseExecution
	var phaseNames []string
	for name := range byPhase {
		phaseNames = append(phaseNames, name)
	}
	sort.Strings(phaseNames)

	summary := &ExecutionSummary{}

	for _, phaseName := range phaseNames {
		seqTasks := byPhase[phaseName]
		phaseExec := b.buildPhaseExecution(phaseName, seqTasks)
		phases = append(phases, phaseExec)
		summary.TotalPhases++
		summary.TotalSequences += len(phaseExec.Sequences)
		summary.TotalTasks += phaseExec.TotalTasks
		for _, seq := range phaseExec.Sequences {
			summary.TotalSteps += len(seq.Steps)
			for _, step := range seq.Steps {
				if step.Parallel && len(step.Tasks) > 1 {
					summary.ParallelGroups++
				}
			}
		}
		if phaseExec.QualityGate != nil {
			summary.QualityGates++
		}
	}

	return &ExecutionPlan{
		FestivalPath: b.festivalPath,
		Phases:       phases,
		Summary:      summary,
	}, nil
}

// BuildPlanForPhase creates execution plan for a specific phase
func (b *PlanBuilder) BuildPlanForPhase(ctx context.Context, phasePath string) (*PhaseExecution, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	graph, err := b.resolver.ResolveFestival()
	if err != nil {
		return nil, err
	}

	// Update task statuses from progress system (YAML source of truth)
	if err := b.updateTaskStatusesFromProgress(ctx, graph); err != nil {
		return nil, err
	}

	// Filter tasks for this phase
	byPhase := b.groupTasksByPhase(graph.Tasks)
	phaseName := filepath.Base(phasePath)

	seqTasks, ok := byPhase[phaseName]
	if !ok {
		// Try with path
		for name, tasks := range byPhase {
			if tasks[filepath.Base(name)] != nil {
				seqTasks = tasks
				phaseName = name
				break
			}
		}
	}

	return b.buildPhaseExecution(phaseName, seqTasks), nil
}

// BuildPlanForSequence creates execution plan for a specific sequence
func (b *PlanBuilder) BuildPlanForSequence(ctx context.Context, seqPath string) (*SequenceExecution, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	graph, err := b.resolver.ResolveSequence(seqPath)
	if err != nil {
		return nil, err
	}

	// Update task statuses from progress system (YAML source of truth)
	if err := b.updateTaskStatusesFromProgress(ctx, graph); err != nil {
		return nil, err
	}

	// Convert to plan tasks
	var tasks []*PlanTask
	for _, task := range graph.Tasks {
		tasks = append(tasks, b.taskToPlanTask(task))
	}

	return b.buildSequenceExecution(filepath.Base(seqPath), tasks), nil
}

// groupTasksByPhase organizes tasks into a hierarchy by phase and sequence
func (b *PlanBuilder) groupTasksByPhase(tasks map[string]*deps.Task) map[string]map[string][]*PlanTask {
	byPhase := make(map[string]map[string][]*PlanTask)

	for _, task := range tasks {
		phaseName := filepath.Base(task.PhasePath)
		seqName := filepath.Base(task.SequencePath)

		if byPhase[phaseName] == nil {
			byPhase[phaseName] = make(map[string][]*PlanTask)
		}

		planTask := b.taskToPlanTask(task)
		byPhase[phaseName][seqName] = append(byPhase[phaseName][seqName], planTask)
	}

	return byPhase
}

// buildPhaseExecution creates a PhaseExecution from tasks
func (b *PlanBuilder) buildPhaseExecution(phaseName string, seqTasks map[string][]*PlanTask) *PhaseExecution {
	phase := &PhaseExecution{
		Name:   phaseName,
		Path:   filepath.Join(b.festivalPath, phaseName),
		Number: extractPhaseNumber(phaseName),
		Status: "pending",
	}

	// Build sequence executions
	var seqNames []string
	for name := range seqTasks {
		seqNames = append(seqNames, name)
	}
	sort.Strings(seqNames)

	for _, seqName := range seqNames {
		tasks := seqTasks[seqName]
		seqExec := b.buildSequenceExecution(seqName, tasks)
		phase.Sequences = append(phase.Sequences, seqExec)
		phase.TotalTasks += seqExec.TotalTasks
	}

	return phase
}

// buildSequenceExecution creates a SequenceExecution from tasks
func (b *PlanBuilder) buildSequenceExecution(seqName string, tasks []*PlanTask) *SequenceExecution {
	seq := &SequenceExecution{
		Name:       seqName,
		Number:     extractSeqNumber(seqName),
		TotalTasks: len(tasks),
		Status:     "pending",
	}

	// Group tasks by number for step grouping
	byNumber := make(map[int][]*PlanTask)
	for _, task := range tasks {
		byNumber[task.Number] = append(byNumber[task.Number], task)
	}

	// Get sorted numbers
	var numbers []int
	for num := range byNumber {
		numbers = append(numbers, num)
	}
	sort.Ints(numbers)

	// Create step groups
	for _, num := range numbers {
		stepTasks := byNumber[num]
		parallel := len(stepTasks) > 1

		stepType := "sequential"
		if parallel {
			stepType = "parallel"
		}

		step := &StepGroup{
			Number:   num,
			Type:     stepType,
			Tasks:    stepTasks,
			Parallel: parallel,
		}
		seq.Steps = append(seq.Steps, step)
	}

	return seq
}

// taskToPlanTask converts a deps.Task to a PlanTask
func (b *PlanBuilder) taskToPlanTask(task *deps.Task) *PlanTask {
	return &PlanTask{
		ID:            task.ID,
		Name:          task.Name,
		Path:          task.Path,
		Number:        task.Number,
		AutonomyLevel: task.AutonomyLevel,
		Dependencies:  task.Dependencies,
		Status:        task.Status,
	}
}

// extractPhaseNumber extracts the numeric prefix from a phase name
func extractPhaseNumber(name string) int {
	num := 0
	for _, c := range name {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	return num
}

// extractSeqNumber extracts the numeric prefix from a sequence name
func extractSeqNumber(name string) int {
	return extractPhaseNumber(name)
}

// updateTaskStatusesFromProgress updates all task statuses in the graph
// by querying the progress tracking system (YAML source of truth)
func (b *PlanBuilder) updateTaskStatusesFromProgress(ctx context.Context, graph *deps.Graph) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Create progress manager
	mgr, err := progress.NewManager(ctx, b.festivalPath)
	if err != nil {
		return err
	}

	// Update each task's status from YAML (or markdown fallback)
	for _, task := range graph.Tasks {
		// ResolveTaskStatus checks YAML first, falls back to markdown
		status := progress.ResolveTaskStatus(mgr.Store(), b.festivalPath, task.Path)

		// Map progress status to execute status constants
		switch status {
		case progress.StatusCompleted:
			task.Status = StatusCompleted // "completed"
		case progress.StatusInProgress:
			task.Status = StatusInProgress // "in_progress"
		case progress.StatusBlocked:
			task.Status = "blocked"
		default:
			task.Status = StatusPending // "pending"
		}
	}

	return nil
}
