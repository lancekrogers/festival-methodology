package festival

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// ElementType represents the type of festival element
type ElementType int

const (
	PhaseType ElementType = iota
	SequenceType
	TaskType
)

func (e ElementType) String() string {
	switch e {
	case PhaseType:
		return "phase"
	case SequenceType:
		return "sequence"
	case TaskType:
		return "task"
	default:
		return "unknown"
	}
}

// FestivalElement represents a numbered element in the festival structure
type FestivalElement struct {
	Type     ElementType
	Number   int
	Name     string
	Path     string
	FullName string // Original name with number prefix
	Children []FestivalElement
}

// Parser handles parsing of festival directory structures
type Parser struct {
	phasePattern    *regexp.Regexp
	sequencePattern *regexp.Regexp
	taskPattern     *regexp.Regexp
}

// NewParser creates a new festival parser
func NewParser() *Parser {
	return &Parser{
		phasePattern:    regexp.MustCompile(`^(\d{3})_(.+)$`),
		sequencePattern: regexp.MustCompile(`^(\d{2})_(.+)$`),
		taskPattern:     regexp.MustCompile(`^(\d{2})_(.+)\.md$`),
	}
}

// ParseFestival parses the entire festival structure
func (p *Parser) ParseFestival(ctx context.Context, festivalDir string) ([]FestivalElement, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("Parser.ParseFestival")
	}

	if !isDir(festivalDir) {
		return nil, errors.NotFound("festival directory").WithField("path", festivalDir)
	}

	phases, err := p.ParsePhases(ctx, festivalDir)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse phases").
			WithOp("Parser.ParseFestival").
			WithCode(errors.ErrCodeParse)
	}

	// Parse sequences within each phase
	for i := range phases {
		if ctx.Err() != nil {
			return nil, errors.Wrap(ctx.Err(), "context cancelled").WithOp("Parser.ParseFestival")
		}

		sequences, err := p.ParseSequences(ctx, phases[i].Path)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse sequences").
				WithOp("Parser.ParseFestival").
				WithField("phase", phases[i].Name).
				WithCode(errors.ErrCodeParse)
		}
		phases[i].Children = sequences

		// Parse tasks within each sequence
		for j := range phases[i].Children {
			if ctx.Err() != nil {
				return nil, errors.Wrap(ctx.Err(), "context cancelled").WithOp("Parser.ParseFestival")
			}

			tasks, err := p.ParseTasks(ctx, phases[i].Children[j].Path)
			if err != nil {
				return nil, errors.Wrap(err, "failed to parse tasks").
					WithOp("Parser.ParseFestival").
					WithField("sequence", phases[i].Children[j].Name).
					WithCode(errors.ErrCodeParse)
			}
			phases[i].Children[j].Children = tasks
		}
	}

	return phases, nil
}

// ParsePhases parses phase directories
func (p *Parser) ParsePhases(ctx context.Context, festivalDir string) ([]FestivalElement, error) {
	return p.parseElements(ctx, festivalDir, p.phasePattern, PhaseType, true)
}

// ParseSequences parses sequence directories within a phase
func (p *Parser) ParseSequences(ctx context.Context, phaseDir string) ([]FestivalElement, error) {
	return p.parseElements(ctx, phaseDir, p.sequencePattern, SequenceType, true)
}

// ParseTasks parses task files within a sequence
func (p *Parser) ParseTasks(ctx context.Context, sequenceDir string) ([]FestivalElement, error) {
	return p.parseElements(ctx, sequenceDir, p.taskPattern, TaskType, false)
}

// parseElements is the generic parser for numbered elements
func (p *Parser) parseElements(ctx context.Context, dir string, pattern *regexp.Regexp, elemType ElementType, isDirectory bool) ([]FestivalElement, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("Parser.parseElements")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, errors.IO("Parser.parseElements", err).WithField("dir", dir)
	}

	var elements []FestivalElement

	for _, entry := range entries {
		// Skip if type doesn't match
		if isDirectory && !entry.IsDir() {
			continue
		}
		if !isDirectory && entry.IsDir() {
			continue
		}

		// Check if name matches pattern
		matches := pattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}

		// Parse number
		num, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}

		// Create element
		element := FestivalElement{
			Type:     elemType,
			Number:   num,
			Name:     matches[2],
			Path:     filepath.Join(dir, entry.Name()),
			FullName: entry.Name(),
		}

		elements = append(elements, element)
	}

	// Sort by number
	sort.Slice(elements, func(i, j int) bool {
		return elements[i].Number < elements[j].Number
	})

	return elements, nil
}

// GetNextNumber returns the next available number for an element type
func (p *Parser) GetNextNumber(ctx context.Context, dir string, elemType ElementType) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, errors.Wrap(err, "context cancelled").WithOp("Parser.GetNextNumber")
	}

	var elements []FestivalElement
	var err error

	switch elemType {
	case PhaseType:
		elements, err = p.ParsePhases(ctx, dir)
	case SequenceType:
		elements, err = p.ParseSequences(ctx, dir)
	case TaskType:
		elements, err = p.ParseTasks(ctx, dir)
	default:
		return 0, errors.Validation("unknown element type").
			WithOp("Parser.GetNextNumber").
			WithField("type", elemType.String())
	}

	if err != nil {
		return 0, err
	}

	if len(elements) == 0 {
		return 1, nil
	}

	// Return the highest number + 1
	return elements[len(elements)-1].Number + 1, nil
}

// FindElement finds an element by number
func (p *Parser) FindElement(ctx context.Context, dir string, number int, elemType ElementType) (*FestivalElement, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("Parser.FindElement")
	}

	var elements []FestivalElement
	var err error

	switch elemType {
	case PhaseType:
		elements, err = p.ParsePhases(ctx, dir)
	case SequenceType:
		elements, err = p.ParseSequences(ctx, dir)
	case TaskType:
		elements, err = p.ParseTasks(ctx, dir)
	default:
		return nil, errors.Validation("unknown element type").
			WithOp("Parser.FindElement").
			WithField("type", elemType.String())
	}

	if err != nil {
		return nil, err
	}

	for _, elem := range elements {
		if elem.Number == number {
			return &elem, nil
		}
	}

	return nil, errors.NotFound(elemType.String()).
		WithField("number", number).
		WithField("dir", dir)
}

// FormatNumber formats a number based on element type
func FormatNumber(number int, elemType ElementType) string {
	switch elemType {
	case PhaseType:
		return fmt.Sprintf("%03d", number)
	case SequenceType, TaskType:
		return fmt.Sprintf("%02d", number)
	default:
		return strconv.Itoa(number)
	}
}

// BuildElementName constructs the full name with number prefix
// Normalizes name case based on element type: UPPERCASE for phases, lowercase for sequences/tasks
func BuildElementName(number int, name string, elemType ElementType) string {
	numStr := FormatNumber(number, elemType)

	// Normalize name based on element type
	normalized := normalizeElementName(name, elemType)

	return fmt.Sprintf("%s_%s", numStr, normalized)
}

func normalizeElementName(name string, elemType ElementType) string {
	trimmed := strings.TrimSpace(name)

	switch elemType {
	case PhaseType:
		trimmed = stripNumericPrefix(trimmed, 3)
		return strings.ToUpper(strings.ReplaceAll(trimmed, " ", "_"))
	case SequenceType, TaskType:
		trimmed = stripNumericPrefix(trimmed, 2)
		return strings.ToLower(strings.ReplaceAll(trimmed, " ", "_"))
	default:
		return strings.ReplaceAll(trimmed, " ", "_")
	}
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

// ParseElementName extracts number and name from a numbered element
func ParseElementName(fullName string, elemType ElementType) (int, string, error) {
	var pattern *regexp.Regexp

	switch elemType {
	case PhaseType:
		pattern = regexp.MustCompile(`^(\d{3})_(.+)$`)
	case SequenceType:
		pattern = regexp.MustCompile(`^(\d{2})_(.+)$`)
	case TaskType:
		pattern = regexp.MustCompile(`^(\d{2})_(.+)\.md$`)
	default:
		return 0, "", errors.Validation("unknown element type").
			WithOp("ParseElementName").
			WithField("type", elemType.String())
	}

	matches := pattern.FindStringSubmatch(fullName)
	if matches == nil {
		return 0, "", errors.Validation("name does not match pattern").
			WithOp("ParseElementName").
			WithField("name", fullName).
			WithField("type", elemType.String())
	}

	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", errors.Parse("failed to parse number", err).
			WithField("name", fullName)
	}

	return num, matches[2], nil
}

// HasParallelTasks checks if a sequence has parallel tasks (multiple tasks with same number)
func (p *Parser) HasParallelTasks(ctx context.Context, sequenceDir string) (map[int][]FestivalElement, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("Parser.HasParallelTasks")
	}

	tasks, err := p.ParseTasks(ctx, sequenceDir)
	if err != nil {
		return nil, err
	}

	parallel := make(map[int][]FestivalElement)
	for _, task := range tasks {
		parallel[task.Number] = append(parallel[task.Number], task)
	}

	// Remove entries with only one task
	for num, tasks := range parallel {
		if len(tasks) <= 1 {
			delete(parallel, num)
		}
	}

	return parallel, nil
}

// isDir checks if a path is a directory
func isDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// isFile checks if a path is a file
func isFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// NormalizeName removes number prefix from element name
func NormalizeName(fullName string) string {
	// Remove phase prefix (3 digits)
	if matches := regexp.MustCompile(`^\d{3}_(.+)$`).FindStringSubmatch(fullName); matches != nil {
		return matches[1]
	}

	// Remove sequence/task prefix (2 digits)
	if matches := regexp.MustCompile(`^\d{2}_(.+?)(?:\.md)?$`).FindStringSubmatch(fullName); matches != nil {
		return matches[1]
	}

	// Return as-is if no pattern matches
	return strings.TrimSuffix(fullName, ".md")
}

// IsPhase checks if a directory name matches the phase pattern (3-digit prefix)
func IsPhase(name string) bool {
	pattern := regexp.MustCompile(`^\d{3}_`)
	return pattern.MatchString(name)
}

// IsSequence checks if a directory name matches the sequence pattern (2-digit prefix)
func IsSequence(name string) bool {
	pattern := regexp.MustCompile(`^\d{2}_`)
	return pattern.MatchString(name)
}

// ParseTaskNumber extracts the task number from a task filename
func ParseTaskNumber(name string) int {
	pattern := regexp.MustCompile(`^(\d{2})_`)
	matches := pattern.FindStringSubmatch(name)
	if matches == nil {
		return 0
	}
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return num
}

// ParsePhaseNumber extracts the phase number from a phase directory name
func ParsePhaseNumber(name string) int {
	pattern := regexp.MustCompile(`^(\d{3})_`)
	matches := pattern.FindStringSubmatch(name)
	if matches == nil {
		return 0
	}
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return num
}

// ParseSequenceNumber extracts the sequence number from a sequence directory name
func ParseSequenceNumber(name string) int {
	pattern := regexp.MustCompile(`^(\d{2})_`)
	matches := pattern.FindStringSubmatch(name)
	if matches == nil {
		return 0
	}
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return num
}
