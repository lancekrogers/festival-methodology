# !/bin/bash

# Release script for fest CLI

set -euo pipefail

VERSION=${1:-}

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v0.1.0"
    exit 1
fi

echo "🚀 Releasing fest ${VERSION}"

# Ensure working directory is clean

if [ -n "$(git status --porcelain)" ]; then
    echo "❌ Working directory is not clean. Commit or stash changes first."
    exit 1
fi

# Tag the release

echo "📦 Creating git tag ${VERSION}"
git tag -a "${VERSION}" -m "Release ${VERSION}"

# Build release binaries

echo "🔨 Building release binaries..."
just build-all

# Create release archives

echo "📦 Creating release archives..."
mkdir -p dist

# macOS ARM64

tar -czf "dist/fest-${VERSION}-darwin-arm64.tar.gz" -C bin fest-darwin-arm64

# macOS Intel

tar -czf "dist/fest-${VERSION}-darwin-amd64.tar.gz" -C bin fest-darwin-amd64

# Linux

tar -czf "dist/fest-${VERSION}-linux-amd64.tar.gz" -C bin fest-linux-amd64

# Windows

(cd bin && zip -q "../dist/fest-${VERSION}-windows-amd64.zip" fest-windows-amd64.exe)

# Generate checksums

echo "🔐 Generating checksums..."
(cd dist && shasum -a 256 *.tar.gz*.zip > checksums.txt)

echo "✅ Release ${VERSION} prepared in dist/"
echo ""
echo "Next steps:"
echo "1. Push tag: git push origin ${VERSION}"
echo "2. Create GitHub release with the archives in dist/"
echo "3. Update Homebrew formula with new URLs and SHA256 checksums"
echo ""
echo "SHA256 checksums:"
cat dist/checksums.txt
