package deps

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
)

// Resolver handles dependency resolution for a festival
type Resolver struct {
	festivalPath string
	graph        *Graph
}

// NewResolver creates a new dependency resolver
func NewResolver(festivalPath string) *Resolver {
	return &Resolver{
		festivalPath: festivalPath,
		graph:        NewGraph(),
	}
}

// ResolveFestival builds the complete dependency graph for a festival
func (r *Resolver) ResolveFestival() (*Graph, error) {
	// Walk through all phases and sequences
	phases, err := r.listPhases()
	if err != nil {
		return nil, err
	}

	for _, phasePath := range phases {
		sequences, err := r.listSequences(phasePath)
		if err != nil {
			continue
		}

		for _, seqPath := range sequences {
			tasks, err := r.loadSequenceTasks(seqPath, phasePath)
			if err != nil {
				continue
			}

			// Add tasks to graph
			for _, task := range tasks {
				r.graph.AddTask(task)
			}

			// Add implicit dependencies (from numbering)
			r.addImplicitDependencies(tasks)
		}
	}

	// Add explicit dependencies (from frontmatter)
	for _, task := range r.graph.Tasks {
		r.addExplicitDependencies(task)
	}

	return r.graph, nil
}

// ResolveSequence builds the dependency graph for a single sequence
func (r *Resolver) ResolveSequence(seqPath string) (*Graph, error) {
	phasePath := filepath.Dir(seqPath)
	tasks, err := r.loadSequenceTasks(seqPath, phasePath)
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		r.graph.AddTask(task)
	}

	r.addImplicitDependencies(tasks)

	for _, task := range r.graph.Tasks {
		r.addExplicitDependencies(task)
	}

	return r.graph, nil
}

// listPhases returns all phase directories in the festival
func (r *Resolver) listPhases() ([]string, error) {
	entries, err := os.ReadDir(r.festivalPath)
	if err != nil {
		return nil, err
	}

	var phases []string
	for _, entry := range entries {
		if entry.IsDir() && isNumberedDir(entry.Name()) {
			phases = append(phases, filepath.Join(r.festivalPath, entry.Name()))
		}
	}

	sort.Strings(phases)
	return phases, nil
}

// listSequences returns all sequence directories in a phase
func (r *Resolver) listSequences(phasePath string) ([]string, error) {
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, err
	}

	var sequences []string
	for _, entry := range entries {
		if entry.IsDir() && isNumberedDir(entry.Name()) {
			sequences = append(sequences, filepath.Join(phasePath, entry.Name()))
		}
	}

	sort.Strings(sequences)
	return sequences, nil
}

// loadSequenceTasks loads all tasks in a sequence
func (r *Resolver) loadSequenceTasks(seqPath, phasePath string) ([]*Task, error) {
	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".md") || strings.Contains(strings.ToUpper(name), "GOAL") {
			continue
		}

		num := extractNumber(name)
		if num == 0 {
			continue // Not a numbered task file
		}

		taskPath := filepath.Join(seqPath, name)

		// Skip untracked files (files with tracking: false in frontmatter)
		if !progress.IsTracked(taskPath) {
			continue
		}

		task := &Task{
			ID:            taskPath,
			Name:          strings.TrimSuffix(name, ".md"),
			Number:        num,
			Path:          taskPath,
			SequencePath:  seqPath,
			PhasePath:     phasePath,
			ParallelGroup: num,
			Status:        "pending",
		}

		// Try to parse task file for additional metadata
		r.parseTaskMetadata(task)

		tasks = append(tasks, task)
	}

	// Sort by number
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].Number < tasks[j].Number
	})

	return tasks, nil
}

// parseTaskMetadata reads the task file and extracts frontmatter metadata
func (r *Resolver) parseTaskMetadata(task *Task) {
	content, err := os.ReadFile(task.Path)
	if err != nil {
		return
	}

	text := string(content)

	// Extract dependencies from frontmatter or blockquote header
	// Format: Dependencies: task1, task2
	depsRe := regexp.MustCompile(`(?i)Dependencies[:\s*]+([^\|]+)`)
	if match := depsRe.FindStringSubmatch(text); len(match) > 1 {
		deps := strings.TrimSpace(match[1])
		if deps != "None" && deps != "none" && deps != "" {
			for _, dep := range strings.Split(deps, ",") {
				dep = strings.TrimSpace(dep)
				if dep != "" {
					task.Dependencies = append(task.Dependencies, dep)
				}
			}
		}
	}

	// Extract fest_dependencies from YAML frontmatter
	if strings.HasPrefix(strings.TrimSpace(text), "---") {
		parts := strings.SplitN(text, "---", 3)
		if len(parts) >= 3 {
			frontmatter := parts[1]

			// Extract fest_dependencies
			festDepsRe := regexp.MustCompile(`fest_dependencies:\s*\n((?:\s+-\s+.+\n)+)`)
			if match := festDepsRe.FindStringSubmatch(frontmatter); len(match) > 1 {
				lines := strings.Split(match[1], "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "- ") {
						dep := strings.TrimPrefix(line, "- ")
						task.Dependencies = append(task.Dependencies, strings.TrimSpace(dep))
					}
				}
			}

			// Extract fest_soft_dependencies
			softDepsRe := regexp.MustCompile(`fest_soft_dependencies:\s*\n((?:\s+-\s+.+\n)+)`)
			if match := softDepsRe.FindStringSubmatch(frontmatter); len(match) > 1 {
				lines := strings.Split(match[1], "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if strings.HasPrefix(line, "- ") {
						dep := strings.TrimPrefix(line, "- ")
						task.SoftDeps = append(task.SoftDeps, strings.TrimSpace(dep))
					}
				}
			}

			// Extract fest_parallel_group
			parallelRe := regexp.MustCompile(`fest_parallel_group:\s*(\d+)`)
			if match := parallelRe.FindStringSubmatch(frontmatter); len(match) > 1 {
				if num, err := strconv.Atoi(match[1]); err == nil {
					task.ParallelGroup = num
				}
			}
		}
	}

	// Extract autonomy level
	autonomyRe := regexp.MustCompile(`(?i)Autonomy\s+Level\*{0,2}[:\s*]+(\w+)`)
	if match := autonomyRe.FindStringSubmatch(text); len(match) > 1 {
		task.AutonomyLevel = strings.ToLower(match[1])
	}
}

// addImplicitDependencies adds dependencies based on task numbering
// Tasks with number N depend on all tasks with number N-1
func (r *Resolver) addImplicitDependencies(tasks []*Task) {
	// Group tasks by number
	byNumber := make(map[int][]*Task)
	for _, task := range tasks {
		byNumber[task.Number] = append(byNumber[task.Number], task)
	}

	// Get sorted list of numbers
	var numbers []int
	for num := range byNumber {
		numbers = append(numbers, num)
	}
	sort.Ints(numbers)

	// Add dependencies: tasks at number N depend on tasks at number N-1
	for i := 1; i < len(numbers); i++ {
		prevNum := numbers[i-1]
		currNum := numbers[i]

		for _, currTask := range byNumber[currNum] {
			for _, prevTask := range byNumber[prevNum] {
				r.graph.AddDependency(prevTask, currTask, DepImplicit, true)
			}
		}
	}
}

// addExplicitDependencies adds dependencies declared in frontmatter
func (r *Resolver) addExplicitDependencies(task *Task) {
	for _, depRef := range task.Dependencies {
		depTask := r.resolveTaskReference(task, depRef)
		if depTask != nil {
			depType := DepExplicit
			if depTask.SequencePath != task.SequencePath {
				if depTask.PhasePath != task.PhasePath {
					depType = DepCrossPhase
				} else {
					depType = DepCrossSequence
				}
			}
			r.graph.AddDependency(depTask, task, depType, true)
		}
	}

	for _, depRef := range task.SoftDeps {
		depTask := r.resolveTaskReference(task, depRef)
		if depTask != nil {
			depType := DepExplicit
			if depTask.SequencePath != task.SequencePath {
				if depTask.PhasePath != task.PhasePath {
					depType = DepCrossPhase
				} else {
					depType = DepCrossSequence
				}
			}
			r.graph.AddDependency(depTask, task, depType, false)
		}
	}
}

// resolveTaskReference resolves a task reference to a task in the graph
// Supports formats:
//   - task_name (same sequence)
//   - 01_task_name (same sequence)
//   - ../sequence/task (cross-sequence)
//   - ../../phase/sequence/task (cross-phase)
func (r *Resolver) resolveTaskReference(from *Task, ref string) *Task {
	ref = strings.TrimSpace(ref)

	// Handle relative paths
	if strings.HasPrefix(ref, "..") {
		absPath := filepath.Join(from.SequencePath, ref)
		if !strings.HasSuffix(absPath, ".md") {
			absPath += ".md"
		}
		if task, ok := r.graph.GetTask(absPath); ok {
			return task
		}
		return nil
	}

	// Handle same-sequence references
	// Try exact match first
	for _, task := range r.graph.Tasks {
		if task.SequencePath != from.SequencePath {
			continue
		}
		if task.Name == ref || task.Name == strings.TrimSuffix(ref, ".md") {
			return task
		}
	}

	// Try with sequence path
	fullPath := filepath.Join(from.SequencePath, ref)
	if !strings.HasSuffix(fullPath, ".md") {
		fullPath += ".md"
	}
	if task, ok := r.graph.GetTask(fullPath); ok {
		return task
	}

	return nil
}

// isNumberedDir checks if a directory name starts with a number
func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// extractNumber extracts the leading number from a filename
func extractNumber(filename string) int {
	num := 0
	for _, c := range filename {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	return num
}
