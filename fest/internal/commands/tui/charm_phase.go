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
)

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

	required := uniqueStrings(collectRequiredVars(context.TODO(), tmplRoot, []string{filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md")}))
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
	return shared.RunCreatePhase(context.TODO(), opts)
}
