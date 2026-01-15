package ui

import "testing"

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		minutes  int
		expected string
	}{
		{"zero", 0, "0m"},
		{"negative", -5, "0m"},
		{"minutes only", 45, "45m"},
		{"one hour", 60, "1h"},
		{"hours and minutes", 339, "5h 39m"},
		{"one day", 1440, "1d"},
		{"days and hours", 1500, "1d 1h"},
		{"one month", 43200, "1mo"},
		{"month and days", 63249, "1mo 13d"},
		{"large value", 100000, "2mo 9d"},
		{"exact hours", 120, "2h"},
		{"exact days", 2880, "2d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatDuration(tt.minutes)
			if result != tt.expected {
				t.Errorf("FormatDuration(%d) = %q, want %q", tt.minutes, result, tt.expected)
			}
		})
	}
}
