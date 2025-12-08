#!/bin/bash
# Installation script for fest CLI

set -euo pipefail

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

# Map architecture names
case "$ARCH" in
    x86_64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

# Set binary name based on OS
case "$OS" in
    darwin)
        BINARY="fest-darwin-${ARCH}"
        ;;
    linux)
        BINARY="fest-linux-${ARCH}"
        ;;
    *)
        echo "Unsupported OS: $OS"
        exit 1
        ;;
esac

# GitHub release URL
GITHUB_REPO="lancekrogers/festival-methodology"
VERSION=${1:-latest}

if [ "$VERSION" = "latest" ]; then
    # Get latest release version from GitHub API
    VERSION=$(curl -s "https://api.github.com/repos/${GITHUB_REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
fi

echo "Installing fest ${VERSION} for ${OS}/${ARCH}..."

# Download URL
DOWNLOAD_URL="https://github.com/${GITHUB_REPO}/releases/download/${VERSION}/fest-${VERSION}-${OS}-${ARCH}.tar.gz"

# Download and extract
TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

echo "Downloading from ${DOWNLOAD_URL}..."
curl -L -o "${TEMP_DIR}/fest.tar.gz" "${DOWNLOAD_URL}"

echo "Extracting..."
tar -xzf "${TEMP_DIR}/fest.tar.gz" -C "${TEMP_DIR}"

# Install to /usr/local/bin (may require sudo)
INSTALL_PATH="/usr/local/bin/fest"

if [ -w "/usr/local/bin" ]; then
    mv "${TEMP_DIR}/fest" "${INSTALL_PATH}"
else
    echo "Installing to /usr/local/bin requires sudo access..."
    sudo mv "${TEMP_DIR}/fest" "${INSTALL_PATH}"
fi

chmod +x "${INSTALL_PATH}"

echo "✅ fest installed successfully to ${INSTALL_PATH}"
echo "Run 'fest --version' to verify the installation"