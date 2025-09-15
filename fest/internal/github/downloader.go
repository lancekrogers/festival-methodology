package github

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ProgressFunc is called during download to report progress
type ProgressFunc func(current, total int64, file string)

// Downloader handles downloading from GitHub
type Downloader struct {
	repoURL string
	branch  string
	client  *http.Client
	timeout int
	retry   int
}

// NewDownloader creates a new GitHub downloader
func NewDownloader(repoURL, branch string) *Downloader {
	return &Downloader{
		repoURL: repoURL,
		branch:  branch,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		timeout: 30,
		retry:   3,
	}
}

// SetTimeout sets the download timeout in seconds
func (d *Downloader) SetTimeout(seconds int) {
	d.timeout = seconds
	d.client.Timeout = time.Duration(seconds) * time.Second
}

// SetRetry sets the number of retry attempts
func (d *Downloader) SetRetry(count int) {
	d.retry = count
}

// Download downloads the festival directory structure
func (d *Downloader) Download(targetDir string, progress ProgressFunc) error {
	// Parse repository URL to get owner and repo name
	owner, repo, err := parseRepoURL(d.repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %w", err)
	}
	
	// Build raw content base URL
	baseURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s", owner, repo, d.branch)
	
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Get file list (simplified for this implementation)
	// In a full implementation, we would use GitHub API to list files
	files := getDefaultFileList()
	
	totalFiles := int64(len(files))
	currentFile := int64(0)
	
	// Download each file
	for _, file := range files {
		currentFile++
		if progress != nil {
			progress(currentFile, totalFiles, file)
		}
		
		fileURL := fmt.Sprintf("%s/festivals/%s", baseURL, file)
		targetPath := filepath.Join(targetDir, file)
		
		// Download with retry
		var lastErr error
		for attempt := 0; attempt < d.retry; attempt++ {
			if err := d.downloadFile(fileURL, targetPath); err != nil {
				lastErr = err
				time.Sleep(time.Duration(attempt+1) * time.Second) // Exponential backoff
				continue
			}
			lastErr = nil
			break
		}
		
		if lastErr != nil {
			return fmt.Errorf("failed to download %s after %d attempts: %w", file, d.retry, lastErr)
		}
	}
	
	return nil
}

// downloadFile downloads a single file
func (d *Downloader) downloadFile(url, targetPath string) error {
	// Create directory if needed
	targetDir := filepath.Dir(targetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Make HTTP request
	resp, err := d.client.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()
	
	// Check status code
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}
	
	// Create target file
	out, err := os.Create(targetPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer out.Close()
	
	// Copy content
	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	
	return out.Sync()
}

// parseRepoURL parses a GitHub repository URL
func parseRepoURL(url string) (owner, repo string, err error) {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")
	
	// Remove github.com
	url = strings.TrimPrefix(url, "github.com/")
	
	// Split owner and repo
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", fmt.Errorf("invalid repository URL format")
	}
	
	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")
	
	return owner, repo, nil
}

// getDefaultFileList returns the default festival file structure
// In a real implementation, this would query GitHub API
func getDefaultFileList() []string {
	return []string{
		"README.md",
		".festival/README.md",
		".festival/FESTIVAL_SOFTWARE_PROJECT_MANAGEMENT.md",
		".festival/PROJECT_MANAGEMENT_SYSTEM.md",
		".festival/VALIDATION_CHECKLIST.md",
		".festival/agents/festival_planning_agent.md",
		".festival/agents/festival_review_agent.md",
		".festival/agents/festival_methodology_manager.md",
		".festival/agents/INDEX.md",
		".festival/templates/FESTIVAL_OVERVIEW_TEMPLATE.md",
		".festival/templates/FESTIVAL_GOAL_TEMPLATE.md",
		".festival/templates/FESTIVAL_RULES_TEMPLATE.md",
		".festival/templates/COMMON_INTERFACES_TEMPLATE.md",
		".festival/templates/TASK_TEMPLATE.md",
		".festival/templates/TASK_TEMPLATE_SIMPLE.md",
		".festival/templates/PHASE_GOAL_TEMPLATE.md",
		".festival/templates/SEQUENCE_GOAL_TEMPLATE.md",
		".festival/templates/CONTEXT_TEMPLATE.md",
		".festival/templates/FESTIVAL_TODO_TEMPLATE.md",
		".festival/templates/INDEX.md",
		".festival/examples/TASK_EXAMPLES.md",
		".festival/examples/PHASE_GOAL_EXAMPLE.md",
		".festival/examples/SEQUENCE_GOAL_EXAMPLE.md",
		".festival/examples/INDEX.md",
		"active/.gitkeep",
		"planned/.gitkeep",
		"completed/.gitkeep",
		"archived/.gitkeep",
	}
}