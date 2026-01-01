package config

import "time"

// FestivalMetadata contains identity and history for a festival.
// This struct enables traceability of code comments back to specific
// festival nodes (e.g., GU0001:P002.S01.T03) and tracks status history.
type FestivalMetadata struct {
	// ID is the unique festival identifier (e.g., "GU0001").
	// Format: 2 uppercase letters (from name initials) + 4 digit counter.
	ID string `yaml:"id,omitempty"`

	// UUID is a globally unique identifier for cross-system references.
	UUID string `yaml:"uuid,omitempty"`

	// Name is the original festival name without the ID suffix.
	Name string `yaml:"name,omitempty"`

	// CreatedAt is when the festival was first created.
	CreatedAt time.Time `yaml:"created_at,omitempty"`

	// StatusHistory tracks all status transitions for this festival.
	StatusHistory []StatusChange `yaml:"status_history,omitempty"`
}

// StatusChange records a festival status transition.
// Each time a festival moves between statuses (planned, active, completed, archived),
// a new StatusChange is appended to the history.
type StatusChange struct {
	// Status is the new status value (planned, active, completed, archived).
	Status string `yaml:"status"`

	// Timestamp is when this status change occurred.
	Timestamp time.Time `yaml:"timestamp"`

	// Path is the filesystem path after this status change (optional).
	Path string `yaml:"path,omitempty"`

	// Notes provides additional context for this status change (optional).
	Notes string `yaml:"notes,omitempty"`
}

// HasMetadata returns true if this metadata has been populated with an ID.
func (m *FestivalMetadata) HasMetadata() bool {
	return m != nil && m.ID != ""
}

// CurrentStatus returns the most recent status, or empty string if no history.
func (m *FestivalMetadata) CurrentStatus() string {
	if m == nil || len(m.StatusHistory) == 0 {
		return ""
	}
	return m.StatusHistory[len(m.StatusHistory)-1].Status
}

// AddStatusChange appends a new status change to the history.
func (m *FestivalMetadata) AddStatusChange(status, path, notes string) {
	m.StatusHistory = append(m.StatusHistory, StatusChange{
		Status:    status,
		Timestamp: time.Now(),
		Path:      path,
		Notes:     notes,
	})
}
