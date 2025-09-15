package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	// Global flags
	configFile string
	verbose    bool
	noColor    bool
	debug      bool
	
	// Version information (set at build time)
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "fest",
	Short: "Festival Methodology CLI tool",
	Long: `fest is a CLI tool for managing Festival Methodology files.
	
It helps you initialize, sync, and update festival directories while
preserving your modifications and ensuring you always have the latest
templates available.`,
	Version: fmt.Sprintf("%s (built %s, commit %s)", Version, BuildTime, GitCommit),
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default: ~/.config/fest/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	
	// Add commands
	rootCmd.AddCommand(NewSyncCommand())
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewUpdateCommand())
	rootCmd.AddCommand(NewCountCommand())
	rootCmd.AddCommand(NewRenumberCommand())
	rootCmd.AddCommand(NewInsertCommand())
	rootCmd.AddCommand(NewRemoveCommand())
}