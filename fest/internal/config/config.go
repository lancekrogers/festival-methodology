package config

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// File permission constants (kept local to avoid import cycles with registry)
const (
	dirPermissions  os.FileMode = 0755
	filePermissions os.FileMode = 0644
)

const (
	// DefaultRepositoryURL is the default festival methodology repository
	DefaultRepositoryURL = "https://github.com/lancekrogers/festival-methodology"

	// DefaultBranch is the default git branch
	DefaultBranch = "main"

	// ConfigFileName is the name of the config file
	ConfigFileName = "config.json"
)

// Config represents the fest configuration
type Config struct {
	Version    string     `json:"version"`
	Repository Repository `json:"repository"`
	Local      Local      `json:"local"`
	Behavior   Behavior   `json:"behavior"`
	Network    Network    `json:"network"`
	LastSync   string     `json:"last_sync,omitempty"`
}

// Repository contains repository information
type Repository struct {
	URL    string `json:"url"`
	Branch string `json:"branch"`
	Path   string `json:"path"`
}

// Local contains local path configuration
type Local struct {
	CacheDir     string `json:"cache_dir"`
	BackupDir    string `json:"backup_dir"`
	ChecksumFile string `json:"checksum_file"`
}

// Behavior contains behavior configuration
type Behavior struct {
	AutoBackup  bool `json:"auto_backup"`
	Interactive bool `json:"interactive"`
	UseColor    bool `json:"use_color"`
	Verbose     bool `json:"verbose"`
}

// Network contains network configuration
type Network struct {
	Timeout    int `json:"timeout"`
	RetryCount int `json:"retry_count"`
	RetryDelay int `json:"retry_delay"`
}

// ConfigDir returns the configuration directory path
func ConfigDir() string {
	if dir := os.Getenv("FEST_CONFIG_DIR"); dir != "" {
		return dir
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return ".fest"
	}

	return filepath.Join(home, ".config", "fest")
}

// Load loads configuration from file
func Load(ctx context.Context, customPath string) (*Config, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	configPath := customPath
	if configPath == "" {
		configPath = filepath.Join(ConfigDir(), ConfigFileName)
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.IO("reading config", err).WithField("path", configPath)
	}

	// Parse JSON
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Parse("parsing config", err).WithField("path", configPath)
	}

	// Apply defaults for missing values
	applyDefaults(&cfg)

	return &cfg, nil
}

// Save saves configuration to file
func Save(ctx context.Context, cfg *Config) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	configDir := ConfigDir()
	configPath := filepath.Join(configDir, ConfigFileName)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(configDir, dirPermissions); err != nil {
		return errors.IO("creating config directory", err).WithField("path", configDir)
	}

	// Marshal to JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling config")
	}

	// Write file
	if err := os.WriteFile(configPath, data, filePermissions); err != nil {
		return errors.IO("writing config", err).WithField("path", configPath)
	}

	return nil
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		Version: "1.0.0",
		Repository: Repository{
			URL:    DefaultRepositoryURL,
			Branch: DefaultBranch,
			Path:   "festivals",
		},
		Local: Local{
			CacheDir:     filepath.Join(ConfigDir(), "cache"),
			BackupDir:    ".fest-backup",
			ChecksumFile: ".fest-checksums.json",
		},
		Behavior: Behavior{
			AutoBackup:  false,
			Interactive: true,
			UseColor:    true,
			Verbose:     false,
		},
		Network: Network{
			Timeout:    30,
			RetryCount: 3,
			RetryDelay: 1,
		},
	}
}

// applyDefaults applies default values to missing configuration fields
func applyDefaults(cfg *Config) {
	defaults := DefaultConfig()

	if cfg.Version == "" {
		cfg.Version = defaults.Version
	}

	if cfg.Repository.URL == "" {
		cfg.Repository.URL = defaults.Repository.URL
	}

	if cfg.Repository.Branch == "" {
		cfg.Repository.Branch = defaults.Repository.Branch
	}

	if cfg.Repository.Path == "" {
		cfg.Repository.Path = defaults.Repository.Path
	}

	if cfg.Local.CacheDir == "" {
		cfg.Local.CacheDir = defaults.Local.CacheDir
	}

	if cfg.Local.BackupDir == "" {
		cfg.Local.BackupDir = defaults.Local.BackupDir
	}

	if cfg.Local.ChecksumFile == "" {
		cfg.Local.ChecksumFile = defaults.Local.ChecksumFile
	}

	if cfg.Network.Timeout == 0 {
		cfg.Network.Timeout = defaults.Network.Timeout
	}

	if cfg.Network.RetryCount == 0 {
		cfg.Network.RetryCount = defaults.Network.RetryCount
	}

	if cfg.Network.RetryDelay == 0 {
		cfg.Network.RetryDelay = defaults.Network.RetryDelay
	}
}
