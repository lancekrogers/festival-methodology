package gates

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTemplateResolver(t *testing.T) {
	resolver := NewTemplateResolver("/festivals")
	if resolver.FestivalsRoot() != "/festivals" {
		t.Errorf("FestivalsRoot = %q, want %q", resolver.FestivalsRoot(), "/festivals")
	}
}

func TestTemplateResolver_Resolve(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	festivalPath := filepath.Join(tmpDir, "test-festival")
	phasePath := filepath.Join(festivalPath, "002_IMPLEMENT")
	sequencePath := filepath.Join(phasePath, "01_core")
	os.MkdirAll(sequencePath, 0755)

	// Create template at sequence level
	seqTemplateDir := filepath.Join(sequencePath, ".fest.templates")
	os.MkdirAll(seqTemplateDir, 0755)
	templatePath := filepath.Join(seqTemplateDir, "CUSTOM_GATE.md")
	os.WriteFile(templatePath, []byte("# Custom Gate"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	// Should find sequence-level template
	result, err := resolver.Resolve("CUSTOM_GATE", festivalPath, phasePath, sequencePath)
	if err != nil {
		t.Fatalf("Resolve error: %v", err)
	}
	if !result.Exists {
		t.Error("Expected template to exist")
	}
	if result.Level != PolicyLevelSequence {
		t.Errorf("Expected level %q, got %q", PolicyLevelSequence, result.Level)
	}
	if result.Path != templatePath {
		t.Errorf("Expected path %q, got %q", templatePath, result.Path)
	}
}

func TestTemplateResolver_ResolveNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	resolver := NewTemplateResolver(tmpDir)

	result, err := resolver.Resolve("NONEXISTENT", tmpDir, tmpDir, tmpDir)
	if err == nil {
		t.Error("Expected error for nonexistent template")
	}
	if result.Exists {
		t.Error("Expected Exists to be false")
	}

	// Should be TemplateNotFoundError
	if _, ok := err.(*TemplateNotFoundError); !ok {
		t.Errorf("Expected TemplateNotFoundError, got %T", err)
	}
}

func TestTemplateResolver_Cache(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create template
	templateDir := filepath.Join(tmpDir, ".festival", "templates")
	os.MkdirAll(templateDir, 0755)
	os.WriteFile(filepath.Join(templateDir, "CACHED.md"), []byte("# Cached"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	// First call - should cache
	result1, _ := resolver.Resolve("CACHED", tmpDir, tmpDir, tmpDir)

	// Second call - should use cache
	result2, _ := resolver.Resolve("CACHED", tmpDir, tmpDir, tmpDir)

	if result1.Path != result2.Path {
		t.Error("Cache not working - different paths returned")
	}

	// Clear cache and verify
	resolver.ClearCache()
}

func TestTemplateResolver_HierarchyPrecedence(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	festivalPath := filepath.Join(tmpDir, "festival")
	phasePath := filepath.Join(festivalPath, "phase")
	sequencePath := filepath.Join(phasePath, "sequence")

	// Create templates at multiple levels
	// Built-in
	builtinDir := filepath.Join(tmpDir, ".festival", "templates")
	os.MkdirAll(builtinDir, 0755)
	os.WriteFile(filepath.Join(builtinDir, "TEST.md"), []byte("builtin"), 0644)

	// Festival
	festivalTemplateDir := filepath.Join(festivalPath, ".festival", "templates")
	os.MkdirAll(festivalTemplateDir, 0755)
	os.WriteFile(filepath.Join(festivalTemplateDir, "TEST.md"), []byte("festival"), 0644)

	// Phase
	phaseTemplateDir := filepath.Join(phasePath, ".fest.templates")
	os.MkdirAll(phaseTemplateDir, 0755)
	os.WriteFile(filepath.Join(phaseTemplateDir, "TEST.md"), []byte("phase"), 0644)

	// Sequence
	seqTemplateDir := filepath.Join(sequencePath, ".fest.templates")
	os.MkdirAll(seqTemplateDir, 0755)
	os.WriteFile(filepath.Join(seqTemplateDir, "TEST.md"), []byte("sequence"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	// Should find sequence level (most specific)
	result, _ := resolver.Resolve("TEST", festivalPath, phasePath, sequencePath)
	if result.Level != PolicyLevelSequence {
		t.Errorf("Expected sequence level, got %q", result.Level)
	}

	// Clear cache for phase test
	resolver.ClearCache()

	// Remove sequence template, should find phase
	os.Remove(filepath.Join(seqTemplateDir, "TEST.md"))
	result, _ = resolver.Resolve("TEST", festivalPath, phasePath, sequencePath)
	if result.Level != PolicyLevelPhase {
		t.Errorf("Expected phase level, got %q", result.Level)
	}
}

func TestTemplateNotFoundError(t *testing.T) {
	err := &TemplateNotFoundError{
		TemplateID: "MISSING",
		Searched:   []string{"/path1", "/path2"},
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error message is empty")
	}
	if !containsSubstring(msg, "MISSING") {
		t.Error("Error message should contain template ID")
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestTemplateResolver_GatesPrefixResolution(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-gates-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create directory structure
	festivalPath := filepath.Join(tmpDir, "festival")
	phasePath := filepath.Join(festivalPath, "002_IMPLEMENT")
	sequencePath := filepath.Join(phasePath, "01_core")
	os.MkdirAll(sequencePath, 0755)

	// Create template in festival's gates/ directory
	gatesDir := filepath.Join(festivalPath, "gates")
	os.MkdirAll(gatesDir, 0755)
	gatesTemplatePath := filepath.Join(gatesDir, "QUALITY_GATE_TESTING.md")
	os.WriteFile(gatesTemplatePath, []byte("# Testing Gate"), 0644)

	// Also create fallback in global templates
	globalTemplateDir := filepath.Join(tmpDir, ".festival", "templates")
	os.MkdirAll(globalTemplateDir, 0755)
	globalTemplatePath := filepath.Join(globalTemplateDir, "QUALITY_GATE_TESTING.md")
	os.WriteFile(globalTemplatePath, []byte("# Global Testing Gate"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	t.Run("gates/ prefix resolves to festival gates directory", func(t *testing.T) {
		result, err := resolver.Resolve("gates/QUALITY_GATE_TESTING", festivalPath, phasePath, sequencePath)
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if !result.Exists {
			t.Error("Expected template to exist")
		}
		if result.Path != gatesTemplatePath {
			t.Errorf("Expected path %q, got %q", gatesTemplatePath, result.Path)
		}
		if result.Level != PolicyLevelFestival {
			t.Errorf("Expected level %q, got %q", PolicyLevelFestival, result.Level)
		}
	})

	t.Run("gates/ prefix falls back to global when not in festival gates/", func(t *testing.T) {
		// Remove the festival-level template
		os.Remove(gatesTemplatePath)
		resolver.ClearCache()

		result, err := resolver.Resolve("gates/QUALITY_GATE_TESTING", festivalPath, phasePath, sequencePath)
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if !result.Exists {
			t.Error("Expected template to exist via fallback")
		}
		if result.Path != globalTemplatePath {
			t.Errorf("Expected fallback path %q, got %q", globalTemplatePath, result.Path)
		}
	})

	t.Run("without gates/ prefix searches standard hierarchy", func(t *testing.T) {
		resolver.ClearCache()

		result, err := resolver.Resolve("QUALITY_GATE_TESTING", festivalPath, phasePath, sequencePath)
		if err != nil {
			t.Fatalf("Resolve error: %v", err)
		}
		if !result.Exists {
			t.Error("Expected template to exist")
		}
		// Should find global template since no gates/ prefix
		if result.Path != globalTemplatePath {
			t.Errorf("Expected path %q, got %q", globalTemplatePath, result.Path)
		}
	})
}

func TestTemplateResolver_ResolveForFestival_GatesPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-festival-gates-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create festival with gates directory
	festivalPath := filepath.Join(tmpDir, "festival")
	gatesDir := filepath.Join(festivalPath, "gates")
	os.MkdirAll(gatesDir, 0755)
	gatesTemplatePath := filepath.Join(gatesDir, "MY_GATE.md")
	os.WriteFile(gatesTemplatePath, []byte("# My Gate"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	result, err := resolver.ResolveForFestival("gates/MY_GATE", festivalPath)
	if err != nil {
		t.Fatalf("ResolveForFestival error: %v", err)
	}
	if !result.Exists {
		t.Error("Expected template to exist")
	}
	if result.Path != gatesTemplatePath {
		t.Errorf("Expected path %q, got %q", gatesTemplatePath, result.Path)
	}
}

func TestTemplateResolver_ResolveForPhase_GatesPrefix(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "resolver-phase-gates-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create festival with gates directory
	festivalPath := filepath.Join(tmpDir, "festival")
	phasePath := filepath.Join(festivalPath, "001_PLAN")
	os.MkdirAll(phasePath, 0755)

	gatesDir := filepath.Join(festivalPath, "gates")
	os.MkdirAll(gatesDir, 0755)
	gatesTemplatePath := filepath.Join(gatesDir, "PHASE_GATE.md")
	os.WriteFile(gatesTemplatePath, []byte("# Phase Gate"), 0644)

	resolver := NewTemplateResolver(tmpDir)

	result, err := resolver.ResolveForPhase("gates/PHASE_GATE", festivalPath, phasePath)
	if err != nil {
		t.Fatalf("ResolveForPhase error: %v", err)
	}
	if !result.Exists {
		t.Error("Expected template to exist")
	}
	if result.Path != gatesTemplatePath {
		t.Errorf("Expected path %q, got %q", gatesTemplatePath, result.Path)
	}
}
