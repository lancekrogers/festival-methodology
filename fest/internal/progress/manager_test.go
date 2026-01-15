package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_UpdateProgress(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test initial progress update
	err = mgr.UpdateProgress(ctx, "01_test.md", 25)
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
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Update to 100%
	err = mgr.UpdateProgress(ctx, "01_test.md", 100)
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
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Test invalid progress values
	err = mgr.UpdateProgress(ctx, "01_test.md", -1)
	if err == nil {
		t.Error("UpdateProgress() should error for negative value")
	}

	err = mgr.UpdateProgress(ctx, "01_test.md", 101)
	if err == nil {
		t.Error("UpdateProgress() should error for value > 100")
	}
}

func TestManager_MarkComplete(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.MarkComplete(ctx, "01_test.md")
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
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.ReportBlocker(ctx, "01_test.md", "Waiting on API spec")
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
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.ReportBlocker(ctx, "01_test.md", "")
	if err == nil {
		t.Error("ReportBlocker() should error for empty message")
	}
}

func TestManager_ClearBlocker(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Report blocker
	mgr.ReportBlocker(ctx, "01_test.md", "Blocker")

	// Clear it
	err = mgr.ClearBlocker(ctx, "01_test.md")
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
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	err = mgr.MarkInProgress(ctx, "01_test.md")
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

func TestManager_MarkComplete_UsesFileModTime(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a task file with a known modification time
	taskPath := filepath.Join(tmpDir, "01_test.md")
	if err := os.WriteFile(taskPath, []byte("# Test Task\n"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Set modification time to 10 minutes ago
	oldTime := time.Now().Add(-10 * time.Minute)
	if err := os.Chtimes(taskPath, oldTime, oldTime); err != nil {
		t.Fatalf("Failed to set file mod time: %v", err)
	}

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Mark complete directly without prior MarkInProgress
	err = mgr.MarkComplete(ctx, "01_test.md")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")

	// TimeSpentMinutes should be approximately 10 minutes (not 0)
	if task.TimeSpentMinutes < 9 || task.TimeSpentMinutes > 11 {
		t.Errorf("TimeSpentMinutes = %d, want approximately 10", task.TimeSpentMinutes)
	}

	// StartedAt should be close to the file modification time
	if task.StartedAt == nil {
		t.Fatal("StartedAt should be set")
	}

	timeDiff := task.StartedAt.Sub(oldTime)
	if timeDiff < -time.Second || timeDiff > time.Second {
		t.Errorf("StartedAt = %v, expected close to file mod time %v", task.StartedAt, oldTime)
	}
}

func TestManager_MarkComplete_PreservesExistingStartTime(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	mgr, err := NewManager(ctx, tmpDir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// First mark in progress (sets StartedAt to current time)
	err = mgr.MarkInProgress(ctx, "01_test.md")
	if err != nil {
		t.Fatalf("MarkInProgress() error = %v", err)
	}

	task, _ := mgr.GetTaskProgress("01_test.md")
	originalStartedAt := *task.StartedAt

	// Wait a moment to ensure time difference
	time.Sleep(10 * time.Millisecond)

	// Now mark complete
	err = mgr.MarkComplete(ctx, "01_test.md")
	if err != nil {
		t.Fatalf("MarkComplete() error = %v", err)
	}

	task, _ = mgr.GetTaskProgress("01_test.md")

	// StartedAt should NOT have changed
	if !task.StartedAt.Equal(originalStartedAt) {
		t.Errorf("StartedAt changed from %v to %v, should be preserved", originalStartedAt, task.StartedAt)
	}
}

func TestGetTaskFileModTime_FallbackToCurrentTime(t *testing.T) {
	// Test that getTaskFileModTime returns current time for non-existent file
	now := time.Now().UTC()
	result := getTaskFileModTime("/nonexistent/path", "nonexistent.md")

	// Should be within a second of now
	if result.Before(now.Add(-time.Second)) || result.After(now.Add(time.Second)) {
		t.Errorf("getTaskFileModTime for nonexistent file = %v, want close to %v", result, now)
	}
}
