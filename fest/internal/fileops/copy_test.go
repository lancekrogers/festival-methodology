package fileops

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create source file
	srcPath := filepath.Join(tmpDir, "source.txt")
	srcContent := []byte("test content")
	if err := os.WriteFile(srcPath, srcContent, 0644); err != nil {
		t.Fatal(err)
	}

	// Copy file
	dstPath := filepath.Join(tmpDir, "dest.txt")
	if err := CopyFile(ctx, srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	// Verify destination file
	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(dstContent) != string(srcContent) {
		t.Error("Content mismatch after copy")
	}
}

func TestCopyDirectory(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create source directory structure
	srcDir := filepath.Join(tmpDir, "source")
	files := map[string]string{
		"file1.txt":         "content1",
		"subdir/file2.txt":  "content2",
		".hidden/file3.txt": "content3",
	}

	for path, content := range files {
		fullPath := filepath.Join(srcDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Copy directory
	dstDir := filepath.Join(tmpDir, "dest")
	copier := NewCopier()
	if err := copier.CopyDirectory(ctx, srcDir, dstDir); err != nil {
		t.Fatalf("CopyDirectory failed: %v", err)
	}

	// Verify all files were copied
	for path, expectedContent := range files {
		dstPath := filepath.Join(dstDir, path)
		content, err := os.ReadFile(dstPath)
		if err != nil {
			t.Errorf("Failed to read %s: %v", path, err)
			continue
		}
		if string(content) != expectedContent {
			t.Errorf("Content mismatch for %s", path)
		}
	}
}

func TestCopyDirectoryCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	dstDir := filepath.Join(tmpDir, "dest")
	copier := NewCopier()
	err := copier.CopyDirectory(ctx, srcDir, dstDir)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()

	// Test existing file
	existingFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(existingFile, []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	if !Exists(existingFile) {
		t.Error("Exists returned false for existing file")
	}

	// Test non-existing file
	nonExistingFile := filepath.Join(tmpDir, "does-not-exist.txt")
	if Exists(nonExistingFile) {
		t.Error("Exists returned true for non-existing file")
	}
}

func TestCreateBackup(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create source directory
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create test file
	testFile := filepath.Join(srcDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create backup
	backupDir := filepath.Join(tmpDir, "backup")
	if err := CreateBackup(ctx, srcDir, backupDir); err != nil {
		t.Fatalf("CreateBackup failed: %v", err)
	}

	// Verify backup
	backupFile := filepath.Join(backupDir, "test.txt")
	if !Exists(backupFile) {
		t.Error("Backup file not created")
	}

	// Check manifest
	manifestFile := filepath.Join(backupDir, "manifest.json")
	if !Exists(manifestFile) {
		t.Error("Manifest file not created")
	}
}

func TestCreateBackupCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "source")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	backupDir := filepath.Join(tmpDir, "backup")
	err := CreateBackup(ctx, srcDir, backupDir)
	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}
