// Package deps provides dependency graph management for festival tasks.
package deps

import "time"

// DependencyType describes the type of dependency
type DependencyType string

const (
	DepImplicit      DependencyType = "implicit"       // From task numbering
	DepExplicit      DependencyType = "explicit"       // From frontmatter
	DepCrossSequence DependencyType = "cross_sequence" // Cross-sequence reference
	DepCrossPhase    DependencyType = "cross_phase"    // Cross-phase reference
)

// Task represents a task node in the dependency graph
type Task struct {
	ID            string     `json:"id"`             // Full path identifier
	Name          string     `json:"name"`           // Task name from filename
	Number        int        `json:"number"`         // Task number prefix
	Path          string     `json:"path"`           // File path
	SequencePath  string     `json:"sequence_path"`  // Parent sequence path
	PhasePath     string     `json:"phase_path"`     // Parent phase path
	ParallelGroup int        `json:"parallel_group"` // Tasks in same group can run in parallel
	Status        string     `json:"status"`         // pending, in_progress, complete
	Dependencies  []string   `json:"dependencies"`   // Explicit dependencies from frontmatter
	SoftDeps      []string   `json:"soft_deps"`      // Soft dependencies (preferred but not required)
	AutonomyLevel string     `json:"autonomy_level"` // high, medium, low
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// Dependency represents an edge in the dependency graph
type Dependency struct {
	From     *Task          `json:"from"`
	To       *Task          `json:"to"`
	Type     DependencyType `json:"type"`
	Required bool           `json:"required"` // true = hard dep, false = soft dep
}

// Graph represents the dependency graph
type Graph struct {
	Tasks    map[string]*Task   `json:"tasks"` // All tasks by ID
	Edges    []*Dependency      `json:"edges"` // All dependencies
	Incoming map[string]int     `json:"-"`     // In-degree for each task (for topological sort)
	Outgoing map[string][]*Task `json:"-"`     // Adjacency list
}

// NewGraph creates a new empty dependency graph
func NewGraph() *Graph {
	return &Graph{
		Tasks:    make(map[string]*Task),
		Edges:    []*Dependency{},
		Incoming: make(map[string]int),
		Outgoing: make(map[string][]*Task),
	}
}

// AddTask adds a task node to the graph
func (g *Graph) AddTask(task *Task) {
	if _, exists := g.Tasks[task.ID]; !exists {
		g.Tasks[task.ID] = task
		g.Incoming[task.ID] = 0
		g.Outgoing[task.ID] = []*Task{}
	}
}

// AddDependency adds a dependency edge to the graph
func (g *Graph) AddDependency(from, to *Task, depType DependencyType, required bool) {
	dep := &Dependency{
		From:     from,
		To:       to,
		Type:     depType,
		Required: required,
	}
	g.Edges = append(g.Edges, dep)
	g.Incoming[to.ID]++
	g.Outgoing[from.ID] = append(g.Outgoing[from.ID], to)
}

// GetTask returns a task by ID
func (g *Graph) GetTask(id string) (*Task, bool) {
	task, ok := g.Tasks[id]
	return task, ok
}

// GetDependencies returns all dependencies for a task
func (g *Graph) GetDependencies(taskID string) []*Task {
	var deps []*Task
	for _, edge := range g.Edges {
		if edge.To.ID == taskID {
			deps = append(deps, edge.From)
		}
	}
	return deps
}

// GetDependents returns all tasks that depend on this task
func (g *Graph) GetDependents(taskID string) []*Task {
	return g.Outgoing[taskID]
}

// ValidationError represents a dependency validation error
type ValidationError struct {
	TaskID   string `json:"task_id"`
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // error, warning
}

// CycleError represents a circular dependency error
type CycleError struct {
	Cycle []string `json:"cycle"`
}

func (e *CycleError) Error() string {
	if len(e.Cycle) == 0 {
		return "circular dependency detected"
	}
	return "circular dependency: " + e.Cycle[0] + " -> ... -> " + e.Cycle[len(e.Cycle)-1]
}
