// Package gates provides named policy registry for hierarchical gate management.
package gates

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// PolicyInfo describes a named policy
type PolicyInfo struct {
	Name        string `json:"name"`                  // Policy name (e.g., "default", "strict")
	Description string `json:"description,omitempty"` // Human-readable description
	Source      string `json:"source"`                // "built-in", "global", "config-repo"
	Path        string `json:"path,omitempty"`        // File path if not built-in
}

// PolicyRegistry manages available named policies
type PolicyRegistry struct {
	festivalsRoot string
	configRoot    string // ~/.config/fest/active/
	policies      map[string]*PolicyInfo
	mu            sync.RWMutex
}

// NewPolicyRegistry creates a registry scanning all policy sources
func NewPolicyRegistry(festivalsRoot, configRoot string) (*PolicyRegistry, error) {
	r := &PolicyRegistry{
		festivalsRoot: festivalsRoot,
		configRoot:    configRoot,
		policies:      make(map[string]*PolicyInfo),
	}

	// Register built-in policies
	r.registerBuiltin()

	// Scan additional sources (errors logged, not fatal)
	r.scanSources()

	return r, nil
}

// registerBuiltin registers the built-in policies
func (r *PolicyRegistry) registerBuiltin() {
	r.policies["default"] = &PolicyInfo{
		Name:        "default",
		Description: "Standard quality gates: testing, code review, iteration",
		Source:      "built-in",
	}
	r.policies["strict"] = &PolicyInfo{
		Name:        "strict",
		Description: "Strict code review with security audit and performance check",
		Source:      "built-in",
	}
	r.policies["lightweight"] = &PolicyInfo{
		Name:        "lightweight",
		Description: "Minimal gates for research and exploration phases",
		Source:      "built-in",
	}
}

// StrictPolicy returns gates with additional security and performance checks
func StrictPolicy() *GatePolicy {
	return &GatePolicy{
		Version:     1,
		Name:        "strict",
		Description: "Strict code review with security audit and performance check",
		Append: []GateTask{
			{ID: "testing_and_verify", Template: "QUALITY_GATE_TESTING", Name: "Testing and Verification", Enabled: true},
			{ID: "code_review", Template: "QUALITY_GATE_REVIEW", Name: "Code Review", Enabled: true},
			{ID: "security_audit", Template: "SECURITY_AUDIT", Name: "Security Audit", Enabled: true},
			{ID: "performance_check", Template: "PERFORMANCE_CHECK", Name: "Performance Check", Enabled: true},
			{ID: "review_results_iterate", Template: "QUALITY_GATE_ITERATE", Name: "Review Results and Iterate", Enabled: true},
		},
	}
}

// LightweightPolicy returns minimal gates for research work
func LightweightPolicy() *GatePolicy {
	return &GatePolicy{
		Version:     1,
		Name:        "lightweight",
		Description: "Minimal gates for research and exploration phases",
		Append: []GateTask{
			{ID: "code_review", Template: "QUALITY_GATE_REVIEW", Name: "Code Review", Enabled: true},
		},
	}
}

// scanSources scans all policy source directories
func (r *PolicyRegistry) scanSources() {
	// Scan global policies (overrides built-in)
	globalPath := filepath.Join(r.festivalsRoot, ".festival", "gates", "policies")
	_ = r.scanDirectory(globalPath, "global")

	// Scan user config repo policies (highest priority)
	if r.configRoot != "" {
		userPath := filepath.Join(r.configRoot, "user", "policies", "gates")
		_ = r.scanDirectory(userPath, "config-repo")
	}
}

// scanDirectory scans a directory for policy files
func (r *PolicyRegistry) scanDirectory(dir, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Directory doesn't exist, not an error
		}
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".yml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		name := strings.TrimSuffix(entry.Name(), ".yml")

		policy, err := LoadPolicy(path)
		if err != nil {
			// Log warning but continue scanning
			continue
		}

		r.mu.Lock()
		r.policies[name] = &PolicyInfo{
			Name:        name,
			Description: policy.Description,
			Source:      source,
			Path:        path,
		}
		r.mu.Unlock()
	}

	return nil
}

// Get retrieves a policy info by name
func (r *PolicyRegistry) Get(name string) (*PolicyInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, ok := r.policies[name]
	return info, ok
}

// GetPolicy retrieves the full policy by name
func (r *PolicyRegistry) GetPolicy(name string) (*GatePolicy, error) {
	r.mu.RLock()
	info, ok := r.policies[name]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("policy not found: %s", name)
	}

	// Built-in policies
	if info.Source == "built-in" {
		return r.getBuiltinPolicy(name)
	}

	// Load from file
	return LoadPolicy(info.Path)
}

// getBuiltinPolicy returns a built-in policy by name
func (r *PolicyRegistry) getBuiltinPolicy(name string) (*GatePolicy, error) {
	switch name {
	case "default":
		return DefaultPolicy(), nil
	case "strict":
		return StrictPolicy(), nil
	case "lightweight":
		return LightweightPolicy(), nil
	default:
		return nil, fmt.Errorf("unknown built-in policy: %s", name)
	}
}

// List returns all registered policy names
func (r *PolicyRegistry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.policies))
	for name := range r.policies {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ListInfo returns all registered policies with their info
func (r *PolicyRegistry) ListInfo() []*PolicyInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	infos := make([]*PolicyInfo, 0, len(r.policies))
	for _, info := range r.policies {
		infos = append(infos, info)
	}

	// Sort by name for consistent ordering
	sort.Slice(infos, func(i, j int) bool {
		return infos[i].Name < infos[j].Name
	})

	return infos
}

// Register adds a named policy to the registry
func (r *PolicyRegistry) Register(name string, info *PolicyInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.policies[name] = info
}

// Refresh rescans all policy sources
func (r *PolicyRegistry) Refresh() {
	r.mu.Lock()
	// Clear non-builtin policies
	for name, info := range r.policies {
		if info.Source != "built-in" {
			delete(r.policies, name)
		}
	}
	r.mu.Unlock()

	// Rescan sources
	r.scanSources()
}
