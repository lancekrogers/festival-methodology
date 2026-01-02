package show

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

const (
	// FestivalGoalFile is the primary festival marker file
	FestivalGoalFile = "FESTIVAL_GOAL.md"
	// FestivalOverviewFile is an alternative festival marker file
	FestivalOverviewFile = "FESTIVAL_OVERVIEW.md"
	// FestivalConfigFile is the festival configuration file
	FestivalConfigFile = "fest.yaml"
	// PhaseGoalFile marks a phase directory
	PhaseGoalFile = "PHASE_GOAL.md"
	// SequenceGoalFile marks a sequence directory
	SequenceGoalFile = "SEQUENCE_GOAL.md"
)

// DetectCurrentFestival walks up from the given directory to find a festival root.
// Returns the festival information if found, or an error if not in a festival.
func DetectCurrentFestival(startDir string) (*FestivalInfo, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return nil, errors.IO("getting absolute path", err)
	}

	dir := absStart
	for {
		if isValidFestival(dir) {
			return parseFestivalInfo(dir)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			// Reached filesystem root
			return nil, errors.NotFound("festival")
		}
		dir = parent
	}
}

// isValidFestival checks if a directory is a valid festival root.
func isValidFestival(dir string) bool {
	// Check for FESTIVAL_GOAL.md or FESTIVAL_OVERVIEW.md
	goalPath := filepath.Join(dir, FestivalGoalFile)
	if info, err := os.Stat(goalPath); err == nil && !info.IsDir() {
		return true
	}

	overviewPath := filepath.Join(dir, FestivalOverviewFile)
	if info, err := os.Stat(overviewPath); err == nil && !info.IsDir() {
		return true
	}

	// Also check for fest.yaml as a fallback
	configPath := filepath.Join(dir, FestivalConfigFile)
	if info, err := os.Stat(configPath); err == nil && !info.IsDir() {
		return true
	}

	return false
}

// FindFestivalByName searches for a festival by name in all status directories.
func FindFestivalByName(festivalsDir, name string) (*FestivalInfo, error) {
	statusDirs := []string{"active", "planned", "completed", "dungeon"}

	for _, status := range statusDirs {
		statusDir := filepath.Join(festivalsDir, status)
		entries, err := os.ReadDir(statusDir)
		if err != nil {
			continue // Skip inaccessible directories
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}

			// Check exact match or prefix match
			if entry.Name() == name || strings.HasPrefix(entry.Name(), name+"_") || strings.Contains(entry.Name(), name) {
				festivalDir := filepath.Join(statusDir, entry.Name())
				if isValidFestival(festivalDir) {
					info, err := parseFestivalInfo(festivalDir)
					if err != nil {
						continue
					}
					info.Status = status
					return info, nil
				}
			}
		}
	}

	return nil, errors.NotFound("festival").WithField("name", name)
}

// ListFestivalsByStatus returns all festivals in a given status directory.
func ListFestivalsByStatus(festivalsDir, status string) ([]*FestivalInfo, error) {
	statusDir := filepath.Join(festivalsDir, status)
	entries, err := os.ReadDir(statusDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []*FestivalInfo{}, nil
		}
		return nil, errors.IO("reading status directory", err).WithField("status", status)
	}

	var festivals []*FestivalInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		festivalDir := filepath.Join(statusDir, entry.Name())
		if !isValidFestival(festivalDir) {
			continue
		}

		info, err := parseFestivalInfo(festivalDir)
		if err != nil {
			// Include with minimal info on parse error
			info = &FestivalInfo{
				ID:     entry.Name(),
				Name:   entry.Name(),
				Status: status,
				Path:   festivalDir,
			}
		}
		info.Status = status
		festivals = append(festivals, info)
	}

	return festivals, nil
}

// parseFestivalInfo parses festival information from a directory.
func parseFestivalInfo(festivalDir string) (*FestivalInfo, error) {
	info := &FestivalInfo{
		ID:   filepath.Base(festivalDir),
		Name: filepath.Base(festivalDir),
		Path: festivalDir,
	}

	// Determine status from parent directory
	parentDir := filepath.Dir(festivalDir)
	parentName := filepath.Base(parentDir)
	switch parentName {
	case "active", "planned", "completed", "dungeon":
		info.Status = parentName
	default:
		info.Status = "unknown"
	}

	// Try to load fest.yaml to get metadata ID
	festConfig, err := config.LoadFestivalConfig(festivalDir)
	if err == nil && festConfig != nil {
		// Extract metadata ID if present
		if festConfig.Metadata.ID != "" {
			info.MetadataID = festConfig.Metadata.ID
		}
		// Keep metadata name separate from directory name (used for linking)
		if festConfig.Metadata.Name != "" {
			info.MetadataName = festConfig.Metadata.Name
		}
	}

	// Calculate statistics
	ctx := context.Background()
	stats, err := CalculateFestivalStats(ctx, festivalDir)
	if err == nil {
		info.Stats = stats
	}

	return info, nil
}

// DetectCurrentLocation determines where we are in a festival hierarchy.
// Returns the current location type (festival, phase, sequence, task) and path info.
func DetectCurrentLocation(startDir string) (*LocationInfo, error) {
	absStart, err := filepath.Abs(startDir)
	if err != nil {
		return nil, errors.IO("getting absolute path", err)
	}

	// First find the festival root
	festival, err := DetectCurrentFestival(absStart)
	if err != nil {
		return nil, err
	}

	loc := &LocationInfo{
		Festival: festival,
	}

	// Determine relative position within the festival
	relPath, err := filepath.Rel(festival.Path, absStart)
	if err != nil {
		return loc, nil // At festival root
	}

	if relPath == "." {
		loc.Type = "festival"
		return loc, nil
	}

	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) >= 1 {
		// First level is phase
		phaseDir := filepath.Join(festival.Path, parts[0])
		if isPhaseDir(phaseDir) {
			loc.Type = "phase"
			loc.Phase = parts[0]
		}
	}

	if len(parts) >= 2 && loc.Phase != "" {
		// Second level is sequence
		seqDir := filepath.Join(festival.Path, parts[0], parts[1])
		if isSequenceDir(seqDir) {
			loc.Type = "sequence"
			loc.Sequence = parts[1]
		}
	}

	if len(parts) >= 3 && loc.Sequence != "" {
		// Could be in a task file's directory
		loc.Type = "task"
		loc.Task = parts[2]
	}

	if loc.Type == "" {
		loc.Type = "festival"
	}

	return loc, nil
}

func isPhaseDir(dir string) bool {
	goalPath := filepath.Join(dir, PhaseGoalFile)
	if info, err := os.Stat(goalPath); err == nil && !info.IsDir() {
		return true
	}
	return false
}

func isSequenceDir(dir string) bool {
	goalPath := filepath.Join(dir, SequenceGoalFile)
	if info, err := os.Stat(goalPath); err == nil && !info.IsDir() {
		return true
	}
	return false
}

// LocationInfo describes the current location within a festival hierarchy.
type LocationInfo struct {
	Type     string        `json:"type"`               // festival, phase, sequence, task
	Festival *FestivalInfo `json:"festival,omitempty"` // Always present if in a festival
	Phase    string        `json:"phase,omitempty"`    // Phase directory name
	Sequence string        `json:"sequence,omitempty"` // Sequence directory name
	Task     string        `json:"task,omitempty"`     // Task file or directory
}
