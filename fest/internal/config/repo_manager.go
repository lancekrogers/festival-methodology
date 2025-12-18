package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// RepoManager handles config repo operations
type RepoManager struct {
	manifest *ConfigRepoManifest
}

// NewRepoManager creates a new repo manager
func NewRepoManager() (*RepoManager, error) {
	manifest, err := LoadRepoManifest()
	if err != nil {
		return nil, err
	}
	return &RepoManager{manifest: manifest}, nil
}

// Manifest returns the current manifest
func (rm *RepoManager) Manifest() *ConfigRepoManifest {
	return rm.manifest
}

// Add adds a new config repo (git clone or symlink local path)
func (rm *RepoManager) Add(name, source string) (*ConfigRepo, error) {
	// Check if name already exists
	if existing := rm.manifest.GetRepo(name); existing != nil {
		return nil, fmt.Errorf("config repo '%s' already exists", name)
	}

	// Ensure config-repos directory exists
	reposDir := ConfigReposPath()
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config-repos directory: %w", err)
	}

	localPath := RepoLocalPath(name)
	isGitRepo := IsGitURL(source)

	if isGitRepo {
		// Git clone
		cmd := exec.Command("git", "clone", source, localPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("failed to clone repo: %w", err)
		}
	} else {
		// Local path - create symlink
		absSource, err := filepath.Abs(source)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve source path: %w", err)
		}

		// Verify source exists
		info, err := os.Stat(absSource)
		if err != nil {
			return nil, fmt.Errorf("source path does not exist: %w", err)
		}
		if !info.IsDir() {
			return nil, fmt.Errorf("source path is not a directory")
		}

		// Create symlink
		if err := os.Symlink(absSource, localPath); err != nil {
			return nil, fmt.Errorf("failed to create symlink: %w", err)
		}
	}

	repo := ConfigRepo{
		Name:      name,
		Source:    source,
		LocalPath: localPath,
		IsGitRepo: isGitRepo,
		LastSync:  time.Now().UTC(),
	}

	rm.manifest.AddRepo(repo)
	if err := SaveRepoManifest(rm.manifest); err != nil {
		return nil, err
	}

	return &repo, nil
}

// Sync syncs a config repo (git pull for git repos, no-op for local)
func (rm *RepoManager) Sync(name string) error {
	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return fmt.Errorf("config repo '%s' not found", name)
	}

	if !repo.IsGitRepo {
		// Local symlink - nothing to sync
		repo.LastSync = time.Now().UTC()
		rm.manifest.AddRepo(*repo)
		return SaveRepoManifest(rm.manifest)
	}

	// Git pull
	cmd := exec.Command("git", "-C", repo.LocalPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to sync repo: %w", err)
	}

	repo.LastSync = time.Now().UTC()
	rm.manifest.AddRepo(*repo)
	return SaveRepoManifest(rm.manifest)
}

// SyncAll syncs all config repos
func (rm *RepoManager) SyncAll() error {
	var lastErr error
	for _, repo := range rm.manifest.Repos {
		if err := rm.Sync(repo.Name); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Use sets a config repo as the active one (creates symlink)
func (rm *RepoManager) Use(name string) error {
	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return fmt.Errorf("config repo '%s' not found", name)
	}

	// Verify local path exists
	if _, err := os.Stat(repo.LocalPath); err != nil {
		return fmt.Errorf("repo local path does not exist: %w", err)
	}

	activePath := ActiveConfigPath()

	// Remove existing symlink if present
	if _, err := os.Lstat(activePath); err == nil {
		if err := os.Remove(activePath); err != nil {
			return fmt.Errorf("failed to remove existing active symlink: %w", err)
		}
	}

	// Create new symlink
	if err := os.Symlink(repo.LocalPath, activePath); err != nil {
		return fmt.Errorf("failed to create active symlink: %w", err)
	}

	rm.manifest.Active = name
	return SaveRepoManifest(rm.manifest)
}

// Remove removes a config repo
func (rm *RepoManager) Remove(name string) error {
	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return fmt.Errorf("config repo '%s' not found", name)
	}

	// If this is the active repo, remove the active symlink first
	if rm.manifest.Active == name {
		activePath := ActiveConfigPath()
		if _, err := os.Lstat(activePath); err == nil {
			if err := os.Remove(activePath); err != nil {
				return fmt.Errorf("failed to remove active symlink: %w", err)
			}
		}
		rm.manifest.Active = ""
	}

	// Remove local path (only if it's a cloned repo, not a symlink to user's dir)
	if repo.IsGitRepo {
		if err := os.RemoveAll(repo.LocalPath); err != nil {
			return fmt.Errorf("failed to remove repo directory: %w", err)
		}
	} else {
		// For symlinks, just remove the symlink
		if err := os.Remove(repo.LocalPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove repo symlink: %w", err)
		}
	}

	rm.manifest.RemoveRepo(name)
	return SaveRepoManifest(rm.manifest)
}

// List returns all configured repos
func (rm *RepoManager) List() []ConfigRepo {
	return rm.manifest.Repos
}

// GetActive returns the currently active repo
func (rm *RepoManager) GetActive() *ConfigRepo {
	return rm.manifest.GetActiveRepo()
}

// GetActiveName returns the name of the active repo
func (rm *RepoManager) GetActiveName() string {
	return rm.manifest.Active
}

// ActiveUserPath returns the path to user/ directory in active config
// Returns empty string if no active config
func ActiveUserPath() string {
	activePath := ActiveConfigPath()
	if _, err := os.Stat(activePath); err != nil {
		return ""
	}
	userPath := filepath.Join(activePath, "user")
	if _, err := os.Stat(userPath); err != nil {
		return ""
	}
	return userPath
}

// ActiveFestivalsPath returns the path to festivals/.festival/ in active config
// Returns empty string if no active config
func ActiveFestivalsPath() string {
	activePath := ActiveConfigPath()
	if _, err := os.Stat(activePath); err != nil {
		return ""
	}
	festPath := filepath.Join(activePath, "festivals", ".festival")
	if _, err := os.Stat(festPath); err != nil {
		return ""
	}
	return festPath
}
