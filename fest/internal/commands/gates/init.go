package gates

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

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
