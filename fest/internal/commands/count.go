package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/festival-methodology/fest/internal/tokens"
	"github.com/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type countOptions struct {
	model         string
	all           bool
	jsonOutput    bool
	showCost      bool
	charsPerToken float64
	wordsPerToken float64
}

// NewCountCommand creates the count command
func NewCountCommand() *cobra.Command {
	opts := &countOptions{}
	
	cmd := &cobra.Command{
		Use:   "count [file]",
		Short: "Count tokens in a file using various methods",
		Long: `Count tokens in a file using multiple tokenization methods.
		
This command provides token counts using different LLM tokenizers and
approximation methods, helping you understand token usage and estimate costs.`,
		Example: `  fest count document.md           # Count tokens in a file
  fest count --model gpt-4 doc.md  # Use specific model tokenizer
  fest count --all --cost doc.md   # Show all methods with costs
  fest count --json doc.md         # Output as JSON`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCount(args[0], opts)
		},
	}
	
	cmd.Flags().StringVar(&opts.model, "model", "", "specific model to use for tokenization (gpt-4, gpt-3.5-turbo, claude-3)")
	cmd.Flags().BoolVar(&opts.all, "all", false, "show all counting methods")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&opts.showCost, "cost", false, "include cost estimates")
	cmd.Flags().Float64Var(&opts.charsPerToken, "chars-per-token", 4.0, "characters per token ratio for approximation")
	cmd.Flags().Float64Var(&opts.wordsPerToken, "words-per-token", 0.75, "words per token ratio for approximation")
	
	return cmd
}

func runCount(filePath string, opts *countOptions) error {
	// Read file
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	
	// Create UI handler
	display := ui.New(noColor, verbose)
	
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
	
	// Add file info
	result.FilePath = filePath
	result.FileSize = len(content)
	
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
	display.Info("Token Count Report for: %s", result.FilePath)
	display.Info(strings.Repeat("═", 55))
	display.Info("")
	
	// Basic statistics
	display.Info("Basic Statistics:")
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
		} else if method.Name == "Claude-3" {
			accuracy = "Estimated"
		}
		
		display.Info("  │ %-23s │ %-8d │ %-10s │", 
			method.Name, method.Tokens, accuracy)
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