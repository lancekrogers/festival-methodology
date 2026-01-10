package wizard

import (
	"context"
	"reflect"
	"testing"
)

func TestGetEditorArgs(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		opts  *FillOptions
		files []string
		want  []string
	}{
		{
			name:  "default buffer mode",
			opts:  &FillOptions{},
			files: []string{"file1.md", "file2.md"},
			want:  []string{"file1.md", "file2.md"},
		},
		{
			name:  "buffer mode explicit",
			opts:  &FillOptions{EditorMode: "buffer"},
			files: []string{"file1.md"},
			want:  []string{"file1.md"},
		},
		{
			name:  "tab mode",
			opts:  &FillOptions{EditorMode: "tab"},
			files: []string{"file1.md", "file2.md"},
			want:  []string{"-p", "file1.md", "file2.md"},
		},
		{
			name:  "split mode (backwards compat)",
			opts:  &FillOptions{EditorMode: "split"},
			files: []string{"file1.md"},
			want:  []string{"-O", "file1.md"},
		},
		{
			name:  "hsplit mode",
			opts:  &FillOptions{EditorMode: "hsplit"},
			files: []string{"file1.md", "file2.md"},
			want:  []string{"-o", "file1.md", "file2.md"},
		},
		{
			name: "custom flags override mode",
			opts: &FillOptions{
				EditorMode:  "tab", // Should be ignored
				EditorFlags: []string{"-c", "set nu"},
			},
			files: []string{"file1.md"},
			want:  []string{"-c", "set nu", "file1.md"},
		},
		{
			name:  "invalid mode falls back to buffer",
			opts:  &FillOptions{EditorMode: "invalid"},
			files: []string{"file1.md"},
			want:  []string{"file1.md"},
		},
		{
			name:  "empty files list",
			opts:  &FillOptions{EditorMode: "tab"},
			files: []string{},
			want:  []string{"-p"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getEditorArgs(ctx, tt.files, tt.opts)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getEditorArgs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetEditorArgs_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	opts := &FillOptions{EditorMode: "tab"}
	files := []string{"file1.md"}

	// With cancelled context, should fail fast and return default buffer mode
	got := getEditorArgs(ctx, files, opts)
	want := []string{"file1.md"} // Buffer mode (no flags) when context cancelled

	if !reflect.DeepEqual(got, want) {
		t.Errorf("getEditorArgs() with cancelled context = %v, want %v", got, want)
	}
}

func TestGetEditorArgs_PriorityOrder(t *testing.T) {
	ctx := context.Background()

	// Test 1: EditorFlags has highest priority
	opts := &FillOptions{
		EditorMode:  "tab",
		EditorFlags: []string{"--custom"},
	}
	files := []string{"file.md"}

	got := getEditorArgs(ctx, files, opts)
	want := []string{"--custom", "file.md"}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("EditorFlags priority: got %v, want %v", got, want)
	}

	// Test 2: EditorMode is used when EditorFlags is empty
	opts2 := &FillOptions{
		EditorMode:  "tab",
		EditorFlags: nil,
	}

	got2 := getEditorArgs(ctx, files, opts2)
	want2 := []string{"-p", "file.md"}

	if !reflect.DeepEqual(got2, want2) {
		t.Errorf("EditorMode priority: got %v, want %v", got2, want2)
	}
}
