package github

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
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

// GitHubTreeResponse represents the GitHub API tree response
type GitHubTreeResponse struct {
	SHA  string           `json:"sha"`
	Tree []GitHubTreeItem `json:"tree"`
}

// GitHubTreeItem represents a single item in the GitHub tree
type GitHubTreeItem struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Size int64  `json:"size,omitempty"`
	SHA  string `json:"sha,omitempty"`
}

// FileInfo contains path and SHA for a file
type FileInfo struct {
	Path string
	SHA  string
}

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

	// Get file list from GitHub API
	files, err := d.getFilesFromGitHub(owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get file list: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no files found in festivals/ directory")
	}

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

// getFilesFromGitHub fetches the file list from GitHub API
func (d *Downloader) getFilesFromGitHub(owner, repo string) ([]string, error) {
	// Build API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, d.branch)

	// Make API request
	resp, err := d.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file tree from GitHub API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var treeResp GitHubTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Filter for files in festivals/ directory
	var files []string
	for _, item := range treeResp.Tree {
		// Only include files (blobs), not directories (trees)
		if item.Type != "blob" {
			continue
		}

		// Only include files under festivals/ directory
		if strings.HasPrefix(item.Path, "festivals/") {
			// Remove "festivals/" prefix to get relative path
			relativePath := strings.TrimPrefix(item.Path, "festivals/")
			files = append(files, relativePath)
		}
	}

	return files, nil
}

// getFilesWithSHA fetches the file list with SHA hashes from GitHub API
func (d *Downloader) getFilesWithSHA(owner, repo string) (map[string]string, error) {
	// Build API URL
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/git/trees/%s?recursive=1", owner, repo, d.branch)

	// Make API request
	resp, err := d.client.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch file tree from GitHub API: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API request failed with status: %d", resp.StatusCode)
	}

	// Parse response
	var treeResp GitHubTreeResponse
	if err := json.NewDecoder(resp.Body).Decode(&treeResp); err != nil {
		return nil, fmt.Errorf("failed to parse GitHub API response: %w", err)
	}

	// Build map of path -> SHA for files in festivals/ directory
	files := make(map[string]string)
	for _, item := range treeResp.Tree {
		// Only include files (blobs), not directories (trees)
		if item.Type != "blob" {
			continue
		}

		// Only include files under festivals/ directory
		if strings.HasPrefix(item.Path, "festivals/") {
			// Remove "festivals/" prefix to get relative path
			relativePath := strings.TrimPrefix(item.Path, "festivals/")
			files[relativePath] = item.SHA
		}
	}

	return files, nil
}

// calculateGitBlobSHA calculates the Git blob SHA1 hash for a file
// Git blob SHA = SHA1("blob <size>\0<content>")
func calculateGitBlobSHA(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Git blob format: "blob <size>\0<content>"
	header := fmt.Sprintf("blob %d\x00", len(content))

	h := sha1.New()
	h.Write([]byte(header))
	h.Write(content)

	return hex.EncodeToString(h.Sum(nil)), nil
}

// CheckForUpdates checks if remote repository has changes compared to local cache
func (d *Downloader) CheckForUpdates(owner, repo, targetDir string) (bool, []string, error) {
	// Get remote file list with SHAs
	remoteFiles, err := d.getFilesWithSHA(owner, repo)
	if err != nil {
		return false, nil, fmt.Errorf("failed to get remote file list: %w", err)
	}

	changes := []string{}

	// Check for new or modified files
	for remotePath, remoteSHA := range remoteFiles {
		localPath := filepath.Join(targetDir, remotePath)
		info, err := os.Stat(localPath)
		if os.IsNotExist(err) {
			// File doesn't exist locally (new file)
			changes = append(changes, fmt.Sprintf("+ %s (new)", remotePath))
			continue
		}
		if err != nil {
			// Other error, skip this file
			continue
		}
		if info.IsDir() {
			// Unexpected: remote is a file but local is a directory
			continue
		}

		// File exists locally - compare SHA hashes
		localSHA, err := calculateGitBlobSHA(localPath)
		if err != nil {
			// Can't calculate local SHA, skip comparison
			continue
		}

		if localSHA != remoteSHA {
			changes = append(changes, fmt.Sprintf("~ %s (modified)", remotePath))
		}
	}

	// Check for deleted files (in local but not in remote)
	err = filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if info.IsDir() {
			return nil // Skip directories
		}

		// Get relative path
		relPath, err := filepath.Rel(targetDir, path)
		if err != nil {
			return nil
		}

		// Skip hidden files and checksums
		if strings.HasPrefix(filepath.Base(relPath), ".") {
			return nil
		}

		// Check if file exists in remote
		if _, exists := remoteFiles[relPath]; !exists {
			changes = append(changes, fmt.Sprintf("- %s (deleted)", relPath))
		}

		return nil
	})

	if err != nil {
		return false, nil, fmt.Errorf("failed to walk local directory: %w", err)
	}

	hasUpdates := len(changes) > 0
	return hasUpdates, changes, nil
}
