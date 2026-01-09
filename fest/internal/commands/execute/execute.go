// Package execute provides the fest execute command for festival orchestration.
package execute

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/execute"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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
		fmt.Println(ui.Success("Execution state cleared."))
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
	if err := shared.EncodeJSON(os.Stdout, data); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

func showStatus(runner *execute.Runner) error {
	state := runner.GetState()
	plan := runner.GetPlan()

	fmt.Println(ui.H1("Execution Status"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(filepath.Base(plan.FestivalPath), ui.FestivalColor))
	fmt.Printf("%s %s\n", ui.Label("Path"), ui.Dim(plan.FestivalPath))
	fmt.Println()

	fmt.Println(ui.H2("Progress"))
	fmt.Printf("%s %s\n", ui.Label("Total tasks"), ui.Value(fmt.Sprintf("%d", state.TotalTasks)))
	fmt.Printf("%s %s\n", ui.Label("Completed"), ui.Value(fmt.Sprintf("%d", state.CompletedTasks), ui.SuccessColor))
	fmt.Printf("%s %s\n", ui.Label("Skipped"), ui.Dim(fmt.Sprintf("%d", state.SkippedTasks)))
	fmt.Printf("%s %s\n", ui.Label("Failed"), ui.Value(fmt.Sprintf("%d", state.FailedTasks), ui.ErrorColor))
	fmt.Printf("%s %s\n", ui.Label("Pending"), ui.Value(fmt.Sprintf("%d", state.PendingTasks()), ui.PendingColor))
	fmt.Printf("%s %s\n", ui.Label("Progress"), ui.Value(fmt.Sprintf("%.1f%%", state.Progress())))
	fmt.Println()

	// Show next step
	step, seq, phase, err := runner.GetNextStep()
	if err != nil {
		return err
	}

	if step == nil {
		fmt.Println(ui.Success("All tasks completed."))
		return nil
	}

	fmt.Println(ui.H2("Next Step"))
	fmt.Printf("%s %s\n", ui.Label("Phase"), ui.Value(phase.Name, ui.PhaseColor))
	fmt.Printf("%s %s\n", ui.Label("Sequence"), ui.Value(seq.Name, ui.SequenceColor))
	fmt.Printf("%s %s\n", ui.Label("Step"), ui.Value(fmt.Sprintf("%d", step.Number)))
	fmt.Println()

	fmt.Println(ui.H3("Tasks"))
	for _, task := range step.Tasks {
		fmt.Printf("  - %s\n", ui.Value(task.Name, ui.TaskColor))
	}
	fmt.Println()

	fmt.Println(ui.H2("Commands"))
	fmt.Printf("  %s %s\n", ui.Value("fest execute --agent"), ui.Dim("# Get agent instructions"))
	fmt.Printf("  %s %s\n", ui.Value("fest execute --dry-run"), ui.Dim("# Preview full plan"))
	fmt.Printf("  %s %s\n", ui.Value("fest execute --json"), ui.Dim("# Export as JSON"))

	return nil
}
