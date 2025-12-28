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
	Version      int               `yaml:"version"`
	UpdatedAt    time.Time         `yaml:"updated_at"`
	Links        map[string]*Link  `yaml:"links"`                  // festival name -> project path
	ProjectLinks map[string]string `yaml:"project_links,omitempty"` // project path -> festival name (reverse lookup)
	Shortcuts    map[string]string `yaml:"shortcuts,omitempty"`
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
			Version:      1,
			UpdatedAt:    time.Now().UTC(),
			Links:        make(map[string]*Link),
			ProjectLinks: make(map[string]string),
			Shortcuts:    make(map[string]string),
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
	if nav.ProjectLinks == nil {
		nav.ProjectLinks = make(map[string]string)
	}
	if nav.Shortcuts == nil {
		nav.Shortcuts = make(map[string]string)
	}

	// Rebuild ProjectLinks from Links for backward compatibility
	nav.rebuildProjectLinks()

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

// SetLink creates or updates a bidirectional festival-to-project link
func (n *Navigation) SetLink(festivalName, projectPath string) {
	// Remove old reverse link if this festival was linked elsewhere
	if oldLink, ok := n.Links[festivalName]; ok {
		delete(n.ProjectLinks, oldLink.Path)
	}

	// Remove old link if this project was linked to another festival
	if oldFestival, ok := n.ProjectLinks[projectPath]; ok {
		delete(n.Links, oldFestival)
	}

	// Set forward link: festival -> project
	n.Links[festivalName] = &Link{
		Path:     projectPath,
		LinkedAt: time.Now().UTC(),
	}

	// Set reverse link: project -> festival
	n.ProjectLinks[projectPath] = festivalName
}

// GetLink retrieves a project link for a festival
func (n *Navigation) GetLink(festivalName string) (*Link, bool) {
	link, ok := n.Links[festivalName]
	return link, ok
}

// RemoveLink removes a bidirectional festival-to-project link
func (n *Navigation) RemoveLink(festivalName string) bool {
	if link, ok := n.Links[festivalName]; ok {
		// Remove reverse link
		delete(n.ProjectLinks, link.Path)
		// Remove forward link
		delete(n.Links, festivalName)
		return true
	}
	return false
}

// ListLinks returns all festival-project links
func (n *Navigation) ListLinks() map[string]*Link {
	return n.Links
}

// GetLinkedProject returns the project path linked to a festival
func (n *Navigation) GetLinkedProject(festivalName string) string {
	if link, ok := n.Links[festivalName]; ok {
		return link.Path
	}
	return ""
}

// GetLinkedFestival returns the festival name linked to a project path
func (n *Navigation) GetLinkedFestival(projectPath string) string {
	if festivalName, ok := n.ProjectLinks[projectPath]; ok {
		return festivalName
	}
	return ""
}

// FindFestivalForPath walks up the directory tree to find a linked festival.
// This handles the case where user is in a subdirectory of a linked project.
func (n *Navigation) FindFestivalForPath(path string) string {
	// Clean and absolutize path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return ""
	}

	// Walk up directory tree
	current := absPath
	for {
		if festivalName, ok := n.ProjectLinks[current]; ok {
			return festivalName
		}

		parent := filepath.Dir(current)
		if parent == current {
			// Reached root
			break
		}
		current = parent
	}

	return ""
}

// rebuildProjectLinks rebuilds the reverse lookup map from Links.
// This provides backward compatibility when loading old navigation files.
func (n *Navigation) rebuildProjectLinks() {
	n.ProjectLinks = make(map[string]string)
	for festivalName, link := range n.Links {
		n.ProjectLinks[link.Path] = festivalName
	}
}
