package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNormalizeTaskID(t *testing.T) {
	festivalDir := t.TempDir()
	taskRel := filepath.Join("001_PHASE", "01_seq", "01_task.md")
	taskAbs := filepath.Join(festivalDir, taskRel)

	got, err := NormalizeTaskID(festivalDir, taskAbs)
	if err != nil {
		t.Fatalf("NormalizeTaskID(abs) error = %v", err)
	}
	if got != filepath.ToSlash(taskRel) {
		t.Fatalf("NormalizeTaskID(abs) = %q, want %q", got, filepath.ToSlash(taskRel))
	}

	got, err = NormalizeTaskID(festivalDir, taskRel)
	if err != nil {
		t.Fatalf("NormalizeTaskID(rel) error = %v", err)
	}
	if got != filepath.ToSlash(taskRel) {
		t.Fatalf("NormalizeTaskID(rel) = %q, want %q", got, filepath.ToSlash(taskRel))
	}

	got, err = NormalizeTaskID(festivalDir, "01_task.md")
	if err != nil {
		t.Fatalf("NormalizeTaskID(name) error = %v", err)
	}
	if got != "01_task.md" {
		t.Fatalf("NormalizeTaskID(name) = %q, want %q", got, "01_task.md")
	}
}

func TestResolveTaskStatus_UsesProgressStore(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()
	seqDir := filepath.Join(festivalDir, "001_PHASE", "01_seq")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	taskPath := filepath.Join(seqDir, "01_task.md")
	if err := os.WriteFile(taskPath, []byte("- [ ] not done\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}

	relKey := filepath.ToSlash(filepath.Join("001_PHASE", "01_seq", "01_task.md"))
	if err := mgr.MarkComplete(ctx, relKey); err != nil {
		t.Fatalf("MarkComplete error = %v", err)
	}

	status := ResolveTaskStatus(mgr.Store(), festivalDir, taskPath)
	if status != StatusCompleted {
		t.Fatalf("ResolveTaskStatus = %q, want %q", status, StatusCompleted)
	}
}

func TestResolveTaskStatus_FallsBackToLegacyKey(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()
	seqDir := filepath.Join(festivalDir, "001_PHASE", "01_seq")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	taskPath := filepath.Join(seqDir, "01_task.md")
	if err := os.WriteFile(taskPath, []byte("- [ ] not done\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}

	if err := mgr.MarkComplete(ctx, "01_task.md"); err != nil {
		t.Fatalf("MarkComplete legacy error = %v", err)
	}

	status := ResolveTaskStatus(mgr.Store(), festivalDir, taskPath)
	if status != StatusCompleted {
		t.Fatalf("ResolveTaskStatus = %q, want %q", status, StatusCompleted)
	}
}

func TestResolveTaskTime_ExplicitDataTakesPrecedence(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()
	seqDir := filepath.Join(festivalDir, "001_PHASE", "01_seq")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	taskPath := filepath.Join(seqDir, "01_task.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}

	// Set explicit time via store
	relKey := filepath.ToSlash(filepath.Join("001_PHASE", "01_seq", "01_task.md"))
	store := mgr.Store()
	store.SetTask(&TaskProgress{
		TaskID:           relKey,
		Status:           StatusCompleted,
		TimeSpentMinutes: 45, // Explicit time from API
	})

	got := ResolveTaskTime(store, festivalDir, taskPath)
	if got != 45 {
		t.Fatalf("ResolveTaskTime = %d, want 45 (explicit)", got)
	}
}

func TestResolveTaskTime_InfersFromFileTimestamps(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()
	seqDir := filepath.Join(festivalDir, "001_PHASE", "01_seq")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	taskPath := filepath.Join(seqDir, "01_task.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}

	// Mark complete with StartedAt but no TimeSpentMinutes
	relKey := filepath.ToSlash(filepath.Join("001_PHASE", "01_seq", "01_task.md"))
	store := mgr.Store()

	// Started 1 hour before file mod time
	modTime := GetFileModTime(taskPath)
	startTime := modTime.Add(-time.Hour) // 1 hour before

	store.SetTask(&TaskProgress{
		TaskID:    relKey,
		Status:    StatusCompleted,
		StartedAt: &startTime,
		// No CompletedAt, no TimeSpentMinutes - should be inferred
	})

	got := ResolveTaskTime(store, festivalDir, taskPath)
	// Should be approximately 60 minutes (1 hour)
	if got < 55 || got > 65 {
		t.Fatalf("ResolveTaskTime = %d, expected ~60 (inferred)", got)
	}
}

func TestResolveTaskTime_NoProgressReturnsZero(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()

	// Empty manager = empty store
	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}
	store := mgr.Store()

	got := ResolveTaskTime(store, festivalDir, "nonexistent.md")
	if got != 0 {
		t.Fatalf("ResolveTaskTime for missing task = %d, want 0", got)
	}
}

func TestResolveTaskProgress_AppliesTimeInference(t *testing.T) {
	ctx := context.Background()
	festivalDir := t.TempDir()
	seqDir := filepath.Join(festivalDir, "001_PHASE", "01_seq")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	taskPath := filepath.Join(seqDir, "01_task.md")
	if err := os.WriteFile(taskPath, []byte("# Task\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	mgr, err := NewManager(ctx, festivalDir)
	if err != nil {
		t.Fatalf("NewManager error = %v", err)
	}

	// Create task marked complete but without CompletedAt timestamp
	relKey := filepath.ToSlash(filepath.Join("001_PHASE", "01_seq", "01_task.md"))
	store := mgr.Store()
	store.SetTask(&TaskProgress{
		TaskID: relKey,
		Status: StatusCompleted,
		// No timestamps, no time - should be inferred from file
	})

	task, ok := ResolveTaskProgress(store, festivalDir, taskPath)
	if !ok {
		t.Fatal("ResolveTaskProgress returned false")
	}

	// CompletedAt should be set from file ModTime
	if task.CompletedAt == nil {
		t.Error("CompletedAt should be inferred from file ModTime")
	}
}
