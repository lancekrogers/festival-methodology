package template

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// FindWorkspaceRoot walks up from startDir until it finds a directory containing .festival/
// Note: This legacy helper returns the directory that directly contains a .festival/ folder.
// Some commands historically assumed this to be the project root. Newer commands should
// prefer FindFestivalsRoot which anchors on festivals/.festival.
func FindWorkspaceRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if dir == "" || dir == "/" || dir == "." {
			break
		}
		if info, err := os.Stat(filepath.Join(dir, ".festival")); err == nil && info.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.NotFound("festival workspace").
		WithField("start_dir", startDir)
}

// FindFestivalsRoot walks up from startDir until it finds the festivals directory
// that contains a .festival/ subdirectory. It only matches if you are inside the
// festivals/ tree (or at the festivals directory itself). It will not match when
// festivals/ exists only as a child of your current directory (enforces being inside).
func FindFestivalsRoot(startDir string) (string, error) {
	dir := startDir
	for {
		if dir == "" || dir == "/" || dir == "." {
			break
		}
		if filepath.Base(dir) == "festivals" {
			if info, err := os.Stat(filepath.Join(dir, ".festival")); err == nil && info.IsDir() {
				return dir, nil
			}
			// Found a festivals dir but missing .festival metadata
			return "", errors.Validation("detected festivals/ but missing .festival directory; run 'fest init'").
				WithField("dir", dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", errors.NotFound("festivals directory").
		WithField("start_dir", startDir)
}

// LocalTemplateRoot returns <festivals_root>/.festival/templates
func LocalTemplateRoot(startDir string) (string, error) {
	root, err := FindFestivalsRoot(startDir)
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".festival", "templates"), nil
}
