package taskfilter

import (
	"testing"
)

// TestProgressConsistency verifies that all commands use the same logic
// for determining what files to count in progress calculations.
// This test ensures the fix for the progress tracking bug (P0.1) works correctly.
func TestProgressConsistency(t *testing.T) {
	// Test files that should be counted in progress (tasks + gates)
	trackedFiles := []string{
		"01_design.md",
		"02_implement.md",
		"03_test.md",
		"04_testing_and_verify.md",     // gate
		"05_code_review.md",            // gate
		"06_review_results_iterate.md", // gate
		"07_commit.md",                 // gate
	}

	// Test files that should NOT be counted in progress
	untrackedFiles := []string{
		"SEQUENCE_GOAL.md",
		"PHASE_GOAL.md",
		"FESTIVAL_GOAL.md",
		"README.md",
		"notes.md",
	}

	// Verify all tracked files are counted consistently
	for _, filename := range trackedFiles {
		// These are the three functions used by different commands:
		// - ShouldTrack: used by status/collectors.go and show/stats.go
		// - IsTask: includes both tasks and gates (used for progress counting)

		shouldTrack := ShouldTrack(filename)
		isTask := IsTask(filename)

		if !shouldTrack {
			t.Errorf("ShouldTrack(%q) = false, want true (file should be tracked)", filename)
		}
		if !isTask {
			t.Errorf("IsTask(%q) = false, want true (file should be counted as task)", filename)
		}

		// Verify consistency: ShouldTrack and IsTask should agree on tracked files
		if shouldTrack != isTask {
			t.Errorf("Inconsistency: ShouldTrack(%q) = %v, IsTask(%q) = %v - these should match",
				filename, shouldTrack, filename, isTask)
		}
	}

	// Verify all untracked files are excluded consistently
	for _, filename := range untrackedFiles {
		shouldTrack := ShouldTrack(filename)
		isTask := IsTask(filename)

		if shouldTrack {
			t.Errorf("ShouldTrack(%q) = true, want false (file should not be tracked)", filename)
		}
		if isTask {
			t.Errorf("IsTask(%q) = true, want false (file should not be counted as task)", filename)
		}
	}
}

// TestGateCountingConsistency ensures gates are counted properly in both
// separate gate totals AND in unified task totals for progress.
func TestGateCountingConsistency(t *testing.T) {
	files := []string{
		"01_design.md",
		"02_implement.md",
		"04_testing_and_verify.md",
		"05_code_review.md",
		"06_review_results_iterate.md",
		"07_commit.md",
		"08_quality_gate.md",
	}

	var taskCount, gateCount, trackedCount int

	for _, f := range files {
		ft := ClassifyFile(f)

		if ShouldTrack(f) {
			trackedCount++
		}

		switch ft {
		case FileTypeTask:
			taskCount++
		case FileTypeGate:
			gateCount++
		}
	}

	// Expected counts:
	// - 2 regular tasks (01_design, 02_implement)
	// - 5 gates (testing_and_verify, code_review, review_results_iterate, commit, quality_gate)
	expectedTasks := 2
	expectedGates := 5
	expectedTracked := 7 // All 7 should be tracked

	if taskCount != expectedTasks {
		t.Errorf("Regular task count = %d, want %d", taskCount, expectedTasks)
	}
	if gateCount != expectedGates {
		t.Errorf("Gate count = %d, want %d", gateCount, expectedGates)
	}
	if trackedCount != expectedTracked {
		t.Errorf("Tracked count = %d, want %d (tasks + gates)", trackedCount, expectedTracked)
	}

	// The key insight: trackedCount should equal taskCount + gateCount
	// This ensures unified progress calculation works correctly
	if trackedCount != taskCount+gateCount {
		t.Errorf("Tracked count (%d) != tasks (%d) + gates (%d) - progress calculation would be inconsistent",
			trackedCount, taskCount, gateCount)
	}
}

// TestPhaseSequencePatterns verifies phase and sequence detection is consistent
func TestPhaseSequencePatterns(t *testing.T) {
	// Phase directories (3-digit prefix)
	phases := []string{
		"001_PLANNING",
		"002_IMPLEMENTATION",
		"003_TESTING",
	}

	// Sequence directories (2-digit prefix)
	sequences := []string{
		"01_setup",
		"02_core_feature",
		"03_integration",
	}

	// Neither phase nor sequence
	other := []string{
		"1_task",     // 1 digit
		"0001_thing", // 4 digits
		"notes",      // no prefix
		"README.md",  // file
	}

	for _, name := range phases {
		if !IsPhaseDir(name) {
			t.Errorf("IsPhaseDir(%q) = false, want true", name)
		}
		if IsSequenceDir(name) {
			t.Errorf("IsSequenceDir(%q) = true, want false (phases are not sequences)", name)
		}
	}

	for _, name := range sequences {
		if !IsSequenceDir(name) {
			t.Errorf("IsSequenceDir(%q) = false, want true", name)
		}
		if IsPhaseDir(name) {
			t.Errorf("IsPhaseDir(%q) = true, want false (sequences are not phases)", name)
		}
	}

	for _, name := range other {
		if IsPhaseDir(name) {
			t.Errorf("IsPhaseDir(%q) = true, want false", name)
		}
		if IsSequenceDir(name) {
			t.Errorf("IsSequenceDir(%q) = true, want false", name)
		}
	}
}
