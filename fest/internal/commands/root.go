package commands

import (
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/structure"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/validation"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
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

// IsVerbose returns the global verbose flag value.
func IsVerbose() bool {
	return verbose
}

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
	// Disable the completion command - defaults are not useful for fest
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Enforce being inside a festivals/ tree for most commands
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Sync global flags to shared package for subpackages
		shared.SetVerbose(verbose)
		shared.SetNoColor(noColor)

		// Allow root (help/version), init, sync, count, go, shell-init, understand, config, extension, index, gates, and validate to run anywhere
		// Also allow subcommands of understand, config, extension, index, gates, remove, renumber, reorder, and validate
		if cmd == rootCmd || cmd.Name() == "init" || cmd.Name() == "sync" || cmd.Name() == "help" || cmd.Name() == "tui" || cmd.Name() == "count" || cmd.Name() == "go" || cmd.Name() == "shell-init" || cmd.Name() == "understand" || cmd.Name() == "config" || cmd.Name() == "extension" || cmd.Name() == "index" || cmd.Name() == "gates" || cmd.Name() == "remove" || cmd.Name() == "renumber" || cmd.Name() == "reorder" || cmd.Name() == "validate" {
			return nil
		}
		// Check if parent is understand, config, extension, index, gates, remove, renumber, reorder, or validate (for subcommands)
		if cmd.Parent() != nil && (cmd.Parent().Name() == "understand" || cmd.Parent().Name() == "config" || cmd.Parent().Name() == "extension" || cmd.Parent().Name() == "index" || cmd.Parent().Name() == "gates" || cmd.Parent().Name() == "remove" || cmd.Parent().Name() == "renumber" || cmd.Parent().Name() == "reorder" || cmd.Parent().Name() == "validate") {
			return nil
		}
		cwd, _ := os.Getwd()
		if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
			// Standardize the message expected by callers
			return fmt.Errorf("no festivals/ directory detected")
		}
		return nil
	}
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default: ~/.config/fest/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// Add commands
	rootCmd.AddCommand(NewSyncCommand())
	rootCmd.AddCommand(NewInitCommand())
	rootCmd.AddCommand(NewTUICommand())
	rootCmd.AddCommand(NewUpdateCommand())
	rootCmd.AddCommand(NewCountCommand())
	rootCmd.AddCommand(structure.NewRenumberCommand())
	rootCmd.AddCommand(structure.NewReorderCommand())
	rootCmd.AddCommand(structure.NewInsertCommand())
	rootCmd.AddCommand(structure.NewRemoveCommand())
	// Headless-first creation commands
	rootCmd.AddCommand(NewApplyCommand())
	// Grouped under 'create'
	createCmd := &cobra.Command{Use: "create", Short: "Create festival elements",
		RunE: func(cmd *cobra.Command, args []string) error {
			return StartCreateTUI()
		},
	}
	createCmd.AddCommand(NewCreateFestivalCommand())
	createCmd.AddCommand(NewCreatePhaseCommand())
	createCmd.AddCommand(NewCreateSequenceCommand())
	createCmd.AddCommand(NewCreateTaskCommand())
	rootCmd.AddCommand(createCmd)

	// Task management commands
	rootCmd.AddCommand(NewTaskCommand())

	// Methodology learning command
	rootCmd.AddCommand(NewUnderstandCommand())

	// Validation command
	rootCmd.AddCommand(validation.NewValidateCommand())

	// Navigation command
	rootCmd.AddCommand(navigation.NewGoCommand())

	// Shell integration command
	rootCmd.AddCommand(NewShellInitCommand())

	// Config repo management
	rootCmd.AddCommand(NewConfigCommand())

	// Extension management
	rootCmd.AddCommand(NewExtensionCommand())

	// Index generation for Guild integration
	rootCmd.AddCommand(navigation.NewIndexCommand())

	// Gates policy management
	rootCmd.AddCommand(gates.NewGatesCommand())
}
