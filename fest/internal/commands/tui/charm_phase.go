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

func charmCreatePhase(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Use hierarchical selector to select festival
	selector, err := NewHierarchySelectorFromCwd(SelectToFestival(false))
	if err != nil {
		// Fallback to manual path entry
		return charmCreatePhaseManual(ctx)
	}

	state, err := selector.Run(ctx)
	if err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Continue with phase creation in selected festival
	return createPhaseInFestival(ctx, state.FestivalPath)
}

// createPhaseInFestival creates a phase in the specified festival.
func createPhaseInFestival(ctx context.Context, festivalPath string) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	var name string
	phaseTypes := []string{"planning", "implementation", "review", "deployment", "research"}
	var phaseType string = phaseTypes[0]

	// Phase name and type form
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Phase name").
				Placeholder("PLANNING").
				Description(fmt.Sprintf("Creating in: %s", filepath.Base(festivalPath))).
				Value(&name).
				Validate(notEmpty),
			huh.NewSelect[string]().
				Title("Phase type").
				Options(toOptions(phaseTypes)...).
				Value(&phaseType),
		),
	).WithTheme(theme())

	if err := form.Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Position selection
	defAfter := nextPhaseAfter(ctx, festivalPath)
	var posPhase string
	var afterStr string

	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Position").
			Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&posPhase),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	if posPhase == "insert" {
		afterStr = fmt.Sprintf("%d", defAfter)
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().
				Title("Insert after number (0 to insert at beginning)").
				Value(&afterStr),
		)).WithTheme(theme()).Run(); err != nil {
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
	} else {
		afterStr = fmt.Sprintf("%d", defAfter)
	}

	after := atoiDefault(afterStr, defAfter)

	// Collect additional template variables
	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot,
		[]string{filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "phase_number" || k == "phase_name" || k == "phase_type" {
			continue
		}
		var v string
		if err := huh.NewForm(huh.NewGroup(
			huh.NewInput().Title(k).Value(&v),
		)).WithTheme(theme()).Run(); err != nil {
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

	opts := &shared.CreatePhaseOpts{
		After:     after,
		Name:      name,
		PhaseType: phaseType,
		Path:      festivalPath,
		VarsFile:  varsFile,
	}
	return shared.RunCreatePhase(ctx, opts)
}

// charmCreatePhaseManual is the fallback when hierarchical selector can't initialize.
func charmCreatePhaseManual(ctx context.Context) error {
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
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}
	basePath := path
	if strings.TrimSpace(basePath) == "" || basePath == "." {
		basePath = findFestivalDir(cwd)
	}
	defAfter := nextPhaseAfter(ctx, basePath)
	// Choose position for phase
	var posPhase string
	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Position").Options(
			huh.NewOption("Append at end", "append"),
			huh.NewOption("Insert after number", "insert"),
		).Value(&posPhase),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}
	if posPhase == "insert" {
		afterStr = fmt.Sprintf("%d", defAfter)
		if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Insert after number (0 to insert at beginning)").Value(&afterStr))).WithTheme(theme()).Run(); err != nil {
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
	} else {
		afterStr = fmt.Sprintf("%d", defAfter)
	}
	after := atoiDefault(afterStr, defAfter)

	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot, []string{filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "phase_number" || k == "phase_name" || k == "phase_type" {
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
	opts := &shared.CreatePhaseOpts{After: after, Name: name, PhaseType: phaseType, Path: fallbackDot(path), VarsFile: varsFile}
	return shared.RunCreatePhase(ctx, opts)
}
