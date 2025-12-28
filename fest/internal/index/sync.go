package index

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	"github.com/lancekrogers/festival-methodology/fest/internal/gates"
)

// TreeSyncer builds a TreeIndex from the workspace filesystem
type TreeSyncer struct {
	workspacePath string
}

// NewTreeSyncer creates a new tree syncer
func NewTreeSyncer(workspacePath string) *TreeSyncer {
	return &TreeSyncer{
		workspacePath: workspacePath,
	}
}

// Sync builds a complete TreeIndex from the workspace
func (s *TreeSyncer) Sync() (*TreeIndex, error) {
	tree := NewTreeIndex(s.workspacePath)

	// Scan each status directory
	statusDirs := []string{"planned", "active", "completed", "dungeon"}
	for _, status := range statusDirs {
		statusPath := filepath.Join(s.workspacePath, status)
		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if !entry.IsDir() || entry.Name()[0] == '.' {
				continue
			}

			festivalPath := filepath.Join(statusPath, entry.Name())
			festNode, err := s.syncFestival(festivalPath, status)
			if err != nil {
				continue
			}

			tree.AddFestival(*festNode, status)
		}
	}

	tree.SortFestivals()
	tree.CalculateSummary()

	return tree, nil
}

// syncFestival syncs a single festival
func (s *TreeSyncer) syncFestival(festivalPath, status string) (*FestivalNode, error) {
	festivalName := filepath.Base(festivalPath)

	node := &FestivalNode{
		ID:     festivalName,
		Name:   toDisplayName(festivalName),
		Status: status,
		Path:   festivalPath,
		Phases: []PhaseNode{},
	}

	// Try to read fest_ref from fest.yaml or goal file
	node.Ref = s.extractRef(festivalPath)

	// Scan phases
	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return node, nil // Return partial node
	}

	for _, entry := range entries {
		if !entry.IsDir() || !festival.IsPhase(entry.Name()) {
			continue
		}

		phasePath := filepath.Join(festivalPath, entry.Name())
		phaseNode := s.syncPhase(phasePath, festival.ParsePhaseNumber(entry.Name()))
		node.Phases = append(node.Phases, *phaseNode)
		node.PhaseCount++

		// Aggregate task counts
		for _, seq := range phaseNode.Sequences {
			for _, task := range seq.Tasks {
				node.TaskCount++
				if task.Status == "completed" {
					node.CompletedTasks++
				}
			}
		}
	}

	// Calculate progress
	if node.TaskCount > 0 {
		node.Progress = float64(node.CompletedTasks) / float64(node.TaskCount)
	}

	return node, nil
}

// syncPhase syncs a single phase
func (s *TreeSyncer) syncPhase(phasePath string, order int) *PhaseNode {
	phaseName := filepath.Base(phasePath)

	node := &PhaseNode{
		ID:        phaseName,
		Name:      toDisplayName(phaseName),
		Order:     order,
		Status:    determineStatus(phasePath),
		Path:      phasePath,
		Sequences: []SequenceNode{},
	}

	// Try to read fest_ref from goal file
	goalPath := filepath.Join(phasePath, "PHASE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		if fm, _, parseErr := frontmatter.ParseFile(content); parseErr == nil && fm != nil {
			node.Ref = fm.Ref
		}
	}

	// Scan sequences
	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return node
	}

	for _, entry := range entries {
		if !entry.IsDir() || !festival.IsSequence(entry.Name()) {
			continue
		}

		seqPath := filepath.Join(phasePath, entry.Name())
		seqNode := s.syncSequence(seqPath, festival.ParseSequenceNumber(entry.Name()))
		node.Sequences = append(node.Sequences, *seqNode)
	}

	return node
}

// syncSequence syncs a single sequence
func (s *TreeSyncer) syncSequence(seqPath string, order int) *SequenceNode {
	seqName := filepath.Base(seqPath)

	node := &SequenceNode{
		ID:     seqName,
		Name:   toDisplayName(seqName),
		Order:  order,
		Status: determineStatus(seqPath),
		Path:   seqPath,
		Tasks:  []TaskNode{},
	}

	// Try to read fest_ref from goal file
	goalPath := filepath.Join(seqPath, "SEQUENCE_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		if fm, _, parseErr := frontmatter.ParseFile(content); parseErr == nil && fm != nil {
			node.Ref = fm.Ref
		}
	}

	// Scan tasks
	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return node
	}

	taskNum := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		if entry.Name() == "SEQUENCE_GOAL.md" || entry.Name() == "PHASE_GOAL.md" {
			continue
		}

		taskNum++
		taskPath := filepath.Join(seqPath, entry.Name())
		taskNode := s.syncTask(taskPath, taskNum)
		node.Tasks = append(node.Tasks, *taskNode)
	}

	return node
}

// syncTask syncs a single task
func (s *TreeSyncer) syncTask(taskPath string, order int) *TaskNode {
	taskName := filepath.Base(taskPath)

	node := &TaskNode{
		ID:     strings.TrimSuffix(taskName, ".md"),
		Name:   toDisplayName(strings.TrimSuffix(taskName, ".md")),
		Order:  order,
		Type:   "task",
		Status: "pending",
		Path:   taskPath,
	}

	// Try to read frontmatter
	if content, err := os.ReadFile(taskPath); err == nil {
		if fm, _, parseErr := frontmatter.ParseFile(content); parseErr == nil && fm != nil {
			node.Ref = fm.Ref
			if fm.Status != "" {
				node.Status = string(fm.Status)
			}
			if fm.Priority != "" {
				node.Autonomy = string(fm.Priority)
			}
		}
	}

	// Check if gate
	if gates.IsManaged(taskPath) {
		node.Type = "gate"
		node.GateType = gates.GetGateID(taskPath)
	}

	return node
}

// extractRef tries to extract a reference ID from the festival
func (s *TreeSyncer) extractRef(festivalPath string) string {
	// Try fest.yaml
	yamlPath := filepath.Join(festivalPath, "fest.yaml")
	if content, err := os.ReadFile(yamlPath); err == nil {
		if fm, _, parseErr := frontmatter.ParseFile(content); parseErr == nil && fm != nil && fm.Ref != "" {
			return fm.Ref
		}
	}

	// Try FESTIVAL_GOAL.md
	goalPath := filepath.Join(festivalPath, "FESTIVAL_GOAL.md")
	if content, err := os.ReadFile(goalPath); err == nil {
		if fm, _, parseErr := frontmatter.ParseFile(content); parseErr == nil && fm != nil && fm.Ref != "" {
			return fm.Ref
		}
	}

	return ""
}

// toDisplayName converts a filesystem name to a display name
func toDisplayName(name string) string {
	// Remove numeric prefix
	parts := strings.SplitN(name, "_", 2)
	if len(parts) == 2 {
		// Check if first part is numeric
		if _, err := filepath.Match("[0-9]*", parts[0]); err == nil {
			name = parts[1]
		}
	}

	// Replace underscores with spaces
	name = strings.ReplaceAll(name, "_", " ")

	// Title case
	return strings.Title(strings.ToLower(name))
}

// determineStatus determines the status of a phase/sequence based on its tasks
func determineStatus(path string) string {
	// For now, return "active" as default
	// A more sophisticated implementation would check task statuses
	return "active"
}
