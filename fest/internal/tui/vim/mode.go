// Package vim provides vim mode state management and UI indicators for fest TUI.
package vim

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Mode represents the current vim editing mode.
type Mode int

const (
	// ModeNormal is the default navigation mode where keystrokes perform actions.
	ModeNormal Mode = iota
	// ModeInsert is the text entry mode where keystrokes insert characters.
	ModeInsert
)

// String returns the display name for the mode.
func (m Mode) String() string {
	switch m {
	case ModeInsert:
		return "INSERT"
	default:
		return "NORMAL"
	}
}

// ModeIndicator renders a styled mode indicator for TUI display.
type ModeIndicator struct {
	// Mode is the current vim mode.
	Mode Mode
	// Enabled controls whether the indicator is shown.
	Enabled bool
}

// NewIndicator creates a new mode indicator with the given enabled state.
func NewIndicator(enabled bool) *ModeIndicator {
	return &ModeIndicator{
		Mode:    ModeNormal,
		Enabled: enabled,
	}
}

// SetMode updates the current mode.
func (m *ModeIndicator) SetMode(mode Mode) {
	m.Mode = mode
}

// Toggle switches between normal and insert modes.
func (m *ModeIndicator) Toggle() {
	if m.Mode == ModeNormal {
		m.Mode = ModeInsert
	} else {
		m.Mode = ModeNormal
	}
}

// Render returns the styled mode indicator string.
// Returns empty string if vim mode is disabled.
func (m *ModeIndicator) Render() string {
	if !m.Enabled {
		return ""
	}
	return renderMode(m.Mode)
}

// renderMode returns the styled mode string.
func renderMode(mode Mode) string {
	var style lipgloss.Style

	switch mode {
	case ModeInsert:
		// INSERT mode: green/accent to indicate active text entry
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#00AF00", Dark: "#00FF5F"}).
			Bold(true)
	default:
		// NORMAL mode: subtle gray to indicate navigation mode
		style = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#808080", Dark: "#949494"})
	}

	return style.Render(mode.String())
}

// TitleWithMode returns a title string with the mode indicator appended.
// If vim mode is disabled, returns the original title unchanged.
//
// Example output:
//   - Vim enabled, NORMAL mode: "Description (NORMAL)"
//   - Vim enabled, INSERT mode: "Description (INSERT)"
//   - Vim disabled: "Description"
func (m *ModeIndicator) TitleWithMode(title string) string {
	if !m.Enabled {
		return title
	}
	return fmt.Sprintf("%s (%s)", title, m.Render())
}

// FormatTitleWithMode formats a title with mode indicator based on config.
// This is a convenience function for one-off formatting without creating an indicator.
func FormatTitleWithMode(title string, mode Mode, enabled bool) string {
	if !enabled {
		return title
	}
	return fmt.Sprintf("%s (%s)", title, renderMode(mode))
}
