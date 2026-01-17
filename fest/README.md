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

## Quick Start

```bash
# 1. Learn the methodology (do this first!)
fest understand methodology
fest understand structure

# 2. Initialize a festivals directory
fest init

# 3. Create your first festival
fest create festival

# 4. Navigate and work
fgo                    # Navigate to festivals
fest status            # Check progress
fest next              # Find next task
```

## Core Commands

| Command | Purpose |
|---------|---------|
| `fest understand` | Learn methodology (run first!) |
| `fest create` | Create festivals/phases/sequences (TUI) |
| `fest validate` | Check structure for issues |
| `fest go` / `fgo` | Navigate to festivals |
| `fest status` | View progress |
| `fest next` | Find next task to work on |

Run `fest --help` for all commands grouped by workflow.

## Shell Integration

Add to your shell config for `fgo` navigation:

```bash
# Zsh/Bash
eval "$(fest shell-init zsh)"

# Fish
fest shell-init fish | source
```

Enable tab completion:

```bash
# Bash
source <(fest completion bash)

# Zsh
source <(fest completion zsh)

# Fish
fest completion fish | source
```

## Configuration

Config stored at `~/.config/fest/config.json`. Run `fest config show` to view.

### Environment Variables

- `FEST_CONFIG_DIR` - Override config directory
- `NO_COLOR` - Disable colored output

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
