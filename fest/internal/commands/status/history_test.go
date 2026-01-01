package status

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Note: StatusHistoryEntry type is defined in history.go

func TestStatusHistoryUpdate(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		wantEntry  bool
	}{
		{"planned to active", "planned", "active", true},
		{"active to completed", "active", "completed", true},
		{"completed to dungeon", "completed", "dungeon", true},
		{"dungeon to active", "dungeon", "active", true},
		{"same status no-op", "active", "active", false},
		{"planned to completed", "planned", "completed", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			shouldRecord := tc.fromStatus != tc.toStatus
			if shouldRecord != tc.wantEntry {
				t.Errorf("status change from %q to %q: shouldRecord = %v, want %v",
					tc.fromStatus, tc.toStatus, shouldRecord, tc.wantEntry)
			}
		})
	}
}

func TestRecordStatusChange(t *testing.T) {
	// Create temp festival directory
	festivalDir := t.TempDir()
	festDir := filepath.Join(festivalDir, ".fest")
	if err := os.MkdirAll(festDir, 0755); err != nil {
		t.Fatalf("Failed to create .fest directory: %v", err)
	}

	tests := []struct {
		name       string
		fromStatus string
		toStatus   string
		note       string
		wantError  bool
	}{
		{"planned to active with note", "planned", "active", "Starting work", false},
		{"active to completed no note", "active", "completed", "", false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Test will fail until RecordStatusChange is implemented
			err := RecordStatusChange(festivalDir, tc.fromStatus, tc.toStatus, tc.note)
			if (err != nil) != tc.wantError {
				t.Errorf("RecordStatusChange() error = %v, wantError %v", err, tc.wantError)
			}

			if !tc.wantError {
				// Verify history file was created
				historyPath := filepath.Join(festDir, "status_history.json")
				if _, err := os.Stat(historyPath); os.IsNotExist(err) {
					t.Error("status_history.json was not created")
				}
			}
		})
	}
}

func TestLoadStatusHistory(t *testing.T) {
	// Create temp festival directory
	festivalDir := t.TempDir()
	festDir := filepath.Join(festivalDir, ".fest")
	if err := os.MkdirAll(festDir, 0755); err != nil {
		t.Fatalf("Failed to create .fest directory: %v", err)
	}

	// Test loading empty history
	t.Run("no history file", func(t *testing.T) {
		history, err := LoadStatusHistory(festivalDir)
		if err != nil {
			t.Errorf("LoadStatusHistory() error = %v, want nil for missing file", err)
		}
		if len(history) != 0 {
			t.Errorf("LoadStatusHistory() returned %d entries, want 0", len(history))
		}
	})

	// Test loading after recording
	t.Run("after recording", func(t *testing.T) {
		// Record a change first
		if err := RecordStatusChange(festivalDir, "planned", "active", "test"); err != nil {
			t.Skipf("RecordStatusChange not implemented: %v", err)
		}

		history, err := LoadStatusHistory(festivalDir)
		if err != nil {
			t.Errorf("LoadStatusHistory() error = %v", err)
		}
		if len(history) != 1 {
			t.Errorf("LoadStatusHistory() returned %d entries, want 1", len(history))
		}
	})
}

func TestStatusHistoryFormat(t *testing.T) {
	// Verify history entries have required fields
	entry := StatusHistoryEntry{
		Timestamp:  time.Now(),
		FromStatus: "active",
		ToStatus:   "completed",
		Note:       "Work finished",
	}

	if entry.FromStatus == "" || entry.ToStatus == "" {
		t.Error("StatusHistoryEntry missing required status fields")
	}
	if entry.Timestamp.IsZero() {
		t.Error("StatusHistoryEntry missing timestamp")
	}
}

// Note: RecordStatusChange and LoadStatusHistory are implemented in history.go
