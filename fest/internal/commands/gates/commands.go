package gates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	gatescore "github.com/lancekrogers/festival-methodology/fest/internal/gates"
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

	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return fmt.Errorf("creating policy registry: %w", err)
	}

	loader, err := gatescore.NewHierarchicalLoader(festivalsRoot, registry)
	if err != nil {
		return fmt.Errorf("creating hierarchical loader: %w", err)
	}

	var effective *gatescore.EffectivePolicy
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

func printGatesShowJSON(cmd *cobra.Command, effective *gatescore.EffectivePolicy) error {
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

func printGatesShowTable(cmd *cobra.Command, effective *gatescore.EffectivePolicy, phase, sequence string) error {
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

	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
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

// Apply and init commands moved to gates_apply.go

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

		if info.Name() == gatescore.PhaseOverrideFileName {
			if _, loadErr := gatescore.LoadPolicy(path); loadErr != nil {
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

// Note: helper types and functions moved to gates_helpers.go
