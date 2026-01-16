package graduate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// topicMappingEntry defines how a topic maps to a sequence.
type topicMappingEntry struct {
	SequenceName string
	Priority     int
}

// TopicMapping defines how topics map to sequences.
var TopicMapping = map[string]topicMappingEntry{
	"requirements": {"core", 1},
	"architecture": {"core", 1},
	"core":         {"core", 1},
	"database":     {"core", 2},
	"models":       {"core", 2},
	"api":          {"api", 3},
	"endpoints":    {"api", 3},
	"integration":  {"integration", 4},
	"testing":      {"testing", 5},
	"deployment":   {"deployment", 6},
}

// TaskPriority defines ordering within sequences.
var TaskPriority = []string{
	"setup", "init", "config",
	"schema", "database", "migration",
	"model", "entity", "type",
	"service", "handler", "controller",
	"endpoint", "route", "api",
	"validation", "middleware",
	"integration", "connect",
	"test", "verify",
}

// Generator creates implementation structure from planning analysis.
type Generator struct {
	festivalPath string
}

// NewGenerator creates a new structure generator.
func NewGenerator(festivalPath string) *Generator {
	return &Generator{festivalPath: festivalPath}
}

// Generate creates a graduation plan from planning source.
func (g *Generator) Generate(ctx context.Context, source *PlanningSource) (*GraduationPlan, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	plan := &GraduationPlan{
		Source: *source,
	}

	// Determine target phase
	plan.Target = g.determineTarget(source)

	// Generate phase goal
	plan.PhaseGoal = g.generatePhaseGoal(source)

	// Map topics to sequences
	sequences, warnings := g.mapTopicsToSequences(source)
	plan.Sequences = sequences

	// Add collision warning if applicable
	if plan.Target.CollisionDetected {
		collisionWarning := fmt.Sprintf(
			"Phase %03d already exists (%s). Suggesting %03d instead.",
			plan.Target.OriginalNumber,
			plan.Target.ExistingPhase,
			plan.Target.Number,
		)
		warnings = append([]string{collisionWarning}, warnings...)
	}
	plan.Warnings = warnings

	// Calculate confidence
	plan.Confidence = g.calculateConfidence(source, sequences, warnings)

	return plan, nil
}

// determineTarget determines the implementation phase target.
// It checks for existing phases and adjusts the target number to avoid collisions.
func (g *Generator) determineTarget(source *PlanningSource) ImplementationTarget {
	// Extract phase number from source
	sourceNum := extractPhaseNumber(source.PhaseName)
	targetNum := sourceNum + 1

	// Scan for existing phases to detect collisions
	existingPhases := g.scanExistingPhases()

	// Check for collision and find next available number
	originalNum := targetNum
	existingName := ""
	collisionDetected := false

	if name, exists := existingPhases[targetNum]; exists {
		collisionDetected = true
		existingName = name
		// Find next available number
		targetNum = g.findNextAvailablePhaseNumber(existingPhases, targetNum)
	}

	targetName := fmt.Sprintf("%03d_IMPLEMENTATION", targetNum)

	return ImplementationTarget{
		Path:              filepath.Join(g.festivalPath, targetName),
		PhaseName:         targetName,
		Number:            targetNum,
		CollisionDetected: collisionDetected,
		OriginalNumber:    originalNum,
		ExistingPhase:     existingName,
	}
}

// scanExistingPhases scans the festival directory for existing phases.
// Returns a map of phase number -> phase name.
func (g *Generator) scanExistingPhases() map[int]string {
	phases := make(map[int]string)

	entries, err := os.ReadDir(g.festivalPath)
	if err != nil {
		return phases
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Skip hidden directories
		if strings.HasPrefix(name, ".") {
			continue
		}
		// Skip directories that don't start with a digit (not a phase)
		if len(name) == 0 || name[0] < '0' || name[0] > '9' {
			continue
		}

		// Extract phase number from directory name
		num := extractPhaseNumber(name)
		if num > 0 {
			phases[num] = name
		}
	}

	return phases
}

// findNextAvailablePhaseNumber finds the next available phase number starting from targetNum.
func (g *Generator) findNextAvailablePhaseNumber(existingPhases map[int]string, startNum int) int {
	nextNum := startNum
	for {
		if _, exists := existingPhases[nextNum]; !exists {
			return nextNum
		}
		nextNum++
		// Safety limit to prevent infinite loop
		if nextNum > 999 {
			return nextNum
		}
	}
}

// generatePhaseGoal creates content for PHASE_GOAL.md.
func (g *Generator) generatePhaseGoal(source *PlanningSource) GeneratedContent {
	goal := "Implement the system as planned"
	if source.Summary != nil && source.Summary.Goal != "" {
		goal = "Implement: " + source.Summary.Goal
	}

	content := GeneratedContent{
		Title: "Phase Goal: IMPLEMENTATION",
		Goal:  goal,
	}

	// Add planning reference section
	content.Sections = append(content.Sections,
		fmt.Sprintf("## Planning Reference\n\nGraduated from: `%s`", source.PhaseName))

	// Add key decisions if available
	if len(source.Decisions) > 0 {
		var decisionLines []string
		for _, d := range source.Decisions {
			decisionLines = append(decisionLines,
				fmt.Sprintf("- [%s] %s (%s)", d.ID, d.Title, d.Status))
		}
		content.Sections = append(content.Sections,
			"## Key Decisions\n\n"+strings.Join(decisionLines, "\n"))
	}

	return content
}

// mapTopicsToSequences converts planning topics to implementation sequences.
func (g *Generator) mapTopicsToSequences(source *PlanningSource) ([]ProposedSequence, []string) {
	var warnings []string

	// Group topics by target sequence
	sequenceMap := make(map[string]*ProposedSequence)
	var unmapped []string

	for _, topic := range source.TopicDirs {
		mapping, ok := TopicMapping[strings.ToLower(topic.Name)]
		if !ok {
			unmapped = append(unmapped, topic.Name)
			// Default: create sequence from topic name
			mapping = topicMappingEntry{
				SequenceName: topic.Name,
				Priority:     99,
			}
		}

		seqKey := mapping.SequenceName
		if seq, exists := sequenceMap[seqKey]; exists {
			// Add documents to existing sequence
			tasks := g.generateTasks(topic)
			seq.Tasks = append(seq.Tasks, tasks...)
		} else {
			// Create new sequence
			seq := &ProposedSequence{
				Name:        seqKey,
				SourceTopic: topic.Name,
				Goal: GeneratedContent{
					Title: fmt.Sprintf("Sequence Goal: %s", seqKey),
					Goal:  fmt.Sprintf("Implement %s functionality", seqKey),
				},
			}
			seq.Tasks = g.generateTasks(topic)
			sequenceMap[seqKey] = seq
		}
	}

	// Convert map to sorted slice
	sequences := make([]ProposedSequence, 0, len(sequenceMap))
	for _, seq := range sequenceMap {
		sequences = append(sequences, *seq)
	}

	// Sort by priority
	sort.Slice(sequences, func(i, j int) bool {
		pi := getSequencePriority(sequences[i].Name)
		pj := getSequencePriority(sequences[j].Name)
		return pi < pj
	})

	// Assign sequence numbers
	for i := range sequences {
		sequences[i].Number = i + 1
		sequences[i].FullName = fmt.Sprintf("%02d_%s", i+1, sequences[i].Name)

		// Re-number tasks
		for j := range sequences[i].Tasks {
			sequences[i].Tasks[j].Number = j + 1
			sequences[i].Tasks[j].FullName = fmt.Sprintf("%02d_%s.md",
				j+1, sequences[i].Tasks[j].Name)
		}
	}

	// Generate warnings
	if len(unmapped) > 0 {
		warnings = append(warnings,
			fmt.Sprintf("Topics not mapped to standard sequences: %s", strings.Join(unmapped, ", ")))
	}

	return sequences, warnings
}

// generateTasks creates tasks from topic documents.
func (g *Generator) generateTasks(topic TopicDirectory) []ProposedTask {
	var tasks []ProposedTask
	cleanRe := regexp.MustCompile(`[^a-z0-9]+`)

	for _, doc := range topic.Documents {
		taskName := strings.TrimSuffix(doc, ".md")
		taskName = strings.ToLower(taskName)
		taskName = cleanRe.ReplaceAllString(taskName, "_")
		taskName = strings.Trim(taskName, "_")

		// Try to extract objective from document content
		docPath := filepath.Join(topic.Path, doc)
		objective := g.extractObjectiveFromDoc(docPath, taskName)

		task := ProposedTask{
			Name:       taskName,
			Objective:  objective,
			SourceDocs: []string{filepath.Join(topic.Name, doc)},
		}
		tasks = append(tasks, task)
	}

	// Sort tasks by priority keywords
	sort.Slice(tasks, func(i, j int) bool {
		return getTaskPriority(tasks[i].Name) < getTaskPriority(tasks[j].Name)
	})

	return tasks
}

// calculateConfidence computes a confidence score for the plan.
func (g *Generator) calculateConfidence(source *PlanningSource, sequences []ProposedSequence, warnings []string) float64 {
	confidence := 1.0

	// Reduce confidence for warnings
	confidence -= float64(len(warnings)) * 0.1

	// Reduce if no planning summary
	if source.Summary == nil {
		confidence -= 0.1
	}

	// Reduce if no decisions documented
	if len(source.Decisions) == 0 {
		confidence -= 0.1
	}

	// Reduce if few documents
	if source.TotalDocs < 3 {
		confidence -= 0.2
	}

	// Clamp to 0.0-1.0
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 1 {
		confidence = 1
	}

	return confidence
}

// extractObjectiveFromDoc attempts to extract a meaningful objective from a planning document.
// It looks for sections like "Objective", "Goal", "Purpose", or extracts from the first
// paragraph or key bullet points. Falls back to a generic objective if extraction fails.
func (g *Generator) extractObjectiveFromDoc(docPath, taskName string) string {
	content, err := os.ReadFile(docPath)
	if err != nil {
		return fmt.Sprintf("Implement %s as specified in planning", taskName)
	}

	text := string(content)

	// Try to extract from common objective sections
	objectivePatterns := []string{
		// Match "## Objective" or "## Goal" section (first line after heading)
		`(?mi)^##\s*(?:Objective|Goal|Purpose)\s*\n+([^\n#]+)`,
		// Match "**Objective:**" or "**Goal:**" inline format
		`(?mi)\*\*(?:Objective|Goal|Purpose):\*\*\s*([^\n]+)`,
		// Match "> Objective:" or "> Goal:" blockquote format
		`(?m)^>\s*(?:Objective|Goal|Purpose):\s*([^\n]+)`,
	}

	for _, pattern := range objectivePatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			objective := strings.TrimSpace(matches[1])
			if objective != "" && len(objective) > 10 {
				return cleanObjective(objective)
			}
		}
	}

	// Try to extract from first heading (after the title)
	titleRe := regexp.MustCompile(`(?m)^#\s+([^\n]+)`)
	if matches := titleRe.FindStringSubmatch(text); len(matches) > 1 {
		title := strings.TrimSpace(matches[1])
		// If title contains a colon, use what's after it
		if idx := strings.Index(title, ":"); idx > 0 && idx < len(title)-1 {
			after := strings.TrimSpace(title[idx+1:])
			if len(after) > 10 {
				return cleanObjective(after)
			}
		}
		// Otherwise use the title directly if it's descriptive
		if len(title) > 15 && !strings.HasPrefix(strings.ToLower(title), "adr-") {
			return cleanObjective("Implement " + strings.ToLower(title))
		}
	}

	// Try to extract first meaningful bullet point from Requirements or Deliverables
	bulletPatterns := []string{
		`(?mi)^##\s*(?:Requirements|Deliverables|Features)\s*\n+\s*[-*]\s+([^\n]+)`,
	}

	for _, pattern := range bulletPatterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(text); len(matches) > 1 {
			bullet := strings.TrimSpace(matches[1])
			if bullet != "" && len(bullet) > 10 {
				return cleanObjective(bullet)
			}
		}
	}

	// Fallback to generic objective
	return fmt.Sprintf("Implement %s as specified in planning", taskName)
}

// cleanObjective cleans up an extracted objective string.
func cleanObjective(s string) string {
	// Remove markdown formatting
	s = regexp.MustCompile(`\*+`).ReplaceAllString(s, "")
	s = regexp.MustCompile("`+").ReplaceAllString(s, "")
	s = strings.TrimSpace(s)

	// Ensure it starts with a capital letter
	if len(s) > 0 && s[0] >= 'a' && s[0] <= 'z' {
		s = strings.ToUpper(string(s[0])) + s[1:]
	}

	// Truncate if too long
	if len(s) > 150 {
		s = s[:147] + "..."
	}

	return s
}

// Helper functions

func extractPhaseNumber(phaseName string) int {
	num := 0
	for _, c := range phaseName {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	if num == 0 {
		return 1 // Default
	}
	return num
}

func getSequencePriority(name string) int {
	if mapping, ok := TopicMapping[strings.ToLower(name)]; ok {
		return mapping.Priority
	}
	return 99
}

func getTaskPriority(name string) int {
	for i, keyword := range TaskPriority {
		if strings.Contains(name, keyword) {
			return i
		}
	}
	return len(TaskPriority)
}
