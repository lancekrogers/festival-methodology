package fileops

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGenerateChecksums(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()
	
	// Create test files
	files := map[string]string{
		"file1.txt": "content1",
		"file2.txt": "content2",
		"subdir/file3.txt": "content3",
	}
	
	for path, content := range files {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
	
	// Generate checksums
	checksums, err := GenerateChecksums(tmpDir)
	if err != nil {
		t.Fatalf("GenerateChecksums failed: %v", err)
	}
	
	// Verify checksums
	if len(checksums) != 3 {
		t.Errorf("Expected 3 checksums, got %d", len(checksums))
	}
	
	// Check specific file
	if entry, ok := checksums["file1.txt"]; !ok {
		t.Error("file1.txt not found in checksums")
	} else {
		if entry.Size != 8 {
			t.Errorf("Expected size 8, got %d", entry.Size)
		}
		if entry.Hash == "" {
			t.Error("Hash is empty")
		}
	}
}

func TestSaveAndLoadChecksums(t *testing.T) {
	tmpDir := t.TempDir()
	checksumFile := filepath.Join(tmpDir, "checksums.json")
	
	// Create test checksums
	original := map[string]ChecksumEntry{
		"file1.txt": {
			Hash:     "abc123",
			Size:     100,
			Modified: time.Now(),
			Original: true,
		},
		"file2.txt": {
			Hash:     "def456",
			Size:     200,
			Modified: time.Now(),
			Original: false,
		},
	}
	
	// Save checksums
	if err := SaveChecksums(checksumFile, original); err != nil {
		t.Fatalf("SaveChecksums failed: %v", err)
	}
	
	// Load checksums
	loaded, err := LoadChecksums(checksumFile)
	if err != nil {
		t.Fatalf("LoadChecksums failed: %v", err)
	}
	
	// Compare
	if len(loaded) != len(original) {
		t.Errorf("Expected %d entries, got %d", len(original), len(loaded))
	}
	
	for path, orig := range original {
		if load, ok := loaded[path]; !ok {
			t.Errorf("Path %s not found in loaded", path)
		} else {
			if load.Hash != orig.Hash {
				t.Errorf("Hash mismatch for %s", path)
			}
			if load.Size != orig.Size {
				t.Errorf("Size mismatch for %s", path)
			}
			if load.Original != orig.Original {
				t.Errorf("Original flag mismatch for %s", path)
			}
		}
	}
}

func TestCompareChecksums(t *testing.T) {
	stored := map[string]ChecksumEntry{
		"unchanged.txt": {Hash: "hash1", Size: 10},
		"modified.txt":  {Hash: "hash2", Size: 20},
		"deleted.txt":   {Hash: "hash3", Size: 30},
	}
	
	current := map[string]ChecksumEntry{
		"unchanged.txt": {Hash: "hash1", Size: 10},
		"modified.txt":  {Hash: "hash2-mod", Size: 25},
		"new.txt":       {Hash: "hash4", Size: 40},
	}
	
	unchanged, modified, added, deleted := CompareChecksums(stored, current)
	
	// Check unchanged
	if len(unchanged) != 1 || unchanged[0] != "unchanged.txt" {
		t.Errorf("Unexpected unchanged files: %v", unchanged)
	}
	
	// Check modified
	if len(modified) != 1 || modified[0] != "modified.txt" {
		t.Errorf("Unexpected modified files: %v", modified)
	}
	
	// Check added
	if len(added) != 1 || added[0] != "new.txt" {
		t.Errorf("Unexpected added files: %v", added)
	}
	
	// Check deleted
	if len(deleted) != 1 || deleted[0] != "deleted.txt" {
		t.Errorf("Unexpected deleted files: %v", deleted)
	}
}