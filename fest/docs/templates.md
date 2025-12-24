# Fest Template System

This document describes the fest template system, including variable reference, template creation, resolution order, and YAML frontmatter format.

## Template Variables

Variables are accessed using Go template syntax `{{ .VariableName }}`.

### Festival-Level Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.FestivalName}}` | string | Festival name | `api-refactor` |
| `{{.FestivalGoal}}` | string | Festival goal | `Modernize API layer` |
| `{{.FestivalTags}}` | []string | Festival tags | `["api", "backend"]` |
| `{{.FestivalDescription}}` | string | Extended description | |

### Phase-Level Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.PhaseNumber}}` | int | Phase number | `1` |
| `{{.PhaseName}}` | string | Phase name | `PLANNING` |
| `{{.PhaseID}}` | string | Formatted ID | `001_PLANNING` |
| `{{.PhaseType}}` | string | Phase type | `planning`, `implementation` |
| `{{.PhaseStructure}}` | string | Structure type | `freeform`, `structured` |
| `{{.PhaseObjective}}` | string | Phase objective | |

### Sequence-Level Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.SequenceNumber}}` | int | Sequence number | `1` |
| `{{.SequenceName}}` | string | Sequence name | `requirements` |
| `{{.SequenceID}}` | string | Formatted ID | `01_requirements` |
| `{{.SequenceObjective}}` | string | Sequence objective | |
| `{{.SequenceDependencies}}` | []string | Dependencies | |

### Task-Level Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.TaskNumber}}` | int | Task number | `1` |
| `{{.TaskName}}` | string | Task name | `user_research` |
| `{{.TaskID}}` | string | Formatted ID | `01_user_research.md` |
| `{{.TaskObjective}}` | string | Task objective | |
| `{{.TaskDeliverables}}` | []string | Task deliverables | |
| `{{.TaskParallel}}` | bool | Can run in parallel | `true` |
| `{{.TaskDependencies}}` | []string | Task dependencies | |

### Computed Variables

| Variable | Type | Description | Example |
|----------|------|-------------|---------|
| `{{.CurrentLevel}}` | string | Current hierarchy level | `festival`, `phase`, `sequence`, `task` |
| `{{.ParentPhaseID}}` | string | Parent phase ID | `001_PLANNING` |
| `{{.ParentSequenceID}}` | string | Parent sequence ID | `01_requirements` |
| `{{.FullPath}}` | string | Full relative path | `001_PLANNING/01_requirements/01_task.md` |

### Custom Variables

Custom variables can be passed via `--vars-file` and accessed through the `{{.Custom}}` map:

```bash
fest create festival --name test --vars-file vars.json
```

vars.json:

```json
{
  "author": "Team A",
  "priority": "high"
}
```

Template access:

```text
Author: {{.Custom.author}}
Priority: {{.Custom.priority}}
```

---

## YAML Frontmatter Format

Templates can include YAML frontmatter to define metadata, IDs, and required variables.

### Basic Format

```markdown
---
template_id: MY_TEMPLATE
description: Template for API endpoints
required_variables:
  - FestivalName
  - PhaseName
optional_variables:
  - PhaseObjective
aliases:
  - api_template
  - endpoint
---

# {{.FestivalName}} - {{.PhaseName}}

Content here...
```

### Frontmatter Fields

| Field | Type | Description |
|-------|------|-------------|
| `template_id` | string | Primary template identifier |
| `id` | string | Alternative to `template_id` |
| `aliases` | []string | Alternative names for lookup |
| `template_version` | string | Template version |
| `description` | string | Template description |
| `required_variables` | []string | Variables that must be provided |
| `optional_variables` | []string | Variables that may be omitted |

### ID Resolution

Templates can be referenced by:

1. `template_id` field
2. `id` field (fallback)
3. Any value in `aliases` array

```bash
# All of these work if configured in frontmatter
fest apply --template-id MY_TEMPLATE --to output.md
fest apply --template-id api_template --to output.md
fest apply --template-id endpoint --to output.md
```

---

## Template Resolution Order

When rendering templates, fest searches in this order:

### 1. Catalog Lookup (Preferred)

Templates are indexed by ID/alias from the `.festival/templates/` directory:

```text
.festival/
└── templates/
    ├── TASK_TEMPLATE.md          # template_id: TASK
    ├── PHASE_GOAL_TEMPLATE.md    # template_id: PHASE_GOAL
    └── SEQUENCE_GOAL_TEMPLATE.md # template_id: SEQUENCE_GOAL
```

### 2. Direct File Path (Fallback)

If catalog lookup fails, fest tries the direct file path:

```go
// Internal resolution logic
content, err := m.RenderWithFallback(catalog, "PHASE_GOAL",
    filepath.Join(tmplRoot, "PHASE_GOAL_TEMPLATE.md"), ctx)
```

### 3. Copy Mode

If the file exists but contains no `{{` markers, it's copied as-is without rendering.

---

## Creating Custom Templates

### Step 1: Create Template File

Create a `.md` file in `.festival/templates/`:

```markdown
---
template_id: API_TASK
description: Template for API development tasks
required_variables:
  - TaskName
  - TaskNumber
  - SequenceName
aliases:
  - api
  - rest_task
---

# Task: {{.TaskName}}

## Objective

[FILL: Describe the API task objective]

## Context

- **Sequence**: {{.SequenceID}}
- **Task**: {{.TaskID}}

## Implementation Steps

1. [FILL: First step]
2. [FILL: Second step]
3. [FILL: Third step]

## Definition of Done

- [ ] [FILL: Completion criteria]
```

### Step 2: Use the Template

```bash
# By ID
fest apply --template-id API_TASK --to ./01_api/01_design.md

# By alias
fest apply --template-id api --to ./01_api/01_design.md

# With variables file
fest apply --template-id API_TASK --to ./output.md --vars-file vars.json
```

---

## Built-in Templates

Fest includes these built-in templates:

### Festival Templates

| ID | File | Description |
|----|------|-------------|
| `FESTIVAL_OVERVIEW` | `FESTIVAL_OVERVIEW_TEMPLATE.md` | Festival overview document |
| `FESTIVAL_RULES` | `FESTIVAL_RULES_TEMPLATE.md` | Festival rules and guidelines |
| `TODO` | `TODO_TEMPLATE.md` | Progress tracking TODO |

### Phase Templates

| ID | File | Description |
|----|------|-------------|
| `PHASE_GOAL` | `PHASE_GOAL_TEMPLATE.md` | Phase goal document |

### Sequence Templates

| ID | File | Description |
|----|------|-------------|
| `SEQUENCE_GOAL` | `SEQUENCE_GOAL_TEMPLATE.md` | Sequence goal document |

### Task Templates

| ID | File | Description |
|----|------|-------------|
| `TASK` | `TASK_TEMPLATE.md` | Generic task template |

---

## Template Markers

Templates use these marker patterns for content that needs filling:

### FILL Markers

```text
[FILL: description]
```

Indicates content that must be replaced by the user.

### Markers: Template Variables

```text
{{.VariableName}}
```

Go template syntax variables that are automatically rendered.

### Validation

The `fest validate` command checks for unfilled markers:

```bash
fest validate templates  # Check for [FILL:] markers
fest validate           # Includes template validation
```

---

## Common Patterns

### Conditional Content

```text
{{if .PhaseObjective}}
## Objective

{{.PhaseObjective}}
{{end}}
```

### Iterating Lists

```text
## Dependencies

{{range .SequenceDependencies}}
- {{.}}
{{end}}
```

### Default Values

```text
Priority: {{if .Custom.priority}}{{.Custom.priority}}{{else}}medium{{end}}
```

### String Manipulation

```text
# {{.PhaseName | upper}}
```

Note: Go template functions must be registered with the renderer.

---

## Template Development Workflow

### 1. Create and Test Locally

```bash
# Create template
vim .festival/templates/MY_TEMPLATE.md

# Test with fest apply
fest apply --template-path ./.festival/templates/MY_TEMPLATE.md \
    --to ./test_output.md \
    --vars-file test_vars.json

# Check output
cat test_output.md
```

### 2. Add to Catalog

Add `template_id` frontmatter for catalog indexing:

```yaml
---
template_id: MY_TEMPLATE
aliases:
  - my
  - custom
---
```

### 3. Validate

```bash
# Check for unfilled markers in generated files
fest validate templates
```

### 4. Share

Templates in `.festival/templates/` are shared across the festival. For project-wide templates, use configuration repositories:

```bash
fest config add https://github.com/org/fest-templates
```

---

## API Reference

### Manager Methods

```go
// RenderFile renders a template file with context
func (m *Manager) RenderFile(templatePath string, ctx *Context) (string, error)

// RenderByID renders by template ID from catalog
func (m *Manager) RenderByID(catalog *Catalog, id string, ctx *Context) (string, error)

// RenderWithFallback tries catalog, then falls back to file path
func (m *Manager) RenderWithFallback(catalog *Catalog, id, fallbackPath string, ctx *Context) (string, error)

// RenderFileOrCopy renders or copies based on content
func (m *Manager) RenderFileOrCopy(templatePath string, ctx *Context) (string, error)
```

### Context Methods

```go
// Create context and set variables
ctx := template.NewContext()
ctx.SetFestival("my-fest", "Goal here", []string{"tag1"})
ctx.SetPhase(1, "PLANNING", "planning")
ctx.SetSequence(1, "requirements")
ctx.SetTask(1, "gather_specs")

// Add custom variables
ctx.Custom["author"] = "Team A"
```

### Catalog Methods

```go
// Load catalog from template directory
catalog, err := template.LoadCatalog(".festival/templates")

// Resolve template ID to path
path, ok := catalog.Resolve("TASK")
```
