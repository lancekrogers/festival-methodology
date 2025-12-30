package progress

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// Checkbox patterns for detection
var (
	// Standard markdown checkboxes
	checkedBoxPattern   = regexp.MustCompile(`^[\s]*[-*]\s*\[(x|X)\]`)
	uncheckedBoxPattern = regexp.MustCompile(`^[\s]*[-*]\s*\[\s*\]`)

	// Emoji checkboxes (per PROJECT_MANAGEMENT_SYSTEM.md)
	emojiCompletedPattern  = regexp.MustCompile(`\[✅\]`)
	emojiInProgressPattern = regexp.MustCompile(`\[🚧\]`)
	emojiBlockedPattern    = regexp.MustCompile(`\[❌\]`)
	emojiNotStartedPattern = regexp.MustCompile(`\[\s*\]`)

	// Section headers to look for (in priority order)
	statusSections = []string{
		"definition of done",
		"requirements",
		"acceptance criteria",
		"deliverables",
		"checklist",
	}
)

// CheckboxCounts holds checkbox statistics from a markdown file
type CheckboxCounts struct {
	Checked   int
	Unchecked int
	Total     int
}

// ParseTaskStatus reads a task markdown file and determines its status
// based on checkbox completion. Returns StatusCompleted, StatusInProgress,
// or StatusPending.
func ParseTaskStatus(taskPath string) string {
	file, err := os.Open(taskPath)
	if err != nil {
		return StatusPending
	}
	defer file.Close()

	// First pass: look for status sections
	sectionCounts := extractSectionCheckboxes(file)
	if sectionCounts.Total > 0 {
		return statusFromCounts(sectionCounts)
	}

	// Reset file for second pass
	file.Seek(0, 0)

	// Second pass: count all checkboxes in the file
	allCounts := extractAllCheckboxes(file)
	if allCounts.Total > 0 {
		return statusFromCounts(allCounts)
	}

	// No checkboxes found - default to pending
	return StatusPending
}

// extractSectionCheckboxes looks for priority sections and extracts checkbox counts
func extractSectionCheckboxes(file *os.File) CheckboxCounts {
	scanner := bufio.NewScanner(file)
	counts := CheckboxCounts{}

	inTargetSection := false
	currentSectionLevel := 0

	for scanner.Scan() {
		line := scanner.Text()

		// Check if this is a header
		headerLevel, headerText := parseHeader(line)
		if headerLevel > 0 {
			// Check if we're entering a target section
			lowerHeader := strings.ToLower(headerText)
			isTargetSection := false
			for _, section := range statusSections {
				if strings.Contains(lowerHeader, section) {
					isTargetSection = true
					break
				}
			}

			if isTargetSection {
				inTargetSection = true
				currentSectionLevel = headerLevel
			} else if headerLevel <= currentSectionLevel && inTargetSection {
				// We've left the target section (same or higher level header)
				inTargetSection = false
			}
			continue
		}

		// If we're in a target section, count checkboxes
		if inTargetSection {
			addCheckboxCounts(&counts, line)
		}
	}

	return counts
}

// extractAllCheckboxes counts all checkboxes in the file
func extractAllCheckboxes(file *os.File) CheckboxCounts {
	scanner := bufio.NewScanner(file)
	counts := CheckboxCounts{}

	for scanner.Scan() {
		addCheckboxCounts(&counts, scanner.Text())
	}

	return counts
}

// addCheckboxCounts checks a line for checkboxes and updates counts
func addCheckboxCounts(counts *CheckboxCounts, line string) {
	// Standard markdown checkboxes
	if checkedBoxPattern.MatchString(line) {
		counts.Checked++
		counts.Total++
		return
	}
	if uncheckedBoxPattern.MatchString(line) {
		counts.Unchecked++
		counts.Total++
		return
	}

	// Emoji checkboxes
	if emojiCompletedPattern.MatchString(line) {
		counts.Checked++
		counts.Total++
		return
	}
	if emojiInProgressPattern.MatchString(line) || emojiBlockedPattern.MatchString(line) {
		// In progress or blocked counts as unchecked but started
		counts.Unchecked++
		counts.Total++
		return
	}
}

// parseHeader extracts header level and text from a markdown header line
// Returns (0, "") if not a header
func parseHeader(line string) (int, string) {
	trimmed := strings.TrimLeft(line, " \t")
	if !strings.HasPrefix(trimmed, "#") {
		return 0, ""
	}

	level := 0
	for _, c := range trimmed {
		if c == '#' {
			level++
		} else {
			break
		}
	}

	if level > 6 {
		return 0, "" // Not a valid header
	}

	text := strings.TrimSpace(strings.TrimLeft(trimmed, "#"))
	return level, text
}

// statusFromCounts determines status based on checkbox completion
func statusFromCounts(counts CheckboxCounts) string {
	if counts.Total == 0 {
		return StatusPending
	}

	if counts.Checked == counts.Total {
		return StatusCompleted
	}

	if counts.Checked > 0 {
		return StatusInProgress
	}

	return StatusPending
}
