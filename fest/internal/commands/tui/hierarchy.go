package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
	uitheme "github.com/lancekrogers/festival-methodology/fest/internal/ui/theme"
)

// HierarchyLevel represents a level in the selection hierarchy.
type HierarchyLevel int

const (
	LevelFestival HierarchyLevel = iota
	LevelPhase
	LevelSequence
	LevelTask
)

// SelectionState holds the current state of hierarchical selection.
type SelectionState struct {
	FestivalsRoot string // e.g., /path/to/festivals
	Festival      string // e.g., "active/my-festival"
	Phase         string // e.g., "001_PLANNING"
	Sequence      string // e.g., "01_requirements"

	// Resolved paths
	FestivalPath string
	PhasePath    string
	SequencePath string
}

// HierarchyConfig configures which levels to show.
type HierarchyConfig struct {
	StartLevel   HierarchyLevel // Where to start selection
	EndLevel     HierarchyLevel // Where to end selection
	AllowCreate  bool           // Show "Create new..." option
	FilterActive bool           // Only show active festivals
}

// LevelOption represents a selectable item at any level.
type LevelOption struct {
	Label string // Display label
	Value string // Directory name
	Path  string // Full resolved path
}

// HierarchySelector provides cascading dropdown selection.
type HierarchySelector struct {
	config HierarchyConfig
	state  SelectionState
}

// NewHierarchySelector creates a new selector with the given config.
func NewHierarchySelector(festivalsRoot string, config HierarchyConfig) *HierarchySelector {
	return &HierarchySelector{
		config: config,
		state: SelectionState{
			FestivalsRoot: festivalsRoot,
		},
	}
}

// NewHierarchySelectorFromCwd creates a selector initialized from current directory.
func NewHierarchySelectorFromCwd(config HierarchyConfig) (*HierarchySelector, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, errors.IO("getting working directory", err)
	}

	// Find festivals root
	festivalsRoot, err := tpl.FindFestivalsRoot(cwd)
	if err != nil {
		return nil, errors.Wrap(err, "finding festivals root")
	}

	h := &HierarchySelector{
		config: config,
		state: SelectionState{
			FestivalsRoot: festivalsRoot,
		},
	}

	// Initialize state from CWD - errors are non-fatal
	_ = h.initFromContext(cwd)

	return h, nil
}

// initFromContext populates state from the current working directory.
func (h *HierarchySelector) initFromContext(cwd string) error {
	absCwd, err := filepath.Abs(cwd)
	if err != nil {
		return err
	}

	// Check if we're in a sequence directory
	if isSequenceDirPath(absCwd) {
		h.state.SequencePath = absCwd
		h.state.Sequence = filepath.Base(absCwd)

		// Sequence parent is phase
		phasePath := filepath.Dir(absCwd)
		h.state.PhasePath = phasePath
		h.state.Phase = filepath.Base(phasePath)

		// Phase parent is festival
		festivalPath := filepath.Dir(phasePath)
		h.state.FestivalPath = festivalPath
		h.state.Festival = h.deriveFestivalName(festivalPath)

		return nil
	}

	// Check if we're in a phase directory
	if isPhaseDirPath(absCwd) {
		h.state.PhasePath = absCwd
		h.state.Phase = filepath.Base(absCwd)

		// Phase parent is festival
		festivalPath := filepath.Dir(absCwd)
		h.state.FestivalPath = festivalPath
		h.state.Festival = h.deriveFestivalName(festivalPath)

		return nil
	}

	// Check if we're in a festival directory
	if looksLikeFestivalDir(absCwd) {
		h.state.FestivalPath = absCwd
		h.state.Festival = h.deriveFestivalName(absCwd)
		return nil
	}

	return errors.NotFound("festival context")
}

// deriveFestivalName extracts festival name relative to festivals root.
func (h *HierarchySelector) deriveFestivalName(festivalPath string) string {
	// Try to get relative path from festivals root
	rel, err := filepath.Rel(h.state.FestivalsRoot, festivalPath)
	if err == nil {
		return rel // e.g., "active/my-festival"
	}
	return filepath.Base(festivalPath)
}

// State returns the current selection state.
func (h *HierarchySelector) State() *SelectionState {
	return &h.state
}

// SetFestival sets the selected festival and updates the path.
func (h *HierarchySelector) SetFestival(festival string) {
	h.state.Festival = festival
	h.state.FestivalPath = filepath.Join(h.state.FestivalsRoot, festival)
}

// SetPhase sets the selected phase and updates the path.
func (h *HierarchySelector) SetPhase(phase string) {
	h.state.Phase = phase
	if h.state.FestivalPath != "" {
		h.state.PhasePath = filepath.Join(h.state.FestivalPath, phase)
	}
}

// SetSequence sets the selected sequence and updates the path.
func (h *HierarchySelector) SetSequence(sequence string) {
	h.state.Sequence = sequence
	if h.state.PhasePath != "" {
		h.state.SequencePath = filepath.Join(h.state.PhasePath, sequence)
	}
}

// ListFestivals returns available festivals.
func (h *HierarchySelector) ListFestivals(ctx context.Context) ([]LevelOption, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	options := []LevelOption{}

	// List festivals in standard locations (active, planned, etc.)
	locations := []string{"active", "planned"}
	if !h.config.FilterActive {
		locations = append(locations, "completed", "dungeon")
	}

	for _, loc := range locations {
		locPath := filepath.Join(h.state.FestivalsRoot, loc)
		entries, err := os.ReadDir(locPath)
		if err != nil {
			continue // Skip if directory doesn't exist
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			// Check if it looks like a festival (has FESTIVAL_GOAL.md or phase dirs)
			festPath := filepath.Join(locPath, entry.Name())
			if looksLikeFestivalDir(festPath) {
				relPath := filepath.Join(loc, entry.Name())
				options = append(options, LevelOption{
					Label: formatFestivalLabel(loc, entry.Name()),
					Value: relPath,
					Path:  festPath,
				})
			}
		}
	}

	// Sort by label
	sort.Slice(options, func(i, j int) bool {
		return options[i].Label < options[j].Label
	})

	return options, nil
}

// ListPhases returns phases for the selected festival.
func (h *HierarchySelector) ListPhases(ctx context.Context, festivalPath string) ([]LevelOption, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	phases, err := listPhaseInfos(festivalPath)
	if err != nil {
		return nil, err
	}

	options := make([]LevelOption, 0, len(phases))

	for _, p := range phases {
		// Format: [impl] 001 Implementation - Build the core features
		label := fmt.Sprintf("%s %s", phaseTypeIcon(p.Type), formatPhaseLabel(p.Name))
		if p.Goal != "" {
			label = fmt.Sprintf("%s - %s", label, p.Goal)
		}

		options = append(options, LevelOption{
			Label: label,
			Value: p.Name,
			Path:  p.Path,
		})
	}

	return options, nil
}

// ListSequences returns sequences for the selected phase.
func (h *HierarchySelector) ListSequences(ctx context.Context, phasePath string) ([]LevelOption, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	sequences, err := listSequenceInfos(phasePath)
	if err != nil {
		return nil, err
	}

	options := make([]LevelOption, 0, len(sequences))

	for _, s := range sequences {
		// Format: [todo] 01 Requirements (0/5 tasks) - Define requirements
		indicator := sequenceProgressIndicator(s.Completed, s.TaskCount)
		label := fmt.Sprintf("%s %s (%d/%d tasks)", indicator, formatSequenceLabel(s.Name), s.Completed, s.TaskCount)
		if s.Goal != "" {
			label = fmt.Sprintf("%s - %s", label, s.Goal)
		}

		options = append(options, LevelOption{
			Label: label,
			Value: s.Name,
			Path:  s.Path,
		})
	}

	return options, nil
}

// ListTasks returns tasks for the selected sequence.
func (h *HierarchySelector) ListTasks(ctx context.Context, sequencePath string) ([]LevelOption, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	entries, err := os.ReadDir(sequencePath)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`^[0-9]{2}_.+\.md$`)
	options := []LevelOption{}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if re.MatchString(name) {
			options = append(options, LevelOption{
				Label: formatTaskLabel(name),
				Value: name,
				Path:  filepath.Join(sequencePath, name),
			})
		}
	}

	return options, nil
}

// formatFestivalLabel creates a display label for a festival.
func formatFestivalLabel(location, name string) string {
	// Format: [active] my-festival-MF0001
	prefix := ""
	switch location {
	case "active":
		prefix = "[active] "
	case "planned":
		prefix = "[planned] "
	case "completed":
		prefix = "[done] "
	case "dungeon":
		prefix = "[archived] "
	}
	return prefix + name
}

// formatPhaseLabel creates a display label for a phase.
func formatPhaseLabel(phase string) string {
	// Convert "001_PLANNING" to "001 Planning"
	parts := strings.SplitN(phase, "_", 2)
	if len(parts) == 2 {
		name := strings.ReplaceAll(parts[1], "_", " ")
		name = strings.Title(strings.ToLower(name))
		return parts[0] + " " + name
	}
	return phase
}

// formatSequenceLabel creates a display label for a sequence.
func formatSequenceLabel(sequence string) string {
	// Convert "01_requirements" to "01 Requirements"
	parts := strings.SplitN(sequence, "_", 2)
	if len(parts) == 2 {
		name := strings.ReplaceAll(parts[1], "_", " ")
		name = strings.Title(strings.ToLower(name))
		return parts[0] + " " + name
	}
	return sequence
}

// formatTaskLabel creates a display label for a task.
func formatTaskLabel(task string) string {
	// Convert "01_design.md" to "01 Design"
	task = strings.TrimSuffix(task, ".md")
	parts := strings.SplitN(task, "_", 2)
	if len(parts) == 2 {
		name := strings.ReplaceAll(parts[1], "_", " ")
		name = strings.Title(strings.ToLower(name))
		return parts[0] + " " + name
	}
	return task
}

// Run executes the full hierarchical selection flow.
func (h *HierarchySelector) Run(ctx context.Context) (*SelectionState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Determine what we need to select based on config and pre-populated state
	needFestival := h.config.StartLevel <= LevelFestival && h.config.EndLevel >= LevelFestival && h.state.FestivalPath == ""
	needPhase := h.config.EndLevel >= LevelPhase && h.state.PhasePath == ""
	needSequence := h.config.EndLevel >= LevelSequence && h.state.SequencePath == ""

	// Step 1: Select festival (if needed)
	if needFestival {
		if err := h.selectFestival(ctx); err != nil {
			return nil, errors.Wrap(err, "selecting festival")
		}
	}

	// Step 2: Select phase (if needed and config allows)
	if h.config.EndLevel >= LevelPhase && needPhase {
		// Ensure we have a festival selected
		if h.state.FestivalPath == "" {
			return nil, errors.Validation("no festival selected for phase selection")
		}

		if err := h.selectPhase(ctx); err != nil {
			return nil, errors.Wrap(err, "selecting phase")
		}
	}

	// Step 3: Select sequence (if needed and config allows)
	if h.config.EndLevel >= LevelSequence && needSequence {
		// Ensure we have a phase selected
		if h.state.PhasePath == "" {
			return nil, errors.Validation("no phase selected for sequence selection")
		}

		if err := h.selectSequence(ctx); err != nil {
			return nil, errors.Wrap(err, "selecting sequence")
		}
	}

	return &h.state, nil
}

// SelectToFestival creates a config that only selects a festival.
func SelectToFestival(allowCreate bool) HierarchyConfig {
	return HierarchyConfig{
		StartLevel:  LevelFestival,
		EndLevel:    LevelFestival,
		AllowCreate: allowCreate,
	}
}

// SelectToPhase creates a config that selects festival and phase.
func SelectToPhase(allowCreate bool) HierarchyConfig {
	return HierarchyConfig{
		StartLevel:  LevelFestival,
		EndLevel:    LevelPhase,
		AllowCreate: allowCreate,
	}
}

// SelectToSequence creates a config that selects festival, phase, and sequence.
func SelectToSequence(allowCreate bool) HierarchyConfig {
	return HierarchyConfig{
		StartLevel:  LevelFestival,
		EndLevel:    LevelSequence,
		AllowCreate: allowCreate,
	}
}

// SelectPhaseToSequence creates a config starting from phase (skips festival).
func SelectPhaseToSequence(allowCreate bool) HierarchyConfig {
	return HierarchyConfig{
		StartLevel:  LevelPhase,
		EndLevel:    LevelSequence,
		AllowCreate: allowCreate,
	}
}

// GetPath returns the path at the specified level.
func (s *SelectionState) GetPath(level HierarchyLevel) string {
	switch level {
	case LevelFestival:
		return s.FestivalPath
	case LevelPhase:
		return s.PhasePath
	case LevelSequence:
		return s.SequencePath
	default:
		return ""
	}
}

// IsComplete checks if selection reached the expected level.
func (s *SelectionState) IsComplete(targetLevel HierarchyLevel) bool {
	return s.GetPath(targetLevel) != ""
}

// selectFestival displays the festival selection form.
func (h *HierarchySelector) selectFestival(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	options, err := h.ListFestivals(ctx)
	if err != nil {
		return err
	}

	if len(options) == 0 && !h.config.AllowCreate {
		return errors.NotFound("festivals").WithField("path", h.state.FestivalsRoot)
	}

	// Build huh options
	huhOpts := make([]huh.Option[string], 0, len(options)+1)
	for _, opt := range options {
		huhOpts = append(huhOpts, huh.NewOption(opt.Label, opt.Value))
	}

	// Add create option if allowed
	if h.config.AllowCreate {
		huhOpts = append(huhOpts, huh.NewOption("+ Create new festival...", "__create__"))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select Festival").
				Description("Use arrow keys to navigate, enter to select").
				Options(huhOpts...).
				Value(&selected),
		),
	)

	if err := uitheme.RunForm(ctx, form); err != nil {
		return err
	}

	if selected == "__create__" {
		// Trigger festival creation flow
		if err := charmCreateFestival(ctx); err != nil {
			return err
		}
		// After creation, re-run festival selection
		return h.selectFestival(ctx)
	}

	// Find the selected option and update state
	for _, opt := range options {
		if opt.Value == selected {
			h.state.Festival = opt.Value
			h.state.FestivalPath = opt.Path
			break
		}
	}

	return nil
}

// selectPhase displays the phase selection form.
func (h *HierarchySelector) selectPhase(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if h.state.FestivalPath == "" {
		return errors.Validation("festival must be selected first")
	}

	options, err := h.ListPhases(ctx, h.state.FestivalPath)
	if err != nil {
		return err
	}

	if len(options) == 0 && !h.config.AllowCreate {
		return errors.NotFound("phases").WithField("path", h.state.FestivalPath)
	}

	// Build huh options
	huhOpts := make([]huh.Option[string], 0, len(options)+1)
	for _, opt := range options {
		huhOpts = append(huhOpts, huh.NewOption(opt.Label, opt.Value))
	}

	// Add create option if allowed
	if h.config.AllowCreate {
		huhOpts = append(huhOpts, huh.NewOption("+ Create new phase...", "__create__"))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Select Phase in %s", filepath.Base(h.state.FestivalPath))).
				Description("Use arrow keys to navigate, enter to select").
				Options(huhOpts...).
				Value(&selected),
		),
	)

	if err := uitheme.RunForm(ctx, form); err != nil {
		return err
	}

	if selected == "__create__" {
		// Trigger phase creation flow
		if err := charmCreatePhase(ctx); err != nil {
			return err
		}
		// After creation, re-run phase selection
		return h.selectPhase(ctx)
	}

	// Find the selected option and update state
	for _, opt := range options {
		if opt.Value == selected {
			h.state.Phase = opt.Value
			h.state.PhasePath = opt.Path
			break
		}
	}

	return nil
}

// selectSequence displays the sequence selection form.
func (h *HierarchySelector) selectSequence(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if h.state.PhasePath == "" {
		return errors.Validation("phase must be selected first")
	}

	options, err := h.ListSequences(ctx, h.state.PhasePath)
	if err != nil {
		return err
	}

	if len(options) == 0 && !h.config.AllowCreate {
		return errors.NotFound("sequences").WithField("path", h.state.PhasePath)
	}

	// Build huh options
	huhOpts := make([]huh.Option[string], 0, len(options)+1)
	for _, opt := range options {
		huhOpts = append(huhOpts, huh.NewOption(opt.Label, opt.Value))
	}

	// Add create option if allowed
	if h.config.AllowCreate {
		huhOpts = append(huhOpts, huh.NewOption("+ Create new sequence...", "__create__"))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title(fmt.Sprintf("Select Sequence in %s", filepath.Base(h.state.PhasePath))).
				Description("Use arrow keys to navigate, enter to select").
				Options(huhOpts...).
				Value(&selected),
		),
	)

	if err := uitheme.RunForm(ctx, form); err != nil {
		return err
	}

	if selected == "__create__" {
		// Trigger sequence creation flow
		if err := charmCreateSequence(ctx); err != nil {
			return err
		}
		// After creation, re-run sequence selection
		return h.selectSequence(ctx)
	}

	// Find the selected option and update state
	for _, opt := range options {
		if opt.Value == selected {
			h.state.Sequence = opt.Value
			h.state.SequencePath = opt.Path
			break
		}
	}

	return nil
}
