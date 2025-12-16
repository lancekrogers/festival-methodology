# Build recipes
root := source_directory() / ".."
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

[private]
@default:
    @just --list --unsorted --justfile {{source_file()}}

# Build fest binary
build:
    cd {{root}} && go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Quick development build (no optimization)
dev:
    cd {{root}} && go build -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Build for Linux AMD64
linux:
    cd {{root}} && GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-linux-amd64 {{main_path}}

# Build for macOS Intel
darwin:
    cd {{root}} && GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-amd64 {{main_path}}

# Build for macOS Apple Silicon
darwin-arm64:
    cd {{root}} && GOOS=darwin GOARCH=arm64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-darwin-arm64 {{main_path}}

# Build for Windows AMD64
windows:
    cd {{root}} && GOOS=windows GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}}-windows-amd64.exe {{main_path}}

# Build for all platforms
all: linux darwin darwin-arm64 windows

# Build Linux binary for container testing
test-binary:
    cd {{root}} && GOOS=linux GOARCH=amd64 go build -ldflags '-s -w' -o {{bin_dir}}/linux/{{binary_name}} {{main_path}}
