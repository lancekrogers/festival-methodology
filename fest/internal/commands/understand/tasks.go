package understand

import (
	"fmt"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

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

func printTasks() {
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("tasks.txt"))
}
