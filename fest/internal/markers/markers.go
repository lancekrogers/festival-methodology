// Package markers provides REPLACE marker parsing and replacement for festival templates.
package markers

// Marker represents a single [REPLACE: hint] marker in content.
type Marker struct {
	FullMatch   string // "[REPLACE: hint text]"
	Hint        string // "hint text"
	LineNumber  int    // Line where marker appears (1-indexed)
	StartOffset int    // Character offset in content
	EndOffset   int    // End character offset
}

// MarkerValue holds a marker and its replacement value.
type MarkerValue struct {
	Marker Marker
	Value  string
}

// Result contains the outcome of marker processing.
type Result struct {
	FilePath       string   `json:"file_path"`
	TotalMarkers   int      `json:"total_markers"`
	FilledMarkers  int      `json:"filled_markers"`
	UnfilledHints  []string `json:"unfilled_hints,omitempty"`
}
