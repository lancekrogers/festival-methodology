package template

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_Load(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantMetadata *Metadata
		wantContent  string
		wantErr      bool
	}{
		{
			name: "template with frontmatter",
			content: `---
template_id: TEST_TEMPLATE
template_version: 1.0.0
required_variables:
  - name
  - goal
optional_variables:
  - description
---
# Festival: {{.name}}

Goal: {{.goal}}

[GUIDANCE: Add description here]`,
			wantMetadata: &Metadata{
				TemplateID:        "TEST_TEMPLATE",
				TemplateVersion:   "1.0.0",
				RequiredVariables: []string{"name", "goal"},
				OptionalVariables: []string{"description"},
			},
			wantContent: `# Festival: {{.name}}

Goal: {{.goal}}

[GUIDANCE: Add description here]`,
			wantErr: false,
		},
		{
			name: "template without frontmatter",
			content: `# Simple Template

Content: {{.value}}`,
			wantMetadata: nil,
			wantContent: `# Simple Template

Content: {{.value}}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "test.md")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			// Load template
			loader := NewLoader()
			tmpl, err := loader.Load(context.Background(), tmpFile)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tmpFile, tmpl.Path)
			assert.Equal(t, tt.wantMetadata, tmpl.Metadata)
			assert.Equal(t, tt.wantContent, tmpl.Content)
		})
	}
}

func TestLoader_LoadAll(t *testing.T) {
	// Create temporary directory with multiple templates
	tmpDir := t.TempDir()

	// Create template files
	templates := map[string]string{
		"template1.md": `---
template_id: TEMPLATE_1
---
Content 1`,
		"template2.md": `---
template_id: TEMPLATE_2
---
Content 2`,
		"subdir/template3.md": `---
template_id: TEMPLATE_3
---
Content 3`,
		"not_template.txt": "This should be ignored",
	}

	for path, content := range templates {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Load all templates
	loader := NewLoader()
	tmplList, err := loader.LoadAll(context.Background(), tmpDir)
	require.NoError(t, err)

	// Should have loaded 3 markdown files
	assert.Len(t, tmplList, 3)

	// Verify template IDs
	ids := make(map[string]bool)
	for _, tmpl := range tmplList {
		if tmpl.Metadata != nil {
			ids[tmpl.Metadata.TemplateID] = true
		}
	}

	assert.True(t, ids["TEMPLATE_1"])
	assert.True(t, ids["TEMPLATE_2"])
	assert.True(t, ids["TEMPLATE_3"])
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		wantMetadata   bool
		wantContentLen int
		wantErr        bool
	}{
		{
			name: "valid frontmatter",
			content: `---
template_id: TEST
---
Content here`,
			wantMetadata:   true,
			wantContentLen: 12,
			wantErr:        false,
		},
		{
			name: "no frontmatter",
			content: `Just content
No frontmatter`,
			wantMetadata:   false,
			wantContentLen: 27, // "Just content\nNo frontmatter"
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary file
			tmpFile := filepath.Join(t.TempDir(), "test.md")
			err := os.WriteFile(tmpFile, []byte(tt.content), 0644)
			require.NoError(t, err)

			file, err := os.Open(tmpFile)
			require.NoError(t, err)
			defer file.Close()

			metadata, content, err := parseFrontmatter(file)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tt.wantMetadata {
				assert.NotNil(t, metadata)
			} else {
				assert.Nil(t, metadata)
			}

			assert.Equal(t, tt.wantContentLen, len(content))
		})
	}
}
