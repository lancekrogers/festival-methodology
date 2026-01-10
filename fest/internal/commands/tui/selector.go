//go:build !no_charm

// Package tui provides terminal user interface components for fest.
package tui

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	uitheme "github.com/lancekrogers/festival-methodology/fest/internal/ui/theme"
	"github.com/lancekrogers/festival-methodology/fest/internal/workspace"
)

// statusPriority defines sort order for festival statuses.
var statusPriority = map[string]int{
	"active":    0, // Active festivals first
	"planned":   1, // Then planned
	"completed": 2, // Then completed
	"dungeon":   3, // Dungeon last
}

// allStatuses lists all valid festival statuses in priority order.
var allStatuses = []string{"active", "planned", "completed", "dungeon"}

// FestivalSelectorConfig configures the festival selector behavior.
type FestivalSelectorConfig struct {
	// Title displayed at the top of the selector
	Title string
	// Description shown below the title
	Description string
	// FilterByStatus limits selection to specific statuses (empty = all)
	FilterByStatus []string
	// AllowCancel enables Esc to cancel selection
	AllowCancel bool
	// ShowStats shows festival statistics in the list
	ShowStats bool
}

// DefaultSelectorConfig returns the default configuration.
func DefaultSelectorConfig() FestivalSelectorConfig {
	return FestivalSelectorConfig{
		Title:          "Select Festival",
		Description:    "Choose a festival from the list",
		FilterByStatus: nil, // All statuses
		AllowCancel:    true,
		ShowStats:      true,
	}
}

// FestivalSelectorResult holds the result of a festival selection.
type FestivalSelectorResult struct {
	// Selected is the chosen festival (nil if cancelled)
	Selected *show.FestivalInfo
	// Cancelled is true if user pressed Esc
	Cancelled bool
}

// FestivalSelector provides interactive festival selection.
type FestivalSelector struct {
	config       FestivalSelectorConfig
	festivalsDir string
	festivals    []*show.FestivalInfo
}

// NewFestivalSelector creates a new festival selector.
func NewFestivalSelector(festivalsDir string, config FestivalSelectorConfig) *FestivalSelector {
	return &FestivalSelector{
		config:       config,
		festivalsDir: festivalsDir,
	}
}

// NewFestivalSelectorFromCwd creates a festival selector using the festivals directory
// detected from the current working directory.
func NewFestivalSelectorFromCwd(config FestivalSelectorConfig) (*FestivalSelector, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	festivalsDir, err := workspace.FindFestivals(cwd)
	if err != nil {
		return nil, err
	}
	return NewFestivalSelector(festivalsDir, config), nil
}

// Run executes the interactive festival selection.
func (s *FestivalSelector) Run(ctx context.Context) (*FestivalSelectorResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Load all festivals
	if err := s.loadFestivals(ctx); err != nil {
		return nil, err
	}

	if len(s.festivals) == 0 {
		return &FestivalSelectorResult{Cancelled: true}, nil
	}

	// Build options for the selector
	options := s.buildOptions()

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(s.config.Title).
				Description(s.config.Description).
				Options(options...).
				Value(&selected),
		),
	).WithTheme(theme())

	if err := form.Run(); err != nil {
		if uitheme.IsCancelled(err) {
			return &FestivalSelectorResult{Cancelled: true}, nil
		}
		return nil, err
	}

	// Find the selected festival
	for _, f := range s.festivals {
		if f.Path == selected {
			return &FestivalSelectorResult{Selected: f}, nil
		}
	}

	return &FestivalSelectorResult{Cancelled: true}, nil
}

// loadFestivals loads all festivals based on the configuration.
func (s *FestivalSelector) loadFestivals(ctx context.Context) error {
	statuses := s.config.FilterByStatus
	if len(statuses) == 0 {
		statuses = allStatuses
	}

	s.festivals = nil
	for _, status := range statuses {
		festivals, err := show.ListFestivalsByStatus(ctx, s.festivalsDir, status)
		if err != nil {
			continue // Skip inaccessible directories
		}
		s.festivals = append(s.festivals, festivals...)
	}

	// Sort by status priority, then by name
	sortFestivals(s.festivals)

	return nil
}

// sortFestivals sorts festivals by status priority, then alphabetically by name.
func sortFestivals(festivals []*show.FestivalInfo) {
	sort.Slice(festivals, func(i, j int) bool {
		// First sort by status priority
		pi := statusPriority[festivals[i].Status]
		pj := statusPriority[festivals[j].Status]
		if pi != pj {
			return pi < pj
		}
		// Then by name (alphabetically)
		return festivals[i].Name < festivals[j].Name
	})
}

// Festivals returns the loaded festivals (call loadFestivals first).
func (s *FestivalSelector) Festivals() []*show.FestivalInfo {
	return s.festivals
}

// FestivalCount returns the number of loaded festivals.
func (s *FestivalSelector) FestivalCount() int {
	return len(s.festivals)
}

// GroupedFestivals groups festivals by status for organized display.
type GroupedFestivals struct {
	Status    string
	Festivals []*show.FestivalInfo
}

// GroupByStatus groups festivals by their status.
func GroupByStatus(festivals []*show.FestivalInfo) []GroupedFestivals {
	groups := make(map[string][]*show.FestivalInfo)

	for _, f := range festivals {
		groups[f.Status] = append(groups[f.Status], f)
	}

	// Return in priority order
	var result []GroupedFestivals
	for _, status := range allStatuses {
		if fests, ok := groups[status]; ok && len(fests) > 0 {
			result = append(result, GroupedFestivals{
				Status:    status,
				Festivals: fests,
			})
		}
	}

	return result
}

// buildOptions creates huh.Option slice for the selector.
func (s *FestivalSelector) buildOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(s.festivals))

	for _, f := range s.festivals {
		opt := formatFestivalOption(f, s.config.ShowStats)
		options = append(options, huh.NewOption(opt.Label, f.Path))
	}

	return options
}

// FestivalOption represents a festival in the selection list.
type FestivalOption struct {
	Label    string             // Display text (with badge and icon)
	Value    string             // Festival path for selection
	Festival *show.FestivalInfo // Full festival info
}

// formatFestivalOption creates a formatted option for display.
func formatFestivalOption(f *show.FestivalInfo, showStats bool) FestivalOption {
	// Format: 🔥 [active] my-festival-name (3 phases, 12 tasks)
	icon := statusIcon(f.Status)
	badge := statusBadge(f.Status)

	label := fmt.Sprintf("%s %s %s", icon, badge, f.Name)

	if showStats && f.Stats != nil {
		label = fmt.Sprintf("%s (%d phases, %d tasks)",
			label, f.Stats.Phases.Total, f.Stats.Tasks.Total)
	}

	return FestivalOption{
		Label:    label,
		Value:    f.Path,
		Festival: f,
	}
}

// statusBadge returns a formatted status badge string.
func statusBadge(status string) string {
	switch status {
	case "active":
		return "[active]"
	case "planned":
		return "[planned]"
	case "completed":
		return "[completed]"
	case "dungeon":
		return "[dungeon]"
	default:
		return "[" + status + "]"
	}
}

// statusIcon returns an icon for the status.
func statusIcon(status string) string {
	switch status {
	case "active":
		return "🔥"
	case "planned":
		return "📋"
	case "completed":
		return "✅"
	case "dungeon":
		return "🏰"
	default:
		return "📁"
	}
}

// FilterFestivals returns festivals matching the search text.
func FilterFestivals(festivals []*show.FestivalInfo, searchText string) []*show.FestivalInfo {
	if searchText == "" {
		return festivals
	}

	search := strings.ToLower(searchText)
	var filtered []*show.FestivalInfo

	for _, f := range festivals {
		if strings.Contains(strings.ToLower(f.Name), search) {
			filtered = append(filtered, f)
		}
	}

	return filtered
}

// SelectFestivalInteractive runs the festival selector and returns the selected path.
// This is a convenience function for commands that need simple festival selection.
func SelectFestivalInteractive(ctx context.Context, title, description string) (string, error) {
	config := DefaultSelectorConfig()
	if title != "" {
		config.Title = title
	}
	if description != "" {
		config.Description = description
	}

	selector, err := NewFestivalSelectorFromCwd(config)
	if err != nil {
		return "", err
	}

	result, err := selector.Run(ctx)
	if err != nil {
		return "", err
	}

	if result.Cancelled || result.Selected == nil {
		return "", nil
	}

	return result.Selected.Path, nil
}

// SelectActiveFestival runs the selector filtered to only active festivals.
func SelectActiveFestival(ctx context.Context) (string, error) {
	config := FestivalSelectorConfig{
		Title:          "Select Active Festival",
		Description:    "Choose an active festival to work on",
		FilterByStatus: []string{"active"},
		AllowCancel:    true,
		ShowStats:      true,
	}

	selector, err := NewFestivalSelectorFromCwd(config)
	if err != nil {
		return "", err
	}

	result, err := selector.Run(ctx)
	if err != nil {
		return "", err
	}

	if result.Cancelled || result.Selected == nil {
		return "", nil
	}

	return result.Selected.Path, nil
}
