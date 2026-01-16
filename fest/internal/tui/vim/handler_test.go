package vim

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewHandler(t *testing.T) {
	// Test enabled
	h := NewHandler(true)
	if !h.Enabled {
		t.Error("NewHandler(true) should set Enabled=true")
	}
	if h.Mode != ModeNormal {
		t.Error("NewHandler should default to ModeNormal")
	}
	if h.Clipboard != "" {
		t.Error("NewHandler should have empty clipboard")
	}

	// Test disabled
	h = NewHandler(false)
	if h.Enabled {
		t.Error("NewHandler(false) should set Enabled=false")
	}
}

func TestHandleKeyDisabled(t *testing.T) {
	h := NewHandler(false)

	// All keys should return ActionNone when disabled
	keys := []string{"h", "j", "k", "l", "i", "a", "esc"}
	for _, k := range keys {
		result := h.HandleKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)})
		if result.Action != ActionNone {
			t.Errorf("HandleKey(%q) with disabled vim should return ActionNone, got %v", k, result.Action)
		}
	}
}

func TestHandleKeyNormalModeNavigation(t *testing.T) {
	h := NewHandler(true)

	tests := []struct {
		key      string
		expected Result
	}{
		{"h", Result{Action: ActionHandled, CursorMove: -1}},
		{"l", Result{Action: ActionHandled, CursorMove: 1}},
		{"j", Result{Action: ActionHandled, CursorLine: 1}},
		{"k", Result{Action: ActionHandled, CursorLine: -1}},
		{"0", Result{Action: ActionHandled, GotoStart: true}},
		{"$", Result{Action: ActionHandled, GotoEnd: true}},
		{"w", Result{Action: ActionHandled, WordForward: true}},
		{"b", Result{Action: ActionHandled, WordBackward: true}},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			h.Mode = ModeNormal // Reset to normal mode
			result := h.HandleKey(keyMsg(tt.key))

			if result.Action != tt.expected.Action {
				t.Errorf("HandleKey(%q).Action = %v, want %v", tt.key, result.Action, tt.expected.Action)
			}
			if result.CursorMove != tt.expected.CursorMove {
				t.Errorf("HandleKey(%q).CursorMove = %v, want %v", tt.key, result.CursorMove, tt.expected.CursorMove)
			}
			if result.CursorLine != tt.expected.CursorLine {
				t.Errorf("HandleKey(%q).CursorLine = %v, want %v", tt.key, result.CursorLine, tt.expected.CursorLine)
			}
			if result.GotoStart != tt.expected.GotoStart {
				t.Errorf("HandleKey(%q).GotoStart = %v, want %v", tt.key, result.GotoStart, tt.expected.GotoStart)
			}
			if result.GotoEnd != tt.expected.GotoEnd {
				t.Errorf("HandleKey(%q).GotoEnd = %v, want %v", tt.key, result.GotoEnd, tt.expected.GotoEnd)
			}
			if result.WordForward != tt.expected.WordForward {
				t.Errorf("HandleKey(%q).WordForward = %v, want %v", tt.key, result.WordForward, tt.expected.WordForward)
			}
			if result.WordBackward != tt.expected.WordBackward {
				t.Errorf("HandleKey(%q).WordBackward = %v, want %v", tt.key, result.WordBackward, tt.expected.WordBackward)
			}
		})
	}
}

func TestHandleKeyNormalModeInsert(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantMode Mode
		wantMove int
		wantEnd  bool
		wantStart bool
	}{
		{"i enters insert at cursor", "i", ModeInsert, 0, false, false},
		{"a enters insert after cursor", "a", ModeInsert, 1, false, false},
		{"A enters insert at end", "A", ModeInsert, 0, true, false},
		{"I enters insert at start", "I", ModeInsert, 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(true)
			h.Mode = ModeNormal

			result := h.HandleKey(keyMsg(tt.key))

			if h.Mode != tt.wantMode {
				t.Errorf("After %q, mode = %v, want %v", tt.key, h.Mode, tt.wantMode)
			}
			if result.Action != ActionEnterInsert {
				t.Errorf("HandleKey(%q).Action = %v, want ActionEnterInsert", tt.key, result.Action)
			}
			if result.CursorMove != tt.wantMove {
				t.Errorf("HandleKey(%q).CursorMove = %v, want %v", tt.key, result.CursorMove, tt.wantMove)
			}
			if result.GotoEnd != tt.wantEnd {
				t.Errorf("HandleKey(%q).GotoEnd = %v, want %v", tt.key, result.GotoEnd, tt.wantEnd)
			}
			if result.GotoStart != tt.wantStart {
				t.Errorf("HandleKey(%q).GotoStart = %v, want %v", tt.key, result.GotoStart, tt.wantStart)
			}
		})
	}
}

func TestHandleKeyNormalModeEditing(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantDelete int
		wantLine   bool
		wantYank   bool
		wantPaste  bool
	}{
		{"x deletes character", "x", 1, false, false, false},
		{"d deletes line", "d", 0, true, false, false},
		{"y yanks line", "y", 0, false, true, false},
		{"p pastes", "p", 0, false, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewHandler(true)
			h.Mode = ModeNormal

			result := h.HandleKey(keyMsg(tt.key))

			if result.Action != ActionHandled {
				t.Errorf("HandleKey(%q).Action = %v, want ActionHandled", tt.key, result.Action)
			}
			if result.DeleteCount != tt.wantDelete {
				t.Errorf("HandleKey(%q).DeleteCount = %v, want %v", tt.key, result.DeleteCount, tt.wantDelete)
			}
			if result.DeleteLine != tt.wantLine {
				t.Errorf("HandleKey(%q).DeleteLine = %v, want %v", tt.key, result.DeleteLine, tt.wantLine)
			}
			if result.YankLine != tt.wantYank {
				t.Errorf("HandleKey(%q).YankLine = %v, want %v", tt.key, result.YankLine, tt.wantYank)
			}
			if result.Paste != tt.wantPaste {
				t.Errorf("HandleKey(%q).Paste = %v, want %v", tt.key, result.Paste, tt.wantPaste)
			}
		})
	}
}

func TestHandleKeyInsertMode(t *testing.T) {
	h := NewHandler(true)
	h.Mode = ModeInsert

	// Escape should exit insert mode
	result := h.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if result.Action != ActionEnterNormal {
		t.Errorf("Escape in insert mode should return ActionEnterNormal, got %v", result.Action)
	}
	if h.Mode != ModeNormal {
		t.Errorf("After Escape, mode should be ModeNormal, got %v", h.Mode)
	}

	// Other keys should pass through
	h.Mode = ModeInsert
	result = h.HandleKey(keyMsg("a"))
	if result.Action != ActionNone {
		t.Errorf("Regular key in insert mode should return ActionNone, got %v", result.Action)
	}
}

func TestHandleKeyNormalModeUnknown(t *testing.T) {
	h := NewHandler(true)
	h.Mode = ModeNormal

	// Unknown keys should be handled (not passed through) in normal mode
	result := h.HandleKey(keyMsg("z"))
	if result.Action != ActionHandled {
		t.Errorf("Unknown key in normal mode should be handled (not passed through), got %v", result.Action)
	}
}

func TestModeSwitch(t *testing.T) {
	h := NewHandler(true)

	// Start in normal mode
	if h.Mode != ModeNormal {
		t.Error("Should start in ModeNormal")
	}

	// Press 'i' to enter insert mode
	h.HandleKey(keyMsg("i"))
	if h.Mode != ModeInsert {
		t.Error("After 'i', should be in ModeInsert")
	}

	// Press Escape to return to normal mode
	h.HandleKey(tea.KeyMsg{Type: tea.KeyEscape})
	if h.Mode != ModeNormal {
		t.Error("After Escape, should be in ModeNormal")
	}

	// Press 'a' to enter insert mode
	h.HandleKey(keyMsg("a"))
	if h.Mode != ModeInsert {
		t.Error("After 'a', should be in ModeInsert")
	}
}

func TestClipboard(t *testing.T) {
	h := NewHandler(true)

	// Set clipboard directly
	h.Clipboard = "test content"

	// Verify paste result references clipboard
	result := h.HandleKey(keyMsg("p"))
	if !result.Paste {
		t.Error("Paste key should set Paste=true")
	}
}

func TestIsWordChar(t *testing.T) {
	tests := []struct {
		r    rune
		want bool
	}{
		{'a', true},
		{'Z', true},
		{'5', true},
		{'_', true},
		{' ', false},
		{'-', false},
		{'.', false},
		{'!', false},
	}

	for _, tt := range tests {
		got := IsWordChar(tt.r)
		if got != tt.want {
			t.Errorf("IsWordChar(%q) = %v, want %v", tt.r, got, tt.want)
		}
	}
}

func TestAbs(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{5, 5},
		{-5, 5},
		{0, 0},
		{-1, 1},
	}

	for _, tt := range tests {
		got := abs(tt.n)
		if got != tt.want {
			t.Errorf("abs(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

// keyMsg creates a KeyMsg from a string
func keyMsg(s string) tea.KeyMsg {
	if len(s) == 1 {
		return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
}
