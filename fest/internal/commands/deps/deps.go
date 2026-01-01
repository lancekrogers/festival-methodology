// Package deps provides the fest deps command for dependency visualization.
package deps

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/deps"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

var (
	jsonOutput   bool
	showAll      bool
	criticalPath bool
)

// NewDepsCommand creates the deps command
func NewDepsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps [task]",
		Short: "Show task dependencies",
		Long: `Display dependency information for tasks in the festival.

Without arguments, shows the dependency graph for the current sequence.
With a task name, shows dependencies for that specific task.

Examples:
  fest deps                    # Show all deps in current sequence
  fest deps 02_implement       # Show deps for specific task
  fest deps --all              # Show all deps in festival
  fest deps --json             # Output as JSON
  fest deps --critical-path    # Show critical path through the DAG`,
		RunE: runDeps,
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&showAll, "all", false, "show all dependencies in festival")
	cmd.Flags().BoolVar(&criticalPath, "critical-path", false, "show the critical path")

	return cmd
}

func runDeps(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := tpl.FindFestivalRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	resolver := deps.NewResolver(festivalPath)

	var graph *deps.Graph
	if showAll {
		graph, err = resolver.ResolveFestival()
	} else {
		// Try to resolve just the current sequence
		seqPath := findSequencePath(cwd, festivalPath)
		if seqPath != "" {
			graph, err = resolver.ResolveSequence(seqPath)
		} else {
			graph, err = resolver.ResolveFestival()
		}
	}

	if err != nil {
		return errors.Wrap(err, "resolving dependencies")
	}

	// If a specific task was requested
	if len(args) > 0 {
		taskName := args[0]
		return showTaskDeps(graph, taskName)
	}

	// Show critical path if requested
	if criticalPath {
		return showCriticalPath(graph)
	}

	// Show full graph
	return showGraph(graph)
}

func showGraph(graph *deps.Graph) error {
	if jsonOutput {
		data, err := json.MarshalIndent(graph, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Text output
	sorted, err := graph.TopologicalSort()
	if err != nil {
		fmt.Printf("Warning: %v\n\n", err)
	}

	fmt.Println("=== DEPENDENCY GRAPH ===")
	fmt.Println()

	// Show parallel groups
	groups := graph.GetParallelGroups()
	for i, group := range groups {
		fmt.Printf("Level %d (can run in parallel):\n", i)
		for _, task := range group {
			deps := graph.GetDependencies(task.ID)
			depNames := make([]string, len(deps))
			for j, d := range deps {
				depNames[j] = d.Name
			}
			if len(depNames) > 0 {
				fmt.Printf("  - %s <- [%s]\n", task.Name, strings.Join(depNames, ", "))
			} else {
				fmt.Printf("  - %s\n", task.Name)
			}
		}
		fmt.Println()
	}

	if sorted != nil {
		fmt.Println("Execution order:")
		for i, task := range sorted {
			fmt.Printf("  %d. %s\n", i+1, task.Name)
		}
	}

	return nil
}

func showTaskDeps(graph *deps.Graph, taskName string) error {
	// Find the task
	var task *deps.Task
	for _, t := range graph.Tasks {
		if t.Name == taskName || strings.TrimSuffix(t.Name, ".md") == taskName ||
			filepath.Base(t.Path) == taskName || filepath.Base(t.Path) == taskName+".md" {
			task = t
			break
		}
	}

	if task == nil {
		return errors.NotFound("task not found").
			WithField("task", taskName)
	}

	if jsonOutput {
		output := struct {
			Task       *deps.Task   `json:"task"`
			DependsOn  []*deps.Task `json:"depends_on"`
			DependedBy []*deps.Task `json:"depended_by"`
		}{
			Task:       task,
			DependsOn:  graph.GetDependencies(task.ID),
			DependedBy: graph.GetDependents(task.ID),
		}
		data, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Text output
	fmt.Printf("=== DEPENDENCIES: %s ===\n\n", task.Name)

	fmt.Println("Task Info:")
	fmt.Printf("  Number: %d\n", task.Number)
	fmt.Printf("  Parallel Group: %d\n", task.ParallelGroup)
	if task.AutonomyLevel != "" {
		fmt.Printf("  Autonomy: %s\n", task.AutonomyLevel)
	}
	fmt.Println()

	deps := graph.GetDependencies(task.ID)
	if len(deps) > 0 {
		fmt.Println("Depends On (must complete first):")
		for _, dep := range deps {
			fmt.Printf("  - %s\n", dep.Name)
		}
		fmt.Println()
	} else {
		fmt.Println("No dependencies (can start immediately)")
		fmt.Println()
	}

	dependents := graph.GetDependents(task.ID)
	if len(dependents) > 0 {
		fmt.Println("Depended By (waiting on this task):")
		for _, dep := range dependents {
			fmt.Printf("  - %s\n", dep.Name)
		}
		fmt.Println()
	} else {
		fmt.Println("No dependents (nothing waiting on this task)")
		fmt.Println()
	}

	// Show the dependency chain
	fmt.Println("Dependency Tree:")
	printDepTree(graph, task, "  ", make(map[string]bool))

	return nil
}

func printDepTree(graph *deps.Graph, task *deps.Task, indent string, visited map[string]bool) {
	if visited[task.ID] {
		fmt.Printf("%s└─ %s (cycle)\n", indent, task.Name)
		return
	}
	visited[task.ID] = true

	deps := graph.GetDependencies(task.ID)
	if len(deps) == 0 {
		fmt.Printf("%s└─ %s (root)\n", indent, task.Name)
		return
	}

	fmt.Printf("%s└─ %s\n", indent, task.Name)
	for _, dep := range deps {
		printDepTree(graph, dep, indent+"  ", visited)
	}
}

func showCriticalPath(graph *deps.Graph) error {
	path := graph.CriticalPath()

	if jsonOutput {
		data, err := json.MarshalIndent(path, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	fmt.Println("=== CRITICAL PATH ===")
	fmt.Println()
	fmt.Println("The critical path is the longest chain of dependencies.")
	fmt.Println("Optimizing these tasks reduces overall completion time.")
	fmt.Println()

	if len(path) == 0 {
		fmt.Println("No critical path (no dependencies or cycle detected)")
		return nil
	}

	for i, task := range path {
		if i < len(path)-1 {
			fmt.Printf("  %d. %s\n     ↓\n", i+1, task.Name)
		} else {
			fmt.Printf("  %d. %s\n", i+1, task.Name)
		}
	}

	fmt.Printf("\nCritical path length: %d tasks\n", len(path))

	return nil
}

func findSequencePath(cwd, festivalPath string) string {
	rel, err := filepath.Rel(festivalPath, cwd)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) >= 2 {
		// We're in a sequence directory
		return filepath.Join(festivalPath, parts[0], parts[1])
	}

	return ""
}
