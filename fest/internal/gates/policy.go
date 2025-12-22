// Package gates provides quality gate policy management with phase-level overrides.
package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// HierarchicalPolicyLoader loads effective policies from the hierarchy.
type HierarchicalPolicyLoader interface {
	LoadForSequence(ctx context.Context, festivalPath, phasePath, sequencePath string) (*EffectivePolicy, error)
	LoadForPhase(ctx context.Context, festivalPath, phasePath string) (*EffectivePolicy, error)
	LoadForFestival(ctx context.Context, festivalPath string) (*EffectivePolicy, error)
}

// HierarchicalTemplateResolver resolves gate templates from the hierarchy.
type HierarchicalTemplateResolver interface {
	Resolve(templateID, festivalPath, phasePath, sequencePath string) (*ResolveResult, error)
	ResolveForPhase(templateID, festivalPath, phasePath string) (*ResolveResult, error)
	ResolveForFestival(templateID, festivalPath string) (*ResolveResult, error)
	ClearCache()
}

// PolicyRegistrar provides access to named policies.
type PolicyRegistrar interface {
	Get(name string) (*PolicyInfo, bool)
	GetPolicy(name string) (*GatePolicy, error)
	List() []string
	ListInfo() []*PolicyInfo
	Refresh()
}

const (
	// PolicyFileName is the name of the gate policy file
	PolicyFileName = "policy.yml"
	// PhaseOverrideFileName is the name of the phase-level override file
	PhaseOverrideFileName = ".fest.gates.yml"
	// DefaultPolicyName is the name of the default policy
	DefaultPolicyName = "default"
)

// PolicyLevel represents the hierarchical level where a policy is defined
type PolicyLevel string

const (
	// PolicyLevelBuiltin is the built-in default level
	PolicyLevelBuiltin PolicyLevel = "builtin"
	// PolicyLevelGlobal is the global festivals root level
	PolicyLevelGlobal PolicyLevel = "global"
	// PolicyLevelFestival is the festival-specific level
	PolicyLevelFestival PolicyLevel = "festival"
	// PolicyLevelPhase is the phase-specific level
	PolicyLevelPhase PolicyLevel = "phase"
	// PolicyLevelSequence is the sequence-specific level
	PolicyLevelSequence PolicyLevel = "sequence"
)

// PolicySource tracks where a policy or gate originated
type PolicySource struct {
	Level PolicyLevel // Hierarchy level where defined
	Path  string      // File path where defined (empty for builtin)
	Name  string      // Policy name if from named policy
}

// GateTask represents a quality gate task definition
type GateTask struct {
	ID             string         `yaml:"id" json:"id"`
	Template       string         `yaml:"template" json:"template"`
	Name           string         `yaml:"name,omitempty" json:"name,omitempty"`
	Enabled        bool           `yaml:"enabled" json:"enabled"`
	Customizations map[string]any `yaml:"customizations,omitempty" json:"customizations,omitempty"`
	// Hierarchical tracking fields (not serialized)
	Source  *PolicySource `yaml:"-" json:"-"` // Origin tracking
	Removed bool          `yaml:"-" json:"-"` // Marked for removal at child level
}

// GatePolicy represents a complete quality gate policy
type GatePolicy struct {
	Version         int        `yaml:"version" json:"version"`
	Name            string     `yaml:"name" json:"name"`
	Description     string     `yaml:"description,omitempty" json:"description,omitempty"`
	Append          []GateTask `yaml:"append" json:"append"`
	ExcludePatterns []string   `yaml:"exclude_patterns,omitempty" json:"exclude_patterns,omitempty"`
	// Hierarchical control fields
	Inherit *bool         `yaml:"inherit,omitempty" json:"inherit,omitempty"` // nil = true (inherit from parent)
	Source  *PolicySource `yaml:"-" json:"-"`                                 // Origin tracking
}

// ShouldInherit returns whether this policy inherits from parent levels.
// Returns true if Inherit is nil (default) or explicitly true.
func (p *GatePolicy) ShouldInherit() bool {
	if p.Inherit == nil {
		return true
	}
	return *p.Inherit
}

// GateOperation represents a phase-level override operation
type GateOperation struct {
	Add    *GateAddOp    `yaml:"add,omitempty" json:"add,omitempty"`
	Remove *GateRemoveOp `yaml:"remove,omitempty" json:"remove,omitempty"`
}

// GateAddOp represents an add operation
type GateAddOp struct {
	Task   GateTask `yaml:"task" json:"task"`
	After  string   `yaml:"after,omitempty" json:"after,omitempty"`   // Insert after this task ID
	Before string   `yaml:"before,omitempty" json:"before,omitempty"` // Insert before this task ID
}

// GateRemoveOp represents a remove operation
type GateRemoveOp struct {
	ID string `yaml:"id" json:"id"`
}

// PhaseOverride represents phase-level gate configuration
type PhaseOverride struct {
	Ops []GateOperation `yaml:"ops,omitempty" json:"ops,omitempty"`
}

// DefaultPolicy returns the built-in default policy
func DefaultPolicy() *GatePolicy {
	return &GatePolicy{
		Version:     1,
		Name:        DefaultPolicyName,
		Description: "Default quality gates: testing, code review, and iteration",
		Append: []GateTask{
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
		ExcludePatterns: []string{
			// Common planning/documentation patterns
			"*_planning",
			"*_research",
			"*_requirements",
			"*_docs",
			"*_design",
			// Additional patterns for scope/review work
			"*_scope",
			"*_validation",
			"*_signoff",
			"*_review",
			"*_analysis",
			"*_assessment",
			"*_discovery",
		},
	}
}

// LoadPolicy loads a gate policy from a file
func LoadPolicy(path string) (*GatePolicy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read policy file: %w", err)
	}

	var policy GatePolicy
	if err := yaml.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("failed to parse policy file: %w", err)
	}

	// Apply defaults
	if policy.Version == 0 {
		policy.Version = 1
	}

	return &policy, nil
}

// SavePolicy saves a gate policy to a file
func SavePolicy(path string, policy *GatePolicy) error {
	data, err := yaml.Marshal(policy)
	if err != nil {
		return fmt.Errorf("failed to marshal policy: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write policy file: %w", err)
	}

	return nil
}

// LoadPhaseOverride loads a phase-level override from a phase directory
func LoadPhaseOverride(phasePath string) (*PhaseOverride, error) {
	overridePath := filepath.Join(phasePath, PhaseOverrideFileName)

	if _, err := os.Stat(overridePath); os.IsNotExist(err) {
		return nil, nil // No override file
	}

	data, err := os.ReadFile(overridePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read phase override: %w", err)
	}

	var override PhaseOverride
	if err := yaml.Unmarshal(data, &override); err != nil {
		return nil, fmt.Errorf("failed to parse phase override: %w", err)
	}

	return &override, nil
}

// GetEnabledTasks returns only enabled tasks from the policy
func (p *GatePolicy) GetEnabledTasks() []GateTask {
	var enabled []GateTask
	for _, task := range p.Append {
		if task.Enabled {
			enabled = append(enabled, task)
		}
	}
	return enabled
}

// Clone creates a deep copy of the policy
func (p *GatePolicy) Clone() *GatePolicy {
	clone := &GatePolicy{
		Version:         p.Version,
		Name:            p.Name,
		Description:     p.Description,
		ExcludePatterns: make([]string, len(p.ExcludePatterns)),
		Append:          make([]GateTask, len(p.Append)),
	}
	copy(clone.ExcludePatterns, p.ExcludePatterns)
	for i, task := range p.Append {
		clone.Append[i] = task
		if task.Customizations != nil {
			clone.Append[i].Customizations = make(map[string]any)
			for k, v := range task.Customizations {
				clone.Append[i].Customizations[k] = v
			}
		}
	}
	return clone
}
