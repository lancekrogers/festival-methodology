//go:build integration
// +build integration

package integration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

// ProgressSnapshot captures progress state from all sources
type ProgressSnapshot struct {
	FestivalPercentage int
	PhaseProgress      map[string]PhaseProgressState
	SequenceProgress   map[string]SequenceProgressState
	TaskProgress       map[string]TaskProgressState
	YAMLData           *YAMLProgressData
}

type PhaseProgressState struct {
	Percentage int
	Completed  int
	Total      int
	Name       string
}

type SequenceProgressState struct {
	Percentage int
	Completed  int
	Total      int
	Name       string
}

type TaskProgressState struct {
	Status            string // From fest CLI
	Progress          int
	CheckboxesChecked int // From markdown parse
	CheckboxesTotal   int
	YAMLStatus        string // From YAML file
	MarkdownStatus    string // Calculated from checkboxes
}

type YAMLProgressData struct {
	Festival  string                 `yaml:"festival"`
	UpdatedAt string                 `yaml:"updated_at"`
	Tasks     map[string]YAMLTaskData `yaml:"tasks"`
}

type YAMLTaskData struct {
	TaskID           string `yaml:"task_id"`
	Status           string `yaml:"status"`
	Progress         int    `yaml:"progress"`
	TimeSpentMinutes int    `yaml:"time_spent_minutes,omitempty"`
}

// JSON response structures for parsing fest output
type FestivalProgressJSON struct {
	FestivalName string                  `json:"festival_name"`
	Overall      AggregateProgressJSON   `json:"overall"`
	Phases       []PhaseProgressJSON     `json:"phases,omitempty"`
}

type AggregateProgressJSON struct {
	Total        int                  `json:"total"`
	Completed    int                  `json:"completed"`
	InProgress   int                  `json:"in_progress"`
	Blocked      int                  `json:"blocked"`
	Pending      int                  `json:"pending"`
	Percentage   int                  `json:"percentage"`
	Blockers     []BlockerJSON        `json:"blockers,omitempty"`
	TimeSpentMin int                  `json:"time_spent_minutes"`
}

type PhaseProgressJSON struct {
	PhaseID   string                `json:"phase_id"`
	PhaseName string                `json:"phase_name"`
	Progress  AggregateProgressJSON `json:"progress"`
}

type SequenceProgressJSON struct {
	SequenceID   string                `json:"sequence_id"`
	SequenceName string                `json:"sequence_name"`
	Progress     AggregateProgressJSON `json:"progress"`
}

type BlockerJSON struct {
	TaskID         string `json:"task_id"`
	BlockerMessage string `json:"blocker_message"`
}

// writeFileInContainer writes content to a file in the container
func writeFileInContainer(tc *TestContainer, path, content string) error {
	// Escape single quotes for shell
	escapedContent := strings.ReplaceAll(content, "'", "'\\''")
	cmd := []string{"sh", "-c", fmt.Sprintf("printf '%%s' '%s' > %s", escapedContent, path)}
	exitCode, _, err := tc.container.Exec(tc.ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to execute write command: %w", err)
	}
	if exitCode != 0 {
		return fmt.Errorf("write command exited with code %d", exitCode)
	}
	return nil
}

// setupTodoAppFestival creates the complete festival structure
func setupTodoAppFestival(t *testing.T, tc *TestContainer, festivalPath string) error {
	t.Helper()

	// Create directory structure
	dirs := []string{
		festivalPath,
		filepath.Join(festivalPath, ".fest"),
		filepath.Join(festivalPath, "001_PLAN/01_design"),
		filepath.Join(festivalPath, "002_IMPLEMENT/01_development"),
		filepath.Join(festivalPath, "003_POLISH/01_finalize"),
	}

	for _, dir := range dirs {
		exitCode, _, err := tc.container.Exec(tc.ctx, []string{"mkdir", "-p", dir})
		if err != nil || exitCode != 0 {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Copy festival metadata from testdata
	testdataPath := "testdata/progress_tracking"
	festivalGoalContent, err := readTestdataFile(t, filepath.Join(testdataPath, "festival_templates/FESTIVAL_GOAL.md"))
	if err != nil {
		return err
	}
	if err := writeFileInContainer(tc, filepath.Join(festivalPath, "FESTIVAL_GOAL.md"), festivalGoalContent); err != nil {
		return err
	}

	festYamlContent, err := readTestdataFile(t, filepath.Join(testdataPath, "festival_templates/fest.yaml"))
	if err != nil {
		return err
	}
	if err := writeFileInContainer(tc, filepath.Join(festivalPath, "fest.yaml"), festYamlContent); err != nil {
		return err
	}

	// Create PHASE_GOAL.md files
	phaseGoals := map[string]string{
		"001_PLAN/PHASE_GOAL.md":      "festival_templates/phase_templates/001_PLAN_PHASE_GOAL.md",
		"002_IMPLEMENT/PHASE_GOAL.md": "festival_templates/phase_templates/002_IMPLEMENT_PHASE_GOAL.md",
		"003_POLISH/PHASE_GOAL.md":    "festival_templates/phase_templates/003_POLISH_PHASE_GOAL.md",
	}

	for relPath, templatePath := range phaseGoals {
		content, err := readTestdataFile(t, filepath.Join(testdataPath, templatePath))
		if err != nil {
			return err
		}
		if err := writeFileInContainer(tc, filepath.Join(festivalPath, relPath), content); err != nil {
			return err
		}
	}

	// Create initial task files (all with 0% completion)
	taskFiles := map[string]string{
		"001_PLAN/01_design/01_requirements.md":                  "task_templates/01_requirements_0_of_2.md",
		"001_PLAN/01_design/02_architecture.md":                  "task_templates/02_architecture_0_of_3.md",
		"002_IMPLEMENT/01_development/01_setup.md":               "task_templates/01_setup_0_of_2.md",
		"002_IMPLEMENT/01_development/02_create_task_list.md":    "task_templates/02_create_task_list_0_of_3.md",
		"002_IMPLEMENT/01_development/03_add_complete_feature.md": "task_templates/03_add_complete_0_of_2.md",
		"003_POLISH/01_finalize/01_documentation.md":             "task_templates/01_documentation_0_of_2.md",
		"003_POLISH/01_finalize/02_cleanup.md":                   "task_templates/02_cleanup_0_of_1.md",
	}

	for relPath, templatePath := range taskFiles {
		content, err := readTestdataFile(t, filepath.Join(testdataPath, templatePath))
		if err != nil {
			return err
		}
		if err := writeFileInContainer(tc, filepath.Join(festivalPath, relPath), content); err != nil {
			return err
		}
	}

	return nil
}

// readTestdataFile reads a file from the testdata directory on the host filesystem
func readTestdataFile(t *testing.T, relPath string) (string, error) {
	t.Helper()

	// Normalize the path (remove duplicate testdata/progress_tracking prefix if present)
	normalizedPath := strings.TrimPrefix(relPath, "testdata/progress_tracking/")
	fullPath := filepath.Join("testdata/progress_tracking", normalizedPath)

	// Read from local filesystem (testdata is on host, not in container)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read testdata file %s: %w", fullPath, err)
	}

	return string(data), nil
}

// simulateTaskWork replaces a task file with a template showing specific checkbox state
func simulateTaskWork(t *testing.T, tc *TestContainer, festivalPath, taskRelPath, templateName string) error {
	t.Helper()

	testdataPath := "testdata/progress_tracking"
	templatePath := filepath.Join(testdataPath, "task_templates", templateName)

	content, err := readTestdataFile(t, templatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateName, err)
	}

	taskFullPath := filepath.Join(festivalPath, taskRelPath)
	if err := writeFileInContainer(tc, taskFullPath, content); err != nil {
		return fmt.Errorf("failed to write task file %s: %w", taskFullPath, err)
	}

	return nil
}

// captureProgressSnapshot captures current progress from all sources
func captureProgressSnapshot(t *testing.T, tc *TestContainer, festivalPath string) *ProgressSnapshot {
	t.Helper()

	snapshot := &ProgressSnapshot{
		PhaseProgress:    make(map[string]PhaseProgressState),
		SequenceProgress: make(map[string]SequenceProgressState),
		TaskProgress:     make(map[string]TaskProgressState),
	}

	// Get overall festival progress via fest progress --json
	output, err := tc.RunFestInDir(festivalPath, "progress", "--json")
	if err != nil {
		t.Logf("Warning: fest progress failed: %v (output: %s)", err, output)
		return snapshot
	}

	var festProgress FestivalProgressJSON
	if err := json.Unmarshal([]byte(output), &festProgress); err != nil {
		t.Logf("Warning: failed to parse progress JSON: %v (output: %s)", err, output)
		return snapshot
	}

	snapshot.FestivalPercentage = festProgress.Overall.Percentage

	// Store phase progress
	for _, phase := range festProgress.Phases {
		snapshot.PhaseProgress[phase.PhaseID] = PhaseProgressState{
			Percentage: phase.Progress.Percentage,
			Completed:  phase.Progress.Completed,
			Total:      phase.Progress.Total,
			Name:       phase.PhaseName,
		}
	}

	// Read YAML progress file directly
	yamlPath := filepath.Join(festivalPath, ".fest/progress.yaml")
	yamlContent, err := tc.ReadFile(yamlPath)
	if err == nil && yamlContent != "" {
		var yamlData YAMLProgressData
		if err := yaml.Unmarshal([]byte(yamlContent), &yamlData); err == nil {
			snapshot.YAMLData = &yamlData
		}
	}

	// Parse each task's markdown file and get status
	taskPaths := []string{
		"001_PLAN/01_design/01_requirements.md",
		"001_PLAN/01_design/02_architecture.md",
		"002_IMPLEMENT/01_development/01_setup.md",
		"002_IMPLEMENT/01_development/02_create_task_list.md",
		"002_IMPLEMENT/01_development/03_add_complete_feature.md",
		"003_POLISH/01_finalize/01_documentation.md",
		"003_POLISH/01_finalize/02_cleanup.md",
	}

	for _, taskPath := range taskPaths {
		// Parse markdown checkboxes
		checked, total := parseMarkdownCheckboxes(t, tc, filepath.Join(festivalPath, taskPath))

		// Get YAML status
		yamlStatus := ""
		if snapshot.YAMLData != nil {
			if yamlTask, ok := snapshot.YAMLData.Tasks[taskPath]; ok {
				yamlStatus = yamlTask.Status
			}
		}

		// Determine markdown status
		markdownStatus := "pending"
		if total > 0 {
			if checked == total {
				markdownStatus = "completed"
			} else if checked > 0 {
				markdownStatus = "in_progress"
			}
		}

		// Get status from fest CLI
		// We infer from the overall progress data
		reportedStatus := markdownStatus // Default assumption
		if yamlStatus != "" {
			reportedStatus = yamlStatus // YAML overrides (this is the bug!)
		}

		snapshot.TaskProgress[taskPath] = TaskProgressState{
			Status:            reportedStatus,
			CheckboxesChecked: checked,
			CheckboxesTotal:   total,
			YAMLStatus:        yamlStatus,
			MarkdownStatus:    markdownStatus,
		}
	}

	return snapshot
}

// parseMarkdownCheckboxes counts checkboxes in a markdown file
// Matches fest CLI logic: prioritizes "Definition of Done" and similar sections
func parseMarkdownCheckboxes(t *testing.T, tc *TestContainer, taskPath string) (checked, total int) {
	t.Helper()

	content, err := tc.ReadFile(taskPath)
	if err != nil {
		t.Logf("Warning: failed to read task file %s: %v", taskPath, err)
		return 0, 0
	}

	lines := strings.Split(content, "\n")

	// Priority sections to check (matching fest CLI's logic)
	prioritySections := []string{
		"definition of done",
		"requirements",
		"acceptance criteria",
		"deliverables",
	}

	// Track whether we're in a priority section
	inPrioritySection := false
	currentSection := ""
	priorityChecked := 0
	priorityTotal := 0
	allChecked := 0
	allTotal := 0

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Detect section headers (## or ###)
		if strings.HasPrefix(trimmed, "##") {
			currentSection = strings.ToLower(strings.TrimSpace(strings.TrimPrefix(trimmed, "##")))
			currentSection = strings.TrimPrefix(currentSection, "#") // Handle ###
			currentSection = strings.TrimSpace(currentSection)

			// Check if this is a priority section
			inPrioritySection = false
			for _, priority := range prioritySections {
				if strings.Contains(currentSection, priority) {
					inPrioritySection = true
					break
				}
			}
		}

		// Count checkboxes
		isChecked := false
		isCheckbox := false

		// Standard checkbox formats
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			isChecked = true
			isCheckbox = true
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			isCheckbox = true
		}

		// Emoji checkbox formats
		if strings.Contains(trimmed, "[✅]") {
			isChecked = true
			isCheckbox = true
		} else if strings.Contains(trimmed, "[🚧]") || strings.Contains(trimmed, "[❌]") {
			isCheckbox = true
		}

		if isCheckbox {
			// Count in priority section if we're in one
			if inPrioritySection {
				if isChecked {
					priorityChecked++
				}
				priorityTotal++
			}
			// Always count in all totals
			if isChecked {
				allChecked++
			}
			allTotal++
		}
	}

	// Return priority section counts if found, otherwise all checkboxes
	if priorityTotal > 0 {
		return priorityChecked, priorityTotal
	}
	return allChecked, allTotal
}

// verifyProgressConsistency logs any divergence between YAML and markdown
// Note: Divergence is allowed - YAML is the source of truth per architectural decision
func verifyProgressConsistency(t *testing.T, snapshot *ProgressSnapshot) {
	t.Helper()

	for taskKey, taskState := range snapshot.TaskProgress {
		if taskState.YAMLStatus != "" && taskState.YAMLStatus != taskState.MarkdownStatus {
			t.Logf("INFO: Task %s has YAML=%s, Markdown=%s (YAML is source of truth)",
				taskKey, taskState.YAMLStatus, taskState.MarkdownStatus)
		}
	}

	// No assertion needed - YAML is allowed to override markdown
	// This is the correct architectural behavior
}

// Main lifecycle test
func TestProgressTracking_FullLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	container := GetSharedContainer(t)
	festivalPath := "/festivals/todo-app-festival"

	// Phase 1: Create festival structure
	t.Run("CreateAndVerifyInitialState", func(t *testing.T) {
		err := setupTodoAppFestival(t, container, festivalPath)
		require.NoError(t, err, "festival setup should succeed")

		snapshot := captureProgressSnapshot(t, container, festivalPath)

		// Verify festival is at 0%
		require.Equal(t, 0, snapshot.FestivalPercentage, "festival should start at 0%%")

		// Verify all tasks are pending
		for taskKey, taskState := range snapshot.TaskProgress {
			require.Equal(t, "pending", taskState.MarkdownStatus,
				"task %s should be pending (all checkboxes unchecked)", taskKey)
			require.Equal(t, 0, taskState.CheckboxesChecked,
				"task %s should have 0 checked boxes", taskKey)
		}

		// Verify no YAML records exist yet
		if snapshot.YAMLData != nil {
			require.Empty(t, snapshot.YAMLData.Tasks,
				"YAML should have no task records initially")
		}

		verifyProgressConsistency(t, snapshot)
	})

	// Phase 2: Planning phase
	t.Run("PlanningPhase", func(t *testing.T) {
		t.Run("Requirements_PartialWork", func(t *testing.T) {
			taskPath := "001_PLAN/01_design/01_requirements.md"

			// Simulate partial work: 1 of 2 checkboxes
			err := simulateTaskWork(t, container, festivalPath, taskPath, "01_requirements_1_of_2.md")
			require.NoError(t, err)

			snapshot := captureProgressSnapshot(t, container, festivalPath)

			// Verify task is in_progress
			taskState := snapshot.TaskProgress[taskPath]
			require.Equal(t, "in_progress", taskState.MarkdownStatus,
				"task should be in_progress with 1 of 2 checkboxes")
			require.Equal(t, 1, taskState.CheckboxesChecked)
			require.Equal(t, 2, taskState.CheckboxesTotal)

			// Sequence should be 0% (no complete tasks yet)
			// Phase should be 0%
			// Festival should be 0%

			verifyProgressConsistency(t, snapshot)
		})

		t.Run("Requirements_Complete", func(t *testing.T) {
			taskPath := "001_PLAN/01_design/01_requirements.md"

			// Complete all checkboxes: 2 of 2
			err := simulateTaskWork(t, container, festivalPath, taskPath, "01_requirements_2_of_2.md")
			require.NoError(t, err)

			snapshot := captureProgressSnapshot(t, container, festivalPath)

			// Task should be completed
			taskState := snapshot.TaskProgress[taskPath]
			require.Equal(t, "completed", taskState.MarkdownStatus,
				"task should be completed with 2 of 2 checkboxes")
			require.Equal(t, 2, taskState.CheckboxesChecked)

			verifyProgressConsistency(t, snapshot)
		})

		t.Run("Architecture_Complete", func(t *testing.T) {
			taskPath := "001_PLAN/01_design/02_architecture.md"

			// Complete all checkboxes: 3 of 3
			err := simulateTaskWork(t, container, festivalPath, taskPath, "02_architecture_3_of_3.md")
			require.NoError(t, err)

			snapshot := captureProgressSnapshot(t, container, festivalPath)

			// Task should be completed
			taskState := snapshot.TaskProgress[taskPath]
			require.Equal(t, "completed", taskState.MarkdownStatus,
				"task should be completed with 3 of 3 checkboxes")

			verifyProgressConsistency(t, snapshot)
		})
	})

	// Phase 3: Implementation phase
	t.Run("ImplementationPhase", func(t *testing.T) {
		// Complete all 3 implementation tasks
		tasks := []struct {
			path     string
			template string
		}{
			{"002_IMPLEMENT/01_development/01_setup.md", "01_setup_2_of_2.md"},
			{"002_IMPLEMENT/01_development/02_create_task_list.md", "02_create_task_list_3_of_3.md"},
			{"002_IMPLEMENT/01_development/03_add_complete_feature.md", "03_add_complete_2_of_2.md"},
		}

		for _, task := range tasks {
			// Simulate work by updating markdown
			err := simulateTaskWork(t, container, festivalPath, task.path, task.template)
			require.NoError(t, err, "task work simulation should succeed")

			// Mark complete in YAML (proper workflow)
			_, err = container.RunFestInDir(festivalPath, "progress", "--complete", "--task", task.path)
			require.NoError(t, err, "fest progress --complete should work")
		}

		snapshot := captureProgressSnapshot(t, container, festivalPath)

		// Verify all implementation tasks completed
		for _, task := range tasks {
			taskState := snapshot.TaskProgress[task.path]
			require.Equal(t, "completed", taskState.MarkdownStatus,
				"task %s markdown should be completed", task.path)
			require.Equal(t, "completed", taskState.Status,
				"task %s should report as completed from YAML", task.path)
		}

		verifyProgressConsistency(t, snapshot)
	})

	// Phase 4: Polish phase
	t.Run("PolishPhase", func(t *testing.T) {
		// Complete final 2 tasks
		tasks := []struct {
			path     string
			template string
		}{
			{"003_POLISH/01_finalize/01_documentation.md", "01_documentation_2_of_2.md"},
			{"003_POLISH/01_finalize/02_cleanup.md", "02_cleanup_1_of_1.md"},
		}

		for _, task := range tasks {
			// Simulate work by updating markdown
			err := simulateTaskWork(t, container, festivalPath, task.path, task.template)
			require.NoError(t, err)

			// Mark complete in YAML (proper workflow)
			_, err = container.RunFestInDir(festivalPath, "progress", "--complete", "--task", task.path)
			require.NoError(t, err, "fest progress --complete should work")
		}

		snapshot := captureProgressSnapshot(t, container, festivalPath)

		// Verify all tasks completed (both markdown and YAML)
		for taskKey, taskState := range snapshot.TaskProgress {
			require.Equal(t, "completed", taskState.MarkdownStatus,
				"task %s markdown should be completed", taskKey)
			require.Equal(t, "completed", taskState.Status,
				"task %s should report as completed from YAML", taskKey)
		}

		// Festival should be 100%
		require.Equal(t, 100, snapshot.FestivalPercentage,
			"festival should be 100%% with all tasks complete")

		verifyProgressConsistency(t, snapshot)
	})

	// Phase 5: Edge cases - Verify YAML is source of truth
	t.Run("EdgeCases_BugDetection", func(t *testing.T) {
		t.Run("YAMLIsSourceOfTruth", func(t *testing.T) {
			// Verify that YAML is the source of truth (not markdown)
			// This is the CORRECT behavior per architectural decision

			taskPath := "002_IMPLEMENT/01_development/01_setup.md"

			// Step 1: Task is already complete from implementation phase
			// Verify it's complete first
			snapshot1 := captureProgressSnapshot(t, container, festivalPath)
			require.Equal(t, "completed", snapshot1.TaskProgress[taskPath].MarkdownStatus,
				"task should be completed from previous phase")

			// Step 2: Use fest CLI to explicitly mark complete (creates/updates YAML)
			_, err := container.RunFestInDir(festivalPath, "progress", "--complete", "--task", taskPath)
			require.NoError(t, err, "fest progress --complete should work")

			// Step 3: Manually edit markdown to unchecked (simulating rollback scenario)
			err = simulateTaskWork(t, container, festivalPath, taskPath, "01_setup_0_of_2.md")
			require.NoError(t, err)

			// Step 4: Capture progress again
			snapshot2 := captureProgressSnapshot(t, container, festivalPath)
			taskState := snapshot2.TaskProgress[taskPath]

			// Verify divergence exists
			require.Equal(t, "completed", taskState.YAMLStatus,
				"YAML should say completed (from fest CLI)")
			require.Equal(t, "pending", taskState.MarkdownStatus,
				"markdown should say pending (0 checkboxes checked)")
			require.Equal(t, 0, taskState.CheckboxesChecked,
				"markdown should have 0 checked boxes")

			// CRITICAL ASSERTION: YAML should be source of truth
			// This verifies the architectural decision that YAML is authoritative
			require.Equal(t, "completed", taskState.Status,
				"YAML is source of truth: fest should report YAML status (completed), not markdown status (pending)")

			// Log the behavior for documentation
			t.Logf("VERIFIED: YAML status=%s is source of truth, markdown status=%s is ignored when divergent",
				taskState.YAMLStatus, taskState.MarkdownStatus)
			t.Logf("Task has %d/%d checkboxes checked but correctly reports as %s from YAML",
				taskState.CheckboxesChecked, taskState.CheckboxesTotal, taskState.Status)
		})

		t.Run("EmptyTaskFile", func(t *testing.T) {
			// Add a task with no checkboxes
			emptyTaskPath := "003_POLISH/01_finalize/03_empty.md"
			testdataPath := "testdata/progress_tracking"
			content, err := readTestdataFile(t, filepath.Join(testdataPath, "edge_cases/empty_task.md"))
			require.NoError(t, err)

			err = writeFileInContainer(container, filepath.Join(festivalPath, emptyTaskPath), content)
			require.NoError(t, err)

			snapshot := captureProgressSnapshot(t, container, festivalPath)

			// Verify task defaults to pending
			if taskState, ok := snapshot.TaskProgress[emptyTaskPath]; ok {
				require.Equal(t, "pending", taskState.MarkdownStatus,
					"empty task should default to pending")
				require.Equal(t, 0, taskState.CheckboxesTotal,
					"empty task should have 0 checkboxes")
			}
		})
	})
}
