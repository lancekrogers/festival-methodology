package template

import (
	"context"
	"os"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// RenderByIDOrFallback renders a template by catalog ID if available, otherwise
// falls back to rendering or copying a specific file path.
// If the template contains Go template delimiters or declares required variables,
// it is rendered with the provided context; otherwise the file content is copied.
func RenderByIDOrFallback(ctx context.Context, catalog *Catalog, id string, fallbackPath string, tmplCtx *Context) (string, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return "", errors.Wrap(err, "context cancelled").
			WithOp("RenderByIDOrFallback")
	}

	mgr := NewManager()
	// Try by ID via catalog
	if catalog != nil && id != "" {
		if content, err := mgr.RenderByID(ctx, catalog, id, tmplCtx); err == nil && content != "" {
			return content, nil
		}
	}

	// Fallback to explicit file
	// Load template file and decide render vs copy based on content/metadata
	loader := NewLoader()
	t, err := loader.Load(ctx, fallbackPath)
	if err != nil {
		// As a last resort, try to read raw content (if not a template with frontmatter)
		b, rerr := os.ReadFile(fallbackPath)
		if rerr == nil {
			return string(b), nil
		}
		return "", err
	}

	// If the template has required variables or contains delimiters, render
	requires := t.Metadata != nil && len(t.Metadata.RequiredVariables) > 0
	if requires || containsDelims(t.Content) {
		out, err := mgr.Render(t, tmplCtx)
		if err != nil {
			return "", err
		}
		return out, nil
	}

	return t.Content, nil
}

func containsDelims(s string) bool {
	for i := 0; i+1 < len(s); i++ {
		if s[i] == '{' && s[i+1] == '{' {
			return true
		}
	}
	return false
}
