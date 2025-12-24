package shared

import "github.com/spf13/cobra"

// verbose holds the global verbose flag value.
// This is set by the root command initialization.
var verbose bool

// noColor holds the global no-color flag value.
// This is set by the root command initialization.
var noColor bool

// configFile holds the path to the config file.
// This is set by the root command initialization.
var configFile string

// Option types for command execution.
// These are defined here so both commands and tui packages can use them
// without creating import cycles.

// InitOpts holds options for the init command.
type InitOpts struct {
	From        string
	Minimal     bool
	NoChecksums bool
}

// CreateFestivalOpts holds options for the create festival command.
type CreateFestivalOpts struct {
	Name       string
	Goal       string
	Tags       string
	VarsFile   string
	JSONOutput bool
	Dest       string
}

// CreatePhaseOpts holds options for the create phase command.
type CreatePhaseOpts struct {
	After      int
	Name       string
	PhaseType  string
	Path       string
	VarsFile   string
	JSONOutput bool
}

// CreateSequenceOpts holds options for the create sequence command.
type CreateSequenceOpts struct {
	After      int
	Name       string
	Path       string
	VarsFile   string
	JSONOutput bool
}

// CreateTaskOpts holds options for the create task command.
type CreateTaskOpts struct {
	After      int
	Names      []string
	Path       string
	VarsFile   string
	JSONOutput bool
}

// ApplyOpts holds options for the apply command.
type ApplyOpts struct {
	TemplateID   string
	TemplatePath string
	DestPath     string
	VarsFile     string
	JSONOutput   bool
}

// TUI function hooks - set by tui package to break import cycle.
// These are initialized by the tui package's init() function.
var (
	// StartCreateTUI launches the TUI for 'create' subcommand selection.
	StartCreateTUI func() error

	// StartCreateFestivalTUI launches the TUI for creating a festival.
	StartCreateFestivalTUI func() error

	// StartCreatePhaseTUI launches the TUI for creating a phase.
	StartCreatePhaseTUI func() error

	// StartCreateSequenceTUI launches the TUI for creating a sequence.
	StartCreateSequenceTUI func() error

	// StartCreateTaskTUI launches the TUI for creating a task.
	StartCreateTaskTUI func() error

	// NewTUICommand creates the TUI cobra command.
	// Set by tui package's init() function.
	NewTUICommand func() *cobra.Command
)

// Command execution hooks - set by commands package to break import cycle.
// These are initialized by the commands package's init() function.
var (
	// RunInit executes the init command.
	RunInit func(path string, opts *InitOpts) error

	// RunCreateFestival executes the create festival command.
	RunCreateFestival func(opts *CreateFestivalOpts) error

	// RunCreatePhase executes the create phase command.
	RunCreatePhase func(opts *CreatePhaseOpts) error

	// RunCreateSequence executes the create sequence command.
	RunCreateSequence func(opts *CreateSequenceOpts) error

	// RunCreateTask executes the create task command.
	RunCreateTask func(opts *CreateTaskOpts) error

	// RunApply executes the apply command.
	RunApply func(opts *ApplyOpts) error
)

// SetVerbose sets the global verbose flag value.
func SetVerbose(v bool) {
	verbose = v
}

// IsVerbose returns the global verbose flag value.
func IsVerbose() bool {
	return verbose
}

// SetNoColor sets the global no-color flag value.
func SetNoColor(v bool) {
	noColor = v
}

// IsNoColor returns the global no-color flag value.
func IsNoColor() bool {
	return noColor
}

// SetConfigFile sets the global config file path.
func SetConfigFile(c string) {
	configFile = c
}

// GetConfigFile returns the global config file path.
func GetConfigFile() string {
	return configFile
}
