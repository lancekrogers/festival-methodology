# Test recipes
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

# Show available test commands (default)
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run unit tests only
[no-cd]
unit:
    go test -v ./...

# Run tests with coverage report
[no-cd]
coverage:
    go test -v -cover ./...

# Generate HTML coverage report
[no-cd]
coverage-html:
    go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Build Linux binary for container testing
[private, no-cd]
build-test-binary:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}

# Run integration tests
[no-cd]
integration: build-test-binary
    go test -v -tags=integration -timeout 10m ./tests/integration/...
