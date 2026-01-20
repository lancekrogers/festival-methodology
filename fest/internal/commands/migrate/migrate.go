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
  metadata      Add ID and metadata to existing festivals
  times         Populate time tracking data from file modification times

Examples:
  fest migrate frontmatter              # Add frontmatter to all docs
  fest migrate frontmatter --dry-run    # Preview changes
  fest migrate frontmatter --fix        # Auto-fix existing frontmatter
  fest migrate metadata                 # Add IDs to all festivals
  fest migrate metadata --dry-run       # Preview ID migrations
  fest migrate times                    # Populate time data from file stats
  fest migrate times --dry-run          # Preview time migrations`,
	}

	cmd.AddCommand(NewFrontmatterCommand())
	cmd.AddCommand(NewMetadataCommand())
	cmd.AddCommand(NewTimesCommand())

	return cmd
}
