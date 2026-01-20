package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	gatescore "github.com/lancekrogers/festival-methodology/fest/internal/gates"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type applyOptions struct {
	policyName string // Optional named policy to apply
	phase      string // --phase: target phase
	sequence   string // --sequence: target sequence
	dryRun     bool   // --dry-run (default: true)
	approve    bool   // --approve: actually apply
	force      bool   // --force: overwrite existing files
	jsonOutput bool   // --json: output as JSON
	policyOnly bool   // --policy-only: only write policy file
}

type applyResult struct {
	OK       bool                       `json:"ok"`
	Action   string                     `json:"action"`
	DryRun   bool                       `json:"dry_run"`
	Changes  []gatescore.GenerateResult `json:"changes,omitempty"`
	Summary  gatescore.GenerateSummary  `json:"summary,omitempty"`
	Errors   []map[string]any           `json:"errors,omitempty"`
	Warnings []string                   `json:"warnings,omitempty"`
}

func newGatesApplyCmd() *cobra.Command {
	opts := &applyOptions{}

	cmd := &cobra.Command{
		Use:   "apply [policy]",
		Short: "Apply quality gates to sequences",
		Long: `Apply quality gate task files to implementation sequences.

By default, runs in dry-run mode showing what would change.
Use --approve to actually apply the changes.

Without a policy name, uses fest.yaml and built-in defaults.
With a policy name (default, strict, lightweight), applies that policy.

Quality gates are only added to sequences not matching excluded_patterns.`,
		Example: `  # Preview changes (dry-run is default)
  fest gates apply

  # Apply to all sequences
  fest gates apply --approve

  # Apply strict policy
  fest gates apply strict --approve

  # Apply to specific sequence
  fest gates apply --sequence 002_IMPL/01_core --approve

  # Force overwrite modified files
  fest gates apply --approve --force

  # Only write policy file (no task files)
  fest gates apply strict --policy-only

  # JSON output for automation
  fest gates apply --json`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.policyName = args[0]
			}
			// If approve is set, explicitly disable dry-run
			// (since --dry-run flag defaults to true)
			if opts.approve {
				opts.dryRun = false
			}
			return runGatesApply(cmd.Context(), cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.phase, "phase", "", "Apply to specific phase")
	cmd.Flags().StringVar(&opts.sequence, "sequence", "", "Apply to specific sequence (format: phase/sequence)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", true, "Preview changes without applying (default)")
	cmd.Flags().BoolVar(&opts.approve, "approve", false, "Apply changes")
	cmd.Flags().BoolVar(&opts.force, "force", false, "Overwrite modified files")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "Output JSON")
	cmd.Flags().BoolVar(&opts.policyOnly, "policy-only", false, "Only write policy file, no task files")

	return cmd
}

func runGatesApply(ctx context.Context, cmd *cobra.Command, opts *applyOptions) error {
	if err := ctx.Err(); err != nil {
		return emitApplyError(opts, errors.Wrap(err, "context cancelled").WithOp("runGatesApply"))
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	cwd, err := os.Getwd()
	if err != nil {
		return emitApplyError(opts, errors.IO("getting working directory", err))
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "finding festivals root").WithOp("runGatesApply"))
	}

	// Resolve paths
	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, opts.phase, opts.sequence)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "resolving paths").WithOp("runGatesApply"))
	}

	// If policy-only mode with a named policy, just write the policy file
	if opts.policyOnly && opts.policyName != "" {
		return runPolicyOnlyApply(ctx, cmd, opts, festivalsRoot, festivalPath, phasePath, sequencePath)
	}

	// Create merger and load effective policy
	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "creating policy registry").WithOp("runGatesApply"))
	}

	merger, err := gatescore.NewConfigMerger(festivalsRoot, registry)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "creating config merger").WithOp("runGatesApply"))
	}

	// If a named policy is specified, we need to handle it differently
	var mergedPolicy *gatescore.MergedPolicy
	mergeOpts := gatescore.DefaultMergeOptions()

	if opts.policyName != "" {
		// Get the named policy and use it instead of fest.yaml
		policy, err := registry.GetPolicy(opts.policyName)
		if err != nil {
			return emitApplyError(opts, errors.Wrap(err, "policy not found").WithField("policy", opts.policyName))
		}

		// Use named policy gates
		mergedPolicy = &gatescore.MergedPolicy{
			Gates:           policy.GetEnabledTasks(),
			ExcludePatterns: policy.ExcludePatterns,
			FestYAMLEnabled: true,
		}
	} else {
		// Merge from fest.yaml and hierarchy
		if sequencePath != "" {
			mergedPolicy, err = merger.MergeForSequence(ctx, festivalPath, phasePath, sequencePath, mergeOpts)
		} else if phasePath != "" {
			mergedPolicy, err = merger.MergeForPhase(ctx, festivalPath, phasePath, mergeOpts)
		} else {
			mergedPolicy, err = merger.MergeForFestival(ctx, festivalPath, mergeOpts)
		}
		if err != nil {
			return emitApplyError(opts, errors.Wrap(err, "loading gate configuration").WithOp("runGatesApply"))
		}
	}

	// Get active gates
	activeGates := mergedPolicy.GetActiveGates()
	if len(activeGates) == 0 {
		return emitApplyResult(opts, applyResult{
			OK:       true,
			Action:   "gates_apply",
			DryRun:   opts.dryRun,
			Warnings: []string{"No active quality gates configured"},
		})
	}

	// Find sequences to process
	var sequences []gatescore.SequenceInfo
	if sequencePath != "" {
		// Single sequence
		sequences = []gatescore.SequenceInfo{{
			Path:      sequencePath,
			PhasePath: phasePath,
			Name:      filepath.Base(sequencePath),
		}}
	} else {
		// Find all implementation sequences
		sequences, err = gatescore.FindSequencesWithInfo(festivalPath, mergedPolicy.ExcludePatterns)
		if err != nil {
			return emitApplyError(opts, errors.Wrap(err, "finding sequences").WithOp("runGatesApply"))
		}
	}

	if len(sequences) == 0 {
		return emitApplyResult(opts, applyResult{
			OK:       true,
			Action:   "gates_apply",
			DryRun:   opts.dryRun,
			Warnings: []string{"No implementation sequences found"},
		})
	}

	// Get template root
	tmplRoot, err := tpl.LocalTemplateRoot(festivalPath)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "finding template root").WithOp("runGatesApply"))
	}

	// Load festival config for gate ordering
	festCfg, err := config.LoadFestivalConfig(festivalPath)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "loading festival config").WithOp("runGatesApply"))
	}

	// Create generator
	generator, err := gatescore.NewTaskGenerator(ctx, tmplRoot)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "creating task generator").WithOp("runGatesApply"))
	}

	// Process each sequence
	var allResults []gatescore.GenerateResult
	var warnings []string
	summary := gatescore.GenerateSummary{TotalSequences: len(sequences)}

	genOpts := gatescore.GenerateOptions{
		DryRun:  opts.dryRun,
		Force:   opts.force,
		Verbose: shared.IsVerbose(),
	}

	for _, seq := range sequences {
		// Phase type is required - error if not detected
		if seq.PhaseType == "" {
			errMsg := fmt.Sprintf("Phase %s: no phase type detected - add fest_phase_type to PHASE_GOAL.md frontmatter or use a phase name that indicates the type", seq.PhaseName)
			warnings = append(warnings, errMsg)
			continue
		}

		// Get gates in configured order for this phase type
		var sequenceGates []gatescore.GateTask
		configGates := festCfg.GetGatesForPhaseType(seq.PhaseType)
		if len(configGates) > 0 {
			// Use configured gate ordering
			sequenceGates = make([]gatescore.GateTask, len(configGates))
			for i, qt := range configGates {
				sequenceGates[i] = gatescore.GateTaskFromQualityGateTask(qt)
			}
			if shared.IsVerbose() {
				fmt.Fprintf(cmd.OutOrStdout(), "  Phase %s: using %d configured %s gates\n", seq.PhaseName, len(sequenceGates), seq.PhaseType)
			}
		} else {
			// Fallback to filesystem discovery if no config
			var discoverErr error
			sequenceGates, discoverErr = gatescore.DiscoverGatesForPhaseType(tmplRoot, seq.PhaseType)
			if discoverErr != nil {
				warnings = append(warnings, fmt.Sprintf("Phase %s: %v", seq.PhaseName, discoverErr))
				continue
			}
			if shared.IsVerbose() {
				fmt.Fprintf(cmd.OutOrStdout(), "  Phase %s: discovered %d %s gates (fallback)\n", seq.PhaseName, len(sequenceGates), seq.PhaseType)
			}
		}

		// Skip sequences in phases with no gates
		if len(sequenceGates) == 0 {
			if shared.IsVerbose() {
				fmt.Fprintf(cmd.OutOrStdout(), "  Skipping %s (no gates for %s phase)\n", seq.Name, seq.PhaseType)
			}
			continue
		}

		results, seqWarnings, err := generator.GenerateForSequence(ctx, seq.Path, sequenceGates, genOpts, festivalPath)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("Sequence %s: %v", seq.Name, err))
			continue
		}

		if len(results) > 0 {
			summary.SequencesUpdated++
		}

		for _, r := range results {
			if r.Type == "create" {
				summary.FilesCreated++
			} else if r.Type == "skip" {
				summary.FilesSkipped++
			}
		}

		allResults = append(allResults, results...)
		warnings = append(warnings, seqWarnings...)
	}

	// Output result
	result := applyResult{
		OK:       true,
		Action:   "gates_apply",
		DryRun:   opts.dryRun,
		Changes:  allResults,
		Summary:  summary,
		Warnings: warnings,
	}

	if opts.jsonOutput {
		return emitApplyResult(opts, result)
	}

	// Human-readable output
	out := cmd.OutOrStdout()

	if opts.dryRun {
		display.Info("Dry-run mode (use --approve to apply changes)")
	}

	display.Info("Found %d sequences, %d will be updated", summary.TotalSequences, summary.SequencesUpdated)

	for _, r := range allResults {
		switch r.Type {
		case "create":
			display.Success("  + %s", r.Path)
		case "skip":
			display.Warning("  ~ Skipped %s (%s)", r.Path, r.Reason)
		case "exists":
			if shared.IsVerbose() {
				display.Info("  = %s (already exists)", r.Path)
			}
		}
	}

	for _, w := range warnings {
		display.Warning("  Warning: %s", w)
	}

	fmt.Fprintln(out)
	display.Info("Summary: %d files created, %d skipped", summary.FilesCreated, summary.FilesSkipped)

	return nil
}

// runPolicyOnlyApply writes just the policy file (original behavior).
func runPolicyOnlyApply(ctx context.Context, cmd *cobra.Command, opts *applyOptions, festivalsRoot, festivalPath, phasePath, sequencePath string) error {
	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "creating policy registry").WithOp("runPolicyOnlyApply"))
	}

	// Verify policy exists
	info, ok := registry.Get(opts.policyName)
	if !ok {
		return emitApplyError(opts, errors.NotFound("policy").WithField("policy", opts.policyName).
			WithField("hint", "run 'fest gates list' to see available policies"))
	}

	// Determine target path
	targetPath, overrideFile := resolveTargetPath(festivalPath, phasePath, sequencePath)
	overridePath := filepath.Join(targetPath, overrideFile)

	out := cmd.OutOrStdout()
	if opts.dryRun {
		fmt.Fprintln(out, ui.H1("Gate Policy (Dry Run)"))
		fmt.Fprintf(out, "%s %s\n", ui.Label("Policy"), ui.Value(info.Name))
		if info.Description != "" {
			fmt.Fprintf(out, "%s %s\n", ui.Label("Description"), ui.Dim(info.Description))
		}
		fmt.Fprintf(out, "%s %s\n", ui.Label("Target"), ui.Dim(overridePath))
		return nil
	}

	// Get the policy
	policy, err := registry.GetPolicy(opts.policyName)
	if err != nil {
		return emitApplyError(opts, errors.Wrap(err, "loading policy").WithField("policy", opts.policyName))
	}

	// Ensure parent directory exists for festival-level override
	if sequencePath == "" && phasePath == "" {
		gatesDir := filepath.Join(festivalPath, ".festival")
		if err := os.MkdirAll(gatesDir, 0755); err != nil {
			return emitApplyError(opts, errors.IO("creating .festival directory", err).WithField("path", gatesDir))
		}
	}

	// Write the override file
	if err := gatescore.SavePolicy(overridePath, policy); err != nil {
		return emitApplyError(opts, errors.Wrap(err, "saving policy").WithField("path", overridePath))
	}

	fmt.Fprintln(out, ui.Success("✓ Gate policy applied"))
	fmt.Fprintf(out, "%s %s\n", ui.Label("Policy"), ui.Value(info.Name))
	fmt.Fprintf(out, "%s %s\n", ui.Label("Path"), ui.Dim(overridePath))
	return nil
}
