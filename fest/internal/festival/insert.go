package festival

import (
	"context"
	"path/filepath"
	"sort"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// InsertPhase inserts a new phase after the specified number
func (r *Renumberer) InsertPhase(ctx context.Context, festivalDir string, afterNumber int, name string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("Renumberer.InsertPhase")
	}

	phases, err := r.parser.ParsePhases(ctx, festivalDir)
	if err != nil {
		return errors.Wrap(err, "failed to parse phases").
			WithOp("Renumberer.InsertPhase").
			WithCode(errors.ErrCodeParse)
	}

	// Find insertion point
	insertAt := afterNumber + 1

	// Build changes
	r.changes = []Change{}

	// Create new phase
	newPhaseName := BuildElementName(insertAt, name, PhaseType)
	newPhasePath := filepath.Join(festivalDir, newPhaseName)

	r.changes = append(r.changes, Change{
		Type:    ChangeCreate,
		NewPath: newPhasePath,
	})

	// Renumber subsequent phases
	for _, phase := range phases {
		if phase.Number >= insertAt {
			newNumber := phase.Number + 1
			newName := BuildElementName(newNumber, phase.Name, PhaseType)
			newPath := filepath.Join(filepath.Dir(phase.Path), newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: phase.Path,
				NewPath: newPath,
				Element: phase,
			})
		}
	}

	// Sort changes to rename in reverse order (avoid conflicts)
	sort.Slice(r.changes, func(i, j int) bool {
		if r.changes[i].Type == ChangeCreate {
			return false // Create operations go last
		}
		if r.changes[j].Type == ChangeCreate {
			return true
		}
		// Rename higher numbers first
		return r.changes[i].Element.Number > r.changes[j].Element.Number
	})

	return r.executeChanges()
}

// InsertSequence inserts a new sequence after the specified number
func (r *Renumberer) InsertSequence(ctx context.Context, phaseDir string, afterNumber int, name string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("Renumberer.InsertSequence")
	}

	sequences, err := r.parser.ParseSequences(ctx, phaseDir)
	if err != nil {
		return errors.Wrap(err, "failed to parse sequences").
			WithOp("Renumberer.InsertSequence").
			WithCode(errors.ErrCodeParse)
	}

	// Find insertion point
	insertAt := afterNumber + 1

	// Build changes
	r.changes = []Change{}

	// Create new sequence
	newSeqName := BuildElementName(insertAt, name, SequenceType)
	newSeqPath := filepath.Join(phaseDir, newSeqName)

	r.changes = append(r.changes, Change{
		Type:    ChangeCreate,
		NewPath: newSeqPath,
	})

	// Renumber subsequent sequences
	for _, seq := range sequences {
		if seq.Number >= insertAt {
			newNumber := seq.Number + 1
			newName := BuildElementName(newNumber, seq.Name, SequenceType)
			newPath := filepath.Join(filepath.Dir(seq.Path), newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: seq.Path,
				NewPath: newPath,
				Element: seq,
			})
		}
	}

	// Sort changes to rename in reverse order
	sort.Slice(r.changes, func(i, j int) bool {
		if r.changes[i].Type == ChangeCreate {
			return false
		}
		if r.changes[j].Type == ChangeCreate {
			return true
		}
		return r.changes[i].Element.Number > r.changes[j].Element.Number
	})

	return r.executeChanges()
}

// InsertTask inserts a new task after the specified number in a sequence directory
func (r *Renumberer) InsertTask(ctx context.Context, sequenceDir string, afterNumber int, name string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("Renumberer.InsertTask")
	}

	tasks, err := r.parser.ParseTasks(ctx, sequenceDir)
	if err != nil {
		return errors.Wrap(err, "failed to parse tasks").
			WithOp("Renumberer.InsertTask").
			WithCode(errors.ErrCodeParse)
	}

	// Find insertion point
	insertAt := afterNumber + 1

	// Build changes
	r.changes = []Change{}

	// Create new task (as file, not directory)
	newTaskName := BuildElementName(insertAt, name, TaskType) + ".md"
	newTaskPath := filepath.Join(sequenceDir, newTaskName)

	r.changes = append(r.changes, Change{
		Type:    ChangeCreate,
		NewPath: newTaskPath,
	})

	// Renumber subsequent tasks
	for _, task := range tasks {
		if task.Number >= insertAt {
			newNumber := task.Number + 1
			newName := BuildElementName(newNumber, task.Name, TaskType) + ".md"
			newPath := filepath.Join(filepath.Dir(task.Path), newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: task.Path,
				NewPath: newPath,
				Element: task,
			})
		}
	}

	// Sort changes to rename in reverse order
	sort.Slice(r.changes, func(i, j int) bool {
		if r.changes[i].Type == ChangeCreate {
			return false
		}
		if r.changes[j].Type == ChangeCreate {
			return true
		}
		return r.changes[i].Element.Number > r.changes[j].Element.Number
	})

	return r.executeChanges()
}
