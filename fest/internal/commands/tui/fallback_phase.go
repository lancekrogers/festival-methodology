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
	types := []string{"planning", "implementation", "review", "deployment", "research"}
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

	required := uniqueStrings(collectRequiredVars(context.TODO(), tmplRoot, []string{
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
	return shared.RunCreatePhase(context.TODO(), opts)
}
