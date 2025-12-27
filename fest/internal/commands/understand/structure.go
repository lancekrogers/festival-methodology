package understand

import (
	"fmt"
	"path/filepath"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

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
