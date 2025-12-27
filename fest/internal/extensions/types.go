// Package extensions provides loadable methodology extension packs.
package extensions

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

const (
	// ExtensionManifestFileName is the manifest filename
	ExtensionManifestFileName = "extension.yml"
	// ExtensionsDirName is the extensions directory name
	ExtensionsDirName = "extensions"
)

// Extension represents a loaded extension pack
type Extension struct {
	Name        string   `yaml:"name" json:"name"`
	Version     string   `yaml:"version" json:"version"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Author      string   `yaml:"author,omitempty" json:"author,omitempty"`
	Type        string   `yaml:"type,omitempty" json:"type,omitempty"` // workflow, template, agent, etc.
	Tags        []string `yaml:"tags,omitempty" json:"tags,omitempty"`

	// Paths and metadata
	Path   string `yaml:"-" json:"path"`   // Directory containing the extension
	Source string `yaml:"-" json:"source"` // project, user, built-in

	// Extension contents
	Files []ExtensionFile `yaml:"files,omitempty" json:"files,omitempty"`
}

// ExtensionFile represents a file in an extension pack
type ExtensionFile struct {
	Path        string `yaml:"path" json:"path"`
	Description string `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string `yaml:"type,omitempty" json:"type,omitempty"` // template, readme, config, etc.
}

// ExtensionManifest is the extension.yml file content
type ExtensionManifest struct {
	Name        string          `yaml:"name"`
	Version     string          `yaml:"version"`
	Description string          `yaml:"description,omitempty"`
	Author      string          `yaml:"author,omitempty"`
	Type        string          `yaml:"type,omitempty"`
	Tags        []string        `yaml:"tags,omitempty"`
	Files       []ExtensionFile `yaml:"files,omitempty"`
}

// LoadExtensionManifest loads an extension manifest from a file
func LoadExtensionManifest(path string) (*ExtensionManifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.IO("reading manifest", err).WithField("path", path)
	}

	var manifest ExtensionManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, errors.Parse("parsing manifest", err).WithField("path", path)
	}

	return &manifest, nil
}

// SaveExtensionManifest saves an extension manifest to a file
func SaveExtensionManifest(path string, manifest *ExtensionManifest) error {
	data, err := yaml.Marshal(manifest)
	if err != nil {
		return errors.Wrap(err, "marshaling manifest")
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return errors.IO("writing manifest", err).WithField("path", path)
	}

	return nil
}

// LoadExtensionFromDir loads an extension from a directory
func LoadExtensionFromDir(dir string, source string) (*Extension, error) {
	manifestPath := filepath.Join(dir, ExtensionManifestFileName)

	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		// No manifest - create extension from directory name
		name := filepath.Base(dir)
		return &Extension{
			Name:   name,
			Path:   dir,
			Source: source,
		}, nil
	}

	// Load manifest
	manifest, err := LoadExtensionManifest(manifestPath)
	if err != nil {
		return nil, err
	}

	return &Extension{
		Name:        manifest.Name,
		Version:     manifest.Version,
		Description: manifest.Description,
		Author:      manifest.Author,
		Type:        manifest.Type,
		Tags:        manifest.Tags,
		Files:       manifest.Files,
		Path:        dir,
		Source:      source,
	}, nil
}

// ListFiles returns all files in the extension directory
func (e *Extension) ListFiles() ([]string, error) {
	var files []string

	err := filepath.Walk(e.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() != ExtensionManifestFileName {
			relPath, _ := filepath.Rel(e.Path, path)
			files = append(files, relPath)
		}
		return nil
	})

	return files, err
}

// GetFile returns the full path to a file in the extension
func (e *Extension) GetFile(name string) string {
	return filepath.Join(e.Path, name)
}

// HasFile checks if a file exists in the extension
func (e *Extension) HasFile(name string) bool {
	path := e.GetFile(name)
	_, err := os.Stat(path)
	return err == nil
}
