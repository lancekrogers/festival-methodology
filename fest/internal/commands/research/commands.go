package research

import (
	"github.com/spf13/cobra"
)

// NewResearchCommand creates the research command group
func NewResearchCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "research",
		Short: "Manage research phase documents",
		Long: `Manage research documents within research phases.

Research phases use flexible subdirectory structures instead of numbered
sequences. This command group helps create, organize, and link research
documents.

Available Commands:
  create    Create a new research document from template
  summary   Generate summary/index of research documents
  link      Link research findings to implementation phases`,
	}

	cmd.AddCommand(newResearchCreateCmd())
	cmd.AddCommand(newResearchSummaryCmd())
	cmd.AddCommand(newResearchLinkCmd())

	return cmd
}
