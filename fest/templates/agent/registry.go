package agent

import (
	"bytes"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"text/template"
)

// Registry holds all parsed agent instruction templates.
var Registry = make(map[string]*template.Template)

func init() {
	if err := loadAllTemplates(); err != nil {
		// Templates are critical - panic if they fail to load
		panic(fmt.Sprintf("failed to load agent templates: %v", err))
	}
}

// loadAllTemplates walks the embedded filesystem and loads all .tmpl files.
func loadAllTemplates() error {
	return fs.WalkDir(Templates, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".tmpl") {
			return nil
		}

		content, err := Templates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		}

		// Template name is path without .tmpl extension
		name := strings.TrimSuffix(path, ".tmpl")
		tmpl, err := template.New(name).Parse(string(content))
		if err != nil {
			return fmt.Errorf("parsing %s: %w", path, err)
		}

		Registry[name] = tmpl
		return nil
	})
}

// Get retrieves a template by name (e.g., "next/task", "gates/implementation/testing").
// Returns nil if template not found.
func Get(name string) *template.Template {
	return Registry[name]
}

// MustGet retrieves a template by name, panicking if not found.
func MustGet(name string) *template.Template {
	tmpl := Registry[name]
	if tmpl == nil {
		panic(fmt.Sprintf("template %q not found", name))
	}
	return tmpl
}

// Render executes a template with the given data and returns the result.
func Render(name string, data any) (string, error) {
	tmpl := Get(name)
	if tmpl == nil {
		return "", fmt.Errorf("template %q not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("executing template %q: %w", name, err)
	}
	return buf.String(), nil
}

// MustRender executes a template, panicking on error.
func MustRender(name string, data any) string {
	result, err := Render(name, data)
	if err != nil {
		panic(err)
	}
	return result
}

// List returns all registered template names.
func List() []string {
	names := make([]string, 0, len(Registry))
	for name := range Registry {
		names = append(names, name)
	}
	return names
}

// ListByPrefix returns template names matching a prefix (e.g., "gates/implementation").
func ListByPrefix(prefix string) []string {
	var names []string
	for name := range Registry {
		if strings.HasPrefix(name, prefix) {
			names = append(names, name)
		}
	}
	return names
}

// GetGateTemplate returns a gate template for a specific phase type and gate name.
// phaseType: "implementation", "research", "planning", "review", "non_coding_action"
// gateName: "testing", "review", "iterate", "commit", etc.
func GetGateTemplate(phaseType, gateName string) *template.Template {
	name := filepath.Join("gates", phaseType, gateName)
	return Get(name)
}
