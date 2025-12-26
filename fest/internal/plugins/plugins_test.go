package plugins

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPluginCommandParts(t *testing.T) {
	tests := []struct {
		command       string
		expectedGroup string
		expectedName  string
	}{
		{"export jira", "export", "jira"},
		{"review parallel", "review", "parallel"},
		{"simple", "simple", ""},
	}

	for _, tc := range tests {
		p := Plugin{Command: tc.command}
		group, name := p.CommandParts()
		if group != tc.expectedGroup {
			t.Errorf("CommandParts(%q) group = %q, want %q", tc.command, group, tc.expectedGroup)
		}
		if name != tc.expectedName {
			t.Errorf("CommandParts(%q) name = %q, want %q", tc.command, name, tc.expectedName)
		}
	}
}

func TestPluginMatchesArgs(t *testing.T) {
	tests := []struct {
		command  string
		args     []string
		expected bool
	}{
		{"export jira", []string{"export", "jira"}, true},
		{"export jira", []string{"export", "csv"}, false},
		{"export jira", []string{"export"}, false},
		{"export jira", []string{}, false},
		{"simple", []string{"simple"}, true},
		{"simple", []string{"other"}, false},
	}

	for _, tc := range tests {
		p := Plugin{Command: tc.command}
		result := p.MatchesArgs(tc.args)
		if result != tc.expected {
			t.Errorf("MatchesArgs(%q, %v) = %v, want %v", tc.command, tc.args, result, tc.expected)
		}
	}
}

func TestPluginConsumedArgs(t *testing.T) {
	tests := []struct {
		command  string
		expected int
	}{
		{"export jira", 2},
		{"simple", 1},
	}

	for _, tc := range tests {
		p := Plugin{Command: tc.command}
		result := p.ConsumedArgs()
		if result != tc.expected {
			t.Errorf("ConsumedArgs(%q) = %d, want %d", tc.command, result, tc.expected)
		}
	}
}

func TestLoadSaveManifest(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "plugins-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	manifest := &PluginManifest{
		Version: 1,
		Plugins: []Plugin{
			{
				Command:     "export jira",
				Exec:        "fest-export-jira",
				Summary:     "Export to Jira",
				Description: "Export festival to Jira",
				WhenToUse:   []string{"Need Jira tracking"},
				Examples:    []string{"fest export jira --phase 002"},
			},
		},
	}

	manifestPath := filepath.Join(tmpDir, "manifest.yml")

	// Save
	if err := SaveManifest(manifestPath, manifest); err != nil {
		t.Fatalf("SaveManifest error: %v", err)
	}

	// Load
	loaded, err := LoadManifest(manifestPath)
	if err != nil {
		t.Fatalf("LoadManifest error: %v", err)
	}

	if loaded.Version != 1 {
		t.Errorf("Loaded version = %d, want 1", loaded.Version)
	}
	if len(loaded.Plugins) != 1 {
		t.Errorf("Loaded plugins count = %d, want 1", len(loaded.Plugins))
	}
	if loaded.Plugins[0].Command != "export jira" {
		t.Errorf("Loaded plugin command = %q, want %q", loaded.Plugins[0].Command, "export jira")
	}
}

func TestManifestFindPlugin(t *testing.T) {
	manifest := &PluginManifest{
		Version: 1,
		Plugins: []Plugin{
			{Command: "export jira", Exec: "fest-export-jira"},
			{Command: "review parallel", Exec: "fest-review-parallel"},
		},
	}

	// Find existing
	p := manifest.FindPlugin("export jira")
	if p == nil {
		t.Error("FindPlugin returned nil for existing plugin")
	}
	if p != nil && p.Exec != "fest-export-jira" {
		t.Errorf("FindPlugin exec = %q, want %q", p.Exec, "fest-export-jira")
	}

	// Find non-existing
	p = manifest.FindPlugin("nonexistent")
	if p != nil {
		t.Error("FindPlugin returned non-nil for non-existing plugin")
	}
}

func TestExecToCommand(t *testing.T) {
	tests := []struct {
		exec     string
		expected string
	}{
		{"fest-export-jira", "export jira"},
		{"fest-review-parallel", "review parallel"},
		{"fest-simple", "simple"},
	}

	for _, tc := range tests {
		result := execToCommand(tc.exec)
		if result != tc.expected {
			t.Errorf("execToCommand(%q) = %q, want %q", tc.exec, result, tc.expected)
		}
	}
}

func TestCommandToExec(t *testing.T) {
	tests := []struct {
		command  string
		expected string
	}{
		{"export jira", "fest-export-jira"},
		{"review parallel", "fest-review-parallel"},
		{"simple", "fest-simple"},
	}

	for _, tc := range tests {
		result := commandToExec(tc.command)
		if result != tc.expected {
			t.Errorf("commandToExec(%q) = %q, want %q", tc.command, result, tc.expected)
		}
	}
}

func TestPluginDiscoveryFindByCommand(t *testing.T) {
	pd := NewPluginDiscovery()
	pd.plugins = []DiscoveredPlugin{
		{Plugin: Plugin{Command: "export jira", Exec: "fest-export-jira"}, Source: "test"},
		{Plugin: Plugin{Command: "review parallel", Exec: "fest-review-parallel"}, Source: "test"},
	}

	// Find existing
	p := pd.FindByCommand("export jira")
	if p == nil {
		t.Error("FindByCommand returned nil for existing plugin")
	}

	// Find non-existing
	p = pd.FindByCommand("nonexistent")
	if p != nil {
		t.Error("FindByCommand returned non-nil for non-existing plugin")
	}
}

func TestPluginDiscoveryFindByArgs(t *testing.T) {
	pd := NewPluginDiscovery()
	pd.plugins = []DiscoveredPlugin{
		{Plugin: Plugin{Command: "export jira", Exec: "fest-export-jira"}, Source: "test"},
		{Plugin: Plugin{Command: "review parallel", Exec: "fest-review-parallel"}, Source: "test"},
	}

	tests := []struct {
		args     []string
		expected string // empty if should not find
	}{
		{[]string{"export", "jira"}, "export jira"},
		{[]string{"review", "parallel"}, "review parallel"},
		{[]string{"export", "csv"}, ""},
		{[]string{}, ""},
	}

	for _, tc := range tests {
		p := pd.FindByArgs(tc.args)
		if tc.expected == "" {
			if p != nil {
				t.Errorf("FindByArgs(%v) = %q, want nil", tc.args, p.Command)
			}
		} else {
			if p == nil {
				t.Errorf("FindByArgs(%v) = nil, want %q", tc.args, tc.expected)
			} else if p.Command != tc.expected {
				t.Errorf("FindByArgs(%v) = %q, want %q", tc.args, p.Command, tc.expected)
			}
		}
	}
}

func TestDispatcherCanHandle(t *testing.T) {
	d := NewDispatcher()
	d.discovery.plugins = []DiscoveredPlugin{
		{Plugin: Plugin{Command: "export jira", Exec: "fest-export-jira"}, Source: "test"},
	}

	if !d.CanHandle([]string{"export", "jira"}) {
		t.Error("CanHandle returned false for registered plugin")
	}

	if d.CanHandle([]string{"unknown", "cmd"}) {
		t.Error("CanHandle returned true for unknown command")
	}
}
