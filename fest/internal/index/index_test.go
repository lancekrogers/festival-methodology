package index

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFestivalIndex(t *testing.T) {
	idx := NewFestivalIndex("test-festival")

	if idx.FestivalID != "test-festival" {
		t.Errorf("FestivalID = %q, want %q", idx.FestivalID, "test-festival")
	}
	if idx.FestSpec != CurrentSpecVersion {
		t.Errorf("FestSpec = %d, want %d", idx.FestSpec, CurrentSpecVersion)
	}
	if len(idx.Phases) != 0 {
		t.Errorf("Phases should be empty, got %d", len(idx.Phases))
	}
}

func TestFestivalIndexSaveLoad(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "index-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	idx := NewFestivalIndex("save-test")
	idx.AddPhase(PhaseIndex{
		PhaseID: "001_DESIGN",
		Path:    "001_DESIGN",
		Sequences: []SequenceIndex{
			{
				SequenceID: "01_requirements",
				Path:       "001_DESIGN/01_requirements",
				Tasks: []TaskIndex{
					{TaskID: "01_gather.md", Path: "001_DESIGN/01_requirements/01_gather.md"},
				},
			},
		},
	})

	indexPath := filepath.Join(tmpDir, IndexFileName)
	if err := idx.Save(indexPath); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	loaded, err := LoadIndex(indexPath)
	if err != nil {
		t.Fatalf("LoadIndex error: %v", err)
	}

	if loaded.FestivalID != "save-test" {
		t.Errorf("FestivalID = %q, want %q", loaded.FestivalID, "save-test")
	}
	if len(loaded.Phases) != 1 {
		t.Errorf("Phases count = %d, want 1", len(loaded.Phases))
	}
	if loaded.Phases[0].PhaseID != "001_DESIGN" {
		t.Errorf("PhaseID = %q, want %q", loaded.Phases[0].PhaseID, "001_DESIGN")
	}
}

func TestFestivalIndexGetPhase(t *testing.T) {
	idx := NewFestivalIndex("test")
	idx.AddPhase(PhaseIndex{PhaseID: "phase1"})
	idx.AddPhase(PhaseIndex{PhaseID: "phase2"})

	phase := idx.GetPhase("phase1")
	if phase == nil {
		t.Fatal("GetPhase returned nil for existing phase")
	}
	if phase.PhaseID != "phase1" {
		t.Errorf("PhaseID = %q, want %q", phase.PhaseID, "phase1")
	}

	notFound := idx.GetPhase("nonexistent")
	if notFound != nil {
		t.Error("GetPhase should return nil for nonexistent phase")
	}
}

func TestPhaseIndexGetSequence(t *testing.T) {
	phase := &PhaseIndex{PhaseID: "test-phase"}
	phase.AddSequence(SequenceIndex{SequenceID: "seq1"})
	phase.AddSequence(SequenceIndex{SequenceID: "seq2"})

	seq := phase.GetSequence("seq1")
	if seq == nil {
		t.Fatal("GetSequence returned nil for existing sequence")
	}
	if seq.SequenceID != "seq1" {
		t.Errorf("SequenceID = %q, want %q", seq.SequenceID, "seq1")
	}

	notFound := phase.GetSequence("nonexistent")
	if notFound != nil {
		t.Error("GetSequence should return nil for nonexistent sequence")
	}
}

func TestSequenceIndexGetTask(t *testing.T) {
	seq := &SequenceIndex{SequenceID: "test-seq"}
	seq.AddTask(TaskIndex{TaskID: "task1.md"})
	seq.AddTask(TaskIndex{TaskID: "task2.md"})

	task := seq.GetTask("task1.md")
	if task == nil {
		t.Fatal("GetTask returned nil for existing task")
	}
	if task.TaskID != "task1.md" {
		t.Errorf("TaskID = %q, want %q", task.TaskID, "task1.md")
	}

	notFound := seq.GetTask("nonexistent.md")
	if notFound != nil {
		t.Error("GetTask should return nil for nonexistent task")
	}
}

func TestIndexSummary(t *testing.T) {
	idx := NewFestivalIndex("test")

	// Add a complex structure
	phase1 := PhaseIndex{PhaseID: "phase1"}
	seq1 := SequenceIndex{SequenceID: "seq1"}
	seq1.AddTask(TaskIndex{TaskID: "task1.md"})
	seq1.AddTask(TaskIndex{TaskID: "task2.md", Managed: true, GateID: "gate1"})
	phase1.AddSequence(seq1)

	seq2 := SequenceIndex{SequenceID: "seq2"}
	seq2.AddTask(TaskIndex{TaskID: "task3.md", Managed: true, GateID: "gate2"})
	phase1.AddSequence(seq2)

	idx.AddPhase(phase1)

	phase2 := PhaseIndex{PhaseID: "phase2"}
	seq3 := SequenceIndex{SequenceID: "seq3"}
	seq3.AddTask(TaskIndex{TaskID: "task4.md"})
	phase2.AddSequence(seq3)
	idx.AddPhase(phase2)

	summary := idx.Summary()

	if summary.PhaseCount != 2 {
		t.Errorf("PhaseCount = %d, want 2", summary.PhaseCount)
	}
	if summary.SequenceCount != 3 {
		t.Errorf("SequenceCount = %d, want 3", summary.SequenceCount)
	}
	if summary.TaskCount != 4 {
		t.Errorf("TaskCount = %d, want 4", summary.TaskCount)
	}
	if summary.ManagedCount != 2 {
		t.Errorf("ManagedCount = %d, want 2", summary.ManagedCount)
	}
}

func TestIndexWriter(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "my-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create phase
	phaseDir := filepath.Join(festivalDir, "001_DESIGN")
	os.MkdirAll(phaseDir, 0755)
	os.WriteFile(filepath.Join(phaseDir, "PHASE_GOAL.md"), []byte("# Phase Goal"), 0644)

	// Create sequence
	seqDir := filepath.Join(phaseDir, "01_requirements")
	os.MkdirAll(seqDir, 0755)
	os.WriteFile(filepath.Join(seqDir, "SEQUENCE_GOAL.md"), []byte("# Sequence Goal"), 0644)

	// Create tasks
	os.WriteFile(filepath.Join(seqDir, "01_gather.md"), []byte("# Task 1"), 0644)
	os.WriteFile(filepath.Join(seqDir, "02_analyze.md"), []byte("# Task 2"), 0644)

	// Generate index
	writer := NewIndexWriter(festivalDir)
	idx, err := writer.Generate()
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	if idx.FestivalID != "my-festival" {
		t.Errorf("FestivalID = %q, want %q", idx.FestivalID, "my-festival")
	}
	if len(idx.Phases) != 1 {
		t.Fatalf("Phases count = %d, want 1", len(idx.Phases))
	}

	phase := idx.Phases[0]
	if phase.PhaseID != "001_DESIGN" {
		t.Errorf("PhaseID = %q, want %q", phase.PhaseID, "001_DESIGN")
	}
	if phase.GoalFile != "PHASE_GOAL.md" {
		t.Errorf("GoalFile = %q, want %q", phase.GoalFile, "PHASE_GOAL.md")
	}
	if len(phase.Sequences) != 1 {
		t.Fatalf("Sequences count = %d, want 1", len(phase.Sequences))
	}

	seq := phase.Sequences[0]
	if seq.SequenceID != "01_requirements" {
		t.Errorf("SequenceID = %q, want %q", seq.SequenceID, "01_requirements")
	}
	if seq.GoalFile != "SEQUENCE_GOAL.md" {
		t.Errorf("GoalFile = %q, want %q", seq.GoalFile, "SEQUENCE_GOAL.md")
	}
	if len(seq.Tasks) != 2 {
		t.Fatalf("Tasks count = %d, want 2", len(seq.Tasks))
	}

	// Tasks should be sorted by number
	if seq.Tasks[0].TaskID != "01_gather.md" {
		t.Errorf("First task = %q, want %q", seq.Tasks[0].TaskID, "01_gather.md")
	}
	if seq.Tasks[1].TaskID != "02_analyze.md" {
		t.Errorf("Second task = %q, want %q", seq.Tasks[1].TaskID, "02_analyze.md")
	}
}

func TestIndexWriterWriteIndex(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "test-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create minimal structure
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	os.MkdirAll(phaseDir, 0755)

	seqDir := filepath.Join(phaseDir, "01_init")
	os.MkdirAll(seqDir, 0755)
	os.WriteFile(filepath.Join(seqDir, "01_start.md"), []byte("# Start"), 0644)

	// Write index
	writer := NewIndexWriter(festivalDir)
	if err := writer.WriteIndex(); err != nil {
		t.Fatalf("WriteIndex error: %v", err)
	}

	// Verify index file was created
	indexPath := filepath.Join(festivalDir, ".festival", IndexFileName)
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Error("Index file was not created")
	}

	// Load and verify
	idx, err := LoadIndex(indexPath)
	if err != nil {
		t.Fatalf("LoadIndex error: %v", err)
	}
	if idx.FestivalID != "test-festival" {
		t.Errorf("FestivalID = %q, want %q", idx.FestivalID, "test-festival")
	}
}

func TestIndexValidator(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "valid-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create structure
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	os.MkdirAll(phaseDir, 0755)

	seqDir := filepath.Join(phaseDir, "01_init")
	os.MkdirAll(seqDir, 0755)
	os.WriteFile(filepath.Join(seqDir, "01_task.md"), []byte("# Task"), 0644)

	// Create matching index
	idx := NewFestivalIndex("valid-festival")
	phase := PhaseIndex{
		PhaseID: "001_PLAN",
		Path:    "001_PLAN",
	}
	seq := SequenceIndex{
		SequenceID: "01_init",
		Path:       "001_PLAN/01_init",
	}
	seq.AddTask(TaskIndex{
		TaskID: "01_task.md",
		Path:   "001_PLAN/01_init/01_task.md",
	})
	phase.AddSequence(seq)
	idx.AddPhase(phase)

	// Validate
	validator := NewIndexValidator(festivalDir, idx)
	result := validator.Validate()

	if !result.Valid {
		t.Errorf("Validation should pass for matching structure, errors: %v", result.Errors)
	}
}

func TestIndexValidatorMissingFiles(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "missing-files-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create partial structure (missing task file)
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	os.MkdirAll(phaseDir, 0755)

	seqDir := filepath.Join(phaseDir, "01_init")
	os.MkdirAll(seqDir, 0755)
	// Note: NOT creating the task file

	// Create index that references the missing file
	idx := NewFestivalIndex("missing-files-festival")
	phase := PhaseIndex{
		PhaseID: "001_PLAN",
		Path:    "001_PLAN",
	}
	seq := SequenceIndex{
		SequenceID: "01_init",
		Path:       "001_PLAN/01_init",
	}
	seq.AddTask(TaskIndex{
		TaskID: "01_task.md",
		Path:   "001_PLAN/01_init/01_task.md",
	})
	phase.AddSequence(seq)
	idx.AddPhase(phase)

	// Validate
	validator := NewIndexValidator(festivalDir, idx)
	result := validator.Validate()

	if result.Valid {
		t.Error("Validation should fail for missing files")
	}
	if len(result.MissingInFS) != 1 {
		t.Errorf("MissingInFS count = %d, want 1", len(result.MissingInFS))
	}
}

func TestIndexValidatorExtraFiles(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "extra-files-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create structure with extra file
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	os.MkdirAll(phaseDir, 0755)

	seqDir := filepath.Join(phaseDir, "01_init")
	os.MkdirAll(seqDir, 0755)
	os.WriteFile(filepath.Join(seqDir, "01_task.md"), []byte("# Task"), 0644)
	os.WriteFile(filepath.Join(seqDir, "02_extra.md"), []byte("# Extra"), 0644) // Not in index

	// Create index without the extra file
	idx := NewFestivalIndex("extra-files-festival")
	phase := PhaseIndex{
		PhaseID: "001_PLAN",
		Path:    "001_PLAN",
	}
	seq := SequenceIndex{
		SequenceID: "01_init",
		Path:       "001_PLAN/01_init",
	}
	seq.AddTask(TaskIndex{
		TaskID: "01_task.md",
		Path:   "001_PLAN/01_init/01_task.md",
	})
	phase.AddSequence(seq)
	idx.AddPhase(phase)

	// Validate
	validator := NewIndexValidator(festivalDir, idx)
	result := validator.Validate()

	// Extra files don't make validation fail, but are reported
	if len(result.ExtraInFS) != 1 {
		t.Errorf("ExtraInFS count = %d, want 1", len(result.ExtraInFS))
	}
}

func TestValidateFromFile(t *testing.T) {
	// Create a mock festival structure
	tmpDir, err := os.MkdirTemp("", "festival-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	festivalDir := filepath.Join(tmpDir, "file-validation-festival")
	os.MkdirAll(festivalDir, 0755)

	// Create structure
	phaseDir := filepath.Join(festivalDir, "001_PLAN")
	os.MkdirAll(phaseDir, 0755)

	seqDir := filepath.Join(phaseDir, "01_init")
	os.MkdirAll(seqDir, 0755)
	os.WriteFile(filepath.Join(seqDir, "01_task.md"), []byte("# Task"), 0644)

	// Generate and save index
	writer := NewIndexWriter(festivalDir)
	if err := writer.WriteIndex(); err != nil {
		t.Fatalf("WriteIndex error: %v", err)
	}

	// Validate from file
	indexPath := filepath.Join(festivalDir, ".festival", IndexFileName)
	result, err := ValidateFromFile(festivalDir, indexPath)
	if err != nil {
		t.Fatalf("ValidateFromFile error: %v", err)
	}

	if !result.Valid {
		t.Errorf("Validation should pass, errors: %v", result.Errors)
	}
}

func TestLoadIndexInvalidJSON(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "index-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	indexPath := filepath.Join(tmpDir, IndexFileName)
	os.WriteFile(indexPath, []byte("invalid json"), 0644)

	_, err = LoadIndex(indexPath)
	if err == nil {
		t.Error("LoadIndex should fail for invalid JSON")
	}
}

func TestLoadIndexNotFound(t *testing.T) {
	_, err := LoadIndex("/nonexistent/path/index.json")
	if err == nil {
		t.Error("LoadIndex should fail for missing file")
	}
}

func TestIndexGeneratedAtTimestamp(t *testing.T) {
	before := time.Now().UTC()
	idx := NewFestivalIndex("timestamp-test")
	after := time.Now().UTC()

	if idx.GeneratedAt.Before(before) || idx.GeneratedAt.After(after) {
		t.Errorf("GeneratedAt should be between %v and %v, got %v", before, after, idx.GeneratedAt)
	}
}
