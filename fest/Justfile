#!/usr/bin/env just --justfile
# fest CLI build and development tasks

set dotenv-load := true

# Configuration
binary_name := "fest"
bin_dir := "bin"
gobin := env_var_or_default("GOBIN", `go env GOPATH` + "/bin")
BUILDTOOL := "go run ./internal/buildutil"

# Modules
[doc('Cross-platform builds')]
mod xbuild '.justfiles/build.just'

[doc('Testing (unit, integration, coverage)')]
mod test '.justfiles/test.just'

[doc('Release packaging')]
mod release '.justfiles/release.just'

[private]
default:
    #!/usr/bin/env bash
    echo "fest CLI - Festival Methodology Tool"
    echo ""
    just --list --unsorted

# Build fest binary with visual dashboard
build:
    @{{BUILDTOOL}} build

# Build fest binary only (fast, no vet)
build-only:
    @{{BUILDTOOL}} build-only

# Format Go code
fmt:
    go fmt ./...

# Run go vet
vet:
    go vet ./...

# Run formatting and vetting
lint: fmt vet
    @echo "✅ Linting complete"

# Clean build artifacts with visual dashboard
clean:
    @{{BUILDTOOL}} clean

# Update and tidy dependencies
deps:
    go get -u ./...
    go mod tidy

# Install fest to $GOBIN
install:
    #!/usr/bin/env bash
    set -euo pipefail
    just build-only
    echo "Installing fest..."
    mkdir -p {{gobin}}
    cp bin/{{binary_name}} {{gobin}}/{{binary_name}}
    if [[ "$(uname)" == "Darwin" ]]; then
        echo "Signing fest binary for macOS..."
        codesign --force --sign - {{gobin}}/{{binary_name}} 2>/dev/null || \
        echo "Warning: Could not sign binary (non-fatal)"
    fi
    echo "✅ fest installed to {{gobin}}/{{binary_name}}"

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
