package gates

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	gatescore "github.com/lancekrogers/festival-methodology/fest/internal/gates"
)

// resolvePaths resolves festival, phase, and sequence paths from flags or cwd.
func resolvePaths(festivalsRoot, cwd, phase, sequence string) (festivalPath, phasePath, sequencePath string, err error) {
	// Try to detect current festival from cwd
	festivalPath = findCurrentFestival(festivalsRoot, cwd)
	if festivalPath == "" {
		// Default to first active festival
		activeDir := filepath.Join(festivalsRoot, "active")
		entries, err := os.ReadDir(activeDir)
		if err == nil {
			for _, e := range entries {
				if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
					festivalPath = filepath.Join(activeDir, e.Name())
					break
				}
			}
		}
	}

	if festivalPath == "" {
		return "", "", "", errors.NotFound("festival")
	}

	if sequence != "" {
		// Parse sequence as "phase/sequence"
		parts := strings.SplitN(sequence, "/", 2)
		if len(parts) != 2 {
			return "", "", "", errors.Validation("sequence must be in format 'phase/sequence'").
				WithField("sequence", sequence)
		}
		phasePath = filepath.Join(festivalPath, parts[0])
		sequencePath = filepath.Join(phasePath, parts[1])

		// Verify paths exist
		if _, err := os.Stat(phasePath); os.IsNotExist(err) {
			return "", "", "", errors.NotFound("phase").WithField("phase", parts[0])
		}
		if _, err := os.Stat(sequencePath); os.IsNotExist(err) {
			return "", "", "", errors.NotFound("sequence").WithField("sequence", sequence)
		}
	} else if phase != "" {
		phasePath = filepath.Join(festivalPath, phase)
		if _, err := os.Stat(phasePath); os.IsNotExist(err) {
			return "", "", "", errors.NotFound("phase").WithField("phase", phase)
		}
	}

	return festivalPath, phasePath, sequencePath, nil
}

// findCurrentFestival finds the festival directory containing cwd.
func findCurrentFestival(festivalsRoot, cwd string) string {
	// Check if we're inside a festival directory
	rel, err := filepath.Rel(festivalsRoot, cwd)
	if err != nil {
		return ""
	}

	// Walk up from cwd looking for festival markers
	parts := strings.Split(rel, string(filepath.Separator))
	for i := len(parts); i > 0; i-- {
		candidate := filepath.Join(festivalsRoot, filepath.Join(parts[:i]...))
		// Check for FESTIVAL_GOAL.md or FESTIVAL_OVERVIEW.md
		if _, err := os.Stat(filepath.Join(candidate, "FESTIVAL_GOAL.md")); err == nil {
			return candidate
		}
		if _, err := os.Stat(filepath.Join(candidate, "FESTIVAL_OVERVIEW.md")); err == nil {
			return candidate
		}
	}

	return ""
}

// getConfigRoot returns the user config root directory.
func getConfigRoot() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "fest", "active")
}

// resolveTargetPath determines the target path and override file name.
func resolveTargetPath(festivalPath, phasePath, sequencePath string) (targetPath, overrideFile string) {
	if sequencePath != "" {
		return sequencePath, gatescore.PhaseOverrideFileName
	}
	if phasePath != "" {
		return phasePath, gatescore.PhaseOverrideFileName
	}
	return festivalPath, filepath.Join(".festival", "gates.yml")
}

// gateOutput is the JSON output format for a gate.
type gateOutput struct {
	ID       string `json:"id"`
	Template string `json:"template"`
	Name     string `json:"name,omitempty"`
	Enabled  bool   `json:"enabled"`
	Removed  bool   `json:"removed,omitempty"`
	Source   string `json:"source,omitempty"`
}

// sourceOutput is the JSON output format for a policy source.
type sourceOutput struct {
	Level string `json:"level"`
	Path  string `json:"path,omitempty"`
	Name  string `json:"name,omitempty"`
}

// validationIssue represents a validation error or warning.
type validationIssue struct {
	Path     string `json:"path"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
}
