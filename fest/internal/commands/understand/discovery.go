package understand

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/extensions"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/plugins"
	"github.com/spf13/cobra"
)

func newUnderstandGatesCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "gates",
		Short: "Show quality gate configuration",
		Long: `Show the quality gate policy that will be applied to sequences.

Quality gates are tasks automatically appended to implementation sequences.
The default gates are: testing_and_verify, code_review, review_results_iterate, commit.

Gates can be customized at multiple levels:
  1. Built-in defaults (always available)
  2. User config repo (~/.config/fest/active/user/policies/gates/)
  3. Project-local (.festival/policies/gates/)
  4. Phase override (.fest.gates.yml in phase directory)`,
		Run: func(cmd *cobra.Command, args []string) {
			printGates(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

func printGates(jsonOutput bool) {
	policy := gates.DefaultPolicy()

	if jsonOutput {
		output := map[string]interface{}{
			"policy":  policy,
			"source":  "built-in",
			"enabled": policy.GetEnabledTasks(),
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Print(`
Quality Gates Configuration
===========================

Quality gates are tasks automatically appended to implementation sequences
to ensure consistent quality standards across your festivals.

`)

	fmt.Printf("Policy: %s\n", policy.Name)
	fmt.Printf("Description: %s\n", policy.Description)
	fmt.Printf("Source: [BUILT-IN]\n\n")

	fmt.Println("Active Gates:")
	fmt.Println("-------------")
	for i, task := range policy.GetEnabledTasks() {
		fmt.Printf("  %d. %s\n", i+1, task.Name)
		fmt.Printf("     ID: %s\n", task.ID)
		fmt.Printf("     Template: %s\n", task.Template)
	}

	fmt.Println("\nExclude Patterns (sequences that skip gates):")
	fmt.Println("----------------------------------------------")
	for _, pattern := range policy.ExcludePatterns {
		fmt.Printf("  %s\n", pattern)
	}

	fmt.Print(`
Customization
-------------

Create a custom policy in your config repo:
  ~/.config/fest/active/user/policies/gates/custom.yml

Or add phase-level overrides:
  <phase>/.fest.gates.yml

See: fest help gates
`)
}

func newUnderstandPluginsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "plugins",
		Short: "Show discovered plugins",
		Long: `Show all plugins discovered from various sources.

Plugins extend fest with additional commands. They are discovered from:
  1. User config repo manifest (~/.config/fest/active/user/plugins/manifest.yml)
  2. User config repo bin directory (~/.config/fest/active/user/plugins/bin/)
  3. System PATH (executables named fest-*)

Plugin executables follow the naming convention:
  fest-<group>-<name>  →  "fest <group> <name>"
  fest-export-jira     →  "fest export jira"`,
		Run: func(cmd *cobra.Command, args []string) {
			printPlugins(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

func printPlugins(jsonOutput bool) {
	discovery := plugins.NewPluginDiscovery()
	_ = discovery.DiscoverAll()
	discovered := discovery.Plugins()

	if jsonOutput {
		output := map[string]interface{}{
			"count":   len(discovered),
			"plugins": discovered,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Print(`
Discovered Plugins
==================

Plugins extend fest with additional commands. Discovery sources:
  1. User config repo manifest
  2. User config repo bin directory
  3. System PATH (fest-* executables)

`)

	if len(discovered) == 0 {
		fmt.Println("No plugins discovered.")
		fmt.Print(`
To add plugins:
  1. Create manifest at ~/.config/fest/active/user/plugins/manifest.yml
  2. Place executables in ~/.config/fest/active/user/plugins/bin/
  3. Add fest-* executables to your PATH

See: fest help plugins
`)
		return
	}

	fmt.Printf("Found %d plugin(s):\n\n", len(discovered))

	for i, p := range discovered {
		fmt.Printf("  %d. %s\n", i+1, p.Command)
		fmt.Printf("     Executable: %s\n", p.Exec)
		fmt.Printf("     Source: %s\n", p.Source)
		if p.Summary != "" {
			fmt.Printf("     Summary: %s\n", p.Summary)
		}
		if p.Path != "" {
			fmt.Printf("     Path: %s\n", p.Path)
		}
		if len(p.WhenToUse) > 0 {
			fmt.Println("     When to use:")
			for _, hint := range p.WhenToUse {
				fmt.Printf("       - %s\n", hint)
			}
		}
		if len(p.Examples) > 0 {
			fmt.Println("     Examples:")
			for _, ex := range p.Examples {
				fmt.Printf("       $ %s\n", ex)
			}
		}
		fmt.Println()
	}
}

func newUnderstandExtensionsCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "extensions",
		Short: "Show loaded extensions",
		Long: `Show all methodology extensions loaded from various sources.

Extensions are workflow pattern packs containing templates, agents, and rules.
They are loaded from three sources with the following precedence:

  1. Project-local: .festival/extensions/ (highest priority)
  2. User config: ~/.config/fest/active/festivals/.festival/extensions/
  3. Built-in: ~/.config/fest/festivals/.festival/extensions/ (lowest priority)

Higher priority sources override lower ones when extensions have the same name.`,
		Run: func(cmd *cobra.Command, args []string) {
			printExtensions(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	return cmd
}

func printExtensions(jsonOutput bool) {
	// Get festival root if available
	cwd, _ := os.Getwd()
	festivalRoot := ""
	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".festival")); err == nil {
			festivalRoot = dir
			break
		}
		if _, err := os.Stat(filepath.Join(dir, "festivals", ".festival")); err == nil {
			festivalRoot = filepath.Join(dir, "festivals")
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	loader := extensions.NewExtensionLoader()
	_ = loader.LoadAll(festivalRoot)
	exts := loader.List()

	if jsonOutput {
		output := map[string]interface{}{
			"count":      len(exts),
			"extensions": exts,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
		return
	}

	fmt.Print(`
Loaded Extensions
=================

Extensions are methodology pattern packs containing templates, agents, and rules.
Loaded from three sources with precedence: project > user > built-in.

`)

	if len(exts) == 0 {
		fmt.Println("No extensions loaded.")
		fmt.Print(`
To add extensions:
  1. Create extension directories in .festival/extensions/
  2. Add extension.yml manifest with name, version, description
  3. Include templates, agents, or configuration files

See: fest extension list
`)
		return
	}

	fmt.Printf("Found %d extension(s):\n\n", len(exts))

	for i, ext := range exts {
		fmt.Printf("  %d. %s", i+1, ext.Name)
		if ext.Version != "" {
			fmt.Printf(" (v%s)", ext.Version)
		}
		fmt.Printf("\n")
		fmt.Printf("     Source: %s\n", ext.Source)
		if ext.Description != "" {
			fmt.Printf("     Description: %s\n", ext.Description)
		}
		if ext.Type != "" {
			fmt.Printf("     Type: %s\n", ext.Type)
		}
		if ext.Author != "" {
			fmt.Printf("     Author: %s\n", ext.Author)
		}
		if len(ext.Tags) > 0 {
			fmt.Printf("     Tags: %v\n", ext.Tags)
		}
		fmt.Println()
	}

	fmt.Println("For detailed info: fest extension info <name>")
}
