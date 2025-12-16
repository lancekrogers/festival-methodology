# Release recipes
root := justfile_directory()
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

# Show available release commands (default)
[private]
default:
    @just --list --unsorted --justfile {{source_file()}}

# Build for macOS Intel
[private]
build-darwin:
    cd {{root}} && GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

# Build for macOS Apple Silicon
[private]
build-darwin-arm64:
    cd {{root}} && GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

# Build for Linux AMD64
[private]
build-linux:
    cd {{root}} && GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

# Build for Windows AMD64
[private]
build-windows:
    cd {{root}} && GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-windows-amd64.exe {{main_path}}

# Sign macOS release binaries
sign: build-darwin build-darwin-arm64
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{root}}
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
darwin: build-darwin sign
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{root}}
    echo "Creating macOS release package..."
    mkdir -p dist
    tar -czf dist/{{binary_name}}-darwin-amd64.tar.gz -C {{bin_dir}} {{binary_name}}-darwin-amd64
    tar -czf dist/{{binary_name}}-darwin-arm64.tar.gz -C {{bin_dir}} {{binary_name}}-darwin-arm64
    echo "Release packages created in dist/"

# Create Linux release package
linux: build-linux
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{root}}
    echo "Creating Linux release package..."
    mkdir -p dist
    tar -czf dist/{{binary_name}}-linux-amd64.tar.gz -C {{bin_dir}} {{binary_name}}-linux-amd64

# Create Windows release package
windows: build-windows
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{root}}
    echo "Creating Windows release package..."
    mkdir -p dist
    cd {{bin_dir}} && zip -q ../dist/{{binary_name}}-windows-amd64.zip {{binary_name}}-windows-amd64.exe

# Create all release packages
release: darwin linux windows
    #!/usr/bin/env bash
    set -euo pipefail
    cd {{root}}
    echo "All release packages created in dist/"
    ls -la dist/
