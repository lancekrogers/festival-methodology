package commands

import (
	"fmt"
	"os"

	commitcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/commit"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/config"
	contextcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/context"
	depscmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/deps"
	executecmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/execute"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/extensions"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/festival"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/gates"
	listcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/list"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/markers"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/migrate"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/navigation"
	nextcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/next"
	parsecmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/parse"
	progresscmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/research"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/status"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/structure"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/system"
	understandcmd "github.com/lancekrogers/festival-methodology/fest/internal/commands/understand"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/validation"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/wizard"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
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
	// Disable the default completion command - we provide our own with better docs
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	// Add custom completion command with better documentation
	completionCmd := system.NewCompletionCommand(rootCmd)
	completionCmd.GroupID = "system"
	rootCmd.AddCommand(completionCmd)

	// Define command groups for organized help output
	rootCmd.AddGroup(
		&cobra.Group{ID: "learning", Title: "Learning Commands:"},
		&cobra.Group{ID: "creation", Title: "Creation Commands:"},
		&cobra.Group{ID: "structure", Title: "Structure Commands:"},
		&cobra.Group{ID: "workflow", Title: "Workflow Commands:"},
		&cobra.Group{ID: "query", Title: "Query Commands:"},
		&cobra.Group{ID: "navigation", Title: "Navigation Commands:"},
		&cobra.Group{ID: "system", Title: "System Commands:"},
	)

	// Enforce being inside a festivals/ tree for most commands
	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		// Sync global flags to shared package for subpackages
		shared.SetVerbose(verbose)
		shared.SetNoColor(noColor)
		shared.SetConfigFile(configFile)
		ui.SetNoColor(noColor)

		// Allow root (help/version), init, system, count, go, shell-init, understand, config, extension, index, gates, research, show, status, progress, context, deps, next, execute, migrate, link, unlink, links, commit, commits, parse, completion, list, validate, and markers to run anywhere
		// Also allow subcommands of system, understand, config, extension, index, gates, research, show, status, progress, context, deps, next, execute, migrate, remove, renumber, reorder, validate, markers, wizard, and go
		if cmd == rootCmd || cmd.Name() == "init" || cmd.Name() == "system" || cmd.Name() == "help" || cmd.Name() == "tui" || cmd.Name() == "count" || cmd.Name() == "go" || cmd.Name() == "shell-init" || cmd.Name() == "understand" || cmd.Name() == "config" || cmd.Name() == "extension" || cmd.Name() == "index" || cmd.Name() == "gates" || cmd.Name() == "research" || cmd.Name() == "show" || cmd.Name() == "status" || cmd.Name() == "progress" || cmd.Name() == "context" || cmd.Name() == "deps" || cmd.Name() == "next" || cmd.Name() == "execute" || cmd.Name() == "migrate" || cmd.Name() == "link" || cmd.Name() == "unlink" || cmd.Name() == "links" || cmd.Name() == "commit" || cmd.Name() == "commits" || cmd.Name() == "parse" || cmd.Name() == "remove" || cmd.Name() == "renumber" || cmd.Name() == "reorder" || cmd.Name() == "validate" || cmd.Name() == "markers" || cmd.Name() == "wizard" || cmd.Name() == "completion" || cmd.Name() == "list" {
			return nil
		}
		// Check if parent is system, understand, config, extension, index, gates, research, show, status, context, deps, next, execute, migrate, remove, renumber, reorder, validate, markers, wizard, or go (for subcommands)
		if cmd.Parent() != nil && (cmd.Parent().Name() == "system" || cmd.Parent().Name() == "understand" || cmd.Parent().Name() == "config" || cmd.Parent().Name() == "extension" || cmd.Parent().Name() == "index" || cmd.Parent().Name() == "gates" || cmd.Parent().Name() == "research" || cmd.Parent().Name() == "show" || cmd.Parent().Name() == "status" || cmd.Parent().Name() == "context" || cmd.Parent().Name() == "deps" || cmd.Parent().Name() == "next" || cmd.Parent().Name() == "execute" || cmd.Parent().Name() == "migrate" || cmd.Parent().Name() == "remove" || cmd.Parent().Name() == "renumber" || cmd.Parent().Name() == "reorder" || cmd.Parent().Name() == "validate" || cmd.Parent().Name() == "markers" || cmd.Parent().Name() == "wizard" || cmd.Parent().Name() == "go") {
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

	markersCmd := markers.NewMarkersCommand()
	markersCmd.GroupID = "learning"
	rootCmd.AddCommand(markersCmd)

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

	linkCmd := navigation.NewLinkCommand()
	linkCmd.GroupID = "navigation"
	rootCmd.AddCommand(linkCmd)

	unlinkCmd := navigation.NewUnlinkCommand()
	unlinkCmd.GroupID = "navigation"
	rootCmd.AddCommand(unlinkCmd)

	linksCmd := navigation.NewLinksCommand()
	linksCmd.GroupID = "navigation"
	rootCmd.AddCommand(linksCmd)

	// === WORKFLOW COMMANDS ===
	executeCmd := executecmd.NewExecuteCommand()
	executeCmd.GroupID = "workflow"
	rootCmd.AddCommand(executeCmd)

	nextCmd := nextcmd.NewNextCommand()
	nextCmd.GroupID = "workflow"
	rootCmd.AddCommand(nextCmd)

	progressCmd := progresscmd.NewProgressCommand()
	progressCmd.GroupID = "workflow"
	rootCmd.AddCommand(progressCmd)

	// === QUERY COMMANDS ===
	showCmd := show.NewShowCommand()
	showCmd.GroupID = "query"
	rootCmd.AddCommand(showCmd)

	statusCmd := status.NewStatusCommand()
	statusCmd.GroupID = "query"
	rootCmd.AddCommand(statusCmd)

	contextCmd := contextcmd.NewContextCommand()
	contextCmd.GroupID = "query"
	rootCmd.AddCommand(contextCmd)

	depsCmd := depscmd.NewDepsCommand()
	depsCmd.GroupID = "query"
	rootCmd.AddCommand(depsCmd)

	listCmd := listcmd.NewListCommand()
	listCmd.GroupID = "query"
	rootCmd.AddCommand(listCmd)

	indexCmd := navigation.NewIndexCommand()
	indexCmd.GroupID = "system"
	rootCmd.AddCommand(indexCmd)

	migrateCmd := migrate.NewMigrateCommand()
	migrateCmd.GroupID = "system"
	rootCmd.AddCommand(migrateCmd)

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

	// Commit traceability commands
	commitCmd := commitcmd.NewCommitCommand()
	commitCmd.GroupID = "workflow"
	rootCmd.AddCommand(commitCmd)

	commitsCmd := commitcmd.NewCommitsCommand()
	commitsCmd.GroupID = "query"
	rootCmd.AddCommand(commitsCmd)

	// Parse command for structured output
	parseCmd := parsecmd.NewParseCommand()
	parseCmd.GroupID = "query"
	rootCmd.AddCommand(parseCmd)

	// Wizard command for guided assistance
	wizardCmd := wizard.NewWizardCommand()
	wizardCmd.GroupID = "learning"
	rootCmd.AddCommand(wizardCmd)
}
