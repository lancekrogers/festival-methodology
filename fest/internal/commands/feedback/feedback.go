package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/feedback"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewFeedbackCommand creates the feedback command group
func NewFeedbackCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "feedback",
		Short: "Manage structured feedback collection",
		Long: `Collect and manage structured feedback during festival execution.

Feedback allows agents to record observations based on defined criteria
for later aggregation and analysis.

Examples:
  fest feedback init --criteria "Code quality" --criteria "Performance"
  fest feedback add --criteria "Code quality" --observation "Found duplication"
  fest feedback view
  fest feedback export --format markdown`,
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newAddCmd())
	cmd.AddCommand(newViewCmd())
	cmd.AddCommand(newExportCmd())

	return cmd
}

func newInitCmd() *cobra.Command {
	var criteria []string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize feedback collection",
		Long: `Initialize feedback collection with defined criteria.

Creates a feedback/ directory in the current festival with
configuration for the specified criteria.

Examples:
  fest feedback init --criteria "Code quality observations"
  fest feedback init --criteria "Performance concerns" --criteria "Methodology suggestions"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), criteria)
		},
	}

	cmd.Flags().StringSliceVar(&criteria, "criteria", nil, "feedback criteria (required)")
	_ = cmd.MarkFlagRequired("criteria")

	return cmd
}

func runInit(ctx context.Context, criteria []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := feedback.NewStore(festivalPath)

	if store.IsInitialized() {
		return errors.Validation("feedback already initialized").
			WithField("hint", "run 'fest feedback view' to see existing feedback")
	}

	config, err := store.Init(ctx, criteria)
	if err != nil {
		return err
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	display.Success("Initialized feedback collection")
	fmt.Printf("\nCriteria:\n")
	for _, c := range config.Criteria {
		fmt.Printf("  - %s\n", c.Name)
	}
	fmt.Printf("\nAdd feedback: fest feedback add --criteria \"...\" --observation \"...\"\n")

	return nil
}

func newAddCmd() *cobra.Command {
	var (
		criteria    string
		observation string
		jsonInput   string
		task        string
		severity    string
		suggestion  string
	)

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a feedback observation",
		Long: `Add a feedback observation for a defined criteria.

Use either flags or JSON input to add an observation.

Examples:
  fest feedback add --criteria "Code quality" --observation "Found duplicate logic"
  fest feedback add --json '{"criteria": "Performance", "observation": "N+1 query"}'
  fest feedback add --criteria "Code quality" --observation "..." --severity high --suggestion "Refactor"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(cmd.Context(), criteria, observation, jsonInput, task, severity, suggestion)
		},
	}

	cmd.Flags().StringVar(&criteria, "criteria", "", "criteria category")
	cmd.Flags().StringVar(&observation, "observation", "", "observation text")
	cmd.Flags().StringVar(&jsonInput, "json", "", "JSON observation object")
	cmd.Flags().StringVar(&task, "task", "", "related task path")
	cmd.Flags().StringVar(&severity, "severity", "", "severity: low, medium, high")
	cmd.Flags().StringVar(&suggestion, "suggestion", "", "suggested action")

	return cmd
}

func runAdd(ctx context.Context, criteria, observation, jsonInput, task, severity, suggestion string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := feedback.NewStore(festivalPath)

	var obs *feedback.Observation

	if jsonInput != "" {
		// Parse JSON input
		obs, err = feedback.ParseObservationJSON(jsonInput)
		if err != nil {
			return err
		}
	} else {
		// Use flags
		if criteria == "" || observation == "" {
			return errors.Validation("criteria and observation are required").
				WithField("hint", "use --criteria and --observation, or --json")
		}
		obs = &feedback.Observation{
			Criteria:    criteria,
			Observation: observation,
			Task:        task,
			Severity:    severity,
			Suggestion:  suggestion,
		}
	}

	if err := store.AddObservation(ctx, obs); err != nil {
		return err
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	display.Success("Added observation %s", obs.ID)

	return nil
}

func newViewCmd() *cobra.Command {
	var (
		criteria   string
		severity   string
		jsonOutput bool
		summary    bool
	)

	cmd := &cobra.Command{
		Use:   "view",
		Short: "View collected feedback",
		Long: `View collected feedback observations.

Filter by criteria or severity, or view a summary of all feedback.

Examples:
  fest feedback view
  fest feedback view --criteria "Code quality"
  fest feedback view --severity high
  fest feedback view --summary
  fest feedback view --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runView(cmd.Context(), criteria, severity, jsonOutput, summary)
		},
	}

	cmd.Flags().StringVar(&criteria, "criteria", "", "filter by criteria")
	cmd.Flags().StringVar(&severity, "severity", "", "filter by severity")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVar(&summary, "summary", false, "show summary only")

	return cmd
}

func runView(ctx context.Context, criteria, severity string, jsonOutput, summary bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := feedback.NewStore(festivalPath)

	observations, err := store.ListObservations(ctx, criteria, severity)
	if err != nil {
		return err
	}

	if len(observations) == 0 {
		display := ui.New(shared.IsNoColor(), shared.IsVerbose())
		display.Info("No feedback collected yet. Add some with 'fest feedback add'")
		return nil
	}

	if jsonOutput {
		data, _ := json.MarshalIndent(observations, "", "  ")
		fmt.Println(string(data))
		return nil
	}

	if summary {
		// Show summary by criteria
		byCategory := make(map[string]int)
		for _, obs := range observations {
			byCategory[obs.Criteria]++
		}
		fmt.Printf("Feedback Summary (%d total):\n\n", len(observations))
		for cat, count := range byCategory {
			fmt.Printf("  %s: %d observations\n", cat, count)
		}
		return nil
	}

	// Full view
	fmt.Printf("Feedback (%d observations):\n\n", len(observations))
	for _, obs := range observations {
		fmt.Printf("[%s] %s\n", obs.ID, obs.Criteria)
		fmt.Printf("    %s\n", obs.Observation)
		if obs.Task != "" {
			fmt.Printf("    Task: %s\n", obs.Task)
		}
		if obs.Severity != "" {
			fmt.Printf("    Severity: %s\n", obs.Severity)
		}
		if obs.Suggestion != "" {
			fmt.Printf("    Suggestion: %s\n", obs.Suggestion)
		}
		fmt.Println()
	}

	return nil
}

func newExportCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export collected feedback",
		Long: `Export collected feedback to a file format.

Supports markdown, JSON, and YAML output formats.

Examples:
  fest feedback export --format markdown > report.md
  fest feedback export --format json > report.json
  fest feedback export --format yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(cmd.Context(), format)
		},
	}

	cmd.Flags().StringVar(&format, "format", "markdown", "output format: markdown, json, yaml")

	return cmd
}

func runExport(ctx context.Context, format string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	festivalPath, err := shared.ResolveFestivalPath(cwd, "")
	if err != nil {
		return errors.Wrap(err, "detecting festival").
			WithField("hint", "run from within a festival directory")
	}

	store := feedback.NewStore(festivalPath)

	output, err := store.Export(ctx, format)
	if err != nil {
		return err
	}

	fmt.Print(output)
	return nil
}
