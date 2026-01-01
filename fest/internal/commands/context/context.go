// Package context provides the fest context command for agent context loading.
package context

import (
	"fmt"
	"os"

	ctx "github.com/lancekrogers/festival-methodology/fest/internal/context"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

var (
	jsonOutput bool
	verbose    bool
	depth      string
	taskName   string
)

// NewContextCommand creates the context command
func NewContextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "context",
		Short: "Get context for the current location or task",
		Long: `Provides AI agents with context for the current location in a festival.

Context includes:
  - Festival, phase, and sequence goals
  - Relevant rules from FESTIVAL_RULES.md
  - Recent decisions from CONTEXT.md
  - Dependency outputs (what prior tasks produced)

Depth levels:
  minimal   - Immediate goals, dependencies, autonomy level
  standard  - + Rules, recent decisions (5)
  full      - + All decisions, dependency outputs

Examples:
  fest context                    # Context for current location
  fest context --depth full       # Full context with all details
  fest context --task my_task     # Context for a specific task
  fest context --json             # Output as JSON
  fest context --verbose          # Explanatory output for agents`,
		RunE: runContext,
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "output with explanatory text for agents")
	cmd.Flags().StringVar(&depth, "depth", "standard", "context depth (minimal, standard, full)")
	cmd.Flags().StringVar(&taskName, "task", "", "get context for a specific task")

	return cmd
}

func runContext(cmd *cobra.Command, args []string) error {
	// Find festival root
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := tpl.FindFestivalRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	// Validate depth
	d := ctx.Depth(depth)
	switch d {
	case ctx.DepthMinimal, ctx.DepthStandard, ctx.DepthFull:
		// Valid
	default:
		return errors.Validation("invalid depth").
			WithField("depth", depth).
			WithField("valid", "minimal, standard, full")
	}

	// Build context
	builder := ctx.NewBuilder(festivalPath, d)

	var output *ctx.ContextOutput
	if taskName != "" {
		output, err = builder.BuildForTask(taskName)
	} else {
		output, err = builder.Build()
	}

	if err != nil {
		return errors.Wrap(err, "building context")
	}

	// Format and output
	formatter := ctx.NewFormatter(verbose)

	if jsonOutput {
		jsonStr, err := formatter.FormatJSON(output)
		if err != nil {
			return errors.Parse("formatting JSON", err)
		}
		fmt.Println(jsonStr)
	} else if verbose {
		fmt.Print(formatter.FormatVerbose(output))
	} else {
		fmt.Print(formatter.FormatText(output))
	}

	return nil
}
