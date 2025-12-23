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

```
fest create festival --name NAME [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--name` | *required* | Festival name |
| `--dest` | `active` | Destination: `active` or `planned` |
| `--goal` | | Festival goal description |
| `--tags` | | Comma-separated tags |
| `--vars-file` | | JSON file with variables |
| `--json` | `false` | Emit JSON output |

#### Examples

```bash
# Create festival in active/
fest create festival --name "api-refactor"

# Create in planned/ with goal
fest create festival --name "ui-redesign" --dest planned --goal "Modernize UI components"

# With JSON output
fest create festival --name "test" --json
```

#### JSON Output

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

```
fest create phase --festival FESTIVAL --name NAME [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--festival` | *required* | Festival path or name |
| `--name` | *required* | Phase name (UPPERCASE) |
| `--position` | *append* | Position number |
| `--goal` | | Phase goal |
| `--json` | `false` | Emit JSON output |

#### Examples

```bash
fest create phase --festival my-fest --name IMPLEMENTATION
fest create phase --festival ./active/my-fest --name PLANNING --position 1
```

---

### fest create sequence

Insert a new sequence into a phase.

```
fest create sequence --phase PHASE --name NAME [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--phase` | *required* | Phase path |
| `--name` | *required* | Sequence name (lowercase) |
| `--position` | *append* | Position number |
| `--goal` | | Sequence goal |
| `--json` | `false` | Emit JSON output |

#### Examples

```bash
fest create sequence --phase ./001_PLANNING --name requirements
fest create sequence --phase ./002_IMPLEMENTATION --name api_layer --position 1
```

---

### fest create task

Insert new task file(s) into a sequence.

```
fest create task --sequence SEQUENCE --name NAME [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--sequence` | *required* | Sequence path |
| `--name` | *required* | Task name(s), comma-separated |
| `--position` | *append* | Position number |
| `--batch` | `false` | Create multiple tasks from names |
| `--json` | `false` | Emit JSON output |

#### Examples

```bash
fest create task --sequence ./01_requirements --name gather_specs
fest create task --sequence ./01_api --name "design,implement,test" --batch
```

---

## Validation Commands

### fest validate

Validate festival methodology compliance.

```
fest validate [festival-path] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--fix` | `false` | Automatically apply safe fixes |
| `--json` | `false` | Output results as JSON |

#### Subcommands

| Command | Description |
|---------|-------------|
| `structure` | Validate naming conventions and hierarchy |
| `completeness` | Validate required files exist |
| `tasks` | Validate task files exist (CRITICAL) |
| `quality-gates` | Validate quality gates exist |
| `checklist` | Post-completion questionnaire |

#### Examples

```bash
fest validate                        # Validate current festival
fest validate ./active/my-fest       # Validate specific festival
fest validate --fix                  # Auto-fix safe issues
fest validate --json                 # JSON output
fest validate tasks                  # Only check task files
fest validate quality-gates --fix    # Add missing quality gates
```

#### JSON Output

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

```
fest understand [topic] [flags]
```

#### Topics

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

#### Examples

```bash
fest understand                      # Show all topics
fest understand methodology          # Core principles
fest understand tasks                # Critical task file guidance
fest understand gates                # Quality gate setup
```

---

## Quality Gate Commands

### fest task defaults

Manage quality gate default tasks.

```
fest task defaults [command] [flags]
```

#### Subcommands

| Command | Description |
|---------|-------------|
| `show` | Show current `fest.yaml` configuration |
| `init` | Create a default `fest.yaml` file |
| `sync` | Sync quality gate tasks to all sequences |
| `add` | Add quality gate tasks to specific sequence |

#### Examples

```bash
fest task defaults show              # Show configuration
fest task defaults init              # Create fest.yaml
fest task defaults sync              # Sync all sequences
fest task defaults add ./01_api      # Add to specific sequence
fest task defaults sync --json       # JSON output
```

---

### fest gates

Manage hierarchical quality gate policies.

```
fest gates [command] [flags]
```

#### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List available named policies |
| `show` | Show effective gate policy |
| `apply` | Apply a named gate policy |
| `init` | Initialize an override file |
| `validate` | Validate gate configuration |

#### Examples

```bash
fest gates list                      # List all policies
fest gates list --json               # JSON output
fest gates show                      # Show effective policy
fest gates apply minimal             # Apply named policy
fest gates init                      # Create override file
```

#### JSON Output (list)

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

```
fest go [target] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--workspace` | `false` | Show detected workspace |
| `--all` | `false` | List all registered festivals |
| `--json` | `false` | Output in JSON format |

#### Shell Integration

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

#### Without Shell Integration

```bash
cd $(fest go)
cd $(fest go 002)
```

---

## Organization Commands

### fest reorder

Reorder phases, sequences, or tasks.

```
fest reorder [phase|sequence|task] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--skip-dry-run` | `false` | Skip preview, apply immediately |
| `--backup` | `false` | Create backup before reordering |
| `--force` | `false` | Skip confirmation prompts |
| `--verbose` | `false` | Show detailed output |

#### Examples

```bash
fest reorder phase --from 3 --to 1           # Move phase 3 to position 1
fest reorder sequence --phase ./001 --from 2 --to 4
fest reorder task --sequence ./01_api --from 1 --to 3
fest reorder phase --from 2 --to 1 --skip-dry-run  # Apply immediately
```

---

### fest renumber

Renumber elements after manual changes.

```
fest renumber [phase|sequence|task] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--skip-dry-run` | `false` | Skip preview, apply immediately |
| `--backup` | `false` | Create backup before renumbering |
| `--start` | `1` | Starting number |
| `--verbose` | `false` | Show detailed output |

#### Examples

```bash
fest renumber phase                          # Renumber all phases
fest renumber sequence --phase ./001_PLAN    # Renumber sequences in phase
fest renumber task --sequence ./01_api       # Renumber tasks in sequence
fest renumber phase --start 100              # Start from 100
```

---

### fest remove

Remove elements and renumber.

```
fest remove [phase|sequence|task] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--force` | `false` | Skip confirmation prompts |
| `--backup` | `false` | Create backup before removal |
| `--verbose` | `false` | Show detailed output |

#### Examples

```bash
fest remove phase 003                        # Remove phase 003
fest remove sequence --phase ./001 02        # Remove sequence 02
fest remove task --sequence ./01_api 03      # Remove task 03
```

---

### fest insert

Insert elements and renumber.

```
fest insert [phase|sequence|task] [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--dry-run` | `true` | Preview changes without applying |
| `--backup` | `false` | Create backup before changes |
| `--verbose` | `false` | Show detailed output |

#### Examples

```bash
fest insert phase --name TESTING --at 2
fest insert sequence --phase ./001 --name design --at 1
fest insert task --sequence ./01_api --name validate --at 2
```

---

## Utility Commands

### fest apply

Apply a template to a destination file.

```
fest apply --to DEST [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--to` | *required* | Destination file path |
| `--template-id` | | Template ID or alias |
| `--template-path` | | Path to template file |
| `--vars-file` | | JSON file with variables |
| `--json` | `false` | Emit JSON output |

#### Examples

```bash
fest apply --template-id TASK --to ./01_new_task.md
fest apply --template-path ./custom.md --to ./output.md --vars-file vars.json
```

---

### fest count

Count tokens in files or directories.

```
fest count [file|directory] [flags]
```

#### Flags

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

#### Examples

```bash
fest count document.md               # Count tokens in file
fest count --model gpt-4 doc.md      # Use GPT-4 tokenizer
fest count --all --cost doc.md       # All methods with costs
fest count -r ./src                  # Count directory recursively
fest count -r --json ./project       # Directory with JSON output
```

#### JSON Output

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

```
fest init [flags]
```

#### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--register` | `false` | Register as active workspace |

#### Examples

```bash
fest init                            # Create festivals/ structure
fest init --register                 # Also register workspace
```

---

### fest sync

Download latest templates from GitHub.

```
fest sync [flags]
```

#### Examples

```bash
fest sync                            # Download latest templates
fest sync --verbose                  # Show detailed progress
```

---

## Configuration Commands

### fest config

Manage configuration repositories.

```
fest config [command] [flags]
```

#### Subcommands

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

```
fest extension [command] [flags]
```

#### Subcommands

| Command | Description |
|---------|-------------|
| `list` | List available extensions |
| `enable` | Enable an extension |
| `disable` | Disable an extension |

---

### fest index

Manage festival indices for Guild integration.

```
fest index [command] [flags]
```

#### Subcommands

| Command | Description |
|---------|-------------|
| `generate` | Generate festival index |
| `validate` | Validate index matches filesystem |

---

## Interactive Mode

### fest tui

Launch interactive terminal UI.

```
fest tui [flags]
```

#### Examples

```bash
fest tui                             # Launch interactive mode
fest create                          # Also launches create TUI
```

---

## Shell Integration

### fest shell-init

Output shell integration code.

```
fest shell-init [shell] [flags]
```

#### Supported Shells

- `zsh`
- `bash`
- `fish`

#### Setup

```bash
# Add to ~/.zshrc
eval "$(fest shell-init zsh)"

# Add to ~/.bashrc
eval "$(fest shell-init bash)"

# Add to ~/.config/fish/config.fish
fest shell-init fish | source
```

This provides the `fgo` function for quick navigation.
