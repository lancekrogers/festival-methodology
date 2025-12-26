//go:build !no_charm

package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
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

// NewTUICommand (charm version) provides a richer interactive UI using Charm libs
func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive UI (Charm) for festival creation and editing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCharmTUI(cmd.Context())
		},
	}
	return cmd
}

func runCharmTUI(ctx context.Context) error {
	// Validate inside festivals workspace; if absent, offer to init
	cwd, _ := os.Getwd()
	if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
		var initNow bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().Title("No festivals/ directory detected. Initialize here?").Value(&initNow),
			),
		).WithTheme(theme())
		if err := form.Run(); err != nil {
			return err
		}
		if initNow {
			if err := shared.RunInit(ctx, ".", &shared.InitOpts{}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no festivals/ directory detected")
		}
	}

	// main menu loop
	for {
		var action string
		menu := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Plan a New Festival (wizard)", "plan_festival"),
						huh.NewOption("Create a Festival (quick)", "create_festival"),
						huh.NewOption("Add a Phase", "create_phase"),
						huh.NewOption("Add a Sequence", "create_sequence"),
						huh.NewOption("Add a Task", "create_task"),
						huh.NewOption("Generate Festival Goal", "festival_goal"),
						huh.NewOption("Quit", "quit"),
					).
					Value(&action),
			),
		).WithTheme(huh.ThemeBase())

		if err := menu.Run(); err != nil {
			return err
		}

		switch action {
		case "plan_festival":
			if err := charmPlanFestivalWizard(); err != nil {
				return err
			}
		case "create_festival":
			if err := charmCreateFestival(); err != nil {
				return err
			}
		case "create_phase":
			if err := charmCreatePhase(); err != nil {
				return err
			}
		case "create_sequence":
			if err := charmCreateSequence(); err != nil {
				return err
			}
		case "create_task":
			if err := charmCreateTask(); err != nil {
				return err
			}
		case "festival_goal":
			if err := charmGenerateFestivalGoal(ctx); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func charmCreateFestival() error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	var name, goal, tags string
	var dest string = "active"
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Festival name").Placeholder("e.g., ecommerce-platform").Value(&name).Validate(notEmpty),
			huh.NewInput().Title("Festival goal").Placeholder("e.g., Launch MVP").Value(&goal),
			huh.NewInput().Title("Tags (comma-separated)").Placeholder("backend,security").Value(&tags),
			huh.NewSelect[string]().Title("Destination").Options(
				huh.NewOption("active", "active"),
				huh.NewOption("planned", "planned"),
			).Value(&dest),
		),
	).WithTheme(theme())
	if err := form.Run(); err != nil {
		return err
	}

	// Additional variables from templates
	required := uniqueStrings(collectRequiredVars(tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))
	vars := map[string]interface{}{"festival_name": name, "festival_goal": goal}
	if strings.TrimSpace(tags) != "" {
		vars["festival_tags"] = strings.Split(tags, ",")
	}
	// Collect missing variables one-by-one
	for _, k := range required {
		if k == "festival_name" || k == "festival_goal" || k == "festival_tags" || k == "festival_description" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title(k).Value(&v))).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}

	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	opts := &shared.CreateFestivalOpts{Name: name, Goal: goal, Tags: tags, VarsFile: varsFile, Dest: dest}
	return shared.RunCreateFestival(opts)
}

func charmPlanFestivalWizard() error {
	cwd, _ := os.Getwd()
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return err
	}
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, goal, tags string
	var dest string = "planned"
	base := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Festival name").Placeholder("e.g., ecommerce-platform").Value(&name).Validate(notEmpty),
			huh.NewInput().Title("Festival goal").Placeholder("e.g., Launch MVP").Value(&goal),
			huh.NewInput().Title("Tags (comma-separated)").Placeholder("backend,security").Value(&tags),
			huh.NewSelect[string]().Title("Destination").Options(
				huh.NewOption("planned", "planned"),
				huh.NewOption("active", "active"),
			).Value(&dest),
		),
	).WithTheme(theme())
	if err := base.Run(); err != nil {
		return err
	}

	required := uniqueStrings(collectRequiredVars(tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))
	vars := map[string]interface{}{"festival_name": name, "festival_goal": goal}
	if strings.TrimSpace(tags) != "" {
		vars["festival_tags"] = strings.Split(tags, ",")
	}
	for _, k := range required {
		if k == "festival_name" || k == "festival_goal" || k == "festival_tags" || k == "festival_description" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title(k).Value(&v))).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}

	if err := shared.RunCreateFestival(&shared.CreateFestivalOpts{Name: name, Goal: goal, Tags: tags, VarsFile: varsFile, Dest: dest}); err != nil {
		return err
	}
	slug := slugify(name)
	festivalDir := filepath.Join(festivalsRoot, dest, slug)

	// Add phases
	var addPhases bool
	var countStr string
	phasesForm := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().Title("Add initial phases now?").Value(&addPhases),
			huh.NewInput().Title("How many phases?").Value(&countStr),
		),
	).WithTheme(theme())
	if err := phasesForm.Run(); err != nil {
		return err
	}
	count := atoiDefault(countStr, 0)
	if addPhases && count > 0 {
		after := 0
		for i := 0; i < count; i++ {
			var pname, ptype string
			ptype = "planning"
			pf := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title(fmt.Sprintf("Phase %d name", i+1)).Placeholder("PLAN").Value(&pname).Validate(notEmpty),
					huh.NewSelect[string]().Title("Phase type").Options(toOptions([]string{"planning", "implementation", "review", "deployment", "research"})...).Value(&ptype),
				),
			).WithTheme(theme())
			if err := pf.Run(); err != nil {
				return err
			}
			if err := shared.RunCreatePhase(&shared.CreatePhaseOpts{After: after, Name: pname, PhaseType: ptype, Path: festivalDir}); err != nil {
				return err
			}
			after++
		}
	}
	return nil
}

func charmGenerateFestivalGoal(ctx context.Context) error {
	cwd, _ := os.Getwd()
	if _, err := tpl.LocalTemplateRoot(cwd); err != nil {
		return err
	}
	var festDir, name, goal string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Festival directory").Placeholder(".").Value(&festDir),
			huh.NewInput().Title("festival_name").Value(&name),
			huh.NewInput().Title("festival_goal").Value(&goal),
		),
	).WithTheme(theme())
	if err := form.Run(); err != nil {
		return err
	}
	vars := map[string]interface{}{}
	if strings.TrimSpace(name) != "" {
		vars["festival_name"] = name
	}
	if strings.TrimSpace(goal) != "" {
		vars["festival_goal"] = goal
	}
	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}
	destPath := filepath.Join(festDir, "FESTIVAL_GOAL.md")
	return shared.RunApply(ctx, &shared.ApplyOpts{TemplatePath: "FESTIVAL_GOAL_TEMPLATE.md", DestPath: destPath, VarsFile: varsFile})
}

func charmCreatePhase() error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, path, afterStr string
	phaseTypes := []string{"planning", "implementation", "review", "deployment", "research"}
	var phaseType string = phaseTypes[0]

	// Two-step: first fields; then compute default 'after'
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("Phase name").Placeholder("PLAN").Value(&name).Validate(notEmpty),
			huh.NewSelect[string]().Title("Phase type").Options(
				toOptions(phaseTypes)...,
			).Value(&phaseType),
			huh.NewInput().Title("Festival directory (contains numbered phases)").Placeholder(".").Value(&path),
		),
	).WithTheme(theme())
	if err := form.Run(); err != nil {
		return err
	}
	basePath := path
	if strings.TrimSpace(basePath) == "" || basePath == "." {
		basePath = findFestivalDir(cwd)
	}
	defAfter := nextPhaseAfter(basePath)
	// Choose position for phase
	var posPhase string
	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Position").Options(
			huh.NewOption("Append at end", "append"),
			huh.NewOption("Insert after number", "insert"),
		).Value(&posPhase),
	)).WithTheme(theme()).Run(); err != nil {
		return err
	}
	if posPhase == "insert" {
		afterStr = fmt.Sprintf("%d", defAfter)
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
			return err
		}
	} else {
		afterStr = fmt.Sprintf("%d", defAfter)
	}
	after := atoiDefault(afterStr, defAfter)

	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "phase_number" || k == "phase_name" || k == "phase_type" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title(k).Value(&v))).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}

	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}
	opts := &shared.CreatePhaseOpts{After: after, Name: name, PhaseType: phaseType, Path: fallbackDot(path), VarsFile: varsFile}
	return shared.RunCreatePhase(opts)
}

func charmCreateSequence() error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, path, afterStr string
	inPhase := isPhaseDirPath(cwd)

	if inPhase {
		// Name first
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Sequence name").Placeholder("requirements").Value(&name).Validate(notEmpty),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		// Position selection with default append
		defAfter := nextSequenceAfter(cwd)
		var pos string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Position").Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&pos),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if pos == "insert" {
			afterStr = fmt.Sprintf("%d", defAfter)
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
				return err
			}
		} else {
			afterStr = fmt.Sprintf("%d", defAfter)
		}
	} else {
		// Offer phase picker if available
		festDir := findFestivalDir(cwd)
		phases := listPhaseDirs(festDir)
		if len(phases) > 0 {
			var selected string
			opts := make([]huh.Option[string], 0, len(phases)+1)
			for _, p := range phases {
				opts = append(opts, huh.NewOption(p, filepath.Join(festDir, p)))
			}
			opts = append(opts, huh.NewOption("Other...", "__other__"))
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Sequence name").Placeholder("requirements").Value(&name).Validate(notEmpty),
					huh.NewSelect[string]().Title("Select phase").Options(opts...).Value(&selected),
				),
			).WithTheme(theme())
			if err := form.Run(); err != nil {
				return err
			}
			if selected == "__other__" {
				// Ask for manual path/number
				if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Phase (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
					return err
				}
			} else {
				path = selected
			}
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Sequence name").Placeholder("requirements").Value(&name).Validate(notEmpty),
					huh.NewInput().Title("Phase (dir or number, e.g., 002 or 002_IMPLEMENT)").Placeholder(".").Value(&path),
				),
			).WithTheme(theme())
			if err := form.Run(); err != nil {
				return err
			}
		}
		// After selection
		rp, rerr := resolvePhaseDirInput(path, cwd)
		if rerr != nil {
			return rerr
		}
		defAfter := nextSequenceAfter(rp)
		var pos string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Position").Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&pos),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if pos == "insert" {
			afterStr = fmt.Sprintf("%d", defAfter)
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
				return err
			}
		} else {
			afterStr = fmt.Sprintf("%d", defAfter)
		}
	}
	after := atoiDefault(afterStr, 0)
	resolvedPath := cwd
	if !inPhase {
		rp, rerr := resolvePhaseDirInput(path, cwd)
		if rerr != nil {
			return rerr
		}
		resolvedPath = rp
	}

	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "sequence_number" || k == "sequence_name" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title(k).Value(&v))).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}

	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}
	opts := &shared.CreateSequenceOpts{After: after, Name: name, Path: fallbackDot(resolvedPath), VarsFile: varsFile}
	return shared.RunCreateSequence(opts)
}

func charmCreateTask() error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, path, afterStr string
	inSequence := isSequenceDirPath(cwd)

	if inSequence {
		// Name first
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title("Task name").Placeholder("user_research").Value(&name).Validate(notEmpty),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		// Position select with default append
		defAfter := nextTaskAfter(cwd)
		var pos string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Position").Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&pos),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if pos == "insert" {
			afterStr = fmt.Sprintf("%d", defAfter)
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
				return err
			}
		} else {
			afterStr = fmt.Sprintf("%d", defAfter)
		}
	} else {
		// Not in a sequence: choose phase (if needed), then sequence
		var phasePath string
		if isPhaseDirPath(cwd) {
			phasePath = cwd
		} else {
			festDir := findFestivalDir(cwd)
			phases := listPhaseDirs(festDir)
			if len(phases) > 0 {
				var pSel string
				pOpts := make([]huh.Option[string], 0, len(phases)+1)
				for _, p := range phases {
					pOpts = append(pOpts, huh.NewOption(p, filepath.Join(festDir, p)))
				}
				pOpts = append(pOpts, huh.NewOption("Other...", "__other__"))
				pf := huh.NewForm(huh.NewGroup(
					huh.NewSelect[string]().Title("Select phase").Options(pOpts...).Value(&pSel),
				)).WithTheme(theme())
				if err := pf.Run(); err != nil {
					return err
				}
				if pSel == "__other__" {
					if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Phase (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
						return err
					}
					rp, rerr := resolvePhaseDirInput(path, cwd)
					if rerr != nil {
						return rerr
					}
					phasePath = rp
				} else {
					phasePath = pSel
				}
			} else {
				if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Phase (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
					return err
				}
				rp, rerr := resolvePhaseDirInput(path, cwd)
				if rerr != nil {
					return rerr
				}
				phasePath = rp
			}
		}

		// Now choose sequence within the selected phase
		seqs := listSequenceDirs(phasePath)
		if len(seqs) > 0 {
			var sSel string
			sOpts := make([]huh.Option[string], 0, len(seqs)+1)
			for _, s := range seqs {
				sOpts = append(sOpts, huh.NewOption(s, filepath.Join(phasePath, s)))
			}
			sOpts = append(sOpts, huh.NewOption("Other...", "__other__"))
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Task name").Placeholder("user_research").Value(&name).Validate(notEmpty),
					huh.NewSelect[string]().Title("Select sequence").Options(sOpts...).Value(&sSel),
				),
			).WithTheme(theme())
			if err := form.Run(); err != nil {
				return err
			}
			if sSel == "__other__" {
				if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Sequence (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
					return err
				}
			} else {
				path = sSel
			}
		} else {
			form := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title("Task name").Placeholder("user_research").Value(&name).Validate(notEmpty),
					huh.NewInput().Title("Sequence (dir or number, e.g., 01 or 01_requirements)").Placeholder(".").Value(&path),
				),
			).WithTheme(theme())
			if err := form.Run(); err != nil {
				return err
			}
		}
		// Compute default after based on resolved sequence path, then choose position
		rs, rerr := resolveSequenceDirInput(path, cwd)
		if rerr != nil {
			return rerr
		}
		defAfter := nextTaskAfter(rs)
		var pos string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewSelect[string]().Title("Position").Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&pos),
		)).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if pos == "insert" {
			afterStr = fmt.Sprintf("%d", defAfter)
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
				return err
			}
		} else {
			afterStr = fmt.Sprintf("%d", defAfter)
		}
	}
	after := atoiDefault(afterStr, 0)

	required := uniqueStrings(collectRequiredVars(tmplRoot, []string{filepath.Join(tmplRoot, "TASK_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "task_number" || k == "task_name" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title(k).Value(&v))).WithTheme(theme()).Run(); err != nil {
			return err
		}
		if strings.TrimSpace(v) != "" {
			vars[k] = v
		}
	}

	varsFile, err := writeTempVarsFile(vars)
	if err != nil {
		return err
	}
	resolvedSeq := cwd
	if !inSequence {
		// If we selected a phase and sequence via pickers above, 'path' will be a full directory.
		// Otherwise, resolve the user's input relative to current cwd (phase-aware if cwd is a phase)
		rs, rerr := resolveSequenceDirInput(path, cwd)
		if rerr != nil {
			return rerr
		}
		resolvedSeq = rs
	}
	opts := &shared.CreateTaskOpts{After: after, Names: []string{name}, Path: fallbackDot(resolvedSeq), VarsFile: varsFile}
	return shared.RunCreateTask(opts)
}

func notEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("value required")
	}
	return nil
}

func toOptions(values []string) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(values))
	for _, v := range values {
		opts = append(opts, huh.NewOption(v, v))
	}
	return opts
}

func theme() *huh.Theme {
	th := huh.ThemeBase()
	th.Focused.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	th.Focused.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	th.Blurred.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	return th
}

func fallbackDot(s string) string {
	if strings.TrimSpace(s) == "" {
		return "."
	}
	return s
}

// StartCreateTUI shows a create-only menu (Charm implementation)
func StartCreateTUI() error {
	for {
		var action string
		menu := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Create what?").
					Options(
						huh.NewOption("Festival", "festival"),
						huh.NewOption("Phase", "phase"),
						huh.NewOption("Sequence", "sequence"),
						huh.NewOption("Task", "task"),
						huh.NewOption("Back", "back"),
					).
					Value(&action),
			),
		).WithTheme(theme())
		if err := menu.Run(); err != nil {
			return err
		}
		switch action {
		case "festival":
			if err := charmCreateFestival(); err != nil {
				return err
			}
		case "phase":
			if err := charmCreatePhase(); err != nil {
				return err
			}
		case "sequence":
			if err := charmCreateSequence(); err != nil {
				return err
			}
		case "task":
			if err := charmCreateTask(); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func StartCreateFestivalTUI() error { return charmCreateFestival() }
func StartCreatePhaseTUI() error    { return charmCreatePhase() }
func StartCreateSequenceTUI() error { return charmCreateSequence() }
func StartCreateTaskTUI() error     { return charmCreateTask() }
