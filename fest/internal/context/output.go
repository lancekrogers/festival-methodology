package context

import (
	"encoding/json"
	"fmt"
	"strings"
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

	// Location header
	sb.WriteString("=== CONTEXT ===\n\n")
	sb.WriteString(fmt.Sprintf("Location: %s (level: %s)\n", ctx.Location.FestivalName, ctx.Location.Level))
	if ctx.Location.PhaseName != "" {
		sb.WriteString(fmt.Sprintf("Phase: %s\n", ctx.Location.PhaseName))
	}
	if ctx.Location.SequenceName != "" {
		sb.WriteString(fmt.Sprintf("Sequence: %s\n", ctx.Location.SequenceName))
	}
	if ctx.Location.TaskName != "" {
		sb.WriteString(fmt.Sprintf("Task: %s\n", ctx.Location.TaskName))
	}
	sb.WriteString(fmt.Sprintf("Depth: %s\n\n", ctx.Depth))

	// Festival context
	if ctx.Festival != nil && ctx.Festival.Goal != nil {
		sb.WriteString("--- FESTIVAL ---\n")
		if ctx.Festival.Goal.Title != "" {
			sb.WriteString(fmt.Sprintf("Title: %s\n", ctx.Festival.Goal.Title))
		}
		if ctx.Festival.Goal.Objective != "" {
			sb.WriteString(fmt.Sprintf("Objective: %s\n", truncate(ctx.Festival.Goal.Objective, 200)))
		}
		sb.WriteString(fmt.Sprintf("Phases: %d\n", ctx.Festival.PhaseCount))
		sb.WriteString("\n")
	}

	// Phase context
	if ctx.Phase != nil {
		sb.WriteString("--- PHASE ---\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", ctx.Phase.Name))
		if ctx.Phase.PhaseType != "" {
			sb.WriteString(fmt.Sprintf("Type: %s\n", ctx.Phase.PhaseType))
		}
		if ctx.Phase.Goal != nil && ctx.Phase.Goal.Objective != "" {
			sb.WriteString(fmt.Sprintf("Objective: %s\n", truncate(ctx.Phase.Goal.Objective, 200)))
		}
		sb.WriteString(fmt.Sprintf("Sequences: %d\n", ctx.Phase.SequenceCount))
		sb.WriteString("\n")
	}

	// Sequence context
	if ctx.Sequence != nil {
		sb.WriteString("--- SEQUENCE ---\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", ctx.Sequence.Name))
		if ctx.Sequence.Goal != nil && ctx.Sequence.Goal.Objective != "" {
			sb.WriteString(fmt.Sprintf("Objective: %s\n", truncate(ctx.Sequence.Goal.Objective, 200)))
		}
		sb.WriteString(fmt.Sprintf("Tasks: %d\n", ctx.Sequence.TaskCount))
		sb.WriteString("\n")
	}

	// Task context
	if ctx.Task != nil {
		sb.WriteString("--- TASK ---\n")
		sb.WriteString(fmt.Sprintf("Name: %s\n", ctx.Task.Name))
		sb.WriteString(fmt.Sprintf("Number: %d\n", ctx.Task.TaskNumber))
		if ctx.Task.AutonomyLevel != "" {
			sb.WriteString(fmt.Sprintf("Autonomy: %s\n", ctx.Task.AutonomyLevel))
		}
		sb.WriteString(fmt.Sprintf("Parallel: %v\n", ctx.Task.ParallelAllowed))
		if ctx.Task.Objective != "" {
			sb.WriteString(fmt.Sprintf("Objective: %s\n", truncate(ctx.Task.Objective, 300)))
		}
		if len(ctx.Task.Dependencies) > 0 {
			sb.WriteString(fmt.Sprintf("Dependencies: %s\n", strings.Join(ctx.Task.Dependencies, ", ")))
		}
		if len(ctx.Task.Deliverables) > 0 {
			sb.WriteString("Deliverables:\n")
			for _, d := range ctx.Task.Deliverables {
				sb.WriteString(fmt.Sprintf("  - %s\n", d))
			}
		}
		sb.WriteString("\n")
	}

	// Rules (standard and full depth)
	if len(ctx.Rules) > 0 {
		sb.WriteString("--- RULES ---\n")
		for _, rule := range ctx.Rules {
			sb.WriteString(fmt.Sprintf("[%s] %s\n", rule.Category, rule.Title))
			if f.verbose && rule.Description != "" {
				sb.WriteString(fmt.Sprintf("  %s\n", truncate(rule.Description, 150)))
			}
		}
		sb.WriteString("\n")
	}

	// Decisions
	if len(ctx.Decisions) > 0 {
		sb.WriteString("--- DECISIONS ---\n")
		for _, d := range ctx.Decisions {
			date := d.Date.Format("2006-01-02")
			sb.WriteString(fmt.Sprintf("[%s] %s\n", date, d.Summary))
			if f.verbose && d.Rationale != "" {
				sb.WriteString(fmt.Sprintf("  Rationale: %s\n", truncate(d.Rationale, 100)))
			}
		}
		sb.WriteString("\n")
	}

	// Dependency outputs (full depth)
	if len(ctx.DependencyOutputs) > 0 {
		sb.WriteString("--- DEPENDENCY OUTPUTS ---\n")
		for _, dep := range ctx.DependencyOutputs {
			sb.WriteString(fmt.Sprintf("%s:\n", dep.TaskName))
			for _, output := range dep.Outputs {
				sb.WriteString(fmt.Sprintf("  - %s\n", output))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
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
