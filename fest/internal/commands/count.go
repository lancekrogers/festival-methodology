package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/fileops"
	"github.com/lancekrogers/festival-methodology/fest/internal/tokens"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type countOptions struct {
	model         string
	all           bool
	jsonOutput    bool
	showCost      bool
	recursive     bool
	charsPerToken float64
	wordsPerToken float64
}

// NewCountCommand creates the count command
func NewCountCommand() *cobra.Command {
	opts := &countOptions{}

	cmd := &cobra.Command{
		Use:   "count [file|directory]",
		Short: "Count tokens in a file or directory using various methods",
		Long: `Count tokens in a file or directory using multiple tokenization methods.

This command provides token counts using different LLM tokenizers and
approximation methods, helping you understand token usage and estimate costs.

When counting a directory with --recursive, the command:
- Respects .gitignore files
- Skips binary files automatically
- Returns aggregated totals for all text files`,
		Example: `  fest count document.md           # Count tokens in a file
  fest count --model gpt-4 doc.md  # Use specific model tokenizer
  fest count --all --cost doc.md   # Show all methods with costs
  fest count --json doc.md         # Output as JSON
  fest count -r ./src              # Count all files in directory
  fest count -r --json ./project   # Directory with JSON output`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(cmd.Context(), args[0], opts)
		},
	}

	cmd.Flags().StringVar(&opts.model, "model", "", "specific model to use for tokenization (gpt-4, gpt-3.5-turbo, claude-3)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "show all counting methods")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.showCost, "cost", false, "include cost estimates")
	cmd.Flags().BoolVarP(&opts.recursive, "recursive", "r", false, "recursively count tokens in directory (respects .gitignore)")
	// Alias: -d / --directory for recursive directory counting
	cmd.Flags().BoolVarP(&opts.recursive, "directory", "d", false, "alias for --recursive: count all files in a directory")
	cmd.Flags().Float64Var(&opts.charsPerToken, "chars-per-token", 4.0, "characters per token ratio for approximation")
	cmd.Flags().Float64Var(&opts.wordsPerToken, "words-per-token", 0.75, "words per token ratio for approximation")

	return cmd
}

func runCount(ctx context.Context, path string, opts *countOptions) error {
	// Create UI handler
	display := ui.New(noColor, verbose)

	// Check if path is a file or directory
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to access path: %w", err)
	}

	var content []byte
	var fileCount int
	isDirectory := info.IsDir()

	if isDirectory {
		// Require recursive flag for directories
		if !opts.recursive {
			return fmt.Errorf("path is a directory; use --recursive flag to count tokens in all files")
		}

		// Walk directory and collect files
		walkResult, err := fileops.WalkDirectory(ctx, path)
		if err != nil {
			return fmt.Errorf("failed to walk directory: %w", err)
		}

		if len(walkResult.Files) == 0 {
			return fmt.Errorf("no text files found in directory")
		}

		if verbose {
			display.Info("Found %d text files (skipped %d binary, %d ignored)",
				len(walkResult.Files), walkResult.SkippedBinary, walkResult.SkippedIgnore)
		}

		// Aggregate all file contents
		content, err = fileops.AggregateFileContents(ctx, walkResult.Files)
		if err != nil {
			return fmt.Errorf("failed to read files: %w", err)
		}

		fileCount = len(walkResult.Files)
	} else {
		// Read single file
		content, err = os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}
		fileCount = 1
	}

	// Create counter with options
	counter := tokens.NewCounter(tokens.CounterOptions{
		CharsPerToken: opts.charsPerToken,
		WordsPerToken: opts.wordsPerToken,
	})

	// Get counts
	result, err := counter.Count(string(content), opts.model, opts.all)
	if err != nil {
		return fmt.Errorf("failed to count tokens: %w", err)
	}

	// Add path info
	result.FilePath = path
	result.FileSize = len(content)
	result.IsDirectory = isDirectory
	if isDirectory {
		result.FileCount = fileCount
	}

	// Add costs if requested
	if opts.showCost {
		result.Costs = tokens.CalculateCosts(result.Methods)
	}

	// Output results
	if opts.jsonOutput {
		return outputJSON(result)
	}

	return outputTable(display, result)
}

func outputJSON(result *tokens.CountResult) error {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}

func outputTable(display *ui.UI, result *tokens.CountResult) error {
	// Header
	if result.IsDirectory {
		display.Info("Token Count Report for: %s (directory)", result.FilePath)
	} else {
		display.Info("Token Count Report for: %s", result.FilePath)
	}
	display.Info("%s", strings.Repeat("═", 55))
	display.Info("")

	// Basic statistics
	display.Info("Basic Statistics:")
	if result.IsDirectory {
		display.Info("  Files:          %d", result.FileCount)
	}
	display.Info("  Characters:     %d", result.Characters)
	display.Info("  Words:          %d", result.Words)
	display.Info("  Lines:          %d", result.Lines)
	display.Info("")

	// Token counts
	display.Info("Token Counts by Method:")
	display.Info("  ┌─────────────────────────┬──────────┬────────────┐")
	display.Info("  │ Method                  │ Tokens   │ Accuracy   │")
	display.Info("  ├─────────────────────────┼──────────┼────────────┤")

	for _, method := range result.Methods {
		accuracy := "Approx"
		if method.IsExact {
			accuracy = "Exact"
		} else if method.Name == "claude_3_approx" {
			accuracy = "Estimated"
		}

		display.Info("  │ %-23s │ %-8d │ %-10s │",
			method.DisplayName, method.Tokens, accuracy)
	}

	display.Info("  └─────────────────────────┴──────────┴────────────┘")

	// Cost estimates
	if len(result.Costs) > 0 {
		display.Info("")
		display.Info("Cost Estimates (Input):")

		for _, cost := range result.Costs {
			display.Info("  %-16s $%.3f ($%.4f/1K tokens)",
				cost.Model+":", cost.Cost, cost.RatePer1K)
		}
	}

	return nil
}
