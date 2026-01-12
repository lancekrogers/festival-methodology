package graduate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/graduate"
	"github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

type graduateOptions struct {
	from          string
	to            string
	path          string
	dryRun        bool
	jsonOutput    bool
	noInteractive bool
}

// NewGraduateCommand creates the graduate command.
func NewGraduateCommand() *cobra.Command {
	opts := &graduateOptions{}

	cmd := &cobra.Command{
		Use:   "graduate",
		Short: "Graduate planning phase to implementation structure",
		Long: `Graduate a freeform planning phase into a structured implementation phase.

The graduate command analyzes planning phase documents and proposes an implementation
structure with properly numbered sequences and tasks.

WORKFLOW:
1. Analyzes planning phase documents (topics, decisions, summary)
2. Proposes implementation structure based on planning content
3. Shows the plan for confirmation (unless --no-interactive)
4. Creates the implementation phase with sequences and task stubs

EXAMPLES:
  fest graduate                            # Interactive mode, auto-detect planning phase
  fest graduate --from 001_PLANNING        # Specify source phase
  fest graduate --dry-run                  # Preview without creating files
  fest graduate --json --dry-run           # Output plan as JSON
  fest graduate --no-interactive           # Create without prompts`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			if ctx == nil {
				ctx = context.Background()
			}
			return runGraduate(ctx, opts)
		},
	}

	cmd.Flags().StringVar(&opts.from, "from", "", "source planning phase directory")
	cmd.Flags().StringVar(&opts.to, "to", "", "target implementation phase name")
	cmd.Flags().StringVar(&opts.path, "path", "", "festival path (default: current directory)")
	cmd.Flags().BoolVar(&opts.dryRun, "dry-run", false, "preview plan without creating files")
	cmd.Flags().BoolVar(&opts.jsonOutput, "json", false, "output plan as JSON")
	cmd.Flags().BoolVar(&opts.noInteractive, "no-interactive", false, "skip all prompts")

	return cmd
}

func runGraduate(ctx context.Context, opts *graduateOptions) error {
	// Resolve festival path
	var festivalPath string
	var err error

	if opts.path != "" {
		// Explicit path provided - use it directly
		festivalPath, err = filepath.Abs(opts.path)
		if err != nil {
			return errors.Wrap(err, "resolving path")
		}
	} else {
		// Try to resolve from current directory
		cwd, err := os.Getwd()
		if err != nil {
			return errors.Wrap(err, "getting current directory")
		}
		festivalPath, err = shared.ResolveFestivalPath(cwd, "")
		if err != nil {
			return err
		}
	}

	// Create analyzer
	analyzer := graduate.NewAnalyzer(festivalPath)

	// Find or use specified planning phase
	var phasePath string
	if opts.from != "" {
		phasePath = filepath.Join(festivalPath, opts.from)
	} else {
		phasePath, err = analyzer.FindPlanningPhase(ctx)
		if err != nil {
			return errors.Wrap(err, "no planning phase found; use --from to specify")
		}
	}

	// Analyze planning phase
	source, err := analyzer.Analyze(ctx, phasePath)
	if err != nil {
		return errors.Wrap(err, "analyzing planning phase")
	}

	// Generate graduation plan
	generator := graduate.NewGenerator(festivalPath)
	plan, err := generator.Generate(ctx, source)
	if err != nil {
		return errors.Wrap(err, "generating graduation plan")
	}

	// Override target if specified
	if opts.to != "" {
		plan.Target.PhaseName = opts.to
		plan.Target.Path = filepath.Join(festivalPath, opts.to)
	}

	// Output plan as JSON if requested
	if opts.jsonOutput {
		return outputJSON(plan)
	}

	// Dry-run mode with enhanced output
	if opts.dryRun {
		displayDryRun(plan, shared.IsVerbose())
		return nil
	}

	// Interactive loop
	for {
		displayPlan(plan)

		if opts.noInteractive {
			break // Skip prompts, execute directly
		}

		choice := promptChoice()
		switch choice {
		case ChoiceAccept:
			goto execute
		case ChoiceEdit:
			plan = editPlan(plan)
			continue // Re-display plan
		case ChoiceCancel:
			fmt.Println("Cancelled.")
			return nil
		case ChoiceHelp:
			printHelp()
			continue
		}
	}

execute:
	// Execute the plan
	executor := graduate.NewExecutor(festivalPath)
	if err := executor.Execute(ctx, plan); err != nil {
		return errors.Wrap(err, "executing graduation plan")
	}

	// Success message
	display := ui.New(shared.IsNoColor(), shared.IsVerbose())
	display.Success("Graduation complete!")
	fmt.Printf("\nCreated: %s\n", plan.Target.PhaseName)
	fmt.Printf("  Sequences: %d\n", len(plan.Sequences))

	totalTasks := 0
	for _, seq := range plan.Sequences {
		totalTasks += len(seq.Tasks)
	}
	fmt.Printf("  Tasks: %d\n", totalTasks)

	return nil
}

func outputJSON(plan *graduate.GraduationPlan) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(plan)
}

func displayPlan(plan *graduate.GraduationPlan) {
	fmt.Println(strings.Repeat("=", 70))
	fmt.Println("                    GRADUATION PLAN")
	fmt.Println(strings.Repeat("=", 70))

	fmt.Printf("\nSOURCE: %s\n", plan.Source.PhaseName)
	fmt.Printf("  Topics: %d\n", len(plan.Source.TopicDirs))
	fmt.Printf("  Documents: %d\n", plan.Source.TotalDocs)
	if len(plan.Source.Decisions) > 0 {
		fmt.Printf("  Decisions: %d ADRs\n", len(plan.Source.Decisions))
	}

	fmt.Printf("\nTARGET: %s", plan.Target.PhaseName)
	if _, err := os.Stat(plan.Target.Path); os.IsNotExist(err) {
		fmt.Print(" (will be created)")
	}
	fmt.Println()

	fmt.Println(strings.Repeat("-", 70))
	fmt.Println()
	fmt.Println("PROPOSED STRUCTURE:")
	fmt.Println()

	fmt.Printf("%s/\n", plan.Target.PhaseName)
	fmt.Println("|-- PHASE_GOAL.md")

	for i, seq := range plan.Sequences {
		prefix := "|--"
		if i == len(plan.Sequences)-1 {
			prefix = "`--"
		}
		fmt.Printf("%s %s/\n", prefix, seq.FullName)

		taskPrefix := "|   "
		if i == len(plan.Sequences)-1 {
			taskPrefix = "    "
		}

		fmt.Printf("%s|-- SEQUENCE_GOAL.md\n", taskPrefix)
		for j, task := range seq.Tasks {
			taskLine := "|--"
			if j == len(seq.Tasks)-1 {
				taskLine = "`--"
			}
			sourceInfo := ""
			if len(task.SourceDocs) > 0 {
				sourceInfo = fmt.Sprintf(" (from: %s)", task.SourceDocs[0])
			}
			fmt.Printf("%s%s %s%s\n", taskPrefix, taskLine, task.FullName, sourceInfo)
		}
	}

	fmt.Println(strings.Repeat("-", 70))
	fmt.Printf("Confidence: %.0f%%", plan.Confidence*100)
	if len(plan.Warnings) > 0 {
		fmt.Printf(" (%d warnings)", len(plan.Warnings))
	}
	fmt.Println()

	if len(plan.Warnings) > 0 {
		fmt.Println("\nWarnings:")
		for _, w := range plan.Warnings {
			fmt.Printf("  * %s\n", w)
		}
	}

	fmt.Println(strings.Repeat("=", 70))
}

// InteractiveChoice represents a user's choice.
type InteractiveChoice int

const (
	ChoiceAccept InteractiveChoice = iota
	ChoiceEdit
	ChoiceCancel
	ChoiceHelp
)

func promptChoice() InteractiveChoice {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("\n[A] Accept and create  [E] Edit plan  [C] Cancel  [?] Help")
	fmt.Print("Choice: ")

	response, _ := reader.ReadString('\n')
	response = strings.ToLower(strings.TrimSpace(response))

	switch response {
	case "a", "accept", "y", "yes":
		return ChoiceAccept
	case "e", "edit":
		return ChoiceEdit
	case "c", "cancel", "n", "no":
		return ChoiceCancel
	case "?", "h", "help":
		return ChoiceHelp
	default:
		fmt.Println("Invalid choice. Use A, E, C, or ?")
		return promptChoice()
	}
}

func confirm(prompt string) bool {
	fmt.Printf("\n%s [y/N]: ", prompt)
	var response string
	_, _ = fmt.Scanln(&response)
	return strings.ToLower(response) == "y" || strings.ToLower(response) == "yes"
}

func editPlan(plan *graduate.GraduationPlan) *graduate.GraduationPlan {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Println("\n" + strings.Repeat("─", 50))
		fmt.Println("EDIT MODE")
		fmt.Println(strings.Repeat("─", 50))
		fmt.Println()

		// Show sequences with numbers
		for i, seq := range plan.Sequences {
			fmt.Printf("  %d. %s (%d tasks)\n", i+1, seq.FullName, len(seq.Tasks))
		}

		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  rename <n> <name>  - Rename sequence n")
		fmt.Println("  remove <n>         - Remove sequence n")
		fmt.Println("  move <n> <pos>     - Move sequence n to position")
		fmt.Println("  show <n>           - Show sequence n details")
		fmt.Println("  done               - Return to main menu")
		fmt.Println()
		fmt.Print("Edit> ")

		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || input == "done" {
			break
		}

		parts := strings.Fields(input)
		if len(parts) == 0 {
			continue
		}

		cmd := strings.ToLower(parts[0])
		switch cmd {
		case "rename":
			if len(parts) < 3 {
				fmt.Println("Usage: rename <n> <name>")
				continue
			}
			plan = renameSequence(plan, parts[1], strings.Join(parts[2:], "_"))

		case "remove":
			if len(parts) < 2 {
				fmt.Println("Usage: remove <n>")
				continue
			}
			plan = removeSequence(plan, parts[1])

		case "move":
			if len(parts) < 3 {
				fmt.Println("Usage: move <n> <position>")
				continue
			}
			plan = moveSequence(plan, parts[1], parts[2])

		case "show":
			if len(parts) < 2 {
				fmt.Println("Usage: show <n>")
				continue
			}
			showSequenceDetail(plan, parts[1])

		default:
			fmt.Printf("Unknown command: %s\n", cmd)
		}
	}

	// Renumber sequences after edits
	return renumberSequences(plan)
}

func renameSequence(plan *graduate.GraduationPlan, nStr, newName string) *graduate.GraduationPlan {
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 || n > len(plan.Sequences) {
		fmt.Println("Invalid sequence number")
		return plan
	}

	plan.Sequences[n-1].Name = newName
	fmt.Printf("Renamed sequence %d to %s\n", n, newName)
	return plan
}

func removeSequence(plan *graduate.GraduationPlan, nStr string) *graduate.GraduationPlan {
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 || n > len(plan.Sequences) {
		fmt.Println("Invalid sequence number")
		return plan
	}

	seq := plan.Sequences[n-1]
	plan.Sequences = append(plan.Sequences[:n-1], plan.Sequences[n:]...)
	fmt.Printf("Removed sequence: %s\n", seq.FullName)
	return plan
}

func moveSequence(plan *graduate.GraduationPlan, fromStr, toStr string) *graduate.GraduationPlan {
	from, err1 := strconv.Atoi(fromStr)
	to, err2 := strconv.Atoi(toStr)
	if err1 != nil || err2 != nil || from < 1 || to < 1 ||
		from > len(plan.Sequences) || to > len(plan.Sequences) {
		fmt.Println("Invalid position")
		return plan
	}

	// Remove from old position
	seq := plan.Sequences[from-1]
	plan.Sequences = append(plan.Sequences[:from-1], plan.Sequences[from:]...)

	// Insert at new position
	if to-1 >= len(plan.Sequences) {
		plan.Sequences = append(plan.Sequences, seq)
	} else {
		plan.Sequences = append(plan.Sequences[:to-1],
			append([]graduate.ProposedSequence{seq}, plan.Sequences[to-1:]...)...)
	}

	fmt.Printf("Moved sequence to position %d\n", to)
	return plan
}

func showSequenceDetail(plan *graduate.GraduationPlan, nStr string) {
	n, err := strconv.Atoi(nStr)
	if err != nil || n < 1 || n > len(plan.Sequences) {
		fmt.Println("Invalid sequence number")
		return
	}

	seq := plan.Sequences[n-1]
	fmt.Printf("\nSequence: %s\n", seq.FullName)
	fmt.Printf("Source Topic: %s\n", seq.SourceTopic)
	fmt.Println("\nTasks:")
	for i, task := range seq.Tasks {
		fmt.Printf("  %d. %s\n", i+1, task.Name)
		if len(task.SourceDocs) > 0 {
			fmt.Printf("     From: %s\n", task.SourceDocs[0])
		}
	}
}

func renumberSequences(plan *graduate.GraduationPlan) *graduate.GraduationPlan {
	for i := range plan.Sequences {
		plan.Sequences[i].Number = i + 1
		plan.Sequences[i].FullName = fmt.Sprintf("%02d_%s", i+1, plan.Sequences[i].Name)

		for j := range plan.Sequences[i].Tasks {
			plan.Sequences[i].Tasks[j].Number = j + 1
			plan.Sequences[i].Tasks[j].FullName = fmt.Sprintf("%02d_%s.md",
				j+1, plan.Sequences[i].Tasks[j].Name)
		}
	}
	return plan
}

func printHelp() {
	fmt.Print(`
GRADUATION HELP

The graduate command creates an implementation phase from your planning documents.

Options:
  A - Accept the plan and create the structure
  E - Edit the plan before creating
  C - Cancel and exit

In Edit Mode:
  rename <n> <name> - Change sequence name
  remove <n>        - Remove a sequence
  move <n> <pos>    - Reorder sequences
  show <n>          - View sequence details
  done              - Return to main menu

Tip: Use --dry-run to preview without creating files.
`)
}

func displayDryRun(plan *graduate.GraduationPlan, verbose bool) {
	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("                    DRY RUN PREVIEW")
	fmt.Println(strings.Repeat("═", 70))

	// Check if target exists
	if _, err := os.Stat(plan.Target.Path); err == nil {
		fmt.Println("\n⚠️  WARNING: Target phase already exists!")
		fmt.Printf("   Path: %s\n", plan.Target.Path)
		fmt.Println("   Existing files will NOT be overwritten.")
	}

	// Summary statistics
	totalFiles := 1 // PHASE_GOAL.md
	for _, seq := range plan.Sequences {
		totalFiles += 1 + len(seq.Tasks) // SEQUENCE_GOAL.md + tasks
	}

	totalTasks := 0
	for _, seq := range plan.Sequences {
		totalTasks += len(seq.Tasks)
	}

	fmt.Println("\n📊 Summary:")
	fmt.Printf("   Files to create:     %d\n", totalFiles)
	fmt.Printf("   Directories:         %d\n", len(plan.Sequences)+1)
	fmt.Printf("   Sequences:           %d\n", len(plan.Sequences))
	fmt.Printf("   Tasks:               %d\n", totalTasks)

	// File tree
	fmt.Println("\n📁 Files to be created:")
	fmt.Println()
	displayFileTree(plan)

	// Verbose: show file contents
	if verbose {
		fmt.Println("\n" + strings.Repeat("─", 70))
		fmt.Println("📄 File Contents Preview")
		fmt.Println(strings.Repeat("─", 70))

		displayFileContents(plan)
	}

	fmt.Println("\n" + strings.Repeat("═", 70))
	fmt.Println("[DRY RUN] No files were created.")
	fmt.Println("Run without --dry-run to create these files.")
	fmt.Println(strings.Repeat("═", 70))
}

func displayFileTree(plan *graduate.GraduationPlan) {
	fmt.Printf("%s/\n", plan.Target.PhaseName)

	// PHASE_GOAL.md
	phaseStatus := "new"
	phaseGoalPath := filepath.Join(plan.Target.Path, "PHASE_GOAL.md")
	if _, err := os.Stat(phaseGoalPath); err == nil {
		phaseStatus = "exists, will skip"
	}
	fmt.Printf("├── PHASE_GOAL.md (%s)\n", phaseStatus)

	for i, seq := range plan.Sequences {
		isLast := i == len(plan.Sequences)-1
		prefix := "├──"
		if isLast {
			prefix = "└──"
		}

		seqPath := filepath.Join(plan.Target.Path, seq.FullName)
		seqStatus := "new"
		if _, err := os.Stat(seqPath); err == nil {
			seqStatus = "exists"
		}
		fmt.Printf("%s %s/ (%s)\n", prefix, seq.FullName, seqStatus)

		childPrefix := "│   "
		if isLast {
			childPrefix = "    "
		}

		// SEQUENCE_GOAL.md
		seqGoalPath := filepath.Join(seqPath, "SEQUENCE_GOAL.md")
		seqGoalStatus := "new"
		if _, err := os.Stat(seqGoalPath); err == nil {
			seqGoalStatus = "exists, will skip"
		}
		fmt.Printf("%s├── SEQUENCE_GOAL.md (%s)\n", childPrefix, seqGoalStatus)

		// Tasks
		for j, task := range seq.Tasks {
			isLastTask := j == len(seq.Tasks)-1
			taskPrefix := "├──"
			if isLastTask {
				taskPrefix = "└──"
			}

			taskPath := filepath.Join(seqPath, task.FullName)
			taskStatus := "new"
			if _, err := os.Stat(taskPath); err == nil {
				taskStatus = "exists, will skip"
			}

			sourceInfo := ""
			if len(task.SourceDocs) > 0 {
				sourceInfo = fmt.Sprintf(" ← %s", task.SourceDocs[0])
			}
			fmt.Printf("%s%s %s (%s)%s\n", childPrefix, taskPrefix, task.FullName, taskStatus, sourceInfo)
		}
	}
}

func displayFileContents(plan *graduate.GraduationPlan) {
	// Show PHASE_GOAL.md preview
	fmt.Printf("\n── PHASE_GOAL.md ──\n")
	phaseContent := generatePhaseGoalPreview(plan)
	fmt.Println(truncate(phaseContent, 20))

	// Show first sequence as example
	if len(plan.Sequences) > 0 {
		seq := plan.Sequences[0]

		fmt.Printf("\n── %s/SEQUENCE_GOAL.md ──\n", seq.FullName)
		seqContent := generateSequenceGoalPreview(&seq)
		fmt.Println(truncate(seqContent, 15))

		// Show first task as example
		if len(seq.Tasks) > 0 {
			task := seq.Tasks[0]
			fmt.Printf("\n── %s/%s ──\n", seq.FullName, task.FullName)
			taskContent := generateTaskPreview(&task)
			fmt.Println(truncate(taskContent, 15))
		}
	}

	fmt.Println("\n(Showing first examples only. Full content will be generated on execute.)")
}

func generatePhaseGoalPreview(plan *graduate.GraduationPlan) string {
	return fmt.Sprintf(`# Phase Goal: %s

**Status:** Not Started | **Graduated From:** %s

## Phase Objective

**Primary Goal:** %s

## Planning Reference

Graduated from: %s
`,
		plan.Target.PhaseName,
		plan.Source.PhaseName,
		plan.PhaseGoal.Goal,
		plan.Source.PhaseName)
}

func generateSequenceGoalPreview(seq *graduate.ProposedSequence) string {
	return fmt.Sprintf(`# Sequence Goal: %s

**Sequence:** %s | **Status:** Not Started

## Sequence Objective

**Primary Goal:** %s

**Source Topic:** %s
`,
		seq.Name,
		seq.FullName,
		seq.Goal.Goal,
		seq.SourceTopic)
}

func generateTaskPreview(task *graduate.ProposedTask) string {
	sourceRef := ""
	if len(task.SourceDocs) > 0 {
		sourceRef = fmt.Sprintf("\n**Planning Reference:** `%s`", task.SourceDocs[0])
	}

	return fmt.Sprintf(`# Task: %s

> **Task Number**: %02d | **Dependencies**: None

## Objective

%s%s

## Requirements

- [ ] Implement functionality
- [ ] Add tests
`,
		task.Name,
		task.Number,
		task.Objective,
		sourceRef)
}

func truncate(content string, maxLines int) string {
	lines := strings.Split(content, "\n")
	if len(lines) <= maxLines {
		return content
	}
	return strings.Join(lines[:maxLines], "\n") + "\n  ... (truncated)"
}
