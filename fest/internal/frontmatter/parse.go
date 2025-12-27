package frontmatter

import (
	"bytes"
	"fmt"

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
		return nil, content, fmt.Errorf("frontmatter missing closing delimiter")
	}

	fmContent := rest[:endIdx]
	remaining := rest[endIdx+len("\n"+delimiter):]
	remaining = bytes.TrimPrefix(remaining, []byte("\n"))

	// Parse the YAML
	var fm Frontmatter
	if err := yaml.Unmarshal(fmContent, &fm); err != nil {
		return nil, content, fmt.Errorf("failed to parse frontmatter: %w", err)
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
		return nil, content, fmt.Errorf("frontmatter missing closing delimiter")
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
