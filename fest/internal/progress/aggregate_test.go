package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// createTestFestivalWithTasks creates a test festival structure with tasks
func createTestFestivalWithTasks(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// Create fest.yaml
	festYAML := `name: test-festival
id: TEST-001
`
	if err := os.WriteFile(filepath.Join(dir, "fest.yaml"), []byte(festYAML), 0644); err != nil {
		t.Fatalf("Failed to create fest.yaml: %v", err)
	}

	// Create phase directory
	phaseDir := filepath.Join(dir, "001_PHASE")
	if err := os.MkdirAll(phaseDir, 0755); err != nil {
		t.Fatalf("Failed to create phase directory: %v", err)
	}

	// Create sequence directory
	seqDir := filepath.Join(phaseDir, "01_sequence")
	if err := os.MkdirAll(seqDir, 0755); err != nil {
		t.Fatalf("Failed to create sequence directory: %v", err)
	}

	// Create task files
	taskContent := `---
fest_type: task
fest_status: pending
---
# Task 01
`
	if err := os.WriteFile(filepath.Join(seqDir, "01_task.md"), []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(seqDir, "02_task.md"), []byte(taskContent), 0644); err != nil {
		t.Fatalf("Failed to create task file: %v", err)
	}

	return dir
}

// TestGetFestivalProgress_TimeMetricsInitialized verifies TimeMetrics is always present
func TestGetFestivalProgress_TimeMetricsInitialized(t *testing.T) {
	ctx := context.Background()
	festPath := createTestFestivalWithTasks(t)

	mgr, err := NewManager(ctx, festPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	progress, err := mgr.GetFestivalProgress(ctx, festPath)
	if err != nil {
		t.Fatalf("GetFestivalProgress() error = %v", err)
	}

	// TimeMetrics should always be initialized
	if progress.TimeMetrics == nil {
		t.Error("TimeMetrics should not be nil")
	}

	// CreatedAt should be set
	if progress.TimeMetrics.CreatedAt.IsZero() {
		t.Error("TimeMetrics.CreatedAt should not be zero")
	}
}

// TestGetFestivalProgress_TotalWorkMinutesAggregation verifies work time sums correctly
func TestGetFestivalProgress_TotalWorkMinutesAggregation(t *testing.T) {
	ctx := context.Background()
	festPath := createTestFestivalWithTasks(t)

	mgr, err := NewManager(ctx, festPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Complete tasks with known time values
	now := time.Now()
	thirtyMinsAgo := now.Add(-30 * time.Minute)
	sixtyMinsAgo := now.Add(-60 * time.Minute)

	// Set task 1: 30 minutes
	store := mgr.Store()
	store.SetTask(&TaskProgress{
		TaskID:           "001_PHASE/01_sequence/01_task.md",
		Status:           StatusCompleted,
		Progress:         100,
		StartedAt:        &thirtyMinsAgo,
		CompletedAt:      &now,
		TimeSpentMinutes: 30,
	})

	// Set task 2: 15 minutes
	fifteenMinsAgo := now.Add(-15 * time.Minute)
	store.SetTask(&TaskProgress{
		TaskID:           "001_PHASE/01_sequence/02_task.md",
		Status:           StatusCompleted,
		Progress:         100,
		StartedAt:        &sixtyMinsAgo,
		CompletedAt:      &fifteenMinsAgo,
		TimeSpentMinutes: 45,
	})

	// Update total work minutes
	store.UpdateTotalWorkMinutes()

	if err := store.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Get fresh progress
	progress, err := mgr.GetFestivalProgress(ctx, festPath)
	if err != nil {
		t.Fatalf("GetFestivalProgress() error = %v", err)
	}

	// Verify total work minutes
	expectedTotal := 75 // 30 + 45
	if progress.TimeMetrics.TotalWorkMinutes != expectedTotal {
		t.Errorf("TotalWorkMinutes = %d, want %d", progress.TimeMetrics.TotalWorkMinutes, expectedTotal)
	}
}

// TestGetFestivalProgress_LifecycleDurationOngoing verifies ongoing festivals have zero stored duration
func TestGetFestivalProgress_LifecycleDurationOngoing(t *testing.T) {
	ctx := context.Background()
	festPath := createTestFestivalWithTasks(t)

	mgr, err := NewManager(ctx, festPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	progress, err := mgr.GetFestivalProgress(ctx, festPath)
	if err != nil {
		t.Fatalf("GetFestivalProgress() error = %v", err)
	}

	// Ongoing (newly created) festivals have 0 stored lifecycle duration
	// The display layer uses FormatDurationWithStatus to show "ongoing" or calculate current days
	if progress.LifecycleDuration != 0 {
		t.Errorf("LifecycleDuration = %d, want 0 for newly created festivals", progress.LifecycleDuration)
	}

	// CompletedAt should be nil for ongoing festivals
	if progress.TimeMetrics.CompletedAt != nil {
		t.Error("CompletedAt should be nil for ongoing festivals")
	}
}

// TestGetFestivalProgress_LifecycleDurationCompleted verifies completed festivals have positive duration
func TestGetFestivalProgress_LifecycleDurationCompleted(t *testing.T) {
	ctx := context.Background()
	festPath := createTestFestivalWithTasks(t)

	mgr, err := NewManager(ctx, festPath)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	// Mark festival as completed
	store := mgr.Store()
	metrics := store.EnsureTimeMetrics()

	// Set CreatedAt to 2 days ago
	twoDaysAgo := time.Now().Add(-48 * time.Hour)
	metrics.CreatedAt = twoDaysAgo

	// Mark completed now
	store.MarkFestivalCompleted()

	if err := store.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Get fresh progress
	progress, err := mgr.GetFestivalProgress(ctx, festPath)
	if err != nil {
		t.Fatalf("GetFestivalProgress() error = %v", err)
	}

	// Completed festivals should have positive lifecycle duration (2 days)
	if progress.LifecycleDuration < 2 {
		t.Errorf("LifecycleDuration = %d, want >= 2 for 2-day old completed festival", progress.LifecycleDuration)
	}
}

// TestGetFestivalProgress_EmptyFestival verifies empty festivals are handled correctly
func TestGetFestivalProgress_EmptyFestival(t *testing.T) {
	ctx := context.Background()
	dir := t.TempDir()

	// Create minimal festival
	festYAML := `name: empty-festival
id: EMPTY-001
`
	if err := os.WriteFile(filepath.Join(dir, "fest.yaml"), []byte(festYAML), 0644); err != nil {
		t.Fatalf("Failed to create fest.yaml: %v", err)
	}

	mgr, err := NewManager(ctx, dir)
	if err != nil {
		t.Fatalf("NewManager() error = %v", err)
	}

	progress, err := mgr.GetFestivalProgress(ctx, dir)
	if err != nil {
		t.Fatalf("GetFestivalProgress() error = %v", err)
	}

	// Should have zero totals but valid structure
	if progress.Overall == nil {
		t.Error("Overall should not be nil")
	}
	if progress.Overall.Total != 0 {
		t.Errorf("Total = %d, want 0 for empty festival", progress.Overall.Total)
	}
	if progress.TimeMetrics == nil {
		t.Error("TimeMetrics should not be nil even for empty festivals")
	}
}
