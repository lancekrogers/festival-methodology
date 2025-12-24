package config

import (
	"strings"
	"testing"
)

func TestBashZshInit(t *testing.T) {
	output := bashZshInit()

	// Check for required components
	tests := []struct {
		name     string
		contains string
	}{
		{"function definition", "fgo()"},
		{"local variable", "local dest"},
		{"fest go call", "fest go"},
		{"cd command", "cd \"$dest\""},
		{"directory check", "-d \"$dest\""},
		{"error handling", "return 1"},
		{"comment header", "fest shell integration"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(output, tc.contains) {
				t.Errorf("bash/zsh init output should contain %q", tc.contains)
			}
		})
	}
}

func TestFishInit(t *testing.T) {
	output := fishInit()

	// Check for required components
	tests := []struct {
		name     string
		contains string
	}{
		{"function definition", "function fgo"},
		{"local variable", "set -l dest"},
		{"fest go call", "fest go"},
		{"cd command", "cd $dest"},
		{"directory check", "-d \"$dest\""},
		{"error handling", "return 1"},
		{"comment header", "fest shell integration"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if !strings.Contains(output, tc.contains) {
				t.Errorf("fish init output should contain %q", tc.contains)
			}
		})
	}
}

func TestShellInitValidShells(t *testing.T) {
	// Test that supported shells don't error
	validShells := []string{"zsh", "bash", "fish"}

	for _, shell := range validShells {
		t.Run(shell, func(t *testing.T) {
			cmd := NewShellInitCommand()
			cmd.SetArgs([]string{shell})
			err := cmd.Execute()
			if err != nil {
				t.Errorf("shell-init %s should not error: %v", shell, err)
			}
		})
	}
}

func TestShellInitInvalidShell(t *testing.T) {
	cmd := NewShellInitCommand()
	cmd.SetArgs([]string{"powershell"})

	// Capture the error
	err := cmd.Execute()
	if err == nil {
		t.Error("shell-init with invalid shell should return error")
	}

	if !strings.Contains(err.Error(), "unsupported shell") {
		t.Errorf("error should mention 'unsupported shell', got: %v", err)
	}
}

func TestShellInitNoArgs(t *testing.T) {
	cmd := NewShellInitCommand()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("shell-init with no args should return error")
	}
}

func TestBashZshInitSyntax(t *testing.T) {
	output := bashZshInit()

	// Check for proper bash/zsh syntax patterns
	if !strings.Contains(output, "[[ ") || !strings.Contains(output, " ]]") {
		t.Error("bash/zsh init should use [[ ]] for conditionals")
	}

	if !strings.Contains(output, "$?") {
		t.Error("bash/zsh init should check exit code with $?")
	}
}

func TestFishInitSyntax(t *testing.T) {
	output := fishInit()

	// Check for proper fish syntax patterns
	if !strings.Contains(output, "test ") {
		t.Error("fish init should use 'test' for conditionals")
	}

	if !strings.Contains(output, "$status") {
		t.Error("fish init should check exit code with $status")
	}

	if !strings.Contains(output, "end") {
		t.Error("fish init should use 'end' to close blocks")
	}
}
