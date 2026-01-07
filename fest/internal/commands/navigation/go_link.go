package navigation

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	festErrors "github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// Styles for festival status in TUI
var (
	activeStyle  = lipgloss.NewStyle().Foreground(ui.ActiveColor).Bold(true)
	plannedStyle = lipgloss.NewStyle().Foreground(ui.PlannedColor).Bold(true)
	pathStyle    = lipgloss.NewStyle().Foreground(ui.MetadataColor)
)

// NewGoLinkCommand creates the context-aware link subcommand for fest go
func NewGoLinkCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "link [path]",
		Short: "Link current festival to a project directory (or vice versa)",
		Long: `Create a bidirectional link between a festival and a project directory.

When run inside a festival:
  - Links the festival to the specified project directory
  - If no path provided, prompts for directory input

When run inside a project directory (non-festival):
  - Shows an interactive picker to select a festival to link
  - Links the current project to the selected festival

After linking:
  - 'fgo' in the festival jumps to the project
  - 'fgo' in the project jumps to the festival

Examples:
  # Inside a festival, link to project:
  fgo link /path/to/project
  fgo link .                    # Link to current directory (if not in festival)

  # Inside a project, show festival picker:
  fgo link`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			path := ""
			if len(args) > 0 {
				path = args[0]
			}
			return runGoLink(cmd.Context(), path)
		},
	}

	return cmd
}

func runGoLink(ctx context.Context, targetPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return festErrors.IO("getting current directory", err)
	}

	// Detect context: are we inside a festival?
	if isInsideFestival(cwd) {
		return linkFestivalToProject(ctx, cwd, targetPath)
	}

	// We're in a project directory, show festival picker
	return linkProjectToFestival(cwd)
}

// isInsideFestival checks if the path is within a festivals/ directory
func isInsideFestival(path string) bool {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	// Walk up to find festivals/ directory
	current := absPath
	for {
		base := filepath.Base(current)
		if base == "festivals" {
			return true
		}

		// Check if parent contains festivals/ (we're inside it)
		parent := filepath.Dir(current)
		if parent == current {
			break
		}

		// Check if we just passed through a festivals directory
		parentBase := filepath.Base(parent)
		if parentBase == "festivals" {
			return true
		}

		current = parent
	}

	return false
}

// linkFestivalToProject links the current festival to a project directory
func linkFestivalToProject(ctx context.Context, cwd, targetPath string) error {
	// Detect current festival
	loc, err := show.DetectCurrentLocation(ctx, cwd)
	if err != nil || loc == nil || loc.Festival == nil || loc.Festival.Name == "" {
		return festErrors.Validation("not inside a recognized festival")
	}
	festivalName := loc.Festival.Name

	var projectPath string

	if targetPath != "" {
		// Use provided path
		absPath, err := filepath.Abs(targetPath)
		if err != nil {
			return festErrors.Wrap(err, "resolving path").WithField("path", targetPath)
		}

		// Validate path exists and is a directory
		info, err := os.Stat(absPath)
		if err != nil {
			return festErrors.NotFound("directory").WithField("path", absPath)
		}
		if !info.IsDir() {
			return festErrors.Validation("path is not a directory").WithField("path", absPath)
		}

		projectPath = absPath
	} else {
		// Show interactive directory picker
		selectedPath, err := selectProjectDirectory(cwd, festivalName)
		if err != nil {
			// Silent exit on user cancel (Ctrl-C or Esc)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		projectPath = selectedPath
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return festErrors.Wrap(err, "loading navigation state")
	}

	// Get the festival path for reverse navigation
	festivalPath := loc.Festival.Path

	// Set the bidirectional link with festival path
	nav.SetLinkWithPath(festivalName, projectPath, festivalPath)

	// Save
	if err := nav.Save(); err != nil {
		return festErrors.Wrap(err, "saving navigation state")
	}

	fmt.Println(ui.H1("Link Created"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(festivalName, ui.FestivalColor))
	fmt.Printf("%s %s\n", ui.Label("Project"), ui.Dim(projectPath))
	fmt.Println()
	fmt.Println(ui.Dim("Use 'fgo' to navigate between them."))

	return nil
}

// linkProjectToFestival shows a picker to select a festival to link
func linkProjectToFestival(cwd string) error {
	absPath, err := filepath.Abs(cwd)
	if err != nil {
		return festErrors.Wrap(err, "resolving current directory")
	}

	// Find festivals directory
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil || festivalsDir == "" {
		return festErrors.NotFound("festivals directory").WithField("hint", "run 'fest init' first")
	}

	// Collect all festivals from active/ and planned/
	festivals, err := collectFestivals(festivalsDir)
	if err != nil {
		return festErrors.Wrap(err, "collecting festivals")
	}

	if len(festivals) == 0 {
		return festErrors.NotFound("festivals").WithField("hint", "create a festival with 'fest create festival'")
	}

	// Build options for picker with color-coded status
	options := make([]huh.Option[string], 0, len(festivals))

	for _, f := range festivals {
		var label string
		if f.status == "active" {
			label = activeStyle.Render("● "+f.name) + " " + pathStyle.Render("(active)")
		} else {
			label = plannedStyle.Render("○ "+f.name) + " " + pathStyle.Render("(planned)")
		}
		options = append(options, huh.NewOption(label, f.name))
	}

	// Show picker
	var selectedFestival string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a festival to link to this project").
				Description(pathStyle.Render(absPath)).
				Options(options...).
				Value(&selectedFestival),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		// Silent exit on user cancel (Ctrl-C or Esc)
		if errors.Is(err, huh.ErrUserAborted) {
			return nil
		}
		return err
	}

	if selectedFestival == "" {
		// User cancelled without selecting
		return nil
	}

	// Find the festival path from the selected festival
	var festivalPath string
	for _, f := range festivals {
		if f.name == selectedFestival {
			festivalPath = f.path
			break
		}
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return festErrors.Wrap(err, "loading navigation state")
	}

	// Set the bidirectional link with festival path
	nav.SetLinkWithPath(selectedFestival, absPath, festivalPath)

	// Save
	if err := nav.Save(); err != nil {
		return festErrors.Wrap(err, "saving navigation state")
	}

	fmt.Println(ui.H1("Link Created"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(selectedFestival, ui.FestivalColor))
	fmt.Printf("%s %s\n", ui.Label("Project"), ui.Dim(absPath))
	fmt.Println()
	fmt.Println(ui.Dim("Use 'fgo' to navigate between them."))

	return nil
}

type festivalInfo struct {
	name        string
	displayName string
	status      string
	path        string
}

// collectFestivals finds all festivals in active/ and planned/ directories
func collectFestivals(festivalsDir string) ([]festivalInfo, error) {
	var festivals []festivalInfo

	statusDirs := []string{"active", "planned"}

	for _, status := range statusDirs {
		statusPath := filepath.Join(festivalsDir, status)
		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			name := entry.Name()
			// Skip hidden directories
			if strings.HasPrefix(name, ".") {
				continue
			}

			festivals = append(festivals, festivalInfo{
				name:        name,
				displayName: fmt.Sprintf("%s (%s)", name, status),
				status:      status,
				path:        filepath.Join(statusPath, name),
			})
		}
	}

	return festivals, nil
}

// resolveFestivalPath finds the full path of a festival by name
func resolveFestivalPath(festivalsDir, festivalName string) string {
	statusDirs := []string{"active", "planned", "completed"}

	for _, status := range statusDirs {
		statusPath := filepath.Join(festivalsDir, status)
		festPath := filepath.Join(statusPath, festivalName)
		if info, err := os.Stat(festPath); err == nil && info.IsDir() {
			return festPath
		}
	}

	return ""
}

// selectProjectDirectory shows an interactive directory picker for selecting a project directory.
// It starts from the campaign root (parent of festivals/ directory) and allows selecting
// any directory except the festivals/ directory itself.
func selectProjectDirectory(festivalPath, festivalName string) (string, error) {
	// Find campaign root by walking up from festival path
	// festivalPath is like: /path/to/guild-framework/festivals/active/festival-name
	// We want campaign root: /path/to/guild-framework
	festivalsDir := filepath.Dir(filepath.Dir(festivalPath)) // → festivals/
	campaignRoot := filepath.Dir(festivalsDir)               // → guild-framework/

	// Collect directories from campaign root (excluding festivals/)
	directories, err := collectDirectories(campaignRoot, festivalsDir)
	if err != nil {
		return "", festErrors.Wrap(err, "collecting directories")
	}

	if len(directories) == 0 {
		return "", festErrors.NotFound("project directories").
			WithField("hint", "no suitable directories found in campaign root")
	}

	// Build options for picker
	options := make([]huh.Option[string], 0, len(directories))
	for _, dir := range directories {
		// Show relative path from campaign root for cleaner display
		relPath, _ := filepath.Rel(campaignRoot, dir)
		label := pathStyle.Render(relPath)
		options = append(options, huh.NewOption(label, dir))
	}

	// Show picker
	var selectedDir string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Select project directory to link to '%s'", festivalName)).
				Description(pathStyle.Render("Campaign: " + campaignRoot)).
				Options(options...).
				Value(&selectedDir),
		),
	).WithTheme(huh.ThemeCharm())

	if err := form.Run(); err != nil {
		return "", err
	}

	if selectedDir == "" {
		return "", festErrors.Validation("no directory selected")
	}

	return selectedDir, nil
}

// collectDirectories recursively collects directories from root, excluding excludePath
func collectDirectories(root, excludePath string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		fullPath := filepath.Join(root, entry.Name())

		// Skip the festivals directory
		if fullPath == excludePath {
			continue
		}

		// Add this directory
		dirs = append(dirs, fullPath)

		// Recursively add subdirectories (one level deep for performance)
		subEntries, err := os.ReadDir(fullPath)
		if err != nil {
			continue
		}
		for _, subEntry := range subEntries {
			if !subEntry.IsDir() {
				continue
			}
			if strings.HasPrefix(subEntry.Name(), ".") {
				continue
			}
			subPath := filepath.Join(fullPath, subEntry.Name())
			// Skip festivals subdirectory if somehow inside
			if strings.Contains(subPath, "festivals") {
				continue
			}
			dirs = append(dirs, subPath)
		}
	}

	return dirs, nil
}
