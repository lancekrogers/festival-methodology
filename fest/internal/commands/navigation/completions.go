package navigation

import (
	"fmt"

	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/spf13/cobra"
)

// NewGoCompletionsCommand creates the hidden completions subcommand for shell integration
func NewGoCompletionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "completions",
		Short:  "Output completion words for shell integration",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoCompletions()
		},
	}

	return cmd
}

func runGoCompletions() error {
	// Subcommands
	subcommands := []string{
		"list",
		"link",
		"map",
		"unmap",
		"project",
		"fest",
		"help",
	}

	// Status directories
	statuses := []string{
		"active",
		"planned",
		"completed",
		"dungeon",
	}

	// Output subcommands
	for _, cmd := range subcommands {
		fmt.Println(cmd)
	}

	// Output status directories
	for _, status := range statuses {
		fmt.Println(status)
	}

	// Load navigation state for shortcuts
	nav, err := navigation.LoadNavigation()
	if err != nil {
		// Silently skip shortcuts if navigation fails
		return nil
	}

	// Output shortcuts with - prefix
	for name := range nav.Shortcuts {
		fmt.Printf("-%s\n", name)
	}

	return nil
}
