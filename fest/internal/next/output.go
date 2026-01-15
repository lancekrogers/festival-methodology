package next

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/templates/agent"
)

// FormatText formats the result as human-readable text
func FormatText(result *NextTaskResult) string {
	switch {
	case result.FestivalComplete:
		return formatTextComplete(result)
	case result.BlockingGate != nil:
		return formatTextBlockingGate(result)
	case result.Task == nil:
		return formatTextNoTask(result)
	default:
		return formatTextTask(result)
	}
}

// FormatJSON formats the result as JSON
func FormatJSON(result *NextTaskResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatVerbose formats the result with additional details
func FormatVerbose(result *NextTaskResult) string {
	switch {
	case result.FestivalComplete:
		return formatVerboseComplete(result)
	case result.BlockingGate != nil:
		return formatVerboseBlockingGate(result)
	case result.Task == nil:
		return formatVerboseNoTask(result)
	default:
		return formatVerboseTask(result)
	}
}

func formatTextComplete(result *NextTaskResult) string {
	var reasonLine string
	if result.Reason != "" {
		reasonLine = labelValue("Reason", ui.Info(result.Reason))
	}

	data := struct {
		Header     string
		Message    string
		ReasonLine string
	}{
		Header:     ui.H2("Festival Complete"),
		Message:    ui.Success("All tasks have been completed."),
		ReasonLine: reasonLine,
	}

	var buf bytes.Buffer
	agent.MustGet("next/complete").Execute(&buf, data)
	return buf.String()
}

func formatTextBlockingGate(result *NextTaskResult) string {
	gate := result.BlockingGate

	var criteriaSection string
	if len(gate.Criteria) > 0 {
		var sb strings.Builder
		sb.WriteString(ui.H3("Criteria"))
		sb.WriteString("\n")
		for _, c := range gate.Criteria {
			sb.WriteString(fmt.Sprintf("  - %s\n", ui.Info(c)))
		}
		criteriaSection = sb.String()
	}

	data := struct {
		Header          string
		PhaseLine       string
		TypeLine        string
		DescriptionLine string
		CriteriaSection string
	}{
		Header:          ui.H2("Quality Gate Required"),
		PhaseLine:       labelValue("Phase", ui.Value(gate.Phase, ui.PhaseColor)),
		TypeLine:        labelValue("Type", ui.Value(gate.GateType)),
		DescriptionLine: labelValue("Description", ui.Value(gate.Description)),
		CriteriaSection: criteriaSection,
	}

	var buf bytes.Buffer
	agent.MustGet("next/blocked").Execute(&buf, data)
	return buf.String()
}

func formatTextNoTask(result *NextTaskResult) string {
	var reasonLine string
	if result.Reason != "" {
		reasonLine = labelValue("Reason", ui.Info(result.Reason))
	}

	var locationSection string
	if result.Location != nil {
		var sb strings.Builder
		sb.WriteString(ui.H3("Location"))
		sb.WriteString("\n")
		ui.WriteLabelValue(&sb, "Festival", ui.Dim(result.Location.FestivalPath))
		if result.Location.PhasePath != "" {
			ui.WriteLabelValue(&sb, "Phase", ui.Dim(filepath.Base(result.Location.PhasePath)))
		}
		if result.Location.SequencePath != "" {
			ui.WriteLabelValue(&sb, "Sequence", ui.Dim(filepath.Base(result.Location.SequencePath)))
		}
		locationSection = sb.String()
	}

	data := struct {
		Header          string
		ReasonLine      string
		LocationSection string
	}{
		Header:          ui.H2("No Tasks Available"),
		ReasonLine:      reasonLine,
		LocationSection: locationSection,
	}

	var buf bytes.Buffer
	agent.MustGet("next/no_task").Execute(&buf, data)
	return buf.String()
}

func formatTextTask(result *NextTaskResult) string {
	// Build parallel tasks section if present
	var parallelSection string
	if len(result.ParallelTasks) > 0 {
		var sb strings.Builder
		sb.WriteString(ui.H3("Parallel Tasks"))
		sb.WriteString("\n")
		for _, task := range result.ParallelTasks {
			sb.WriteString(fmt.Sprintf("  - %s %s\n", ui.Value(task.Name, ui.TaskColor), ui.Dim(task.SequenceName)))
		}
		parallelSection = sb.String()
	}

	// Build autonomy line if present
	var autonomyLine string
	if result.Task.AutonomyLevel != "" {
		var sb strings.Builder
		ui.WriteLabelValue(&sb, "Autonomy", ui.Value(result.Task.AutonomyLevel))
		autonomyLine = strings.TrimSuffix(sb.String(), "\n")
	}

	// Build context files section
	contextSection := buildContextSection(result.Location, result.Task)

	// Build progress line if available
	var progressLine string
	if result.Progress != nil {
		progressLine = labelValue("Progress", ui.Info(fmt.Sprintf("%.1f%% (%d/%d tasks)",
			result.Progress.Percentage,
			result.Progress.CompletedTasks,
			result.Progress.TotalTasks)))
	}

	// Build label lines
	taskRelPath := filepath.Join(result.Task.PhaseName, result.Task.SequenceName, result.Task.Name+".md")

	data := struct {
		Header             string
		TaskLine           string
		PathLine           string
		SequenceLine       string
		PhaseLine          string
		AutonomyLine       string
		ProgressLine       string
		RecommendationLine string
		ParallelSection    string
		ActionInstruction  string
		ProgressCmd        string
		ContextSection     string
	}{
		Header:             ui.H1("Next Task"),
		TaskLine:           labelValue("Task", ui.Value(result.Task.Name, ui.TaskColor)),
		PathLine:           labelValue("Path", ui.Dim(result.Task.Path)),
		SequenceLine:       labelValue("Sequence", ui.Value(result.Task.SequenceName, ui.SequenceColor)),
		PhaseLine:          labelValue("Phase", ui.Value(result.Task.PhaseName, ui.PhaseColor)),
		AutonomyLine:       autonomyLine,
		ProgressLine:       progressLine,
		RecommendationLine: labelValue("Recommendation", ui.Info(result.Reason)),
		ParallelSection:    parallelSection,
		ActionInstruction:  ui.Info("Read this file and follow the instructions laid out exactly."),
		ProgressCmd:        ui.Value(fmt.Sprintf("fest progress --task %s --complete", taskRelPath)),
		ContextSection:     contextSection,
	}

	var buf bytes.Buffer
	agent.MustGet("next/task").Execute(&buf, data)
	return buf.String()
}

// buildContextSection creates the context files section showing goal files
func buildContextSection(loc *LocationInfo, task *TaskInfo) string {
	var sb strings.Builder
	sb.WriteString(ui.H3("Context Files"))
	sb.WriteString("\n")

	// Festival goal
	if loc != nil && loc.FestivalPath != "" {
		festivalGoal := filepath.Join(loc.FestivalPath, "FESTIVAL_GOAL.md")
		sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(festivalGoal)))
	}

	// Phase goal
	if task.PhasePath != "" {
		phaseGoal := filepath.Join(task.PhasePath, "PHASE_GOAL.md")
		sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(phaseGoal)))
	}

	// Sequence goal
	if task.SequencePath != "" {
		sequenceGoal := filepath.Join(task.SequencePath, "SEQUENCE_GOAL.md")
		sb.WriteString(fmt.Sprintf("  - %s\n", ui.Dim(sequenceGoal)))
	}

	return sb.String()
}

// labelValue formats a label-value pair without trailing newline
func labelValue(label, value string) string {
	var sb strings.Builder
	ui.WriteLabelValue(&sb, label, value)
	return strings.TrimSuffix(sb.String(), "\n")
}

func formatVerboseComplete(result *NextTaskResult) string {
	var sb strings.Builder

	sb.WriteString(ui.H2("Festival Complete"))
	sb.WriteString("\n")
	sb.WriteString(ui.Success("All tasks in the festival have been completed."))
	sb.WriteString("\n")
	sb.WriteString(ui.Info("Congratulations on finishing the festival!"))
	if result.Reason != "" {
		sb.WriteString("\n")
		ui.WriteLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}

	return sb.String()
}

func formatVerboseBlockingGate(result *NextTaskResult) string {
	var sb strings.Builder
	gate := result.BlockingGate

	sb.WriteString(ui.H2("Quality Gate Required"))
	sb.WriteString("\n")
	ui.WriteLabelValue(&sb, "Phase", ui.Value(gate.Phase, ui.PhaseColor))
	ui.WriteLabelValue(&sb, "Gate Type", ui.Value(gate.GateType))
	sb.WriteString("\n")
	sb.WriteString(ui.Info(gate.Description))
	sb.WriteString("\n")

	if len(gate.Criteria) > 0 {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Criteria to Pass"))
		sb.WriteString("\n")
		for i, c := range gate.Criteria {
			sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, ui.Info(c)))
		}
	}

	sb.WriteString("\n")
	sb.WriteString(ui.Warning("Complete the quality gate before proceeding."))

	return sb.String()
}

func formatVerboseNoTask(result *NextTaskResult) string {
	var sb strings.Builder

	sb.WriteString(ui.H2("No Tasks Available"))
	sb.WriteString("\n")
	if result.Reason != "" {
		ui.WriteLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}

	if result.Location != nil {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Location"))
		sb.WriteString("\n")
		ui.WriteLabelValue(&sb, "Festival", ui.Dim(result.Location.FestivalPath))
		if result.Location.PhasePath != "" {
			ui.WriteLabelValue(&sb, "Phase", ui.Dim(filepath.Base(result.Location.PhasePath)))
		}
		if result.Location.SequencePath != "" {
			ui.WriteLabelValue(&sb, "Sequence", ui.Dim(filepath.Base(result.Location.SequencePath)))
		}
	}

	return sb.String()
}

func formatVerboseTask(result *NextTaskResult) string {
	var sb strings.Builder

	sb.WriteString(ui.H1("Next Task"))
	sb.WriteString("\n")
	writeTaskDetails(&sb, result.Task)
	writeTaskLocation(&sb, result.Task)
	writeTaskProperties(&sb, result.Task)
	writeTaskDependencies(&sb, result.Task.Dependencies)
	writeRecommendation(&sb, result.Reason)
	writeParallelTasks(&sb, result.ParallelTasks)
	writeCurrentLocation(&sb, result.Location)

	return sb.String()
}

func writeTaskDetails(sb *strings.Builder, task *TaskInfo) {
	sb.WriteString(ui.H2("Task Details"))
	sb.WriteString("\n")
	ui.WriteLabelValue(sb, "Task", ui.Value(task.Name, ui.TaskColor))
	ui.WriteLabelValue(sb, "Number", ui.Value(fmt.Sprintf("%d", task.Number)))
	ui.WriteLabelValue(sb, "Path", ui.Dim(task.Path))
	sb.WriteString("\n")
}

func writeTaskLocation(sb *strings.Builder, task *TaskInfo) {
	sb.WriteString(ui.H2("Location"))
	sb.WriteString("\n")
	ui.WriteLabelValue(sb, "Phase", ui.Value(task.PhaseName, ui.PhaseColor))
	ui.WriteLabelValue(sb, "Sequence", ui.Value(task.SequenceName, ui.SequenceColor))
	sb.WriteString("\n")
}

func writeTaskProperties(sb *strings.Builder, task *TaskInfo) {
	if task.AutonomyLevel == "" && task.ParallelGroup == 0 {
		return
	}

	sb.WriteString(ui.H2("Properties"))
	sb.WriteString("\n")
	if task.AutonomyLevel != "" {
		ui.WriteLabelValue(sb, "Autonomy", ui.Value(task.AutonomyLevel))
	}
	if task.ParallelGroup > 0 {
		ui.WriteLabelValue(sb, "Parallel Group", ui.Value(fmt.Sprintf("%d", task.ParallelGroup)))
	}
	sb.WriteString("\n")
}

func writeTaskDependencies(sb *strings.Builder, deps []string) {
	if len(deps) == 0 {
		return
	}

	sb.WriteString(ui.H2("Dependencies"))
	sb.WriteString("\n")
	for _, dep := range deps {
		sb.WriteString(fmt.Sprintf("  %s %s\n", ui.StateIcon("completed"), ui.Info(dep)))
	}
	sb.WriteString("\n")
}

func writeRecommendation(sb *strings.Builder, reason string) {
	sb.WriteString(ui.H2("Recommendation"))
	sb.WriteString("\n")
	if reason == "" {
		sb.WriteString(ui.Dim("No recommendation available."))
	} else {
		sb.WriteString(ui.Info(reason))
	}
	sb.WriteString("\n\n")
}

func writeParallelTasks(sb *strings.Builder, tasks []*TaskInfo) {
	if len(tasks) == 0 {
		return
	}

	sb.WriteString(ui.H2("Parallel Tasks"))
	sb.WriteString("\n")
	for _, task := range tasks {
		sb.WriteString(fmt.Sprintf("  - %s\n", ui.Value(task.Name, ui.TaskColor)))
		sb.WriteString(fmt.Sprintf("    %s %s\n", ui.Label("Path"), ui.Dim(task.Path)))
		if task.AutonomyLevel != "" {
			sb.WriteString(fmt.Sprintf("    %s %s\n", ui.Label("Autonomy"), ui.Value(task.AutonomyLevel)))
		}
		sb.WriteString("\n")
	}
}

func writeCurrentLocation(sb *strings.Builder, loc *LocationInfo) {
	sb.WriteString(ui.H2("Current Location"))
	sb.WriteString("\n")
	if loc == nil {
		sb.WriteString(ui.Dim("Unknown location\n"))
		return
	}
	ui.WriteLabelValue(sb, "Festival", ui.Dim(filepath.Base(loc.FestivalPath)))
	if loc.PhasePath != "" {
		ui.WriteLabelValue(sb, "Phase", ui.Dim(filepath.Base(loc.PhasePath)))
	}
	if loc.SequencePath != "" {
		ui.WriteLabelValue(sb, "Sequence", ui.Dim(filepath.Base(loc.SequencePath)))
	}
}

// FormatShort formats a minimal one-line output
func FormatShort(result *NextTaskResult) string {
	if result.FestivalComplete {
		return "Festival complete"
	}
	if result.BlockingGate != nil {
		return fmt.Sprintf("Blocked: Quality gate in %s", result.BlockingGate.Phase)
	}
	if result.Task == nil {
		return "No tasks available"
	}
	return result.Task.Path
}

// FormatCD formats output suitable for shell cd command
func FormatCD(result *NextTaskResult) string {
	if result.Task == nil {
		return ""
	}
	// Return the directory containing the task file
	return filepath.Dir(result.Task.Path)
}
