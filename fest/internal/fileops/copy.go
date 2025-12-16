package fileops

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
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
func (c *Copier) CopyDirectory(src, dst string) error {
	// Get source directory info
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("failed to stat source: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source is not a directory: %s", src)
	}

	// Create destination directory
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Walk source directory
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
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
				return fmt.Errorf("failed to create directory %s: %w", dstPath, err)
			}
			return nil
		}

		// Handle files
		return c.copyFile(path, dstPath, info.Mode())
	})
}

// CopyFile copies a single file
func CopyFile(src, dst string) error {
	return NewCopier().copyFile(src, dst, 0644)
}

// copyFile copies a single file with permissions
func (c *Copier) copyFile(src, dst string, mode os.FileMode) error {
	// Open source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	// Create destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy content
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	// Sync to ensure write
	if err := dstFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync file: %w", err)
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
func CreateBackup(src, backupDir string) error {
	// Create backup directory
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Copy directory to backup
	copier := NewCopier()
	if err := copier.CopyDirectory(src, backupDir); err != nil {
		return fmt.Errorf("failed to copy to backup: %w", err)
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
func (u *Updater) UpdateFile(relPath string) error {
	srcPath := filepath.Join(u.sourceDir, relPath)
	dstPath := filepath.Join(u.targetDir, relPath)

	// Check if source file exists
	if !Exists(srcPath) {
		return fmt.Errorf("source file does not exist: %s", srcPath)
	}

	// Get source file info
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return fmt.Errorf("failed to stat source file: %w", err)
	}

	// Copy file
	return NewCopier().copyFile(srcPath, dstPath, srcInfo.Mode())
}

func timeNow() string {
	return time.Now().Format(time.RFC3339)
}
