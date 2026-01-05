// Package ui provides UI components and styling for the fest CLI.
// This file defines the status color scheme used throughout the fest CLI
// to provide visual distinction between different festival statuses.
package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Status color scheme for festival statuses.
// These colors provide visual distinction across different terminal themes
// using the ANSI 256-color palette which is widely supported in modern terminals.
var (
	ActiveColor    = lipgloss.Color("42")  // Green - currently executing
	PlannedColor   = lipgloss.Color("33")  // Blue - future work
	CompletedColor = lipgloss.Color("205") // Purple/Magenta - finished successfully
	ArchivedColor  = lipgloss.Color("245") // Grey - deprioritized/paused
	DungeonColor   = lipgloss.Color("240") // Dark grey - deep archived storage
)

// Entity type colors for visual hierarchy in fest CLI output.
// Each entity type gets a distinct color to make nested structures easier to scan.
//
// Usage examples:
//   - FestivalColor: Use for festival names in list/show commands
//   - PhaseColor: Use for phase headers and identifiers
//   - SequenceColor: Use for sequence names in progress/status output
//   - TaskColor: Use for task items in lists and progress displays
//   - GateColor: Use for quality gate indicators and validation output
var (
	FestivalColor = ActiveColor           // Green (42) - reuse active color for top-level entities
	PhaseColor    = PlannedColor          // Blue (33) - reuse planned color for major divisions
	SequenceColor = lipgloss.Color("51")  // Cyan - distinct color for mid-level groupings
	TaskColor     = lipgloss.Color("141") // Purple - for individual work items
	GateColor     = lipgloss.Color("214") // Orange - for quality gates and checkpoints
)

// State colors for progress and status indication across all entity types.
// These supplement the festival status colors with more granular workflow states.
//
// Usage examples:
//   - PendingColor: Tasks/sequences not yet started
//   - InProgressColor: Currently being worked on (use with animation/indicators)
//   - BlockedColor: Waiting on dependencies or resolution
//   - CompletedColor: Successfully finished (shared with festival completed status)
var (
	PendingColor    = lipgloss.Color("245") // Grey/dim - not yet started
	InProgressColor = lipgloss.Color("220") // Yellow/amber - actively being worked
	BlockedColor    = lipgloss.Color("196") // Red - blocked/waiting
	// CompletedColor already defined in status colors above (205 - purple/magenta)
)

// Structural element colors for UI components and formatting.
// These support common visual patterns across all fest CLI commands.
//
// Usage examples:
//   - BorderColor: Panel borders, separators, boxes
//   - ValueColor: Important values (progress percentages, counts)
//   - MetadataColor: Less important info (paths, IDs, timestamps)
//   - SuccessColor: Success messages and positive indicators
//   - WarningColor: Warning messages and caution indicators
//   - ErrorColor: Error messages and failure indicators
var (
	BorderColor   = lipgloss.Color("240") // Subtle grey - borders, separators
	ValueColor    = lipgloss.Color("255") // Bright white - emphasized values
	MetadataColor = lipgloss.Color("245") // Dim grey - paths, IDs, secondary info
	SuccessColor  = ActiveColor           // Green (42) - reuse active color for success
	WarningColor  = InProgressColor       // Yellow (220) - reuse in-progress color for warnings
	ErrorColor    = BlockedColor          // Red (196) - reuse blocked color for errors
)

// Status styles with colors and bold formatting for headers.
// These pre-configured styles can be used directly or as templates.
var (
	ActiveStyle    = lipgloss.NewStyle().Foreground(ActiveColor).Bold(true)
	PlannedStyle   = lipgloss.NewStyle().Foreground(PlannedColor).Bold(true)
	CompletedStyle = lipgloss.NewStyle().Foreground(CompletedColor).Bold(true)
	ArchivedStyle  = lipgloss.NewStyle().Foreground(ArchivedColor).Bold(true)
	DungeonStyle   = lipgloss.NewStyle().Foreground(DungeonColor).Bold(true)
)

// GetStatusStyle returns the appropriate lipgloss style for a given status string.
// Status matching is case-insensitive. Returns an unstyled lipgloss.Style for unknown statuses.
func GetStatusStyle(status string) lipgloss.Style {
	switch strings.ToLower(status) {
	case "active":
		return ActiveStyle
	case "planned":
		return PlannedStyle
	case "completed":
		return CompletedStyle
	case "archived":
		return ArchivedStyle
	case "dungeon":
		return DungeonStyle
	default:
		return lipgloss.NewStyle()
	}
}

// GetStatusColor returns the appropriate color for a given status string.
// This is useful when you need just the color without the full style.
func GetStatusColor(status string) lipgloss.Color {
	switch strings.ToLower(status) {
	case "active":
		return ActiveColor
	case "planned":
		return PlannedColor
	case "completed":
		return CompletedColor
	case "archived":
		return ArchivedColor
	case "dungeon":
		return DungeonColor
	default:
		return lipgloss.Color("")
	}
}
