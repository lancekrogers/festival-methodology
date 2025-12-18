package plugins

import (
	"fmt"
	"os"
	"os/exec"
)

// Dispatcher handles plugin command dispatch
type Dispatcher struct {
	discovery *PluginDiscovery
}

// NewDispatcher creates a new plugin dispatcher
func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		discovery: NewPluginDiscovery(),
	}
}

// Initialize discovers all available plugins
func (d *Dispatcher) Initialize() error {
	return d.discovery.DiscoverAll()
}

// CanHandle checks if the given args match a plugin
func (d *Dispatcher) CanHandle(args []string) bool {
	return d.discovery.FindByArgs(args) != nil
}

// Dispatch executes a plugin with the given args
func (d *Dispatcher) Dispatch(args []string) error {
	plugin := d.discovery.FindByArgs(args)
	if plugin == nil {
		return fmt.Errorf("no plugin found for: %v", args)
	}

	return d.Execute(plugin, args)
}

// Execute runs a discovered plugin
func (d *Dispatcher) Execute(plugin *DiscoveredPlugin, allArgs []string) error {
	// Find the executable
	execPath := plugin.Path
	if execPath == "" {
		// Try to find in PATH
		var err error
		execPath, err = exec.LookPath(plugin.Exec)
		if err != nil {
			return fmt.Errorf("plugin executable not found: %s", plugin.Exec)
		}
	}

	// Determine how many args the plugin command consumes
	consumedArgs := plugin.ConsumedArgs()

	// Build args for the plugin (remaining args after command)
	var pluginArgs []string
	if len(allArgs) > consumedArgs {
		pluginArgs = allArgs[consumedArgs:]
	}

	// Create command
	cmd := exec.Command(execPath, pluginArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set environment
	cmd.Env = append(os.Environ(),
		"FEST_PLUGIN=1",
		fmt.Sprintf("FEST_PLUGIN_COMMAND=%s", plugin.Command),
	)

	// Run and return exit code
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	return nil
}

// ListPlugins returns all discovered plugins
func (d *Dispatcher) ListPlugins() []DiscoveredPlugin {
	return d.discovery.Plugins()
}

// FindPlugin finds a plugin by command
func (d *Dispatcher) FindPlugin(command string) *DiscoveredPlugin {
	return d.discovery.FindByCommand(command)
}

// GetDiscovery returns the underlying discovery instance
func (d *Dispatcher) GetDiscovery() *PluginDiscovery {
	return d.discovery
}
