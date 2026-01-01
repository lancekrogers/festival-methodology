// Package status provides atomic status change operations for festivals.
package status

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/id"
	"github.com/lancekrogers/festival-methodology/fest/internal/registry"
)

// AtomicStatusChange performs an atomic status change for a festival.
// For "completed" status, it uses date-based directories.
// Returns the new path of the festival.
func AtomicStatusChange(festivalPath, fromStatus, toStatus string) (string, error) {
	festivalName := filepath.Base(festivalPath)
	festivalsRoot := filepath.Dir(filepath.Dir(festivalPath))

	// Record status history before move
	if err := RecordStatusChange(festivalPath, fromStatus, toStatus, ""); err != nil {
		// Log warning but don't fail - history is optional
		_ = err
	}

	var newPath string
	if toStatus == "completed" {
		// Use date-based directory for completed festivals
		dateDir := CalculateCompletionDateDir(time.Now())
		completedDir := filepath.Join(festivalsRoot, "completed")
		var err error
		newPath, err = MoveToDateDirectory(festivalPath, completedDir, dateDir)
		if err != nil {
			return "", err
		}
	} else {
		newPath = filepath.Join(festivalsRoot, toStatus, festivalName)

		// Check if destination exists
		if _, err := os.Stat(newPath); err == nil {
			return "", os.ErrExist
		}

		// Create parent and move
		if err := os.MkdirAll(filepath.Dir(newPath), 0755); err != nil {
			return "", err
		}
		if err := os.Rename(festivalPath, newPath); err != nil {
			return "", err
		}
	}

	// Update registry with new path/status
	updateRegistry(festivalsRoot, festivalName, toStatus, newPath)

	return newPath, nil
}

// updateRegistry updates the ID registry with the new festival status and path.
// This is best-effort - registry updates are non-blocking.
func updateRegistry(festivalsRoot, festivalName, newStatus, newPath string) {
	ctx := context.Background()
	regPath := registry.GetRegistryPath(festivalsRoot)
	reg, err := registry.Load(ctx, regPath)
	if err != nil {
		return
	}

	// Try to extract ID from directory name
	festivalID, err := id.ExtractIDFromDirName(festivalName)
	if err != nil {
		return
	}

	// Update existing entry or add new one
	entry, err := reg.Get(ctx, festivalID)
	if err != nil {
		// Entry doesn't exist - this shouldn't happen for proper festivals
		return
	}

	entry.Status = newStatus
	entry.Path = newPath
	entry.UpdatedAt = time.Now()

	if err := reg.Update(ctx, entry); err != nil {
		return
	}

	_ = reg.Save(ctx)
}

// CopyDeleteMove performs a copy+delete operation for cross-filesystem moves.
// This is used as a fallback when os.Rename fails across filesystems.
func CopyDeleteMove(sourcePath, destDir, festivalName string) (string, error) {
	destPath := filepath.Join(destDir, festivalName)

	// Create destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return "", err
	}

	// Copy all files recursively
	err := filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(destPath, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		// Copy file
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(targetPath, content, info.Mode())
	})
	if err != nil {
		// Cleanup on failure
		os.RemoveAll(destPath)
		return "", err
	}

	// Delete source
	if err := os.RemoveAll(sourcePath); err != nil {
		return "", err
	}

	return destPath, nil
}
