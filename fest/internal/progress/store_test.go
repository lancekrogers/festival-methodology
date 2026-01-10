package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_LoadNew(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if store.Data() == nil {
		t.Fatal("Data() returned nil")
	}

	if store.Data().Tasks == nil {
		t.Error("Tasks map should not be nil")
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create and save
	store1 := NewStore(tmpDir)
	if err := store1.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	now := time.Now().UTC()
	store1.SetTask(&TaskProgress{
		TaskID:    "01_test.md",
		Status:    StatusInProgress,
		Progress:  50,
		StartedAt: &now,
	})

	if err := store1.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	progressPath := filepath.Join(tmpDir, ProgressDir, ProgressFileName)
	if _, err := os.Stat(progressPath); os.IsNotExist(err) {
		t.Fatalf("Progress file not created at %s", progressPath)
	}

	// Load in new store
	store2 := NewStore(tmpDir)
	if err := store2.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	task, found := store2.GetTask("01_test.md")
	if !found {
		t.Fatal("Task not found after load")
	}

	if task.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusInProgress)
	}

	if task.Progress != 50 {
		t.Errorf("Progress = %d, want 50", task.Progress)
	}
}

func TestStore_GetTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Non-existent task
	_, found := store.GetTask("nonexistent")
	if found {
		t.Error("GetTask() should return false for non-existent task")
	}

	// Add and retrieve
	store.SetTask(&TaskProgress{TaskID: "01_test.md", Status: StatusPending})
	task, found := store.GetTask("01_test.md")
	if !found {
		t.Error("GetTask() should return true for existing task")
	}
	if task.TaskID != "01_test.md" {
		t.Errorf("TaskID = %q, want %q", task.TaskID, "01_test.md")
	}
}

func TestStore_AllTasks(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	store.SetTask(&TaskProgress{TaskID: "01_test.md"})
	store.SetTask(&TaskProgress{TaskID: "02_test.md"})
	store.SetTask(&TaskProgress{TaskID: "03_test.md"})

	tasks := store.AllTasks()
	if len(tasks) != 3 {
		t.Errorf("AllTasks() returned %d tasks, want 3", len(tasks))
	}
}
