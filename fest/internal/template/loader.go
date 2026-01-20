package template

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
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
	Load(ctx context.Context, path string) (*Template, error)
	LoadAll(ctx context.Context, dir string) ([]*Template, error)
}

type loaderImpl struct{}

// NewLoader creates a new template loader
func NewLoader() Loader {
	return &loaderImpl{}
}

// Load loads a single template file
func (l *loaderImpl) Load(ctx context.Context, path string) (*Template, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("Loader.Load")
	}

	// Read file
	file, err := os.Open(path)
	if err != nil {
		return nil, errors.IO("opening template file", err).
			WithField("path", path)
	}
	defer file.Close()

	// Parse frontmatter and content. If YAML frontmatter parsing fails,
	// fall back to treating the entire file as content without metadata.
	metadata, content, err := parseFrontmatter(file)
	if err != nil {
		// Tolerant fallback for non‑YAML frontmatter: load full file content
		b, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil, errors.IO("reading template file", err).
				WithField("path", path)
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
func (l *loaderImpl) LoadAll(ctx context.Context, dir string) ([]*Template, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("Loader.LoadAll")
	}

	templates := []*Template{}

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return ctxErr
		}

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
		tmpl, err := l.Load(ctx, path)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Warning: failed to load template %s: %v\n", path, err)
			return nil
		}

		templates = append(templates, tmpl)
		return nil
	})

	if err != nil {
		return nil, errors.IO("walking template directory", err).
			WithField("path", dir)
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
		return nil, "", errors.IO("reading file", err)
	}

	// Parse frontmatter YAML if present
	var metadata *Metadata
	if len(frontmatterLines) > 0 {
		metadata = &Metadata{}
		yamlContent := strings.Join(frontmatterLines, "\n")
		if err := yaml.Unmarshal([]byte(yamlContent), metadata); err != nil {
			return nil, "", errors.Parse("parsing YAML frontmatter", err)
		}
	}

	// Join content lines
	content := strings.Join(contentLines, "\n")

	// Trim leading newlines (common when templates have blank lines after metadata frontmatter)
	content = strings.TrimLeft(content, "\n")

	return metadata, content, nil
}
