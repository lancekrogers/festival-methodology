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

// PanelOptions configures how a panel component is rendered.
type PanelOptions struct {
	// Title appears at the top of the panel.
	Title string
	// TitleColor sets the title text color (defaults to ValueColor).
	TitleColor lipgloss.Color
	// BorderStyle determines the panel border style.
	BorderStyle BorderStyle
	// BorderColor sets the border color (defaults to BorderColor).
	BorderColor lipgloss.Color
	// ContentPadding sets internal content padding [top, right, bottom, left].
	ContentPadding [4]int
	// Width sets the panel width (0 = auto-fit).
	Width int
}

// DefaultPanelOptions returns sensible defaults for panel rendering.
func DefaultPanelOptions() PanelOptions {
	return PanelOptions{
		Title:          "",
		TitleColor:     ValueColor,
		BorderStyle:    BorderRounded,
		BorderColor:    BorderColor,
		ContentPadding: [4]int{0, 1, 0, 1},
		Width:          0,
	}
}

// Panel creates a styled panel with optional title and border.
// Panels are containers for grouped content with visual separation.
//
// Usage:
//   // Simple panel with title
//   opts := DefaultPanelOptions()
//   opts.Title = "Status"
//   output := Panel("All systems operational", opts)
//
//   // Panel with custom colors and width
//   opts := DefaultPanelOptions()
//   opts.Title = "Error"
//   opts.TitleColor = lipgloss.Color("196")
//   opts.BorderColor = lipgloss.Color("196")
//   opts.Width = 60
//   output := Panel("Connection failed", opts)
func Panel(content string, opts PanelOptions) string {
	// Build the content with padding
	contentStyle := lipgloss.NewStyle().
		Padding(opts.ContentPadding[0], opts.ContentPadding[1], opts.ContentPadding[2], opts.ContentPadding[3])

	paddedContent := contentStyle.Render(content)

	// If there's a title, add it above the content
	if opts.Title != "" {
		titleStyle := lipgloss.NewStyle().
			Foreground(opts.TitleColor).
			Bold(true).
			Padding(0, 1)

		title := titleStyle.Render(opts.Title)
		paddedContent = title + "\n" + paddedContent
	}

	// Wrap in border
	borderOpts := BorderOptions{
		Style:   opts.BorderStyle,
		Color:   opts.BorderColor,
		Padding: [4]int{0, 0, 0, 0}, // No extra padding - content already padded
		Width:   opts.Width,
	}

	return Border(paddedContent, borderOpts)
}

// TitledPanel is a convenience function for creating a panel with a title.
func TitledPanel(title, content string) string {
	opts := DefaultPanelOptions()
	opts.Title = title
	return Panel(content, opts)
}

// InfoPanel creates a panel styled for informational content.
func InfoPanel(title, content string) string {
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = SuccessColor
	opts.BorderColor = SuccessColor
	return Panel(content, opts)
}

// WarningPanel creates a panel styled for warning content.
func WarningPanel(title, content string) string {
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = WarningColor
	opts.BorderColor = WarningColor
	return Panel(content, opts)
}

// ErrorPanel creates a panel styled for error content.
func ErrorPanel(title, content string) string {
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = ErrorColor
	opts.BorderColor = ErrorColor
	return Panel(content, opts)
}
