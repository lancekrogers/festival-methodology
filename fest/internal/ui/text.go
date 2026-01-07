package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	labelStyle   = lipgloss.NewStyle().Foreground(MetadataColor).Bold(true)
	valueStyle   = lipgloss.NewStyle().Foreground(ValueColor).Bold(true)
	dimStyle     = lipgloss.NewStyle().Foreground(MetadataColor).Faint(true)
	infoStyle    = lipgloss.NewStyle().Foreground(MetadataColor)
	successStyle = lipgloss.NewStyle().Foreground(SuccessColor).Bold(true)
	warningStyle = lipgloss.NewStyle().Foreground(WarningColor).Bold(true)
	errorStyle   = lipgloss.NewStyle().Foreground(ErrorColor).Bold(true)
)

// Label styles a short label for key/value pairs or section metadata.
func Label(text string) string {
	return labelStyle.Render(text)
}

// Value styles a primary value. Pass an optional color override.
func Value(text string, colors ...lipgloss.Color) string {
	if len(colors) > 0 && colors[0] != "" {
		return valueStyle.Foreground(colors[0]).Render(text)
	}
	return valueStyle.Render(text)
}

// Dim styles secondary metadata (paths, IDs, timestamps).
func Dim(text string) string {
	return dimStyle.Render(text)
}

// ColoredText renders text in a specific color without other styling.
func ColoredText(text string, color lipgloss.Color) string {
	return lipgloss.NewStyle().Foreground(color).Render(text)
}

// Success styles a success message fragment.
func Success(text string) string {
	return successStyle.Render(text)
}

// Warning styles a warning message fragment.
func Warning(text string) string {
	return warningStyle.Render(text)
}

// Error styles an error message fragment.
func Error(text string) string {
	return errorStyle.Render(text)
}

// Info styles informational message fragments without strong emphasis.
func Info(text string) string {
	return infoStyle.Render(text)
}

// StateIcon returns a colored symbol representing a progress state.
func StateIcon(state string) string {
	switch normalizeState(state) {
	case "completed", "complete", "done":
		return ColoredText("✓", SuccessColor)
	case "in_progress", "inprogress", "active":
		return ColoredText("●", InProgressColor)
	case "blocked", "error", "failed":
		return ColoredText("■", BlockedColor)
	case "pending", "todo", "queued":
		return ColoredText("○", PendingColor)
	default:
		return ColoredText("•", MetadataColor)
	}
}

func normalizeState(state string) string {
	normalized := strings.ToLower(strings.TrimSpace(state))
	normalized = strings.ReplaceAll(normalized, "-", "_")
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

// WriteLabelValue writes a labeled value line to a builder.
func WriteLabelValue(sb *strings.Builder, label, value string) {
	if sb == nil {
		return
	}
	sb.WriteString(Label(label))
	sb.WriteByte(' ')
	sb.WriteString(value)
	sb.WriteByte('\n')
}
