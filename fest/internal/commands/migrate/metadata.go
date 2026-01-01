package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	"github.com/lancekrogers/festival-methodology/fest/internal/registry"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"
)

// NewMetadataCommand creates the migrate metadata subcommand
func NewMetadataCommand() *cobra.Command {
	var dryRun bool
	var verbose bool

	cmd := &cobra.Command{
		Use:   "metadata [path]",
		Short: "Add metadata to existing festivals",
		Long: `Migrate existing festivals to use the ID system.

This command:
1. Generates a unique ID for festivals without one
2. Adds metadata to fest.yaml (ID, UUID, creation time)
3. Renames the festival directory to include the ID suffix

The migration is idempotent - running it multiple times is safe.

Examples:
  fest migrate metadata                    # Migrate all festivals
  fest migrate metadata ./active/my-fest   # Migrate specific festival
  fest migrate metadata --dry-run          # Preview changes only`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}

			// Find festivals root
			cwd, err := os.Getwd()
			if err != nil {
				return errors.IO("getting working directory", err)
			}
			festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
			if err != nil {
				return errors.IO("finding festivals root", err)
			}

			// If a specific path is provided, migrate only that festival
			if len(args) > 0 {
				targetPath := args[0]
				if !filepath.IsAbs(targetPath) {
					cwd, _ := os.Getwd()
					targetPath = filepath.Join(cwd, targetPath)
				}
				return migrateSingleFestival(ctx, festivalsRoot, targetPath, dryRun, verbose)
			}

			// Migrate all festivals
			return migrateAllFestivals(ctx, festivalsRoot, dryRun, verbose)
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "preview changes without making them")
	cmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "show detailed progress")

	return cmd
}

// migrationContext holds context for a single festival migration
type migrationContext struct {
	festivalsRoot string
	festivalPath  string
	dirName       string
	newID         string
	newPath       string
	festivalName  string
	festConfig    *config.FestivalConfig
}

// migrateSingleFestival migrates a single festival to use the ID system
func migrateSingleFestival(ctx context.Context, festivalsRoot, festivalPath string, dryRun, verbose bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Check if already migrated
	migCtx, skip, err := checkExistingMigration(festivalPath, festivalsRoot)
	if err != nil {
		return err
	}
	if skip {
		return nil
	}

	// Handle dry-run mode
	if dryRun {
		printDryRunOutput(migCtx)
		return nil
	}

	// Build and apply metadata
	buildMigrationMetadata(migCtx)

	// Perform the actual migration (save config, rename directory)
	if err := performMigration(migCtx); err != nil {
		return err
	}

	// Update registry (non-blocking)
	updateRegistryAfterMigration(ctx, migCtx)

	// Print success message
	printMigrationSuccess(migCtx, verbose)

	return nil
}

// checkExistingMigration checks if a festival is already migrated
// Returns migration context, whether to skip, and any error
func checkExistingMigration(festivalPath, festivalsRoot string) (*migrationContext, bool, error) {
	dirName := filepath.Base(festivalPath)

	// Check if directory already has an ID suffix
	existingID, err := id.ExtractIDFromDirName(dirName)
	if err == nil && existingID != "" {
		fmt.Printf("Festival %s already has ID %s\n", dirName, existingID)
		return nil, true, nil
	}

	// Load existing config
	festConfig, err := config.LoadFestivalConfig(festivalPath)
	if err != nil {
		return nil, false, errors.IO("loading festival config", err).
			WithField("path", festivalPath)
	}

	// Check if metadata already exists
	if festConfig.Metadata.ID != "" {
		fmt.Printf("Festival %s already has metadata ID %s\n", dirName, festConfig.Metadata.ID)
		return nil, true, nil
	}

	// Determine festival name
	festivalName := dirName
	if festConfig.Metadata.Name != "" {
		festivalName = festConfig.Metadata.Name
	}

	// Generate new ID
	newID, err := id.GenerateID(festivalName, festivalsRoot)
	if err != nil {
		return nil, false, errors.Wrap(err, "generating festival ID").
			WithField("name", festivalName)
	}

	// Calculate new path
	newDirName := fmt.Sprintf("%s_%s", dirName, newID)
	newPath := filepath.Join(filepath.Dir(festivalPath), newDirName)

	return &migrationContext{
		festivalsRoot: festivalsRoot,
		festivalPath:  festivalPath,
		dirName:       dirName,
		newID:         newID,
		newPath:       newPath,
		festivalName:  festivalName,
		festConfig:    festConfig,
	}, false, nil
}

// printDryRunOutput prints what would be migrated in dry-run mode
func printDryRunOutput(migCtx *migrationContext) {
	fmt.Printf("[DRY-RUN] Would migrate:\n")
	fmt.Printf("  Festival: %s\n", migCtx.dirName)
	fmt.Printf("  New ID: %s\n", migCtx.newID)
	fmt.Printf("  New path: %s\n", migCtx.newPath)
}

// buildMigrationMetadata populates the festival config with migration metadata
func buildMigrationMetadata(migCtx *migrationContext) {
	now := time.Now()
	migCtx.festConfig.Metadata = config.FestivalMetadata{
		ID:        migCtx.newID,
		UUID:      uuid.New().String(),
		Name:      migCtx.festivalName,
		CreatedAt: now,
		StatusHistory: []config.StatusChange{
			{
				Status:    detectStatusFromPath(migCtx.festivalPath),
				Timestamp: now,
				Path:      migCtx.newPath,
				Notes:     "Migrated to ID system",
			},
		},
	}
}

// performMigration saves the config and renames the directory
func performMigration(migCtx *migrationContext) error {
	// Save updated config
	if err := config.SaveFestivalConfig(migCtx.festivalPath, migCtx.festConfig); err != nil {
		return errors.IO("saving festival config", err).
			WithField("path", migCtx.festivalPath)
	}

	// Rename directory
	if err := os.Rename(migCtx.festivalPath, migCtx.newPath); err != nil {
		return errors.IO("renaming festival directory", err).
			WithField("from", migCtx.festivalPath).
			WithField("to", migCtx.newPath)
	}

	return nil
}

// updateRegistryAfterMigration updates the registry with the migrated festival
// This is non-blocking - errors are printed as warnings
func updateRegistryAfterMigration(ctx context.Context, migCtx *migrationContext) {
	regPath := registry.GetRegistryPath(migCtx.festivalsRoot)
	reg, err := registry.Load(ctx, regPath)
	if err != nil {
		fmt.Printf("Warning: could not load registry: %v\n", err)
		return
	}

	entry := registry.RegistryEntry{
		ID:        migCtx.newID,
		Name:      migCtx.festivalName,
		Status:    detectStatusFromPath(migCtx.newPath),
		Path:      migCtx.newPath,
		CreatedAt: migCtx.festConfig.Metadata.CreatedAt,
		UpdatedAt: time.Now(),
	}

	if err := reg.Add(ctx, entry); err != nil {
		fmt.Printf("Warning: could not add to registry: %v\n", err)
		return
	}

	if err := reg.Save(ctx); err != nil {
		fmt.Printf("Warning: could not save registry: %v\n", err)
	}
}

// printMigrationSuccess prints the success message
func printMigrationSuccess(migCtx *migrationContext, verbose bool) {
	newDirName := filepath.Base(migCtx.newPath)
	if verbose {
		fmt.Printf("Migrated %s -> %s (ID: %s)\n", migCtx.dirName, newDirName, migCtx.newID)
	} else {
		fmt.Printf("Migrated: %s\n", newDirName)
	}
}

// migrateAllFestivals migrates all festivals without IDs
func migrateAllFestivals(ctx context.Context, festivalsRoot string, dryRun, verbose bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var migrated, skipped, failed int

	for _, status := range id.StatusDirectories {
		statusPath := filepath.Join(festivalsRoot, status)

		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		// Handle completed/ with date subdirectories
		if status == "completed" {
			if err := migrateCompletedDirectory(ctx, festivalsRoot, statusPath, dryRun, verbose, &migrated, &skipped, &failed); err != nil {
				return err
			}
			continue
		}

		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}

			if !entry.IsDir() {
				continue
			}

			festivalPath := filepath.Join(statusPath, entry.Name())
			err := migrateSingleFestival(ctx, festivalsRoot, festivalPath, dryRun, verbose)
			if err != nil {
				if verbose {
					fmt.Printf("Error migrating %s: %v\n", entry.Name(), err)
				}
				failed++
			} else {
				// Check if we actually migrated or just skipped
				if _, idErr := id.ExtractIDFromDirName(entry.Name()); idErr == nil {
					skipped++
				} else {
					migrated++
				}
			}
		}
	}

	fmt.Printf("\nMigration complete: %d migrated, %d skipped (already have ID), %d failed\n", migrated, skipped, failed)
	return nil
}

// migrateCompletedDirectory handles migration within date-organized completed/ directory
func migrateCompletedDirectory(ctx context.Context, festivalsRoot, completedPath string, dryRun, verbose bool, migrated, skipped, failed *int) error {
	dateDirs, err := os.ReadDir(completedPath)
	if err != nil {
		return nil
	}

	for _, dateDir := range dateDirs {
		if !dateDir.IsDir() {
			continue
		}

		datePath := filepath.Join(completedPath, dateDir.Name())
		festivals, err := os.ReadDir(datePath)
		if err != nil {
			continue
		}

		for _, festival := range festivals {
			if err := ctx.Err(); err != nil {
				return err
			}

			if !festival.IsDir() {
				continue
			}

			festivalPath := filepath.Join(datePath, festival.Name())
			err := migrateSingleFestival(ctx, festivalsRoot, festivalPath, dryRun, verbose)
			if err != nil {
				if verbose {
					fmt.Printf("Error migrating %s: %v\n", festival.Name(), err)
				}
				*failed++
			} else {
				if _, idErr := id.ExtractIDFromDirName(festival.Name()); idErr == nil {
					*skipped++
				} else {
					*migrated++
				}
			}
		}
	}

	return nil
}

// detectStatusFromPath determines the festival status from its path
func detectStatusFromPath(path string) string {
	for _, status := range id.StatusDirectories {
		// Direct child of status directory
		if filepath.Base(filepath.Dir(path)) == status {
			return status
		}
		// Date subdirectory in completed/
		if filepath.Base(filepath.Dir(filepath.Dir(path))) == status {
			return status
		}
	}
	return "unknown"
}
