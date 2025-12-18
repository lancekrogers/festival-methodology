// Package index provides festival structure indexing for Guild integration.
package index

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

const (
	// IndexFileName is the name of the index file
	IndexFileName = "index.json"
	// CurrentSpecVersion is the current index specification version
	CurrentSpecVersion = 1
)

// FestivalIndex is the root index structure
type FestivalIndex struct {
	FestSpec    int          `json:"fest_spec"`
	FestivalID  string       `json:"festival_id"`
	GeneratedAt time.Time    `json:"generated_at"`
	Phases      []PhaseIndex `json:"phases"`
}

// PhaseIndex represents a phase in the index
type PhaseIndex struct {
	PhaseID   string          `json:"phase_id"`
	Path      string          `json:"path"`
	GoalFile  string          `json:"goal_file,omitempty"`
	Sequences []SequenceIndex `json:"sequences"`
}

// SequenceIndex represents a sequence in the index
type SequenceIndex struct {
	SequenceID   string      `json:"sequence_id"`
	Path         string      `json:"path"`
	GoalFile     string      `json:"goal_file,omitempty"`
	Tasks        []TaskIndex `json:"tasks"`
	ManagedGates []string    `json:"managed_gates,omitempty"`
}

// TaskIndex represents a task in the index
type TaskIndex struct {
	TaskID  string `json:"task_id"`
	Path    string `json:"path"`
	Managed bool   `json:"managed"`
	GateID  string `json:"gate_id,omitempty"`
}

// NewFestivalIndex creates a new empty festival index
func NewFestivalIndex(festivalID string) *FestivalIndex {
	return &FestivalIndex{
		FestSpec:    CurrentSpecVersion,
		FestivalID:  festivalID,
		GeneratedAt: time.Now().UTC(),
		Phases:      []PhaseIndex{},
	}
}

// LoadIndex loads an index from a file
func LoadIndex(path string) (*FestivalIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read index: %w", err)
	}

	var index FestivalIndex
	if err := json.Unmarshal(data, &index); err != nil {
		return nil, fmt.Errorf("failed to parse index: %w", err)
	}

	return &index, nil
}

// Save saves the index to a file
func (idx *FestivalIndex) Save(path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal index: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write index: %w", err)
	}

	return nil
}

// AddPhase adds a phase to the index
func (idx *FestivalIndex) AddPhase(phase PhaseIndex) {
	idx.Phases = append(idx.Phases, phase)
}

// GetPhase returns a phase by ID
func (idx *FestivalIndex) GetPhase(phaseID string) *PhaseIndex {
	for i := range idx.Phases {
		if idx.Phases[i].PhaseID == phaseID {
			return &idx.Phases[i]
		}
	}
	return nil
}

// AddSequence adds a sequence to a phase
func (p *PhaseIndex) AddSequence(seq SequenceIndex) {
	p.Sequences = append(p.Sequences, seq)
}

// GetSequence returns a sequence by ID
func (p *PhaseIndex) GetSequence(seqID string) *SequenceIndex {
	for i := range p.Sequences {
		if p.Sequences[i].SequenceID == seqID {
			return &p.Sequences[i]
		}
	}
	return nil
}

// AddTask adds a task to a sequence
func (s *SequenceIndex) AddTask(task TaskIndex) {
	s.Tasks = append(s.Tasks, task)
}

// GetTask returns a task by ID
func (s *SequenceIndex) GetTask(taskID string) *TaskIndex {
	for i := range s.Tasks {
		if s.Tasks[i].TaskID == taskID {
			return &s.Tasks[i]
		}
	}
	return nil
}

// Summary returns a summary of the index
type IndexSummary struct {
	PhaseCount    int `json:"phase_count"`
	SequenceCount int `json:"sequence_count"`
	TaskCount     int `json:"task_count"`
	ManagedCount  int `json:"managed_count"`
}

// Summary returns a summary of the index
func (idx *FestivalIndex) Summary() IndexSummary {
	summary := IndexSummary{
		PhaseCount: len(idx.Phases),
	}

	for _, phase := range idx.Phases {
		summary.SequenceCount += len(phase.Sequences)
		for _, seq := range phase.Sequences {
			summary.TaskCount += len(seq.Tasks)
			for _, task := range seq.Tasks {
				if task.Managed {
					summary.ManagedCount++
				}
			}
		}
	}

	return summary
}
