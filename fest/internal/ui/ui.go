package ui

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// UI handles user interface operations
type UI struct {
	noColor bool
	verbose bool
	reader  *bufio.Reader
}

// New creates a new UI handler
func New(noColor, verbose bool) *UI {
	SetNoColor(noColor)

	return &UI{
		noColor: noColor,
		verbose: verbose,
		reader:  bufio.NewReader(os.Stdin),
	}
}

// Info prints an info message
func (u *UI) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(infoStyle.Render(message))
}

// Success prints a success message in green
func (u *UI) Success(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(successStyle.Render("✓ " + message))
}

// Warning prints a warning message in yellow
func (u *UI) Warning(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Println(warningStyle.Render("⚠ " + message))
}

// Error prints an error message in red
func (u *UI) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	fmt.Fprintln(os.Stderr, errorStyle.Render("✗ "+message))
}

// Confirm asks for yes/no confirmation
func (u *UI) Confirm(format string, args ...interface{}) bool {
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s [Y/n]: ", message)

	response, err := u.reader.ReadString('\n')
	if err != nil {
		return false
	}

	response = strings.TrimSpace(strings.ToLower(response))
	return response == "" || response == "y" || response == "yes"
}

// Choose presents options and returns the selected index
func (u *UI) Choose(message string, options []string) int {
	fmt.Println(message)
	for i, option := range options {
		fmt.Printf("  [%d] %s\n", i+1, option)
	}
	fmt.Print("Choice: ")

	response, err := u.reader.ReadString('\n')
	if err != nil {
		return 0
	}

	response = strings.TrimSpace(response)
	var choice int
	if _, err := fmt.Sscanf(response, "%d", &choice); err != nil {
		return 0
	}

	if choice < 1 || choice > len(options) {
		return 0
	}

	return choice - 1
}

// Prompt asks the user to input a value
func (u *UI) Prompt(label string) string {
	fmt.Printf("%s: ", label)
	response, err := u.reader.ReadString('\n')
	if err != nil {
		return ""
	}
	return strings.TrimSpace(response)
}

// PromptDefault asks the user to input a value with a default
func (u *UI) PromptDefault(label, def string) string {
	if def != "" {
		fmt.Printf("%s [%s]: ", label, def)
	} else {
		fmt.Printf("%s: ", label)
	}
	response, err := u.reader.ReadString('\n')
	if err != nil {
		return def
	}
	response = strings.TrimSpace(response)
	if response == "" {
		return def
	}
	return response
}

// ShowDiff shows the difference between two files
func (u *UI) ShowDiff(file1, file2 string) error {
	// Simplified diff display
	// In a full implementation, we would use go-diff package
	fmt.Printf("\n--- %s\n+++ %s\n", file1, file2)
	fmt.Println("(Diff display not fully implemented)")
	return nil
}

// ProgressBar represents a progress bar
type ProgressBar struct {
	label   string
	total   int64
	current int64
}

// NewProgressBar creates a new progress bar
func (u *UI) NewProgressBar(label string, total int64) *ProgressBar {
	return &ProgressBar{
		label: label,
		total: total,
	}
}

// Update updates the progress bar
func (p *ProgressBar) Update(current, total int64, file string) {
	p.current = current
	if total > 0 {
		p.total = total
	}

	if p.total > 0 {
		percent := int(float64(p.current) * 100 / float64(p.total))
		fmt.Printf("\r%s: [%3d%%] %d/%d files - %s", p.label, percent, p.current, p.total, file)
	} else {
		fmt.Printf("\r%s: %d files - %s", p.label, p.current, file)
	}
}

// Finish completes the progress bar
func (p *ProgressBar) Finish() {
	fmt.Println() // New line after progress
}

// Verbose returns true if verbose mode is enabled
func (u *UI) Verbose() bool {
	return u.verbose
}
