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

	required := uniqueStrings(collectRequiredVars(context.TODO(), tmplRoot, []string{
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
	return shared.RunCreateSequence(context.TODO(), opts)
}
