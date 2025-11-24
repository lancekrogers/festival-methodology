package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create context with custom variables for tests
func newTestContext(vars map[string]interface{}) *Context {
	ctx := NewContext()
	for k, v := range vars {
		ctx.SetCustom(k, v)
	}
	return ctx
}

func TestRenderer_RenderString(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		context  *Context
		expected string
		wantErr  bool
	}{
		{
			name:    "simple variable substitution",
			content: "Hello {{.name}}!",
			context: newTestContext(map[string]interface{}{
				"name": "World",
			}),
			expected: "Hello World!",
			wantErr:  false,
		},
		{
			name:    "multiple variables",
			content: "Festival: {{.festival_name}}\nGoal: {{.goal}}",
			context: newTestContext(map[string]interface{}{
				"festival_name": "my-festival",
				"goal":          "Build awesome things",
			}),
			expected: "Festival: my-festival\nGoal: Build awesome things",
			wantErr:  false,
		},
		{
			name:    "sprig functions - default",
			content: "Name: {{.name | default \"Unknown\"}}",
			context: newTestContext(map[string]interface{}{}),
			expected: "Name: Unknown",
			wantErr:  false,
		},
		{
			name:    "sprig functions - upper",
			content: "Name: {{.name | upper}}",
			context: newTestContext(map[string]interface{}{
				"name": "festival",
			}),
			expected: "Name: FESTIVAL",
			wantErr:  false,
		},
		{
			name:    "preserve guidance markers",
			content: "# Title\n\n[GUIDANCE: Fill this in]\n\n[FILL: Add description]",
			context: newTestContext(map[string]interface{}{}),
			expected: "# Title\n\n[GUIDANCE: Fill this in]\n\n[FILL: Add description]",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderer := NewRenderer()
			output, err := renderer.RenderString(tt.content, tt.context)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestRenderer_Render(t *testing.T) {
	tmpl := &Template{
		Path: "test.md",
		Metadata: &Metadata{
			TemplateID:        "TEST",
			RequiredVariables: []string{"name"},
		},
		Content: "Hello {{.name}}!",
	}

	ctx := newTestContext(map[string]interface{}{
		"name": "World",
	})

	renderer := NewRenderer()
	output, err := renderer.Render(tmpl, ctx)

	require.NoError(t, err)
	assert.Equal(t, "Hello World!", output)
}

func TestValidateTemplate(t *testing.T) {
	tests := []struct {
		name     string
		metadata *Metadata
		context  *Context
		wantErr  bool
	}{
		{
			name: "all required variables present",
			metadata: &Metadata{
				RequiredVariables: []string{"name", "goal"},
			},
			context: newTestContext(map[string]interface{}{
				"name": "test",
				"goal": "test goal",
			}),
			wantErr: false,
		},
		{
			name: "missing required variable",
			metadata: &Metadata{
				RequiredVariables: []string{"name", "goal"},
			},
			context: newTestContext(map[string]interface{}{
				"name": "test",
			}),
			wantErr: true,
		},
		{
			name:     "no metadata",
			metadata: nil,
			context: newTestContext(map[string]interface{}{}),
			wantErr: false,
		},
		{
			name: "optional variables missing",
			metadata: &Metadata{
				RequiredVariables: []string{"name"},
				OptionalVariables: []string{"description"},
			},
			context: newTestContext(map[string]interface{}{
				"name": "test",
			}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpl := &Template{
				Metadata: tt.metadata,
			}

			err := ValidateTemplate(tmpl, tt.context)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckUnrenderedVariables(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		wantUnrendered []string
	}{
		{
			name:       "no unrendered variables",
			content:    "Hello World! This is content.",
			wantUnrendered: []string{},
		},
		{
			name:       "one unrendered variable",
			content:    "Hello {{name}}!",
			wantUnrendered: []string{"name"},
		},
		{
			name:       "multiple unrendered variables",
			content:    "{{festival_name}} - {{goal}} - {{description}}",
			wantUnrendered: []string{"festival_name", "goal", "description"},
		},
		{
			name:       "guidance markers not counted",
			content:    "[GUIDANCE: Fill this in]\n[FILL: Add description]",
			wantUnrendered: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			unrendered := CheckUnrenderedVariables(tt.content)
			assert.ElementsMatch(t, tt.wantUnrendered, unrendered)
		})
	}
}

func TestRenderWithDefaults(t *testing.T) {
	tmpl := &Template{
		Metadata: &Metadata{
			OptionalVariables: []string{"description", "owner"},
		},
		Content: "Name: {{.name}}\nDescription: {{.description}}\nOwner: {{.owner}}",
	}

	ctx := newTestContext(map[string]interface{}{
		"name": "Test Festival",
	})

	output, err := RenderWithDefaults(tmpl, ctx)
	require.NoError(t, err)

	// Optional variables should have empty defaults
	assert.Contains(t, output, "Name: Test Festival")
	assert.Contains(t, output, "Description: ")
	assert.Contains(t, output, "Owner: ")
}

func TestPreserveGuidance(t *testing.T) {
	content := `# Title

[GUIDANCE: This should be preserved]

Some content

[FILL: This should also be preserved]`

	result := PreserveGuidance(content)
	assert.Equal(t, content, result)
}
