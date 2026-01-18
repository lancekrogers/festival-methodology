# fest - Festival Methodology CLI

A CLI for managing Festival Methodology - goal-oriented project management for AI agent workflows.

## Installation

```bash
# Install from source
git clone https://github.com/festival-methodology/fest
cd fest
just install
```

Or with Go:

```bash
go install github.com/festival-methodology/fest/cmd/fest@latest
```

## Shell Integration (Recommended)

Add to your shell config for quick navigation commands:

```bash
# Zsh/Bash
eval "$(fest shell-init zsh)"

# Fish
fest shell-init fish | source
```

This gives you:
- `fgo` - Quick navigation (`fest go`)
- `fls` - Quick listing (`fest list`)
- Tab completion for all fest commands

## Agent Workflow

The typical workflow for AI agents:

### 1. Learn the Methodology

```bash
fest intro                    # Start here - getting started guide
fest understand methodology   # Core principles
fest understand structure     # 3-level hierarchy
```

### 2. Initialize & Create

```bash
fest init                     # Initialize festivals directory
fest create festival          # Create a new festival (TUI)
fest create phase             # Add phases
fest create sequence          # Add sequences
```

### 3. Plan & Validate

```bash
fest validate                 # Check structure for issues
fest validate --fix           # Auto-fix common problems
fest status                   # View festival progress
```

### 4. Execute

```bash
fest execute                  # Execute festival tasks
fest next                     # Find next task to work on
fest progress                 # Track execution progress
```

## Quick Commands

After shell integration:

| Command | Full Form | Purpose |
|---------|-----------|---------|
| `fgo` | `fest go` | Navigate to festivals directory |
| `fgo 2` | `fest go 2` | Go to phase 002 |
| `fgo 2/1` | `fest go 2/1` | Go to phase 2, sequence 1 |
| `fgo active` | `fest go active` | Go to active festivals |
| `fls` | `fest list` | List festivals by status |
| `fls active` | `fest list active` | List active festivals |

## Core Commands

| Command | Purpose |
|---------|---------|
| `fest intro` | Getting started guide (run first!) |
| `fest understand` | Learn methodology concepts |
| `fest create` | Create festivals/phases/sequences (TUI) |
| `fest validate` | Check structure for issues |
| `fest execute` | Execute festival tasks |
| `fest status` | View progress |
| `fest next` | Find next task |

Run `fest --help` for all commands grouped by workflow.

## Configuration

Config stored at `~/.config/fest/config.json`. Run `fest config show` to view.

## Development

Uses `just` for all build/test commands:

```bash
just              # List all commands
just build        # Build fest binary
just test::all    # Run all tests
just install      # Install to $GOBIN
just lint         # Format and vet
just clean        # Clean build artifacts
```

Subcommand modules:

```bash
just test::       # Testing commands
just xbuild::     # Cross-platform builds
just release::    # Release packaging
```

## Learn More

The CLI is self-documenting:

```bash
fest --help              # All commands with workflows
fest understand          # Methodology learning hub
fest [command] --help    # Detailed command help
```

## License

MIT
