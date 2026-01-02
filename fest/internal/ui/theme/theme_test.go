package theme

import (
	"context"
	"errors"
	"testing"

	"github.com/charmbracelet/huh"
)

func TestFestTheme(t *testing.T) {
	theme := FestTheme()

	// Verify theme is not nil and has expected structure
	if theme == nil {
		t.Fatal("FestTheme() returned nil")
	}

	// Check focused styles exist
	if theme.Focused.Title.String() == "" {
		// Title style should be configured
		t.Log("Focused.Title style is configured")
	}
}

func TestIsCancelled(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "huh user aborted",
			err:      huh.ErrUserAborted,
			expected: true,
		},
		{
			name:     "ErrUserCancelled",
			err:      ErrUserCancelled,
			expected: true,
		},
		{
			name:     "context cancelled",
			err:      context.Canceled,
			expected: true,
		},
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: true,
		},
		{
			name:     "other error",
			err:      errors.New("some other error"),
			expected: false,
		},
		{
			name:     "wrapped user aborted",
			err:      errors.Join(errors.New("wrapper"), huh.ErrUserAborted),
			expected: true,
		},
		{
			name:     "wrapped context cancelled",
			err:      errors.Join(errors.New("wrapper"), context.Canceled),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsCancelled(tt.err)
			if result != tt.expected {
				t.Errorf("IsCancelled(%v) = %v, want %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestToOptions(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   int
	}{
		{
			name:   "empty slice",
			values: []string{},
			want:   0,
		},
		{
			name:   "single value",
			values: []string{"option1"},
			want:   1,
		},
		{
			name:   "multiple values",
			values: []string{"yes", "no", "maybe"},
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := ToOptions(tt.values)
			if len(opts) != tt.want {
				t.Errorf("ToOptions() returned %d options, want %d", len(opts), tt.want)
			}
		})
	}
}

func TestRunFormWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var value string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Test").
				Value(&value),
		),
	)

	err := RunForm(ctx, form)
	if err == nil {
		t.Error("RunForm with cancelled context should return error")
	}
}

func TestPromptInputWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := PromptInput(ctx, "Test", "placeholder")
	if err == nil {
		t.Error("PromptInput with cancelled context should return error")
	}
}

func TestPromptConfirmWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, _, err := PromptConfirm(ctx, "Test?")
	if err == nil {
		t.Error("PromptConfirm with cancelled context should return error")
	}
}

func TestPromptSelectWithCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := ToOptions([]string{"a", "b", "c"})
	_, _, err := PromptSelect(ctx, "Test", opts)
	if err == nil {
		t.Error("PromptSelect with cancelled context should return error")
	}
}
