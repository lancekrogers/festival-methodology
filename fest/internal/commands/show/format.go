package show

import (
	"fmt"
	"strings"
)

// FormatFestivalDetails formats a single festival with full details.
func FormatFestivalDetails(festival *FestivalInfo, verbose bool) string {
	var sb strings.Builder

	// Header
	sb.WriteString(fmt.Sprintf("Festival: %s\n", festival.Name))
	sb.WriteString(fmt.Sprintf("  Status: %s\n", festival.Status))
	sb.WriteString(fmt.Sprintf("  Path:   %s\n", festival.Path))

	if festival.Priority != "" {
		sb.WriteString(fmt.Sprintf("  Priority: %s\n", festival.Priority))
	}

	// Statistics
	if festival.Stats != nil {
		sb.WriteString("\nProgress:\n")
		sb.WriteString(fmt.Sprintf("  Overall: %.1f%%\n", festival.Stats.Progress))

		sb.WriteString("\nPhases:\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Phases))

		sb.WriteString("\nSequences:\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Sequences))

		sb.WriteString("\nTasks:\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Tasks))

		if festival.Stats.Gates.Total > 0 {
			sb.WriteString("\nGates:\n")
			sb.WriteString(fmt.Sprintf("  Total:  %d\n", festival.Stats.Gates.Total))
			sb.WriteString(fmt.Sprintf("  Passed: %d\n", festival.Stats.Gates.Passed))
			sb.WriteString(fmt.Sprintf("  Failed: %d\n", festival.Stats.Gates.Failed))
		}
	}

	return sb.String()
}

func formatStatusCounts(prefix string, counts StatusCounts) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%sTotal:       %d\n", prefix, counts.Total))
	sb.WriteString(fmt.Sprintf("%sCompleted:   %d\n", prefix, counts.Completed))
	sb.WriteString(fmt.Sprintf("%sIn Progress: %d\n", prefix, counts.InProgress))
	sb.WriteString(fmt.Sprintf("%sPending:     %d\n", prefix, counts.Pending))
	if counts.Blocked > 0 {
		sb.WriteString(fmt.Sprintf("%sBlocked:     %d\n", prefix, counts.Blocked))
	}
	return sb.String()
}

// FormatFestivalList formats a list of festivals for a single status.
func FormatFestivalList(status string, festivals []*FestivalInfo) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s Festivals (%d)\n", strings.Title(status), len(festivals)))
	sb.WriteString(strings.Repeat("-", 40) + "\n")

	if len(festivals) == 0 {
		sb.WriteString("  (none)\n")
		return sb.String()
	}

	for _, f := range festivals {
		progress := ""
		if f.Stats != nil {
			progress = fmt.Sprintf(" [%.0f%%]", f.Stats.Progress)
		}
		sb.WriteString(fmt.Sprintf("  %s%s\n", f.Name, progress))
	}

	return sb.String()
}

// FormatAllFestivals formats all festivals grouped by status.
func FormatAllFestivals(allFestivals map[string][]*FestivalInfo, statusOrder []string) string {
	var sb strings.Builder

	total := 0
	for _, festivals := range allFestivals {
		total += len(festivals)
	}

	sb.WriteString(fmt.Sprintf("All Festivals (%d total)\n", total))
	sb.WriteString(strings.Repeat("=", 40) + "\n\n")

	for _, status := range statusOrder {
		festivals := allFestivals[status]
		sb.WriteString(FormatFestivalList(status, festivals))
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatLocation formats the current location within a festival.
func FormatLocation(loc *LocationInfo) string {
	var sb strings.Builder

	if loc.Festival == nil {
		return "Not in a festival directory\n"
	}

	sb.WriteString(fmt.Sprintf("Festival: %s\n", loc.Festival.Name))
	sb.WriteString(fmt.Sprintf("Location: %s\n", loc.Type))

	if loc.Phase != "" {
		sb.WriteString(fmt.Sprintf("  Phase: %s\n", loc.Phase))
	}
	if loc.Sequence != "" {
		sb.WriteString(fmt.Sprintf("  Sequence: %s\n", loc.Sequence))
	}
	if loc.Task != "" {
		sb.WriteString(fmt.Sprintf("  Task: %s\n", loc.Task))
	}

	return sb.String()
}
