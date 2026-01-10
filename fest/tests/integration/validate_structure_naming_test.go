//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type validateStructureOutput struct {
	OK     bool   `json:"ok"`
	Action string `json:"action"`
	Valid  bool   `json:"valid"`
	Issues []struct {
		Code    string `json:"code"`
		Path    string `json:"path"`
		Message string `json:"message"`
	} `json:"issues"`
}

func TestValidateStructureNamingViolations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := GetSharedContainer(t)

	festivalPath := "/festivals/naming-violations"
	phasePath := filepath.Join(festivalPath, "001_plan")
	sequencePath := filepath.Join(phasePath, "01_Design")

	exitCode, _, err := container.container.Exec(container.ctx, []string{"mkdir", "-p", sequencePath})
	require.NoError(t, err)
	require.Equal(t, 0, exitCode)

	writeFile := func(path, content string) {
		exitCode, _, err := container.container.Exec(container.ctx, []string{
			"sh", "-c",
			"printf '%s' '" + content + "' > " + path,
		})
		require.NoError(t, err)
		require.Equal(t, 0, exitCode)
	}

	writeFile(filepath.Join(festivalPath, "FESTIVAL_OVERVIEW.md"), "# Naming Violations\n")
	writeFile(filepath.Join(sequencePath, "SEQUENCE_GOAL.md"), "# Sequence Goal\n")
	writeFile(filepath.Join(sequencePath, "01_Task.md"), "# Task\n")

	output, err := container.RunFest("validate", "structure", festivalPath, "--json")
	require.NoError(t, err, "validate structure should not fail")

	var result validateStructureOutput
	err = json.Unmarshal([]byte(strings.TrimSpace(output)), &result)
	require.NoError(t, err, "validate structure output should be valid JSON")
	require.False(t, result.Valid, "structure validation should fail for naming violations")

	var hasPhase, hasSequence, hasTask bool
	for _, issue := range result.Issues {
		if issue.Code != "naming_convention" {
			continue
		}
		if strings.Contains(issue.Path, "001_plan") {
			hasPhase = true
		}
		if strings.Contains(issue.Path, "01_Design") {
			hasSequence = true
		}
		if strings.Contains(issue.Path, "01_Task.md") {
			hasTask = true
		}
	}

	require.True(t, hasPhase, "phase naming violation should be reported")
	require.True(t, hasSequence, "sequence naming violation should be reported")
	require.True(t, hasTask, "task naming violation should be reported")
}
