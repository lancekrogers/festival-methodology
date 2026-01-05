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

// Panel component tests

func TestDefaultPanelOptions(t *testing.T) {
	opts := DefaultPanelOptions()

	if opts.Title != "" {
		t.Errorf("Expected empty title, got %q", opts.Title)
	}

	if opts.TitleColor != ValueColor {
		t.Error("Expected title color to be ValueColor")
	}

	if opts.BorderStyle != BorderRounded {
		t.Errorf("Expected BorderRounded, got %v", opts.BorderStyle)
	}

	if opts.BorderColor != BorderColor {
		t.Error("Expected border color to be BorderColor")
	}

	expectedPadding := [4]int{0, 1, 0, 1}
	if opts.ContentPadding != expectedPadding {
		t.Errorf("Expected padding %v, got %v", expectedPadding, opts.ContentPadding)
	}
}

func TestPanel_WithoutTitle(t *testing.T) {
	content := "Panel content"
	opts := DefaultPanelOptions()

	result := Panel(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Panel should contain content %q", content)
	}

	// Without title, should just be bordered content
	lines := strings.Split(result, "\n")
	if len(lines) < 3 {
		t.Error("Panel should have at least 3 lines (border + content + border)")
	}
}

func TestPanel_WithTitle(t *testing.T) {
	title := "Test Panel"
	content := "Panel content"
	opts := DefaultPanelOptions()
	opts.Title = title

	result := Panel(content, opts)

	if !strings.Contains(result, title) {
		t.Errorf("Panel should contain title %q", title)
	}

	if !strings.Contains(result, content) {
		t.Errorf("Panel should contain content %q", content)
	}

	// With title, should have more lines
	lines := strings.Split(result, "\n")
	if len(lines) < 4 {
		t.Errorf("Titled panel should have at least 4 lines, got %d", len(lines))
	}
}

func TestPanel_WithCustomColors(t *testing.T) {
	content := "Colored panel"
	opts := DefaultPanelOptions()
	opts.Title = "Warning"
	opts.TitleColor = lipgloss.Color("220")
	opts.BorderColor = lipgloss.Color("220")

	result := Panel(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Panel should contain content %q", content)
	}

	if !strings.Contains(result, "Warning") {
		t.Error("Panel should contain title")
	}
}

func TestPanel_WithWidth(t *testing.T) {
	content := "Short"
	opts := DefaultPanelOptions()
	opts.Width = 40

	result := Panel(content, opts)

	if !strings.Contains(result, content) {
		t.Errorf("Panel should contain content %q", content)
	}

	// Panel should respect width setting
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Panel should have multiple lines")
	}
}

func TestPanel_MultipleBorderStyles(t *testing.T) {
	content := "Content"

	tests := []struct {
		name  string
		style BorderStyle
	}{
		{"Rounded", BorderRounded},
		{"Square", BorderSquare},
		{"Minimal", BorderMinimal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := DefaultPanelOptions()
			opts.BorderStyle = tt.style

			result := Panel(content, opts)

			if !strings.Contains(result, content) {
				t.Errorf("Panel with %s border should contain content", tt.name)
			}
		})
	}
}

func TestTitledPanel_Convenience(t *testing.T) {
	title := "Title"
	content := "Content"

	result := TitledPanel(title, content)

	if !strings.Contains(result, title) {
		t.Errorf("TitledPanel should contain title %q", title)
	}

	if !strings.Contains(result, content) {
		t.Errorf("TitledPanel should contain content %q", content)
	}

	// Should match Panel with title set
	opts := DefaultPanelOptions()
	opts.Title = title
	expected := Panel(content, opts)
	if result != expected {
		t.Error("TitledPanel should match Panel with title set")
	}
}

func TestInfoPanel(t *testing.T) {
	title := "Info"
	content := "Information message"

	result := InfoPanel(title, content)

	if !strings.Contains(result, title) || !strings.Contains(result, content) {
		t.Error("InfoPanel should contain title and content")
	}

	// Info panel uses success colors
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = SuccessColor
	opts.BorderColor = SuccessColor
	expected := Panel(content, opts)
	if result != expected {
		t.Error("InfoPanel should use success colors")
	}
}

func TestWarningPanel(t *testing.T) {
	title := "Warning"
	content := "Warning message"

	result := WarningPanel(title, content)

	if !strings.Contains(result, title) || !strings.Contains(result, content) {
		t.Error("WarningPanel should contain title and content")
	}

	// Warning panel uses warning colors
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = WarningColor
	opts.BorderColor = WarningColor
	expected := Panel(content, opts)
	if result != expected {
		t.Error("WarningPanel should use warning colors")
	}
}

func TestErrorPanel(t *testing.T) {
	title := "Error"
	content := "Error message"

	result := ErrorPanel(title, content)

	if !strings.Contains(result, title) || !strings.Contains(result, content) {
		t.Error("ErrorPanel should contain title and content")
	}

	// Error panel uses error colors
	opts := DefaultPanelOptions()
	opts.Title = title
	opts.TitleColor = ErrorColor
	opts.BorderColor = ErrorColor
	expected := Panel(content, opts)
	if result != expected {
		t.Error("ErrorPanel should use error colors")
	}
}

func TestPanel_MultilineContent(t *testing.T) {
	title := "Multi-line"
	content := "Line 1\nLine 2\nLine 3"
	opts := DefaultPanelOptions()
	opts.Title = title

	result := Panel(content, opts)

	if !strings.Contains(result, "Line 1") || !strings.Contains(result, "Line 2") || !strings.Contains(result, "Line 3") {
		t.Error("Panel should preserve multiline content")
	}

	if !strings.Contains(result, title) {
		t.Error("Panel should contain title")
	}
}
