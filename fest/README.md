# Fest CLI

![Festival Methodology Banner](docs/images/banner.jpg)

Fest is a CLI tool for working with **Festival Methodology** - a hierarchical agentic planning and execution system designed for AI agent workflows.

## What is Festival Methodology?

Festival Methodology is a structured approach to **hierarchical agentic planning and execution**. It organizes complex projects into a three-level hierarchy that AI agents can systematically work through:

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
- **Just-in-Time Context**: Agents only load the context they need for the current task, minimizing token usage while maintaining full awareness of the project structure.

**Key Concepts:**

- **Festival**: A complete project or initiative with a defined outcome
- **Phase**: A major milestone (e.g., "Design", "Implementation", "Testing")
- **Sequence**: A group of related tasks that accomplish a specific goal
- **Task**: A markdown document containing a unit of work. Similar to a Claude Code plan - each task document contains multiple actions, acceptance criteria, and context. Tasks are not single to-dos; they're comprehensive work units that may include many steps.
- **Quality Gates**: Validation checkpoints at the end of sequences (testing, code review, etc.)

## What fest Does

Fest is both a **project scaffolding tool** and an **agent guidance system**. It teaches agents how to work with Festival Methodology and guides them through execution with minimal context overhead.

**Agent Guidance System:**
- Built-in documentation teaches agents the methodology on-demand (`fest intro`, `fest understand`)
- Agents learn what they need, when they need it - no upfront context dump
- `fest next` shows agents exactly what to work on next with just the context they need
- `fest execute` orchestrates task execution, guiding agents through the festival systematically
- Self-documenting commands guide agents through proper usage

**Project Management:**
- **Create**: Interactive TUI for scaffolding festivals, phases, sequences, and tasks
- **Validate**: Check festival structure for issues and auto-fix common problems
- **Navigate**: Quick commands to jump between festivals, phases, and sequences
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
