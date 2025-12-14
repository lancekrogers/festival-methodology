package commands

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

// NewUnderstandCommand creates the understand command group
func NewUnderstandCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "understand [topic]",
		Short: "Learn Festival Methodology",
		Long: `Learn about Festival Methodology - a goal-oriented project management
system designed for AI agent development workflows.

The understand command helps you grasp the methodology so you know
WHEN and WHY to use specific approaches. For command usage, use --help.

Content is pulled from your local .festival/ directory when available,
ensuring you see the current methodology design and any customizations.

Topics:
  methodology   Core principles and philosophy
  structure     Three-level hierarchy (Phases → Sequences → Tasks)
  workflow      Just-in-time reading and execution patterns
  resources     What's in the .festival/ directory`,
		Example: `  # Overview of Festival Methodology
  fest understand

# Learn core principles

  fest understand methodology

# Understand the three-level hierarchy

  fest understand structure

# Learn the just-in-time workflow

  fest understand workflow

# See available resources

  fest understand resources`,
		Run: func(cmd *cobra.Command, args []string) {
			dotFestival := findDotFestivalDir()

			if len(args) == 0 {
				printOverview(dotFestival)
				return
			}

			topic := strings.ToLower(args[0])
			switch topic {
			case "methodology":
				printMethodology(dotFestival)
			case "structure":
				printStructure(dotFestival)
			case "workflow":
				printWorkflow(dotFestival)
			case "resources":
				printResources(dotFestival)
			default:
				fmt.Printf("Unknown topic: %s\n\n", topic)
				fmt.Println("Available topics: methodology, structure, workflow, resources")
			}
		},
	}

	return cmd
}

func printOverview(dotFestival string) {
	fmt.Print(`
Festival Methodology - AI Agent Build System
=============================================

Festival Methodology is a goal-oriented project management system designed
for human-AI collaboration and long-running autonomous development cycles.

`)

	// Try to pull core concepts from .festival/README.md
	if dotFestival != "" {
		if content := extractSection(filepath.Join(dotFestival, "README.md"), "Core Concepts", "Quick Navigation"); content != "" {
			fmt.Println(content)
		} else {
			printDefaultCoreConcepts()
		}
	} else {
		printDefaultCoreConcepts()
	}

	fmt.Print(`

Quick Start
-----------

1. Run 'fest understand methodology' to learn the principles
2. Run 'fest understand structure' to understand the hierarchy
3. Run 'fest understand workflow' to learn execution patterns
4. Use 'fest create festival --json' to scaffold a new festival

Token-Efficient Workflow
------------------------

Use the fest CLI instead of manual file creation:

  fest create festival --name "my-project" --json
  fest create phase --name "IMPLEMENT" --json
  fest create sequence --name "api" --json
  fest task defaults sync --approve --json

Each command is self-documenting. Use --help for details.

Learn More
----------

  fest understand methodology   # Core principles
  fest understand structure     # Three-level hierarchy
  fest understand workflow      # Just-in-time patterns
  fest understand resources     # What's in .festival/

`)

	if dotFestival != "" {
		fmt.Printf("Source: %s\n", dotFestival)
	} else {
		fmt.Println("Note: No .festival/ directory found. Showing default content.")
		fmt.Println("      Run from a festivals/ tree to see your methodology resources.")
	}
}

func printDefaultCoreConcepts() {
	fmt.Print(`Core Concepts
-------------

• Three-level hierarchy: Phases → Sequences → Tasks
• Requirements-driven: Implementation only after requirements are defined
• Just-in-time context: Read templates and docs only when needed
• Quality gates: Every sequence ends with testing, review, iteration
`)
}

func printMethodology(dotFestival string) {
	fmt.Print(`
Festival Methodology - Core Principles
======================================

`)

	// Try to read from FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md
	if dotFestival != "" {
		docPath := filepath.Join(dotFestival, "FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md")
		if content := readFileContent(docPath); content != "" {
			// Extract key sections
			if section := extractSection(content, "## Core Principles", "## "); section != "" {
				fmt.Println(section)
			} else {
				// Show the whole file summary if no specific section
				fmt.Printf("See: %s\n\n", docPath)
				printDefaultMethodology()
			}
		} else {
			printDefaultMethodology()
		}

		// Also check README.md for methodology guidance
		readmePath := filepath.Join(dotFestival, "README.md")
		if content := extractSection(readmePath, "Requirements-Driven Implementation", "### Standard 3-Phase"); content != "" {
			fmt.Println("\nRequirements-Driven Implementation")
			fmt.Println("-----------------------------------")
			fmt.Println(content)
		}

		if content := extractSection(readmePath, "Phase Adaptability", "## Goal Files"); content != "" {
			fmt.Println("\nPhase Adaptability")
			fmt.Println("------------------")
			fmt.Println(content)
		}
	} else {
		printDefaultMethodology()
	}

	if dotFestival != "" {
		fmt.Printf("\nSource: %s\n", dotFestival)
	}
}

func printDefaultMethodology() {
	fmt.Print(`1. Goal-Oriented Development
----------------------------

Everything is organized around achieving a specific goal. A "festival"
represents the complete journey from problem to solution.

2. Requirements-Driven Implementation

-------------------------------------
CRITICAL: Implementation sequences can ONLY be created AFTER requirements
are defined. Never guess what to implement.

  Human provides:
  • Project goals and vision
  • Requirements and specifications
  • Architectural decisions
  • Feedback and iteration guidance

  AI agent creates:
  • Structured sequences from requirements
  • Detailed task specifications
  • Implementation code
  • Progress tracking and documentation

3. Phase Adaptability

---------------------
Phases are guidelines, not rigid requirements. Adapt to your needs.

4. Quality Gates

----------------
Every implementation sequence ends with testing, review, and iteration tasks.
Use 'fest task defaults sync' to add these automatically.

5. Step-Based Progress

----------------------
Think in steps, not time. Track progress by completed tasks, not hours.
`)
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

Naming Conventions
------------------

  Phases:     3-digit prefix → 001_PLAN, 002_IMPLEMENT, 003_REVIEW
  Sequences:  2-digit prefix → 01_requirements, 02_api, 03_frontend
  Tasks:      2-digit prefix → 01_design_schema.md, 02_implement_models.md

Parallel Execution
------------------

Tasks with the same number execute in parallel:

  01_frontend_setup.md  ┐
  01_backend_setup.md   ├── Can run simultaneously
  01_database_setup.md  ┘

`)

	if dotFestival != "" {
		fmt.Printf("Source: %s\n", dotFestival)
	}
}

func printScaffoldTree(variant string) {
	switch variant {
	case "simple":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md
├── FESTIVAL_GOAL.md
├── TODO.md
│
├── 001_PLAN/
│   ├── PHASE_GOAL.md
│   └── 01_requirements/
│       ├── SEQUENCE_GOAL.md
│       └── 01_gather_requirements.md
│
└── 002_IMPLEMENT/
    ├── PHASE_GOAL.md
    └── 01_core/
        ├── SEQUENCE_GOAL.md
        ├── 01_implement_feature.md
        └── 02_add_tests.md
`)
	case "standard":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md
├── FESTIVAL_GOAL.md
├── TODO.md
├── fest.yaml                         ← Quality gate config
│
├── 001_PLAN/
│   ├── PHASE_GOAL.md
│   └── 01_requirements/
│       ├── SEQUENCE_GOAL.md
│       ├── 01_gather_requirements.md
│       └── 02_document_specs.md
│
├── 002_IMPLEMENT/
│   ├── PHASE_GOAL.md
│   ├── 01_backend/
│   │   ├── SEQUENCE_GOAL.md
│   │   ├── 01_create_models.md
│   │   ├── 02_implement_api.md
│   │   ├── 03_testing_and_verify.md  ← Quality gate
│   │   ├── 04_code_review.md         ← Quality gate
│   │   └── 05_review_results_iterate.md ← Quality gate
│   └── 02_frontend/
│       ├── SEQUENCE_GOAL.md
│       ├── 01_create_components.md
│       ├── 02_add_styling.md
│       ├── 03_testing_and_verify.md
│       ├── 04_code_review.md
│       └── 05_review_results_iterate.md
│
└── 003_REVIEW_AND_UAT/
    ├── PHASE_GOAL.md
    └── 01_final_validation/
        ├── SEQUENCE_GOAL.md
        └── 01_user_acceptance_testing.md
`)
	case "complex":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md
├── FESTIVAL_GOAL.md
├── TODO.md
├── fest.yaml
│
├── 001_RESEARCH/
│   ├── PHASE_GOAL.md
│   └── 01_discovery/
│       └── [research documents]
│
├── 002_PLAN/
│   ├── PHASE_GOAL.md
│   ├── 01_requirements/
│   │   └── [requirement docs]
│   └── 02_architecture/
│       └── [design docs]
│
├── 003_IMPLEMENT_CORE/
│   ├── PHASE_GOAL.md
│   ├── 01_foundation/
│   │   ├── [tasks...]
│   │   └── [quality gates]
│   └── 02_data_layer/
│       ├── [tasks...]
│       └── [quality gates]
│
├── 004_IMPLEMENT_FEATURES/
│   ├── PHASE_GOAL.md
│   ├── 01_feature_a/
│   │   └── [tasks + quality gates]
│   └── 02_feature_b/
│       └── [tasks + quality gates]
│
└── 005_FINAL_REVIEW/
    ├── PHASE_GOAL.md
    └── 01_integration_testing/
        └── [validation tasks]
`)
	}
}

func printWorkflow(dotFestival string) {
	fmt.Print(`
Festival Workflow - Just-in-Time Reading
========================================

The just-in-time approach preserves context window for actual work.

`)

	// Try to read workflow guidance from .festival/README.md
	if dotFestival != "" {
		readmePath := filepath.Join(dotFestival, "README.md")
		if content := extractSection(readmePath, "When to Read What", "### Never Do This"); content != "" {
			fmt.Println("When to Read What (from .festival/)")
			fmt.Println("------------------------------------")
			fmt.Println(content)
		} else {
			printDefaultWorkflowTable()
		}

		if content := extractSection(readmePath, "### Never Do This", "### Always Do This"); content != "" {
			fmt.Println("\nNEVER Do This")
			fmt.Println("-------------")
			fmt.Println(content)
		}

		if content := extractSection(readmePath, "### Always Do This", "## Quick Navigation"); content != "" {
			fmt.Println("\nALWAYS Do This")
			fmt.Println("--------------")
			fmt.Println(content)
		}
	} else {
		printDefaultWorkflowTable()
		printDefaultWorkflowRules()
	}

	fmt.Print(`

Execution Pattern
-----------------

1. Read FESTIVAL_GOAL.md to understand the objective
2. Check TODO.md for current status
3. Navigate to the current phase/sequence
4. Read ONLY the current task file
5. Execute the task
6. Mark complete and move to next

Using fest CLI Efficiently
--------------------------

The CLI is self-documenting. Use --help instead of reading docs:

  fest --help                    # All commands
  fest create --help             # Create subcommands
  fest create festival --help    # Specific command details
  fest task defaults --help      # Quality gate management

JSON output provides structured responses:

  fest create phase --name "IMPLEMENT" --json

`)

	if dotFestival != "" {
		fmt.Printf("Source: %s\n", dotFestival)
	}
}

func printDefaultWorkflowTable() {
	fmt.Print(`
Resource              Read When
  ────────────────────  ─────────────────────────────────────
  fest understand       Now - provides methodology overview
  Templates             ONLY when creating that document type
  Examples              ONLY when stuck or need clarification
  Agents                ONLY when using that specific agent
`)
}

func printDefaultWorkflowRules() {
	fmt.Print(`
NEVER Do This
-------------

  ✗ Reading all templates upfront "to understand them"
  ✗ Loading all agent files at once
  ✗ Reading examples before trying yourself

ALWAYS Do This
--------------

  ✓ Read templates one at a time as you create documents
  ✓ Read examples only when stuck
  ✓ Close files after extracting what you need
  ✓ Focus context on actual work, not documentation
`)
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
