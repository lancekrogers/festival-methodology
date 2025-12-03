package template

import (
    "fmt"
    "os"
    "path/filepath"
)

// FindWorkspaceRoot walks up from startDir until it finds a directory containing .festival/
func FindWorkspaceRoot(startDir string) (string, error) {
    dir := startDir
    for {
        if dir == "" || dir == "/" || dir == "." {
            break
        }
        // Check for .festival directory
        if info, err := os.Stat(filepath.Join(dir, ".festival")); err == nil && info.IsDir() {
            return dir, nil
        }
        parent := filepath.Dir(dir)
        if parent == dir {
            break
        }
        dir = parent
    }
    return "", fmt.Errorf("not a festival workspace (missing .festival/)")
}

// LocalTemplateRoot returns <workspace_root>/.festival/templates
func LocalTemplateRoot(startDir string) (string, error) {
    root, err := FindWorkspaceRoot(startDir)
    if err != nil {
        return "", err
    }
    return filepath.Join(root, ".festival", "templates"), nil
}

