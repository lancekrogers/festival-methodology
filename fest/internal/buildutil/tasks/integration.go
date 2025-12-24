// internal/buildutil/tasks/integration.go
package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/buildutil/ui"
)

// IntegrationResult tracks integration test results
type IntegrationResult struct {
	Suite    string
	Pass     bool
	Duration time.Duration
}

// Integration runs integration tests
func Integration(verbose bool) error {
	ui.Section("Running Integration Tests")

	// Build Linux binary for Docker-based integration tests
	ui.Task("Building", "Linux binary for Docker tests")
	if err := os.MkdirAll("bin/linux", 0o755); err != nil {
		ui.TaskFail()
		return fmt.Errorf("failed to create bin/linux directory: %w", err)
	}

	cmd := exec.Command("go", "build", "-ldflags", "-s -w", "-o", "bin/linux/fest", "./cmd/fest")
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		ui.TaskFail()
		return fmt.Errorf("failed to build Linux binary: %w", err)
	}
	ui.TaskPass()

	suites, err := discoverIntegrationSuites()
	if err != nil {
		return fmt.Errorf("failed to discover integration test suites: %w", err)
	}

	if len(suites) == 0 {
		ui.Status("No integration tests found", true)
		return nil
	}

	if verbose {
		fmt.Printf("Found %d integration test suites\n", len(suites))
	}

	results := make([]IntegrationResult, 0, len(suites))
	total := len(suites)
	failures := 0

	// Run each test suite
	for i, suite := range suites {
		name := strings.TrimPrefix(suite, "tests/integration/")

		ui.Progress(i+1, total, fmt.Sprintf("Testing %s", name))

		start := time.Now()
		cmd := exec.Command("go", "test", "-tags", "integration", "-timeout", "2m", "./"+suite)

		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		pass := cmd.Run() == nil
		duration := time.Since(start)

		results = append(results, IntegrationResult{
			Suite:    name,
			Pass:     pass,
			Duration: duration,
		})

		if !pass {
			failures++
		}
	}

	ui.ClearProgress()

	// Calculate totals
	var totalTime time.Duration
	passed := 0
	for _, r := range results {
		totalTime += r.Duration
		if r.Pass {
			passed++
		}
	}

	// Display summary
	rows := [][]string{
		{"Test Suite", "Status", "Time"},
	}

	for _, r := range results {
		status := "✓ PASS"
		if !r.Pass {
			status = "✗ FAIL"
		}
		if ui.ColourEnabled() {
			if r.Pass {
				status = ui.Green + status + ui.Reset
			} else {
				status = ui.Red + status + ui.Reset
			}
		}

		rows = append(rows, []string{
			r.Suite,
			status,
			fmt.Sprintf("%.2fs", r.Duration.Seconds()),
		})
	}

	// Add totals row
	totalStatus := fmt.Sprintf("%d/%d passed", passed, len(results))
	if ui.ColourEnabled() {
		if failures > 0 {
			totalStatus = ui.Red + totalStatus + ui.Reset
		} else {
			totalStatus = ui.Green + totalStatus + ui.Reset
		}
	}

	rows = append(rows, []string{
		"TOTAL",
		totalStatus,
		fmt.Sprintf("%.2fs", totalTime.Seconds()),
	})

	success := failures == 0

	// Use custom status messages for integration test results
	successMsg := "✓ ALL INTEGRATION TESTS PASSED"
	failMsg := fmt.Sprintf("✗ %d INTEGRATION TESTS FAILED", failures)

	ui.SummaryCardWithStatus("Integration Test Summary", rows, fmt.Sprintf("%.2fs", totalTime.Seconds()), success, successMsg, failMsg)

	if failures > 0 {
		return fmt.Errorf("%d integration test suites failed", failures)
	}

	return nil
}

// discoverIntegrationSuites finds all integration test directories
func discoverIntegrationSuites() ([]string, error) {
	var suites []string

	// Check if tests/integration directory exists
	if _, err := os.Stat("tests/integration"); os.IsNotExist(err) {
		return suites, nil
	}

	// First check if tests/integration itself has test files (flat structure)
	matches, _ := filepath.Glob("tests/integration/*_test.go")
	if len(matches) > 0 {
		suites = append(suites, "tests/integration")
	}

	// Also walk for subdirectories with tests
	err := filepath.Walk("tests/integration", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Look for subdirectories with _test.go files
		if info.IsDir() && path != "tests/integration" {
			// Check if directory has test files
			subMatches, _ := filepath.Glob(filepath.Join(path, "*_test.go"))
			if len(subMatches) > 0 {
				suites = append(suites, path)
			}
		}

		return nil
	})

	return suites, err
}
