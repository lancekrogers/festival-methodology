# Fest CLI Architecture

This document describes the internal architecture, package structure, and data flow of the `fest` CLI tool.

## Package Dependency Graph

```
cmd/fest/
    └── main.go ─────────────────────────────────────┐
                                                     │
                                                     ▼
internal/commands/  ◄──────────────────────────────────────────────────────────────┐
    ├── root.go        (Cobra setup, CLIConfig injection)                         │
    ├── cliconfig.go   (Context-based configuration)                              │
    ├── create*.go     (Festival/phase/sequence/task creation)                    │
    ├── validate.go    (Validation orchestration)                                 │
    ├── understand.go  (Methodology learning)                                     │
    ├── gates.go       (Quality gate management)                                  │
    ├── tui*.go        (Interactive terminal UI)                                  │
    └── ... (40+ command files)                                                   │
            │                                                                     │
            ├─────────────────────┬────────────────────┬───────────────────┬──────┘
            ▼                     ▼                    ▼                   ▼
    internal/config/      internal/template/   internal/validator/  internal/response/
    ├── config.go         ├── manager.go       ├── types.go         └── json.go
    └── loader.go         ├── helpers.go       ├── structure.go
                          ├── renderer.go      ├── completeness.go
                          └── catalog.go       ├── templates.go
                                               └── gates.go
            │                     │                    │
            └─────────────────────┴────────────────────┘
                                  │
                                  ▼
                          internal/fileops/
                          ├── fileops.go
                          └── (file I/O utilities)
```

## Package Responsibilities

### Entry Point

| Package | File | Purpose |
|---------|------|---------|
| `cmd/fest` | `main.go` | Binary entry point; calls `commands.Execute()` |

### Command Layer (`internal/commands/`)

| File | LOC | Responsibility |
|------|-----|----------------|
| `root.go` | 129 | Cobra root command, CLI flag binding, context injection |
| `cliconfig.go` | 97 | `CLIConfig` struct, context storage/retrieval |
| `validate.go` | 634 | Validation command orchestration |
| `understand.go` | 876 | Methodology learning and explanation |
| `task_defaults.go` | 743 | Quality gate task management |
| `tui_charm.go` | 735 | BubbleTea-based interactive UI |
| `gates.go` | 350 | Gate policy listing and management |
| `config_repo.go` | 368 | Configuration repository management |

### Domain Logic

| Package | Purpose | Key Types |
|---------|---------|-----------|
| `internal/config` | Configuration loading/saving | `Config`, `Repository`, `Behavior` |
| `internal/template` | Template rendering | `Manager`, `Renderer`, `Loader`, `Context` |
| `internal/validator` | Festival validation | `Validator` interface, `Issue`, `Result` |
| `internal/festival` | Festival model and operations | Parser, Renumber, Reorder |
| `internal/gates` | Quality gate policies | `GatePolicy`, `GateTask`, `PolicyLoader` |
| `internal/errors` | Structured error types | `Error`, error codes |
| `internal/response` | JSON output formatting | `Response`, `Emitter`, `Encode()` |

### Infrastructure

| Package | Purpose |
|---------|---------|
| `internal/fileops` | File system operations |
| `internal/github` | GitHub API integration for downloads |
| `internal/extensions` | Extension loading and management |
| `internal/plugins` | Plugin system |
| `internal/workspace` | Workspace detection and configuration |
| `internal/index` | Guild integration index generation |
| `internal/tokens` | Token counting for LLM context |
| `internal/ui` | Terminal UI utilities |

## Data Flow

### Festival Creation

```
User: fest create festival --name "my-fest"
         │
         ▼
    commands/create_festival.go
    ├── Parse CLI flags
    ├── Validate input
    │       ▼
    template.Manager
    ├── Load template from catalog (or file fallback)
    ├── Create Context with variables
    ├── Render template
    │       ▼
    fileops
    ├── Create directory structure
    ├── Write rendered files
    │       ▼
    response.Encode()
    └── Emit JSON result to stdout
```

### Validation Workflow

```
User: fest validate --json
         │
         ▼
    commands/validate.go
    ├── Find festival root
    ├── Create validator instances:
    │   ├── StructureValidator
    │   ├── CompletenessValidator
    │   ├── TemplateValidator
    │   └── GatesValidator
    │       ▼
    validator.Validator interface
    ├── Each validator.Validate(ctx, path)
    ├── Collect []Issue from all validators
    │       ▼
    validator.Result
    ├── Aggregate issues
    ├── Calculate score
    │       ▼
    response.Encode()
    └── Emit JSON result
```

### Template Rendering (with Fallback)

```
commands/create_phase.go
         │
         ▼
    template.Manager.RenderWithFallback()
    ├── Try catalog lookup by ID
    │   ├── Found: Load from embedded catalog
    │   └── Not found: Fall through
    ├── Try file path
    │   ├── Found: Load from filesystem
    │   └── Not found: Return error
    │       ▼
    template.Renderer.Render()
    ├── Parse Mustache/Go template syntax
    ├── Apply Context variables
    └── Return rendered content
```

## Configuration Loading

### Sequence

```
1. User invokes command with --config flag (optional)
2. root.go PersistentPreRunE:
   a. Create CLIConfig with flag values
   b. Store in context via ContextWithConfig()
3. Command handler:
   a. Retrieve config via ConfigFromContext(ctx)
   b. Use config values for behavior
4. config.Load() (if needed):
   a. Check $FEST_CONFIG_DIR or ~/.config/fest/
   b. Load config.json if exists
   c. Apply defaults for missing values
```

### Configuration Precedence

1. CLI flags (highest)
2. Environment variables (`FEST_CONFIG_DIR`)
3. Config file (`~/.config/fest/config.json`)
4. Built-in defaults (lowest)

## Error Handling Strategy

### Error Package (`internal/errors`)

Provides structured errors with:
- **Code**: Categorization (`NOT_FOUND`, `VALIDATION`, `IO`, etc.)
- **Op**: Operation name for stack context
- **Fields**: Key-value metadata
- **Err**: Wrapped underlying error

```go
// Create structured error
err := errors.NotFound("festival").
    WithOp("validate").
    WithField("path", "/path/to/festival")

// Check error code
if errors.Is(err, errors.ErrCodeNotFound) {
    // Handle not found
}
```

### Error Codes

| Code | Meaning |
|------|---------|
| `NOT_FOUND` | Resource does not exist |
| `VALIDATION` | Input validation failure |
| `IO` | File system operation failure |
| `CONFIG` | Configuration error |
| `TEMPLATE` | Template rendering failure |
| `PARSE` | Parsing error |
| `INTERNAL` | Unexpected internal error |
| `PERMISSION` | Permission denied |

### JSON Error Output

Commands with `--json` emit errors in a consistent format:

```json
{
  "ok": false,
  "action": "validate",
  "errors": [
    {
      "code": "NOT_FOUND",
      "message": "festival not found",
      "path": "/path/to/festival"
    }
  ]
}
```

## Context Propagation

### Pattern

All I/O and long-running operations accept `context.Context`:

```go
func (v *StructureValidator) Validate(ctx context.Context, path string) ([]Issue, error) {
    if err := ctx.Err(); err != nil {
        return nil, err  // Respect cancellation
    }
    // ... validation logic
}
```

### Context Storage

CLI configuration flows through context:

```go
// In root.go PersistentPreRunE:
cmd.SetContext(ContextWithConfig(ctx, globalCfg))

// In command handler:
cfg := ConfigFromContext(cmd.Context())
```

## Key Interfaces

### Validator

```go
type Validator interface {
    Validate(ctx context.Context, path string) ([]Issue, error)
}
```

Implementations:
- `StructureValidator` - Directory naming conventions
- `CompletenessValidator` - Required files present
- `TemplateValidator` - No unfilled template markers
- `GatesValidator` - Quality gate compliance

### Template Loader/Renderer

```go
type Loader interface {
    Load(path string) (*Template, error)
    LoadAll(dir string) ([]*Template, error)
}

type Renderer interface {
    Render(tmpl *Template, ctx *Context) (string, error)
}
```

### Gate Interfaces

```go
type HierarchicalPolicyLoader interface {
    LoadForSequence(ctx context.Context, festivalPath, phasePath, sequencePath string) (*EffectivePolicy, error)
    LoadForPhase(ctx context.Context, festivalPath, phasePath string) (*EffectivePolicy, error)
    LoadForFestival(ctx context.Context, festivalPath string) (*EffectivePolicy, error)
}

type PolicyRegistrar interface {
    Get(name string) (*PolicyInfo, bool)
    GetPolicy(name string) (*GatePolicy, error)
    List() []string
    Refresh()
}
```

## Directory Structure Requirements

```
festivals/
├── active/                    # Currently executing festivals
│   └── my-festival/
│       ├── FESTIVAL_OVERVIEW.md
│       ├── FESTIVAL_RULES.md  (optional)
│       ├── TODO.md            (optional)
│       ├── 001_PHASE_NAME/
│       │   ├── PHASE_GOAL.md
│       │   └── 01_sequence_name/
│       │       └── 01_task_name.md
│       └── ...
├── completed/
├── planned/
└── archived/
```

### Naming Conventions

- **Phases**: `NNN_UPPERCASE_NAME/` (e.g., `001_PLANNING/`)
- **Sequences**: `NN_lowercase_name/` (e.g., `01_requirements/`)
- **Tasks**: `NN_lowercase_name.md` (e.g., `01_gather_requirements.md`)

## Testing Strategy

| Test Type | Location | Coverage |
|-----------|----------|----------|
| Unit | `*_test.go` alongside code | Response: 94%, Validator: 30% |
| Integration | `tests/integration/` | Container-based workflow tests |
| TUI | `*_test.go` using teatest | Interactive UI testing |

### Test Patterns

- Table-driven tests for multiple scenarios
- `t.TempDir()` for filesystem isolation
- Context cancellation verification
- Docker containers for full workflow testing

## Build Information

Build-time variables injected via `-ldflags`:
- `Version`: Semantic version
- `BuildTime`: Build timestamp
- `GitCommit`: Git commit hash

Access via `fest --version` or programmatically through `commands.Version`.
