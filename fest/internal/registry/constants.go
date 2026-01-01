// Package registry provides the central ID registry for fast festival lookups.
package registry

import "os"

// Version constants
const (
	// CurrentVersion is the current registry file format version.
	CurrentVersion = "1.0"
)

// File permission constants
const (
	// DirPermissions is the default permission for registry directories.
	DirPermissions os.FileMode = 0755

	// FilePermissions is the default permission for registry files.
	FilePermissions os.FileMode = 0644
)

// Status constants for festival lifecycle
const (
	StatusActive    = "active"
	StatusPlanned   = "planned"
	StatusCompleted = "completed"
	StatusDungeon   = "dungeon"
	StatusUnknown   = "unknown"
)
