package extensions

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadExtensionManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manifestContent := `name: test-extension
version: "1.0.0"
description: A test extension
author: Test Author
type: workflow
tags:
  - testing
  - example
files:
  - path: README.md
    description: Documentation
`
	manifestPath := filepath.Join(tmpDir, ExtensionManifestFileName)
	if err := os.WriteFile(manifestPath, []byte(manifestContent), 0644); err != nil {
		t.Fatal(err)
	}

	manifest, err := LoadExtensionManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadExtensionManifest error: %v", err)
	}

	if manifest.Name != "test-extension" {
		t.Errorf("Name = %q, want %q", manifest.Name, "test-extension")
	}
	if manifest.Version != "1.0.0" {
		t.Errorf("Version = %q, want %q", manifest.Version, "1.0.0")
	}
	if manifest.Type != "workflow" {
		t.Errorf("Type = %q, want %q", manifest.Type, "workflow")
	}
	if len(manifest.Tags) != 2 {
		t.Errorf("Tags count = %d, want 2", len(manifest.Tags))
	}
}

func TestSaveExtensionManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manifest := &ExtensionManifest{
		Name:        "my-extension",
		Version:     "2.0.0",
		Description: "My extension",
		Type:        "template",
	}

	manifestPath := filepath.Join(tmpDir, ExtensionManifestFileName)
	if err := SaveExtensionManifest(manifestPath, manifest); err != nil {
		t.Fatalf("SaveExtensionManifest error: %v", err)
	}

	// Load and verify
	loaded, err := LoadExtensionManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadExtensionManifest error: %v", err)
	}
	if loaded.Name != "my-extension" {
		t.Errorf("Name = %q, want %q", loaded.Name, "my-extension")
	}
}

func TestLoadExtensionFromDir_WithManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create manifest
	manifestContent := `name: with-manifest
version: "1.0"
description: Has a manifest
type: agent
`
	manifestPath := filepath.Join(tmpDir, ExtensionManifestFileName)
	os.WriteFile(manifestPath, []byte(manifestContent), 0644)

	ext, err := LoadExtensionFromDir(tmpDir, "test")
	if err != nil {
		t.Fatalf("LoadExtensionFromDir error: %v", err)
	}

	if ext.Name != "with-manifest" {
		t.Errorf("Name = %q, want %q", ext.Name, "with-manifest")
	}
	if ext.Source != "test" {
		t.Errorf("Source = %q, want %q", ext.Source, "test")
	}
	if ext.Type != "agent" {
		t.Errorf("Type = %q, want %q", ext.Type, "agent")
	}
}

func TestLoadExtensionFromDir_WithoutManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create extension dir without manifest
	extDir := filepath.Join(tmpDir, "my-extension")
	os.MkdirAll(extDir, 0755)

	ext, err := LoadExtensionFromDir(extDir, "builtin")
	if err != nil {
		t.Fatalf("LoadExtensionFromDir error: %v", err)
	}

	// Should use directory name as extension name
	if ext.Name != "my-extension" {
		t.Errorf("Name = %q, want %q", ext.Name, "my-extension")
	}
	if ext.Source != "builtin" {
		t.Errorf("Source = %q, want %q", ext.Source, "builtin")
	}
}

func TestExtensionListFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some files
	os.WriteFile(filepath.Join(tmpDir, "README.md"), []byte("# Test"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "template.md"), []byte("Template"), 0644)
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "nested.md"), []byte("Nested"), 0644)

	ext := &Extension{Path: tmpDir}
	files, err := ext.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles error: %v", err)
	}

	if len(files) != 3 {
		t.Errorf("ListFiles count = %d, want 3", len(files))
	}
}

func TestExtensionHasFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	os.WriteFile(filepath.Join(tmpDir, "exists.md"), []byte("exists"), 0644)

	ext := &Extension{Path: tmpDir}

	if !ext.HasFile("exists.md") {
		t.Error("HasFile returned false for existing file")
	}
	if ext.HasFile("not-exists.md") {
		t.Error("HasFile returned true for non-existing file")
	}
}

func TestExtensionLoaderLoadAll(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "ext-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create extensions directory structure
	extDir := filepath.Join(tmpDir, ".festival", "extensions")
	os.MkdirAll(extDir, 0755)

	// Create an extension
	ext1Dir := filepath.Join(extDir, "ext1")
	os.MkdirAll(ext1Dir, 0755)
	os.WriteFile(filepath.Join(ext1Dir, ExtensionManifestFileName),
		[]byte("name: ext1\nversion: '1.0'\ntype: workflow"), 0644)

	ext2Dir := filepath.Join(extDir, "ext2")
	os.MkdirAll(ext2Dir, 0755)
	// No manifest - should use directory name

	loader := NewExtensionLoader()
	loader.loadFromDirectory(extDir, "test")

	if loader.Count() != 2 {
		t.Errorf("Count = %d, want 2", loader.Count())
	}

	ext1 := loader.Get("ext1")
	if ext1 == nil {
		t.Error("ext1 not loaded")
	} else if ext1.Type != "workflow" {
		t.Errorf("ext1 type = %q, want %q", ext1.Type, "workflow")
	}

	ext2 := loader.Get("ext2")
	if ext2 == nil {
		t.Error("ext2 not loaded")
	}
}

func TestExtensionLoaderListBySource(t *testing.T) {
	loader := NewExtensionLoader()
	loader.extensions = map[string]*Extension{
		"a": {Name: "a", Source: "project"},
		"b": {Name: "b", Source: "user"},
		"c": {Name: "c", Source: "project"},
	}

	project := loader.ListBySource("project")
	if len(project) != 2 {
		t.Errorf("ListBySource('project') count = %d, want 2", len(project))
	}

	user := loader.ListBySource("user")
	if len(user) != 1 {
		t.Errorf("ListBySource('user') count = %d, want 1", len(user))
	}
}

func TestExtensionLoaderListByType(t *testing.T) {
	loader := NewExtensionLoader()
	loader.extensions = map[string]*Extension{
		"a": {Name: "a", Type: "workflow"},
		"b": {Name: "b", Type: "template"},
		"c": {Name: "c", Type: "workflow"},
	}

	workflow := loader.ListByType("workflow")
	if len(workflow) != 2 {
		t.Errorf("ListByType('workflow') count = %d, want 2", len(workflow))
	}
}

func TestExtensionLoaderHasExtension(t *testing.T) {
	loader := NewExtensionLoader()
	loader.extensions = map[string]*Extension{
		"exists": {Name: "exists"},
	}

	if !loader.HasExtension("exists") {
		t.Error("HasExtension returned false for existing extension")
	}
	if loader.HasExtension("not-exists") {
		t.Error("HasExtension returned true for non-existing extension")
	}
}
