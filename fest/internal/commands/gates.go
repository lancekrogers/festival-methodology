package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

// NewGatesCommand creates the gates command group
func NewGatesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gates",
		Short: "Manage quality gate policies",
		Long: `Manage hierarchical quality gate policies for festivals.

Quality gates are validation steps that run at the end of implementation
sequences. Gates can be customized at any level: festival, phase, or sequence.

Available Commands:
  show      Show effective gate policy
  list      List available named policies
  apply     Apply a named gate policy
  init      Initialize an override file
  validate  Validate gate configuration`,
	}

	cmd.AddCommand(newGatesShowCmd())
	cmd.AddCommand(newGatesListCmd())
	cmd.AddCommand(newGatesApplyCmd())
	cmd.AddCommand(newGatesInitCmd())
	cmd.AddCommand(newGatesValidateCmd())

	return cmd
}

// --- SHOW COMMAND ---

func newGatesShowCmd() *cobra.Command {
	var phase, sequence string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show effective gate policy",
		Long: `Display the effective gate policy for a festival, phase, or sequence.
Shows which gates are active and where each gate originated from.`,
		Example: `  fest gates show
  fest gates show --phase 002_IMPLEMENT
  fest gates show --sequence 002_IMPLEMENT/01_core --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGatesShow(cmd.Context(), cmd, phase, sequence, jsonOutput)
		},
	}

	cmd.Flags().StringVar(&phase, "phase", "", "Show gates for specific phase")
	cmd.Flags().StringVar(&sequence, "sequence", "", "Show gates for specific sequence (format: phase/sequence)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runGatesShow(ctx context.Context, cmd *cobra.Command, phase, sequence string, jsonOutput bool) error {
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

	registry, err := gates.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return fmt.Errorf("creating policy registry: %w", err)
	}

	loader, err := gates.NewHierarchicalLoader(festivalsRoot, registry)
	if err != nil {
		return fmt.Errorf("creating hierarchical loader: %w", err)
	}

	var effective *gates.EffectivePolicy
	if sequencePath != "" {
		effective, err = loader.LoadForSequence(ctx, festivalPath, phasePath, sequencePath)
	} else if phasePath != "" {
		effective, err = loader.LoadForPhase(ctx, festivalPath, phasePath)
	} else {
		effective, err = loader.LoadForFestival(ctx, festivalPath)
	}
	if err != nil {
		return fmt.Errorf("loading effective policy: %w", err)
	}

	if jsonOutput {
		return printGatesShowJSON(cmd, effective)
	}
	return printGatesShowTable(cmd, effective, phase, sequence)
}

func printGatesShowJSON(cmd *cobra.Command, effective *gates.EffectivePolicy) error {
	output := struct {
		Gates   []gateOutput   `json:"gates"`
		Sources []sourceOutput `json:"sources"`
		Level   string         `json:"level"`
	}{
		Gates:   make([]gateOutput, 0, len(effective.Gates)),
		Sources: make([]sourceOutput, 0, len(effective.Sources)),
		Level:   string(effective.Level),
	}

	for _, gate := range effective.Gates {
		g := gateOutput{
			ID:       gate.ID,
			Template: gate.Template,
			Name:     gate.Name,
			Enabled:  gate.Enabled,
			Removed:  gate.Removed,
		}
		if gate.Source != nil {
			g.Source = string(gate.Source.Level)
		}
		output.Gates = append(output.Gates, g)
	}

	for _, src := range effective.Sources {
		output.Sources = append(output.Sources, sourceOutput{
			Level: string(src.Level),
			Path:  src.Path,
			Name:  src.Name,
		})
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

type gateOutput struct {
	ID       string `json:"id"`
	Template string `json:"template"`
	Name     string `json:"name,omitempty"`
	Enabled  bool   `json:"enabled"`
	Removed  bool   `json:"removed,omitempty"`
	Source   string `json:"source,omitempty"`
}

type sourceOutput struct {
	Level string `json:"level"`
	Path  string `json:"path,omitempty"`
	Name  string `json:"name,omitempty"`
}

func printGatesShowTable(cmd *cobra.Command, effective *gates.EffectivePolicy, phase, sequence string) error {
	out := cmd.OutOrStdout()

	// Header
	location := "festival"
	if sequence != "" {
		location = sequence
	} else if phase != "" {
		location = phase
	}
	fmt.Fprintf(out, "Effective gates for %s:\n\n", location)

	// Table header
	fmt.Fprintf(out, "  %-24s %-12s %-30s\n", "Gate", "Source", "Template")
	fmt.Fprintf(out, "  %-24s %-12s %-30s\n", strings.Repeat("-", 24), strings.Repeat("-", 12), strings.Repeat("-", 30))

	// Gates
	activeGates := effective.GetActiveGates()
	for _, gate := range activeGates {
		source := "builtin"
		if gate.Source != nil {
			source = string(gate.Source.Level)
		}
		fmt.Fprintf(out, "  %-24s [%-10s] %s\n", gate.ID, source, gate.Template)
	}

	if len(activeGates) == 0 {
		fmt.Fprintf(out, "  (no gates active)\n")
	}

	// Show removed gates if any
	for _, gate := range effective.Gates {
		if gate.Removed {
			source := "unknown"
			if gate.Source != nil {
				source = string(gate.Source.Level)
			}
			fmt.Fprintf(out, "  %-24s [%-10s] (removed)\n", gate.ID, source)
		}
	}

	fmt.Fprintln(out)
	return nil
}

// --- LIST COMMAND ---

func newGatesListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available named policies",
		Long:  `Display all available named gate policies that can be applied.`,
		Example: `  fest gates list
  fest gates list --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGatesList(cmd.Context(), cmd, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runGatesList(ctx context.Context, cmd *cobra.Command, jsonOutput bool) error {
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

	policies := registry.ListInfo()

	if jsonOutput {
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(policies)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Available gate policies:\n\n")
	fmt.Fprintf(out, "  %-16s %-12s %s\n", "Name", "Source", "Description")
	fmt.Fprintf(out, "  %-16s %-12s %s\n", strings.Repeat("-", 16), strings.Repeat("-", 12), strings.Repeat("-", 40))

	for _, info := range policies {
		fmt.Fprintf(out, "  %-16s [%-10s] %s\n", info.Name, info.Source, info.Description)
	}

	fmt.Fprintln(out)
	return nil
}

// --- APPLY COMMAND ---

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

	var targetPath, overrideFile string
	if sequencePath != "" {
		targetPath = sequencePath
		overrideFile = gates.PhaseOverrideFileName
	} else if phasePath != "" {
		targetPath = phasePath
		overrideFile = gates.PhaseOverrideFileName
	} else {
		targetPath = festivalPath
		overrideFile = filepath.Join(".festival", "gates.yml")
	}

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

// --- INIT COMMAND ---

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

	var targetPath, overrideFile string
	if sequencePath != "" {
		targetPath = sequencePath
		overrideFile = gates.PhaseOverrideFileName
	} else if phasePath != "" {
		targetPath = phasePath
		overrideFile = gates.PhaseOverrideFileName
	} else {
		targetPath = festivalPath
		overrideFile = filepath.Join(".festival", "gates.yml")
	}

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

// --- VALIDATE COMMAND ---

func newGatesValidateCmd() *cobra.Command {
	var fix bool
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "validate",
		Short: "Validate gate configuration",
		Long:  `Check gate configuration files for errors and inconsistencies.`,
		Example: `  fest gates validate
  fest gates validate --fix
  fest gates validate --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGatesValidate(cmd.Context(), cmd, fix, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&fix, "fix", false, "Attempt to fix issues automatically")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")

	return cmd
}

func runGatesValidate(ctx context.Context, cmd *cobra.Command, fix, jsonOutput bool) error {
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

	var issues []validationIssue

	// Check for override files and validate them
	err = filepath.Walk(festivalsRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip inaccessible paths
		}

		if info.Name() == gates.PhaseOverrideFileName {
			if _, loadErr := gates.LoadPolicy(path); loadErr != nil {
				issues = append(issues, validationIssue{
					Path:     path,
					Severity: "error",
					Message:  fmt.Sprintf("Invalid gate override: %v", loadErr),
				})
			}
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walking directory: %w", err)
	}

	if jsonOutput {
		output := struct {
			Valid  bool              `json:"valid"`
			Issues []validationIssue `json:"issues"`
		}{
			Valid:  len(issues) == 0,
			Issues: issues,
		}
		enc := json.NewEncoder(cmd.OutOrStdout())
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	out := cmd.OutOrStdout()
	if len(issues) == 0 {
		fmt.Fprintf(out, "Gate configuration is valid.\n")
		return nil
	}

	fmt.Fprintf(out, "Found %d issue(s):\n\n", len(issues))
	for _, issue := range issues {
		fmt.Fprintf(out, "  [%s] %s\n    %s\n\n", issue.Severity, issue.Path, issue.Message)
	}

	return nil
}

type validationIssue struct {
	Path     string `json:"path"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}

// --- HELPERS ---

func resolvePaths(festivalsRoot, cwd, phase, sequence string) (festivalPath, phasePath, sequencePath string, err error) {
	// Try to detect current festival from cwd
	festivalPath = findCurrentFestival(festivalsRoot, cwd)
	if festivalPath == "" {
		// Default to first active festival
		activeDir := filepath.Join(festivalsRoot, "active")
		entries, err := os.ReadDir(activeDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
					festivalPath = filepath.Join(activeDir, e.Name())
					break
				}
			}
		}
	}

	if festivalPath == "" {
		return "", "", "", fmt.Errorf("no festival found")
	}

	if sequence != "" {
		// Parse sequence as "phase/sequence"
		parts := strings.SplitN(sequence, "/", 2)
		if len(parts) != 2 {
			return "", "", "", fmt.Errorf("sequence must be in format 'phase/sequence'")
		}
		phasePath = filepath.Join(festivalPath, parts[0])
		sequencePath = filepath.Join(phasePath, parts[1])

		// Verify paths exist
		if _, err := os.Stat(phasePath); os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("phase not found: %s", parts[0])
		}
		if _, err := os.Stat(sequencePath); os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("sequence not found: %s", sequence)
		}
	} else if phase != "" {
		phasePath = filepath.Join(festivalPath, phase)
		if _, err := os.Stat(phasePath); os.IsNotExist(err) {
			return "", "", "", fmt.Errorf("phase not found: %s", phase)
		}
	}

	return festivalPath, phasePath, sequencePath, nil
}

func findCurrentFestival(festivalsRoot, cwd string) string {
	// Check if we're inside a festival directory
	rel, err := filepath.Rel(festivalsRoot, cwd)
	if err != nil {
		return ""
	}

	// Walk up from cwd looking for festival markers
	parts := strings.Split(rel, string(filepath.Separator))
	for i := len(parts); i > 0; i-- {
		candidate := filepath.Join(festivalsRoot, filepath.Join(parts[:i]...))
		// Check for FESTIVAL_GOAL.md or FESTIVAL_OVERVIEW.md
		if _, err := os.Stat(filepath.Join(candidate, "FESTIVAL_GOAL.md")); err == nil {
			return candidate
		}
		if _, err := os.Stat(filepath.Join(candidate, "FESTIVAL_OVERVIEW.md")); err == nil {
			return candidate
		}
	}

	return ""
}

func getConfigRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fest", "active")
}
