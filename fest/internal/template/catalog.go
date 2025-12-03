package template

import (
    "fmt"
    "path/filepath"
)

// Catalog maps template IDs (and aliases) to file paths
type Catalog struct {
    byID map[string]string
}

// LoadCatalog scans the given templateRoot and builds an ID -> path map.
func LoadCatalog(templateRoot string) (*Catalog, error) {
    l := NewLoader()
    tmpls, err := l.LoadAll(templateRoot)
    if err != nil {
        return nil, err
    }
    c := &Catalog{byID: make(map[string]string)}
    for _, t := range tmpls {
        rel := t.Path
        if filepath.IsAbs(t.Path) {
            if r, err := filepath.Rel(templateRoot, t.Path); err == nil {
                rel = filepath.Join(templateRoot, r)
            }
        }
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
func (m *Manager) RenderByID(catalog *Catalog, id string, ctx *Context) (string, error) {
    if catalog == nil {
        return "", fmt.Errorf("template catalog is nil")
    }
    path, ok := catalog.Resolve(id)
    if !ok {
        return "", fmt.Errorf("unknown template id: %s", id)
    }
    return m.RenderFile(path, ctx)
}

