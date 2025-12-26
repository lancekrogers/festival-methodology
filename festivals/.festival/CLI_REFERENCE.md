# fest CLI Reference

Complete reference for the fest command-line interface. This guide is optimized for AI agents using Festival Methodology.

## Quick Start for AI Agents

```bash
# Create complete festival structure
fest create festival --name "my-project" --goal "Description" --json
fest create phase --name "PLAN" --json
fest create phase --name "IMPLEMENT" --json
fest create sequence --name "requirements" --json
fest create task --name "analyze" --json

# Add quality gates
fest task defaults sync --approve --json
```

**Always use `--json`** for machine-readable output.

## Installation

```bash
# From festival-methodology directory
cd fest
just build
just install  # Installs to ~/go/bin/fest

# Or run directly
./bin/fest --help
```

## Global Flags

All commands support these flags:

| Flag | Description |
|------|-------------|
| `--json` | Output JSON for automation |
| `--verbose` | Enable verbose logging |
| `--debug` | Enable debug logging |
| `--no-color` | Disable colored output |
| `--config` | Custom config file path |

## Commands

### fest create festival

Create a new festival scaffold.

```bash
fest create festival --name "auth-system" --goal "Build authentication" --json
```

**Flags:**

- `--name` (required): Festival name
- `--goal`: Festival goal/description
- `--tags`: Comma-separated tags
- `--dest`: Destination (active or planned, default: active)
- `--vars-file`: JSON file with template variables

**Output:**

```json
{
  "ok": true,
  "action": "create_festival",
  "festival": {
    "name": "auth-system",
    "slug": "auth-system",
    "dest": "active"
  },
  "created": [
    "FESTIVAL_OVERVIEW.md",
    "FESTIVAL_GOAL.md",
    "FESTIVAL_RULES.md",
    "TODO.md"
  ]
}
```

### fest create phase

Add a phase to the current festival.

```bash
fest create phase --name "IMPLEMENT" --json
```

**Flags:**

- `--name` (required): Phase name (e.g., PLAN, IMPLEMENT, REVIEW)
- `--type`: Phase type (planning or implementation)
- `--after`: Insert after phase number (default: append)
- `--path`: Festival path (default: current directory)

**Output:**

```json
{
  "ok": true,
  "action": "create_phase",
  "phase": {
    "id": "002_IMPLEMENT",
    "name": "IMPLEMENT",
    "number": 2,
    "type": "implementation"
  },
  "created": ["002_IMPLEMENT/PHASE_GOAL.md"]
}
```

### fest create sequence

Add a sequence to the current phase.

```bash
fest create sequence --name "api-development" --json
```

**Flags:**

- `--name` (required): Sequence name
- `--after`: Insert after sequence number (default: append)
- `--path`: Phase path (default: current directory)

**Output:**

```json
{
  "ok": true,
  "action": "create_sequence",
  "sequence": {
    "id": "01_api_development",
    "name": "api-development",
    "number": 1
  },
  "created": ["01_api_development/SEQUENCE_GOAL.md"]
}
```

### fest create task

Add a task to the current sequence.

```bash
fest create task --name "create-endpoints" --json
```

**Flags:**

- `--name` (required): Task name
- `--after`: Insert after task number (default: append)
- `--path`: Sequence path (default: current directory)

**Output:**

```json
{
  "ok": true,
  "action": "create_task",
  "task": {
    "id": "01_create_endpoints",
    "name": "create-endpoints",
    "number": 1
  },
  "created": ["01_create_endpoints.md"]
}
```

### fest task defaults sync

Sync quality gate tasks to all sequences in a festival.

```bash
# Preview changes (default - dry run)
fest task defaults sync --dry-run --json

# Apply changes
fest task defaults sync --approve --json

# Interactive mode
fest task defaults sync --interactive

# Force overwrite modified files
fest task defaults sync --approve --force --json
```

**Flags:**

- `--path`: Festival root (default: current directory)
- `--dry-run`: Preview changes without applying (DEFAULT)
- `--approve`: Actually apply changes
- `--interactive`: Prompt for each change
- `--force`: Overwrite modified files
- `--verbose`: Show detailed logging

**Output:**

```json
{
  "ok": true,
  "action": "task_defaults_sync",
  "dry_run": false,
  "changes": [
    {
      "type": "create",
      "path": "002_IMPLEMENT/01_api/04_testing_and_verify.md",
      "template": "QUALITY_GATE_TESTING"
    },
    {
      "type": "skip",
      "path": "002_IMPLEMENT/02_ui/04_testing_and_verify.md",
      "reason": "file_modified"
    }
  ],
  "summary": {
    "sequences_updated": 3,
    "files_created": 9,
    "files_skipped": 2
  }
}
```

### fest task defaults add

Add quality gate tasks to a specific sequence.

```bash
fest task defaults add --sequence ./002_IMPLEMENT/01_api --approve --json
```

**Flags:**

- `--sequence` (required): Path to target sequence
- `--dry-run`: Preview only (DEFAULT)
- `--approve`: Apply changes

### fest task defaults show

Display current fest.yaml configuration.

```bash
fest task defaults show --json
```

**Output:**

```json
{
  "ok": true,
  "action": "task_defaults_show",
  "config": {
    "version": "1.0",
    "quality_gates": {
      "enabled": true,
      "tasks": [
        {"id": "testing_and_verify", "enabled": true},
        {"id": "code_review", "enabled": true},
        {"id": "review_results_iterate", "enabled": true}
      ]
    }
  }
}
```

### fest task defaults init

Create a starter fest.yaml file.

```bash
fest task defaults init --json
```

### fest renumber

Fix numbering gaps after removing elements.

```bash
# Renumber phases
fest renumber phase ./my-festival --skip-dry-run

# Renumber sequences
fest renumber sequence ./my-festival/002_IMPLEMENT --skip-dry-run

# Renumber tasks
fest renumber task ./my-festival/002_IMPLEMENT/01_api --skip-dry-run
```

**Flags:**

- `--dry-run`: Preview changes (default: true)
- `--skip-dry-run`: Apply changes
- `--backup`: Create backup before changes
- `--start`: Starting number (default: 1)

### fest insert

Insert a new element and renumber subsequent elements.

```bash
# Insert phase after phase 1
fest insert phase --after 1 --name "NEW_PHASE" --path ./my-festival

# Insert sequence after sequence 2
fest insert sequence --after 2 --name "new_seq" --path ./my-festival/002_IMPLEMENT
```

### fest remove

Remove an element and renumber remaining elements.

```bash
fest remove phase --number 2 --path ./my-festival --skip-dry-run
fest remove sequence --number 1 --path ./my-festival/002_IMPLEMENT --skip-dry-run
```

**Flags:**

- `--dry-run`: Preview only (default: true)
- `--force`: Skip confirmation
- `--backup`: Create backup

### fest apply

Apply a template by ID or path.

```bash
# By template ID
fest apply --template-id TASK_TEMPLATE --to ./01_new_task.md

# By path
fest apply --template-path ./custom-template.md --to ./output.md

# With variables
fest apply --template-id FESTIVAL_GOAL --to ./FESTIVAL_GOAL.md --vars-file vars.json
```

### fest count

Count tokens for LLM cost estimation.

```bash
# Single file
fest count ./FESTIVAL_OVERVIEW.md

# Directory (recursive)
fest count --recursive ./my-festival
fest count -d ./my-festival

# With cost estimation
fest count --cost --all ./my-festival

# JSON output
fest count --json ./my-festival
```

**Flags:**

- `--recursive`, `-d`: Scan directories
- `--cost`: Show cost estimates
- `--all`: Show all tokenizer results
- `--model`: Specific model (gpt-4, gpt-3.5, claude-3)

## Configuration

### fest.yaml

Place `fest.yaml` in festival root to customize defaults.

```yaml
version: "1.0"

# Quality gate tasks for implementation sequences
quality_gates:
  enabled: true
  auto_append: true  # Auto-add when creating sequences

  tasks:
    - id: testing_and_verify
      template: QUALITY_GATE_TESTING
      enabled: true
      customizations:
        test_command: "make test"
        coverage_threshold: 80

    - id: code_review
      template: QUALITY_GATE_REVIEW
      enabled: true
      customizations:
        lint_command: "make lint"

    - id: review_results_iterate
      template: QUALITY_GATE_ITERATE
      enabled: true

# Exclude these sequences from quality gates
excluded_patterns:
  - "*_planning"
  - "*_research"
  - "*_requirements"

# Template preferences
templates:
  task_default: TASK_TEMPLATE_SIMPLE
```

### Global Config

`~/.config/fest/config.json`:

```json
{
  "version": "1.0.0",
  "repository": {
    "url": "https://github.com/lancekrogers/festival-methodology",
    "branch": "main"
  },
  "behavior": {
    "auto_backup": false,
    "interactive": true
  }
}
```

## Agent Workflow Patterns

### Pattern 1: Create Full Festival

```bash
# 1. Create festival
fest create festival --name "api-service" --goal "Build REST API" --json

# 2. Add phases
cd active/api-service
fest create phase --name "PLAN" --json
fest create phase --name "IMPLEMENT" --json
fest create phase --name "REVIEW" --json

# 3. Add sequences
cd 002_IMPLEMENT
fest create sequence --name "endpoints" --json
fest create sequence --name "authentication" --json

# 4. Add tasks
cd 01_endpoints
fest create task --name "define_routes" --json
fest create task --name "implement_handlers" --json
fest create task --name "add_validation" --json

# 5. Add quality gates
cd ../..
fest task defaults sync --approve --json
```

### Pattern 2: Update Existing Festival

```bash
# Add new sequence and sync quality gates
cd my-festival/002_IMPLEMENT
fest create sequence --name "new-feature" --json
cd ..
fest task defaults sync --approve --json
```

### Pattern 3: Token Budget Check

```bash
# Check token usage before loading context
fest count --recursive ./my-festival --cost --json
```

## Error Handling

All errors return JSON with `ok: false`:

```json
{
  "ok": false,
  "error": "festival not found: no festivals/ directory",
  "action": "create_phase"
}
```

Common errors:

- `festival not found`: Not in a festivals/ directory
- `phase not found`: Path doesn't point to a phase
- `template not found`: Invalid template ID
- `file_modified`: File was manually edited (use `--force` to overwrite)

## Best Practices

1. **Always use `--json`** for automation and parsing
2. **Run `--dry-run` first** before applying changes
3. **Use `fest task defaults sync`** after creating sequences
4. **Edit only content sections** of generated files
5. **Run `fest count`** before loading large festivals into context
