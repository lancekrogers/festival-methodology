package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
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
				return fmt.Errorf("could not get working directory: %w", err)
			}
			festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
			if err != nil {
				return fmt.Errorf("could not find festivals root: %w", err)
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

// migrateSingleFestival migrates a single festival to use the ID system
func migrateSingleFestival(ctx context.Context, festivalsRoot, festivalPath string, dryRun, verbose bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	// Check if festival already has an ID
	dirName := filepath.Base(festivalPath)
	existingID, err := id.ExtractIDFromDirName(dirName)
	if err == nil && existingID != "" {
		fmt.Printf("Festival %s already has ID %s\n", dirName, existingID)
		return nil
	}

	// Load existing config
	festConfig, err := config.LoadFestivalConfig(festivalPath)
	if err != nil {
		return fmt.Errorf("failed to load festival config: %w", err)
	}

	// Check if metadata already exists
	if festConfig.Metadata.ID != "" {
		fmt.Printf("Festival %s already has metadata ID %s\n", dirName, festConfig.Metadata.ID)
		return nil
	}

	// Generate new ID
	festivalName := dirName
	if festConfig.Metadata.Name != "" {
		festivalName = festConfig.Metadata.Name
	}
	newID, err := id.GenerateID(festivalName, festivalsRoot)
	if err != nil {
		return fmt.Errorf("failed to generate ID: %w", err)
	}

	// Calculate new directory name
	newDirName := fmt.Sprintf("%s_%s", dirName, newID)
	newPath := filepath.Join(filepath.Dir(festivalPath), newDirName)

	if dryRun {
		fmt.Printf("[DRY-RUN] Would migrate:\n")
		fmt.Printf("  Festival: %s\n", dirName)
		fmt.Printf("  New ID: %s\n", newID)
		fmt.Printf("  New path: %s\n", newPath)
		return nil
	}

	// Update metadata in config
	festConfig.Metadata = config.FestivalMetadata{
		ID:        newID,
		UUID:      uuid.New().String(),
		Name:      festivalName,
		CreatedAt: time.Now(),
		StatusHistory: []config.StatusChange{
			{
				Status:    detectStatusFromPath(festivalPath),
				Timestamp: time.Now(),
				Path:      newPath,
				Notes:     "Migrated to ID system",
			},
		},
	}

	// Save updated config
	if err := config.SaveFestivalConfig(festivalPath, festConfig); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	// Rename directory
	if err := os.Rename(festivalPath, newPath); err != nil {
		return fmt.Errorf("failed to rename directory: %w", err)
	}

	// Update registry
	regPath := registry.GetRegistryPath(festivalsRoot)
	reg, err := registry.Load(ctx, regPath)
	if err != nil {
		fmt.Printf("Warning: could not load registry: %v\n", err)
	} else {
		entry := registry.RegistryEntry{
			ID:        newID,
			Name:      festivalName,
			Status:    detectStatusFromPath(newPath),
			Path:      newPath,
			CreatedAt: festConfig.Metadata.CreatedAt,
			UpdatedAt: time.Now(),
		}
		if err := reg.Add(ctx, entry); err != nil {
			fmt.Printf("Warning: could not add to registry: %v\n", err)
		} else {
			if err := reg.Save(ctx); err != nil {
				fmt.Printf("Warning: could not save registry: %v\n", err)
			}
		}
	}

	if verbose {
		fmt.Printf("Migrated %s -> %s (ID: %s)\n", dirName, newDirName, newID)
	} else {
		fmt.Printf("Migrated: %s\n", newDirName)
	}

	return nil
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
