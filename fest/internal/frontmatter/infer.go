package frontmatter

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// InferFromPath infers frontmatter from a file path
func InferFromPath(path string) (*Frontmatter, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	fm := &Frontmatter{
		Created: info.ModTime(),
		Updated: info.ModTime(),
	}

	// Determine type from filename and path
	filename := filepath.Base(path)
	dir := filepath.Dir(path)

	switch {
	case strings.HasPrefix(strings.ToUpper(filename), "FESTIVAL_"):
		fm.Type = TypeFestival
		fm.ID = inferFestivalID(dir)
		fm.Name = inferName(filename)
		fm.Status = StatusActive

	case strings.HasPrefix(strings.ToUpper(filename), "PHASE_"):
		fm.Type = TypePhase
		fm.ID = filepath.Base(dir)
		fm.Name = inferName(filename)
		fm.Parent = inferFestivalID(filepath.Dir(dir))
		fm.Order = extractOrder(filepath.Base(dir))
		fm.Status = StatusPending

	case strings.HasPrefix(strings.ToUpper(filename), "SEQUENCE_"):
		fm.Type = TypeSequence
		fm.ID = filepath.Base(dir)
		fm.Name = inferName(filename)
		fm.Parent = filepath.Base(filepath.Dir(dir))
		fm.Order = extractOrder(filepath.Base(dir))
		fm.Status = StatusPending

	case strings.HasSuffix(filename, ".md"):
		// Task or gate file
		fm.Type = TypeTask
		fm.ID = strings.TrimSuffix(filename, ".md")
		fm.Name = inferName(filename)
		fm.Parent = filepath.Base(dir)
		fm.Order = extractOrder(filename)
		fm.Status = StatusPending

		// Check if it's a gate
		if isGateFile(filename) {
			fm.Type = TypeGate
			fm.GateType = inferGateType(filename)
		}
	}

	return fm, nil
}

// inferFestivalID extracts festival ID from path
func inferFestivalID(festivalPath string) string {
	return filepath.Base(festivalPath)
}

// inferName creates a human-readable name from filename
func inferName(filename string) string {
	// Remove extension
	name := strings.TrimSuffix(filename, ".md")

	// Remove prefix like "FESTIVAL_", "PHASE_", "SEQUENCE_"
	prefixes := []string{"FESTIVAL_", "PHASE_", "SEQUENCE_", "GOAL"}
	for _, prefix := range prefixes {
		name = strings.TrimPrefix(strings.ToUpper(name), prefix)
		name = strings.TrimPrefix(name, prefix)
	}

	// Remove numeric prefix
	re := regexp.MustCompile(`^\d+[_\-]?`)
	name = re.ReplaceAllString(name, "")

	// Convert underscores to spaces and title case
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.TrimSpace(name)

	return name
}

// extractOrder extracts numeric order from a filename or directory name
func extractOrder(name string) int {
	// Extract leading digits
	var numStr string
	for _, c := range name {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		} else {
			break
		}
	}

	if numStr == "" {
		return 0
	}

	num, err := strconv.Atoi(numStr)
	if err != nil {
		return 0
	}

	return num
}

// isGateFile checks if a filename appears to be a quality gate
func isGateFile(filename string) bool {
	lower := strings.ToLower(filename)
	gateIndicators := []string{
		"quality_gate",
		"gate_",
		"_gate",
		"testing_and_verify",
		"code_review",
		"review_results_iterate",
		"review_gate",
		"commit",
	}

	for _, indicator := range gateIndicators {
		if strings.Contains(lower, indicator) {
			return true
		}
	}

	return false
}

// inferGateType determines gate type from filename
func inferGateType(filename string) GateType {
	lower := strings.ToLower(filename)

	switch {
	case strings.Contains(lower, "testing") || strings.Contains(lower, "verify"):
		return GateTesting
	case strings.Contains(lower, "review"):
		return GateReview
	case strings.Contains(lower, "iterate"):
		return GateIterate
	case strings.Contains(lower, "security"):
		return GateSecurity
	case strings.Contains(lower, "performance") || strings.Contains(lower, "perf"):
		return GatePerformance
	}

	return GateTesting
}

// InferFromDirectory infers frontmatter for a directory (phase or sequence)
func InferFromDirectory(dirPath string) (*Frontmatter, error) {
	info, err := os.Stat(dirPath)
	if err != nil {
		return nil, err
	}

	if !info.IsDir() {
		return InferFromPath(dirPath)
	}

	fm := &Frontmatter{
		ID:      filepath.Base(dirPath),
		Created: info.ModTime(),
		Updated: info.ModTime(),
		Order:   extractOrder(filepath.Base(dirPath)),
		Status:  StatusPending,
	}

	// Determine type from directory depth
	parent := filepath.Dir(dirPath)
	grandparent := filepath.Dir(parent)

	// Check for goal files to determine type
	if hasGoalFile(dirPath, "PHASE_GOAL.md") {
		fm.Type = TypePhase
		fm.Parent = inferFestivalID(parent)
		fm.Name = inferName(filepath.Base(dirPath))
	} else if hasGoalFile(dirPath, "SEQUENCE_GOAL.md") {
		fm.Type = TypeSequence
		fm.Parent = filepath.Base(parent)
		fm.Name = inferName(filepath.Base(dirPath))
	} else if hasGoalFile(dirPath, "FESTIVAL_GOAL.md") || hasGoalFile(dirPath, "FESTIVAL_OVERVIEW.md") {
		fm.Type = TypeFestival
		fm.Name = inferName(filepath.Base(dirPath))
		fm.Status = StatusActive
	} else {
		// Try to infer from depth
		if isNumberedDir(filepath.Base(parent)) && isNumberedDir(filepath.Base(grandparent)) {
			// We're in a task directory? Unusual but possible
			fm.Type = TypeSequence
			fm.Parent = filepath.Base(parent)
		} else if isNumberedDir(filepath.Base(parent)) {
			fm.Type = TypeSequence
			fm.Parent = filepath.Base(parent)
		} else {
			fm.Type = TypePhase
			fm.Parent = inferFestivalID(parent)
		}
		fm.Name = inferName(filepath.Base(dirPath))
	}

	return fm, nil
}

// hasGoalFile checks if a directory contains a specific goal file
func hasGoalFile(dir, filename string) bool {
	_, err := os.Stat(filepath.Join(dir, filename))
	return err == nil
}

// isNumberedDir checks if directory name starts with a number
func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// InferCreatedTime tries to infer creation time from file
func InferCreatedTime(path string) time.Time {
	info, err := os.Stat(path)
	if err != nil {
		return time.Now()
	}
	return info.ModTime()
}
