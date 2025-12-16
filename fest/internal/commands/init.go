package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/fileops"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

type initOptions struct {
	force       bool
	from        string
	minimal     bool
	noChecksums bool
	register    bool
	unregister  bool
}

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	opts := &initOptions{}

	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new festival directory structure",
		Long: `Initialize a new festival directory structure in the current or specified directory.

This command copies the festival template structure from your local cache
(populated by 'fest sync') and creates initial checksum tracking.

Workspace Registration:
  Use --register to mark an existing festivals/ directory as your active workspace.
  This enables 'fest go' to navigate directly to it from anywhere in the project.

  Use --unregister to remove the workspace marker, making the festivals/
  directory invisible to 'fest go' (useful for source repositories).`,
		Example: `  fest init                      # Initialize in current directory
  fest init ./my-project         # Initialize in specified directory
  fest init --force              # Overwrite existing festival
  fest init --minimal            # Create minimal structure only
  fest init --register           # Register existing festivals as workspace
  fest init --unregister         # Remove workspace registration`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			targetPath := "."
			if len(args) > 0 {
				targetPath = args[0]
			}
			return runInit(targetPath, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "overwrite existing festival directory")
	cmd.Flags().StringVar(&opts.from, "from", "", "source directory (default: ~/.config/fest)")
	cmd.Flags().BoolVar(&opts.minimal, "minimal", false, "create minimal structure only")
	cmd.Flags().BoolVar(&opts.noChecksums, "no-checksums", false, "skip checksum generation")
	cmd.Flags().BoolVar(&opts.register, "register", false, "register existing festivals as active workspace")
	cmd.Flags().BoolVar(&opts.unregister, "unregister", false, "remove workspace registration")

	return cmd
}

func runInit(targetPath string, opts *initOptions) error {
	// Create UI handler
	display := ui.New(noColor, verbose)

	// Convert to absolute path
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("invalid path: %w", err)
	}

	// Handle --register flag: register existing festivals directory
	if opts.register {
		return runRegister(absPath, display)
	}

	// Handle --unregister flag: remove workspace marker
	if opts.unregister {
		return runUnregister(absPath, display)
	}

	// Check if festival already exists
	festivalPath := filepath.Join(absPath, "festivals")
	if fileops.Exists(festivalPath) && !opts.force {
		if !display.Confirm("Festival directory already exists at %s. Overwrite?", festivalPath) {
			display.Warning("Initialization cancelled")
			return nil
		}
	}

	// Determine source directory
	sourceDir := opts.from
	if sourceDir == "" {
		sourceDir = filepath.Join(config.ConfigDir(), "festivals")
	}

	// Check if source exists
	if !fileops.Exists(sourceDir) {
		return fmt.Errorf("source directory not found at %s. Run 'fest sync' first", sourceDir)
	}

	display.Info("Initializing festival structure at %s...", festivalPath)

	// Create festivals directory if it doesn't exist
	if err := os.MkdirAll(festivalPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Copy structure
	copier := fileops.NewCopier()
	if opts.minimal {
		// Copy only essential directories
		essentialDirs := []string{".festival", "active", "planned"}
		for _, dir := range essentialDirs {
			src := filepath.Join(sourceDir, dir)
			dst := filepath.Join(festivalPath, dir)
			if fileops.Exists(src) {
				if err := copier.CopyDirectory(src, dst); err != nil {
					return fmt.Errorf("failed to copy %s: %w", dir, err)
				}
			}
		}
	} else {
		// Copy everything
		if err := copier.CopyDirectory(sourceDir, festivalPath); err != nil {
			return fmt.Errorf("failed to copy festival structure: %w", err)
		}
	}

	// Generate checksums unless disabled
	if !opts.noChecksums {
		display.Info("Generating .festival checksums...")
		checksumFile := filepath.Join(festivalPath, ".festival", ".fest-checksums.json")

		// Only checksum the .festival directory
		festivalMetaDir := filepath.Join(festivalPath, ".festival")
		checksums, err := fileops.GenerateChecksums(festivalMetaDir)
		if err != nil {
			return fmt.Errorf("failed to generate checksums: %w", err)
		}

		if err := fileops.SaveChecksums(checksumFile, checksums); err != nil {
			return fmt.Errorf("failed to save checksums: %w", err)
		}

		display.Info("Created checksum tracking at %s", checksumFile)
	}

	// Auto-register the new festivals directory as workspace
	if err := workspace.RegisterFestivals(festivalPath); err != nil {
		display.Warning("Could not register workspace marker: %v", err)
	} else {
		marker, _ := workspace.ReadMarker(festivalPath)
		if marker != nil {
			display.Info("Registered as workspace: %s", marker.Workspace)
		}
	}

	// Show summary
	display.Success("Successfully initialized festival structure at %s", festivalPath)
	display.Info("\nNext steps:")
	display.Info("  1. cd %s", absPath)
	display.Info("  2. Review festivals/.festival/README.md")
	display.Info("  3. Start planning your festival in festivals/planned/")
	display.Info("\nWorkspace navigation:")
	display.Info("  cd $(fest go)              # Navigate to festivals from anywhere")

	return nil
}

// runRegister registers an existing festivals directory as the active workspace
func runRegister(targetPath string, display *ui.UI) error {
	// Find the festivals directory
	festivalsDir, err := findFestivalsDir(targetPath)
	if err != nil {
		return err
	}

	// Check if already registered
	if workspace.HasMarker(festivalsDir) {
		marker, _ := workspace.ReadMarker(festivalsDir)
		if marker != nil {
			display.Info("Already registered as workspace: %s", marker.Workspace)
			return nil
		}
	}

	// Register
	if err := workspace.RegisterFestivals(festivalsDir); err != nil {
		return fmt.Errorf("failed to register workspace: %w", err)
	}

	marker, _ := workspace.ReadMarker(festivalsDir)
	wsName := ""
	if marker != nil {
		wsName = marker.Workspace
	}

	display.Success("Registered %s as workspace: %s", festivalsDir, wsName)
	display.Info("You can now use 'cd $(fest go)' from anywhere in this project")

	return nil
}

// runUnregister removes the workspace marker from a festivals directory
func runUnregister(targetPath string, display *ui.UI) error {
	// Find the festivals directory
	festivalsDir, err := findFestivalsDir(targetPath)
	if err != nil {
		return err
	}

	// Check if registered
	if !workspace.HasMarker(festivalsDir) {
		display.Info("No workspace marker found at %s", festivalsDir)
		return nil
	}

	// Get workspace name before removing
	marker, _ := workspace.ReadMarker(festivalsDir)
	wsName := ""
	if marker != nil {
		wsName = marker.Workspace
	}

	// Unregister
	if err := workspace.UnregisterFestivals(festivalsDir); err != nil {
		return fmt.Errorf("failed to unregister workspace: %w", err)
	}

	display.Success("Unregistered workspace: %s", wsName)
	display.Info("This festivals directory will no longer be found by 'fest go'")

	return nil
}

// findFestivalsDir locates the festivals directory from a given path
func findFestivalsDir(targetPath string) (string, error) {
	// Check if target is already a festivals directory
	if filepath.Base(targetPath) == "festivals" {
		if info, err := os.Stat(targetPath); err == nil && info.IsDir() {
			return targetPath, nil
		}
	}

	// Check if target contains a festivals directory
	festivalsDir := filepath.Join(targetPath, "festivals")
	if info, err := os.Stat(festivalsDir); err == nil && info.IsDir() {
		return festivalsDir, nil
	}

	// Walk up looking for festivals directory
	nearest, err := workspace.FindNearestFestivals(targetPath)
	if err != nil {
		return "", fmt.Errorf("failed to find festivals directory: %w", err)
	}
	if nearest == "" {
		return "", fmt.Errorf("no festivals directory found from %s", targetPath)
	}

	return nearest, nil
}
