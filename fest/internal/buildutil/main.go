// internal/buildutil/main.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/buildutil/tasks"
	"github.com/lancekrogers/festival-methodology/fest/internal/buildutil/ui"
)

var (
	noColor bool
	verbose bool
)

func main() {
	flag.BoolVar(&noColor, "no-color", false, "disable ANSI colours")
	flag.BoolVar(&verbose, "v", false, "verbose output")
	flag.Parse()

	// Initialize UI with color preferences
	ui.Init(noColor)

	if flag.NArg() == 0 {
		log.Fatalf("usage: buildutil <build|build-only|test|integration|clean|all>")
	}

	cmd := flag.Arg(0)
	startTime := time.Now()

	// Hide cursor during operations
	if ui.ColourEnabled() {
		fmt.Print(ui.HideCursor)
		defer fmt.Print(ui.ShowCursor)
	}

	var err error

	switch cmd {
	case "build":
		err = tasks.Build(verbose)

	case "build-only":
		err = tasks.BuildOnly(verbose)

	case "test":
		err = tasks.Test(verbose)

	case "integration":
		err = tasks.Integration(verbose)

	case "clean":
		err = tasks.Clean(verbose)

	case "all":
		// Run all tasks in sequence
		var errors []error

		fmt.Println("\n🧹 Cleaning...")
		if cleanErr := tasks.Clean(verbose); cleanErr != nil {
			errors = append(errors, fmt.Errorf("clean failed: %w", cleanErr))
		}

		fmt.Println("\n🔨 Building...")
		if buildErr := tasks.Build(verbose); buildErr != nil {
			errors = append(errors, fmt.Errorf("build failed: %w", buildErr))
			// Don't continue if build fails - can't test broken code
			err = fmt.Errorf("stopping due to build failure: %w", buildErr)
			break
		}

		fmt.Println("\n🧪 Testing...")
		if testErr := tasks.Test(verbose); testErr != nil {
			errors = append(errors, fmt.Errorf("tests failed: %w", testErr))
			// Continue to integration tests even if unit tests fail
		}

		fmt.Println("\n🔗 Integration Testing...")
		if integrationErr := tasks.Integration(verbose); integrationErr != nil {
			errors = append(errors, fmt.Errorf("integration tests failed: %w", integrationErr))
		}

		// Set overall error if any step failed
		if len(errors) > 0 {
			err = fmt.Errorf("%d tasks failed", len(errors))
		}

		// Show overall summary
		if err == nil {
			totalTime := time.Since(startTime)
			cleanStatus := "✓ Complete"
			buildStatus := "✓ Complete"
			testStatus := "✓ Complete"
			integrationStatus := "✓ Complete"

			if ui.ColourEnabled() {
				cleanStatus = ui.Green + cleanStatus + ui.Reset
				buildStatus = ui.Green + buildStatus + ui.Reset
				testStatus = ui.Green + testStatus + ui.Reset
				integrationStatus = ui.Green + integrationStatus + ui.Reset
			}

			rows := [][]string{
				{"Task", "Status"},
				{"Clean", cleanStatus},
				{"Build", buildStatus},
				{"Test", testStatus},
				{"Integration", integrationStatus},
			}
			ui.SummaryCard("All Tasks Complete", rows, fmt.Sprintf("%.2fs", totalTime.Seconds()), true)
		}

	default:
		log.Fatalf("unknown command %q", cmd)
	}

	if err != nil {
		if ui.ColourEnabled() {
			fmt.Printf("\n%s✗ Error: %v%s\n", ui.Red, err, ui.Reset)
		} else {
			fmt.Printf("\nError: %v\n", err)
		}
		os.Exit(1)
	}
}
