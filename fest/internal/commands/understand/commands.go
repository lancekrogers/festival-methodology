package understand

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/lancekrogers/festival-methodology/fest/internal/extensions"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/plugins"
	"github.com/spf13/cobra"
)

// NewUnderstandCommand creates the understand command group with subcommands
func NewUnderstandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "understand",
		Short: "Learn methodology FIRST - run before executing festival tasks",
		Long: `Learn about Festival Methodology - a goal-oriented project management
system designed for AI agent development workflows.

START HERE if you're new to Festival Methodology:
  fest understand methodology    Core principles and philosophy
  fest understand structure      3-level hierarchy explained
  fest understand tasks          CRITICAL: When to create task files

QUICK REFERENCE:
  fest understand checklist      Validation checklist before starting
  fest understand rules          Naming conventions for automation
  fest understand workflow       Just-in-time reading pattern

The understand command helps you grasp WHEN and WHY to use specific
approaches. For command syntax, use --help on any command.

Content is pulled from your local .festival/ directory when available,
ensuring you see the current methodology design and any customizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			dotFestival := findDotFestivalDir()
			printOverview(dotFestival)
		},
	}

	// Add subcommands - ordered by importance
	// CRITICAL commands first
	cmd.AddCommand(newUnderstandTasksCmd())     // Most common mistake
	cmd.AddCommand(newUnderstandStructureCmd()) // Core hierarchy
	cmd.AddCommand(newUnderstandRulesCmd())     // Mandatory rules
	cmd.AddCommand(newUnderstandChecklistCmd()) // Quick validation

	// Learning commands
	cmd.AddCommand(newUnderstandMethodologyCmd())
	cmd.AddCommand(newUnderstandWorkflowCmd())
	cmd.AddCommand(newUnderstandTemplatesCmd())
	cmd.AddCommand(newUnderstandResourcesCmd())

	// Extension/plugin discovery
	cmd.AddCommand(newUnderstandGatesCmd())
	cmd.AddCommand(newUnderstandPluginsCmd())
	cmd.AddCommand(newUnderstandExtensionsCmd())

	return cmd
}

func newUnderstandTasksCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "tasks",
		Short: "When and how to create task files (CRITICAL)",
		Long: `Learn when to create task files vs. goal documents.

THIS IS THE MOST COMMON MISTAKE: Creating sequences with only
SEQUENCE_GOAL.md but no task files.

  Goals define WHAT to achieve.
  Tasks define HOW to execute.

AI agents EXECUTE TASK FILES. Without them, agents know the
objective but don't know what specific work to perform.`,
		Run: func(cmd *cobra.Command, args []string) {
			printTasks()
		},
	}
}

func newUnderstandMethodologyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "methodology",
		Short: "Core principles - START HERE for new agents",
		Long: `Learn the core principles of Festival Methodology.

This is the STARTING POINT for agents new to Festival Methodology.
Covers goal-oriented development, requirements-driven implementation,
and quality gates.

After reading this, proceed to:
  fest understand structure   - Learn the 3-level hierarchy
  fest understand tasks       - Learn when to create task files`,
		Run: func(cmd *cobra.Command, args []string) {
			printMethodology(findDotFestivalDir())
		},
	}
}

func newUnderstandStructureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "structure",
		Short: "3-level hierarchy: Festival → Phase → Sequence → Task",
		Long: `Understand the Festival Methodology structure.

HIERARCHY:
  Festival    - A complete project with a goal
  └─ Phase    - Major milestone (001_PLANNING, 002_IMPLEMENTATION)
     └─ Sequence - Related tasks grouped together
        └─ Task   - Individual executable work item

Includes visual scaffold examples for simple, standard, and complex festivals.`,
		Run: func(cmd *cobra.Command, args []string) {
			printStructure(findDotFestivalDir())
		},
	}
}

func newUnderstandWorkflowCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "workflow",
		Short: "Just-in-time reading and execution patterns",
		Long:  `Learn the just-in-time approach to reading templates and documentation, preserving context for actual work.`,
		Run: func(cmd *cobra.Command, args []string) {
			printWorkflow(findDotFestivalDir())
		},
	}
}

func newUnderstandResourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resources",
		Short: "What's in the .festival/ directory",
		Long:  `List the templates, agents, and examples available in your .festival/ directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			printResources(findDotFestivalDir())
		},
	}
}

func newUnderstandRulesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rules",
		Short: "MANDATORY structure rules for automation",
		Long:  `Learn the RIGID structure requirements that enable Festival automation: naming conventions, required files, quality gates, and parallel execution.`,
		Run: func(cmd *cobra.Command, args []string) {
			printRules()
		},
	}
}

func newUnderstandTemplatesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "templates",
		Short: "Template variables that save tokens",
		Long:  `Learn how to pass variables to fest create commands to generate pre-filled documents, minimizing post-creation editing and saving tokens.`,
		Run: func(cmd *cobra.Command, args []string) {
			printTemplates()
		},
	}
}

func newUnderstandChecklistCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checklist",
		Short: "Quick festival validation checklist",
		Long: `Show a quick checklist for validating your festival structure.

This is a quick reference. For full validation, run 'fest validate checklist'.

Checklist:
  1. FESTIVAL_OVERVIEW.md exists and is filled out
  2. Each phase has PHASE_GOAL.md
  3. Each sequence has SEQUENCE_GOAL.md
  4. Implementation sequences have TASK FILES (not just goals!)
  5. Quality gates present in implementation sequences
  6. No unfilled template markers ([FILL:], {{ }})`,
		Run: func(cmd *cobra.Command, args []string) {
			printChecklist()
		},
	}
}

func printChecklist() {
	fmt.Print(`
Festival Validation Checklist
=============================

Before executing a festival, verify:

  ✓ Festival Level
    □ FESTIVAL_OVERVIEW.md exists and defines goals
    □ FESTIVAL_RULES.md exists with quality standards

  ✓ Phase Level
    □ Each phase has PHASE_GOAL.md
    □ Phases are numbered correctly (001_, 002_, ...)

  ✓ Sequence Level
    □ Each sequence has SEQUENCE_GOAL.md
    □ Sequences are numbered correctly (01_, 02_, ...)

  ✗ CRITICAL: Task Files
    □ Implementation sequences have TASK FILES
    □ Not just SEQUENCE_GOAL.md - actual task files!
    □ Goals define WHAT; tasks define HOW
    □ AI agents execute TASK FILES

  ✓ Quality Gates
    □ Implementation sequences end with quality gates
    □ XX_testing_and_verify.md
    □ XX_code_review.md
    □ XX_review_results_iterate.md

  ✓ Templates
    □ No [FILL:] markers remaining
    □ No {{ }} template syntax in final docs


Quick Validation Commands
-------------------------

  fest validate                # Full validation
  fest validate tasks          # Check task files exist
  fest validate checklist      # Detailed checklist with auto-checks

`)
}

func printOverview(dotFestival string) {
	// Use embedded overview content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("overview.txt"))

	if dotFestival != "" {
		fmt.Printf("\nSource: %s\n", dotFestival)
	} else {
		fmt.Println("\nNote: No .festival/ directory found. Showing default content.")
		fmt.Println("      Run from a festivals/ tree to see your methodology resources.")
	}
}

func printMethodology(dotFestival string) {
	// Hybrid: try .festival/ supplements first, fall back to embedded defaults
	if dotFestival != "" {
		docPath := filepath.Join(dotFestival, "FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md")
		if content := readFileContent(docPath); content != "" {
			// Extract key sections from .festival/
			if section := extractSection(content, "## Core Principles", "## "); section != "" {
				fmt.Print("\nFestival Methodology - Core Principles\n")
				fmt.Println("======================================")
				fmt.Println(section)

				// Also check README.md for additional methodology guidance
				readmePath := filepath.Join(dotFestival, "README.md")
				if rdContent := extractSection(readmePath, "Requirements-Driven Implementation", "### Standard 3-Phase"); rdContent != "" {
					fmt.Println("\nRequirements-Driven Implementation")
					fmt.Println("-----------------------------------")
					fmt.Println(rdContent)
				}

				if paContent := extractSection(readmePath, "Phase Adaptability", "## Goal Files"); paContent != "" {
					fmt.Println("\nPhase Adaptability")
					fmt.Println("------------------")
					fmt.Println(paContent)
				}

				fmt.Printf("\nSource: %s\n", dotFestival)
				return
			}
		}
	}

	// Default: use embedded content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("methodology.txt"))
}

func printStructure(dotFestival string) {
	fmt.Print(`
Festival Structure - Three-Level Hierarchy
==========================================

Festival Methodology uses a three-level hierarchy:

  FESTIVAL (the project)
    └── PHASE (major stage of work)
          └── SEQUENCE (group of related tasks)
                └── TASK (atomic unit of work)

`)

	// Show scaffold trees
	fmt.Print(`

Scaffold: Simple Festival
-------------------------

`)
	printScaffoldTree("simple")

	fmt.Print(`

Scaffold: Standard Festival with Quality Gates
----------------------------------------------

`)
	printScaffoldTree("standard")

	fmt.Print(`

Scaffold: Complex Multi-Phase Festival
--------------------------------------

`)
	printScaffoldTree("complex")

	// Try to get naming conventions from .festival/
	if dotFestival != "" {
		if content := extractSection(filepath.Join(dotFestival, "README.md"), "### Standard 3-Phase Pattern", "### Phase Flexibility"); content != "" {
			fmt.Println("\nPhase Patterns (from .festival/)")
			fmt.Println("---------------------------------")
			fmt.Println(content)
		}
	}

	fmt.Print(`

Naming Conventions (MANDATORY)
------------------------------

  Phases:     NNN_PHASE_NAME      3-digit, UPPERCASE
  Sequences:  NN_sequence_name    2-digit, lowercase
  Tasks:      NN_task_name.md     2-digit, lowercase, .md extension

Parallel Execution
------------------

Tasks with the same number execute in parallel:

  01_frontend_setup.md  ┐
  01_backend_setup.md   ├── Run simultaneously
  01_database_setup.md  ┘
  02_integration.md     ← Waits for all 01_ tasks

For detailed requirements: fest understand rules
For template variables:    fest understand templates
`)

	if dotFestival != "" {
		fmt.Printf("\nSource: %s\n", dotFestival)
	}
}

func printScaffoldTree(variant string) {
	fmt.Print(understanddocs.LoadScaffold(variant))
}

func printWorkflow(dotFestival string) {
	// Hybrid: try .festival/ supplements first, fall back to embedded defaults
	if dotFestival != "" {
		readmePath := filepath.Join(dotFestival, "README.md")
		hasContent := false

		// Check if .festival/ has workflow content
		if content := extractSection(readmePath, "When to Read What", "### Never Do This"); content != "" {
			fmt.Print("\nFestival Workflow - Just-in-Time Reading\n")
			fmt.Println("========================================")
			fmt.Println("\nThe just-in-time approach preserves context window for actual work.")
			fmt.Println("\nWhen to Read What (from .festival/)")
			fmt.Println("------------------------------------")
			fmt.Println(content)
			hasContent = true

			if never := extractSection(readmePath, "### Never Do This", "### Always Do This"); never != "" {
				fmt.Println("\nNEVER Do This")
				fmt.Println("-------------")
				fmt.Println(never)
			}

			if always := extractSection(readmePath, "### Always Do This", "## Quick Navigation"); always != "" {
				fmt.Println("\nALWAYS Do This")
				fmt.Println("--------------")
				fmt.Println(always)
			}
		}

		if hasContent {
			fmt.Printf("\nSource: %s\n", dotFestival)
			return
		}
	}

	// Default: use embedded content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("workflow.txt"))
}

func printRules() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("rules.txt"))
}

func printTemplates() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("templates.txt"))
}

func printTasks() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("tasks.txt"))
}

func newUnderstandGatesCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "gates",
		Short: "Show quality gate configuration",
		Long: `Show the quality gate policy that will be applied to sequences.

Quality gates are tasks automatically appended to implementation sequences.
The default gates are: testing_and_verify, code_review, review_results_iterate.

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

func printResources(dotFestival string) {
	if dotFestival == "" {
		fmt.Println("\nNo .festival/ directory found.")
		fmt.Println("Run from a festivals/ tree to see your methodology resources.")
		fmt.Println("\nExpected location: festivals/.festival/")
		return
	}

	fmt.Printf("\nFestival Resources: %s\n", dotFestival)
	fmt.Println(strings.Repeat("=", 50))

	// List actual directory structure
	fmt.Println("\nDirectory Structure:")
	printDirectoryTree(dotFestival, "", 0)

	// Show templates with descriptions
	fmt.Println("\nTemplates (read when creating that document type):")
	fmt.Println("-" + strings.Repeat("-", 49))
	templatesDir := filepath.Join(dotFestival, "templates")
	if entries, err := os.ReadDir(templatesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				desc := getTemplateDescription(filepath.Join(templatesDir, name))
				if desc != "" {
					fmt.Printf("  %-35s %s\n", name, desc)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}
		}
	}

	// Show agents
	fmt.Println("\nAgents (read when using that agent):")
	fmt.Println("-" + strings.Repeat("-", 49))
	agentsDir := filepath.Join(dotFestival, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "INDEX.md" {
				name := entry.Name()
				desc := getTemplateDescription(filepath.Join(agentsDir, name))
				if desc != "" {
					fmt.Printf("  %-35s %s\n", name, desc)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}
		}
	}

	// Show examples
	fmt.Println("\nExamples (read when stuck or need patterns):")
	fmt.Println("-" + strings.Repeat("-", 49))
	examplesDir := filepath.Join(dotFestival, "examples")
	if entries, err := os.ReadDir(examplesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				fmt.Printf("  %s\n", entry.Name())
			}
		}
	}
}

func printDirectoryTree(dir string, prefix string, depth int) {
	if depth > 2 {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Filter to show only directories and key files
	var items []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".md") {
			items = append(items, entry)
		}
	}

	for i, entry := range items {
		isLast := i == len(items)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if entry.IsDir() {
			fmt.Printf("%s%s%s/\n", prefix, connector, entry.Name())
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			printDirectoryTree(filepath.Join(dir, entry.Name()), newPrefix, depth+1)
		} else {
			fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())
		}
	}
}

func getTemplateDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return ""
}

func findDotFestivalDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up looking for festivals/.festival or .festival
	dir := cwd
	for {
		// Check for .festival in current dir
		candidate := filepath.Join(dir, ".festival")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}

		// Check for festivals/.festival
		candidate = filepath.Join(dir, "festivals", ".festival")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}

		// Move up
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

func extractSection(pathOrContent string, startMarker, endMarker string) string {
	var content string
	if strings.Contains(pathOrContent, "\n") {
		content = pathOrContent
	} else {
		content = readFileContent(pathOrContent)
	}
	if content == "" {
		return ""
	}

	// Find start
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return ""
	}

	// Find content after start marker
	afterStart := content[startIdx+len(startMarker):]

	// Find end (look for next heading or end marker)
	endIdx := len(afterStart)
	if endMarker != "" {
		// Look for end marker or next same-level heading
		re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(endMarker))
		if loc := re.FindStringIndex(afterStart); loc != nil {
			endIdx = loc[0]
		}
	}

	section := strings.TrimSpace(afterStart[:endIdx])
	return section
}
