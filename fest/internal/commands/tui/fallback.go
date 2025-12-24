//go:build no_charm

package tui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	// Register TUI hooks with the shared package
	shared.NewTUICommand = NewTUICommand
	shared.StartCreateTUI = StartCreateTUI
	shared.StartCreateFestivalTUI = StartCreateFestivalTUI
	shared.StartCreatePhaseTUI = StartCreatePhaseTUI
	shared.StartCreateSequenceTUI = StartCreateSequenceTUI
	shared.StartCreateTaskTUI = StartCreateTaskTUI
}

// NewTUICommand launches an interactive text UI for common actions
func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive UI for creating festivals, phases, sequences, and tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI()
		},
	}
	return cmd
}

func runTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Ensure we are inside a festivals workspace; if not, offer to init
	if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
		display.Warning("No festivals/ directory detected.")
		if display.Confirm("Initialize a new festival workspace here?") {
			if err := shared.RunInit(".", &shared.InitOpts{}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no festivals/ directory detected")
		}
	}

	for {
		choice := display.Choose("What would you like to do?", []string{
			"Plan a New Festival (wizard)",
			"Create a Festival (quick)",
			"Add a Phase",
			"Add a Sequence",
			"Add a Task",
			"Generate Festival Goal",
			"Quit",
		})

		switch choice {
		case 0:
			if err := tuiPlanFestivalWizard(display); err != nil {
				return err
			}
		case 1:
			if err := tuiCreateFestival(display); err != nil {
				return err
			}
		case 2:
			if err := tuiCreatePhase(display); err != nil {
				return err
			}
		case 3:
			if err := tuiCreateSequence(display); err != nil {
				return err
			}
		case 4:
			if err := tuiCreateTask(display); err != nil {
				return err
			}
		case 5:
			if err := tuiGenerateFestivalGoal(display); err != nil {
				return err
			}
		default:
			return nil
		}

		if !display.Confirm("Do you want to perform another action?") {
			break
		}
	}
	return nil
}

func tuiCreateFestival(display *ui.UI) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(display.Prompt("Festival name"))
	if name == "" {
		return fmt.Errorf("festival name is required")
	}
	goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
	tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
	dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "active")))
	if dest != "planned" && dest != "active" {
		dest = "active"
	}

	// Collect additional variables required by core festival templates
	required := uniqueStrings(collectRequiredVars(tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))

	vars := map[string]interface{}{}
	// Pre-populate typical variables
	vars["festival_name"] = name
	vars["festival_goal"] = goal
	if tags != "" {
		// keep as string; create command handles tags flag for standardized parsing
		vars["festival_tags"] = strings.Split(tags, ",")
	}

	// Ask for any missing variables not already filled
	for _, v := range required {
		if v == "festival_name" || v == "festival_goal" || v == "festival_tags" || v == "festival_description" {
			continue
		}
		if _, ok := vars[v]; ok {
			continue
		}
		val := strings.TrimSpace(display.PromptDefault(fmt.Sprintf("%s", v), ""))
		if val != "" {
			vars[v] = val
		}
	}

	// Write variables to a temporary JSON file
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	opts := &shared.CreateFestivalOpts{
		Name:     name,
		Goal:     goal,
		Tags:     tags,
		VarsFile: varsFile,
		Dest:     dest,
	}
	return shared.RunCreateFestival(opts)
}

// Wizard: create festival then optionally add phases
func tuiPlanFestivalWizard(display *ui.UI) error {
	cwd, _ := os.Getwd()
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return err
	}
	// First create festival (quick)
	cwdTmpl, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	name := strings.TrimSpace(display.Prompt("Festival name"))
	if name == "" {
		return fmt.Errorf("festival name is required")
	}
	goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
	tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
	dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "planned")))
	if dest != "planned" && dest != "active" {
		dest = "planned"
	}

	// gather extra vars from templates
	required := uniqueStrings(collectRequiredVars(cwdTmpl, defaultFestivalTemplatePaths(cwdTmpl)))
	vars := map[string]interface{}{"festival_name": name, "festival_goal": goal}
	if tags != "" {
		vars["festival_tags"] = strings.Split(tags, ",")
	}
	for _, v := range required {
		if v == "festival_name" || v == "festival_goal" || v == "festival_tags" || v == "festival_description" {
			continue
		}
		val := strings.TrimSpace(display.PromptDefault(v, ""))
		if val != "" {
			vars[v] = val
		}
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	if err := shared.RunCreateFestival(&shared.CreateFestivalOpts{Name: name, Goal: goal, Tags: tags, VarsFile: varsFile, Dest: dest}); err != nil {
		return err
	}

	// Compute created path
	slug := slugify(name)
	festivalDir := filepath.Join(festivalsRoot, dest, slug)

	// Optionally add phases
	if display.Confirm("Add initial phases now?") {
		countStr := display.PromptDefault("How many phases?", "0")
		count := atoiDefault(countStr, 0)
		after := 0
		for i := 0; i < count; i++ {
			pname := strings.TrimSpace(display.Prompt(fmt.Sprintf("Phase %d name (e.g., PLAN)", i+1)))
			if pname == "" {
				pname = fmt.Sprintf("PHASE_%d", i+1)
			}
			ptype := strings.TrimSpace(display.PromptDefault("Phase type (planning|implementation|review|deployment)", "planning"))
			if ptype == "" {
				ptype = "planning"
			}
			if err := shared.RunCreatePhase(&shared.CreatePhaseOpts{After: after, Name: pname, PhaseType: ptype, Path: festivalDir}); err != nil {
				return err
			}
			after++
		}
	}
	display.Success("Festival planned: %s (%s)", slug, dest)
	display.Info("Location: %s", festivalDir)
	return nil
}

func tuiGenerateFestivalGoal(display *ui.UI) error {
	cwd, _ := os.Getwd()
	if _, err := tpl.LocalTemplateRoot(cwd); err != nil {
		return err
	}
	festDir := strings.TrimSpace(display.PromptDefault("Festival directory (where to write FESTIVAL_GOAL.md)", "."))
	if festDir == "" {
		festDir = "."
	}
	name := strings.TrimSpace(display.PromptDefault("festival_name", ""))
	goal := strings.TrimSpace(display.PromptDefault("festival_goal", ""))
	vars := map[string]interface{}{}
	if name != "" {
		vars["festival_name"] = name
	}
	if goal != "" {
		vars["festival_goal"] = goal
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}
	// Use apply to render template to destination
	destPath := filepath.Join(festDir, "FESTIVAL_GOAL.md")
	return shared.RunApply(&shared.ApplyOpts{TemplatePath: "FESTIVAL_GOAL_TEMPLATE.md", DestPath: destPath, VarsFile: varsFile})
}

func tuiCreatePhase(display *ui.UI) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(display.Prompt("Phase name (e.g., PLAN)"))
	if name == "" {
		return fmt.Errorf("phase name is required")
	}
	// Choose phase type
	types := []string{"planning", "implementation", "review", "deployment"}
	tIdx := display.Choose("Phase type:", types)
	if tIdx < 0 || tIdx >= len(types) {
		tIdx = 0
	}
	phaseType := types[tIdx]

	path := strings.TrimSpace(display.PromptDefault("Festival directory (contains numbered phases)", "."))
	// Default to appending at end
	festDir := path
	if festDir == "." || festDir == "" {
		festDir = findFestivalDir(cwd)
	}
	defAfter := nextPhaseAfter(festDir)
	after := defAfter
	if !display.Confirm("Append at end?") {
		afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", fmt.Sprintf("%d", defAfter)))
		after = atoiDefault(afterStr, defAfter)
	}

	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
		filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md"),
	}))
	vars := map[string]interface{}{}
	// Gather missing variables
	for _, v := range required {
		// Context will already set phase numbering/name/type; ask for extras only
		if v == "phase_number" || v == "phase_name" || v == "phase_type" {
			continue
		}
		val := strings.TrimSpace(display.PromptDefault(v, ""))
		if val != "" {
			vars[v] = val
		}
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	opts := &shared.CreatePhaseOpts{
		After:     after,
		Name:      name,
		PhaseType: phaseType,
		Path:      path,
		VarsFile:  varsFile,
	}
	return shared.RunCreatePhase(opts)
}

func tuiCreateSequence(display *ui.UI) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(display.Prompt("Sequence name (e.g., requirements)"))
	if name == "" {
		return fmt.Errorf("sequence name is required")
	}
	var resolvedPhase string
	if isPhaseDirPath(cwd) {
		resolvedPhase = cwd
	} else {
		// Offer a quick picker of phases if available
		festDir := findFestivalDir(cwd)
		phases := listPhaseDirs(festDir)
		if len(phases) > 0 {
			items := append(append([]string{}, phases...), "Other...")
			idx := display.Choose("Select a phase:", items)
			if idx >= 0 && idx < len(phases) {
				resolvedPhase = filepath.Join(festDir, phases[idx])
			} else {
				// Fallback to manual input
				path := strings.TrimSpace(display.PromptDefault("Phase (dir or number, e.g., 002 or 002_IMPLEMENT)", "."))
				rp, rerr := resolvePhaseDirInput(path, cwd)
				if rerr != nil {
					return fmt.Errorf("invalid phase: %w", rerr)
				}
				resolvedPhase = rp
			}
		} else {
			path := strings.TrimSpace(display.PromptDefault("Phase (dir or number, e.g., 002 or 002_IMPLEMENT)", "."))
			rp, rerr := resolvePhaseDirInput(path, cwd)
			if rerr != nil {
				return fmt.Errorf("invalid phase: %w", rerr)
			}
			resolvedPhase = rp
		}
	}
	// Default to append after last sequence in resolved phase
	defAfter := nextSequenceAfter(resolvedPhase)
	after := defAfter
	if !display.Confirm("Append at end?") {
		afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", fmt.Sprintf("%d", defAfter)))
		after = atoiDefault(afterStr, defAfter)
	}

	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
		filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md"),
	}))
	vars := map[string]interface{}{}
	for _, v := range required {
		if v == "sequence_number" || v == "sequence_name" {
			continue
		}
		val := strings.TrimSpace(display.PromptDefault(v, ""))
		if val != "" {
			vars[v] = val
		}
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	opts := &shared.CreateSequenceOpts{
		After:    after,
		Name:     name,
		Path:     resolvedPhase,
		VarsFile: varsFile,
	}
	return shared.RunCreateSequence(opts)
}

func tuiCreateTask(display *ui.UI) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(display.Prompt("Task name (e.g., user_research)"))
	if name == "" {
		return fmt.Errorf("task name is required")
	}
	var resolvedSeq string
	if isSequenceDirPath(cwd) {
		// Already in a sequence directory
		resolvedSeq = cwd
	} else if isPhaseDirPath(cwd) {
		// In a phase directory: offer sequence picker within this phase
		seqs := listSequenceDirs(cwd)
		if len(seqs) > 0 {
			items := append(append([]string{}, seqs...), "Other...")
			idx := display.Choose("Select a sequence:", items)
			if idx >= 0 && idx < len(seqs) {
				resolvedSeq = filepath.Join(cwd, seqs[idx])
			} else {
				path := strings.TrimSpace(display.PromptDefault("Sequence (dir or number, e.g., 01 or 01_requirements)", "."))
				rs, rerr := resolveSequenceDirInput(path, cwd)
				if rerr != nil {
					return fmt.Errorf("invalid sequence: %w", rerr)
				}
				resolvedSeq = rs
			}
		} else {
			path := strings.TrimSpace(display.PromptDefault("Sequence (dir or number, e.g., 01 or 01_requirements)", "."))
			rs, rerr := resolveSequenceDirInput(path, cwd)
			if rerr != nil {
				return fmt.Errorf("invalid sequence: %w", rerr)
			}
			resolvedSeq = rs
		}
	} else {
		// Not in phase or sequence: pick a phase first, then a sequence within it
		festDir := findFestivalDir(cwd)
		phases := listPhaseDirs(festDir)
		var chosenPhase string
		if len(phases) > 0 {
			items := append(append([]string{}, phases...), "Other...")
			idx := display.Choose("Select a phase:", items)
			if idx >= 0 && idx < len(phases) {
				chosenPhase = filepath.Join(festDir, phases[idx])
			} else {
				p := strings.TrimSpace(display.PromptDefault("Phase (dir or number, e.g., 002 or 002_IMPLEMENT)", "."))
				rp, rerr := resolvePhaseDirInput(p, cwd)
				if rerr != nil {
					return fmt.Errorf("invalid phase: %w", rerr)
				}
				chosenPhase = rp
			}
		} else {
			p := strings.TrimSpace(display.PromptDefault("Phase (dir or number, e.g., 002 or 002_IMPLEMENT)", "."))
			rp, rerr := resolvePhaseDirInput(p, cwd)
			if rerr != nil {
				return fmt.Errorf("invalid phase: %w", rerr)
			}
			chosenPhase = rp
		}

		// Now pick sequence within chosen phase
		seqs := listSequenceDirs(chosenPhase)
		if len(seqs) > 0 {
			items := append(append([]string{}, seqs...), "Other...")
			idx := display.Choose("Select a sequence:", items)
			if idx >= 0 && idx < len(seqs) {
				resolvedSeq = filepath.Join(chosenPhase, seqs[idx])
			} else {
				s := strings.TrimSpace(display.PromptDefault("Sequence (dir or number, e.g., 01 or 01_requirements)", "."))
				rs, rerr := resolveSequenceDirInput(s, chosenPhase)
				if rerr != nil {
					return fmt.Errorf("invalid sequence: %w", rerr)
				}
				resolvedSeq = rs
			}
		} else {
			s := strings.TrimSpace(display.PromptDefault("Sequence (dir or number, e.g., 01 or 01_requirements)", "."))
			rs, rerr := resolveSequenceDirInput(s, chosenPhase)
			if rerr != nil {
				return fmt.Errorf("invalid sequence: %w", rerr)
			}
			resolvedSeq = rs
		}
	}
	// Default to append after last task in resolved sequence
	defAfter := nextTaskAfter(resolvedSeq)
	after := defAfter
	if !display.Confirm("Append at end?") {
		afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", fmt.Sprintf("%d", defAfter)))
		after = atoiDefault(afterStr, defAfter)
	}

	// Prefer TASK_TEMPLATE.md for required vars
	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
		filepath.Join(tmplRoot, "TASK_TEMPLATE.md"),
	}))
	vars := map[string]interface{}{}
	for _, v := range required {
		if v == "task_number" || v == "task_name" {
			continue
		}
		val := strings.TrimSpace(display.PromptDefault(v, ""))
		if val != "" {
			vars[v] = val
		}
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	opts := &shared.CreateTaskOpts{
		After:    after,
		Names:    []string{name},
		Path:     resolvedSeq,
		VarsFile: varsFile,
	}
	return shared.RunCreateTask(opts)
}

// StartCreateTUI shows a create-only menu (fallback implementation)
func StartCreateTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	for {
		choice := display.Choose("Create what?", []string{
			"Festival",
			"Phase",
			"Sequence",
			"Task",
			"Back",
		})
		switch choice {
		case 0:
			if err := tuiCreateFestival(display); err != nil {
				return err
			}
		case 1:
			if err := tuiCreatePhase(display); err != nil {
				return err
			}
		case 2:
			if err := tuiCreateSequence(display); err != nil {
				return err
			}
		case 3:
			if err := tuiCreateTask(display); err != nil {
				return err
			}
		default:
			return nil
		}
		if !display.Confirm("Create another?") {
			return nil
		}
	}
}

func StartCreateFestivalTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateFestival(display)
}

func StartCreatePhaseTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreatePhase(display)
}

func StartCreateSequenceTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateSequence(display)
}

func StartCreateTaskTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateTask(display)
}
