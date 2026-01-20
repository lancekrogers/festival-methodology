package progress

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFileModTime_ExistingFile(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Get the expected mod time
	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}
	expected := info.ModTime()

	// Test GetFileModTime
	got := GetFileModTime(testFile)
	if !got.Equal(expected) {
		t.Errorf("GetFileModTime() = %v, want %v", got, expected)
	}
}

func TestGetFileModTime_NonExistentFile(t *testing.T) {
	got := GetFileModTime("/nonexistent/path/to/file.md")
	if !got.IsZero() {
		t.Errorf("GetFileModTime() for non-existent file = %v, want zero time", got)
	}
}

func TestGetFileModTime_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	got := GetFileModTime(tmpDir)
	if got.IsZero() {
		t.Error("GetFileModTime() for directory should return non-zero time")
	}
}

func TestFileTimeCache_GetModTime(t *testing.T) {
	// Create a temp file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileTimeCache()

	// First call should hit filesystem
	t1 := cache.GetModTime(testFile)
	if t1.IsZero() {
		t.Error("First GetModTime() should return non-zero time")
	}

	// Cache should have one entry
	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1", cache.Size())
	}

	// Second call should return same time (cached)
	t2 := cache.GetModTime(testFile)
	if !t1.Equal(t2) {
		t.Errorf("Second GetModTime() = %v, want %v (cached)", t2, t1)
	}
}

func TestFileTimeCache_NonExistentFile(t *testing.T) {
	cache := NewFileTimeCache()
	got := cache.GetModTime("/nonexistent/path")
	if !got.IsZero() {
		t.Error("GetModTime() for non-existent file should return zero time")
	}

	// Zero time should also be cached
	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1 (zero time should be cached)", cache.Size())
	}
}

func TestFileTimeCache_Invalidate(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileTimeCache()
	cache.GetModTime(testFile)

	if cache.Size() != 1 {
		t.Errorf("Cache size = %d, want 1", cache.Size())
	}

	cache.Invalidate(testFile)

	if cache.Size() != 0 {
		t.Errorf("Cache size after invalidate = %d, want 0", cache.Size())
	}
}

func TestFileTimeCache_Clear(t *testing.T) {
	cache := NewFileTimeCache()
	cache.GetModTime("/path1")
	cache.GetModTime("/path2")
	cache.GetModTime("/path3")

	if cache.Size() != 3 {
		t.Errorf("Cache size = %d, want 3", cache.Size())
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Cache size after clear = %d, want 0", cache.Size())
	}
}

func TestFileTimeCache_CacheReflectsFileChanges(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(testFile, []byte("initial"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cache := NewFileTimeCache()
	t1 := cache.GetModTime(testFile)

	// Wait a moment and modify the file
	time.Sleep(10 * time.Millisecond)
	if err := os.WriteFile(testFile, []byte("modified"), 0644); err != nil {
		t.Fatalf("Failed to modify test file: %v", err)
	}

	// Cache still returns old time
	t2 := cache.GetModTime(testFile)
	if !t1.Equal(t2) {
		t.Error("Cached time should not change when file changes")
	}

	// After invalidation, should get new time
	cache.Invalidate(testFile)
	t3 := cache.GetModTime(testFile)
	if t1.Equal(t3) {
		t.Error("After invalidate, should get updated file time")
	}
}

// TestInferTaskTime tests the time inference logic
func TestInferTaskTime_ExplicitTimeNotOverwritten(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(testFile, []byte("# Task"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Task with explicit time data
	task := &TaskProgress{
		TaskID:           "task.md",
		Status:           StatusCompleted,
		TimeSpentMinutes: 60, // Explicit time from API
	}

	applied := InferTaskTime(testFile, task)

	if applied {
		t.Error("InferTaskTime should not modify tasks with explicit time data")
	}
	if task.TimeSpentMinutes != 60 {
		t.Errorf("TimeSpentMinutes = %d, want 60 (unchanged)", task.TimeSpentMinutes)
	}
}

func TestInferTaskTime_SetsCompletedAt(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(testFile, []byte("# Task"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Task completed but no timestamps
	task := &TaskProgress{
		TaskID: "task.md",
		Status: StatusCompleted,
	}

	applied := InferTaskTime(testFile, task)

	if !applied {
		t.Error("InferTaskTime should apply inference to completed task without timestamps")
	}
	if task.CompletedAt == nil {
		t.Error("CompletedAt should be set from file ModTime")
	}
}

func TestInferTaskTime_CalculatesTimeSpan(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(testFile, []byte("# Task"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Task with StartedAt but no CompletedAt or TimeSpentMinutes
	startTime := time.Now().Add(-2 * time.Hour)
	task := &TaskProgress{
		TaskID:    "task.md",
		Status:    StatusCompleted,
		StartedAt: &startTime,
	}

	applied := InferTaskTime(testFile, task)

	if !applied {
		t.Error("InferTaskTime should apply inference")
	}
	if task.CompletedAt == nil {
		t.Error("CompletedAt should be set from file ModTime")
	}
	// Time should be roughly 2 hours (120 minutes) - allow some variance
	if task.TimeSpentMinutes < 110 || task.TimeSpentMinutes > 130 {
		t.Errorf("TimeSpentMinutes = %d, expected ~120", task.TimeSpentMinutes)
	}
}

func TestInferTaskTime_NoArbitraryCaps(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "task.md")
	if err := os.WriteFile(testFile, []byte("# Task"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Task that took 3 days
	startTime := time.Now().Add(-72 * time.Hour) // 3 days ago
	completedTime := time.Now()
	task := &TaskProgress{
		TaskID:      "task.md",
		Status:      StatusCompleted,
		StartedAt:   &startTime,
		CompletedAt: &completedTime,
	}

	applied := InferTaskTime(testFile, task)

	if !applied {
		t.Error("InferTaskTime should apply inference")
	}
	// Should report approximately 4320 minutes (72 hours)
	// NO 8-hour (480 min) caps should be applied
	if task.TimeSpentMinutes < 4300 {
		t.Errorf("TimeSpentMinutes = %d, expected ~4320 (no arbitrary caps)", task.TimeSpentMinutes)
	}
}

func TestNeedsTimeInference(t *testing.T) {
	tests := []struct {
		name string
		task *TaskProgress
		want bool
	}{
		{
			name: "nil task",
			task: nil,
			want: false,
		},
		{
			name: "has explicit time",
			task: &TaskProgress{
				Status:           StatusCompleted,
				TimeSpentMinutes: 30,
			},
			want: false,
		},
		{
			name: "not completed",
			task: &TaskProgress{
				Status: StatusInProgress,
			},
			want: false,
		},
		{
			name: "completed without timestamp",
			task: &TaskProgress{
				Status: StatusCompleted,
			},
			want: true,
		},
		{
			name: "completed with timestamp but no time",
			task: &TaskProgress{
				Status:      StatusCompleted,
				CompletedAt: func() *time.Time { t := time.Now(); return &t }(),
			},
			want: false, // Has timestamp, inference would just calculate from existing data
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NeedsTimeInference(tt.task)
			if got != tt.want {
				t.Errorf("NeedsTimeInference() = %v, want %v", got, tt.want)
			}
		})
	}
}
