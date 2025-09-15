#!/bin/bash
# Sign macOS binary for distribution
# This script handles code signing for macOS binaries to avoid Gatekeeper warnings

set -e

BINARY_NAME="${1:-fest}"
IDENTITY="${CODESIGN_IDENTITY:-}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}macOS Binary Signing Script${NC}"
echo "================================"

# Check if binary exists
if [ ! -f "$BINARY_NAME" ]; then
    echo -e "${RED}Error: Binary '$BINARY_NAME' not found${NC}"
    echo "Please build the binary first with: make build"
    exit 1
fi

# Check if running on macOS
if [[ "$OSTYPE" != "darwin"* ]]; then
    echo -e "${YELLOW}Warning: Not running on macOS, skipping signing${NC}"
    exit 0
fi

# Check if codesign is available
if ! command -v codesign &> /dev/null; then
    echo -e "${RED}Error: codesign command not found${NC}"
    echo "This script must be run on macOS with Xcode Command Line Tools installed"
    exit 1
fi

# Ad-hoc signing if no identity provided
if [ -z "$IDENTITY" ]; then
    echo -e "${YELLOW}No signing identity provided, using ad-hoc signing${NC}"
    echo "For distribution, set CODESIGN_IDENTITY environment variable"
    
    # Ad-hoc sign the binary
    echo "Signing $BINARY_NAME with ad-hoc signature..."
    codesign --force --deep --sign - "$BINARY_NAME"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Binary signed successfully (ad-hoc)${NC}"
    else
        echo -e "${RED}✗ Failed to sign binary${NC}"
        exit 1
    fi
else
    # Sign with provided identity
    echo "Signing $BINARY_NAME with identity: $IDENTITY"
    
    # Sign the binary with hardened runtime for notarization
    codesign --force \
             --deep \
             --sign "$IDENTITY" \
             --options runtime \
             --timestamp \
             "$BINARY_NAME"
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Binary signed successfully with identity${NC}"
    else
        echo -e "${RED}✗ Failed to sign binary with identity${NC}"
        exit 1
    fi
fi

# Verify the signature
echo "Verifying signature..."
codesign --verify --deep --strict --verbose=2 "$BINARY_NAME"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Signature verification passed${NC}"
    
    # Display signature info
    echo ""
    echo "Signature Information:"
    codesign --display --verbose=2 "$BINARY_NAME" 2>&1 | grep -E "^(Identifier|Format|Signature|Info.plist)"
    
    # Check if binary will pass Gatekeeper
    echo ""
    echo "Checking Gatekeeper assessment..."
    spctl --assess --verbose=4 --type execute "$BINARY_NAME" 2>&1 || true
    
else
    echo -e "${RED}✗ Signature verification failed${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}Signing complete!${NC}"
echo ""
echo "Next steps for distribution:"
echo "1. For personal use: The ad-hoc signed binary can be distributed"
echo "2. For wider distribution: Use a Developer ID certificate"
echo "3. For App Store: Submit for notarization with Apple"
echo ""
echo "To avoid Gatekeeper warnings for users:"
echo "  - Users can right-click and select 'Open' on first run"
echo "  - Or remove quarantine: xattr -d com.apple.quarantine $BINARY_NAME"