package fileops

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Copier handles file copying operations
type Copier struct {
	preservePermissions bool
	skipHidden          bool
}

// NewCopier creates a new file copier
func NewCopier() *Copier {
	return &Copier{
		preservePermissions: true,
		skipHidden:          false,
	}
}

// CopyDirectory recursively copies a directory
func (c *Copier) CopyDirectory(ctx context.Context, src, dst string) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return err
	}

	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return errors.IO("reading source directory", err).WithField("path", src)
	}

	if !srcInfo.IsDir() {
		return errors.Validation("source is not a directory").WithField("path", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return errors.IO("creating destination directory", err).WithField("path", dst)
	}

	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err != nil {
			return err
		}

		// Calculate relative path
		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		// Skip root directory (already created)
		if relPath == "." {
			return nil
		}

		// Calculate destination path
		dstPath := filepath.Join(dst, relPath)

		// Skip hidden files if configured
		if c.skipHidden && isHidden(filepath.Base(path)) && filepath.Base(path) != ".festival" {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Handle directories
		if info.IsDir() {
			if err := os.MkdirAll(dstPath, info.Mode()); err != nil {
				return errors.IO("creating directory", err).WithField("path", dstPath)
			}
			return nil
		}

		// Handle files
		return c.copyFile(ctx, path, dstPath, info.Mode())
	})
}

// CopyFile copies a single file
func CopyFile(ctx context.Context, src, dst string) error {
	return NewCopier().copyFile(ctx, src, dst, 0644)
}

// copyFile copies a single file with permissions
func (c *Copier) copyFile(ctx context.Context, src, dst string, mode os.FileMode) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return err
	}

	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return errors.IO("opening source file", err).WithField("path", src)
	}
	defer srcFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return errors.IO("creating destination directory", err).WithField("path", dstDir)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return errors.IO("creating destination file", err).WithField("path", dst)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return errors.IO("copying file content", err).WithField("src", src).WithField("dst", dst)
	}

	// Sync to ensure write
	if err := dstFile.Sync(); err != nil {
		return errors.IO("syncing file", err).WithField("path", dst)
	}

	// Set permissions if preserving
	if c.preservePermissions && mode != 0 {
		if err := os.Chmod(dst, mode); err != nil {
			// Non-fatal error
			fmt.Fprintf(os.Stderr, "Warning: failed to set permissions on %s: %v\n", dst, err)
		}
	}

	return nil
}

// Exists checks if a file or directory exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// isHidden checks if a file/directory name starts with a dot
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// CreateBackup creates a backup of the specified directory
func CreateBackup(ctx context.Context, src, backupDir string) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return err
	}

	// Create backup directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return errors.IO("creating backup directory", err).WithField("path", backupDir)
	}

	// Copy directory to backup
	copier := NewCopier()
	if err := copier.CopyDirectory(ctx, src, backupDir); err != nil {
		return errors.Wrap(err, "copying to backup").WithField("src", src).WithField("dst", backupDir)
	}

	// Create manifest
	manifest := filepath.Join(backupDir, "manifest.json")
	manifestData := fmt.Sprintf(`{
  "timestamp": "%s",
  "source": "%s",
  "reason": "manual backup"
}`, timeNow(), src)

	if err := os.WriteFile(manifest, []byte(manifestData), 0644); err != nil {
		// Non-fatal error
		fmt.Fprintf(os.Stderr, "Warning: failed to create backup manifest: %v\n", err)
	}

	return nil
}

// Updater handles file updates
type Updater struct {
	sourceDir string
	targetDir string
}

// NewUpdater creates a new file updater
func NewUpdater(sourceDir, targetDir string) *Updater {
	return &Updater{
		sourceDir: sourceDir,
		targetDir: targetDir,
	}
}

// UpdateFile updates a single file from source to target
func (u *Updater) UpdateFile(ctx context.Context, relPath string) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return err
	}

	srcPath := filepath.Join(u.sourceDir, relPath)
	dstPath := filepath.Join(u.targetDir, relPath)

	// Check if source file exists
	if !Exists(srcPath) {
		return errors.NotFound("source file").WithField("path", srcPath)
	}

	// Get source file info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return errors.IO("reading source file", err).WithField("path", srcPath)
	}

	// Copy file
	return NewCopier().copyFile(ctx, srcPath, dstPath, srcInfo.Mode())
}

func timeNow() string {
	return time.Now().Format(time.RFC3339)
}
