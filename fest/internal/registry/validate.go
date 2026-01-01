package registry

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/id"
)

// ValidationError represents a discrepancy between registry and filesystem.
type ValidationError struct {
	ID          string
	Type        string // "missing_in_fs", "missing_in_registry", "status_mismatch", "path_mismatch"
	Description string
	Expected    string
	Actual      string
}

// Validate checks registry entries against the filesystem.
// Returns a list of discrepancies found.
func (r *Registry) Validate(ctx context.Context, festivalsRoot string) ([]ValidationError, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	var validationErrors []ValidationError

	r.mu.RLock()
	entries := make(map[string]RegistryEntry, len(r.Entries))
	for k, v := range r.Entries {
		entries[k] = v
	}
	r.mu.RUnlock()

	// Check each registry entry exists on filesystem
	for id, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, errors.Wrap(err, "context cancelled")
		}

		if _, err := os.Stat(entry.Path); os.IsNotExist(err) {
			validationErrors = append(validationErrors, ValidationError{
				ID:          id,
				Type:        "missing_in_fs",
				Description: "Festival exists in registry but not on filesystem",
				Expected:    entry.Path,
				Actual:      "not found",
			})
		}
	}

	// Scan filesystem for festivals not in registry
	fsEntries, err := scanFilesystemForFestivals(ctx, festivalsRoot)
	if err != nil {
		return validationErrors, err
	}

	for fsID, fsPath := range fsEntries {
		if _, exists := entries[fsID]; !exists {
			validationErrors = append(validationErrors, ValidationError{
				ID:          fsID,
				Type:        "missing_in_registry",
				Description: "Festival exists on filesystem but not in registry",
				Expected:    "in registry",
				Actual:      fsPath,
			})
		}
	}

	return validationErrors, nil
}

// Rebuild regenerates the registry from filesystem scan.
// This is useful for recovering from corruption or initializing the registry.
func Rebuild(ctx context.Context, festivalsRoot string, registryPath string) (*Registry, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	reg := New(registryPath)

	// Scan all status directories
	for _, status := range id.StatusDirectories {
		statusPath := filepath.Join(festivalsRoot, status)

		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		// For completed, we need to scan date subdirectories
		if status == "completed" {
			if err := scanCompletedDirectory(ctx, reg, statusPath, status); err != nil {
				return nil, err
			}
			continue
		}

		// Scan regular status directory
		entries, err := os.ReadDir(statusPath)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return nil, errors.Wrap(err, "context cancelled")
			}

			if !entry.IsDir() {
				continue
			}

			if err := addFestivalToRegistry(ctx, reg, filepath.Join(statusPath, entry.Name()), status); err != nil {
				// Log but continue scanning
				continue
			}
		}
	}

	return reg, nil
}

// scanCompletedDirectory handles the date-based subdirectory structure in completed/
func scanCompletedDirectory(ctx context.Context, reg *Registry, completedPath string, status string) error {
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
				return errors.Wrap(err, "context cancelled")
			}

			if !festival.IsDir() {
				continue
			}

			if err := addFestivalToRegistry(ctx, reg, filepath.Join(datePath, festival.Name()), status); err != nil {
				continue
			}
		}
	}

	return nil
}

// addFestivalToRegistry adds a festival from filesystem to the registry
func addFestivalToRegistry(ctx context.Context, reg *Registry, festivalPath string, status string) error {
	dirName := filepath.Base(festivalPath)

	// Try to extract ID from directory name
	festivalID, err := id.ExtractIDFromDirName(dirName)
	if err != nil {
		// Festival doesn't have an ID yet (legacy)
		return nil
	}

	// Try to load fest.yaml for additional metadata
	name := dirName
	festConfig, configErr := config.LoadFestivalConfig(festivalPath)
	if configErr == nil && festConfig.Metadata.Name != "" {
		name = festConfig.Metadata.Name
	}

	entry := RegistryEntry{
		ID:     festivalID,
		Name:   name,
		Status: status,
		Path:   festivalPath,
	}

	if configErr == nil && !festConfig.Metadata.CreatedAt.IsZero() {
		entry.CreatedAt = festConfig.Metadata.CreatedAt
	}

	return reg.Add(ctx, entry)
}

// scanFilesystemForFestivals scans all festival directories for festivals with IDs
func scanFilesystemForFestivals(ctx context.Context, festivalsRoot string) (map[string]string, error) {
	result := make(map[string]string)

	for _, status := range id.StatusDirectories {
		statusPath := filepath.Join(festivalsRoot, status)

		if _, err := os.Stat(statusPath); os.IsNotExist(err) {
			continue
		}

		err := filepath.WalkDir(statusPath, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if err := ctx.Err(); err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			festivalID, extractErr := id.ExtractIDFromDirName(d.Name())
			if extractErr != nil {
				return nil
			}

			result[festivalID] = path
			return nil
		})

		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

// Sync validates and repairs the registry against the filesystem.
// It adds missing entries and removes stale entries.
func (r *Registry) Sync(ctx context.Context, festivalsRoot string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	validationErrors, err := r.Validate(ctx, festivalsRoot)
	if err != nil {
		return err
	}

	for _, valErr := range validationErrors {
		switch valErr.Type {
		case "missing_in_fs":
			// Remove stale entry
			if err := r.Delete(ctx, valErr.ID); err != nil {
				return err
			}
		case "missing_in_registry":
			// Add missing entry by rescanning the specific path
			if err := addFestivalToRegistry(ctx, r, valErr.Actual, detectStatus(valErr.Actual)); err != nil {
				continue
			}
		}
	}

	return r.Save(ctx)
}

// detectStatus determines the status from a festival path
func detectStatus(path string) string {
	for _, status := range id.StatusDirectories {
		if filepath.Base(filepath.Dir(path)) == status {
			return status
		}
		// Check for completed/YYYY-MM/festival pattern
		if filepath.Base(filepath.Dir(filepath.Dir(path))) == status {
			return status
		}
	}
	return "unknown"
}
