package template

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// AgentTemplatesDir is the directory name for agent-created templates
const AgentTemplatesDir = ".templates"

// AgentTemplate represents a simple agent-created template
type AgentTemplate struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
	Variables   []string `json:"variables"` // Extracted {{variable}} names
	Content     string   `json:"-"`         // Template content
}

// AgentTemplateStore manages agent templates within a festival
type AgentTemplateStore struct {
	festivalPath string
	templatesDir string
}

// NewAgentTemplateStore creates a store for the given festival path
func NewAgentTemplateStore(festivalPath string) *AgentTemplateStore {
	return &AgentTemplateStore{
		festivalPath: festivalPath,
		templatesDir: filepath.Join(festivalPath, AgentTemplatesDir),
	}
}

// EnsureDir creates the templates directory if it doesn't exist
func (s *AgentTemplateStore) EnsureDir() error {
	if err := os.MkdirAll(s.templatesDir, 0755); err != nil {
		return errors.IO("creating templates directory", err).
			WithField("path", s.templatesDir)
	}
	return nil
}

// List returns all agent templates in the festival
func (s *AgentTemplateStore) List(ctx context.Context) ([]*AgentTemplate, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("AgentTemplateStore.List")
	}

	entries, err := os.ReadDir(s.templatesDir)
	if os.IsNotExist(err) {
		return []*AgentTemplate{}, nil
	}
	if err != nil {
		return nil, errors.IO("reading templates directory", err).
			WithField("path", s.templatesDir)
	}

	var templates []*AgentTemplate
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".md")
		tmpl, err := s.Load(ctx, name)
		if err != nil {
			continue // Skip invalid templates
		}
		templates = append(templates, tmpl)
	}

	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})

	return templates, nil
}

// Load loads a specific template by name
func (s *AgentTemplateStore) Load(ctx context.Context, name string) (*AgentTemplate, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("AgentTemplateStore.Load")
	}

	path := filepath.Join(s.templatesDir, name+".md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, errors.NotFound("template").WithField("name", name)
	}
	if err != nil {
		return nil, errors.IO("reading template file", err).WithField("path", path)
	}

	return &AgentTemplate{
		Name:      name,
		Path:      path,
		Variables: ExtractVariables(string(content)),
		Content:   string(content),
	}, nil
}

// Save saves a template with the given name and content
func (s *AgentTemplateStore) Save(ctx context.Context, name, content string) (*AgentTemplate, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("AgentTemplateStore.Save")
	}

	if err := s.EnsureDir(); err != nil {
		return nil, err
	}

	path := filepath.Join(s.templatesDir, name+".md")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return nil, errors.IO("writing template file", err).WithField("path", path)
	}

	return &AgentTemplate{
		Name:      name,
		Path:      path,
		Variables: ExtractVariables(content),
		Content:   content,
	}, nil
}

// Delete removes a template by name
func (s *AgentTemplateStore) Delete(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("AgentTemplateStore.Delete")
	}

	path := filepath.Join(s.templatesDir, name+".md")
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			return errors.NotFound("template").WithField("name", name)
		}
		return errors.IO("deleting template file", err).WithField("path", path)
	}
	return nil
}

// variableRegex matches {{variable_name}} patterns
var variableRegex = regexp.MustCompile(`\{\{([a-zA-Z_][a-zA-Z0-9_]*)\}\}`)

// ExtractVariables extracts all {{variable}} names from content
func ExtractVariables(content string) []string {
	matches := variableRegex.FindAllStringSubmatch(content, -1)
	seen := make(map[string]bool)
	var vars []string

	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			seen[match[1]] = true
			vars = append(vars, match[1])
		}
	}

	sort.Strings(vars)
	return vars
}

// ApplyVariables substitutes {{variable}} placeholders with values
func ApplyVariables(content string, vars map[string]string) string {
	result := content
	for name, value := range vars {
		placeholder := "{{" + name + "}}"
		result = strings.ReplaceAll(result, placeholder, value)
	}
	return result
}

// ParseVariablesJSON parses a JSON string or @file path into variable map
func ParseVariablesJSON(input string) (map[string]string, error) {
	var data []byte
	var err error

	if strings.HasPrefix(input, "@") {
		// Read from file
		filePath := strings.TrimPrefix(input, "@")
		data, err = os.ReadFile(filePath)
		if err != nil {
			return nil, errors.IO("reading variables file", err).
				WithField("path", filePath)
		}
	} else {
		data = []byte(input)
	}

	var vars map[string]string
	if err := json.Unmarshal(data, &vars); err != nil {
		return nil, errors.Parse("parsing variables JSON", err)
	}

	return vars, nil
}
