package ui

import (
	"fmt"
	"strings"
)

// FormatDuration formats minutes as human-readable duration.
// Examples: "45m", "5h 39m", "2d 3h", "1mo 13d"
// Returns the two most significant units for readability.
func FormatDuration(minutes int) string {
	if minutes <= 0 {
		return "0m"
	}

	const (
		minutesPerHour  = 60
		minutesPerDay   = 60 * 24      // 1440
		minutesPerMonth = 60 * 24 * 30 // 43200 (approx 30-day month)
	)

	var parts []string

	if minutes >= minutesPerMonth {
		months := minutes / minutesPerMonth
		minutes %= minutesPerMonth
		parts = append(parts, fmt.Sprintf("%dmo", months))
	}

	if minutes >= minutesPerDay {
		days := minutes / minutesPerDay
		minutes %= minutesPerDay
		parts = append(parts, fmt.Sprintf("%dd", days))
	}

	if minutes >= minutesPerHour {
		hours := minutes / minutesPerHour
		minutes %= minutesPerHour
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
	}

	// Limit to 2 most significant units for readability
	if len(parts) > 2 {
		parts = parts[:2]
	}

	return strings.Join(parts, " ")
}
