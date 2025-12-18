# fest - Festival Methodology CLI

A minimal CLI tool for managing Festival Methodology files, providing safe initialization, syncing, and updating of festival directory structures.

## Features

- **Safe Operations**: Uses SHA256 checksums to track file changes and prevent accidental overwrites
- **Smart Updates**: Only updates unchanged files, preserving user modifications
- **Interactive Mode**: Prompts for decisions on modified files during updates
- **Backup Support**: Optional automatic backups before updates
- **Token Counting**: Multiple tokenization methods for LLM cost estimation
- **Festival Management**: Intelligent renumbering for phases, sequences, and tasks
- **Element Manipulation**: Insert, remove, and reorder festival elements with automatic renumbering
- **Configurable**: Flexible configuration via JSON config file
- **Colored Output**: Clear, colored terminal output for better visibility
- **Config Repos**: Manage private configuration repositories for templates, policies, and extensions
- **Plugin System**: Extend fest with custom commands via external executables
- **Extension System**: Load methodology extensions from project, user, or built-in sources
- **Quality Gates**: Configurable quality gate policies with phase-level overrides
- **Festival Index**: Generate machine-readable indices for tool integration

## Installation

```bash
go install github.com/festival-methodology/fest/cmd/fest@latest
```

Or build from source:

```bash
git clone https://github.com/festival-methodology/fest
cd fest
go build -o fest cmd/fest/main.go
```

## Usage

### Interactive TUI

```bash
# Launch interactive menu (inside a festivals/ workspace)
fest tui

# Create a festival, add phases/sequences/tasks with guided prompts.
# The TUI inspects template requirements and asks only for missing variables.
```

Advanced TUI (Charm-based): build with tag `charm` after installing dependencies:

```bash
go get github.com/charmbracelet/huh@latest github.com/charmbracelet/lipgloss@latest github.com/charmbracelet/bubbletea@latest
go build -tags charm -o fest cmd/fest/main.go
```

### Navigation

Navigate to your festivals directory from anywhere:

```bash
# One-time setup - add to ~/.zshrc or ~/.bashrc:
eval "$(fest shell-init zsh)"
```

Then use `fgo` to navigate:

```bash
fgo              # Go to festivals root
fgo 2            # Go to phase 002
fgo 2/1          # Go to phase 2, sequence 1
fgo active       # Go to active directory
```

Without shell integration, use command substitution:

```bash
cd $(fest go)
cd $(fest go 2)
```

Register a workspace for cross-project navigation:

```bash
fest init --register /path/to/project/festivals
```

### Initialize a new festival directory

```bash
# Initialize in current directory
fest init

# Initialize with specific path
fest init /path/to/project

# Skip confirmation prompt
fest init --yes
```

### Sync templates from GitHub

```bash
# Sync using default repository
fest sync

# Sync from specific repository
fest sync --source github.com/user/repo --branch main

# Force overwrite existing cache
fest sync --force
```

### Update festival files

```bash
# Interactive update (default)
fest update

# Preview changes without modifying files
fest update --dry-run

# Update only unchanged files, skip modified
fest update --no-interactive

# Create backup before updating
fest update --backup

# Show diffs for modified files
fest update --diff
```

### Count tokens in files

```bash
# Basic token counting
fest count document.md

# Use specific model tokenizer
fest count --model gpt-4 document.md

# Show all counting methods
fest count --all document.md

# Include cost estimates
fest count --cost document.md

# Output as JSON for scripting
fest count --json document.md

# Custom approximation ratios
fest count --chars-per-token 3.5 --words-per-token 0.8 document.md
```

### Festival Renumbering

```bash
# Renumber phases (dry-run by default)
fest renumber phase ./my-festival

# Renumber sequences in a phase
fest renumber sequence --phase 001_PLAN

# Renumber tasks in a sequence
fest renumber task --sequence 001_PLAN/01_requirements

# Apply changes (disable dry-run)
fest renumber phase --dry-run=false ./my-festival

# Create backup before renumbering
fest renumber phase --backup ./my-festival
```

### Insert new elements

```bash
# Insert a new phase after phase 001
fest insert phase --after 1 --name "DESIGN_REVIEW"

# Insert a sequence in a phase
fest insert sequence --phase 001_PLAN --after 1 --name "validation"

# Insert at the beginning (after 0)
fest insert phase --after 0 --name "PREREQUISITES"
```

### Remove elements

```bash
# Remove a phase and renumber subsequent phases
fest remove phase 2

# Remove a sequence from a phase
fest remove sequence --phase 001_PLAN 02

# Remove with automatic confirmation
fest remove phase --force 002_DEFINE_INTERFACES

# Create backup before removal
fest remove phase --backup 2
```

## Configuration

Configuration is stored in `~/.config/fest/config.json`:

```json
{
  "version": "1.0.0",
  "repository": {
    "url": "https://github.com/user/festival-methodology",
    "branch": "main",
    "path": "festivals"
  },
  "local": {
    "cache_dir": "~/.config/fest/cache",
    "backup_dir": ".fest-backup",
    "checksum_file": ".fest-checksums.json"
  },
  "behavior": {
    "auto_backup": false,
    "interactive": true,
    "use_color": true,
    "verbose": false
  },
  "network": {
    "timeout": 30,
    "retry_count": 3,
    "retry_delay": 1
  }
}
```

## Safety Mechanisms

1. **Checksum Tracking**: SHA256 checksums track all files to detect modifications
2. **Three-Tier Classification**: Files are classified as unchanged, modified, or new
3. **Interactive Prompts**: User confirmation required for modified files
4. **Atomic Operations**: File operations are atomic to prevent corruption
5. **Backup Creation**: Optional backups with timestamped directories
6. **Dry Run Mode**: Preview changes without modifying files

## File Categories

During updates, files are categorized as:

- **Unchanged**: Files matching original checksums (safe to update)
- **Modified**: Files with user changes (require decision)
- **New**: User-created files (always preserved)
- **Deleted**: Files removed by user (not recreated)

## Festival Management

The `fest` tool includes powerful commands for managing Festival Methodology structures:

### Three-Level Hierarchy

- **Phases**: 3-digit directories (001_PLAN, 002_DEFINE_INTERFACES, etc.)
- **Sequences**: 2-digit directories within phases (01_requirements, 02_architecture, etc.)
- **Tasks**: 2-digit markdown files within sequences (01_analyze.md, 02_implement.md, etc.)

### Renumbering Features

- **Automatic Renumbering**: When inserting or removing elements, all subsequent items are automatically renumbered
- **Parallel Task Support**: Multiple tasks with the same number (parallel execution) are preserved
- **Dry-Run by Default**: Preview changes before applying them
- **Backup Creation**: Optional backups before making changes
- **Smart Detection**: Automatically identifies element types based on naming patterns

### Common Use Cases

1. **Insert a new phase**: All subsequent phases shift forward
2. **Remove a sequence**: Following sequences move up to fill the gap
3. **Reorder tasks**: Maintain proper numbering after reorganization
4. **Fix numbering gaps**: Renumber to create continuous sequences

## Token Counting

The `fest count` command provides comprehensive token counting using multiple methods:

### Exact Tokenizers

- **GPT-4/GPT-3.5**: Uses OpenAI's tiktoken library for exact token counts
- **Claude**: Approximation based on character ratios (API required for exact counts)

### Approximation Methods

- **Character-based**: Divides total characters by ratio (default: 4 chars = 1 token)
- **Word-based**: Multiplies word count by ratio (default: 1 word = 1.33 tokens)
- **Whitespace split**: Simple word count based on whitespace

### Cost Estimation

When using `--cost`, the tool estimates API costs for popular models:

- OpenAI GPT-4, GPT-3.5-turbo
- Anthropic Claude-3 Opus, Sonnet, Haiku
- Prices are for input tokens only

### Output Formats

- **Table format** (default): Human-readable table with all metrics
- **JSON format** (`--json`): Machine-readable for scripting and automation

## Config Repositories

Config repos allow you to maintain private templates, policies, and extensions in a separate repository:

```bash
# Add a config repo (git URL or local path)
fest config add my-config https://github.com/user/my-fest-config
fest config add local-config /path/to/local/config

# Sync all config repos (pull latest)
fest config sync

# Sync a specific repo
fest config sync my-config

# Set active config repo
fest config use my-config

# Show current active config
fest config show

# List all config repos
fest config list
```

### Config Repo Structure

```
my-fest-config/
├── festivals/                    # Overrides for .festival/ structure
│   └── .festival/
│       ├── templates/            # Custom templates
│       ├── extensions/           # Custom extensions
│       └── agents/               # Custom agents
└── user/                         # User-specific customizations
    ├── config.yaml               # User settings
    └── policies/                 # Quality gate policies
        └── gates/
            ├── default.yml       # Default gates
            └── variants/         # Named variants
                ├── backend.yml
                └── frontend.yml
```

### Precedence Rules

1. **Project-local**: `.festival/` in current festival
2. **User-level**: Active config repo's `festivals/.festival/`
3. **Built-in**: Default templates from fest installation

## Quality Gates

Quality gates are tasks automatically appended to implementation sequences:

```bash
# Show resolved quality gates
fest task defaults show

# Show gates for a specific phase
fest task defaults show --phase 002_IMPLEMENT

# Sync gates to sequences
fest task defaults sync

# Sync with a specific policy variant
fest task defaults sync --policy backend
```

### Gate Policy Format

Create policies in `policies/gates/default.yml`:

```yaml
version: 1
name: default

append:
  - id: testing_and_verify
    template: QUALITY_GATE_TESTING
    enabled: true
  - id: code_review
    template: QUALITY_GATE_REVIEW
    enabled: true
  - id: review_results_iterate
    template: QUALITY_GATE_ITERATE
    enabled: true

exclude_patterns:
  - "*_planning"
  - "*_research"
  - "*_docs"
```

### Phase Overrides

Add `.fest.gates.yml` to a phase for phase-specific modifications:

```yaml
ops:
  - add:
      step:
        id: security_review
        template: SECURITY_REVIEW
      after: code_review

  - remove:
      id: review_results_iterate
```

## Plugin System

Extend fest with custom commands via external executables:

### Plugin Discovery

Plugins are discovered from:
1. Config repo `user/plugins/bin/` directory
2. System PATH (executables named `fest-*`)

### Plugin Manifest

Document plugins in `user/plugins/manifest.yml`:

```yaml
version: 1
plugins:
  - command: "export jira"
    exec: "fest-export-jira"
    summary: "Export festival to Jira"
    description: |
      Converts festival structure to Jira artifacts.
    when_to_use:
      - "Festival needs Jira tracking"
    examples:
      - "fest export jira --phase 002"
```

### Using Plugins

```bash
# Execute a plugin
fest export jira --phase 002

# Alternative hyphenated form
fest export-jira --phase 002

# View available plugins
fest understand plugins
```

## Extension System

Extensions provide reusable methodology components:

```bash
# List all loaded extensions
fest extension list

# Show extension details
fest extension info my-extension

# Filter by source
fest extension list --source project
fest extension list --source user

# Filter by type
fest extension list --type workflow
```

### Extension Sources

Extensions are loaded from (in precedence order):
1. **Project**: `.festival/extensions/` in current festival
2. **User**: Config repo `festivals/.festival/extensions/`
3. **Built-in**: Default extensions from fest installation

### Extension Manifest

Create `extension.yml` in your extension directory:

```yaml
name: my-workflow
version: "1.0.0"
description: Custom workflow for my team
author: Your Name
type: workflow
tags:
  - automation
  - ci-cd
files:
  - path: README.md
    description: Documentation
  - path: templates/
    description: Workflow templates
```

## Festival Index

Generate machine-readable indices for tool integration:

```bash
# Generate index for current festival
fest index write

# Generate index for specific festival
fest index write /path/to/festival

# Write to custom location
fest index write --output custom/index.json

# Validate index against filesystem
fest index validate

# Show index contents
fest index show

# Show as JSON
fest index show --json
```

### Index Schema

The index file (`.festival/index.json`) provides:

```json
{
  "fest_spec": 1,
  "festival_id": "my-festival",
  "generated_at": "2025-12-17T10:30:00Z",
  "phases": [
    {
      "phase_id": "001_DESIGN",
      "path": "001_DESIGN",
      "goal_file": "PHASE_GOAL.md",
      "sequences": [
        {
          "sequence_id": "01_requirements",
          "path": "001_DESIGN/01_requirements",
          "goal_file": "SEQUENCE_GOAL.md",
          "tasks": [
            {
              "task_id": "01_gather.md",
              "path": "001_DESIGN/01_requirements/01_gather.md",
              "managed": false
            }
          ],
          "managed_gates": []
        }
      ]
    }
  ]
}
```

### Validation

The validator checks:
- All indexed entries exist on disk
- Files on disk that aren't in the index
- Missing goal files (warnings)

## Environment Variables

- `FEST_CONFIG_DIR`: Override default config directory
- `NO_COLOR`: Disable colored output
- `TIKTOKEN_CACHE_DIR`: Cache directory for tiktoken vocabularies

## Command-Line Flags

### Global Flags

- `--config`: Specify custom config file
- `--verbose`: Enable verbose output
- `--no-color`: Disable colored output
- `--debug`: Enable debug logging

### Command-Specific Flags

#### init

- `--source`: Source directory for templates
- `--yes`: Skip confirmation prompt
- `--force`: Overwrite existing files

#### sync

- `--source`: GitHub repository URL
- `--branch`: Git branch to sync from
- `--force`: Overwrite existing cache
- `--timeout`: Download timeout in seconds
- `--retry`: Number of retry attempts
- `--dry-run`: Preview without downloading

#### update

- `--dry-run`: Preview changes without modifying
- `--force`: Update all files regardless of modifications
- `--backup`: Create backup before updating
- `--interactive`: Prompt for each modified file
- `--no-interactive`: Skip all modified files
- `--diff`: Show diffs for modified files

#### count

- `--model`: Specific model for tokenization (gpt-4, gpt-3.5-turbo, claude-3)
- `--all`: Show all counting methods
- `--json`: Output in JSON format
- `--cost`: Include cost estimates
- `--chars-per-token`: Characters per token ratio (default: 4.0)
- `--words-per-token`: Words per token ratio (default: 0.75)

#### config

- `add <name> <source>`: Add a config repo (git URL or local path)
- `sync [name]`: Sync config repos (pull latest)
- `use <name>`: Set active config repo
- `show`: Show current active config
- `list`: List all config repos

#### extension

- `list`: List all loaded extensions
- `info <name>`: Show extension details
- `--source`: Filter by source (project, user, builtin)
- `--type`: Filter by type (workflow, template, agent)

#### index

- `write [path]`: Generate festival index
- `validate [path]`: Validate index against filesystem
- `show [path]`: Show index contents
- `--output`: Output path for write command
- `--index`: Index file path for validate command
- `--json`: Output as JSON for show command

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package tests
go test ./internal/fileops -v
```

### Building

```bash
# Build for current platform
go build -o fest cmd/fest/main.go

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o fest-linux
GOOS=darwin GOARCH=amd64 go build -o fest-darwin
GOOS=windows GOARCH=amd64 go build -o fest.exe
```

## License

MIT

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
