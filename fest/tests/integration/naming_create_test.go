//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

type createFestivalOutput struct {
	OK       bool `json:"ok"`
	Festival struct {
		Directory string `json:"directory"`
		Dest      string `json:"dest"`
	} `json:"festival"`
}

type createSequenceOutput struct {
	OK      bool     `json:"ok"`
	Created []string `json:"created"`
}

type createTaskOutput struct {
	OK      bool     `json:"ok"`
	Created []string `json:"created"`
}

func setupFestivalTemplates(t *testing.T, container *TestContainer) {
	t.Helper()

	tmpDir := t.TempDir()
	templateRoot := filepath.Join(tmpDir, ".festival", "templates")

	templates := map[string]string{
		"FESTIVAL_OVERVIEW_TEMPLATE.md": `---
template_id: FESTIVAL_OVERVIEW
required_variables:
  - festival_name
  - festival_goal
---
# {{.festival_name}}

Goal: {{.festival_goal}}
`,
		"FESTIVAL_GOAL_TEMPLATE.md": `---
template_id: FESTIVAL_GOAL
required_variables:
  - festival_name
  - festival_goal
---
# {{.festival_name}}

Goal: {{.festival_goal}}
`,
		"FESTIVAL_RULES_TEMPLATE.md": `# Festival Rules

- Follow naming conventions
`,
		"FESTIVAL_TODO_TEMPLATE.md": `# TODO

- [ ] Initialize festival
`,
		"PHASE_GOAL_TEMPLATE.md": `---
template_id: PHASE_GOAL
required_variables:
  - phase_id
---
# Phase {{.phase_id}}
`,
		"SEQUENCE_GOAL_TEMPLATE.md": `---
template_id: SEQUENCE_GOAL
required_variables:
  - sequence_id
---
# Sequence {{.sequence_id}}
`,
		"TASK_TEMPLATE.md": `---
template_id: TASK
required_variables:
  - task_id
---
# Task {{.task_id}}
`,
		"phases/implementation/gates/QUALITY_GATE_TESTING.md": "# Testing Gate\n",
		"phases/implementation/gates/QUALITY_GATE_REVIEW.md":  "# Review Gate\n",
		"phases/implementation/gates/QUALITY_GATE_ITERATE.md": "# Iterate Gate\n",
		"phases/implementation/gates/QUALITY_GATE_COMMIT.md":  "# Commit Gate\n",
	}

	for path, content := range templates {
		fullPath := filepath.Join(templateRoot, path)
		require.NoError(t, os.MkdirAll(filepath.Dir(fullPath), 0755))
		require.NoError(t, os.WriteFile(fullPath, []byte(content), 0644))
	}

	require.NoError(t, container.CopyDirToContainer(filepath.Join(tmpDir, ".festival"), "/festivals/.festival"))
}

func parseCreateFestivalOutput(t *testing.T, output string) createFestivalOutput {
	t.Helper()

	var res createFestivalOutput
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &res)
	require.NoError(t, err, "create festival output should be valid JSON")
	require.True(t, res.OK, "create festival should succeed")
	require.NotEmpty(t, res.Festival.Directory, "festival directory should be present")
	require.NotEmpty(t, res.Festival.Dest, "festival dest should be present")

	return res
}

func parseCreateSequenceOutput(t *testing.T, output string) createSequenceOutput {
	t.Helper()

	var res createSequenceOutput
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &res)
	require.NoError(t, err, "create sequence output should be valid JSON")
	require.True(t, res.OK, "create sequence should succeed")
	require.NotEmpty(t, res.Created, "sequence should report created paths")

	return res
}

func parseCreateTaskOutput(t *testing.T, output string) createTaskOutput {
	t.Helper()

	var res createTaskOutput
	err := json.Unmarshal([]byte(strings.TrimSpace(output)), &res)
	require.NoError(t, err, "create task output should be valid JSON")
	require.True(t, res.OK, "create task should succeed")
	require.NotEmpty(t, res.Created, "task should report created paths")

	return res
}

func TestFestCreateNamingNormalizationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := GetSharedContainer(t)
	setupFestivalTemplates(t, container)

	output, err := container.RunFestInDir(
		"/festivals",
		"create", "festival",
		"--name", "naming-fixture",
		"--goal", "Validate-naming-normalization",
		"--json",
		"--skip-markers",
	)
	require.NoError(t, err, "create festival should succeed")

	result := parseCreateFestivalOutput(t, output)
	festivalPath := filepath.Join("/festivals", result.Festival.Dest, result.Festival.Directory)

	exists, err := container.CheckFileExists(filepath.Join(festivalPath, "fest.yaml"))
	require.NoError(t, err)
	require.True(t, exists, "fest.yaml should be created")

	_, err = container.RunFestInDir(
		festivalPath,
		"create", "phase",
		"--name", "001-Plan",
		"--json",
		"--skip-markers",
	)
	require.NoError(t, err, "create phase should succeed")

	phasePath := filepath.Join(festivalPath, "001_PLAN")
	exists, err = container.CheckDirExists(phasePath)
	require.NoError(t, err)
	require.True(t, exists, "phase directory should be normalized")

	doublePhasePath := filepath.Join(festivalPath, "001_001_PLAN")
	exists, err = container.CheckDirExists(doublePhasePath)
	require.NoError(t, err)
	require.False(t, exists, "phase directory should not double-prefix")

	output, err = container.RunFestInDir(
		phasePath,
		"create", "sequence",
		"--name", "01-Requirements",
		"--json",
		"--skip-markers",
	)
	require.NoError(t, err, "create sequence should succeed")

	sequenceResult := parseCreateSequenceOutput(t, output)
	sequenceGoalPath := sequenceResult.Created[0]
	require.Contains(t, sequenceGoalPath, "/01_requirements/", "sequence directory should be normalized")
	require.NotContains(t, sequenceGoalPath, "/01_01_requirements/", "sequence directory should not double-prefix")

	sequencePath := filepath.Dir(sequenceGoalPath)
	exists, err = container.CheckDirExists(sequencePath)
	require.NoError(t, err)
	require.True(t, exists, "sequence directory should exist")

	output, err = container.RunFestInDir(
		sequencePath,
		"create", "task",
		"--name", "01-setup",
		"--json",
		"--skip-markers",
	)
	require.NoError(t, err, "create task should succeed")

	taskResult := parseCreateTaskOutput(t, output)
	taskPath := taskResult.Created[0]
	require.True(t, strings.HasSuffix(taskPath, "/01_setup.md"), "task file should be normalized")
	require.NotContains(t, taskPath, "/01_01_setup.md", "task file should not double-prefix")

	exists, err = container.CheckFileExists(taskPath)
	require.NoError(t, err)
	require.True(t, exists, "task file should be normalized")
}
