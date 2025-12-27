package progress

import (
	"testing"
)

func TestManager_UpdateProgress(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test initial progress update
	err = mgr.UpdateProgress("01_test.md", 25)
	if err != nil {
		t.Fatalf("UpdateProgress() error = %v", err)
	}

	task, found := mgr.GetTaskProgress("01_test.md")
	if !found {
		t.Fatal("Task not found after update")
	}

	if task.Progress != 25 {
		t.Errorf("Progress = %d, want 25", task.Progress)
	}

	if task.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusInProgress)
	}

	if task.StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}

func TestManager_UpdateProgress_Complete(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Update to 100%
	err = mgr.UpdateProgress("01_test.md", 100)
	if err != nil {
		t.Fatalf("UpdateProgress() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	if task.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", task.Status, StatusCompleted)
	}

	if task.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestManager_UpdateProgress_Invalid(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test invalid progress values
	err = mgr.UpdateProgress("01_test.md", -1)
	if err == nil {
		t.Error("UpdateProgress() should error for negative value")
	}

	err = mgr.UpdateProgress("01_test.md", 101)
	if err == nil {
		t.Error("UpdateProgress() should error for value > 100")
	}
}

func TestManager_MarkComplete(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.MarkComplete("01_test.md")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	if task.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", task.Status, StatusCompleted)
	}

	if task.Progress != 100 {
		t.Errorf("Progress = %d, want 100", task.Progress)
	}

	if task.CompletedAt == nil {
		t.Error("CompletedAt should be set")
	}
}

func TestManager_ReportBlocker(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.ReportBlocker("01_test.md", "Waiting on API spec")
	if err != nil {
		t.Fatalf("ReportBlocker() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	if task.Status != StatusBlocked {
		t.Errorf("Status = %q, want %q", task.Status, StatusBlocked)
	}

	if task.BlockerMessage != "Waiting on API spec" {
		t.Errorf("BlockerMessage = %q, want %q", task.BlockerMessage, "Waiting on API spec")
	}

	if task.BlockedAt == nil {
		t.Error("BlockedAt should be set")
	}
}

func TestManager_ReportBlocker_EmptyMessage(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.ReportBlocker("01_test.md", "")
	if err == nil {
		t.Error("ReportBlocker() should error for empty message")
	}
}

func TestManager_ClearBlocker(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Report blocker
	mgr.ReportBlocker("01_test.md", "Blocker")

	// Clear it
	err = mgr.ClearBlocker("01_test.md")
	if err != nil {
		t.Fatalf("ClearBlocker() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	if task.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusInProgress)
	}

	if task.BlockerMessage != "" {
		t.Errorf("BlockerMessage should be empty, got %q", task.BlockerMessage)
	}
}

func TestManager_MarkInProgress(t *testing.T) {
	tmpDir := t.TempDir()

	mgr, err := NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.MarkInProgress("01_test.md")
	if err != nil {
		t.Fatalf("MarkInProgress() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	if task.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusInProgress)
	}

	if task.StartedAt == nil {
		t.Error("StartedAt should be set")
	}
}
