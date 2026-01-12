//go:build !no_charm

package ui

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/term"
)

// SpinnerState represents the current state of a managed spinner.
type SpinnerState int

const (
	SpinnerIdle SpinnerState = iota
	SpinnerRunning
	SpinnerStopped
)

// ManagedSpinner provides a terminal spinner with Start/Stop/UpdateMessage methods.
// It automatically handles non-TTY environments by disabling animation.
type ManagedSpinner struct {
	message string
	frame   int
	state   SpinnerState
	isTTY   bool
	mu      sync.Mutex
	done    chan struct{}
	wg      sync.WaitGroup
}

// NewSpinner creates a new managed spinner with the given initial message.
func NewSpinner(message string) *ManagedSpinner {
	return &ManagedSpinner{
		message: message,
		state:   SpinnerIdle,
		isTTY:   term.IsTerminal(int(os.Stdout.Fd())),
		done:    make(chan struct{}),
	}
}

// Start begins the spinner animation in a background goroutine.
// For non-TTY environments, it prints the message once without animation.
func (s *ManagedSpinner) Start() {
	s.mu.Lock()
	if s.state == SpinnerRunning {
		s.mu.Unlock()
		return
	}
	s.state = SpinnerRunning
	s.done = make(chan struct{})
	s.mu.Unlock()

	if !s.isTTY {
		// Non-TTY: just print the message once
		fmt.Printf("%s...\n", s.message)
		return
	}

	s.wg.Add(1)
	go s.run()
}

// run is the internal goroutine that animates the spinner.
func (s *ManagedSpinner) run() {
	defer s.wg.Done()

	ticker := time.NewTicker(80 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.done:
			// Clear the line before exiting
			fmt.Print("\r\033[K")
			return
		case <-ticker.C:
			s.mu.Lock()
			frame := s.frame
			msg := s.message
			s.frame++
			s.mu.Unlock()

			// Print spinner frame and message
			spinnerChar := Spinner(frame)
			fmt.Printf("\r%s %s", spinnerChar, msg)
		}
	}
}

// UpdateMessage changes the spinner's message while it's running.
func (s *ManagedSpinner) UpdateMessage(message string) {
	s.mu.Lock()
	s.message = message
	s.mu.Unlock()

	if !s.isTTY && s.state == SpinnerRunning {
		// Non-TTY: print the new message
		fmt.Printf("%s...\n", message)
	}
}

// Stop halts the spinner animation and clears the line.
// Optionally prints a final message.
func (s *ManagedSpinner) Stop() {
	s.mu.Lock()
	if s.state != SpinnerRunning {
		s.mu.Unlock()
		return
	}
	s.state = SpinnerStopped
	s.mu.Unlock()

	if s.isTTY {
		close(s.done)
		s.wg.Wait()
	}
}

// StopWithMessage stops the spinner and prints a final message.
func (s *ManagedSpinner) StopWithMessage(message string) {
	s.Stop()
	if message != "" {
		fmt.Println(message)
	}
}

// StopSuccess stops the spinner and prints a success message with checkmark.
func (s *ManagedSpinner) StopSuccess(message string) {
	s.StopWithMessage(SuccessIcon + " " + message)
}

// StopError stops the spinner and prints an error message with X mark.
func (s *ManagedSpinner) StopError(message string) {
	s.StopWithMessage(ErrorIcon + " " + message)
}

// StopWarning stops the spinner and prints a warning message.
func (s *ManagedSpinner) StopWarning(message string) {
	s.StopWithMessage(WarningIcon + " " + message)
}

// Icon constants for spinner completion messages.
const (
	SuccessIcon = "✓"
	ErrorIcon   = "✗"
	WarningIcon = "⚠"
)

// RunWithSpinner runs a function with a spinner showing progress.
// Returns the result of the function and any error.
// The spinner is automatically stopped when the function completes.
func RunWithSpinner[T any](ctx context.Context, message string, fn func() (T, error)) (T, error) {
	spinner := NewSpinner(message)
	spinner.Start()

	result, err := fn()

	if err != nil {
		spinner.StopError(fmt.Sprintf("%s: %v", message, err))
	} else {
		spinner.StopSuccess(message)
	}

	return result, err
}

// RunWithSpinnerNoResult runs a function with a spinner for operations that don't return a value.
func RunWithSpinnerNoResult(ctx context.Context, message string, fn func() error) error {
	spinner := NewSpinner(message)
	spinner.Start()

	err := fn()

	if err != nil {
		spinner.StopError(fmt.Sprintf("%s: %v", message, err))
	} else {
		spinner.StopSuccess(message)
	}

	return err
}
