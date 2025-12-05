package commands

import (
    "fmt"
    "os"
    "path/filepath"
    "strings"

    tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
    "github.com/lancekrogers/festival-methodology/fest/internal/ui"
    "github.com/spf13/cobra"
)

// NewTUICommand launches an interactive text UI for common actions
func NewTUICommand() *cobra.Command {
    cmd := &cobra.Command{
        Use:   "tui",
        Short: "Interactive UI for creating festivals, phases, sequences, and tasks",
        RunE: func(cmd *cobra.Command, args []string) error {
            return runTUI()
        },
    }
    return cmd
}

func runTUI() error {
    display := ui.New(noColor, verbose)
    cwd, _ := os.Getwd()

    // Ensure we are inside a festivals workspace; if not, offer to init
    if _, err := tpl.FindFestivalsRoot(cwd); err != nil {
        display.Warning("No festivals/ directory detected.")
        if display.Confirm("Initialize a new festival workspace here?") {
            if err := runInit(".", &initOptions{}); err != nil {
                return err
            }
        } else {
            return fmt.Errorf("no festivals/ directory detected")
        }
    }

    for {
        choice := display.Choose("What would you like to do?", []string{
            "Plan a New Festival (wizard)",
            "Create a Festival (quick)",
            "Add a Phase",
            "Add a Sequence",
            "Add a Task",
            "Generate Festival Goal",
            "Quit",
        })

        switch choice {
        case 0:
            if err := tuiPlanFestivalWizard(display); err != nil {
                return err
            }
        case 1:
            if err := tuiCreateFestival(display); err != nil {
                return err
            }
        case 2:
            if err := tuiCreatePhase(display); err != nil {
                return err
            }
        case 3:
            if err := tuiCreateSequence(display); err != nil {
                return err
            }
        case 4:
            if err := tuiCreateTask(display); err != nil {
                return err
            }
        case 5:
            if err := tuiGenerateFestivalGoal(display); err != nil {
                return err
            }
        default:
            return nil
        }

        if !display.Confirm("Do you want to perform another action?") {
            break
        }
    }
    return nil
}

func tuiCreateFestival(display *ui.UI) error {
    cwd, _ := os.Getwd()
    tmplRoot, err := tpl.LocalTemplateRoot(cwd)
    if err != nil {
        return err
    }

    name := strings.TrimSpace(display.Prompt("Festival name"))
    if name == "" {
        return fmt.Errorf("festival name is required")
    }
    goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
    tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
    dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "active")))
    if dest != "planned" && dest != "active" {
        dest = "active"
    }

    // Collect additional variables required by core festival templates
    required := uniqueStrings(collectRequiredVars(tmplRoot, defaultFestivalTemplatePaths(tmplRoot)))

    vars := map[string]interface{}{}
    // Pre-populate typical variables
    vars["festival_name"] = name
    vars["festival_goal"] = goal
    if tags != "" {
        // keep as string; create command handles tags flag for standardized parsing
        vars["festival_tags"] = strings.Split(tags, ",")
    }

    // Ask for any missing variables not already filled
    for _, v := range required {
        if v == "festival_name" || v == "festival_goal" || v == "festival_tags" || v == "festival_description" {
            continue
        }
        if _, ok := vars[v]; ok {
            continue
        }
        val := strings.TrimSpace(display.PromptDefault(fmt.Sprintf("%s", v), ""))
        if val != "" {
            vars[v] = val
        }
    }

    // Write variables to a temporary JSON file
    varsFile, err := writeTempVarsFile(vars)
    if err != nil {
        return err
    }

    opts := &createFestivalOptions{
        name:     name,
        goal:     goal,
        tags:     tags,
        varsFile: varsFile,
        dest:     dest,
    }
    return runCreateFestival(opts)
}

// Wizard: create festival then optionally add phases
func tuiPlanFestivalWizard(display *ui.UI) error {
    cwd, _ := os.Getwd()
    festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
    if err != nil {
        return err
    }
    // First create festival (quick)
    cwdTmpl, err := tpl.LocalTemplateRoot(cwd)
    if err != nil {
        return err
    }
    name := strings.TrimSpace(display.Prompt("Festival name"))
    if name == "" {
        return fmt.Errorf("festival name is required")
    }
    goal := strings.TrimSpace(display.PromptDefault("Festival goal", ""))
    tags := strings.TrimSpace(display.PromptDefault("Tags (comma-separated)", ""))
    dest := strings.ToLower(strings.TrimSpace(display.PromptDefault("Destination (active|planned)", "planned")))
    if dest != "planned" && dest != "active" {
        dest = "planned"
    }

    // gather extra vars from templates
    required := uniqueStrings(collectRequiredVars(cwdTmpl, defaultFestivalTemplatePaths(cwdTmpl)))
    vars := map[string]interface{}{"festival_name": name, "festival_goal": goal}
    if tags != "" { vars["festival_tags"] = strings.Split(tags, ",") }
    for _, v := range required {
        if v == "festival_name" || v == "festival_goal" || v == "festival_tags" || v == "festival_description" { continue }
        val := strings.TrimSpace(display.PromptDefault(v, ""))
        if val != "" { vars[v] = val }
    }
    varsFile, err := writeTempVarsFile(vars)
    if err != nil { return err }

    if err := runCreateFestival(&createFestivalOptions{name: name, goal: goal, tags: tags, varsFile: varsFile, dest: dest}); err != nil {
        return err
    }

    // Compute created path
    slug := slugify(name)
    festivalDir := filepath.Join(festivalsRoot, dest, slug)

    // Optionally add phases
    if display.Confirm("Add initial phases now?") {
        countStr := display.PromptDefault("How many phases?", "0")
        count := atoiDefault(countStr, 0)
        after := 0
        for i := 0; i < count; i++ {
            pname := strings.TrimSpace(display.Prompt(fmt.Sprintf("Phase %d name (e.g., PLAN)", i+1)))
            if pname == "" { pname = fmt.Sprintf("PHASE_%d", i+1) }
            ptype := strings.TrimSpace(display.PromptDefault("Phase type (planning|implementation|review|deployment)", "planning"))
            if ptype == "" { ptype = "planning" }
            if err := runCreatePhase(&createPhaseOptions{after: after, name: pname, phaseType: ptype, path: festivalDir}); err != nil {
                return err
            }
            after++
        }
    }
    display.Success("Festival planned: %s (%s)", slug, dest)
    display.Info("Location: %s", festivalDir)
    return nil
}

func tuiGenerateFestivalGoal(display *ui.UI) error {
    cwd, _ := os.Getwd()
    if _, err := tpl.LocalTemplateRoot(cwd); err != nil {
        return err
    }
    festDir := strings.TrimSpace(display.PromptDefault("Festival directory (where to write FESTIVAL_GOAL.md)", "."))
    if festDir == "" { festDir = "." }
    name := strings.TrimSpace(display.PromptDefault("festival_name", ""))
    goal := strings.TrimSpace(display.PromptDefault("festival_goal", ""))
    vars := map[string]interface{}{}
    if name != "" { vars["festival_name"] = name }
    if goal != "" { vars["festival_goal"] = goal }
    varsFile, err := writeTempVarsFile(vars)
    if err != nil { return err }
    // Use apply to render template to destination
    destPath := filepath.Join(festDir, "FESTIVAL_GOAL.md")
    return runApply(&applyOptions{templatePath: "FESTIVAL_GOAL_TEMPLATE.md", destPath: destPath, varsFile: varsFile})
}

func tuiCreatePhase(display *ui.UI) error {
    cwd, _ := os.Getwd()
    tmplRoot, err := tpl.LocalTemplateRoot(cwd)
    if err != nil {
        return err
    }

    name := strings.TrimSpace(display.Prompt("Phase name (e.g., PLAN)"))
    if name == "" {
        return fmt.Errorf("phase name is required")
    }
    // Choose phase type
    types := []string{"planning", "implementation", "review", "deployment"}
    tIdx := display.Choose("Phase type:", types)
    if tIdx < 0 || tIdx >= len(types) {
        tIdx = 0
    }
    phaseType := types[tIdx]

    path := strings.TrimSpace(display.PromptDefault("Festival directory (contains numbered phases)", "."))
    afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", "0"))
    after := atoiDefault(afterStr, 0)

    required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
        filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md"),
    }))
    vars := map[string]interface{}{}
    // Gather missing variables
    for _, v := range required {
        // Context will already set phase numbering/name/type; ask for extras only
        if v == "phase_number" || v == "phase_name" || v == "phase_type" {
            continue
        }
        val := strings.TrimSpace(display.PromptDefault(v, ""))
        if val != "" {
            vars[v] = val
        }
    }
    varsFile, err := writeTempVarsFile(vars)
    if err != nil {
        return err
    }

    opts := &createPhaseOptions{
        after:     after,
        name:      name,
        phaseType: phaseType,
        path:      path,
        varsFile:  varsFile,
    }
    return runCreatePhase(opts)
}

func tuiCreateSequence(display *ui.UI) error {
    cwd, _ := os.Getwd()
    tmplRoot, err := tpl.LocalTemplateRoot(cwd)
    if err != nil {
        return err
    }

    name := strings.TrimSpace(display.Prompt("Sequence name (e.g., requirements)"))
    if name == "" {
        return fmt.Errorf("sequence name is required")
    }
    path := strings.TrimSpace(display.PromptDefault("Phase directory (contains numbered sequences)", "."))
    afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", "0"))
    after := atoiDefault(afterStr, 0)

    required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
        filepath.Join(tmplRoot, "SEQUENCE_GOAL_TEMPLATE.md"),
    }))
    vars := map[string]interface{}{}
    for _, v := range required {
        if v == "sequence_number" || v == "sequence_name" {
            continue
        }
        val := strings.TrimSpace(display.PromptDefault(v, ""))
        if val != "" {
            vars[v] = val
        }
    }
    varsFile, err := writeTempVarsFile(vars)
    if err != nil {
        return err
    }

    opts := &createSequenceOptions{
        after:     after,
        name:      name,
        path:      path,
        varsFile:  varsFile,
    }
    return runCreateSequence(opts)
}

func tuiCreateTask(display *ui.UI) error {
    cwd, _ := os.Getwd()
    tmplRoot, err := tpl.LocalTemplateRoot(cwd)
    if err != nil {
        return err
    }

    name := strings.TrimSpace(display.Prompt("Task name (e.g., user_research)"))
    if name == "" {
        return fmt.Errorf("task name is required")
    }
    path := strings.TrimSpace(display.PromptDefault("Sequence directory (contains numbered task files)", "."))
    afterStr := strings.TrimSpace(display.PromptDefault("Insert after number (0 to insert at beginning)", "0"))
    after := atoiDefault(afterStr, 0)

    // Prefer TASK_TEMPLATE.md for required vars
    required := uniqueStrings(collectRequiredVars(tmplRoot, []string{
        filepath.Join(tmplRoot, "TASK_TEMPLATE.md"),
    }))
    vars := map[string]interface{}{}
    for _, v := range required {
        if v == "task_number" || v == "task_name" {
            continue
        }
        val := strings.TrimSpace(display.PromptDefault(v, ""))
        if val != "" {
            vars[v] = val
        }
    }
    varsFile, err := writeTempVarsFile(vars)
    if err != nil {
        return err
    }

    opts := &createTaskOptions{
        after:     after,
        name:      name,
        path:      path,
        varsFile:  varsFile,
    }
    return runCreateTask(opts)
}
