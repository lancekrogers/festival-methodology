package extensions

import (
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	tpl "github.com/lancekrogers/festival-methodology/fest/internal/template"
)

// ExtensionLoader handles loading extensions from multiple sources
type ExtensionLoader struct {
	extensions map[string]*Extension // keyed by name
	sources    []string              // source priority order
}

// NewExtensionLoader creates a new extension loader
func NewExtensionLoader() *ExtensionLoader {
	return &ExtensionLoader{
		extensions: make(map[string]*Extension),
		sources:    []string{},
	}
}

// LoadAll loads extensions from all sources with proper precedence
// 1. Project-local (.festival/extensions/)
// 2. User config repo (festivals/.festival/extensions/)
// 3. Built-in (~/.config/fest/festivals/.festival/extensions/)
func (el *ExtensionLoader) LoadAll(festivalRoot string) error {
	// Load built-in extensions first (lowest priority)
	builtInPath := filepath.Join(config.ConfigDir(), "festivals", ".festival", ExtensionsDirName)
	el.loadFromDirectory(builtInPath, "built-in")

	// Load user config repo extensions (medium priority)
	if userFestPath := config.ActiveFestivalsPath(); userFestPath != "" {
		userExtPath := filepath.Join(userFestPath, ExtensionsDirName)
		el.loadFromDirectory(userExtPath, "user")
	}

	// Load project-local extensions (highest priority)
	if festivalRoot != "" {
		localRoot, err := tpl.LocalTemplateRoot(festivalRoot)
		if err == nil {
			localExtPath := filepath.Join(localRoot, ExtensionsDirName)
			el.loadFromDirectory(localExtPath, "project")
		}
	}

	return nil
}

// loadFromDirectory loads all extensions from a directory
func (el *ExtensionLoader) loadFromDirectory(dir string, source string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil // Directory doesn't exist or can't be read
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		extPath := filepath.Join(dir, entry.Name())
		ext, err := LoadExtensionFromDir(extPath, source)
		if err != nil {
			continue // Skip invalid extensions
		}

		// Higher priority sources override lower ones
		el.extensions[ext.Name] = ext
	}

	return nil
}

// Get returns an extension by name
func (el *ExtensionLoader) Get(name string) *Extension {
	return el.extensions[name]
}

// List returns all loaded extensions
func (el *ExtensionLoader) List() []*Extension {
	result := make([]*Extension, 0, len(el.extensions))
	for _, ext := range el.extensions {
		result = append(result, ext)
	}
	return result
}

// ListBySource returns extensions from a specific source
func (el *ExtensionLoader) ListBySource(source string) []*Extension {
	var result []*Extension
	for _, ext := range el.extensions {
		if ext.Source == source {
			result = append(result, ext)
		}
	}
	return result
}

// ListByType returns extensions of a specific type
func (el *ExtensionLoader) ListByType(extType string) []*Extension {
	var result []*Extension
	for _, ext := range el.extensions {
		if ext.Type == extType {
			result = append(result, ext)
		}
	}
	return result
}

// Count returns the number of loaded extensions
func (el *ExtensionLoader) Count() int {
	return len(el.extensions)
}

// HasExtension checks if an extension is loaded
func (el *ExtensionLoader) HasExtension(name string) bool {
	_, ok := el.extensions[name]
	return ok
}
