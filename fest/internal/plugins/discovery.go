package plugins

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
)

const (
	// PluginPrefix is the prefix for fest plugin executables
	PluginPrefix = "fest-"
)

// DiscoveredPlugin represents a plugin found during discovery
type DiscoveredPlugin struct {
	Plugin
	Source string // "manifest", "path", or path to manifest
	Path   string // Full path to executable
}

// PluginDiscovery handles plugin discovery from multiple sources
type PluginDiscovery struct {
	plugins []DiscoveredPlugin
}

// NewPluginDiscovery creates a new plugin discovery instance
func NewPluginDiscovery() *PluginDiscovery {
	return &PluginDiscovery{
		plugins: []DiscoveredPlugin{},
	}
}

// DiscoverAll discovers plugins from all sources
func (pd *PluginDiscovery) DiscoverAll() error {
	// 1. Load from user config repo manifest
	if userPath := config.ActiveUserPath(); userPath != "" {
		manifestPath := filepath.Join(userPath, "plugins", ManifestFileName)
		if err := pd.loadManifest(manifestPath, "user"); err == nil {
			// Manifest loaded successfully
		}
	}

	// 2. Scan user config repo bin directory
	if userPath := config.ActiveUserPath(); userPath != "" {
		binPath := filepath.Join(userPath, "plugins", "bin")
		pd.scanDirectory(binPath, "user-bin")
	}

	// 3. Scan PATH for fest-* executables
	pd.scanPath()

	return nil
}

// loadManifest loads plugins from a manifest file
func (pd *PluginDiscovery) loadManifest(manifestPath, source string) error {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return err
	}

	for _, p := range manifest.Plugins {
		// Try to find executable
		execPath, _ := exec.LookPath(p.Exec)
		if execPath == "" {
			// Check config repo bin directory
			if userPath := config.ActiveUserPath(); userPath != "" {
				binPath := filepath.Join(userPath, "plugins", "bin", p.Exec)
				if _, err := os.Stat(binPath); err == nil {
					execPath = binPath
				}
			}
		}

		pd.plugins = append(pd.plugins, DiscoveredPlugin{
			Plugin: p,
			Source: source,
			Path:   execPath,
		})
	}

	return nil
}

// scanDirectory scans a directory for fest-* executables
func (pd *PluginDiscovery) scanDirectory(dir, source string) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasPrefix(name, PluginPrefix) {
			continue
		}

		fullPath := filepath.Join(dir, name)

		// Check if executable
		info, err := os.Stat(fullPath)
		if err != nil {
			continue
		}
		if info.Mode()&0111 == 0 {
			continue // Not executable
		}

		// Skip if already discovered (from manifest)
		if pd.hasExec(name) {
			continue
		}

		// Create plugin from executable name
		// fest-export-jira -> "export jira"
		command := execToCommand(name)

		pd.plugins = append(pd.plugins, DiscoveredPlugin{
			Plugin: Plugin{
				Command: command,
				Exec:    name,
				Summary: "Plugin: " + name,
			},
			Source: source,
			Path:   fullPath,
		})
	}
}

// scanPath scans PATH for fest-* executables
func (pd *PluginDiscovery) scanPath() {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return
	}

	paths := filepath.SplitList(pathEnv)
	for _, dir := range paths {
		pd.scanDirectory(dir, "path")
	}
}

// hasExec checks if an executable is already discovered
func (pd *PluginDiscovery) hasExec(exec string) bool {
	for _, p := range pd.plugins {
		if p.Exec == exec {
			return true
		}
	}
	return false
}

// execToCommand converts an executable name to a command string
// fest-export-jira -> "export jira"
func execToCommand(exec string) string {
	// Remove "fest-" prefix
	name := strings.TrimPrefix(exec, PluginPrefix)

	// Replace first "-" with space
	for i, r := range name {
		if r == '-' {
			return name[:i] + " " + name[i+1:]
		}
	}

	return name
}

// commandToExec converts a command to executable name
// "export jira" -> "fest-export-jira"
func commandToExec(command string) string {
	// Replace space with "-"
	name := strings.ReplaceAll(command, " ", "-")
	return PluginPrefix + name
}

// Plugins returns all discovered plugins
func (pd *PluginDiscovery) Plugins() []DiscoveredPlugin {
	return pd.plugins
}

// FindByCommand finds a plugin by command string
func (pd *PluginDiscovery) FindByCommand(command string) *DiscoveredPlugin {
	for i := range pd.plugins {
		if pd.plugins[i].Command == command {
			return &pd.plugins[i]
		}
	}
	return nil
}

// FindByArgs finds a plugin matching command-line args
func (pd *PluginDiscovery) FindByArgs(args []string) *DiscoveredPlugin {
	if len(args) == 0 {
		return nil
	}

	// Try "group name" format first
	if len(args) >= 2 {
		command := args[0] + " " + args[1]
		if p := pd.FindByCommand(command); p != nil {
			return p
		}
	}

	// Try single arg format
	command := args[0]
	if p := pd.FindByCommand(command); p != nil {
		return p
	}

	// Try converting "group-name" to "group name"
	if strings.Contains(args[0], "-") {
		parts := strings.SplitN(args[0], "-", 2)
		if len(parts) == 2 {
			command := parts[0] + " " + parts[1]
			if p := pd.FindByCommand(command); p != nil {
				return p
			}
		}
	}

	return nil
}
