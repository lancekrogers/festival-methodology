# Fest CLI Styling Guidelines

This document provides guidelines for applying consistent styling to fest command output using the `internal/ui` package.

## Table of Contents

- [Color Usage Guidelines](#color-usage-guidelines)
- [Component Usage Guidelines](#component-usage-guidelines)
- [Code Examples](#code-examples)
- [Best Practices](#best-practices)

## Color Usage Guidelines

The `internal/ui/styles.go` file defines semantic color constants for consistent styling across fest commands.

### Entity Type Colors

Use these colors to identify different festival methodology entities in output:

| Color | Constant | Usage | Example Output |
|-------|----------|-------|----------------|
| Green (42) | `FestivalColor` | Festival names, festival-level information | "Festival: my-festival-FES0001" |
| Blue (33) | `PhaseColor` | Phase names, phase-level information | "Phase: 001_FOUNDATION" |
| Cyan (51) | `SequenceColor` | Sequence names, sequence-level information | "Sequence: 02_create_components" |
| Purple (141) | `TaskColor` | Task names, task-level information | "Task: 01_write_guidelines" |
| Orange (214) | `GateColor` | Quality gate names, gate requirements | "Gate: code_review" |

**When to use:**
- Apply entity colors when displaying hierarchical festival structure
- Use in list views to visually distinguish entity types
- Apply to labels, headers, and names

**Example:**
```go
fmt.Printf("%s: %s\n",
    ui.Label("Festival"),
    ui.Value("my-festival", ui.FestivalColor))
```

### State Colors

Use these colors to indicate progress and status:

| Color | Constant | Usage | Example Context |
|-------|----------|-------|-----------------|
| Grey (245) | `PendingColor` | Not started, waiting, queued | Pending tasks, future phases |
| Yellow (220) | `InProgressColor` | Currently executing, active | Active tasks, running commands |
| Red (196) | `BlockedColor` | Failed, errored, blocked | Failed tests, blocked tasks |
| Green (42) | `SuccessColor` | Completed, passed, successful | Completed tasks, passing tests |

**When to use:**
- Progress bars and completion indicators
- Status badges and labels
- Task/sequence/phase state displays

**Example:**
```go
status := "in_progress"
color := ui.InProgressColor
if status == "completed" {
    color = ui.SuccessColor
}
fmt.Println(ui.ColoredText(status, color))
```

### Structural Element Colors

Use these colors for UI structure and metadata:

| Color | Constant | Usage | Example Context |
|-------|----------|-------|-----------------|
| Grey (240) | `BorderColor` | Borders, separators, dividers | Panel borders, section separators |
| White (255) | `ValueColor` | Important values, highlighted text | Festival names, counts, primary data |
| Grey (245) | `MetadataColor` | Secondary info, timestamps, IDs | Created dates, file paths, identifiers |
| Green (42) | `SuccessColor` | Success messages, confirmations | "✓ Task completed", "All tests passed" |
| Yellow (220) | `WarningColor` | Warnings, advisories, cautions | "⚠ Missing dependencies", "Review needed" |
| Red (196) | `ErrorColor` | Errors, failures, critical issues | "✗ Validation failed", "Error: file not found" |

**When to use:**
- Borders: Panel and border components
- Values: Important data points requiring emphasis
- Metadata: Supporting information that doesn't need emphasis
- Success/Warning/Error: Semantic message highlighting

**Example:**
```go
// Metadata (dim)
fmt.Printf("%s: %s\n",
    ui.Label("Created"),
    ui.Dim("2025-01-05"))

// Important value (bright)
fmt.Printf("%s: %s\n",
    ui.Label("Status"),
    ui.Value("Active"))

// Success message
fmt.Println(ui.Success("✓ All tests passed"))
```

### Legacy Colors (Active/Planned/Archived)

These colors are used for festival lifecycle states:

| Color | Constant | Usage |
|-------|----------|-------|
| Green (42) | `ActiveColor` | Active festivals |
| Blue (33) | `PlannedColor` | Planned festivals |
| Grey (240) | `ArchivedColor` | Archived/completed festivals |

**Note:** `ActiveColor` is aliased to `SuccessColor` and `FestivalColor` for consistency.

### Color Selection Decision Tree

When choosing a color, ask:

1. **Is it a festival entity?** → Use entity color (Festival/Phase/Sequence/Task/Gate)
2. **Is it showing progress/status?** → Use state color (Pending/InProgress/Blocked/Success)
3. **Is it a message?** → Use semantic color (Success/Warning/Error)
4. **Is it structural?** → Use Border/Value/Metadata colors

### TTY vs Non-TTY Behavior

All lipgloss styling (including colors) automatically detects TTY environments:

- **TTY (interactive terminal):** Colors and ANSI codes are applied
- **Non-TTY (pipes, redirects, CI/CD):** Colors are automatically stripped, output is plain text

**No special handling is required** - lipgloss handles this automatically. Commands will not fail in non-TTY environments.

### Color Accessibility

Consider accessibility when applying colors:

- Don't rely solely on color to convey information (use icons, text labels too)
- Ensure sufficient contrast between foreground and background
- Test output in different terminal color schemes
- Provide `--no-color` flag for users who need plain output

## Component Usage Guidelines

The `internal/ui/components.go` file provides reusable UI components for consistent styling.

### Border Components

Use borders to visually separate content or highlight important information.

| Function | Use Case | Example |
|----------|----------|---------|
| `Border(content, opts)` | Custom border with full control | Complex layouts, custom styling |
| `RoundedBorder(content)` | Quick rounded border | Modern, friendly appearance |
| `SquareBorder(content)` | Classic ASCII border | Traditional terminal look |
| `MinimalBorder(content)` | Maximum compatibility | Environments with limited UTF-8 support |

**When to use:**
- Highlighting important output (errors, warnings, summaries)
- Visually separating sections in command output
- Creating focus points in TUI interfaces

**Example:**
```go
// Simple rounded border
output := ui.RoundedBorder("Status: All tests passed")

// Custom border with color
opts := ui.DefaultBorderOptions()
opts.Style = ui.BorderSquare
opts.Color = ui.SuccessColor
output := ui.Border("Task completed successfully", opts)
```

### Panel Components

Use panels for grouped content with titles and semantic meaning.

| Function | Use Case |
|----------|----------|
| `Panel(content, opts)` | Custom panel with full control |
| `TitledPanel(title, content)` | Basic panel with title |
| `InfoPanel(title, content)` | Informational content (green border) |
| `WarningPanel(title, content)` | Warnings, advisories (yellow border) |
| `ErrorPanel(title, content)` | Errors, failures (red border) |

**When to use:**
- Displaying structured information with context
- Grouping related data points
- Semantic messaging (info/warning/error)

**Example:**
```go
// Info panel for status
info := ui.InfoPanel("Status",
    "Festival: my-festival\nPhase: 001_FOUNDATION\nProgress: 25%")

// Error panel for validation failures
errors := ui.ErrorPanel("Validation Failed",
    "- 3 tasks missing\n- 2 naming violations\n- 1 dependency issue")
```

### Header Components

Use headers to create visual hierarchy in command output.

| Function | Level | Use Case |
|----------|-------|----------|
| `H1(text)` | Top-level | Command titles, major sections (uppercase, underlined) |
| `H2(text)` | Section | Primary sections within output (bold) |
| `H3(text)` | Subsection | Subsections, categories (dimmed) |
| `Header(text, opts)` | Custom | Full control over styling |

**When to use:**
- Creating visual hierarchy in output
- Separating major sections
- Making output scannable

**Example:**
```go
fmt.Println(ui.H1("Festival Status Report"))
fmt.Println()
fmt.Println(ui.H2("Active Festivals"))
// ... content ...
fmt.Println()
fmt.Println(ui.H3("Recently Completed"))
// ... content ...
```

### Progress Bar Components

Use progress bars to visualize completion and progress.

| Function | Use Case |
|----------|----------|
| `RenderProgressBar(opts)` | Full control over appearance |
| `SimpleProgressBar(current, total)` | Quick percentage bar |
| `Spinner(frame)` | Animated loading indicator |

**When to use:**
- Showing task/sequence/phase completion percentages
- Long-running operations (with spinner)
- Visual feedback for progress

**Example:**
```go
// Simple progress bar (0-100)
bar := ui.SimpleProgressBar(42, 100)  // "42%"

// Custom progress bar with fraction
opts := ui.DefaultProgressBarOptions()
opts.Current = 15
opts.Total = 23
opts.ShowPercentage = false
opts.ShowFraction = true
bar := ui.RenderProgressBar(opts)  // "15/23"

// Animated spinner for loading
for i := 0; i < 10; i++ {
    fmt.Printf("\r%s Loading...", ui.Spinner(i))
    time.Sleep(100 * time.Millisecond)
}
```

## Code Examples

### Example 1: Festival Status Output

```go
package main

import (
    "fmt"
    "github.com/lancekrogers/festival-methodology/fest/internal/ui"
)

func displayFestivalStatus(name string, phase string, progress int, total int) {
    // Title
    fmt.Println(ui.H1("Festival Status"))
    fmt.Println()

    // Info panel with details
    info := fmt.Sprintf(
        "%s: %s\n%s: %s\n%s: %d/%d",
        ui.Label("Festival"),
        ui.ColoredText(name, ui.FestivalColor),
        ui.Label("Phase"),
        ui.ColoredText(phase, ui.PhaseColor),
        ui.Label("Progress"),
        progress, total,
    )
    fmt.Println(ui.InfoPanel("Details", info))
    fmt.Println()

    // Progress bar
    fmt.Println(ui.H2("Overall Progress"))
    bar := ui.SimpleProgressBar(progress, total)
    fmt.Println(bar)
}
```

### Example 2: Validation Error Display

```go
func displayValidationErrors(errors []string) {
    if len(errors) == 0 {
        fmt.Println(ui.Success("✓ Validation passed"))
        return
    }

    // Error panel
    content := ""
    for _, err := range errors {
        content += fmt.Sprintf("- %s\n", err)
    }

    panel := ui.ErrorPanel(
        fmt.Sprintf("Validation Failed (%d errors)", len(errors)),
        content,
    )
    fmt.Println(panel)
}
```

### Example 3: Task List with State Colors

```go
func displayTasks(tasks []Task) {
    fmt.Println(ui.H2("Tasks"))

    for _, task := range tasks {
        // Choose color based on state
        var stateColor lipgloss.Color
        var icon string

        switch task.State {
        case "completed":
            stateColor = ui.SuccessColor
            icon = "✓"
        case "in_progress":
            stateColor = ui.InProgressColor
            icon = "⟳"
        case "blocked":
            stateColor = ui.BlockedColor
            icon = "✗"
        default:
            stateColor = ui.PendingColor
            icon = "○"
        }

        fmt.Printf("%s %s %s\n",
            ui.ColoredText(icon, stateColor),
            ui.ColoredText(task.Name, ui.TaskColor),
            ui.Dim(task.Path),
        )
    }
}
```

## Best Practices

### 1. Consistency is Key

- **Always use the same color for the same concept** across all commands
- **Reuse components** instead of inline styling
- **Follow the established patterns** shown in existing commands

### 2. Don't Overuse Colors

- **Use color to highlight, not decorate** - too many colors create visual noise
- **Stick to semantic colors** - don't use arbitrary colors
- **Leave most text unstyled** - color should draw attention to important elements

### 3. Structure Before Styling

- **Get the information architecture right first** - structure matters more than color
- **Use components for structure** - borders, panels, headers create hierarchy
- **Style enhances structure** - don't let styling obscure the message

### 4. Test in Multiple Environments

- **Test in different terminals** - colors render differently
- **Test in non-TTY** - ensure output is readable without colors (pipes, redirects)
- **Test with `--no-color`** - verify flag works correctly

### 5. Maintain Performance

- **Don't style inside tight loops** - lipgloss rendering has overhead
- **Cache styled strings** if reusing - don't re-render the same content
- **Keep output concise** - don't generate excessive styled output

### 6. Make it Accessible

- **Don't rely on color alone** - use icons, labels, structure too
- **Provide plain output options** - `--no-color` or `--plain` flags
- **Use high-contrast colors** - ensure readability

### 7. Update Guidelines

When adding new patterns:
- **Document the pattern** in this file
- **Add code examples** showing correct usage
- **Update related commands** to follow the new pattern

## Quick Reference

### Common Patterns

```go
// Label-value pair
fmt.Printf("%s: %s\n", ui.Label("Name"), ui.Value("my-festival"))

// Success message
fmt.Println(ui.Success("✓ Task completed"))

// Warning message
fmt.Println(ui.Warning("⚠ Dependencies missing"))

// Error message
fmt.Println(ui.Error("✗ Validation failed"))

// Section header
fmt.Println(ui.H2("Section Title"))

// Info panel
fmt.Println(ui.InfoPanel("Title", "Content here"))

// Progress bar
fmt.Println(ui.SimpleProgressBar(42, 100))

// Bordered content
fmt.Println(ui.RoundedBorder("Important message"))
```

### Color Quick Reference

- **Green (42):** Success, completion, festivals, active state
- **Blue (33):** Phases, planned state, information
- **Cyan (51):** Sequences
- **Purple (141):** Tasks
- **Orange (214):** Quality gates
- **Yellow (220):** In progress, warnings
- **Red (196):** Errors, blocked, failures
- **Grey (240/245):** Borders, metadata, pending, dimmed text
- **White (255):** Important values, highlighted text

## Visual Examples

This section will be populated with before/after screenshots as commands are styled.

### Command Output Improvements

Visual examples will demonstrate the impact of consistent styling:

- **Before:** Plain text output without colors or structure
- **After:** Styled output with semantic colors and visual hierarchy

Screenshots will be added for:
- `fest status` - Festival status with progress bars and entity colors
- `fest progress` - Progress tracking with state colors
- `fest list` - Festival listing with color-coded states
- `fest validation` - Validation results with error/warning panels
- `fest next` - Task recommendation with highlighted information

*Note: Visual examples will be added as commands are styled in phases 002-005.*

---

*For implementation details, see:*
- `internal/ui/styles.go` - Color definitions
- `internal/ui/components.go` - Component implementations
- `internal/ui/ui.go` - UI helper functions
