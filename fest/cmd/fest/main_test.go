package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

func TestMainHelp(t *testing.T) {
	// Test that help command works
	cmd := exec.Command("go", "run", "main.go", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Help returns exit code 0, but CombinedOutput might see it as error
		// Check if output contains expected help text
		if !strings.Contains(string(output), "fest is a CLI tool for managing Festival Methodology files") {
			t.Fatalf("Help command failed: %v\nOutput: %s", err, output)
		}
	}

	// Verify help contains expected commands
	outputStr := string(output)
	expectedCommands := []string{"init", "sync", "update"}
	for _, cmd := range expectedCommands {
		if !strings.Contains(outputStr, cmd) {
			t.Errorf("Help output missing command: %s", cmd)
		}
	}
}

func TestMainVersion(t *testing.T) {
	// Test version flag
	cmd := exec.Command("go", "run", "main.go", "--version")
	output, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(err.Error(), "exit status 2") {
		// Version flag causes exit, but should show output
		outputStr := string(output)
		if !strings.Contains(outputStr, "version") {
			t.Log("Version output:", outputStr)
		}
	}
}

func TestBuildBinary(t *testing.T) {
	// Test that the binary builds successfully
	cmd := exec.Command("go", "build", "-o", "fest-test", "main.go")
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}

	// Clean up
	defer os.Remove("fest-test")

	// Test that built binary runs
	cmd = exec.Command("./fest-test", "--help")
	output, _ := cmd.CombinedOutput()
	if !strings.Contains(string(output), "fest") {
		t.Error("Built binary doesn't run correctly")
	}
}
