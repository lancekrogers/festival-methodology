package registry

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	reg := New("/path/to/registry.yaml")

	if reg.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", reg.Version, CurrentVersion)
	}
	if reg.Entries == nil {
		t.Error("Entries should be initialized, got nil")
	}
	if len(reg.Entries) != 0 {
		t.Errorf("Entries length = %d, want 0", len(reg.Entries))
	}
}

func TestRegistry_Add(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	entry := RegistryEntry{
		ID:        "GU0001",
		Name:      "guild-usable",
		Status:    "active",
		Path:      "/festivals/active/guild-usable_GU0001",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := reg.Add(ctx, entry); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify entry was added
	if !reg.Exists(ctx, "GU0001") {
		t.Error("Entry should exist after Add")
	}

	// Test duplicate add
	if err := reg.Add(ctx, entry); err == nil {
		t.Error("Add() should return error for duplicate ID")
	}
}

func TestRegistry_Get(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	entry := RegistryEntry{
		ID:        "GU0001",
		Name:      "guild-usable",
		Status:    "active",
		Path:      "/festivals/active/guild-usable_GU0001",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = reg.Add(ctx, entry)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{"existing entry", "GU0001", false},
		{"non-existent entry", "XX9999", true},
		{"empty ID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := reg.Get(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got.ID != tt.id {
				t.Errorf("Get() ID = %v, want %v", got.ID, tt.id)
			}
		})
	}
}

func TestRegistry_Update(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	entry := RegistryEntry{
		ID:        "GU0001",
		Name:      "guild-usable",
		Status:    "active",
		Path:      "/festivals/active/guild-usable_GU0001",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_ = reg.Add(ctx, entry)

	// Update status
	entry.Status = "completed"
	entry.Path = "/festivals/completed/2025-01/guild-usable_GU0001"
	if err := reg.Update(ctx, entry); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Verify update
	got, _ := reg.Get(ctx, "GU0001")
	if got.Status != "completed" {
		t.Errorf("Status = %q, want %q", got.Status, "completed")
	}

	// Test update non-existent
	nonExistent := RegistryEntry{ID: "XX9999"}
	if err := reg.Update(ctx, nonExistent); err == nil {
		t.Error("Update() should return error for non-existent ID")
	}
}

func TestRegistry_Delete(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	entry := RegistryEntry{
		ID:     "GU0001",
		Name:   "guild-usable",
		Status: "active",
		Path:   "/festivals/active/guild-usable_GU0001",
	}
	_ = reg.Add(ctx, entry)

	// Delete entry
	if err := reg.Delete(ctx, "GU0001"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify deletion
	if reg.Exists(ctx, "GU0001") {
		t.Error("Entry should not exist after Delete")
	}

	// Test delete non-existent
	if err := reg.Delete(ctx, "XX9999"); err == nil {
		t.Error("Delete() should return error for non-existent ID")
	}
}

func TestRegistry_Exists(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	entry := RegistryEntry{ID: "GU0001", Name: "test", Status: "active"}
	_ = reg.Add(ctx, entry)

	tests := []struct {
		id   string
		want bool
	}{
		{"GU0001", true},
		{"XX9999", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			if got := reg.Exists(ctx, tt.id); got != tt.want {
				t.Errorf("Exists(%q) = %v, want %v", tt.id, got, tt.want)
			}
		})
	}
}

func TestRegistry_List(t *testing.T) {
	ctx := context.Background()
	reg := New("/tmp/test-registry.yaml")

	// Empty list
	list := reg.List(ctx)
	if len(list) != 0 {
		t.Errorf("List() length = %d, want 0 for empty registry", len(list))
	}

	// Add entries
	entries := []RegistryEntry{
		{ID: "GU0001", Name: "first", Status: "active"},
		{ID: "GU0002", Name: "second", Status: "active"},
		{ID: "FN0001", Name: "fest-node", Status: "active"},
	}
	for _, e := range entries {
		_ = reg.Add(ctx, e)
	}

	list = reg.List(ctx)
	if len(list) != 3 {
		t.Errorf("List() length = %d, want 3", len(list))
	}
}

func TestRegistry_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	reg := New("/tmp/test-registry.yaml")
	entry := RegistryEntry{ID: "GU0001", Name: "test", Status: "active"}

	// All operations should fail with cancelled context
	if err := reg.Add(ctx, entry); err == nil {
		t.Error("Add() should return error with cancelled context")
	}
	if _, err := reg.Get(ctx, "GU0001"); err == nil {
		t.Error("Get() should return error with cancelled context")
	}
	if err := reg.Update(ctx, entry); err == nil {
		t.Error("Update() should return error with cancelled context")
	}
	if err := reg.Delete(ctx, "GU0001"); err == nil {
		t.Error("Delete() should return error with cancelled context")
	}
	if reg.Exists(ctx, "GU0001") {
		t.Error("Exists() should return false with cancelled context")
	}
	if reg.Count(ctx) != 0 {
		t.Error("Count() should return 0 with cancelled context")
	}
	if reg.List(ctx) != nil {
		t.Error("List() should return nil with cancelled context")
	}
	if reg.ByStatus(ctx, "active") != nil {
		t.Error("ByStatus() should return nil with cancelled context")
	}
}

func TestLoad_NewRegistry(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "id_registry.yaml")

	// Load non-existent file should return new registry
	reg, err := Load(ctx, path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if reg == nil {
		t.Fatal("Load() returned nil registry")
	}
	if reg.Version != CurrentVersion {
		t.Errorf("Version = %q, want %q", reg.Version, CurrentVersion)
	}
}

func TestRegistry_SaveLoad(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "id_registry.yaml")

	// Create and save registry
	reg := New(path)
	entry := RegistryEntry{
		ID:        "GU0001",
		Name:      "guild-usable",
		Status:    "active",
		Path:      "/festivals/active/guild-usable_GU0001",
		CreatedAt: time.Now().Truncate(time.Second),
		UpdatedAt: time.Now().Truncate(time.Second),
	}
	_ = reg.Add(ctx, entry)

	if err := reg.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("Registry file not created")
	}

	// Load and verify
	loaded, err := Load(ctx, path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	got, err := loaded.Get(ctx, "GU0001")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Name != entry.Name {
		t.Errorf("Name = %q, want %q", got.Name, entry.Name)
	}
	if got.Status != entry.Status {
		t.Errorf("Status = %q, want %q", got.Status, entry.Status)
	}
}

func TestRegistry_AtomicSave(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "id_registry.yaml")

	reg := New(path)
	_ = reg.Add(ctx, RegistryEntry{ID: "GU0001", Name: "test", Status: "active"})

	// Save should use atomic write
	if err := reg.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify no temp file remains
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("Temp file should not exist after successful save")
	}
}

func TestLoad_CorruptedFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "id_registry.yaml")

	// Write corrupted YAML
	if err := os.WriteFile(path, []byte("not: valid: yaml: {{{{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := Load(ctx, path)
	if err == nil {
		t.Error("Load() should return error for corrupted file")
	}
}

func TestLoad_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := Load(ctx, "/any/path")
	if err == nil {
		t.Error("Load() should return error with cancelled context")
	}
}

// Benchmarks for performance-critical operations

// setupBenchmarkRegistry creates a registry with n entries for benchmarking
func setupBenchmarkRegistry(n int) *Registry {
	ctx := context.Background()
	reg := New("/tmp/benchmark-registry.yaml")

	for i := 0; i < n; i++ {
		entry := RegistryEntry{
			ID:     generateTestID(i),
			Name:   "test-festival",
			Status: StatusActive,
			Path:   "/festivals/active/test",
		}
		_ = reg.Add(ctx, entry)
	}

	return reg
}

// generateTestID creates a test ID like TE0001, TE0002, etc.
func generateTestID(n int) string {
	return fmt.Sprintf("TE%04d", n+1)
}

func BenchmarkRegistry_Get(b *testing.B) {
	reg := setupBenchmarkRegistry(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = reg.Get(ctx, "TE0500")
	}
}

func BenchmarkRegistry_Add(b *testing.B) {
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reg := New("/tmp/benchmark-registry.yaml")
		_ = reg.Add(ctx, RegistryEntry{
			ID:     generateTestID(i),
			Name:   "test",
			Status: StatusActive,
		})
	}
}

func BenchmarkRegistry_Exists(b *testing.B) {
	reg := setupBenchmarkRegistry(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.Exists(ctx, "TE0500")
	}
}

func BenchmarkRegistry_List(b *testing.B) {
	reg := setupBenchmarkRegistry(100)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.List(ctx)
	}
}

func BenchmarkRegistry_ByStatus(b *testing.B) {
	reg := setupBenchmarkRegistry(1000)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = reg.ByStatus(ctx, StatusActive)
	}
}
