package frontmatter

import (
	"strings"
	"testing"
	"time"
)

func TestInject(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		fm          *Frontmatter
		wantContain []string
	}{
		{
			name:    "inject into content without frontmatter",
			content: "# Task Title\n\nContent here",
			fm: &Frontmatter{
				Type:    TypeTask,
				ID:      "01_test",
				Status:  StatusPending,
				Created: time.Date(2025, 1, 15, 10, 0, 0, 0, time.UTC),
			},
			wantContain: []string{
				"---\n",
				"fest_type: task",
				"fest_id: 01_test",
				"fest_status: pending",
				"# Task Title",
			},
		},
		{
			name:    "inject with all fields",
			content: "# Festival\n\nDescription",
			fm: &Frontmatter{
				Type:     TypeFestival,
				ID:       "my-festival",
				Name:     "My Festival",
				Status:   StatusActive,
				Priority: PriorityHigh,
				Created:  time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Tags:     []string{"v1", "important"},
			},
			wantContain: []string{
				"fest_type: festival",
				"fest_name: My Festival",
				"fest_priority: high",
				"# Festival",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Inject([]byte(tc.content), tc.fm)
			if err != nil {
				t.Fatalf("Inject() error = %v", err)
			}

			resultStr := string(result)
			for _, want := range tc.wantContain {
				if !strings.Contains(resultStr, want) {
					t.Errorf("Result should contain %q\nGot:\n%s", want, resultStr)
				}
			}
		})
	}
}

func TestBuilder(t *testing.T) {
	t.Run("task builder", func(t *testing.T) {
		fm := NewBuilder(TypeTask).
			ID("01_test").
			Name("Test Task").
			Parent("01_seq").
			Order(1).
			Status(StatusPending).
			Build()

		if fm.Type != TypeTask {
			t.Errorf("Type = %q, want task", fm.Type)
		}
		if fm.ID != "01_test" {
			t.Errorf("ID = %q, want 01_test", fm.ID)
		}
		if fm.Parent != "01_seq" {
			t.Errorf("Parent = %q, want 01_seq", fm.Parent)
		}
		if fm.Order != 1 {
			t.Errorf("Order = %d, want 1", fm.Order)
		}
	})

	t.Run("festival builder with tags", func(t *testing.T) {
		fm := NewBuilder(TypeFestival).
			ID("test-fest").
			Name("Test Festival").
			Priority(PriorityHigh).
			Tags("urgent", "v1").
			Build()

		if fm.Type != TypeFestival {
			t.Errorf("Type = %q, want festival", fm.Type)
		}
		if fm.Priority != PriorityHigh {
			t.Errorf("Priority = %q, want high", fm.Priority)
		}
		if len(fm.Tags) != 2 {
			t.Errorf("Tags len = %d, want 2", len(fm.Tags))
		}
	})

	t.Run("builder sets default status", func(t *testing.T) {
		fm := NewBuilder(TypeSequence).
			ID("01_seq").
			Build()

		// Builder sets default status based on type
		if fm.Status != StatusPending {
			t.Errorf("Status = %q, want pending", fm.Status)
		}
	})
}

func TestMarshal(t *testing.T) {
	fm := &Frontmatter{
		Type:    TypeTask,
		ID:      "01_test",
		Name:    "Test Task",
		Status:  StatusPending,
		Created: time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
	}

	data, err := Marshal(fm)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	result := string(data)
	if !strings.Contains(result, "fest_type: task") {
		t.Error("Should contain fest_type: task")
	}
	if !strings.Contains(result, "fest_id: 01_test") {
		t.Error("Should contain fest_id: 01_test")
	}
}

func TestFormatBlock(t *testing.T) {
	fm := &Frontmatter{
		Type:   TypePhase,
		ID:     "001_planning",
		Status: StatusActive,
	}

	result, err := FormatBlock(fm)
	if err != nil {
		t.Fatalf("FormatBlock() error = %v", err)
	}

	if !strings.HasPrefix(result, "---\n") {
		t.Error("Should start with ---")
	}
	if !strings.HasSuffix(result, "---\n") {
		t.Error("Should end with ---")
	}
}

func TestInjectString(t *testing.T) {
	fm := &Frontmatter{
		Type:    TypeTask,
		ID:      "01_test",
		Status:  StatusPending,
		Created: time.Now(),
	}

	result, err := InjectString("# Content", fm)
	if err != nil {
		t.Fatalf("InjectString() error = %v", err)
	}

	if !strings.Contains(result, "fest_type: task") {
		t.Error("Should contain fest_type")
	}
	if !strings.Contains(result, "# Content") {
		t.Error("Should contain original content")
	}
}

func TestFormat(t *testing.T) {
	fm := &Frontmatter{
		Type:   TypeGate,
		ID:     "gate_review",
		Status: StatusPending,
	}

	result, err := Format(fm)
	if err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	if !strings.Contains(result, "fest_type: gate") {
		t.Error("Should contain fest_type: gate")
	}
}
