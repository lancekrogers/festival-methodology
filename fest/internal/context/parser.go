package context

import (
	"bufio"
	"regexp"
	"strings"
	"time"
)

// ParseGoalFile extracts goal context from a goal markdown file
func ParseGoalFile(content []byte) (*GoalContext, error) {
	text := string(content)
	goal := &GoalContext{}

	// Extract title from first heading
	titleRe := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	if match := titleRe.FindStringSubmatch(text); len(match) > 1 {
		goal.Title = strings.TrimSpace(match[1])
	}

	// Extract objective section
	goal.Objective = extractSection(text, "Objective", "##")
	if goal.Objective == "" {
		// Try ## Goal section
		goal.Objective = extractSection(text, "Goal", "##")
	}

	// Extract success criteria (checklist items)
	goal.SuccessCriteria = extractChecklist(text)

	// Parse frontmatter for status/priority
	goal.Status = extractFrontmatterField(text, "status")
	goal.Priority = extractFrontmatterField(text, "priority")

	return goal, nil
}

// ParseRulesFile extracts rules from FESTIVAL_RULES.md
func ParseRulesFile(content []byte) ([]Rule, error) {
	text := string(content)
	var rules []Rule

	// Split by ## headers to get categories
	categoryRe := regexp.MustCompile(`(?m)^##\s+(.+)$`)
	matches := categoryRe.FindAllStringSubmatchIndex(text, -1)

	for i, match := range matches {
		categoryStart := match[2]
		categoryEnd := match[3]
		category := strings.TrimSpace(text[categoryStart:categoryEnd])

		// Get content between this category and the next
		contentStart := match[1]
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(text)
		}

		categoryContent := text[contentStart:contentEnd]
		categoryRules := extractRulesFromCategory(category, categoryContent)
		rules = append(rules, categoryRules...)
	}

	return rules, nil
}

// extractRulesFromCategory extracts individual rules from a category section
func extractRulesFromCategory(category, content string) []Rule {
	var rules []Rule

	// Look for ### sub-sections or bullet points
	lines := strings.Split(content, "\n")
	var currentRule *Rule

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Sub-section header defines a rule title
		if strings.HasPrefix(trimmed, "### ") {
			if currentRule != nil && currentRule.Title != "" {
				rules = append(rules, *currentRule)
			}
			currentRule = &Rule{
				Category: category,
				Title:    strings.TrimPrefix(trimmed, "### "),
			}
			continue
		}

		// Bullet points with bold title
		if strings.HasPrefix(trimmed, "- **") || strings.HasPrefix(trimmed, "* **") {
			if currentRule != nil && currentRule.Title != "" {
				rules = append(rules, *currentRule)
			}
			// Extract title between ** markers
			titleRe := regexp.MustCompile(`\*\*(.+?)\*\*`)
			if match := titleRe.FindStringSubmatch(trimmed); len(match) > 1 {
				currentRule = &Rule{
					Category:    category,
					Title:       match[1],
					Description: strings.TrimSpace(strings.Replace(trimmed, "**"+match[1]+"**", "", 1)),
				}
				// Clean up description
				currentRule.Description = strings.TrimPrefix(currentRule.Description, "- ")
				currentRule.Description = strings.TrimPrefix(currentRule.Description, "* ")
				currentRule.Description = strings.TrimPrefix(currentRule.Description, ": ")
				currentRule.Description = strings.TrimPrefix(currentRule.Description, "- ")
			}
			continue
		}

		// Accumulate description for current rule
		if currentRule != nil && trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			if currentRule.Description != "" {
				currentRule.Description += " " + trimmed
			} else {
				currentRule.Description = trimmed
			}
		}
	}

	// Don't forget the last rule
	if currentRule != nil && currentRule.Title != "" {
		rules = append(rules, *currentRule)
	}

	return rules
}

// ParseContextFile extracts decisions from CONTEXT.md
func ParseContextFile(content []byte) ([]Decision, error) {
	text := string(content)
	var decisions []Decision

	// Look for decision entries with dates
	// Common formats: "## YYYY-MM-DD: Summary" or "### Decision: Summary (YYYY-MM-DD)"
	dateHeaderRe := regexp.MustCompile(`(?m)^###?\s*(\d{4}-\d{2}-\d{2})[:\s]+(.+)$`)
	matches := dateHeaderRe.FindAllStringSubmatchIndex(text, -1)

	for i, match := range matches {
		dateStr := text[match[2]:match[3]]
		summary := strings.TrimSpace(text[match[4]:match[5]])

		date, _ := time.Parse("2006-01-02", dateStr)

		// Get content between this decision and the next
		contentStart := match[1]
		var contentEnd int
		if i+1 < len(matches) {
			contentEnd = matches[i+1][0]
		} else {
			contentEnd = len(text)
		}

		decisionContent := text[contentStart:contentEnd]

		decision := Decision{
			Date:      date,
			Summary:   summary,
			Rationale: extractSection(decisionContent, "Rationale", "-"),
			Impact:    extractSection(decisionContent, "Impact", "-"),
		}

		decisions = append(decisions, decision)
	}

	// Also look for bullet-style decisions
	// Format: "- **YYYY-MM-DD**: Summary"
	bulletRe := regexp.MustCompile(`(?m)^[-*]\s+\*\*(\d{4}-\d{2}-\d{2})\*\*[:\s]+(.+)$`)
	bulletMatches := bulletRe.FindAllStringSubmatch(text, -1)

	for _, match := range bulletMatches {
		dateStr := match[1]
		summary := strings.TrimSpace(match[2])
		date, _ := time.Parse("2006-01-02", dateStr)

		// Avoid duplicates
		isDuplicate := false
		for _, d := range decisions {
			if d.Date.Equal(date) && d.Summary == summary {
				isDuplicate = true
				break
			}
		}

		if !isDuplicate {
			decisions = append(decisions, Decision{
				Date:    date,
				Summary: summary,
			})
		}
	}

	return decisions, nil
}

// extractSection extracts content from a markdown section
func extractSection(text, sectionName, prefix string) string {
	// Build regex to match section header
	pattern := `(?mi)^` + regexp.QuoteMeta(prefix) + `\s*` + sectionName + `\s*$`
	re := regexp.MustCompile(pattern)

	loc := re.FindStringIndex(text)
	if loc == nil {
		return ""
	}

	// Find the end of the section (next heading or end of text)
	remainder := text[loc[1]:]
	endPattern := `(?m)^#`
	endRe := regexp.MustCompile(endPattern)
	endLoc := endRe.FindStringIndex(remainder)

	var content string
	if endLoc != nil {
		content = remainder[:endLoc[0]]
	} else {
		content = remainder
	}

	// Clean up
	content = strings.TrimSpace(content)
	return content
}

// extractChecklist extracts checklist items from markdown
func extractChecklist(text string) []string {
	var items []string
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Match "- [ ]" or "- [x]" patterns
		if strings.HasPrefix(line, "- [ ]") || strings.HasPrefix(line, "- [x]") ||
			strings.HasPrefix(line, "- [X]") {
			item := strings.TrimPrefix(line, "- [ ]")
			item = strings.TrimPrefix(item, "- [x]")
			item = strings.TrimPrefix(item, "- [X]")
			items = append(items, strings.TrimSpace(item))
		}
	}

	return items
}

// extractFrontmatterField extracts a field from YAML frontmatter
func extractFrontmatterField(text, field string) string {
	// Check if text starts with frontmatter
	if !strings.HasPrefix(strings.TrimSpace(text), "---") {
		return ""
	}

	// Find frontmatter boundaries
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return ""
	}

	frontmatter := parts[1]
	pattern := `(?m)^` + field + `:\s*(.+)$`
	re := regexp.MustCompile(pattern)

	if match := re.FindStringSubmatch(frontmatter); len(match) > 1 {
		return strings.TrimSpace(match[1])
	}

	return ""
}

// ParseTaskFile extracts task-specific context from a task markdown file
func ParseTaskFile(content []byte) (*TaskContext, error) {
	text := string(content)
	task := &TaskContext{}

	// Extract title
	titleRe := regexp.MustCompile(`(?m)^#\s+(?:Task:\s*)?(.+)$`)
	if match := titleRe.FindStringSubmatch(text); len(match) > 1 {
		task.Name = strings.TrimSpace(match[1])
	}

	// Extract objective
	task.Objective = extractSection(text, "Objective", "##")

	// Extract autonomy level from frontmatter or header
	task.AutonomyLevel = extractFrontmatterField(text, "autonomy_level")
	if task.AutonomyLevel == "" {
		// Try to find in header blockquote - handles **Autonomy Level**: value format
		autonomyRe := regexp.MustCompile(`(?i)Autonomy\s+Level\*{0,2}[:\s*]+(\w+)`)
		if match := autonomyRe.FindStringSubmatch(text); len(match) > 1 {
			task.AutonomyLevel = strings.ToLower(match[1])
		}
	}

	// Extract parallel execution flag - handles **Parallel Execution**: Yes format
	parallelRe := regexp.MustCompile(`(?i)Parallel\s+Execution\*{0,2}[:\s*]+(\w+)`)
	if match := parallelRe.FindStringSubmatch(text); len(match) > 1 {
		task.ParallelAllowed = strings.ToLower(match[1]) == "yes" || strings.ToLower(match[1]) == "true"
	}

	// Extract dependencies
	dependenciesSection := extractSection(text, "Dependencies", "##")
	if dependenciesSection != "" {
		task.Dependencies = extractBulletList(dependenciesSection)
	} else {
		// Try header blockquote format
		depsRe := regexp.MustCompile(`(?i)Dependencies[:\s]+([^\|]+)`)
		if match := depsRe.FindStringSubmatch(text); len(match) > 1 {
			deps := strings.TrimSpace(match[1])
			if deps != "None" && deps != "none" && deps != "" {
				task.Dependencies = strings.Split(deps, ",")
				for i := range task.Dependencies {
					task.Dependencies[i] = strings.TrimSpace(task.Dependencies[i])
				}
			}
		}
	}

	// Extract deliverables
	deliverablesSection := extractSection(text, "Deliverables", "##")
	if deliverablesSection != "" {
		task.Deliverables = extractBulletList(deliverablesSection)
	}

	return task, nil
}

// extractBulletList extracts items from a bullet list
func extractBulletList(text string) []string {
	var items []string
	scanner := bufio.NewScanner(strings.NewReader(text))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			item := strings.TrimPrefix(line, "- ")
			item = strings.TrimPrefix(item, "* ")
			items = append(items, strings.TrimSpace(item))
		}
	}

	return items
}
