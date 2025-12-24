package shared

// verbose holds the global verbose flag value.
// This is set by the root command initialization.
var verbose bool

// noColor holds the global no-color flag value.
// This is set by the root command initialization.
var noColor bool

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
