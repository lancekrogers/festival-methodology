// internal/buildutil/tasks/integration.go
package tasks

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/buildutil/ui"
)

// integrationTestEvent represents a go test -json output line
type integrationTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
	Output  string `json:"Output"`
}

// IntegrationResult tracks integration test results
type IntegrationResult struct {
	Suite       string
	Pass        bool
	Duration    time.Duration
	TestsPassed int
	TestsFailed int
	FailedTests []string // Names of failed tests
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
		if name == "" {
			name = "tests/integration"
		}

		start := time.Now()

		var pass bool
		var testsPassed, testsFailed int
		var failedTests []string

		// Docker environment variables for Colima/testcontainers compatibility
		dockerEnv := append(os.Environ(),
			"DOCKER_HOST=unix://"+os.Getenv("HOME")+"/.colima/default/docker.sock",
			"TESTCONTAINERS_RYUK_DISABLED=true",
		)

		if verbose {
			// In verbose mode, show output directly
			cmd := exec.Command("go", "test", "-v", "-tags", "integration", "-timeout", "2m", "./"+suite)
			cmd.Env = dockerEnv
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			ui.Progress(i+1, total, fmt.Sprintf("Testing %s", name))
			pass = cmd.Run() == nil
		} else {
			// Run with -json for real-time progress
			cmd := exec.Command("go", "test", "-json", "-tags", "integration", "-timeout", "2m", "./"+suite)
			cmd.Env = dockerEnv
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				return fmt.Errorf("failed to create stdout pipe: %w", err)
			}

			if err := cmd.Start(); err != nil {
				return fmt.Errorf("failed to start test: %w", err)
			}

			// Track state for progress display
			var currentTest string
			var currentOutput string
			var mu sync.Mutex

			// Spinner characters for visual feedback
			spinnerChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinnerIdx := 0

			// Print initial two lines (will be updated in place)
			fmt.Println("  → Starting...")
			fmt.Printf("[%d/%d] ⠋ Starting... 0s", i+1, total)

			// Start a goroutine to update elapsed time
			done := make(chan bool)
			go func() {
				ticker := time.NewTicker(100 * time.Millisecond)
				defer ticker.Stop()
				for {
					select {
					case <-done:
						return
					case <-ticker.C:
						mu.Lock()
						elapsed := time.Since(start).Seconds()
						testName := currentTest
						output := currentOutput
						passed := testsPassed
						failed := testsFailed
						mu.Unlock()

						// Cycle spinner
						spinner := spinnerChars[spinnerIdx%len(spinnerChars)]
						spinnerIdx++

						status := fmt.Sprintf("%d✓", passed)
						if failed > 0 {
							status += fmt.Sprintf(" %d✗", failed)
						}

						var progressLine string
						if testName != "" {
							progressLine = fmt.Sprintf("%s %s (%s) %.0fs", spinner, testName, status, elapsed)
						} else {
							progressLine = fmt.Sprintf("%s Starting... %.0fs", spinner, elapsed)
						}

						if output == "" {
							output = "waiting for output..."
						}
						ui.ProgressWithOutput(i+1, total, output, progressLine)
					}
				}
			}()

			// Parse JSON output in real-time
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				var event integrationTestEvent
				if err := json.Unmarshal(scanner.Bytes(), &event); err != nil {
					continue
				}

				mu.Lock()
				// Capture output lines (strip newlines for display)
				if event.Action == "output" && event.Output != "" {
					trimmed := strings.TrimSpace(event.Output)
					// Show subtest runs and actual test output (not just framework markers)
					if strings.HasPrefix(trimmed, "=== RUN") {
						// Show subtest being run
						currentOutput = strings.TrimPrefix(trimmed, "=== RUN   ")
					} else if trimmed != "" && !strings.HasPrefix(trimmed, "---") && !strings.HasPrefix(trimmed, "PASS") && !strings.HasPrefix(trimmed, "FAIL") {
						// Show actual test output
						currentOutput = trimmed
					}
				}

				// Track all tests (including subtests for failure reporting)
				if event.Test != "" {
					switch event.Action {
					case "run":
						if !strings.Contains(event.Test, "/") {
							currentTest = event.Test
						}
					case "pass":
						if !strings.Contains(event.Test, "/") {
							testsPassed++
						}
					case "fail":
						// Track all failed tests (including subtests)
						failedTests = append(failedTests, event.Test)
						if !strings.Contains(event.Test, "/") {
							testsFailed++
						}
					}
				}
				mu.Unlock()
			}

			close(done)
			cmd.Wait()
			pass = testsFailed == 0
			ui.ClearProgressWithOutput()
		}

		duration := time.Since(start)

		results = append(results, IntegrationResult{
			Suite:       name,
			Pass:        pass,
			Duration:    duration,
			TestsPassed: testsPassed,
			TestsFailed: testsFailed,
			FailedTests: failedTests,
		})

		if !pass {
			failures++
		}
	}

	ui.ClearProgress()

	// Calculate totals
	var totalTime time.Duration
	totalTestsPassed := 0
	totalTestsFailed := 0
	for _, r := range results {
		totalTime += r.Duration
		totalTestsPassed += r.TestsPassed
		totalTestsFailed += r.TestsFailed
	}
	totalTests := totalTestsPassed + totalTestsFailed

	// Display summary - show failed tests as individual rows
	rows := [][]string{}
	hasFailures := failures > 0

	for _, r := range results {
		if !r.Pass && len(r.FailedTests) > 0 {
			// Show each failed test as a row
			for _, testName := range r.FailedTests {
				status := "✗ FAILED"
				if ui.ColourEnabled() {
					status = ui.Red + status + ui.Reset
				}
				rows = append(rows, []string{
					testName,
					status,
					"",
				})
			}
		}
	}

	// Add header only if there are failures to show
	if hasFailures && len(rows) > 0 {
		rows = append([][]string{{"Failed Test", "Status", ""}}, rows...)
	}

	// Add totals row
	totalStatus := fmt.Sprintf("%d/%d tests passed", totalTestsPassed, totalTests)
	if ui.ColourEnabled() {
		if totalTestsFailed > 0 {
			totalStatus = ui.Red + totalStatus + ui.Reset
		} else {
			totalStatus = ui.Green + totalStatus + ui.Reset
		}
	}

	rows = append(rows, []string{
		fmt.Sprintf("%d suites", len(results)),
		totalStatus,
		fmt.Sprintf("%.2fs", totalTime.Seconds()),
	})

	success := failures == 0

	// Use custom status messages for integration test results
	successMsg := fmt.Sprintf("✓ ALL %d TESTS PASSED", totalTestsPassed)
	failMsg := fmt.Sprintf("✗ %d/%d TESTS FAILED", totalTestsFailed, totalTests)

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
