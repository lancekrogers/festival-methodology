package gates

import (
	"bufio"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	// FrontmatterDelimiter is the YAML frontmatter delimiter
	FrontmatterDelimiter = "---"
)

// FileMarkers represents the managed file markers in frontmatter
type FileMarkers struct {
	FestManaged bool   `yaml:"fest_managed,omitempty"`
	FestGateID  string `yaml:"fest_gate_id,omitempty"`
	FestVersion string `yaml:"fest_version,omitempty"`
}

// ParseMarkers extracts markers from a file's YAML frontmatter
func ParseMarkers(filePath string) (*FileMarkers, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Check for frontmatter start
	if !scanner.Scan() {
		return nil, nil // Empty file
	}
	if strings.TrimSpace(scanner.Text()) != FrontmatterDelimiter {
		return nil, nil // No frontmatter
	}

	// Read frontmatter content
	var frontmatter strings.Builder
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == FrontmatterDelimiter {
			break // End of frontmatter
		}
		frontmatter.WriteString(line)
		frontmatter.WriteString("\n")
	}

	if frontmatter.Len() == 0 {
		return nil, nil // Empty frontmatter
	}

	// Parse YAML
	var markers FileMarkers
	if err := yaml.Unmarshal([]byte(frontmatter.String()), &markers); err != nil {
		return nil, err
	}

	return &markers, nil
}

// IsManaged checks if a file is managed by fest
func IsManaged(filePath string) bool {
	markers, err := ParseMarkers(filePath)
	if err != nil || markers == nil {
		return false
	}
	return markers.FestManaged
}

// GetGateID returns the gate ID from a file's markers
func GetGateID(filePath string) string {
	markers, err := ParseMarkers(filePath)
	if err != nil || markers == nil {
		return ""
	}
	return markers.FestGateID
}

// AddMarkers adds fest markers to content
func AddMarkers(content string, gateID string) string {
	markers := FileMarkers{
		FestManaged: true,
		FestGateID:  gateID,
		FestVersion: "1.0",
	}

	markerYAML, err := yaml.Marshal(markers)
	if err != nil {
		return content
	}

	// Check if content already has frontmatter
	if strings.HasPrefix(strings.TrimSpace(content), FrontmatterDelimiter) {
		// Insert markers into existing frontmatter
		return insertMarkersIntoFrontmatter(content, markers)
	}

	// Add new frontmatter with markers
	return FrontmatterDelimiter + "\n" + string(markerYAML) + FrontmatterDelimiter + "\n\n" + content
}

// insertMarkersIntoFrontmatter adds markers to existing frontmatter
func insertMarkersIntoFrontmatter(content string, markers FileMarkers) string {
	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return content
	}

	// Find the end of frontmatter
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == FrontmatterDelimiter {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return content // Malformed frontmatter
	}

	// Parse existing frontmatter
	existingYAML := strings.Join(lines[1:endIdx], "\n")
	var existing map[string]any
	if err := yaml.Unmarshal([]byte(existingYAML), &existing); err != nil {
		existing = make(map[string]any)
	}

	// Add markers
	existing["fest_managed"] = markers.FestManaged
	existing["fest_gate_id"] = markers.FestGateID
	existing["fest_version"] = markers.FestVersion

	// Serialize
	newYAML, err := yaml.Marshal(existing)
	if err != nil {
		return content
	}

	// Reconstruct
	result := FrontmatterDelimiter + "\n" + string(newYAML) + FrontmatterDelimiter
	if endIdx+1 < len(lines) {
		result += "\n" + strings.Join(lines[endIdx+1:], "\n")
	}

	return result
}

// StripMarkers removes fest markers from content
func StripMarkers(content string) string {
	if !strings.HasPrefix(strings.TrimSpace(content), FrontmatterDelimiter) {
		return content
	}

	lines := strings.Split(content, "\n")
	if len(lines) < 2 {
		return content
	}

	// Find the end of frontmatter
	endIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == FrontmatterDelimiter {
			endIdx = i
			break
		}
	}

	if endIdx == -1 {
		return content
	}

	// Parse and filter frontmatter
	existingYAML := strings.Join(lines[1:endIdx], "\n")
	var existing map[string]any
	if err := yaml.Unmarshal([]byte(existingYAML), &existing); err != nil {
		return content
	}

	// Remove markers
	delete(existing, "fest_managed")
	delete(existing, "fest_gate_id")
	delete(existing, "fest_version")

	// If nothing left in frontmatter, remove it entirely
	if len(existing) == 0 {
		if endIdx+1 < len(lines) {
			return strings.TrimLeft(strings.Join(lines[endIdx+1:], "\n"), "\n")
		}
		return ""
	}

	// Serialize remaining frontmatter
	newYAML, err := yaml.Marshal(existing)
	if err != nil {
		return content
	}

	result := FrontmatterDelimiter + "\n" + string(newYAML) + FrontmatterDelimiter
	if endIdx+1 < len(lines) {
		result += "\n" + strings.Join(lines[endIdx+1:], "\n")
	}

	return result
}
