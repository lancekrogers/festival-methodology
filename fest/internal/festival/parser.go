package festival

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
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
func (p *Parser) ParseFestival(festivalDir string) ([]FestivalElement, error) {
	if !isDir(festivalDir) {
		return nil, fmt.Errorf("festival directory does not exist: %s", festivalDir)
	}
	
	phases, err := p.ParsePhases(festivalDir)
	if err != nil {
		return nil, fmt.Errorf("failed to parse phases: %w", err)
	}
	
	// Parse sequences within each phase
	for i := range phases {
		sequences, err := p.ParseSequences(phases[i].Path)
		if err != nil {
			return nil, fmt.Errorf("failed to parse sequences in %s: %w", phases[i].Name, err)
		}
		phases[i].Children = sequences
		
		// Parse tasks within each sequence
		for j := range phases[i].Children {
			tasks, err := p.ParseTasks(phases[i].Children[j].Path)
			if err != nil {
				return nil, fmt.Errorf("failed to parse tasks in %s: %w", phases[i].Children[j].Name, err)
			}
			phases[i].Children[j].Children = tasks
		}
	}
	
	return phases, nil
}

// ParsePhases parses phase directories
func (p *Parser) ParsePhases(festivalDir string) ([]FestivalElement, error) {
	return p.parseElements(festivalDir, p.phasePattern, PhaseType, true)
}

// ParseSequences parses sequence directories within a phase
func (p *Parser) ParseSequences(phaseDir string) ([]FestivalElement, error) {
	return p.parseElements(phaseDir, p.sequencePattern, SequenceType, true)
}

// ParseTasks parses task files within a sequence
func (p *Parser) ParseTasks(sequenceDir string) ([]FestivalElement, error) {
	return p.parseElements(sequenceDir, p.taskPattern, TaskType, false)
}

// parseElements is the generic parser for numbered elements
func (p *Parser) parseElements(dir string, pattern *regexp.Regexp, elemType ElementType, isDirectory bool) ([]FestivalElement, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory %s: %w", dir, err)
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
func (p *Parser) GetNextNumber(dir string, elemType ElementType) (int, error) {
	var elements []FestivalElement
	var err error
	
	switch elemType {
	case PhaseType:
		elements, err = p.ParsePhases(dir)
	case SequenceType:
		elements, err = p.ParseSequences(dir)
	case TaskType:
		elements, err = p.ParseTasks(dir)
	default:
		return 0, fmt.Errorf("unknown element type: %v", elemType)
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
func (p *Parser) FindElement(dir string, number int, elemType ElementType) (*FestivalElement, error) {
	var elements []FestivalElement
	var err error
	
	switch elemType {
	case PhaseType:
		elements, err = p.ParsePhases(dir)
	case SequenceType:
		elements, err = p.ParseSequences(dir)
	case TaskType:
		elements, err = p.ParseTasks(dir)
	default:
		return nil, fmt.Errorf("unknown element type: %v", elemType)
	}
	
	if err != nil {
		return nil, err
	}
	
	for _, elem := range elements {
		if elem.Number == number {
			return &elem, nil
		}
	}
	
	return nil, fmt.Errorf("%s %d not found", elemType, number)
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
func BuildElementName(number int, name string, elemType ElementType) string {
	numStr := FormatNumber(number, elemType)
	return fmt.Sprintf("%s_%s", numStr, name)
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
		return 0, "", fmt.Errorf("unknown element type: %v", elemType)
	}
	
	matches := pattern.FindStringSubmatch(fullName)
	if matches == nil {
		return 0, "", fmt.Errorf("name does not match %s pattern: %s", elemType, fullName)
	}
	
	num, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0, "", fmt.Errorf("failed to parse number: %w", err)
	}
	
	return num, matches[2], nil
}

// HasParallelTasks checks if a sequence has parallel tasks (multiple tasks with same number)
func (p *Parser) HasParallelTasks(sequenceDir string) (map[int][]FestivalElement, error) {
	tasks, err := p.ParseTasks(sequenceDir)
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