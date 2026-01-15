// Package feedback provides structured feedback collection for festivals.
package feedback

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/lancekrogers/festival-methodology/fest/internal/errors"
	"gopkg.in/yaml.v3"
)

const (
	// FeedbackDir is the directory name for feedback storage
	FeedbackDir = "feedback"
	// ConfigFile is the feedback configuration file name
	ConfigFile = "feedback.yaml"
	// ObservationsDir is the directory for observations
	ObservationsDir = "observations"
)

// Config represents the feedback configuration
type Config struct {
	Version  int        `yaml:"version"`
	Criteria []Criteria `yaml:"criteria"`
	Settings Settings   `yaml:"settings"`
}

// Criteria represents a feedback criteria category
type Criteria struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description,omitempty"`
}

// Settings represents feedback settings
type Settings struct {
	InjectReminder bool `yaml:"inject_reminder"`
	AutoLinkTask   bool `yaml:"auto_link_task"`
}

// Observation represents a single feedback observation
type Observation struct {
	ID          string `yaml:"id" json:"id"`
	Criteria    string `yaml:"criteria" json:"criteria"`
	Observation string `yaml:"observation" json:"observation"`
	Task        string `yaml:"task,omitempty" json:"task,omitempty"`
	Timestamp   string `yaml:"timestamp" json:"timestamp"`
	Severity    string `yaml:"severity,omitempty" json:"severity,omitempty"`
	Suggestion  string `yaml:"suggestion,omitempty" json:"suggestion,omitempty"`
}

// Store manages feedback storage within a festival
type Store struct {
	festivalPath    string
	feedbackDir     string
	observationsDir string
}

// NewStore creates a feedback store for the given festival path
func NewStore(festivalPath string) *Store {
	feedbackDir := filepath.Join(festivalPath, FeedbackDir)
	return &Store{
		festivalPath:    festivalPath,
		feedbackDir:     feedbackDir,
		observationsDir: filepath.Join(feedbackDir, ObservationsDir),
	}
}

// IsInitialized checks if feedback is configured for this festival
func (s *Store) IsInitialized() bool {
	configPath := filepath.Join(s.feedbackDir, ConfigFile)
	info, err := os.Stat(configPath)
	return err == nil && !info.IsDir()
}

// Init initializes feedback with the given criteria
func (s *Store) Init(ctx context.Context, criteria []string) (*Config, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("feedback.Init")
	}

	// Create directories
	if err := os.MkdirAll(s.observationsDir, 0755); err != nil {
		return nil, errors.IO("creating feedback directory", err).WithField("path", s.feedbackDir)
	}

	// Create config
	config := &Config{
		Version:  1,
		Criteria: make([]Criteria, len(criteria)),
		Settings: Settings{
			InjectReminder: true,
			AutoLinkTask:   true,
		},
	}
	for i, c := range criteria {
		config.Criteria[i] = Criteria{Name: c}
	}

	// Save config
	configPath := filepath.Join(s.feedbackDir, ConfigFile)
	data, err := yaml.Marshal(config)
	if err != nil {
		return nil, errors.Wrap(err, "marshaling config")
	}
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return nil, errors.IO("writing config file", err).WithField("path", configPath)
	}

	return config, nil
}

// LoadConfig loads the feedback configuration
func (s *Store) LoadConfig(ctx context.Context) (*Config, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("feedback.LoadConfig")
	}

	configPath := filepath.Join(s.feedbackDir, ConfigFile)
	data, err := os.ReadFile(configPath)
	if os.IsNotExist(err) {
		return nil, errors.NotFound("feedback configuration").
			WithField("hint", "run 'fest feedback init' first")
	}
	if err != nil {
		return nil, errors.IO("reading config file", err).WithField("path", configPath)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, errors.Parse("parsing feedback config", err)
	}

	return &config, nil
}

// AddObservation adds a new observation
func (s *Store) AddObservation(ctx context.Context, obs *Observation) error {
	if err := ctx.Err(); err != nil {
		return errors.Wrap(err, "context cancelled").WithOp("feedback.AddObservation")
	}

	// Load config to validate criteria
	config, err := s.LoadConfig(ctx)
	if err != nil {
		return err
	}

	// Validate criteria
	validCriteria := false
	for _, c := range config.Criteria {
		if strings.EqualFold(c.Name, obs.Criteria) {
			obs.Criteria = c.Name // Normalize case
			validCriteria = true
			break
		}
	}
	if !validCriteria {
		names := make([]string, len(config.Criteria))
		for i, c := range config.Criteria {
			names[i] = c.Name
		}
		return errors.Validation("unknown criteria").
			WithField("criteria", obs.Criteria).
			WithField("available", strings.Join(names, ", "))
	}

	// Generate ID
	nextID, err := s.nextObservationID()
	if err != nil {
		return err
	}
	obs.ID = fmt.Sprintf("%03d", nextID)

	// Set timestamp if not provided
	if obs.Timestamp == "" {
		obs.Timestamp = time.Now().UTC().Format(time.RFC3339)
	}

	// Validate severity
	if obs.Severity != "" && obs.Severity != "low" && obs.Severity != "medium" && obs.Severity != "high" {
		return errors.Validation("invalid severity").
			WithField("severity", obs.Severity).
			WithField("allowed", "low, medium, high")
	}

	// Save observation
	obsPath := filepath.Join(s.observationsDir, obs.ID+".yaml")
	data, err := yaml.Marshal(obs)
	if err != nil {
		return errors.Wrap(err, "marshaling observation")
	}
	if err := os.WriteFile(obsPath, data, 0644); err != nil {
		return errors.IO("writing observation", err).WithField("path", obsPath)
	}

	return nil
}

// ListObservations returns all observations, optionally filtered
func (s *Store) ListObservations(ctx context.Context, criteria, severity string) ([]*Observation, error) {
	if err := ctx.Err(); err != nil {
		return nil, errors.Wrap(err, "context cancelled").WithOp("feedback.ListObservations")
	}

	entries, err := os.ReadDir(s.observationsDir)
	if os.IsNotExist(err) {
		return []*Observation{}, nil
	}
	if err != nil {
		return nil, errors.IO("reading observations directory", err)
	}

	var observations []*Observation
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		obs, err := s.loadObservation(filepath.Join(s.observationsDir, entry.Name()))
		if err != nil {
			continue // Skip invalid files
		}

		// Apply filters
		if criteria != "" && !strings.EqualFold(obs.Criteria, criteria) {
			continue
		}
		if severity != "" && !strings.EqualFold(obs.Severity, severity) {
			continue
		}

		observations = append(observations, obs)
	}

	// Sort by ID
	sort.Slice(observations, func(i, j int) bool {
		return observations[i].ID < observations[j].ID
	})

	return observations, nil
}

// Export exports observations in the specified format
func (s *Store) Export(ctx context.Context, format string) (string, error) {
	observations, err := s.ListObservations(ctx, "", "")
	if err != nil {
		return "", err
	}

	switch format {
	case "json":
		data, err := json.MarshalIndent(observations, "", "  ")
		if err != nil {
			return "", errors.Wrap(err, "marshaling JSON")
		}
		return string(data), nil

	case "yaml":
		data, err := yaml.Marshal(observations)
		if err != nil {
			return "", errors.Wrap(err, "marshaling YAML")
		}
		return string(data), nil

	case "markdown":
		return s.exportMarkdown(observations), nil

	default:
		return "", errors.Validation("unknown format").
			WithField("format", format).
			WithField("allowed", "json, yaml, markdown")
	}
}

// GetReminderText returns the feedback reminder for agent injection
func (s *Store) GetReminderText(ctx context.Context) (string, error) {
	config, err := s.LoadConfig(ctx)
	if err != nil {
		return "", nil // No config = no reminder
	}

	if !config.Settings.InjectReminder {
		return "", nil
	}

	var sb strings.Builder
	sb.WriteString("\nRemember, you have been asked to supply feedback based on the below criteria:\n\n")
	for _, c := range config.Criteria {
		sb.WriteString(fmt.Sprintf("- %s\n", c.Name))
	}
	sb.WriteString("\nWhen you have new feedback to add please use:\n")
	sb.WriteString("  fest feedback add --json '{\"criteria\": \"...\", \"observation\": \"...\"}'\n")
	sb.WriteString("\nExisting feedback can be viewed by running:\n")
	sb.WriteString("  fest feedback view\n")

	return sb.String(), nil
}

func (s *Store) nextObservationID() (int, error) {
	entries, err := os.ReadDir(s.observationsDir)
	if os.IsNotExist(err) {
		return 1, nil
	}
	if err != nil {
		return 0, errors.IO("reading observations directory", err)
	}

	maxID := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".yaml")
		id, err := strconv.Atoi(name)
		if err == nil && id > maxID {
			maxID = id
		}
	}

	return maxID + 1, nil
}

func (s *Store) loadObservation(path string) (*Observation, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var obs Observation
	if err := yaml.Unmarshal(data, &obs); err != nil {
		return nil, err
	}

	return &obs, nil
}

func (s *Store) exportMarkdown(observations []*Observation) string {
	var sb strings.Builder

	sb.WriteString("# Feedback Report\n\n")
	sb.WriteString(fmt.Sprintf("Generated: %s\n\n", time.Now().Format(time.RFC3339)))

	// Group by criteria
	byCategory := make(map[string][]*Observation)
	for _, obs := range observations {
		byCategory[obs.Criteria] = append(byCategory[obs.Criteria], obs)
	}

	for criteria, obs := range byCategory {
		sb.WriteString(fmt.Sprintf("## %s\n\n", criteria))
		for _, o := range obs {
			sb.WriteString(fmt.Sprintf("### %s\n\n", o.ID))
			sb.WriteString(fmt.Sprintf("**Observation:** %s\n\n", o.Observation))
			if o.Task != "" {
				sb.WriteString(fmt.Sprintf("**Task:** %s\n\n", o.Task))
			}
			if o.Severity != "" {
				sb.WriteString(fmt.Sprintf("**Severity:** %s\n\n", o.Severity))
			}
			if o.Suggestion != "" {
				sb.WriteString(fmt.Sprintf("**Suggestion:** %s\n\n", o.Suggestion))
			}
			sb.WriteString("---\n\n")
		}
	}

	return sb.String()
}

// ParseObservationJSON parses a JSON observation
func ParseObservationJSON(input string) (*Observation, error) {
	var obs Observation
	if err := json.Unmarshal([]byte(input), &obs); err != nil {
		return nil, errors.Parse("parsing observation JSON", err)
	}
	return &obs, nil
}
