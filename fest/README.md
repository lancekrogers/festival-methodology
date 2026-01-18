# fest - Festival Methodology CLI

fest is a CLI tool for working with **Festival Methodology** - a structured approach to project management designed for AI agent workflows.

## What is Festival Methodology?

Festival Methodology organizes work into a three-level hierarchy:

```
Festival (the project)
├── Phase (major milestone)
│   ├── Sequence (related tasks)
│   │   ├── Task 1
│   │   ├── Task 2
│   │   └── Task 3
│   └── Sequence
└── Phase
```

**Why this structure?**

- **Context Management**: AI agents have limited context windows. Festivals break work into digestible chunks that fit within agent context limits.
- **Goal-Oriented**: Each level (festival, phase, sequence, task) has explicit goals. Agents always know what they're working toward.
- **Resumable**: Work can be paused and resumed. A new agent session can pick up exactly where the last one left off.
- **Traceable**: Every task links to its parent sequence and phase. Progress is trackable across the entire project.

**Key Concepts:**

- **Festival**: A complete project or initiative with a defined outcome
- **Phase**: A major milestone (e.g., "Design", "Implementation", "Testing")
- **Sequence**: A group of related tasks that accomplish a specific goal
- **Task**: A single unit of work with clear acceptance criteria
- **Quality Gates**: Validation checkpoints at the end of sequences (testing, code review, etc.)

## What fest Does

fest helps you create, navigate, validate, and execute festivals:

- **Learn**: Built-in documentation teaches agents the methodology (`fest intro`, `fest understand`)
- **Create**: Interactive TUI for scaffolding festivals, phases, sequences, and tasks
- **Validate**: Check festival structure for issues and auto-fix common problems
- **Navigate**: Quick commands to jump between festivals, phases, and sequences
- **Execute**: Orchestrate task execution with progress tracking
- **Track**: Monitor completion status across all levels

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
