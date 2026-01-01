package festival

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
)

// TestCreateFestival_DirectoryNaming verifies that festival directories
// are created with the format {slug}_{ID} where ID is XX0001 format.
func TestCreateFestival_DirectoryNaming(t *testing.T) {
	tests := []struct {
		name           string
		festivalName   string
		expectedPrefix string // Expected 2-letter prefix
	}{
		{
			name:           "two word name",
			festivalName:   "my project",
			expectedPrefix: "MP",
		},
		{
			name:           "hyphenated name",
			festivalName:   "guild-usable",
			expectedPrefix: "GU",
		},
		{
			name:           "single word",
			festivalName:   "onboarding",
			expectedPrefix: "ON",
		},
		{
			name:           "three word name",
			festivalName:   "fest node ids",
			expectedPrefix: "FN",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()

			// Create festivals directory structure
			festivalsRoot := filepath.Join(tmpDir, "festivals")
			for _, status := range []string{"planned", "active", "completed", "dungeon"} {
				if err := os.MkdirAll(filepath.Join(festivalsRoot, status), 0755); err != nil {
					t.Fatalf("Failed to create status dir: %v", err)
				}
			}

			// Create minimal .festival/templates directory
			templatesDir := filepath.Join(festivalsRoot, ".festival", "templates")
			if err := os.MkdirAll(templatesDir, 0755); err != nil {
				t.Fatalf("Failed to create templates dir: %v", err)
			}

			// Change to festivals directory
			origDir, _ := os.Getwd()
			if err := os.Chdir(festivalsRoot); err != nil {
				t.Fatalf("Failed to chdir: %v", err)
			}
			defer os.Chdir(origDir)

			// Run create festival
			opts := &CreateFestivalOptions{
				Name:        tt.festivalName,
				Dest:        "active",
				SkipMarkers: true,
				JSONOutput:  true, // Suppress console output
			}

			err := RunCreateFestival(context.Background(), opts)
			if err != nil {
				t.Fatalf("RunCreateFestival failed: %v", err)
			}

			// Verify directory was created with ID suffix
			slug := Slugify(tt.festivalName)
			activeDir := filepath.Join(festivalsRoot, "active")
			entries, err := os.ReadDir(activeDir)
			if err != nil {
				t.Fatalf("Failed to read active dir: %v", err)
			}

			if len(entries) != 1 {
				t.Fatalf("Expected 1 entry in active/, got %d", len(entries))
			}

			dirName := entries[0].Name()

			// Directory should start with slug
			if !strings.HasPrefix(dirName, slug) {
				t.Errorf("Directory %q should start with slug %q", dirName, slug)
			}

			// Directory should have underscore separator and ID suffix
			if !strings.Contains(dirName, "_") {
				t.Errorf("Directory %q should contain underscore separator", dirName)
			}

			// Extract the ID suffix (after last underscore)
			parts := strings.Split(dirName, "_")
			if len(parts) < 2 {
				t.Fatalf("Directory %q should have format {slug}_{id}", dirName)
			}
			idPart := parts[len(parts)-1]

			// ID should be 6 characters: 2 letters + 4 digits
			if len(idPart) != 6 {
				t.Errorf("ID %q should be 6 characters (XX0001 format)", idPart)
			}

			// First two characters should be the expected prefix
			if !strings.HasPrefix(idPart, tt.expectedPrefix) {
				t.Errorf("ID %q should start with prefix %q", idPart, tt.expectedPrefix)
			}

			// Last 4 characters should be digits (0001)
			counter := idPart[2:]
			if counter != "0001" {
				t.Errorf("First festival should have counter 0001, got %q", counter)
			}
		})
	}
}

// TestCreateFestival_MetadataPopulation verifies that fest.yaml
// includes the metadata section with ID, UUID, name, and timestamps.
func TestCreateFestival_MetadataPopulation(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals directory structure
	festivalsRoot := filepath.Join(tmpDir, "festivals")
	for _, status := range []string{"planned", "active", "completed", "dungeon"} {
		if err := os.MkdirAll(filepath.Join(festivalsRoot, status), 0755); err != nil {
			t.Fatalf("Failed to create status dir: %v", err)
		}
	}

	// Create minimal .festival/templates directory
	templatesDir := filepath.Join(festivalsRoot, ".festival", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Change to festivals directory
	origDir, _ := os.Getwd()
	if err := os.Chdir(festivalsRoot); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	// Run create festival
	opts := &CreateFestivalOptions{
		Name:        "test festival",
		Goal:        "Test the metadata population",
		Dest:        "active",
		SkipMarkers: true,
		JSONOutput:  true,
	}

	err := RunCreateFestival(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunCreateFestival failed: %v", err)
	}

	// Find the created directory
	activeDir := filepath.Join(festivalsRoot, "active")
	entries, err := os.ReadDir(activeDir)
	if err != nil || len(entries) != 1 {
		t.Fatalf("Expected 1 entry in active/")
	}

	festivalDir := filepath.Join(activeDir, entries[0].Name())

	// Load the fest.yaml
	cfg, err := config.LoadFestivalConfig(festivalDir)
	if err != nil {
		t.Fatalf("Failed to load fest.yaml: %v", err)
	}

	// Verify metadata fields
	if cfg.Metadata.ID == "" {
		t.Error("Metadata.ID should not be empty")
	}

	if cfg.Metadata.UUID == "" {
		t.Error("Metadata.UUID should not be empty")
	}

	if cfg.Metadata.Name == "" {
		t.Error("Metadata.Name should not be empty")
	}

	if cfg.Metadata.Name != "test festival" {
		t.Errorf("Metadata.Name = %q, want %q", cfg.Metadata.Name, "test festival")
	}

	if cfg.Metadata.CreatedAt.IsZero() {
		t.Error("Metadata.CreatedAt should not be zero")
	}

	// Verify status history
	if len(cfg.Metadata.StatusHistory) == 0 {
		t.Error("Metadata.StatusHistory should have at least one entry")
	}

	if len(cfg.Metadata.StatusHistory) > 0 {
		firstChange := cfg.Metadata.StatusHistory[0]
		if firstChange.Status != "active" && firstChange.Status != "planned" {
			t.Errorf("First status should be 'active' or 'planned', got %q", firstChange.Status)
		}
		if firstChange.Timestamp.IsZero() {
			t.Error("Status change timestamp should not be zero")
		}
	}
}

// TestCreateFestival_UniqueIDs verifies that multiple festivals
// get unique IDs with incrementing counters.
func TestCreateFestival_UniqueIDs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals directory structure
	festivalsRoot := filepath.Join(tmpDir, "festivals")
	for _, status := range []string{"planned", "active", "completed", "dungeon"} {
		if err := os.MkdirAll(filepath.Join(festivalsRoot, status), 0755); err != nil {
			t.Fatalf("Failed to create status dir: %v", err)
		}
	}

	// Create minimal .festival/templates directory
	templatesDir := filepath.Join(festivalsRoot, ".festival", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Change to festivals directory
	origDir, _ := os.Getwd()
	if err := os.Chdir(festivalsRoot); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	// Create first festival with "GU" prefix
	opts1 := &CreateFestivalOptions{
		Name:        "guild usable",
		Dest:        "active",
		SkipMarkers: true,
		JSONOutput:  true,
	}
	if err := RunCreateFestival(context.Background(), opts1); err != nil {
		t.Fatalf("First RunCreateFestival failed: %v", err)
	}

	// Create second festival with same "GU" prefix
	opts2 := &CreateFestivalOptions{
		Name:        "guild ui",
		Dest:        "planned",
		SkipMarkers: true,
		JSONOutput:  true,
	}
	if err := RunCreateFestival(context.Background(), opts2); err != nil {
		t.Fatalf("Second RunCreateFestival failed: %v", err)
	}

	// Create third festival with different prefix
	opts3 := &CreateFestivalOptions{
		Name:        "fest node",
		Dest:        "active",
		SkipMarkers: true,
		JSONOutput:  true,
	}
	if err := RunCreateFestival(context.Background(), opts3); err != nil {
		t.Fatalf("Third RunCreateFestival failed: %v", err)
	}

	// Verify directory names
	activeEntries, _ := os.ReadDir(filepath.Join(festivalsRoot, "active"))
	plannedEntries, _ := os.ReadDir(filepath.Join(festivalsRoot, "planned"))

	// Should have 2 in active, 1 in planned
	if len(activeEntries) != 2 {
		t.Errorf("Expected 2 entries in active/, got %d", len(activeEntries))
	}
	if len(plannedEntries) != 1 {
		t.Errorf("Expected 1 entry in planned/, got %d", len(plannedEntries))
	}

	// Collect all IDs
	ids := make(map[string]bool)
	for _, e := range activeEntries {
		parts := strings.Split(e.Name(), "_")
		id := parts[len(parts)-1]
		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}
	for _, e := range plannedEntries {
		parts := strings.Split(e.Name(), "_")
		id := parts[len(parts)-1]
		if ids[id] {
			t.Errorf("Duplicate ID found: %s", id)
		}
		ids[id] = true
	}

	// Verify we have GU0001, GU0002, FN0001
	expectedIDs := []string{"GU0001", "GU0002", "FN0001"}
	for _, expectedID := range expectedIDs {
		if !ids[expectedID] {
			t.Errorf("Expected ID %s not found in %v", expectedID, ids)
		}
	}
}

// TestCreateFestival_BackwardsCompatibility verifies that old festivals
// without IDs in their directory names still work.
func TestCreateFestival_BackwardsCompatibility(t *testing.T) {
	tmpDir := t.TempDir()

	// Create festivals directory structure
	festivalsRoot := filepath.Join(tmpDir, "festivals")
	for _, status := range []string{"planned", "active", "completed", "dungeon"} {
		if err := os.MkdirAll(filepath.Join(festivalsRoot, status), 0755); err != nil {
			t.Fatalf("Failed to create status dir: %v", err)
		}
	}

	// Create an old-style festival without ID suffix
	oldFestivalDir := filepath.Join(festivalsRoot, "active", "old-festival")
	if err := os.MkdirAll(oldFestivalDir, 0755); err != nil {
		t.Fatalf("Failed to create old festival dir: %v", err)
	}

	// Create minimal fest.yaml without metadata
	oldConfig := &config.FestivalConfig{
		Version: "1.0",
		QualityGates: config.QualityGatesConfig{
			Enabled: true,
		},
	}
	if err := config.SaveFestivalConfig(oldFestivalDir, oldConfig); err != nil {
		t.Fatalf("Failed to save old fest.yaml: %v", err)
	}

	// Create minimal .festival/templates directory
	templatesDir := filepath.Join(festivalsRoot, ".festival", "templates")
	if err := os.MkdirAll(templatesDir, 0755); err != nil {
		t.Fatalf("Failed to create templates dir: %v", err)
	}

	// Change to festivals directory
	origDir, _ := os.Getwd()
	if err := os.Chdir(festivalsRoot); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	// Create a new festival - should still work even with old festival present
	opts := &CreateFestivalOptions{
		Name:        "new festival",
		Dest:        "active",
		SkipMarkers: true,
		JSONOutput:  true,
	}

	err := RunCreateFestival(context.Background(), opts)
	if err != nil {
		t.Fatalf("RunCreateFestival failed: %v", err)
	}

	// Verify both old and new festivals exist
	entries, _ := os.ReadDir(filepath.Join(festivalsRoot, "active"))
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries in active/, got %d", len(entries))
	}

	// Old festival should still be there unchanged
	oldExists := false
	for _, e := range entries {
		if e.Name() == "old-festival" {
			oldExists = true
		}
	}
	if !oldExists {
		t.Error("Old festival directory should still exist")
	}
}
