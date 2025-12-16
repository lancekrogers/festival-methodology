# Install recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
main_path := "cmd/fest/main.go"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")
bin_dir := "bin"

[private]
@default:
    @echo "Install Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Build fest binary (private helper)
[private]
build:
    go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Install fest to $GOBIN
install: build
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Installing fest..."
    mkdir -p {{gobin}}
    cp bin/{{binary_name}} {{gobin}}/{{binary_name}}
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "Signing fest binary for macOS..."
        codesign --force --sign - {{gobin}}/{{binary_name}} 2>/dev/null || \
        echo "Warning: Could not sign binary (non-fatal)"
    fi
    echo "fest installed to {{gobin}}/{{binary_name}}"

# Uninstall fest from $GOBIN
uninstall:
    #!/usr/bin/env bash
    set -euo pipefail
    echo "Uninstalling fest..."
    if [ -f {{gobin}}/{{binary_name}} ]; then
        rm {{gobin}}/{{binary_name}}
        echo "fest uninstalled from {{gobin}}"
    else
        echo "fest not found in {{gobin}}"
    fi
