package navigation

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
	"github.com/spf13/cobra"
)

type moveOptions struct {
	copy      bool
	force     bool
	toProject bool
	verbose   bool
	json      bool
}

// MoveResult represents the outcome of a move operation
type MoveResult struct {
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Operation   string `json:"operation"` // "move" or "copy"
	Success     bool   `json:"success"`
	Message     string `json:"message,omitempty"`
}

// NewGoMoveCommand creates the go move subcommand
func NewGoMoveCommand() *cobra.Command {
	opts := &moveOptions{}

	cmd := &cobra.Command{
		Use:   "move <source> [destination]",
		Short: "Move files between festival and linked project",
		Long: `Move or copy files between a festival and its linked project directory.

The command automatically detects which direction to move based on your
current directory:

  - In festival directory: moves TO linked project
  - In linked project: moves TO festival

Examples:
  # In project directory, move file to festival
  fest go move ./analysis.md

  # In festival directory, move file to project
  fest go move ./specs/api.go --to-project

  # Copy instead of move (keeps original)
  fest go move --copy ./notes.md

  # Force overwrite existing files
  fest go move --force ./config.yml

Requirements:
  - Festival must have project_path set in fest.yaml
  - Must be in either festival or linked project directory`,
		Args: cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			source := args[0]
			destination := ""
			if len(args) > 1 {
				destination = args[1]
			}
			return runMove(cmd.Context(), source, destination, opts)
		},
	}

	cmd.Flags().BoolVarP(&opts.copy, "copy", "c", false, "copy instead of move")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "overwrite existing files")
	cmd.Flags().BoolVar(&opts.toProject, "to-project", false, "move from festival to project")
	cmd.Flags().BoolVarP(&opts.verbose, "verbose", "v", false, "show detailed output")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

// LocationInfo contains information about the current location
type LocationInfo struct {
	Type         string // "festival" or "project"
	FestivalPath string // Path to festival root
	ProjectPath  string // Path to linked project
	RelativePath string // Current path relative to root
}

func runMove(ctx context.Context, source, destination string, opts *moveOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location (festival or project)
	loc, err := detectLocation(ctx, cwd)
	if err != nil {
		return err
	}

	// Resolve source path
	sourcePath := source
	if !filepath.IsAbs(source) {
		sourcePath = filepath.Join(cwd, source)
	}

	// Check source exists
	sourceInfo, err := os.Stat(sourcePath)
	if os.IsNotExist(err) {
		return errors.NotFound("source file").WithField("path", source)
	}
	if err != nil {
		return errors.IO("checking source file", err).WithField("path", source)
	}

	// Determine target directory
	var targetDir string
	switch loc.Type {
	case "festival":
		if opts.toProject || destination != "" {
			targetDir = loc.ProjectPath
		} else {
			// Default: stay in festival (might want to move within festival)
			return errors.Validation("use --to-project to move from festival to project")
		}
	case "project":
		targetDir = loc.FestivalPath
	default:
		return errors.Validation("not in a festival or linked project directory")
	}

	// Compute destination path
	destPath := computeDestination(sourcePath, destination, targetDir, loc)

	// Check destination
	if _, err := os.Stat(destPath); err == nil {
		if !opts.force {
			return errors.Validation("destination already exists").
				WithField("path", destPath).
				WithField("hint", "use --force to overwrite")
		}
	}

	// Ensure destination directory exists
	destDir := filepath.Dir(destPath)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.IO("creating destination directory", err).WithField("path", destDir)
	}

	// Execute move or copy
	operation := "move"
	if opts.copy {
		operation = "copy"
		if err := copyFile(sourcePath, destPath, sourceInfo.IsDir()); err != nil {
			return err
		}
	} else {
		if err := os.Rename(sourcePath, destPath); err != nil {
			// Rename fails across filesystems, fall back to copy+delete
			if err := copyFile(sourcePath, destPath, sourceInfo.IsDir()); err != nil {
				return err
			}
			if err := os.RemoveAll(sourcePath); err != nil {
				return errors.IO("removing source after copy", err).WithField("path", sourcePath)
			}
		}
	}

	// Output result
	result := MoveResult{
		Source:      sourcePath,
		Destination: destPath,
		Operation:   operation,
		Success:     true,
		Message:     fmt.Sprintf("%s -> %s", source, destPath),
	}

	if opts.json {
		fmt.Printf(`{"source": "%s", "destination": "%s", "operation": "%s", "success": true}%s`,
			result.Source, result.Destination, result.Operation, "\n")
	} else {
		display := ui.New(shared.IsNoColor(), opts.verbose)
		if opts.verbose {
			display.Success("%s: %s → %s", strings.Title(operation), sourcePath, destPath)
		} else {
			display.Success("%s → %s", filepath.Base(sourcePath), destPath)
		}
	}
	_ = result // use result to avoid unused variable warning

	return nil
}

func detectLocation(ctx context.Context, cwd string) (*LocationInfo, error) {
	// Check if we're inside a festival
	if isInsideFestival(cwd) {
		loc, err := show.DetectCurrentLocation(ctx, cwd)
		if err != nil {
			return nil, errors.Wrap(err, "detecting festival location")
		}
		if loc == nil || loc.Festival == nil {
			return nil, errors.NotFound("festival")
		}

		// Get project path from fest.yaml
		festivalPath := loc.Festival.Path
		cfg, err := config.LoadFestivalConfig(festivalPath)
		if err != nil {
			return nil, errors.Wrap(err, "loading festival config")
		}

		if cfg.ProjectPath == "" {
			return nil, errors.Validation("festival has no linked project").
				WithField("hint", "set project_path in fest.yaml")
		}

		// Compute relative path within festival
		relPath, _ := filepath.Rel(festivalPath, cwd)

		return &LocationInfo{
			Type:         "festival",
			FestivalPath: festivalPath,
			ProjectPath:  cfg.ProjectPath,
			RelativePath: relPath,
		}, nil
	}

	// Check if we're in a linked project
	nav, err := navigation.LoadNavigation()
	if err != nil {
		return nil, errors.Wrap(err, "loading navigation state")
	}

	festivalName := nav.FindFestivalForPath(cwd)
	if festivalName == "" {
		return nil, errors.Validation("not in a festival or linked project directory").
			WithField("hint", "run from festival or linked project, or set up a link")
	}

	// Find the festival's path
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return nil, errors.Wrap(err, "finding festivals directory")
	}

	festPath := resolveFestivalPath(festivalsDir, festivalName)
	if festPath == "" {
		return nil, errors.NotFound("festival").WithField("name", festivalName)
	}

	// Get the linked project path from navigation
	projectPath := nav.GetLinkedProject(festivalName)
	if projectPath == "" {
		return nil, errors.Validation("no linked project path found")
	}

	// Compute relative path within project
	relPath, _ := filepath.Rel(projectPath, cwd)

	return &LocationInfo{
		Type:         "project",
		FestivalPath: festPath,
		ProjectPath:  projectPath,
		RelativePath: relPath,
	}, nil
}

func computeDestination(source, destination, targetDir string, loc *LocationInfo) string {
	filename := filepath.Base(source)

	// If explicit destination provided
	if destination != "" {
		if filepath.IsAbs(destination) {
			return destination
		}
		// Relative to target directory
		destPath := filepath.Join(targetDir, destination)
		// If destination looks like a directory path, append filename
		if strings.HasSuffix(destination, "/") || !strings.Contains(filepath.Base(destination), ".") {
			destPath = filepath.Join(destPath, filename)
		}
		return destPath
	}

	// Default: same relative path in target
	if loc.RelativePath != "" && loc.RelativePath != "." {
		return filepath.Join(targetDir, loc.RelativePath, filename)
	}

	return filepath.Join(targetDir, filename)
}

func copyFile(src, dst string, isDir bool) error {
	if isDir {
		return copyDir(src, dst)
	}
	return copySingleFile(src, dst)
}

func copySingleFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return errors.IO("opening source file", err).WithField("path", src)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return errors.IO("creating destination file", err).WithField("path", dst)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return errors.IO("copying file", err).WithField("from", src).WithField("to", dst)
	}

	// Preserve permissions
	sourceInfo, _ := os.Stat(src)
	return os.Chmod(dst, sourceInfo.Mode())
}

func copyDir(src, dst string) error {
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return errors.IO("checking source directory", err).WithField("path", src)
	}

	if err := os.MkdirAll(dst, sourceInfo.Mode()); err != nil {
		return errors.IO("creating destination directory", err).WithField("path", dst)
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return errors.IO("reading source directory", err).WithField("path", src)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copySingleFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// Note: resolveFestivalPath is defined in go_link.go
