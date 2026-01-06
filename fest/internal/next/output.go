package next

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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
	var sb strings.Builder
	sb.WriteString(ui.H2("Festival Complete"))
	sb.WriteString("\n")
	sb.WriteString(ui.Success("All tasks have been completed."))
	if result.Reason != "" {
		sb.WriteString("\n")
		writeLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}
	return sb.String()
}

func formatTextBlockingGate(result *NextTaskResult) string {
	var sb strings.Builder
	gate := result.BlockingGate

	sb.WriteString(ui.H2("Quality Gate Required"))
	sb.WriteString("\n")
	writeLabelValue(&sb, "Phase", ui.Value(gate.Phase, ui.PhaseColor))
	writeLabelValue(&sb, "Type", ui.Value(gate.GateType))
	writeLabelValue(&sb, "Description", ui.Value(gate.Description))

	if len(gate.Criteria) > 0 {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Criteria"))
		sb.WriteString("\n")
		for _, c := range gate.Criteria {
			sb.WriteString(fmt.Sprintf("  - %s\n", ui.Info(c)))
		}
	}

	return sb.String()
}

func formatTextNoTask(result *NextTaskResult) string {
	var sb strings.Builder

	sb.WriteString(ui.H2("No Tasks Available"))
	sb.WriteString("\n")
	if result.Reason != "" {
		writeLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}
	if result.Location != nil {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Location"))
		sb.WriteString("\n")
		writeLabelValue(&sb, "Festival", ui.Dim(result.Location.FestivalPath))
		if result.Location.PhasePath != "" {
			writeLabelValue(&sb, "Phase", ui.Dim(filepath.Base(result.Location.PhasePath)))
		}
		if result.Location.SequencePath != "" {
			writeLabelValue(&sb, "Sequence", ui.Dim(filepath.Base(result.Location.SequencePath)))
		}
	}

	return sb.String()
}

func formatTextTask(result *NextTaskResult) string {
	var sb strings.Builder

	sb.WriteString(ui.H1("Next Task"))
	sb.WriteString("\n")
	writeLabelValue(&sb, "Task", ui.Value(result.Task.Name, ui.TaskColor))
	writeLabelValue(&sb, "Path", ui.Dim(result.Task.Path))
	writeLabelValue(&sb, "Sequence", ui.Value(result.Task.SequenceName, ui.SequenceColor))
	writeLabelValue(&sb, "Phase", ui.Value(result.Task.PhaseName, ui.PhaseColor))

	if result.Task.AutonomyLevel != "" {
		writeLabelValue(&sb, "Autonomy", ui.Value(result.Task.AutonomyLevel))
	}

	sb.WriteString("\n")
	writeLabelValue(&sb, "Recommendation", ui.Info(result.Reason))

	if len(result.ParallelTasks) > 0 {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Parallel Tasks"))
		sb.WriteString("\n")
		for _, task := range result.ParallelTasks {
			sb.WriteString(fmt.Sprintf("  - %s %s\n", ui.Value(task.Name, ui.TaskColor), ui.Dim(task.SequenceName)))
		}
	}

	return sb.String()
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
		writeLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}

	return sb.String()
}

func formatVerboseBlockingGate(result *NextTaskResult) string {
	var sb strings.Builder
	gate := result.BlockingGate

	sb.WriteString(ui.H2("Quality Gate Required"))
	sb.WriteString("\n")
	writeLabelValue(&sb, "Phase", ui.Value(gate.Phase, ui.PhaseColor))
	writeLabelValue(&sb, "Gate Type", ui.Value(gate.GateType))
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
		writeLabelValue(&sb, "Reason", ui.Info(result.Reason))
	}

	if result.Location != nil {
		sb.WriteString("\n")
		sb.WriteString(ui.H3("Location"))
		sb.WriteString("\n")
		writeLabelValue(&sb, "Festival", ui.Dim(result.Location.FestivalPath))
		if result.Location.PhasePath != "" {
			writeLabelValue(&sb, "Phase", ui.Dim(filepath.Base(result.Location.PhasePath)))
		}
		if result.Location.SequencePath != "" {
			writeLabelValue(&sb, "Sequence", ui.Dim(filepath.Base(result.Location.SequencePath)))
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

func writeLabelValue(sb *strings.Builder, label, value string) {
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label(label), value))
}

func writeTaskDetails(sb *strings.Builder, task *TaskInfo) {
	sb.WriteString(ui.H2("Task Details"))
	sb.WriteString("\n")
	writeLabelValue(sb, "Task", ui.Value(task.Name, ui.TaskColor))
	writeLabelValue(sb, "Number", ui.Value(fmt.Sprintf("%d", task.Number)))
	writeLabelValue(sb, "Path", ui.Dim(task.Path))
	sb.WriteString("\n")
}

func writeTaskLocation(sb *strings.Builder, task *TaskInfo) {
	sb.WriteString(ui.H2("Location"))
	sb.WriteString("\n")
	writeLabelValue(sb, "Phase", ui.Value(task.PhaseName, ui.PhaseColor))
	writeLabelValue(sb, "Sequence", ui.Value(task.SequenceName, ui.SequenceColor))
	sb.WriteString("\n")
}

func writeTaskProperties(sb *strings.Builder, task *TaskInfo) {
	if task.AutonomyLevel == "" && task.ParallelGroup == 0 {
		return
	}

	sb.WriteString(ui.H2("Properties"))
	sb.WriteString("\n")
	if task.AutonomyLevel != "" {
		writeLabelValue(sb, "Autonomy", ui.Value(task.AutonomyLevel))
	}
	if task.ParallelGroup > 0 {
		writeLabelValue(sb, "Parallel Group", ui.Value(fmt.Sprintf("%d", task.ParallelGroup)))
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
	writeLabelValue(sb, "Festival", ui.Dim(filepath.Base(loc.FestivalPath)))
	if loc.PhasePath != "" {
		writeLabelValue(sb, "Phase", ui.Dim(filepath.Base(loc.PhasePath)))
	}
	if loc.SequencePath != "" {
		writeLabelValue(sb, "Sequence", ui.Dim(filepath.Base(loc.SequencePath)))
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
