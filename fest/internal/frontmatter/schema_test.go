package frontmatter

import (
	"testing"
	"time"
)

func TestNewFrontmatter(t *testing.T) {
	fm := NewFrontmatter(TypeTask, "01_test", "Test Task")

	if fm.Type != TypeTask {
		t.Errorf("Type = %q, want %q", fm.Type, TypeTask)
	}
	if fm.ID != "01_test" {
		t.Errorf("ID = %q, want 01_test", fm.ID)
	}
	if fm.Name != "Test Task" {
		t.Errorf("Name = %q, want 'Test Task'", fm.Name)
	}
	if fm.Status != StatusPending {
		t.Errorf("Status = %q, want %q", fm.Status, StatusPending)
	}
	if fm.Created.IsZero() {
		t.Error("Created should not be zero")
	}
}

func TestDefaultStatus(t *testing.T) {
	tests := []struct {
		docType  Type
		expected Status
	}{
		{TypeFestival, StatusPlanned},
		{TypePhase, StatusPending},
		{TypeSequence, StatusPending},
		{TypeTask, StatusPending},
		{TypeGate, StatusPending},
	}

	for _, tc := range tests {
		t.Run(string(tc.docType), func(t *testing.T) {
			got := DefaultStatus(tc.docType)
			if got != tc.expected {
				t.Errorf("DefaultStatus(%q) = %q, want %q", tc.docType, got, tc.expected)
			}
		})
	}
}

func TestFrontmatter_Validate(t *testing.T) {
	tests := []struct {
		name      string
		fm        *Frontmatter
		wantError bool
		errCount  int
	}{
		{
			name: "valid festival",
			fm: &Frontmatter{
				Type:    TypeFestival,
				ID:      "test-festival",
				Status:  StatusActive,
				Created: time.Now(),
			},
			wantError: false,
		},
		{
			name: "valid task",
			fm: &Frontmatter{
				Type:    TypeTask,
				ID:      "01_test",
				Parent:  "01_seq",
				Order:   1,
				Status:  StatusPending,
				Created: time.Now(),
			},
			wantError: false,
		},
		{
			name: "missing type",
			fm: &Frontmatter{
				ID:      "test",
				Status:  StatusPending,
				Created: time.Now(),
			},
			wantError: true,
			errCount:  2, // type required + status invalid for unknown type
		},
		{
			name: "missing id",
			fm: &Frontmatter{
				Type:    TypeTask,
				Status:  StatusPending,
				Created: time.Now(),
			},
			wantError: true,
			errCount:  3, // id + parent + order
		},
		{
			name: "task missing parent",
			fm: &Frontmatter{
				Type:    TypeTask,
				ID:      "test",
				Order:   1,
				Status:  StatusPending,
				Created: time.Now(),
			},
			wantError: true,
			errCount:  1,
		},
		{
			name: "invalid status",
			fm: &Frontmatter{
				Type:    TypeFestival,
				ID:      "test",
				Status:  Status("invalid"),
				Created: time.Now(),
			},
			wantError: true,
			errCount:  1,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			errors := tc.fm.Validate()
			hasErrors := len(errors) > 0

			if hasErrors != tc.wantError {
				t.Errorf("Validate() hasErrors = %v, want %v, errors: %v", hasErrors, tc.wantError, errors)
			}

			if tc.wantError && tc.errCount > 0 && len(errors) != tc.errCount {
				t.Errorf("Validate() error count = %d, want %d, errors: %v", len(errors), tc.errCount, errors)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		docType Type
		status  Status
		valid   bool
	}{
		// Festival statuses
		{TypeFestival, StatusPlanned, true},
		{TypeFestival, StatusActive, true},
		{TypeFestival, StatusCompleted, true},
		{TypeFestival, StatusDungeon, true},
		{TypeFestival, StatusPending, false},

		// Task statuses
		{TypeTask, StatusPending, true},
		{TypeTask, StatusInProgress, true},
		{TypeTask, StatusBlocked, true},
		{TypeTask, StatusCompleted, true},
		{TypeTask, StatusActive, false},

		// Gate statuses
		{TypeGate, StatusPending, true},
		{TypeGate, StatusPassed, true},
		{TypeGate, StatusFailed, true},
		{TypeGate, StatusCompleted, false},
	}

	for _, tc := range tests {
		t.Run(string(tc.docType)+"_"+string(tc.status), func(t *testing.T) {
			got := isValidStatus(tc.docType, tc.status)
			if got != tc.valid {
				t.Errorf("isValidStatus(%q, %q) = %v, want %v", tc.docType, tc.status, got, tc.valid)
			}
		})
	}
}

func TestType_Constants(t *testing.T) {
	if TypeFestival != "festival" {
		t.Error("TypeFestival should be 'festival'")
	}
	if TypePhase != "phase" {
		t.Error("TypePhase should be 'phase'")
	}
	if TypeSequence != "sequence" {
		t.Error("TypeSequence should be 'sequence'")
	}
	if TypeTask != "task" {
		t.Error("TypeTask should be 'task'")
	}
	if TypeGate != "gate" {
		t.Error("TypeGate should be 'gate'")
	}
}
