# Test recipes
root := justfile_directory()
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

# Show available test commands (default)
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Run all tests
test:
    cd {{root}} && go test -v ./...

# Run unit tests only
unit:
    cd {{root}} && go test -v ./...

# Run tests with coverage report
coverage:
    cd {{root}} && go test -v -cover ./...

# Generate HTML coverage report
coverage-html:
    cd {{root}} && go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out -o coverage.html

# Build Linux binary for container testing
[private]
build-test-binary:
    cd {{root}} && GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}

# Run integration tests
integration: build-test-binary
    cd {{root}} && go test -v -tags=integration -timeout 10m ./tests/integration/...
