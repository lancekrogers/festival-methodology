package navigation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

type goOptions struct {
	showWorkspace bool
	showAll       bool
	json          bool
}

// NewGoCommand creates the go navigation command
func NewGoCommand() *cobra.Command {
	opts := &goOptions{}

	cmd := &cobra.Command{
		Use:   "go [target]",
		Short: "Navigate to festivals/ - use 'fgo' after shell-init setup",
		Long: `Navigate to your workspace's festivals directory.

The go command finds the festivals/ directory that has been registered
as your active workspace using 'fest init --register'.

NOTE: This command prints the path. To actually change directories,
set up shell integration (one-time):

  # Add to ~/.zshrc or ~/.bashrc:
  eval "$(fest shell-init zsh)"

Then use 'fgo' to navigate:
  fgo              Navigate to festivals root
  fgo 002          Navigate to phase 002
  fgo 2/1          Navigate to phase 2, sequence 1

Without shell integration, use command substitution:
  cd $(fest go)
  cd $(fest go 002)

If no registered festivals are found, falls back to nearest festivals/.`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			target := ""
			if len(args) > 0 {
				target = args[0]
			}
			return runGo(target, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.showWorkspace, "workspace", false, "show which workspace was detected")
	cmd.Flags().BoolVar(&opts.showAll, "all", false, "list all registered festivals directories")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	// Add subcommands for navigation shortcuts
	cmd.AddCommand(NewGoMapCommand())
	cmd.AddCommand(NewGoUnmapCommand())
	cmd.AddCommand(NewGoListCommand())
	cmd.AddCommand(NewGoShortcutCommand())
	cmd.AddCommand(NewGoProjectCommand())
	cmd.AddCommand(NewGoFestCommand())
	cmd.AddCommand(NewGoLinkCommand())

	return cmd
}

func runGo(target string, opts *goOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Handle --all flag
	if opts.showAll {
		return runGoAll(cwd, opts)
	}

	// Find the appropriate festivals directory
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals directory")
	}

	if festivalsDir == "" {
		return errors.NotFound("festivals directory")
	}

	// Handle --workspace flag
	if opts.showWorkspace {
		return runGoWorkspace(festivalsDir, opts)
	}

	// Smart navigation: no target provided
	if target == "" {
		// Try smart bidirectional navigation
		if resultPath := trySmartNavigation(cwd, festivalsDir); resultPath != "" {
			if opts.json {
				fmt.Printf(`{"path": "%s"}%s`, resultPath, "\n")
			} else {
				fmt.Println(resultPath)
			}
			return nil
		}
		// Fall back to festivals root
		if opts.json {
			fmt.Printf(`{"path": "%s"}%s`, festivalsDir, "\n")
		} else {
			fmt.Println(festivalsDir)
		}
		return nil
	}

	// Resolve target if provided
	resolved, err := resolveGoTarget(target, festivalsDir)
	if err != nil {
		return err
	}

	// Output the path
	if opts.json {
		fmt.Printf(`{"path": "%s"}%s`, resolved, "\n")
	} else {
		fmt.Println(resolved)
	}

	return nil
}

// trySmartNavigation attempts bidirectional navigation based on current location
func trySmartNavigation(cwd, festivalsDir string) string {
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return ""
	}

	// Check if we're inside a festival
	if isInsideFestival(cwd) {
		// Try to find the festival name
		loc, err := show.DetectCurrentLocation(cwd)
		if err == nil && loc != nil && loc.Festival != nil && loc.Festival.Name != "" {
			// Check if there's a linked project
			if projectPath := nav.GetLinkedProject(loc.Festival.Name); projectPath != "" {
				// Verify the path exists
				if info, err := os.Stat(projectPath); err == nil && info.IsDir() {
					return projectPath
				}
			}
		}
		return "" // No linked project, fall through to default
	}

	// Check if we're in a linked project
	if festivalName := nav.FindFestivalForPath(cwd); festivalName != "" {
		// Find the festival's path
		festPath := resolveFestivalPath(festivalsDir, festivalName)
		if festPath != "" {
			return festPath
		}
	}

	return "" // No smart navigation available
}

func runGoAll(cwd string, opts *goOptions) error {
	allFestivals, err := workspace.FindAllMarkedFestivals(cwd)
	if err != nil {
		return errors.Wrap(err, "finding festivals directories")
	}

	if len(allFestivals) == 0 {
		// Fall back to showing nearest
		nearest, err := workspace.FindNearestFestivals(cwd)
		if err != nil || nearest == "" {
			return errors.NotFound("festivals directories")
		}
		if opts.json {
			fmt.Printf(`{"festivals": [{"path": "%s", "registered": false}]}%s`, nearest, "\n")
		} else {
			fmt.Printf("%s (not registered)\n", nearest)
		}
		return nil
	}

	if opts.json {
		fmt.Print(`{"festivals": [`)
		for i, f := range allFestivals {
			marker, _ := workspace.ReadMarker(f)
			ws := ""
			if marker != nil {
				ws = marker.Workspace
			}
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf(`{"workspace": "%s", "path": "%s", "registered": true}`, ws, f)
		}
		fmt.Println("]}")
	} else {
		for _, f := range allFestivals {
			marker, _ := workspace.ReadMarker(f)
			if marker != nil {
				fmt.Printf("%s → %s\n", marker.Workspace, f)
			} else {
				fmt.Println(f)
			}
		}
	}

	return nil
}

func runGoWorkspace(festivalsDir string, opts *goOptions) error {
	marker, err := workspace.ReadMarker(festivalsDir)
	ws := "(not registered)"
	if err == nil && marker != nil {
		ws = marker.Workspace
	}

	if opts.json {
		registered := marker != nil
		fmt.Printf(`{"workspace": "%s", "path": "%s", "registered": %t}%s`, ws, festivalsDir, registered, "\n")
	} else {
		fmt.Printf("%s → %s\n", ws, festivalsDir)
	}

	return nil
}

func resolveGoTarget(target, festivalsDir string) (string, error) {
	// Check if target looks like a phase shortcut (numeric)
	if isPhaseShortcut(target) {
		return resolvePhaseShortcut(target, festivalsDir)
	}

	// Check if target looks like phase/sequence shortcut
	if strings.Contains(target, "/") {
		parts := strings.SplitN(target, "/", 2)
		if isPhaseShortcut(parts[0]) {
			phaseDir, err := resolvePhaseShortcut(parts[0], festivalsDir)
			if err != nil {
				return "", err
			}
			if len(parts) > 1 && isSequenceShortcut(parts[1]) {
				return resolveSequenceShortcut(parts[1], phaseDir)
			}
			return phaseDir, nil
		}
	}

	// Try to resolve as festival name (searches active/, planned/, completed/)
	if festPath := resolveFestivalByName(target, festivalsDir); festPath != "" {
		return festPath, nil
	}

	// Treat as a relative path within festivals
	fullPath := filepath.Join(festivalsDir, target)
	if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
		return fullPath, nil
	}

	return "", errors.NotFound("target").WithField("target", target)
}

// resolveFestivalByName searches for a festival by name in status directories
func resolveFestivalByName(name, festivalsDir string) string {
	statusDirs := []string{"active", "planned", "completed"}
	for _, status := range statusDirs {
		festPath := filepath.Join(festivalsDir, status, name)
		if info, err := os.Stat(festPath); err == nil && info.IsDir() {
			return festPath
		}
	}
	return ""
}

func isPhaseShortcut(s string) bool {
	// Phase shortcuts are 1-3 digit numbers
	if len(s) == 0 || len(s) > 3 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func isSequenceShortcut(s string) bool {
	// Sequence shortcuts are 1-2 digit numbers
	if len(s) == 0 || len(s) > 2 {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func resolvePhaseShortcut(shortcut, festivalsDir string) (string, error) {
	// Pad to 3 digits
	padded := fmt.Sprintf("%03s", shortcut)
	if len(shortcut) < 3 {
		// Convert "2" to "002", "02" to "002"
		n := 0
		fmt.Sscanf(shortcut, "%d", &n)
		padded = fmt.Sprintf("%03d", n)
	}

	// Search in active/, planned/, completed/ subdirectories
	searchDirs := []string{
		filepath.Join(festivalsDir, "active"),
		filepath.Join(festivalsDir, "planned"),
		filepath.Join(festivalsDir, "completed"),
		festivalsDir, // Also search root
	}

	for _, searchDir := range searchDirs {
		entries, err := os.ReadDir(searchDir)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() && strings.HasPrefix(entry.Name(), padded+"_") {
				return filepath.Join(searchDir, entry.Name()), nil
			}
		}
	}

	return "", errors.NotFound("phase").WithField("shortcut", shortcut)
}

func resolveSequenceShortcut(shortcut, phaseDir string) (string, error) {
	// Pad to 2 digits
	n := 0
	fmt.Sscanf(shortcut, "%d", &n)
	padded := fmt.Sprintf("%02d", n)

	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return "", errors.IO("reading phase directory", err).WithField("path", phaseDir)
	}

	for _, entry := range entries {
		if entry.IsDir() && strings.HasPrefix(entry.Name(), padded+"_") {
			return filepath.Join(phaseDir, entry.Name()), nil
		}
	}

	return "", errors.NotFound("sequence").WithField("shortcut", shortcut).WithField("phase", filepath.Base(phaseDir))
}
