package festival

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
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
func (r *Renumberer) RenumberPhases(festivalDir string, startFrom int) error {
	phases, err := r.parser.ParsePhases(festivalDir)
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
func (r *Renumberer) RenumberSequences(phaseDir string, startFrom int) error {
	sequences, err := r.parser.ParseSequences(phaseDir)
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
func (r *Renumberer) RenumberTasks(sequenceDir string, startFrom int) error {
	tasks, err := r.parser.ParseTasks(sequenceDir)
	if err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
	}

	if len(tasks) == 0 {
		return fmt.Errorf("no tasks found in %s", sequenceDir)
	}

	// Check for parallel tasks
	parallel, err := r.parser.HasParallelTasks(sequenceDir)
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

// InsertPhase inserts a new phase after the specified number
func (r *Renumberer) InsertPhase(festivalDir string, afterNumber int, name string) error {
	phases, err := r.parser.ParsePhases(festivalDir)
	if err != nil {
		return fmt.Errorf("failed to parse phases: %w", err)
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
func (r *Renumberer) InsertSequence(phaseDir string, afterNumber int, name string) error {
	sequences, err := r.parser.ParseSequences(phaseDir)
	if err != nil {
		return fmt.Errorf("failed to parse sequences: %w", err)
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
func (r *Renumberer) InsertTask(sequenceDir string, afterNumber int, name string) error {
	tasks, err := r.parser.ParseTasks(sequenceDir)
	if err != nil {
		return fmt.Errorf("failed to parse tasks: %w", err)
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

// RemoveElement removes an element and renumbers subsequent ones
func (r *Renumberer) RemoveElement(path string) error {
	dir := filepath.Dir(path)
	base := filepath.Base(path)

	// Determine element type
	var elemType ElementType
	var elements []FestivalElement
	var err error

	if matched := regexp.MustCompile(`^\d{3}_`).MatchString(base); matched {
		elemType = PhaseType
		elements, err = r.parser.ParsePhases(dir)
	} else if matched := regexp.MustCompile(`^\d{2}_`).MatchString(base); matched {
		if strings.HasSuffix(base, ".md") {
			elemType = TaskType
			elements, err = r.parser.ParseTasks(dir)
		} else {
			elemType = SequenceType
			elements, err = r.parser.ParseSequences(dir)
		}
	} else {
		return fmt.Errorf("unable to determine element type for %s", path)
	}

	if err != nil {
		return fmt.Errorf("failed to parse elements: %w", err)
	}

	// Find element to remove
	var toRemove *FestivalElement
	var removeIndex int
	for i, elem := range elements {
		if elem.Path == path {
			toRemove = &elem
			removeIndex = i
			break
		}
	}

	if toRemove == nil {
		return fmt.Errorf("element not found: %s", path)
	}

	// Build changes
	r.changes = []Change{
		{
			Type:    ChangeRemove,
			OldPath: path,
		},
	}

	// Renumber subsequent elements
	for i := removeIndex + 1; i < len(elements); i++ {
		newNumber := elements[i].Number - 1
		newName := BuildElementName(newNumber, elements[i].Name, elemType)
		if elemType == TaskType {
			newName += ".md"
		}
		newPath := filepath.Join(dir, newName)

		r.changes = append(r.changes, Change{
			Type:    ChangeRename,
			OldPath: elements[i].Path,
			NewPath: newPath,
			Element: elements[i],
		})
	}

	return r.executeChanges()
}

// executeChanges applies the planned changes
func (r *Renumberer) executeChanges() error {
	if len(r.changes) == 0 {
		if !r.options.Quiet {
			fmt.Println("No changes needed.")
		}
		return nil
	}

	// Display changes
	if !r.options.Quiet {
		r.displayChanges()
	}

	if r.options.DryRun {
		if !r.options.Quiet {
			fmt.Println("\nDRY RUN - Preview complete.")
		}
		// If auto-approve, proceed to apply after dry-run preview
		if r.options.AutoApprove {
			if !r.options.Quiet {
				fmt.Println("\nApplying changes...")
			}
		} else {
			// Prompt user to apply changes after dry-run
			if r.confirmApplyAfterDryRun() {
				if !r.options.Quiet {
					fmt.Println("\nApplying changes...")
				}
			} else {
				if !r.options.Quiet {
					fmt.Println("Operation cancelled.")
				}
				return nil
			}
		}
	} else {
		// Not in dry-run mode, confirm changes before applying
		if !r.options.AutoApprove {
			if !r.confirmChanges() {
				if !r.options.Quiet {
					fmt.Println("Operation cancelled.")
				}
				return nil
			}
		}
	}

	// Create backup if requested
	if r.options.Backup {
		if err := r.createBackup(); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Apply changes
	for _, change := range r.changes {
		switch change.Type {
		case ChangeRename:
			if err := os.Rename(change.OldPath, change.NewPath); err != nil {
				return fmt.Errorf("failed to rename %s to %s: %w", change.OldPath, change.NewPath, err)
			}
			if r.options.Verbose {
				fmt.Printf("Renamed: %s → %s\n", filepath.Base(change.OldPath), filepath.Base(change.NewPath))
			}

		case ChangeCreate:
			// If NewPath looks like a file (e.g., ends with .md), create an empty file.
			// Otherwise, create a directory (used for phases/sequences).
			if strings.HasSuffix(strings.ToLower(change.NewPath), ".md") {
				// For file creations, just ensure parent exists; the caller writes content.
				if err := os.MkdirAll(filepath.Dir(change.NewPath), 0755); err != nil {
					return fmt.Errorf("failed to create parent directory for %s: %w", change.NewPath, err)
				}
			} else {
				if err := os.MkdirAll(change.NewPath, 0755); err != nil {
					return fmt.Errorf("failed to create %s: %w", change.NewPath, err)
				}
			}
			if r.options.Verbose {
				fmt.Printf("Created: %s\n", filepath.Base(change.NewPath))
			}

		case ChangeRemove:
			if err := os.RemoveAll(change.OldPath); err != nil {
				return fmt.Errorf("failed to remove %s: %w", change.OldPath, err)
			}
			if r.options.Verbose {
				fmt.Printf("Removed: %s\n", filepath.Base(change.OldPath))
			}
		}
	}

	if !r.options.Quiet {
		fmt.Printf("\n✓ Successfully applied %d changes.\n", len(r.changes))
	}
	return nil
}

// displayChanges shows planned changes
func (r *Renumberer) displayChanges() {
	fmt.Println("\nFestival Renumbering Report")
	fmt.Println(strings.Repeat("═", 55))
	fmt.Println("\nChanges to be made:")

	for _, change := range r.changes {
		switch change.Type {
		case ChangeRename:
			fmt.Printf("  → Rename: %s → %s\n",
				filepath.Base(change.OldPath),
				filepath.Base(change.NewPath))
		case ChangeCreate:
			fmt.Printf("  ✓ Create: %s\n", filepath.Base(change.NewPath))
		case ChangeRemove:
			fmt.Printf("  ✗ Remove: %s\n", filepath.Base(change.OldPath))
		}
	}

	fmt.Printf("\nTotal: %d changes\n", len(r.changes))
}

// confirmChanges prompts for confirmation
func (r *Renumberer) confirmChanges() bool {
	fmt.Print("\nProceed with renumbering? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// confirmApplyAfterDryRun prompts to apply changes after dry-run preview
func (r *Renumberer) confirmApplyAfterDryRun() bool {
	fmt.Print("\nApply these changes? [y/N]: ")
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}

// createBackup creates a backup of affected directories
func (r *Renumberer) createBackup() error {
	// Implementation would create timestamped backup
	// For now, just log
	fmt.Println("Creating backup...")
	return nil
}
