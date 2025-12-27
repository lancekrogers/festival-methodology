// Package navigation provides festival-project linking and navigation state management.
package navigation

import (
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/config"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	// NavigationFileName is the navigation state file name
	NavigationFileName = "navigation.yaml"
)

// Navigation represents the navigation state for festivals
type Navigation struct {
	Version   int               `yaml:"version"`
	UpdatedAt time.Time         `yaml:"updated_at"`
	Links     map[string]*Link  `yaml:"links"`
	Shortcuts map[string]string `yaml:"shortcuts,omitempty"`
}

// Link represents a festival-to-project link
type Link struct {
	Path     string    `yaml:"path"`
	LinkedAt time.Time `yaml:"linked_at"`
}

// NavigationPath returns the path to the navigation state file
func NavigationPath() string {
	return filepath.Join(config.ConfigDir(), NavigationFileName)
}

// LoadNavigation loads navigation state from disk
func LoadNavigation() (*Navigation, error) {
	navPath := NavigationPath()

	// Check if file exists
	if _, err := os.Stat(navPath); os.IsNotExist(err) {
		// Return empty navigation if file doesn't exist
		return &Navigation{
			Version:   1,
			UpdatedAt: time.Now().UTC(),
			Links:     make(map[string]*Link),
			Shortcuts: make(map[string]string),
		}, nil
	}

	data, err := os.ReadFile(navPath)
	if err != nil {
		return nil, errors.IO("reading navigation file", err).WithField("path", navPath)
	}

	var nav Navigation
	if err := yaml.Unmarshal(data, &nav); err != nil {
		return nil, errors.Parse("parsing navigation file", err).WithField("path", navPath)
	}

	// Initialize maps if nil
	if nav.Links == nil {
		nav.Links = make(map[string]*Link)
	}
	if nav.Shortcuts == nil {
		nav.Shortcuts = make(map[string]string)
	}

	return &nav, nil
}

// Save writes the navigation state to disk
func (n *Navigation) Save() error {
	n.UpdatedAt = time.Now().UTC()

	data, err := yaml.Marshal(n)
	if err != nil {
		return errors.Wrap(err, "marshaling navigation state")
	}

	navPath := NavigationPath()

	// Ensure config directory exists
	configDir := filepath.Dir(navPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return errors.IO("creating config directory", err).WithField("path", configDir)
	}

	if err := os.WriteFile(navPath, data, 0644); err != nil {
		return errors.IO("writing navigation file", err).WithField("path", navPath)
	}

	return nil
}

// SetLink creates or updates a festival-to-project link
func (n *Navigation) SetLink(festivalName, projectPath string) {
	n.Links[festivalName] = &Link{
		Path:     projectPath,
		LinkedAt: time.Now().UTC(),
	}
}

// GetLink retrieves a project link for a festival
func (n *Navigation) GetLink(festivalName string) (*Link, bool) {
	link, ok := n.Links[festivalName]
	return link, ok
}

// RemoveLink removes a festival-to-project link
func (n *Navigation) RemoveLink(festivalName string) bool {
	if _, ok := n.Links[festivalName]; ok {
		delete(n.Links, festivalName)
		return true
	}
	return false
}

// ListLinks returns all festival-project links
func (n *Navigation) ListLinks() map[string]*Link {
	return n.Links
}
