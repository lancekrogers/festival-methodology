package festival

import (
	"context"
	"fmt"
	"path/filepath"
)

// RenumberOptions configures renumbering behavior
type RenumberOptions struct {
	DryRun    bool
	Backup    bool
	StartFrom int
	Verbose   bool
	// Quiet suppresses all printouts (no report, no success lines)
	Quiet bool
	// AutoApprove skips confirmation prompts and applies changes immediately
	AutoApprove bool
}

// Renumberer handles renumbering operations
type Renumberer struct {
	parser  *Parser
	options RenumberOptions
	changes []Change
}

// Change represents a renumbering change
type Change struct {
	Type    ChangeType
	OldPath string
	NewPath string
	Element FestivalElement
}

// ChangeType represents the type of change
type ChangeType int

const (
	ChangeRename ChangeType = iota
	ChangeCreate
	ChangeRemove
)

func (c ChangeType) String() string {
	switch c {
	case ChangeRename:
		return "Rename"
	case ChangeCreate:
		return "Create"
	case ChangeRemove:
		return "Remove"
	default:
		return "Unknown"
	}
}

// NewRenumberer creates a new renumberer
func NewRenumberer(options RenumberOptions) *Renumberer {
	return &Renumberer{
		parser:  NewParser(),
		options: options,
		changes: []Change{},
	}
}

// RenumberPhases renumbers all phases starting from a given number
func (r *Renumberer) RenumberPhases(ctx context.Context, festivalDir string, startFrom int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	phases, err := r.parser.ParsePhases(ctx, festivalDir)
	if err != nil {
		return fmt.Errorf("failed to parse phases: %w", err)
	}

	if len(phases) == 0 {
		return fmt.Errorf("no phases found in %s", festivalDir)
	}

	// Build renumbering plan
	r.changes = []Change{}
	newNumber := startFrom

	for _, phase := range phases {
		if phase.Number != newNumber {
			newName := BuildElementName(newNumber, phase.Name, PhaseType)
			newPath := filepath.Join(filepath.Dir(phase.Path), newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: phase.Path,
				NewPath: newPath,
				Element: phase,
			})
		}
		newNumber++
	}

	// Execute changes
	return r.executeChanges()
}

// RenumberSequences renumbers sequences within a phase
func (r *Renumberer) RenumberSequences(ctx context.Context, phaseDir string, startFrom int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	sequences, err := r.parser.ParseSequences(ctx, phaseDir)
	if err != nil {
		return fmt.Errorf("failed to parse sequences: %w", err)
	}

	if len(sequences) == 0 {
		return fmt.Errorf("no sequences found in %s", phaseDir)
	}

	// Build renumbering plan
	r.changes = []Change{}
	newNumber := startFrom

	for _, seq := range sequences {
		if seq.Number != newNumber {
			newName := BuildElementName(newNumber, seq.Name, SequenceType)
			newPath := filepath.Join(filepath.Dir(seq.Path), newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: seq.Path,
				NewPath: newPath,
				Element: seq,
			})
		}
		newNumber++
	}

	// Execute changes
	return r.executeChanges()
}

// RenumberTasks renumbers tasks within a sequence
func (r *Renumberer) RenumberTasks(ctx context.Context, sequenceDir string, startFrom int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	tasks, err := r.parser.ParseTasks(ctx, sequenceDir)
	if err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found in %s", sequenceDir)
	}

	// Check for parallel tasks
	parallel, err := r.parser.HasParallelTasks(ctx, sequenceDir)
	if err != nil {
		return fmt.Errorf("failed to check parallel tasks: %w", err)
	}

	if len(parallel) > 0 && !r.options.DryRun {
		fmt.Println("Warning: Parallel tasks detected. They will be preserved with the same number.")
	}

	// Build renumbering plan
	r.changes = []Change{}
	newNumber := startFrom
	processedNumbers := make(map[int]bool)

	for _, task := range tasks {
		// Skip if we've already processed this number (parallel tasks)
		if processedNumbers[task.Number] {
			continue
		}
		processedNumbers[task.Number] = true

		// Get all tasks with this number
		tasksWithNumber := []FestivalElement{task}
		for _, t := range tasks {
			if t.Number == task.Number && t.Path != task.Path {
				tasksWithNumber = append(tasksWithNumber, t)
			}
		}

		// Renumber all tasks with this number
		for _, t := range tasksWithNumber {
			if t.Number != newNumber {
				newName := BuildElementName(newNumber, t.Name+".md", TaskType)
				newPath := filepath.Join(filepath.Dir(t.Path), newName)

				r.changes = append(r.changes, Change{
					Type:    ChangeRename,
					OldPath: t.Path,
					NewPath: newPath,
					Element: t,
				})
			}
		}

		newNumber++
	}

	// Execute changes
	return r.executeChanges()
}
