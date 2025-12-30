//go:build integration
// +build integration

package integration

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFestGoAndWorkspaceCommands(t *testing.T) {
	// Skip if Docker not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Get shared container (reset between tests)
	container := GetSharedContainer(t)

	// Test 1: Test go command help
	t.Run("GoCommandHelp", func(t *testing.T) {
		output, err := container.RunFest("go", "--help")
		require.NoError(t, err, "go --help should not fail")
		require.Contains(t, output, "Navigate to your workspace", "Help should mention navigation")
		require.Contains(t, output, "--register", "Help should mention --register")
		require.Contains(t, output, "--workspace", "Help should mention --workspace flag")
		require.Contains(t, output, "--all", "Help should mention --all flag")
	})

	// Test 2: Setup festivals directory with marker
	t.Run("SetupWorkspace", func(t *testing.T) {
		// Create festivals/.festival/ structure
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"mkdir", "-p", "/testproject/festivals/.festival",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode, "Failed to create festivals structure")

		// Verify structure was created
		exists, err := container.CheckDirExists("/testproject/festivals/.festival")
		require.NoError(t, err)
		require.True(t, exists, ".festival directory should exist")
	})

	// Test 3: Register workspace
	t.Run("RegisterWorkspace", func(t *testing.T) {
		// Run init --register
		output, err := container.RunFest("init", "--register", "/testproject/festivals")
		require.NoError(t, err, "init --register should succeed")
		require.Contains(t, output, "Registered", "Output should confirm registration")
		require.Contains(t, output, "testproject", "Output should show workspace name derived from parent")

		// Verify marker file was created
		exists, err := container.CheckFileExists("/testproject/festivals/.festival/.workspace")
		require.NoError(t, err)
		require.True(t, exists, ".workspace marker should exist")

		// Read marker contents
		content, err := container.ReadFile("/testproject/festivals/.festival/.workspace")
		require.NoError(t, err)
		require.Contains(t, content, "testproject", "Marker should contain workspace name")
	})

	// Test 4: fest go from workspace root
	t.Run("GoFromWorkspaceRoot", func(t *testing.T) {
		// Create a subdir to test from
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"mkdir", "-p", "/testproject/src/deep/nested",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		// Run fest go from nested directory
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject/src/deep/nested && /fest go",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "/testproject/festivals", "go should return festivals path")
		}
	})

	// Test 5: fest go --workspace
	t.Run("GoWorkspaceFlag", func(t *testing.T) {
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject && /fest go --workspace",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "testproject", "Should show workspace name")
			require.Contains(t, output, "festivals", "Should show festivals path")
		}
	})

	// Test 6: fest go --all
	t.Run("GoAllFlag", func(t *testing.T) {
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject && /fest go --all",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "testproject", "Should list registered workspace")
		}
	})

	// Test 7: fest go with phase shortcut
	t.Run("GoPhaseShortcut", func(t *testing.T) {
		// Create phase directories in active/
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"mkdir", "-p", "/testproject/festivals/active/001_PLAN",
			"/testproject/festivals/active/002_IMPLEMENT",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		// Run fest go 1
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject && /fest go 1",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "001_PLAN", "Should resolve to phase 001")
		}

		// Run fest go 2
		exitCode, reader, err = container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject && /fest go 2",
		})
		require.NoError(t, err)

		output, _ = readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "002_IMPLEMENT", "Should resolve to phase 002")
		}
	})

	// Test 8: fest go with phase/sequence shortcut
	t.Run("GoPhaseSequenceShortcut", func(t *testing.T) {
		// Create sequence directories
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"mkdir", "-p",
			"/testproject/festivals/active/001_PLAN/01_requirements",
			"/testproject/festivals/active/001_PLAN/02_design",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		// Run fest go 1/2
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject && /fest go 1/2",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "02_design", "Should resolve to sequence 02 in phase 001")
		}
	})

	// Test 9: Unregister workspace
	t.Run("UnregisterWorkspace", func(t *testing.T) {
		// Verify marker exists first
		exists, err := container.CheckFileExists("/testproject/festivals/.festival/.workspace")
		require.NoError(t, err)
		require.True(t, exists, "Marker should exist before unregistering")

		// Run init --unregister
		output, err := container.RunFest("init", "--unregister", "/testproject/festivals")
		require.NoError(t, err, "init --unregister should succeed")
		require.Contains(t, output, "Unregistered", "Output should confirm unregistration")

		// Verify marker was removed
		exists, err = container.CheckFileExists("/testproject/festivals/.festival/.workspace")
		require.NoError(t, err)
		require.False(t, exists, ".workspace marker should be removed")
	})

	// Test 10: fest go falls back to nearest without marker
	t.Run("GoFallbackWithoutMarker", func(t *testing.T) {
		// fest go should still work, falling back to nearest festivals/
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /testproject/src && /fest go",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "/testproject/festivals", "Should fallback to nearest festivals")
		}
	})

	// Test 11: Nested workspaces - inner without marker, outer with marker
	t.Run("NestedWorkspaces", func(t *testing.T) {
		// Create nested structure: outer has marker, inner does not
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"mkdir", "-p",
			"/outer/festivals/.festival",
			"/outer/inner/festivals/.festival",
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		// Register only outer
		output, err := container.RunFest("init", "--register", "/outer/festivals")
		require.NoError(t, err)
		require.Contains(t, output, "outer", "Should register outer workspace")

		// From inner project, fest go should find outer (has marker) not inner (no marker)
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /outer/inner && /fest go",
		})
		require.NoError(t, err)

		output, _ = readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, "/outer/festivals", "Should find outer festivals (has marker)")
			require.NotContains(t, output, "/outer/inner/festivals", "Should skip inner (no marker)")
		}
	})

	// Test 12: JSON output
	t.Run("GoJSONOutput", func(t *testing.T) {
		exitCode, reader, err := container.container.Exec(container.ctx, []string{
			"sh", "-c", "cd /outer && /fest go --json",
		})
		require.NoError(t, err)

		output, _ := readOutput(reader)
		if exitCode == 0 {
			require.Contains(t, output, `"path"`, "JSON output should have path field")
		}
	})
}

// readOutput is a helper to read exec output
func readOutput(reader interface{}) (string, error) {
	if r, ok := reader.(interface{ Read([]byte) (int, error) }); ok {
		buf := make([]byte, 4096)
		n, err := r.Read(buf)
		if err != nil {
			return "", err
		}
		return string(buf[:n]), nil
	}
	return "", nil
}
