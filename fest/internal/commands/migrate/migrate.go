// Package migrate provides the fest migrate command for document migrations.
package migrate

import (
	"github.com/spf13/cobra"
)

// NewMigrateCommand creates the migrate command with subcommands
func NewMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrate festival documents",
		Long: `Migrate festival documents to add missing features or update formats.

Available migrations:
  frontmatter   Add YAML frontmatter to existing documents

Examples:
  fest migrate frontmatter              # Add frontmatter to all docs
  fest migrate frontmatter --dry-run    # Preview changes
  fest migrate frontmatter --fix        # Auto-fix existing frontmatter`,
	}

	cmd.AddCommand(NewFrontmatterCommand())

	return cmd
}
