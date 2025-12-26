package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadCatalog(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()

	// Create a template with frontmatter
	template1 := `---
template_id: TEST_TEMPLATE
aliases:
  - test
  - my_test
---
# Test Template

Content here.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "test_template.md"), []byte(template1), 0644); err != nil {
		t.Fatalf("failed to write template1: %v", err)
	}

	// Create another template
	template2 := `---
template_id: ANOTHER_TEMPLATE
---
# Another Template
`
	if err := os.WriteFile(filepath.Join(tmpDir, "another.md"), []byte(template2), 0644); err != nil {
		t.Fatalf("failed to write template2: %v", err)
	}

	// Create template without frontmatter
	template3 := `# No Frontmatter Template
Just content.
`
	if err := os.WriteFile(filepath.Join(tmpDir, "no_meta.md"), []byte(template3), 0644); err != nil {
		t.Fatalf("failed to write template3: %v", err)
	}

	catalog, err := LoadCatalog(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	tests := []struct {
		name    string
		id      string
		wantOK  bool
		wantExt string // expected file extension/name
	}{
		{
			name:    "finds by template_id",
			id:      "TEST_TEMPLATE",
			wantOK:  true,
			wantExt: "test_template.md",
		},
		{
			name:    "finds by alias",
			id:      "test",
			wantOK:  true,
			wantExt: "test_template.md",
		},
		{
			name:    "finds by second alias",
			id:      "my_test",
			wantOK:  true,
			wantExt: "test_template.md",
		},
		{
			name:    "finds another template",
			id:      "ANOTHER_TEMPLATE",
			wantOK:  true,
			wantExt: "another.md",
		},
		{
			name:   "unknown id returns false",
			id:     "UNKNOWN",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, ok := catalog.Resolve(tt.id)

			if ok != tt.wantOK {
				t.Errorf("Resolve(%q) ok = %v, want %v", tt.id, ok, tt.wantOK)
				return
			}

			if tt.wantOK && !contains(path, tt.wantExt) {
				t.Errorf("Resolve(%q) = %q, want path containing %q", tt.id, path, tt.wantExt)
			}
		})
	}
}

func TestCatalog_Resolve_EmptyID(t *testing.T) {
	catalog := &Catalog{byID: map[string]string{
		"TEST": "/path/to/test.md",
	}}

	_, ok := catalog.Resolve("")
	if ok {
		t.Error("Resolve with empty ID should return false")
	}
}

func TestManager_RenderByID(t *testing.T) {
	// Create temp directory with template
	tmpDir := t.TempDir()

	template1 := `---
template_id: GREETING
---
Hello, {{.name}}!
`
	if err := os.WriteFile(filepath.Join(tmpDir, "greeting.md"), []byte(template1), 0644); err != nil {
		t.Fatalf("failed to write template: %v", err)
	}

	catalog, err := LoadCatalog(context.Background(), tmpDir)
	if err != nil {
		t.Fatalf("LoadCatalog failed: %v", err)
	}

	mgr := NewManager()
	tmplCtx := NewContext()
	tmplCtx.SetCustom("name", "World")

	result, err := mgr.RenderByID(context.Background(), catalog, "GREETING", tmplCtx)
	if err != nil {
		t.Fatalf("RenderByID failed: %v", err)
	}

	expected := "Hello, World!"
	if result != expected {
		t.Errorf("got %q, want %q", result, expected)
	}
}

func TestManager_RenderByID_UnknownID(t *testing.T) {
	catalog := &Catalog{byID: map[string]string{}}
	mgr := NewManager()
	tmplCtx := NewContext()

	_, err := mgr.RenderByID(context.Background(), catalog, "UNKNOWN", tmplCtx)
	if err == nil {
		t.Error("expected error for unknown ID")
	}
}

func TestManager_RenderByID_NilCatalog(t *testing.T) {
	mgr := NewManager()
	tmplCtx := NewContext()

	_, err := mgr.RenderByID(context.Background(), nil, "TEST", tmplCtx)
	if err == nil {
		t.Error("expected error for nil catalog")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
