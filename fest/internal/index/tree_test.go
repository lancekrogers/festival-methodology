package index

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewTreeIndex(t *testing.T) {
	tree := NewTreeIndex("/test/workspace")

	if tree.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", tree.Version)
	}
	if tree.Workspace.Path != "/test/workspace" {
		t.Errorf("Workspace.Path = %q, want /test/workspace", tree.Workspace.Path)
	}
}

func TestAddFestival(t *testing.T) {
	tree := NewTreeIndex("/test")

	festival := FestivalNode{
		ID:   "test-festival",
		Name: "Test Festival",
	}

	tree.AddFestival(festival, "active")

	if len(tree.Festivals.Active) != 1 {
		t.Fatalf("Active count = %d, want 1", len(tree.Festivals.Active))
	}
	if tree.Festivals.Active[0].ID != "test-festival" {
		t.Errorf("Festival ID = %q, want test-festival", tree.Festivals.Active[0].ID)
	}
}

func TestAddFestivalByStatus(t *testing.T) {
	tests := []struct {
		status   string
		expected int
	}{
		{"planned", 1},
		{"active", 1},
		{"completed", 1},
		{"dungeon", 1},
		{"unknown", 1}, // defaults to active
	}

	for _, tc := range tests {
		t.Run(tc.status, func(t *testing.T) {
			tree := NewTreeIndex("/test")
			festival := FestivalNode{ID: "test"}
			tree.AddFestival(festival, tc.status)

			var count int
			switch tc.status {
			case "planned":
				count = len(tree.Festivals.Planned)
			case "active", "unknown":
				count = len(tree.Festivals.Active)
			case "completed":
				count = len(tree.Festivals.Completed)
			case "dungeon":
				count = len(tree.Festivals.Dungeon)
			}

			if count != tc.expected {
				t.Errorf("count = %d, want %d", count, tc.expected)
			}
		})
	}
}

func TestGetAllFestivals(t *testing.T) {
	tree := NewTreeIndex("/test")

	tree.AddFestival(FestivalNode{ID: "planned1"}, "planned")
	tree.AddFestival(FestivalNode{ID: "active1"}, "active")
	tree.AddFestival(FestivalNode{ID: "completed1"}, "completed")
	tree.AddFestival(FestivalNode{ID: "dungeon1"}, "dungeon")

	all := tree.GetAllFestivals()
	if len(all) != 4 {
		t.Errorf("GetAllFestivals count = %d, want 4", len(all))
	}
}

func TestGetFestivalByID(t *testing.T) {
	tree := NewTreeIndex("/test")

	tree.AddFestival(FestivalNode{ID: "test1", Name: "Test One"}, "active")
	tree.AddFestival(FestivalNode{ID: "test2", Name: "Test Two"}, "planned")

	found := tree.GetFestivalByID("test2")
	if found == nil {
		t.Fatal("GetFestivalByID returned nil")
	}
	if found.Name != "Test Two" {
		t.Errorf("Name = %q, want Test Two", found.Name)
	}

	notFound := tree.GetFestivalByID("nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent ID")
	}
}

func TestGetFestivalByRef(t *testing.T) {
	tree := NewTreeIndex("/test")

	tree.AddFestival(FestivalNode{ID: "test1", Ref: "FEST-abc123"}, "active")

	found := tree.GetFestivalByRef("FEST-abc123")
	if found == nil {
		t.Fatal("GetFestivalByRef returned nil")
	}
	if found.ID != "test1" {
		t.Errorf("ID = %q, want test1", found.ID)
	}

	notFound := tree.GetFestivalByRef("FEST-nonexistent")
	if notFound != nil {
		t.Error("Expected nil for nonexistent ref")
	}
}

func TestCalculateSummary(t *testing.T) {
	tree := NewTreeIndex("/test")

	tree.AddFestival(FestivalNode{
		ID:             "test1",
		TaskCount:      10,
		CompletedTasks: 5,
	}, "active")
	tree.AddFestival(FestivalNode{
		ID:             "test2",
		TaskCount:      8,
		CompletedTasks: 8,
	}, "completed")

	tree.CalculateSummary()

	if tree.Workspace.FestivalCount != 2 {
		t.Errorf("FestivalCount = %d, want 2", tree.Workspace.FestivalCount)
	}
	if tree.Workspace.TotalTasks != 18 {
		t.Errorf("TotalTasks = %d, want 18", tree.Workspace.TotalTasks)
	}
	if tree.Workspace.CompletedTasks != 13 {
		t.Errorf("CompletedTasks = %d, want 13", tree.Workspace.CompletedTasks)
	}
}

func TestSortFestivals(t *testing.T) {
	tree := NewTreeIndex("/test")

	tree.AddFestival(FestivalNode{ID: "z-last", Name: "Z Last"}, "active")
	tree.AddFestival(FestivalNode{ID: "a-first", Name: "A First"}, "active")
	tree.AddFestival(FestivalNode{ID: "m-middle", Name: "M Middle"}, "active")

	tree.SortFestivals()

	if tree.Festivals.Active[0].Name != "A First" {
		t.Errorf("First festival = %q, want A First", tree.Festivals.Active[0].Name)
	}
	if tree.Festivals.Active[2].Name != "Z Last" {
		t.Errorf("Last festival = %q, want Z Last", tree.Festivals.Active[2].Name)
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	indexPath := filepath.Join(tmpDir, "tree-index.json")

	tree := NewTreeIndex("/test/workspace")
	tree.AddFestival(FestivalNode{
		ID:        "test-festival",
		Name:      "Test Festival",
		Ref:       "FEST-abc123",
		TaskCount: 5,
	}, "active")

	// Save
	if err := tree.Save(indexPath); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Check file exists
	if _, err := os.Stat(indexPath); os.IsNotExist(err) {
		t.Fatal("Index file was not created")
	}

	// Load
	loaded, err := LoadTreeIndex(indexPath)
	if err != nil {
		t.Fatalf("LoadTreeIndex error: %v", err)
	}

	if loaded.Version != tree.Version {
		t.Errorf("Loaded version = %q, want %q", loaded.Version, tree.Version)
	}
	if len(loaded.Festivals.Active) != 1 {
		t.Fatalf("Loaded active count = %d, want 1", len(loaded.Festivals.Active))
	}
	if loaded.Festivals.Active[0].Ref != "FEST-abc123" {
		t.Errorf("Loaded ref = %q, want FEST-abc123", loaded.Festivals.Active[0].Ref)
	}
}

func TestLoadTreeIndex_NotFound(t *testing.T) {
	_, err := LoadTreeIndex("/nonexistent/path.json")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
