// Package help provides context-aware help functionality for fest CLI.
package help

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/spf13/cobra"
)

// ContextualHelp provides location-aware help information.
type ContextualHelp struct {
	location *show.LocationInfo
}

// NewContextualHelp creates a new contextual help instance based on current directory.
func NewContextualHelp(ctx context.Context) (*ContextualHelp, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return &ContextualHelp{}, nil // Fall back to generic help
	}

	location, err := show.DetectCurrentLocation(ctx, cwd)
	if err != nil {
		return &ContextualHelp{}, nil // Fall back to generic help
	}

	return &ContextualHelp{location: location}, nil
}

// GetContext returns a description of the current context.
func (h *ContextualHelp) GetContext() string {
	if h.location == nil {
		return "Not currently in a festival directory"
	}

	var parts []string
	if h.location.Festival != nil {
		parts = append(parts, fmt.Sprintf("Festival: %s", h.location.Festival.Name))
	}
	if h.location.Phase != "" {
		parts = append(parts, fmt.Sprintf("Phase: %s", h.location.Phase))
	}
	if h.location.Sequence != "" {
		parts = append(parts, fmt.Sprintf("Sequence: %s", h.location.Sequence))
	}
	if h.location.Task != "" {
		parts = append(parts, fmt.Sprintf("Task: %s", h.location.Task))
	}

	if len(parts) == 0 {
		return "Not currently in a festival directory"
	}

	return strings.Join(parts, " > ")
}

// GetSuggestedCommands returns suggested commands based on current location.
func (h *ContextualHelp) GetSuggestedCommands() []CommandSuggestion {
	if h.location == nil {
		return genericSuggestions()
	}

	switch h.location.Type {
	case "festival":
		return festivalSuggestions(h.location)
	case "phase":
		return phaseSuggestions(h.location)
	case "sequence":
		return sequenceSuggestions(h.location)
	case "task":
		return taskSuggestions(h.location)
	default:
		return genericSuggestions()
	}
}

// CommandSuggestion represents a suggested command with context.
type CommandSuggestion struct {
	Command     string
	Description string
}

func genericSuggestions() []CommandSuggestion {
	return []CommandSuggestion{
		{"fest tui", "Interactive festival management"},
		{"fest show all", "List all available festivals"},
		{"fest create festival", "Create a new festival"},
		{"fest understand", "Learn about Festival Methodology"},
		{"fest init", "Initialize a festival workspace"},
	}
}

func festivalSuggestions(loc *show.LocationInfo) []CommandSuggestion {
	suggestions := []CommandSuggestion{
		{"fest status", "View festival status and progress"},
		{"fest validate", "Validate festival structure"},
		{"fest next", "Show next task to work on"},
	}

	if loc.Festival != nil && loc.Festival.Stats != nil {
		if loc.Festival.Stats.Phases.Total == 0 {
			suggestions = append(suggestions,
				CommandSuggestion{"fest create phase", "Add a phase to this festival"})
		}
	}

	suggestions = append(suggestions,
		CommandSuggestion{"fest go <phase>", "Navigate to a specific phase"})

	return suggestions
}

func phaseSuggestions(loc *show.LocationInfo) []CommandSuggestion {
	phase := loc.Phase
	return []CommandSuggestion{
		{"fest status", fmt.Sprintf("View %s phase status", phase)},
		{"fest next", "Show next task in this phase"},
		{"fest go ..", "Navigate to festival root"},
		{"fest create sequence", "Add a sequence to this phase"},
		{"fest validate", "Validate phase structure"},
	}
}

func sequenceSuggestions(loc *show.LocationInfo) []CommandSuggestion {
	seq := loc.Sequence
	return []CommandSuggestion{
		{"fest status", fmt.Sprintf("View %s sequence status", seq)},
		{"fest next", "Show next task in this sequence"},
		{"fest progress --complete", "Mark current task complete"},
		{"fest go ..", "Navigate to parent phase"},
		{"fest create task", "Add a task to this sequence"},
	}
}

func taskSuggestions(loc *show.LocationInfo) []CommandSuggestion {
	return []CommandSuggestion{
		{"fest progress --complete", "Mark this task complete"},
		{"fest next", "Show next task after this one"},
		{"fest go ..", "Navigate to parent sequence"},
		{"fest markers fill", "Fill task markers"},
	}
}

// FormatContextualHelp formats help text with context information.
func (h *ContextualHelp) FormatContextualHelp() string {
	var sb strings.Builder

	// Current context
	ctx := h.GetContext()
	if ctx != "" && ctx != "Not currently in a festival directory" {
		sb.WriteString("CURRENT CONTEXT:\n")
		sb.WriteString("  " + ctx + "\n\n")
	}

	// Suggested commands
	suggestions := h.GetSuggestedCommands()
	if len(suggestions) > 0 {
		sb.WriteString("SUGGESTED NEXT STEPS:\n")
		for _, s := range suggestions {
			sb.WriteString(fmt.Sprintf("  %-30s %s\n", s.Command, s.Description))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// EnhanceCobraHelp adds contextual information to a cobra command's help output.
// This can be called from a command's PreRun or added to the help template.
func EnhanceCobraHelp(cmd *cobra.Command) {
	ctx := context.Background()
	help, err := NewContextualHelp(ctx)
	if err != nil {
		return
	}

	// Get the original help function
	originalHelp := cmd.HelpFunc()

	// Create enhanced help function
	cmd.SetHelpFunc(func(c *cobra.Command, args []string) {
		// Print contextual help first
		contextHelp := help.FormatContextualHelp()
		if contextHelp != "" {
			fmt.Print(contextHelp)
		}

		// Then print the original help
		originalHelp(c, args)
	})
}

// AddContextToLongDescription enhances a command's long description with context.
func AddContextToLongDescription(cmd *cobra.Command) {
	ctx := context.Background()
	help, err := NewContextualHelp(ctx)
	if err != nil {
		return
	}

	contextInfo := help.FormatContextualHelp()
	if contextInfo != "" && cmd.Long != "" {
		cmd.Long = cmd.Long + "\n\n" + contextInfo
	}
}
