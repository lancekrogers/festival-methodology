//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestContainer wraps container operations for testing
type TestContainer struct {
	container testcontainers.Container
	ctx       context.Context
	t         *testing.T
}

// NewTestContainer creates a new Alpine container for testing fest
func NewTestContainer(t *testing.T) (*TestContainer, error) {
	ctx := context.Background()

	// Get the absolute path to the Linux binary
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	festBinaryPath := filepath.Join(cwd, "../../bin/linux", "fest")
	festBinaryPath, err = filepath.Abs(festBinaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Create Linux binary directory if it doesn't exist
	linuxBinDir := filepath.Dir(festBinaryPath)
	if err := os.MkdirAll(linuxBinDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create linux bin directory: %w", err)
	}

	// Check if binary exists
	if _, err := os.Stat(festBinaryPath); err != nil {
		return nil, fmt.Errorf("fest binary not found at %s: %w - run 'just build-test-binary' first", festBinaryPath, err)
	}

	req := testcontainers.ContainerRequest{
		Image:      "alpine:latest",
		Cmd:        []string{"sleep", "3600"}, // Keep container running
		WaitingFor: wait.ForExec([]string{"true"}).WithStartupTimeout(30 * time.Second),
		AutoRemove: true,
		Mounts: testcontainers.ContainerMounts{
			{
				Source:   testcontainers.GenericBindMountSource{HostPath: festBinaryPath},
				Target:   "/fest",
				ReadOnly: false,
			},
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	// Make fest executable in container
	exitCode, output, err := container.Exec(ctx, []string{"chmod", "+x", "/fest"})
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to make fest executable: %w", err)
	}
	if exitCode != 0 {
		outputBytes, _ := io.ReadAll(output)
		container.Terminate(ctx)
		return nil, fmt.Errorf("chmod failed with exit code %d, output: %s", exitCode, string(outputBytes))
	}

	return &TestContainer{
		container: container,
		ctx:       ctx,
		t:         t,
	}, nil
}

// RunFest executes the fest command in the container
func (tc *TestContainer) RunFest(args ...string) (string, error) {
	cmd := append([]string{"/fest"}, args...)

	exitCode, reader, err := tc.container.Exec(tc.ctx, cmd)
	if err != nil {
		return "", fmt.Errorf("failed to execute fest: %w", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	if exitCode != 0 {
		return string(output), fmt.Errorf("fest exited with code %d: %s", exitCode, output)
	}

	return string(output), nil
}

// CopyToContainer copies a file to the container
func (tc *TestContainer) CopyToContainer(sourcePath, targetPath string) error {
	fileContent, err := os.ReadFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to read source file: %w", err)
	}

	return tc.container.CopyToContainer(
		tc.ctx,
		fileContent,
		targetPath,
		0644,
	)
}

// CopyDirToContainer copies a directory to the container
func (tc *TestContainer) CopyDirToContainer(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			// Create directory in container
			exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", targetPath})
			if err != nil || exitCode != 0 {
				return fmt.Errorf("failed to create directory %s: %w", targetPath, err)
			}
			return nil
		}

		// Copy file
		return tc.CopyToContainer(path, targetPath)
	})
}

// ListDirectory lists files in a container directory recursively
func (tc *TestContainer) ListDirectory(path string) ([]string, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{"find", path, "-type", "f"})
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	if exitCode != 0 {
		return nil, fmt.Errorf("find command failed with exit code %d", exitCode)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read output: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var files []string
	for _, line := range lines {
		if line != "" && line != path {
			files = append(files, line)
		}
	}

	return files, nil
}

// ReadFile reads a file from the container
func (tc *TestContainer) ReadFile(path string) (string, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{"cat", path})
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to read output: %w", err)
	}

	if exitCode != 0 {
		return "", fmt.Errorf("cat command failed with exit code %d: %s", exitCode, output)
	}

	return string(output), nil
}

// CheckFileExists checks if a file exists in the container
func (tc *TestContainer) CheckFileExists(path string) (bool, error) {
	exitCode, _, err := tc.container.Exec(tc.ctx, []string{"test", "-f", path})
	if err != nil {
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	return exitCode == 0, nil
}

// CheckDirExists checks if a directory exists in the container
func (tc *TestContainer) CheckDirExists(path string) (bool, error) {
	exitCode, _, err := tc.container.Exec(tc.ctx, []string{"test", "-d", path})
	if err != nil {
		return false, fmt.Errorf("failed to check directory: %w", err)
	}

	return exitCode == 0, nil
}

// FileSystemState captures the state of a directory tree
type FileSystemState struct {
	Directories []string          // directory paths
	Files       map[string]string // path -> content
}

// CaptureState captures the complete filesystem state in the container
func (tc *TestContainer) CaptureState(rootPath string) (*FileSystemState, error) {
	state := &FileSystemState{
		Files:       make(map[string]string),
		Directories: []string{},
	}

	// Find all directories
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{"find", rootPath, "-type", "d"})
	if err != nil {
		return nil, fmt.Errorf("failed to find directories: %w", err)
	}

	if exitCode == 0 {
		output, _ := io.ReadAll(reader)
		for _, dir := range strings.Split(strings.TrimSpace(string(output)), "\n") {
			if dir != "" && dir != rootPath {
				relPath, _ := filepath.Rel(rootPath, dir)
				if relPath != "." && relPath != "" {
					state.Directories = append(state.Directories, relPath)
				}
			}
		}
	}

	// Find all files and read their content
	files, err := tc.ListDirectory(rootPath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		content, err := tc.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}

		relPath, _ := filepath.Rel(rootPath, file)
		state.Files[relPath] = content
	}

	return state, nil
}

// CountPhases counts the number of phases in a festival directory
func (tc *TestContainer) CountPhases(festivalPath string) (int, error) {
	// First check what directories exist
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		fmt.Sprintf("find %s -maxdepth 1 -type d -name '[0-9][0-9][0-9]_*' | wc -l", festivalPath),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count phases: %w", err)
	}

	if exitCode != 0 {
		return 0, nil // No phases found
	}

	output, _ := io.ReadAll(reader)
	count := 0
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count, nil
}

// CountSequences counts the number of sequences in a phase
func (tc *TestContainer) CountSequences(phasePath string) (int, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		fmt.Sprintf("ls -d %s/[0-9][0-9]_* 2>/dev/null | wc -l", phasePath),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count sequences: %w", err)
	}

	if exitCode != 0 {
		return 0, nil // No sequences found
	}

	output, _ := io.ReadAll(reader)
	count := 0
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count, nil
}

// CountTasks counts the number of tasks in a sequence
func (tc *TestContainer) CountTasks(sequencePath string) (int, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		fmt.Sprintf("ls %s/[0-9][0-9]_*.md 2>/dev/null | wc -l", sequencePath),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	if exitCode != 0 {
		return 0, nil // No tasks found
	}

	output, _ := io.ReadAll(reader)
	count := 0
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count, nil
}

// VerifyParallelItems checks that parallel tasks/sequences with the same number exist
func (tc *TestContainer) VerifyParallelItems(path string, prefix string) (int, error) {
	exitCode, reader, err := tc.container.Exec(tc.ctx, []string{
		"sh", "-c",
		fmt.Sprintf("ls %s/%s* 2>/dev/null | wc -l", path, prefix),
	})
	if err != nil {
		return 0, fmt.Errorf("failed to verify parallel items: %w", err)
	}

	if exitCode != 0 {
		return 0, nil // No items found
	}

	output, _ := io.ReadAll(reader)
	count := 0
	fmt.Sscanf(strings.TrimSpace(string(output)), "%d", &count)
	return count, nil
}

// VerifyStructure validates the directory structure is correct
func (tc *TestContainer) VerifyStructure(festivalPath string) error {
	// Check if festival directory exists
	exists, err := tc.CheckDirExists(festivalPath)
	if err != nil {
		return fmt.Errorf("failed to check festival directory: %w", err)
	}
	if !exists {
		return fmt.Errorf("festival directory does not exist: %s", festivalPath)
	}

	// Get all phases
	phaseCount, err := tc.CountPhases(festivalPath)
	if err != nil {
		return fmt.Errorf("failed to count phases: %w", err)
	}

	if phaseCount == 0 {
		// Valid - empty festival
		return nil
	}

	// Check that phases are sequentially numbered
	for i := 1; i <= phaseCount; i++ {
		phasePattern := fmt.Sprintf("%s/%03d_*", festivalPath, i)
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{
			"sh", "-c",
			fmt.Sprintf("ls -d %s 2>/dev/null | head -1", phasePattern),
		})
		if err != nil || exitCode != 0 {
			return fmt.Errorf("phase %03d not found or not sequential", i)
		}
	}

	return nil
}

// Cleanup terminates the container
func (tc *TestContainer) Cleanup() {
	if tc.container != nil {
		tc.container.Terminate(tc.ctx)
	}
}

// ValidateSnapshot compares actual state with expected state
func ValidateSnapshot(t *testing.T, actual, expected *FileSystemState) {
	// Check directories
	require.ElementsMatch(t, expected.Directories, actual.Directories, "directory structure mismatch")

	// Check files
	require.Equal(t, len(expected.Files), len(actual.Files), "file count mismatch")

	for path, expectedContent := range expected.Files {
		actualContent, exists := actual.Files[path]
		require.True(t, exists, "missing file: %s", path)
		require.Equal(t, expectedContent, actualContent, "content mismatch in file: %s", path)
	}
}

// CreateFestivalGoalFile creates a basic FESTIVAL_GOAL.md content
func CreateFestivalGoalFile() string {
	return `---
id: COMPLEX_TEST_FESTIVAL
---

# Complex Test Festival

## Goal
Test the fest CLI with a complex, realistic festival structure including:
- Multiple implementation phases
- Parallel sequences and tasks
- Deep nesting (3 levels)
- Renumbering operations
- Phase/sequence removal

## Success Criteria
- All fest commands work correctly in isolation
- Renumbering handles parallel items correctly
- Structure remains valid after operations
`
}

// CreatePhaseGoalFile creates a PHASE_GOAL.md for a given phase
func CreatePhaseGoalFile(phaseName string) string {
	return fmt.Sprintf(`---
id: %s_PHASE
---

# %s Phase

## Objective
Complete all %s tasks and sequences.

## Sequences
Multiple sequences with parallel execution paths.
`, strings.ToUpper(phaseName), phaseName, strings.ToLower(phaseName))
}

// CreateTaskFile creates a task markdown file
func CreateTaskFile(taskName string) string {
	return fmt.Sprintf(`---
id: %s
---

# %s

## Task Description
This is a test task for %s.

## Implementation
- Step 1
- Step 2
- Step 3

## Verification
- Test passes
- Code reviewed
`, strings.ToUpper(taskName), taskName, taskName)
}