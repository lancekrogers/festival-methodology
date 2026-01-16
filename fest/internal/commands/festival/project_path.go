package festival

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ResolveProjectPath resolves a project path to an absolute path.
// Supports:
//   - Absolute paths: /Users/user/projects
//   - Home expansion: ~/projects
//   - Relative to workspace: projects/my-app (resolved from workspace root)
func ResolveProjectPath(rawPath, workspaceRoot string) (string, error) {
	if rawPath == "" {
		return "", nil
	}

	// Handle home directory expansion
	if strings.HasPrefix(rawPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.IO("getting home directory", err)
		}
		rawPath = filepath.Join(home, rawPath[2:])
	} else if rawPath == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", errors.IO("getting home directory", err)
		}
		rawPath = home
	}

	// If already absolute, use as-is
	if filepath.IsAbs(rawPath) {
		return filepath.Clean(rawPath), nil
	}

	// Relative paths: resolve from workspace root
	if workspaceRoot == "" {
		// If no workspace root, resolve from cwd
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.IO("getting current directory", err)
		}
		return filepath.Join(cwd, rawPath), nil
	}

	return filepath.Join(workspaceRoot, rawPath), nil
}

// ValidateProjectPath checks if the project path is valid.
// Returns nil if path is empty (valid, just not set).
// Returns nil if path exists and is a directory.
// Returns error if path exists but is not a directory.
// Returns a special "not found" error if path doesn't exist (can be warning).
func ValidateProjectPath(path string) error {
	if path == "" {
		return nil
	}

	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return errors.NotFound("project path").WithField("path", path)
	}
	if err != nil {
		return errors.IO("checking project path", err).WithField("path", path)
	}
	if !info.IsDir() {
		return errors.Validation("project path must be a directory").WithField("path", path)
	}

	return nil
}
