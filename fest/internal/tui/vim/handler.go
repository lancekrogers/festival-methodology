package vim

import (
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
)

// Handler processes key events in vim mode, managing mode switching and keybindings.
type Handler struct {
	// Mode is the current vim mode (Normal or Insert).
	Mode Mode
	// Enabled controls whether vim mode is active.
	Enabled bool
	// Clipboard holds yanked content for paste operations.
	Clipboard string
}

// NewHandler creates a new vim key handler.
func NewHandler(enabled bool) *Handler {
	return &Handler{
		Mode:    ModeNormal,
		Enabled: enabled,
	}
}

// KeyAction represents the result of handling a key event.
type KeyAction int

const (
	// ActionNone means the key was not handled (pass through).
	ActionNone KeyAction = iota
	// ActionHandled means the key was handled, don't pass to textarea.
	ActionHandled
	// ActionEnterInsert means switch to insert mode.
	ActionEnterInsert
	// ActionEnterNormal means switch to normal mode.
	ActionEnterNormal
)

// Result contains the action to take and any cursor movement.
type Result struct {
	Action KeyAction
	// CursorMove is the number of positions to move cursor (negative = left, positive = right).
	CursorMove int
	// CursorLine is the number of lines to move cursor (negative = up, positive = down).
	CursorLine int
	// DeleteCount is the number of characters to delete.
	DeleteCount int
	// DeleteLine indicates whether to delete the entire current line.
	DeleteLine bool
	// YankLine indicates whether to yank the current line.
	YankLine bool
	// Paste indicates whether to paste from clipboard.
	Paste bool
	// GotoStart indicates whether to go to start of line.
	GotoStart bool
	// GotoEnd indicates whether to go to end of line.
	GotoEnd bool
	// WordForward indicates whether to move word forward.
	WordForward bool
	// WordBackward indicates whether to move word backward.
	WordBackward bool
}

// HandleKey processes a key event based on the current mode.
// Returns a Result indicating what action to take.
func (h *Handler) HandleKey(msg tea.KeyMsg) Result {
	if !h.Enabled {
		return Result{Action: ActionNone}
	}

	switch h.Mode {
	case ModeNormal:
		return h.handleNormalMode(msg)
	case ModeInsert:
		return h.handleInsertMode(msg)
	default:
		return Result{Action: ActionNone}
	}
}

// handleNormalMode processes keys in normal (navigation) mode.
func (h *Handler) handleNormalMode(msg tea.KeyMsg) Result {
	switch msg.String() {
	// Mode switching
	case "i":
		h.Mode = ModeInsert
		return Result{Action: ActionEnterInsert}
	case "a":
		h.Mode = ModeInsert
		return Result{Action: ActionEnterInsert, CursorMove: 1}
	case "A":
		h.Mode = ModeInsert
		return Result{Action: ActionEnterInsert, GotoEnd: true}
	case "I":
		h.Mode = ModeInsert
		return Result{Action: ActionEnterInsert, GotoStart: true}

	// Navigation
	case "h", "left":
		return Result{Action: ActionHandled, CursorMove: -1}
	case "l", "right":
		return Result{Action: ActionHandled, CursorMove: 1}
	case "j", "down":
		return Result{Action: ActionHandled, CursorLine: 1}
	case "k", "up":
		return Result{Action: ActionHandled, CursorLine: -1}
	case "w":
		return Result{Action: ActionHandled, WordForward: true}
	case "b":
		return Result{Action: ActionHandled, WordBackward: true}
	case "0":
		return Result{Action: ActionHandled, GotoStart: true}
	case "$":
		return Result{Action: ActionHandled, GotoEnd: true}

	// Editing
	case "x":
		return Result{Action: ActionHandled, DeleteCount: 1}
	case "d":
		// Wait for second 'd' for line delete (simplified: just delete line)
		return Result{Action: ActionHandled, DeleteLine: true}
	case "y":
		// Simplified: yank line
		return Result{Action: ActionHandled, YankLine: true}
	case "p":
		return Result{Action: ActionHandled, Paste: true}

	// Escape stays in normal mode
	case "esc":
		return Result{Action: ActionHandled}

	default:
		// Don't pass unknown keys through in normal mode
		return Result{Action: ActionHandled}
	}
}

// handleInsertMode processes keys in insert (text entry) mode.
func (h *Handler) handleInsertMode(msg tea.KeyMsg) Result {
	switch msg.String() {
	case "esc":
		h.Mode = ModeNormal
		return Result{Action: ActionEnterNormal}
	default:
		// Pass all other keys through to the textarea for normal text entry
		return Result{Action: ActionNone}
	}
}

// ApplyToTextarea applies the result to a textarea model.
// Returns the updated model and any command to execute.
func (h *Handler) ApplyToTextarea(ta textarea.Model, result Result, content string) (textarea.Model, tea.Cmd) {
	switch result.Action {
	case ActionNone:
		return ta, nil

	case ActionHandled:
		// Apply cursor movements and editing operations
		return h.applyOperations(ta, result, content)

	case ActionEnterInsert:
		// Apply any position changes then switch to insert
		ta, _ = h.applyOperations(ta, result, content)
		return ta, nil

	case ActionEnterNormal:
		return ta, nil

	default:
		return ta, nil
	}
}

// applyOperations applies cursor movements and editing to the textarea.
func (h *Handler) applyOperations(ta textarea.Model, result Result, content string) (textarea.Model, tea.Cmd) {
	// Handle cursor movements
	if result.CursorMove != 0 {
		for i := 0; i < abs(result.CursorMove); i++ {
			if result.CursorMove < 0 {
				ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyLeft})
			} else {
				ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyRight})
			}
		}
	}

	if result.CursorLine != 0 {
		for i := 0; i < abs(result.CursorLine); i++ {
			if result.CursorLine < 0 {
				ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyUp})
			} else {
				ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyDown})
			}
		}
	}

	// Handle line start/end
	if result.GotoStart {
		ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyHome})
	}
	if result.GotoEnd {
		ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyEnd})
	}

	// Handle word movements
	if result.WordForward {
		ta = h.moveWordForward(ta, content)
	}
	if result.WordBackward {
		ta = h.moveWordBackward(ta, content)
	}

	// Handle deletions
	if result.DeleteCount > 0 {
		for i := 0; i < result.DeleteCount; i++ {
			ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyDelete})
		}
	}

	if result.DeleteLine {
		// Go to start of line, select to end, delete
		ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyHome})
		// Delete entire line content using keyboard shortcuts
		ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyCtrlK}) // Kill line
	}

	// Handle yank
	if result.YankLine {
		// Store current line in clipboard
		lines := strings.Split(content, "\n")
		row := ta.Line()
		if row >= 0 && row < len(lines) {
			h.Clipboard = lines[row]
		}
	}

	// Handle paste
	if result.Paste && h.Clipboard != "" {
		// Insert clipboard content by simulating key presses
		for _, r := range h.Clipboard {
			ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		}
	}

	return ta, nil
}

// moveWordForward moves the cursor to the start of the next word.
func (h *Handler) moveWordForward(ta textarea.Model, content string) textarea.Model {
	// Use the built-in word forward key binding
	ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyCtrlRight})
	return ta
}

// moveWordBackward moves the cursor to the start of the previous word.
func (h *Handler) moveWordBackward(ta textarea.Model, content string) textarea.Model {
	// Use the built-in word backward key binding
	ta, _ = ta.Update(tea.KeyMsg{Type: tea.KeyCtrlLeft})
	return ta
}

// abs returns the absolute value of an integer.
func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// VimKeyMap returns a textarea KeyMap customized for vim normal mode.
// This disables most default bindings to prevent conflicts.
func VimKeyMap() textarea.KeyMap {
	return textarea.KeyMap{
		// Disable most keybindings in normal mode - vim handler takes over
		CharacterForward:        key.NewBinding(key.WithKeys("right")),
		CharacterBackward:       key.NewBinding(key.WithKeys("left")),
		DeleteCharacterBackward: key.NewBinding(key.WithDisabled()),
		DeleteCharacterForward:  key.NewBinding(key.WithKeys("delete")),
		DeleteAfterCursor:       key.NewBinding(key.WithKeys("ctrl+k")),
		DeleteBeforeCursor:      key.NewBinding(key.WithDisabled()),
		DeleteWordBackward:      key.NewBinding(key.WithDisabled()),
		DeleteWordForward:       key.NewBinding(key.WithDisabled()),
		InsertNewline:           key.NewBinding(key.WithKeys("enter")),
		LineEnd:                 key.NewBinding(key.WithKeys("end")),
		LineNext:                key.NewBinding(key.WithKeys("down")),
		LinePrevious:            key.NewBinding(key.WithKeys("up")),
		LineStart:               key.NewBinding(key.WithKeys("home")),
		Paste:                   key.NewBinding(key.WithKeys("ctrl+v")),
		WordBackward:            key.NewBinding(key.WithKeys("ctrl+left")),
		WordForward:             key.NewBinding(key.WithKeys("ctrl+right")),
	}
}

// DefaultKeyMap returns the standard textarea KeyMap for insert mode / non-vim mode.
func DefaultKeyMap() textarea.KeyMap {
	return textarea.DefaultKeyMap
}

// IsWordChar returns true if the rune is a word character (letter, digit, underscore).
func IsWordChar(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_'
}
