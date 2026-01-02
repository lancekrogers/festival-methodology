package progress

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

func TestResolveTaskID_WithTaskPath(t *testing.T) {
	festivalDir := setupFestival(t, []string{
		"002_FOUNDATION/01_alpha/01_design.md",
	})

	taskPath := filepath.Join(festivalDir, "002_FOUNDATION", "01_alpha", "01_design.md")
	opts := &progressOptions{taskPath: taskPath}

	got, err := resolveTaskID(festivalDir, opts)
	if err != nil {
		t.Fatalf("resolveTaskID() error = %v", err)
	}

	want := filepath.ToSlash(filepath.Join("002_FOUNDATION", "01_alpha", "01_design.md"))
	if got != want {
		t.Errorf("resolveTaskID() = %q, want %q", got, want)
	}
}

func TestResolveTaskID_PhaseSequence(t *testing.T) {
	festivalDir := setupFestival(t, []string{
		"002_FOUNDATION/01_alpha/01_design.md",
	})

	opts := &progressOptions{
		taskID:   "01_design",
		phase:    "002_FOUNDATION",
		sequence: "01_alpha",
	}

	got, err := resolveTaskID(festivalDir, opts)
	if err != nil {
		t.Fatalf("resolveTaskID() error = %v", err)
	}

	want := filepath.ToSlash(filepath.Join("002_FOUNDATION", "01_alpha", "01_design.md"))
	if got != want {
		t.Errorf("resolveTaskID() = %q, want %q", got, want)
	}
}

func TestResolveTaskID_UniqueBareName(t *testing.T) {
	festivalDir := setupFestival(t, []string{
		"002_FOUNDATION/01_alpha/02_unique.md",
		"003_OTHER/01_beta/01_design.md",
	})

	opts := &progressOptions{taskID: "02_unique.md"}

	got, err := resolveTaskID(festivalDir, opts)
	if err != nil {
		t.Fatalf("resolveTaskID() error = %v", err)
	}

	want := filepath.ToSlash(filepath.Join("002_FOUNDATION", "01_alpha", "02_unique.md"))
	if got != want {
		t.Errorf("resolveTaskID() = %q, want %q", got, want)
	}
}

func TestResolveTaskID_AmbiguousBareName(t *testing.T) {
	festivalDir := setupFestival(t, []string{
		"002_FOUNDATION/01_alpha/01_design.md",
		"003_OTHER/01_beta/01_design.md",
	})

	opts := &progressOptions{taskID: "01_design.md"}

	_, err := resolveTaskID(festivalDir, opts)
	if err == nil {
		t.Fatal("resolveTaskID() expected error, got nil")
	}
	if errors.Code(err) != errors.ErrCodeValidation {
		t.Fatalf("resolveTaskID() error code = %q, want %q", errors.Code(err), errors.ErrCodeValidation)
	}
	if !strings.Contains(err.Error(), "ambiguous") {
		t.Fatalf("resolveTaskID() error = %q, want ambiguous error", err.Error())
	}
}

func setupFestival(t *testing.T, taskPaths []string) string {
	t.Helper()

	dir := t.TempDir()
	for _, rel := range taskPaths {
		writeTask(t, filepath.Join(dir, rel))
	}
	return dir
}

func writeTask(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("# Task\n"), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}
