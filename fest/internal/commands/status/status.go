// Package status implements the fest status command for managing entity statuses.
package status

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/progress"
	"github.com/spf13/cobra"
)

// EntityType represents the type of festival entity.
type EntityType string

const (
	EntityFestival EntityType = "festival"
	EntityPhase    EntityType = "phase"
	EntitySequence EntityType = "sequence"
	EntityTask     EntityType = "task"
	EntityGate     EntityType = "gate"
)

// ValidStatuses defines valid status values per entity type.
var ValidStatuses = map[EntityType][]string{
	EntityFestival: {"planned", "active", "completed", "dungeon"},
	EntityPhase:    {"pending", "in_progress", "completed"},
	EntitySequence: {"pending", "in_progress", "completed"},
	EntityTask:     {"pending", "in_progress", "blocked", "completed"},
	EntityGate:     {"pending", "passed", "failed"},
}

type statusOptions struct {
	json       bool
	entityType string
	force      bool
	path       string
}

// NewStatusCommand creates the status command with all subcommands.
func NewStatusCommand() *cobra.Command {
	opts := &statusOptions{}

	cmd := &cobra.Command{
		Use:   "status [path]",
		Short: "Manage and query festival entity statuses",
		Long: `Manage and query status for festivals, phases, sequences, tasks, and gates.

When run without arguments, shows the status of the current entity based on
your working directory location.

EXAMPLES:
  fest status                                  # Status from current directory
  fest status ./festivals/active/my-festival   # Status for specific path
  fest status active/my-festival               # Relative to festivals/ root

SUBCOMMANDS:
  fest status              Show current entity status
  fest status set <status> Change entity status
  fest status list         List entities by status
  fest status history      View status change history`,
		Args: cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				opts.path = args[0]
			}
			return runStatusShow(cmd, opts)
		},
	}

	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "", "entity type to query (festival, phase, sequence, task, gate)")

	// Add subcommands
	cmd.AddCommand(newStatusSetCommand(opts))
	cmd.AddCommand(newStatusListCommand(opts))
	cmd.AddCommand(newStatusHistoryCommand(opts))

	return cmd
}

func runStatusShow(cmd *cobra.Command, opts *statusOptions) error {
	// Resolve target path
	targetPath, err := resolveStatusPath(opts.path)
	if err != nil {
		if opts.json {
			return emitErrorJSON(err.Error())
		}
		return err
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(targetPath)
	if err != nil {
		if opts.json {
			return emitErrorJSON("not in a festival directory")
		}
		return err
	}

	if opts.json {
		return emitLocationJSON(loc)
	}
	return emitLocationText(loc)
}

// resolveStatusPath resolves the target path for status commands.
// If pathArg is empty, uses current working directory.
// If pathArg is relative to a festivals/ root (e.g., "active/my-festival"),
// it resolves from the festivals root.
func resolveStatusPath(pathArg string) (string, error) {
	if pathArg == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", errors.IO("getting current directory", err)
		}
		return cwd, nil
	}

	// Try as absolute or relative path first
	absPath, err := filepath.Abs(pathArg)
	if err != nil {
		return "", errors.Wrap(err, "resolving path").WithField("path", pathArg)
	}

	// Check if path exists
	if _, err := os.Stat(absPath); err == nil {
		return absPath, nil
	}

	// Try resolving relative to festivals/ root
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.IO("getting current directory", err)
	}

	// Find festivals root and try pathArg relative to it
	festivalsRoot := findFestivalsRoot(cwd)
	if festivalsRoot != "" {
		candidatePath := filepath.Join(festivalsRoot, pathArg)
		if _, err := os.Stat(candidatePath); err == nil {
			return candidatePath, nil
		}
	}

	return "", errors.NotFound("path").WithField("path", pathArg)
}

// findFestivalsRoot walks up from startPath looking for a festivals/ directory
func findFestivalsRoot(startPath string) string {
	current := startPath
	for {
		// Check if current is festivals/ or contains festivals/
		if filepath.Base(current) == "festivals" {
			return current
		}
		festivalsDir := filepath.Join(current, "festivals")
		if info, err := os.Stat(festivalsDir); err == nil && info.IsDir() {
			return festivalsDir
		}

		parent := filepath.Dir(current)
		if parent == current {
			break
		}
		current = parent
	}
	return ""
}

func emitErrorJSON(message string) error {
	result := map[string]interface{}{
		"error": message,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

func emitLocationJSON(loc *show.LocationInfo) error {
	result := map[string]interface{}{
		"type": loc.Type,
	}

	if loc.Festival != nil {
		result["festival"] = map[string]interface{}{
			"name":   loc.Festival.Name,
			"status": loc.Festival.Status,
			"path":   loc.Festival.Path,
		}
		if loc.Festival.Stats != nil {
			result["progress"] = loc.Festival.Stats.Progress
		}
	}

	if loc.Phase != "" {
		result["phase"] = loc.Phase
	}
	if loc.Sequence != "" {
		result["sequence"] = loc.Sequence
	}
	if loc.Task != "" {
		result["task"] = loc.Task
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling status to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitLocationText(loc *show.LocationInfo) error {
	if loc.Festival == nil {
		fmt.Println("Not in a festival directory")
		return nil
	}

	fmt.Printf("Festival: %s\n", loc.Festival.Name)
	fmt.Printf("Status:   %s\n", loc.Festival.Status)
	fmt.Printf("Location: %s\n", loc.Type)

	if loc.Phase != "" {
		fmt.Printf("Phase:    %s\n", loc.Phase)
	}
	if loc.Sequence != "" {
		fmt.Printf("Sequence: %s\n", loc.Sequence)
	}
	if loc.Task != "" {
		fmt.Printf("Task:     %s\n", loc.Task)
	}

	if loc.Festival.Stats != nil {
		fmt.Printf("\nProgress: %.1f%%\n", loc.Festival.Stats.Progress)
		fmt.Printf("  Phases:    %d/%d completed\n",
			loc.Festival.Stats.Phases.Completed,
			loc.Festival.Stats.Phases.Total)
		fmt.Printf("  Sequences: %d/%d completed\n",
			loc.Festival.Stats.Sequences.Completed,
			loc.Festival.Stats.Sequences.Total)
		fmt.Printf("  Tasks:     %d/%d completed\n",
			loc.Festival.Stats.Tasks.Completed,
			loc.Festival.Stats.Tasks.Total)
	}

	return nil
}

func newStatusSetCommand(opts *statusOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <status>",
		Short: "Change entity status",
		Long: `Change the status of the current entity.

For festivals, this will move the directory between status folders
(planned, active, completed, dungeon).

For other entities, this updates the frontmatter in the relevant files.`,
		Example: `  fest status set active               # Set current festival to active
  fest status set completed --force    # Set without confirmation
  fest status set in_progress          # Set phase/sequence/task status`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusSet(cmd, args[0], opts)
		},
	}

	cmd.Flags().BoolVar(&opts.force, "force", false, "skip confirmation prompts")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

func runStatusSet(cmd *cobra.Command, newStatus string, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// Determine entity type and validate status
	entityType := EntityType(loc.Type)
	if !isValidStatus(entityType, newStatus) {
		validOptions := ValidStatuses[entityType]
		return errors.Validation("invalid status").
			WithField("status", newStatus).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Handle festival status changes (directory moves)
	if loc.Type == "festival" {
		return handleFestivalStatusChange(loc.Festival, newStatus, opts)
	}

	// For other entities, update frontmatter (placeholder - needs frontmatter support)
	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"entity":     loc.Type,
			"new_status": newStatus,
			"note":       "frontmatter updates not yet implemented",
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Status change recorded: %s -> %s\n", loc.Type, newStatus)
		fmt.Println("Note: Frontmatter updates pending implementation")
	}

	return nil
}

func handleFestivalStatusChange(festival *show.FestivalInfo, newStatus string, opts *statusOptions) error {
	if festival.Status == newStatus {
		if opts.json {
			result := map[string]interface{}{
				"success": true,
				"message": "already at requested status",
				"status":  newStatus,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Printf("Festival is already %s\n", newStatus)
		}
		return nil
	}

	// Confirm unless forced
	if !opts.force {
		fmt.Printf("Move festival '%s' from %s to %s? [y/N]: ", festival.Name, festival.Status, newStatus)
		var response string
		fmt.Scanln(&response)
		if !strings.HasPrefix(strings.ToLower(response), "y") {
			fmt.Println("Operation cancelled.")
			return nil
		}
	}

	// Calculate new path
	festivalsRoot := filepath.Dir(filepath.Dir(festival.Path))
	newPath := filepath.Join(festivalsRoot, newStatus, festival.Name)

	// Check if destination exists
	if _, err := os.Stat(newPath); err == nil {
		return errors.Validation("destination already exists").WithField("path", newPath)
	}

	// Create destination directory if needed
	destDir := filepath.Join(festivalsRoot, newStatus)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return errors.IO("creating destination directory", err)
	}

	// Move the directory
	if err := os.Rename(festival.Path, newPath); err != nil {
		return errors.IO("moving festival directory", err)
	}

	if opts.json {
		result := map[string]interface{}{
			"success":    true,
			"festival":   festival.Name,
			"old_status": festival.Status,
			"new_status": newStatus,
			"old_path":   festival.Path,
			"new_path":   newPath,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Moved festival '%s' from %s to %s\n", festival.Name, festival.Status, newStatus)
		fmt.Printf("New path: %s\n", newPath)
	}

	return nil
}

func isValidStatus(entityType EntityType, status string) bool {
	validStatuses, ok := ValidStatuses[entityType]
	if !ok {
		return false
	}
	for _, valid := range validStatuses {
		if valid == status {
			return true
		}
	}
	return false
}

// PhaseInfo holds information about a phase.
type PhaseInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	Status    string       `json:"status"`
	TaskStats StatusCounts `json:"task_stats,omitempty"`
}

// SequenceInfo holds information about a sequence.
type SequenceInfo struct {
	Name      string       `json:"name"`
	Path      string       `json:"path"`
	PhaseName string       `json:"phase_name"`
	Status    string       `json:"status"`
	TaskStats StatusCounts `json:"task_stats,omitempty"`
}

// TaskInfo holds information about a task.
type TaskInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	PhaseName    string `json:"phase_name"`
	SequenceName string `json:"sequence_name"`
	Status       string `json:"status"`
}

// StatusCounts tracks entity completion.
type StatusCounts struct {
	Total      int `json:"total"`
	Completed  int `json:"completed"`
	InProgress int `json:"in_progress"`
	Pending    int `json:"pending"`
	Blocked    int `json:"blocked,omitempty"`
}

// collectPhases collects all phases from a festival directory.
func collectPhases(festivalPath string) ([]*PhaseInfo, error) {
	var phases []*PhaseInfo

	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		phaseDir := filepath.Join(festivalPath, entry.Name())

		// Check if it's a phase directory (numeric prefix)
		if !hasNumericPrefix(entry.Name()) {
			continue
		}

		phase, err := collectPhaseInfo(phaseDir, entry.Name())
		if err != nil {
			// Skip phases that can't be parsed
			continue
		}

		phases = append(phases, phase)
	}

	return phases, nil
}

// collectPhaseInfo collects information about a single phase.
func collectPhaseInfo(phasePath, phaseName string) (*PhaseInfo, error) {
	// Calculate phase stats using show package
	festStats, err := show.CalculateFestivalStats(phasePath)
	var taskStats StatusCounts
	if err == nil && festStats != nil {
		taskStats = StatusCounts{
			Total:      festStats.Tasks.Total,
			Completed:  festStats.Tasks.Completed,
			InProgress: festStats.Tasks.InProgress,
			Pending:    festStats.Tasks.Pending,
			Blocked:    festStats.Tasks.Blocked,
		}
	}

	// Determine phase status
	status := "pending"
	if taskStats.Total > 0 {
		if taskStats.Completed == taskStats.Total {
			status = "completed"
		} else if taskStats.InProgress > 0 || taskStats.Completed > 0 {
			status = "in_progress"
		}
	}

	return &PhaseInfo{
		Name:      phaseName,
		Path:      phasePath,
		Status:    status,
		TaskStats: taskStats,
	}, nil
}

// collectSequencesFromFestival collects all sequences across all phases in a festival.
func collectSequencesFromFestival(festivalPath string) ([]*SequenceInfo, error) {
	var allSequences []*SequenceInfo

	entries, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	for _, entry := range entries {
		if !entry.IsDir() || !hasNumericPrefix(entry.Name()) {
			continue
		}

		phaseDir := filepath.Join(festivalPath, entry.Name())
		sequences, err := collectSequences(phaseDir, entry.Name())
		if err != nil {
			continue
		}

		allSequences = append(allSequences, sequences...)
	}

	return allSequences, nil
}

// collectSequences collects all sequences from a phase directory.
func collectSequences(phasePath, phaseName string) ([]*SequenceInfo, error) {
	var sequences []*SequenceInfo

	entries, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, errors.IO("reading phase directory", err).WithField("path", phasePath)
	}

	for _, entry := range entries {
		if !entry.IsDir() || !hasNumericPrefix(entry.Name()) {
			continue
		}

		seqDir := filepath.Join(phasePath, entry.Name())
		seq, err := collectSequenceInfo(seqDir, phaseName, entry.Name())
		if err != nil {
			continue
		}

		sequences = append(sequences, seq)
	}

	return sequences, nil
}

// collectSequenceInfo collects information about a single sequence.
func collectSequenceInfo(seqPath, phaseName, seqName string) (*SequenceInfo, error) {
	// Count tasks in sequence
	taskStats, err := countSequenceTasks(seqPath)
	if err != nil {
		taskStats = StatusCounts{}
	}

	// Determine sequence status
	status := "pending"
	if taskStats.Total > 0 {
		if taskStats.Completed == taskStats.Total {
			status = "completed"
		} else if taskStats.InProgress > 0 || taskStats.Completed > 0 {
			status = "in_progress"
		}
	}

	return &SequenceInfo{
		Name:      seqName,
		Path:      seqPath,
		PhaseName: phaseName,
		Status:    status,
		TaskStats: taskStats,
	}, nil
}

// countSequenceTasks counts tasks in a sequence directory.
func countSequenceTasks(seqDir string) (StatusCounts, error) {
	counts := StatusCounts{}

	entries, err := os.ReadDir(seqDir)
	if err != nil {
		return counts, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := entry.Name()
		// Skip goal files and gate files
		if name == "SEQUENCE_GOAL.md" || name == "PHASE_GOAL.md" ||
		   name == "FESTIVAL_GOAL.md" || name == "FESTIVAL_OVERVIEW.md" ||
		   strings.Contains(strings.ToLower(name), "gate") {
			continue
		}

		counts.Total++
		taskPath := filepath.Join(seqDir, name)
		status := progress.ParseTaskStatus(taskPath)

		switch status {
		case "completed":
			counts.Completed++
		case "in_progress":
			counts.InProgress++
		case "blocked":
			counts.Blocked++
		default:
			counts.Pending++
		}
	}

	return counts, nil
}

// collectTasksFromFestival collects all tasks across all sequences in a festival.
func collectTasksFromFestival(festivalPath string) ([]*TaskInfo, error) {
	var allTasks []*TaskInfo

	phases, err := os.ReadDir(festivalPath)
	if err != nil {
		return nil, errors.IO("reading festival directory", err).WithField("path", festivalPath)
	}

	for _, phaseEntry := range phases {
		if !phaseEntry.IsDir() || !hasNumericPrefix(phaseEntry.Name()) {
			continue
		}

		phaseDir := filepath.Join(festivalPath, phaseEntry.Name())
		tasks, err := collectTasksFromPhase(phaseDir, phaseEntry.Name())
		if err != nil {
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// collectTasksFromPhase collects all tasks from all sequences in a phase.
func collectTasksFromPhase(phasePath, phaseName string) ([]*TaskInfo, error) {
	var allTasks []*TaskInfo

	sequences, err := os.ReadDir(phasePath)
	if err != nil {
		return nil, errors.IO("reading phase directory", err).WithField("path", phasePath)
	}

	for _, seqEntry := range sequences {
		if !seqEntry.IsDir() || !hasNumericPrefix(seqEntry.Name()) {
			continue
		}

		seqDir := filepath.Join(phasePath, seqEntry.Name())
		tasks, err := collectTasks(seqDir, phaseName, seqEntry.Name())
		if err != nil {
			continue
		}

		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

// collectTasks collects all tasks from a sequence directory.
func collectTasks(seqPath, phaseName, seqName string) ([]*TaskInfo, error) {
	var tasks []*TaskInfo

	entries, err := os.ReadDir(seqPath)
	if err != nil {
		return nil, errors.IO("reading sequence directory", err).WithField("path", seqPath)
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := entry.Name()
		// Skip goal files and gate files
		if name == "SEQUENCE_GOAL.md" || name == "PHASE_GOAL.md" ||
		   name == "FESTIVAL_GOAL.md" || name == "FESTIVAL_OVERVIEW.md" ||
		   strings.Contains(strings.ToLower(name), "gate") {
			continue
		}

		taskPath := filepath.Join(seqPath, name)
		status := progress.ParseTaskStatus(taskPath)

		tasks = append(tasks, &TaskInfo{
			Name:         strings.TrimSuffix(name, ".md"),
			Path:         taskPath,
			PhaseName:    phaseName,
			SequenceName: seqName,
			Status:       status,
		})
	}

	return tasks, nil
}

// hasNumericPrefix checks if a directory name starts with digits.
func hasNumericPrefix(name string) bool {
	if len(name) == 0 {
		return false
	}
	return name[0] >= '0' && name[0] <= '9'
}

// Filter functions

func filterPhasesByStatus(phases []*PhaseInfo, status string) []*PhaseInfo {
	if status == "" {
		return phases
	}

	var filtered []*PhaseInfo
	for _, phase := range phases {
		if phase.Status == status {
			filtered = append(filtered, phase)
		}
	}
	return filtered
}

func filterSequencesByStatus(sequences []*SequenceInfo, status string) []*SequenceInfo {
	if status == "" {
		return sequences
	}

	var filtered []*SequenceInfo
	for _, seq := range sequences {
		if seq.Status == status {
			filtered = append(filtered, seq)
		}
	}
	return filtered
}

func filterTasksByStatus(tasks []*TaskInfo, status string) []*TaskInfo {
	if status == "" {
		return tasks
	}

	var filtered []*TaskInfo
	for _, task := range tasks {
		if task.Status == status {
			filtered = append(filtered, task)
		}
	}
	return filtered
}

// Presentation functions

func emitPhasesJSON(phases []*PhaseInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":  "phase",
		"count": len(phases),
		"phases": phases,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling phases to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitPhasesText(phases []*PhaseInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Phases with status '%s' (%d)\n", filterStatus, len(phases))
	} else {
		fmt.Printf("Phases (%d)\n", len(phases))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, phase := range phases {
		fmt.Printf("  %s [%s]", phase.Name, phase.Status)
		if phase.TaskStats.Total > 0 {
			fmt.Printf(" (%d/%d tasks)", phase.TaskStats.Completed, phase.TaskStats.Total)
		}
		fmt.Println()
	}

	return nil
}

func emitSequencesJSON(sequences []*SequenceInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":      "sequence",
		"count":     len(sequences),
		"sequences": sequences,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling sequences to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitSequencesText(sequences []*SequenceInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Sequences with status '%s' (%d)\n", filterStatus, len(sequences))
	} else {
		fmt.Printf("Sequences (%d)\n", len(sequences))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, seq := range sequences {
		fmt.Printf("  %s/%s [%s]", seq.PhaseName, seq.Name, seq.Status)
		if seq.TaskStats.Total > 0 {
			fmt.Printf(" (%d/%d tasks)", seq.TaskStats.Completed, seq.TaskStats.Total)
		}
		fmt.Println()
	}

	return nil
}

func emitTasksJSON(tasks []*TaskInfo, filterStatus string) error {
	result := map[string]interface{}{
		"type":  "task",
		"count": len(tasks),
		"tasks": tasks,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
	}

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return errors.Wrap(err, "marshaling tasks to JSON")
	}
	fmt.Println(string(data))
	return nil
}

func emitTasksText(tasks []*TaskInfo, filterStatus string) error {
	if filterStatus != "" {
		fmt.Printf("Tasks with status '%s' (%d)\n", filterStatus, len(tasks))
	} else {
		fmt.Printf("Tasks (%d)\n", len(tasks))
	}
	fmt.Println(strings.Repeat("─", 60))

	for _, task := range tasks {
		fmt.Printf("  %s/%s/%s [%s]\n",
			task.PhaseName,
			task.SequenceName,
			task.Name,
			task.Status)
	}

	return nil
}

func emitEmptyJSON(entityType, filterStatus string) error {
	result := map[string]interface{}{
		"type":  entityType,
		"count": 0,
	}
	if filterStatus != "" {
		result["status"] = filterStatus
		result["message"] = fmt.Sprintf("no %ss found with status '%s'", entityType, filterStatus)
	} else {
		result["message"] = fmt.Sprintf("no %ss found", entityType)
	}

	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}

// Routing functions

func runFestivalListing(festivalsRoot, filterStatus string, opts *statusOptions) error {
	if filterStatus != "" {
		festivals, err := show.ListFestivalsByStatus(festivalsRoot, filterStatus)
		if err != nil {
			return err
		}

		if opts.json {
			result := map[string]interface{}{
				"status":    filterStatus,
				"count":     len(festivals),
				"festivals": festivals,
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println(show.FormatFestivalList(filterStatus, festivals))
		}
	} else {
		fmt.Println("Use 'fest show all' to see all festivals grouped by status")
	}

	return nil
}

func runPhaseListing(loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	var phases []*PhaseInfo
	var err error

	// Collect phases based on current location
	if loc.Type == "festival" {
		// List all phases in festival
		phases, err = collectPhases(loc.Festival.Path)
	} else {
		// In phase/sequence/task - list just current phase
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		phase, err := collectPhaseInfo(phasePath, loc.Phase)
		if err != nil {
			return err
		}
		phases = []*PhaseInfo{phase}
	}

	if err != nil {
		return err
	}

	// Filter by status
	phases = filterPhasesByStatus(phases, filterStatus)

	// Handle empty results
	if len(phases) == 0 {
		if opts.json {
			return emitEmptyJSON("phase", filterStatus)
		}
		fmt.Printf("No phases found")
		if filterStatus != "" {
			fmt.Printf(" with status '%s'", filterStatus)
		}
		fmt.Println()
		return nil
	}

	// Emit output
	if opts.json {
		return emitPhasesJSON(phases, filterStatus)
	}
	return emitPhasesText(phases, filterStatus)
}

func runSequenceListing(loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	var sequences []*SequenceInfo
	var err error

	// Collect sequences based on current location
	switch loc.Type {
	case "festival":
		// List all sequences in festival
		sequences, err = collectSequencesFromFestival(loc.Festival.Path)
	case "phase":
		// List sequences in current phase
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		sequences, err = collectSequences(phasePath, loc.Phase)
	default:
		// In sequence or task - list current sequence
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		seqPath := filepath.Join(phasePath, loc.Sequence)
		seq, err := collectSequenceInfo(seqPath, loc.Phase, loc.Sequence)
		if err != nil {
			return err
		}
		sequences = []*SequenceInfo{seq}
	}

	if err != nil {
		return err
	}

	// Filter by status
	sequences = filterSequencesByStatus(sequences, filterStatus)

	// Handle empty results
	if len(sequences) == 0 {
		if opts.json {
			return emitEmptyJSON("sequence", filterStatus)
		}
		fmt.Printf("No sequences found")
		if filterStatus != "" {
			fmt.Printf(" with status '%s'", filterStatus)
		}
		fmt.Println()
		return nil
	}

	// Emit output
	if opts.json {
		return emitSequencesJSON(sequences, filterStatus)
	}
	return emitSequencesText(sequences, filterStatus)
}

func runTaskListing(loc *show.LocationInfo, filterStatus string, opts *statusOptions) error {
	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	var tasks []*TaskInfo
	var err error

	// Collect tasks based on current location
	switch loc.Type {
	case "festival":
		// List all tasks in festival
		tasks, err = collectTasksFromFestival(loc.Festival.Path)
	case "phase":
		// List tasks in current phase
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		tasks, err = collectTasksFromPhase(phasePath, loc.Phase)
	case "sequence":
		// List tasks in current sequence
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		seqPath := filepath.Join(phasePath, loc.Sequence)
		tasks, err = collectTasks(seqPath, loc.Phase, loc.Sequence)
	default:
		// In task - could list siblings or just the current task
		phasePath := filepath.Join(loc.Festival.Path, loc.Phase)
		seqPath := filepath.Join(phasePath, loc.Sequence)
		tasks, err = collectTasks(seqPath, loc.Phase, loc.Sequence)
	}

	if err != nil {
		return err
	}

	// Filter by status
	tasks = filterTasksByStatus(tasks, filterStatus)

	// Handle empty results
	if len(tasks) == 0 {
		if opts.json {
			return emitEmptyJSON("task", filterStatus)
		}
		fmt.Printf("No tasks found")
		if filterStatus != "" {
			fmt.Printf(" with status '%s'", filterStatus)
		}
		fmt.Println()
		return nil
	}

	// Emit output
	if opts.json {
		return emitTasksJSON(tasks, filterStatus)
	}
	return emitTasksText(tasks, filterStatus)
}

func newStatusListCommand(opts *statusOptions) *cobra.Command {
	var filterStatus string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entities by status",
		Long: `List festivals, phases, sequences, or tasks filtered by status.

Without filters, lists all festivals grouped by status.`,
		Example: `  fest status list                     # List all festivals by status
  fest status list --status active     # List active festivals only
  fest status list --type task --status pending  # List pending tasks`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusList(cmd, filterStatus, opts)
		},
	}

	cmd.Flags().StringVar(&filterStatus, "status", "", "filter by status")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")
	cmd.Flags().StringVar(&opts.entityType, "type", "festival", "entity type (festival, phase, sequence, task)")

	return cmd
}

func runStatusList(cmd *cobra.Command, filterStatus string, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	// Detect current location
	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		// Not in a festival - handle festival listing or error
		festivalsDir := findFestivalsRoot(cwd)
		if festivalsDir == "" {
			return errors.NotFound("festival or festivals directory").
				WithField("hint", "navigate to a festival directory to list phases/sequences/tasks")
		}
		if opts.entityType == "festival" || opts.entityType == "" {
			return runFestivalListing(festivalsDir, filterStatus, opts)
		}
		return errors.NotFound("festival").
			WithField("hint", "navigate to a festival directory to list phases/sequences/tasks")
	}

	// Validate entity type and status
	entityType := EntityType(opts.entityType)
	if filterStatus != "" && !isValidStatus(entityType, filterStatus) {
		validOptions := ValidStatuses[entityType]
		return errors.Validation("invalid status for entity type").
			WithField("status", filterStatus).
			WithField("type", opts.entityType).
			WithField("valid_options", strings.Join(validOptions, ", "))
	}

	// Route based on entity type
	switch opts.entityType {
	case "festival", "":
		festivalsRoot := filepath.Dir(filepath.Dir(loc.Festival.Path))
		return runFestivalListing(festivalsRoot, filterStatus, opts)

	case "phase":
		return runPhaseListing(loc, filterStatus, opts)

	case "sequence":
		return runSequenceListing(loc, filterStatus, opts)

	case "task":
		return runTaskListing(loc, filterStatus, opts)

	default:
		return errors.Validation("invalid entity type").
			WithField("type", opts.entityType).
			WithField("valid_types", "festival, phase, sequence, task")
	}
}

func newStatusHistoryCommand(opts *statusOptions) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "history",
		Short: "View status change history",
		Long: `View the history of status changes for the current entity.

History is stored in .fest/status_history.json within each festival.`,
		Example: `  fest status history            # Show status history
  fest status history --limit 10 # Show last 10 changes`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatusHistory(cmd, limit, opts)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "maximum number of entries to show")
	cmd.Flags().BoolVar(&opts.json, "json", false, "output in JSON format")

	return cmd
}

func runStatusHistory(cmd *cobra.Command, limit int, opts *statusOptions) error {
	cwd, err := os.Getwd()
	if err != nil {
		return errors.IO("getting current directory", err)
	}

	loc, err := show.DetectCurrentLocation(cwd)
	if err != nil {
		return err
	}

	if loc.Festival == nil {
		return errors.NotFound("festival")
	}

	// History file path
	historyPath := filepath.Join(loc.Festival.Path, ".fest", "status_history.json")

	// Check if history exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		if opts.json {
			result := map[string]interface{}{
				"history": []interface{}{},
				"message": "no status history found",
			}
			data, _ := json.MarshalIndent(result, "", "  ")
			fmt.Println(string(data))
		} else {
			fmt.Println("No status history found for this festival.")
		}
		return nil
	}

	// Read history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return errors.IO("reading history file", err)
	}

	var history []map[string]interface{}
	if err := json.Unmarshal(data, &history); err != nil {
		return errors.Wrap(err, "parsing history file")
	}

	// Apply limit
	if limit > 0 && len(history) > limit {
		history = history[len(history)-limit:]
	}

	if opts.json {
		result := map[string]interface{}{
			"festival": loc.Festival.Name,
			"count":    len(history),
			"history":  history,
		}
		data, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("Status History for %s:\n", loc.Festival.Name)
		fmt.Println(strings.Repeat("-", 50))
		for _, entry := range history {
			fmt.Printf("%s: %s -> %s\n",
				entry["timestamp"],
				entry["from_status"],
				entry["to_status"])
			if note, ok := entry["note"].(string); ok && note != "" {
				fmt.Printf("  Note: %s\n", note)
			}
		}
	}

	return nil
}
