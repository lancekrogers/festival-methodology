// Package next provides the fest next command for task navigation.
package next

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/next"
	"github.com/spf13/cobra"
)

var (
	jsonOutput    bool
	verboseOutput bool
	shortOutput   bool
	cdOutput      bool
	sequenceOnly  bool
)

// NewNextCommand creates the next command
func NewNextCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "next",
		Short: "Find the next task to work on",
		Long: `Determine the next task to work on based on dependencies and progress.

The command analyzes the festival structure, checks task completion status,
and recommends the next task following the priority order:

1. Tasks in current sequence with satisfied dependencies
2. Next incomplete task in current phase
3. First incomplete task in earliest phase
4. Quality gates before phase transitions

Examples:
  fest next                    # Find next task in festival
  fest next --sequence         # Only consider current sequence
  fest next --json             # Output as JSON
  fest next --verbose          # Detailed output
  fest next --short            # Just the task path
  fest next --cd               # Output directory for shell cd`,
		RunE: runNext,
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().BoolVar(&verboseOutput, "verbose", false, "show detailed information")
	cmd.Flags().BoolVar(&shortOutput, "short", false, "output only the task path")
	cmd.Flags().BoolVar(&cdOutput, "cd", false, "output directory path for cd command")
	cmd.Flags().BoolVar(&sequenceOnly, "sequence", false, "only consider current sequence")

	return cmd
}

func runNext(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	selector := next.NewSelector(festivalPath)

	var result *next.NextTaskResult
	if sequenceOnly {
		seqPath := findSequencePath(cwd, festivalPath)
		if seqPath == "" {
			return errors.NotFound("not inside a sequence directory")
		}
		result, err = selector.FindNextInSequence(seqPath)
	} else {
		result, err = selector.FindNext(cwd)
	}

	if err != nil {
		return errors.Wrap(err, "finding next task")
	}

	// Output formatting
	if cdOutput {
		output := next.FormatCD(result)
		if output == "" {
			return errors.NotFound("no task available to navigate to")
		}
		fmt.Println(output)
		return nil
	}

	if shortOutput {
		fmt.Println(next.FormatShort(result))
		return nil
	}

	if jsonOutput {
		output, err := next.FormatJSON(result)
		if err != nil {
			return errors.Parse("formatting JSON", err)
		}
		fmt.Println(output)
		return nil
	}

	if verboseOutput {
		fmt.Print(next.FormatVerbose(result))
		return nil
	}

	fmt.Print(next.FormatText(result))
	return nil
}

// findSequencePath finds the sequence path from current directory
func findSequencePath(cwd, festivalPath string) string {
	// Walk up from cwd looking for a sequence directory
	current := cwd
	for {
		// Check if current is a sequence (numbered dir inside a numbered phase dir)
		parent := filepath.Dir(current)
		if parent == festivalPath {
			// Current is a phase, not a sequence
			return ""
		}
		grandparent := filepath.Dir(parent)
		if grandparent == festivalPath {
			// Parent is a phase, current might be a sequence
			if isNumberedDir(filepath.Base(parent)) && isNumberedDir(filepath.Base(current)) {
				return current
			}
		}
		if current == festivalPath || current == "/" || current == "." {
			break
		}
		current = parent
	}
	return ""
}

// isNumberedDir checks if directory name starts with a number
func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}
