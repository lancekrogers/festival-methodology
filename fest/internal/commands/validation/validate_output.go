package validation

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

// formatNodeReference creates a node reference string from festival ID and location.
// Format: ID:P###.S##.T## (e.g., GU0001:P002.S01.T03)
func formatNodeReference(festivalID string, phase, sequence, task int) string {
	if festivalID == "" {
		return ""
	}
	return fmt.Sprintf("%s:P%03d.S%02d.T%02d", festivalID, phase, sequence, task)
}

// getCurrentLocationNumbers determines phase, sequence, and task numbers from cwd relative to festival
func getCurrentLocationNumbers(festivalPath string) (phase, sequence, task int) {
	cwd, err := os.Getwd()
	if err != nil {
		return 0, 0, 0
	}

	rel, err := filepath.Rel(festivalPath, cwd)
	if err != nil || strings.HasPrefix(rel, "..") {
		return 0, 0, 0
	}

	parts := strings.Split(rel, string(filepath.Separator))
	if len(parts) == 0 || parts[0] == "." {
		return 0, 0, 0
	}

	// Extract phase number from first part (e.g., "001_Research" -> 1)
	if len(parts) >= 1 {
		phase = extractLeadingNumber(parts[0])
	}

	// Extract sequence number from second part
	if len(parts) >= 2 {
		sequence = extractLeadingNumber(parts[1])
	}

	// Task number would be from third part or task file - for now just location
	return phase, sequence, task
}

// extractLeadingNumber extracts the leading number from a string (e.g., "001_Name" -> 1)
func extractLeadingNumber(s string) int {
	num := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			num = num*10 + int(c-'0')
		} else {
			break
		}
	}
	return num
}

// printContextHeader prints the current context header with node reference
func printContextHeader(festivalPath string) {
	fmt.Println()
	fmt.Println(ui.H2("Current Context"))

	// Load festival metadata
	festConfig, err := config.LoadFestivalConfig(festivalPath)
	if err != nil || festConfig.Metadata.ID == "" {
		fmt.Println(ui.Dim("No ID detected (legacy festival)."))
		return
	}

	// Get current location numbers
	phase, seq, task := getCurrentLocationNumbers(festivalPath)

	// Format and print node reference
	nodeRef := formatNodeReference(festConfig.Metadata.ID, phase, seq, task)
	fmt.Printf("%s %s\n", ui.Label("Node"), ui.Value(nodeRef))

	// Print agent guidance with TODO example
	fmt.Println()
	fmt.Println(ui.H3("Self-Check Guidance"))
	fmt.Println(ui.Dim("Use this reference in code comments for traceability:"))
	fmt.Printf("  %s\n", ui.Dim(fmt.Sprintf("// TODO(%s): Description of work needed", nodeRef)))
}

func printValidationResult(display *ui.UI, festivalPath string, result *ValidationResult) {
	// Print context header with node reference
	printContextHeader(festivalPath)

	fmt.Println()
	fmt.Println(ui.H1("Festival Validation"))
	fmt.Printf("%s %s\n", ui.Label("Festival"), ui.Value(result.Festival, ui.FestivalColor))
	fmt.Printf("%s %s\n", ui.Label("Path"), ui.Dim(festivalPath))

	// Group issues by category
	structureIssues := filterIssuesByCode(result.Issues, CodeNamingConvention)
	completenessIssues := filterIssuesByCode(result.Issues, CodeMissingFile, CodeMissingGoal)
	taskIssues := filterIssuesByCode(result.Issues, CodeMissingTaskFiles)
	gateIssues := filterIssuesByCode(result.Issues, CodeMissingQualityGate)
	templateIssues := filterIssuesByCode(result.Issues, CodeUnfilledTemplate)
	orderingIssues := filterIssuesByCode(result.Issues, CodeNumberingGap)

	printValidationSection(display, "STRUCTURE", structureIssues)
	printValidationSection(display, "COMPLETENESS", completenessIssues)
	printTaskValidationSection(display, taskIssues)
	printValidationSection(display, "QUALITY GATES", gateIssues)
	printMarkerValidationSection(display, templateIssues, result.MarkerInfo)
	printValidationSection(display, "ORDERING", orderingIssues)

	// Score and summary
	fmt.Println()
	fmt.Printf("%s %s\n", ui.Label("Score"), ui.Value(fmt.Sprintf("%d/100", result.Score)))
	if result.Valid {
		fmt.Println(ui.Success("Festival structure is valid"))
	} else {
		fmt.Println(ui.Warning("Festival structure needs attention"))
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Println()
		fmt.Println(ui.H3("Suggestions"))
		for _, s := range result.Suggestions {
			fmt.Printf("  • %s\n", ui.Info(s))
		}
	}

	// Agent self-check prompts
	printAgentReflection(display, result)
}

func printValidationSection(display *ui.UI, title string, issues []ValidationIssue) {
	printSectionHeader(title, issues)

	if len(issues) == 0 {
		display.Success("All checks passed")
		return
	}

	for _, issue := range issues {
		printValidationIssue(display, issue)
	}
}

func printMarkerValidationSection(display *ui.UI, issues []ValidationIssue, markerInfo *MarkerInfo) {
	printSectionHeader("Markers", issues)
	fmt.Println(ui.Dim("Template completion status"))

	if len(issues) == 0 {
		if markerInfo != nil && markerInfo.TotalCount > 0 {
			display.Success("All template markers have been filled (%d markers found)", markerInfo.TotalCount)
		} else {
			display.Success("All template markers have been filled (0 markers found)")
		}
		return
	}

	// Show error with count
	if markerInfo != nil {
		display.Error("Found %d unfilled markers in %d files", markerInfo.TotalCount, markerInfo.TotalFiles)
	} else {
		display.Error("Files contain unfilled template markers")
	}

	fmt.Println()
	fmt.Println(ui.Info("Template markers are placeholders that must be replaced with actual content."))
	fmt.Println(ui.Info("A festival with unfilled markers is incomplete."))
	fmt.Println()

	if len(issues) > 0 {
		fmt.Println(ui.H3("Files needing attention"))
		for _, issue := range issues {
			fmt.Printf("  • %s\n", ui.Dim(issue.Path))
			fmt.Printf("    %s\n", ui.Info(issue.Message))
		}
	}

	fmt.Println()
	fmt.Println(ui.H3("Common marker types to replace"))
	fmt.Println("  [FILL: description]    → Write actual content")
	fmt.Println("  [REPLACE: guidance]    → Replace with your content")
	fmt.Println("  [GUIDANCE: hint]       → Remove and write real content")
	fmt.Println("  {{ placeholder }}      → Fill in the value")
	fmt.Println()
	fmt.Println(ui.Info("Run 'fest markers list' to see all unfilled markers"))
}

func printAgentReflection(display *ui.UI, result *ValidationResult) {
	fmt.Println()
	fmt.Println(ui.H2("Agent Self-Check"))

	if !result.Valid || result.Score < 100 {
		display.Warning("Before continuing, reflect on these questions:")
		fmt.Println()
		fmt.Println(ui.Info("  1. Did you follow the festival methodology exactly as written?"))
		fmt.Println()
		fmt.Println(ui.Info("  2. Did you understand the purpose of each level in the hierarchy?"))
		fmt.Println(ui.Dim("     • Phases organize major milestones"))
		fmt.Println(ui.Dim("     • Sequences group related work toward a goal"))
		fmt.Println(ui.Dim("     • Tasks define specific executable work"))
		fmt.Println()
		fmt.Println(ui.Info("  3. If someone follows this festival structure exactly,"))
		fmt.Println(ui.Info("     will the goals at each level be achieved?"))
		fmt.Println()
		fmt.Println(ui.Info("  4. Are all placeholder markers filled with real content?"))
		fmt.Println(ui.Dim("     Templates exist to guide you, not to remain as-is."))
		fmt.Println()
	} else {
		display.Success("Festival structure passes all checks.")
		fmt.Println()
		fmt.Println(ui.Info("  Verify: If an agent executes each task in order,"))
		fmt.Println(ui.Info("  will the sequence goals, phase goals, and festival goal be achieved?"))
	}
}

func printSectionHeader(title string, issues []ValidationIssue) {
	state := sectionState(issues)
	fmt.Printf("\n%s %s\n", ui.StateIcon(state), ui.H2(title))
}

func sectionState(issues []ValidationIssue) string {
	state := "completed"
	for _, issue := range issues {
		switch issue.Level {
		case LevelError:
			return "blocked"
		case LevelWarning:
			state = "in_progress"
		case LevelInfo:
			if state == "completed" {
				state = "pending"
			}
		}
	}
	return state
}

func printValidationIssue(display *ui.UI, issue ValidationIssue) {
	switch issue.Level {
	case LevelError:
		display.Error("%s", issue.Message)
	case LevelWarning:
		display.Warning("%s", issue.Message)
	case LevelInfo:
		display.Info("%s", issue.Message)
	}
	if issue.Path != "" {
		fmt.Printf("  %s %s\n", ui.Label("Path"), ui.Dim(issue.Path))
	}
	if issue.Fix != "" {
		fmt.Printf("  %s %s\n", ui.Label("Fix"), ui.Dim(issue.Fix))
	}
}

func filterIssuesByCode(issues []ValidationIssue, codes ...string) []ValidationIssue {
	codeSet := make(map[string]bool)
	for _, c := range codes {
		codeSet[c] = true
	}

	var filtered []ValidationIssue
	for _, issue := range issues {
		if codeSet[issue.Code] {
			filtered = append(filtered, issue)
		}
	}
	return filtered
}

// countFileMarkers counts unfilled template markers in a file
func countFileMarkers(path string) int {
	markers := []string{"[FILL:", "[GUIDANCE:", "{{ "}
	count := 0

	file, err := os.Open(path)
	if err != nil {
		return 0
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		for _, marker := range markers {
			count += strings.Count(line, marker)
		}
	}

	return count
}
