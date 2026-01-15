// Package next provides task selection and navigation for festival workflows.
package next

import (
	"bufio"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/deps"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
)

// NextTaskResult represents the result of finding the next task
type NextTaskResult struct {
	// Primary task recommendation
	Task *TaskInfo `json:"task,omitempty"`

	// Alternative tasks that can be done in parallel
	ParallelTasks []*TaskInfo `json:"parallel_tasks,omitempty"`

	// Quality gate if one is blocking progress
	BlockingGate *GateInfo `json:"blocking_gate,omitempty"`

	// Planning phase info (when in a planning/research phase)
	Planning *PlanningPhaseResult `json:"planning,omitempty"`

	// Reason for the recommendation
	Reason string `json:"reason"`

	// Whether we're at the end of the festival
	FestivalComplete bool `json:"festival_complete"`

	// Current location context
	Location *LocationInfo `json:"location"`

	// Progress statistics
	Progress *ProgressInfo `json:"progress,omitempty"`
}

// PlanningPhaseResult contains information for planning/research phases
type PlanningPhaseResult struct {
	PhaseName       string               `json:"phase_name"`
	PhasePath       string               `json:"phase_path"`
	PhaseType       string               `json:"phase_type"`
	Objectives      []*PlanningObjective `json:"objectives"`
	Progress        *PlanningProgress    `json:"progress"`
	GraduationReady bool                 `json:"graduation_ready"`
}

// PlanningObjective represents a planning objective from PHASE_GOAL.md
type PlanningObjective struct {
	Category string `json:"category"` // "question", "decision", "artifact"
	Text     string `json:"text"`
	Resolved bool   `json:"resolved"`
}

// PlanningProgress contains progress info for planning objectives
type PlanningProgress struct {
	TotalObjectives    int     `json:"total_objectives"`
	ResolvedObjectives int     `json:"resolved_objectives"`
	Percentage         float64 `json:"percentage"`
}

// ProgressInfo contains festival progress statistics
type ProgressInfo struct {
	TotalTasks     int     `json:"total_tasks"`
	CompletedTasks int     `json:"completed_tasks"`
	Percentage     float64 `json:"percentage"`
}

// TaskInfo contains information about a task
type TaskInfo struct {
	Name          string   `json:"name"`
	Path          string   `json:"path"`
	Number        int      `json:"number"`
	SequenceName  string   `json:"sequence_name"`
	SequencePath  string   `json:"sequence_path"`
	PhaseName     string   `json:"phase_name"`
	PhasePath     string   `json:"phase_path"`
	Status        string   `json:"status"`
	AutonomyLevel string   `json:"autonomy_level,omitempty"`
	ParallelGroup int      `json:"parallel_group"`
	Dependencies  []string `json:"dependencies,omitempty"`
}

// GateInfo contains information about a quality gate
type GateInfo struct {
	Phase       string   `json:"phase"`
	GateType    string   `json:"gate_type"`
	Description string   `json:"description"`
	Criteria    []string `json:"criteria,omitempty"`
}

// LocationInfo contains current location context
type LocationInfo struct {
	FestivalPath string `json:"festival_path"`
	PhasePath    string `json:"phase_path,omitempty"`
	SequencePath string `json:"sequence_path,omitempty"`
	CurrentPath  string `json:"current_path"`
}

// Selector finds the next task to work on
type Selector struct {
	festivalPath string
	resolver     *deps.Resolver
}

// NewSelector creates a new task selector
func NewSelector(festivalPath string) *Selector {
	return &Selector{
		festivalPath: festivalPath,
		resolver:     deps.NewResolver(festivalPath),
	}
}

// FindNext finds the next task to work on from the current location
func (s *Selector) FindNext(ctx context.Context, currentPath string) (*NextTaskResult, error) {
	// Build the dependency graph
	graph, err := s.resolver.ResolveFestival()
	if err != nil {
		return nil, err
	}

	// Update task statuses from progress system (YAML source of truth)
	if err := s.updateTaskStatusesFromProgress(ctx, graph); err != nil {
		return nil, err
	}

	// Determine current location
	location := s.determineLocation(currentPath)

	// Check if current phase is a planning/research phase
	if location.PhasePath != "" {
		phaseType, err := s.getPhaseType(location.PhasePath)
		if err == nil && isPlanningPhaseType(phaseType) {
			return s.findNextPlanning(ctx, location, phaseType)
		}
	}

	// Get all ready tasks (those with all dependencies satisfied)
	readyTasks := graph.GetReadyTasks()

	if len(readyTasks) == 0 {
		// Check if festival is complete
		if s.isFestivalComplete(graph) {
			return &NextTaskResult{
				FestivalComplete: true,
				Reason:           "All tasks in the festival are complete",
				Location:         location,
			}, nil
		}

		// Check for blocking quality gate
		gate := s.findBlockingGate(graph)
		if gate != nil {
			return &NextTaskResult{
				BlockingGate: gate,
				Reason:       "Quality gate must be passed before proceeding",
				Location:     location,
			}, nil
		}

		return &NextTaskResult{
			Reason:   "No tasks are currently ready (dependencies not satisfied)",
			Location: location,
		}, nil
	}

	// Prioritize tasks
	prioritized := s.prioritizeTasks(readyTasks, location)

	// Get the primary task
	primary := prioritized[0]
	taskInfo := s.taskToInfo(primary)

	// Find parallel tasks
	parallelTasks := s.findParallelTasks(prioritized, primary)

	// Calculate progress
	progress := s.calculateProgress(graph)

	result := &NextTaskResult{
		Task:          taskInfo,
		ParallelTasks: parallelTasks,
		Reason:        s.generateReason(primary, location),
		Location:      location,
		Progress:      progress,
	}

	return result, nil
}

// calculateProgress computes progress statistics from the graph
func (s *Selector) calculateProgress(graph *deps.Graph) *ProgressInfo {
	if graph == nil || len(graph.Tasks) == 0 {
		return nil
	}

	total := len(graph.Tasks)
	completed := 0
	for _, task := range graph.Tasks {
		if task.Status == "complete" {
			completed++
		}
	}

	percentage := 0.0
	if total > 0 {
		percentage = float64(completed) / float64(total) * 100
	}

	return &ProgressInfo{
		TotalTasks:     total,
		CompletedTasks: completed,
		Percentage:     percentage,
	}
}

// FindNextInSequence finds the next task within the current sequence
func (s *Selector) FindNextInSequence(ctx context.Context, seqPath string) (*NextTaskResult, error) {
	graph, err := s.resolver.ResolveSequence(seqPath)
	if err != nil {
		return nil, err
	}

	// Update task statuses from progress system (YAML source of truth)
	if err := s.updateTaskStatusesFromProgress(ctx, graph); err != nil {
		return nil, err
	}

	location := s.determineLocation(seqPath)
	readyTasks := graph.GetReadyTasks()

	if len(readyTasks) == 0 {
		// Check if sequence is complete
		allComplete := true
		for _, task := range graph.Tasks {
			if task.Status != "complete" {
				allComplete = false
				break
			}
		}

		if allComplete {
			return &NextTaskResult{
				Reason:   "All tasks in sequence are complete",
				Location: location,
			}, nil
		}

		return &NextTaskResult{
			Reason:   "No tasks are ready (dependencies not satisfied)",
			Location: location,
		}, nil
	}

	// Sort by task number
	sort.Slice(readyTasks, func(i, j int) bool {
		return readyTasks[i].Number < readyTasks[j].Number
	})

	primary := readyTasks[0]
	taskInfo := s.taskToInfo(primary)

	// Find parallel tasks in same group
	var parallelTasks []*TaskInfo
	for _, task := range readyTasks[1:] {
		if task.ParallelGroup == primary.ParallelGroup {
			parallelTasks = append(parallelTasks, s.taskToInfo(task))
		}
	}

	return &NextTaskResult{
		Task:          taskInfo,
		ParallelTasks: parallelTasks,
		Reason:        "Next task in sequence",
		Location:      location,
	}, nil
}

// determineLocation identifies the current location context
func (s *Selector) determineLocation(currentPath string) *LocationInfo {
	location := &LocationInfo{
		FestivalPath: s.festivalPath,
		CurrentPath:  currentPath,
	}

	// Try to determine if we're in a phase or sequence
	rel, err := filepath.Rel(s.festivalPath, currentPath)
	if err != nil || strings.HasPrefix(rel, "..") {
		return location
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) >= 1 && isNumberedDir(parts[0]) {
		location.PhasePath = filepath.Join(s.festivalPath, parts[0])
	}
	if len(parts) >= 2 && isNumberedDir(parts[1]) {
		location.SequencePath = filepath.Join(s.festivalPath, parts[0], parts[1])
	}

	return location
}

// prioritizeTasks orders tasks by priority
func (s *Selector) prioritizeTasks(tasks []*deps.Task, location *LocationInfo) []*deps.Task {
	// Create a copy to sort
	sorted := make([]*deps.Task, len(tasks))
	copy(sorted, tasks)

	sort.Slice(sorted, func(i, j int) bool {
		ti, tj := sorted[i], sorted[j]

		// Priority 1: Current sequence tasks first
		if location.SequencePath != "" {
			inSeqI := ti.SequencePath == location.SequencePath
			inSeqJ := tj.SequencePath == location.SequencePath
			if inSeqI != inSeqJ {
				return inSeqI
			}
		}

		// Priority 2: Current phase tasks
		if location.PhasePath != "" {
			inPhaseI := ti.PhasePath == location.PhasePath
			inPhaseJ := tj.PhasePath == location.PhasePath
			if inPhaseI != inPhaseJ {
				return inPhaseI
			}
		}

		// Priority 3: Earlier phases first
		if ti.PhasePath != tj.PhasePath {
			return ti.PhasePath < tj.PhasePath
		}

		// Priority 4: Earlier sequences first
		if ti.SequencePath != tj.SequencePath {
			return ti.SequencePath < tj.SequencePath
		}

		// Priority 5: Lower task number first
		return ti.Number < tj.Number
	})

	return sorted
}

// findParallelTasks finds tasks that can be done in parallel with the primary
func (s *Selector) findParallelTasks(tasks []*deps.Task, primary *deps.Task) []*TaskInfo {
	var parallel []*TaskInfo

	for _, task := range tasks {
		if task.ID == primary.ID {
			continue
		}

		// Same parallel group in same sequence
		if task.SequencePath == primary.SequencePath &&
			task.ParallelGroup == primary.ParallelGroup {
			parallel = append(parallel, s.taskToInfo(task))
		}
	}

	return parallel
}

// taskToInfo converts a deps.Task to TaskInfo
func (s *Selector) taskToInfo(task *deps.Task) *TaskInfo {
	return &TaskInfo{
		Name:          task.Name,
		Path:          task.Path,
		Number:        task.Number,
		SequenceName:  filepath.Base(task.SequencePath),
		SequencePath:  task.SequencePath,
		PhaseName:     filepath.Base(task.PhasePath),
		PhasePath:     task.PhasePath,
		Status:        task.Status,
		AutonomyLevel: task.AutonomyLevel,
		ParallelGroup: task.ParallelGroup,
		Dependencies:  task.Dependencies,
	}
}

// isFestivalComplete checks if all tasks are done
func (s *Selector) isFestivalComplete(graph *deps.Graph) bool {
	for _, task := range graph.Tasks {
		if task.Status != "complete" {
			return false
		}
	}
	return true
}

// findBlockingGate checks for quality gates blocking progress
func (s *Selector) findBlockingGate(graph *deps.Graph) *GateInfo {
	// Group tasks by phase
	byPhase := make(map[string][]*deps.Task)
	for _, task := range graph.Tasks {
		byPhase[task.PhasePath] = append(byPhase[task.PhasePath], task)
	}

	// Sort phases
	var phases []string
	for phase := range byPhase {
		phases = append(phases, phase)
	}
	sort.Strings(phases)

	// Check each phase for incomplete tasks and gates
	for i, phase := range phases {
		tasks := byPhase[phase]

		// Check if phase is complete
		allComplete := true
		for _, task := range tasks {
			if task.Status != "complete" {
				allComplete = false
				break
			}
		}

		// If phase complete and there's a next phase, check for gate
		if allComplete && i < len(phases)-1 {
			gateFile := filepath.Join(phase, "QUALITY_GATE.md")
			if _, err := os.Stat(gateFile); err == nil {
				return &GateInfo{
					Phase:       filepath.Base(phase),
					GateType:    "phase_transition",
					Description: "Quality gate must be passed before moving to next phase",
				}
			}
		}
	}

	return nil
}

// generateReason creates a human-readable reason for the recommendation
func (s *Selector) generateReason(task *deps.Task, location *LocationInfo) string {
	if location.SequencePath != "" && task.SequencePath == location.SequencePath {
		return "Next task in current sequence"
	}
	if location.PhasePath != "" && task.PhasePath == location.PhasePath {
		return "Next task in current phase (sequence change)"
	}
	return "Next available task in festival"
}

// ProgressStats contains progress information for the festival
type ProgressStats struct {
	TotalTasks      int     `json:"total_tasks"`
	CompletedTasks  int     `json:"completed_tasks"`
	InProgressTasks int     `json:"in_progress_tasks"`
	PendingTasks    int     `json:"pending_tasks"`
	PercentComplete float64 `json:"percent_complete"`
}

// GetProgress returns progress information for the festival
func (s *Selector) GetProgress() (*ProgressStats, error) {
	graph, err := s.resolver.ResolveFestival()
	if err != nil {
		return nil, err
	}

	stats := &ProgressStats{}
	for _, task := range graph.Tasks {
		stats.TotalTasks++
		if task.Status == "complete" {
			stats.CompletedTasks++
		} else if task.Status == "in_progress" {
			stats.InProgressTasks++
		} else {
			stats.PendingTasks++
		}
	}

	if stats.TotalTasks > 0 {
		stats.PercentComplete = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	return stats, nil
}

// updateTaskStatusesFromProgress updates all task statuses in the graph
// by querying the progress tracking system (YAML source of truth)
func (s *Selector) updateTaskStatusesFromProgress(ctx context.Context, graph *deps.Graph) error {
	if ctx == nil {
		ctx = context.Background()
	}
	// Create progress manager
	mgr, err := progress.NewManager(ctx, s.festivalPath)
	if err != nil {
		return err
	}

	// Update each task's status from YAML (or markdown fallback)
	for _, task := range graph.Tasks {
		// ResolveTaskStatus checks YAML first, falls back to markdown
		status := progress.ResolveTaskStatus(mgr.Store(), s.festivalPath, task.Path)

		// Map progress status to deps status
		switch status {
		case progress.StatusCompleted:
			task.Status = "complete" // GetReadyTasks() will skip this
		case progress.StatusInProgress:
			task.Status = "in_progress"
		case progress.StatusBlocked:
			task.Status = "blocked"
		default:
			task.Status = "pending"
		}
	}

	return nil
}

// isNumberedDir checks if directory name starts with a number
func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// getPhaseType reads the phase type from PHASE_GOAL.md frontmatter
func (s *Selector) getPhaseType(phasePath string) (frontmatter.PhaseType, error) {
	goalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
	content, err := os.ReadFile(goalPath)
	if err != nil {
		return "", err
	}

	fm, _, err := frontmatter.Parse(content)
	if err != nil {
		return "", err
	}

	if fm == nil || fm.PhaseType == "" {
		return frontmatter.PhaseTypeImplementation, nil
	}
	return fm.PhaseType, nil
}

// isPlanningPhaseType returns true if the phase type is planning or research
func isPlanningPhaseType(phaseType frontmatter.PhaseType) bool {
	return phaseType == frontmatter.PhaseTypePlanning ||
		phaseType == frontmatter.PhaseTypeResearch
}

// findNextPlanning returns planning guidance for planning/research phases
func (s *Selector) findNextPlanning(_ context.Context, location *LocationInfo, phaseType frontmatter.PhaseType) (*NextTaskResult, error) {
	goalPath := filepath.Join(location.PhasePath, "PHASE_GOAL.md")

	// Parse planning objectives from PHASE_GOAL.md
	objectives, err := parsePlanningObjectives(goalPath)
	if err != nil {
		// Fall back to showing basic info if parsing fails
		objectives = []*PlanningObjective{}
	}

	// Calculate progress
	total := len(objectives)
	resolved := 0
	for _, obj := range objectives {
		if obj.Resolved {
			resolved++
		}
	}

	percentage := 0.0
	if total > 0 {
		percentage = float64(resolved) / float64(total) * 100
	}

	planningResult := &PlanningPhaseResult{
		PhaseName: filepath.Base(location.PhasePath),
		PhasePath: location.PhasePath,
		PhaseType: string(phaseType),
		Objectives: objectives,
		Progress: &PlanningProgress{
			TotalObjectives:    total,
			ResolvedObjectives: resolved,
			Percentage:         percentage,
		},
		GraduationReady: total > 0 && resolved == total,
	}

	reason := "Planning phase - review objectives and explore"
	if planningResult.GraduationReady {
		reason = "Planning complete! Run 'fest graduate' to generate implementation phase"
	}

	return &NextTaskResult{
		Planning: planningResult,
		Reason:   reason,
		Location: location,
	}, nil
}

// parsePlanningObjectives extracts planning objectives from PHASE_GOAL.md
// Looks for markdown checkboxes in sections like "Questions to Answer", "Decisions to Make", "Artifacts to Produce"
func parsePlanningObjectives(goalPath string) ([]*PlanningObjective, error) {
	file, err := os.Open(goalPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var objectives []*PlanningObjective
	currentCategory := ""

	// Regex patterns for detecting sections and checkboxes
	sectionRegex := regexp.MustCompile(`^###?\s*(Questions?|Decisions?|Artifacts?|Objectives?)`)
	checkboxRegex := regexp.MustCompile(`^[-*]\s*\[([ xX])\]\s*(.+)`)

	scanner := bufio.NewScanner(file)
	inPlanningSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check for Planning Objectives header
		if strings.Contains(strings.ToLower(line), "planning objectives") ||
			strings.Contains(strings.ToLower(line), "objectives to achieve") {
			inPlanningSection = true
			continue
		}

		// Check for section headers
		if matches := sectionRegex.FindStringSubmatch(line); len(matches) > 0 {
			inPlanningSection = true
			sectionName := strings.ToLower(matches[1])
			if strings.HasPrefix(sectionName, "question") {
				currentCategory = "question"
			} else if strings.HasPrefix(sectionName, "decision") {
				currentCategory = "decision"
			} else if strings.HasPrefix(sectionName, "artifact") {
				currentCategory = "artifact"
			} else {
				currentCategory = "objective"
			}
			continue
		}

		// Exit planning section on major headers
		if strings.HasPrefix(line, "## ") && !strings.Contains(strings.ToLower(line), "planning") {
			if !strings.Contains(strings.ToLower(line), "question") &&
				!strings.Contains(strings.ToLower(line), "decision") &&
				!strings.Contains(strings.ToLower(line), "artifact") &&
				!strings.Contains(strings.ToLower(line), "objective") {
				inPlanningSection = false
			}
		}

		// Parse checkboxes if we're in a planning section
		if inPlanningSection {
			if matches := checkboxRegex.FindStringSubmatch(line); len(matches) > 0 {
				resolved := matches[1] != " "
				text := strings.TrimSpace(matches[2])
				category := currentCategory
				if category == "" {
					category = "objective"
				}
				objectives = append(objectives, &PlanningObjective{
					Category: category,
					Text:     text,
					Resolved: resolved,
				})
			}
		}
	}

	return objectives, scanner.Err()
}
