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
	fmt.Println(strings.Repeat("=", 80))

	// Load festival metadata
	festConfig, err := config.LoadFestivalConfig(festivalPath)
	if err != nil || festConfig.Metadata.ID == "" {
		fmt.Println("Current Context: (no ID - legacy festival)")
		fmt.Println(strings.Repeat("=", 80))
		return
	}

	// Get current location numbers
	phase, seq, task := getCurrentLocationNumbers(festivalPath)

	// Format and print node reference
	nodeRef := formatNodeReference(festConfig.Metadata.ID, phase, seq, task)
	fmt.Printf("Current Context: %s\n", nodeRef)
	fmt.Println(strings.Repeat("=", 80))

	// Print agent guidance with TODO example
	fmt.Println()
	fmt.Println("Self-Check Guidance:")
	fmt.Println("  Use this reference in code comments for traceability:")
	fmt.Printf("  // TODO(%s): Description of work needed\n", nodeRef)
}

func printValidationResult(display *ui.UI, festivalPath string, result *ValidationResult) {
	// Print context header with node reference
	printContextHeader(festivalPath)

	fmt.Printf("\nFestival Validation: %s\n", result.Festival)
	fmt.Println(strings.Repeat("=", 50))

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
	printTemplateValidationSection(display, templateIssues)
	printValidationSection(display, "ORDERING", orderingIssues)

	// Score and summary
	fmt.Printf("\nScore: %d/100", result.Score)
	if result.Valid {
		fmt.Println(" - Festival structure is valid")
	} else {
		fmt.Println(" - Festival structure needs attention")
	}

	// Suggestions
	if len(result.Suggestions) > 0 {
		fmt.Println("\nSuggestions:")
		for _, s := range result.Suggestions {
			fmt.Printf("  • %s\n", s)
		}
	}

	// Agent self-check prompts
	printAgentReflection(display, result)
}

func printValidationSection(display *ui.UI, title string, issues []ValidationIssue) {
	fmt.Printf("\n%s\n", title)

	if len(issues) == 0 {
		display.Success("[OK] All checks passed")
		return
	}

	for _, issue := range issues {
		switch issue.Level {
		case LevelError:
			display.Error("[ERROR] %s", issue.Message)
		case LevelWarning:
			display.Warning("[WARN] %s", issue.Message)
		case LevelInfo:
			display.Info("[INFO] %s", issue.Message)
		}
		if issue.Path != "" {
			fmt.Printf("        Path: %s\n", issue.Path)
		}
		if issue.Fix != "" {
			fmt.Printf("        FIX: %s\n", issue.Fix)
		}
	}
}

func printTemplateValidationSection(display *ui.UI, issues []ValidationIssue) {
	fmt.Println("\nTEMPLATES (Required for Valid Festival)")

	if len(issues) == 0 {
		display.Success("[OK] All template markers have been filled")
		return
	}

	display.Error("[ERROR] Files contain unfilled template markers")
	fmt.Println()
	fmt.Println("        Template markers are PLACEHOLDERS that MUST be replaced with actual content.")
	fmt.Println("        A festival with unfilled markers is INCOMPLETE.")
	fmt.Println()
	fmt.Println("        Files needing attention:")

	for _, issue := range issues {
		fmt.Printf("        - %s\n", issue.Path)
	}

	fmt.Println()
	fmt.Println("        For each file, edit and replace:")
	fmt.Println("          [FILL: description]  → Write actual content")
	fmt.Println("          [REPLACE: guidance]  → Replace with your content")
	fmt.Println("          {{ placeholder }}    → Fill in the value")
}

func printAgentReflection(display *ui.UI, result *ValidationResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println("AGENT SELF-CHECK")
	fmt.Println(strings.Repeat("─", 50))
	fmt.Println()

	if !result.Valid || result.Score < 100 {
		display.Warning("Before continuing, reflect on these questions:")
		fmt.Println()
		fmt.Println("  1. Did you follow the festival methodology exactly as written?")
		fmt.Println()
		fmt.Println("  2. Did you understand the PURPOSE of each level in the hierarchy?")
		fmt.Println("     • Phases organize major milestones")
		fmt.Println("     • Sequences group related work toward a goal")
		fmt.Println("     • Tasks define specific executable work")
		fmt.Println()
		fmt.Println("  3. If someone follows this festival structure exactly,")
		fmt.Println("     will the goals at each level be achieved?")
		fmt.Println()
		fmt.Println("  4. Are ALL placeholder markers filled with real content?")
		fmt.Println("     Templates exist to guide you, not to remain as-is.")
		fmt.Println()
	} else {
		display.Success("Festival structure passes all checks.")
		fmt.Println()
		fmt.Println("  Verify: If an agent executes each task in order,")
		fmt.Println("  will the sequence goals, phase goals, and festival goal be achieved?")
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
