# Fest Plugins and Extensions

This document describes the fest plugin system for external commands and the extension system for methodology packs.

## Overview

Fest supports two types of extensibility:

| System | Purpose | Discovery | Format |
|--------|---------|-----------|--------|
| **Plugins** | External CLI commands | `PATH` + manifest | Executables (`fest-*`) |
| **Extensions** | Methodology packs | `.festival/extensions/` | Directories with `extension.yml` |

---

## Plugins

Plugins are external executables that extend fest with new commands.

### Plugin Discovery

Plugins are discovered from three sources (in priority order):

1. **User manifest**: `~/.config/fest-repos/<active>/plugins/manifest.yml`
2. **User bin directory**: `~/.config/fest-repos/<active>/plugins/bin/`
3. **System PATH**: Any executable matching `fest-*`

### Naming Convention

Plugin executables must be named with the `fest-` prefix:

```
fest-<group>-<name>     # Two-part command
fest-<name>             # Single command
```

Examples:

```
fest-export-jira        # Command: "export jira"
fest-import-confluence  # Command: "import confluence"
fest-stats              # Command: "stats"
```

### Plugin Manifest

The manifest file (`manifest.yml`) provides metadata for plugins:

```yaml
version: 1
plugins:
  - command: "export jira"
    exec: "fest-export-jira"
    summary: "Export festival to Jira"
    description: |
      Export festival phases and tasks to Jira issues.
      Creates epics for phases and stories for sequences.
    when_to_use:
      - "Syncing work to issue tracker"
      - "Generating Jira backlog from festival"
    examples:
      - "fest export jira --project PROJ"
      - "fest export jira --project PROJ --dry-run"
    version: "1.2.0"

  - command: "import confluence"
    exec: "fest-import-confluence"
    summary: "Import requirements from Confluence"
    description: "Parse Confluence pages and generate task files"
    version: "1.0.0"
```

### Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `command` | Yes | The fest subcommand (e.g., "export jira") |
| `exec` | Yes | Executable filename |
| `summary` | Yes | One-line description |
| `description` | No | Full description |
| `when_to_use` | No | Usage hints for AI agents |
| `examples` | No | Example command invocations |
| `version` | No | Plugin version |

### Creating a Plugin

#### Step 1: Create the Executable

```bash
#!/bin/bash
# fest-export-jira

echo "Exporting to Jira..."
# Implementation here
```

Or in Go:

```go
package main

import (
    "fmt"
    "os"
)

func main() {
    // args[0] = executable name
    // args[1:] = user-provided arguments
    args := os.Args[1:]

    fmt.Printf("Exporting to Jira with args: %v\n", args)
    // Implementation
}
```

#### Step 2: Make Executable

```bash
chmod +x fest-export-jira
```

#### Step 3: Install

Option A: Add to PATH

```bash
mv fest-export-jira /usr/local/bin/
```

Option B: Add to config repo

```bash
mv fest-export-jira ~/.config/fest-repos/default/plugins/bin/
```

#### Step 4: Add to Manifest (Optional)

```yaml
# ~/.config/fest-repos/default/plugins/manifest.yml
version: 1
plugins:
  - command: "export jira"
    exec: "fest-export-jira"
    summary: "Export festival to Jira"
```

### Plugin Dispatch

When fest receives an unknown command, it checks for plugins:

```go
// Internal dispatch logic
discovery := plugins.NewPluginDiscovery()
discovery.DiscoverAll()

plugin := discovery.FindByArgs(os.Args[1:])
if plugin != nil {
    // Dispatch to external executable
    return dispatch.Run(plugin, remainingArgs)
}
```

### View Discovered Plugins

```bash
fest understand plugins  # List all discovered plugins
```

---

## Extensions

Extensions are methodology packs that provide templates, agents, or workflow configurations.

### Extension Discovery

Extensions are loaded from three locations (in priority order):

| Priority | Location | Source Name |
|----------|----------|-------------|
| 1 (highest) | `.festival/extensions/` (project) | `project` |
| 2 | User config repo `.festival/extensions/` | `user` |
| 3 (lowest) | `~/.config/fest/festivals/.festival/extensions/` | `built-in` |

Higher priority sources override extensions with the same name.

### Extension Structure

```
my-extension/
├── extension.yml       # Manifest (optional)
├── templates/          # Custom templates
│   ├── TASK_TEMPLATE.md
│   └── PHASE_GOAL_TEMPLATE.md
├── agents/             # AI agent prompts
│   └── reviewer.md
└── README.md           # Documentation
```

### Extension Manifest

The `extension.yml` file provides metadata:

```yaml
name: api-workflow
version: 1.0.0
description: Templates and agents for API development
author: Team Backend
type: workflow
tags:
  - api
  - backend
  - rest

files:
  - path: templates/API_TASK_TEMPLATE.md
    description: Task template for API endpoints
    type: template
  - path: agents/api-reviewer.md
    description: API code review agent prompt
    type: agent
```

### Manifest Fields

| Field | Required | Description |
|-------|----------|-------------|
| `name` | Yes | Extension identifier |
| `version` | No | Semantic version |
| `description` | No | Extension description |
| `author` | No | Author or team name |
| `type` | No | Extension type (workflow, template, agent) |
| `tags` | No | Searchable tags |
| `files` | No | File listing with metadata |

### Extension Types

| Type | Purpose |
|------|---------|
| `workflow` | Complete methodology pack |
| `template` | Custom templates only |
| `agent` | AI agent prompts |
| `config` | Configuration presets |

### Creating an Extension

#### Step 1: Create Directory

```bash
mkdir -p .festival/extensions/my-extension
cd .festival/extensions/my-extension
```

#### Step 2: Create Manifest

```yaml
# extension.yml
name: my-extension
version: 1.0.0
description: Custom templates for my team
author: My Team
type: template
```

#### Step 3: Add Content

```markdown
# templates/CUSTOM_TASK.md
---
template_id: CUSTOM_TASK
description: Custom task template
---

# Task: {{.TaskName}}

## Team-Specific Section

[FILL: Custom content here]

## Standard Sections

...
```

#### Step 4: Verify Loading

```bash
fest understand extensions  # List loaded extensions
```

### View Loaded Extensions

```bash
# List all extensions
fest understand extensions

# Output includes:
# - Extension name
# - Version
# - Description
# - Source (project/user/built-in)
# - File count
```

---

## Extension Loader API

### Basic Usage

```go
import "github.com/lancekrogers/festival-methodology/fest/internal/extensions"

loader := extensions.NewExtensionLoader()
loader.LoadAll("/path/to/festival")

// Get extension by name
ext := loader.Get("my-extension")

// List all extensions
all := loader.List()

// Filter by source
projectExts := loader.ListBySource("project")

// Filter by type
workflowExts := loader.ListByType("workflow")
```

### Extension Methods

```go
// Check if file exists in extension
if ext.HasFile("templates/TASK.md") {
    path := ext.GetFile("templates/TASK.md")
    // Use file
}

// List all files
files, _ := ext.ListFiles()
```

---

## Plugin Discovery API

### Basic Usage

```go
import "github.com/lancekrogers/festival-methodology/fest/internal/plugins"

discovery := plugins.NewPluginDiscovery()
discovery.DiscoverAll()

// Get all plugins
all := discovery.Plugins()

// Find by command
plugin := discovery.FindByCommand("export jira")

// Find by CLI args
plugin := discovery.FindByArgs([]string{"export", "jira", "--project", "PROJ"})
```

### Dispatch to Plugin

```go
import "github.com/lancekrogers/festival-methodology/fest/internal/plugins"

// Run plugin with arguments
err := plugins.Dispatch(plugin, []string{"--project", "PROJ"})
```

---

## Best Practices

### Plugins

1. **Follow naming conventions**: Use `fest-<group>-<name>` format
2. **Handle errors gracefully**: Exit with non-zero codes on failure
3. **Support --help**: Provide usage information
4. **Use manifest**: Include metadata for better discovery
5. **Version your plugins**: Track compatibility

### Extensions

1. **Include manifest**: Always provide `extension.yml`
2. **Use semantic versioning**: Track changes properly
3. **Document files**: Use the `files` array in manifest
4. **Test templates**: Verify variable rendering
5. **Keep focused**: One extension per concern

---

## Commands Reference

```bash
# Plugins
fest understand plugins              # List discovered plugins

# Extensions
fest understand extensions           # List loaded extensions
fest extension list                  # Alternative: list extensions
fest extension enable <name>         # Enable an extension
fest extension disable <name>        # Disable an extension
```
