package status

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

// emitErrorJSON outputs an error message in JSON format.
func emitErrorJSON(message string) error {
	result := map[string]interface{}{
		"error": message,
	}
	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
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

	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

// emitLocationText outputs location information in human-readable text format.
func emitLocationText(ctx context.Context, loc *show.LocationInfo) error {
	if loc.Festival == nil {
		fmt.Println("Not in a festival directory")
		return nil
	}

	fmt.Println(ui.H1("Status"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(loc.Festival.Name, ui.FestivalColor))
	fmt.Printf("%s %s\n", ui.Label("Status"), ui.GetStatusStyle(loc.Festival.Status).Render(loc.Festival.Status))
	fmt.Printf("%s %s\n", ui.Label("Location"), ui.Value(loc.Type))

	if loc.Phase != "" {
		fmt.Printf("%s %s\n", ui.Label("Phase"), ui.Value(loc.Phase, ui.PhaseColor))
	}
	if loc.Sequence != "" {
		fmt.Printf("%s %s\n", ui.Label("Sequence"), ui.Value(loc.Sequence, ui.SequenceColor))
	}
	if loc.Task != "" {
		fmt.Printf("%s %s\n", ui.Label("Task"), ui.Value(loc.Task, ui.TaskColor))
	}

	// Display context-appropriate progress
	switch loc.Type {
	case "sequence":
		return emitSequenceProgress(ctx, loc)
	case "phase":
		return emitPhaseProgress(ctx, loc)
	case "festival", "task":
		return emitFestivalProgress(loc)
	default:
		return emitFestivalProgress(loc)
	}
}

func emitFestivalProgress(loc *show.LocationInfo) error {
	if loc.Festival.Stats != nil {
		fmt.Printf("\n%s %s\n",
			ui.Label("Progress"),
			ui.Value(fmt.Sprintf("%.1f%%", loc.Festival.Stats.Progress)))
		fmt.Printf("%s %s\n",
			ui.Label("Phases"),
			ui.Dim(fmt.Sprintf("%d/%d completed", loc.Festival.Stats.Phases.Completed, loc.Festival.Stats.Phases.Total)))
		fmt.Printf("%s %s\n",
			ui.Label("Sequences"),
			ui.Dim(fmt.Sprintf("%d/%d completed", loc.Festival.Stats.Sequences.Completed, loc.Festival.Stats.Sequences.Total)))
		fmt.Printf("%s %s\n",
			ui.Label("Tasks"),
			ui.Dim(fmt.Sprintf("%d/%d completed", loc.Festival.Stats.Tasks.Completed, loc.Festival.Stats.Tasks.Total)))
	}
	return nil
}

func emitPhaseProgress(ctx context.Context, loc *show.LocationInfo) error {
	if ctx == nil {
		ctx = context.Background()
	}
	mgr, err := progress.NewManager(ctx, loc.Festival.Path)
	if err != nil {
		// Fall back to festival stats if progress manager fails
		return emitFestivalProgress(loc)
	}

	phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
	phaseProgress, err := mgr.GetPhaseProgress(ctx, phasePath)
	if err != nil {
		// Fall back to festival stats if phase progress fails
		return emitFestivalProgress(loc)
	}

	prog := phaseProgress.Progress
	fmt.Printf("\n%s %s %s\n",
		ui.Label("Progress"),
		ui.Value(fmt.Sprintf("%d%%", prog.Percentage)),
		ui.Dim(fmt.Sprintf("(%d/%d tasks)", prog.Completed, prog.Total)))

	if prog.InProgress > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("In progress"),
			ui.GetStateStyle("in_progress").Render(fmt.Sprintf("%d", prog.InProgress)))
	}
	if prog.Pending > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Pending"),
			ui.GetStateStyle("pending").Render(fmt.Sprintf("%d", prog.Pending)))
	}
	if prog.Blocked > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Blocked"),
			ui.GetStateStyle("blocked").Render(fmt.Sprintf("%d", prog.Blocked)))
	}

	return nil
}

func emitSequenceProgress(ctx context.Context, loc *show.LocationInfo) error {
	if ctx == nil {
		ctx = context.Background()
	}
	mgr, err := progress.NewManager(ctx, loc.Festival.Path)
	if err != nil {
		// Fall back to festival stats if progress manager fails
		return emitFestivalProgress(loc)
	}

	seqPath := filepath.Join(loc.Festival.Path, loc.Phase, loc.Sequence)
	seqProgress, err := mgr.GetSequenceProgress(ctx, seqPath)
	if err != nil {
		// Fall back to festival stats if sequence progress fails
		return emitFestivalProgress(loc)
	}

	prog := seqProgress.Progress
	fmt.Printf("\n%s %s %s\n",
		ui.Label("Progress"),
		ui.Value(fmt.Sprintf("%d%%", prog.Percentage)),
		ui.Dim(fmt.Sprintf("(%d/%d tasks)", prog.Completed, prog.Total)))

	if prog.InProgress > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("In progress"),
			ui.GetStateStyle("in_progress").Render(fmt.Sprintf("%d", prog.InProgress)))
	}
	if prog.Pending > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Pending"),
			ui.GetStateStyle("pending").Render(fmt.Sprintf("%d", prog.Pending)))
	}
	if prog.Blocked > 0 {
		fmt.Printf("%s %s\n",
			ui.Label("Blocked"),
			ui.GetStateStyle("blocked").Render(fmt.Sprintf("%d", prog.Blocked)))
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

	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

// emitPhasesText outputs phase information in human-readable text format.
func emitPhasesText(phases []*PhaseInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Println(ui.H2(fmt.Sprintf("Phases with status '%s' (%d)", filterStatus, len(phases))))
	} else {
		fmt.Println(ui.H2(fmt.Sprintf("Phases (%d)", len(phases))))
	}
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	for _, phase := range phases {
		fmt.Printf("%s %s [%s]",
			ui.StateIcon(phase.Status),
			ui.Value(phase.Name, ui.PhaseColor),
			renderStateLabel(phase.Status))
		if phase.TaskStats.Total > 0 {
			fmt.Printf(" %s", ui.Dim(fmt.Sprintf("(%d/%d tasks)", phase.TaskStats.Completed, phase.TaskStats.Total)))
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

	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

// emitSequencesText outputs sequence information in human-readable text format.
func emitSequencesText(sequences []*SequenceInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Println(ui.H2(fmt.Sprintf("Sequences with status '%s' (%d)", filterStatus, len(sequences))))
	} else {
		fmt.Println(ui.H2(fmt.Sprintf("Sequences (%d)", len(sequences))))
	}
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	for _, seq := range sequences {
		fmt.Printf("%s %s/%s [%s]",
			ui.StateIcon(seq.Status),
			ui.Value(seq.PhaseName, ui.PhaseColor),
			ui.Value(seq.Name, ui.SequenceColor),
			renderStateLabel(seq.Status))
		if seq.TaskStats.Total > 0 {
			fmt.Printf(" %s", ui.Dim(fmt.Sprintf("(%d/%d tasks)", seq.TaskStats.Completed, seq.TaskStats.Total)))
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

	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

// emitTasksText outputs task information in human-readable text format.
func emitTasksText(tasks []*TaskInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Println(ui.H2(fmt.Sprintf("Tasks with status '%s' (%d)", filterStatus, len(tasks))))
	} else {
		fmt.Println(ui.H2(fmt.Sprintf("Tasks (%d)", len(tasks))))
	}
	fmt.Println(ui.Dim(strings.Repeat("─", 60)))

	for _, task := range tasks {
		fmt.Printf("%s %s/%s/%s [%s]\n",
			ui.StateIcon(task.Status),
			ui.Value(task.PhaseName, ui.PhaseColor),
			ui.Value(task.SequenceName, ui.SequenceColor),
			ui.Value(task.Name, ui.TaskColor),
			renderStateLabel(task.Status))
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

	if err := shared.EncodeJSON(os.Stdout, result); err != nil {
		return errors.Wrap(err, "encoding JSON output")
	}
	return nil
}

func renderStateLabel(status string) string {
	label := strings.ReplaceAll(status, "_", " ")
	return ui.GetStateStyle(status).Render(label)
}
