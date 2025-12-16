package festival

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReorderPhase_MoveUp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phases: 001, 002, 003
	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "003_REVIEW"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move phase 3 to position 1
	err := r.ReorderPhase(tmpDir, 3, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify structure: REVIEW should now be 001
	if _, err := os.Stat(filepath.Join(tmpDir, "001_REVIEW")); err != nil {
		t.Error("expected 001_REVIEW to exist (moved from 003)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "002_PLANNING")); err != nil {
		t.Error("expected 002_PLANNING to exist (shifted from 001)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "003_IMPLEMENT")); err != nil {
		t.Error("expected 003_IMPLEMENT to exist (shifted from 002)")
	}
}

func TestReorderPhase_MoveDown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phases: 001, 002, 003
	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "003_REVIEW"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move phase 1 to position 3
	err := r.ReorderPhase(tmpDir, 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify structure: PLANNING should now be 003
	if _, err := os.Stat(filepath.Join(tmpDir, "001_IMPLEMENT")); err != nil {
		t.Error("expected 001_IMPLEMENT to exist (shifted from 002)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "002_REVIEW")); err != nil {
		t.Error("expected 002_REVIEW to exist (shifted from 003)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "003_PLANNING")); err != nil {
		t.Error("expected 003_PLANNING to exist (moved from 001)")
	}
}

func TestReorderPhase_SamePosition(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move phase 1 to position 1 (no-op)
	err := r.ReorderPhase(tmpDir, 1, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify nothing changed
	if _, err := os.Stat(filepath.Join(tmpDir, "001_PLANNING")); err != nil {
		t.Error("expected 001_PLANNING to remain unchanged")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "002_IMPLEMENT")); err != nil {
		t.Error("expected 002_IMPLEMENT to remain unchanged")
	}
}

func TestReorderPhase_InvalidFrom(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	err := r.ReorderPhase(tmpDir, 5, 1)
	if err == nil {
		t.Error("expected error for invalid source position")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' in error, got: %v", err)
	}
}

func TestReorderPhase_InvalidTo(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	err := r.ReorderPhase(tmpDir, 1, 10)
	if err == nil {
		t.Error("expected error for invalid destination position")
	}
	if !strings.Contains(err.Error(), "out of range") {
		t.Errorf("expected 'out of range' in error, got: %v", err)
	}
}

func TestReorderSequence_MoveUp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sequences: 01, 02, 03, 04
	os.MkdirAll(filepath.Join(tmpDir, "01_requirements"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "02_design"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "03_implementation"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "04_tests"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move sequence 4 (tests) to position 1
	err := r.ReorderSequence(tmpDir, 4, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify: tests should be first, others shifted down
	if _, err := os.Stat(filepath.Join(tmpDir, "01_tests")); err != nil {
		t.Error("expected 01_tests to exist (moved from 04)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "02_requirements")); err != nil {
		t.Error("expected 02_requirements to exist (shifted from 01)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "03_design")); err != nil {
		t.Error("expected 03_design to exist (shifted from 02)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "04_implementation")); err != nil {
		t.Error("expected 04_implementation to exist (shifted from 03)")
	}
}

func TestReorderSequence_MoveDown(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "01_tests"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "02_requirements"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "03_implementation"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move sequence 1 (tests) to position 3
	err := r.ReorderSequence(tmpDir, 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify: tests should be last
	if _, err := os.Stat(filepath.Join(tmpDir, "01_requirements")); err != nil {
		t.Error("expected 01_requirements to exist (shifted from 02)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "02_implementation")); err != nil {
		t.Error("expected 02_implementation to exist (shifted from 03)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "03_tests")); err != nil {
		t.Error("expected 03_tests to exist (moved from 01)")
	}
}

func TestReorderTask_MoveUp(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tasks
	os.WriteFile(filepath.Join(tmpDir, "01_setup.md"), []byte("setup"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "02_implement.md"), []byte("implement"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "03_test.md"), []byte("test"), 0644)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move task 3 to position 1
	err := r.ReorderTask(tmpDir, 3, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	if _, err := os.Stat(filepath.Join(tmpDir, "01_test.md")); err != nil {
		t.Error("expected 01_test.md to exist (moved from 03)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "02_setup.md")); err != nil {
		t.Error("expected 02_setup.md to exist (shifted from 01)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "03_implement.md")); err != nil {
		t.Error("expected 03_implement.md to exist (shifted from 02)")
	}
}

func TestReorderTask_MoveDown(t *testing.T) {
	tmpDir := t.TempDir()

	os.WriteFile(filepath.Join(tmpDir, "01_test.md"), []byte("test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "02_setup.md"), []byte("setup"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "03_implement.md"), []byte("implement"), 0644)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move task 1 to position 3
	err := r.ReorderTask(tmpDir, 1, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	if _, err := os.Stat(filepath.Join(tmpDir, "01_setup.md")); err != nil {
		t.Error("expected 01_setup.md to exist (shifted from 02)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "02_implement.md")); err != nil {
		t.Error("expected 02_implement.md to exist (shifted from 03)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "03_test.md")); err != nil {
		t.Error("expected 03_test.md to exist (moved from 01)")
	}
}

func TestReorderTask_ParallelTasks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create tasks including parallel tasks (same number)
	os.WriteFile(filepath.Join(tmpDir, "01_setup.md"), []byte("setup"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "02_task_a.md"), []byte("a"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "02_task_b.md"), []byte("b"), 0644) // parallel
	os.WriteFile(filepath.Join(tmpDir, "03_finish.md"), []byte("finish"), 0644)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move parallel tasks (02) to position 1
	err := r.ReorderTask(tmpDir, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify parallel tasks moved together
	entries, _ := os.ReadDir(tmpDir)
	count01 := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "01_task_") {
			count01++
		}
	}
	if count01 != 2 {
		t.Errorf("expected 2 parallel tasks with prefix 01_, got %d", count01)
	}

	// Verify setup shifted
	if _, err := os.Stat(filepath.Join(tmpDir, "02_setup.md")); err != nil {
		t.Error("expected 02_setup.md to exist (shifted from 01)")
	}
}

func TestReorderPhase_DryRunMode(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)

	r := NewRenumberer(RenumberOptions{
		DryRun: true,
		Quiet:  true,
		// Note: without AutoApprove, it would prompt - but Quiet+DryRun shows preview only
	})

	output := captureOutput(func() {
		// This will show preview but not apply since AutoApprove is false
		_ = r.ReorderPhase(tmpDir, 2, 1)
	})

	// Since Quiet is true, output should be empty even in dry-run
	if output != "" {
		t.Logf("dry-run output: %q", output)
	}

	// Verify nothing changed (no AutoApprove)
	if _, err := os.Stat(filepath.Join(tmpDir, "001_PLANNING")); err != nil {
		t.Error("expected 001_PLANNING to remain (dry-run)")
	}
	if _, err := os.Stat(filepath.Join(tmpDir, "002_IMPLEMENT")); err != nil {
		t.Error("expected 002_IMPLEMENT to remain (dry-run)")
	}
}

func TestReorderPhase_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()

	os.MkdirAll(filepath.Join(tmpDir, "001_PLANNING"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_IMPLEMENT"), 0755)

	r := NewRenumberer(RenumberOptions{
		Verbose:     true,
		AutoApprove: true,
	})

	output := captureOutput(func() {
		_ = r.ReorderPhase(tmpDir, 2, 1)
	})

	// Verbose mode should show rename operations
	if !strings.Contains(output, "Renamed:") {
		t.Errorf("expected verbose output to contain 'Renamed:', got: %q", output)
	}
}

func TestReorderPhase_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	err := r.ReorderPhase(tmpDir, 1, 2)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestReorderSequence_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	err := r.ReorderSequence(tmpDir, 1, 2)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestReorderTask_EmptyDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	err := r.ReorderTask(tmpDir, 1, 2)
	if err == nil {
		t.Error("expected error for empty directory")
	}
}

func TestReorderPhase_MoveMiddle(t *testing.T) {
	tmpDir := t.TempDir()

	// Create phases: 001, 002, 003, 004, 005
	os.MkdirAll(filepath.Join(tmpDir, "001_A"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "002_B"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "003_C"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "004_D"), 0755)
	os.MkdirAll(filepath.Join(tmpDir, "005_E"), 0755)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move phase 2 to position 4
	err := r.ReorderPhase(tmpDir, 2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Expected: A=1, C=2, D=3, B=4, E=5
	expected := map[string]string{
		"001_A": "001_A",
		"002_C": "002_C",
		"003_D": "003_D",
		"004_B": "004_B",
		"005_E": "005_E",
	}

	for want := range expected {
		if _, err := os.Stat(filepath.Join(tmpDir, want)); err != nil {
			t.Errorf("expected %s to exist", want)
		}
	}
}

func TestReorderSequence_PreservesContents(t *testing.T) {
	tmpDir := t.TempDir()

	// Create sequences with content
	seq1 := filepath.Join(tmpDir, "01_first")
	seq2 := filepath.Join(tmpDir, "02_second")
	os.MkdirAll(seq1, 0755)
	os.MkdirAll(seq2, 0755)

	// Add tasks inside sequences
	os.WriteFile(filepath.Join(seq1, "01_task.md"), []byte("first task"), 0644)
	os.WriteFile(filepath.Join(seq2, "01_task.md"), []byte("second task"), 0644)

	r := NewRenumberer(RenumberOptions{
		Quiet:       true,
		AutoApprove: true,
	})

	// Move sequence 2 to position 1
	err := r.ReorderSequence(tmpDir, 2, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify contents preserved
	content, err := os.ReadFile(filepath.Join(tmpDir, "01_second", "01_task.md"))
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if string(content) != "second task" {
		t.Errorf("expected 'second task', got %q", content)
	}

	content, err = os.ReadFile(filepath.Join(tmpDir, "02_first", "01_task.md"))
	if err != nil {
		t.Fatalf("failed to read task: %v", err)
	}
	if string(content) != "first task" {
		t.Errorf("expected 'first task', got %q", content)
	}
}
