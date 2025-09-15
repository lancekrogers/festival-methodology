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