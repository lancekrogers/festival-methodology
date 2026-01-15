package understand

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

// findDotFestivalDir searches for the .festival directory by walking up the directory tree.
func findDotFestivalDir() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up looking for festivals/.festival or .festival
	dir := cwd
	for {
		// Check for .festival in current dir
		candidate := filepath.Join(dir, ".festival")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}

		// Check for festivals/.festival
		candidate = filepath.Join(dir, "festivals", ".festival")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate
		}

		// Move up
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// readFileContent reads the entire file content as a string.
func readFileContent(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(content)
}

// extractSection extracts a section from content between markers.
func extractSection(pathOrContent string, startMarker, endMarker string) string {
	var content string
	if strings.Contains(pathOrContent, "\n") {
		content = pathOrContent
	} else {
		content = readFileContent(pathOrContent)
	}
	if content == "" {
		return ""
	}

	// Find start
	startIdx := strings.Index(content, startMarker)
	if startIdx == -1 {
		return ""
	}

	// Find content after start marker
	afterStart := content[startIdx+len(startMarker):]

	// Find end (look for next heading or end marker)
	endIdx := len(afterStart)
	if endMarker != "" {
		// Look for end marker or next same-level heading
		re := regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(endMarker))
		if loc := re.FindStringIndex(afterStart); loc != nil {
			endIdx = loc[0]
		}
	}

	section := strings.TrimSpace(afterStart[:endIdx])
	return section
}

// findCurrentFestival locates the festival root directory from cwd.
func findCurrentFestival() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Use template package's festival root finder
	root, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return ""
	}

	// Check if we're inside an active festival
	rel, err := filepath.Rel(root, cwd)
	if err != nil {
		return ""
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) >= 2 && (parts[0] == "active" || parts[0] == "planned" || parts[0] == "completed") {
		festivalPath := filepath.Join(root, parts[0], parts[1])
		// Verify fest.yaml exists
		if _, err := os.Stat(filepath.Join(festivalPath, "fest.yaml")); err == nil {
			return festivalPath
		}
	}

	return ""
}

// findFestivalRulesFile searches for FESTIVAL_RULES.md in order of preference.
func findFestivalRulesFile(dotFestival string) string {
	// Try current festival first
	if festivalDir := findCurrentFestival(); festivalDir != "" {
		rulesPath := filepath.Join(festivalDir, "FESTIVAL_RULES.md")
		if _, err := os.Stat(rulesPath); err == nil {
			return rulesPath
		}
	}

	// Try workspace .festival templates directory (for default rules)
	if dotFestival != "" {
		rulesPath := filepath.Join(dotFestival, "templates", "FESTIVAL_RULES_TEMPLATE.md")
		if _, err := os.Stat(rulesPath); err == nil {
			return rulesPath
		}
	}

	return ""
}

// hasSignificantUnfilledMarkers checks if file has too many unfilled template markers.
// Returns true if the file appears to be mostly unfilled template content.
func hasSignificantUnfilledMarkers(content string) bool {
	// Count unfilled markers
	markers := []string{"[FILL:", "[REPLACE:", "[GUIDANCE:", "{{ "}
	count := 0
	for _, marker := range markers {
		count += strings.Count(content, marker)
	}

	// If no markers, file is ready to use
	if count == 0 {
		return false
	}

	// If more than 20% of lines have markers, treat as unfilled template
	lines := strings.Count(content, "\n") + 1
	threshold := lines / 5 // 20%
	if threshold < 3 {
		threshold = 3 // Minimum threshold
	}

	return count > threshold
}
