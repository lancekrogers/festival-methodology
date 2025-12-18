package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

// NewUnderstandCommand creates the understand command group with subcommands
func NewUnderstandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "understand",
		Short: "Learn Festival Methodology",
		Long: `Learn about Festival Methodology - a goal-oriented project management
system designed for AI agent development workflows.

The understand command helps you grasp the methodology so you know
WHEN and WHY to use specific approaches. For command usage, use --help.

Content is pulled from your local .festival/ directory when available,
ensuring you see the current methodology design and any customizations.`,
		Run: func(cmd *cobra.Command, args []string) {
			dotFestival := findDotFestivalDir()
			printOverview(dotFestival)
		},
	}

	// Add subcommands
	cmd.AddCommand(newUnderstandMethodologyCmd())
	cmd.AddCommand(newUnderstandStructureCmd())
	cmd.AddCommand(newUnderstandWorkflowCmd())
	cmd.AddCommand(newUnderstandResourcesCmd())
	cmd.AddCommand(newUnderstandRulesCmd())
	cmd.AddCommand(newUnderstandTemplatesCmd())
	cmd.AddCommand(newUnderstandTasksCmd())

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
		Short: "Core principles and philosophy",
		Long:  `Learn the core principles of Festival Methodology including goal-oriented development, requirements-driven implementation, and quality gates.`,
		Run: func(cmd *cobra.Command, args []string) {
			printMethodology(findDotFestivalDir())
		},
	}
}

func newUnderstandStructureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "structure",
		Short: "Three-level hierarchy with scaffold examples",
		Long:  `Understand the Phases → Sequences → Tasks hierarchy with visual scaffold examples for simple, standard, and complex festivals.`,
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

func printOverview(dotFestival string) {
	// Use embedded overview content
	fmt.Print("\n")
	fmt.Print(understand.Load("overview.txt"))

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
	fmt.Print(understand.Load("methodology.txt"))
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
	fmt.Print(understand.LoadScaffold(variant))
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
	fmt.Print(understand.Load("workflow.txt"))
}

func printRules() {
	fmt.Print("\n")
	fmt.Print(understand.Load("rules.txt"))
}

func printTemplates() {
	fmt.Print("\n")
	fmt.Print(understand.Load("templates.txt"))
}

func printTasks() {
	fmt.Print("\n")
	fmt.Print(understand.Load("tasks.txt"))
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
