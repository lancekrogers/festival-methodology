package understand

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func newUnderstandResourcesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "resources",
		Short: "What's in the .festival/ directory",
		Long:  `List the templates, agents, and examples available in your .festival/ directory.`,
		Run: func(cmd *cobra.Command, args []string) {
			printResources(findDotFestivalDir())
		},
	}
}

func printResources(dotFestival string) {
	if dotFestival == "" {
		fmt.Println("\nNo .festival/ directory found.")
		fmt.Println("Run from a festivals/ tree to see your methodology resources.")
		fmt.Println("\nExpected location: festivals/.festival/")
		return
	}

	fmt.Printf("\nFestival Resources: %s\n", dotFestival)
	fmt.Println(strings.Repeat("=", 50))

	// List actual directory structure
	fmt.Println("\nDirectory Structure:")
	printDirectoryTree(dotFestival, "", 0)

	// Show templates with descriptions
	fmt.Println("\nTemplates (read when creating that document type):")
	fmt.Println("-" + strings.Repeat("-", 49))
	templatesDir := filepath.Join(dotFestival, "templates")
	if entries, err := os.ReadDir(templatesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				name := entry.Name()
				desc := getTemplateDescription(filepath.Join(templatesDir, name))
				if desc != "" {
					fmt.Printf("  %-35s %s\n", name, desc)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}
		}
	}

	// Show agents
	fmt.Println("\nAgents (read when using that agent):")
	fmt.Println("-" + strings.Repeat("-", 49))
	agentsDir := filepath.Join(dotFestival, "agents")
	if entries, err := os.ReadDir(agentsDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") && entry.Name() != "INDEX.md" {
				name := entry.Name()
				desc := getTemplateDescription(filepath.Join(agentsDir, name))
				if desc != "" {
					fmt.Printf("  %-35s %s\n", name, desc)
				} else {
					fmt.Printf("  %s\n", name)
				}
			}
		}
	}

	// Show examples
	fmt.Println("\nExamples (read when stuck or need patterns):")
	fmt.Println("-" + strings.Repeat("-", 49))
	examplesDir := filepath.Join(dotFestival, "examples")
	if entries, err := os.ReadDir(examplesDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".md") {
				fmt.Printf("  %s\n", entry.Name())
			}
		}
	}
}

func printDirectoryTree(dir string, prefix string, depth int) {
	if depth > 2 {
		return
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Filter to show only directories and key files
	var items []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() || strings.HasSuffix(entry.Name(), ".md") {
			items = append(items, entry)
		}
	}

	for i, entry := range items {
		isLast := i == len(items)-1
		connector := "├── "
		if isLast {
			connector = "└── "
		}

		if entry.IsDir() {
			fmt.Printf("%s%s%s/\n", prefix, connector, entry.Name())
			newPrefix := prefix + "│   "
			if isLast {
				newPrefix = prefix + "    "
			}
			printDirectoryTree(filepath.Join(dir, entry.Name()), newPrefix, depth+1)
		} else {
			fmt.Printf("%s%s%s\n", prefix, connector, entry.Name())
		}
	}
}

func getTemplateDescription(path string) string {
	file, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false
	for scanner.Scan() {
		line := scanner.Text()
		if line == "---" {
			if inFrontmatter {
				break
			}
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.HasPrefix(line, "description:") {
			return strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		}
	}
	return ""
}
