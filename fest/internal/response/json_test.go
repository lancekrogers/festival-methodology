package response

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{
			name: "simple struct",
			input: struct {
				OK     bool   `json:"ok"`
				Action string `json:"action"`
			}{OK: true, Action: "test"},
		},
		{
			name: "map",
			input: map[string]interface{}{
				"ok":     true,
				"action": "test",
			},
		},
		{
			name:  "response type",
			input: Response{OK: true, Action: "create_festival"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Encode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Verify output is valid JSON
			var result map[string]interface{}
			if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
				t.Errorf("Encode() produced invalid JSON: %v", err)
			}
		})
	}
}

func TestEmitter_Success(t *testing.T) {
	var buf bytes.Buffer
	e := NewEmitter(&buf)

	err := e.Success("test_action", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("Success() error = %v", err)
	}

	var result Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !result.OK {
		t.Error("Expected OK to be true")
	}
	if result.Action != "test_action" {
		t.Errorf("Action = %q, want %q", result.Action, "test_action")
	}
	if len(result.Errors) != 0 {
		t.Error("Expected no errors")
	}
}

func TestEmitter_Failure(t *testing.T) {
	var buf bytes.Buffer
	e := NewEmitter(&buf)

	err := e.Failure("test_action", errors.New("something went wrong"))
	if err != nil {
		t.Fatalf("Failure() error = %v", err)
	}

	var result Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.OK {
		t.Error("Expected OK to be false")
	}
	if result.Action != "test_action" {
		t.Errorf("Action = %q, want %q", result.Action, "test_action")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Message != "something went wrong" {
		t.Errorf("Error message = %q, want %q", result.Errors[0].Message, "something went wrong")
	}
}

func TestEmitter_SuccessWithWarnings(t *testing.T) {
	var buf bytes.Buffer
	e := NewEmitter(&buf)

	warnings := []string{"warning 1", "warning 2"}
	err := e.SuccessWithWarnings("test_action", nil, warnings)
	if err != nil {
		t.Fatalf("SuccessWithWarnings() error = %v", err)
	}

	var result Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if !result.OK {
		t.Error("Expected OK to be true")
	}
	if len(result.Warnings) != 2 {
		t.Errorf("Expected 2 warnings, got %d", len(result.Warnings))
	}
}

func TestEmitter_FailureWithCode(t *testing.T) {
	var buf bytes.Buffer
	e := NewEmitter(&buf)

	err := e.FailureWithCode("test_action", "INVALID_INPUT", "name is required")
	if err != nil {
		t.Fatalf("FailureWithCode() error = %v", err)
	}

	var result Response
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if result.OK {
		t.Error("Expected OK to be false")
	}
	if len(result.Errors) != 1 {
		t.Fatalf("Expected 1 error, got %d", len(result.Errors))
	}
	if result.Errors[0].Code != "INVALID_INPUT" {
		t.Errorf("Error code = %q, want %q", result.Errors[0].Code, "INVALID_INPUT")
	}
}

func TestExtractErrors(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantCode  string
		wantMsg   string
		wantCount int
	}{
		{
			name:      "nil error",
			err:       nil,
			wantCount: 0,
		},
		{
			name:      "simple error",
			err:       errors.New("test error"),
			wantCode:  "error",
			wantMsg:   "test error",
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := ExtractErrors(tt.err)
			if len(errs) != tt.wantCount {
				t.Errorf("ExtractErrors() returned %d errors, want %d", len(errs), tt.wantCount)
				return
			}
			if tt.wantCount > 0 {
				if errs[0].Code != tt.wantCode {
					t.Errorf("Error code = %q, want %q", errs[0].Code, tt.wantCode)
				}
				if errs[0].Message != tt.wantMsg {
					t.Errorf("Error message = %q, want %q", errs[0].Message, tt.wantMsg)
				}
			}
		})
	}
}

func TestConvenienceFunctions(t *testing.T) {
	t.Run("EmitSuccess", func(t *testing.T) {
		var buf bytes.Buffer
		err := EmitSuccess(&buf, "action", nil)
		if err != nil {
			t.Errorf("EmitSuccess() error = %v", err)
		}
	})

	t.Run("EmitFailure", func(t *testing.T) {
		var buf bytes.Buffer
		err := EmitFailure(&buf, "action", errors.New("fail"))
		if err != nil {
			t.Errorf("EmitFailure() error = %v", err)
		}
	})
}

func TestJSONFormatting(t *testing.T) {
	var buf bytes.Buffer
	err := Encode(&buf, Response{OK: true, Action: "test"})
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	// Verify pretty printing (indentation)
	output := buf.String()
	if output[0] != '{' {
		t.Error("Expected JSON to start with {")
	}
	// Check for newlines (pretty printed)
	if !bytes.Contains(buf.Bytes(), []byte("\n")) {
		t.Error("Expected pretty-printed JSON with newlines")
	}
}
