// Package parser provides structured parsing of festival documents.
package parser

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
)

// ParseOptions controls what to include in parsed output
type ParseOptions struct {
	IncludeContent bool   // Include document content in output
	TypeFilter     string // Filter by entity type (task, gate, phase, etc.)
	Compact        bool   // Compact output (summary only)
	Format         string // Output format: json, yaml
	InferMissing   bool   // Infer frontmatter when missing
}

// ParsedEntity represents a generic parsed document
type ParsedEntity struct {
	Type        string                 `json:"type" yaml:"type"`
	ID          string                 `json:"id" yaml:"id"`
	Ref         string                 `json:"ref,omitempty" yaml:"ref,omitempty"`
	Name        string                 `json:"name" yaml:"name"`
	Status      string                 `json:"status" yaml:"status"`
	Path        string                 `json:"path" yaml:"path"`
	Order       int                    `json:"order,omitempty" yaml:"order,omitempty"`
	Frontmatter map[string]interface{} `json:"frontmatter,omitempty" yaml:"frontmatter,omitempty"`
	Content     string                 `json:"content,omitempty" yaml:"content,omitempty"`
}

// ParsedFestival represents a parsed festival with hierarchy
type ParsedFestival struct {
	ParsedEntity `yaml:",inline"`
	Phases       []ParsedPhase    `json:"phases,omitempty" yaml:"phases,omitempty"`
	Summary      *FestivalSummary `json:"summary,omitempty" yaml:"summary,omitempty"`
}

// FestivalSummary holds aggregate statistics
type FestivalSummary struct {
	PhaseCount     int     `json:"phase_count" yaml:"phase_count"`
	SequenceCount  int     `json:"sequence_count" yaml:"sequence_count"`
	TaskCount      int     `json:"task_count" yaml:"task_count"`
	GateCount      int     `json:"gate_count" yaml:"gate_count"`
	CompletedTasks int     `json:"completed_tasks" yaml:"completed_tasks"`
	Progress       float64 `json:"progress" yaml:"progress"`
}

// ParsedPhase represents a parsed phase with sequences
type ParsedPhase struct {
	ParsedEntity `yaml:",inline"`
	Sequences    []ParsedSequence `json:"sequences,omitempty" yaml:"sequences,omitempty"`
	Goal         *GoalContent     `json:"goal,omitempty" yaml:"goal,omitempty"`
}

// ParsedSequence represents a parsed sequence with tasks
type ParsedSequence struct {
	ParsedEntity `yaml:",inline"`
	Tasks        []ParsedTask `json:"tasks,omitempty" yaml:"tasks,omitempty"`
	Goal         *GoalContent `json:"goal,omitempty" yaml:"goal,omitempty"`
}

// ParsedTask represents a parsed task or gate
type ParsedTask struct {
	ParsedEntity `yaml:",inline"`
	Autonomy     string   `json:"autonomy,omitempty" yaml:"autonomy,omitempty"`
	GateType     string   `json:"gate_type,omitempty" yaml:"gate_type,omitempty"`
	Dependencies []string `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
}

// GoalContent represents extracted goal information
type GoalContent struct {
	Objective       string   `json:"objective,omitempty" yaml:"objective,omitempty"`
	SuccessCriteria []string `json:"success_criteria,omitempty" yaml:"success_criteria,omitempty"`
	Context         string   `json:"context,omitempty" yaml:"context,omitempty"`
}

// Parser handles parsing of festival structures
type Parser struct {
	opts ParseOptions
}

// NewParser creates a new parser with options
func NewParser(opts ParseOptions) *Parser {
	return &Parser{opts: opts}
}

// ParseFestival parses a complete festival from a path
func (p *Parser) ParseFestival(path string) (*ParsedFestival, error) {
	// Get festival info
	festivalDir := filepath.Clean(path)
	festivalName := filepath.Base(festivalDir)

	festival := &ParsedFestival{
		ParsedEntity: ParsedEntity{
			Type:   "festival",
			ID:     festivalName,
			Name:   strings.ReplaceAll(festivalName, "-", " "),
			Path:   festivalDir,
			Status: "active",
		},
		Summary: &FestivalSummary{},
	}

	// Parse FESTIVAL_GOAL.md if exists
	goalPath := filepath.Join(festivalDir, "FESTIVAL_GOAL.md")
	if _, err := os.Stat(goalPath); err == nil {
		if content, err := os.ReadFile(goalPath); err == nil {
			fm, _, _ := frontmatter.Parse(content)
			if fm != nil {
				festival.Status = string(fm.Status)
				festival.Ref = fm.Ref
			}
		}
	}

	// Parse fest.yaml for metadata
	festYaml := filepath.Join(festivalDir, "fest.yaml")
	if _, err := os.Stat(festYaml); err == nil {
		if content, err := os.ReadFile(festYaml); err == nil {
			fm, _, _ := frontmatter.Parse(content)
			if fm != nil && fm.Name != "" {
				festival.Name = fm.Name
			}
		}
	}

	// Enumerate and parse phases
	entries, err := os.ReadDir(festivalDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		// Look for numbered directories (phases)
		if !isNumberedDir(entry.Name()) {
			continue
		}

		phase, err := p.ParsePhase(filepath.Join(festivalDir, entry.Name()))
		if err != nil {
			continue
		}

		festival.Phases = append(festival.Phases, *phase)
		festival.Summary.PhaseCount++
		festival.Summary.SequenceCount += len(phase.Sequences)

		for _, seq := range phase.Sequences {
			for _, task := range seq.Tasks {
				festival.Summary.TaskCount++
				if task.Type == "gate" {
					festival.Summary.GateCount++
				}
				if task.Status == "completed" {
					festival.Summary.CompletedTasks++
				}
			}
		}
	}

	// Sort phases by order
	sort.Slice(festival.Phases, func(i, j int) bool {
		return festival.Phases[i].Order < festival.Phases[j].Order
	})

	// Calculate progress
	if festival.Summary.TaskCount > 0 {
		festival.Summary.Progress = float64(festival.Summary.CompletedTasks) / float64(festival.Summary.TaskCount)
	}

	return festival, nil
}

// ParsePhase parses a phase directory
func (p *Parser) ParsePhase(path string) (*ParsedPhase, error) {
	phaseName := filepath.Base(path)

	phase := &ParsedPhase{
		ParsedEntity: ParsedEntity{
			Type:   "phase",
			ID:     phaseName,
			Name:   extractName(phaseName),
			Path:   path,
			Status: "pending",
			Order:  extractOrder(phaseName),
		},
	}

	// Parse PHASE_GOAL.md
	goalPath := filepath.Join(path, "PHASE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		fm, body, _ := frontmatter.Parse(content)
		if fm != nil {
			phase.Status = string(fm.Status)
			phase.Ref = fm.Ref
		}
		phase.Goal = parseGoalContent(string(body))
	}

	// Enumerate sequences
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		if !isNumberedDir(entry.Name()) {
			continue
		}

		seq, err := p.ParseSequence(filepath.Join(path, entry.Name()))
		if err != nil {
			continue
		}
		phase.Sequences = append(phase.Sequences, *seq)
	}

	// Sort sequences by order
	sort.Slice(phase.Sequences, func(i, j int) bool {
		return phase.Sequences[i].Order < phase.Sequences[j].Order
	})

	return phase, nil
}

// ParseSequence parses a sequence directory
func (p *Parser) ParseSequence(path string) (*ParsedSequence, error) {
	seqName := filepath.Base(path)

	seq := &ParsedSequence{
		ParsedEntity: ParsedEntity{
			Type:   "sequence",
			ID:     seqName,
			Name:   extractName(seqName),
			Path:   path,
			Status: "pending",
			Order:  extractOrder(seqName),
		},
	}

	// Parse SEQUENCE_GOAL.md
	goalPath := filepath.Join(path, "SEQUENCE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		fm, body, _ := frontmatter.Parse(content)
		if fm != nil {
			seq.Status = string(fm.Status)
			seq.Ref = fm.Ref
		}
		seq.Goal = parseGoalContent(string(body))
	}

	// Enumerate tasks
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if strings.HasPrefix(strings.ToUpper(entry.Name()), "SEQUENCE_") {
			continue
		}

		task, err := p.ParseTask(filepath.Join(path, entry.Name()))
		if err != nil {
			continue
		}

		// Apply type filter
		if p.opts.TypeFilter != "" && task.Type != p.opts.TypeFilter {
			continue
		}

		seq.Tasks = append(seq.Tasks, *task)
	}

	// Sort tasks by order
	sort.Slice(seq.Tasks, func(i, j int) bool {
		return seq.Tasks[i].Order < seq.Tasks[j].Order
	})

	return seq, nil
}

// ParseTask parses a task file
func (p *Parser) ParseTask(path string) (*ParsedTask, error) {
	fileName := filepath.Base(path)
	taskName := strings.TrimSuffix(fileName, ".md")

	task := &ParsedTask{
		ParsedEntity: ParsedEntity{
			Type:   "task",
			ID:     taskName,
			Name:   extractName(taskName),
			Path:   path,
			Status: "pending",
			Order:  extractOrder(taskName),
		},
	}

	// Read file
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// Parse frontmatter
	fm, body, _ := frontmatter.Parse(content)
	if fm != nil {
		task.Status = string(fm.Status)
		task.Ref = fm.Ref
		task.Autonomy = string(fm.Autonomy)
		task.GateType = string(fm.GateType)
		if fm.Type == frontmatter.TypeGate {
			task.Type = "gate"
		}
	} else if p.opts.InferMissing {
		// Infer from path
		inferred, _ := frontmatter.InferFromPath(path)
		if inferred != nil {
			if inferred.Type == frontmatter.TypeGate {
				task.Type = "gate"
				task.GateType = string(inferred.GateType)
			}
		}
	}

	// Detect gate from filename if not already set
	if task.Type != "gate" && isGateFile(fileName) {
		task.Type = "gate"
	}

	// Include content if requested
	if p.opts.IncludeContent {
		task.Content = string(body)
	}

	return task, nil
}

// Helper functions

func isNumberedDir(name string) bool {
	if len(name) < 2 {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

func extractOrder(name string) int {
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
	var order int
	for _, c := range numStr {
		order = order*10 + int(c-'0')
	}
	return order
}

func extractName(name string) string {
	// Remove numeric prefix
	for i, c := range name {
		if c < '0' || c > '9' {
			if c == '_' || c == '-' {
				name = name[i+1:]
			} else {
				name = name[i:]
			}
			break
		}
	}
	// Convert underscores to spaces
	name = strings.ReplaceAll(name, "_", " ")
	return strings.TrimSpace(name)
}

func isGateFile(name string) bool {
	lower := strings.ToLower(name)
	indicators := []string{"quality_gate", "gate_", "_gate", "testing_and_verify", "review_gate"}
	for _, ind := range indicators {
		if strings.Contains(lower, ind) {
			return true
		}
	}
	return false
}

func parseGoalContent(content string) *GoalContent {
	goal := &GoalContent{}

	lines := strings.Split(content, "\n")
	var currentSection string
	var sectionContent strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check for section headers
		if strings.HasPrefix(trimmed, "## ") {
			// Save previous section
			if currentSection != "" {
				text := strings.TrimSpace(sectionContent.String())
				switch currentSection {
				case "objective":
					goal.Objective = text
				case "context":
					goal.Context = text
				}
			}

			// Start new section
			header := strings.ToLower(strings.TrimPrefix(trimmed, "## "))
			switch {
			case strings.Contains(header, "objective"):
				currentSection = "objective"
			case strings.Contains(header, "context"):
				currentSection = "context"
			case strings.Contains(header, "success"):
				currentSection = "success"
			default:
				currentSection = ""
			}
			sectionContent.Reset()
			continue
		}

		// Collect section content
		if currentSection != "" {
			if currentSection == "success" {
				// Parse checkboxes as success criteria
				if strings.HasPrefix(trimmed, "- [") && len(trimmed) > 4 {
					criterion := strings.TrimPrefix(trimmed[4:], " ")
					criterion = strings.TrimPrefix(criterion, "] ")
					if criterion != "" {
						goal.SuccessCriteria = append(goal.SuccessCriteria, criterion)
					}
				}
			} else {
				sectionContent.WriteString(line + "\n")
			}
		}
	}

	// Save final section
	if currentSection != "" && currentSection != "success" {
		text := strings.TrimSpace(sectionContent.String())
		switch currentSection {
		case "objective":
			goal.Objective = text
		case "context":
			goal.Context = text
		}
	}

	return goal
}
