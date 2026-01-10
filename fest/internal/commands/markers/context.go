package markers

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// MarkerGoalContext holds extracted goal information for markers command
type MarkerGoalContext struct {
	Name string `json:"name"`
	Goal string `json:"goal"`
}

// MarkerFileContext holds the context hierarchy for a file
type MarkerFileContext struct {
	Festival *MarkerGoalContext `json:"festival,omitempty"`
	Phase    *MarkerGoalContext `json:"phase,omitempty"`
	Sequence *MarkerGoalContext `json:"sequence,omitempty"`
}

// FileWithContext represents a file's markers with its context
type FileWithContext struct {
	Path       string             `json:"path"`
	FullPath   string             `json:"full_path"`
	Context    *MarkerFileContext `json:"context"`
	Markers    []MarkerOccurrence `json:"markers"`
	Position   int                `json:"position"`
	TotalFiles int                `json:"total_files"`
}

// extractPrimaryGoal extracts the Primary Goal text from a goal markdown file
func extractPrimaryGoal(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	text := string(content)

	// Look for "Primary Goal" section heading
	// Pattern matches: "## Primary Goal", "**Primary Goal:**", "**Primary Goal**:", etc.
	primaryGoalRe := regexp.MustCompile(`(?mi)^#+\s*\*{0,2}\s*Primary\s+Goal\s*\*{0,2}\s*:?\s*$`)
	loc := primaryGoalRe.FindStringIndex(text)

	if loc == nil {
		// Fallback: try inline format "**Primary Goal:** Some text"
		inlineRe := regexp.MustCompile(`(?mi)\*{0,2}\s*Primary\s+Goal\s*\*{0,2}\s*:\s*\*{0,2}\s*(.+?)(\*{0,2})$`)
		if match := inlineRe.FindStringSubmatch(text); len(match) > 1 {
			goal := strings.TrimSpace(match[1])
			// Clean up markdown formatting from start and end
			goal = strings.Trim(goal, "*_ ")
			// Don't return markers as goals
			if !strings.HasPrefix(goal, "[") {
				return goal, nil
			}
		}
		return "", nil // No primary goal found
	}

	// Extract content after "Primary Goal" heading
	remainder := text[loc[1]:]

	// Find end of section (next heading or end of text)
	endRe := regexp.MustCompile(`(?m)^#+\s`)
	endLoc := endRe.FindStringIndex(remainder)

	var goalContent string
	if endLoc != nil {
		goalContent = remainder[:endLoc[0]]
	} else {
		goalContent = remainder
	}

	// Clean up the goal content
	lines := strings.Split(goalContent, "\n")
	var goalLines []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// Skip empty lines, template markers, and metadata
		if trimmed == "" ||
			strings.HasPrefix(trimmed, "[") ||
			strings.HasPrefix(trimmed, "{{") ||
			strings.HasPrefix(trimmed, "---") {
			continue
		}
		// Remove markdown formatting (bold, italic)
		// Replace **text** with text, *text* with text
		trimmed = regexp.MustCompile(`\*\*(.+?)\*\*`).ReplaceAllString(trimmed, "$1")
		trimmed = regexp.MustCompile(`\*(.+?)\*`).ReplaceAllString(trimmed, "$1")
		trimmed = strings.Trim(trimmed, "*_")
		goalLines = append(goalLines, trimmed)
	}

	goal := strings.Join(goalLines, " ")
	goal = strings.TrimSpace(goal)

	// If goal is too long, truncate it
	if len(goal) > 500 {
		goal = goal[:497] + "..."
	}

	return goal, nil
}

// buildMarkerFileContext builds the context hierarchy for a file
func buildMarkerFileContext(festivalPath, filePath string) (*MarkerFileContext, error) {
	ctx := &MarkerFileContext{}

	// Festival context (always present)
	festivalGoalPath := filepath.Join(festivalPath, "FESTIVAL_GOAL.md")
	festivalGoal, _ := extractPrimaryGoal(festivalGoalPath)
	ctx.Festival = &MarkerGoalContext{
		Name: filepath.Base(festivalPath),
		Goal: festivalGoal,
	}

	// Determine file level and extract phase/sequence context
	relPath, err := filepath.Rel(festivalPath, filePath)
	if err != nil {
		return ctx, nil
	}

	parts := strings.Split(filepath.Clean(relPath), string(filepath.Separator))

	// If file is in a phase directory (numbered directory)
	if len(parts) >= 2 && isNumberedDir(parts[0]) {
		phaseName := parts[0]
		phaseGoalPath := filepath.Join(festivalPath, phaseName, "PHASE_GOAL.md")
		phaseGoal, _ := extractPrimaryGoal(phaseGoalPath)
		ctx.Phase = &MarkerGoalContext{
			Name: phaseName,
			Goal: phaseGoal,
		}
	}

	// If file is in a sequence directory (numbered directory within phase)
	if len(parts) >= 3 && isNumberedDir(parts[0]) && isNumberedDir(parts[1]) {
		seqName := parts[1]
		seqGoalPath := filepath.Join(festivalPath, parts[0], seqName, "SEQUENCE_GOAL.md")
		seqGoal, _ := extractPrimaryGoal(seqGoalPath)
		ctx.Sequence = &MarkerGoalContext{
			Name: seqName,
			Goal: seqGoal,
		}
	}

	return ctx, nil
}

// isNumberedDir checks if a directory name starts with a number
func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// getNextFileWithContext returns the first file with markers and its context
func getNextFileWithContext(festivalPath string, sortedFiles []string, summary *MarkerSummary) (*FileWithContext, error) {
	if len(sortedFiles) == 0 {
		return nil, nil
	}

	nextFile := sortedFiles[0]
	fullPath := filepath.Join(festivalPath, nextFile)

	ctx, err := buildMarkerFileContext(festivalPath, fullPath)
	if err != nil {
		return nil, err
	}

	// Get markers for this file
	var fileMarkers []MarkerOccurrence
	for _, occ := range summary.Occurrences {
		if occ.File == nextFile {
			fileMarkers = append(fileMarkers, occ)
		}
	}

	return &FileWithContext{
		Path:       nextFile,
		FullPath:   fullPath,
		Context:    ctx,
		Markers:    fileMarkers,
		Position:   1,
		TotalFiles: len(sortedFiles),
	}, nil
}
