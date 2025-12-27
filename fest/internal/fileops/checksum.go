package fileops

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ChecksumEntry represents a file checksum entry
type ChecksumEntry struct {
	Hash     string    `json:"hash"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	Original bool      `json:"original,omitempty"`
}

// ChecksumData represents the complete checksum file data
type ChecksumData struct {
	Version string                   `json:"version"`
	Created time.Time                `json:"created"`
	Updated time.Time                `json:"updated"`
	Source  ChecksumSource           `json:"source,omitempty"`
	Files   map[string]ChecksumEntry `json:"files"`
}

// ChecksumSource represents the source of the checksums
type ChecksumSource struct {
	Repository string `json:"repository,omitempty"`
	Branch     string `json:"branch,omitempty"`
	Commit     string `json:"commit,omitempty"`
}

// GenerateChecksums generates checksums for all files in a directory
func GenerateChecksums(ctx context.Context, rootPath string) (map[string]ChecksumEntry, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	checksums := make(map[string]ChecksumEntry)

	err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Skip hidden files and special files
		if shouldSkipFile(path, rootPath) {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return err
		}

		// Calculate checksum
		hash, err := calculateFileChecksum(ctx, path)
		if err != nil {
			return errors.IO("calculating checksum", err).WithField("path", relPath)
		}

		checksums[relPath] = ChecksumEntry{
			Hash:     hash,
			Size:     info.Size(),
			Modified: info.ModTime(),
			Original: true,
		}

		return nil
	})

	return checksums, err
}

// LoadChecksums loads checksums from a JSON file
func LoadChecksums(ctx context.Context, path string) (map[string]ChecksumEntry, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.IO("reading checksum file", err).WithField("path", path)
	}

	var checksumData ChecksumData
	if err := json.Unmarshal(data, &checksumData); err != nil {
		// Try to unmarshal as simple map for backward compatibility
		var simpleChecksums map[string]ChecksumEntry
		if err2 := json.Unmarshal(data, &simpleChecksums); err2 == nil {
			return simpleChecksums, nil
		}
		return nil, errors.Parse("parsing checksum file", err).WithField("path", path)
	}

	return checksumData.Files, nil
}

// SaveChecksums saves checksums to a JSON file
func SaveChecksums(ctx context.Context, path string, checksums map[string]ChecksumEntry) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return err
	}

	checksumData := ChecksumData{
		Version: "1.0.0",
		Created: time.Now(),
		Updated: time.Now(),
		Files:   checksums,
	}

	// Check if file exists to preserve creation time
	if existingData, err := os.ReadFile(path); err == nil {
		var existing ChecksumData
		if json.Unmarshal(existingData, &existing) == nil {
			checksumData.Created = existing.Created
		}
	}

	data, err := json.MarshalIndent(checksumData, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling checksums")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.IO("writing checksum file", err).WithField("path", path)
	}

	return nil
}

// CompareChecksums compares two checksum maps and returns differences
func CompareChecksums(stored, current map[string]ChecksumEntry) (unchanged, modified, added, deleted []string) {
	// Check existing files
	for path, currentEntry := range current {
		if storedEntry, exists := stored[path]; exists {
			if currentEntry.Hash == storedEntry.Hash {
				unchanged = append(unchanged, path)
			} else {
				modified = append(modified, path)
			}
		} else {
			added = append(added, path)
		}
	}

	// Check deleted files
	for path := range stored {
		if _, exists := current[path]; !exists {
			deleted = append(deleted, path)
		}
	}

	return
}

// calculateFileChecksum calculates SHA256 checksum of a file
func calculateFileChecksum(ctx context.Context, path string) (string, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return "", err
	}

	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// shouldSkipFile determines if a file should be skipped
func shouldSkipFile(path, rootPath string) bool {
	relPath, err := filepath.Rel(rootPath, path)
	if err != nil {
		return true
	}

	// Skip checksum file itself
	if strings.HasSuffix(relPath, ".fest-checksums.json") {
		return true
	}

	// Skip backup directories
	if strings.Contains(relPath, ".fest-backup") {
		return true
	}

	// Skip git files
	if strings.Contains(relPath, ".git") {
		return true
	}

	// Skip lock files
	if strings.HasSuffix(relPath, ".fest.lock") {
		return true
	}

	// Skip temporary files
	if strings.HasSuffix(relPath, ".tmp") || strings.HasSuffix(relPath, "~") {
		return true
	}

	return false
}

// VerifyChecksum verifies a file against its stored checksum
func VerifyChecksum(ctx context.Context, path string, expected string) (bool, error) {
	actual, err := calculateFileChecksum(ctx, path)
	if err != nil {
		return false, err
	}
	return actual == expected, nil
}
