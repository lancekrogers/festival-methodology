package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

// Manager is the high-level API for template operations
type Manager struct {
	loader   Loader
	renderer Renderer
}

// NewManager creates a new template manager
func NewManager() *Manager {
	return &Manager{
		loader:   NewLoader(),
		renderer: NewRenderer(),
	}
}

// Render renders an already-loaded template using the manager's renderer
func (m *Manager) Render(t *Template, ctx *Context) (string, error) {
	return m.renderer.Render(t, ctx)
}

// RenderFile renders a template file with the given context
func (m *Manager) RenderFile(ctx context.Context, templatePath string, tmplCtx *Context) (string, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return "", fmt.Errorf("context cancelled: %w", err)
	}

	// Load template
	tmpl, err := m.loader.Load(ctx, templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template: %w", err)
	}

	// Validate required variables
	if err := ValidateTemplate(tmpl, tmplCtx); err != nil {
		return "", fmt.Errorf("template validation failed: %w", err)
	}

	// Render template
	output, err := m.renderer.Render(tmpl, tmplCtx)
	if err != nil {
		return "", fmt.Errorf("failed to render template: %w", err)
	}

	// Check for unrendered variables
	unrendered := CheckUnrenderedVariables(output)
	if len(unrendered) > 0 {
		return "", fmt.Errorf("template has unrendered variables: %v", unrendered)
	}

	return output, nil
}

// RenderFileToFile renders a template and writes it to an output file
func (m *Manager) RenderFileToFile(ctx context.Context, templatePath, outputPath string, tmplCtx *Context) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Render template
	output, err := m.RenderFile(ctx, templatePath, tmplCtx)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// RenderDirectory renders all templates in a directory to an output directory
func (m *Manager) RenderDirectory(ctx context.Context, templateDir, outputDir string, tmplCtx *Context) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	// Load all templates
	templates, err := m.loader.LoadAll(ctx, templateDir)
	if err != nil {
		return fmt.Errorf("failed to load templates: %w", err)
	}

	// Render each template
	for _, tmpl := range templates {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return fmt.Errorf("context cancelled: %w", ctxErr)
		}

		// Compute relative path
		relPath, err := filepath.Rel(templateDir, tmpl.Path)
		if err != nil {
			return fmt.Errorf("failed to compute relative path: %w", err)
		}

		// Compute output path
		outputPath := filepath.Join(outputDir, relPath)

		// Validate template
		if err := ValidateTemplate(tmpl, tmplCtx); err != nil {
			return fmt.Errorf("template validation failed for %s: %w", relPath, err)
		}

		// Render template
		output, err := m.renderer.Render(tmpl, tmplCtx)
		if err != nil {
			return fmt.Errorf("failed to render %s: %w", relPath, err)
		}

		// Check for unrendered variables
		unrendered := CheckUnrenderedVariables(output)
		if len(unrendered) > 0 {
			return fmt.Errorf("template %s has unrendered variables: %v", relPath, unrendered)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return fmt.Errorf("failed to create directory for %s: %w", relPath, err)
		}

		// Write output file
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", outputPath, err)
		}
	}

	return nil
}

// RenderString renders a template string with the given context
func (m *Manager) RenderString(content string, ctx *Context) (string, error) {
	output, err := m.renderer.RenderString(content, ctx)
	if err != nil {
		return "", err
	}

	// Check for unrendered variables
	unrendered := CheckUnrenderedVariables(output)
	if len(unrendered) > 0 {
		return "", fmt.Errorf("template has unrendered variables: %v", unrendered)
	}

	return output, nil
}

// GetTemplateInfo loads a template and returns its metadata
func (m *Manager) GetTemplateInfo(ctx context.Context, templatePath string) (*Metadata, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	tmpl, err := m.loader.Load(ctx, templatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load template: %w", err)
	}

	if tmpl.Metadata == nil {
		return &Metadata{}, nil
	}

	return tmpl.Metadata, nil
}

// ListTemplates lists all templates in a directory
func (m *Manager) ListTemplates(ctx context.Context, dir string) ([]*Template, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	return m.loader.LoadAll(ctx, dir)
}
