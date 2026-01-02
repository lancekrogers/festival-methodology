//go:build !no_charm

package tui

import (
	"context"
	"fmt"
	"sort"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	festErrors "github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	uitheme "github.com/lancekrogers/festival-methodology/fest/internal/ui/theme"
)

// Styles for list items
var (
	shortcutStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true) // Pink
	linkStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)  // Green
	pathStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))            // Gray
)

func init() {
	// Register go list TUI hook with the shared package
	shared.StartGoListTUI = StartGoListTUI
}

// StartGoListTUI launches an interactive selector for navigation shortcuts and links.
// Returns the selected path or empty string if cancelled.
func StartGoListTUI(ctx context.Context) (string, error) {
	// Check context
	if err := ctx.Err(); err != nil {
		return "", err
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return "", festErrors.Wrap(err, "loading navigation state")
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
		label := shortcutStyle.Render(fmt.Sprintf("-%-12s", name)) + " → " + pathStyle.Render(path)
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
		label := linkStyle.Render(fmt.Sprintf("%-13s", festivalName)) + " → " + pathStyle.Render(link.Path)
		opts = append(opts, huh.NewOption(label, link.Path))
	}

	// Check if we have any options
	if len(opts) == 0 {
		return "", festErrors.Validation("no shortcuts or links configured").
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
	).WithTheme(uitheme.FestTheme())

	if err := form.Run(); err != nil {
		// Silent exit on user cancel (Ctrl-C or Esc)
		if uitheme.IsCancelled(err) {
			return "", nil
		}
		return "", err
	}

	return selected, nil
}
