package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindWorkspaceRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create .festival directory
	festivalDir := filepath.Join(tmpDir, ".festival")
	if err := os.MkdirAll(festivalDir, 0755); err != nil {
		t.Fatalf("failed to create .festival dir: %v", err)
	}

	// Create nested directory
	nestedDir := filepath.Join(tmpDir, "sub", "deep", "path")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	tests := []struct {
		name      string
		startDir  string
		wantRoot  string
		wantError bool
	}{
		{
			name:      "finds root from workspace root",
			startDir:  tmpDir,
			wantRoot:  tmpDir,
			wantError: false,
		},
		{
			name:      "finds root from nested directory",
			startDir:  nestedDir,
			wantRoot:  tmpDir,
			wantError: false,
		},
		{
			name:      "finds root from .festival directory",
			startDir:  festivalDir,
			wantRoot:  tmpDir,
			wantError: false,
		},
		{
			name:      "error when no .festival found",
			startDir:  "/tmp",
			wantRoot:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindWorkspaceRoot(tt.startDir)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.wantRoot {
				t.Errorf("got %q, want %q", got, tt.wantRoot)
			}
		})
	}
}

func TestLocalTemplateRoot(t *testing.T) {
	// Create temp directory structure
	tmpDir := t.TempDir()

	// Create .festival/templates directory
	templatesDir := filepath.Join(tmpDir, ".festival", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	tests := []struct {
		name          string
		startDir      string
		wantTemplates string
		wantError     bool
	}{
		{
			name:          "returns templates path from workspace root",
			startDir:      tmpDir,
			wantTemplates: templatesDir,
			wantError:     false,
		},
		{
			name:          "error when no workspace found",
			startDir:      "/tmp",
			wantTemplates: "",
			wantError:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LocalTemplateRoot(tt.startDir)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if got != tt.wantTemplates {
				t.Errorf("got %q, want %q", got, tt.wantTemplates)
			}
		})
	}
}
