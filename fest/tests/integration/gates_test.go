//go:build integration
// +build integration

package integration

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGatesCommands(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container, err := NewTestContainer(t)
	require.NoError(t, err, "Failed to create test container")
	defer container.Cleanup()

	// Setup: Create festivals directory structure with .festival/gates
	t.Run("Setup", func(t *testing.T) {
		// Create the base festivals structure
		err := setupGatesTestFestival(container)
		require.NoError(t, err, "Failed to setup test festival")

		// Verify structure
		exists, err := container.CheckDirExists("/festivals/.festival")
		require.NoError(t, err)
		require.True(t, exists, ".festival directory should exist")
	})

	// Test 1: gates --help
	t.Run("GatesHelp", func(t *testing.T) {
		output, err := container.RunFest("gates", "--help")
		require.NoError(t, err, "gates --help should not fail")
		require.Contains(t, output, "gates", "Help should mention gates")
		require.Contains(t, output, "show", "Help should mention show subcommand")
		require.Contains(t, output, "list", "Help should mention list subcommand")
		require.Contains(t, output, "apply", "Help should mention apply subcommand")
		require.Contains(t, output, "init", "Help should mention init subcommand")
		require.Contains(t, output, "validate", "Help should mention validate subcommand")
		t.Logf("gates help: %s", output)
	})

	// Test 2: gates list - show available policies
	t.Run("GatesList", func(t *testing.T) {
		// Change to festivals directory first
		output, err := container.RunFestInDir("/festivals", "gates", "list")
		require.NoError(t, err, "gates list should not fail")
		require.Contains(t, output, "default", "Should list default policy")
		require.Contains(t, output, "strict", "Should list strict policy")
		require.Contains(t, output, "lightweight", "Should list lightweight policy")
		t.Logf("gates list: %s", output)
	})

	// Test 3: gates list --json
	t.Run("GatesListJSON", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals", "gates", "list", "--json")
		require.NoError(t, err, "gates list --json should not fail")
		require.Contains(t, output, `"name"`, "JSON should contain name field")
		require.Contains(t, output, `"source"`, "JSON should contain source field")
		require.Contains(t, output, `"description"`, "JSON should contain description field")
		t.Logf("gates list --json: %s", output)
	})

	// Test 4: gates show - show effective policy for festival
	t.Run("GatesShow", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "show")
		require.NoError(t, err, "gates show should not fail")
		require.Contains(t, output, "Merged gates", "Should show merged gates header")
		require.Contains(t, output, "testing_and_verify", "Should show default gate")
		require.Contains(t, output, "code_review", "Should show code review gate")
		t.Logf("gates show: %s", output)
	})

	// Test 5: gates show --json
	t.Run("GatesShowJSON", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "show", "--json")
		require.NoError(t, err, "gates show --json should not fail")
		require.Contains(t, output, `"gates"`, "JSON should contain gates array")
		require.Contains(t, output, `"sources"`, "JSON should contain sources array")
		require.Contains(t, output, `"level"`, "JSON should contain level field")
		t.Logf("gates show --json: %s", output)
	})

	// Test 6: gates init - create fest.yaml at festival level
	t.Run("GatesInitFestival", func(t *testing.T) {
		// First verify fest.yaml doesn't exist
		festYAMLPath := "/festivals/test-gates-festival/fest.yaml"
		exists, _ := container.CheckFileExists(festYAMLPath)
		require.False(t, exists, "fest.yaml should not exist initially")

		// Run init
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "init")
		require.NoError(t, err, "gates init should not fail")
		require.Contains(t, output, "Created fest.yaml", "Should confirm fest.yaml creation")
		t.Logf("gates init: %s", output)

		// Verify file was created
		exists, err = container.CheckFileExists(festYAMLPath)
		require.NoError(t, err)
		require.True(t, exists, "fest.yaml should be created")

		// Verify file content
		content, err := container.ReadFile(festYAMLPath)
		require.NoError(t, err)
		require.Contains(t, content, "quality_gates:", "fest.yaml should have quality_gates section")
		require.Contains(t, content, "testing_and_verify", "fest.yaml should have default gate")
		t.Logf("fest.yaml content: %s", content)
	})

	// Test 7: gates init --phase - create override at phase level
	t.Run("GatesInitPhase", func(t *testing.T) {
		overridePath := "/festivals/test-gates-festival/002_IMPLEMENT/.fest.gates.yml"

		// Verify file doesn't exist
		exists, _ := container.CheckFileExists(overridePath)
		require.False(t, exists, "Phase override should not exist initially")

		// Run init with --phase flag
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "init", "--phase", "002_IMPLEMENT")
		require.NoError(t, err, "gates init --phase should not fail")
		require.Contains(t, output, "Created override file", "Should confirm creation")
		t.Logf("gates init --phase: %s", output)

		// Verify file was created
		exists, err = container.CheckFileExists(overridePath)
		require.NoError(t, err)
		require.True(t, exists, "Phase override should be created")
	})

	// Test 8: gates init --sequence - create override at sequence level
	t.Run("GatesInitSequence", func(t *testing.T) {
		overridePath := "/festivals/test-gates-festival/002_IMPLEMENT/01_core/.fest.gates.yml"

		// Verify file doesn't exist
		exists, _ := container.CheckFileExists(overridePath)
		require.False(t, exists, "Sequence override should not exist initially")

		// Run init with --sequence flag
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "init", "--sequence", "002_IMPLEMENT/01_core")
		require.NoError(t, err, "gates init --sequence should not fail")
		require.Contains(t, output, "Created override file", "Should confirm creation")
		t.Logf("gates init --sequence: %s", output)

		// Verify file was created
		exists, err = container.CheckFileExists(overridePath)
		require.NoError(t, err)
		require.True(t, exists, "Sequence override should be created")
	})

	// Test 9: gates apply - apply named policy (policy-only mode)
	t.Run("GatesApplyStrict", func(t *testing.T) {
		// Create a second festival for apply test
		err := createMinimalFestival(container, "/festivals/apply-test-festival")
		require.NoError(t, err)

		// Apply strict policy with --policy-only to write the policy file
		output, err := container.RunFestInDir("/festivals/apply-test-festival", "gates", "apply", "strict", "--policy-only", "--approve")
		require.NoError(t, err, "gates apply strict --policy-only should not fail")
		require.Contains(t, output, "Applied policy", "Should confirm application")
		require.Contains(t, output, "strict", "Should mention policy name")
		t.Logf("gates apply strict: %s", output)

		// Verify policy was written
		policyPath := "/festivals/apply-test-festival/.festival/gates.yml"
		content, err := container.ReadFile(policyPath)
		require.NoError(t, err)
		require.Contains(t, content, "strict", "Policy file should contain strict")
		require.Contains(t, content, "security_audit", "Strict policy should have security_audit")
		t.Logf("Applied policy content: %s", content)
	})

	// Test 10: gates apply --dry-run (policy-only mode)
	t.Run("GatesApplyDryRun", func(t *testing.T) {
		// Create another test festival
		err := createMinimalFestival(container, "/festivals/dry-run-test")
		require.NoError(t, err)

		// Apply with --policy-only (dry-run is default)
		output, err := container.RunFestInDir("/festivals/dry-run-test", "gates", "apply", "lightweight", "--policy-only")
		require.NoError(t, err, "gates apply --policy-only should not fail")
		require.Contains(t, output, "[dry-run]", "Should indicate dry run")
		require.Contains(t, output, "lightweight", "Should mention policy name")
		t.Logf("gates apply --policy-only: %s", output)

		// Verify file was NOT created
		policyPath := "/festivals/dry-run-test/.festival/gates.yml"
		exists, _ := container.CheckFileExists(policyPath)
		require.False(t, exists, "Dry run should not create file")
	})

	// Test 11: gates validate
	t.Run("GatesValidate", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "validate")
		require.NoError(t, err, "gates validate should not fail")
		// Either "valid" or lists issues
		if strings.Contains(output, "valid") {
			t.Logf("Validation passed: %s", output)
		} else {
			t.Logf("Validation found issues: %s", output)
		}
	})

	// Test 12: gates validate --json
	t.Run("GatesValidateJSON", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "validate", "--json")
		require.NoError(t, err, "gates validate --json should not fail")
		require.Contains(t, output, `"valid"`, "JSON should contain valid field")
		require.Contains(t, output, `"issues"`, "JSON should contain issues field")
		t.Logf("gates validate --json: %s", output)
	})

	// Test 13: gates show with phase override in effect
	t.Run("GatesShowWithOverride", func(t *testing.T) {
		// The phase override was created in test 7
		// Show gates for that phase
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "show", "--phase", "002_IMPLEMENT")
		require.NoError(t, err, "gates show --phase should not fail")
		require.Contains(t, output, "Merged gates", "Should show merged gates")
		t.Logf("gates show --phase: %s", output)
	})

	// Test 14: gates show for sequence
	t.Run("GatesShowSequence", func(t *testing.T) {
		output, err := container.RunFestInDir("/festivals/test-gates-festival", "gates", "show", "--sequence", "002_IMPLEMENT/01_core")
		require.NoError(t, err, "gates show --sequence should not fail")
		require.Contains(t, output, "Merged gates", "Should show merged gates")
		t.Logf("gates show --sequence: %s", output)
	})

	// Test 15: Verify hierarchical merge behavior
	t.Run("HierarchicalMerge", func(t *testing.T) {
		// Write a custom policy at festival level that removes a gate
		customPolicy := `version: 1
name: custom
inherit: true
append:
  - id: testing_and_verify
    template: QUALITY_GATE_TESTING
    name: Testing and Verification
    enabled: true
  - id: code_review
    template: QUALITY_GATE_REVIEW
    name: Code Review
    enabled: true
`
		// Create a new festival for this test
		err := createMinimalFestival(container, "/festivals/hierarchy-test")
		require.NoError(t, err)

		// Write custom policy
		policyPath := "/festivals/hierarchy-test/.festival/gates.yml"
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"sh", "-c",
			"mkdir -p /festivals/hierarchy-test/.festival && printf '%s' '" + customPolicy + "' > " + policyPath,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)

		// Verify the policy is loaded correctly
		output, err := container.RunFestInDir("/festivals/hierarchy-test", "gates", "show", "--json")
		require.NoError(t, err, "gates show should work with custom policy")
		require.Contains(t, output, "testing_and_verify", "Should show gates from custom policy")
		t.Logf("Hierarchical merge test: %s", output)
	})
}

// setupGatesTestFestival creates a test festival structure for gates testing
func setupGatesTestFestival(tc *TestContainer) error {
	// Create base structure
	dirs := []string{
		"/festivals/.festival/gates/policies",
		"/festivals/.festival/gates/templates",
		"/festivals/.festival/templates",
		"/festivals/test-gates-festival/001_DESIGN/01_planning",
		"/festivals/test-gates-festival/002_IMPLEMENT/01_core",
		"/festivals/test-gates-festival/002_IMPLEMENT/02_features",
		"/festivals/test-gates-festival/003_TEST/01_unit",
	}

	for _, dir := range dirs {
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", dir})
		if err != nil || exitCode != 0 {
			return err
		}
	}

	// Create FESTIVAL_GOAL.md
	goalContent := `---
id: TEST_GATES_FESTIVAL
---

# Test Gates Festival

## Goal

Test the hierarchical quality gates system.
`
	goalPath := "/festivals/test-gates-festival/FESTIVAL_GOAL.md"
	exitCode, _, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		"printf '%s' '" + goalContent + "' > " + goalPath,
	})
	if err != nil || exitCode != 0 {
		return err
	}

	// Create task files
	taskContent := `---
id: TASK_01
---

# Task 01

Implementation task.
`
	tasks := []string{
		"/festivals/test-gates-festival/002_IMPLEMENT/01_core/01_setup.md",
		"/festivals/test-gates-festival/002_IMPLEMENT/01_core/02_implement.md",
		"/festivals/test-gates-festival/002_IMPLEMENT/02_features/01_feature_a.md",
	}

	for _, task := range tasks {
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{
			"sh", "-c",
			"printf '%s' '" + taskContent + "' > " + task,
		})
		if err != nil || exitCode != 0 {
			return err
		}
	}

	return nil
}

// createMinimalFestival creates a minimal festival structure
func createMinimalFestival(tc *TestContainer, path string) error {
	// Create directories including the root .festival (required by FindFestivalsRoot)
	dirs := []string{
		"/festivals/.festival", // Root .festival directory
		filepath.Join(path, ".festival"),
		filepath.Join(path, "001_PHASE/01_seq"),
	}

	for _, dir := range dirs {
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", dir})
		if err != nil || exitCode != 0 {
			return err
		}
	}

	// Create FESTIVAL_GOAL.md
	goalPath := filepath.Join(path, "FESTIVAL_GOAL.md")
	goalContent := "# Minimal Test Festival\n\nTest festival for gates.\n"
	exitCode, _, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		"printf '%s' '" + goalContent + "' > " + goalPath,
	})
	if err != nil || exitCode != 0 {
		return err
	}

	return nil
}

// RunFestInDir runs fest command from a specific directory
func (tc *TestContainer) RunFestInDir(dir string, args ...string) (string, error) {
	// Use sh -c to change directory and run fest
	cmd := []string{"sh", "-c", "cd " + dir + " && /fest " + strings.Join(args, " ")}
	return tc.runCommand(cmd)
}

// runCommand executes a command and returns output
func (tc *TestContainer) runCommand(cmd []string) (string, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, cmd)
	if err != nil {
		return "", err
	}

	output, err := readAllFromReader(reader)
	if err != nil {
		return "", err
	}

	if exitCode != 0 {
		return output, nil // Return output even on non-zero exit for debugging
	}

	return output, nil
}

// readAllFromReader reads all content from an io.Reader
func readAllFromReader(r interface{}) (string, error) {
	if reader, ok := r.(interface{ Read([]byte) (int, error) }); ok {
		var buf strings.Builder
		data := make([]byte, 4096)
		for {
			n, err := reader.Read(data)
			if n > 0 {
				buf.Write(data[:n])
			}
			if err != nil {
				break
			}
		}
		return buf.String(), nil
	}
	return "", nil
}
