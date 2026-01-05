// Package execute provides the fest execute command for festival orchestration.
package execute

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/execute"
	"github.com/spf13/cobra"
)

var (
	dryRun       bool
	agentMode    bool
	jsonOutput   bool
	maxParallel  int
	continueExec bool
	phaseName    string
	seqName      string
	reset        bool
)

// NewExecuteCommand creates the execute command
func NewExecuteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute",
		Short: "Execute festival tasks with orchestration",
		Long: `Orchestrate festival execution with support for parallel tasks,
quality gates, and execution modes.

The execute command builds an execution plan from the festival structure,
tracks progress, and provides instructions for completing tasks.

Examples:
  fest execute --dry-run          # Preview execution plan
  fest execute --agent            # Agent-friendly output
  fest execute --json             # JSON output
  fest execute --parallel 4       # Limit parallel execution
  fest execute --phase 01         # Execute specific phase
  fest execute --continue         # Resume from saved state
  fest execute --reset            # Clear saved state`,
		RunE: runExecute,
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview execution plan without executing")
	cmd.Flags().BoolVar(&agentMode, "agent", false, "output agent-friendly instructions")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output as JSON")
	cmd.Flags().IntVar(&maxParallel, "parallel", 1, "maximum parallel tasks")
	cmd.Flags().BoolVar(&continueExec, "continue", false, "continue from saved state")
	cmd.Flags().StringVar(&phaseName, "phase", "", "execute specific phase")
	cmd.Flags().StringVar(&seqName, "sequence", "", "execute specific sequence")
	cmd.Flags().BoolVar(&reset, "reset", false, "clear saved execution state")

	return cmd
}

func runExecute(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Resolve festival path (supports linked festivals via fest link)
	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "not inside a festival")
	}

	// Handle reset
	if reset {
		stateManager := execute.NewStateManager(festivalPath)
		if err := stateManager.Clear(ctx); err != nil {
			return errors.IO("clearing execution state", err)
		}
		fmt.Println("Execution state cleared.")
		return nil
	}

	// Create config
	config := execute.DefaultConfig()
	config.DryRun = dryRun

	// Create runner
	runner := execute.NewRunner(festivalPath, config)

	// Initialize
	if err := runner.Initialize(ctx); err != nil {
		return errors.Wrap(err, "initializing execution runner")
	}

	// Handle different output modes
	if dryRun {
		if jsonOutput {
			return outputJSON(runner.GetPlan())
		}
		fmt.Print(runner.FormatDryRun())
		return nil
	}

	if agentMode {
		instructions, err := runner.FormatAgentInstructions()
		if err != nil {
			return errors.Wrap(err, "formatting agent instructions")
		}
		fmt.Print(instructions)
		return nil
	}

	if jsonOutput {
		return outputJSON(runner.GetPlan())
	}

	// Default: show plan overview and current status
	return showStatus(runner)
}

func outputJSON(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return errors.Parse("formatting JSON", err)
	}
	fmt.Println(string(output))
	return nil
}

func showStatus(runner *execute.Runner) error {
	state := runner.GetState()
	plan := runner.GetPlan()

	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println("                    EXECUTION STATUS                        ")
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()

	fmt.Printf("Festival: %s\n\n", plan.FestivalPath)

	fmt.Println("PROGRESS")
	fmt.Println("────────")
	fmt.Printf("  Total Tasks:     %d\n", state.TotalTasks)
	fmt.Printf("  Completed:       %d\n", state.CompletedTasks)
	fmt.Printf("  Skipped:         %d\n", state.SkippedTasks)
	fmt.Printf("  Failed:          %d\n", state.FailedTasks)
	fmt.Printf("  Pending:         %d\n", state.PendingTasks())
	fmt.Printf("  Progress:        %.1f%%\n", state.Progress())
	fmt.Println()

	// Show next step
	step, seq, phase, err := runner.GetNextStep()
	if err != nil {
		return err
	}

	if step == nil {
		fmt.Println("✓ All tasks completed!")
		return nil
	}

	fmt.Println("NEXT STEP")
	fmt.Println("─────────")
	fmt.Printf("  Phase:    %s\n", phase.Name)
	fmt.Printf("  Sequence: %s\n", seq.Name)
	fmt.Printf("  Step:     %d\n", step.Number)
	fmt.Println()

	fmt.Println("  Tasks:")
	for _, task := range step.Tasks {
		fmt.Printf("    • %s\n", task.Name)
	}
	fmt.Println()

	fmt.Println("COMMANDS")
	fmt.Println("────────")
	fmt.Println("  fest execute --agent      # Get agent instructions")
	fmt.Println("  fest execute --dry-run    # Preview full plan")
	fmt.Println("  fest execute --json       # Export as JSON")

	return nil
}
