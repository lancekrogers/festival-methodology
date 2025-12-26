//go:build no_charm

package tui

import (
	"context"
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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

// NewTUICommand launches an interactive text UI for common actions
func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive UI for creating festivals, phases, sequences, and tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTUI(cmd.Context())
		},
	}
	return cmd
}

func runTUI(ctx context.Context) error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	cwd, _ := os.Getwd()

	// Ensure we are inside a festivals workspace; if not, offer to init
	if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
		display.Warning("No festivals/ directory detected.")
		if display.Confirm("Initialize a new festival workspace here?") {
			if err := shared.RunInit(ctx, ".", &shared.InitOpts{}); err != nil {
				return err
			}
		} else {
			return fmt.Errorf("no festivals/ directory detected")
		}
	}

	for {
		choice := display.Choose("What would you like to do?", []string{
			"Plan a New Festival (wizard)",
			"Create a Festival (quick)",
			"Add a Phase",
			"Add a Sequence",
			"Add a Task",
			"Generate Festival Goal",
			"Quit",
		})

		switch choice {
		case 0:
			if err := tuiPlanFestivalWizard(display); err != nil {
				return err
			}
		case 1:
			if err := tuiCreateFestival(display); err != nil {
				return err
			}
		case 2:
			if err := tuiCreatePhase(display); err != nil {
				return err
			}
		case 3:
			if err := tuiCreateSequence(display); err != nil {
				return err
			}
		case 4:
			if err := tuiCreateTask(display); err != nil {
				return err
			}
		case 5:
			if err := tuiGenerateFestivalGoal(ctx, display); err != nil {
				return err
			}
		default:
			return nil
		}

		if !display.Confirm("Do you want to perform another action?") {
			break
		}
	}
	return nil
}

// StartCreateTUI shows a create-only menu (fallback implementation)
func StartCreateTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	for {
		choice := display.Choose("Create what?", []string{
			"Festival",
			"Phase",
			"Sequence",
			"Task",
			"Back",
		})
		switch choice {
		case 0:
			if err := tuiCreateFestival(display); err != nil {
				return err
			}
		case 1:
			if err := tuiCreatePhase(display); err != nil {
				return err
			}
		case 2:
			if err := tuiCreateSequence(display); err != nil {
				return err
			}
		case 3:
			if err := tuiCreateTask(display); err != nil {
				return err
			}
		default:
			return nil
		}
		if !display.Confirm("Create another?") {
			return nil
		}
	}
}

// StartCreateFestivalTUI starts the festival creation TUI directly
func StartCreateFestivalTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateFestival(display)
}

// StartCreatePhaseTUI starts the phase creation TUI directly
func StartCreatePhaseTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreatePhase(display)
}

// StartCreateSequenceTUI starts the sequence creation TUI directly
func StartCreateSequenceTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateSequence(display)
}

// StartCreateTaskTUI starts the task creation TUI directly
func StartCreateTaskTUI() error {
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	return tuiCreateTask(display)
}
