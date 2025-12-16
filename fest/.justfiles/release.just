# Release recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

[private]
@default:
    @echo "Release Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Build for macOS Intel (private helper)
[private]
build-darwin:
    GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

# Build for macOS Apple Silicon (private helper)
[private]
build-darwin-arm64:
    GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

# Build for Linux AMD64 (private helper)
[private]
build-linux:
    GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

# Build for Windows AMD64 (private helper)
[private]
build-windows:
    GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-windows-amd64.exe {{main_path}}

# Sign macOS release binaries
sign-release: build-darwin build-darwin-arm64
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Signing macOS release binaries..."
    if [ -f ./scripts/sign-macos.sh ]; then
        ./scripts/sign-macos.sh {{bin_dir}}/{{binary_name}}-darwin-amd64
        ./scripts/sign-macos.sh {{bin_dir}}/{{binary_name}}-darwin-arm64
    else
        echo "Warning: scripts/sign-macos.sh not found, using fallback signing"
        codesign --force --sign - {{bin_dir}}/{{binary_name}}-darwin-amd64 2>/dev/null || true
        codesign --force --sign - {{bin_dir}}/{{binary_name}}-darwin-arm64 2>/dev/null || true
    fi

# Create macOS release packages
release-darwin: build-darwin sign-release
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Creating macOS release package..."
    mkdir -p dist
    tar -czf dist/{{binary_name}}-darwin-amd64.tar.gz -C {{bin_dir}} {{binary_name}}-darwin-amd64
    tar -czf dist/{{binary_name}}-darwin-arm64.tar.gz -C {{bin_dir}} {{binary_name}}-darwin-arm64
    echo "Release packages created in dist/"

# Create Linux release package
release-linux: build-linux
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Creating Linux release package..."
    mkdir -p dist
    tar -czf dist/{{binary_name}}-linux-amd64.tar.gz -C {{bin_dir}} {{binary_name}}-linux-amd64

# Create Windows release package
release-windows: build-windows
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Creating Windows release package..."
    mkdir -p dist
    cd {{bin_dir}} && zip -q ../dist/{{binary_name}}-windows-amd64.zip {{binary_name}}-windows-amd64.exe

# Create all release packages
release: release-darwin release-linux release-windows
    #!/usr/bin/env bash
    set -euo pipefail
    echo "All release packages created in dist/"
    ls -la dist/
