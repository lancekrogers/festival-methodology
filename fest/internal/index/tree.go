package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// TreeIndex represents the complete workspace tree for Guild v3
type TreeIndex struct {
	Version   string           `json:"version"`
	IndexedAt time.Time        `json:"indexed_at"`
	Workspace WorkspaceInfo    `json:"workspace"`
	Festivals FestivalsByStatus `json:"festivals"`
}

// WorkspaceInfo contains workspace metadata
type WorkspaceInfo struct {
	Path           string `json:"path"`
	FestivalCount  int    `json:"festival_count"`
	TotalTasks     int    `json:"total_tasks"`
	CompletedTasks int    `json:"completed_tasks"`
}

// FestivalsByStatus groups festivals by their status
type FestivalsByStatus struct {
	Planned   []FestivalNode `json:"planned"`
	Active    []FestivalNode `json:"active"`
	Completed []FestivalNode `json:"completed"`
	Dungeon   []FestivalNode `json:"dungeon"`
}

// FestivalNode represents a festival in the tree
type FestivalNode struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Ref            string      `json:"ref,omitempty"`
	Status         string      `json:"status"`
	Path           string      `json:"path"`
	PhaseCount     int         `json:"phase_count"`
	TaskCount      int         `json:"task_count"`
	CompletedTasks int         `json:"completed_tasks"`
	Progress       float64     `json:"progress"`
	Phases         []PhaseNode `json:"phases,omitempty"`
}

// PhaseNode represents a phase in the tree
type PhaseNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Ref       string         `json:"ref,omitempty"`
	Order     int            `json:"order"`
	Status    string         `json:"status"`
	Path      string         `json:"path"`
	Sequences []SequenceNode `json:"sequences,omitempty"`
}

// SequenceNode represents a sequence in the tree
type SequenceNode struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Ref    string     `json:"ref,omitempty"`
	Order  int        `json:"order"`
	Status string     `json:"status"`
	Path   string     `json:"path"`
	Tasks  []TaskNode `json:"tasks,omitempty"`
}

// TaskNode represents a task in the tree
type TaskNode struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Ref      string `json:"ref,omitempty"`
	Order    int    `json:"order"`
	Type     string `json:"type"` // task or gate
	Status   string `json:"status"`
	Path     string `json:"path"`
	Autonomy string `json:"autonomy,omitempty"`
	GateType string `json:"gate_type,omitempty"`
}

// NewTreeIndex creates a new tree index
func NewTreeIndex(workspacePath string) *TreeIndex {
	return &TreeIndex{
		Version:   "1.0",
		IndexedAt: time.Now().UTC(),
		Workspace: WorkspaceInfo{
			Path: workspacePath,
		},
		Festivals: FestivalsByStatus{},
	}
}

// AddFestival adds a festival to the appropriate status group
func (t *TreeIndex) AddFestival(festival FestivalNode, status string) {
	// Ensure status is set on the node
	festival.Status = status

	switch status {
	case "planned":
		t.Festivals.Planned = append(t.Festivals.Planned, festival)
	case "active":
		t.Festivals.Active = append(t.Festivals.Active, festival)
	case "completed":
		t.Festivals.Completed = append(t.Festivals.Completed, festival)
	case "dungeon":
		t.Festivals.Dungeon = append(t.Festivals.Dungeon, festival)
	default:
		// Default to active
		festival.Status = "active"
		t.Festivals.Active = append(t.Festivals.Active, festival)
	}
}

// GetAllFestivals returns all festivals across all statuses
func (t *TreeIndex) GetAllFestivals() []FestivalNode {
	var all []FestivalNode
	all = append(all, t.Festivals.Planned...)
	all = append(all, t.Festivals.Active...)
	all = append(all, t.Festivals.Completed...)
	all = append(all, t.Festivals.Dungeon...)
	return all
}

// GetFestivalByID finds a festival by ID
func (t *TreeIndex) GetFestivalByID(id string) *FestivalNode {
	for i := range t.Festivals.Active {
		if t.Festivals.Active[i].ID == id {
			return &t.Festivals.Active[i]
		}
	}
	for i := range t.Festivals.Planned {
		if t.Festivals.Planned[i].ID == id {
			return &t.Festivals.Planned[i]
		}
	}
	for i := range t.Festivals.Completed {
		if t.Festivals.Completed[i].ID == id {
			return &t.Festivals.Completed[i]
		}
	}
	for i := range t.Festivals.Dungeon {
		if t.Festivals.Dungeon[i].ID == id {
			return &t.Festivals.Dungeon[i]
		}
	}
	return nil
}

// GetFestivalByRef finds a festival by reference ID
func (t *TreeIndex) GetFestivalByRef(ref string) *FestivalNode {
	for _, f := range t.GetAllFestivals() {
		if f.Ref == ref {
			return &f
		}
	}
	return nil
}

// CalculateSummary updates workspace summary statistics
func (t *TreeIndex) CalculateSummary() {
	t.Workspace.FestivalCount = 0
	t.Workspace.TotalTasks = 0
	t.Workspace.CompletedTasks = 0

	for _, f := range t.GetAllFestivals() {
		t.Workspace.FestivalCount++
		t.Workspace.TotalTasks += f.TaskCount
		t.Workspace.CompletedTasks += f.CompletedTasks
	}
}

// Save saves the tree index to a file
func (t *TreeIndex) Save(path string) error {
	t.CalculateSummary()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// LoadTreeIndex loads a tree index from a file
func LoadTreeIndex(path string) (*TreeIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var index TreeIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, err
	}

	return &index, nil
}

// SortFestivals sorts festivals by name within each status group
func (t *TreeIndex) SortFestivals() {
	sortFn := func(a, b FestivalNode) bool {
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	}
	sort.Slice(t.Festivals.Planned, func(i, j int) bool {
		return sortFn(t.Festivals.Planned[i], t.Festivals.Planned[j])
	})
	sort.Slice(t.Festivals.Active, func(i, j int) bool {
		return sortFn(t.Festivals.Active[i], t.Festivals.Active[j])
	})
	sort.Slice(t.Festivals.Completed, func(i, j int) bool {
		return sortFn(t.Festivals.Completed[i], t.Festivals.Completed[j])
	})
	sort.Slice(t.Festivals.Dungeon, func(i, j int) bool {
		return sortFn(t.Festivals.Dungeon[i], t.Festivals.Dungeon[j])
	})
}
