package template

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Metadata represents the YAML frontmatter of a template
type Metadata struct {
	TemplateID        string   `yaml:"template_id"`
	ID                string   `yaml:"id"`
	Aliases           []string `yaml:"aliases"`
	TemplateVersion   string   `yaml:"template_version"`
	RequiredVariables []string `yaml:"required_variables"`
	OptionalVariables []string `yaml:"optional_variables"`
	Description       string   `yaml:"description"`
}

// Template represents a loaded template with its metadata and content
type Template struct {
	Path     string
	Metadata *Metadata
	Content  string // Raw template content (without frontmatter)
}

// Loader loads templates from the filesystem
type Loader interface {
	Load(path string) (*Template, error)
	LoadAll(dir string) ([]*Template, error)
}

type loaderImpl struct{}

// NewLoader creates a new template loader
func NewLoader() Loader {
	return &loaderImpl{}
}

// Load loads a single template file
func (l *loaderImpl) Load(path string) (*Template, error) {
	// Read file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open template file %s: %w", path, err)
	}
	defer file.Close()

	// Parse frontmatter and content. If YAML frontmatter parsing fails,
	// fall back to treating the entire file as content without metadata.
	metadata, content, err := parseFrontmatter(file)
	if err != nil {
		// Tolerant fallback for non‑YAML frontmatter: load full file content
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil, fmt.Errorf("failed to read template %s: %w", path, err)
		}
		return &Template{Path: path, Metadata: nil, Content: string(b)}, nil
	}

	t := &Template{
		Path:     path,
		Metadata: metadata,
		Content:  content,
	}

	// Normalize metadata: prefer template_id, fallback to id
	if t.Metadata != nil {
		if t.Metadata.TemplateID == "" && t.Metadata.ID != "" {
			t.Metadata.TemplateID = t.Metadata.ID
		}
	}

	return t, nil
}

// LoadAll loads all markdown templates from a directory
func (l *loaderImpl) LoadAll(dir string) ([]*Template, error) {
	templates := []*Template{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Only process markdown files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Load template
		tmpl, err := l.Load(path)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to load template %s: %v\n", path, err)
			return nil
		}

		templates = append(templates, tmpl)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory %s: %w", dir, err)
	}

	return templates, nil
}

// parseFrontmatter parses YAML frontmatter from a markdown file
func parseFrontmatter(file *os.File) (*Metadata, string, error) {
	scanner := bufio.NewScanner(file)
	var frontmatterLines []string
	var contentLines []string
	inFrontmatter := false
	frontmatterFound := false

	for scanner.Scan() {
		raw := scanner.Text()
		// Trim BOM (if present) and surrounding whitespace for delimiter checks
		line := strings.TrimSpace(strings.TrimPrefix(raw, "\uFEFF"))

		// Check for frontmatter delimiter
		if line == "---" {
			if !frontmatterFound {
				// First delimiter - start of frontmatter
				inFrontmatter = true
				frontmatterFound = true
				continue
			} else if inFrontmatter {
				// Second delimiter - end of frontmatter
				inFrontmatter = false
				continue
			}
		}

		// Collect lines
		if inFrontmatter {
			// Use the raw line content for YAML (preserve spacing)
			frontmatterLines = append(frontmatterLines, raw)
		} else if frontmatterFound {
			// After frontmatter ends, collect content
			contentLines = append(contentLines, raw)
		} else {
			// No frontmatter - treat as content
			contentLines = append(contentLines, raw)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, "", fmt.Errorf("error reading file: %w", err)
	}

	// Parse frontmatter YAML if present
	var metadata *Metadata
	if len(frontmatterLines) > 0 {
		metadata = &Metadata{}
		yamlContent := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), metadata); err != nil {
			return nil, "", fmt.Errorf("failed to parse YAML frontmatter: %w", err)
		}
	}

	// Join content lines
	content := strings.Join(contentLines, "\n")

	return metadata, content, nil
}
