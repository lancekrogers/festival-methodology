package commands

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

func newGatesApplyCmd() *cobra.Command {
	var phase, sequence string
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "apply <policy>",
		Short: "Apply a named gate policy",
		Long:  `Apply a named gate policy at the specified level (festival, phase, or sequence).`,
		Example: `  fest gates apply strict
  fest gates apply lightweight --phase 001_RESEARCH
  fest gates apply strict --sequence 002_IMPLEMENT/01_core --dry-run`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGatesApply(cmd.Context(), cmd, args[0], phase, sequence, dryRun)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Apply to specific phase")
	cmd.Flags().StringVar(&sequence, "sequence", "", "Apply to specific sequence (format: phase/sequence)")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be done without making changes")

	return cmd
}

func runGatesApply(ctx context.Context, cmd *cobra.Command, policyName, phase, sequence string, dryRun bool) error {
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

	registry, err := gates.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return fmt.Errorf("creating policy registry: %w", err)
	}

	// Verify policy exists
	info, ok := registry.Get(policyName)
	if !ok {
		return fmt.Errorf("policy %q not found; run 'fest gates list' to see available policies", policyName)
	}

	// Determine target path
	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, phase, sequence)
	if err != nil {
		return fmt.Errorf("resolving paths: %w", err)
	}

	targetPath, overrideFile := resolveTargetPath(festivalPath, phasePath, sequencePath)
	overridePath := filepath.Join(targetPath, overrideFile)

	out := cmd.OutOrStdout()
	if dryRun {
		fmt.Fprintf(out, "[dry-run] Would apply policy %q to %s\n", policyName, overridePath)
		fmt.Fprintf(out, "Policy: %s - %s\n", info.Name, info.Description)
		return nil
	}

	// Get the policy
	policy, err := registry.GetPolicy(policyName)
	if err != nil {
		return fmt.Errorf("loading policy: %w", err)
	}

	// Ensure parent directory exists for festival-level override
	if sequencePath == "" && phasePath == "" {
		gatesDir := filepath.Join(festivalPath, ".festival")
		if err := os.MkdirAll(gatesDir, 0755); err != nil {
			return fmt.Errorf("creating .festival directory: %w", err)
		}
	}

	// Write the override file
	if err := gates.SavePolicy(overridePath, policy); err != nil {
		return fmt.Errorf("saving policy: %w", err)
	}

	fmt.Fprintf(out, "Applied policy %q to %s\n", policyName, overridePath)
	return nil
}

func newGatesInitCmd() *cobra.Command {
	var phase, sequence string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize an override file",
		Long:  `Create a template override file at the specified level for customization.`,
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

	// Create template content
	template := `# Gate policy override file
# See: fest understand gates

version: 1
inherit: true  # Set to false to not inherit from parent levels

# Add gates (insert after inherited gates)
# add:
#   - id: security_audit
#     template: SECURITY_AUDIT
#     after: code_review

# Remove gates from inherited set
# remove:
#   - testing_and_verify

# Replace all inherited gates (mutually exclusive with add/remove)
# replace:
#   - id: custom_only
#     template: CUSTOM_GATE
`

	if err := os.WriteFile(overridePath, []byte(template), 0644); err != nil {
		return fmt.Errorf("writing override file: %w", err)
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created override file: %s\n", overridePath)
	return nil
}
