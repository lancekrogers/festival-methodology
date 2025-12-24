// Package gates provides quality gate policy management.
package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigMerger merges gate configuration from fest.yaml and policy files.
// It provides a unified view of all gate sources with proper precedence.
type ConfigMerger struct {
	festivalsRoot string
	loader        *HierarchicalLoader
}

// NewConfigMerger creates a new config merger with the given festivals root.
func NewConfigMerger(festivalsRoot string, registry *PolicyRegistry) (*ConfigMerger, error) {
	loader, err := NewHierarchicalLoader(festivalsRoot, registry)
	if err != nil {
		return nil, fmt.Errorf("creating hierarchical loader: %w", err)
	}

	return &ConfigMerger{
		festivalsRoot: festivalsRoot,
		loader:        loader,
	}, nil
}

// MergeOptions controls how configuration is merged.
type MergeOptions struct {
	IncludeFestYAML bool // Include fest.yaml as a source (default: true)
	IncludeBuiltin  bool // Include built-in defaults (default: true)
}

// DefaultMergeOptions returns the default merge options.
func DefaultMergeOptions() MergeOptions {
	return MergeOptions{
		IncludeFestYAML: true,
		IncludeBuiltin:  true,
	}
}

// MergedPolicy represents the result of merging all configuration sources.
type MergedPolicy struct {
	Gates           []GateTask      // Final merged list of gates
	ExcludePatterns []string        // Merged exclude patterns
	Sources         []PolicySource  // All sources that contributed
	Level           PolicyLevel     // Most specific level that contributed
	FestYAMLEnabled bool            // Whether fest.yaml has quality gates enabled
}

// GetActiveGates returns only enabled and non-removed gates.
func (m *MergedPolicy) GetActiveGates() []GateTask {
	active := make([]GateTask, 0, len(m.Gates))
	for _, gate := range m.Gates {
		if !gate.Removed && gate.Enabled {
			active = append(active, gate)
		}
	}
	return active
}

// MergeForSequence loads and merges all configuration for a specific sequence.
// Precedence (lowest to highest):
// 1. Built-in defaults
// 2. fest.yaml quality_gates.tasks
// 3. Global policy (.festival/gates/policies/default.yml)
// 4. Festival-level policy (.festival/gates.yml)
// 5. Phase-level override (.fest.gates.yml)
// 6. Sequence-level override (.fest.gates.yml)
func (m *ConfigMerger) MergeForSequence(
	ctx context.Context,
	festivalPath, phasePath, sequencePath string,
	opts MergeOptions,
) (*MergedPolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	merged := m.initMergedPolicy(opts)

	// Step 2: Apply fest.yaml if enabled
	if opts.IncludeFestYAML {
		if err := m.applyFestYAML(festivalPath, merged); err != nil {
			return nil, err
		}
	}

	// Step 3-6: Apply policy files from hierarchy
	if err := m.applyPolicyHierarchy(ctx, festivalPath, phasePath, sequencePath, merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// MergeForPhase loads and merges configuration for a phase.
func (m *ConfigMerger) MergeForPhase(
	ctx context.Context,
	festivalPath, phasePath string,
	opts MergeOptions,
) (*MergedPolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	merged := m.initMergedPolicy(opts)

	if opts.IncludeFestYAML {
		if err := m.applyFestYAML(festivalPath, merged); err != nil {
			return nil, err
		}
	}

	if err := m.applyPolicyHierarchyPhase(ctx, festivalPath, phasePath, merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// MergeForFestival loads and merges configuration for a festival.
func (m *ConfigMerger) MergeForFestival(
	ctx context.Context,
	festivalPath string,
	opts MergeOptions,
) (*MergedPolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	merged := m.initMergedPolicy(opts)

	if opts.IncludeFestYAML {
		if err := m.applyFestYAML(festivalPath, merged); err != nil {
			return nil, err
		}
	}

	if err := m.applyPolicyHierarchyFestival(ctx, festivalPath, merged); err != nil {
		return nil, err
	}

	return merged, nil
}

// initMergedPolicy creates a MergedPolicy with built-in defaults.
func (m *ConfigMerger) initMergedPolicy(opts MergeOptions) *MergedPolicy {
	merged := &MergedPolicy{
		Gates:           make([]GateTask, 0),
		ExcludePatterns: make([]string, 0),
		Sources:         make([]PolicySource, 0),
		Level:           PolicyLevelBuiltin,
	}

	if opts.IncludeBuiltin {
		defaultPolicy := DefaultPolicy()
		builtinSource := PolicySource{
			Level: PolicyLevelBuiltin,
			Path:  "",
			Name:  DefaultPolicyName,
		}

		for _, task := range defaultPolicy.Append {
			task.Source = &builtinSource
			merged.Gates = append(merged.Gates, task)
		}
		merged.ExcludePatterns = append(merged.ExcludePatterns, defaultPolicy.ExcludePatterns...)
		merged.Sources = append(merged.Sources, builtinSource)
	}

	return merged
}

// applyFestYAML loads and applies fest.yaml configuration.
func (m *ConfigMerger) applyFestYAML(festivalPath string, merged *MergedPolicy) error {
	policy, err := GatePolicyFromFestConfig(festivalPath)
	if err != nil {
		return fmt.Errorf("loading fest.yaml: %w", err)
	}

	if policy == nil {
		// Quality gates disabled in fest.yaml
		merged.FestYAMLEnabled = false
		return nil
	}

	merged.FestYAMLEnabled = true

	// fest.yaml gates replace built-in defaults (they're more specific)
	if len(policy.Append) > 0 {
		// Replace gates with same ID, add new ones
		for _, gate := range policy.Append {
			merged.Gates = m.addOrReplaceGate(merged.Gates, gate)
		}
	}

	// Merge exclude patterns
	if len(policy.ExcludePatterns) > 0 {
		merged.ExcludePatterns = mergePatterns(merged.ExcludePatterns, policy.ExcludePatterns)
	}

	if policy.Source != nil {
		merged.Sources = append(merged.Sources, *policy.Source)
		merged.Level = policy.Source.Level
	}

	return nil
}

// applyPolicyHierarchy applies policies from file hierarchy for sequence level.
func (m *ConfigMerger) applyPolicyHierarchy(
	ctx context.Context,
	festivalPath, phasePath, sequencePath string,
	merged *MergedPolicy,
) error {
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, m.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
		{PolicyLevelPhase, phasePath, PhaseOverrideFileName},
		{PolicyLevelSequence, sequencePath, PhaseOverrideFileName},
	}

	return m.applyPolicyLevels(ctx, levels, merged)
}

// applyPolicyHierarchyPhase applies policies for phase level.
func (m *ConfigMerger) applyPolicyHierarchyPhase(
	ctx context.Context,
	festivalPath, phasePath string,
	merged *MergedPolicy,
) error {
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, m.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
		{PolicyLevelPhase, phasePath, PhaseOverrideFileName},
	}

	return m.applyPolicyLevels(ctx, levels, merged)
}

// applyPolicyHierarchyFestival applies policies for festival level.
func (m *ConfigMerger) applyPolicyHierarchyFestival(
	ctx context.Context,
	festivalPath string,
	merged *MergedPolicy,
) error {
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, m.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
	}

	return m.applyPolicyLevels(ctx, levels, merged)
}

// applyPolicyLevels applies a series of policy levels to merged policy.
func (m *ConfigMerger) applyPolicyLevels(
	ctx context.Context,
	levels []struct {
		level PolicyLevel
		path  string
		file  string
	},
	merged *MergedPolicy,
) error {
	for _, lvl := range levels {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("context cancelled: %w", err)
		}

		policyPath := filepath.Join(lvl.path, lvl.file)
		if _, err := os.Stat(policyPath); os.IsNotExist(err) {
			continue
		}

		policy, err := LoadPolicy(policyPath)
		if err != nil {
			return fmt.Errorf("loading policy at %s: %w", policyPath, err)
		}

		policy.Source = &PolicySource{
			Level: lvl.level,
			Path:  policyPath,
			Name:  policy.Name,
		}

		// Handle inherit: false
		if !policy.ShouldInherit() {
			merged.Gates = nil
		}

		// Apply gates
		for _, gate := range policy.Append {
			gate.Source = policy.Source
			merged.Gates = m.addOrReplaceGate(merged.Gates, gate)
		}

		// Merge exclude patterns
		if len(policy.ExcludePatterns) > 0 {
			merged.ExcludePatterns = mergePatterns(merged.ExcludePatterns, policy.ExcludePatterns)
		}

		merged.Sources = append(merged.Sources, *policy.Source)
		merged.Level = policy.Source.Level
	}

	return nil
}

// addOrReplaceGate adds a gate or replaces if one with same ID exists.
func (m *ConfigMerger) addOrReplaceGate(gates []GateTask, newGate GateTask) []GateTask {
	for i, gate := range gates {
		if gate.ID == newGate.ID {
			gates[i] = newGate
			return gates
		}
	}
	return append(gates, newGate)
}

// mergePatterns merges two pattern lists, deduplicating.
func mergePatterns(existing, additional []string) []string {
	seen := make(map[string]bool)
	for _, p := range existing {
		seen[p] = true
	}

	result := existing
	for _, p := range additional {
		if !seen[p] {
			result = append(result, p)
			seen[p] = true
		}
	}
	return result
}

// Loader returns the underlying HierarchicalLoader.
func (m *ConfigMerger) Loader() *HierarchicalLoader {
	return m.loader
}

// FestivalsRoot returns the configured festivals root path.
func (m *ConfigMerger) FestivalsRoot() string {
	return m.festivalsRoot
}
