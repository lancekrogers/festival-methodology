# Fest CLI Command Reference

Complete reference for all `fest` commands with flags, examples, and JSON output formats.

## Global Flags

These flags are available on all commands:

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `~/.config/fest/config.json` | Config file path |
| `--verbose` | `false` | Enable verbose output |
| `--no-color` | `false` | Disable colored output |
| `--debug` | `false` | Enable debug logging |

---

## Creation Commands

### fest create festival

Create a new festival scaffold under `festivals/`.

```bash
fest create festival --name NAME [flags]
```

#### create festival: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | *required* | Festival name |
| `--dest` | `active` | Destination: `active` or `planned` |
| `--goal` | | Festival goal description |
| `--tags` | | Comma-separated tags |
| `--vars-file` | | JSON file with variables |
| `--json` | `false` | Emit JSON output |

#### create festival: Examples

```bash
# Create festival in active/
fest create festival --name "api-refactor"

# Create in planned/ with goal
fest create festival --name "ui-redesign" --dest planned --goal "Modernize UI components"

# With JSON output
fest create festival --name "test" --json
```

#### create festival: JSON Output

```json
{
  "ok": true,
  "action": "create_festival",
  "festival": "api-refactor",
  "path": "/path/to/festivals/active/api-refactor",
  "files_created": [
    "FESTIVAL_OVERVIEW.md",
    "FESTIVAL_RULES.md",
    "TODO.md"
  ]
}
```

---

### fest create phase

Insert a new phase into a festival.

```bash
fest create phase --festival FESTIVAL --name NAME [flags]
```

#### create phase: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--festival` | *required* | Festival path or name |
| `--name` | *required* | Phase name (UPPERCASE) |
| `--position` | *append* | Position number |
| `--goal` | | Phase goal |
| `--json` | `false` | Emit JSON output |

#### create phase: Examples

```bash
fest create phase --festival my-fest --name IMPLEMENTATION
fest create phase --festival ./active/my-fest --name PLANNING --position 1
```

---

### fest create sequence

Insert a new sequence into a phase.

```bash
fest create sequence --phase PHASE --name NAME [flags]
```

#### create sequence: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--phase` | *required* | Phase path |
| `--name` | *required* | Sequence name (lowercase) |
| `--position` | *append* | Position number |
| `--goal` | | Sequence goal |
| `--json` | `false` | Emit JSON output |

#### create sequence: Examples

```bash
fest create sequence --phase ./001_PLANNING --name requirements
fest create sequence --phase ./002_IMPLEMENTATION --name api_layer --position 1
```

---

### fest create task

Insert new task file(s) into a sequence.

```bash
fest create task --sequence SEQUENCE --name NAME [flags]
```

#### create task: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--sequence` | *required* | Sequence path |
| `--name` | *required* | Task name(s), comma-separated |
| `--position` | *append* | Position number |
| `--batch` | `false` | Create multiple tasks from names |
| `--json` | `false` | Emit JSON output |

#### create task: Examples

```bash
fest create task --sequence ./01_requirements --name gather_specs
fest create task --sequence ./01_api --name "design,implement,test" --batch
```

---

## Validation Commands

### fest validate

Validate festival methodology compliance.

```bash
fest validate [festival-path] [flags]
```

#### validate: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--fix` | `false` | Automatically apply safe fixes |
| `--json` | `false` | Output results as JSON |

#### validate: Subcommands

| Command | Description |
|---------|-------------|
| `structure` | Validate naming conventions and hierarchy |
| `completeness` | Validate required files exist |
| `tasks` | Validate task files exist (CRITICAL) |
| `quality-gates` | Validate quality gates exist |
| `checklist` | Post-completion questionnaire |

#### validate: Examples

```bash
fest validate                        # Validate current festival
fest validate ./active/my-fest       # Validate specific festival
fest validate --fix                  # Auto-fix safe issues
fest validate --json                 # JSON output
fest validate tasks                  # Only check task files
fest validate quality-gates --fix    # Add missing quality gates
```

#### validate: JSON Output

```json
{
  "ok": true,
  "action": "validate",
  "festival": "my-fest",
  "valid": true,
  "score": 85,
  "issues": [
    {
      "level": "warning",
      "code": "missing_quality_gates",
      "path": "002_IMPL/01_api",
      "message": "Sequence missing quality gates",
      "fix": "Run fest validate --fix",
      "auto_fixable": true
    }
  ]
}
```

---

## Learning Commands

### fest understand

Learn Festival Methodology concepts.

```bash
fest understand [topic] [flags]
```

#### understand: Topics

| Topic | Description |
|-------|-------------|
| `methodology` | Core principles and philosophy |
| `structure` | Three-level hierarchy with examples |
| `tasks` | When and how to create task files (CRITICAL) |
| `templates` | Template variables that save tokens |
| `workflow` | Just-in-time reading patterns |
| `rules` | MANDATORY structure rules |
| `gates` | Quality gate configuration |
| `extensions` | Loaded extensions |
| `plugins` | Discovered plugins |
| `resources` | What's in `.festival/` |
| `checklist` | Quick validation checklist |

#### understand: Examples

```bash
fest understand                      # Show all topics
fest understand methodology          # Core principles
fest understand tasks                # Critical task file guidance
fest understand gates                # Quality gate setup
```

---

## Quality Gate Commands

### fest gates

Manage quality gate policies and create gate task files.

Quality gates are validation steps appended to implementation sequences.
Configuration is merged from multiple sources with precedence:
1. Built-in defaults
2. `fest.yaml` `quality_gates.tasks`
3. Festival-level policy (`.festival/gates.yml`)
4. Phase-level override (`.fest.gates.yml`)
5. Sequence-level override (`.fest.gates.yml`)

```bash
fest gates [command] [flags]
```

#### gates: Subcommands

| Command | Description |
|---------|-------------|
| `show` | Show merged gate policy with sources |
| `list` | List available named policies |
| `apply` | Create quality gate task files in sequences |
| `init` | Initialize fest.yaml or override file |
| `validate` | Validate gate configuration |

#### gates: Examples

```bash
fest gates show                      # Show merged policy
fest gates show --json               # JSON output with sources
fest gates list                      # List all policies
fest gates apply                     # Preview gate creation (dry-run)
fest gates apply --approve           # Create gate task files
fest gates apply --sequence 002_IMPLEMENT/01_core --approve
fest gates init                      # Create fest.yaml
fest gates init --phase 002_IMPLEMENT  # Create phase override
```

#### gates: JSON Output

```json
{
  "ok": true,
  "action": "gates_list",
  "policies": [
    {
      "name": "default",
      "description": "Standard quality gates",
      "gates": ["testing", "code_review", "iterate"]
    },
    {
      "name": "minimal",
      "description": "Minimal validation only",
      "gates": ["testing"]
    }
  ]
}
```

---

## Navigation Commands

### fest go

Navigate to festivals directory.

```bash
fest go [target] [flags]
```

#### go: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workspace` | `false` | Show detected workspace |
| `--all` | `false` | List all registered festivals |
| `--json` | `false` | Output in JSON format |

#### go: Shell Integration

Add to `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(fest shell-init zsh)"
```

Then use `fgo`:

```bash
fgo              # Navigate to festivals root
fgo 002          # Navigate to phase 002
fgo 2/1          # Navigate to phase 2, sequence 1
```

#### go: Without Shell Integration

```bash
cd $(fest go)
cd $(fest go 002)
```

---

## Organization Commands

### fest reorder

Reorder phases, sequences, or tasks.

```bash
fest reorder [phase|sequence|task] [flags]
```

#### reorder: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--skip-dry-run` | `false` | Skip preview, apply immediately |
| `--backup` | `false` | Create backup before reordering |
| `--force` | `false` | Skip confirmation prompts |
| `--verbose` | `false` | Show detailed output |

#### reorder: Examples

```bash
fest reorder phase --from 3 --to 1           # Move phase 3 to position 1
fest reorder sequence --phase ./001 --from 2 --to 4
fest reorder task --sequence ./01_api --from 1 --to 3
fest reorder phase --from 2 --to 1 --skip-dry-run  # Apply immediately
```

---

### fest renumber

Renumber elements after manual changes.

```bash
fest renumber [phase|sequence|task] [flags]
```

#### renumber: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--skip-dry-run` | `false` | Skip preview, apply immediately |
| `--backup` | `false` | Create backup before renumbering |
| `--start` | `1` | Starting number |
| `--verbose` | `false` | Show detailed output |

#### renumber: Examples

```bash
fest renumber phase                          # Renumber all phases
fest renumber sequence --phase ./001_PLAN    # Renumber sequences in phase
fest renumber task --sequence ./01_api       # Renumber tasks in sequence
fest renumber phase --start 100              # Start from 100
```

---

### fest remove

Remove elements and renumber.

```bash
fest remove [phase|sequence|task] [flags]
```

#### remove: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--force` | `false` | Skip confirmation prompts |
| `--backup` | `false` | Create backup before removal |
| `--verbose` | `false` | Show detailed output |

#### remove: Examples

```bash
fest remove phase 003                        # Remove phase 003
fest remove sequence --phase ./001 02        # Remove sequence 02
fest remove task --sequence ./01_api 03      # Remove task 03
```

---

### fest insert

Insert elements and renumber.

```bash
fest insert [phase|sequence|task] [flags]
```

#### insert: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--backup` | `false` | Create backup before changes |
| `--verbose` | `false` | Show detailed output |

#### insert: Examples

```bash
fest insert phase --name TESTING --at 2
fest insert sequence --phase ./001 --name design --at 1
fest insert task --sequence ./01_api --name validate --at 2
```

---

## Utility Commands

### fest apply

Apply a template to a destination file.

```bash
fest apply --to DEST [flags]
```

#### apply: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--to` | *required* | Destination file path |
| `--template-id` | | Template ID or alias |
| `--template-path` | | Path to template file |
| `--vars-file` | | JSON file with variables |
| `--json` | `false` | Emit JSON output |

#### apply: Examples

```bash
fest apply --template-id TASK --to ./01_new_task.md
fest apply --template-path ./custom.md --to ./output.md --vars-file vars.json
```

---

### fest count

Count tokens in files or directories.

```bash
fest count [file|directory] [flags]
```

#### count: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-r, --recursive` | `false` | Count all files in directory |
| `-d, --directory` | `false` | Alias for `--recursive` |
| `--model` | | Tokenizer model (gpt-4, gpt-3.5-turbo, claude-3) |
| `--all` | `false` | Show all counting methods |
| `--cost` | `false` | Include cost estimates |
| `--json` | `false` | Output in JSON format |
| `--chars-per-token` | `4` | Characters per token ratio |
| `--words-per-token` | `0.75` | Words per token ratio |

#### count: Examples

```bash
fest count document.md               # Count tokens in file
fest count --model gpt-4 doc.md      # Use GPT-4 tokenizer
fest count --all --cost doc.md       # All methods with costs
fest count -r ./src                  # Count directory recursively
fest count -r --json ./project       # Directory with JSON output
```

#### count: JSON Output

```json
{
  "ok": true,
  "action": "count",
  "file": "document.md",
  "counts": {
    "tiktoken_gpt4": 1234,
    "chars_approx": 1180,
    "words_approx": 1050
  },
  "cost_estimates": {
    "gpt-4-input": "$0.03",
    "gpt-4-output": "$0.06"
  }
}
```

---

### fest init

Initialize a new festivals directory structure.

```bash
fest init [flags]
```

#### init: Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--register` | `false` | Register as active workspace |

#### init: Examples

```bash
fest init                            # Create festivals/ structure
fest init --register                 # Also register workspace
```

---

## System Commands

These commands maintain the fest tool itself (templates, configuration) - NOT your festival content.

### fest system

Parent command for system maintenance operations.

```bash
fest system [command]
```

#### system: Subcommands

| Command | Description |
|---------|-------------|
| `sync` | Download latest templates from GitHub |
| `update` | Update .festival/ methodology files from templates |

---

### fest system sync

Download latest fest templates from GitHub to local cache.

```bash
fest system sync [flags]
```

#### system sync: Examples

```bash
fest system sync                     # Download latest templates
fest system sync --verbose           # Show detailed progress
fest system sync --force             # Overwrite existing cache
```

---

### fest system update

Update .festival/ methodology files from cached templates.

```bash
fest system update [flags]
```

#### system update: Examples

```bash
fest system update                   # Interactive update
fest system update --dry-run         # Preview changes
fest system update --no-interactive  # Skip modified files
fest system update --backup          # Create backup first
```

---

## Configuration Commands

### fest config

Manage configuration repositories.

```bash
fest config [command] [flags]
```

#### config: Subcommands

| Command | Description |
|---------|-------------|
| `show` | Show current configuration |
| `list` | List available config repos |
| `add` | Add a config repository |
| `remove` | Remove a config repository |
| `set-default` | Set default config repo |

---

### fest extension

Manage methodology extensions.

```bash
fest extension [command] [flags]
```

#### extension: Subcommands

| Command | Description |
|---------|-------------|
| `list` | List available extensions |
| `enable` | Enable an extension |
| `disable` | Disable an extension |

---

### fest index

Manage festival indices for Guild integration.

```bash
fest index [command] [flags]
```

#### index: Subcommands

| Command | Description |
|---------|-------------|
| `generate` | Generate festival index |
| `validate` | Validate index matches filesystem |

---

## Interactive Mode

### fest tui

Launch interactive terminal UI.

```bash
fest tui [flags]
```

#### tui: Examples

```bash
fest tui                             # Launch interactive mode
fest create                          # Also launches create TUI
```

---

## Shell Integration

### fest shell-init

Output shell integration code.

```bash
fest shell-init [shell] [flags]
```

#### shell-init: Supported Shells

- `zsh`
- `bash`
- `fish`

#### shell-init: Setup

```bash
# Add to ~/.zshrc
eval "$(fest shell-init zsh)"

# Add to ~/.bashrc
eval "$(fest shell-init bash)"

# Add to ~/.config/fish/config.fish
fest shell-init fish | source
```

This provides the `fgo` function for quick navigation.
