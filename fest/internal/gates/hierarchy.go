// Package gates provides hierarchical quality gate policy management.
package gates

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// EffectivePolicy represents the merged result of policies from all hierarchy levels
type EffectivePolicy struct {
	Gates   []GateTask     // Final list of gates to apply
	Sources []PolicySource // All sources that contributed
	Level   PolicyLevel    // Most specific level that contributed
}

// GetActiveGates returns only gates that are not marked as removed
func (e *EffectivePolicy) GetActiveGates() []GateTask {
	active := make([]GateTask, 0, len(e.Gates))
	for _, gate := range e.Gates {
		if !gate.Removed {
			active = append(active, gate)
		}
	}
	return active
}

// HierarchicalLoader loads and merges policies from all hierarchy levels
type HierarchicalLoader struct {
	festivalsRoot string          // Root of festivals directory
	registry      *PolicyRegistry // For loading named policies (optional)
}

// NewHierarchicalLoader creates a loader with the given festivals root
func NewHierarchicalLoader(festivalsRoot string, registry *PolicyRegistry) (*HierarchicalLoader, error) {
	if festivalsRoot == "" {
		return nil, errors.Validation("festivalsRoot cannot be empty").
			WithOp("NewHierarchicalLoader")
	}

	// Validate path exists
	if _, err := os.Stat(festivalsRoot); os.IsNotExist(err) {
		return nil, errors.NotFound("festivals root").
			WithField("path", festivalsRoot)
	}

	return &HierarchicalLoader{
		festivalsRoot: festivalsRoot,
		registry:      registry,
	}, nil
}

// LoadForSequence loads the effective policy for a specific sequence
func (h *HierarchicalLoader) LoadForSequence(
	ctx context.Context,
	festivalPath, phasePath, sequencePath string,
) (*EffectivePolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForSequence")
	}

	// Start with built-in default
	effective := h.initEffectivePolicy()

	// Define level order
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, h.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
		{PolicyLevelPhase, phasePath, PhaseOverrideFileName},
		{PolicyLevelSequence, sequencePath, PhaseOverrideFileName},
	}

	// Walk each level
	for _, lvl := range levels {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForSequence")
		}

		policyPath := filepath.Join(lvl.path, lvl.file)
		if _, err := os.Stat(policyPath); os.IsNotExist(err) {
			continue // No override at this level
		}

		policy, err := LoadPolicy(policyPath)
		if err != nil {
			return nil, errors.Wrap(err, "loading policy").
				WithOp("HierarchicalLoader.LoadForSequence").
				WithField("path", policyPath).
				WithCode(errors.ErrCodeConfig)
		}

		policy.Source = &PolicySource{
			Level: lvl.level,
			Path:  policyPath,
			Name:  policy.Name,
		}

		effective = h.mergePolicy(effective, policy)
	}

	return effective, nil
}

// LoadForPhase loads effective policy for a phase (stops at phase level)
func (h *HierarchicalLoader) LoadForPhase(
	ctx context.Context,
	festivalPath, phasePath string,
) (*EffectivePolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForPhase")
	}

	// Start with built-in default
	effective := h.initEffectivePolicy()

	// Define level order (no sequence level)
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, h.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
		{PolicyLevelPhase, phasePath, PhaseOverrideFileName},
	}

	for _, lvl := range levels {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForPhase")
		}

		policyPath := filepath.Join(lvl.path, lvl.file)
		if _, err := os.Stat(policyPath); os.IsNotExist(err) {
			continue
		}

		policy, err := LoadPolicy(policyPath)
		if err != nil {
			return nil, errors.Wrap(err, "loading policy").
				WithOp("HierarchicalLoader.LoadForPhase").
				WithField("path", policyPath).
				WithCode(errors.ErrCodeConfig)
		}

		policy.Source = &PolicySource{
			Level: lvl.level,
			Path:  policyPath,
			Name:  policy.Name,
		}

		effective = h.mergePolicy(effective, policy)
	}

	return effective, nil
}

// LoadForFestival loads effective policy for a festival (stops at festival level)
func (h *HierarchicalLoader) LoadForFestival(
	ctx context.Context,
	festivalPath string,
) (*EffectivePolicy, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForFestival")
	}

	// Start with built-in default
	effective := h.initEffectivePolicy()

	// Define level order (no phase or sequence)
	levels := []struct {
		level PolicyLevel
		path  string
		file  string
	}{
		{PolicyLevelGlobal, h.festivalsRoot, filepath.Join(".festival", "gates", "policies", "default.yml")},
		{PolicyLevelFestival, festivalPath, filepath.Join(".festival", "gates.yml")},
	}

	for _, lvl := range levels {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled").WithOp("HierarchicalLoader.LoadForFestival")
		}

		policyPath := filepath.Join(lvl.path, lvl.file)
		if _, err := os.Stat(policyPath); os.IsNotExist(err) {
			continue
		}

		policy, err := LoadPolicy(policyPath)
		if err != nil {
			return nil, errors.Wrap(err, "loading policy").
				WithOp("HierarchicalLoader.LoadForFestival").
				WithField("path", policyPath).
				WithCode(errors.ErrCodeConfig)
		}

		policy.Source = &PolicySource{
			Level: lvl.level,
			Path:  policyPath,
			Name:  policy.Name,
		}

		effective = h.mergePolicy(effective, policy)
	}

	return effective, nil
}

// initEffectivePolicy creates an EffectivePolicy with built-in defaults
func (h *HierarchicalLoader) initEffectivePolicy() *EffectivePolicy {
	defaultPolicy := DefaultPolicy()
	builtinSource := PolicySource{
		Level: PolicyLevelBuiltin,
		Path:  "",
		Name:  DefaultPolicyName,
	}

	gates := make([]GateTask, len(defaultPolicy.Append))
	for i, task := range defaultPolicy.Append {
		gates[i] = task
		gates[i].Source = &builtinSource
	}

	return &EffectivePolicy{
		Gates:   gates,
		Sources: []PolicySource{builtinSource},
		Level:   PolicyLevelBuiltin,
	}
}

// mergePolicy applies a policy to the current effective state
func (h *HierarchicalLoader) mergePolicy(current *EffectivePolicy, policy *GatePolicy) *EffectivePolicy {
	// Handle inherit: false FIRST, before any other operations
	if !policy.ShouldInherit() {
		current.Gates = nil
	}

	// Handle Append (from base policies)
	if len(policy.Append) > 0 {
		for _, gate := range policy.Append {
			gate.Source = policy.Source
			current.Gates = h.addGate(current.Gates, gate)
		}
	}

	// Track this policy as a source
	if policy.Source != nil {
		current.Sources = append(current.Sources, *policy.Source)
		current.Level = policy.Source.Level
	}

	return current
}

// addGate adds a gate to the list, respecting After positioning if specified
func (h *HierarchicalLoader) addGate(gates []GateTask, newGate GateTask) []GateTask {
	// Check if gate with same ID already exists
	for i, gate := range gates {
		if gate.ID == newGate.ID {
			// Replace existing gate
			gates[i] = newGate
			return gates
		}
	}

	// No positioning specified, append to end
	return append(gates, newGate)
}

// removeGate marks a gate as removed (keeps it for display purposes)
func (h *HierarchicalLoader) removeGate(gates []GateTask, removeID string, source *PolicySource) []GateTask {
	for i := range gates {
		if gates[i].ID == removeID {
			gates[i].Removed = true
			gates[i].Source = source
		}
	}
	return gates
}

// FestivalsRoot returns the configured festivals root path
func (h *HierarchicalLoader) FestivalsRoot() string {
	return h.festivalsRoot
}
