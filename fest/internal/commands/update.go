package commands

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/fileops"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type updateOptions struct {
	dryRun        bool
	force         bool
	backup        bool
	interactive   bool
	noInteractive bool
	diff          bool
}

// NewUpdateCommand creates the update command
func NewUpdateCommand() *cobra.Command {
	opts := &updateOptions{}

	cmd := &cobra.Command{
		Use:   "update [path]",
		Short: "Update festival files from latest templates",
		Long: `Update festival files from latest templates, preserving user modifications.
  
This command compares your festival files with the latest templates and updates
only the files you haven't modified. For modified files, it will prompt you
for action unless --no-interactive is specified.`,
		Example: `  fest update                  # Interactive update
  fest update --dry-run        # Preview changes
  fest update --no-interactive # Skip all modified files
  fest update --backup         # Create backups before updating`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := "."
			if len(args) > 0 {
				targetPath = args[0]
			}
			return runUpdate(targetPath, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "show what would be updated without making changes")
	cmd.Flags().BoolVar(&opts.force, "force", false, "update all files regardless of modifications")
	cmd.Flags().BoolVar(&opts.backup, "backup", false, "create backups before updating")
	cmd.Flags().BoolVar(&opts.interactive, "interactive", true, "prompt for each modified file")
	cmd.Flags().BoolVar(&opts.noInteractive, "no-interactive", false, "update only unchanged files, skip modified")
	cmd.Flags().BoolVar(&opts.diff, "diff", false, "show diffs for modified files")

	return cmd
}

func runUpdate(targetPath string, opts *updateOptions) error {
	// Create UI handler
	display := ui.New(noColor, verbose)

	// If no-interactive is set, disable interactive
	if opts.noInteractive {
		opts.interactive = false
	}

	// Resolve festivals root from targetPath (works from any subdirectory under festivals/)
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}
	festivalsRoot, err := tpl.FindFestivalsRoot(absTarget)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	festivalDir := filepath.Join(festivalsRoot, ".festival")

	// Load checksums (stored in .festival/ directory)
	checksumFile := filepath.Join(festivalDir, ".fest-checksums.json")
	if !fileops.Exists(checksumFile) {
		return fmt.Errorf("no checksum file found. Run 'fest init' first")
	}

	storedChecksums, err := fileops.LoadChecksums(checksumFile)
	if err != nil {
		return fmt.Errorf("failed to load checksums: %w", err)
	}

	// Get source directory (.festival only)
	sourceDir := filepath.Join(config.ConfigDir(), "festivals", ".festival")
	if !fileops.Exists(sourceDir) {
		return fmt.Errorf("no cached .festival templates found. Run 'fest sync' first")
	}

	display.Info("Analyzing .festival files...")

	// Calculate current checksums for the .festival directory only
	currentChecksums, err := fileops.GenerateChecksums(festivalDir)
	if err != nil {
		return fmt.Errorf("failed to generate checksums: %w", err)
	}

	// Categorize files
	changes := categorizeChanges(storedChecksums, currentChecksums)

	// Show summary
	display.Info("\nFile status:")
	display.Info("  Unchanged: %d files (safe to update)", len(changes.unchanged))
	display.Info("  Modified:  %d files (need decision)", len(changes.modified))
	display.Info("  New:       %d files (user created, will skip)", len(changes.new))
	display.Info("  Deleted:   %d files (user removed, will skip)", len(changes.deleted))

	if opts.dryRun {
		display.Warning("\nDRY RUN - No files will be modified")
		displayChanges(display, changes)
		return nil
	}

	// Create backup if requested (backup only .festival directory)
	if opts.backup {
		backupDir := filepath.Join(festivalDir, ".fest-backup", timeStamp())
		display.Info("\nCreating backup at %s...", backupDir)
		if err := fileops.CreateBackup(festivalDir, backupDir); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
	}

	// Process updates (update only .festival files)
	updater := fileops.NewUpdater(sourceDir, festivalDir)
	updatedFiles := []string{}
	skippedFiles := []string{}

	// Update unchanged files
	for _, file := range changes.unchanged {
		if verbose {
			display.Info("Updating %s...", file)
		}
		if err := updater.UpdateFile(file); err != nil {
			display.Warning("Failed to update %s: %v", file, err)
		} else {
			updatedFiles = append(updatedFiles, file)
		}
	}

	// Handle modified files
	acceptAll := false
	for _, file := range changes.modified {
		if opts.force || acceptAll {
			// Force update or accept all
			if err := updater.UpdateFile(file); err != nil {
				display.Warning("Failed to update %s: %v", file, err)
			} else {
				updatedFiles = append(updatedFiles, file)
			}
		} else if opts.interactive {
			// Interactive prompt
			action := promptForFile(display, file)
			switch action {
			case "yes":
				if err := updater.UpdateFile(file); err != nil {
					display.Warning("Failed to update %s: %v", file, err)
				} else {
					updatedFiles = append(updatedFiles, file)
				}
			case "skip":
				skippedFiles = append(skippedFiles, file)
			case "all":
				acceptAll = true
				if err := updater.UpdateFile(file); err != nil {
					display.Warning("Failed to update %s: %v", file, err)
				} else {
					updatedFiles = append(updatedFiles, file)
				}
			default:
				skippedFiles = append(skippedFiles, file)
			}
		} else {
			// Non-interactive - skip modified
			skippedFiles = append(skippedFiles, file)
		}
	}

	// Update checksums for updated files
	if len(updatedFiles) > 0 {
		display.Info("\nUpdating .festival checksums...")
		newChecksums, err := fileops.GenerateChecksums(festivalDir)
		if err != nil {
			display.Warning("Failed to update checksums: %v", err)
		} else {
			if err := fileops.SaveChecksums(checksumFile, newChecksums); err != nil {
				display.Warning("Failed to save checksums: %v", err)
			}
		}
	}

	// Show summary
	display.Success("\nUpdate complete:")
	display.Info("  Updated: %d files", len(updatedFiles))
	display.Info("  Skipped: %d files", len(skippedFiles))

	return nil
}

type fileChanges struct {
	unchanged []string
	modified  []string
	new       []string
	deleted   []string
}

func categorizeChanges(stored, current map[string]fileops.ChecksumEntry) fileChanges {
	changes := fileChanges{
		unchanged: []string{},
		modified:  []string{},
		new:       []string{},
		deleted:   []string{},
	}

	// Check existing files
	for path, currentEntry := range current {
		if storedEntry, exists := stored[path]; exists {
			if currentEntry.Hash == storedEntry.Hash {
				changes.unchanged = append(changes.unchanged, path)
			} else {
				changes.modified = append(changes.modified, path)
			}
		} else {
			changes.new = append(changes.new, path)
		}
	}

	// Check deleted files
	for path := range stored {
		if _, exists := current[path]; !exists {
			changes.deleted = append(changes.deleted, path)
		}
	}

	return changes
}

func promptForFile(display *ui.UI, file string) string {
	fmt.Printf("\nFile modified: %s\n", file)
	fmt.Print("Update this file? [Y/s/a] (Y=yes, s=skip, a=accept all): ")

	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "", "y", "yes":
		return "yes"
	case "s", "skip":
		return "skip"
	case "a", "all", "accept":
		return "all"
	default:
		return "skip"
	}
}

func displayChanges(display *ui.UI, changes fileChanges) {
	if len(changes.unchanged) > 0 {
		display.Info("\nFiles to update:")
		for _, file := range changes.unchanged {
			display.Info("  ✓ %s", file)
		}
	}

	if len(changes.modified) > 0 {
		display.Info("\nModified files (would skip):")
		for _, file := range changes.modified {
			display.Info("  ⚠ %s", file)
		}
	}
}

func timeStamp() string {
	return time.Now().Format("2006-01-02_150405")
}
