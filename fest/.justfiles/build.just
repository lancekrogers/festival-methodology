# Build recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

[private]
@default:
    @echo "Build Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Build fest binary into bin directory
build:
    go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Quick development build (no optimization)
dev:
    go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Build for Linux AMD64
build-linux:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

# Build for macOS Intel
build-darwin:
    GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

# Build for macOS Apple Silicon
build-darwin-arm64:
    GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

# Build for Windows AMD64
build-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-windows-amd64.exe {{main_path}}

# Build for all platforms
build-all: build-linux build-darwin build-darwin-arm64 build-windows

# Build Linux binary for container testing
build-test-binary:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}
