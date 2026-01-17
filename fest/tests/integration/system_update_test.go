//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSystemUpdate_DeletesOrphanedFiles verifies that `fest system update --force`
// deletes files that exist in the workspace but not in the source templates.
// This is a regression test for the bug where old template files were not cleaned up.
func TestSystemUpdate_DeletesOrphanedFiles(t *testing.T) {
	tc := GetSharedContainer(t)

	// Step 1: Create a minimal source template structure (simulating ~/.config/fest/festivals/)
	// This represents the "new" template structure from upstream
	_, err := tc.runCommand([]string{"sh", "-c", `
		mkdir -p /root/.config/fest/festivals/.festival/templates/festival
		mkdir -p /root/.config/fest/festivals/.festival/templates/phases/implementation
		mkdir -p /root/.config/fest/festivals/.festival/templates/tasks

		echo "# Festival Goal Template" > /root/.config/fest/festivals/.festival/templates/festival/GOAL.md
		echo "# Implementation Phase" > /root/.config/fest/festivals/.festival/templates/phases/implementation/GOAL.md
		echo "# Task Template" > /root/.config/fest/festivals/.festival/templates/tasks/TASK.md
	`})
	require.NoError(t, err, "failed to create source templates")

	// Step 2: Create a workspace with the source files PLUS orphaned files
	// The orphaned files represent old templates that have been removed from upstream
	_, err = tc.runCommand([]string{"sh", "-c", `
		mkdir -p /workspace/festivals/.festival/templates/festival
		mkdir -p /workspace/festivals/.festival/templates/phases/implementation
		mkdir -p /workspace/festivals/.festival/templates/tasks
		mkdir -p /workspace/festivals/.festival/templates/gates

		# Copy "good" files (exist in source)
		echo "# Festival Goal Template" > /workspace/festivals/.festival/templates/festival/GOAL.md
		echo "# Implementation Phase" > /workspace/festivals/.festival/templates/phases/implementation/GOAL.md
		echo "# Task Template" > /workspace/festivals/.festival/templates/tasks/TASK.md

		# Create orphaned files (don't exist in source)
		echo "OLD TEMPLATE - SHOULD BE DELETED" > /workspace/festivals/.festival/templates/FESTIVAL_GOAL_TEMPLATE.md
		echo "OLD TEMPLATE - SHOULD BE DELETED" > /workspace/festivals/.festival/templates/PHASE_GOAL_TEMPLATE.md
		echo "OLD TEMPLATE - SHOULD BE DELETED" > /workspace/festivals/.festival/templates/gates/QUALITY_GATE.md
	`})
	require.NoError(t, err, "failed to create workspace with orphaned files")

	// Step 3: Create the checksums file (required by update command)
	// The checksums should match the current workspace files so they appear "unchanged"
	_, err = tc.runCommand([]string{"sh", "-c", `
		mkdir -p /workspace/festivals/.festival
		cat > /workspace/festivals/.festival/.fest-checksums.json << 'EOF'
{
  "templates/festival/GOAL.md": {"hash": "abc123", "modified": "2024-01-01T00:00:00Z"},
  "templates/phases/implementation/GOAL.md": {"hash": "def456", "modified": "2024-01-01T00:00:00Z"},
  "templates/tasks/TASK.md": {"hash": "ghi789", "modified": "2024-01-01T00:00:00Z"},
  "templates/FESTIVAL_GOAL_TEMPLATE.md": {"hash": "old111", "modified": "2024-01-01T00:00:00Z"},
  "templates/PHASE_GOAL_TEMPLATE.md": {"hash": "old222", "modified": "2024-01-01T00:00:00Z"},
  "templates/gates/QUALITY_GATE.md": {"hash": "old333", "modified": "2024-01-01T00:00:00Z"}
}
EOF
	`})
	require.NoError(t, err, "failed to create checksums file")

	// Verify orphaned files exist before update
	orphan1Exists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/FESTIVAL_GOAL_TEMPLATE.md")
	require.NoError(t, err)
	assert.True(t, orphan1Exists, "orphaned file should exist before update")

	orphan2Exists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/PHASE_GOAL_TEMPLATE.md")
	require.NoError(t, err)
	assert.True(t, orphan2Exists, "orphaned file should exist before update")

	gatesOrphanExists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/gates/QUALITY_GATE.md")
	require.NoError(t, err)
	assert.True(t, gatesOrphanExists, "orphaned gates file should exist before update")

	// Step 4: Run fest system update --force
	output, err := tc.RunFestInDir("/workspace/festivals", "system", "update", "--force")
	t.Logf("Update output: %s", output)
	// Note: We don't require NoError because the update might succeed but return non-zero

	// Step 5: Verify orphaned files were deleted
	orphan1Exists, err = tc.CheckFileExists("/workspace/festivals/.festival/templates/FESTIVAL_GOAL_TEMPLATE.md")
	require.NoError(t, err)
	assert.False(t, orphan1Exists, "orphaned FESTIVAL_GOAL_TEMPLATE.md should be deleted after update")

	orphan2Exists, err = tc.CheckFileExists("/workspace/festivals/.festival/templates/PHASE_GOAL_TEMPLATE.md")
	require.NoError(t, err)
	assert.False(t, orphan2Exists, "orphaned PHASE_GOAL_TEMPLATE.md should be deleted after update")

	gatesOrphanExists, err = tc.CheckFileExists("/workspace/festivals/.festival/templates/gates/QUALITY_GATE.md")
	require.NoError(t, err)
	assert.False(t, gatesOrphanExists, "orphaned gates/QUALITY_GATE.md should be deleted after update")

	// Verify the empty gates directory was cleaned up
	gatesDirExists, err := tc.CheckDirExists("/workspace/festivals/.festival/templates/gates")
	require.NoError(t, err)
	assert.False(t, gatesDirExists, "empty gates directory should be removed after update")

	// Step 6: Verify good files still exist
	goodFile1Exists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/festival/GOAL.md")
	require.NoError(t, err)
	assert.True(t, goodFile1Exists, "festival/GOAL.md should still exist after update")

	goodFile2Exists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/phases/implementation/GOAL.md")
	require.NoError(t, err)
	assert.True(t, goodFile2Exists, "phases/implementation/GOAL.md should still exist after update")

	goodFile3Exists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/tasks/TASK.md")
	require.NoError(t, err)
	assert.True(t, goodFile3Exists, "tasks/TASK.md should still exist after update")
}

// TestSystemUpdate_ShowsOrphanedInDryRun verifies that `fest system update --dry-run`
// correctly reports orphaned files that would be deleted.
func TestSystemUpdate_ShowsOrphanedInDryRun(t *testing.T) {
	tc := GetSharedContainer(t)

	// Setup same as above
	_, err := tc.runCommand([]string{"sh", "-c", `
		mkdir -p /root/.config/fest/festivals/.festival/templates/festival
		echo "# Festival Goal" > /root/.config/fest/festivals/.festival/templates/festival/GOAL.md

		mkdir -p /workspace/festivals/.festival/templates/festival
		echo "# Festival Goal" > /workspace/festivals/.festival/templates/festival/GOAL.md
		echo "OLD FILE" > /workspace/festivals/.festival/templates/OLD_ORPHAN.md

		cat > /workspace/festivals/.festival/.fest-checksums.json << 'EOF'
{
  "templates/festival/GOAL.md": {"hash": "abc", "modified": "2024-01-01T00:00:00Z"},
  "templates/OLD_ORPHAN.md": {"hash": "xyz", "modified": "2024-01-01T00:00:00Z"}
}
EOF
	`})
	require.NoError(t, err, "failed to setup test")

	// Run dry-run
	output, err := tc.RunFestInDir("/workspace/festivals", "system", "update", "--dry-run")
	t.Logf("Dry-run output: %s", output)

	// Verify output mentions orphaned files
	assert.Contains(t, output, "Orphaned", "dry-run should report orphaned files")
	assert.Contains(t, output, "1 files", "should show 1 orphaned file")

	// Verify orphaned file still exists (dry-run shouldn't delete)
	orphanExists, err := tc.CheckFileExists("/workspace/festivals/.festival/templates/OLD_ORPHAN.md")
	require.NoError(t, err)
	assert.True(t, orphanExists, "orphaned file should still exist after dry-run")
}
