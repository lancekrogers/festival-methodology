// Package theme provides a high-contrast, adaptive huh theme for fest CLI.
// The theme uses colors that work on any terminal background, avoiding purple
// which conflicts with the user's festivals/ directory indicator.
package theme

import (
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

// FestTheme returns a high-contrast theme that works on any terminal background.
// It uses adaptive colors that automatically adjust for light/dark terminals
// and specifically avoids purple/magenta colors which can conflict with custom
// terminal backgrounds.
func FestTheme() *huh.Theme {
	// High-contrast adaptive colors that work on any background
	textColor := lipgloss.AdaptiveColor{Light: "#000000", Dark: "#FFFFFF"}
	titleColor := lipgloss.AdaptiveColor{Light: "#005FAF", Dark: "#00D7FF"}
	placeholderColor := lipgloss.AdaptiveColor{Light: "#808080", Dark: "#949494"}
	focusColor := lipgloss.AdaptiveColor{Light: "#FF8700", Dark: "#FFD700"}
	errorColor := lipgloss.AdaptiveColor{Light: "#D70000", Dark: "#FF5F87"}
	selectedColor := lipgloss.AdaptiveColor{Light: "#00AF00", Dark: "#00FF5F"}
	borderColor := lipgloss.AdaptiveColor{Light: "#005FAF", Dark: "#00D7FF"}

	// Start from ThemeBase for structure
	t := huh.ThemeBase()

	// === Focused field styles (active input) ===
	t.Focused.Title = lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	t.Focused.Description = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Focused.TextInput.Text = lipgloss.NewStyle().
		Foreground(textColor)

	t.Focused.TextInput.Cursor = lipgloss.NewStyle().
		Foreground(focusColor)

	t.Focused.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Focused.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(focusColor)

	t.Focused.SelectSelector = lipgloss.NewStyle().
		Foreground(focusColor).
		SetString("> ")

	t.Focused.SelectedOption = lipgloss.NewStyle().
		Foreground(selectedColor).
		Bold(true)

	t.Focused.Option = lipgloss.NewStyle().
		Foreground(textColor)

	t.Focused.ErrorMessage = lipgloss.NewStyle().
		Foreground(errorColor)

	t.Focused.ErrorIndicator = lipgloss.NewStyle().
		Foreground(errorColor).
		SetString("! ")

	t.Focused.FocusedButton = lipgloss.NewStyle().
		Background(focusColor).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Padding(0, 1)

	t.Focused.BlurredButton = lipgloss.NewStyle().
		Foreground(placeholderColor).
		Padding(0, 1)

	t.Focused.Base = lipgloss.NewStyle().
		BorderForeground(borderColor).
		BorderStyle(lipgloss.RoundedBorder()).
		Padding(0, 1)

	// === Blurred field styles (inactive inputs) ===
	t.Blurred.Title = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Blurred.Description = lipgloss.NewStyle().
		Foreground(placeholderColor).
		Faint(true)

	t.Blurred.TextInput.Text = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Blurred.TextInput.Placeholder = lipgloss.NewStyle().
		Foreground(placeholderColor).
		Faint(true)

	t.Blurred.TextInput.Prompt = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Blurred.Option = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Blurred.SelectedOption = lipgloss.NewStyle().
		Foreground(placeholderColor)

	t.Blurred.SelectSelector = lipgloss.NewStyle().
		Foreground(placeholderColor).
		SetString("  ")

	return t
}
