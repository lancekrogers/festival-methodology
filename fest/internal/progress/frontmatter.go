// Package progress provides task progress tracking and status resolution for festivals.
package progress

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// FrontmatterMeta holds parsed frontmatter metadata from markdown files.
// Only fields relevant to progress tracking are parsed.
type FrontmatterMeta struct {
	// Tracking indicates whether the file should be tracked for progress.
	// nil = not specified (default to true)
	// true = explicitly tracked
	// false = explicitly excluded from tracking
	Tracking *bool `yaml:"tracking,omitempty"`
}

// ParseFrontmatter extracts YAML frontmatter from a markdown file.
// Returns nil with no error if the file has no frontmatter.
// Returns an error if the file cannot be read or frontmatter is malformed.
func ParseFrontmatter(filePath string) (*FrontmatterMeta, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Check for opening delimiter
	if !scanner.Scan() {
		return nil, nil // Empty file, no frontmatter
	}
	firstLine := strings.TrimSpace(scanner.Text())
	if firstLine != "---" {
		return nil, nil // No frontmatter
	}

	// Collect frontmatter content until closing delimiter
	var yamlContent strings.Builder
	foundClosing := false
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "---" {
			foundClosing = true
			break
		}
		yamlContent.WriteString(line)
		yamlContent.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	if !foundClosing {
		return nil, nil // Unclosed frontmatter, treat as no frontmatter
	}

	// Parse YAML content
	var meta FrontmatterMeta
	if err := yaml.Unmarshal([]byte(yamlContent.String()), &meta); err != nil {
		// Malformed YAML - fail-safe to treat as tracked
		return nil, nil
	}

	return &meta, nil
}

// IsTracked returns true if the file should be tracked for progress.
// Returns true if: file has no frontmatter, frontmatter has no tracking field,
// or tracking is not explicitly false.
// Returns true on any error (fail-safe to track).
func IsTracked(filePath string) bool {
	meta, err := ParseFrontmatter(filePath)
	if err != nil {
		// Fail-safe: track on error
		return true
	}
	if meta == nil {
		// No frontmatter, default to tracked
		return true
	}
	if meta.Tracking == nil {
		// No tracking field, default to tracked
		return true
	}
	// Return the explicit tracking value
	return *meta.Tracking
}
