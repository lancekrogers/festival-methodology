//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestValidateOrderingIntegration tests the fest validate ordering command in a container
func TestValidateOrderingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container, err := NewTestContainer(t)
	require.NoError(t, err, "Failed to create test container")
	defer container.Cleanup()

	// Create base festivals directory
	_, _, err = container.container.Exec(container.ctx, []string{"mkdir", "-p", "/festivals"})
	require.NoError(t, err)

	t.Run("ValidOrderingPasses", func(t *testing.T) {
		// Create a festival with valid sequential numbering
		festivalPath := "/festivals/valid-festival"

		// Create festival with sequential phases 001, 002, 003
		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_requirements"},
			{"mkdir", "-p", festivalPath + "/002_DESIGN/01_architecture"},
			{"mkdir", "-p", festivalPath + "/003_IMPLEMENT/01_core"},
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create required files
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Valid Festival\n\nTest festival with valid ordering."},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning Phase"},
			{festivalPath + "/001_PLANNING/01_requirements/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/01_requirements/01_gather.md", "# Gather Requirements"},
			{festivalPath + "/001_PLANNING/01_requirements/02_analyze.md", "# Analyze Requirements"},
			{festivalPath + "/002_DESIGN/PHASE_GOAL.md", "# Design Phase"},
			{festivalPath + "/002_DESIGN/01_architecture/SEQUENCE_GOAL.md", "# Architecture"},
			{festivalPath + "/002_DESIGN/01_architecture/01_design.md", "# Design"},
			{festivalPath + "/003_IMPLEMENT/PHASE_GOAL.md", "# Implementation Phase"},
			{festivalPath + "/003_IMPLEMENT/01_core/SEQUENCE_GOAL.md", "# Core Implementation"},
			{festivalPath + "/003_IMPLEMENT/01_core/01_implement.md", "# Implement"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err, "Failed to create %s", file.path)
		}

		// Run validate ordering - should pass (exit 0)
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.NoError(t, err, "validate ordering should pass for valid festival: %s", output)
		t.Logf("Valid ordering output: %s", output)

		// Run full validate - should also pass ordering section
		output, err = container.RunFest("validate", festivalPath)
		require.NoError(t, err, "validate should pass: %s", output)
		require.Contains(t, output, "ORDERING", "Output should contain ORDERING section")
	})

	t.Run("PhaseGapDetected", func(t *testing.T) {
		// Create a festival with a gap in phase numbering
		festivalPath := "/festivals/phase-gap-festival"

		// Create phases 001 and 003 (missing 002)
		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING"},
			{"mkdir", "-p", festivalPath + "/003_TESTING"}, // Gap - missing 002
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create required files
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Gap Festival\n\nTest festival with phase gap."},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/003_TESTING/PHASE_GOAL.md", "# Testing"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering - should fail and report the gap
		output, err := container.RunFest("validate", "ordering", festivalPath)
		// The command should fail (return error) because there's a gap
		require.Error(t, err, "validate ordering should fail for festival with gap")
		require.Contains(t, output, "numbering gap", "Output should mention numbering gap")
		require.Contains(t, output, "003", "Output should mention the phase number after gap")
		t.Logf("Phase gap detection output: %s", output)
	})

	t.Run("SequenceGapDetected", func(t *testing.T) {
		// Create a festival with a gap in sequence numbering
		festivalPath := "/festivals/sequence-gap-festival"

		// Create sequences 01 and 04 (missing 02, 03)
		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_requirements"},
			{"mkdir", "-p", festivalPath + "/001_PLANNING/04_review"}, // Gap
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create required files
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Sequence Gap Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/001_PLANNING/01_requirements/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/04_review/SEQUENCE_GOAL.md", "# Review"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering - should fail
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.Error(t, err, "validate ordering should fail for festival with sequence gap")
		require.Contains(t, output, "sequence", "Output should mention sequence")
		require.Contains(t, output, "numbering gap", "Output should mention numbering gap")
		t.Logf("Sequence gap detection output: %s", output)
	})

	t.Run("TaskGapDetected", func(t *testing.T) {
		// Create a festival with a gap in task numbering
		festivalPath := "/festivals/task-gap-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_requirements"},
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create tasks with gap: 01, 05 (missing 02, 03, 04)
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Task Gap Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/001_PLANNING/01_requirements/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/01_requirements/01_gather.md", "# Gather"},
			{festivalPath + "/001_PLANNING/01_requirements/05_verify.md", "# Verify"}, // Gap
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering - should fail
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.Error(t, err, "validate ordering should fail for festival with task gap")
		require.Contains(t, output, "task", "Output should mention task")
		require.Contains(t, output, "numbering gap", "Output should mention numbering gap")
		t.Logf("Task gap detection output: %s", output)
	})

	t.Run("MustStartFromOne", func(t *testing.T) {
		// Create a festival that doesn't start from 001
		festivalPath := "/festivals/bad-start-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/002_DESIGN"},  // Should start at 001
			{"mkdir", "-p", festivalPath + "/003_IMPLEMENT"},
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Bad Start Festival"},
			{festivalPath + "/002_DESIGN/PHASE_GOAL.md", "# Design"},
			{festivalPath + "/003_IMPLEMENT/PHASE_GOAL.md", "# Implement"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering - should fail because it doesn't start at 001
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.Error(t, err, "validate ordering should fail when not starting from 001")
		require.Contains(t, output, "must start at", "Output should mention must start at")
		t.Logf("Must start from 1 detection output: %s", output)
	})

	t.Run("ParallelTasksValid", func(t *testing.T) {
		// Create a festival with valid parallel tasks (consecutive duplicates)
		festivalPath := "/festivals/parallel-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_requirements"},
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create parallel tasks with same number (01_a, 01_b) then sequential (02)
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Parallel Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/001_PLANNING/01_requirements/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/01_requirements/01_task_a.md", "# Task A"},
			{festivalPath + "/001_PLANNING/01_requirements/01_task_b.md", "# Task B"}, // Parallel with above
			{festivalPath + "/001_PLANNING/01_requirements/02_next.md", "# Next Task"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering - should pass (parallel tasks are valid)
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.NoError(t, err, "validate ordering should pass for festival with valid parallel tasks: %s", output)
		t.Logf("Parallel tasks valid output: %s", output)
	})

	t.Run("NonConsecutiveDuplicatesLimitation", func(t *testing.T) {
		// NOTE: This test documents a limitation of filesystem-based validation.
		// Non-consecutive duplicates cannot be reliably detected because files
		// starting with the same number prefix always sort together alphabetically.
		// Example: 01_a.md, 02_b.md, 01_c.md sorts to: 01_a.md, 01_c.md, 02_b.md
		festivalPath := "/festivals/nonconsec-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_requirements"},
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		// Create tasks: 01_a, 02_b, 01_c - after sorting these appear as 01_a, 01_c, 02_b
		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Non-Consecutive Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/001_PLANNING/01_requirements/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/01_requirements/01_task_a.md", "# Task A"},
			{festivalPath + "/001_PLANNING/01_requirements/02_task_b.md", "# Task B"},
			{festivalPath + "/001_PLANNING/01_requirements/01_task_c.md", "# Task C"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// After alphabetical sorting, duplicates appear consecutive - this passes
		output, err := container.RunFest("validate", "ordering", festivalPath)
		require.NoError(t, err, "validate ordering passes because alphabetical sort makes duplicates consecutive: %s", output)
		t.Logf("Non-consecutive duplicates limitation documented: %s", output)
	})

	t.Run("FullValidateIncludesOrdering", func(t *testing.T) {
		// Test that the main validate command includes ordering checks
		festivalPath := "/festivals/full-validate-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING/01_req"},
			{"mkdir", "-p", festivalPath + "/003_TESTING/01_test"}, // Gap in phases
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# Full Validate Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/001_PLANNING/01_req/SEQUENCE_GOAL.md", "# Requirements"},
			{festivalPath + "/001_PLANNING/01_req/01_task.md", "# Task"},
			{festivalPath + "/003_TESTING/PHASE_GOAL.md", "# Testing"},
			{festivalPath + "/003_TESTING/01_test/SEQUENCE_GOAL.md", "# Test"},
			{festivalPath + "/003_TESTING/01_test/01_test.md", "# Test Task"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run full validate - should include ordering issues
		output, _ := container.RunFest("validate", festivalPath)
		require.Contains(t, output, "ORDERING", "Full validate output should include ORDERING section")
		require.Contains(t, output, "numbering gap", "Full validate should detect ordering gap")
		t.Logf("Full validate with ordering output: %s", output)
	})

	t.Run("JSONOutputFormat", func(t *testing.T) {
		// Test that JSON output works for ordering validation
		festivalPath := "/festivals/json-test-festival"

		cmds := [][]string{
			{"mkdir", "-p", festivalPath + "/001_PLANNING"},
			{"mkdir", "-p", festivalPath + "/003_TESTING"}, // Gap
		}
		for _, cmd := range cmds {
			exitCode, _, err := container.container.Exec(container.ctx, cmd)
			require.NoError(t, err)
			require.Equal(t, 0, exitCode)
		}

		createFiles := []struct {
			path    string
			content string
		}{
			{festivalPath + "/FESTIVAL_OVERVIEW.md", "# JSON Test Festival"},
			{festivalPath + "/001_PLANNING/PHASE_GOAL.md", "# Planning"},
			{festivalPath + "/003_TESTING/PHASE_GOAL.md", "# Testing"},
		}
		for _, file := range createFiles {
			err := container.container.CopyToContainer(container.ctx, []byte(file.content), file.path, 0644)
			require.NoError(t, err)
		}

		// Run validate ordering with --json flag
		output, _ := container.RunFest("validate", "ordering", "--json", festivalPath)
		require.Contains(t, output, `"code"`, "JSON output should contain code field")
		require.Contains(t, output, `"numbering_gap"`, "JSON output should contain numbering_gap code")
		require.Contains(t, output, `"valid"`, "JSON output should contain valid field")
		t.Logf("JSON output: %s", output)
	})
}
