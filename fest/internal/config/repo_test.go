package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsGitURL(t *testing.T) {
	tests := []struct {
		source   string
		expected bool
	}{
		{"https://github.com/user/repo", true},
		{"git@github.com:user/repo.git", true},
		{"http://github.com/user/repo", true},
		{"ssh://git@github.com/user/repo", true},
		{"/absolute/path", false},
		{"./relative/path", false},
		{"../parent/path", false},
		{"~/home/path", false},
		{"", false},
	}

	for _, tc := range tests {
		result := IsGitURL(tc.source)
		if result != tc.expected {
			t.Errorf("IsGitURL(%q) = %v, want %v", tc.source, result, tc.expected)
		}
	}
}

func TestConfigRepoManifest(t *testing.T) {
	manifest := &ConfigRepoManifest{
		Version: 1,
		Repos:   []ConfigRepo{},
	}

	// Test AddRepo
	repo1 := ConfigRepo{
		Name:      "test1",
		Source:    "/path/to/test1",
		LocalPath: "/local/test1",
		IsGitRepo: false,
	}
	manifest.AddRepo(repo1)

	if len(manifest.Repos) != 1 {
		t.Errorf("AddRepo: expected 1 repo, got %d", len(manifest.Repos))
	}

	// Test GetRepo
	got := manifest.GetRepo("test1")
	if got == nil {
		t.Error("GetRepo: expected to find test1, got nil")
	}
	if got != nil && got.Name != "test1" {
		t.Errorf("GetRepo: expected name 'test1', got %q", got.Name)
	}

	// Test GetRepo for non-existent
	got = manifest.GetRepo("nonexistent")
	if got != nil {
		t.Error("GetRepo: expected nil for nonexistent repo")
	}

	// Test AddRepo update (same name)
	repo1Updated := ConfigRepo{
		Name:      "test1",
		Source:    "/new/path/to/test1",
		LocalPath: "/local/test1",
		IsGitRepo: false,
	}
	manifest.AddRepo(repo1Updated)
	if len(manifest.Repos) != 1 {
		t.Errorf("AddRepo update: expected 1 repo, got %d", len(manifest.Repos))
	}
	got = manifest.GetRepo("test1")
	if got.Source != "/new/path/to/test1" {
		t.Errorf("AddRepo update: expected updated source, got %q", got.Source)
	}

	// Test RemoveRepo
	removed := manifest.RemoveRepo("test1")
	if !removed {
		t.Error("RemoveRepo: expected true, got false")
	}
	if len(manifest.Repos) != 0 {
		t.Errorf("RemoveRepo: expected 0 repos, got %d", len(manifest.Repos))
	}

	// Test RemoveRepo non-existent
	removed = manifest.RemoveRepo("nonexistent")
	if removed {
		t.Error("RemoveRepo: expected false for nonexistent, got true")
	}
}

func TestConfigRepoManifestActive(t *testing.T) {
	manifest := &ConfigRepoManifest{
		Version: 1,
		Active:  "test1",
		Repos: []ConfigRepo{
			{Name: "test1", Source: "/path1"},
			{Name: "test2", Source: "/path2"},
		},
	}

	// Test GetActiveRepo
	active := manifest.GetActiveRepo()
	if active == nil {
		t.Error("GetActiveRepo: expected repo, got nil")
	}
	if active != nil && active.Name != "test1" {
		t.Errorf("GetActiveRepo: expected 'test1', got %q", active.Name)
	}

	// Test GetActiveRepo with no active
	manifest.Active = ""
	active = manifest.GetActiveRepo()
	if active != nil {
		t.Error("GetActiveRepo: expected nil when no active, got repo")
	}

	// Test GetActiveRepo with invalid active
	manifest.Active = "nonexistent"
	active = manifest.GetActiveRepo()
	if active != nil {
		t.Error("GetActiveRepo: expected nil for invalid active name")
	}
}

func TestRepoLocalPath(t *testing.T) {
	path := RepoLocalPath("myrepo")
	expected := filepath.Join(ConfigReposPath(), "myrepo")
	if path != expected {
		t.Errorf("RepoLocalPath: expected %q, got %q", expected, path)
	}
}

func TestLoadSaveRepoManifest(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "fest-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Override config dir
	os.Setenv("FEST_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("FEST_CONFIG_DIR")

	// Test loading non-existent (should return empty manifest)
	manifest, err := LoadRepoManifest()
	if err != nil {
		t.Errorf("LoadRepoManifest: unexpected error %v", err)
	}
	if manifest.Version != 1 {
		t.Errorf("LoadRepoManifest: expected version 1, got %d", manifest.Version)
	}
	if len(manifest.Repos) != 0 {
		t.Errorf("LoadRepoManifest: expected 0 repos, got %d", len(manifest.Repos))
	}

	// Add repo and save
	manifest.AddRepo(ConfigRepo{
		Name:      "test",
		Source:    "https://github.com/test/repo",
		LocalPath: "/local/test",
		IsGitRepo: true,
	})
	manifest.Active = "test"

	if err := SaveRepoManifest(manifest); err != nil {
		t.Errorf("SaveRepoManifest: unexpected error %v", err)
	}

	// Load again and verify
	loaded, err := LoadRepoManifest()
	if err != nil {
		t.Errorf("LoadRepoManifest after save: unexpected error %v", err)
	}
	if loaded.Active != "test" {
		t.Errorf("LoadRepoManifest: expected active 'test', got %q", loaded.Active)
	}
	if len(loaded.Repos) != 1 {
		t.Errorf("LoadRepoManifest: expected 1 repo, got %d", len(loaded.Repos))
	}
	if loaded.Repos[0].Name != "test" {
		t.Errorf("LoadRepoManifest: expected repo name 'test', got %q", loaded.Repos[0].Name)
	}
}
