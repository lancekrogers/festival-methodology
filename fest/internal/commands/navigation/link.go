package navigation

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

type linkOptions struct {
	showLink bool
	json     bool
}

// NewLinkCommand creates the link command
func NewLinkCommand() *cobra.Command {
	opts := &linkOptions{}

	cmd := &cobra.Command{
		Use:   "link [path]",
		Short: "Link festival to project directory (context-aware)",
		Long: `Link a festival to a project directory with context-aware behavior.

When run inside a festival:
  - Links the festival to the specified project path
  - If no path provided, prompts for directory input

When run inside a project directory (non-festival):
  - Shows an interactive picker to select a festival to link
  - Links the current project to the selected festival

After linking, use 'fgo' to navigate between them.

The link is stored centrally in ~/.config/fest/navigation.yaml.
Use 'fest links' to see all festival-project links.
Use 'fest unlink' to remove the link for current festival.`,
		Example: `  fest link /path/to/project   # Inside festival: link to project
  fest link .                  # Inside festival: link to cwd
  fest link                    # Inside festival: prompt for path
  fest link                    # Inside project: show festival picker
  fest link --show             # Display current link`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.showLink {
				return runLinkShow(opts)
			}
			targetPath := ""
			if len(args) > 0 {
				targetPath = args[0]
			}
			return runLink(targetPath, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.showLink, "show", false, "show current link")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

func runLink(targetPath string, opts *linkOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// Detect context: are we inside a festival?
	loc, _ := show.DetectCurrentLocation(cwd)

	if loc == nil || loc.Festival == nil {
		// Not in a festival - show picker to select festival to link
		// Use targetPath as project directory, or cwd if not specified
		projectDir := cwd
		if targetPath != "" && targetPath != "." {
			absPath, err := filepath.Abs(targetPath)
			if err != nil {
				return errors.IO("resolving path", err).WithField("path", targetPath)
			}
			projectDir = absPath
		}
		return linkProjectToFestival(projectDir)
	}

	// Inside a festival - if no path provided, use the TUI prompt from go_link
	if targetPath == "" {
		return linkFestivalToProject(cwd, "")
	}

	// Resolve target path
	var absPath string
	if targetPath == "." {
		absPath = cwd
	} else {
		absPath, err = filepath.Abs(targetPath)
		if err != nil {
			return errors.IO("resolving path", err).WithField("path", targetPath)
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

	// Warn if path is inside festivals directory
	festivalsRoot := filepath.Dir(filepath.Dir(loc.Festival.Path))
	if strings.HasPrefix(absPath, festivalsRoot) {
		display.Warning("Linking to a path inside festivals directory may be unintended.")
		fmt.Printf("%s %s\n", ui.Label("Path"), ui.Dim(absPath))
		if !display.Confirm("Continue") {
			display.Info("Operation cancelled.")
			return nil
		}
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	// Create link with festival path for reverse navigation
	festivalName := loc.Festival.Name
	festivalPath := loc.Festival.Path
	nav.SetLinkWithPath(festivalName, absPath, festivalPath)

	// Save navigation state
	if err := nav.Save(); err != nil {
		return errors.Wrap(err, "saving navigation state")
	}

	// Output result
	if opts.json {
		result := map[string]interface{}{
			"success":   true,
			"festival":  festivalName,
			"project":   absPath,
			"linked_at": time.Now().UTC().Format(time.RFC3339),
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println(ui.H1("Festival Link"))
		fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(festivalName, ui.FestivalColor))
		fmt.Printf("%s %s\n", ui.Label("Project"), ui.Dim(absPath))
		fmt.Println()
		fmt.Println(ui.Dim("Use 'fgo project' to navigate to the project (after shell-init setup)"))
		fmt.Println(ui.Dim("Use 'fest link --show' to view this link"))
	}

	return nil
}

func runLinkShow(opts *linkOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current festival
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		if opts.json {
			result := map[string]interface{}{
				"error": "not in a festival directory",
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		return errors.Wrap(err, "detecting festival")
	}

	if loc.Festival == nil {
		if opts.json {
			result := map[string]interface{}{
				"error": "not in a festival directory",
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
			return nil
		}
		return errors.NotFound("festival")
	}

	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	festivalName := loc.Festival.Name
	link, found := nav.GetLink(festivalName)

	if opts.json {
		result := map[string]interface{}{
			"festival": festivalName,
			"linked":   found,
		}
		if found {
			result["project"] = link.Path
			result["linked_at"] = link.LinkedAt.Format(time.RFC3339)
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println(ui.H1("Festival Link"))
		fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(festivalName, ui.FestivalColor))
		if found {
			fmt.Printf("%s %s\n", ui.Label("Project"), ui.Dim(link.Path))
			fmt.Printf("%s %s\n", ui.Label("Linked"), ui.Dim(link.LinkedAt.Format("2006-01-02 15:04:05")))
		} else {
			fmt.Printf("%s %s\n", ui.Label("Project"), ui.Dim("(not linked)"))
			fmt.Println()
			fmt.Println(ui.Dim("Use 'fest link <path>' to link a project directory"))
		}
	}

	return nil
}

// NewUnlinkCommand creates the unlink command
func NewUnlinkCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "unlink",
		Short: "Remove festival-project link (context-aware)",
		Long: `Remove the project link for the current location.

Context-aware behavior:
  - Inside a festival: unlinks that festival from its project
  - Inside a linked project: unlinks the project from its festival

This removes the association between the festival and its project directory.`,
		Example: `  fest unlink   # Remove link for current festival or project`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUnlink(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func runUnlink(jsonOutput bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Load navigation state first - we need it for both paths
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	var festivalName string

	// Try to detect current festival
	loc, _ := show.DetectCurrentLocation(cwd)

	if loc != nil && loc.Festival != nil {
		// Inside a festival - unlink it
		festivalName = loc.Festival.Name
	} else {
		// Not in a festival - check if we're in a linked project
		for name, link := range nav.Links {
			if strings.HasPrefix(cwd, link.Path) || cwd == link.Path {
				festivalName = name
				break
			}
		}

		if festivalName == "" {
			return errors.NotFound("link").
				WithField("hint", "run from inside a festival or linked project directory")
		}
	}

	removed := nav.RemoveLink(festivalName)

	if removed {
		// Save navigation state
		if err := nav.Save(); err != nil {
			return errors.Wrap(err, "saving navigation state")
		}
	}

	if jsonOutput {
		result := map[string]interface{}{
			"success":  true,
			"festival": festivalName,
			"removed":  removed,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		if removed {
			fmt.Printf("%s %s\n", ui.Success("✓ Unlinked"), ui.Value(festivalName, ui.FestivalColor))
		} else {
			fmt.Printf("%s %s\n", ui.Warning("Not linked"), ui.Value(festivalName, ui.FestivalColor))
		}
	}

	return nil
}

// NewLinksCommand creates the links command to list all links
func NewLinksCommand() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "links",
		Short: "List all festival-project links",
		Long: `Display all festival-project links.

Shows a table of all festivals that have been linked to project directories.`,
		Example: `  fest links        # List all links
  fest links --json # List in JSON format`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinks(jsonOutput)
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")

	return cmd
}

func runLinks(jsonOutput bool) error {
	// Load navigation state
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return errors.Wrap(err, "loading navigation state")
	}

	links := nav.ListLinks()

	if jsonOutput {
		type linkInfo struct {
			Festival string `json:"festival"`
			Project  string `json:"project"`
			LinkedAt string `json:"linked_at"`
		}
		var linkList []linkInfo
		for name, link := range links {
			linkList = append(linkList, linkInfo{
				Festival: name,
				Project:  link.Path,
				LinkedAt: link.LinkedAt.Format(time.RFC3339),
			})
		}
		result := map[string]interface{}{
			"count": len(linkList),
			"links": linkList,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		if len(links) == 0 {
			fmt.Println(ui.H1("Festival Links"))
			fmt.Println(ui.Dim("No festival-project links configured."))
			fmt.Println()
			fmt.Println(ui.Dim("Use 'fest link <path>' from within a festival to create a link."))
			return nil
		}

		fmt.Println(ui.H1("Festival Links"))
		fmt.Printf("%s %s\n", ui.Label("Total"), ui.Value(fmt.Sprintf("%d", len(links))))
		fmt.Println(ui.Dim(strings.Repeat("─", 60)))

		// Try to find festivals directory to look up statuses
		cwd, _ := os.Getwd()
		festivalsDir, _ := workspace.FindFestivals(cwd)

		index := 0
		for name, link := range links {
			if index > 0 {
				fmt.Println()
			}
			// Try to get festival status for coloring
			styledName := ui.Value(name, ui.FestivalColor)
			if festivalsDir != "" {
				status := findFestivalStatus(festivalsDir, name)
				if status != "" {
					styledName = ui.GetStatusStyle(status).Render(name)
				}
			}
			fmt.Printf("%s %s %s\n", styledName, ui.Dim("→"), ui.Dim(link.Path))
			fmt.Printf("%s %s\n", ui.Dim("Linked"), ui.Dim(link.LinkedAt.Format("2006-01-02")))
			index++
		}
	}

	return nil
}

// findFestivalStatus looks up a festival's status by searching status directories.
// Returns empty string if festival not found.
func findFestivalStatus(festivalsDir, festivalName string) string {
	statuses := []string{"active", "planned", "completed", "dungeon"}
	for _, status := range statuses {
		statusDir := filepath.Join(festivalsDir, status)
		festivalPath := filepath.Join(statusDir, festivalName)
		if info, err := os.Stat(festivalPath); err == nil && info.IsDir() {
			return status
		}
	}
	return ""
}
