// Package status provides status history tracking for festivals.
package status

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/shared"
	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// StatusHistoryEntry represents a single status change record.
type StatusHistoryEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	FromStatus string    `json:"from_status"`
	ToStatus   string    `json:"to_status"`
	Note       string    `json:"note,omitempty"`
}

// RecordStatusChange appends a new entry to the festival's status history.
// It creates the .fest directory if it doesn't exist.
func RecordStatusChange(festivalDir, fromStatus, toStatus, note string) error {
	// Don't record if status hasn't changed
	if fromStatus == toStatus {
		return nil
	}

	festDir := filepath.Join(festivalDir, ".fest")
	historyPath := filepath.Join(festDir, "status_history.json")

	// Load existing history
	history, err := LoadStatusHistory(festivalDir)
	if err != nil {
		return errors.Wrap(err, "loading status history")
	}

	// Append new entry
	entry := StatusHistoryEntry{
		Timestamp:  time.Now(),
		FromStatus: fromStatus,
		ToStatus:   toStatus,
		Note:       note,
	}
	history = append(history, entry)

	// Ensure .fest directory exists
	if err := os.MkdirAll(festDir, 0755); err != nil {
		return errors.IO("creating .fest directory", err)
	}

	// Write updated history
	var buffer bytes.Buffer
	if err := shared.EncodeJSON(&buffer, history); err != nil {
		return errors.Wrap(err, "encoding status history JSON")
	}

	data := bytes.TrimRight(buffer.Bytes(), "\n")
	if err := os.WriteFile(historyPath, data, 0644); err != nil {
		return errors.IO("writing status history", err)
	}

	return nil
}

// LoadStatusHistory loads the status history from a festival's .fest directory.
// Returns an empty slice if no history file exists.
func LoadStatusHistory(festivalDir string) ([]StatusHistoryEntry, error) {
	historyPath := filepath.Join(festivalDir, ".fest", "status_history.json")

	// Check if file exists
	if _, err := os.Stat(historyPath); os.IsNotExist(err) {
		return []StatusHistoryEntry{}, nil
	}

	// Read and parse history
	data, err := os.ReadFile(historyPath)
	if err != nil {
		return nil, errors.IO("reading status history", err)
	}

	var history []StatusHistoryEntry
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, errors.Wrap(err, "parsing status history")
	}

	return history, nil
}

// GetLastStatusChange returns the most recent status change entry, if any.
func GetLastStatusChange(festivalDir string) (*StatusHistoryEntry, error) {
	history, err := LoadStatusHistory(festivalDir)
	if err != nil {
		return nil, err
	}

	if len(history) == 0 {
		return nil, nil
	}

	return &history[len(history)-1], nil
}
