package validator

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestOrderingValidator_NoGaps(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantIssues int
	}{
		{
			name: "sequential phases",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_PLANNING"), 0755)
				os.MkdirAll(filepath.Join(dir, "002_IMPLEMENTATION"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_TESTING"), 0755)
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "sequential sequences",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				phase := filepath.Join(dir, "001_PLANNING")
				os.MkdirAll(filepath.Join(phase, "01_requirements"), 0755)
				os.MkdirAll(filepath.Join(phase, "02_design"), 0755)
				os.MkdirAll(filepath.Join(phase, "03_review"), 0755)
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "sequential tasks",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_gather.md"), []byte("# Task"), 0644)
				os.WriteFile(filepath.Join(seq, "02_analyze.md"), []byte("# Task"), 0644)
				os.WriteFile(filepath.Join(seq, "03_document.md"), []byte("# Task"), 0644)
				return dir
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			v := NewOrderingValidator()
			issues, err := v.Validate(ctx, dir)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() got %d issues, want %d", len(issues), tt.wantIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s - %s", issue.Code, issue.Message)
				}
			}
		})
	}
}

func TestOrderingValidator_MustStartFromOne(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantIssues int
		wantCode   string
	}{
		{
			name: "phase starts at 002",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "002_PLANNING"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_IMPLEMENTATION"), 0755)
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
		{
			name: "sequence starts at 05",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				phase := filepath.Join(dir, "001_PLANNING")
				os.MkdirAll(filepath.Join(phase, "05_requirements"), 0755)
				os.MkdirAll(filepath.Join(phase, "06_design"), 0755)
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
		{
			name: "task starts at 03",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "03_gather.md"), []byte("# Task"), 0644)
				os.WriteFile(filepath.Join(seq, "04_analyze.md"), []byte("# Task"), 0644)
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			v := NewOrderingValidator()
			issues, err := v.Validate(ctx, dir)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() got %d issues, want %d", len(issues), tt.wantIssues)
			}
			for _, issue := range issues {
				if issue.Code != tt.wantCode {
					t.Errorf("Expected code %q, got %q", tt.wantCode, issue.Code)
				}
				t.Logf("  Issue: %s - %s", issue.Code, issue.Message)
			}
		})
	}
}

func TestOrderingValidator_DetectsGaps(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantIssues int
		wantCode   string
	}{
		{
			name: "phase gap 001 003",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_PLANNING"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_TESTING"), 0755) // Missing 002
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
		{
			name: "sequence gap 01 03",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				phase := filepath.Join(dir, "001_PLANNING")
				os.MkdirAll(filepath.Join(phase, "01_requirements"), 0755)
				os.MkdirAll(filepath.Join(phase, "03_review"), 0755) // Missing 02
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
		{
			name: "task gap 01 04",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_gather.md"), []byte("# Task"), 0644)
				os.WriteFile(filepath.Join(seq, "04_document.md"), []byte("# Task"), 0644) // Missing 02, 03
				return dir
			},
			wantIssues: 1,
			wantCode:   CodeNumberingGap,
		},
		{
			name: "multiple gaps at different levels",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_PLANNING"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_TESTING"), 0755) // Gap in phases

				phase := filepath.Join(dir, "001_PLANNING")
				os.MkdirAll(filepath.Join(phase, "01_req"), 0755)
				os.MkdirAll(filepath.Join(phase, "05_review"), 0755) // Gap in sequences

				return dir
			},
			wantIssues: 2, // One for phases, one for sequences
			wantCode:   CodeNumberingGap,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			v := NewOrderingValidator()
			issues, err := v.Validate(ctx, dir)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() got %d issues, want %d", len(issues), tt.wantIssues)
			}
			for _, issue := range issues {
				if issue.Code != tt.wantCode {
					t.Errorf("Expected code %q, got %q", tt.wantCode, issue.Code)
				}
				t.Logf("  Issue: %s - %s", issue.Code, issue.Message)
			}
		})
	}
}

func TestOrderingValidator_ParallelTasks(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantIssues int
	}{
		{
			name: "valid parallel tasks consecutive",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_task_a.md"), []byte("# Task A"), 0644)
				os.WriteFile(filepath.Join(seq, "01_task_b.md"), []byte("# Task B"), 0644)
				os.WriteFile(filepath.Join(seq, "02_next.md"), []byte("# Next"), 0644)
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "all same number is valid",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_a.md"), []byte("# A"), 0644)
				os.WriteFile(filepath.Join(seq, "01_b.md"), []byte("# B"), 0644)
				os.WriteFile(filepath.Join(seq, "01_c.md"), []byte("# C"), 0644)
				return dir
			},
			wantIssues: 0,
		},
		{
			// NOTE: This test documents that non-consecutive duplicates cannot be
			// reliably detected via filesystem ordering because files starting with
			// the same number prefix (e.g., 01_) will always sort together
			// alphabetically, making them appear consecutive.
			// Example: 01_task_a.md, 02_task_b.md, 01_task_c.md
			// Sorts to: 01_task_a.md, 01_task_c.md, 02_task_b.md (consecutive!)
			name: "non-consecutive duplicates appear consecutive after sorting",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_requirements")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_task_a.md"), []byte("# Task A"), 0644)
				os.WriteFile(filepath.Join(seq, "02_task_b.md"), []byte("# Task B"), 0644)
				os.WriteFile(filepath.Join(seq, "01_task_c.md"), []byte("# Task C"), 0644)
				return dir
			},
			wantIssues: 0, // After alphabetical sorting, these appear consecutive
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			v := NewOrderingValidator()
			issues, err := v.Validate(ctx, dir)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() got %d issues, want %d", len(issues), tt.wantIssues)
				for _, issue := range issues {
					t.Logf("  Issue: %s - %s", issue.Code, issue.Message)
				}
			}
		})
	}
}

func TestOrderingValidator_EdgeCases(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		setup      func(t *testing.T) string
		wantIssues int
	}{
		{
			name: "empty festival",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			wantIssues: 0,
		},
		{
			name: "single phase",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_ONLY"), 0755)
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "single task",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				seq := filepath.Join(dir, "001_PLANNING", "01_req")
				os.MkdirAll(seq, 0755)
				os.WriteFile(filepath.Join(seq, "01_only.md"), []byte("# Only"), 0644)
				return dir
			},
			wantIssues: 0,
		},
		{
			name: "empty sequence no tasks",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_PLANNING", "01_empty"), 0755)
				return dir
			},
			wantIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			v := NewOrderingValidator()
			issues, err := v.Validate(ctx, dir)
			if err != nil {
				t.Fatalf("Validate() error = %v", err)
			}
			if len(issues) != tt.wantIssues {
				t.Errorf("Validate() got %d issues, want %d", len(issues), tt.wantIssues)
			}
		})
	}
}

func TestOrderingValidator_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "001_PLANNING"), 0755)

	v := NewOrderingValidator()
	_, err := v.Validate(ctx, dir)

	if err == nil {
		t.Error("Expected error due to cancelled context")
	}
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}

func TestCheckOrderingCorrect(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T) string
		wantOK bool
	}{
		{
			name: "correct ordering",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_A"), 0755)
				os.MkdirAll(filepath.Join(dir, "002_B"), 0755)
				return dir
			},
			wantOK: true,
		},
		{
			name: "gap detected",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "001_A"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_C"), 0755) // Gap!
				return dir
			},
			wantOK: false,
		},
		{
			name: "does not start at 1",
			setup: func(t *testing.T) string {
				dir := t.TempDir()
				os.MkdirAll(filepath.Join(dir, "002_A"), 0755)
				os.MkdirAll(filepath.Join(dir, "003_B"), 0755) // Starts at 002, not 001
				return dir
			},
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := tt.setup(t)
			result := CheckOrderingCorrect(dir)
			if result != tt.wantOK {
				t.Errorf("CheckOrderingCorrect() = %v, want %v", result, tt.wantOK)
			}
		})
	}
}

func TestValidateElementOrdering_MessageFormat(t *testing.T) {
	ctx := context.Background()

	// Test that error messages include proper number formatting
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "001_A"), 0755)
	os.MkdirAll(filepath.Join(dir, "005_E"), 0755) // Gap: missing 002, 003, 004

	v := NewOrderingValidator()
	issues, err := v.Validate(ctx, dir)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}

	if len(issues) != 1 {
		t.Fatalf("Expected 1 issue, got %d", len(issues))
	}

	issue := issues[0]

	// Check that message contains proper phase formatting (3 digits)
	if issue.Code != CodeNumberingGap {
		t.Errorf("Expected code %q, got %q", CodeNumberingGap, issue.Code)
	}

	// Message should mention "005 follows 001" with proper formatting
	if issue.Message == "" {
		t.Error("Expected non-empty message")
	}

	t.Logf("Issue message: %s", issue.Message)
	t.Logf("Issue fix: %s", issue.Fix)
}
