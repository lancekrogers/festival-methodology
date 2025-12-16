# Development run recipes for fest CLI

# Set working directory to project root
set working-directory := '..'

# Configuration
binary_name := "fest"
main_path := "cmd/fest/main.go"
bin_dir := "bin"

[private]
@default:
    @echo "Run Commands"
    @just --list --unsorted --justfile {{source_file()}}

# Build fest binary (private helper)
[private]
build:
    go build -ldflags '-s -w' -o {{bin_dir}}/{{binary_name}} {{main_path}}

# Show fest help
run-help: build
    ./{{bin_dir}}/{{binary_name}} --help

# Initialize a new festival
run-init: build
    ./{{bin_dir}}/{{binary_name}} init

# Sync festival state
run-sync: build
    ./{{bin_dir}}/{{binary_name}} sync

# Update festival
run-update: build
    ./{{bin_dir}}/{{binary_name}} update
