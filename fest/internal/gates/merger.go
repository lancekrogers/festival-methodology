package gates

import (
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
)

// PolicyLoader handles loading and merging gate policies from multiple sources
type PolicyLoader struct {
	// Policy search paths in order of precedence
	searchPaths []string
}

// NewPolicyLoader creates a new policy loader
func NewPolicyLoader() *PolicyLoader {
	return &PolicyLoader{
		searchPaths: []string{},
	}
}

// AddSearchPath adds a path to search for policies
func (pl *PolicyLoader) AddSearchPath(path string) {
	pl.searchPaths = append(pl.searchPaths, path)
}

// LoadWithPrecedence loads a policy by name, checking sources in precedence order:
// 1. Project-local (festival root)
// 2. User config repo (~/.config/fest/active/user/policies/gates/)
// 3. Built-in default
func (pl *PolicyLoader) LoadWithPrecedence(policyName string, festivalRoot string) (*GatePolicy, string, error) {
	// Try project-local
	localPath := filepath.Join(festivalRoot, "policies", "gates", policyName+".yml")
	if policy, err := LoadPolicy(localPath); err == nil {
		return policy, localPath, nil
	}

	// Try user config repo
	userPath := config.ActiveUserPath()
	if userPath != "" {
		userPolicyPath := filepath.Join(userPath, "policies", "gates", policyName+".yml")
		if policy, err := LoadPolicy(userPolicyPath); err == nil {
			return policy, userPolicyPath, nil
		}
	}

	// Try custom search paths
	for _, searchPath := range pl.searchPaths {
		policyPath := filepath.Join(searchPath, policyName+".yml")
		if policy, err := LoadPolicy(policyPath); err == nil {
			return policy, policyPath, nil
		}
	}

	// Fall back to built-in default
	if policyName == DefaultPolicyName || policyName == "" {
		return DefaultPolicy(), "built-in", nil
	}

	// Policy not found - return default with warning
	return DefaultPolicy(), "built-in", nil
}

// ApplyPhaseOverride applies phase-level overrides to a policy
func ApplyPhaseOverride(policy *GatePolicy, override *PhaseOverride) *GatePolicy {
	if override == nil || len(override.Ops) == 0 {
		return policy
	}

	// Clone the policy to avoid modifying the original
	result := policy.Clone()

	for _, op := range override.Ops {
		if op.Add != nil {
			result = applyAddOp(result, op.Add)
		}
		if op.Remove != nil {
			result = applyRemoveOp(result, op.Remove)
		}
	}

	return result
}

// applyAddOp adds a task to the policy
func applyAddOp(policy *GatePolicy, op *GateAddOp) *GatePolicy {
	task := op.Task

	// Find insertion point
	insertIdx := len(policy.Append) // Default: append at end

	if op.After != "" {
		for i, t := range policy.Append {
			if t.ID == op.After {
				insertIdx = i + 1
				break
			}
		}
	} else if op.Before != "" {
		for i, t := range policy.Append {
			if t.ID == op.Before {
				insertIdx = i
				break
			}
		}
	}

	// Insert at position
	newAppend := make([]GateTask, 0, len(policy.Append)+1)
	newAppend = append(newAppend, policy.Append[:insertIdx]...)
	newAppend = append(newAppend, task)
	newAppend = append(newAppend, policy.Append[insertIdx:]...)
	policy.Append = newAppend

	return policy
}

// applyRemoveOp removes a task from the policy
func applyRemoveOp(policy *GatePolicy, op *GateRemoveOp) *GatePolicy {
	newAppend := make([]GateTask, 0, len(policy.Append))
	for _, t := range policy.Append {
		if t.ID != op.ID {
			newAppend = append(newAppend, t)
		}
	}
	policy.Append = newAppend
	return policy
}

// ResolvePolicy loads and resolves a policy for a specific phase
func ResolvePolicy(policyName string, festivalRoot string, phasePath string) (*GatePolicy, error) {
	loader := NewPolicyLoader()

	// Load base policy
	policy, _, err := loader.LoadWithPrecedence(policyName, festivalRoot)
	if err != nil {
		return nil, err
	}

	// Load and apply phase override
	override, err := LoadPhaseOverride(phasePath)
	if err != nil {
		return nil, err
	}

	return ApplyPhaseOverride(policy, override), nil
}
