// Package errors provides structured error types for fest CLI.
package errors

import (
	"encoding/json"
	"fmt"
)

// Error codes for categorization.
const (
	ErrCodeNotFound   = "NOT_FOUND"
	ErrCodeValidation = "VALIDATION"
	ErrCodeIO         = "IO"
	ErrCodeConfig     = "CONFIG"
	ErrCodeTemplate   = "TEMPLATE"
	ErrCodeParse      = "PARSE"
	ErrCodeInternal   = "INTERNAL"
	ErrCodePermission = "PERMISSION"
)

// Standard hints for common error scenarios.
const (
	HintFestivalNotFound     = "Navigate to a festival directory or run 'fest show all' to see available festivals"
	HintPhaseNotFound        = "Run 'fest status list --type phase' to see available phases"
	HintSequenceNotFound     = "Run 'fest status list --type sequence' to see available sequences"
	HintCreateFestival       = "Run 'fest create festival' or 'fest tui' to create a new festival"
	HintCheckPath            = "Check the path and try again, or run from a different directory"
	HintCheckConfig          = "Check your fest.yaml configuration for syntax errors"
	HintCheckTemplate        = "Run 'fest validate' to check for template issues"
	HintRunInit              = "Run 'fest init' to initialize a festival workspace"
	HintCheckPermissions     = "Check file/directory permissions and try again"
	HintUseForce             = "Use --force to skip confirmation prompts"
	HintNavigateToFestival   = "Navigate to a festival directory first"
	HintUseInteractiveMode   = "Use 'fest tui' for interactive mode"
	HintCheckStatus          = "Run 'fest status' to see current location and status"
	HintSeeHelp              = "Run 'fest help' for more information"
	HintUnderstandMethodology = "Run 'fest understand methodology' to learn about the festival structure"
)

// Error is a structured error type with code, context, and chain support.
type Error struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Op      string                 `json:"op,omitempty"`
	Err     error                  `json:"-"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
	Hint    string                 `json:"hint,omitempty"` // actionable suggestion
}

// Error returns the error string with context.
func (e *Error) Error() string {
	var msg string
	if e.Op != "" && e.Err != nil {
		msg = fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	} else if e.Op != "" {
		msg = fmt.Sprintf("%s: %s", e.Op, e.Message)
	} else if e.Err != nil {
		msg = fmt.Sprintf("%s: %v", e.Message, e.Err)
	} else {
		msg = e.Message
	}

	if e.Hint != "" {
		msg = fmt.Sprintf("%s\nHint: %s", msg, e.Hint)
	}
	return msg
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// MarshalJSON implements json.Marshaler with full error chain.
func (e *Error) MarshalJSON() ([]byte, error) {
	type errorJSON struct {
		Code    string                 `json:"code"`
		Message string                 `json:"message"`
		Op      string                 `json:"op,omitempty"`
		Cause   string                 `json:"cause,omitempty"`
		Hint    string                 `json:"hint,omitempty"`
		Fields  map[string]interface{} `json:"fields,omitempty"`
	}

	ej := errorJSON{
		Code:    e.Code,
		Message: e.Message,
		Op:      e.Op,
		Hint:    e.Hint,
		Fields:  e.Fields,
	}
	if e.Err != nil {
		ej.Cause = e.Err.Error()
	}

	return json.Marshal(ej)
}

// New creates a new error with a message.
func New(message string) *Error {
	return &Error{
		Code:    ErrCodeInternal,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with a message.
func Wrap(err error, message string) *Error {
	return &Error{
		Code:    ErrCodeInternal,
		Message: message,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// Wrapf wraps an existing error with a formatted message.
func Wrapf(err error, format string, args ...interface{}) *Error {
	return &Error{
		Code:    ErrCodeInternal,
		Message: fmt.Sprintf(format, args...),
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// WithCode sets the error code.
func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

// WithOp sets the operation name.
func (e *Error) WithOp(op string) *Error {
	e.Op = op
	return e
}

// WithField adds a context field.
func (e *Error) WithField(key string, value interface{}) *Error {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// WithFields adds multiple context fields.
func (e *Error) WithFields(fields map[string]interface{}) *Error {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	for k, v := range fields {
		e.Fields[k] = v
	}
	return e
}

// WithHint adds an actionable suggestion to the error.
func (e *Error) WithHint(hint string) *Error {
	e.Hint = hint
	return e
}

// WithHintf adds a formatted actionable suggestion to the error.
func (e *Error) WithHintf(format string, args ...interface{}) *Error {
	e.Hint = fmt.Sprintf(format, args...)
	return e
}

// NotFound creates a NOT_FOUND error.
func NotFound(resource string) *Error {
	return &Error{
		Code:    ErrCodeNotFound,
		Message: fmt.Sprintf("%s not found", resource),
		Fields:  map[string]interface{}{"resource": resource},
	}
}

// Validation creates a VALIDATION error.
func Validation(message string) *Error {
	return &Error{
		Code:    ErrCodeValidation,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
}

// IO creates an IO error.
func IO(op string, err error) *Error {
	return &Error{
		Code:    ErrCodeIO,
		Message: "I/O operation failed",
		Op:      op,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// Config creates a CONFIG error.
func Config(message string) *Error {
	return &Error{
		Code:    ErrCodeConfig,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
}

// Template creates a TEMPLATE error.
func Template(message string) *Error {
	return &Error{
		Code:    ErrCodeTemplate,
		Message: message,
		Fields:  make(map[string]interface{}),
	}
}

// Parse creates a PARSE error.
func Parse(message string, err error) *Error {
	return &Error{
		Code:    ErrCodeParse,
		Message: message,
		Err:     err,
		Fields:  make(map[string]interface{}),
	}
}

// Code extracts the error code from any error.
// Returns ErrCodeInternal if error is not a structured Error.
func Code(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(*Error); ok {
		return e.Code
	}
	return ErrCodeInternal
}

// Is checks if the error has the given code.
func Is(err error, code string) bool {
	return Code(err) == code
}
