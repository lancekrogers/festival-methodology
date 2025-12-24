package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Version == "" {
		t.Error("Version is empty")
	}

	if cfg.Repository.URL == "" {
		t.Error("Repository URL is empty")
	}

	if cfg.Repository.Branch != "main" {
		t.Errorf("Expected branch 'main', got %s", cfg.Repository.Branch)
	}

	if cfg.Network.Timeout != 30 {
		t.Errorf("Expected timeout 30, got %d", cfg.Network.Timeout)
	}

	if cfg.Network.RetryCount != 3 {
		t.Errorf("Expected retry count 3, got %d", cfg.Network.RetryCount)
	}
}

func TestLoadNonExisting(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "non-existing.json")

	cfg, err := Load(ctx, configPath)
	if err != nil {
		t.Fatalf("Load failed for non-existing file: %v", err)
	}

	// Should return default config
	defaults := DefaultConfig()
	if cfg.Version != defaults.Version {
		t.Error("Did not return default config for non-existing file")
	}
}

func TestSaveAndLoad(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	os.Setenv("FEST_CONFIG_DIR", tmpDir)
	defer os.Unsetenv("FEST_CONFIG_DIR")

	// Create config
	original := &Config{
		Version: "test-1.0",
		Repository: Repository{
			URL:    "https://github.com/test/repo",
			Branch: "develop",
			Path:   "custom/path",
		},
		Local: Local{
			CacheDir:     "/tmp/cache",
			BackupDir:    "/tmp/backup",
			ChecksumFile: "checksums.json",
		},
		Behavior: Behavior{
			AutoBackup:  true,
			Interactive: false,
			UseColor:    false,
			Verbose:     true,
		},
		Network: Network{
			Timeout:    60,
			RetryCount: 5,
			RetryDelay: 2,
		},
		LastSync: "2024-01-01T00:00:00Z",
	}

	// Save config
	if err := Save(ctx, original); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Load config
	loaded, err := Load(ctx, "")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	// Compare
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: %s != %s", loaded.Version, original.Version)
	}

	if loaded.Repository.URL != original.Repository.URL {
		t.Errorf("Repository URL mismatch")
	}

	if loaded.Repository.Branch != original.Repository.Branch {
		t.Errorf("Branch mismatch")
	}

	if loaded.Network.Timeout != original.Network.Timeout {
		t.Errorf("Timeout mismatch")
	}

	if loaded.Behavior.AutoBackup != original.Behavior.AutoBackup {
		t.Errorf("AutoBackup mismatch")
	}

	if loaded.LastSync != original.LastSync {
		t.Errorf("LastSync mismatch")
	}
}

func TestApplyDefaults(t *testing.T) {
	// Create partial config
	cfg := &Config{
		Version: "custom",
		Repository: Repository{
			URL: "https://custom.url",
		},
	}

	// Apply defaults
	applyDefaults(cfg)

	// Check that custom values are preserved
	if cfg.Version != "custom" {
		t.Error("Custom version was overwritten")
	}

	if cfg.Repository.URL != "https://custom.url" {
		t.Error("Custom URL was overwritten")
	}

	// Check that defaults were applied
	if cfg.Repository.Branch != DefaultBranch {
		t.Error("Default branch was not applied")
	}

	if cfg.Network.Timeout != 30 {
		t.Error("Default timeout was not applied")
	}
}

func TestConfigDir(t *testing.T) {
	// Test with environment variable
	testDir := "/custom/dir"
	os.Setenv("FEST_CONFIG_DIR", testDir)
	defer os.Unsetenv("FEST_CONFIG_DIR")

	dir := ConfigDir()
	if dir != testDir {
		t.Errorf("Expected %s, got %s", testDir, dir)
	}

	// Test without environment variable
	os.Unsetenv("FEST_CONFIG_DIR")
	dir = ConfigDir()
	if dir == "" {
		t.Error("ConfigDir returned empty string")
	}
}

func TestJSONMarshaling(t *testing.T) {
	cfg := DefaultConfig()

	// Marshal
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal: %v", err)
	}

	// Unmarshal
	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	// Compare key fields
	if loaded.Version != cfg.Version {
		t.Error("Version mismatch after marshal/unmarshal")
	}

	if loaded.Repository.URL != cfg.Repository.URL {
		t.Error("Repository URL mismatch after marshal/unmarshal")
	}
}
