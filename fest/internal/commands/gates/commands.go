package gates

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	gatescore "github.com/lancekrogers/festival-methodology/fest/internal/gates"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

// NewGatesCommand creates the gates command group
func NewGatesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gates",
		Short: "Manage quality gates - validation steps at sequence end",
		Long: `Manage hierarchical quality gate policies for festivals.

Quality gates are validation steps that run at the end of implementation
sequences. Gates can be customized at any level: festival, phase, or sequence.

Available Commands:
  show      Show effective gate policy
  list      List available named policies
  apply     Apply quality gates to sequences
  remove    Remove quality gate files from sequences
  init      Initialize an override file
  validate  Validate gate configuration`,
	}

	cmd.AddCommand(newGatesShowCmd())
	cmd.AddCommand(newGatesListCmd())
	cmd.AddCommand(newGatesApplyCmd())
	cmd.AddCommand(newGatesRemoveCmd())
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
		return errors.Wrap(err, "context cancelled").WithOp("runGatesShow")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting working directory", err)
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals root").WithOp("runGatesShow")
	}

	festivalPath, phasePath, sequencePath, err := resolvePaths(festivalsRoot, cwd, phase, sequence)
	if err != nil {
		return errors.Wrap(err, "resolving paths").WithOp("runGatesShow")
	}

	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return errors.Wrap(err, "creating policy registry").WithOp("runGatesShow")
	}

	// Use ConfigMerger to show merged configuration from fest.yaml + policy files
	merger, err := gatescore.NewConfigMerger(festivalsRoot, registry)
	if err != nil {
		return errors.Wrap(err, "creating config merger").WithOp("runGatesShow")
	}

	opts := gatescore.DefaultMergeOptions()
	var merged *gatescore.MergedPolicy
	if sequencePath != "" {
		merged, err = merger.MergeForSequence(ctx, festivalPath, phasePath, sequencePath, opts)
	} else if phasePath != "" {
		merged, err = merger.MergeForPhase(ctx, festivalPath, phasePath, opts)
	} else {
		merged, err = merger.MergeForFestival(ctx, festivalPath, opts)
	}
	if err != nil {
		return errors.Wrap(err, "loading merged policy").WithOp("runGatesShow")
	}

	if jsonOutput {
		return printGatesShowMergedJSON(cmd, merged)
	}
	return printGatesShowMergedTable(cmd, merged, phase, sequence)
}

func printGatesShowMergedJSON(cmd *cobra.Command, merged *gatescore.MergedPolicy) error {
	output := struct {
		Gates           []gateOutput   `json:"gates"`
		Sources         []sourceOutput `json:"sources"`
		Level           string         `json:"level"`
		FestYAMLEnabled bool           `json:"fest_yaml_enabled"`
		ExcludePatterns []string       `json:"exclude_patterns,omitempty"`
	}{
		Gates:           make([]gateOutput, 0, len(merged.Gates)),
		Sources:         make([]sourceOutput, 0, len(merged.Sources)),
		Level:           string(merged.Level),
		FestYAMLEnabled: merged.FestYAMLEnabled,
		ExcludePatterns: merged.ExcludePatterns,
	}

	for _, gate := range merged.Gates {
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

	for _, src := range merged.Sources {
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

func printGatesShowMergedTable(cmd *cobra.Command, merged *gatescore.MergedPolicy, phase, sequence string) error {
	out := cmd.OutOrStdout()

	// Header
	location := "festival"
	if sequence != "" {
		location = sequence
	} else if phase != "" {
		location = phase
	}
	fmt.Fprintln(out, ui.H1("Gate Policy"))
	fmt.Fprintf(out, "%s %s\n", ui.Label("Scope"), ui.Value(location))
	fmt.Fprintln(out, ui.Dim(strings.Repeat("─", 60)))

	// Show configuration sources
	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.H2("Configuration Sources"))
	if len(merged.Sources) == 0 {
		fmt.Fprintln(out, ui.Dim("No configuration sources found."))
	} else {
		for _, src := range merged.Sources {
			target := src.Path
			if target == "" {
				target = src.Name
			}
			fmt.Fprintf(out, "%s %s\n", ui.Dim(fmt.Sprintf("[%s]", src.Level)), ui.Dim(target))
		}
	}

	// Active gates
	fmt.Fprintln(out)
	fmt.Fprintln(out, ui.H2("Active Gates"))

	// Gates
	activeGates := merged.GetActiveGates()
	if len(activeGates) == 0 {
		fmt.Fprintln(out, ui.Dim("No active gates."))
	}
	for _, gate := range activeGates {
		source := "builtin"
		if gate.Source != nil {
			source = string(gate.Source.Level)
		}
		fmt.Fprintf(out, "%s %s %s\n",
			ui.Value(gate.ID, ui.GateColor),
			ui.Dim(fmt.Sprintf("[%s]", source)),
			ui.Dim(gate.Template))
	}

	// Show removed gates if any
	hasRemoved := false
	for _, gate := range merged.Gates {
		if gate.Removed {
			if !hasRemoved {
				fmt.Fprintln(out)
				fmt.Fprintln(out, ui.H2("Removed Gates"))
				hasRemoved = true
			}
			source := "unknown"
			if gate.Source != nil {
				source = string(gate.Source.Level)
			}
			fmt.Fprintf(out, "%s %s\n",
				ui.Warning(gate.ID),
				ui.Dim(fmt.Sprintf("removed at %s level", source)))
		}
	}

	// Show exclude patterns if any
	if len(merged.ExcludePatterns) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, ui.H2("Exclude Patterns"))
		for _, pattern := range merged.ExcludePatterns {
			fmt.Fprintf(out, "%s %s\n", ui.Dim("•"), ui.Dim(pattern))
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
		return errors.Wrap(err, "context cancelled").WithOp("runGatesList")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting working directory", err)
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals root").WithOp("runGatesList")
	}

	registry, err := gatescore.NewPolicyRegistry(festivalsRoot, getConfigRoot())
	if err != nil {
		return errors.Wrap(err, "creating policy registry").WithOp("runGatesList")
	}

	policies := registry.ListInfo()

	// Try to find festival root to discover local templates
	var localTemplates []string
	var gatesDir string
	festivalRoot, findErr := tpl.FindFestivalRoot(cwd)
	if findErr == nil {
		gatesDir = filepath.Join(festivalRoot, "gates")
		localTemplates = discoverLocalGateTemplates(gatesDir)
	}

	if jsonOutput {
		return printGatesListJSON(cmd, policies, localTemplates, gatesDir)
	}

	return printGatesListTable(cmd, policies, localTemplates, gatesDir)
}

// discoverLocalGateTemplates finds all .md files in a gates directory
func discoverLocalGateTemplates(gatesDir string) []string {
	var templates []string

	entries, err := os.ReadDir(gatesDir)
	if err != nil {
		return templates // Return empty if directory doesn't exist
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".md") {
			// Return without .md extension
			templates = append(templates, strings.TrimSuffix(name, ".md"))
		}
	}

	return templates
}

func printGatesListJSON(cmd *cobra.Command, policies []*gatescore.PolicyInfo, localTemplates []string, gatesDir string) error {
	output := struct {
		Policies       []*gatescore.PolicyInfo `json:"policies"`
		LocalTemplates []string                `json:"local_templates,omitempty"`
		GatesDirectory string                  `json:"gates_directory,omitempty"`
	}{
		Policies:       policies,
		LocalTemplates: localTemplates,
	}

	if len(localTemplates) > 0 {
		output.GatesDirectory = gatesDir
	}

	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func printGatesListTable(cmd *cobra.Command, policies []*gatescore.PolicyInfo, localTemplates []string, gatesDir string) error {
	out := cmd.OutOrStdout()

	// Show named policies
	fmt.Fprintln(out, ui.H1("Gate Policies"))
	if len(policies) == 0 {
		fmt.Fprintln(out, ui.Dim("No gate policies available."))
	} else {
		fmt.Fprintln(out, ui.H2("Named Policies"))
		for _, info := range policies {
			fmt.Fprintf(out, "%s %s\n",
				ui.Value(info.Name),
				ui.Dim(fmt.Sprintf("[%s]", info.Source)))
			if info.Description != "" {
				fmt.Fprintf(out, "  %s\n", ui.Dim(info.Description))
			}
		}
	}

	// Show local templates if present
	if len(localTemplates) > 0 {
		fmt.Fprintln(out)
		fmt.Fprintln(out, ui.H2("Local Gate Templates"))
		fmt.Fprintf(out, "%s %s\n", ui.Label("Directory"), ui.Dim(gatesDir))
		for _, tmpl := range localTemplates {
			fmt.Fprintf(out, "%s %s\n", ui.Dim("•"), ui.Value(tmpl, ui.GateColor))
		}
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
		return errors.Wrap(err, "context cancelled").WithOp("runGatesValidate")
	}

	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting working directory", err)
	}

	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals root").WithOp("runGatesValidate")
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
		return errors.IO("walking directory", err).WithField("path", festivalsRoot)
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
		fmt.Fprintln(out, ui.Success("✓ Gate configuration is valid."))
		return nil
	}

	fmt.Fprintln(out, ui.H1("Gate Validation"))
	fmt.Fprintf(out, "%s %s\n", ui.Label("Issues"), ui.Value(fmt.Sprintf("%d", len(issues))))
	for _, issue := range issues {
		severity := strings.ToUpper(issue.Severity)
		severityLabel := ui.Warning(severity)
		if strings.EqualFold(issue.Severity, "error") {
			severityLabel = ui.Error(severity)
		}
		fmt.Fprintf(out, "\n%s %s\n", severityLabel, ui.Dim(issue.Path))
		fmt.Fprintf(out, "  %s\n", issue.Message)
	}

	return nil
}

// Note: helper types and functions moved to gates_helpers.go
