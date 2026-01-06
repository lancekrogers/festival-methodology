package template

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// Context holds all variables available for template rendering.
// It focuses exclusively on festival structure variables (festival/phase/sequence/task)
// and does NOT include time-based, user/account, or system variables.
type Context struct {
	// Festival-level variables
	FestivalName        string
	FestivalGoal        string
	FestivalTags        []string
	FestivalDescription string

	// Phase-level variables
	PhaseNumber    int
	PhaseName      string
	PhaseID        string // formatted: "001_PLANNING"
	PhaseType      string // "planning", "implementation", "review", "deployment", "research"
	PhaseStructure string // "freeform" or "structured"
	PhaseObjective string

	// Sequence-level variables
	SequenceNumber       int
	SequenceName         string
	SequenceID           string // formatted: "01_requirements"
	SequenceObjective    string
	SequenceDependencies []string

	// Task-level variables
	TaskNumber       int
	TaskName         string
	TaskID           string // formatted: "01_user_research.md"
	TaskObjective    string
	TaskDeliverables []string
	TaskParallel     bool
	TaskDependencies []string

	// Computed/structure variables
	CurrentLevel     string // "festival", "phase", "sequence", or "task"
	ParentPhaseID    string // e.g., "001_PLANNING"
	ParentSequenceID string // e.g., "01_requirements"
	FullPath         string // e.g., "001_PLANNING/01_requirements/01_task.md"
	FestivalRoot     string // relative path to festival root: ".", "..", "../.."

	// Custom user-provided variables (from TUI or CLI)
	Custom map[string]interface{}
}

// NewContext creates a new empty context
func NewContext() *Context {
	return &Context{
		Custom: make(map[string]interface{}),
	}
}

// SetFestival sets festival-level variables
func (c *Context) SetFestival(name, goal string, tags []string) {
	c.FestivalName = name
	c.FestivalGoal = goal
	c.FestivalTags = tags
	c.CurrentLevel = "festival"
}

// SetFestivalDescription sets the extended festival description
func (c *Context) SetFestivalDescription(description string) {
	c.FestivalDescription = description
}

// SetPhase sets phase-level variables
func (c *Context) SetPhase(number int, name, phaseType string) {
	c.PhaseNumber = number
	c.PhaseName = name
	c.PhaseType = phaseType
	c.PhaseID = FormatPhaseID(number, name)
	c.CurrentLevel = "phase"

	// Automatically set phase structure based on type
	if phaseType == "research" {
		c.PhaseStructure = "freeform"
	} else {
		c.PhaseStructure = "structured"
	}
}

// SetPhaseStructure sets whether the phase is freeform or structured
func (c *Context) SetPhaseStructure(structure string) {
	c.PhaseStructure = structure
}

// SetPhaseObjective sets the phase objective
func (c *Context) SetPhaseObjective(objective string) {
	c.PhaseObjective = objective
}

// SetSequence sets sequence-level variables
func (c *Context) SetSequence(number int, name string) {
	c.SequenceNumber = number
	c.SequenceName = name
	c.SequenceID = FormatSequenceID(number, name)
	c.CurrentLevel = "sequence"
}

// SetSequenceObjective sets the sequence objective
func (c *Context) SetSequenceObjective(objective string) {
	c.SequenceObjective = objective
}

// SetSequenceDependencies sets the sequence dependencies
func (c *Context) SetSequenceDependencies(dependencies []string) {
	c.SequenceDependencies = dependencies
}

// SetTask sets task-level variables
func (c *Context) SetTask(number int, name string) {
	c.TaskNumber = number
	c.TaskName = name
	c.TaskID = FormatTaskID(number, name)
	c.CurrentLevel = "task"
}

// SetTaskObjective sets the task objective
func (c *Context) SetTaskObjective(objective string) {
	c.TaskObjective = objective
}

// SetTaskDeliverables sets the task deliverables
func (c *Context) SetTaskDeliverables(deliverables []string) {
	c.TaskDeliverables = deliverables
}

// SetTaskParallel sets whether the task can run in parallel
func (c *Context) SetTaskParallel(parallel bool) {
	c.TaskParallel = parallel
}

// SetTaskDependencies sets the task dependencies
func (c *Context) SetTaskDependencies(dependencies []string) {
	c.TaskDependencies = dependencies
}

// SetCustom sets a custom user-provided variable
func (c *Context) SetCustom(key string, value interface{}) {
	c.Custom[key] = value
}

// ComputeStructureVariables calculates full_path, parent_*, and festival_root variables based on current level
func (c *Context) ComputeStructureVariables() {
	switch c.CurrentLevel {
	case "task":
		c.ParentPhaseID = c.PhaseID
		c.ParentSequenceID = c.SequenceID
		c.FullPath = filepath.Join(c.PhaseID, c.SequenceID, c.TaskID)
		c.FestivalRoot = "../.."
	case "sequence":
		c.ParentPhaseID = c.PhaseID
		c.FullPath = filepath.Join(c.PhaseID, c.SequenceID)
		c.FestivalRoot = "../.."
	case "phase":
		c.FullPath = c.PhaseID
		c.FestivalRoot = ".."
	case "festival":
		c.FullPath = ""
		c.FestivalRoot = "."
	}
}

// ToTemplateData converts Context to map[string]interface{} for Go templates
func (c *Context) ToTemplateData() map[string]interface{} {
	data := map[string]interface{}{
		// Festival-level
		"festival_name":        c.FestivalName,
		"festival_goal":        c.FestivalGoal,
		"festival_tags":        c.FestivalTags,
		"festival_description": c.FestivalDescription,

		// Phase-level
		"phase_number":    c.PhaseNumber,
		"phase_name":      c.PhaseName,
		"phase_id":        c.PhaseID,
		"phase_type":      c.PhaseType,
		"phase_structure": c.PhaseStructure,
		"phase_objective": c.PhaseObjective,

		// Sequence-level
		"sequence_number":       c.SequenceNumber,
		"sequence_name":         c.SequenceName,
		"sequence_id":           c.SequenceID,
		"sequence_objective":    c.SequenceObjective,
		"sequence_dependencies": c.SequenceDependencies,

		// Task-level
		"task_number":       c.TaskNumber,
		"task_name":         c.TaskName,
		"task_id":           c.TaskID,
		"task_objective":    c.TaskObjective,
		"task_deliverables": c.TaskDeliverables,
		"task_parallel":     c.TaskParallel,
		"task_dependencies": c.TaskDependencies,

		// Structure
		"current_level":      c.CurrentLevel,
		"parent_phase_id":    c.ParentPhaseID,
		"parent_sequence_id": c.ParentSequenceID,
		"full_path":          c.FullPath,
		"festival_root":      c.FestivalRoot,
	}

	// Merge custom variables
	for k, v := range c.Custom {
		data[k] = v
	}

	return data
}

// Get retrieves a variable value by key
func (c *Context) Get(key string) (interface{}, bool) {
	// Try custom variables first
	if val, ok := c.Custom[key]; ok {
		return val, true
	}

	// Convert to template data and look up
	data := c.ToTemplateData()
	val, ok := data[key]
	return val, ok
}

// FormatPhaseID creates formatted phase ID: "001_PLANNING"
func FormatPhaseID(number int, name string) string {
	// Normalize name: strip numeric prefix, uppercase, replace spaces with underscores
	normalized := normalizeName(name, 3, strings.ToUpper)
	return fmt.Sprintf("%03d_%s", number, normalized)
}

// FormatSequenceID creates formatted sequence ID: "01_requirements"
func FormatSequenceID(number int, name string) string {
	// Normalize name: strip numeric prefix, lowercase, replace spaces with underscores
	normalized := normalizeName(name, 2, strings.ToLower)
	return fmt.Sprintf("%02d_%s", number, normalized)
}

func normalizeName(name string, prefixDigits int, transform func(string) string) string {
	trimmed := strings.TrimSpace(name)
	trimmed = stripNumericPrefix(trimmed, prefixDigits)
	trimmed = strings.ReplaceAll(trimmed, " ", "_")
	return transform(trimmed)
}

func stripNumericPrefix(name string, digits int) string {
	if len(name) <= digits {
		return name
	}

	prefix := name[:digits]
	if _, err := strconv.Atoi(prefix); err != nil {
		return name
	}

	if len(name) == digits {
		return name
	}

	sep := name[digits]
	if sep != '_' && sep != '-' && sep != ' ' {
		return name
	}

	remainder := strings.TrimLeft(strings.TrimSpace(name[digits+1:]), "_- ")
	if remainder == "" {
		return name
	}

	return remainder
}

// FormatTaskID creates formatted task ID: "01_user_research.md"
func FormatTaskID(number int, name string) string {
	// Normalize name: lowercase, replace spaces with underscores
	normalized := strings.ToLower(strings.ReplaceAll(name, " ", "_"))
	return fmt.Sprintf("%02d_%s.md", number, normalized)
}
