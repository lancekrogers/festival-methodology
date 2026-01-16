// Package types implements the fest types command for template type discovery.
package types

import (
	"github.com/spf13/cobra"
)

// NewTypesCommand creates the types command group.
func NewTypesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "Discover and explore template types",
		Long: `Explore available template types at each festival level.

Template types define the structure and purpose of festivals, phases,
sequences, and tasks. Custom types can be added in .festival/templates/.

Examples:
  fest types list                        # List all available types
  fest types list --level task           # List task-level types only
  fest types show feature                # Show details about a type
  fest types show implementation --level phase`,
	}

	cmd.AddCommand(newListCmd())
	cmd.AddCommand(newShowCmd())

	return cmd
}
