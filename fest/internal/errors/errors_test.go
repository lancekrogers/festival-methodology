package errors

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name:     "message only",
			err:      New("something failed"),
			expected: "something failed",
		},
		{
			name:     "with op",
			err:      New("something failed").WithOp("LoadConfig"),
			expected: "LoadConfig: something failed",
		},
		{
			name:     "with wrapped error",
			err:      Wrap(errors.New("underlying"), "operation failed"),
			expected: "operation failed: underlying",
		},
		{
			name:     "with op and wrapped error",
			err:      Wrap(errors.New("underlying"), "operation failed").WithOp("LoadConfig"),
			expected: "LoadConfig: operation failed: underlying",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.expected {
				t.Errorf("Error() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestError_Unwrap(t *testing.T) {
	underlying := errors.New("underlying error")
	err := Wrap(underlying, "wrapped")

	if err.Unwrap() != underlying {
		t.Error("Unwrap() should return the underlying error")
	}
}

func TestError_WithCode(t *testing.T) {
	err := New("test").WithCode(ErrCodeNotFound)

	if err.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeNotFound)
	}
}

func TestError_WithField(t *testing.T) {
	err := New("test").WithField("path", "/some/path")

	if err.Fields["path"] != "/some/path" {
		t.Errorf("Fields[path] = %v, want %q", err.Fields["path"], "/some/path")
	}
}

func TestError_WithFields(t *testing.T) {
	err := New("test").WithFields(map[string]interface{}{
		"path": "/some/path",
		"line": 42,
	})

	if err.Fields["path"] != "/some/path" {
		t.Errorf("Fields[path] = %v, want %q", err.Fields["path"], "/some/path")
	}
	if err.Fields["line"] != 42 {
		t.Errorf("Fields[line] = %v, want %d", err.Fields["line"], 42)
	}
}

func TestError_MarshalJSON(t *testing.T) {
	err := Wrap(errors.New("underlying"), "operation failed").
		WithCode(ErrCodeIO).
		WithOp("ReadFile").
		WithField("path", "/test/file.txt")

	data, jsonErr := json.Marshal(err)
	if jsonErr != nil {
		t.Fatalf("MarshalJSON() error = %v", jsonErr)
	}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(data, &result); jsonErr != nil {
		t.Fatalf("Unmarshal error = %v", jsonErr)
	}

	if result["code"] != ErrCodeIO {
		t.Errorf("code = %v, want %q", result["code"], ErrCodeIO)
	}
	if result["message"] != "operation failed" {
		t.Errorf("message = %v, want %q", result["message"], "operation failed")
	}
	if result["op"] != "ReadFile" {
		t.Errorf("op = %v, want %q", result["op"], "ReadFile")
	}
	if result["cause"] != "underlying" {
		t.Errorf("cause = %v, want %q", result["cause"], "underlying")
	}

	fields, ok := result["fields"].(map[string]interface{})
	if !ok {
		t.Fatal("fields should be a map")
	}
	if fields["path"] != "/test/file.txt" {
		t.Errorf("fields.path = %v, want %q", fields["path"], "/test/file.txt")
	}
}

func TestNotFound(t *testing.T) {
	err := NotFound("template")

	if err.Code != ErrCodeNotFound {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeNotFound)
	}
	if err.Fields["resource"] != "template" {
		t.Errorf("Fields[resource] = %v, want %q", err.Fields["resource"], "template")
	}
}

func TestValidation(t *testing.T) {
	err := Validation("invalid input")

	if err.Code != ErrCodeValidation {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeValidation)
	}
}

func TestIO(t *testing.T) {
	underlying := errors.New("permission denied")
	err := IO("ReadFile", underlying)

	if err.Code != ErrCodeIO {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeIO)
	}
	if err.Op != "ReadFile" {
		t.Errorf("Op = %q, want %q", err.Op, "ReadFile")
	}
	if err.Unwrap() != underlying {
		t.Error("should wrap underlying error")
	}
}

func TestCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: "",
		},
		{
			name:     "structured error",
			err:      New("test").WithCode(ErrCodeValidation),
			expected: ErrCodeValidation,
		},
		{
			name:     "plain error",
			err:      errors.New("plain error"),
			expected: ErrCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Code(tt.err); got != tt.expected {
				t.Errorf("Code() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestIs(t *testing.T) {
	err := NotFound("resource")

	if !Is(err, ErrCodeNotFound) {
		t.Error("Is() should return true for matching code")
	}
	if Is(err, ErrCodeIO) {
		t.Error("Is() should return false for non-matching code")
	}
}

func TestWrapf(t *testing.T) {
	underlying := errors.New("underlying")
	err := Wrapf(underlying, "failed to process %s", "item")

	if err.Message != "failed to process item" {
		t.Errorf("Message = %q, want %q", err.Message, "failed to process item")
	}
	if err.Unwrap() != underlying {
		t.Error("should wrap underlying error")
	}
}
