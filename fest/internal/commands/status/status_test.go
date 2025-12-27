package status

import (
	"testing"
)

func TestValidStatuses(t *testing.T) {
	tests := []struct {
		entityType EntityType
		status     string
		expected   bool
	}{
		// Festival statuses
		{EntityFestival, "planned", true},
		{EntityFestival, "active", true},
		{EntityFestival, "completed", true},
		{EntityFestival, "dungeon", true},
		{EntityFestival, "invalid", false},

		// Phase statuses
		{EntityPhase, "pending", true},
		{EntityPhase, "in_progress", true},
		{EntityPhase, "completed", true},
		{EntityPhase, "blocked", false},

		// Sequence statuses
		{EntitySequence, "pending", true},
		{EntitySequence, "in_progress", true},
		{EntitySequence, "completed", true},
		{EntitySequence, "blocked", false},

		// Task statuses
		{EntityTask, "pending", true},
		{EntityTask, "in_progress", true},
		{EntityTask, "blocked", true},
		{EntityTask, "completed", true},
		{EntityTask, "invalid", false},

		// Gate statuses
		{EntityGate, "pending", true},
		{EntityGate, "passed", true},
		{EntityGate, "failed", true},
		{EntityGate, "completed", false},
	}

	for _, tc := range tests {
		t.Run(string(tc.entityType)+"/"+tc.status, func(t *testing.T) {
			result := isValidStatus(tc.entityType, tc.status)
			if result != tc.expected {
				t.Errorf("isValidStatus(%q, %q) = %v, want %v",
					tc.entityType, tc.status, result, tc.expected)
			}
		})
	}
}

func TestEntityTypes(t *testing.T) {
	// Verify all entity types have valid statuses defined
	entityTypes := []EntityType{
		EntityFestival,
		EntityPhase,
		EntitySequence,
		EntityTask,
		EntityGate,
	}

	for _, et := range entityTypes {
		statuses, ok := ValidStatuses[et]
		if !ok {
			t.Errorf("EntityType %q has no valid statuses defined", et)
			continue
		}
		if len(statuses) == 0 {
			t.Errorf("EntityType %q has empty valid statuses", et)
		}
	}
}

func TestIsValidStatus_UnknownEntityType(t *testing.T) {
	result := isValidStatus("unknown", "pending")
	if result {
		t.Error("isValidStatus should return false for unknown entity type")
	}
}
