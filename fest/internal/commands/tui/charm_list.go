//go:build !no_charm

package tui

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
)

func init() {
	// Register go list TUI hook with the shared package
	shared.StartGoListTUI = StartGoListTUI
}

// StartGoListTUI launches an interactive selector for navigation shortcuts and links.
// Returns the selected path or an error if cancelled/failed.
func StartGoListTUI(ctx context.Context) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return "", errors.Wrap(err, "loading navigation state")
	}

	// Build options list
	opts := make([]huh.Option[string], 0)

	// Add shortcuts (sorted by name for consistency)
	shortcutNames := make([]string, 0, len(nav.Shortcuts))
	for name := range nav.Shortcuts {
		shortcutNames = append(shortcutNames, name)
	}
	sort.Strings(shortcutNames)

	for _, name := range shortcutNames {
		path := nav.Shortcuts[name]
		label := fmt.Sprintf("-%-12s → %s", name, path)
		opts = append(opts, huh.NewOption(label, path))
	}

	// Add festival links (sorted by name for consistency)
	linkNames := make([]string, 0, len(nav.Links))
	for name := range nav.Links {
		linkNames = append(linkNames, name)
	}
	sort.Strings(linkNames)

	for _, festivalName := range linkNames {
		link := nav.Links[festivalName]
		label := fmt.Sprintf("%-13s → %s", festivalName, link.Path)
		opts = append(opts, huh.NewOption(label, link.Path))
	}

	// Check if we have any options
	if len(opts) == 0 {
		return "", errors.Validation("no shortcuts or links configured").
			WithField("hint", "use 'fest go map <name>' to create shortcuts or 'fest link <path>' to link festivals")
	}

	// Run the selector
	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select destination").
				Description("Use arrow keys to navigate, Enter to select, Esc to cancel").
				Options(opts...).
				Value(&selected),
		),
	).WithTheme(theme())

	if err := form.Run(); err != nil {
		// User cancelled (Ctrl+C or Esc)
		return "", err
	}

	return selected, nil
}
