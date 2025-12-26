package festival

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
)

// TestCreateSequence_DirectoryCreation tests that creating a sequence
// results in a directory being created (not a file)
func TestCreateSequence_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up phase directory structure
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatalf("failed to create phase dir: %v", err)
	}

	// Use renumberer directly (what create sequence uses internally)
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertSequence(phaseDir, 0, "requirements")
	if err != nil {
		t.Fatalf("InsertSequence failed: %v", err)
	}

	// Verify directory was created (not a file)
	seqPath := filepath.Join(phaseDir, "01_requirements")
	info, err := os.Stat(seqPath)
	if err != nil {
		t.Fatalf("expected sequence directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected 01_requirements to be a directory, got file")
	}
}

// TestCreateTask_FileCreation tests that creating a task
// creates the parent directory but doesn't write file content
func TestCreateTask_FileCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up sequence directory structure
	seqDir := filepath.Join(tmpDir, "01_requirements")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("failed to create sequence dir: %v", err)
	}

	// Use renumberer directly (what create task uses internally)
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertTask(seqDir, 0, "define_requirements")
	if err != nil {
		t.Fatalf("InsertTask failed: %v", err)
	}

	// The renumberer ensures parent exists but doesn't create .md file content
	// (the create command is responsible for writing template content)
	taskPath := filepath.Join(seqDir, "01_define_requirements.md")

	// Check parent directory exists
	if _, err := os.Stat(seqDir); err != nil {
		t.Fatalf("expected parent directory to exist: %v", err)
	}

	// Task file may or may not exist (renumberer behavior)
	// but if it exists, it should be empty (no template content)
	if info, err := os.Stat(taskPath); err == nil {
		if info.IsDir() {
			t.Error("expected task to be a file, not a directory")
		}
		// If file exists, verify it's empty (no content written by renumberer)
		content, _ := os.ReadFile(taskPath)
		if len(content) > 0 {
			t.Errorf("renumberer should not write content to task file, got: %q", content)
		}
	}
}

// TestCreateSequence_NoIsDirectoryError tests that creating a sequence
// doesn't cause "is a directory" errors when the directory already exists
func TestCreateSequence_NoIsDirectoryError(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up phase directory with existing sequence
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	existingSeq := filepath.Join(phaseDir, "01_existing")
	if err := os.MkdirAll(existingSeq, 0755); err != nil {
		t.Fatalf("failed to create existing sequence: %v", err)
	}

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert new sequence at beginning (should shift existing)
	err := ren.InsertSequence(phaseDir, 0, "new_sequence")
	if err != nil {
		t.Fatalf("InsertSequence failed with error: %v", err)
	}

	// Verify both sequences exist as directories
	newSeq := filepath.Join(phaseDir, "01_new_sequence")
	shiftedSeq := filepath.Join(phaseDir, "02_existing")

	if info, err := os.Stat(newSeq); err != nil || !info.IsDir() {
		t.Error("expected 01_new_sequence directory")
	}
	if info, err := os.Stat(shiftedSeq); err != nil || !info.IsDir() {
		t.Error("expected 02_existing directory (renumbered)")
	}
}

// TestCreateTask_NoIsDirectoryError tests that creating a task
// with existing tasks doesn't cause errors
func TestCreateTask_NoIsDirectoryError(t *testing.T) {
	tmpDir := t.TempDir()

	// Set up sequence directory with existing task
	seqDir := tmpDir
	if err := os.WriteFile(filepath.Join(seqDir, "01_existing.md"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create existing task: %v", err)
	}

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert new task at beginning (should shift existing)
	err := ren.InsertTask(seqDir, 0, "new_task")
	if err != nil {
		t.Fatalf("InsertTask failed with error: %v", err)
	}

	// Verify renumbering happened correctly
	shiftedTask := filepath.Join(seqDir, "02_existing.md")
	if _, err := os.Stat(shiftedTask); err != nil {
		t.Error("expected 02_existing.md (renumbered from 01)")
	}

	// Verify content was preserved in renamed file
	content, err := os.ReadFile(shiftedTask)
	if err != nil {
		t.Fatalf("failed to read shifted task: %v", err)
	}
	if string(content) != "content" {
		t.Errorf("shifted task content = %q, want %q", content, "content")
	}
}

// TestCreatePhase_DirectoryCreation tests that creating a phase
// results in a directory being created
func TestCreatePhase_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	err := ren.InsertPhase(tmpDir, 0, "PLANNING")
	if err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Verify directory was created
	phasePath := filepath.Join(tmpDir, "001_PLANNING")
	info, err := os.Stat(phasePath)
	if err != nil {
		t.Fatalf("expected phase directory to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected 001_PLANNING to be a directory, got file")
	}
}

// TestCreate_ChainedOperations tests multiple create operations in sequence
func TestCreate_ChainedOperations(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festival structure progressively
	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Create phases
	if err := ren.InsertPhase(tmpDir, 0, "PLANNING"); err != nil {
		t.Fatalf("InsertPhase PLANNING failed: %v", err)
	}
	if err := ren.InsertPhase(tmpDir, 1, "IMPLEMENT"); err != nil {
		t.Fatalf("InsertPhase IMPLEMENT failed: %v", err)
	}

	// Create sequences in first phase
	phaseDir := filepath.Join(tmpDir, "001_PLANNING")
	if err := ren.InsertSequence(phaseDir, 0, "requirements"); err != nil {
		t.Fatalf("InsertSequence requirements failed: %v", err)
	}
	if err := ren.InsertSequence(phaseDir, 1, "design"); err != nil {
		t.Fatalf("InsertSequence design failed: %v", err)
	}

	// Create tasks in first sequence
	seqDir := filepath.Join(phaseDir, "01_requirements")
	if err := ren.InsertTask(seqDir, 0, "gather_info"); err != nil {
		t.Fatalf("InsertTask gather_info failed: %v", err)
	}
	if err := ren.InsertTask(seqDir, 1, "analyze"); err != nil {
		t.Fatalf("InsertTask analyze failed: %v", err)
	}

	// Verify complete structure
	expectedDirs := []string{
		"001_PLANNING",
		"002_IMPLEMENT",
		"001_PLANNING/01_requirements",
		"001_PLANNING/02_design",
	}

	for _, dir := range expectedDirs {
		path := filepath.Join(tmpDir, dir)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("expected %s to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", dir)
		}
	}
}

// TestCreateOptions_DefaultAfterZero verifies that after=0 creates at position 1
func TestCreateOptions_DefaultAfterZero(t *testing.T) {
	tmpDir := t.TempDir()

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Create with after=0 (default)
	if err := ren.InsertPhase(tmpDir, 0, "FIRST"); err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Should create at position 1 (001_)
	if _, err := os.Stat(filepath.Join(tmpDir, "001_FIRST")); err != nil {
		t.Error("expected 001_FIRST to exist")
	}
}

// TestCreateOptions_InsertInMiddle verifies inserting in the middle of existing elements
func TestCreateOptions_InsertInMiddle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create initial phases
	os.MkdirAll(filepath.Join(tmpDir, "001_FIRST"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_THIRD"), 0755)

	ren := festival.NewRenumberer(festival.RenumberOptions{
		AutoApprove: true,
		Quiet:       true,
	})

	// Insert after position 1
	if err := ren.InsertPhase(tmpDir, 1, "SECOND"); err != nil {
		t.Fatalf("InsertPhase failed: %v", err)
	}

	// Verify structure
	expected := []string{"001_FIRST", "002_SECOND", "003_THIRD"}
	for _, name := range expected {
		if _, err := os.Stat(filepath.Join(tmpDir, name)); err != nil {
			t.Errorf("expected %s to exist", name)
		}
	}
}
