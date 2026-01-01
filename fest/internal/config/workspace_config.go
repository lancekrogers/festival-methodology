package config

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	// WorkspaceConfigFileName is the name of the workspace-level config file
	WorkspaceConfigFileName = "config.yaml"
	// DotFestivalDir is the hidden directory containing workspace config
	DotFestivalDir = ".festival"
)

// WorkspaceConfig represents workspace-level configuration in .festival/config.yaml
type WorkspaceConfig struct {
	Version string      `yaml:"version"`
	Agent   AgentConfig `yaml:"agent"`
}

// AgentConfig controls AI agent behavior for the workspace
type AgentConfig struct {
	// StrictMode is a master switch that enables all strict behaviors
	StrictMode bool `yaml:"strict_mode"`
	// DisableSkipMarkers prevents --skip-markers from being used
	DisableSkipMarkers bool `yaml:"disable_skip_markers"`
	// RequireValidation forces validation after all create operations
	RequireValidation bool `yaml:"require_validation"`
	// BlockOnErrors prevents completion if validation has errors
	BlockOnErrors bool `yaml:"block_on_errors"`
}

// LoadWorkspaceConfig loads workspace configuration from .festival/config.yaml
func LoadWorkspaceConfig(festivalsRoot string) (*WorkspaceConfig, error) {
	configPath := filepath.Join(festivalsRoot, DotFestivalDir, WorkspaceConfigFileName)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultWorkspaceConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.IO("reading workspace config", err).WithField("path", configPath)
	}

	// Parse YAML
	var cfg WorkspaceConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Parse("parsing workspace config", err).WithField("path", configPath)
	}

	// Apply defaults for missing values
	applyWorkspaceDefaults(&cfg)

	return &cfg, nil
}

// SaveWorkspaceConfig saves workspace configuration to .festival/config.yaml
func SaveWorkspaceConfig(festivalsRoot string, cfg *WorkspaceConfig) error {
	dotFestivalPath := filepath.Join(festivalsRoot, DotFestivalDir)

	// Ensure .festival directory exists
	if err := os.MkdirAll(dotFestivalPath, dirPermissions); err != nil {
		return errors.IO("creating .festival directory", err).WithField("path", dotFestivalPath)
	}

	configPath := filepath.Join(dotFestivalPath, WorkspaceConfigFileName)

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return errors.Wrap(err, "marshaling workspace config")
	}

	// Write file
	if err := os.WriteFile(configPath, data, filePermissions); err != nil {
		return errors.IO("writing workspace config", err).WithField("path", configPath)
	}

	return nil
}

// DefaultWorkspaceConfig returns the default workspace configuration
// All agent settings are disabled by default for backward compatibility
func DefaultWorkspaceConfig() *WorkspaceConfig {
	return &WorkspaceConfig{
		Version: "1.0",
		Agent: AgentConfig{
			StrictMode:         false,
			DisableSkipMarkers: false,
			RequireValidation:  false,
			BlockOnErrors:      false,
		},
	}
}

// applyWorkspaceDefaults applies default values to missing configuration fields
func applyWorkspaceDefaults(cfg *WorkspaceConfig) {
	if cfg.Version == "" {
		cfg.Version = "1.0"
	}
	// Agent config fields default to false (zero value) which is correct
}

// WorkspaceConfigExists checks if a config.yaml file exists in .festival/
func WorkspaceConfigExists(festivalsRoot string) bool {
	configPath := filepath.Join(festivalsRoot, DotFestivalDir, WorkspaceConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}

// MergeAgentConfig merges workspace and festival agent configs
// Festival config takes precedence over workspace config
func MergeAgentConfig(workspace, festival *AgentConfig) *AgentConfig {
	if workspace == nil && festival == nil {
		return &AgentConfig{}
	}
	if workspace == nil {
		return festival
	}
	if festival == nil {
		return workspace
	}

	// Start with workspace settings
	merged := &AgentConfig{
		StrictMode:         workspace.StrictMode,
		DisableSkipMarkers: workspace.DisableSkipMarkers,
		RequireValidation:  workspace.RequireValidation,
		BlockOnErrors:      workspace.BlockOnErrors,
	}

	// Festival settings override workspace settings when set to true
	// (more restrictive settings win)
	if festival.StrictMode {
		merged.StrictMode = true
	}
	if festival.DisableSkipMarkers {
		merged.DisableSkipMarkers = true
	}
	if festival.RequireValidation {
		merged.RequireValidation = true
	}
	if festival.BlockOnErrors {
		merged.BlockOnErrors = true
	}

	return merged
}

// EffectiveSkipMarkers determines if markers should be skipped based on config and flags
func EffectiveSkipMarkers(cfg *AgentConfig, agentMode, skipFlag bool) bool {
	if cfg == nil {
		cfg = &AgentConfig{}
	}

	// Agent mode always processes markers
	if agentMode {
		return false
	}

	// Strict mode always processes markers
	if cfg.StrictMode {
		return false
	}

	// Config can disable skip-markers
	if cfg.DisableSkipMarkers {
		return false
	}

	return skipFlag
}

// ShouldValidate determines if validation should run after creation
func ShouldValidate(cfg *AgentConfig, agentMode bool) bool {
	if cfg == nil {
		cfg = &AgentConfig{}
	}

	// Agent mode always validates
	if agentMode {
		return true
	}

	// Strict mode always validates
	if cfg.StrictMode {
		return true
	}

	return cfg.RequireValidation
}

// ShouldBlockOnErrors determines if errors should block completion
func ShouldBlockOnErrors(cfg *AgentConfig, agentMode bool) bool {
	if cfg == nil {
		cfg = &AgentConfig{}
	}

	// Agent mode always blocks on errors
	if agentMode {
		return true
	}

	// Strict mode always blocks on errors
	if cfg.StrictMode {
		return true
	}

	return cfg.BlockOnErrors
}
