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

func charmCreateTask(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Use hierarchical selector to select festival, phase, and sequence
	selector, err := NewHierarchySelectorFromCwd(SelectToSequence(true))
	if err != nil {
		// Fallback to manual path entry
		return charmCreateTaskManual(ctx)
	}

	state, err := selector.Run(ctx)
	if err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Continue with task creation in selected sequence
	return createTaskInSequence(ctx, state.SequencePath)
}

// createTaskInSequence creates a task in the specified sequence.
func createTaskInSequence(ctx context.Context, sequencePath string) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}

	var name string

	// Task name form
	if err := huh.NewForm(huh.NewGroup(
		huh.NewInput().
			Title("Task name").
			Placeholder("implement_feature").
			Description(fmt.Sprintf("Creating in: %s", filepath.Base(sequencePath))).
			Value(&name).
			Validate(notEmpty),
	)).WithTheme(theme()).Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return nil
		}
		return err
	}

	// Position selection
	defAfter := nextTaskAfter(ctx, sequencePath)
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
		[]string{filepath.Join(tmplRoot, "TASK_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "task_number" || k == "task_name" {
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

	opts := &shared.CreateTaskOpts{
		After:    after,
		Names:    []string{name},
		Path:     sequencePath,
		VarsFile: varsFile,
	}
	return shared.RunCreateTask(ctx, opts)
}

// charmCreateTaskManual is the fallback when hierarchical selector can't initialize.
func charmCreateTaskManual(ctx context.Context) error {
	cwd, _ := os.Getwd()
	tmplRoot, err := tpl.LocalTemplateRoot(cwd)
	if err != nil {
		return err
	}
	var name, path, afterStr string

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
				if uitheme.IsCancelled(err) {
					return nil
				}
				return err
			}
			if pSel == "__other__" {
				if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Phase (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
					if uitheme.IsCancelled(err) {
						return nil
					}
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
				if uitheme.IsCancelled(err) {
					return nil
				}
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
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
		if sSel == "__other__" {
			if err := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Sequence (dir or number)").Value(&path))).WithTheme(theme()).Run(); err != nil {
				if uitheme.IsCancelled(err) {
					return nil
				}
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
			if uitheme.IsCancelled(err) {
				return nil
			}
			return err
		}
	}

	rs, rerr := resolveSequenceDirInput(path, cwd)
	if rerr != nil {
		return rerr
	}
	defAfter := nextTaskAfter(ctx, rs)
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

	required := uniqueStrings(collectRequiredVars(ctx, tmplRoot, []string{filepath.Join(tmplRoot, "TASK_TEMPLATE.md")}))
	vars := map[string]interface{}{}
	for _, k := range required {
		if k == "task_number" || k == "task_name" {
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
	opts := &shared.CreateTaskOpts{After: after, Names: []string{name}, Path: fallbackDot(rs), VarsFile: varsFile}
	return shared.RunCreateTask(ctx, opts)
}
