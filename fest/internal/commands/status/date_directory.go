// Package status provides date-based directory management for festival completion.
package status

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// CalculateCompletionDateDir returns the YYYY-MM formatted date directory
// for organizing completed festivals by month.
func CalculateCompletionDateDir(t time.Time) string {
	return t.Format("2006-01")
}

// CreateDateDirectory creates a date-organized directory under the completed folder.
// It's idempotent - calling multiple times with the same arguments is safe.
func CreateDateDirectory(completedDir, dateDir string) error {
	path := filepath.Join(completedDir, dateDir)
	if err := os.MkdirAll(path, 0755); err != nil {
		return errors.IO("creating date directory", err).WithField("path", path)
	}
	return nil
}

// MoveToDateDirectory moves a festival to a date-organized completed directory.
// It returns the new path on success.
func MoveToDateDirectory(sourcePath, completedDir, dateDir string) (string, error) {
	festivalName := filepath.Base(sourcePath)
	destPath := filepath.Join(completedDir, dateDir, festivalName)

	// Check if destination already exists
	if _, err := os.Stat(destPath); err == nil {
		return "", errors.Validation("destination already exists").
			WithField("destination", destPath)
	}

	// Create parent directory
	if err := os.MkdirAll(filepath.Join(completedDir, dateDir), 0755); err != nil {
		return "", errors.IO("creating date directory", err)
	}

	// Attempt atomic rename first
	if err := os.Rename(sourcePath, destPath); err != nil {
		// Fallback to copy+delete for cross-filesystem moves
		return copyAndDelete(sourcePath, destPath)
	}

	return destPath, nil
}

// GetCompletedPath returns the full path for a completed festival with date organization.
func GetCompletedPath(festivalsRoot, festivalName, dateDir string) string {
	return filepath.Join(festivalsRoot, "completed", dateDir, festivalName)
}

// copyAndDelete performs a copy+delete operation for cross-filesystem moves.
// It copies all files recursively, then deletes the source.
// If copy fails, no deletion occurs (atomic behavior).
func copyAndDelete(sourcePath, destPath string) (string, error) {
	// Create destination directory
	if err := os.MkdirAll(destPath, 0755); err != nil {
		return "", errors.IO("creating destination directory", err)
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
		return copyFile(path, targetPath, info.Mode())
	})
	if err != nil {
		// Cleanup on failure - remove partial copy
		os.RemoveAll(destPath)
		return "", errors.Wrap(err, "copying festival files")
	}

	// Delete source only after successful copy
	if err := os.RemoveAll(sourcePath); err != nil {
		return "", errors.IO("removing source after copy", err)
	}

	return destPath, nil
}

// copyFile copies a single file preserving permissions.
func copyFile(src, dst string, mode os.FileMode) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
