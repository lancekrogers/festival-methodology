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

func charmCreateSequence(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Use hierarchical selector to select festival and phase
	selector, err := NewHierarchySelectorFromCwd(SelectToPhase(true))
	if err != nil {
		// Fallback to manual path entry
		return charmCreateSequenceManual(ctx)
	}

	state, err := selector.Run(ctx)
	if err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Continue with sequence creation in selected phase
	return createSequenceInPhase(ctx, state.PhasePath)
}

// createSequenceInPhase creates a sequence in the specified phase.
func createSequenceInPhase(ctx context.Context, phasePath string) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	var name string

	// Sequence name form
	if err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Sequence name").
			Placeholder("requirements").
			Description(fmt.Sprintf("Creating in: %s", filepath.Base(phasePath))).
			Value(&name).
			Validate(notEmpty),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Position selection
	defAfter := nextSequenceAfter(ctx, phasePath)
	var pos string
	var afterStr string

	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().
			Title("Position").
			Options(
				huh.NewOption("Append at end", "append"),
				huh.NewOption("Insert after number", "insert"),
			).Value(&pos),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	if pos == "insert" {
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

	after := atoiDefault(afterStr, 0)

	// Collect additional template variables
	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot,
		[]string{filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "sequence_number" || k == "sequence_name" {
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

	opts := &shared.CreateSequenceOpts{
		After:    after,
		Name:     name,
		Path:     phasePath,
		VarsFile: varsFile,
	}
	return shared.RunCreateSequence(ctx, opts)
}

// charmCreateSequenceManual is the fallback when hierarchical selector can't initialize.
func charmCreateSequenceManual(ctx context.Context) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, path, afterStr string

	// Get festival directory and list phases
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
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
		if selected == "__other__" {
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Phase (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
				if uitheme.IsCancelled(err) {
					return nil
				}
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
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
	}

	rp, rerr := resolvePhaseDirInput(path, cwd)
	if rerr != nil {
		return rerr
	}

	defAfter := nextSequenceAfter(ctx, rp)
	var pos string
	if err := huh.NewForm(huh.NewGroup(
		huh.NewSelect[string]().Title("Position").Options(
			huh.NewOption("Append at end", "append"),
			huh.NewOption("Insert after number", "insert"),
		).Value(&pos),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}
	if pos == "insert" {
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

	after := atoiDefault(afterStr, 0)

	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot, []string{filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "sequence_number" || k == "sequence_name" {
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
	opts := &shared.CreateSequenceOpts{After: after, Name: name, Path: fallbackDot(rp), VarsFile: varsFile}
	return shared.RunCreateSequence(ctx, opts)
}
