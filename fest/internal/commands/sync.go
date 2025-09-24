package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/config"
	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/github"
	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/ui"
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
		Short: "Download latest festival templates from GitHub",
		Long: `Download the latest festival templates from GitHub to ~/.config/fest/
		
This command fetches the complete festivals/ directory structure from the
configured repository and stores it locally for use with init and update commands.`,
		Example: `  fest sync                          # Use defaults from config
  fest sync --source github.com/user/repo  # Sync from specific repo
  fest sync --force                       # Overwrite existing cache`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSync(opts)
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

func runSync(opts *syncOptions) error {
	// Load configuration
	cfg, err := config.Load(configFile)
	if err != nil && opts.source == "" {
		return fmt.Errorf("no config found and no --source specified: %w", err)
	}
	
	// Determine repository URL
	repoURL := opts.source
	if repoURL == "" && cfg != nil {
		repoURL = cfg.Repository.URL
	}
	if repoURL == "" {
		repoURL = config.DefaultRepositoryURL
	}
	
	// Create UI handler
	display := ui.New(noColor, verbose)
	
	if opts.dryRun {
		display.Info("DRY RUN: Would sync from %s (branch: %s)", repoURL, opts.branch)
		return nil
	}
	
	display.Info("Syncing from %s (branch: %s)...", repoURL, opts.branch)
	
	// Create downloader
	downloader := github.NewDownloader(repoURL, opts.branch)
	downloader.SetTimeout(opts.timeout)
	downloader.SetRetry(opts.retry)
	
	// Determine target directory
	targetDir := filepath.Join(config.ConfigDir(), "festivals")
	
	// Check if exists and not forcing
	if !opts.force && fileExists(targetDir) {
		if !display.Confirm("Local cache exists. Overwrite?") {
			display.Warning("Sync cancelled")
			return nil
		}
	}
	
	// Download with progress
	progressBar := display.NewProgressBar("Downloading", -1)
	err = downloader.Download(targetDir, func(current, total int64, file string) {
		progressBar.Update(current, total, file)
	})
	progressBar.Finish()
	
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	
	// Update config with sync time
	if cfg != nil {
		cfg.LastSync = timeNow()
		if err := config.Save(cfg); err != nil {
			display.Warning("Failed to update config: %v", err)
		}
	}
	
	display.Success("Successfully synced festival templates to %s", targetDir)
	return nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func timeNow() string {
	return time.Now().Format(time.RFC3339)
}