package validator

import (
	"os"
	"path/filepath"
	"strings"
)

// stripInlineCode removes text inside backticks from a line.
// This prevents markers inside inline code examples from being counted.
func stripInlineCode(line string) string {
	result := strings.Builder{}
	inBacktick := false
	for _, ch := range line {
		if ch == '`' {
			inBacktick = !inBacktick
			continue
		}
		if !inBacktick {
			result.WriteRune(ch)
		}
	}
	return result.String()
}

// ValidateTemplateMarkers scans .md files for unfilled markers
func ValidateTemplateMarkers(festivalPath string) ([]Issue, error) {
	issues := []Issue{}
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}

	_ = filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(festivalPath, path)
		if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") {
			return nil
		}
		// Skip gates/ directory - these are intentional template files
		if strings.HasPrefix(rel, "gates/") || strings.HasPrefix(rel, "gates"+string(filepath.Separator)) {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// Scan line-by-line, skipping code blocks
		lines := strings.Split(string(content), "\n")
		inCodeBlock := false
		for _, line := range lines {
			// Toggle fence state on ``` lines
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			// Skip markers inside code blocks - they're documentation examples
			if inCodeBlock {
				continue
			}
			// Strip inline code (backticks) before checking for markers
			lineWithoutCode := stripInlineCode(line)
			for _, m := range markers {
				if strings.Contains(lineWithoutCode, m) {
					issues = append(issues, Issue{
						Level:   LevelWarning,
						Code:    CodeUnfilledTemplate,
						Path:    rel,
						Message: "File contains unfilled template marker: " + m,
						Fix:     "Edit file and replace template markers with actual content",
					})
					break
				}
			}
		}
		return nil
	})
	return issues, nil
}

// Checklist helpers reused by validate command

func CheckTemplatesFilled(festivalPath string) bool {
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}
	filled := true
	_ = filepath.Walk(festivalPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}
		rel, _ := filepath.Rel(festivalPath, path)
		if strings.HasPrefix(rel, ".") || strings.Contains(rel, "/.") {
			return nil
		}
		// Skip gates/ directory - these are intentional template files
		if strings.HasPrefix(rel, "gates/") || strings.HasPrefix(rel, "gates"+string(filepath.Separator)) {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		// Scan line-by-line, skipping code blocks
		lines := strings.Split(string(b), "\n")
		inCodeBlock := false
		for _, line := range lines {
			// Toggle fence state on ``` lines
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			// Skip markers inside code blocks - they're documentation examples
			if inCodeBlock {
				continue
			}
			// Strip inline code (backticks) before checking for markers
			lineWithoutCode := stripInlineCode(line)
			for _, m := range markers {
				if strings.Contains(lineWithoutCode, m) {
					filled = false
					return filepath.SkipAll
				}
			}
		}
		return nil
	})
	return filled
}
