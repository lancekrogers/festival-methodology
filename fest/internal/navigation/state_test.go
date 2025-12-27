package navigation

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadNavigation_NewFile(t *testing.T) {
	// Create a temporary directory for test config
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	nav, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	if nav == nil {
		t.Fatal("LoadNavigation() returned nil")
	}

	if nav.Version != 1 {
		t.Errorf("Version = %d, want 1", nav.Version)
	}

	if nav.Links == nil {
		t.Error("Links map should not be nil")
	}

	if nav.Shortcuts == nil {
		t.Error("Shortcuts map should not be nil")
	}
}

func TestNavigation_SetAndGetLink(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	nav, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	// Set a link
	nav.SetLink("test-festival", "/path/to/project")

	// Get the link
	link, found := nav.GetLink("test-festival")
	if !found {
		t.Fatal("GetLink() returned false, want true")
	}

	if link.Path != "/path/to/project" {
		t.Errorf("Link.Path = %q, want %q", link.Path, "/path/to/project")
	}

	// Check LinkedAt is recent
	if time.Since(link.LinkedAt) > time.Minute {
		t.Error("LinkedAt should be recent")
	}
}

func TestNavigation_SaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	// Create and save navigation with a link
	nav1, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	nav1.SetLink("my-festival", "/home/user/projects/my-project")
	nav1.Shortcuts["p"] = "/path/to/shortcut"

	if err := nav1.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	navPath := filepath.Join(tmpDir, NavigationFileName)
	if _, err := os.Stat(navPath); os.IsNotExist(err) {
		t.Fatalf("Navigation file was not created at %s", navPath)
	}

	// Load navigation again
	nav2, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	// Verify link was preserved
	link, found := nav2.GetLink("my-festival")
	if !found {
		t.Fatal("Link was not preserved after save/load")
	}

	if link.Path != "/home/user/projects/my-project" {
		t.Errorf("Link.Path = %q, want %q", link.Path, "/home/user/projects/my-project")
	}

	// Verify shortcut was preserved
	if nav2.Shortcuts["p"] != "/path/to/shortcut" {
		t.Errorf("Shortcut not preserved: got %q", nav2.Shortcuts["p"])
	}
}

func TestNavigation_RemoveLink(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	nav, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	// Set a link
	nav.SetLink("test-festival", "/path/to/project")

	// Verify link exists
	if _, found := nav.GetLink("test-festival"); !found {
		t.Fatal("Link should exist before removal")
	}

	// Remove the link
	removed := nav.RemoveLink("test-festival")
	if !removed {
		t.Error("RemoveLink() returned false, want true")
	}

	// Verify link is gone
	if _, found := nav.GetLink("test-festival"); found {
		t.Error("Link should not exist after removal")
	}

	// Remove non-existent link
	removed = nav.RemoveLink("non-existent")
	if removed {
		t.Error("RemoveLink() returned true for non-existent link")
	}
}

func TestNavigation_ListLinks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	nav, err := LoadNavigation()
	if err != nil {
		t.Fatalf("LoadNavigation() error = %v", err)
	}

	// Add multiple links
	nav.SetLink("festival-1", "/path/to/project1")
	nav.SetLink("festival-2", "/path/to/project2")
	nav.SetLink("festival-3", "/path/to/project3")

	links := nav.ListLinks()
	if len(links) != 3 {
		t.Errorf("ListLinks() returned %d links, want 3", len(links))
	}

	// Verify each link
	expectedLinks := map[string]string{
		"festival-1": "/path/to/project1",
		"festival-2": "/path/to/project2",
		"festival-3": "/path/to/project3",
	}

	for name, expectedPath := range expectedLinks {
		link, ok := links[name]
		if !ok {
			t.Errorf("Missing link for %q", name)
			continue
		}
		if link.Path != expectedPath {
			t.Errorf("Link[%q].Path = %q, want %q", name, link.Path, expectedPath)
		}
	}
}

func TestNavigationPath(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("FEST_CONFIG_DIR", tmpDir)

	path := NavigationPath()
	expected := filepath.Join(tmpDir, NavigationFileName)

	if path != expected {
		t.Errorf("NavigationPath() = %q, want %q", path, expected)
	}
}
