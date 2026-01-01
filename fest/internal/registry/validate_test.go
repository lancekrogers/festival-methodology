package registry

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestRegistry_Validate(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	festivalsRoot := tmpDir

	// Create festival directory structure
	activePath := filepath.Join(festivalsRoot, "active")
	os.MkdirAll(activePath, 0755)

	// Create a festival with ID in directory name
	festivalPath := filepath.Join(activePath, "guild-usable_GU0001")
	os.MkdirAll(festivalPath, 0755)

	regPath := filepath.Join(festivalsRoot, ".festival", "id_registry.yaml")
	os.MkdirAll(filepath.Dir(regPath), 0755)

	reg := New(regPath)
	entry := RegistryEntry{
		ID:     "GU0001",
		Name:   "guild-usable",
		Status: "active",
		Path:   festivalPath,
	}
	_ = reg.Add(ctx, entry)

	// Add stale entry (doesn't exist on filesystem)
	staleEntry := RegistryEntry{
		ID:     "ST0001",
		Name:   "stale",
		Status: "active",
		Path:   filepath.Join(activePath, "stale_ST0001"),
	}
	_ = reg.Add(ctx, staleEntry)

	// Validate should find the stale entry
	errors, err := reg.Validate(ctx, festivalsRoot)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	found := false
	for _, e := range errors {
		if e.ID == "ST0001" && e.Type == "missing_in_fs" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Validate() should find missing_in_fs for stale entry")
	}
}

func TestRebuild(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	festivalsRoot := tmpDir

	// Create status directories
	activePath := filepath.Join(festivalsRoot, "active")
	completedPath := filepath.Join(festivalsRoot, "completed", "2025-01")
	os.MkdirAll(activePath, 0755)
	os.MkdirAll(completedPath, 0755)

	// Create festivals with IDs
	festival1 := filepath.Join(activePath, "first_GU0001")
	festival2 := filepath.Join(activePath, "second_FN0002")
	festival3 := filepath.Join(completedPath, "done_DO0003")
	os.MkdirAll(festival1, 0755)
	os.MkdirAll(festival2, 0755)
	os.MkdirAll(festival3, 0755)

	regPath := filepath.Join(festivalsRoot, ".festival", "id_registry.yaml")
	os.MkdirAll(filepath.Dir(regPath), 0755)

	reg, err := Rebuild(ctx, festivalsRoot, regPath)
	if err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}

	if reg.Count(ctx) != 3 {
		t.Errorf("Rebuild() registry count = %d, want 3", reg.Count(ctx))
	}

	// Check specific entries
	if !reg.Exists(ctx, "GU0001") {
		t.Error("Rebuild() should include GU0001")
	}
	if !reg.Exists(ctx, "FN0002") {
		t.Error("Rebuild() should include FN0002")
	}
	if !reg.Exists(ctx, "DO0003") {
		t.Error("Rebuild() should include DO0003")
	}
}

func TestRebuild_NoFestivals(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	festivalsRoot := tmpDir

	regPath := filepath.Join(festivalsRoot, ".festival", "id_registry.yaml")

	reg, err := Rebuild(ctx, festivalsRoot, regPath)
	if err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}

	if reg.Count(ctx) != 0 {
		t.Errorf("Rebuild() registry count = %d, want 0 for empty festivals", reg.Count(ctx))
	}
}

func TestRebuild_SkipsLegacyFestivals(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	festivalsRoot := tmpDir

	activePath := filepath.Join(festivalsRoot, "active")
	os.MkdirAll(activePath, 0755)

	// Create festival without ID (legacy)
	legacy := filepath.Join(activePath, "legacy-festival")
	os.MkdirAll(legacy, 0755)

	// Create festival with ID
	withID := filepath.Join(activePath, "modern_MO0001")
	os.MkdirAll(withID, 0755)

	regPath := filepath.Join(festivalsRoot, ".festival", "id_registry.yaml")

	reg, err := Rebuild(ctx, festivalsRoot, regPath)
	if err != nil {
		t.Fatalf("Rebuild() error = %v", err)
	}

	if reg.Count(ctx) != 1 {
		t.Errorf("Rebuild() should only include 1 festival (with ID), got %d", reg.Count(ctx))
	}
	if !reg.Exists(ctx, "MO0001") {
		t.Error("Rebuild() should include MO0001")
	}
}

func TestRegistry_Sync(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	festivalsRoot := tmpDir

	activePath := filepath.Join(festivalsRoot, "active")
	os.MkdirAll(activePath, 0755)

	// Create festival on filesystem
	festival := filepath.Join(activePath, "test_TE0001")
	os.MkdirAll(festival, 0755)

	regPath := filepath.Join(festivalsRoot, ".festival", "id_registry.yaml")
	os.MkdirAll(filepath.Dir(regPath), 0755)

	reg := New(regPath)

	// Add stale entry
	_ = reg.Add(ctx, RegistryEntry{
		ID:     "ST0001",
		Name:   "stale",
		Status: "active",
		Path:   filepath.Join(activePath, "stale_ST0001"),
	})

	// Sync should add TE0001 and remove ST0001
	if err := reg.Sync(ctx, festivalsRoot); err != nil {
		t.Fatalf("Sync() error = %v", err)
	}

	if reg.Exists(ctx, "ST0001") {
		t.Error("Sync() should remove stale entry ST0001")
	}
	if !reg.Exists(ctx, "TE0001") {
		t.Error("Sync() should add missing entry TE0001")
	}
}

func TestValidate_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	reg := New("/tmp/test.yaml")
	_, err := reg.Validate(ctx, "/tmp/festivals")
	if err == nil {
		t.Error("Validate() should return error with cancelled context")
	}
}

func TestRebuild_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Rebuild(ctx, "/tmp/festivals", "/tmp/registry.yaml")
	if err == nil {
		t.Error("Rebuild() should return error with cancelled context")
	}
}

func TestDetectStatus(t *testing.T) {
	tests := []struct {
		path   string
		want   string
	}{
		{"/festivals/active/test_TE0001", "active"},
		{"/festivals/planned/test_TE0001", "planned"},
		{"/festivals/completed/2025-01/test_TE0001", "completed"},
		{"/festivals/dungeon/test_TE0001", "dungeon"},
		{"/random/path/test", "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			got := detectStatus(tt.path)
			if got != tt.want {
				t.Errorf("detectStatus(%q) = %q, want %q", tt.path, got, tt.want)
			}
		})
	}
}
