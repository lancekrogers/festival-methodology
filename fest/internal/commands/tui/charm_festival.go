//go:build !no_charm

package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	uitheme "github.com/lancekrogers/festival-methodology/fest/internal/ui/theme"
)

func charmCreateFestival(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Additional variables from templates
	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))
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
			if uitheme.IsCancelled(err) {
				return nil
			}
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
	return shared.RunCreateFestival(ctx, opts)
}

func charmPlanFestivalWizard(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

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
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))
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
			if uitheme.IsCancelled(err) {
				return nil
			}
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

	if err := shared.RunCreateFestival(ctx, &shared.CreateFestivalOpts{Name: name, Goal: goal, Tags: tags, VarsFile: varsFile, Dest: dest}); err != nil {
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
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}
	count := atoiDefault(countStr, 0)
	if addPhases && count > 0 {
		after := 0
		for i := 0; i < count; i++ {
			if err := ctx.Err(); err != nil {
				return err
			}
			var pname, ptype string
			ptype = "planning"
			pf := huh.NewForm(
				huh.NewGroup(
					huh.NewInput().Title(fmt.Sprintf("Phase %d name", i+1)).Placeholder("PLAN").Value(&pname).Validate(notEmpty),
					huh.NewSelect[string]().Title("Phase type").Options(toOptions([]string{"planning", "implementation", "review", "deployment", "research"})...).Value(&ptype),
				),
			).WithTheme(theme())
			if err := pf.Run(); err != nil {
				if uitheme.IsCancelled(err) {
					return nil
				}
				return err
			}
			if err := shared.RunCreatePhase(ctx, &shared.CreatePhaseOpts{After: after, Name: pname, PhaseType: ptype, Path: festivalDir}); err != nil {
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
		if uitheme.IsCancelled(err) {
			return nil
		}
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
