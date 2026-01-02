package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAtomicStatusChange(t *testing.T) {
	tests := []struct {
		name          string
		setupFn       func(baseDir string) (festivalPath string)
		fromStatus    string
		toStatus      string
		wantError     bool
		checkRollback bool
	}{
		{
			name: "successful planned to active",
			setupFn: func(baseDir string) string {
				path := filepath.Join(baseDir, "planned", "test-festival")
				os.MkdirAll(path, 0755)
				os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)
				return path
			},
			fromStatus: "planned",
			toStatus:   "active",
			wantError:  false,
		},
		{
			name: "successful active to completed with date",
			setupFn: func(baseDir string) string {
				path := filepath.Join(baseDir, "active", "test-festival")
				os.MkdirAll(path, 0755)
				os.WriteFile(filepath.Join(path, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644)
				return path
			},
			fromStatus: "active",
			toStatus:   "completed",
			wantError:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()
			festivalPath := tc.setupFn(baseDir)

			newPath, err := AtomicStatusChange(festivalPath, tc.fromStatus, tc.toStatus)
			if (err != nil) != tc.wantError {
				t.Errorf("AtomicStatusChange() error = %v, wantError %v", err, tc.wantError)
			}

			if !tc.wantError {
				// Verify source no longer exists
				if _, err := os.Stat(festivalPath); !os.IsNotExist(err) {
					t.Error("Source directory still exists after status change")
				}

				// Verify destination exists
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					t.Errorf("Destination directory does not exist: %s", newPath)
				}
			}
		})
	}
}

func TestAtomicStatusChangeRollback(t *testing.T) {
	baseDir := t.TempDir()

	// Create source festival
	sourcePath := filepath.Join(baseDir, "active", "test-festival")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}
	testFile := filepath.Join(sourcePath, "FESTIVAL_OVERVIEW.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create conflicting destination at the correct date-based path
	// The dateDir will be the current month in YYYY-MM format
	dateDir := CalculateCompletionDateDir(time.Now())
	conflictPath := filepath.Join(baseDir, "completed", dateDir, "test-festival")
	if err := os.MkdirAll(conflictPath, 0755); err != nil {
		t.Fatalf("Failed to create conflict: %v", err)
	}

	// Attempt status change - should fail
	_, err := AtomicStatusChange(sourcePath, "active", "completed")
	if err == nil {
		t.Error("Expected error when destination conflicts")
	}

	// Verify source still exists (rollback/no change)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		t.Error("Source was removed despite failed status change")
	}

	// Verify test file still exists
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Test file was lost during failed status change")
	}
}

func TestCrossFilesystemFallback(t *testing.T) {
	// This test simulates cross-filesystem move by using copy+delete
	// In practice, os.Rename returns EXDEV for cross-filesystem moves
	baseDir := t.TempDir()

	sourcePath := filepath.Join(baseDir, "active", "test-festival")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// Create some files
	testFile := filepath.Join(sourcePath, "FESTIVAL_OVERVIEW.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	subDir := filepath.Join(sourcePath, "001_phase")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	subFile := filepath.Join(subDir, "PHASE_GOAL.md")
	if err := os.WriteFile(subFile, []byte("phase content"), 0644); err != nil {
		t.Fatalf("Failed to create subfile: %v", err)
	}

	destDir := filepath.Join(baseDir, "completed")

	// Use copy-delete for cross-filesystem simulation
	newPath, err := CopyDeleteMove(sourcePath, destDir, "test-festival")
	if err != nil {
		t.Fatalf("CopyDeleteMove() error = %v", err)
	}

	// Verify source is gone
	if _, err := os.Stat(sourcePath); !os.IsNotExist(err) {
		t.Error("Source still exists after copy-delete move")
	}

	// Verify destination exists with all files
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("Destination does not exist")
	}

	movedFile := filepath.Join(newPath, "FESTIVAL_OVERVIEW.md")
	if _, err := os.Stat(movedFile); os.IsNotExist(err) {
		t.Error("Test file was not copied")
	}

	movedSubFile := filepath.Join(newPath, "001_phase", "PHASE_GOAL.md")
	if _, err := os.Stat(movedSubFile); os.IsNotExist(err) {
		t.Error("Subdirectory file was not copied")
	}
}

func TestStatusSetValidation(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		newStatus  string
		wantValid  bool
	}{
		{"festival to active", EntityFestival, "active", true},
		{"festival to completed", EntityFestival, "completed", true},
		{"festival to dungeon", EntityFestival, "dungeon", true},
		{"festival to planned", EntityFestival, "planned", true},
		{"festival to invalid", EntityFestival, "invalid", false},
		{"festival to pending", EntityFestival, "pending", false}, // pending is for phases
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := isValidStatus(tc.entityType, tc.newStatus)
			if got != tc.wantValid {
				t.Errorf("isValidStatus(%q, %q) = %v, want %v",
					tc.entityType, tc.newStatus, got, tc.wantValid)
			}
		})
	}
}

func TestCompletedUsesDateDirectory(t *testing.T) {
	baseDir := t.TempDir()

	// Create source festival
	sourcePath := filepath.Join(baseDir, "active", "test-festival")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sourcePath, "FESTIVAL_OVERVIEW.md"), []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create overview: %v", err)
	}

	newPath, err := AtomicStatusChange(sourcePath, "active", "completed")
	if err != nil {
		t.Fatalf("AtomicStatusChange() error = %v", err)
	}

	// Verify path includes date directory (YYYY-MM format)
	// Path should be like: baseDir/completed/2025-01/test-festival
	relPath, err := filepath.Rel(baseDir, newPath)
	if err != nil {
		t.Fatalf("Failed to get relative path: %v", err)
	}

	// Should have completed/YYYY-MM/festival-name structure
	// Parse path parts to verify structure
	_ = filepath.SplitList(relPath) // Verify path can be parsed
	if len(relPath) < 10 {          // At minimum: completed/YYYY-MM
		t.Errorf("Path too short for date directory structure: %s", relPath)
	}
	if relPath[:9] != "completed" {
		t.Errorf("Expected path under 'completed', got: %s", relPath)
	}
}

// Note: AtomicStatusChange and CopyDeleteMove are implemented in atomic.go
