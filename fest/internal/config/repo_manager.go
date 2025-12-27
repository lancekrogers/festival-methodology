package config

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// RepoManager handles config repo operations
type RepoManager struct {
	manifest *ConfigRepoManifest
}

// NewRepoManager creates a new repo manager
func NewRepoManager(ctx context.Context) (*RepoManager, error) {
	manifest, err := LoadRepoManifest(ctx)
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
func (rm *RepoManager) Add(ctx context.Context, name, source string) (*ConfigRepo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Check if name already exists
	if existing := rm.manifest.GetRepo(name); existing != nil {
		return nil, errors.Validation("config repo already exists").WithField("name", name)
	}

	// Ensure config-repos directory exists
	reposDir := ConfigReposPath()
	if err := os.MkdirAll(reposDir, 0755); err != nil {
		return nil, errors.IO("creating config-repos directory", err).WithField("path", reposDir)
	}

	localPath := RepoLocalPath(name)
	isGitRepo := IsGitURL(source)

	if isGitRepo {
		// Git clone with context
		cmd := exec.CommandContext(ctx, "git", "clone", source, localPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, errors.IO("cloning repo", err).WithField("source", source)
		}
	} else {
		// Local path - create symlink
		absSource, err := filepath.Abs(source)
		if err != nil {
			return nil, errors.IO("resolving source path", err).WithField("source", source)
		}

		// Verify source exists
		info, err := os.Stat(absSource)
		if err != nil {
			return nil, errors.NotFound("source path").WithField("path", absSource)
		}
		if !info.IsDir() {
			return nil, errors.Validation("source path is not a directory").WithField("path", absSource)
		}

		// Create symlink
		if err := os.Symlink(absSource, localPath); err != nil {
			return nil, errors.IO("creating symlink", err).WithField("source", absSource).WithField("target", localPath)
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
	if err := SaveRepoManifest(ctx, rm.manifest); err != nil {
		return nil, err
	}

	return &repo, nil
}

// Sync syncs a config repo (git pull for git repos, no-op for local)
func (rm *RepoManager) Sync(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return errors.NotFound("config repo").WithField("name", name)
	}

	if !repo.IsGitRepo {
		// Local symlink - nothing to sync
		repo.LastSync = time.Now().UTC()
		rm.manifest.AddRepo(*repo)
		return SaveRepoManifest(ctx, rm.manifest)
	}

	// Git pull with context
	cmd := exec.CommandContext(ctx, "git", "-C", repo.LocalPath, "pull")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return errors.IO("syncing repo", err).WithField("name", name)
	}

	repo.LastSync = time.Now().UTC()
	rm.manifest.AddRepo(*repo)
	return SaveRepoManifest(ctx, rm.manifest)
}

// SyncAll syncs all config repos
func (rm *RepoManager) SyncAll(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	var lastErr error
	for _, repo := range rm.manifest.Repos {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := rm.Sync(ctx, repo.Name); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// Use sets a config repo as the active one (creates symlink)
func (rm *RepoManager) Use(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return errors.NotFound("config repo").WithField("name", name)
	}

	// Verify local path exists
	if _, err := os.Stat(repo.LocalPath); err != nil {
		return errors.NotFound("repo local path").WithField("path", repo.LocalPath)
	}

	activePath := ActiveConfigPath()

	// Remove existing symlink if present
	if _, err := os.Lstat(activePath); err == nil {
		if err := os.Remove(activePath); err != nil {
			return errors.IO("removing existing active symlink", err).WithField("path", activePath)
		}
	}

	// Create new symlink
	if err := os.Symlink(repo.LocalPath, activePath); err != nil {
		return errors.IO("creating active symlink", err).WithField("source", repo.LocalPath).WithField("target", activePath)
	}

	rm.manifest.Active = name
	return SaveRepoManifest(ctx, rm.manifest)
}

// Remove removes a config repo
func (rm *RepoManager) Remove(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	repo := rm.manifest.GetRepo(name)
	if repo == nil {
		return errors.NotFound("config repo").WithField("name", name)
	}

	// If this is the active repo, remove the active symlink first
	if rm.manifest.Active == name {
		activePath := ActiveConfigPath()
		if _, err := os.Lstat(activePath); err == nil {
			if err := os.Remove(activePath); err != nil {
				return errors.IO("removing active symlink", err).WithField("path", activePath)
			}
		}
		rm.manifest.Active = ""
	}

	// Remove local path (only if it's a cloned repo, not a symlink to user's dir)
	if repo.IsGitRepo {
		if err := os.RemoveAll(repo.LocalPath); err != nil {
			return errors.IO("removing repo directory", err).WithField("path", repo.LocalPath)
		}
	} else {
		// For symlinks, just remove the symlink
		if err := os.Remove(repo.LocalPath); err != nil && !os.IsNotExist(err) {
			return errors.IO("removing repo symlink", err).WithField("path", repo.LocalPath)
		}
	}

	rm.manifest.RemoveRepo(name)
	return SaveRepoManifest(ctx, rm.manifest)
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
