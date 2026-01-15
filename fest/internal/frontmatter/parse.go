package frontmatter

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

// Delimiter for frontmatter
const delimiter = "---"

// Parse extracts frontmatter from document content
// Returns the parsed frontmatter, remaining content, and any error
func Parse(content []byte) (*Frontmatter, []byte, error) {
	if !HasFrontmatter(content) {
		return nil, content, nil
	}

	// Find the start and end delimiters
	trimmed := bytes.TrimSpace(content)
	if !bytes.HasPrefix(trimmed, []byte(delimiter)) {
		return nil, content, nil
	}

	// Skip the opening delimiter
	rest := trimmed[len(delimiter):]
	rest = bytes.TrimPrefix(rest, []byte("\n"))

	// Find the closing delimiter
	endIdx := bytes.Index(rest, []byte("\n"+delimiter))
	if endIdx == -1 {
		return nil, content, errors.Parse("frontmatter missing closing delimiter", nil)
	}

	fmContent := rest[:endIdx]
	remaining := rest[endIdx+len("\n"+delimiter):]
	remaining = bytes.TrimPrefix(remaining, []byte("\n"))

	// Parse the YAML
	var fm Frontmatter
	if err := yaml.Unmarshal(fmContent, &fm); err != nil {
		return nil, content, errors.Parse("parsing frontmatter YAML", err)
	}

	return &fm, remaining, nil
}

// HasFrontmatter checks if content starts with frontmatter delimiter
func HasFrontmatter(content []byte) bool {
	trimmed := bytes.TrimSpace(content)
	return bytes.HasPrefix(trimmed, []byte(delimiter))
}

// Extract returns just the frontmatter without parsing
func Extract(content []byte) ([]byte, []byte, error) {
	if !HasFrontmatter(content) {
		return nil, content, nil
	}

	trimmed := bytes.TrimSpace(content)
	if !bytes.HasPrefix(trimmed, []byte(delimiter)) {
		return nil, content, nil
	}

	rest := trimmed[len(delimiter):]
	rest = bytes.TrimPrefix(rest, []byte("\n"))

	endIdx := bytes.Index(rest, []byte("\n"+delimiter))
	if endIdx == -1 {
		return nil, content, errors.Parse("frontmatter missing closing delimiter", nil)
	}

	fmContent := rest[:endIdx]
	remaining := rest[endIdx+len("\n"+delimiter):]
	remaining = bytes.TrimPrefix(remaining, []byte("\n"))

	return fmContent, remaining, nil
}

// Replace replaces existing frontmatter with new frontmatter
func Replace(content []byte, fm *Frontmatter) ([]byte, error) {
	// Remove existing frontmatter if present
	_, remaining, err := Extract(content)
	if err != nil {
		return nil, err
	}

	// Inject new frontmatter
	return Inject(remaining, fm)
}

// StripFrontmatter removes frontmatter from content
func StripFrontmatter(content []byte) ([]byte, error) {
	_, remaining, err := Extract(content)
	if err != nil {
		return nil, err
	}
	return remaining, nil
}

// ParseFile is a convenience function that parses frontmatter from file content
// and returns both the frontmatter and remaining content as a string
func ParseFile(content []byte) (*Frontmatter, string, error) {
	fm, remaining, err := Parse(content)
	if err != nil {
		return nil, "", err
	}
	return fm, string(remaining), nil
}

// ParseWithFallback parses frontmatter with legacy text format fallback
// Use this for reading documents that may not have YAML frontmatter
func ParseWithFallback(content []byte) (*Frontmatter, []byte, error) {
	// First try YAML frontmatter
	fm, remaining, err := Parse(content)
	if err != nil {
		return nil, content, err
	}
	if fm != nil {
		return fm, remaining, nil
	}

	// Fallback to legacy text parsing
	fm = parseLegacyText(content)
	return fm, content, nil
}

// parseLegacyText extracts metadata from legacy inline text format
// e.g., **Phase:** 003 | **Status:** Active | **Type:** Implementation
func parseLegacyText(content []byte) *Frontmatter {
	text := string(content)
	fm := &Frontmatter{}

	// Parse inline metadata patterns
	// Pattern: **Field:** Value or **Field:** Value |
	patterns := map[string]*regexp.Regexp{
		"status":     regexp.MustCompile(`\*\*Status:\*\*\s*([^|\n]+)`),
		"type":       regexp.MustCompile(`\*\*Type:\*\*\s*([^|\n]+)`),
		"phase":      regexp.MustCompile(`\*\*Phase:\*\*\s*([^|\n]+)`),
		"sequence":   regexp.MustCompile(`\*\*Sequence:\*\*\s*([^|\n]+)`),
		"created":    regexp.MustCompile(`\*\*Created:\*\*\s*([^|\n]+)`),
		"phase_type": regexp.MustCompile(`\*\*Phase Type:\*\*\s*([^|\n]+)`),
	}

	// Extract values
	for field, pattern := range patterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			value := strings.TrimSpace(matches[1])
			switch field {
			case "status":
				fm.Status = parseStatus(value)
			case "type":
				fm.PhaseType = parsePhaseType(value)
			case "phase":
				fm.Parent = value
			case "sequence":
				fm.Parent = value
			case "phase_type":
				fm.PhaseType = parsePhaseType(value)
			}
		}
	}

	// Try to infer document type from content
	if strings.Contains(text, "# Phase Goal:") || strings.Contains(text, "# Phase:") {
		fm.Type = TypePhase
	} else if strings.Contains(text, "# Sequence Goal:") || strings.Contains(text, "## Sequence Objective") {
		fm.Type = TypeSequence
	} else if strings.Contains(text, "# Task:") || strings.Contains(text, "## Task Objective") {
		fm.Type = TypeTask
	} else if strings.Contains(text, "# Festival") || strings.Contains(text, "## Festival Objective") {
		fm.Type = TypeFestival
	}

	// Return nil if we couldn't extract anything meaningful
	if fm.Type == "" && fm.Status == "" && fm.PhaseType == "" {
		return nil
	}

	return fm
}

// parseStatus converts status string to Status type
func parseStatus(s string) Status {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "planned", "planning":
		return StatusPlanned
	case "active", "in progress", "in_progress":
		return StatusInProgress
	case "complete", "completed":
		return StatusCompleted
	case "dungeon", "archived":
		return StatusDungeon
	case "pending":
		return StatusPending
	case "blocked":
		return StatusBlocked
	case "passed":
		return StatusPassed
	case "failed":
		return StatusFailed
	default:
		return ""
	}
}

// parsePhaseType converts phase type string to PhaseType
func parsePhaseType(s string) PhaseType {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "planning", "plan":
		return PhaseTypePlanning
	case "implementation", "implement", "build":
		return PhaseTypeImplementation
	case "research", "discovery":
		return PhaseTypeResearch
	case "review", "qa":
		return PhaseTypeReview
	case "deployment", "deploy", "release":
		return PhaseTypeDeployment
	default:
		return ""
	}
}
