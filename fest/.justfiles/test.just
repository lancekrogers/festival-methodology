# Test recipes
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

# Show available test commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run all tests (unit + integration)
[no-cd]
all: unit integration
    @echo "✅ All tests passed"

# Run unit tests (quick feedback)
[no-cd]
unit:
    gotestsum --format pkgname-and-test-fails -- ./...


# Run integration tests
[no-cd]
integration:
    #!/usr/bin/env bash
    set -euo pipefail
    # Build test binary first
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}
    gotestsum --format pkgname-and-test-fails -- -tags=integration -timeout 10m ./tests/integration/...

# Run tests with coverage
[no-cd]
coverage:
    gotestsum --format pkgname-and-test-fails -- -cover ./...

# Generate HTML coverage report
[no-cd]
coverage-html:
    go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Watch mode - rerun tests on file changes
[no-cd]
watch:
    gotestsum --watch --format pkgname-and-test-fails
