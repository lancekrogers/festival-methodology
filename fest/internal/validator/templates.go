package validator

import (
	"os"
	"path/filepath"
	"strings"
)

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
		s := string(content)
		for _, m := range markers {
			if strings.Contains(s, m) {
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
		s := string(b)
		for _, m := range markers {
			if strings.Contains(s, m) {
				filled = false
				return filepath.SkipAll
			}
		}
		return nil
	})
	return filled
}
