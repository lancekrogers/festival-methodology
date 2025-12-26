//go:build !no_charm

package tui

import (
	"context"
	"os"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
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

// NewTUICommand (charm version) provides a richer interactive UI using Charm libs
func NewTUICommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tui",
		Short: "Interactive UI (Charm) for festival creation and editing",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCharmTUI(cmd.Context())
		},
	}
	return cmd
}

func runCharmTUI(ctx context.Context) error {
	// Validate inside festivals workspace; if absent, offer to init
	cwd, _ := os.Getwd()
	if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
		var initNow bool
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().Title("No festivals/ directory detected. Initialize here?").Value(&initNow),
			),
		).WithTheme(theme())
		if err := form.Run(); err != nil {
			return err
		}
		if initNow {
			if err := shared.RunInit(ctx, ".", &shared.InitOpts{}); err != nil {
				return err
			}
		} else {
			return errors.NotFound("festivals directory")
		}
	}

	// main menu loop
	for {
		var action string
		menu := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("What would you like to do?").
					Options(
						huh.NewOption("Plan a New Festival (wizard)", "plan_festival"),
						huh.NewOption("Create a Festival (quick)", "create_festival"),
						huh.NewOption("Add a Phase", "create_phase"),
						huh.NewOption("Add a Sequence", "create_sequence"),
						huh.NewOption("Add a Task", "create_task"),
						huh.NewOption("Generate Festival Goal", "festival_goal"),
						huh.NewOption("Quit", "quit"),
					).
					Value(&action),
			),
		).WithTheme(huh.ThemeBase())

		if err := menu.Run(); err != nil {
			return err
		}

		switch action {
		case "plan_festival":
			if err := charmPlanFestivalWizard(); err != nil {
				return err
			}
		case "create_festival":
			if err := charmCreateFestival(); err != nil {
				return err
			}
		case "create_phase":
			if err := charmCreatePhase(); err != nil {
				return err
			}
		case "create_sequence":
			if err := charmCreateSequence(); err != nil {
				return err
			}
		case "create_task":
			if err := charmCreateTask(); err != nil {
				return err
			}
		case "festival_goal":
			if err := charmGenerateFestivalGoal(ctx); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// notEmpty validates that a string is not empty
func notEmpty(s string) error {
	if strings.TrimSpace(s) == "" {
		return errors.Validation("value required")
	}
	return nil
}

// toOptions converts a slice of strings to huh options
func toOptions(values []string) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(values))
	for _, v := range values {
		opts = append(opts, huh.NewOption(v, v))
	}
	return opts
}

// theme returns the custom theme for TUI forms
func theme() *huh.Theme {
	th := huh.ThemeBase()
	th.Focused.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)
	th.Focused.Description = lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	th.Blurred.Title = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	return th
}

// fallbackDot returns "." if string is empty, otherwise returns the string
func fallbackDot(s string) string {
	if strings.TrimSpace(s) == "" {
		return "."
	}
	return s
}

// StartCreateTUI shows a create-only menu (Charm implementation)
func StartCreateTUI() error {
	for {
		var action string
		menu := huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[string]().
					Title("Create what?").
					Options(
						huh.NewOption("Festival", "festival"),
						huh.NewOption("Phase", "phase"),
						huh.NewOption("Sequence", "sequence"),
						huh.NewOption("Task", "task"),
						huh.NewOption("Back", "back"),
					).
					Value(&action),
			),
		).WithTheme(theme())
		if err := menu.Run(); err != nil {
			return err
		}
		switch action {
		case "festival":
			if err := charmCreateFestival(); err != nil {
				return err
			}
		case "phase":
			if err := charmCreatePhase(); err != nil {
				return err
			}
		case "sequence":
			if err := charmCreateSequence(); err != nil {
				return err
			}
		case "task":
			if err := charmCreateTask(); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

// StartCreateFestivalTUI starts the festival creation TUI directly
func StartCreateFestivalTUI() error { return charmCreateFestival() }

// StartCreatePhaseTUI starts the phase creation TUI directly
func StartCreatePhaseTUI() error { return charmCreatePhase() }

// StartCreateSequenceTUI starts the sequence creation TUI directly
func StartCreateSequenceTUI() error { return charmCreateSequence() }

// StartCreateTaskTUI starts the task creation TUI directly
func StartCreateTaskTUI() error { return charmCreateTask() }
