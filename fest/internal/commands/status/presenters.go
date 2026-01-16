package status

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

// isFreeformPhase checks if the phase uses freeform structure.
func isFreeformPhase(phaseName string) bool {
	name := strings.ToLower(phaseName)
	return strings.Contains(name, "planning") ||
		strings.Contains(name, "research") ||
		strings.Contains(name, "design")
}

func emitPhaseProgress(ctx context.Context, loc *show.LocationInfo) error {
	// Check if this is a freeform phase
	if isFreeformPhase(loc.Phase) {
		return emitFreeformPhaseProgress(loc)
	}

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

// emitFreeformPhaseProgress displays progress for planning/research phases.
func emitFreeformPhaseProgress(loc *show.LocationInfo) error {
	phasePath := filepath.Join(loc.Festival.Path, loc.Phase)

	// Count topic directories and documents
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return err
	}

	type topicInfo struct {
		name     string
		docCount int
	}
	var topics []topicInfo
	var totalDocs int

	for _, entry := range entries {
		if entry.IsDir() && !strings.HasPrefix(entry.Name(), ".") {
			topicPath := filepath.Join(phasePath, entry.Name())
			files, _ := filepath.Glob(filepath.Join(topicPath, "*.md"))
			topics = append(topics, topicInfo{entry.Name(), len(files)})
			totalDocs += len(files)
		} else if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
			// Count root-level markdown files (excluding PHASE_GOAL.md)
			if entry.Name() != "PHASE_GOAL.md" {
				totalDocs++
			}
		}
	}

	fmt.Printf("\n%s\n", ui.H2("Planning Progress"))
	fmt.Printf("%s %s\n", ui.Label("Topics"), ui.Value(fmt.Sprintf("%d areas", len(topics))))
	fmt.Printf("%s %s\n", ui.Label("Documents"), ui.Value(fmt.Sprintf("%d created", totalDocs)))

	if len(topics) > 0 {
		fmt.Printf("\n%s\n", ui.Label("Topic Areas"))
		for _, t := range topics {
			fmt.Printf("  • %s %s\n", ui.Value(t.name), ui.Dim(fmt.Sprintf("(%d docs)", t.docCount)))
		}
	}

	// Check for key planning deliverables
	fmt.Printf("\n%s\n", ui.Label("Key Deliverables"))
	deliverables := []struct {
		name string
		path string
	}{
		{"PHASE_GOAL.md", filepath.Join(phasePath, "PHASE_GOAL.md")},
		{"PLANNING_SUMMARY.md", filepath.Join(phasePath, "PLANNING_SUMMARY.md")},
		{"START_HERE.md", filepath.Join(phasePath, "START_HERE.md")},
	}

	for _, d := range deliverables {
		status := "✗"
		if _, err := os.Stat(d.path); err == nil {
			status = "✓"
		}
		fmt.Printf("  %s %s\n", status, d.name)
	}

	// Parse and display planning objectives from PHASE_GOAL.md
	goalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
	objectives, err := parsePlanningObjectives(goalPath)
	if err == nil && len(objectives) > 0 {
		resolved := 0
		for _, obj := range objectives {
			if obj.resolved {
				resolved++
			}
		}
		percentage := float64(resolved) / float64(len(objectives)) * 100

		fmt.Printf("\n%s\n", ui.H2("Planning Objectives"))
		fmt.Printf("%s %s %s\n",
			ui.Label("Progress"),
			ui.Value(fmt.Sprintf("%.0f%%", percentage)),
			ui.Dim(fmt.Sprintf("(%d/%d resolved)", resolved, len(objectives))))

		// Group objectives by category
		categories := map[string][]*planningObjective{
			"question":  {},
			"decision":  {},
			"artifact":  {},
			"objective": {},
		}
		for _, obj := range objectives {
			categories[obj.category] = append(categories[obj.category], obj)
		}

		// Display each category
		categoryLabels := []struct {
			key   string
			label string
		}{
			{"question", "Questions to Answer"},
			{"decision", "Decisions to Make"},
			{"artifact", "Artifacts to Produce"},
			{"objective", "Objectives"},
		}

		for _, cat := range categoryLabels {
			objs := categories[cat.key]
			if len(objs) == 0 {
				continue
			}
			fmt.Printf("\n%s\n", ui.Label(cat.label))
			for _, obj := range objs {
				icon := "○"
				if obj.resolved {
					icon = "●"
				}
				fmt.Printf("  %s %s\n", icon, obj.text)
			}
		}

		// Graduation prompt when 100% complete
		if resolved == len(objectives) {
			fmt.Printf("\n%s\n", ui.Success("All planning objectives resolved!"))
			fmt.Printf("%s Run %s to transition to implementation\n",
				ui.Label("Next"),
				ui.Info("fest graduate"))
		}
	}

	return nil
}

// planningObjective represents a parsed objective from PHASE_GOAL.md
type planningObjective struct {
	category string
	text     string
	resolved bool
}

// parsePlanningObjectives extracts objectives from PHASE_GOAL.md
func parsePlanningObjectives(goalPath string) ([]*planningObjective, error) {
	file, err := os.Open(goalPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()

	var objectives []*planningObjective
	currentCategory := ""

	// Regex patterns for detecting sections and checkboxes
	sectionRegex := regexp.MustCompile(`^###?\s*(Questions?|Decisions?|Artifacts?|Objectives?)`)
	checkboxRegex := regexp.MustCompile(`^[-*]\s*\[([ xX])\]\s*(.+)`)

	scanner := bufio.NewScanner(file)
	inPlanningSection := false

	for scanner.Scan() {
		line := scanner.Text()

		// Check for Planning Objectives header
		if strings.Contains(strings.ToLower(line), "planning objectives") ||
			strings.Contains(strings.ToLower(line), "objectives to achieve") {
			inPlanningSection = true
			continue
		}

		// Check for section headers
		if matches := sectionRegex.FindStringSubmatch(line); len(matches) > 0 {
			inPlanningSection = true
			sectionName := strings.ToLower(matches[1])
			if strings.HasPrefix(sectionName, "question") {
				currentCategory = "question"
			} else if strings.HasPrefix(sectionName, "decision") {
				currentCategory = "decision"
			} else if strings.HasPrefix(sectionName, "artifact") {
				currentCategory = "artifact"
			} else {
				currentCategory = "objective"
			}
			continue
		}

		// Exit planning section on major headers
		if strings.HasPrefix(line, "## ") && !strings.Contains(strings.ToLower(line), "planning") {
			if !strings.Contains(strings.ToLower(line), "question") &&
				!strings.Contains(strings.ToLower(line), "decision") &&
				!strings.Contains(strings.ToLower(line), "artifact") &&
				!strings.Contains(strings.ToLower(line), "objective") {
				inPlanningSection = false
			}
		}

		// Parse checkboxes
		if inPlanningSection || currentCategory != "" {
			if matches := checkboxRegex.FindStringSubmatch(line); len(matches) > 0 {
				resolved := matches[1] == "x" || matches[1] == "X"
				text := strings.TrimSpace(matches[2])
				category := currentCategory
				if category == "" {
					category = "objective"
				}
				objectives = append(objectives, &planningObjective{
					category: category,
					text:     text,
					resolved: resolved,
				})
			}
		}
	}

	return objectives, scanner.Err()
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
