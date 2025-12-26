package index

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
)

// IndexWriter generates festival indices from filesystem structure
type IndexWriter struct {
	festivalRoot string
	index        *FestivalIndex
}

// NewIndexWriter creates a new index writer
func NewIndexWriter(festivalRoot string) *IndexWriter {
	return &IndexWriter{
		festivalRoot: festivalRoot,
	}
}

// Generate generates the complete festival index
func (w *IndexWriter) Generate() (*FestivalIndex, error) {
	// Determine festival ID from directory name
	festivalID := filepath.Base(w.festivalRoot)

	w.index = NewFestivalIndex(festivalID)

	// Scan phases
	entries, err := os.ReadDir(w.festivalRoot)
	if err != nil {
		return nil, err
	}

	var phaseEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && festival.IsPhase(entry.Name()) {
			phaseEntries = append(phaseEntries, entry)
		}
	}

	// Sort phases by number
	sort.Slice(phaseEntries, func(i, j int) bool {
		numI := festival.ParsePhaseNumber(phaseEntries[i].Name())
		numJ := festival.ParsePhaseNumber(phaseEntries[j].Name())
		return numI < numJ
	})

	// Process each phase
	for _, entry := range phaseEntries {
		phase, err := w.scanPhase(entry.Name())
		if err != nil {
			continue // Skip invalid phases
		}
		w.index.AddPhase(*phase)
	}

	return w.index, nil
}

// scanPhase scans a phase directory
func (w *IndexWriter) scanPhase(phaseName string) (*PhaseIndex, error) {
	phasePath := filepath.Join(w.festivalRoot, phaseName)

	phase := &PhaseIndex{
		PhaseID:   phaseName,
		Path:      phaseName,
		Sequences: []SequenceIndex{},
	}

	// Check for goal file
	goalFile := "PHASE_GOAL.md"
	if _, err := os.Stat(filepath.Join(phasePath, goalFile)); err == nil {
		phase.GoalFile = goalFile
	}

	// Scan sequences
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, err
	}

	var seqEntries []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() && festival.IsSequence(entry.Name()) {
			seqEntries = append(seqEntries, entry)
		}
	}

	// Sort sequences by number
	sort.Slice(seqEntries, func(i, j int) bool {
		numI := festival.ParseSequenceNumber(seqEntries[i].Name())
		numJ := festival.ParseSequenceNumber(seqEntries[j].Name())
		return numI < numJ
	})

	// Process each sequence
	for _, entry := range seqEntries {
		seq, err := w.scanSequence(phaseName, entry.Name())
		if err != nil {
			continue // Skip invalid sequences
		}
		phase.AddSequence(*seq)
	}

	return phase, nil
}

// scanSequence scans a sequence directory
func (w *IndexWriter) scanSequence(phaseName, seqName string) (*SequenceIndex, error) {
	seqPath := filepath.Join(w.festivalRoot, phaseName, seqName)

	seq := &SequenceIndex{
		SequenceID:   seqName,
		Path:         filepath.Join(phaseName, seqName),
		Tasks:        []TaskIndex{},
		ManagedGates: []string{},
	}

	// Check for goal file
	goalFile := "SEQUENCE_GOAL.md"
	if _, err := os.Stat(filepath.Join(seqPath, goalFile)); err == nil {
		seq.GoalFile = goalFile
	}

	// Scan tasks
	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, err
	}

	var taskEntries []os.DirEntry
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Skip goal files
			if entry.Name() == "SEQUENCE_GOAL.md" || entry.Name() == "PHASE_GOAL.md" {
				continue
			}
			taskEntries = append(taskEntries, entry)
		}
	}

	// Sort tasks by number
	sort.Slice(taskEntries, func(i, j int) bool {
		numI := festival.ParseTaskNumber(taskEntries[i].Name())
		numJ := festival.ParseTaskNumber(taskEntries[j].Name())
		return numI < numJ
	})

	// Process each task
	for _, entry := range taskEntries {
		task := w.scanTask(phaseName, seqName, entry.Name())
		seq.AddTask(task)

		// Track managed gates
		if task.Managed && task.GateID != "" {
			seq.ManagedGates = append(seq.ManagedGates, task.GateID)
		}
	}

	return seq, nil
}

// scanTask scans a task file
func (w *IndexWriter) scanTask(phaseName, seqName, taskName string) TaskIndex {
	taskPath := filepath.Join(w.festivalRoot, phaseName, seqName, taskName)

	task := TaskIndex{
		TaskID:  taskName,
		Path:    filepath.Join(phaseName, seqName, taskName),
		Managed: false,
	}

	// Check if managed
	if gates.IsManaged(taskPath) {
		task.Managed = true
		task.GateID = gates.GetGateID(taskPath)
	}

	return task
}

// WriteIndex generates and writes the index to .festival/index.json
func (w *IndexWriter) WriteIndex() error {
	index, err := w.Generate()
	if err != nil {
		return err
	}

	// Determine output path
	dotFestival := filepath.Join(w.festivalRoot, ".festival")
	if err := os.MkdirAll(dotFestival, 0755); err != nil {
		return err
	}

	indexPath := filepath.Join(dotFestival, IndexFileName)
	return index.Save(indexPath)
}
