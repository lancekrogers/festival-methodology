# Test recipes
BUILDTOOL := "go run ./internal/buildutil"

# Show available test commands
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run all tests (unit + integration) with visual dashboard
[no-cd]
all:
    @{{BUILDTOOL}} all

# Run unit tests with visual dashboard
[no-cd]
unit:
    @{{BUILDTOOL}} test

# Run unit tests with verbose output
[no-cd]
unit-verbose:
    @{{BUILDTOOL}} -v test

# Run integration tests with visual dashboard
[no-cd]
integration:
    @{{BUILDTOOL}} integration

# Run integration tests with verbose output
[no-cd]
integration-verbose:
    @{{BUILDTOOL}} -v integration

# Run tests with coverage (using go test directly)
[no-cd]
coverage:
    go test -cover ./...

# Generate HTML coverage report
[no-cd]
coverage-html:
    go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# ============================================
# Fallback commands (bypass buildutil)
# ============================================

# Run unit tests directly (fallback)
[no-cd]
unit-raw:
    go test -short -timeout 30s ./...

# Run unit tests with verbose output (fallback)
[no-cd]
unit-raw-verbose:
    go test -v -short -timeout 30s ./...

# Run integration tests directly (fallback)
[no-cd]
integration-raw:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Building Linux binary..."
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o bin/linux/fest ./cmd/fest
    echo "Running integration tests..."
    go test -v -tags=integration -timeout 5m ./tests/integration/...

# Run all tests directly (fallback)
[no-cd]
all-raw: unit-raw integration-raw
    @echo "✅ All tests passed (raw mode)"
