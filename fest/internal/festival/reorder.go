package festival

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
)

// ReorderPhase moves a phase from one position to another within a festival
func (r *Renumberer) ReorderPhase(ctx context.Context, festivalDir string, from, to int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	phases, err := r.parser.ParsePhases(ctx, festivalDir)
	if err != nil {
		return fmt.Errorf("failed to parse phases: %w", err)
	}

	return r.reorderElements(phases, from, to, festivalDir, PhaseType)
}

// ReorderSequence moves a sequence from one position to another within a phase
func (r *Renumberer) ReorderSequence(ctx context.Context, phaseDir string, from, to int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	sequences, err := r.parser.ParseSequences(ctx, phaseDir)
	if err != nil {
		return fmt.Errorf("failed to parse sequences: %w", err)
	}

	return r.reorderElements(sequences, from, to, phaseDir, SequenceType)
}

// ReorderTask moves a task from one position to another within a sequence
func (r *Renumberer) ReorderTask(ctx context.Context, sequenceDir string, from, to int) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	tasks, err := r.parser.ParseTasks(ctx, sequenceDir)
	if err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	// Handle parallel tasks - group them by number
	taskGroups := make(map[int][]FestivalElement)
	for _, task := range tasks {
		taskGroups[task.Number] = append(taskGroups[task.Number], task)
	}

	return r.reorderTasksWithParallel(taskGroups, tasks, from, to, sequenceDir)
}

// reorderElements is the core reordering algorithm for phases and sequences
func (r *Renumberer) reorderElements(elements []FestivalElement, from, to int, dir string, elemType ElementType) error {
	if len(elements) == 0 {
		return fmt.Errorf("no elements found in %s", dir)
	}

	// Validate positions
	if from == to {
		if !r.options.Quiet {
			fmt.Println("No changes needed - source and destination are the same.")
		}
		return nil
	}

	// Find element at 'from' position
	var fromElement *FestivalElement
	var fromIndex int
	for i, elem := range elements {
		if elem.Number == from {
			fromElement = &elements[i]
			fromIndex = i
			break
		}
	}
	if fromElement == nil {
		return fmt.Errorf("element at position %d not found", from)
	}

	// Validate 'to' position - must be within valid range
	minNum := elements[0].Number
	maxNum := elements[len(elements)-1].Number
	if to < minNum || to > maxNum {
		return fmt.Errorf("destination position %d is out of range [%d, %d]", to, minNum, maxNum)
	}

	// Build change plan
	r.changes = []Change{}

	// Use a temporary name for the element being moved to avoid collisions
	tmpName := "_tmp_reorder_" + fromElement.Name
	tmpPath := filepath.Join(dir, tmpName)

	// Step 1: Move source to temporary location
	r.changes = append(r.changes, Change{
		Type:    ChangeRename,
		OldPath: fromElement.Path,
		NewPath: tmpPath,
		Element: *fromElement,
	})

	// Step 2: Shift elements between from and to
	if from < to {
		// Moving down: shift elements between (from, to] up by 1 position
		for i := fromIndex + 1; i < len(elements); i++ {
			elem := elements[i]
			if elem.Number > to {
				break
			}
			newNumber := elem.Number - 1
			newName := BuildElementName(newNumber, elem.Name, elemType)
			if elemType == TaskType {
				newName += ".md"
			}
			newPath := filepath.Join(dir, newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: elem.Path,
				NewPath: newPath,
				Element: elem,
			})
		}
	} else {
		// Moving up: shift elements between [to, from) down by 1 position
		// Process in reverse order to avoid conflicts
		for i := fromIndex - 1; i >= 0; i-- {
			elem := elements[i]
			if elem.Number < to {
				break
			}
			newNumber := elem.Number + 1
			newName := BuildElementName(newNumber, elem.Name, elemType)
			if elemType == TaskType {
				newName += ".md"
			}
			newPath := filepath.Join(dir, newName)

			r.changes = append(r.changes, Change{
				Type:    ChangeRename,
				OldPath: elem.Path,
				NewPath: newPath,
				Element: elem,
			})
		}
	}

	// Step 3: Move source from temp to final destination
	finalName := BuildElementName(to, fromElement.Name, elemType)
	if elemType == TaskType {
		finalName += ".md"
	}
	finalPath := filepath.Join(dir, finalName)

	// Create a modified element for the final move
	movedElement := *fromElement
	movedElement.Number = to

	r.changes = append(r.changes, Change{
		Type:    ChangeRename,
		OldPath: tmpPath,
		NewPath: finalPath,
		Element: movedElement,
	})

	// Sort changes for proper execution order
	// When moving up (to < from): process lower numbers first (ascending)
	// When moving down (to > from): process higher numbers first (descending)
	r.sortChangesForReorder(from, to)

	return r.executeChanges()
}

// reorderTasksWithParallel handles task reordering with parallel task support
func (r *Renumberer) reorderTasksWithParallel(taskGroups map[int][]FestivalElement, allTasks []FestivalElement, from, to int, dir string) error {
	if len(allTasks) == 0 {
		return fmt.Errorf("no tasks found in %s", dir)
	}

	// Validate positions
	if from == to {
		if !r.options.Quiet {
			fmt.Println("No changes needed - source and destination are the same.")
		}
		return nil
	}

	// Check if 'from' position exists
	fromTasks, exists := taskGroups[from]
	if !exists {
		return fmt.Errorf("task at position %d not found", from)
	}

	// Get unique numbers in sorted order
	var numbers []int
	for num := range taskGroups {
		numbers = append(numbers, num)
	}
	sort.Ints(numbers)

	// Validate 'to' position
	minNum := numbers[0]
	maxNum := numbers[len(numbers)-1]
	if to < minNum || to > maxNum {
		return fmt.Errorf("destination position %d is out of range [%d, %d]", to, minNum, maxNum)
	}

	// Build change plan
	r.changes = []Change{}

	// Move all parallel tasks at 'from' to temporary locations
	tmpPaths := make(map[string]string)
	for _, task := range fromTasks {
		tmpName := "_tmp_reorder_" + task.Name + ".md"
		tmpPath := filepath.Join(dir, tmpName)
		tmpPaths[task.Path] = tmpPath

		r.changes = append(r.changes, Change{
			Type:    ChangeRename,
			OldPath: task.Path,
			NewPath: tmpPath,
			Element: task,
		})
	}

	// Shift elements
	if from < to {
		// Moving down: shift elements between (from, to] up by 1
		for _, num := range numbers {
			if num <= from || num > to {
				continue
			}
			for _, task := range taskGroups[num] {
				newNumber := task.Number - 1
				newName := BuildElementName(newNumber, task.Name, TaskType) + ".md"
				newPath := filepath.Join(dir, newName)

				r.changes = append(r.changes, Change{
					Type:    ChangeRename,
					OldPath: task.Path,
					NewPath: newPath,
					Element: task,
				})
			}
		}
	} else {
		// Moving up: shift elements between [to, from) down by 1
		for i := len(numbers) - 1; i >= 0; i-- {
			num := numbers[i]
			if num >= from || num < to {
				continue
			}
			for _, task := range taskGroups[num] {
				newNumber := task.Number + 1
				newName := BuildElementName(newNumber, task.Name, TaskType) + ".md"
				newPath := filepath.Join(dir, newName)

				r.changes = append(r.changes, Change{
					Type:    ChangeRename,
					OldPath: task.Path,
					NewPath: newPath,
					Element: task,
				})
			}
		}
	}

	// Move from temp to final destination
	for _, task := range fromTasks {
		tmpPath := tmpPaths[task.Path]
		finalName := BuildElementName(to, task.Name, TaskType) + ".md"
		finalPath := filepath.Join(dir, finalName)

		movedTask := task
		movedTask.Number = to

		r.changes = append(r.changes, Change{
			Type:    ChangeRename,
			OldPath: tmpPath,
			NewPath: finalPath,
			Element: movedTask,
		})
	}

	// Sort changes for proper execution order
	r.sortChangesForReorder(from, to)

	return r.executeChanges()
}

// sortChangesForReorder orders changes to avoid conflicts during execution
func (r *Renumberer) sortChangesForReorder(from, to int) {
	// Separate changes into groups:
	// 1. Initial move to temp (must happen first)
	// 2. Shifts (order depends on direction)
	// 3. Final move from temp (must happen last)

	var toTemp, shifts, fromTemp []Change

	for _, change := range r.changes {
		oldBase := filepath.Base(change.OldPath)
		newBase := filepath.Base(change.NewPath)

		if hasPrefix(newBase, "_tmp_reorder_") {
			toTemp = append(toTemp, change)
		} else if hasPrefix(oldBase, "_tmp_reorder_") {
			fromTemp = append(fromTemp, change)
		} else {
			shifts = append(shifts, change)
		}
	}

	// Sort shifts based on direction
	if from < to {
		// Moving down: process lower numbers first (ascending)
		sort.Slice(shifts, func(i, j int) bool {
			return shifts[i].Element.Number < shifts[j].Element.Number
		})
	} else {
		// Moving up: process higher numbers first (descending)
		sort.Slice(shifts, func(i, j int) bool {
			return shifts[i].Element.Number > shifts[j].Element.Number
		})
	}

	// Reconstruct changes in proper order
	r.changes = append(toTemp, shifts...)
	r.changes = append(r.changes, fromTemp...)
}

// hasPrefix checks if a string has a given prefix
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
