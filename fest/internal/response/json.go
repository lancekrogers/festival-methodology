// Package response provides JSON output formatting for fest CLI commands.
// It defines a standard Response type with success/failure states, error handling,
// and warnings, ensuring consistent JSON output across all commands.
package response

import (
	"encoding/json"
	"io"
)

// Response is the standard JSON output structure for all fest commands.
// All commands emitting JSON should use these types for consistency.
type Response struct {
	OK       bool        `json:"ok"`
	Action   string      `json:"action"`
	Data     interface{} `json:"data,omitempty"`
	Errors   []Error     `json:"errors,omitempty"`
	Warnings []string    `json:"warnings,omitempty"`
}

// Error represents a structured error in JSON output.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Path    string `json:"path,omitempty"`
}

// Emitter provides JSON response emission with consistent formatting.
type Emitter struct {
	w io.Writer
}

// NewEmitter creates an Emitter that writes to w.
func NewEmitter(w io.Writer) *Emitter {
	return &Emitter{w: w}
}

// Success emits a successful response with optional data.
func (e *Emitter) Success(action string, data interface{}) error {
	return e.emit(Response{
		OK:     true,
		Action: action,
		Data:   data,
	})
}

// SuccessWithWarnings emits a successful response with warnings.
func (e *Emitter) SuccessWithWarnings(action string, data interface{}, warnings []string) error {
	return e.emit(Response{
		OK:       true,
		Action:   action,
		Data:     data,
		Warnings: warnings,
	})
}

// Failure emits a failure response with error details.
func (e *Emitter) Failure(action string, err error) error {
	return e.emit(Response{
		OK:     false,
		Action: action,
		Errors: ExtractErrors(err),
	})
}

// FailureWithCode emits a failure response with a specific error code.
func (e *Emitter) FailureWithCode(action, code, message string) error {
	return e.emit(Response{
		OK:     false,
		Action: action,
		Errors: []Error{{Code: code, Message: message}},
	})
}

// FailureWithErrors emits a failure response with multiple errors.
func (e *Emitter) FailureWithErrors(action string, errors []Error) error {
	return e.emit(Response{
		OK:     false,
		Action: action,
		Errors: errors,
	})
}

func (e *Emitter) emit(r Response) error {
	enc := json.NewEncoder(e.w)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}

// ExtractErrors converts an error to a slice of Error.
// This function can be extended to handle structured error types.
func ExtractErrors(err error) []Error {
	if err == nil {
		return nil
	}
	return []Error{{
		Code:    "error",
		Message: err.Error(),
	}}
}

// EmitSuccess is a convenience function for one-off emissions to w.
func EmitSuccess(w io.Writer, action string, data interface{}) error {
	return NewEmitter(w).Success(action, data)
}

// EmitFailure is a convenience function for one-off emissions to w.
func EmitFailure(w io.Writer, action string, err error) error {
	return NewEmitter(w).Failure(action, err)
}

// Encode writes any struct as pretty-printed JSON to w.
// Use this for command-specific result types that already have the correct structure.
func Encode(w io.Writer, v interface{}) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
