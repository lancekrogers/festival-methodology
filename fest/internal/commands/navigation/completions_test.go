package navigation

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestFindCompletedFestivals(t *testing.T) {
	tmpDir := t.TempDir()
	completedDir := filepath.Join(tmpDir, "festivals", "completed")

	// Create test structure with date directories
	festivals := map[string][]string{
		"2024-11": {"fest-alpha", "fest-beta"},
		"2024-12": {"fest-gamma"},
		"2025-01": {"fest-delta", "fest-epsilon"},
	}

	for dateDir, fests := range festivals {
		for _, fest := range fests {
			path := filepath.Join(completedDir, dateDir, fest)
			if err := os.MkdirAll(path, 0755); err != nil {
				t.Fatal(err)
			}
			// Create FESTIVAL_OVERVIEW.md to make it a valid festival
			if err := os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}

	tests := []struct {
		name      string
		prefix    string
		wantCount int
	}{
		{"all festivals", "", 5},
		{"fest- prefix", "fest-", 5},
		{"alpha prefix", "alpha", 0},
		{"fest-a prefix", "fest-a", 1}, // fest-alpha
		{"fest-d prefix", "fest-d", 1}, // fest-delta
		{"fest-e prefix", "fest-e", 1}, // fest-epsilon
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findCompletedFestivals(completedDir, tt.prefix)
			if len(got) != tt.wantCount {
				t.Errorf("findCompletedFestivals() returned %d, want %d: %v", len(got), tt.wantCount, got)
			}
		})
	}
}

func TestFindCompletedFestivals_MixedStructure(t *testing.T) {
	tmpDir := t.TempDir()
	completedDir := filepath.Join(tmpDir, "festivals", "completed")

	// Old flat structure (legacy festivals)
	legacyFests := []string{"old-fest-1", "old-fest-2"}
	for _, fest := range legacyFests {
		path := filepath.Join(completedDir, fest)
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
		// Create FESTIVAL_OVERVIEW.md to make it a valid festival
		if err := os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// New date-based structure
	dateDir := "2025-01"
	newFest := "new-fest-1"
	path := filepath.Join(completedDir, dateDir, newFest)
	if err := os.MkdirAll(path, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	got := findCompletedFestivals(completedDir, "")

	// Should find both legacy and new festivals
	if len(got) != 3 {
		t.Errorf("Expected 3 festivals, got %d: %v", len(got), got)
	}

	// Verify specific festivals are found
	foundOld1, foundOld2, foundNew := false, false, false
	for _, f := range got {
		switch f {
		case "old-fest-1":
			foundOld1 = true
		case "old-fest-2":
			foundOld2 = true
		case "new-fest-1":
			foundNew = true
		}
	}

	if !foundOld1 || !foundOld2 {
		t.Error("Legacy festivals not found")
	}
	if !foundNew {
		t.Error("New date-based festival not found")
	}
}

func TestFindCompletedFestivals_EmptyDateDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	completedDir := filepath.Join(tmpDir, "festivals", "completed")

	// Create an empty date directory
	if err := os.MkdirAll(filepath.Join(completedDir, "2025-01"), 0755); err != nil {
		t.Fatal(err)
	}

	got := findCompletedFestivals(completedDir, "")

	if len(got) != 0 {
		t.Errorf("Expected 0 festivals from empty directory, got %d: %v", len(got), got)
	}
}

func TestFindCompletedFestivals_NoCompletedDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	completedDir := filepath.Join(tmpDir, "festivals", "completed")
	// Don't create the directory

	got := findCompletedFestivals(completedDir, "")

	if len(got) != 0 {
		t.Errorf("Expected 0 festivals from nonexistent directory, got %d: %v", len(got), got)
	}
}

func TestIsDateDirectory(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"2025-01", true},
		{"2024-12", true},
		{"2020-06", true},
		{"my-festival", false},
		{"2025-1", false}, // Not zero-padded
		{"25-01", false},  // Short year
		{"202501", false}, // No dash
		{"2025-13", true}, // Month validation not enforced by pattern
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDateDirectory(tt.name)
			if got != tt.want {
				t.Errorf("isDateDirectory(%q) = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func BenchmarkFindCompletedFestivals(b *testing.B) {
	tmpDir := b.TempDir()
	completedDir := filepath.Join(tmpDir, "festivals", "completed")

	// Create 72 date directories (6 years * 12 months) with 10 festivals each = 720 festivals
	for year := 2020; year <= 2025; year++ {
		for month := 1; month <= 12; month++ {
			dateDir := fmt.Sprintf("%d-%02d", year, month)
			for i := 0; i < 10; i++ {
				festName := fmt.Sprintf("fest-%d-%02d-%d", year, month, i)
				path := filepath.Join(completedDir, dateDir, festName)
				os.MkdirAll(path, 0755)
				os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)
			}
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		findCompletedFestivals(completedDir, "")
	}
}

// Note: isDateDirectory, findCompletedFestivals, and isValidFestivalDir
// are implemented in completions.go
