package template

import (
	"context"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Catalog maps template IDs (and aliases) to file paths
type Catalog struct {
	byID map[string]string
}

// LoadCatalog scans the given templateRoot and builds an ID -> path map.
func LoadCatalog(ctx context.Context, templateRoot string) (*Catalog, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").
			WithOp("LoadCatalog")
	}

	l := NewLoader()
	tmpls, err := l.LoadAll(ctx, templateRoot)
	if err != nil {
		return nil, err
	}
	c := &Catalog{byID: make(map[string]string)}
	for _, t := range tmpls {
		if t.Metadata != nil {
			id := t.Metadata.TemplateID
			if id != "" {
				c.byID[id] = t.Path
			}
			for _, alias := range t.Metadata.Aliases {
				if alias != "" {
					c.byID[alias] = t.Path
				}
			}
		}
	}
	return c, nil
}

// Resolve returns the template file path for a given ID/alias
func (c *Catalog) Resolve(id string) (string, bool) {
	p, ok := c.byID[id]
	return p, ok
}

// RenderByID renders a template by ID/alias using the catalog
func (m *Manager) RenderByID(ctx context.Context, catalog *Catalog, id string, tmplCtx *Context) (string, error) {
	// Check context early
	if err := ctx.Err(); err != nil {
		return "", errors.Wrap(err, "context cancelled").
			WithOp("Manager.RenderByID")
	}

	if catalog == nil {
		return "", errors.Validation("template catalog is nil").
			WithOp("Manager.RenderByID")
	}
	path, ok := catalog.Resolve(id)
	if !ok {
		return "", errors.NotFound("template").
			WithField("id", id)
	}
	return m.RenderFile(ctx, path, tmplCtx)
}
