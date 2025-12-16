# Quality and maintenance recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
bin_dir := "bin"

[private]
@default:
    @echo "Quality Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Format Go code
fmt:
    go fmt ./...

# Run go vet
vet:
    go vet ./...

# Run formatting and vetting
lint: fmt vet
    @echo "Linting complete"

# Clean build artifacts and generated files (preserve .gitkeep)
clean:
    go clean
    rm -f {{bin_dir}}/{{binary_name}}*
    rm -f coverage.out coverage.html

# Update and tidy dependencies
deps:
    go get -u ./...
    go mod tidy
