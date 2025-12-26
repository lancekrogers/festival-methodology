package template

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
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
		return "", errors.Wrap(err, "context cancelled").
			WithOp("Manager.RenderFile")
	}

	// Load template
	tmpl, err := m.loader.Load(ctx, templatePath)
	if err != nil {
		return "", errors.Wrap(err, "loading template").
			WithOp("Manager.RenderFile").
			WithField("path", templatePath)
	}

	// Validate required variables
	if err := ValidateTemplate(tmpl, tmplCtx); err != nil {
		return "", errors.Wrap(err, "template validation failed").
			WithCode(errors.ErrCodeValidation).
			WithField("path", templatePath)
	}

	// Render template
	output, err := m.renderer.Render(tmpl, tmplCtx)
	if err != nil {
		return "", errors.Wrap(err, "rendering template").
			WithCode(errors.ErrCodeTemplate).
			WithField("path", templatePath)
	}

	// Check for unrendered variables
	unrendered := CheckUnrenderedVariables(output)
	if len(unrendered) > 0 {
		return "", errors.Validation(fmt.Sprintf("template has unrendered variables: %v", unrendered)).
			WithField("path", templatePath)
	}

	return output, nil
}

// RenderFileToFile renders a template and writes it to an output file
func (m *Manager) RenderFileToFile(ctx context.Context, templatePath, outputPath string, tmplCtx *Context) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").
			WithOp("Manager.RenderFileToFile")
	}

	// Render template
	output, err := m.RenderFile(ctx, templatePath, tmplCtx)
	if err != nil {
		return err
	}

	// Ensure output directory exists
	outputDir := filepath.Dir(outputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return errors.IO("creating output directory", err).
			WithField("path", outputDir)
	}

	// Write output file
	if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
		return errors.IO("writing output file", err).
			WithField("path", outputPath)
	}

	return nil
}

// RenderDirectory renders all templates in a directory to an output directory
func (m *Manager) RenderDirectory(ctx context.Context, templateDir, outputDir string, tmplCtx *Context) error {
	// Check context early
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").
			WithOp("Manager.RenderDirectory")
	}

	// Load all templates
	templates, err := m.loader.LoadAll(ctx, templateDir)
	if err != nil {
		return errors.Wrap(err, "loading templates").
			WithOp("Manager.RenderDirectory").
			WithField("dir", templateDir)
	}

	// Render each template
	for _, tmpl := range templates {
		// Check context on each iteration
		if ctxErr := ctx.Err(); ctxErr != nil {
			return errors.Wrap(ctxErr, "context cancelled").
				WithOp("Manager.RenderDirectory")
		}

		// Compute relative path
		relPath, err := filepath.Rel(templateDir, tmpl.Path)
		if err != nil {
			return errors.Wrap(err, "computing relative path").
				WithField("template", tmpl.Path).
				WithField("base", templateDir)
		}

		// Compute output path
		outputPath := filepath.Join(outputDir, relPath)

		// Validate template
		if err := ValidateTemplate(tmpl, tmplCtx); err != nil {
			return errors.Wrap(err, "template validation failed").
				WithCode(errors.ErrCodeValidation).
				WithField("path", relPath)
		}

		// Render template
		output, err := m.renderer.Render(tmpl, tmplCtx)
		if err != nil {
			return errors.Wrap(err, "rendering template").
				WithCode(errors.ErrCodeTemplate).
				WithField("path", relPath)
		}

		// Check for unrendered variables
		unrendered := CheckUnrenderedVariables(output)
		if len(unrendered) > 0 {
			return errors.Validation(fmt.Sprintf("template has unrendered variables: %v", unrendered)).
				WithField("path", relPath)
		}

		// Ensure output directory exists
		if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
			return errors.IO("creating directory", err).
				WithField("path", relPath)
		}

		// Write output file
		if err := os.WriteFile(outputPath, []byte(output), 0644); err != nil {
			return errors.IO("writing output file", err).
				WithField("path", outputPath)
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
		return "", errors.Validation(fmt.Sprintf("template has unrendered variables: %v", unrendered))
	}

	return output, nil
}

// GetTemplateInfo loads a template and returns its metadata
func (m *Manager) GetTemplateInfo(ctx context.Context, templatePath string) (*Metadata, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("Manager.GetTemplateInfo")
	}

	tmpl, err := m.loader.Load(ctx, templatePath)
	if err != nil {
		return nil, errors.Wrap(err, "loading template").
			WithOp("Manager.GetTemplateInfo").
			WithField("path", templatePath)
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
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("Manager.ListTemplates")
	}

	return m.loader.LoadAll(ctx, dir)
}
