package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/frontmatter"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

var (
	dryRun  bool
	fix     bool
	verbose bool
)

// NewFrontmatterCommand creates the frontmatter migration command
func NewFrontmatterCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "frontmatter",
		Short: "Add YAML frontmatter to existing documents",
		Long: `Add YAML frontmatter to festival documents that don't have it.

This command walks through all festival documents and adds frontmatter
to any that are missing it. Existing frontmatter is preserved.

Examples:
  fest migrate frontmatter              # Add frontmatter to all docs
  fest migrate frontmatter --dry-run    # Preview changes without writing
  fest migrate frontmatter --fix        # Update/fix existing frontmatter
  fest migrate frontmatter --verbose    # Show detailed progress`,
		RunE: runFrontmatterMigration,
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without writing")
	cmd.Flags().BoolVar(&fix, "fix", false, "update/fix existing frontmatter")
	cmd.Flags().BoolVar(&verbose, "verbose", false, "show detailed progress")

	return cmd
}

// MigrationResult holds the result of migrating a single file
type MigrationResult struct {
	Path    string
	Action  string // "added", "updated", "skipped", "error"
	Message string
	OldFM   *frontmatter.Frontmatter
	NewFM   *frontmatter.Frontmatter
}

func runFrontmatterMigration(cmd *cobra.Command, args []string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	festivalPath, err := tpl.FindFestivalRoot(cwd)
	if err != nil {
		return fmt.Errorf("not inside a festival: %w", err)
	}

	if dryRun {
		fmt.Println("DRY RUN - No changes will be written")
		fmt.Println()
	}

	results, err := migrateDirectory(festivalPath)
	if err != nil {
		return err
	}

	// Print summary
	printSummary(results)

	return nil
}

func migrateDirectory(dir string) ([]*MigrationResult, error) {
	var results []*MigrationResult

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and files
		if strings.HasPrefix(info.Name(), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip non-markdown files
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".md") {
			return nil
		}

		result := migrateFile(path)
		results = append(results, result)

		if verbose {
			printResult(result)
		}

		return nil
	})

	return results, err
}

func migrateFile(path string) *MigrationResult {
	result := &MigrationResult{
		Path: path,
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to read: %v", err)
		return result
	}

	// Check for existing frontmatter
	existingFM, remaining, err := frontmatter.Parse(content)
	if err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to parse: %v", err)
		return result
	}

	// If frontmatter exists and we're not fixing, skip
	if existingFM != nil && !fix {
		result.Action = "skipped"
		result.Message = "already has frontmatter"
		result.OldFM = existingFM
		return result
	}

	// Infer frontmatter from path
	inferredFM, err := frontmatter.InferFromPath(path)
	if err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to infer: %v", err)
		return result
	}

	// Merge if fixing existing
	var newFM *frontmatter.Frontmatter
	if existingFM != nil && fix {
		newFM = frontmatter.MergeInto(existingFM, inferredFM)
		result.Action = "updated"
		result.Message = "frontmatter updated"
		result.OldFM = existingFM
	} else {
		newFM = inferredFM
		result.Action = "added"
		result.Message = "frontmatter added"
	}
	result.NewFM = newFM

	// Skip if dry run
	if dryRun {
		return result
	}

	// Write the updated file
	newContent, err := frontmatter.Inject(remaining, newFM)
	if err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to inject: %v", err)
		return result
	}

	if err := os.WriteFile(path, newContent, 0o644); err != nil {
		result.Action = "error"
		result.Message = fmt.Sprintf("failed to write: %v", err)
		return result
	}

	return result
}

func printResult(result *MigrationResult) {
	icon := ""
	switch result.Action {
	case "added":
		icon = "✓"
	case "updated":
		icon = "↻"
	case "skipped":
		icon = "○"
	case "error":
		icon = "✗"
	}

	fmt.Printf("%s %s: %s\n", icon, result.Path, result.Message)
}

func printSummary(results []*MigrationResult) {
	added := 0
	updated := 0
	skipped := 0
	errors := 0

	for _, r := range results {
		switch r.Action {
		case "added":
			added++
		case "updated":
			updated++
		case "skipped":
			skipped++
		case "error":
			errors++
		}
	}

	fmt.Println()
	fmt.Println("════════════════════════════════════════")
	fmt.Println("            MIGRATION SUMMARY           ")
	fmt.Println("════════════════════════════════════════")
	fmt.Printf("  Added:   %d\n", added)
	fmt.Printf("  Updated: %d\n", updated)
	fmt.Printf("  Skipped: %d\n", skipped)
	fmt.Printf("  Errors:  %d\n", errors)
	fmt.Printf("  Total:   %d\n", len(results))

	if dryRun && (added > 0 || updated > 0) {
		fmt.Println()
		fmt.Println("Run without --dry-run to apply changes.")
	}
}
