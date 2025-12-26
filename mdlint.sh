#!/usr/bin/env bash
# Markdown linting for festival-methodology
# Usage: ./mdlint.sh [directory] [additional args]
# Examples:
#   ./mdlint.sh              # Lint fest/ directory (default)
#   ./mdlint.sh fest         # Lint fest/ directory
#   ./mdlint.sh .            # Lint entire repo
#   ./mdlint.sh festivals    # Lint festivals/ directory
set -euo pipefail

if ! command -v markdownlint-cli2 &> /dev/null; then
    echo "Error: markdownlint-cli2 not found"
    echo "Install: npm install -g markdownlint-cli2"
    exit 1
fi

# Change to script directory to ensure consistent paths
cd "$(dirname "$0")"

# Default to fest/ directory if no argument provided
TARGET="${1:-fest}"
shift 2>/dev/null || true

markdownlint-cli2 "$TARGET/**/*.md" "$@"
