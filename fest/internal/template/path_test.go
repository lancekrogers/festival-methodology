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

func TestFindFestivalsRoot(t *testing.T) {
	// Create temp directory structure: /tmp/xxx/festivals/.festival/
	tmpDir := t.TempDir()

	festivalsDir := filepath.Join(tmpDir, "festivals")
	festivalMetaDir := filepath.Join(festivalsDir, ".festival")
	if err := os.MkdirAll(festivalMetaDir, 0755); err != nil {
		t.Fatalf("failed to create festivals/.festival dir: %v", err)
	}

	// Create nested directory inside festivals
	nestedDir := filepath.Join(festivalsDir, "active", "my-festival", "001_PLANNING")
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
			name:      "finds root from festivals directory",
			startDir:  festivalsDir,
			wantRoot:  festivalsDir,
			wantError: false,
		},
		{
			name:      "finds root from nested directory inside festivals",
			startDir:  nestedDir,
			wantRoot:  festivalsDir,
			wantError: false,
		},
		{
			name:      "error when not inside festivals tree",
			startDir:  tmpDir,
			wantRoot:  "",
			wantError: true,
		},
		{
			name:      "error when no festivals directory",
			startDir:  "/tmp",
			wantRoot:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FindFestivalsRoot(tt.startDir)

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
	// Create temp directory structure: /tmp/xxx/festivals/.festival/templates/
	tmpDir := t.TempDir()

	festivalsDir := filepath.Join(tmpDir, "festivals")
	templatesDir := filepath.Join(festivalsDir, ".festival", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("failed to create templates dir: %v", err)
	}

	// Create nested directory inside festivals
	nestedDir := filepath.Join(festivalsDir, "active", "my-festival")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("failed to create nested dir: %v", err)
	}

	tests := []struct {
		name          string
		startDir      string
		wantTemplates string
		wantError     bool
	}{
		{
			name:          "returns templates path from festivals root",
			startDir:      festivalsDir,
			wantTemplates: templatesDir,
			wantError:     false,
		},
		{
			name:          "returns templates path from nested directory",
			startDir:      nestedDir,
			wantTemplates: templatesDir,
			wantError:     false,
		},
		{
			name:          "error when not inside festivals tree",
			startDir:      tmpDir,
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
