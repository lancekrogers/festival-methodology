package template

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// Renderer renders templates with variable substitution
type Renderer interface {
	Render(tmpl *Template, ctx *Context) (string, error)
	RenderString(content string, ctx *Context) (string, error)
}

type rendererImpl struct {
	// Template cache for performance
	cache map[string]*template.Template
}

// NewRenderer creates a new template renderer
func NewRenderer() Renderer {
	return &rendererImpl{
		cache: make(map[string]*template.Template),
	}
}

// Render renders a template with the given context
func (r *rendererImpl) Render(tmpl *Template, ctx *Context) (string, error) {
	return r.RenderString(tmpl.Content, ctx)
}

// RenderString renders a template string with the given context
func (r *rendererImpl) RenderString(content string, ctx *Context) (string, error) {
	// Create Go template with Sprig functions
	tmpl, err := template.New("template").
		Funcs(sprig.TxtFuncMap()).
		Parse(content)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	// Flatten context for template execution
	vars := ctx.ToFlatMap()

	// Execute template
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, vars); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// ValidateTemplate checks if a template can be rendered with required variables
func ValidateTemplate(tmpl *Template, ctx *Context) error {
	if tmpl.Metadata == nil {
		// No metadata, can't validate
		return nil
	}

	// Check required variables
	missing := []string{}
	for _, required := range tmpl.Metadata.RequiredVariables {
		if _, ok := ctx.Get(required); !ok {
			missing = append(missing, required)
		}
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required variables: %s", strings.Join(missing, ", "))
	}

	return nil
}

// CheckUnrenderedVariables scans rendered content for unrendered variables
func CheckUnrenderedVariables(content string) []string {
	unrendered := []string{}

	// Find all {{variable}} patterns
	start := 0
	for {
		idx := strings.Index(content[start:], "{{")
		if idx == -1 {
			break
		}

		// Find closing }}
		idx += start
		closeIdx := strings.Index(content[idx:], "}}")
		if closeIdx == -1 {
			break
		}

		// Extract variable name
		variable := content[idx+2 : idx+closeIdx]
		variable = strings.TrimSpace(variable)

		// Skip if it's a comment or conditional
		if !strings.HasPrefix(variable, "#") &&
			!strings.HasPrefix(variable, "/") &&
			!strings.HasPrefix(variable, "!") {
			unrendered = append(unrendered, variable)
		}

		start = idx + closeIdx + 2
	}

	return unrendered
}

// PreserveGuidance ensures [GUIDANCE: ...] and [FILL: ...] markers are preserved
func PreserveGuidance(content string) string {
	// This is a no-op for now since Go templates don't touch [GUIDANCE]
	// markers - they're just plain text
	// This function exists for future enhancements if needed
	return content
}

// RenderWithDefaults renders a template with default values for missing variables
func RenderWithDefaults(tmpl *Template, ctx *Context) (string, error) {
	// Create a new context with defaults for optional variables
	enrichedCtx := &Context{
		Auto:     ctx.Auto,
		User:     make(map[string]interface{}),
		Computed: ctx.Computed,
	}

	// Copy user variables
	for k, v := range ctx.User {
		enrichedCtx.User[k] = v
	}

	// Add defaults for missing optional variables
	if tmpl.Metadata != nil {
		for _, optional := range tmpl.Metadata.OptionalVariables {
			if _, ok := enrichedCtx.Get(optional); !ok {
				// Set default value (empty string or "TBD")
				enrichedCtx.User[optional] = ""
			}
		}
	}

	// Render with enriched context
	renderer := NewRenderer()
	return renderer.Render(tmpl, enrichedCtx)
}
