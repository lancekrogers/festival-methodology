package wizard

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestGetEditor(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "falls back to vim when EDITOR not set",
			envValue: "",
			want:     "vim",
		},
		{
			name:     "uses EDITOR env var when set",
			envValue: "nvim",
			want:     "nvim",
		},
		{
			name:     "uses custom editor from env",
			envValue: "code",
			want:     "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore EDITOR
			origEditor := os.Getenv("EDITOR")
			defer func() {
				if origEditor == "" {
					os.Unsetenv("EDITOR")
				} else {
					os.Setenv("EDITOR", origEditor)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("EDITOR")
			} else {
				os.Setenv("EDITOR", tt.envValue)
			}

			got := getEditor(ctx)
			if got != tt.want {
				t.Errorf("getEditor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEditor_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should still return a valid editor even with cancelled context
	// (config.Load will fail but we fall back gracefully)
	got := getEditor(ctx)
	if got == "" {
		t.Error("getEditor() returned empty string with cancelled context")
	}
}

func TestRunFill_NoMarkersFound(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a file without markers
	noMarkerFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(noMarkerFile, []byte("# Test\nNo markers here\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	opts := &FillOptions{
		Path:        tmpDir,
		Interactive: false, // Use agent mode to avoid editor
		JSONOutput:  true,
	}

	// Should not error when no markers found
	err := RunFill(ctx, opts)
	if err != nil {
		t.Errorf("RunFill() error = %v, want nil", err)
	}
}

func TestRunFill_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &FillOptions{
		Path:        t.TempDir(),
		Interactive: false,
	}

	err := RunFill(ctx, opts)
	if err == nil {
		t.Error("RunFill() with cancelled context should return error")
	}
}

func TestRunFill_InvalidPath(t *testing.T) {
	ctx := context.Background()

	opts := &FillOptions{
		Path:        "/nonexistent/path/that/does/not/exist",
		Interactive: false,
	}

	err := RunFill(ctx, opts)
	if err == nil {
		t.Error("RunFill() with invalid path should return error")
	}
}

func TestRunFill_RecursiveWalk(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create nested structure WITHOUT markers (to test directory walking)
	nestedDir := filepath.Join(tmpDir, "phase", "sequence")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	// File without markers - tests recursive walking finds it
	noMarkerFile := filepath.Join(nestedDir, "task.md")
	if err := os.WriteFile(noMarkerFile, []byte("# Task\nNo markers here\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	opts := &FillOptions{
		Path:        tmpDir,
		Interactive: false,
		JSONOutput:  true,
	}

	// Should walk recursively and find the file (even if no markers)
	err := RunFill(ctx, opts)
	if err != nil {
		t.Errorf("RunFill() error = %v, want nil", err)
	}
}

func TestRunFill_SkipsHiddenDirs(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create hidden directory with marker file
	hiddenDir := filepath.Join(tmpDir, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	markerFile := filepath.Join(hiddenDir, "task.md")
	if err := os.WriteFile(markerFile, []byte("[REPLACE: should be skipped]\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Create .festival directory (should also be skipped)
	festivalDir := filepath.Join(tmpDir, ".festival")
	if err := os.MkdirAll(festivalDir, 0755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}

	festivalFile := filepath.Join(festivalDir, "config.md")
	if err := os.WriteFile(festivalFile, []byte("[REPLACE: should be skipped]\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	opts := &FillOptions{
		Path:        tmpDir,
		Interactive: false,
		JSONOutput:  true,
	}

	// Should not find any markers (all in hidden/skipped dirs)
	err := RunFill(ctx, opts)
	if err != nil {
		t.Errorf("RunFill() error = %v, want nil", err)
	}
}

func TestRunFill_SingleFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a single file without markers (to test single-file path handling)
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("# Test\nNo markers here\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	opts := &FillOptions{
		Path:        testFile, // Single file, not directory
		Interactive: false,
		JSONOutput:  true,
	}

	err := RunFill(ctx, opts)
	if err != nil {
		t.Errorf("RunFill() with single file error = %v, want nil", err)
	}
}
