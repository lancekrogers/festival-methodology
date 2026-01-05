// Package ui provides reusable lipgloss UI components for fest CLI.
// This file defines border, panel, header, and progress bar components
// that can be used across all fest commands for consistent styling.
package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// BorderStyle defines the type of border to render.
type BorderStyle int

const (
	// BorderRounded uses rounded corners for a modern, friendly appearance.
	BorderRounded BorderStyle = iota
	// BorderSquare uses sharp corners for a classic terminal look.
	BorderSquare
	// BorderMinimal uses simple ASCII characters for maximum compatibility.
	BorderMinimal
)

// BorderOptions configures how a border component is rendered.
type BorderOptions struct {
	// Style determines the border character set (rounded, square, minimal).
	Style BorderStyle
	// Color sets the border line color (defaults to BorderColor from styles.go).
	Color lipgloss.Color
	// Padding sets internal padding (top, right, bottom, left).
	Padding [4]int
	// Width sets the total width (0 = auto-fit content).
	Width int
}

// DefaultBorderOptions returns sensible defaults for border rendering.
func DefaultBorderOptions() BorderOptions {
	return BorderOptions{
		Style:   BorderRounded,
		Color:   BorderColor,
		Padding: [4]int{0, 1, 0, 1}, // Horizontal padding only by default
		Width:   0,
	}
}

// Border creates a styled border around the given content.
//
// Usage:
//   // Simple border with defaults
//   output := Border("Hello World", DefaultBorderOptions())
//
//   // Custom border with square style and custom color
//   opts := DefaultBorderOptions()
//   opts.Style = BorderSquare
//   opts.Color = lipgloss.Color("42")
//   output := Border("Status: Active", opts)
func Border(content string, opts BorderOptions) string {
	style := lipgloss.NewStyle().
		BorderForeground(opts.Color).
		Padding(opts.Padding[0], opts.Padding[1], opts.Padding[2], opts.Padding[3])

	// Set border style based on options
	switch opts.Style {
	case BorderRounded:
		style = style.Border(lipgloss.RoundedBorder())
	case BorderSquare:
		style = style.Border(lipgloss.NormalBorder())
	case BorderMinimal:
		style = style.Border(lipgloss.Border{
			Top:         "-",
			Bottom:      "-",
			Left:        "|",
			Right:       "|",
			TopLeft:     "+",
			TopRight:    "+",
			BottomLeft:  "+",
			BottomRight: "+",
		})
	}

	// Set width if specified
	if opts.Width > 0 {
		style = style.Width(opts.Width - 2) // Account for border characters
	}

	return style.Render(content)
}

// RoundedBorder is a convenience function for creating rounded borders.
// Equivalent to Border(content, DefaultBorderOptions()).
func RoundedBorder(content string) string {
	return Border(content, DefaultBorderOptions())
}

// SquareBorder is a convenience function for creating square borders.
func SquareBorder(content string) string {
	opts := DefaultBorderOptions()
	opts.Style = BorderSquare
	return Border(content, opts)
}

// MinimalBorder is a convenience function for creating minimal ASCII borders.
// Useful for maximum terminal compatibility.
func MinimalBorder(content string) string {
	opts := DefaultBorderOptions()
	opts.Style = BorderMinimal
	return Border(content, opts)
}
