package navigation

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

// dateDirectoryPattern matches YYYY-MM format for date-based directory organization
var dateDirectoryPattern = regexp.MustCompile(`^\d{4}-\d{2}$`)

// NewGoCompletionsCommand creates the hidden completions subcommand for shell integration
func NewGoCompletionsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "completions",
		Short:  "Output completion words for shell integration",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGoCompletions()
		},
	}

	return cmd
}

func runGoCompletions() error {
	// Subcommands
	subcommands := []string{
		"list",
		"link",
		"map",
		"unmap",
		"project",
		"fest",
		"help",
	}

	// Status directories
	statuses := []string{
		"active",
		"planned",
		"completed",
		"dungeon",
	}

	// Output subcommands
	for _, cmd := range subcommands {
		fmt.Println(cmd)
	}

	// Output status directories
	for _, status := range statuses {
		fmt.Println(status)
	}

	// Load navigation state for shortcuts
	nav, err := navigation.LoadNavigation()
	if err != nil {
		// Silently skip shortcuts if navigation fails
		return nil
	}

	// Output shortcuts with - prefix
	for name := range nav.Shortcuts {
		fmt.Printf("-%s\n", name)
	}

	return nil
}

// isDateDirectory checks if a directory name matches YYYY-MM pattern
func isDateDirectory(name string) bool {
	return dateDirectoryPattern.MatchString(name)
}

// findCompletedFestivals finds all completed festivals, including those in date subdirectories.
// It supports both legacy flat structure and new date-based organization.
func findCompletedFestivals(completedDir, prefix string) []string {
	entries, err := os.ReadDir(completedDir)
	if err != nil {
		return nil
	}

	var festivals []string

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()

		if isDateDirectory(name) {
			// Search within date directory
			datePath := filepath.Join(completedDir, name)
			subEntries, err := os.ReadDir(datePath)
			if err != nil {
				continue
			}

			for _, subEntry := range subEntries {
				if !subEntry.IsDir() {
					continue
				}
				festName := subEntry.Name()
				// Check if it's a valid festival
				festPath := filepath.Join(datePath, festName)
				if !isValidFestivalDir(festPath) {
					continue
				}
				// Check prefix filter
				if prefix == "" || len(festName) >= len(prefix) && festName[:len(prefix)] == prefix {
					festivals = append(festivals, festName)
				}
			}
		} else {
			// Legacy flat structure
			festPath := filepath.Join(completedDir, name)
			if !isValidFestivalDir(festPath) {
				continue
			}
			if prefix == "" || len(name) >= len(prefix) && name[:len(prefix)] == prefix {
				festivals = append(festivals, name)
			}
		}
	}

	return festivals
}

// isValidFestivalDir checks if a directory contains festival markers
func isValidFestivalDir(path string) bool {
	// Check for FESTIVAL_OVERVIEW.md or FESTIVAL_GOAL.md
	if _, err := os.Stat(filepath.Join(path, "FESTIVAL_OVERVIEW.md")); err == nil {
		return true
	}
	if _, err := os.Stat(filepath.Join(path, "FESTIVAL_GOAL.md")); err == nil {
		return true
	}
	return false
}

// CompleteGoTarget provides fuzzy completions for the go command target argument
func CompleteGoTarget(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	// Get festivals directory
	cwd, err := os.Getwd()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil || festivalsDir == "" {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// Collect targets
	targets := navigation.CollectNavigationTargets(festivalsDir)
	if len(targets) == 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	// If no partial input, return all targets
	if toComplete == "" {
		result := make([]string, len(targets))
		for i, t := range targets {
			result[i] = t.Name
		}
		return result, cobra.ShellCompDirectiveNoFileComp
	}

	// Fuzzy filter
	finder := navigation.NewFuzzyFinder(targets)
	matches := finder.Find(toComplete)

	result := make([]string, len(matches))
	for i, m := range matches {
		result[i] = m.Name
	}

	return result, cobra.ShellCompDirectiveNoFileComp
}
