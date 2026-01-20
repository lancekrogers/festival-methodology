package progress

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStore_LoadNew(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	err := store.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if store.Data() == nil {
		t.Fatal("Data() returned nil")
	}

	if store.Data().Tasks == nil {
		t.Error("Tasks map should not be nil")
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create and save
	store1 := NewStore(tmpDir)
	if err := store1.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	now := time.Now().UTC()
	store1.SetTask(&TaskProgress{
		TaskID:    "01_test.md",
		Status:    StatusInProgress,
		Progress:  50,
		StartedAt: &now,
	})

	if err := store1.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	progressPath := filepath.Join(tmpDir, ProgressDir, ProgressFileName)
	if _, err := os.Stat(progressPath); os.IsNotExist(err) {
		t.Fatalf("Progress file not created at %s", progressPath)
	}

	// Load in new store
	store2 := NewStore(tmpDir)
	if err := store2.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	task, found := store2.GetTask("01_test.md")
	if !found {
		t.Fatal("Task not found after load")
	}

	if task.Status != StatusInProgress {
		t.Errorf("Status = %q, want %q", task.Status, StatusInProgress)
	}

	if task.Progress != 50 {
		t.Errorf("Progress = %d, want 50", task.Progress)
	}
}

func TestStore_GetTask(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Non-existent task
	_, found := store.GetTask("nonexistent")
	if found {
		t.Error("GetTask() should return false for non-existent task")
	}

	// Add and retrieve
	store.SetTask(&TaskProgress{TaskID: "01_test.md", Status: StatusPending})
	task, found := store.GetTask("01_test.md")
	if !found {
		t.Error("GetTask() should return true for existing task")
	}
	if task.TaskID != "01_test.md" {
		t.Errorf("TaskID = %q, want %q", task.TaskID, "01_test.md")
	}
}

func TestStore_AllTasks(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()
	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	store.SetTask(&TaskProgress{TaskID: "01_test.md"})
	store.SetTask(&TaskProgress{TaskID: "02_test.md"})
	store.SetTask(&TaskProgress{TaskID: "03_test.md"})

	tasks := store.AllTasks()
	if len(tasks) != 3 {
		t.Errorf("AllTasks() returned %d tasks, want 3", len(tasks))
	}
}

// TestStore_BackwardCompatibility verifies that legacy progress.yaml files
// without time_metrics field can still be loaded without errors
func TestStore_BackwardCompatibility(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create a legacy progress.yaml without time_metrics
	legacyYAML := `festival: test-festival
updated_at: 2026-01-15T10:00:00Z
tasks:
  01_task.md:
    task_id: 01_task.md
    status: completed
    progress: 100
`
	festDir := filepath.Join(tmpDir, ProgressDir)
	if err := os.MkdirAll(festDir, 0755); err != nil {
		t.Fatalf("mkdir error = %v", err)
	}
	if err := os.WriteFile(filepath.Join(festDir, ProgressFileName), []byte(legacyYAML), 0644); err != nil {
		t.Fatalf("write error = %v", err)
	}

	// Load should succeed without time_metrics
	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v, legacy files should load without errors", err)
	}

	// Verify task data loaded correctly
	task, found := store.GetTask("01_task.md")
	if !found {
		t.Fatal("Task should be found in legacy file")
	}
	if task.Status != StatusCompleted {
		t.Errorf("Status = %q, want %q", task.Status, StatusCompleted)
	}

	// TimeMetrics should be initialized for legacy files (backward compatibility upgrade)
	if store.Data().TimeMetrics == nil {
		t.Error("TimeMetrics should be initialized for legacy files")
	}
	// CreatedAt should use UpdatedAt as fallback for legacy files
	if store.Data().TimeMetrics.CreatedAt.IsZero() {
		t.Error("TimeMetrics.CreatedAt should not be zero")
	}
}

// TestStore_TimeMetricsRoundtrip verifies FestivalTimeMetrics saves and loads correctly
func TestStore_TimeMetricsRoundtrip(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create with time metrics
	store1 := NewStore(tmpDir)
	if err := store1.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	now := time.Now().UTC()
	completed := now.Add(24 * time.Hour)
	store1.data.TimeMetrics = &FestivalTimeMetrics{
		CreatedAt:         now,
		CompletedAt:       &completed,
		LifecycleDuration: 1,
		TotalWorkMinutes:  120,
	}

	if err := store1.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load in new store and verify
	store2 := NewStore(tmpDir)
	if err := store2.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	metrics := store2.Data().TimeMetrics
	if metrics == nil {
		t.Fatal("TimeMetrics should not be nil after roundtrip")
	}
	if metrics.TotalWorkMinutes != 120 {
		t.Errorf("TotalWorkMinutes = %d, want 120", metrics.TotalWorkMinutes)
	}
	if metrics.LifecycleDuration != 1 {
		t.Errorf("LifecycleDuration = %d, want 1", metrics.LifecycleDuration)
	}
	if metrics.CompletedAt == nil {
		t.Error("CompletedAt should not be nil")
	}
}

// TestStore_NewFestivalsGetTimeMetrics verifies new festivals get TimeMetrics initialized
func TestStore_NewFestivalsGetTimeMetrics(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// New festivals should have TimeMetrics initialized
	if store.Data().TimeMetrics == nil {
		t.Fatal("TimeMetrics should be initialized for new festivals")
	}

	// CreatedAt should be set
	if store.Data().TimeMetrics.CreatedAt.IsZero() {
		t.Error("CreatedAt should be set for new festivals")
	}

	// Save and reload should preserve TimeMetrics
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusPending})
	if err := store.Save(ctx); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	store2 := NewStore(tmpDir)
	if err := store2.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if store2.Data().TimeMetrics == nil {
		t.Fatal("TimeMetrics should be preserved after save/load")
	}
	if store2.Data().TimeMetrics.CreatedAt.IsZero() {
		t.Error("CreatedAt should be preserved after save/load")
	}
}

// TestStore_HelperMethods verifies the time metrics helper methods
func TestStore_HelperMethods(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Add some tasks with time
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusCompleted, TimeSpentMinutes: 30})
	store.SetTask(&TaskProgress{TaskID: "02_task.md", Status: StatusCompleted, TimeSpentMinutes: 45})

	// Test UpdateTotalWorkMinutes
	store.UpdateTotalWorkMinutes()
	if store.GetTimeMetrics().TotalWorkMinutes != 75 {
		t.Errorf("TotalWorkMinutes = %d, want 75", store.GetTimeMetrics().TotalWorkMinutes)
	}

	// Test MarkFestivalCompleted
	store.MarkFestivalCompleted()
	if store.GetTimeMetrics().CompletedAt == nil {
		t.Error("CompletedAt should be set after MarkFestivalCompleted")
	}
}

func TestStore_IsFestivalComplete(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Empty festival should not be complete
	if store.IsFestivalComplete() {
		t.Error("Empty festival should not be complete")
	}

	// Festival with pending tasks should not be complete
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusPending})
	if store.IsFestivalComplete() {
		t.Error("Festival with pending tasks should not be complete")
	}

	// Festival with mixed statuses should not be complete
	store.SetTask(&TaskProgress{TaskID: "02_task.md", Status: StatusCompleted})
	if store.IsFestivalComplete() {
		t.Error("Festival with mixed statuses should not be complete")
	}

	// Festival with all completed tasks should be complete
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusCompleted})
	if !store.IsFestivalComplete() {
		t.Error("Festival with all completed tasks should be complete")
	}
}

func TestStore_CheckAndSetCompletion(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	store := NewStore(tmpDir)
	if err := store.Load(ctx); err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Add incomplete tasks
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusPending, TimeSpentMinutes: 30})
	store.SetTask(&TaskProgress{TaskID: "02_task.md", Status: StatusCompleted, TimeSpentMinutes: 45})

	// CheckAndSetCompletion should not mark as complete
	if store.CheckAndSetCompletion() {
		t.Error("CheckAndSetCompletion should return false for incomplete festival")
	}
	if store.GetTimeMetrics().CompletedAt != nil {
		t.Error("CompletedAt should not be set for incomplete festival")
	}

	// Complete all tasks
	store.SetTask(&TaskProgress{TaskID: "01_task.md", Status: StatusCompleted, TimeSpentMinutes: 30})

	// CheckAndSetCompletion should mark as complete
	if !store.CheckAndSetCompletion() {
		t.Error("CheckAndSetCompletion should return true for complete festival")
	}
	if store.GetTimeMetrics().CompletedAt == nil {
		t.Error("CompletedAt should be set for complete festival")
	}
	if store.GetTimeMetrics().TotalWorkMinutes != 75 {
		t.Errorf("TotalWorkMinutes = %d, want 75", store.GetTimeMetrics().TotalWorkMinutes)
	}

	// Second call should be idempotent (return false, not reset timestamp)
	completedAt := store.GetTimeMetrics().CompletedAt
	if store.CheckAndSetCompletion() {
		t.Error("CheckAndSetCompletion should return false on second call (idempotent)")
	}
	if !store.GetTimeMetrics().CompletedAt.Equal(*completedAt) {
		t.Error("CompletedAt should not change on second call")
	}
}

func TestFestivalTimeMetrics_CalculateLifecycleDuration(t *testing.T) {
	tests := []struct {
		name         string
		metrics      *FestivalTimeMetrics
		expectedDays int
	}{
		{
			name:         "nil metrics",
			metrics:      nil,
			expectedDays: -1,
		},
		{
			name: "ongoing festival (no CompletedAt)",
			metrics: &FestivalTimeMetrics{
				CreatedAt: time.Now().Add(-48 * time.Hour),
			},
			expectedDays: -1,
		},
		{
			name: "completed same day",
			metrics: &FestivalTimeMetrics{
				CreatedAt:   time.Now().Add(-12 * time.Hour),
				CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
			expectedDays: 0,
		},
		{
			name: "completed after 1 day",
			metrics: &FestivalTimeMetrics{
				CreatedAt:   time.Now().Add(-36 * time.Hour),
				CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
			expectedDays: 1,
		},
		{
			name: "completed after 30 days",
			metrics: &FestivalTimeMetrics{
				CreatedAt:   time.Now().Add(-30 * 24 * time.Hour),
				CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
			expectedDays: 30,
		},
		{
			name: "completed after 365 days (no cap)",
			metrics: &FestivalTimeMetrics{
				CreatedAt:   time.Now().Add(-365 * 24 * time.Hour),
				CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
			expectedDays: 365,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metrics.CalculateLifecycleDuration()
			if got != tt.expectedDays {
				t.Errorf("CalculateLifecycleDuration() = %d, want %d", got, tt.expectedDays)
			}
		})
	}
}

func TestFormatLifecycleDuration(t *testing.T) {
	tests := []struct {
		days     int
		expected string
	}{
		{-1, "ongoing"},
		{0, "< 1 day"},
		{1, "1 day"},
		{2, "2 days"},
		{30, "30 days"},
		{365, "365 days"},
	}

	for _, tt := range tests {
		got := FormatLifecycleDuration(tt.days)
		if got != tt.expected {
			t.Errorf("FormatLifecycleDuration(%d) = %q, want %q", tt.days, got, tt.expected)
		}
	}
}
