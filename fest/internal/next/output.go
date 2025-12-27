package next

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
)

// FormatText formats the result as human-readable text
func FormatText(result *NextTaskResult) string {
	var sb strings.Builder

	if result.FestivalComplete {
		sb.WriteString("🎉 Festival Complete!\n")
		sb.WriteString("All tasks have been completed.\n")
		return sb.String()
	}

	if result.BlockingGate != nil {
		sb.WriteString("⚠️  Quality Gate Required\n\n")
		sb.WriteString(fmt.Sprintf("Phase: %s\n", result.BlockingGate.Phase))
		sb.WriteString(fmt.Sprintf("Type: %s\n", result.BlockingGate.GateType))
		sb.WriteString(fmt.Sprintf("Description: %s\n", result.BlockingGate.Description))
		if len(result.BlockingGate.Criteria) > 0 {
			sb.WriteString("\nCriteria:\n")
			for _, c := range result.BlockingGate.Criteria {
				sb.WriteString(fmt.Sprintf("  - %s\n", c))
			}
		}
		return sb.String()
	}

	if result.Task == nil {
		sb.WriteString("No tasks available\n")
		sb.WriteString(fmt.Sprintf("Reason: %s\n", result.Reason))
		return sb.String()
	}

	// Header
	sb.WriteString("=== NEXT TASK ===\n\n")

	// Primary task
	sb.WriteString(fmt.Sprintf("Task: %s\n", result.Task.Name))
	sb.WriteString(fmt.Sprintf("Path: %s\n", result.Task.Path))
	sb.WriteString(fmt.Sprintf("Sequence: %s\n", result.Task.SequenceName))
	sb.WriteString(fmt.Sprintf("Phase: %s\n", result.Task.PhaseName))

	if result.Task.AutonomyLevel != "" {
		sb.WriteString(fmt.Sprintf("Autonomy: %s\n", result.Task.AutonomyLevel))
	}

	sb.WriteString(fmt.Sprintf("\nReason: %s\n", result.Reason))

	// Parallel tasks
	if len(result.ParallelTasks) > 0 {
		sb.WriteString("\n--- Can Also Work On (Parallel) ---\n")
		for _, task := range result.ParallelTasks {
			sb.WriteString(fmt.Sprintf("  - %s (%s)\n", task.Name, task.SequenceName))
		}
	}

	return sb.String()
}

// FormatJSON formats the result as JSON
func FormatJSON(result *NextTaskResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// FormatVerbose formats the result with additional details
func FormatVerbose(result *NextTaskResult) string {
	var sb strings.Builder

	if result.FestivalComplete {
		sb.WriteString("════════════════════════════════════════\n")
		sb.WriteString("         🎉 FESTIVAL COMPLETE 🎉        \n")
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString("All tasks in the festival have been completed.\n")
		sb.WriteString("Congratulations on finishing the festival!\n")
		return sb.String()
	}

	if result.BlockingGate != nil {
		sb.WriteString("════════════════════════════════════════\n")
		sb.WriteString("         ⚠️  QUALITY GATE REQUIRED      \n")
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString(fmt.Sprintf("Phase: %s\n", result.BlockingGate.Phase))
		sb.WriteString(fmt.Sprintf("Gate Type: %s\n", result.BlockingGate.GateType))
		sb.WriteString(fmt.Sprintf("\n%s\n", result.BlockingGate.Description))
		if len(result.BlockingGate.Criteria) > 0 {
			sb.WriteString("\nCriteria to Pass:\n")
			for i, c := range result.BlockingGate.Criteria {
				sb.WriteString(fmt.Sprintf("  %d. %s\n", i+1, c))
			}
		}
		sb.WriteString("\nComplete the quality gate before proceeding.\n")
		return sb.String()
	}

	if result.Task == nil {
		sb.WriteString("════════════════════════════════════════\n")
		sb.WriteString("         NO TASKS AVAILABLE             \n")
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString(fmt.Sprintf("Reason: %s\n\n", result.Reason))
		sb.WriteString("Location:\n")
		sb.WriteString(fmt.Sprintf("  Festival: %s\n", result.Location.FestivalPath))
		if result.Location.PhasePath != "" {
			sb.WriteString(fmt.Sprintf("  Phase: %s\n", filepath.Base(result.Location.PhasePath)))
		}
		if result.Location.SequencePath != "" {
			sb.WriteString(fmt.Sprintf("  Sequence: %s\n", filepath.Base(result.Location.SequencePath)))
		}
		return sb.String()
	}

	// Detailed task info
	sb.WriteString("════════════════════════════════════════\n")
	sb.WriteString("              NEXT TASK                 \n")
	sb.WriteString("════════════════════════════════════════\n\n")

	sb.WriteString(fmt.Sprintf("📋 Task: %s\n", result.Task.Name))
	sb.WriteString(fmt.Sprintf("   Number: %d\n", result.Task.Number))
	sb.WriteString(fmt.Sprintf("   Path: %s\n", result.Task.Path))
	sb.WriteString("\n")

	sb.WriteString("📍 Location:\n")
	sb.WriteString(fmt.Sprintf("   Phase: %s\n", result.Task.PhaseName))
	sb.WriteString(fmt.Sprintf("   Sequence: %s\n", result.Task.SequenceName))
	sb.WriteString("\n")

	if result.Task.AutonomyLevel != "" || result.Task.ParallelGroup > 0 {
		sb.WriteString("⚙️  Properties:\n")
		if result.Task.AutonomyLevel != "" {
			sb.WriteString(fmt.Sprintf("   Autonomy Level: %s\n", result.Task.AutonomyLevel))
		}
		if result.Task.ParallelGroup > 0 {
			sb.WriteString(fmt.Sprintf("   Parallel Group: %d\n", result.Task.ParallelGroup))
		}
		sb.WriteString("\n")
	}

	if len(result.Task.Dependencies) > 0 {
		sb.WriteString("🔗 Dependencies (satisfied):\n")
		for _, dep := range result.Task.Dependencies {
			sb.WriteString(fmt.Sprintf("   ✓ %s\n", dep))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(fmt.Sprintf("💡 Reason: %s\n", result.Reason))
	sb.WriteString("\n")

	// Parallel tasks
	if len(result.ParallelTasks) > 0 {
		sb.WriteString("════════════════════════════════════════\n")
		sb.WriteString("       PARALLEL TASKS AVAILABLE         \n")
		sb.WriteString("════════════════════════════════════════\n\n")
		sb.WriteString("These tasks can be worked on simultaneously:\n\n")
		for _, task := range result.ParallelTasks {
			sb.WriteString(fmt.Sprintf("  • %s\n", task.Name))
			sb.WriteString(fmt.Sprintf("    Path: %s\n", task.Path))
			if task.AutonomyLevel != "" {
				sb.WriteString(fmt.Sprintf("    Autonomy: %s\n", task.AutonomyLevel))
			}
			sb.WriteString("\n")
		}
	}

	// Location context
	sb.WriteString("════════════════════════════════════════\n")
	sb.WriteString("         CURRENT LOCATION               \n")
	sb.WriteString("════════════════════════════════════════\n\n")
	sb.WriteString(fmt.Sprintf("Festival: %s\n", filepath.Base(result.Location.FestivalPath)))
	if result.Location.PhasePath != "" {
		sb.WriteString(fmt.Sprintf("Phase: %s\n", filepath.Base(result.Location.PhasePath)))
	}
	if result.Location.SequencePath != "" {
		sb.WriteString(fmt.Sprintf("Sequence: %s\n", filepath.Base(result.Location.SequencePath)))
	}

	return sb.String()
}

// FormatShort formats a minimal one-line output
func FormatShort(result *NextTaskResult) string {
	if result.FestivalComplete {
		return "Festival complete"
	}
	if result.BlockingGate != nil {
		return fmt.Sprintf("Blocked: Quality gate in %s", result.BlockingGate.Phase)
	}
	if result.Task == nil {
		return "No tasks available"
	}
	return result.Task.Path
}

// FormatCD formats output suitable for shell cd command
func FormatCD(result *NextTaskResult) string {
	if result.Task == nil {
		return ""
	}
	// Return the directory containing the task file
	return filepath.Dir(result.Task.Path)
}
