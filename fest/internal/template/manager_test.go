package template

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManager_RenderFile(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.md")

	content := `---
template_id: TEST_TEMPLATE
required_variables:
  - name
  - goal
---
# Festival: {{.name}}

Goal: {{.goal}}

[GUIDANCE: Add more details here]`

	err := os.WriteFile(templatePath, []byte(content), 0644)
	require.NoError(t, err)

	// Create context
	ctx := NewContextBuilder().
		WithUser("name", "my-festival").
		WithUser("goal", "Build awesome things").
		Build()

	// Render template
	manager := NewManager()
	output, err := manager.RenderFile(templatePath, ctx)

	require.NoError(t, err)
	assert.Contains(t, output, "# Festival: my-festival")
	assert.Contains(t, output, "Goal: Build awesome things")
	assert.Contains(t, output, "[GUIDANCE: Add more details here]")
}

func TestManager_RenderFile_MissingVariable(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.md")

	content := `---
template_id: TEST_TEMPLATE
required_variables:
  - name
  - goal
---
# Festival: {{.name}}

Goal: {{.goal}}`

	err := os.WriteFile(templatePath, []byte(content), 0644)
	require.NoError(t, err)

	// Create context with missing variable
	ctx := NewContextBuilder().
		WithUser("name", "my-festival").
		Build()

	// Render template - should fail validation
	manager := NewManager()
	_, err = manager.RenderFile(templatePath, ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "validation failed")
}

func TestManager_RenderFileToFile(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "template.md")
	outputPath := filepath.Join(tmpDir, "output", "rendered.md")

	content := `# Festival: {{.name}}

Goal: {{.goal}}`

	err := os.WriteFile(templatePath, []byte(content), 0644)
	require.NoError(t, err)

	// Create context
	ctx := NewContextBuilder().
		WithUser("name", "my-festival").
		WithUser("goal", "Build awesome things").
		Build()

	// Render to file
	manager := NewManager()
	err = manager.RenderFileToFile(templatePath, outputPath, ctx)
	require.NoError(t, err)

	// Verify output file exists and has correct content
	output, err := os.ReadFile(outputPath)
	require.NoError(t, err)
	assert.Contains(t, string(output), "# Festival: my-festival")
	assert.Contains(t, string(output), "Goal: Build awesome things")
}

func TestManager_RenderDirectory(t *testing.T) {
	// Create temporary directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, "templates")
	outputDir := filepath.Join(tmpDir, "output")

	err := os.MkdirAll(templateDir, 0755)
	require.NoError(t, err)

	// Create multiple template files
	templates := map[string]string{
		"overview.md": `# Festival: {{.name}}`,
		"goals.md":    `Goal: {{.goal}}`,
		"subdir/details.md": `Details for {{.name}}`,
	}

	for path, content := range templates {
		fullPath := filepath.Join(templateDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// Create context
	ctx := NewContextBuilder().
		WithUser("name", "my-festival").
		WithUser("goal", "Build awesome things").
		Build()

	// Render directory
	manager := NewManager()
	err = manager.RenderDirectory(templateDir, outputDir, ctx)
	require.NoError(t, err)

	// Verify all files were created
	assert.FileExists(t, filepath.Join(outputDir, "overview.md"))
	assert.FileExists(t, filepath.Join(outputDir, "goals.md"))
	assert.FileExists(t, filepath.Join(outputDir, "subdir", "details.md"))

	// Verify content
	content, err := os.ReadFile(filepath.Join(outputDir, "overview.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "# Festival: my-festival")

	content, err = os.ReadFile(filepath.Join(outputDir, "goals.md"))
	require.NoError(t, err)
	assert.Contains(t, string(content), "Goal: Build awesome things")
}

func TestManager_RenderString(t *testing.T) {
	manager := NewManager()

	ctx := NewContextBuilder().
		WithUser("name", "Test").
		Build()

	output, err := manager.RenderString("Hello {{.name}}!", ctx)
	require.NoError(t, err)
	assert.Equal(t, "Hello Test!", output)
}

func TestManager_GetTemplateInfo(t *testing.T) {
	// Create temporary template file
	tmpDir := t.TempDir()
	templatePath := filepath.Join(tmpDir, "test.md")

	content := `---
template_id: TEST_TEMPLATE
template_version: 1.0.0
description: Test template for unit tests
required_variables:
  - name
optional_variables:
  - description
---
Content here`

	err := os.WriteFile(templatePath, []byte(content), 0644)
	require.NoError(t, err)

	// Get template info
	manager := NewManager()
	metadata, err := manager.GetTemplateInfo(templatePath)

	require.NoError(t, err)
	assert.Equal(t, "TEST_TEMPLATE", metadata.TemplateID)
	assert.Equal(t, "1.0.0", metadata.TemplateVersion)
	assert.Equal(t, "Test template for unit tests", metadata.Description)
	assert.Equal(t, []string{"name"}, metadata.RequiredVariables)
	assert.Equal(t, []string{"description"}, metadata.OptionalVariables)
}

func TestManager_ListTemplates(t *testing.T) {
	// Create temporary directory with templates
	tmpDir := t.TempDir()

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
	}

	for path, content := range templates {
		fullPath := filepath.Join(tmpDir, path)
		err := os.MkdirAll(filepath.Dir(fullPath), 0755)
		require.NoError(t, err)
		err = os.WriteFile(fullPath, []byte(content), 0644)
		require.NoError(t, err)
	}

	// List templates
	manager := NewManager()
	tmplList, err := manager.ListTemplates(tmpDir)

	require.NoError(t, err)
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
