package wizard

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

func TestListAllMarkers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test files with markers
	file1 := filepath.Join(tmpDir, "test1.md")
	content1 := `# Test 1
Project: [REPLACE: project name]
Type: [REPLACE: Web|CLI|Library]
`
	if err := os.WriteFile(file1, []byte(content1), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	file2 := filepath.Join(tmpDir, "test2.md")
	content2 := `# Test 2
Status: [REPLACE: Active|Inactive]
`
	if err := os.WriteFile(file2, []byte(content2), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &FillOptions{}
	files := []string{file1, file2}

	err := listAllMarkers(ctx, opts, files, tmpDir)

	// Restore stdout
	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listAllMarkers() error = %v", err)
	}

	// Read captured output
	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	// Parse JSON output
	var result struct {
		Files []struct {
			Path         string `json:"path"`
			RelativePath string `json:"relative_path"`
			Markers      []struct {
				Line    int      `json:"line"`
				Hint    string   `json:"hint"`
				Context string   `json:"context"`
				Options []string `json:"options,omitempty"`
			} `json:"markers"`
		} `json:"files"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON output: %v\nOutput: %s", err, output)
	}

	// Verify we have 2 files
	if len(result.Files) != 2 {
		t.Errorf("Expected 2 files, got %d", len(result.Files))
	}

	// Verify first file has 2 markers
	if len(result.Files) > 0 && len(result.Files[0].Markers) != 2 {
		t.Errorf("File 1: expected 2 markers, got %d", len(result.Files[0].Markers))
	}

	// Verify second file has 1 marker
	if len(result.Files) > 1 && len(result.Files[1].Markers) != 1 {
		t.Errorf("File 2: expected 1 marker, got %d", len(result.Files[1].Markers))
	}

	// Verify options are parsed for pipe-separated hints
	if len(result.Files) > 0 && len(result.Files[0].Markers) > 1 {
		marker := result.Files[0].Markers[1]
		if marker.Hint != "Web|CLI|Library" {
			t.Errorf("Expected hint 'Web|CLI|Library', got %q", marker.Hint)
		}
		expectedOptions := []string{"Web", "CLI", "Library"}
		if len(marker.Options) != len(expectedOptions) {
			t.Errorf("Expected %d options, got %d", len(expectedOptions), len(marker.Options))
		}
	}
}

func TestListAllMarkers_NoMarkers(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create file without markers
	file := filepath.Join(tmpDir, "test.md")
	if err := os.WriteFile(file, []byte("# Test\nNo markers\n"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	opts := &FillOptions{}
	files := []string{file}

	err := listAllMarkers(ctx, opts, files, tmpDir)

	w.Close()
	os.Stdout = oldStdout

	if err != nil {
		t.Fatalf("listAllMarkers() error = %v", err)
	}

	// Read output
	var buf strings.Builder
	io.Copy(&buf, r)
	output := buf.String()

	// Parse JSON
	var result struct {
		Files []interface{} `json:"files"`
	}

	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Should have null or empty files array
	if result.Files != nil && len(result.Files) > 0 {
		t.Errorf("Expected no files with markers, got %d", len(result.Files))
	}
}

func TestRunBatchFill(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create test file with markers
	testFile := filepath.Join(tmpDir, "test.md")
	content := `# Test
Project: [REPLACE: project name]
Type: [REPLACE: Web|CLI|Library]
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Create batch input JSON
	batchInput := filepath.Join(tmpDir, "batch.json")
	batchData := map[string]interface{}{
		"replacements": []map[string]interface{}{
			{
				"file": testFile,
				"markers": []map[string]string{
					{"hint": "project name", "value": "Test Project"},
					{"hint": "Web|CLI|Library", "value": "CLI"},
				},
			},
		},
	}
	batchJSON, _ := json.MarshalIndent(batchData, "", "  ")
	if err := os.WriteFile(batchInput, batchJSON, 0644); err != nil {
		t.Fatalf("WriteFile batch input error = %v", err)
	}

	opts := &FillOptions{
		BatchInput: batchInput,
		DryRun:     false,
	}

	// Create a display
	display := ui.New(false, false)

	err := runBatchFill(ctx, opts, display)
	if err != nil {
		t.Fatalf("runBatchFill() error = %v", err)
	}

	// Verify file was modified
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile after fill error = %v", err)
	}

	resultStr := string(result)
	if !strings.Contains(resultStr, "Test Project") {
		t.Errorf("Expected file to contain 'Test Project', got:\n%s", resultStr)
	}
	if !strings.Contains(resultStr, "Type: CLI") {
		t.Errorf("Expected file to contain 'Type: CLI', got:\n%s", resultStr)
	}
	if strings.Contains(resultStr, "[REPLACE:") {
		t.Errorf("Expected no REPLACE markers remaining, got:\n%s", resultStr)
	}
}

func TestRunBatchFill_DryRun(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.md")
	originalContent := `# Test
Project: [REPLACE: project name]
`
	if err := os.WriteFile(testFile, []byte(originalContent), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	batchInput := filepath.Join(tmpDir, "batch.json")
	batchData := map[string]interface{}{
		"replacements": []map[string]interface{}{
			{
				"file": testFile,
				"markers": []map[string]string{
					{"hint": "project name", "value": "Test"},
				},
			},
		},
	}
	batchJSON, _ := json.Marshal(batchData)
	if err := os.WriteFile(batchInput, batchJSON, 0644); err != nil {
		t.Fatalf("WriteFile batch input error = %v", err)
	}

	opts := &FillOptions{
		BatchInput: batchInput,
		DryRun:     true, // Dry run
	}

	display := ui.New(false, false)
	err := runBatchFill(ctx, opts, display)
	if err != nil {
		t.Fatalf("runBatchFill() error = %v", err)
	}

	// Verify file was NOT modified
	result, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("ReadFile error = %v", err)
	}

	if string(result) != originalContent {
		t.Errorf("File should not be modified in dry-run mode\nGot:\n%s\nWant:\n%s",
			string(result), originalContent)
	}
}

func TestRunBatchFill_IncompleteBatch(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.md")
	content := `# Test
Project: [REPLACE: project name]
Type: [REPLACE: Web|CLI]
`
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Batch input missing one marker
	batchInput := filepath.Join(tmpDir, "batch.json")
	batchData := map[string]interface{}{
		"replacements": []map[string]interface{}{
			{
				"file": testFile,
				"markers": []map[string]string{
					{"hint": "project name", "value": "Test"}, // Missing "Web|CLI" marker
				},
			},
		},
	}
	batchJSON, _ := json.Marshal(batchData)
	if err := os.WriteFile(batchInput, batchJSON, 0644); err != nil {
		t.Fatalf("WriteFile batch input error = %v", err)
	}

	opts := &FillOptions{
		BatchInput: batchInput,
	}

	display := ui.New(false, false)
	err := runBatchFill(ctx, opts, display)

	// Should error due to incomplete batch
	if err == nil {
		t.Error("runBatchFill() expected error for incomplete batch, got nil")
	}

	// Error should mention "incomplete batch input"
	if !strings.Contains(err.Error(), "incomplete batch input") {
		t.Errorf("Expected 'incomplete batch input' error, got: %v", err)
	}
}

func TestRunBatchFill_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	batchInput := filepath.Join(tmpDir, "batch.json")
	if err := os.WriteFile(batchInput, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	opts := &FillOptions{
		BatchInput: batchInput,
	}

	display := ui.New(false, false)
	err := runBatchFill(ctx, opts, display)

	if err == nil {
		t.Error("runBatchFill() expected error for invalid JSON, got nil")
	}
}

