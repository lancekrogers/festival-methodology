package config

import (
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func TestFestivalMetadata_MarshalYAML(t *testing.T) {
	now := time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		metadata FestivalMetadata
		wantErr  bool
	}{
		{
			name: "full metadata",
			metadata: FestivalMetadata{
				ID:        "GU0001",
				UUID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:      "guild-usable",
				CreatedAt: now,
				StatusHistory: []StatusChange{
					{
						Status:    "planned",
						Timestamp: now,
						Notes:     "Initial creation",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal metadata",
			metadata: FestivalMetadata{
				ID:   "FN0042",
				Name: "fest-node-ids",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.metadata)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(data) == 0 {
				t.Error("Marshal() produced empty output")
			}
		})
	}
}

func TestFestivalMetadata_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		want    FestivalMetadata
		wantErr bool
	}{
		{
			name: "full metadata",
			yaml: `id: GU0001
uuid: 550e8400-e29b-41d4-a716-446655440000
name: guild-usable
created_at: 2025-12-31T12:00:00Z
status_history:
  - status: planned
    timestamp: 2025-12-31T12:00:00Z
    notes: Initial creation
`,
			want: FestivalMetadata{
				ID:        "GU0001",
				UUID:      "550e8400-e29b-41d4-a716-446655440000",
				Name:      "guild-usable",
				CreatedAt: time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
				StatusHistory: []StatusChange{
					{
						Status:    "planned",
						Timestamp: time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC),
						Notes:     "Initial creation",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "minimal metadata",
			yaml: `id: FN0042
name: fest-node-ids
`,
			want: FestivalMetadata{
				ID:   "FN0042",
				Name: "fest-node-ids",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FestivalMetadata
			err := yaml.Unmarshal([]byte(tt.yaml), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got.ID != tt.want.ID {
				t.Errorf("ID = %v, want %v", got.ID, tt.want.ID)
			}
			if got.Name != tt.want.Name {
				t.Errorf("Name = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}

func TestFestivalMetadata_RoundTrip(t *testing.T) {
	now := time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

	original := FestivalMetadata{
		ID:        "GU0001",
		UUID:      "550e8400-e29b-41d4-a716-446655440000",
		Name:      "guild-usable",
		CreatedAt: now,
		StatusHistory: []StatusChange{
			{
				Status:    "planned",
				Timestamp: now,
				Notes:     "Initial creation",
			},
			{
				Status:    "active",
				Timestamp: now.Add(24 * time.Hour),
				Notes:     "Started work",
			},
		},
	}

	// Marshal
	data, err := yaml.Marshal(&original)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Unmarshal
	var roundTrip FestivalMetadata
	if err := yaml.Unmarshal(data, &roundTrip); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	// Verify
	if roundTrip.ID != original.ID {
		t.Errorf("ID mismatch: got %v, want %v", roundTrip.ID, original.ID)
	}
	if roundTrip.UUID != original.UUID {
		t.Errorf("UUID mismatch: got %v, want %v", roundTrip.UUID, original.UUID)
	}
	if roundTrip.Name != original.Name {
		t.Errorf("Name mismatch: got %v, want %v", roundTrip.Name, original.Name)
	}
	if len(roundTrip.StatusHistory) != len(original.StatusHistory) {
		t.Errorf("StatusHistory length mismatch: got %d, want %d",
			len(roundTrip.StatusHistory), len(original.StatusHistory))
	}
}

func TestStatusChange_MarshalYAML(t *testing.T) {
	now := time.Date(2025, 12, 31, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		change  StatusChange
		wantErr bool
	}{
		{
			name: "with notes",
			change: StatusChange{
				Status:    "active",
				Timestamp: now,
				Notes:     "Started implementation",
			},
			wantErr: false,
		},
		{
			name: "without notes",
			change: StatusChange{
				Status:    "completed",
				Timestamp: now,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := yaml.Marshal(&tt.change)
			if (err != nil) != tt.wantErr {
				t.Errorf("Marshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && len(data) == 0 {
				t.Error("Marshal() produced empty output")
			}
		})
	}
}

func TestFestivalMetadata_EmptyFields(t *testing.T) {
	tests := []struct {
		name     string
		metadata FestivalMetadata
	}{
		{
			name:     "completely empty",
			metadata: FestivalMetadata{},
		},
		{
			name: "empty status history",
			metadata: FestivalMetadata{
				ID:            "XX0001",
				Name:          "test",
				StatusHistory: []StatusChange{},
			},
		},
		{
			name: "nil status history",
			metadata: FestivalMetadata{
				ID:            "XX0001",
				Name:          "test",
				StatusHistory: nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should marshal without error
			data, err := yaml.Marshal(&tt.metadata)
			if err != nil {
				t.Errorf("Marshal() error = %v", err)
				return
			}

			// Should unmarshal without error
			var result FestivalMetadata
			if err := yaml.Unmarshal(data, &result); err != nil {
				t.Errorf("Unmarshal() error = %v", err)
			}
		})
	}
}

func TestFestivalMetadata_BackwardsCompatibility(t *testing.T) {
	// Test that fest.yaml without metadata section still works
	yamlWithoutMetadata := `version: "1.0"
quality_gates:
  enabled: true
  auto_append: true
  tasks:
    - id: testing_and_verify
      template: gates/QUALITY_GATE_TESTING
      enabled: true
excluded_patterns:
  - "*_planning"
`

	var cfg FestivalConfig
	err := yaml.Unmarshal([]byte(yamlWithoutMetadata), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal config without metadata: %v", err)
	}

	// Metadata should be zero value
	if cfg.Metadata.ID != "" {
		t.Errorf("Expected empty metadata ID, got %s", cfg.Metadata.ID)
	}
}

func TestFestivalConfig_WithMetadata(t *testing.T) {
	yamlWithMetadata := `version: "1.0"
metadata:
  id: GU0001
  uuid: 550e8400-e29b-41d4-a716-446655440000
  name: guild-usable
  created_at: 2025-12-31T12:00:00Z
  status_history:
    - status: planned
      timestamp: 2025-12-31T12:00:00Z
      notes: Initial creation
quality_gates:
  enabled: true
`

	var cfg FestivalConfig
	err := yaml.Unmarshal([]byte(yamlWithMetadata), &cfg)
	if err != nil {
		t.Fatalf("Failed to unmarshal config with metadata: %v", err)
	}

	if cfg.Metadata.ID != "GU0001" {
		t.Errorf("Expected metadata ID GU0001, got %s", cfg.Metadata.ID)
	}
	if cfg.Metadata.Name != "guild-usable" {
		t.Errorf("Expected metadata name guild-usable, got %s", cfg.Metadata.Name)
	}
	if len(cfg.Metadata.StatusHistory) != 1 {
		t.Errorf("Expected 1 status history entry, got %d", len(cfg.Metadata.StatusHistory))
	}
}

func TestFestivalMetadata_IDFormat(t *testing.T) {
	// Test valid ID formats
	validIDs := []string{
		"GU0001", // Standard: 2 letters + 4 digits
		"FN0042",
		"AB9999",
		"XY0000",
	}

	for _, id := range validIDs {
		t.Run("valid_"+id, func(t *testing.T) {
			metadata := FestivalMetadata{ID: id, Name: "test"}
			data, err := yaml.Marshal(&metadata)
			if err != nil {
				t.Errorf("Failed to marshal valid ID %s: %v", id, err)
			}
			if len(data) == 0 {
				t.Error("Marshal produced empty output")
			}
		})
	}
}

func TestFestivalMetadata_HasMetadata(t *testing.T) {
	tests := []struct {
		name     string
		metadata *FestivalMetadata
		want     bool
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			want:     false,
		},
		{
			name:     "empty metadata",
			metadata: &FestivalMetadata{},
			want:     false,
		},
		{
			name:     "metadata with ID",
			metadata: &FestivalMetadata{ID: "GU0001"},
			want:     true,
		},
		{
			name:     "metadata with empty ID",
			metadata: &FestivalMetadata{ID: "", Name: "test"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.HasMetadata()
			if got != tt.want {
				t.Errorf("HasMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFestivalMetadata_CurrentStatus(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		metadata *FestivalMetadata
		want     string
	}{
		{
			name:     "nil metadata",
			metadata: nil,
			want:     "",
		},
		{
			name:     "empty history",
			metadata: &FestivalMetadata{StatusHistory: []StatusChange{}},
			want:     "",
		},
		{
			name: "single status",
			metadata: &FestivalMetadata{
				StatusHistory: []StatusChange{
					{Status: "planned", Timestamp: now},
				},
			},
			want: "planned",
		},
		{
			name: "multiple statuses returns last",
			metadata: &FestivalMetadata{
				StatusHistory: []StatusChange{
					{Status: "planned", Timestamp: now},
					{Status: "active", Timestamp: now.Add(time.Hour)},
					{Status: "completed", Timestamp: now.Add(2 * time.Hour)},
				},
			},
			want: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.metadata.CurrentStatus()
			if got != tt.want {
				t.Errorf("CurrentStatus() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFestivalMetadata_AddStatusChange(t *testing.T) {
	metadata := &FestivalMetadata{
		ID:   "GU0001",
		Name: "test-festival",
	}

	// Add first status
	metadata.AddStatusChange("planned", "/path/to/planned", "Initial creation")

	if len(metadata.StatusHistory) != 1 {
		t.Fatalf("Expected 1 status change, got %d", len(metadata.StatusHistory))
	}

	change := metadata.StatusHistory[0]
	if change.Status != "planned" {
		t.Errorf("Status = %v, want planned", change.Status)
	}
	if change.Path != "/path/to/planned" {
		t.Errorf("Path = %v, want /path/to/planned", change.Path)
	}
	if change.Notes != "Initial creation" {
		t.Errorf("Notes = %v, want 'Initial creation'", change.Notes)
	}
	if change.Timestamp.IsZero() {
		t.Error("Timestamp should not be zero")
	}

	// Add second status
	metadata.AddStatusChange("active", "/path/to/active", "Started work")

	if len(metadata.StatusHistory) != 2 {
		t.Fatalf("Expected 2 status changes, got %d", len(metadata.StatusHistory))
	}

	if metadata.CurrentStatus() != "active" {
		t.Errorf("CurrentStatus() = %v, want active", metadata.CurrentStatus())
	}
}
