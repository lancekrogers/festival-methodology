package theme

import (
	"context"
	"errors"

	"github.com/charmbracelet/huh"
	festErrors "github.com/lancekrogers/festival-methodology/fest/internal/errors"
)

// Error codes for theme package.
const (
	ErrCodeCancelled = "CANCELLED"
)

// ErrUserCancelled is returned when user cancels with Ctrl-C or Esc.
var ErrUserCancelled = festErrors.New("operation cancelled by user").WithCode(ErrCodeCancelled)

// RunForm executes a form with the fest theme, context propagation, and proper error handling.
// Context cancellation will interrupt the form. Returns ErrUserCancelled if the user aborts.
func RunForm(ctx context.Context, form *huh.Form) error {
	if err := ctx.Err(); err != nil {
		return festErrors.Wrap(err, "context cancelled").WithOp("RunForm")
	}

	form = form.WithTheme(FestTheme())

	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrUserCancelled
		}
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return festErrors.Wrap(err, "context cancelled").WithOp("RunForm")
		}
		return festErrors.Wrap(err, "form error").WithOp("RunForm")
	}
	return nil
}

// IsCancelled checks if an error indicates user or context cancellation.
// Returns true for ErrUserCancelled, huh.ErrUserAborted, and context cancellation.
func IsCancelled(err error) bool {
	if err == nil {
		return false
	}
	return errors.Is(err, ErrUserCancelled) ||
		errors.Is(err, huh.ErrUserAborted) ||
		errors.Is(err, context.Canceled) ||
		errors.Is(err, context.DeadlineExceeded) ||
		festErrors.Is(err, ErrCodeCancelled)
}

// PromptInput is a convenience function for single-input prompts.
// Returns the value, whether it was skipped (empty or cancelled), and any error.
func PromptInput(ctx context.Context, title, placeholder string) (string, bool, error) {
	if err := ctx.Err(); err != nil {
		return "", false, festErrors.Wrap(err, "context cancelled")
	}

	var value string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title(title).
				Placeholder(placeholder).
				Value(&value),
		),
	)

	if err := RunForm(ctx, form); err != nil {
		if IsCancelled(err) {
			return "", true, nil
		}
		return "", false, err
	}

	return value, value == "", nil
}

// PromptSelect is a convenience function for single-select prompts.
// Returns the selected value, whether it was cancelled, and any error.
func PromptSelect[T comparable](ctx context.Context, title string, options []huh.Option[T]) (T, bool, error) {
	var value T
	if err := ctx.Err(); err != nil {
		return value, false, festErrors.Wrap(err, "context cancelled")
	}

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[T]().
				Title(title).
				Options(options...).
				Value(&value),
		),
	)

	if err := RunForm(ctx, form); err != nil {
		if IsCancelled(err) {
			return value, true, nil
		}
		return value, false, err
	}

	return value, false, nil
}

// PromptConfirm is a convenience function for yes/no prompts.
// Returns the boolean result, whether it was cancelled, and any error.
func PromptConfirm(ctx context.Context, title string) (bool, bool, error) {
	if err := ctx.Err(); err != nil {
		return false, false, festErrors.Wrap(err, "context cancelled")
	}

	var value bool
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewConfirm().
				Title(title).
				Value(&value),
		),
	)

	if err := RunForm(ctx, form); err != nil {
		if IsCancelled(err) {
			return false, true, nil
		}
		return false, false, err
	}

	return value, false, nil
}

// ToOptions converts a slice of strings to huh options.
func ToOptions(values []string) []huh.Option[string] {
	opts := make([]huh.Option[string], 0, len(values))
	for _, v := range values {
		opts = append(opts, huh.NewOption(v, v))
	}
	return opts
}
