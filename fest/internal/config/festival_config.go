package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	// FestivalConfigFileName is the name of the festival-level config file
	FestivalConfigFileName = "fest.yaml"
)

// FestivalConfig represents per-festival configuration
type FestivalConfig struct {
	Version          string             `yaml:"version"`
	QualityGates     QualityGatesConfig `yaml:"quality_gates"`
	ExcludedPatterns []string           `yaml:"excluded_patterns"`
	Templates        TemplatePrefs      `yaml:"templates"`
	Tracking         TrackingConfig     `yaml:"tracking"`
}

// QualityGatesConfig contains quality gate settings
type QualityGatesConfig struct {
	Enabled    bool              `yaml:"enabled"`
	AutoAppend bool              `yaml:"auto_append"`
	Tasks      []QualityGateTask `yaml:"tasks"`
}

// QualityGateTask represents a single quality gate task configuration
type QualityGateTask struct {
	ID             string                 `yaml:"id"`
	Template       string                 `yaml:"template"`
	Name           string                 `yaml:"name,omitempty"`
	Enabled        bool                   `yaml:"enabled"`
	Customizations map[string]interface{} `yaml:"customizations,omitempty"`
}

// TemplatePrefs contains template preference settings
type TemplatePrefs struct {
	TaskDefault  string `yaml:"task_default"`
	PreferSimple bool   `yaml:"prefer_simple"`
}

// TrackingConfig contains file tracking settings
type TrackingConfig struct {
	Enabled      bool   `yaml:"enabled"`
	ChecksumFile string `yaml:"checksum_file"`
}

// LoadFestivalConfig loads festival configuration from fest.yaml
func LoadFestivalConfig(festivalPath string) (*FestivalConfig, error) {
	configPath := filepath.Join(festivalPath, FestivalConfigFileName)

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Return default config if file doesn't exist
		return DefaultFestivalConfig(), nil
	}

	// Read file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read festival config: %w", err)
	}

	// Parse YAML
	var cfg FestivalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse festival config: %w", err)
	}

	// Apply defaults for missing values
	applyFestivalDefaults(&cfg)

	return &cfg, nil
}

// SaveFestivalConfig saves festival configuration to fest.yaml
func SaveFestivalConfig(festivalPath string, cfg *FestivalConfig) error {
	configPath := filepath.Join(festivalPath, FestivalConfigFileName)

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal festival config: %w", err)
	}

	// Write file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write festival config: %w", err)
	}

	return nil
}

// DefaultFestivalConfig returns the default festival configuration
func DefaultFestivalConfig() *FestivalConfig {
	return &FestivalConfig{
		Version: "1.0",
		QualityGates: QualityGatesConfig{
			Enabled:    true,
			AutoAppend: true,
			Tasks: []QualityGateTask{
				{
					ID:       "testing_and_verify",
					Template: "QUALITY_GATE_TESTING",
					Name:     "Testing and Verification",
					Enabled:  true,
				},
				{
					ID:       "code_review",
					Template: "QUALITY_GATE_REVIEW",
					Name:     "Code Review",
					Enabled:  true,
				},
				{
					ID:       "review_results_iterate",
					Template: "QUALITY_GATE_ITERATE",
					Name:     "Review Results and Iterate",
					Enabled:  true,
				},
			},
		},
		ExcludedPatterns: []string{
			"*_planning",
			"*_research",
			"*_requirements",
		},
		Templates: TemplatePrefs{
			TaskDefault:  "TASK_TEMPLATE_SIMPLE",
			PreferSimple: true,
		},
		Tracking: TrackingConfig{
			Enabled:      true,
			ChecksumFile: ".festival-checksums.json",
		},
	}
}

// applyFestivalDefaults applies default values to missing configuration fields
func applyFestivalDefaults(cfg *FestivalConfig) {
	defaults := DefaultFestivalConfig()

	if cfg.Version == "" {
		cfg.Version = defaults.Version
	}

	// If no tasks defined, use defaults
	if len(cfg.QualityGates.Tasks) == 0 {
		cfg.QualityGates.Tasks = defaults.QualityGates.Tasks
	}

	// If no excluded patterns, use defaults
	if len(cfg.ExcludedPatterns) == 0 {
		cfg.ExcludedPatterns = defaults.ExcludedPatterns
	}

	if cfg.Templates.TaskDefault == "" {
		cfg.Templates.TaskDefault = defaults.Templates.TaskDefault
	}

	if cfg.Tracking.ChecksumFile == "" {
		cfg.Tracking.ChecksumFile = defaults.Tracking.ChecksumFile
	}
}

// IsSequenceExcluded checks if a sequence name matches any excluded pattern
func (cfg *FestivalConfig) IsSequenceExcluded(sequenceName string) bool {
	for _, pattern := range cfg.ExcludedPatterns {
		matched, err := filepath.Match(pattern, sequenceName)
		if err != nil {
			continue // Skip invalid patterns
		}
		if matched {
			return true
		}
	}
	return false
}

// GetEnabledTasks returns only enabled quality gate tasks
func (cfg *FestivalConfig) GetEnabledTasks() []QualityGateTask {
	var enabled []QualityGateTask
	for _, task := range cfg.QualityGates.Tasks {
		if task.Enabled {
			enabled = append(enabled, task)
		}
	}
	return enabled
}

// FestivalConfigExists checks if a fest.yaml file exists in the given path
func FestivalConfigExists(festivalPath string) bool {
	configPath := filepath.Join(festivalPath, FestivalConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}
