package index

import (
	"testing"
)

func TestComputeDiff_NoChanges(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{ID: "test", Name: "Test"}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{ID: "test", Name: "Test"}, "active")

	diff := ComputeDiff(old, new)

	// Check that no festival-level changes detected
	festivalChanges := diff.FilterByType("festival")
	if len(festivalChanges) > 0 {
		t.Errorf("Expected no festival changes, got %d", len(festivalChanges))
	}
}

func TestComputeDiff_FestivalAdded(t *testing.T) {
	old := NewTreeIndex("/test")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{ID: "new-festival", Name: "New Festival"}, "active")

	diff := ComputeDiff(old, new)

	if diff.Summary.FestivalsAdded != 1 {
		t.Errorf("FestivalsAdded = %d, want 1", diff.Summary.FestivalsAdded)
	}

	added := diff.FilterByDiffType(DiffAdded)
	if len(added) != 1 {
		t.Fatalf("Added count = %d, want 1", len(added))
	}
	if added[0].EntityID != "new-festival" {
		t.Errorf("EntityID = %q, want new-festival", added[0].EntityID)
	}
}

func TestComputeDiff_FestivalRemoved(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{ID: "old-festival", Name: "Old Festival"}, "active")

	new := NewTreeIndex("/test")

	diff := ComputeDiff(old, new)

	if diff.Summary.FestivalsRemoved != 1 {
		t.Errorf("FestivalsRemoved = %d, want 1", diff.Summary.FestivalsRemoved)
	}

	removed := diff.FilterByDiffType(DiffRemoved)
	if len(removed) != 1 {
		t.Fatalf("Removed count = %d, want 1", len(removed))
	}
	if removed[0].EntityID != "old-festival" {
		t.Errorf("EntityID = %q, want old-festival", removed[0].EntityID)
	}
}

func TestComputeDiff_FestivalMoved(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{ID: "test", Name: "Test"}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{ID: "test", Name: "Test"}, "completed")

	diff := ComputeDiff(old, new)

	if diff.Summary.FestivalsMoved != 1 {
		t.Errorf("FestivalsMoved = %d, want 1", diff.Summary.FestivalsMoved)
	}

	moved := diff.FilterByDiffType(DiffMoved)
	if len(moved) != 1 {
		t.Fatalf("Moved count = %d, want 1", len(moved))
	}
	if moved[0].OldStatus != "active" || moved[0].NewStatus != "completed" {
		t.Errorf("Status change = %s->%s, want active->completed", moved[0].OldStatus, moved[0].NewStatus)
	}
}

func TestComputeDiff_PhaseAdded(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{
		ID:     "test",
		Phases: []PhaseNode{},
	}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{ID: "001_Planning", Name: "Planning"},
		},
	}, "active")

	diff := ComputeDiff(old, new)

	if diff.Summary.PhasesAdded != 1 {
		t.Errorf("PhasesAdded = %d, want 1", diff.Summary.PhasesAdded)
	}
}

func TestComputeDiff_SequenceAdded(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{ID: "001_Planning", Sequences: []SequenceNode{}},
		},
	}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{
				ID: "001_Planning",
				Sequences: []SequenceNode{
					{ID: "01_setup", Name: "Setup"},
				},
			},
		},
	}, "active")

	diff := ComputeDiff(old, new)

	if diff.Summary.SequencesAdded != 1 {
		t.Errorf("SequencesAdded = %d, want 1", diff.Summary.SequencesAdded)
	}
}

func TestComputeDiff_TaskAdded(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{
				ID: "001_Planning",
				Sequences: []SequenceNode{
					{ID: "01_setup", Tasks: []TaskNode{}},
				},
			},
		},
	}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{
				ID: "001_Planning",
				Sequences: []SequenceNode{
					{
						ID: "01_setup",
						Tasks: []TaskNode{
							{ID: "01_task", Name: "Task", Status: "pending"},
						},
					},
				},
			},
		},
	}, "active")

	diff := ComputeDiff(old, new)

	if diff.Summary.TasksAdded != 1 {
		t.Errorf("TasksAdded = %d, want 1", diff.Summary.TasksAdded)
	}
}

func TestComputeDiff_TaskCompleted(t *testing.T) {
	old := NewTreeIndex("/test")
	old.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{
				ID: "001_Planning",
				Sequences: []SequenceNode{
					{
						ID: "01_setup",
						Tasks: []TaskNode{
							{ID: "01_task", Name: "Task", Status: "pending"},
						},
					},
				},
			},
		},
	}, "active")

	new := NewTreeIndex("/test")
	new.AddFestival(FestivalNode{
		ID: "test",
		Phases: []PhaseNode{
			{
				ID: "001_Planning",
				Sequences: []SequenceNode{
					{
						ID: "01_setup",
						Tasks: []TaskNode{
							{ID: "01_task", Name: "Task", Status: "completed"},
						},
					},
				},
			},
		},
	}, "active")

	diff := ComputeDiff(old, new)

	if diff.Summary.TasksCompleted != 1 {
		t.Errorf("TasksCompleted = %d, want 1", diff.Summary.TasksCompleted)
	}

	modified := diff.FilterByDiffType(DiffModified)
	if len(modified) != 1 {
		t.Fatalf("Modified count = %d, want 1", len(modified))
	}
	if modified[0].OldStatus != "pending" || modified[0].NewStatus != "completed" {
		t.Errorf("Status change = %s->%s, want pending->completed", modified[0].OldStatus, modified[0].NewStatus)
	}
}

func TestHasChanges(t *testing.T) {
	diff := &TreeDiff{
		Changes: []DiffChange{},
	}
	if diff.HasChanges() {
		t.Error("HasChanges should be false for empty changes")
	}

	diff.Changes = append(diff.Changes, DiffChange{Type: DiffAdded})
	if !diff.HasChanges() {
		t.Error("HasChanges should be true when changes exist")
	}
}

func TestFilterByType(t *testing.T) {
	diff := &TreeDiff{
		Changes: []DiffChange{
			{Type: DiffAdded, EntityType: "festival"},
			{Type: DiffAdded, EntityType: "phase"},
			{Type: DiffRemoved, EntityType: "task"},
			{Type: DiffModified, EntityType: "task"},
		},
	}

	tasks := diff.FilterByType("task")
	if len(tasks) != 2 {
		t.Errorf("FilterByType(task) count = %d, want 2", len(tasks))
	}

	phases := diff.FilterByType("phase")
	if len(phases) != 1 {
		t.Errorf("FilterByType(phase) count = %d, want 1", len(phases))
	}
}

func TestFilterByDiffType(t *testing.T) {
	diff := &TreeDiff{
		Changes: []DiffChange{
			{Type: DiffAdded, EntityType: "festival"},
			{Type: DiffAdded, EntityType: "phase"},
			{Type: DiffRemoved, EntityType: "task"},
			{Type: DiffModified, EntityType: "task"},
		},
	}

	added := diff.FilterByDiffType(DiffAdded)
	if len(added) != 2 {
		t.Errorf("FilterByDiffType(Added) count = %d, want 2", len(added))
	}

	removed := diff.FilterByDiffType(DiffRemoved)
	if len(removed) != 1 {
		t.Errorf("FilterByDiffType(Removed) count = %d, want 1", len(removed))
	}
}
