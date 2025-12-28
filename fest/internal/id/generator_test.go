package id

import (
	"strings"
	"testing"
)

func TestNewGenerator(t *testing.T) {
	g := NewGenerator()
	if g == nil {
		t.Fatal("NewGenerator() returned nil")
	}
	if g.seen == nil {
		t.Error("seen map should be initialized")
	}
}

func TestGenerator_Generate(t *testing.T) {
	g := NewGenerator()

	id1 := g.Generate("path/to/file.md")

	// Check format
	if !Validate(id1) {
		t.Errorf("Generated ID should be valid: %s", id1)
	}

	// Check prefix
	if !strings.HasPrefix(id1, IDPrefix+"-") {
		t.Errorf("ID should start with %s-: %s", IDPrefix, id1)
	}

	// Check uniqueness
	id2 := g.Generate("path/to/file.md")
	if id1 == id2 {
		t.Error("Subsequent calls should generate unique IDs")
	}
}

func TestGenerator_GenerateFromContent(t *testing.T) {
	g := NewGenerator()

	id1 := g.GenerateFromContent("same content")
	id2 := g.GenerateFromContent("same content")

	// Same content should produce same ID
	if id1 != id2 {
		t.Errorf("Same content should produce same ID: %s != %s", id1, id2)
	}

	// Different content should produce different ID
	id3 := g.GenerateFromContent("different content")
	if id1 == id3 {
		t.Error("Different content should produce different IDs")
	}
}

func TestGenerator_Register(t *testing.T) {
	g := NewGenerator()

	ref := "FEST-abc123"
	g.Register(ref)

	if !g.IsRegistered(ref) {
		t.Error("Registered ID should be found")
	}

	if g.IsRegistered("FEST-xyz789") {
		t.Error("Unregistered ID should not be found")
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		id    string
		valid bool
	}{
		{"FEST-abc123", true},
		{"FEST-000000", true},
		{"FEST-zzz999", true},
		{"FEST-abcdef", true},
		{"fest-abc123", false}, // lowercase prefix
		{"FEST-ABC123", false}, // uppercase hash
		{"FEST-abc12", false},  // too short
		{"FEST-abc1234", false}, // too long
		{"FESTabc123", false},   // missing dash
		{"ABC-abc123", false},   // wrong prefix
		{"", false},
	}

	for _, tc := range tests {
		t.Run(tc.id, func(t *testing.T) {
			if got := Validate(tc.id); got != tc.valid {
				t.Errorf("Validate(%q) = %v, want %v", tc.id, got, tc.valid)
			}
		})
	}
}

func TestParse(t *testing.T) {
	hash, err := Parse("FEST-abc123")
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if hash != "abc123" {
		t.Errorf("Parse() = %q, want %q", hash, "abc123")
	}

	_, err = Parse("invalid")
	if err == nil {
		t.Error("Parse() should error on invalid ID")
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		hash     string
		expected string
	}{
		{"abc123", "FEST-abc123"},
		{"ABCDEF", "FEST-abcdef"}, // Should lowercase
		{"abc1234567", "FEST-abc123"}, // Should truncate
	}

	for _, tc := range tests {
		t.Run(tc.hash, func(t *testing.T) {
			if got := Format(tc.hash); got != tc.expected {
				t.Errorf("Format(%q) = %q, want %q", tc.hash, got, tc.expected)
			}
		})
	}
}

func TestExtractFromMessage(t *testing.T) {
	tests := []struct {
		message  string
		expected []string
	}{
		{
			"[FEST-abc123] Implement feature",
			[]string{"FEST-abc123"},
		},
		{
			"[FEST-abc123] Related to [FEST-def456]",
			[]string{"FEST-abc123", "FEST-def456"},
		},
		{
			"No references here",
			nil,
		},
		{
			"[FEST-abc123] first [FEST-xyz789] second [FEST-111222]",
			[]string{"FEST-abc123", "FEST-xyz789", "FEST-111222"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.message, func(t *testing.T) {
			got := ExtractFromMessage(tc.message)

			if len(got) != len(tc.expected) {
				t.Errorf("ExtractFromMessage() returned %d IDs, want %d", len(got), len(tc.expected))
				return
			}

			for i, id := range got {
				if id != tc.expected[i] {
					t.Errorf("ID[%d] = %q, want %q", i, id, tc.expected[i])
				}
			}
		})
	}
}

func TestFormatForCommit(t *testing.T) {
	result := FormatForCommit("FEST-abc123")
	expected := "[FEST-abc123]"
	if result != expected {
		t.Errorf("FormatForCommit() = %q, want %q", result, expected)
	}
}

func TestGenerator_Uniqueness(t *testing.T) {
	g := NewGenerator()
	seen := make(map[string]bool)

	// Generate 1000 IDs and ensure uniqueness
	for i := 0; i < 1000; i++ {
		id := g.Generate("test/path")
		if seen[id] {
			t.Errorf("Duplicate ID generated: %s", id)
		}
		seen[id] = true
	}
}
