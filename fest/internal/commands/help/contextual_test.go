package help

import (
	"strings"
	"testing"

	"github.com/lancekrogers/festival-methodology/fest/internal/commands/show"
)

func TestGetContext(t *testing.T) {
	tests := []struct {
		name     string
		location *show.LocationInfo
		want     string
	}{
		{
			name:     "nil location",
			location: nil,
			want:     "Not currently in a festival directory",
		},
		{
			name: "festival only",
			location: &show.LocationInfo{
				Type:     "festival",
				Festival: &show.FestivalInfo{Name: "my-festival"},
			},
			want: "Festival: my-festival",
		},
		{
			name: "festival and phase",
			location: &show.LocationInfo{
				Type:     "phase",
				Festival: &show.FestivalInfo{Name: "my-festival"},
				Phase:    "001_PLANNING",
			},
			want: "Festival: my-festival > Phase: 001_PLANNING",
		},
		{
			name: "full path",
			location: &show.LocationInfo{
				Type:     "task",
				Festival: &show.FestivalInfo{Name: "my-festival"},
				Phase:    "001_PLANNING",
				Sequence: "01_setup",
				Task:     "01_init.md",
			},
			want: "Festival: my-festival > Phase: 001_PLANNING > Sequence: 01_setup > Task: 01_init.md",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &ContextualHelp{location: tc.location}
			got := h.GetContext()
			if got != tc.want {
				t.Errorf("GetContext() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestGetSuggestedCommands(t *testing.T) {
	tests := []struct {
		name         string
		location     *show.LocationInfo
		wantContains []string
	}{
		{
			name:         "nil location - generic suggestions",
			location:     nil,
			wantContains: []string{"fest tui", "fest show all"},
		},
		{
			name: "festival location",
			location: &show.LocationInfo{
				Type:     "festival",
				Festival: &show.FestivalInfo{Name: "test"},
			},
			wantContains: []string{"fest status", "fest validate", "fest next"},
		},
		{
			name: "phase location",
			location: &show.LocationInfo{
				Type:     "phase",
				Festival: &show.FestivalInfo{Name: "test"},
				Phase:    "001_PLANNING",
			},
			wantContains: []string{"fest status", "fest next", "fest go .."},
		},
		{
			name: "sequence location",
			location: &show.LocationInfo{
				Type:     "sequence",
				Festival: &show.FestivalInfo{Name: "test"},
				Phase:    "001_PLANNING",
				Sequence: "01_setup",
			},
			wantContains: []string{"fest progress", "fest next"},
		},
		{
			name: "task location",
			location: &show.LocationInfo{
				Type:     "task",
				Festival: &show.FestivalInfo{Name: "test"},
				Phase:    "001_PLANNING",
				Sequence: "01_setup",
				Task:     "01_init.md",
			},
			wantContains: []string{"fest progress", "fest next"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			h := &ContextualHelp{location: tc.location}
			suggestions := h.GetSuggestedCommands()

			// Check that suggestions is not empty
			if len(suggestions) == 0 {
				t.Error("GetSuggestedCommands() returned empty slice")
				return
			}

			// Check that expected commands are present
			for _, want := range tc.wantContains {
				found := false
				for _, s := range suggestions {
					if strings.Contains(s.Command, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("GetSuggestedCommands() missing %q", want)
				}
			}
		})
	}
}

func TestFormatContextualHelp(t *testing.T) {
	h := &ContextualHelp{
		location: &show.LocationInfo{
			Type:     "phase",
			Festival: &show.FestivalInfo{Name: "test-festival"},
			Phase:    "001_PLANNING",
		},
	}

	output := h.FormatContextualHelp()

	// Should contain context section
	if !strings.Contains(output, "CURRENT CONTEXT:") {
		t.Error("FormatContextualHelp() missing CURRENT CONTEXT section")
	}

	// Should contain festival name
	if !strings.Contains(output, "test-festival") {
		t.Error("FormatContextualHelp() missing festival name")
	}

	// Should contain suggestions section
	if !strings.Contains(output, "SUGGESTED NEXT STEPS:") {
		t.Error("FormatContextualHelp() missing SUGGESTED NEXT STEPS section")
	}
}

func TestGenericSuggestions(t *testing.T) {
	suggestions := genericSuggestions()

	if len(suggestions) == 0 {
		t.Error("genericSuggestions() returned empty slice")
	}

	// Should always include fest understand
	found := false
	for _, s := range suggestions {
		if strings.Contains(s.Command, "fest understand") {
			found = true
			break
		}
	}
	if !found {
		t.Error("genericSuggestions() should include 'fest understand'")
	}
}
