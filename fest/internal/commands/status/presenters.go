package status

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// emitErrorJSON outputs an error message in JSON format.
func emitErrorJSON(message string) error {
	result := map[string]interface{}{
		"error": message,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

// emitLocationJSON outputs location information in JSON format.
func emitLocationJSON(loc *show.LocationInfo) error {
	result := map[string]interface{}{
		"type": loc.Type,
	}

	if loc.Festival != nil {
		result["festival"] = map[string]interface{}{
			"name":   loc.Festival.Name,
			"status": loc.Festival.Status,
			"path":   loc.Festival.Path,
		}
		if loc.Festival.Stats != nil {
			result["progress"] = loc.Festival.Stats.Progress
		}
	}

	if loc.Phase != "" {
		result["phase"] = loc.Phase
	}
	if loc.Sequence != "" {
		result["sequence"] = loc.Sequence
	}
	if loc.Task != "" {
		result["task"] = loc.Task
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling status to JSON")
	}
	fmt.Println(string(data))
	return nil
}

// emitLocationText outputs location information in human-readable text format.
func emitLocationText(loc *show.LocationInfo) error {
	if loc.Festival == nil {
		fmt.Println("Not in a festival directory")
		return nil
	}

	fmt.Printf("Festival: %s\n", loc.Festival.Name)
	fmt.Printf("Status:   %s\n", loc.Festival.Status)
	fmt.Printf("Location: %s\n", loc.Type)

	if loc.Phase != "" {
		fmt.Printf("Phase:    %s\n", loc.Phase)
	}
	if loc.Sequence != "" {
		fmt.Printf("Sequence: %s\n", loc.Sequence)
	}
	if loc.Task != "" {
		fmt.Printf("Task:     %s\n", loc.Task)
	}

	if loc.Festival.Stats != nil {
		fmt.Printf("\nProgress: %.1f%%\n", loc.Festival.Stats.Progress)
		fmt.Printf("  Phases:    %d/%d completed\n",
			loc.Festival.Stats.Phases.Completed,
			loc.Festival.Stats.Phases.Total)
		fmt.Printf("  Sequences: %d/%d completed\n",
			loc.Festival.Stats.Sequences.Completed,
			loc.Festival.Stats.Sequences.Total)
		fmt.Printf("  Tasks:     %d/%d completed\n",
			loc.Festival.Stats.Tasks.Completed,
			loc.Festival.Stats.Tasks.Total)
	}

	return nil
}

// emitPhasesJSON outputs phase information in JSON format.
func emitPhasesJSON(phases []*PhaseInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":   "phase",
		"count":  len(phases),
		"phases": phases,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling phases to JSON")
	}
	fmt.Println(string(data))
	return nil
}

// emitPhasesText outputs phase information in human-readable text format.
func emitPhasesText(phases []*PhaseInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Phases with status '%s' (%d)\n", filterStatus, len(phases))
	} else {
		fmt.Printf("Phases (%d)\n", len(phases))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, phase := range phases {
		fmt.Printf("  %s [%s]", phase.Name, phase.Status)
		if phase.TaskStats.Total > 0 {
			fmt.Printf(" (%d/%d tasks)", phase.TaskStats.Completed, phase.TaskStats.Total)
		}
		fmt.Println()
	}

	return nil
}

// emitSequencesJSON outputs sequence information in JSON format.
func emitSequencesJSON(sequences []*SequenceInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":      "sequence",
		"count":     len(sequences),
		"sequences": sequences,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling sequences to JSON")
	}
	fmt.Println(string(data))
	return nil
}

// emitSequencesText outputs sequence information in human-readable text format.
func emitSequencesText(sequences []*SequenceInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Sequences with status '%s' (%d)\n", filterStatus, len(sequences))
	} else {
		fmt.Printf("Sequences (%d)\n", len(sequences))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, seq := range sequences {
		fmt.Printf("  %s/%s [%s]", seq.PhaseName, seq.Name, seq.Status)
		if seq.TaskStats.Total > 0 {
			fmt.Printf(" (%d/%d tasks)", seq.TaskStats.Completed, seq.TaskStats.Total)
		}
		fmt.Println()
	}

	return nil
}

// emitTasksJSON outputs task information in JSON format.
func emitTasksJSON(tasks []*TaskInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":  "task",
		"count": len(tasks),
		"tasks": tasks,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling tasks to JSON")
	}
	fmt.Println(string(data))
	return nil
}

// emitTasksText outputs task information in human-readable text format.
func emitTasksText(tasks []*TaskInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Tasks with status '%s' (%d)\n", filterStatus, len(tasks))
	} else {
		fmt.Printf("Tasks (%d)\n", len(tasks))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, task := range tasks {
		fmt.Printf("  %s/%s/%s [%s]\n",
			task.PhaseName,
			task.SequenceName,
			task.Name,
			task.Status)
	}

	return nil
}

// emitEmptyJSON outputs a message for empty results in JSON format.
func emitEmptyJSON(entityType, filterStatus string) error {
	result := map[string]interface{}{
		"type":  entityType,
		"count": 0,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
		result["message"] = fmt.Sprintf("no %ss found with status '%s'", entityType, filterStatus)
	} else {
		result["message"] = fmt.Sprintf("no %ss found", entityType)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}
