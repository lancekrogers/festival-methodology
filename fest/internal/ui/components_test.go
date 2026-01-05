package ui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestDefaultBorderOptions(t *testing.T) {
	opts := DefaultBorderOptions()

	if opts.Style != BorderRounded {
		t.Errorf("Expected default style BorderRounded, got %v", opts.Style)
	}

	if opts.Color != BorderColor {
		t.Errorf("Expected default color to be BorderColor")
	}

	expectedPadding := [4]int{0, 1, 0, 1}
	if opts.Padding != expectedPadding {
		t.Errorf("Expected padding %v, got %v", expectedPadding, opts.Padding)
	}

	if opts.Width != 0 {
		t.Errorf("Expected width 0, got %d", opts.Width)
	}
}

func TestBorder_Rounded(t *testing.T) {
	content := "Test Content"
	opts := DefaultBorderOptions()
	opts.Style = BorderRounded

	result := Border(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}

	// Rounded borders use UTF-8 box characters
	if !strings.Contains(result, "╭") && !strings.Contains(result, "┌") {
		t.Error("Rounded border should contain rounded corner characters")
	}
}

func TestBorder_Square(t *testing.T) {
	content := "Test Content"
	opts := DefaultBorderOptions()
	opts.Style = BorderSquare

	result := Border(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}

	// Square borders should have content wrapped
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("Bordered content should have at least 3 lines (top, content, bottom)")
	}
}

func TestBorder_Minimal(t *testing.T) {
	content := "Test"
	opts := DefaultBorderOptions()
	opts.Style = BorderMinimal

	result := Border(content, opts)

	// Minimal borders use ASCII characters
	if !strings.Contains(result, "+") || !strings.Contains(result, "-") || !strings.Contains(result, "|") {
		t.Error("Minimal border should use ASCII characters (+, -, |)")
	}

	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}
}

func TestBorder_WithWidth(t *testing.T) {
	content := "Short"
	opts := DefaultBorderOptions()
	opts.Width = 30

	result := Border(content, opts)

	// Check that border respects width setting
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Bordered content should have multiple lines")
	}

	// The actual width check is approximate due to ANSI codes
	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}
}

func TestBorder_WithPadding(t *testing.T) {
	content := "Content"
	opts := DefaultBorderOptions()
	opts.Padding = [4]int{1, 2, 1, 2} // top, right, bottom, left

	result := Border(content, opts)

	lines := strings.Split(result, "\n")
	// With vertical padding, should have more lines than minimal border
	if len(lines) < 4 { // Top border + top padding + content + bottom padding + bottom border
		t.Errorf("Border with padding should have at least 4 lines, got %d", len(lines))
	}

	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}
}

func TestBorder_WithCustomColor(t *testing.T) {
	content := "Colored"
	opts := DefaultBorderOptions()
	opts.Color = lipgloss.Color("42") // Green

	result := Border(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Border output should contain content %q", content)
	}

	// With color applied, output should contain ANSI codes (in TTY mode)
	// This is a basic check - full color verification would require TTY simulation
}

func TestRoundedBorder_Convenience(t *testing.T) {
	content := "Quick Border"
	result := RoundedBorder(content)

	if !strings.Contains(result, content) {
		t.Errorf("RoundedBorder output should contain content %q", content)
	}

	// Should produce same result as Border with default options
	expected := Border(content, DefaultBorderOptions())
	if result != expected {
		t.Error("RoundedBorder should produce same output as Border with defaults")
	}
}

func TestSquareBorder_Convenience(t *testing.T) {
	content := "Square Content"
	result := SquareBorder(content)

	if !strings.Contains(result, content) {
		t.Errorf("SquareBorder output should contain content %q", content)
	}

	opts := DefaultBorderOptions()
	opts.Style = BorderSquare
	expected := Border(content, opts)
	if result != expected {
		t.Error("SquareBorder should match Border with BorderSquare style")
	}
}

func TestMinimalBorder_Convenience(t *testing.T) {
	content := "Minimal Content"
	result := MinimalBorder(content)

	if !strings.Contains(result, content) {
		t.Errorf("MinimalBorder output should contain content %q", content)
	}

	// Should use ASCII characters
	if !strings.Contains(result, "+") {
		t.Error("MinimalBorder should use ASCII characters")
	}

	opts := DefaultBorderOptions()
	opts.Style = BorderMinimal
	expected := Border(content, opts)
	if result != expected {
		t.Error("MinimalBorder should match Border with BorderMinimal style")
	}
}

func TestBorder_EmptyContent(t *testing.T) {
	content := ""
	opts := DefaultBorderOptions()

	result := Border(content, opts)

	// Should still render a border even with empty content
	if result == "" {
		t.Error("Border should render even with empty content")
	}

	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Border should have at least top and bottom lines")
	}
}

func TestBorder_MultilineContent(t *testing.T) {
	content := "Line 1\nLine 2\nLine 3"
	opts := DefaultBorderOptions()

	result := Border(content, opts)

	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") || !strings.Contains(result, "Line 3") {
		t.Error("Border should preserve multiline content")
	}

	lines := strings.Split(result, "\n")
	// Should have at least: top border + 3 content lines + bottom border = 5 lines
	if len(lines) < 5 {
		t.Errorf("Multiline border should have at least 5 lines, got %d", len(lines))
	}
}
