package config

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	// FestivalConfigFileName is the name of the festival-level config file
	FestivalConfigFileName = "fest.yaml"
)

// FestivalConfig represents per-festival configuration
type FestivalConfig struct {
	Version          string             `yaml:"version"`
	Metadata         FestivalMetadata   `yaml:"metadata,omitempty"`
	ProjectPath      string             `yaml:"project_path,omitempty"` // Path to linked project directory
	QualityGates     QualityGatesConfig `yaml:"quality_gates"`
	ExcludedPatterns []string           `yaml:"excluded_patterns"`
	Templates        TemplatePrefs      `yaml:"templates"`
	Tracking         TrackingConfig     `yaml:"tracking"`
	Agent            AgentConfig        `yaml:"agent,omitempty"`
}

// QualityGatesConfig contains quality gate settings
type QualityGatesConfig struct {
	Enabled    bool                         `yaml:"enabled"`
	AutoAppend bool                         `yaml:"auto_append"`
	Tasks      []QualityGateTask            `yaml:"tasks"`                 // Legacy: implementation gates only
	PhaseGates map[string][]QualityGateTask `yaml:"phase_gates,omitempty"` // Per phase type gate ordering
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
		return nil, errors.IO("reading festival config", err).WithField("path", configPath)
	}

	// Parse YAML
	var cfg FestivalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, errors.Parse("parsing festival config", err).WithField("path", configPath)
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
		return errors.Wrap(err, "marshaling festival config")
	}

	// Write file
	if err := os.WriteFile(configPath, data, filePermissions); err != nil {
		return errors.IO("writing festival config", err).WithField("path", configPath)
	}

	return nil
}

// DefaultFestivalConfig returns the default festival configuration.
// Note: Template paths reference the templates directory structure.
func DefaultFestivalConfig() *FestivalConfig {
	return &FestivalConfig{
		Version: "1.0",
		QualityGates: QualityGatesConfig{
			Enabled:    true,
			AutoAppend: true,
			// Legacy Tasks field for backwards compatibility (implementation only)
			// Note: IDs should match template filenames (without .md extension)
			Tasks: []QualityGateTask{
				{ID: "QUALITY_GATE_TESTING", Template: "phases/implementation/gates/QUALITY_GATE_TESTING", Name: "Testing and Verification", Enabled: true},
				{ID: "QUALITY_GATE_REVIEW", Template: "phases/implementation/gates/QUALITY_GATE_REVIEW", Name: "Code Review", Enabled: true},
				{ID: "QUALITY_GATE_ITERATE", Template: "phases/implementation/gates/QUALITY_GATE_ITERATE", Name: "Review Results and Iterate", Enabled: true},
				{ID: "QUALITY_GATE_COMMIT", Template: "phases/implementation/gates/QUALITY_GATE_COMMIT", Name: "Commit Changes", Enabled: true},
			},
			// PhaseGates specifies gate ordering per phase type
			// Note: IDs should match template filenames (without .md extension)
			PhaseGates: map[string][]QualityGateTask{
				"implementation": {
					{ID: "QUALITY_GATE_TESTING", Template: "phases/implementation/gates/QUALITY_GATE_TESTING", Name: "Testing and Verification", Enabled: true},
					{ID: "QUALITY_GATE_REVIEW", Template: "phases/implementation/gates/QUALITY_GATE_REVIEW", Name: "Code Review", Enabled: true},
					{ID: "QUALITY_GATE_ITERATE", Template: "phases/implementation/gates/QUALITY_GATE_ITERATE", Name: "Review Results and Iterate", Enabled: true},
					{ID: "QUALITY_GATE_COMMIT", Template: "phases/implementation/gates/QUALITY_GATE_COMMIT", Name: "Commit Changes", Enabled: true},
				},
				"planning": {
					{ID: "plan_review", Template: "phases/planning/gates/plan_review", Name: "Planning Review", Enabled: true},
					{ID: "decision_validation", Template: "phases/planning/gates/decision_validation", Name: "Decision Validation", Enabled: true},
					{ID: "approval", Template: "phases/planning/gates/approval", Name: "Planning Approval", Enabled: true},
				},
				"research": {
					{ID: "findings_review", Template: "phases/research/gates/findings_review", Name: "Findings Review", Enabled: true},
					{ID: "documentation", Template: "phases/research/gates/documentation", Name: "Documentation", Enabled: true},
					{ID: "summary", Template: "phases/research/gates/summary", Name: "Research Summary", Enabled: true},
				},
				"review": {
					{ID: "checklist", Template: "phases/review/gates/checklist", Name: "Review Checklist", Enabled: true},
					{ID: "sign_off", Template: "phases/review/gates/sign_off", Name: "Sign-off", Enabled: true},
				},
				"non_coding_action": {
					{ID: "action_verify", Template: "phases/non_coding_action/gates/action_verify", Name: "Execution & Verify", Enabled: true},
					{ID: "completion", Template: "phases/non_coding_action/gates/completion", Name: "Completion", Enabled: true},
				},
			},
		},
		ExcludedPatterns: []string{
			"*_planning",
			"*_research",
			"*_requirements",
			"*_docs",
		},
		Templates: TemplatePrefs{
			TaskDefault:  "tasks/SIMPLE",
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

	// If no phase gates defined, use defaults
	if cfg.QualityGates.PhaseGates == nil {
		cfg.QualityGates.PhaseGates = defaults.QualityGates.PhaseGates
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

// GetGatesForPhaseType returns configured gates for a phase type in order.
// Falls back to Tasks (implementation) for backwards compatibility if PhaseGates is not configured.
func (cfg *FestivalConfig) GetGatesForPhaseType(phaseType string) []QualityGateTask {
	// Check PhaseGates map first (preferred)
	if cfg.QualityGates.PhaseGates != nil {
		if gates, ok := cfg.QualityGates.PhaseGates[phaseType]; ok && len(gates) > 0 {
			// Return only enabled gates
			var enabled []QualityGateTask
			for _, gate := range gates {
				if gate.Enabled {
					enabled = append(enabled, gate)
				}
			}
			return enabled
		}
	}

	// Fallback for backwards compatibility: use Tasks for implementation phase
	if phaseType == "implementation" && len(cfg.QualityGates.Tasks) > 0 {
		return cfg.GetEnabledTasks()
	}

	return nil
}

// FestivalConfigExists checks if a fest.yaml file exists in the given path
func FestivalConfigExists(festivalPath string) bool {
	configPath := filepath.Join(festivalPath, FestivalConfigFileName)
	_, err := os.Stat(configPath)
	return err == nil
}
