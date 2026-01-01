// Package registry provides the central ID registry for fast festival lookups.
// The registry maintains a map of festival IDs to their metadata and paths,
// enabling O(1) lookup and uniqueness validation.
package registry

import (
	"context"
	"sync"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Registry holds the central index of all festival IDs and their metadata.
// It provides thread-safe access for concurrent operations.
type Registry struct {
	Version string                   `yaml:"version"`
	Entries map[string]RegistryEntry `yaml:"entries"`

	mu   sync.RWMutex `yaml:"-"`
	path string       `yaml:"-"`
}

// RegistryEntry represents a single festival in the registry.
type RegistryEntry struct {
	ID        string    `yaml:"id"`
	Name      string    `yaml:"name"`
	Status    string    `yaml:"status"`
	Path      string    `yaml:"path"`
	CreatedAt time.Time `yaml:"created_at,omitempty"`
	UpdatedAt time.Time `yaml:"updated_at,omitempty"`
}

// New creates a new empty registry with the given file path.
func New(path string) *Registry {
	return &Registry{
		Version: "1.0",
		Entries: make(map[string]RegistryEntry),
		path:    path,
	}
}

// Path returns the file path where the registry is stored.
func (r *Registry) Path() string {
	return r.path
}

// Add adds a new entry to the registry.
// Returns an error if the ID already exists.
func (r *Registry) Add(ctx context.Context, entry RegistryEntry) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if entry.ID == "" {
		return errors.Validation("entry ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Entries[entry.ID]; exists {
		return errors.Validation("festival ID already exists").
			WithField("id", entry.ID)
	}

	entry.UpdatedAt = time.Now()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = entry.UpdatedAt
	}
	r.Entries[entry.ID] = entry
	return nil
}

// Get retrieves an entry by ID.
// Returns an error if the ID does not exist.
func (r *Registry) Get(ctx context.Context, id string) (RegistryEntry, error) {
	if err := ctx.Err(); err != nil {
		return RegistryEntry{}, errors.Wrap(err, "context cancelled")
	}

	if id == "" {
		return RegistryEntry{}, errors.Validation("ID is required")
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	entry, exists := r.Entries[id]
	if !exists {
		return RegistryEntry{}, errors.NotFound("festival not found in registry").
			WithField("id", id)
	}
	return entry, nil
}

// Update updates an existing entry in the registry.
// Returns an error if the ID does not exist.
func (r *Registry) Update(ctx context.Context, entry RegistryEntry) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if entry.ID == "" {
		return errors.Validation("entry ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.Entries[entry.ID]
	if !exists {
		return errors.NotFound("festival not found in registry").
			WithField("id", entry.ID)
	}

	// Preserve creation time
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = existing.CreatedAt
	}
	entry.UpdatedAt = time.Now()
	r.Entries[entry.ID] = entry
	return nil
}

// Delete removes an entry from the registry.
// Returns an error if the ID does not exist.
func (r *Registry) Delete(ctx context.Context, id string) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled")
	}

	if id == "" {
		return errors.Validation("ID is required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.Entries[id]; !exists {
		return errors.NotFound("festival not found in registry").
			WithField("id", id)
	}

	delete(r.Entries, id)
	return nil
}

// Exists checks if an ID exists in the registry.
func (r *Registry) Exists(id string) bool {
	if id == "" {
		return false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.Entries[id]
	return exists
}

// List returns all entries in the registry.
func (r *Registry) List(ctx context.Context) []RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	entries := make([]RegistryEntry, 0, len(r.Entries))
	for _, entry := range r.Entries {
		entries = append(entries, entry)
	}
	return entries
}

// Count returns the number of entries in the registry.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.Entries)
}

// ByStatus returns all entries with the given status.
func (r *Registry) ByStatus(ctx context.Context, status string) []RegistryEntry {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var entries []RegistryEntry
	for _, entry := range r.Entries {
		if entry.Status == status {
			entries = append(entries, entry)
		}
	}
	return entries
}
