package registry

import (
	"context"
	"os"
	"path/filepath"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

// DefaultRegistryPath returns the default path for the registry file
// relative to the festivals root directory.
const DefaultRegistryFileName = "id_registry.yaml"

// GetRegistryPath returns the full path to the registry file
// within the .festival directory of the festivals root.
func GetRegistryPath(festivalsRoot string) string {
	return filepath.Join(festivalsRoot, ".festival", DefaultRegistryFileName)
}

// Load reads and parses the registry from the given path.
// If the file does not exist, returns a new empty registry.
// Returns an error if the file exists but cannot be parsed.
func Load(ctx context.Context, path string) (*Registry, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled")
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Return new empty registry for non-existent file
		return New(path), nil
	}
	if err != nil {
		return nil, errors.IO("reading registry file", err).
			WithField("path", path)
	}

	var reg Registry
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, errors.Parse("parsing registry YAML", err).
			WithField("path", path)
	}

	// Initialize if nil (empty file case)
	if reg.Entries == nil {
		reg.Entries = make(map[string]RegistryEntry)
	}
	if reg.Version == "" {
		reg.Version = "1.0"
	}
	reg.path = path

	return &reg, nil
}

// Save writes the registry to its file path using atomic operations.
// It writes to a temporary file first, then renames to the final path
// to ensure the registry file is never corrupted.
func (r *Registry) Save(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	// Ensure the directory exists
	dir := filepath.Dir(r.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return errors.IO("creating registry directory", err).
			WithField("path", dir)
	}

	// Marshal to YAML
	r.mu.RLock()
	data, err := yaml.Marshal(r)
	r.mu.RUnlock()
	if err != nil {
		return errors.Parse("marshaling registry to YAML", err)
	}

	// Write to temporary file
	tmpPath := r.path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0644); err != nil {
		return errors.IO("writing temporary registry file", err).
			WithField("path", tmpPath)
	}

	// Atomic rename
	if err := os.Rename(tmpPath, r.path); err != nil {
		// Clean up temp file on rename failure
		os.Remove(tmpPath)
		return errors.IO("renaming registry file", err).
			WithField("from", tmpPath).
			WithField("to", r.path)
	}

	return nil
}

// LoadOrCreate loads an existing registry or creates a new one if it doesn't exist.
// The registry is saved after creation.
func LoadOrCreate(ctx context.Context, path string) (*Registry, error) {
	reg, err := Load(ctx, path)
	if err != nil {
		return nil, err
	}

	// If we created a new registry, save it
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		if err := reg.Save(ctx); err != nil {
			return nil, err
		}
	}

	return reg, nil
}
