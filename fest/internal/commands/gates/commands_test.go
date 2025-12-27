package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverLocalGateTemplates(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(t *testing.T) string // returns gates dir path
		wantTemplates []string
	}{
		{
			name: "discovers .md files in gates directory",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gatesDir := filepath.Join(tmpDir, "gates")
				if err := os.MkdirAll(gatesDir, 0755); err != nil {
					t.Fatalf("failed to create gates dir: %v", err)
				}
				// Create some template files
				for _, name := range []string{"QUALITY_GATE_TESTING.md", "QUALITY_GATE_REVIEW.md", "CUSTOM_GATE.md"} {
					if err := os.WriteFile(filepath.Join(gatesDir, name), []byte("# Template"), 0644); err != nil {
						t.Fatalf("failed to create template: %v", err)
					}
				}
				return gatesDir
			},
			wantTemplates: []string{"CUSTOM_GATE", "QUALITY_GATE_REVIEW", "QUALITY_GATE_TESTING"},
		},
		{
			name: "returns empty for non-existent directory",
			setupFunc: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent", "gates")
			},
			wantTemplates: nil,
		},
		{
			name: "ignores non-.md files",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gatesDir := filepath.Join(tmpDir, "gates")
				if err := os.MkdirAll(gatesDir, 0755); err != nil {
					t.Fatalf("failed to create gates dir: %v", err)
				}
				// Create mixed files
				files := map[string]bool{
					"QUALITY_GATE.md": true,  // should be included
					"README.txt":      false, // should be ignored
					"config.yaml":     false, // should be ignored
					"ANOTHER_GATE.md": true,  // should be included
					".hidden.md":      true,  // hidden files with .md are still included
				}
				for name := range files {
					if err := os.WriteFile(filepath.Join(gatesDir, name), []byte("content"), 0644); err != nil {
						t.Fatalf("failed to create file: %v", err)
					}
				}
				return gatesDir
			},
			wantTemplates: []string{".hidden", "ANOTHER_GATE", "QUALITY_GATE"},
		},
		{
			name: "ignores subdirectories",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gatesDir := filepath.Join(tmpDir, "gates")
				if err := os.MkdirAll(gatesDir, 0755); err != nil {
					t.Fatalf("failed to create gates dir: %v", err)
				}
				// Create a template file
				if err := os.WriteFile(filepath.Join(gatesDir, "GATE.md"), []byte("# Gate"), 0644); err != nil {
					t.Fatalf("failed to create template: %v", err)
				}
				// Create a subdirectory (should be ignored)
				subDir := filepath.Join(gatesDir, "subdir")
				if err := os.MkdirAll(subDir, 0755); err != nil {
					t.Fatalf("failed to create subdir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(subDir, "NESTED.md"), []byte("# Nested"), 0644); err != nil {
					t.Fatalf("failed to create nested file: %v", err)
				}
				return gatesDir
			},
			wantTemplates: []string{"GATE"},
		},
		{
			name: "returns empty for empty directory",
			setupFunc: func(t *testing.T) string {
				tmpDir := t.TempDir()
				gatesDir := filepath.Join(tmpDir, "gates")
				if err := os.MkdirAll(gatesDir, 0755); err != nil {
					t.Fatalf("failed to create gates dir: %v", err)
				}
				return gatesDir
			},
			wantTemplates: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gatesDir := tt.setupFunc(t)
			got := discoverLocalGateTemplates(gatesDir)

			// Check length
			if len(got) != len(tt.wantTemplates) {
				t.Errorf("discoverLocalGateTemplates() returned %d templates, want %d\nGot: %v\nWant: %v",
					len(got), len(tt.wantTemplates), got, tt.wantTemplates)
				return
			}

			// Check contents (templates may be in different order, so compare as sets)
			gotSet := make(map[string]bool)
			for _, tmpl := range got {
				gotSet[tmpl] = true
			}

			for _, want := range tt.wantTemplates {
				if !gotSet[want] {
					t.Errorf("discoverLocalGateTemplates() missing template %q\nGot: %v\nWant: %v",
						want, got, tt.wantTemplates)
				}
			}
		})
	}
}
