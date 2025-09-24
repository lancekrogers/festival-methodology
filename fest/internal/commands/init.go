package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/config"
	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/fileops"
	"github.com/anthropics/guild-framework/projects/festival-methodology/fest/internal/ui"
	"github.com/spf13/cobra"
)

type initOptions struct {
	force       bool
	from        string
	minimal     bool
	noChecksums bool
}

// NewInitCommand creates the init command
func NewInitCommand() *cobra.Command {
	opts := &initOptions{}
	
	cmd := &cobra.Command{
		Use:   "init [path]",
		Short: "Initialize a new festival directory structure",
		Long: `Initialize a new festival directory structure in the current or specified directory.
		
This command copies the festival template structure from your local cache
(populated by 'fest sync') and creates initial checksum tracking.`,
		Example: `  fest init                      # Initialize in current directory
  fest init ./my-project         # Initialize in specified directory
  fest init --force             # Overwrite existing festival
  fest init --minimal           # Create minimal structure only`,
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
		display.Info("Generating checksums...")
		checksumFile := filepath.Join(festivalPath, ".fest-checksums.json")
		
		checksums, err := fileops.GenerateChecksums(festivalPath)
		if err != nil {
			return fmt.Errorf("failed to generate checksums: %w", err)
		}
		
		if err := fileops.SaveChecksums(checksumFile, checksums); err != nil {
			return fmt.Errorf("failed to save checksums: %w", err)
		}
		
		display.Info("Created checksum tracking at %s", checksumFile)
	}
	
	// Show summary
	display.Success("Successfully initialized festival structure at %s", festivalPath)
	display.Info("\nNext steps:")
	display.Info("  1. cd %s", absPath)
	display.Info("  2. Review festivals/.festival/README.md")
	display.Info("  3. Start planning your festival in festivals/planned/")
	
	return nil
}