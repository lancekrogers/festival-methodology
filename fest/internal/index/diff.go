package index

import (
	"time"
)

// DiffType represents the type of change detected
type DiffType string

const (
	DiffAdded    DiffType = "added"
	DiffRemoved  DiffType = "removed"
	DiffModified DiffType = "modified"
	DiffMoved    DiffType = "moved"
)

// TreeDiff represents differences between two tree indexes
type TreeDiff struct {
	OldVersion string       `json:"old_version"`
	NewVersion string       `json:"new_version"`
	Timestamp  time.Time    `json:"timestamp"`
	Changes    []DiffChange `json:"changes"`
	Summary    DiffSummary  `json:"summary"`
}

// DiffChange represents a single change
type DiffChange struct {
	Type       DiffType `json:"type"`
	EntityType string   `json:"entity_type"` // festival, phase, sequence, task
	EntityID   string   `json:"entity_id"`
	EntityRef  string   `json:"entity_ref,omitempty"`
	OldPath    string   `json:"old_path,omitempty"`
	NewPath    string   `json:"new_path,omitempty"`
	OldStatus  string   `json:"old_status,omitempty"`
	NewStatus  string   `json:"new_status,omitempty"`
	Details    string   `json:"details,omitempty"`
}

// DiffSummary summarizes the changes
type DiffSummary struct {
	FestivalsAdded    int `json:"festivals_added"`
	FestivalsRemoved  int `json:"festivals_removed"`
	FestivalsModified int `json:"festivals_modified"`
	FestivalsMoved    int `json:"festivals_moved"`
	PhasesAdded       int `json:"phases_added"`
	PhasesRemoved     int `json:"phases_removed"`
	SequencesAdded    int `json:"sequences_added"`
	SequencesRemoved  int `json:"sequences_removed"`
	TasksAdded        int `json:"tasks_added"`
	TasksRemoved      int `json:"tasks_removed"`
	TasksCompleted    int `json:"tasks_completed"`
}

// ComputeDiff computes the differences between two tree indexes
func ComputeDiff(old, new *TreeIndex) *TreeDiff {
	diff := &TreeDiff{
		OldVersion: old.Version,
		NewVersion: new.Version,
		Timestamp:  time.Now().UTC(),
		Changes:    []DiffChange{},
	}

	// Build lookup maps
	oldFestivals := make(map[string]*FestivalNode)
	newFestivals := make(map[string]*FestivalNode)

	for _, f := range old.GetAllFestivals() {
		oldFestivals[f.ID] = &f
	}
	for _, f := range new.GetAllFestivals() {
		newFestivals[f.ID] = &f
	}

	// Find added and modified festivals
	for id, newF := range newFestivals {
		oldF, exists := oldFestivals[id]
		if !exists {
			diff.Changes = append(diff.Changes, DiffChange{
				Type:       DiffAdded,
				EntityType: "festival",
				EntityID:   id,
				EntityRef:  newF.Ref,
				NewPath:    newF.Path,
				NewStatus:  newF.Status,
			})
			diff.Summary.FestivalsAdded++
			continue
		}

		// Check if moved (status changed)
		if oldF.Status != newF.Status {
			diff.Changes = append(diff.Changes, DiffChange{
				Type:       DiffMoved,
				EntityType: "festival",
				EntityID:   id,
				EntityRef:  newF.Ref,
				OldPath:    oldF.Path,
				NewPath:    newF.Path,
				OldStatus:  oldF.Status,
				NewStatus:  newF.Status,
			})
			diff.Summary.FestivalsMoved++
		}

		// Compare phases
		diff.compareFestivalContents(oldF, newF)
	}

	// Find removed festivals
	for id, oldF := range oldFestivals {
		if _, exists := newFestivals[id]; !exists {
			diff.Changes = append(diff.Changes, DiffChange{
				Type:       DiffRemoved,
				EntityType: "festival",
				EntityID:   id,
				EntityRef:  oldF.Ref,
				OldPath:    oldF.Path,
				OldStatus:  oldF.Status,
			})
			diff.Summary.FestivalsRemoved++
		}
	}

	return diff
}

// compareFestivalContents compares the contents of two festivals
func (d *TreeDiff) compareFestivalContents(old, new *FestivalNode) {
	oldPhases := make(map[string]*PhaseNode)
	newPhases := make(map[string]*PhaseNode)

	for i := range old.Phases {
		oldPhases[old.Phases[i].ID] = &old.Phases[i]
	}
	for i := range new.Phases {
		newPhases[new.Phases[i].ID] = &new.Phases[i]
	}

	// Find added phases
	for id, newP := range newPhases {
		if _, exists := oldPhases[id]; !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffAdded,
				EntityType: "phase",
				EntityID:   id,
				EntityRef:  newP.Ref,
				NewPath:    newP.Path,
			})
			d.Summary.PhasesAdded++
		} else {
			// Compare sequences
			d.comparePhaseContents(oldPhases[id], newP)
		}
	}

	// Find removed phases
	for id, oldP := range oldPhases {
		if _, exists := newPhases[id]; !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffRemoved,
				EntityType: "phase",
				EntityID:   id,
				EntityRef:  oldP.Ref,
				OldPath:    oldP.Path,
			})
			d.Summary.PhasesRemoved++
		}
	}
}

// comparePhaseContents compares the contents of two phases
func (d *TreeDiff) comparePhaseContents(old, new *PhaseNode) {
	oldSeqs := make(map[string]*SequenceNode)
	newSeqs := make(map[string]*SequenceNode)

	for i := range old.Sequences {
		oldSeqs[old.Sequences[i].ID] = &old.Sequences[i]
	}
	for i := range new.Sequences {
		newSeqs[new.Sequences[i].ID] = &new.Sequences[i]
	}

	// Find added sequences
	for id, newS := range newSeqs {
		if _, exists := oldSeqs[id]; !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffAdded,
				EntityType: "sequence",
				EntityID:   id,
				EntityRef:  newS.Ref,
				NewPath:    newS.Path,
			})
			d.Summary.SequencesAdded++
		} else {
			// Compare tasks
			d.compareSequenceContents(oldSeqs[id], newS)
		}
	}

	// Find removed sequences
	for id, oldS := range oldSeqs {
		if _, exists := newSeqs[id]; !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffRemoved,
				EntityType: "sequence",
				EntityID:   id,
				EntityRef:  oldS.Ref,
				OldPath:    oldS.Path,
			})
			d.Summary.SequencesRemoved++
		}
	}
}

// compareSequenceContents compares the contents of two sequences
func (d *TreeDiff) compareSequenceContents(old, new *SequenceNode) {
	oldTasks := make(map[string]*TaskNode)
	newTasks := make(map[string]*TaskNode)

	for i := range old.Tasks {
		oldTasks[old.Tasks[i].ID] = &old.Tasks[i]
	}
	for i := range new.Tasks {
		newTasks[new.Tasks[i].ID] = &new.Tasks[i]
	}

	// Find added and modified tasks
	for id, newT := range newTasks {
		oldT, exists := oldTasks[id]
		if !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffAdded,
				EntityType: "task",
				EntityID:   id,
				EntityRef:  newT.Ref,
				NewPath:    newT.Path,
				NewStatus:  newT.Status,
			})
			d.Summary.TasksAdded++
			continue
		}

		// Check for status change
		if oldT.Status != newT.Status {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffModified,
				EntityType: "task",
				EntityID:   id,
				EntityRef:  newT.Ref,
				OldPath:    oldT.Path,
				NewPath:    newT.Path,
				OldStatus:  oldT.Status,
				NewStatus:  newT.Status,
				Details:    "status changed",
			})
			if newT.Status == "completed" {
				d.Summary.TasksCompleted++
			}
		}
	}

	// Find removed tasks
	for id, oldT := range oldTasks {
		if _, exists := newTasks[id]; !exists {
			d.Changes = append(d.Changes, DiffChange{
				Type:       DiffRemoved,
				EntityType: "task",
				EntityID:   id,
				EntityRef:  oldT.Ref,
				OldPath:    oldT.Path,
				OldStatus:  oldT.Status,
			})
			d.Summary.TasksRemoved++
		}
	}
}

// HasChanges returns true if there are any changes
func (d *TreeDiff) HasChanges() bool {
	return len(d.Changes) > 0
}

// FilterByType returns changes of a specific entity type
func (d *TreeDiff) FilterByType(entityType string) []DiffChange {
	var filtered []DiffChange
	for _, c := range d.Changes {
		if c.EntityType == entityType {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// FilterByDiffType returns changes of a specific diff type
func (d *TreeDiff) FilterByDiffType(diffType DiffType) []DiffChange {
	var filtered []DiffChange
	for _, c := range d.Changes {
		if c.Type == diffType {
			filtered = append(filtered, c)
		}
	}
	return filtered
}
