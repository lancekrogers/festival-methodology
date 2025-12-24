# Contributing to Fest

This document covers development setup, testing, code style guidelines, and the PR review process.

## Development Setup

### Prerequisites

- **Go 1.21+** - [Download](https://go.dev/dl/)
- **just** - Command runner ([Installation](https://github.com/casey/just#installation))
- **Docker** - Required for integration tests

### Getting Started

```bash
# Clone repository
git clone https://github.com/lancekrogers/festival-methodology.git
cd festival-methodology/fest

# Install dependencies
go mod download

# Build and test
just all

# Install fest to $GOBIN
just install
```

### Directory Structure

```text
fest/
├── cmd/fest/           # Binary entry point
├── internal/           # Private packages
│   ├── commands/       # CLI command implementations
│   ├── config/         # Configuration loading
│   ├── errors/         # Structured error types
│   ├── extensions/     # Extension system
│   ├── festival/       # Festival operations
│   ├── fileops/        # File utilities
│   ├── gates/          # Quality gate policies
│   ├── plugins/        # Plugin system
│   ├── response/       # JSON output formatting
│   ├── template/       # Template rendering
│   ├── validator/      # Validation checks
│   └── ...
├── tests/integration/  # Container-based tests
├── docs/               # Documentation
└── Justfile           # Build recipes
```

---

## Build Commands

```bash
# Show all available commands
just

# Build binary to bin/fest
just build

# Install to $GOBIN
just install

# Format code
just fmt

# Run linters
just lint

# Clean build artifacts
just clean

# Update dependencies
just deps
```

---

## Running Tests

```bash
# Run all tests (unit + integration + build)
just all

# Unit tests only
just test unit

# Unit tests with coverage
just test coverage

# HTML coverage report
just test coverage-html

# Integration tests (requires Docker)
just test integration
```

### Test Coverage Targets

| Package | Minimum | Target |
|---------|---------|--------|
| `internal/errors` | 80% | 90%+ |
| `internal/response` | 90% | 95%+ |
| `internal/validator` | 30% | 60%+ |
| `internal/commands` | 20% | 50%+ |
| Overall | 40% | 60%+ |

---

## Code Style Guidelines

### Context Propagation

**MANDATORY**: All I/O and long-running operations must accept `context.Context`:

```go
// Good
func (v *Validator) Validate(ctx context.Context, path string) ([]Issue, error) {
    if err := ctx.Err(); err != nil {
        return nil, err  // Check cancellation early
    }
    // ... validation logic
}

// Bad - missing context
func (v *Validator) Validate(path string) ([]Issue, error) {
    // Cannot be cancelled
}
```

### Error Handling

Use the `internal/errors` package for structured errors:

```go
// Good - structured error with context
return errors.NotFound("festival").
    WithOp("validate").
    WithField("path", festivalPath)

// Bad - plain error
return fmt.Errorf("festival not found: %s", festivalPath)
```

Error codes:

- `NOT_FOUND` - Resource doesn't exist
- `VALIDATION` - Input validation failure
- `IO` - File system error
- `CONFIG` - Configuration error
- `TEMPLATE` - Template rendering error
- `PARSE` - Parsing error
- `INTERNAL` - Unexpected error

### JSON Output

Use `internal/response` for consistent JSON output:

```go
// Good - use response package
import "github.com/lancekrogers/festival-methodology/fest/internal/response"

func emitResult(result MyResult) error {
    return response.Encode(os.Stdout, result)
}

// Bad - inline encoder
enc := json.NewEncoder(os.Stdout)
enc.SetIndent("", "  ")
enc.Encode(result)
```

### File and Function Limits

| Metric | Limit | Action |
|--------|-------|--------|
| File size | 500 LOC | Split into multiple files |
| Function size | 50 LOC | Extract helper functions |
| Function parameters | 5 | Use options struct |
| Interface methods | 5 | Split interface |

### Naming Conventions

| Element | Convention | Example |
|---------|------------|---------|
| Interfaces | `-er` suffix | `Validator`, `Loader`, `Renderer` |
| Constructors | `New` prefix | `NewValidator()`, `NewLoader()` |
| Getters | No `Get` prefix | `Name()` not `GetName()` |
| Constants | UPPER_SNAKE | `ErrCodeNotFound`, `PolicyFileName` |
| Package-private | lowercase | `loaderImpl`, `validatorImpl` |

### Dependencies

- **Inject dependencies**, don't create inline:

```go
// Good - dependency injection
type Service struct {
    loader  Loader
    config  *Config
}

func NewService(loader Loader, config *Config) *Service {
    return &Service{loader: loader, config: config}
}

// Bad - hidden dependency
func (s *Service) Process() {
    loader := NewLoader()  // Hidden dependency
}
```

- Prefer standard library over external dependencies
- External deps require justification: 2-3x value in quality/speed

### Testing Requirements

- **Table-driven tests** for multiple scenarios
- **Context cancellation tests** for I/O operations
- **Error cases first**, happy paths second
- **Test behavior**, not implementation

```go
func TestValidator(t *testing.T) {
    ctx := context.Background()

    tests := []struct {
        name       string
        setup      func(t *testing.T) string
        wantIssues int
        wantErr    bool
    }{
        {
            name: "valid structure",
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                // Setup valid structure
                return dir
            },
            wantIssues: 0,
        },
        {
            name: "invalid structure",
            setup: func(t *testing.T) string {
                dir := t.TempDir()
                // Setup invalid structure
                return dir
            },
            wantIssues: 1,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := tt.setup(t)
            v := NewValidator()
            issues, err := v.Validate(ctx, dir)

            if (err != nil) != tt.wantErr {
                t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
            }
            if len(issues) != tt.wantIssues {
                t.Errorf("Validate() issues = %d, want %d", len(issues), tt.wantIssues)
            }
        })
    }
}
```

---

## PR Checklist

Before submitting a PR, verify:

### Code Quality

- [ ] All tests pass (`just all`)
- [ ] New code has tests
- [ ] Coverage meets package targets
- [ ] No magic numbers or strings
- [ ] Error messages are helpful and contextual

### Style

- [ ] Code formatted (`just fmt`)
- [ ] Linter passes (`just lint`)
- [ ] Functions under 50 LOC
- [ ] Files under 500 LOC
- [ ] Interfaces have 5 or fewer methods

### Patterns

- [ ] Context propagated through I/O operations
- [ ] Errors wrapped with `internal/errors` package
- [ ] JSON output uses `internal/response` package
- [ ] Dependencies injected, not created inline

### Checklist: Documentation

- [ ] Godoc comments on exported types/functions
- [ ] Complex logic has inline comments
- [ ] README updated if adding features

### Testing

- [ ] Table-driven tests for multiple cases
- [ ] Context cancellation tested
- [ ] Error paths tested
- [ ] JSON output format verified

---

## Commit Messages

Follow conventional commit format:

```text
<type>: <description>

[optional body]
```

Types:

- `feat`: New feature
- `fix`: Bug fix
- `refactor`: Code restructuring
- `docs`: Documentation only
- `test`: Adding/updating tests
- `chore`: Maintenance tasks

Examples:

```text
feat: add gates validate subcommand

fix: handle context cancellation in validator

refactor: extract template rendering helpers

docs: add architecture documentation

test: add context cancellation tests for validator
```

---

## Getting Help

- **Issues**: Report bugs or request features
- **Discussions**: Ask questions or propose ideas
- **Code Review**: All PRs require review before merge

---

## Related Documentation

See these docs for more details:

- [Architecture](architecture.md) - Package structure and data flow
- [Commands](commands.md) - Complete CLI reference
- [Templates](templates.md) - Template system guide
- [Plugins](plugins.md) - Plugin and extension development
