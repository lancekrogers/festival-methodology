package system

import "github.com/spf13/cobra"

// NewSystemCommand creates the system command group
func NewSystemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "Manage fest tool configuration and templates",
		Long: `Commands for maintaining the fest tool itself.

These commands manage fest's templates, configuration, and methodology
files - NOT your festival content. Use these to keep fest up to date.

Available subcommands:
  sync   - Download latest templates from GitHub
  update - Update .festival/ files from cached templates`,
	}

	cmd.AddCommand(NewSyncCommand())
	cmd.AddCommand(NewUpdateCommand())

	return cmd
}
