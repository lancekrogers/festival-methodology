package system

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
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
		Short: "System: Update fest methodology files from templates",
		Long: `Update the .festival/ methodology files from latest templates.

This is a SYSTEM command that updates fest's methodology files (templates,
documentation, agents) in your workspace - NOT your festival content.

It compares your .festival/ files with the latest templates and updates
only the files you haven't modified. For modified files, it will prompt you
for action unless --no-interactive is specified.

Your actual festivals (phases, sequences, tasks) are never modified by this command.`,
		Example: `  fest system update                  # Interactive update
  fest system update --dry-run        # Preview changes
  fest system update --no-interactive # Skip all modified files
  fest system update --backup         # Create backups before updating`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := "."
			if len(args) > 0 {
				targetPath = args[0]
			}
			return runUpdate(cmd.Context(), targetPath, opts)
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

func runUpdate(ctx context.Context, targetPath string, opts *updateOptions) error {
	// Create UI handler
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// If no-interactive is set, disable interactive
	if opts.noInteractive {
		opts.interactive = false
	}

	// Resolve festivals root from targetPath (works from any subdirectory under festivals/)
	absTarget, err := filepath.Abs(targetPath)
	if err != nil {
		return errors.Wrap(err, "resolving path").WithField("path", targetPath)
	}
	festivalsRoot, err := tpl.FindFestivalsRoot(absTarget)
	if err != nil {
		return err
	}
	festivalDir := filepath.Join(festivalsRoot, ".festival")

	// Load checksums (stored in .festival/ directory)
	checksumFile := filepath.Join(festivalDir, ".fest-checksums.json")
	if !fileops.Exists(checksumFile) {
		return errors.NotFound("checksum file").WithField("path", checksumFile).WithField("hint", "run 'fest init' first")
	}

	storedChecksums, err := fileops.LoadChecksums(ctx, checksumFile)
	if err != nil {
		return errors.IO("loading checksums", err).WithField("path", checksumFile)
	}

	// Get source directory (.festival only)
	sourceDir := filepath.Join(config.ConfigDir(), "festivals", ".festival")
	if !fileops.Exists(sourceDir) {
		return errors.NotFound(".festival templates").WithField("path", sourceDir).WithField("hint", "run 'fest system sync' first")
	}

	display.Info("Analyzing .festival methodology files...")

	// Calculate current checksums for the .festival directory only
	currentChecksums, err := fileops.GenerateChecksums(ctx, festivalDir)
	if err != nil {
		return errors.Wrap(err, "generating checksums").WithField("path", festivalDir)
	}

	// Categorize files
	changes := categorizeChanges(ctx, storedChecksums, currentChecksums, sourceDir)

	// Filter unchanged files to only those that exist in source
	// (handles case where files were removed from upstream methodology)
	validUnchanged := []string{}
	for _, file := range changes.unchanged {
		srcPath := filepath.Join(sourceDir, file)
		if fileops.Exists(srcPath) {
			validUnchanged = append(validUnchanged, file)
		} else if shared.IsVerbose() {
			display.Warning("Skipping %s (removed from source)", file)
		}
	}
	changes.unchanged = validUnchanged

	// Show summary
	display.Info("\nMethodology file status:")
	display.Info("  Unchanged:   %d files (safe to update)", len(changes.unchanged))
	display.Info("  Modified:    %d files (need decision)", len(changes.modified))
	display.Info("  New:         %d files (user created, will skip)", len(changes.new))
	display.Info("  Deleted:     %d files (user removed, will skip)", len(changes.deleted))
	display.Info("  From source: %d files (new upstream files)", len(changes.fromSource))

	if opts.dryRun {
		display.Warning("\nDRY RUN - No files will be modified")
		displayChanges(display, changes)
		return nil
	}

	// Create backup if requested (backup only .festival directory)
	if opts.backup {
		backupDir := filepath.Join(festivalDir, ".fest-backup", timeStamp())
		display.Info("\nCreating backup at %s...", backupDir)
		if err := fileops.CreateBackup(ctx, festivalDir, backupDir); err != nil {
			return errors.IO("creating backup", err).WithField("path", backupDir)
		}
	}

	// Process updates (update only .festival files)
	updater := fileops.NewUpdater(sourceDir, festivalDir)
	updatedFiles := []string{}
	skippedFiles := []string{}

	// Update unchanged files
	for _, file := range changes.unchanged {
		if shared.IsVerbose() {
			display.Info("Updating %s...", file)
		}
		if err := updater.UpdateFile(ctx, file); err != nil {
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
			if err := updater.UpdateFile(ctx, file); err != nil {
				display.Warning("Failed to update %s: %v", file, err)
			} else {
				updatedFiles = append(updatedFiles, file)
			}
		} else if opts.interactive {
			// Interactive prompt
			action := promptForFile(display, file)
			switch action {
			case "yes":
				if err := updater.UpdateFile(ctx, file); err != nil {
					display.Warning("Failed to update %s: %v", file, err)
				} else {
					updatedFiles = append(updatedFiles, file)
				}
			case "skip":
				skippedFiles = append(skippedFiles, file)
			case "all":
				acceptAll = true
				if err := updater.UpdateFile(ctx, file); err != nil {
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

	// Copy new files from source (these are new upstream files not in workspace yet)
	for _, file := range changes.fromSource {
		if shared.IsVerbose() {
			display.Info("Adding new upstream file %s...", file)
		}
		if err := updater.UpdateFile(ctx, file); err != nil {
			display.Warning("Failed to add %s: %v", file, err)
		} else {
			updatedFiles = append(updatedFiles, file)
		}
	}

	// Update checksums for updated files
	if len(updatedFiles) > 0 {
		display.Info("\nUpdating .festival checksums...")
		newChecksums, err := fileops.GenerateChecksums(ctx, festivalDir)
		if err != nil {
			display.Warning("Failed to update checksums: %v", err)
		} else {
			if err := fileops.SaveChecksums(ctx, checksumFile, newChecksums); err != nil {
				display.Warning("Failed to save checksums: %v", err)
			}
		}
	}

	// Show summary
	display.Success("\nMethodology update complete:")
	display.Info("  Updated: %d files", len(updatedFiles))
	display.Info("  Skipped: %d files", len(skippedFiles))

	return nil
}

type fileChanges struct {
	unchanged  []string
	modified   []string
	new        []string // User-created files (exist in workspace but not in stored checksums)
	deleted    []string // User-deleted files (exist in stored checksums but not in workspace)
	fromSource []string // New from source (exist in source but not in workspace or stored checksums)
}

func categorizeChanges(ctx context.Context, stored, current map[string]fileops.ChecksumEntry, sourceDir string) fileChanges {
	changes := fileChanges{
		unchanged:  []string{},
		modified:   []string{},
		new:        []string{},
		deleted:    []string{},
		fromSource: []string{},
	}

	// Check existing files in workspace
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

	// Check deleted files (in stored but not in current)
	for path := range stored {
		if _, exists := current[path]; !exists {
			changes.deleted = append(changes.deleted, path)
		}
	}

	// Check for new files in source that don't exist locally or in stored checksums
	sourceFiles, err := fileops.ListFiles(ctx, sourceDir)
	if err == nil {
		for _, file := range sourceFiles {
			_, existsInCurrent := current[file]
			_, existsInStored := stored[file]
			if !existsInCurrent && !existsInStored {
				changes.fromSource = append(changes.fromSource, file)
			}
		}
	}

	return changes
}

func promptForFile(display *ui.UI, file string) string {
	fmt.Println()
	fmt.Println(ui.H2("File Modified"))
	fmt.Printf("%s %s\n", ui.Label("File"), ui.Dim(file))
	fmt.Print(ui.Info("Update this file? [Y/s/a] (Y=yes, s=skip, a=accept all): "))

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
			display.Info("  + %s", file)
		}
	}

	if len(changes.fromSource) > 0 {
		display.Info("\nNew files from upstream to add:")
		for _, file := range changes.fromSource {
			display.Info("  ++ %s", file)
		}
	}

	if len(changes.modified) > 0 {
		display.Info("\nModified files (would skip):")
		for _, file := range changes.modified {
			display.Info("  ~ %s", file)
		}
	}
}

func timeStamp() string {
	return time.Now().Format("2006-01-02_150405")
}
