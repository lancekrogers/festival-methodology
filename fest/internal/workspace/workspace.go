// Package workspace provides workspace-aware festivals directory detection.
package workspace

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

const (
	// MarkerFile is the name of the workspace marker file inside .festival/
	MarkerFile = ".workspace"
	// FestivalsDir is the expected name of the festivals directory
	FestivalsDir = "festivals"
	// DotFestival is the hidden directory inside festivals/
	DotFestival = ".festival"
)

// Marker represents the .workspace file content
type Marker struct {
	Workspace  string    `json:"workspace"`
	Registered time.Time `json:"registered"`
}

// MarkerPath returns the full path to the marker file for a given festivals directory
func MarkerPath(festivalsDir string) string {
	return filepath.Join(festivalsDir, DotFestival, MarkerFile)
}

// HasMarker checks if a festivals directory has a workspace marker
func HasMarker(festivalsDir string) bool {
	markerPath := MarkerPath(festivalsDir)
	info, err := os.Stat(markerPath)
	return err == nil && !info.IsDir()
}

// ReadMarker reads and parses the workspace marker from a festivals directory
func ReadMarker(festivalsDir string) (*Marker, error) {
	markerPath := MarkerPath(festivalsDir)
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return nil, err
	}

	var marker Marker
	if err := json.Unmarshal(data, &marker); err != nil {
		return nil, err
	}

	return &marker, nil
}

// RegisterFestivals creates a .workspace marker in festivals/.festival/
// The workspace name is derived from the parent directory of festivals/
func RegisterFestivals(festivalsDir string) error {
	// Ensure the path is absolute
	absPath, err := filepath.Abs(festivalsDir)
	if err != nil {
		return err
	}

	// Derive workspace name from parent directory
	parentDir := filepath.Dir(absPath)
	workspaceName := filepath.Base(parentDir)

	// Create marker
	marker := Marker{
		Workspace:  workspaceName,
		Registered: time.Now().UTC(),
	}

	data, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return err
	}

	// Ensure .festival directory exists
	dotFestivalPath := filepath.Join(absPath, DotFestival)
	if err := os.MkdirAll(dotFestivalPath, 0755); err != nil {
		return err
	}

	// Write marker file
	markerPath := MarkerPath(absPath)
	return os.WriteFile(markerPath, data, 0644)
}

// UnregisterFestivals removes the .workspace marker from a festivals directory
func UnregisterFestivals(festivalsDir string) error {
	absPath, err := filepath.Abs(festivalsDir)
	if err != nil {
		return err
	}

	markerPath := MarkerPath(absPath)
	err = os.Remove(markerPath)
	if os.IsNotExist(err) {
		return nil // Already unregistered
	}
	return err
}

// FindMarkedFestivals walks UP from startDir looking for festivals/.festival/.workspace
// Returns the path to the first festivals/ directory that has a marker, or empty string if none found
func FindMarkedFestivals(startDir string) (string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	dir := absStart
	for {
		// Check if festivals/.festival/.workspace exists at this level
		festivalsPath := filepath.Join(dir, FestivalsDir)
		if HasMarker(festivalsPath) {
			return festivalsPath, nil
		}

		// Move up to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, no marker found
			return "", nil
		}
		dir = parent
	}
}

// FindAllMarkedFestivals walks UP from startDir and collects ALL festivals/ directories with markers
// Used for `fest go --all`
func FindAllMarkedFestivals(startDir string) ([]string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return nil, err
	}

	var results []string
	dir := absStart
	for {
		// Check if festivals/.festival/.workspace exists at this level
		festivalsPath := filepath.Join(dir, FestivalsDir)
		if HasMarker(festivalsPath) {
			results = append(results, festivalsPath)
		}

		// Move up to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root
			break
		}
		dir = parent
	}

	return results, nil
}

// FindNearestFestivals walks UP from startDir looking for any festivals/ directory
// This is the fallback behavior when no markers are found
func FindNearestFestivals(startDir string) (string, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return "", err
	}

	dir := absStart
	for {
		// Check if festivals/ exists at this level
		festivalsPath := filepath.Join(dir, FestivalsDir)
		info, err := os.Stat(festivalsPath)
		if err == nil && info.IsDir() {
			// Also check for .festival/ inside to confirm it's a valid festivals dir
			dotFestivalPath := filepath.Join(festivalsPath, DotFestival)
			if info, err := os.Stat(dotFestivalPath); err == nil && info.IsDir() {
				return festivalsPath, nil
			}
		}

		// Move up to parent
		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached root, no festivals found
			return "", nil
		}
		dir = parent
	}
}

// FindFestivals finds the appropriate festivals directory, preferring marked ones
// Falls back to nearest festivals/ if no markers exist
func FindFestivals(startDir string) (string, error) {
	// First try to find a marked festivals directory
	marked, err := FindMarkedFestivals(startDir)
	if err != nil {
		return "", err
	}
	if marked != "" {
		return marked, nil
	}

	// Fall back to nearest festivals directory
	return FindNearestFestivals(startDir)
}
