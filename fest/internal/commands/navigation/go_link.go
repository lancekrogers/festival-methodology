package navigation

import (
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
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// Styles for festival status in TUI
var (
	activeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true) // Green
	plannedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Bold(true) // Blue
	pathStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))           // Gray
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
			return runGoLink(path)
		},
	}

	return cmd
}

func runGoLink(targetPath string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return festErrors.IO("getting current directory", err)
	}

	// Detect context: are we inside a festival?
	if isInsideFestival(cwd) {
		return linkFestivalToProject(cwd, targetPath)
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
func linkFestivalToProject(cwd, targetPath string) error {
	// Detect current festival
	loc, err := show.DetectCurrentLocation(cwd)
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
		// Prompt for directory
		var inputPath string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title(fmt.Sprintf("Link festival '%s' to which directory?", festivalName)).
					Description("Enter an absolute path to the project directory").
					Placeholder("/path/to/your/project").
					Value(&inputPath).
					Validate(func(s string) error {
						if strings.TrimSpace(s) == "" {
							return festErrors.Validation("path is required")
						}
						return nil
					}),
			),
		).WithTheme(huh.ThemeCharm())

		if err := form.Run(); err != nil {
			// Silent exit on user cancel (Ctrl-C or Esc)
			if errors.Is(err, huh.ErrUserAborted) {
				return nil
			}
			return err
		}

		absPath, err := filepath.Abs(inputPath)
		if err != nil {
			return festErrors.Wrap(err, "resolving path").WithField("path", inputPath)
		}

		// Validate
		info, err := os.Stat(absPath)
		if err != nil {
			return festErrors.NotFound("directory").WithField("path", absPath)
		}
		if !info.IsDir() {
			return festErrors.Validation("path is not a directory").WithField("path", absPath)
		}

		projectPath = absPath
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

	fmt.Printf("Linked: %s ↔ %s\n", festivalName, projectPath)
	fmt.Println("Use 'fgo' to navigate between them.")

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

	fmt.Printf("Linked: %s ↔ %s\n", selectedFestival, absPath)
	fmt.Println("Use 'fgo' to navigate between them.")

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
