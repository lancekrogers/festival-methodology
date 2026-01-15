package show

import (
	"fmt"
	"math"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

// FormatNodeReference creates a node reference string from festival ID and location.
// Format: ID:P###.S##.T## (e.g., GU0001:P002.S01.T03)
// Returns empty string if festivalID is empty.
func FormatNodeReference(festivalID string, phase, sequence, task int) string {
	if festivalID == "" {
		return ""
	}
	return fmt.Sprintf("%s:P%03d.S%02d.T%02d", festivalID, phase, sequence, task)
}

// FormatFestivalDetails formats a single festival with full details.
func FormatFestivalDetails(festival *FestivalInfo, verbose bool) string {
	var sb strings.Builder

	// Header
	sb.WriteString(ui.H1("Festival"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Name"), ui.Value(festival.Name, ui.FestivalColor)))

	// Display festival ID prominently
	if festival.MetadataID != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("ID"), ui.Value(festival.MetadataID)))
	} else {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("ID"), ui.Dim("No ID (run fest migrate to add)")))
	}

	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Status"), ui.GetStatusStyle(festival.Status).Render(festival.Status)))
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Path"), ui.Dim(festival.Path)))

	if festival.Priority != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Priority"), ui.Value(festival.Priority)))
	}

	// Statistics
	if festival.Stats != nil {
		sb.WriteString("\n")
		sb.WriteString(ui.H2("Progress"))
		sb.WriteString("\n")
		sb.WriteString(fmt.Sprintf("%s %s %s\n",
			ui.Label("Overall"),
			renderPercentBar(festival.Stats.Progress),
			ui.Value(fmt.Sprintf("%.1f%%", festival.Stats.Progress))))

		sb.WriteString("\n")
		sb.WriteString(ui.H3("Phases"))
		sb.WriteString("\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Phases))

		sb.WriteString("\n")
		sb.WriteString(ui.H3("Sequences"))
		sb.WriteString("\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Sequences))

		sb.WriteString("\n")
		sb.WriteString(ui.H3("Tasks"))
		sb.WriteString("\n")
		sb.WriteString(formatStatusCounts("  ", festival.Stats.Tasks))

		if festival.Stats.Gates.Total > 0 {
			sb.WriteString("\n")
			sb.WriteString(ui.H3("Gates"))
			sb.WriteString("\n")
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.Label("Total"), ui.Value(fmt.Sprintf("%d", festival.Stats.Gates.Total))))
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.Label("Passed"), ui.GetStateStyle("completed").Render(fmt.Sprintf("%d", festival.Stats.Gates.Passed))))
			sb.WriteString(fmt.Sprintf("  %s %s\n", ui.Label("Failed"), ui.GetStateStyle("blocked").Render(fmt.Sprintf("%d", festival.Stats.Gates.Failed))))
		}
	}

	return sb.String()
}

func formatStatusCounts(prefix string, counts StatusCounts) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, ui.Label("Total"), ui.Value(fmt.Sprintf("%d", counts.Total))))
	sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, ui.Label("Completed"), ui.GetStateStyle("completed").Render(fmt.Sprintf("%d", counts.Completed))))
	sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, ui.Label("In progress"), ui.GetStateStyle("in_progress").Render(fmt.Sprintf("%d", counts.InProgress))))
	sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, ui.Label("Pending"), ui.GetStateStyle("pending").Render(fmt.Sprintf("%d", counts.Pending))))
	if counts.Blocked > 0 {
		sb.WriteString(fmt.Sprintf("%s%s %s\n", prefix, ui.Label("Blocked"), ui.GetStateStyle("blocked").Render(fmt.Sprintf("%d", counts.Blocked))))
	}
	return sb.String()
}

// FormatFestivalList formats a list of festivals for a single status.
func FormatFestivalList(status string, festivals []*FestivalInfo) string {
	var sb strings.Builder

	header := fmt.Sprintf("%s Festivals (%d)", strings.ToUpper(status), len(festivals))
	sb.WriteString(ui.GetStatusStyle(status).Render(header))
	sb.WriteString("\n")
	sb.WriteString(ui.Dim(strings.Repeat("─", 40)))
	sb.WriteString("\n")

	if len(festivals) == 0 {
		sb.WriteString(ui.Dim("  (none)\n"))
		return sb.String()
	}

	for _, f := range festivals {
		progress := ""
		if f.Stats != nil {
			progress = ui.Dim(fmt.Sprintf(" [%.0f%%]", f.Stats.Progress))
		}
		styledName := ui.GetStatusStyle(status).Render(f.Name)
		sb.WriteString(fmt.Sprintf("  %s%s\n", styledName, progress))
	}

	return sb.String()
}

// FormatFestivalListWithProgress formats a list of festivals with detailed progress info.
func FormatFestivalListWithProgress(status string, festivals []*FestivalInfo, progressMap map[string]*progress.FestivalProgress) string {
	var sb strings.Builder

	header := fmt.Sprintf("%s Festivals (%d)", strings.ToUpper(status), len(festivals))
	sb.WriteString(ui.GetStatusStyle(status).Render(header))
	sb.WriteString("\n")
	sb.WriteString(ui.Dim(strings.Repeat("─", 40)))
	sb.WriteString("\n")

	if len(festivals) == 0 {
		sb.WriteString(ui.Dim("  (none)\n"))
		return sb.String()
	}

	for _, f := range festivals {
		styledName := ui.GetStatusStyle(status).Render(f.Name)
		sb.WriteString(fmt.Sprintf("  %s\n", styledName))

		// Show detailed progress if available
		if prog, ok := progressMap[f.Path]; ok && prog != nil && prog.Overall != nil {
			overall := prog.Overall
			// Progress bar with percentage and task counts
			bar := renderPercentBar(float64(overall.Percentage))
			sb.WriteString(fmt.Sprintf("    %s %s %s %s\n",
				ui.Label("Overall"),
				bar,
				ui.Value(fmt.Sprintf("%d%%", overall.Percentage)),
				ui.Dim(fmt.Sprintf("(%d/%d tasks)", overall.Completed, overall.Total))))

			// Total time if available
			if overall.TimeSpentMin > 0 {
				sb.WriteString(fmt.Sprintf("    %s %s\n",
					ui.Label("Total time"),
					ui.Value(ui.FormatDuration(overall.TimeSpentMin))))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// FormatAllFestivalsWithProgress formats all festivals grouped by status with detailed progress.
func FormatAllFestivalsWithProgress(allFestivals map[string][]*FestivalInfo, statusOrder []string, progressMap map[string]*progress.FestivalProgress) string {
	var sb strings.Builder

	total := 0
	for _, festivals := range allFestivals {
		total += len(festivals)
	}

	sb.WriteString(ui.H1("All Festivals"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Total"), ui.Value(fmt.Sprintf("%d", total))))
	sb.WriteString(ui.Dim(strings.Repeat("─", 40)))
	sb.WriteString("\n\n")

	for _, status := range statusOrder {
		festivals := allFestivals[status]
		sb.WriteString(FormatFestivalListWithProgress(status, festivals, progressMap))
		sb.WriteString("\n")
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

	sb.WriteString(ui.H1("All Festivals"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Total"), ui.Value(fmt.Sprintf("%d", total))))
	sb.WriteString(ui.Dim(strings.Repeat("─", 40)))
	sb.WriteString("\n\n")

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

	sb.WriteString(ui.H1("Location"))
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Festival"), ui.Value(loc.Festival.Name, ui.FestivalColor)))
	sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Location"), ui.Value(loc.Type)))

	if loc.Phase != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Phase"), ui.Value(loc.Phase, ui.PhaseColor)))
	}
	if loc.Sequence != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Sequence"), ui.Value(loc.Sequence, ui.SequenceColor)))
	}
	if loc.Task != "" {
		sb.WriteString(fmt.Sprintf("%s %s\n", ui.Label("Task"), ui.Value(loc.Task, ui.TaskColor)))
	}

	return sb.String()
}

func renderPercentBar(progress float64) string {
	opts := ui.DefaultProgressBarOptions()
	opts.Current = int(math.Round(progress))
	opts.Total = 100
	opts.Width = 24
	opts.ShowPercentage = false
	opts.ShowFraction = false
	opts.FilledColor = ui.SuccessColor
	opts.EmptyColor = ui.BorderColor
	return ui.RenderProgressBar(opts)
}
