// internal/buildutil/tasks/clean.go
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

// Clean removes build artifacts
func Clean(verbose bool) error {
	ui.Section("Cleaning Build Artifacts")

	artifacts := []string{
		"bin/",
		"*.test",
		"*.exe",
		"coverage.out",
		"coverage_*.out",
		".test-*",
		"*.disabled",
		"*.old",
		"*.wip",
		"*.backup",
		"*.tmp",
		"*.bak",
		"*~",
		".test-results.tmp",
		".test-timing.tmp",
		"coverage.html",
	}

	total := len(artifacts)
	removed := 0

	for i, pattern := range artifacts {
		ui.Progress(i+1, total, fmt.Sprintf("Removing %s", pattern))

		if strings.Contains(pattern, "*") {
			// Use shell expansion for patterns
			cmd := exec.Command("sh", "-c", fmt.Sprintf("rm -rf %s 2>/dev/null || true", pattern))
			if verbose {
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
			}
			cmd.Run()
			removed++
		} else {
			// Direct removal for specific files/directories
			if err := os.RemoveAll(pattern); err == nil {
				removed++
			}
		}

		time.Sleep(50 * time.Millisecond) // Small delay for visual effect
	}

	ui.ClearProgress()

	// Also clean up any .test binaries in subdirectories
	filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip vendor and .git directories
		if info.IsDir() && (info.Name() == "vendor" || info.Name() == ".git") {
			return filepath.SkipDir
		}

		// Remove .test files
		if strings.HasSuffix(info.Name(), ".test") {
			os.Remove(path)
			removed++
		}

		return nil
	})

	// Display summary
	removeStatus := fmt.Sprintf("✓ %d items removed", removed)
	cleanStatus := "✓ Complete"

	if ui.ColourEnabled() {
		removeStatus = ui.Green + removeStatus + ui.Reset
		cleanStatus = ui.Green + cleanStatus + ui.Reset
	}

	rows := [][]string{
		{"Action", "Status"},
		{"Remove build artifacts", removeStatus},
		{"Clean workspace", cleanStatus},
	}

	ui.SummaryCardWithStatus("Clean Summary", rows, "< 1s", true, "✓ CLEAN SUCCESSFUL", "✗ CLEAN FAILED")

	return nil
}
