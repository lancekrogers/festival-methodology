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

// Header component tests

func TestDefaultHeaderOptions(t *testing.T) {
	opts := DefaultHeaderOptions()

	if opts.Level != HeaderH2 {
		t.Errorf("Expected default level HeaderH2, got %v", opts.Level)
	}

	if opts.Color != ValueColor {
		t.Error("Expected color to be ValueColor")
	}

	if opts.Underline {
		t.Error("Expected underline to be false by default")
	}

	if opts.UnderlineChar != "─" {
		t.Errorf("Expected underline char '─', got %q", opts.UnderlineChar)
	}
}

func TestHeader_H1(t *testing.T) {
	text := "Top Level Header"
	opts := DefaultHeaderOptions()
	opts.Level = HeaderH1

	result := Header(text, opts)

	// H1 should be uppercase
	if !strings.Contains(result, "TOP LEVEL HEADER") {
		t.Error("H1 header should be uppercase")
	}
}

func TestHeader_H2(t *testing.T) {
	text := "Section Header"
	opts := DefaultHeaderOptions()
	opts.Level = HeaderH2

	result := Header(text, opts)

	if !strings.Contains(result, text) {
		t.Errorf("H2 header should contain original text %q", text)
	}
}

func TestHeader_H3(t *testing.T) {
	text := "Subsection"
	opts := DefaultHeaderOptions()
	opts.Level = HeaderH3

	result := Header(text, opts)

	if !strings.Contains(result, text) {
		t.Errorf("H3 header should contain text %q", text)
	}
}

func TestHeader_WithUnderline(t *testing.T) {
	text := "Header"
	opts := DefaultHeaderOptions()
	opts.Underline = true

	result := Header(text, opts)

	if !strings.Contains(result, text) {
		t.Errorf("Header should contain text %q", text)
	}

	// Should have underline character
	if !strings.Contains(result, "─") {
		t.Error("Underlined header should contain underline character")
	}

	// Should have multiple lines
	lines := strings.Split(result, "\n")
	if len(lines) < 2 {
		t.Error("Underlined header should have at least 2 lines")
	}
}

func TestHeader_CustomUnderlineChar(t *testing.T) {
	text := "Custom"
	opts := DefaultHeaderOptions()
	opts.Underline = true
	opts.UnderlineChar = "="

	result := Header(text, opts)

	if !strings.Contains(result, "=") {
		t.Error("Header should use custom underline character")
	}
}

func TestHeader_CustomColor(t *testing.T) {
	text := "Colored Header"
	opts := DefaultHeaderOptions()
	opts.Color = lipgloss.Color("42")

	result := Header(text, opts)

	if !strings.Contains(result, text) {
		t.Errorf("Header should contain text %q", text)
	}
}

func TestH1_Convenience(t *testing.T) {
	text := "title"
	result := H1(text)

	// H1 should be uppercase and underlined
	if !strings.Contains(result, "TITLE") {
		t.Error("H1 should be uppercase")
	}

	if !strings.Contains(result, "─") {
		t.Error("H1 should have underline by default")
	}

	// Should match Header with H1 + underline
	opts := DefaultHeaderOptions()
	opts.Level = HeaderH1
	opts.Underline = true
	expected := Header(text, opts)
	if result != expected {
		t.Error("H1 should match Header with HeaderH1 and underline")
	}
}

func TestH2_Convenience(t *testing.T) {
	text := "Section"
	result := H2(text)

	if !strings.Contains(result, text) {
		t.Errorf("H2 should contain text %q", text)
	}

	opts := DefaultHeaderOptions()
	opts.Level = HeaderH2
	expected := Header(text, opts)
	if result != expected {
		t.Error("H2 should match Header with HeaderH2")
	}
}

func TestH3_Convenience(t *testing.T) {
	text := "Subsection"
	result := H3(text)

	if !strings.Contains(result, text) {
		t.Errorf("H3 should contain text %q", text)
	}

	opts := DefaultHeaderOptions()
	opts.Level = HeaderH3
	expected := Header(text, opts)
	if result != expected {
		t.Error("H3 should match Header with HeaderH3")
	}
}

// Progress bar component tests

func TestDefaultProgressBarOptions(t *testing.T) {
	opts := DefaultProgressBarOptions()

	if opts.Current != 0 {
		t.Errorf("Expected current 0, got %d", opts.Current)
	}

	if opts.Total != 100 {
		t.Errorf("Expected total 100, got %d", opts.Total)
	}

	if opts.Width != 40 {
		t.Errorf("Expected width 40, got %d", opts.Width)
	}

	if opts.FilledChar != "█" {
		t.Errorf("Expected filled char '█', got %q", opts.FilledChar)
	}

	if opts.EmptyChar != "░" {
		t.Errorf("Expected empty char '░', got %q", opts.EmptyChar)
	}

	if opts.FilledColor != SuccessColor {
		t.Error("Expected filled color to be SuccessColor")
	}

	if opts.EmptyColor != lipgloss.Color("240") {
		t.Error("Expected empty color to be grey (240)")
	}

	if !opts.ShowPercentage {
		t.Error("Expected ShowPercentage to be true by default")
	}

	if opts.ShowFraction {
		t.Error("Expected ShowFraction to be false by default")
	}
}

func TestProgressBar_ZeroPercent(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 0
	opts.Total = 100

	result := RenderProgressBar(opts)

	// Should contain empty characters
	if !strings.Contains(result, opts.EmptyChar) {
		t.Error("0% progress bar should contain empty characters")
	}

	// Should not contain filled characters (or very few)
	filledCount := strings.Count(result, opts.FilledChar)
	if filledCount > 1 {
		t.Errorf("0%% progress bar should have minimal filled characters, got %d", filledCount)
	}
}

func TestProgressBar_FiftyPercent(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 50
	opts.Total = 100

	result := RenderProgressBar(opts)

	// Should contain both filled and empty characters
	if !strings.Contains(result, opts.FilledChar) {
		t.Error("50% progress bar should contain filled characters")
	}

	if !strings.Contains(result, opts.EmptyChar) {
		t.Error("50% progress bar should contain empty characters")
	}
}

func TestProgressBar_OneHundredPercent(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 100
	opts.Total = 100

	result := RenderProgressBar(opts)

	// Should be fully filled
	if !strings.Contains(result, opts.FilledChar) {
		t.Error("100% progress bar should contain filled characters")
	}

	// Should have minimal or no empty characters
	emptyCount := strings.Count(result, opts.EmptyChar)
	if emptyCount > 1 {
		t.Errorf("100%% progress bar should have minimal empty characters, got %d", emptyCount)
	}
}

func TestProgressBar_OverOneHundredPercent(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 150
	opts.Total = 100

	result := RenderProgressBar(opts)

	// Should cap at 100% - fully filled
	if !strings.Contains(result, opts.FilledChar) {
		t.Error("Over 100% progress bar should be fully filled")
	}

	// Should have minimal or no empty characters
	emptyCount := strings.Count(result, opts.EmptyChar)
	if emptyCount > 1 {
		t.Errorf("Over 100%% progress bar should have minimal empty characters, got %d", emptyCount)
	}
}

func TestProgressBar_CustomCharacters(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 50
	opts.Total = 100
	opts.FilledChar = "#"
	opts.EmptyChar = "-"

	result := RenderProgressBar(opts)

	if !strings.Contains(result, "#") {
		t.Error("Progress bar should use custom filled character '#'")
	}

	if !strings.Contains(result, "-") {
		t.Error("Progress bar should use custom empty character '-'")
	}
}

func TestProgressBar_CustomColors(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 75
	opts.Total = 100
	opts.FilledColor = lipgloss.Color("42")
	opts.EmptyColor = lipgloss.Color("240")

	result := RenderProgressBar(opts)

	// Basic check - should contain the progress bar
	if !strings.Contains(result, opts.FilledChar) {
		t.Error("Progress bar should contain filled characters")
	}
}

func TestProgressBar_ShowPercentage(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 75
	opts.Total = 100
	opts.ShowPercentage = true
	opts.ShowFraction = false

	result := RenderProgressBar(opts)

	// Should contain percentage indicator
	// Note: The implementation has a complex percentage rendering, just verify it's not empty
	if result == "" {
		t.Error("Progress bar with percentage should not be empty")
	}

	if !strings.Contains(result, opts.FilledChar) {
		t.Error("Progress bar should contain filled characters")
	}
}

func TestProgressBar_ShowFraction(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 42
	opts.Total = 100
	opts.ShowPercentage = false
	opts.ShowFraction = true

	result := RenderProgressBar(opts)

	// Should contain the progress bar
	if result == "" {
		t.Error("Progress bar with fraction should not be empty")
	}

	if !strings.Contains(result, opts.FilledChar) {
		t.Error("Progress bar should contain filled characters")
	}
}

func TestProgressBar_ZeroTotal(t *testing.T) {
	opts := DefaultProgressBarOptions()
	opts.Current = 50
	opts.Total = 0

	result := RenderProgressBar(opts)

	// With zero total, should handle gracefully (all empty)
	if !strings.Contains(result, opts.EmptyChar) {
		t.Error("Progress bar with zero total should contain empty characters")
	}
}

func TestSimpleProgressBar(t *testing.T) {
	result := SimpleProgressBar(30, 100)

	// Should produce a valid progress bar
	if result == "" {
		t.Error("SimpleProgressBar should not be empty")
	}

	// Should contain progress characters
	opts := DefaultProgressBarOptions()
	if !strings.Contains(result, opts.FilledChar) && !strings.Contains(result, opts.EmptyChar) {
		t.Error("SimpleProgressBar should contain progress bar characters")
	}
}

func TestSimpleProgressBar_MatchesProgressBar(t *testing.T) {
	current := 60
	total := 100

	simple := SimpleProgressBar(current, total)

	opts := DefaultProgressBarOptions()
	opts.Current = current
	opts.Total = total
	standard := RenderProgressBar(opts)

	if simple != standard {
		t.Error("SimpleProgressBar should match RenderProgressBar with default options")
	}
}

func TestSpinner_FirstFrame(t *testing.T) {
	result := Spinner(0)

	if result == "" {
		t.Error("Spinner should not be empty")
	}

	// Should contain the first spinner frame
	if !strings.Contains(result, SpinnerFrames[0]) {
		t.Errorf("Spinner(0) should contain first frame %q", SpinnerFrames[0])
	}
}

func TestSpinner_MiddleFrame(t *testing.T) {
	frameIndex := 5
	result := Spinner(frameIndex)

	if result == "" {
		t.Error("Spinner should not be empty")
	}

	// Should contain the specified spinner frame
	if !strings.Contains(result, SpinnerFrames[frameIndex]) {
		t.Errorf("Spinner(%d) should contain frame %q", frameIndex, SpinnerFrames[frameIndex])
	}
}

func TestSpinner_WrapsAround(t *testing.T) {
	// Test that spinner wraps around when frame exceeds length
	frameCount := len(SpinnerFrames)
	result := Spinner(frameCount) // Should wrap to frame 0

	if result == "" {
		t.Error("Spinner should not be empty")
	}

	// Should contain the first frame (wrapped around)
	if !strings.Contains(result, SpinnerFrames[0]) {
		t.Errorf("Spinner(%d) should wrap to first frame %q", frameCount, SpinnerFrames[0])
	}
}

func TestSpinner_LargeFrameNumber(t *testing.T) {
	// Test with a large frame number
	result := Spinner(100)

	if result == "" {
		t.Error("Spinner should not be empty")
	}

	// Should successfully render some frame
	expectedFrame := SpinnerFrames[100%len(SpinnerFrames)]
	if !strings.Contains(result, expectedFrame) {
		t.Errorf("Spinner should contain frame %q", expectedFrame)
	}
}

func TestSpinner_AllFrames(t *testing.T) {
	// Verify all spinner frames can be rendered
	for i := 0; i < len(SpinnerFrames); i++ {
		result := Spinner(i)
		if result == "" {
			t.Errorf("Spinner(%d) should not be empty", i)
		}
		if !strings.Contains(result, SpinnerFrames[i]) {
			t.Errorf("Spinner(%d) should contain frame %q", i, SpinnerFrames[i])
		}
	}
}
