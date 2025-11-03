package fileops

import (
	"fmt"
	"os"
	"path/filepath"

	gitignore "github.com/sabhiram/go-gitignore"
)

// WalkResult contains information about walked files
type WalkResult struct {
	Files         []string // List of text files to process
	TotalFiles    int      // Total files found
	SkippedBinary int      // Files skipped (binary)
	SkippedIgnore int      // Files skipped (gitignore)
}

// WalkDirectory recursively walks a directory, respecting .gitignore files
// and filtering out binary files
func WalkDirectory(rootPath string) (*WalkResult, error) {
	result := &WalkResult{
		Files: []string{},
	}

	// Load gitignore rules from the root directory
	gitignoreFile := filepath.Join(rootPath, ".gitignore")
	var gi *gitignore.GitIgnore
	if _, err := os.Stat(gitignoreFile); err == nil {
		gi, err = gitignore.CompileIgnoreFile(gitignoreFile)
		if err != nil {
			return nil, fmt.Errorf("failed to parse .gitignore: %w", err)
		}
	}

	// Walk the directory tree
	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip .git directory
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		result.TotalFiles++

		// Get relative path for gitignore matching
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Check if file is in .gitignore
		if gi != nil && gi.MatchesPath(relPath) {
			result.SkippedIgnore++
			return nil
		}

		// Check if file is binary
		isBinary, err := IsBinaryFile(path)
		if err != nil {
			// If we can't determine, skip it to be safe
			result.SkippedBinary++
			return nil
		}
		if isBinary {
			result.SkippedBinary++
			return nil
		}

		// Add to result
		result.Files = append(result.Files, path)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return result, nil
}

// AggregateFileContents reads all files and returns combined content
func AggregateFileContents(files []string) ([]byte, error) {
	var totalContent []byte

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %w", file, err)
		}
		totalContent = append(totalContent, content...)
	}

	return totalContent, nil
}
