package ui

import (
	"io"
	"regexp"
	"testing"

	"github.com/muesli/termenv"
)

var ansiRegexp = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestTTYDetection_DisablesANSIWhenNotTTY(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "")
	t.Setenv("NO_COLOR", "")

	original := termenv.DefaultOutput()
	termenv.SetDefaultOutput(termenv.NewOutput(io.Discard))
	t.Cleanup(func() {
		termenv.SetDefaultOutput(original)
		SetNoColor(false)
	})

	SetNoColor(false)

	output := H1("Test Header")
	if ansiRegexp.MatchString(output) {
		t.Errorf("expected no ANSI codes when not a TTY, got: %q", output)
	}
}

func TestTTYDetection_EnablesANSIWhenTTY(t *testing.T) {
	t.Setenv("CLICOLOR_FORCE", "1")
	t.Setenv("NO_COLOR", "")

	original := termenv.DefaultOutput()
	termenv.SetDefaultOutput(termenv.NewOutput(io.Discard, termenv.WithTTY(true)))
	t.Cleanup(func() {
		termenv.SetDefaultOutput(original)
		SetNoColor(false)
	})

	SetNoColor(false)

	output := H1("Test Header")
	if !ansiRegexp.MatchString(output) {
		t.Errorf("expected ANSI codes when in TTY mode, got: %q", output)
	}
}
