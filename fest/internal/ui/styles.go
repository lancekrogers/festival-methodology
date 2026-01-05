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
