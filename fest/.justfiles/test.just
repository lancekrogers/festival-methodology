# Test recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

[private]
@default:
    @echo "Test Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Run tests with verbose output
test:
    go test -v ./...

# Run tests with coverage report
test-coverage:
    go test -v -cover ./...

# Generate HTML coverage report
test-coverage-html:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

# Build Linux binary for container testing
[private]
build-test-binary:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}

# Run integration tests
test-integration: build-test-binary
    go test -v -tags=integration -timeout 10m ./tests/integration/...
