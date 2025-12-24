package gates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
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
			// Default to dry-run unless approve is set
			if !opts.approve {
				opts.dryRun = true
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
		return emitApplyError(opts, fmt.Errorf("context cancelled: %w", err))
	}

	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	cwd, err := os.Getwd()
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("getting working directory: %w", err))
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("finding festivals root: %w", err))
	}

	// Resolve paths
	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, opts.phase, opts.sequence)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("resolving paths: %w", err))
	}

	// If policy-only mode with a named policy, just write the policy file
	if opts.policyOnly && opts.policyName != "" {
		return runPolicyOnlyApply(ctx, cmd, opts, festivalsRoot, festivalPath, phasePath, sequencePath)
	}

	// Create merger and load effective policy
	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("creating policy registry: %w", err))
	}

	merger, err := gatescore.NewConfigMerger(festivalsRoot, registry)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("creating config merger: %w", err))
	}

	// If a named policy is specified, we need to handle it differently
	var mergedPolicy *gatescore.MergedPolicy
	mergeOpts := gatescore.DefaultMergeOptions()

	if opts.policyName != "" {
		// Get the named policy and use it instead of fest.yaml
		policy, err := registry.GetPolicy(opts.policyName)
		if err != nil {
			return emitApplyError(opts, fmt.Errorf("policy %q not found: %w", opts.policyName, err))
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
			return emitApplyError(opts, fmt.Errorf("loading gate configuration: %w", err))
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
			return emitApplyError(opts, fmt.Errorf("finding sequences: %w", err))
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
		return emitApplyError(opts, fmt.Errorf("finding template root: %w", err))
	}

	// Create generator
	generator, err := gatescore.NewTaskGenerator(tmplRoot)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("creating task generator: %w", err))
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
		results, seqWarnings, err := generator.GenerateForSequence(ctx, seq.Path, activeGates, genOpts)
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
				fmt.Fprintf(out, "  = %s (already exists)\n", r.Path)
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
		return emitApplyError(opts, fmt.Errorf("creating policy registry: %w", err))
	}

	// Verify policy exists
	info, ok := registry.Get(opts.policyName)
	if !ok {
		return emitApplyError(opts, fmt.Errorf("policy %q not found; run 'fest gates list' to see available policies", opts.policyName))
	}

	// Determine target path
	targetPath, overrideFile := resolveTargetPath(festivalPath, phasePath, sequencePath)
	overridePath := filepath.Join(targetPath, overrideFile)

	out := cmd.OutOrStdout()
	if opts.dryRun {
		fmt.Fprintf(out, "[dry-run] Would apply policy %q to %s\n", opts.policyName, overridePath)
		fmt.Fprintf(out, "Policy: %s - %s\n", info.Name, info.Description)
		return nil
	}

	// Get the policy
	policy, err := registry.GetPolicy(opts.policyName)
	if err != nil {
		return emitApplyError(opts, fmt.Errorf("loading policy: %w", err))
	}

	// Ensure parent directory exists for festival-level override
	if sequencePath == "" && phasePath == "" {
		gatesDir := filepath.Join(festivalPath, ".festival")
		if err := os.MkdirAll(gatesDir, 0755); err != nil {
			return emitApplyError(opts, fmt.Errorf("creating .festival directory: %w", err))
		}
	}

	// Write the override file
	if err := gatescore.SavePolicy(overridePath, policy); err != nil {
		return emitApplyError(opts, fmt.Errorf("saving policy: %w", err))
	}

	fmt.Fprintf(out, "Applied policy %q to %s\n", opts.policyName, overridePath)
	return nil
}

func newGatesInitCmd() *cobra.Command {
	var phase, sequence string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a gate configuration file",
		Long: `Create a template configuration file at the specified level.

At festival level, creates fest.yaml with quality gate settings.
At phase/sequence level, creates .fest.gates.yml override file.`,
		Example: `  fest gates init
  fest gates init --phase 002_IMPLEMENT
  fest gates init --sequence 002_IMPLEMENT/01_core`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGatesInit(cmd.Context(), cmd, phase, sequence)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Initialize for specific phase")
	cmd.Flags().StringVar(&sequence, "sequence", "", "Initialize for specific sequence (format: phase/sequence)")

	return cmd
}

func runGatesInit(ctx context.Context, cmd *cobra.Command, phase, sequence string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("getting working directory: %w", err)
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return fmt.Errorf("finding festivals root: %w", err)
	}

	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, phase, sequence)
	if err != nil {
		return fmt.Errorf("resolving paths: %w", err)
	}

	// Determine what to create
	if sequencePath != "" || phasePath != "" {
		// Phase or sequence level: create .fest.gates.yml
		return createPhaseOverrideFile(cmd, festivalPath, phasePath, sequencePath)
	}

	// Festival level: create fest.yaml
	return createFestYAMLFile(cmd, festivalPath)
}

func createPhaseOverrideFile(cmd *cobra.Command, festivalPath, phasePath, sequencePath string) error {
	targetPath, overrideFile := resolveTargetPath(festivalPath, phasePath, sequencePath)
	overridePath := filepath.Join(targetPath, overrideFile)

	// Check if file already exists
	if _, err := os.Stat(overridePath); err == nil {
		return fmt.Errorf("override file already exists: %s", overridePath)
	}

	// Ensure parent directory exists
	parentDir := filepath.Dir(overridePath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	template := `# Gate policy override file
# See: fest understand gates

version: 1
inherit: true  # Set to false to not inherit from parent levels

# Add gates (insert after inherited gates)
# append:
#   - id: security_audit
#     template: SECURITY_AUDIT
#     enabled: true

# Exclude patterns for this level
# exclude_patterns:
#   - "*_docs"
`

	if err := os.WriteFile(overridePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("writing override file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created override file: %s\n", overridePath)
	return nil
}

func createFestYAMLFile(cmd *cobra.Command, festivalPath string) error {
	festYAMLPath := filepath.Join(festivalPath, "fest.yaml")

	// Check if file already exists
	if _, err := os.Stat(festYAMLPath); err == nil {
		return fmt.Errorf("fest.yaml already exists: %s", festYAMLPath)
	}

	template := `# Festival Configuration
# See: fest understand config

version: "1.0"

quality_gates:
  enabled: true
  auto_append: true
  tasks:
    - id: testing_and_verify
      template: QUALITY_GATE_TESTING
      name: Testing and Verification
      enabled: true

    - id: code_review
      template: QUALITY_GATE_REVIEW
      name: Code Review
      enabled: true

    - id: review_results_iterate
      template: QUALITY_GATE_ITERATE
      name: Review Results and Iterate
      enabled: true

excluded_patterns:
  - "*_planning"
  - "*_research"
  - "*_requirements"

templates:
  task_default: TASK_TEMPLATE_SIMPLE
  prefer_simple: true

tracking:
  enabled: true
  checksum_file: .festival-checksums.json
`

	if err := os.WriteFile(festYAMLPath, []byte(template), 0644); err != nil {
		return fmt.Errorf("writing fest.yaml: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created fest.yaml: %s\n", festYAMLPath)
	return nil
}

func emitApplyError(opts *applyOptions, err error) error {
	if opts.jsonOutput {
		return emitApplyResult(opts, applyResult{
			OK:     false,
			Action: "gates_apply",
			Errors: []map[string]any{{
				"code":    "error",
				"message": err.Error(),
			}},
		})
	}
	return err
}

func emitApplyResult(opts *applyOptions, result applyResult) error {
	if opts.jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(result)
	}
	return nil
}
