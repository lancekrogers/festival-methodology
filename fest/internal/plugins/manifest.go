// Package plugins provides external plugin discovery and dispatch.
package plugins

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const (
	// ManifestFileName is the name of the plugin manifest file
	ManifestFileName = "manifest.yml"
)

// Plugin represents a plugin definition in the manifest
type Plugin struct {
	Command     string   `yaml:"command" json:"command"`         // e.g., "export jira"
	Exec        string   `yaml:"exec" json:"exec"`               // e.g., "fest-export-jira"
	Summary     string   `yaml:"summary" json:"summary"`         // Short description
	Description string   `yaml:"description,omitempty" json:"description,omitempty"` // Long description
	WhenToUse   []string `yaml:"when_to_use,omitempty" json:"when_to_use,omitempty"` // Usage hints
	Examples    []string `yaml:"examples,omitempty" json:"examples,omitempty"`       // Example commands
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
}

// PluginManifest represents a plugin manifest file
type PluginManifest struct {
	Version int      `yaml:"version" json:"version"`
	Plugins []Plugin `yaml:"plugins" json:"plugins"`
}

// LoadManifest loads a plugin manifest from a file
func LoadManifest(path string) (*PluginManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %w", err)
	}

	var manifest PluginManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest: %w", err)
	}

	// Apply defaults
	if manifest.Version == 0 {
		manifest.Version = 1
	}

	return &manifest, nil
}

// SaveManifest saves a plugin manifest to a file
func SaveManifest(path string, manifest *PluginManifest) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	return nil
}

// FindPlugin searches the manifest for a plugin matching the command
func (m *PluginManifest) FindPlugin(command string) *Plugin {
	for i := range m.Plugins {
		if m.Plugins[i].Command == command {
			return &m.Plugins[i]
		}
	}
	return nil
}

// CommandParts returns the command split into parts (group and name)
func (p *Plugin) CommandParts() (string, string) {
	// Split "export jira" into ["export", "jira"]
	var group, name string
	for i, r := range p.Command {
		if r == ' ' {
			group = p.Command[:i]
			name = p.Command[i+1:]
			return group, name
		}
	}
	return p.Command, ""
}

// MatchesArgs checks if the plugin matches the given command args
func (p *Plugin) MatchesArgs(args []string) bool {
	if len(args) == 0 {
		return false
	}

	group, name := p.CommandParts()

	// Two-part command (e.g., "export jira") requires both parts
	if name != "" {
		if len(args) >= 2 {
			// Check "group name" format
			return args[0] == group && args[1] == name
		}
		if len(args) == 1 {
			// Check "group-name" format (single arg with dash)
			return args[0] == group+"-"+name
		}
		return false
	}

	// Single-part command
	return len(args) >= 1 && args[0] == group
}

// ConsumedArgs returns the number of args consumed by the plugin command
func (p *Plugin) ConsumedArgs() int {
	_, name := p.CommandParts()
	if name == "" {
		return 1
	}
	return 2
}
