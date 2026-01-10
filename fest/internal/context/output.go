package context

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

// Formatter formats context output
type Formatter struct {
	verbose bool
}

// NewFormatter creates a new formatter
func NewFormatter(verbose bool) *Formatter {
	return &Formatter{verbose: verbose}
}

// FormatText formats context as structured text
func (f *Formatter) FormatText(ctx *ContextOutput) string {
	var sb strings.Builder

	sb.WriteString(ui.H1("Context"))

	sections := []string{
		formatLocationSection(ctx),
		formatFestivalSection(ctx),
		formatPhaseSection(ctx),
		formatSequenceSection(ctx),
		formatTaskSection(ctx),
		formatRulesSection(ctx.Rules, f.verbose),
		formatDecisionsSection(ctx.Decisions, f.verbose),
		formatDependencyOutputsSection(ctx.DependencyOutputs),
	}

	body := joinSections(sections...)
	if body != "" {
		sb.WriteString("\n\n")
		sb.WriteString(body)
	}
	sb.WriteString("\n")

	return sb.String()
}

func formatLocationSection(ctx *ContextOutput) string {
	if ctx == nil || ctx.Location == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Location"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Festival"), ui.Value(ctx.Location.FestivalName, ui.FestivalColor)))
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Level"), ui.Value(ctx.Location.Level)))
	if ctx.Location.PhaseName != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Phase"), ui.Value(ctx.Location.PhaseName, ui.PhaseColor)))
	}
	if ctx.Location.SequenceName != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Sequence"), ui.Value(ctx.Location.SequenceName, ui.SequenceColor)))
	}
	if ctx.Location.TaskName != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Task"), ui.Value(ctx.Location.TaskName, ui.TaskColor)))
	}
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Depth"), ui.Value(string(ctx.Depth))))

	return strings.TrimRight(sb.String(), "\n")
}

func formatFestivalSection(ctx *ContextOutput) string {
	if ctx == nil || ctx.Festival == nil || ctx.Festival.Goal == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Festival"))
	sb.WriteString("\n")
	if ctx.Festival.Goal.Title != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Title"), ui.Value(ctx.Festival.Goal.Title, ui.FestivalColor)))
	}
	if ctx.Festival.Goal.Objective != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Objective"), ui.Info(truncate(ctx.Festival.Goal.Objective, 200))))
	}
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Phases"), ui.Value(fmt.Sprintf("%d", ctx.Festival.PhaseCount))))

	return strings.TrimRight(sb.String(), "\n")
}

func formatPhaseSection(ctx *ContextOutput) string {
	if ctx == nil || ctx.Phase == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Phase"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Name"), ui.Value(ctx.Phase.Name, ui.PhaseColor)))
	if ctx.Phase.PhaseType != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Type"), ui.Value(ctx.Phase.PhaseType)))
	}
	if ctx.Phase.Goal != nil && ctx.Phase.Goal.Objective != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Objective"), ui.Info(truncate(ctx.Phase.Goal.Objective, 200))))
	}
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Sequences"), ui.Value(fmt.Sprintf("%d", ctx.Phase.SequenceCount))))

	return strings.TrimRight(sb.String(), "\n")
}

func formatSequenceSection(ctx *ContextOutput) string {
	if ctx == nil || ctx.Sequence == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Sequence"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Name"), ui.Value(ctx.Sequence.Name, ui.SequenceColor)))
	if ctx.Sequence.Goal != nil && ctx.Sequence.Goal.Objective != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Objective"), ui.Info(truncate(ctx.Sequence.Goal.Objective, 200))))
	}
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Tasks"), ui.Value(fmt.Sprintf("%d", ctx.Sequence.TaskCount))))

	return strings.TrimRight(sb.String(), "\n")
}

func formatTaskSection(ctx *ContextOutput) string {
	if ctx == nil || ctx.Task == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Task"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Name"), ui.Value(ctx.Task.Name, ui.TaskColor)))
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Number"), ui.Value(fmt.Sprintf("%d", ctx.Task.TaskNumber))))
	if ctx.Task.AutonomyLevel != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Autonomy"), ui.Value(ctx.Task.AutonomyLevel)))
	}
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Parallel"), ui.Value(fmt.Sprintf("%v", ctx.Task.ParallelAllowed))))
	if ctx.Task.Objective != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Objective"), ui.Info(truncate(ctx.Task.Objective, 300))))
	}
	if len(ctx.Task.Dependencies) > 0 {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Dependencies"), ui.Value(strings.Join(ctx.Task.Dependencies, ", "), ui.TaskColor)))
	}
	if len(ctx.Task.Deliverables) > 0 {
		sb.WriteString(fmt.Sprintf("%s\n", ui.Label("Deliverables")))
		for _, d := range ctx.Task.Deliverables {
			sb.WriteString(fmt.Sprintf("  - %s\n", ui.Value(d)))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatRulesSection(rules []Rule, verbose bool) string {
	if len(rules) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Rules"))
	sb.WriteString("\n")
	for _, rule := range rules {
		sb.WriteString(fmt.Sprintf("  %s %s\n", ui.Label(rule.Category), ui.Value(rule.Title)))
		if verbose && rule.Description != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", ui.Dim(truncate(rule.Description, 150))))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatDecisionsSection(decisions []Decision, verbose bool) string {
	if len(decisions) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Decisions"))
	sb.WriteString("\n")
	for _, d := range decisions {
		date := d.Date.Format("2006-01-02")
		sb.WriteString(fmt.Sprintf("  %s %s\n", ui.Label(date), ui.Value(d.Summary)))
		if verbose && d.Rationale != "" {
			sb.WriteString(fmt.Sprintf("    %s %s\n", ui.Label("Rationale"), ui.Dim(truncate(d.Rationale, 100))))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func formatDependencyOutputsSection(outputs []DepOutput) string {
	if len(outputs) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(ui.H2("Dependency Outputs"))
	sb.WriteString("\n")
	for idx, dep := range outputs {
		if idx > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Task"), ui.Value(dep.TaskName, ui.TaskColor)))
		for _, output := range dep.Outputs {
			sb.WriteString(fmt.Sprintf("  - %s\n", ui.Value(output)))
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

func joinSections(sections ...string) string {
	nonEmpty := make([]string, 0, len(sections))
	for _, section := range sections {
		if strings.TrimSpace(section) != "" {
			nonEmpty = append(nonEmpty, section)
		}
	}

	return strings.Join(nonEmpty, "\n\n")
}

// FormatJSON formats context as JSON
func (f *Formatter) FormatJSON(ctx *ContextOutput) (string, error) {
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatVerbose formats context with explanatory text for generic agents
func (f *Formatter) FormatVerbose(ctx *ContextOutput) string {
	var sb strings.Builder

	sb.WriteString("# AGENT CONTEXT BRIEFING\n\n")
	sb.WriteString("This document provides the context you need to work on your current task.\n\n")

	// Current location explanation
	sb.WriteString("## Your Current Location\n\n")
	sb.WriteString(fmt.Sprintf("You are working on the **%s** festival.\n", ctx.Location.FestivalName))
	if ctx.Location.PhaseName != "" {
		sb.WriteString(fmt.Sprintf("- Current phase: **%s**\n", ctx.Location.PhaseName))
	}
	if ctx.Location.SequenceName != "" {
		sb.WriteString(fmt.Sprintf("- Current sequence: **%s**\n", ctx.Location.SequenceName))
	}
	if ctx.Location.TaskName != "" {
		sb.WriteString(fmt.Sprintf("- Current task: **%s**\n", ctx.Location.TaskName))
	}
	sb.WriteString("\n")

	// Goals hierarchy
	sb.WriteString("## Goals Hierarchy\n\n")
	sb.WriteString("Understanding the goals at each level helps you align your work with the broader objectives.\n\n")

	if ctx.Festival != nil && ctx.Festival.Goal != nil {
		sb.WriteString("### Festival Goal\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", ctx.Festival.Goal.Objective))
	}

	if ctx.Phase != nil && ctx.Phase.Goal != nil {
		sb.WriteString("### Phase Goal\n")
		sb.WriteString(fmt.Sprintf("Phase type: %s\n", ctx.Phase.PhaseType))
		sb.WriteString(fmt.Sprintf("%s\n\n", ctx.Phase.Goal.Objective))
	}

	if ctx.Sequence != nil && ctx.Sequence.Goal != nil {
		sb.WriteString("### Sequence Goal\n")
		sb.WriteString(fmt.Sprintf("%s\n\n", ctx.Sequence.Goal.Objective))
	}

	// Task details
	if ctx.Task != nil {
		sb.WriteString("## Your Task\n\n")
		sb.WriteString(fmt.Sprintf("**Task**: %s (Task #%d)\n\n", ctx.Task.Name, ctx.Task.TaskNumber))

		if ctx.Task.Objective != "" {
			sb.WriteString("### Objective\n")
			sb.WriteString(fmt.Sprintf("%s\n\n", ctx.Task.Objective))
		}

		sb.WriteString("### Task Properties\n")
		sb.WriteString(fmt.Sprintf("- **Autonomy level**: %s\n", ctx.Task.AutonomyLevel))
		if ctx.Task.AutonomyLevel == "high" {
			sb.WriteString("  (You can proceed with minimal human intervention)\n")
		} else if ctx.Task.AutonomyLevel == "low" {
			sb.WriteString("  (Seek human approval before major decisions)\n")
		}
		sb.WriteString(fmt.Sprintf("- **Parallel execution**: %v\n", ctx.Task.ParallelAllowed))
		if ctx.Task.ParallelAllowed {
			sb.WriteString("  (This task can run alongside other tasks)\n")
		}
		sb.WriteString("\n")

		if len(ctx.Task.Dependencies) > 0 {
			sb.WriteString("### Dependencies\n")
			sb.WriteString("Complete these tasks first:\n")
			for _, dep := range ctx.Task.Dependencies {
				sb.WriteString(fmt.Sprintf("- %s\n", dep))
			}
			sb.WriteString("\n")
		}

		if len(ctx.Task.Deliverables) > 0 {
			sb.WriteString("### Expected Deliverables\n")
			sb.WriteString("You should produce:\n")
			for _, d := range ctx.Task.Deliverables {
				sb.WriteString(fmt.Sprintf("- %s\n", d))
			}
			sb.WriteString("\n")
		}
	}

	// Rules
	if len(ctx.Rules) > 0 {
		sb.WriteString("## Applicable Rules\n\n")
		sb.WriteString("These rules apply to your work:\n\n")
		for _, rule := range ctx.Rules {
			sb.WriteString(fmt.Sprintf("**[%s] %s**\n", rule.Category, rule.Title))
			if rule.Description != "" {
				sb.WriteString(fmt.Sprintf("%s\n\n", rule.Description))
			}
		}
	}

	// Decisions
	if len(ctx.Decisions) > 0 {
		sb.WriteString("## Recent Decisions\n\n")
		sb.WriteString("These decisions have been made that may affect your work:\n\n")
		for _, d := range ctx.Decisions {
			date := d.Date.Format("Jan 2, 2006")
			sb.WriteString(fmt.Sprintf("**%s** - %s\n", date, d.Summary))
			if d.Rationale != "" {
				sb.WriteString(fmt.Sprintf("_Rationale: %s_\n", d.Rationale))
			}
			sb.WriteString("\n")
		}
	}

	// Dependency outputs
	if len(ctx.DependencyOutputs) > 0 {
		sb.WriteString("## What Previous Tasks Produced\n\n")
		sb.WriteString("These are outputs from tasks you depend on:\n\n")
		for _, dep := range ctx.DependencyOutputs {
			sb.WriteString(fmt.Sprintf("### %s\n", dep.TaskName))
			for _, output := range dep.Outputs {
				sb.WriteString(fmt.Sprintf("- %s\n", output))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// truncate truncates a string to the given length
func truncate(s string, length int) string {
	// Clean up the string first
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")

	if len(s) <= length {
		return s
	}
	return s[:length-3] + "..."
}
