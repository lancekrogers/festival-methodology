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

	return cmd
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

  fest understand --help        # See all topics
  fest understand rules         # MANDATORY structure requirements
  fest understand templates     # Variables that save tokens
  fest understand structure     # Scaffold examples with annotations

Token-Efficient Workflow
------------------------

Use the fest CLI instead of manual file creation:

  fest create festival --name "my-project" --goal "Build X" --json
  fest create phase --name "IMPLEMENT" --json
  fest create sequence --name "api" --json
  fest task defaults sync --approve --json   # Add quality gates

Variables are passed to templates - no post-creation editing needed.
Run 'fest understand templates' to see all available variables.

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
	switch variant {
	case "simple":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md              [REQUIRED] Goal and scope
├── FESTIVAL_RULES.md                 [REQUIRED] Quality standards
├── TODO.md                           Progress tracking
│
├── 001_PLAN/                         3-digit phase prefix
│   ├── PHASE_GOAL.md                 [REQUIRED] Phase objectives
│   └── 01_requirements/              2-digit sequence prefix
│       ├── SEQUENCE_GOAL.md          [REQUIRED] Sequence objectives
│       └── 01_gather_requirements.md 2-digit task prefix + .md
│
└── 002_IMPLEMENT/
    ├── PHASE_GOAL.md                 [REQUIRED]
    └── 01_core/
        ├── SEQUENCE_GOAL.md          [REQUIRED]
        ├── 01_implement_feature.md
        └── 02_add_tests.md
`)
	case "standard":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md              [REQUIRED]
├── FESTIVAL_RULES.md                 [REQUIRED]
├── TODO.md
├── fest.yaml                         Quality gate config
│
├── 001_PLAN/
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   └── 01_requirements/              (no quality gates - planning)
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       ├── 01_gather_requirements.md
│       └── 02_document_specs.md
│
├── 002_IMPLEMENT/
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   ├── 01_backend/
│   │   ├── SEQUENCE_GOAL.md          [REQUIRED]
│   │   ├── 01_create_models.md       ┐
│   │   ├── 02_implement_api.md       ├── Your tasks
│   │   ├── 03_testing_and_verify.md  ← QUALITY GATE [REQUIRED]
│   │   ├── 04_code_review.md         ← QUALITY GATE [REQUIRED]
│   │   └── 05_review_results_iterate.md ← QUALITY GATE [REQUIRED]
│   └── 02_frontend/
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       ├── 01_create_components.md
│       ├── 02_add_styling.md
│       ├── 03_testing_and_verify.md  ← QUALITY GATE
│       ├── 04_code_review.md         ← QUALITY GATE
│       └── 05_review_results_iterate.md ← QUALITY GATE
│
└── 003_REVIEW_AND_UAT/
    ├── PHASE_GOAL.md                 [REQUIRED]
    └── 01_final_validation/
        ├── SEQUENCE_GOAL.md          [REQUIRED]
        └── 01_user_acceptance_testing.md
`)
	case "complex":
		fmt.Print(`my-festival/
├── FESTIVAL_OVERVIEW.md              [REQUIRED]
├── FESTIVAL_RULES.md                 [REQUIRED]
├── TODO.md
├── fest.yaml
│
├── 001_RESEARCH/                     (no quality gates - research phase)
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   └── 01_discovery/
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       └── [research documents]
│
├── 002_PLAN/                         (no quality gates - planning phase)
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   ├── 01_requirements/
│   │   ├── SEQUENCE_GOAL.md          [REQUIRED]
│   │   └── [requirement docs]
│   └── 02_architecture/
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       └── [design docs]
│
├── 003_IMPLEMENT_CORE/               (quality gates required)
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   ├── 01_foundation/
│   │   ├── SEQUENCE_GOAL.md          [REQUIRED]
│   │   ├── [tasks...]
│   │   └── [quality gates]           ← 3 required tasks
│   └── 02_data_layer/
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       ├── [tasks...]
│       └── [quality gates]           ← 3 required tasks
│
├── 004_IMPLEMENT_FEATURES/           (quality gates required)
│   ├── PHASE_GOAL.md                 [REQUIRED]
│   ├── 01_feature_a/
│   │   ├── SEQUENCE_GOAL.md          [REQUIRED]
│   │   └── [tasks + quality gates]
│   └── 02_feature_b/
│       ├── SEQUENCE_GOAL.md          [REQUIRED]
│       └── [tasks + quality gates]
│
└── 005_FINAL_REVIEW/
    ├── PHASE_GOAL.md                 [REQUIRED]
    └── 01_integration_testing/
        ├── SEQUENCE_GOAL.md          [REQUIRED]
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

func printRules() {
	fmt.Print(`
Festival Structure Rules (MANDATORY)
====================================

The fest CLI enforces a RIGID structure that enables automation.
These rules are NOT suggestions - they are requirements.

NAMING CONVENTIONS
------------------

  Level      Format              Example
  ─────      ──────              ───────
  Phase      NNN_PHASE_NAME      001_PLAN, 002_IMPLEMENT, 003_REVIEW
  Sequence   NN_sequence_name    01_requirements, 02_api, 03_frontend
  Task       NN_task_name.md     01_design.md, 02_build.md, 03_test.md

  • Phases:    3-digit prefix, UPPERCASE name
  • Sequences: 2-digit prefix, lowercase name
  • Tasks:     2-digit prefix, lowercase name, .md extension

REQUIRED FILES
--------------

  Festival Root:
    FESTIVAL_OVERVIEW.md    [REQUIRED] Project goal, scope, stakeholders
    FESTIVAL_RULES.md       [REQUIRED] Quality standards and team rules
    TODO.md                 [Recommended] Progress tracking
    fest.yaml               [Optional] Quality gate configuration

  Each Phase (NNN_NAME/):
    PHASE_GOAL.md           [REQUIRED] Phase objectives and evaluation

  Each Sequence (NN_name/):
    SEQUENCE_GOAL.md        [REQUIRED] Sequence objectives and dependencies

QUALITY GATES
-------------

Every IMPLEMENTATION sequence MUST end with these 3 tasks:

    [your tasks here...]
    XX_testing_and_verify.md       ← Validates all deliverables
    XX+1_code_review.md            ← Reviews implementation quality
    XX+2_review_results_iterate.md ← Addresses review findings

  Use 'fest task defaults sync' to add quality gates automatically.

  Excluded sequences (no quality gates needed):
    - *_planning, *_research, *_requirements, *_docs

PARALLEL EXECUTION
------------------

Tasks with the SAME number execute in parallel:

    01_backend.md   ┐
    01_database.md  ├── Run simultaneously
    01_frontend.md  ┘
    02_integrate.md ← Waits for all 01_ tasks to complete

Use this for independent work streams within a sequence.

WHY THIS MATTERS
----------------

This rigid structure enables:
  • Automated quality gate insertion via 'fest task defaults sync'
  • Consistent navigation across all festivals
  • Parallel task detection and scheduling
  • Progress tracking and reporting
  • Template variable auto-computation (parent_phase_id, full_path, etc.)

Run 'fest understand templates' to learn how to pass variables.
`)
}

func printTemplates() {
	fmt.Print(`
Template Variables (Save Tokens)
================================

Pass variables to 'fest create' commands to generate pre-filled documents.
This eliminates post-creation editing and saves significant tokens.

FESTIVAL VARIABLES
------------------

  fest create festival \
    --name "auth-system" \
    --goal "Build OAuth 2.0 authentication" \
    --json

  Variables available in templates:
    {{ festival_name }}        → "auth-system"
    {{ festival_goal }}        → "Build OAuth 2.0 authentication"

PHASE VARIABLES
---------------

  fest create phase \
    --name "IMPLEMENT" \
    --json

  Auto-formatting: "implement" → "002_IMPLEMENT"

  Variables available:
    {{ phase_name }}           → "IMPLEMENT"
    {{ phase_number }}         → 2
    {{ phase_id }}             → "002_IMPLEMENT" (auto-computed)

SEQUENCE VARIABLES
------------------

  fest create sequence \
    --name "api endpoints" \
    --json

  Auto-formatting: "api endpoints" → "01_api_endpoints"

  Variables available:
    {{ sequence_name }}        → "api_endpoints"
    {{ sequence_number }}      → 1
    {{ sequence_id }}          → "01_api_endpoints" (auto-computed)
    {{ parent_phase_id }}      → "002_IMPLEMENT" (auto-computed)

TASK VARIABLES
--------------

  fest create task \
    --name "login endpoint" \
    --json

  Auto-formatting: "login endpoint" → "01_login_endpoint.md"

  Variables available:
    {{ task_name }}            → "login_endpoint"
    {{ task_number }}          → 1
    {{ task_id }}              → "01_login_endpoint.md" (auto-computed)
    {{ parent_sequence_id }}   → "01_api_endpoints" (auto-computed)
    {{ parent_phase_id }}      → "002_IMPLEMENT" (auto-computed)
    {{ full_path }}            → "002_IMPLEMENT/01_api_endpoints/01_login_endpoint.md"

AUTO-COMPUTED VARIABLES
-----------------------

The template engine auto-computes these (no input needed):

  • phase_id           → "002_IMPLEMENT" (from number + name)
  • sequence_id        → "01_api_endpoints" (from number + name)
  • task_id            → "01_login_endpoint.md" (from number + name)
  • parent_phase_id    → Current phase context
  • parent_sequence_id → Current sequence context
  • full_path          → Complete path from festival root

TOKEN SAVINGS
-------------

  Manual approach:
    1. Read template file            (~200 tokens)
    2. Understand template format    (~100 tokens)
    3. Write file with edits         (~200 tokens)
    Total: ~500 tokens per file

  CLI approach:
    fest create task --name "X" --json
    Total: ~50 tokens per file

  Savings: 90% token reduction per file

EXAMPLE: Full Festival Creation
-------------------------------

  # Create festival with goal
  fest create festival --name "user-auth" --goal "OAuth 2.0 system" --json

  # Create planning phase
  fest create phase --name "PLAN" --json

  # Create requirements sequence
  fest create sequence --name "requirements" --json

  # Create tasks
  fest create task --name "gather requirements" --json
  fest create task --name "document specs" --json

  # Add quality gates to all implementation sequences
  fest task defaults sync --approve --json

All generated files are pre-filled with correct structure and variables.
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
