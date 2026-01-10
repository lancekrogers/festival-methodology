package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"github.com/lancekrogers/festival-methodology/fest/internal/festival"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

func collectRequiredVars(ctx context.Context, templateRoot string, paths []string) []string {
	loader := tpl.NewLoader()
	vars := []string{}
	for _, p := range paths {
		if ctx.Err() != nil {
			break
		}
		if _, err := os.Stat(p); err != nil {
			continue
		}
		t, err := loader.Load(ctx, p)
		if err != nil || t.Metadata == nil {
			continue
		}
		vars = append(vars, t.Metadata.RequiredVariables...)
	}
	return vars
}

func uniqueStrings(in []string) []string {
	m := map[string]struct{}{}
	out := []string{}
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := m[s]; !ok {
			m[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

func writeTempVarsFile(vars map[string]interface{}) (string, error) {
	if len(vars) == 0 {
		return "", nil
	}
	f, err := os.CreateTemp("", "fest-vars-*.json")
	if err != nil {
		return "", errors.IO("creating temp vars file", err)
	}
	enc := json.NewEncoder(f)
	if err := enc.Encode(vars); err != nil {
		_ = f.Close()
		return "", errors.IO("writing temp vars file", err)
	}
	_ = f.Close()
	return f.Name(), nil
}

func atoiDefault(s string, def int) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return def
	}
	return n
}

func defaultFestivalTemplatePaths(tmplRoot string) []string {
	return []string{
		filepath.Join(tmplRoot, "FESTIVAL_OVERVIEW_TEMPLATE.md"),
		filepath.Join(tmplRoot, "FESTIVAL_GOAL_TEMPLATE.md"),
		filepath.Join(tmplRoot, "FESTIVAL_RULES_TEMPLATE.md"),
		filepath.Join(tmplRoot, "FESTIVAL_TODO_TEMPLATE.md"),
	}
}

// slugify mirrors the create_festival.go behavior
func slugify(s string) string {
	lower := strings.ToLower(strings.TrimSpace(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(lower, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = "festival"
	}
	return slug
}

// resolvePhaseDirInput attempts to resolve a user-provided phase path shortcut
// like "02" or "002" or a phase name to a concrete directory path.
// It searches under the detected festival directory near cwd.
func resolvePhaseDirInput(input, cwd string) (string, error) {
	input = strings.TrimSpace(input)
	absCwd, _ := filepath.Abs(cwd)
	festivalDir := findFestivalDir(absCwd)

	// If direct path exists (relative or absolute), use it
	if input == "" || input == "." {
		if isPhaseDirPath(absCwd) {
			return absCwd, nil
		}
		// No specific phase; default to CWD if it looks like a festival directory
		return "", errors.Validation("please specify a phase (e.g., 002_IMPLEMENT or 02)")
	}
	if filepath.IsAbs(input) {
		if info, err := os.Stat(input); err == nil && info.IsDir() {
			return input, nil
		}
	} else {
		try := filepath.Join(absCwd, input)
		if info, err := os.Stat(try); err == nil && info.IsDir() {
			return try, nil
		}
	}

	// Collect phase dirs under festivalDir
	phases := listPhaseDirs(festivalDir)
	if len(phases) == 0 {
		return "", errors.NotFound("phase directories").WithField("path", festivalDir)
	}

	// If numeric, pad to 3 digits and match prefix
	if isDigits(input) {
		n, _ := strconv.Atoi(input)
		code := fmt.Sprintf("%03d", n)
		for _, name := range phases {
			if strings.HasPrefix(name, code+"_") || name == code {
				return filepath.Join(festivalDir, name), nil
			}
		}
		return "", errors.NotFound("phase").WithField("code", code).WithField("path", festivalDir)
	}

	// Match by name suffix after underscore (case-insensitive)
	needle := strings.ToLower(input)
	for _, name := range phases {
		if name == input {
			return filepath.Join(festivalDir, name), nil
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) == 2 {
			if strings.ToLower(parts[1]) == needle {
				return filepath.Join(festivalDir, name), nil
			}
		}
	}

	return "", errors.NotFound("phase").WithField("input", input)
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// findFestivalDir attempts to find the nearest festival directory from cwd.
// If cwd is a phase dir (NNN_NAME), returns its parent. Otherwise, returns cwd
// if it looks like a festival root (has phase dirs or key festival files). Fallback: cwd.
func findFestivalDir(cwd string) string {
	if isPhaseDirPath(cwd) {
		return filepath.Dir(cwd)
	}
	if looksLikeFestivalDir(cwd) {
		return cwd
	}
	// Fallback one level up
	parent := filepath.Dir(cwd)
	if looksLikeFestivalDir(parent) {
		return parent
	}
	return cwd
}

func looksLikeFestivalDir(dir string) bool {
	// If typical files exist or numbered phase dirs present, assume festival dir
	if exists(filepath.Join(dir, "FESTIVAL_OVERVIEW.md")) || exists(filepath.Join(dir, "FESTIVAL_GOAL.md")) {
		return true
	}
	return len(listPhaseDirs(dir)) > 0
}

func isPhaseDirPath(path string) bool {
	base := filepath.Base(path)
	re := regexp.MustCompile(`^[0-9]{3}_.+`)
	return re.MatchString(base)
}

func isSequenceDirPath(path string) bool {
	base := filepath.Base(path)
	re := regexp.MustCompile(`^[0-9]{2}_.+`)
	return re.MatchString(base)
}

func listPhaseDirs(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	re := regexp.MustCompile(`^[0-9]{3}_.+`)
	out := []string{}
	for _, e := range entries {
		if e.IsDir() && re.MatchString(e.Name()) {
			out = append(out, e.Name())
		}
	}
	return out
}

func listSequenceDirs(phaseDir string) []string {
	entries, err := os.ReadDir(phaseDir)
	if err != nil {
		return nil
	}
	re := regexp.MustCompile(`^[0-9]{2}_.+`)
	out := []string{}
	for _, e := range entries {
		if e.IsDir() && re.MatchString(e.Name()) {
			out = append(out, e.Name())
		}
	}
	return out
}

// resolveSequenceDirInput resolves user input like "01" or a sequence name to a real sequence directory
// relative to the nearest phase directory around cwd. If input is "." and cwd is a sequence dir, returns cwd.
func resolveSequenceDirInput(input, cwd string) (string, error) {
	input = strings.TrimSpace(input)
	absCwd, _ := filepath.Abs(cwd)

	if input == "" || input == "." {
		if isSequenceDirPath(absCwd) {
			return absCwd, nil
		}
		return "", errors.Validation("please specify a sequence (e.g., 01_requirements or 01)")
	}

	// If a direct path exists
	if filepath.IsAbs(input) {
		if info, err := os.Stat(input); err == nil && info.IsDir() {
			return input, nil
		}
	} else {
		candidate := filepath.Join(absCwd, input)
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			return candidate, nil
		}
	}

	// Determine base phase directory to search
	var phaseDir string
	switch {
	case isPhaseDirPath(absCwd):
		phaseDir = absCwd
	case isSequenceDirPath(absCwd):
		phaseDir = filepath.Dir(absCwd)
	default:
		phaseDir = findFestivalDir(absCwd) // best-effort fallback
	}

	sequences := listSequenceDirs(phaseDir)
	if len(sequences) == 0 {
		return "", errors.NotFound("sequence directories").WithField("path", phaseDir)
	}

	if isDigits(input) {
		n, _ := strconv.Atoi(input)
		code := fmt.Sprintf("%02d", n)
		for _, name := range sequences {
			if strings.HasPrefix(name, code+"_") || name == code {
				return filepath.Join(phaseDir, name), nil
			}
		}
		return "", errors.NotFound("sequence").WithField("code", code).WithField("path", phaseDir)
	}

	needle := strings.ToLower(input)
	for _, name := range sequences {
		if name == input {
			return filepath.Join(phaseDir, name), nil
		}
		parts := strings.SplitN(name, "_", 2)
		if len(parts) == 2 {
			if strings.ToLower(parts[1]) == needle {
				return filepath.Join(phaseDir, name), nil
			}
		}
	}

	return "", errors.NotFound("sequence").WithField("input", input)
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// nextPhaseAfter returns the number to use for --after when inserting a new phase
// so that the new phase is appended at the end. If no phases exist, returns 0.
func nextPhaseAfter(ctx context.Context, festivalDir string) int {
	parser := festival.NewParser()
	phases, err := parser.ParsePhases(ctx, festivalDir)
	if err != nil || len(phases) == 0 {
		return 0
	}
	max := 0
	for _, p := range phases {
		if p.Number > max {
			max = p.Number
		}
	}
	return max
}

// nextSequenceAfter returns the number to use for --after when inserting a sequence in a phase
func nextSequenceAfter(ctx context.Context, phaseDir string) int {
	parser := festival.NewParser()
	seqs, err := parser.ParseSequences(ctx, phaseDir)
	if err != nil || len(seqs) == 0 {
		return 0
	}
	max := 0
	for _, s := range seqs {
		if s.Number > max {
			max = s.Number
		}
	}
	return max
}

// nextTaskAfter returns the number to use for --after when inserting a task in a sequence
func nextTaskAfter(ctx context.Context, seqDir string) int {
	parser := festival.NewParser()
	tasks, err := parser.ParseTasks(ctx, seqDir)
	if err != nil || len(tasks) == 0 {
		return 0
	}
	max := 0
	for _, t := range tasks {
		if t.Number > max {
			max = t.Number
		}
	}
	return max
}

// FestivalInfo holds information about a discovered festival.
type FestivalInfo struct {
	Name   string // Directory name
	Path   string // Full path
	Status string // active, planned, completed, dungeon
	Goal   string // Extracted from FESTIVAL_GOAL.md if available
}

// listFestivalsInDir lists all festivals in a status directory.
func listFestivalsInDir(statusDir string) ([]FestivalInfo, error) {
	entries, err := os.ReadDir(statusDir)
	if err != nil {
		return nil, err
	}

	var festivals []FestivalInfo
	status := filepath.Base(statusDir)

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		festPath := filepath.Join(statusDir, e.Name())

		// Verify it's a festival (has FESTIVAL_GOAL.md or similar)
		if !isFestivalDir(festPath) {
			continue
		}

		info := FestivalInfo{
			Name:   e.Name(),
			Path:   festPath,
			Status: status,
			Goal:   extractFestivalGoal(festPath),
		}
		festivals = append(festivals, info)
	}

	return festivals, nil
}

// isFestivalDir checks if a directory is a festival.
func isFestivalDir(dir string) bool {
	markers := []string{
		"FESTIVAL_GOAL.md",
		"FESTIVAL_OVERVIEW.md",
		"FESTIVAL_RULES.md",
		"fest.yaml",
	}
	for _, m := range markers {
		if exists(filepath.Join(dir, m)) {
			return true
		}
	}
	// Also check for phase directories
	return len(listPhaseDirs(dir)) > 0
}

// extractFestivalGoal extracts the goal from FESTIVAL_GOAL.md.
func extractFestivalGoal(festDir string) string {
	goalPath := filepath.Join(festDir, "FESTIVAL_GOAL.md")
	data, err := os.ReadFile(goalPath)
	if err != nil {
		return ""
	}

	// Extract first non-empty, non-header line as goal summary
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// Truncate long goals
		if len(line) > 60 {
			return line[:57] + "..."
		}
		return line
	}
	return ""
}

// listAllFestivals lists festivals across all status directories.
func listAllFestivals(festivalsRoot string) (map[string][]FestivalInfo, error) {
	result := make(map[string][]FestivalInfo)

	statuses := []string{"active", "planned", "completed", "dungeon"}
	for _, status := range statuses {
		statusDir := filepath.Join(festivalsRoot, status)
		if !exists(statusDir) {
			continue
		}

		festivals, err := listFestivalsInDir(statusDir)
		if err != nil {
			continue // Skip dirs we can't read
		}

		if len(festivals) > 0 {
			result[status] = festivals
		}
	}

	return result, nil
}

// PhaseInfo holds information about a discovered phase.
type PhaseInfo struct {
	Number    int
	Name      string // Full directory name (e.g., "001_PLANNING")
	ShortName string // Name without number (e.g., "PLANNING")
	Path      string
	Type      string // planning, implementation, research, review, deployment
	Goal      string // Extracted from PHASE_GOAL.md
}

// getPhaseType determines the phase type from its name.
func getPhaseType(phaseName string) string {
	lower := strings.ToLower(phaseName)

	switch {
	case strings.Contains(lower, "planning") || strings.Contains(lower, "plan"):
		return "planning"
	case strings.Contains(lower, "research") || strings.Contains(lower, "design"):
		return "research"
	case strings.Contains(lower, "implementation") || strings.Contains(lower, "implement"):
		return "implementation"
	case strings.Contains(lower, "review") || strings.Contains(lower, "qa"):
		return "review"
	case strings.Contains(lower, "deployment") || strings.Contains(lower, "deploy"):
		return "deployment"
	default:
		return "implementation" // Default
	}
}

// listPhaseInfos returns detailed info about phases in a festival.
func listPhaseInfos(festivalDir string) ([]PhaseInfo, error) {
	phaseNames := listPhaseDirs(festivalDir)
	if len(phaseNames) == 0 {
		return nil, nil
	}

	phases := make([]PhaseInfo, 0, len(phaseNames))

	for _, name := range phaseNames {
		phasePath := filepath.Join(festivalDir, name)

		// Extract number and short name
		parts := strings.SplitN(name, "_", 2)
		number := 0
		shortName := name
		if len(parts) == 2 {
			number, _ = strconv.Atoi(parts[0])
			shortName = parts[1]
		}

		info := PhaseInfo{
			Number:    number,
			Name:      name,
			ShortName: shortName,
			Path:      phasePath,
			Type:      getPhaseType(name),
			Goal:      extractPhaseGoal(phasePath),
		}
		phases = append(phases, info)
	}

	return phases, nil
}

// extractPhaseGoal extracts the goal from PHASE_GOAL.md.
func extractPhaseGoal(phaseDir string) string {
	goalPath := filepath.Join(phaseDir, "PHASE_GOAL.md")
	data, err := os.ReadFile(goalPath)
	if err != nil {
		return ""
	}

	// Look for "Primary Goal:" line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "**Primary Goal:**") {
			goal := strings.TrimPrefix(line, "**Primary Goal:**")
			goal = strings.TrimSpace(goal)
			if len(goal) > 50 {
				return goal[:47] + "..."
			}
			return goal
		}
	}

	return ""
}

// phaseTypeIcon returns an icon for the phase type.
func phaseTypeIcon(phaseType string) string {
	switch phaseType {
	case "planning":
		return "[plan]"
	case "research":
		return "[research]"
	case "implementation":
		return "[impl]"
	case "review":
		return "[review]"
	case "deployment":
		return "[deploy]"
	default:
		return "[phase]"
	}
}

// SequenceInfo holds information about a discovered sequence.
type SequenceInfo struct {
	Number    int
	Name      string // Full directory name (e.g., "01_requirements")
	ShortName string // Name without number (e.g., "requirements")
	Path      string
	TaskCount int
	Completed int    // Number of completed tasks
	Goal      string // Extracted from SEQUENCE_GOAL.md
}

// listSequenceInfos returns detailed info about sequences in a phase.
func listSequenceInfos(phaseDir string) ([]SequenceInfo, error) {
	seqNames := listSequenceDirs(phaseDir)
	if len(seqNames) == 0 {
		return nil, nil
	}

	sequences := make([]SequenceInfo, 0, len(seqNames))

	for _, name := range seqNames {
		seqPath := filepath.Join(phaseDir, name)

		// Extract number and short name
		parts := strings.SplitN(name, "_", 2)
		number := 0
		shortName := name
		if len(parts) == 2 {
			number, _ = strconv.Atoi(parts[0])
			shortName = parts[1]
		}

		// Count tasks
		taskCount, completed := countTasks(seqPath)

		info := SequenceInfo{
			Number:    number,
			Name:      name,
			ShortName: shortName,
			Path:      seqPath,
			TaskCount: taskCount,
			Completed: completed,
			Goal:      extractSequenceGoal(seqPath),
		}
		sequences = append(sequences, info)
	}

	return sequences, nil
}

// countTasks counts task files in a sequence directory.
func countTasks(seqDir string) (total, completed int) {
	entries, err := os.ReadDir(seqDir)
	if err != nil {
		return 0, 0
	}

	taskRe := regexp.MustCompile(`^[0-9]{2}_.*\.md$`)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !taskRe.MatchString(e.Name()) {
			continue
		}
		// Skip SEQUENCE_GOAL.md
		if e.Name() == "SEQUENCE_GOAL.md" {
			continue
		}

		total++

		// Check if completed (look for completion markers in file)
		taskPath := filepath.Join(seqDir, e.Name())
		if isTaskCompleted(taskPath) {
			completed++
		}
	}

	return total, completed
}

// isTaskCompleted checks if a task file is marked as completed.
func isTaskCompleted(taskPath string) bool {
	data, err := os.ReadFile(taskPath)
	if err != nil {
		return false
	}

	content := string(data)
	// Check for common completion patterns
	if strings.Contains(content, "**Status:** Completed") ||
		strings.Contains(content, "Status: completed") {
		return true
	}

	// Count checkboxes
	unchecked := strings.Count(content, "- [ ]")
	checked := strings.Count(content, "- [x]") + strings.Count(content, "- [X]")

	// Consider complete if all checkboxes are checked and there are some
	if checked > 0 && unchecked == 0 {
		return true
	}

	return false
}

// extractSequenceGoal extracts the goal from SEQUENCE_GOAL.md.
func extractSequenceGoal(seqDir string) string {
	goalPath := filepath.Join(seqDir, "SEQUENCE_GOAL.md")
	data, err := os.ReadFile(goalPath)
	if err != nil {
		return ""
	}

	// Look for "Primary Goal:" line
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "**Primary Goal:**") {
			goal := strings.TrimPrefix(line, "**Primary Goal:**")
			goal = strings.TrimSpace(goal)
			if len(goal) > 40 {
				return goal[:37] + "..."
			}
			return goal
		}
	}

	return ""
}

// sequenceProgressIndicator returns a text indicator based on completion.
func sequenceProgressIndicator(completed, total int) string {
	if total == 0 {
		return "[empty]"
	}

	percent := float64(completed) / float64(total) * 100
	switch {
	case percent == 100:
		return "[done]"
	case percent >= 50:
		return "[prog]"
	case percent > 0:
		return "[start]"
	default:
		return "[todo]"
	}
}
