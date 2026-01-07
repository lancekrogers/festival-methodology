package system

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/github"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type syncOptions struct {
	source  string
	branch  string
	force   bool
	timeout int
	retry   int
	dryRun  bool
}

// NewSyncCommand creates the sync command
func NewSyncCommand() *cobra.Command {
	opts := &syncOptions{}

	cmd := &cobra.Command{
		Use:   "sync",
		Short: "System: Download latest fest templates from GitHub",
		Long: `Download the latest fest methodology templates from GitHub to ~/.config/fest/

This is a SYSTEM command that maintains fest itself, not your festival content.
It fetches the complete .festival/ template structure from the configured
repository and stores it locally for use with 'fest init' and 'fest system update'.

Run this periodically to get the latest methodology templates and documentation.`,
		Example: `  fest system sync                          # Use defaults from config
  fest system sync --source github.com/user/repo  # Sync from specific repo
  fest system sync --force                       # Overwrite existing cache`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(cmd.Context(), opts)
		},
	}

	cmd.Flags().StringVar(&opts.source, "source", "", "GitHub repository URL")
	cmd.Flags().StringVar(&opts.branch, "branch", "main", "Git branch to sync from")
	cmd.Flags().BoolVar(&opts.force, "force", false, "overwrite existing files without checking")
	cmd.Flags().IntVar(&opts.timeout, "timeout", 30, "timeout in seconds")
	cmd.Flags().IntVar(&opts.retry, "retry", 3, "number of retry attempts")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "show what would be downloaded")

	return cmd
}

func runSync(ctx context.Context, opts *syncOptions) error {
	if ctx == nil {
		ctx = context.Background()
	}

	// Determine target directory
	targetDir := filepath.Join(config.ConfigDir(), "festivals")

	// Create UI handler
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())

	// Load configuration
	cfg, err := config.Load(ctx, shared.GetConfigFile())
	if err != nil && opts.source == "" {
		return errors.Wrap(err, "no config found and no --source specified - use --source flag")
	}

	// Determine repository URL
	repoURL := opts.source
	if repoURL == "" && cfg != nil {
		repoURL = cfg.Repository.URL
	}
	if repoURL == "" {
		repoURL = config.DefaultRepositoryURL
	}

	// Parse repository to get owner and repo
	owner, repo, err := parseRepoURLForSync(repoURL)
	if err != nil {
		return errors.Validation("invalid repository URL").WithField("url", repoURL)
	}

	// Create downloader
	downloader := github.NewDownloader(repoURL, opts.branch)
	downloader.SetTimeout(opts.timeout)
	downloader.SetRetry(opts.retry)

	// Check if directory exists and we're not forcing
	if fileExists(targetDir) && !opts.force {
		display.Info("Checking for updates from %s...", repoURL)

		hasUpdates, changes, err := downloader.CheckForUpdates(owner, repo, targetDir)
		if err != nil {
			display.Warning("Failed to check for updates: %v", err)
			display.Info("Use --force to re-download")
			return nil
		}

		if !hasUpdates {
			display.Success("Fest templates are up to date!")
			return nil
		}

		// Show changes
		display.Info("\nUpdates available (%d changes):", len(changes))
		for _, change := range changes {
			display.Info("  %s", change)
		}

		// Prompt user
		if !display.Confirm("\nDownload updates?") {
			display.Info("Sync cancelled")
			return nil
		}
	}

	if opts.dryRun {
		display.Info("DRY RUN: Would sync from %s (branch: %s)", repoURL, opts.branch)
		return nil
	}

	display.Info("Syncing from %s (branch: %s)...", repoURL, opts.branch)

	// Download with progress
	progressBar := display.NewProgressBar("Downloading", -1)
	err = downloader.Download(targetDir, func(current, total int64, file string) {
		progressBar.Update(current, total, file)
	})
	progressBar.Finish()

	if err != nil {
		return errors.IO("downloading templates", err).WithField("url", repoURL)
	}

	// Update config with sync time
	if cfg != nil {
		cfg.LastSync = timeNow()
		if err := config.Save(ctx, cfg); err != nil {
			display.Warning("Failed to update config: %v", err)
		}
	}

	display.Success("Successfully synced fest templates to %s", targetDir)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func timeNow() string {
	return time.Now().Format(time.RFC3339)
}

func parseRepoURLForSync(url string) (owner, repo string, err error) {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Remove github.com
	url = strings.TrimPrefix(url, "github.com/")

	// Split owner and repo
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", "", errors.Validation("invalid repository URL format")
	}

	owner = parts[0]
	repo = strings.TrimSuffix(parts[1], ".git")

	return owner, repo, nil
}
