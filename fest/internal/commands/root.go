package commands

import (
	"fmt"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/extensions"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/gates"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/navigation"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/research"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/structure"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/system"
	understandcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/understand"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/validation"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/spf13/cobra"

	// Import tui package for its init() side effects (registers hooks)
	_ "github.com/lancekrogers/festival-methodology/fest/internal/commands/tui"
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
	Short: "Festival Methodology CLI - goal-oriented project management for AI agents",
	Long: `fest manages Festival Methodology - a goal-oriented project management
system designed for AI agent development workflows.

GETTING STARTED (for AI agents):
  fest understand methodology    Learn core principles first
  fest understand structure      Understand the 3-level hierarchy
  fest understand tasks          Learn when/how to create task files
  fest validate                  Check if a festival is properly structured

COMMON WORKFLOWS:
  fest create festival           Create a new festival (interactive TUI)
  fest create phase/sequence     Add phases or sequences to existing festival
  fest validate --fix            Fix common structure issues automatically
  fest go                        Navigate to festivals directory

SYSTEM MAINTENANCE:
  fest system sync               Download latest templates from source
  fest system update             Update .festival/ methodology files

Run 'fest understand' to learn the methodology before executing tasks.`,
	Version: fmt.Sprintf("%s (built %s, commit %s)", Version, BuildTime, GitCommit),
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Disable the completion command - defaults are not useful for fest
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Define command groups for organized help output
	rootCmd.AddGroup(
		&cobra.Group{ID: "learning", Title: "Learning Commands:"},
		&cobra.Group{ID: "creation", Title: "Creation Commands:"},
		&cobra.Group{ID: "structure", Title: "Structure Commands:"},
		&cobra.Group{ID: "navigation", Title: "Navigation Commands:"},
		&cobra.Group{ID: "system", Title: "System Commands:"},
	)

	// Enforce being inside a festivals/ tree for most commands
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Sync global flags to shared package for subpackages
		shared.SetVerbose(verbose)
		shared.SetNoColor(noColor)
		shared.SetConfigFile(configFile)

		// Allow root (help/version), init, system, count, go, shell-init, understand, config, extension, index, gates, research, and validate to run anywhere
		// Also allow subcommands of system, understand, config, extension, index, gates, research, remove, renumber, reorder, and validate
		if cmd == rootCmd || cmd.Name() == "init" || cmd.Name() == "system" || cmd.Name() == "help" || cmd.Name() == "tui" || cmd.Name() == "count" || cmd.Name() == "go" || cmd.Name() == "shell-init" || cmd.Name() == "understand" || cmd.Name() == "config" || cmd.Name() == "extension" || cmd.Name() == "index" || cmd.Name() == "gates" || cmd.Name() == "research" || cmd.Name() == "remove" || cmd.Name() == "renumber" || cmd.Name() == "reorder" || cmd.Name() == "validate" {
			return nil
		}
		// Check if parent is system, understand, config, extension, index, gates, research, remove, renumber, reorder, or validate (for subcommands)
		if cmd.Parent() != nil && (cmd.Parent().Name() == "system" || cmd.Parent().Name() == "understand" || cmd.Parent().Name() == "config" || cmd.Parent().Name() == "extension" || cmd.Parent().Name() == "index" || cmd.Parent().Name() == "gates" || cmd.Parent().Name() == "research" || cmd.Parent().Name() == "remove" || cmd.Parent().Name() == "renumber" || cmd.Parent().Name() == "reorder" || cmd.Parent().Name() == "validate") {
			return nil
		}
		cwd, _ := os.Getwd()
		if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
			// Standardize the message expected by callers
			return errors.NotFound("festivals/ directory")
		}
		return nil
	}
	// Global flags
	rootCmd.PersistentFlags().StringVar(&configFile, "config", "", "config file (default: ~/.config/fest/config.json)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "disable colored output")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")

	// === LEARNING COMMANDS ===
	understandCmd := understandcmd.NewUnderstandCommand()
	understandCmd.GroupID = "learning"
	rootCmd.AddCommand(understandCmd)

	validateCmd := validation.NewValidateCommand()
	validateCmd.GroupID = "learning"
	rootCmd.AddCommand(validateCmd)

	// === CREATION COMMANDS ===
	createCmd := &cobra.Command{
		Use:     "create",
		Short:   "Create festivals, phases, sequences, or tasks (TUI)",
		GroupID: "creation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return shared.StartCreateTUI(cmd.Context())
		},
	}
	createCmd.AddCommand(festival.NewCreateFestivalCommand())
	createCmd.AddCommand(festival.NewCreatePhaseCommand())
	createCmd.AddCommand(festival.NewCreateSequenceCommand())
	createCmd.AddCommand(festival.NewCreateTaskCommand())
	rootCmd.AddCommand(createCmd)

	insertCmd := structure.NewInsertCommand()
	insertCmd.GroupID = "creation"
	rootCmd.AddCommand(insertCmd)

	applyCmd := festival.NewApplyCommand()
	applyCmd.GroupID = "creation"
	rootCmd.AddCommand(applyCmd)

	// === STRUCTURE COMMANDS ===
	renumberCmd := structure.NewRenumberCommand()
	renumberCmd.GroupID = "structure"
	rootCmd.AddCommand(renumberCmd)

	reorderCmd := structure.NewReorderCommand()
	reorderCmd.GroupID = "structure"
	rootCmd.AddCommand(reorderCmd)

	removeCmd := structure.NewRemoveCommand()
	removeCmd.GroupID = "structure"
	rootCmd.AddCommand(removeCmd)

	// === NAVIGATION COMMANDS ===
	goCmd := navigation.NewGoCommand()
	goCmd.GroupID = "navigation"
	rootCmd.AddCommand(goCmd)

	indexCmd := navigation.NewIndexCommand()
	indexCmd.GroupID = "navigation"
	rootCmd.AddCommand(indexCmd)

	// === SYSTEM COMMANDS ===
	initCmd := system.NewInitCommand()
	initCmd.GroupID = "system"
	rootCmd.AddCommand(initCmd)

	systemCmd := system.NewSystemCommand()
	systemCmd.GroupID = "system"
	rootCmd.AddCommand(systemCmd)

	configCmd := config.NewConfigCommand()
	configCmd.GroupID = "system"
	rootCmd.AddCommand(configCmd)

	shellInitCmd := config.NewShellInitCommand()
	shellInitCmd.GroupID = "system"
	rootCmd.AddCommand(shellInitCmd)

	extensionCmd := extensions.NewExtensionCommand()
	extensionCmd.GroupID = "system"
	rootCmd.AddCommand(extensionCmd)

	countCmd := system.NewCountCommand()
	countCmd.GroupID = "system"
	rootCmd.AddCommand(countCmd)

	if shared.NewTUICommand != nil {
		tuiCmd := shared.NewTUICommand()
		tuiCmd.GroupID = "creation"
		rootCmd.AddCommand(tuiCmd)
	}

	// Gates policy management (part of learning/validation)
	gatesCmd := gates.NewGatesCommand()
	gatesCmd.GroupID = "learning"
	rootCmd.AddCommand(gatesCmd)

	// Research document management
	researchCmd := research.NewResearchCommand()
	researchCmd.GroupID = "creation"
	rootCmd.AddCommand(researchCmd)
}
