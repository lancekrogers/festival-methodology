package vim

import (
	"strings"
	"testing"
)

func TestModeString(t *testing.T) {
	tests := []struct {
		mode Mode
		want string
	}{
		{ModeNormal, "NORMAL"},
		{ModeInsert, "INSERT"},
		{Mode(99), "NORMAL"}, // Unknown mode defaults to NORMAL
	}

	for _, tt := range tests {
		got := tt.mode.String()
		if got != tt.want {
			t.Errorf("Mode(%d).String() = %q, want %q", tt.mode, got, tt.want)
		}
	}
}

func TestModeIndicatorNewIndicator(t *testing.T) {
	// Test enabled
	m := NewIndicator(true)
	if !m.Enabled {
		t.Error("NewIndicator(true) should set Enabled=true")
	}
	if m.Mode != ModeNormal {
		t.Error("NewIndicator should default to ModeNormal")
	}

	// Test disabled
	m = NewIndicator(false)
	if m.Enabled {
		t.Error("NewIndicator(false) should set Enabled=false")
	}
}

func TestModeIndicatorSetMode(t *testing.T) {
	m := NewIndicator(true)

	m.SetMode(ModeInsert)
	if m.Mode != ModeInsert {
		t.Error("SetMode(ModeInsert) should set mode to INSERT")
	}

	m.SetMode(ModeNormal)
	if m.Mode != ModeNormal {
		t.Error("SetMode(ModeNormal) should set mode to NORMAL")
	}
}

func TestModeIndicatorToggle(t *testing.T) {
	m := NewIndicator(true)

	// Start in NORMAL
	if m.Mode != ModeNormal {
		t.Error("Should start in ModeNormal")
	}

	// Toggle to INSERT
	m.Toggle()
	if m.Mode != ModeInsert {
		t.Error("After first toggle, should be in ModeInsert")
	}

	// Toggle back to NORMAL
	m.Toggle()
	if m.Mode != ModeNormal {
		t.Error("After second toggle, should be in ModeNormal")
	}
}

func TestModeIndicatorRender(t *testing.T) {
	// Test disabled indicator returns empty
	m := NewIndicator(false)
	if got := m.Render(); got != "" {
		t.Errorf("Disabled indicator Render() = %q, want empty", got)
	}

	// Test enabled indicator renders mode
	m = NewIndicator(true)

	got := m.Render()
	if !strings.Contains(got, "NORMAL") {
		t.Errorf("Enabled NORMAL indicator should contain 'NORMAL', got %q", got)
	}

	m.SetMode(ModeInsert)
	got = m.Render()
	if !strings.Contains(got, "INSERT") {
		t.Errorf("Enabled INSERT indicator should contain 'INSERT', got %q", got)
	}
}

func TestModeIndicatorTitleWithMode(t *testing.T) {
	// Test disabled - returns original title
	m := NewIndicator(false)
	got := m.TitleWithMode("Description")
	if got != "Description" {
		t.Errorf("Disabled TitleWithMode = %q, want 'Description'", got)
	}

	// Test enabled NORMAL mode
	m = NewIndicator(true)
	got = m.TitleWithMode("Description")
	if !strings.HasPrefix(got, "Description (") {
		t.Errorf("Enabled TitleWithMode should start with 'Description (', got %q", got)
	}
	if !strings.Contains(got, "NORMAL") {
		t.Errorf("NORMAL mode TitleWithMode should contain 'NORMAL', got %q", got)
	}

	// Test enabled INSERT mode
	m.SetMode(ModeInsert)
	got = m.TitleWithMode("Description")
	if !strings.Contains(got, "INSERT") {
		t.Errorf("INSERT mode TitleWithMode should contain 'INSERT', got %q", got)
	}
}

func TestFormatTitleWithMode(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		mode    Mode
		enabled bool
		wantIn  string // substring that must be present
		wantOut string // substring that must NOT be present
	}{
		{
			name:    "disabled returns original",
			title:   "Test Title",
			mode:    ModeNormal,
			enabled: false,
			wantIn:  "Test Title",
			wantOut: "NORMAL",
		},
		{
			name:    "enabled NORMAL includes indicator",
			title:   "Test Title",
			mode:    ModeNormal,
			enabled: true,
			wantIn:  "NORMAL",
		},
		{
			name:    "enabled INSERT includes indicator",
			title:   "Test Title",
			mode:    ModeInsert,
			enabled: true,
			wantIn:  "INSERT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTitleWithMode(tt.title, tt.mode, tt.enabled)

			if tt.wantIn != "" && !strings.Contains(got, tt.wantIn) {
				t.Errorf("FormatTitleWithMode() = %q, should contain %q", got, tt.wantIn)
			}
			if tt.wantOut != "" && strings.Contains(got, tt.wantOut) {
				t.Errorf("FormatTitleWithMode() = %q, should NOT contain %q", got, tt.wantOut)
			}
		})
	}
}

func TestModeConstants(t *testing.T) {
	// Verify mode constants have expected values for serialization
	if ModeNormal != 0 {
		t.Errorf("ModeNormal should be 0, got %d", ModeNormal)
	}
	if ModeInsert != 1 {
		t.Errorf("ModeInsert should be 1, got %d", ModeInsert)
	}
}
