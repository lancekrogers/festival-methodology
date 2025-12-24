package shared

// verbose holds the global verbose flag value.
// This is set by the root command initialization.
var verbose bool

// SetVerbose sets the global verbose flag value.
func SetVerbose(v bool) {
	verbose = v
}

// IsVerbose returns the global verbose flag value.
func IsVerbose() bool {
	return verbose
}
