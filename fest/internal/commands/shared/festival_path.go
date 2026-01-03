package shared

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
)

// ResolveFestivalPath resolves the festival path from multiple sources in priority order:
//  1. Explicit path (--festival flag, highest priority)
//  2. Navigation link for current directory (fest link)
//  3. FindFestivalRoot from current directory (fallback - looks for fest.yaml)
//
// This allows fest commands to work from linked project directories without
// requiring the user to cd into the festival directory.
func ResolveFestivalPath(cwd, explicitPath string) (string, error) {
	// 1. Explicit flag takes precedence
	if explicitPath != "" {
		return explicitPath, nil
	}

	// 2. Check for linked festival
	nav, err := navigation.LoadNavigation()
	if err == nil {
		if linkedFestivalName := nav.FindFestivalForPath(cwd); linkedFestivalName != "" {
			// Found a linked festival name, now search for its actual path
			festivalsRoot, err := workspace.FindFestivals(cwd)
			if err == nil && festivalsRoot != "" {
				festivalPath, err := findFestivalByName(festivalsRoot, linkedFestivalName)
				if err == nil {
					return festivalPath, nil
				}
			}
		}
	}

	// 3. Fall back to festival root detection
	return template.FindFestivalRoot(cwd)
}

// findFestivalByName searches for a festival by name in all status directories
func findFestivalByName(festivalsRoot, name string) (string, error) {
	statusDirs := []string{"active", "planned", "completed", "dungeon"}

	for _, status := range statusDirs {
		statusDir := filepath.Join(festivalsRoot, status)
		festivalPath := filepath.Join(statusDir, name)

		// Check if this festival directory exists
		if info, err := os.Stat(festivalPath); err == nil && info.IsDir() {
			// Verify it's a valid festival by checking for festival markers
			if isValidFestival(festivalPath) {
				return festivalPath, nil
			}
		}
	}

	// Festival name not found in any status directory
	return template.FindFestivalRoot(festivalsRoot)
}

// isValidFestival checks if a directory is a valid festival root
func isValidFestival(dir string) bool {
	// Check for FESTIVAL_GOAL.md or FESTIVAL_OVERVIEW.md
	goalPath := filepath.Join(dir, "FESTIVAL_GOAL.md")
	if info, err := os.Stat(goalPath); err == nil && !info.IsDir() {
		return true
	}

	overviewPath := filepath.Join(dir, "FESTIVAL_OVERVIEW.md")
	if info, err := os.Stat(overviewPath); err == nil && !info.IsDir() {
		return true
	}

	// Also check for fest.yaml as a fallback
	configPath := filepath.Join(dir, "fest.yaml")
	if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
		return true
	}

	return false
}
