package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCalculateCompletionDateDir(t *testing.T) {
	tests := []struct {
		name      string
		timestamp time.Time
		want      string
	}{
		{"january 2025", time.Date(2025, 1, 15, 0, 0, 0, 0, time.Local), "2025-01"},
		{"december 2024", time.Date(2024, 12, 31, 23, 59, 59, 0, time.Local), "2024-12"},
		{"month boundary start", time.Date(2025, 2, 1, 0, 0, 0, 0, time.Local), "2025-02"},
		{"month boundary end", time.Date(2025, 3, 31, 23, 59, 59, 0, time.Local), "2025-03"},
		{"leap year february", time.Date(2024, 2, 29, 12, 0, 0, 0, time.Local), "2024-02"},
		{"new year", time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local), "2026-01"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := CalculateCompletionDateDir(tc.timestamp)
			if got != tc.want {
				t.Errorf("CalculateCompletionDateDir(%v) = %q, want %q",
					tc.timestamp, got, tc.want)
			}
		})
	}
}

func TestCalculateCompletionDateDirNow(t *testing.T) {
	now := time.Now()
	expected := now.Format("2006-01")
	got := CalculateCompletionDateDir(now)

	if got != expected {
		t.Errorf("CalculateCompletionDateDir(now) = %q, want %q", got, expected)
	}
}

func TestCreateDateDirectory(t *testing.T) {
	tests := []struct {
		name      string
		dateDir   string
		wantError bool
	}{
		{"valid date dir", "2025-01", false},
		{"another valid", "2024-12", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()
			completedDir := filepath.Join(baseDir, "completed")

			err := CreateDateDirectory(completedDir, tc.dateDir)
			if (err != nil) != tc.wantError {
				t.Errorf("CreateDateDirectory() error = %v, wantError %v", err, tc.wantError)
			}

			if !tc.wantError {
				expectedPath := filepath.Join(completedDir, tc.dateDir)
				if _, err := os.Stat(expectedPath); os.IsNotExist(err) {
					t.Errorf("Date directory was not created at %s", expectedPath)
				}
			}
		})
	}
}

func TestCreateDateDirectoryIdempotent(t *testing.T) {
	baseDir := t.TempDir()
	completedDir := filepath.Join(baseDir, "completed")
	dateDir := "2025-01"

	// Create twice - should not error
	if err := CreateDateDirectory(completedDir, dateDir); err != nil {
		t.Fatalf("First CreateDateDirectory() failed: %v", err)
	}

	if err := CreateDateDirectory(completedDir, dateDir); err != nil {
		t.Errorf("Second CreateDateDirectory() should be idempotent, got error: %v", err)
	}
}

func TestMoveToDateDirectory(t *testing.T) {
	tests := []struct {
		name          string
		festivalName  string
		dateDir       string
		wantError     bool
		setupExisting bool // if true, create conflicting destination
	}{
		{"normal move", "my-festival", "2025-01", false, false},
		{"conflict exists", "my-festival", "2025-01", true, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			baseDir := t.TempDir()

			// Create source festival
			activePath := filepath.Join(baseDir, "active", tc.festivalName)
			if err := os.MkdirAll(activePath, 0755); err != nil {
				t.Fatalf("Failed to create source: %v", err)
			}

			// Create a file to verify move
			testFile := filepath.Join(activePath, "FESTIVAL_OVERVIEW.md")
			if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			completedDir := filepath.Join(baseDir, "completed")
			if err := os.MkdirAll(completedDir, 0755); err != nil {
				t.Fatalf("Failed to create completed dir: %v", err)
			}

			if tc.setupExisting {
				// Create conflicting destination
				existingPath := filepath.Join(completedDir, tc.dateDir, tc.festivalName)
				if err := os.MkdirAll(existingPath, 0755); err != nil {
					t.Fatalf("Failed to create existing path: %v", err)
				}
			}

			newPath, err := MoveToDateDirectory(activePath, completedDir, tc.dateDir)
			if (err != nil) != tc.wantError {
				t.Errorf("MoveToDateDirectory() error = %v, wantError %v", err, tc.wantError)
			}

			if !tc.wantError {
				// Verify source no longer exists
				if _, err := os.Stat(activePath); !os.IsNotExist(err) {
					t.Error("Source directory still exists after move")
				}

				// Verify destination exists
				if _, err := os.Stat(newPath); os.IsNotExist(err) {
					t.Error("Destination directory does not exist after move")
				}

				// Verify file was moved
				movedFile := filepath.Join(newPath, "FESTIVAL_OVERVIEW.md")
				if _, err := os.Stat(movedFile); os.IsNotExist(err) {
					t.Error("Test file was not moved with directory")
				}
			}
		})
	}
}

func TestMoveToDateDirectoryCreatesParent(t *testing.T) {
	baseDir := t.TempDir()

	// Create source
	sourcePath := filepath.Join(baseDir, "active", "test-festival")
	if err := os.MkdirAll(sourcePath, 0755); err != nil {
		t.Fatalf("Failed to create source: %v", err)
	}

	// completedDir doesn't exist yet
	completedDir := filepath.Join(baseDir, "completed")

	_, err := MoveToDateDirectory(sourcePath, completedDir, "2025-01")
	if err != nil {
		t.Errorf("MoveToDateDirectory() should create parent dirs, got error: %v", err)
	}
}

func TestGetCompletedPath(t *testing.T) {
	tests := []struct {
		name         string
		festivalName string
		dateDir      string
		want         string
	}{
		{"simple", "my-festival", "2025-01", "completed/2025-01/my-festival"},
		{"with suffix", "my-project_AB0001", "2024-12", "completed/2024-12/my-project_AB0001"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			festivalsRoot := "/festivals"
			got := GetCompletedPath(festivalsRoot, tc.festivalName, tc.dateDir)
			want := filepath.Join(festivalsRoot, tc.want)
			if got != want {
				t.Errorf("GetCompletedPath() = %q, want %q", got, want)
			}
		})
	}
}

// Note: CalculateCompletionDateDir, CreateDateDirectory, MoveToDateDirectory,
// and GetCompletedPath are implemented in date_directory.go
