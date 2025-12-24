// internal/buildutil/tasks/build.go
package tasks

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/buildutil/ui"
)

// PackageResult tracks build results for a package
type PackageResult struct {
	Package   string
	VetPass   bool
	BuildPass bool
	VetTime   time.Duration
	BuildTime time.Duration
}

// Build runs go vet and go build on all packages
func Build(verbose bool) error {
	ui.Section("Building Fest CLI")

	packages, err := discoverPackages()
	if err != nil {
		return fmt.Errorf("failed to discover packages: %w", err)
	}

	if verbose {
		fmt.Printf("Found %d packages\n", len(packages))
	}

	results := make([]PackageResult, 0, len(packages))
	total := len(packages)

	// Process each package
	for i, pkg := range packages {
		shortName := strings.TrimPrefix(pkg, "./")
		if shortName == "." {
			shortName = "root"
		}

		result := PackageResult{Package: shortName}

		// Vet
		ui.Progress(i+1, total, fmt.Sprintf("Vetting %s", shortName))
		start := time.Now()
		cmd := exec.Command("go", "vet", pkg)
		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		result.VetPass = cmd.Run() == nil
		result.VetTime = time.Since(start)

		if !result.VetPass {
			// Don't return immediately - continue to collect all results for dashboard
		}

		// Build
		ui.Progress(i+1, total, fmt.Sprintf("Building %s", shortName))
		start = time.Now()
		cmd = exec.Command("go", "build", "-o", "/dev/null", pkg)
		if verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		result.BuildPass = cmd.Run() == nil
		result.BuildTime = time.Since(start)

		if !result.BuildPass {
			// Don't return immediately - continue to collect all results for dashboard
		}

		results = append(results, result)
	}

	// Build main binary
	ui.Progress(total, total, "Building main binary")
	start := time.Now()

	// Create bin directory
	os.MkdirAll("bin", 0o755)

	cmd := exec.Command("go", "build", "-o", "bin/fest", "./cmd/fest")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	mainBuildSuccess := cmd.Run() == nil
	mainBuildTime := time.Since(start)

	ui.ClearProgress()

	// Don't return immediately - add to results for dashboard display

	// Add main binary result
	results = append(results, PackageResult{
		Package:   "bin/fest",
		VetPass:   true,
		BuildPass: mainBuildSuccess,
		BuildTime: mainBuildTime,
	})

	// Calculate totals
	var totalTime time.Duration
	for _, r := range results {
		totalTime += r.VetTime + r.BuildTime
	}

	// Display summary - only show packages with errors
	rows := [][]string{}
	hasFailures := false

	for _, r := range results {
		// Only include packages that have failures
		if !r.VetPass || !r.BuildPass {
			hasFailures = true

			vetStatus := "✓"
			if !r.VetPass {
				vetStatus = "✗"
			}
			if ui.ColourEnabled() {
				if r.VetPass {
					vetStatus = ui.Green + vetStatus + ui.Reset
				} else {
					vetStatus = ui.Red + vetStatus + ui.Reset
				}
			}

			buildStatus := "✓"
			if !r.BuildPass {
				buildStatus = "✗"
			}
			if ui.ColourEnabled() {
				if r.BuildPass {
					buildStatus = ui.Green + buildStatus + ui.Reset
				} else {
					buildStatus = ui.Red + buildStatus + ui.Reset
				}
			}

			rows = append(rows, []string{
				r.Package,
				fmt.Sprintf("%s %.2fs", vetStatus, r.VetTime.Seconds()),
				fmt.Sprintf("%s %.2fs", buildStatus, r.BuildTime.Seconds()),
			})
		}
	}

	// Add header only if there are failures to show
	if hasFailures {
		rows = append([][]string{{"Package", "Vet", "Build"}}, rows...)
	}

	// Choose appropriate title based on whether there are failures
	title := "Build Summary"
	if hasFailures {
		title = "Build Failures"
	} else {
		title = "Build Complete - No Errors"
	}

	ui.SummaryCard(title, rows, fmt.Sprintf("%.2fs", totalTime.Seconds()), !hasFailures)

	// Now return error if there were any failures
	if hasFailures {
		failedPackages := []string{}
		for _, r := range results {
			if !r.VetPass || !r.BuildPass {
				failedPackages = append(failedPackages, r.Package)
			}
		}
		return fmt.Errorf("build failed for packages: %s", strings.Join(failedPackages, ", "))
	}

	return nil
}

// discoverPackages finds all Go packages in the project
func discoverPackages() ([]string, error) {
	cmd := exec.Command("go", "list", "./...")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	module := getModuleName()

	var packages []string
	for _, line := range lines {
		if line != "" &&
			!strings.Contains(line, "/vendor/") &&
			!strings.Contains(line, "/testdata") &&
			!strings.Contains(line, "/tests/integration") &&
			!strings.HasSuffix(line, "/tests/integration") &&
			!strings.Contains(line, "_test") &&
			!strings.HasSuffix(line, "/test/benchmark") &&
			!strings.HasSuffix(line, "/test/e2e") {
			// Convert full module paths to relative paths
			if module != "" && strings.HasPrefix(line, module) {
				relativePath := strings.TrimPrefix(line, module)
				if relativePath == "" {
					packages = append(packages, ".")
				} else if strings.HasPrefix(relativePath, "/") {
					packages = append(packages, "."+relativePath)
				}
			}
		}
	}

	return packages, nil
}

// getModuleName reads the module name from go.mod
func getModuleName() string {
	data, err := os.ReadFile("go.mod")
	if err != nil {
		return ""
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

// BuildOnly builds the main fest binary without running go vet (fast user installation)
func BuildOnly(verbose bool) error {
	ui.Section("Building Fest CLI")

	// Create bin directory
	if err := os.MkdirAll("bin", 0o755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	ui.Task("Building", "fest binary")

	// Build main binary only
	cmd := exec.Command("go", "build", "-o", "bin/fest", "./cmd/fest")
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		ui.TaskFail()
		return fmt.Errorf("failed to build fest binary: %w", err)
	}

	ui.TaskPass()

	// Show summary
	ui.SummaryCard(
		"Build Complete - No Errors",
		[][]string{
			{"Task", "Status"},
			{"Binary Build", ui.Green + "✓ Complete" + ui.Reset},
		},
		"< 5s",
		true,
	)

	return nil
}
