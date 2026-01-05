package navigation

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/spf13/cobra"
)

var shortcutNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{1,20}$`)

// Styles for fest go list output
var (
	// Shortcut styling - pink/magenta for user shortcuts
	shortcutNameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Bold(true)

	// Festival link styling - green for festival links
	festivalLinkStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)

	// Path styling - gray for file paths
	pathDisplayStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// Header styling - bold for section headers
	headerStyle = lipgloss.NewStyle().Bold(true)
)

// isValidShortcutName checks if a shortcut name is valid
func isValidShortcutName(name string) bool {
	return shortcutNameRegex.MatchString(name)
}

// NewGoMapCommand creates the go map subcommand
func NewGoMapCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "map <name> [path]",
		Short: "Create a navigation shortcut",
		Long: `Create a shortcut for quick navigation using fgo.

If path is omitted, the current directory is used.
Shortcut names must be alphanumeric (with underscores), 1-20 characters.

Usage with fgo:
  fgo -<name>    Navigate to the shortcut`,
		Example: `  fest go map n                   # Create shortcut 'n' to current directory
  fest go map api /path/to/api    # Create shortcut 'api' to specific path

  # Then navigate with:
  fgo -n      # Navigate to notes
  fgo -api    # Navigate to API directory`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			path := ""
			if len(args) > 1 {
				path = args[1]
			}
			return runGoMap(name, path, jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func runGoMap(name, path string, jsonOutput bool) error {
	// Validate shortcut name
	if !isValidShortcutName(name) {
		return errors.Validation("invalid shortcut name").
			WithField("name", name).
			WithField("hint", "use 1-20 alphanumeric characters or underscores")
	}

	// Resolve path
	var absPath string
	var err error
	if path == "" {
		absPath, err = os.Getwd()
		if err != nil {
			return errors.IO("getting current directory", err)
		}
	} else {
		absPath, err = filepath.Abs(path)
		if err != nil {
			return errors.IO("resolving path", err).WithField("path", path)
		}
	}

	// Validate path exists and is a directory
	info, err := os.Stat(absPath)
	if os.IsNotExist(err) {
		return errors.NotFound("path").WithField("path", absPath)
	}
	if err != nil {
		return errors.IO("checking path", err).WithField("path", absPath)
	}
	if !info.IsDir() {
		return errors.Validation("path must be a directory").WithField("path", absPath)
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	// Create shortcut
	nav.Shortcuts[name] = absPath
	nav.UpdatedAt = time.Now().UTC()

	// Save navigation state
	if err := nav.Save(); err != nil {
		return errors.Wrap(err, "saving navigation state")
	}

	// Output result
	if jsonOutput {
		result := map[string]interface{}{
			"success":  true,
			"shortcut": name,
			"path":     absPath,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Created shortcut '-%s' → %s\n", name, absPath)
		fmt.Println("\nUse 'fgo -" + name + "' to navigate")
	}

	return nil
}

// NewGoUnmapCommand creates the go unmap subcommand
func NewGoUnmapCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "unmap <name>",
		Short: "Remove a navigation shortcut",
		Long:  `Remove a previously created navigation shortcut.`,
		Example: `  fest go unmap n     # Remove shortcut 'n'
  fest go unmap api   # Remove shortcut 'api'`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoUnmap(args[0], jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func runGoUnmap(name string, jsonOutput bool) error {
	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	// Check if shortcut exists
	_, exists := nav.Shortcuts[name]

	if exists {
		delete(nav.Shortcuts, name)
		nav.UpdatedAt = time.Now().UTC()

		// Save navigation state
		if err := nav.Save(); err != nil {
			return errors.Wrap(err, "saving navigation state")
		}
	}

	// Output result
	if jsonOutput {
		result := map[string]interface{}{
			"success":  true,
			"shortcut": name,
			"removed":  exists,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		if exists {
			fmt.Printf("Removed shortcut '-%s'\n", name)
		} else {
			fmt.Printf("Shortcut '%s' not found\n", name)
		}
	}

	return nil
}

// NewGoListCommand creates the go list subcommand
func NewGoListCommand() *cobra.Command {
	var jsonOutput bool
	var interactive bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List navigation shortcuts and links",
		Long: `Display all navigation shortcuts and festival-project links.

SHORTCUTS are user-defined with 'fest go map'.
LINKS are festival-project associations created with 'fest link'.

Use --interactive (-i) to select a destination with an interactive picker.
When used with shell integration (fgo list), this will navigate to the selected path.`,
		Example: `  fest go list           # List all shortcuts and links
  fest go list --json    # Output in JSON format
  fest go list -i        # Interactive picker (with fgo: navigates to selection)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if interactive {
				return runGoListInteractive(cmd.Context())
			}
			return runGoList(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	cmd.Flags().BoolVarP(&interactive, "interactive", "i", false, "interactive picker mode")

	return cmd
}

// runGoListInteractive runs the TUI selector and outputs the selected path
func runGoListInteractive(ctx context.Context) error {
	if shared.StartGoListTUI == nil {
		return errors.Validation("TUI not available - build with charm support or use non-interactive mode")
	}

	selected, err := shared.StartGoListTUI(ctx)
	if err != nil {
		return err
	}

	// Output the selected path for shell wrapper to cd to
	if selected != "" {
		fmt.Println(selected)
	}
	return nil
}

func runGoList(jsonOutput bool) error {
	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	if jsonOutput {
		type shortcutInfo struct {
			Name string `json:"name"`
			Path string `json:"path"`
		}
		type linkInfo struct {
			Festival string `json:"festival"`
			Project  string `json:"project"`
			LinkedAt string `json:"linked_at"`
		}

		var shortcuts []shortcutInfo
		for name, path := range nav.Shortcuts {
			shortcuts = append(shortcuts, shortcutInfo{
				Name: name,
				Path: path,
			})
		}

		var links []linkInfo
		for name, link := range nav.Links {
			links = append(links, linkInfo{
				Festival: name,
				Project:  link.Path,
				LinkedAt: link.LinkedAt.Format(time.RFC3339),
			})
		}

		result := map[string]interface{}{
			"shortcuts": shortcuts,
			"links":     links,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		hasContent := false

		// Print shortcuts section
		if len(nav.Shortcuts) > 0 {
			fmt.Println(headerStyle.Render("SHORTCUTS"))
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("%-12s %s\n", "Shortcut", "Path")
			fmt.Println(strings.Repeat("-", 60))
			for name, path := range nav.Shortcuts {
				styledName := shortcutNameStyle.Render("-" + name)
				styledPath := pathDisplayStyle.Render(path)
				fmt.Printf("%-21s %s\n", styledName, styledPath)
			}
			hasContent = true
		}

		// Print links section
		if len(nav.Links) > 0 {
			if hasContent {
				fmt.Println()
			}
			fmt.Println(headerStyle.Render("FESTIVAL LINKS"))
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("%-25s %s\n", "Festival", "Project Path")
			fmt.Println(strings.Repeat("-", 60))
			for name, link := range nav.Links {
				styledFestival := festivalLinkStyle.Render(name)
				styledPath := pathDisplayStyle.Render(link.Path)
				fmt.Printf("%-34s %s\n", styledFestival, styledPath)
			}
			hasContent = true
		}

		if !hasContent {
			fmt.Println("No shortcuts or festival links configured.")
			fmt.Println()
			fmt.Println("Create a shortcut:    fest go map <name> [path]")
			fmt.Println("Create a link:        fest link <path>  (from within a festival)")
		}
	}

	return nil
}

// NewGoShortcutCommand creates the internal shortcut lookup command
func NewGoShortcutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "shortcut <name>",
		Short:  "Internal: lookup shortcut path",
		Hidden: true,
		Args:   cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoShortcut(args[0])
		},
	}

	return cmd
}

func runGoShortcut(name string) error {
	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	path, exists := nav.Shortcuts[name]
	if !exists {
		return errors.NotFound("shortcut").WithField("name", name)
	}

	// Just print the path for shell function to use
	fmt.Println(path)
	return nil
}

// NewGoProjectCommand creates the go project subcommand
func NewGoProjectCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Navigate to linked project directory",
		Long: `Navigate to the project directory linked to the current festival.

Use 'fest link <path>' to create a link from within a festival.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoProject()
		},
	}

	return cmd
}

func runGoProject() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current festival
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return errors.NotFound("festival").
			WithField("hint", "run from inside a festival directory")
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	link, found := nav.GetLink(loc.Festival.Name)
	if !found {
		return errors.NotFound("project link").
			WithField("festival", loc.Festival.Name).
			WithField("hint", "use 'fest link <path>' to create a link")
	}

	// Print the path for shell function to use
	fmt.Println(link.Path)
	return nil
}

// NewGoFestCommand creates the go fest subcommand
func NewGoFestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fest",
		Short: "Navigate back to festival from linked project",
		Long: `Navigate back to the festival that is linked to the current project directory.

This is the reverse of 'fgo project'.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoFest()
		},
	}

	return cmd
}

func runGoFest() error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	// Check if current directory is in any linked project
	for festivalName, link := range nav.Links {
		// Check if cwd is within the linked project path
		if strings.HasPrefix(cwd, link.Path) || cwd == link.Path {
			// Use stored festival path if available (preferred)
			if link.FestivalPath != "" {
				// Verify the path still exists
				if info, err := os.Stat(link.FestivalPath); err == nil && info.IsDir() {
					fmt.Println(link.FestivalPath)
					return nil
				}
				// Fall through to search if stored path is invalid
			}

			// Fall back to searching for the festival by name
			festivalsPath, err := findFestivalPath(festivalName)
			if err != nil {
				return err
			}
			fmt.Println(festivalsPath)
			return nil
		}
	}

	return errors.NotFound("festival link").
		WithField("hint", "current directory is not in a linked project")
}

// findFestivalPath searches for a festival by name in the festivals directory
func findFestivalPath(festivalName string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting current directory", err)
	}

	// Find festivals directory
	festivalsDir, err := findFestivalsDir(cwd)
	if err != nil {
		return "", err
	}

	// Search in status directories
	statusDirs := []string{"active", "planned", "completed", "dungeon"}
	for _, status := range statusDirs {
		statusPath := filepath.Join(festivalsDir, status)
		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && entry.Name() == festivalName {
				return filepath.Join(statusPath, entry.Name()), nil
			}
		}
	}

	return "", errors.NotFound("festival").WithField("name", festivalName)
}

// findFestivalsDir walks up from the given path to find a festivals directory.
// A valid festivals directory must contain a .festival/ subdirectory with a .workspace file.
// This distinguishes actual working festivals from template libraries (which lack .workspace).
func findFestivalsDir(startPath string) (string, error) {
	dir := startPath
	for {
		festivalsPath := filepath.Join(dir, "festivals")
		if info, err := os.Stat(festivalsPath); err == nil && info.IsDir() {
			// Check for .festival subdirectory
			dotFestivalPath := filepath.Join(festivalsPath, ".festival")
			if info, err := os.Stat(dotFestivalPath); err == nil && info.IsDir() {
				// Check for .workspace file to confirm this is a working festivals directory
				// (not a template library which lacks .workspace)
				workspacePath := filepath.Join(dotFestivalPath, ".workspace")
				if _, err := os.Stat(workspacePath); err == nil {
					return festivalsPath, nil
				}
			}
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", errors.NotFound("festivals directory")
}
