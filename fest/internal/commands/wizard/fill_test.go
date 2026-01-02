package wizard

import (
	"context"
	"os"
	"testing"
)

func TestGetEditor(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "falls back to vim when EDITOR not set",
			envValue: "",
			want:     "vim",
		},
		{
			name:     "uses EDITOR env var when set",
			envValue: "nvim",
			want:     "nvim",
		},
		{
			name:     "uses custom editor from env",
			envValue: "code",
			want:     "code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore EDITOR
			origEditor := os.Getenv("EDITOR")
			defer func() {
				if origEditor == "" {
					os.Unsetenv("EDITOR")
				} else {
					os.Setenv("EDITOR", origEditor)
				}
			}()

			if tt.envValue == "" {
				os.Unsetenv("EDITOR")
			} else {
				os.Setenv("EDITOR", tt.envValue)
			}

			got := getEditor(ctx)
			if got != tt.want {
				t.Errorf("getEditor() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGetEditor_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	// Should still return a valid editor even with cancelled context
	// (config.Load will fail but we fall back gracefully)
	got := getEditor(ctx)
	if got == "" {
		t.Error("getEditor() returned empty string with cancelled context")
	}
}
