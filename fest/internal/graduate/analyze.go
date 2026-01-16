package graduate

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Analyzer reads and analyzes planning phases.
type Analyzer struct {
	festivalPath string
}

// NewAnalyzer creates a new planning analyzer.
func NewAnalyzer(festivalPath string) *Analyzer {
	return &Analyzer{festivalPath: festivalPath}
}

// Analyze reads a planning phase and extracts structured information.
func (a *Analyzer) Analyze(ctx context.Context, phasePath string) (*PlanningSource, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Verify path exists
	info, err := os.Stat(phasePath)
	if err != nil {
		return nil, errors.NotFound("planning phase").WithField("path", phasePath)
	}
	if !info.IsDir() {
		return nil, errors.Validation("path is not a directory").WithField("path", phasePath)
	}

	source := &PlanningSource{
		Path:       phasePath,
		PhaseName:  filepath.Base(phasePath),
		AnalyzedAt: time.Now(),
	}

	// Scan topic directories
	topics, totalDocs, err := a.scanTopics(phasePath)
	if err != nil {
		return nil, errors.Wrap(err, "scanning topics")
	}
	source.TopicDirs = topics
	source.TotalDocs = totalDocs

	// Parse decisions
	decisions, err := a.parseDecisions(phasePath)
	if err == nil && len(decisions) > 0 {
		source.Decisions = decisions
	}

	// Parse planning summary
	summary, err := a.parsePlanningSummary(phasePath)
	if err == nil && summary != nil {
		source.Summary = summary
	}

	return source, nil
}

// scanTopics finds all topic directories and their documents.
func (a *Analyzer) scanTopics(phasePath string) ([]TopicDirectory, int, error) {
	var topics []TopicDirectory
	var totalDocs int

	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, 0, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Skip hidden directories
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		topicPath := filepath.Join(phasePath, entry.Name())
		docs, err := a.findDocuments(topicPath)
		if err != nil {
			continue // Skip directories we can't read
		}

		topic := TopicDirectory{
			Name:      entry.Name(),
			Path:      topicPath,
			Documents: docs,
			DocCount:  len(docs),
		}
		topics = append(topics, topic)
		totalDocs += len(docs)
	}

	return topics, totalDocs, nil
}

// findDocuments finds all markdown documents in a directory.
// Excludes goal files (PHASE_GOAL.md, SEQUENCE_GOAL.md, *_GOAL.md) as these
// are metadata files, not planning documents.
func (a *Analyzer) findDocuments(dirPath string) ([]string, error) {
	var docs []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".md") {
			continue
		}
		// Filter out goal files - these are metadata, not planning content
		if isGoalFile(name) {
			continue
		}
		docs = append(docs, name)
	}

	return docs, nil
}

// isGoalFile returns true if the filename matches a goal file pattern.
// Goal files include PHASE_GOAL.md, SEQUENCE_GOAL.md, FESTIVAL_GOAL.md,
// and any file ending in _GOAL.md.
func isGoalFile(filename string) bool {
	upper := strings.ToUpper(filename)
	// Exact matches
	if upper == "PHASE_GOAL.MD" || upper == "SEQUENCE_GOAL.MD" || upper == "FESTIVAL_GOAL.MD" {
		return true
	}
	// Pattern match: *_GOAL.md
	if strings.HasSuffix(upper, "_GOAL.MD") {
		return true
	}
	return false
}

// parseDecisions looks for ADRs in decisions/ directory.
func (a *Analyzer) parseDecisions(phasePath string) ([]Decision, error) {
	var decisions []Decision

	// Check for decisions directory
	decisionsDir := filepath.Join(phasePath, "decisions")
	info, err := os.Stat(decisionsDir)
	if err != nil || !info.IsDir() {
		return nil, nil // No decisions directory
	}

	entries, err := os.ReadDir(decisionsDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		docPath := filepath.Join(decisionsDir, entry.Name())
		decision, err := a.parseDecisionFile(docPath)
		if err == nil {
			decisions = append(decisions, *decision)
		}
	}

	return decisions, nil
}

// parseDecisionFile parses an ADR document.
func (a *Analyzer) parseDecisionFile(filePath string) (*Decision, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	text := string(content)
	decision := &Decision{
		FilePath: filePath,
	}

	// Extract ID from filename (e.g., ADR-001-database.md)
	base := filepath.Base(filePath)
	idRe := regexp.MustCompile(`^(ADR-\d+)`)
	if matches := idRe.FindStringSubmatch(base); len(matches) > 1 {
		decision.ID = matches[1]
	}

	// Extract title from first # heading
	titleRe := regexp.MustCompile(`(?m)^#\s+(.+)$`)
	if matches := titleRe.FindStringSubmatch(text); len(matches) > 1 {
		decision.Title = strings.TrimSpace(matches[1])
	}

	// Extract status from ## Status section
	statusRe := regexp.MustCompile(`(?mi)^##\s*Status\s*\n+([^\n]+)`)
	if matches := statusRe.FindStringSubmatch(text); len(matches) > 1 {
		decision.Status = strings.ToLower(strings.TrimSpace(matches[1]))
	}

	// Extract summary from ## Context or ## Decision section
	contextRe := regexp.MustCompile(`(?mi)^##\s*(?:Context|Decision)\s*\n+([^\n]+)`)
	if matches := contextRe.FindStringSubmatch(text); len(matches) > 1 {
		decision.Summary = strings.TrimSpace(matches[1])
	}

	return decision, nil
}

// parsePlanningSummary reads PLANNING_SUMMARY.md.
func (a *Analyzer) parsePlanningSummary(phasePath string) (*PlanningSummary, error) {
	summaryPath := filepath.Join(phasePath, "PLANNING_SUMMARY.md")
	content, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil, err // File doesn't exist, not an error
	}

	text := string(content)
	summary := &PlanningSummary{
		FilePath: summaryPath,
	}

	// Extract goal from Primary Goal or first paragraph
	goalRe := regexp.MustCompile(`(?mi)\*\*Primary Goal:\*\*\s*(.+)`)
	if matches := goalRe.FindStringSubmatch(text); len(matches) > 1 {
		summary.Goal = strings.TrimSpace(matches[1])
	}

	// Extract key decisions (bullet list after "Key Decisions" heading)
	decisionsRe := regexp.MustCompile(`(?mi)^##\s*Key Decisions\s*\n((?:\s*[-*]\s+.+\n)+)`)
	if matches := decisionsRe.FindStringSubmatch(text); len(matches) > 1 {
		lines := strings.Split(matches[1], "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
				summary.KeyDecisions = append(summary.KeyDecisions,
					strings.TrimSpace(strings.TrimLeft(line, "-* ")))
			}
		}
	}

	// Extract proposed sequences (numbered list or bullet points)
	seqRe := regexp.MustCompile(`(?mi)^##\s*(?:Proposed )?(?:Implementation )?Sequences?\s*\n((?:\s*(?:\d+\.|[-*])\s+.+\n)+)`)
	if matches := seqRe.FindStringSubmatch(text); len(matches) > 1 {
		lines := strings.Split(matches[1], "\n")
		prefixRe := regexp.MustCompile(`^(?:\d+\.|[-*])\s*`)
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" {
				// Remove bullet or number prefix
				line = prefixRe.ReplaceAllString(line, "")
				summary.ProposedSequences = append(summary.ProposedSequences, strings.TrimSpace(line))
			}
		}
	}

	return summary, nil
}

// FindPlanningPhase finds a planning phase in the festival.
func (a *Analyzer) FindPlanningPhase(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	entries, err := os.ReadDir(a.festivalPath)
	if err != nil {
		return "", errors.Wrap(err, "reading festival directory")
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.Contains(name, "planning") || strings.Contains(name, "design") {
			return filepath.Join(a.festivalPath, entry.Name()), nil
		}
	}

	return "", errors.NotFound("planning phase")
}
