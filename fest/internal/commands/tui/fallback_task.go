//go:build no_charm

package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

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
	required := uniqueStrings(collectRequiredVars(context.TODO(), tmplRoot, []string{
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
	return shared.RunCreateTask(context.TODO(), opts)
}
