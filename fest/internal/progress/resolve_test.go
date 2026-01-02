package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
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
