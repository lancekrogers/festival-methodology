package understand

import (
	"fmt"
	"path/filepath"

	understanddocs "github.com/lancekrogers/festival-methodology/fest/docs/understand"
	"github.com/spf13/cobra"
)

func newUnderstandMethodologyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "methodology",
		Short: "Core principles - START HERE for new agents",
		Long: `Learn the core principles of Festival Methodology.

This is the STARTING POINT for agents new to Festival Methodology.
Covers goal-oriented development, requirements-driven implementation,
and quality gates.

After reading this, proceed to:
  fest understand structure   - Learn the 3-level hierarchy
  fest understand tasks       - Learn when to create task files`,
		Run: func(cmd *cobra.Command, args []string) {
			printMethodology(findDotFestivalDir())
		},
	}
}

func printMethodology(dotFestival string) {
	// Hybrid: try .festival/ supplements first, fall back to embedded defaults
	if dotFestival != "" {
		docPath := filepath.Join(dotFestival, "FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md")
		if content := readFileContent(docPath); content != "" {
			// Extract key sections from .festival/
			if section := extractSection(content, "## Core Principles", "## "); section != "" {
				fmt.Print("\nFestival Methodology - Core Principles\n")
				fmt.Println("======================================")
				fmt.Println(section)

				// Also check README.md for additional methodology guidance
				readmePath := filepath.Join(dotFestival, "README.md")
				if rdContent := extractSection(readmePath, "Requirements-Driven Implementation", "### Standard 3-Phase"); rdContent != "" {
					fmt.Println("\nRequirements-Driven Implementation")
					fmt.Println("-----------------------------------")
					fmt.Println(rdContent)
				}

				if paContent := extractSection(readmePath, "Phase Adaptability", "## Goal Files"); paContent != "" {
					fmt.Println("\nPhase Adaptability")
					fmt.Println("------------------")
					fmt.Println(paContent)
				}

				fmt.Printf("\nSource: %s\n", dotFestival)
				return
			}
		}
	}

	// Default: use embedded content
	fmt.Print("\n")
	fmt.Print(understanddocs.Load("methodology.txt"))
}
