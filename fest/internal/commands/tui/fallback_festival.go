//go:build no_charm

package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

func tuiCreateFestival(display *ui.UI) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	name := strings.TrimSpace(display.Prompt("Festival name"))
	if name == "" {
		return errors.Validation("festival name is required")
	}
	goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
	tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
	dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "active")))
	if dest != "planned" && dest != "active" {
		dest = "active"
	}

	// Collect additional variables required by core festival templates
	required := uniqueStrings(collectRequiredVars(context.TODO(), tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))

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
	return shared.RunCreateFestival(context.TODO(), opts)
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
		return errors.Validation("festival name is required")
	}
	goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
	tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
	dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "planned")))
	if dest != "planned" && dest != "active" {
		dest = "planned"
	}

	// gather extra vars from templates
	required := uniqueStrings(collectRequiredVars(context.TODO(), cwdTmpl, defaultFestivalTemplatePaths(cwdTmpl)))
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

	if err := shared.RunCreateFestival(context.TODO(), &shared.CreateFestivalOpts{Name: name, Goal: goal, Tags: tags, VarsFile: varsFile, Dest: dest}); err != nil {
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
			ptype := strings.TrimSpace(display.PromptDefault("Phase type (planning|implementation|review|deployment|research)", "planning"))
			if ptype == "" {
				ptype = "planning"
			}
			if err := shared.RunCreatePhase(context.TODO(), &shared.CreatePhaseOpts{After: after, Name: pname, PhaseType: ptype, Path: festivalDir}); err != nil {
				return err
			}
			after++
		}
	}
	display.Success("Festival planned: %s (%s)", slug, dest)
	display.Info("Location: %s", festivalDir)
	return nil
}

func tuiGenerateFestivalGoal(ctx context.Context, display *ui.UI) error {
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
	return shared.RunApply(ctx, &shared.ApplyOpts{TemplatePath: "FESTIVAL_GOAL_TEMPLATE.md", DestPath: destPath, VarsFile: varsFile})
}
